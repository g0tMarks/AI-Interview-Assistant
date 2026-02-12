package validation

import (
	"errors"
	"net/mail"
	"strings"
)

// ValidateEmail validates email format according to RFC 5322.
// Returns an error if the email format is invalid, nil otherwise.
func ValidateEmail(email string) error {
	if email == "" {
		return errors.New("email is required")
	}

	// Trim whitespace
	email = strings.TrimSpace(email)

	if email == "" {
		return errors.New("email cannot be empty")
	}

	// Use Go's net/mail package for RFC 5322 compliant validation
	_, err := mail.ParseAddress(email)
	if err != nil {
		return errors.New("invalid email format")
	}

	return nil
}


