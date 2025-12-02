// internal/api/handlers/health.go
package handlers

import (
	"net/http"
)

// HealthHandler holds dependencies for health-related endpoints.
// Empty for now, but you can add logger, build info, etc. later.
type HealthHandler struct {
	// Logger *slog.Logger
}

// NewHealthHandler constructs a HealthHandler.
func NewHealthHandler() *HealthHandler {
	return &HealthHandler{}
}

// Health is a basic liveness endpoint.
// GET /health -> "Hello, World!"
func (h *HealthHandler) Health(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("Hello, World!"))
}
