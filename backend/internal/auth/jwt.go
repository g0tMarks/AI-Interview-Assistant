package auth

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

const (
	RoleStudent = "student"
	// DefaultStudentTokenExpiry is the duration a student JWT is valid.
	DefaultStudentTokenExpiry = 24 * time.Hour
)

// StudentClaims holds JWT claims for an authenticated student.
// Subject (sub) is the student ID as string.
type StudentClaims struct {
	jwt.RegisteredClaims
	Role string `json:"role"`
}

// IssueStudentToken creates a signed JWT for the given student ID.
// secret must be a non-empty key (e.g. from JWT_SECRET env).
func IssueStudentToken(secret string, studentID uuid.UUID, expiry time.Duration) (string, error) {
	if secret == "" {
		return "", errors.New("JWT secret is required")
	}
	if expiry <= 0 {
		expiry = DefaultStudentTokenExpiry
	}
	now := time.Now()
	claims := StudentClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   studentID.String(),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(expiry)),
			ID:        uuid.New().String(),
		},
		Role: RoleStudent,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

// ValidateStudentToken parses and validates a JWT string and returns the student claims.
func ValidateStudentToken(secret string, tokenString string) (*StudentClaims, error) {
	if secret == "" {
		return nil, errors.New("JWT secret is required")
	}
	token, err := jwt.ParseWithClaims(tokenString, &StudentClaims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(secret), nil
	})
	if err != nil {
		return nil, err
	}
	claims, ok := token.Claims.(*StudentClaims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token claims")
	}
	if claims.Role != RoleStudent {
		return nil, errors.New("invalid role")
	}
	return claims, nil
}

// StudentIDFromClaims returns the student UUID from the token subject.
func StudentIDFromClaims(claims *StudentClaims) (uuid.UUID, error) {
	return uuid.Parse(claims.Subject)
}
