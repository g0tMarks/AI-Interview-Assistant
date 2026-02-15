package middleware

import (
	"net/http"
	"strings"

	"github.com/g0tMarks/AI-Interview-Assistant/backend/internal/auth"
)

// RequireStudentAuth returns a middleware that validates a Bearer JWT and sets the student ID on the request context.
// If the token is missing or invalid, it responds with 401 Unauthorized.
func RequireStudentAuth(jwtSecret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authz := r.Header.Get("Authorization")
			if authz == "" {
				http.Error(w, "missing Authorization header", http.StatusUnauthorized)
				return
			}
			const prefix = "Bearer "
			if !strings.HasPrefix(authz, prefix) {
				http.Error(w, "invalid Authorization header", http.StatusUnauthorized)
				return
			}
			tokenString := strings.TrimSpace(authz[len(prefix):])
			if tokenString == "" {
				http.Error(w, "missing token", http.StatusUnauthorized)
				return
			}
			claims, err := auth.ValidateStudentToken(jwtSecret, tokenString)
			if err != nil {
				http.Error(w, "invalid or expired token", http.StatusUnauthorized)
				return
			}
			studentID, err := auth.StudentIDFromClaims(claims)
			if err != nil {
				http.Error(w, "invalid token claims", http.StatusUnauthorized)
				return
			}
			ctx := auth.WithStudentID(r.Context(), studentID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
