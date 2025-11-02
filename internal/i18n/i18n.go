package i18n

import (
	"embed"
	"encoding/json"
	"fmt"
	"sync"
)

type Locale string

const (
	LocaleEN Locale = "en"
	LocaleDE Locale = "de"
)

//go:embed locales/*.json
var localeFS embed.FS

var (
	supportedLocales = []Locale{LocaleEN, LocaleDE}
	mu               sync.RWMutex
	currentLocale    = LocaleEN
	messages         = make(map[string]string)
)

func init() {
	if err := load(LocaleEN); err != nil {
		messages = make(map[string]string)
	}
}

func load(loc Locale) error {
	data, err := localeFS.ReadFile(fmt.Sprintf("locales/%s.json", loc))
	if err != nil {
		return err
	}
	tmp := make(map[string]string)
	if err := json.Unmarshal(data, &tmp); err != nil {
		return err
	}
	mu.Lock()
	messages = tmp
	currentLocale = loc
	mu.Unlock()
	return nil
}

func SetLocale(loc Locale) error {
	mu.RLock()
	if loc == currentLocale {
		mu.RUnlock()
		return nil
	}
	mu.RUnlock()
	return load(loc)
}

func Current() Locale {
	mu.RLock()
	defer mu.RUnlock()
	return currentLocale
}

func Supported() []Locale {
	out := append([]Locale(nil), supportedLocales...)
	return out
}

func T(key string, args ...any) string {
	mu.RLock()
	msg, ok := messages[key]
	mu.RUnlock()
	if !ok {
		msg = key
	}
	if len(args) > 0 {
		return fmt.Sprintf(msg, args...)
	}
	return msg
}

func DisplayName(loc Locale) string {
	switch loc {
	case LocaleEN:
		return T("language.english")
	case LocaleDE:
		return T("language.german")
	default:
		return string(loc)
	}
}
