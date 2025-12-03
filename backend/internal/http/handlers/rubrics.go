package handlers

import (
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/g0tMarks/AI-Interview-Assistant/backend/internal/db" // sqlc package
)

type RubricHandler struct {
	q *db.Queries
}

func NewRubricHandler(q *db.Queries) *RubricHandler {
	return &RubricHandler{q: q}
}

type CreateRubricRequest struct {
	TeacherID   uuid.UUID `json:"teacherId"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
}

type RubricResponse struct {
	RubricID    uuid.UUID          `json:"rubricId"`
	TeacherID   uuid.UUID          `json:"teacherId"`
	Title       string             `json:"title"`
	Description string             `json:"description"`
	CreatedAt   pgtype.Timestamptz `json:"createdAt"`
}
