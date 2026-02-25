package safety

import (
	"fmt"
	"strings"
	"unicode/utf8"
)

const (
	// MaxUserMessageRunes is an upper bound for free-form user text such as interview messages.
	MaxUserMessageRunes = 4000
)

// SanitizeUserMessage trims whitespace, strips control characters (except newlines and tabs),
// and enforces a maximum length. It returns an error if the content is too long.
func SanitizeUserMessage(s string) (string, error) {
	trimmed := strings.TrimSpace(s)
	if trimmed == "" {
		return trimmed, nil
	}

	if utf8.RuneCountInString(trimmed) > MaxUserMessageRunes {
		return "", fmt.Errorf("content is too long; maximum is %d characters", MaxUserMessageRunes)
	}

	var b strings.Builder
	b.Grow(len(trimmed))
	for _, r := range trimmed {
		// Allow printable characters plus common whitespace; drop other control chars.
		if r == '\n' || r == '\r' || r == '\t' || (r >= 32 && r != 127) {
			b.WriteRune(r)
		}
	}
	return b.String(), nil
}

