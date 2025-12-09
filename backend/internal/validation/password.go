package validation

import (
	"errors"
	"regexp"
	"unicode"

	"golang.org/x/crypto/bcrypt"
)

// ValidatePassword validates password strength according to PRD requirements:
// - Minimum 8 characters in length
// - At least one uppercase letter
// - At least one lowercase letter
// - At least one number
// - At least one special character (!@#$%^&*()_+-=[]{}|;:,.<>?)
func ValidatePassword(password string) error {
	if len(password) < 8 {
		return errors.New("password must be at least 8 characters long")
	}

	var (
		hasUpper   = false
		hasLower   = false
		hasNumber  = false
		hasSpecial = false
	)

	specialChars := regexp.MustCompile(`[!@#$%^&*()_+\-=\[\]{}|;:,.<>?]`)

	for _, char := range password {
		switch {
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsLower(char):
			hasLower = true
		case unicode.IsNumber(char):
			hasNumber = true
		case specialChars.MatchString(string(char)):
			hasSpecial = true
		}
	}

	if !hasUpper {
		return errors.New("password must contain at least one uppercase letter")
	}
	if !hasLower {
		return errors.New("password must contain at least one lowercase letter")
	}
	if !hasNumber {
		return errors.New("password must contain at least one number")
	}
	if !hasSpecial {
		return errors.New("password must contain at least one special character (!@#$%^&*()_+-=[]{}|;:,.<>?)")
	}

	return nil
}

// HashPassword hashes a password using bcrypt with cost factor 10.
// Returns the hashed password as a string and any error that occurred.
func HashPassword(password string) (string, error) {
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashedBytes), nil
}
