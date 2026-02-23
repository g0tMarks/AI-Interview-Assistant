// internal/api/server.go
package api

import (
	"context"
	"net/http"

	"github.com/g0tMarks/AI-Interview-Assistant/backend/internal/db"
	"github.com/g0tMarks/AI-Interview-Assistant/backend/internal/services"
	"github.com/g0tMarks/AI-Interview-Assistant/backend/internal/storage"
	"github.com/jackc/pgx/v5"
)

// TxBeginner is implemented by *pgx.Conn for starting transactions.
type TxBeginner interface {
	Begin(ctx context.Context) (pgx.Tx, error)
}

// Dependencies holds things your handlers need (DB, logger, config, etc.).
type Dependencies struct {
	Queries         *db.Queries
	LLMService      services.LLMService
	TxBeginner      TxBeginner       // for handlers that need transactions (e.g. rubric parse)
	JWTSecret       string           // used for signing/validating student JWTs (e.g. from JWT_SECRET env)
	Storage         storage.Store
	UploadsMaxBytes int64 // max upload size for multipart uploads (bytes); if 0, handler default applies
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
