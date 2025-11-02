package id

import (
	"crypto/rand"
	"encoding/hex"
	"strings"
	"time"
)

// New returns a 32 character hex encoded random identifier.
func New() string {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		// rand.Read should not fail, but fall back to time based entropy just in case.
		now := time.Now().UnixNano()
		return strings.ToUpper(hex.EncodeToString([]byte{
			byte(now >> 56), byte(now >> 48), byte(now >> 40), byte(now >> 32),
			byte(now >> 24), byte(now >> 16), byte(now >> 8), byte(now),
		}))
	}
	return strings.ToUpper(hex.EncodeToString(b[:]))
}

// Short returns a slightly shorter identifier suitable for invoice numbers.
func Short() string {
	var b [6]byte
	if _, err := rand.Read(b[:]); err != nil {
		now := time.Now().UnixNano()
		return strings.ToUpper(hex.EncodeToString([]byte{
			byte(now >> 8), byte(now),
		}))
	}
	return strings.ToUpper(hex.EncodeToString(b[:]))
}
