package locale

import (
	"fmt"
	"strconv"
	"strings"
	"unicode"
)

// NumberParser abstracts locale aware floating point parsing so that different
// strategies (DE, EN, etc.) can be swapped without touching call sites.
type NumberParser interface {
	ParseFloat(input string) (float64, error)
}

var activeNumberParser NumberParser = EuroNumberParser{}

// SetNumberParser allows callers to swap the active parser implementation.
// Passing nil resets the parser to the default German implementation.
func SetNumberParser(p NumberParser) {
	if p == nil {
		activeNumberParser = EuroNumberParser{}
		return
	}
	activeNumberParser = p
}

// ParseFloat delegates to the active parser implementation.
func ParseFloat(input string) (float64, error) {
	return activeNumberParser.ParseFloat(input)
}

// EuroNumberParser accepts common continental European number formats, e.g.:
//
//	1.234,56   -> 1234.56
//	1 234,56   -> 1234.56
//	1234,56    -> 1234.56
//	1234.56    -> 1234.56 (fallback for EN style)
type EuroNumberParser struct{}

func (EuroNumberParser) ParseFloat(input string) (float64, error) {
	normalized, err := normalizeEuropeanNumber(input)
	if err != nil {
		return 0, err
	}
	return strconv.ParseFloat(normalized, 64)
}

func normalizeEuropeanNumber(in string) (string, error) {
	value := strings.TrimSpace(in)
	if value == "" {
		return "0", nil
	}

	sign := ""
	switch {
	case strings.HasPrefix(value, "+"):
		value = strings.TrimPrefix(value, "+")
	case strings.HasPrefix(value, "-"):
		sign = "-"
		value = strings.TrimPrefix(value, "-")
	}

	replacer := strings.NewReplacer(" ", "", "\u00A0", "", "'", "")
	value = replacer.Replace(value)
	if value == "" {
		return "0", nil
	}

	useCommaAsDecimal := strings.Contains(value, ",")
	useDotAsDecimal := !useCommaAsDecimal && strings.Contains(value, ".")
	decimalRune := '.'
	if useCommaAsDecimal {
		decimalRune = ','
	}

	var builder strings.Builder
	decimalSeen := false

	for _, r := range value {
		switch {
		case unicode.IsDigit(r):
			builder.WriteRune(r)
		case r == ',' || r == '.':
			asDecimal := (r == decimalRune) || (!useCommaAsDecimal && useDotAsDecimal && r == '.')
			if asDecimal {
				if decimalSeen {
					return "", fmt.Errorf("multiple decimal separators")
				}
				builder.WriteRune('.')
				decimalSeen = true
				continue
			}
			// treat as thousands separator -> skip
		default:
			return "", fmt.Errorf("invalid numeric character %q", r)
		}
	}

	if builder.Len() == 0 {
		return "0", nil
	}
	return sign + builder.String(), nil
}
