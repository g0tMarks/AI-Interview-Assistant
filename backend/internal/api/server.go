// internal/api/server.go
package api

import (
	"net/http"

	"github.com/g0tMarks/AI-Interview-Assistant/backend/internal/db"
)

// Dependencies holds things your handlers need (DB, logger, config, etc.).
// Start empty for the Hello World example and add fields as you go.
type Dependencies struct {
	Queries *db.Queries
	// Logger *slog.Logger
}

// Server wraps the top-level router and implements http.Handler.
type Server struct {
	router http.Handler
}

// NewServer constructs the router and returns a Server.
func NewServer(deps Dependencies) *Server {
	r := NewRouter(deps)
	return &Server{
		router: r,
	}
}

// ServeHTTP makes Server satisfy http.Handler.
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.router.ServeHTTP(w, r)
}
