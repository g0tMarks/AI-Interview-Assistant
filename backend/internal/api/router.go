// internal/api/router.go
package api

import (
	"net/http"

	"github.com/g0tMarks/AI-Interview-Assistant/backend/internal/api/handlers"
	"github.com/go-chi/chi/v5"
)

// NewRouter builds the chi router and registers all routes.
func NewRouter(deps Dependencies) http.Handler {
	r := chi.NewRouter()

	// Middlewares (logging, recover, etc.) can go here later.
	// r.Use(middleware.Logger)
	// r.Use(middleware.Recoverer)

	healthHandler := handlers.NewHealthHandler()
	rubricHandler := handlers.NewRubricHandler(deps.Queries)

	r.Get("/health", healthHandler.Health)
	r.Post("/rubrics", rubricHandler.CreateRubric)

	return r
}
