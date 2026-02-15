package auth

import (
	"context"

	"github.com/google/uuid"
)

type contextKey int

const (
	contextKeyStudentID contextKey = iota
)

// StudentIDFromContext returns the authenticated student's ID from the request context.
// Returns (zero UUID, false) if not set (e.g. no valid JWT).
func StudentIDFromContext(ctx context.Context) (uuid.UUID, bool) {
	v := ctx.Value(contextKeyStudentID)
	if v == nil {
		return uuid.Nil, false
	}
	id, ok := v.(uuid.UUID)
	return id, ok
}

// WithStudentID sets the student ID on the context. Used by auth middleware.
func WithStudentID(ctx context.Context, studentID uuid.UUID) context.Context {
	return context.WithValue(ctx, contextKeyStudentID, studentID)
}
