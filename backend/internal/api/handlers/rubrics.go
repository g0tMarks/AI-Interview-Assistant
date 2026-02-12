package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/g0tMarks/AI-Interview-Assistant/backend/internal/db"
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
	RawText     string    `json:"rawText"`
}

type RubricResponse struct {
	RubricID    uuid.UUID          `json:"rubricId"`
	TeacherID   uuid.UUID          `json:"teacherId"`
	Title       string             `json:"title"`
	Description string             `json:"description"`
	RawText     string             `json:"rawText"`
	IsEnabled   bool               `json:"isEnabled"`
	CreatedAt   pgtype.Timestamptz `json:"createdAt"`
	UpdatedAt   pgtype.Timestamptz `json:"updatedAt"`
}

func (h *RubricHandler) CreateRubric(w http.ResponseWriter, r *http.Request) {
	var req CreateRubricRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid JSON body", http.StatusBadRequest)
		return
	}

	// Minimal validation: we just ensure there's enough info to be useful
	if req.Title == "" {
		http.Error(w, "title is required", http.StatusBadRequest)
		return
	}
	if req.TeacherID == uuid.Nil {
		http.Error(w, "teacherId is required", http.StatusBadRequest)
		return
	}
	if req.RawText == "" {
		http.Error(w, "rawText is required", http.StatusBadRequest)
		return
	}

	ctx := r.Context()

	// Convert uuid.UUID to pgtype.UUID
	teacherID := pgtype.UUID{
		Bytes: req.TeacherID,
		Valid: true,
	}

	// Convert string to pgtype.Text for description
	description := pgtype.Text{}
	if req.Description != "" {
		description.String = req.Description
		description.Valid = true
	}

	rubric, err := h.q.CreateRubric(ctx, db.CreateRubricParams{
		TeacherID:   teacherID,
		Title:       req.Title,
		Description: description,
		RawText:     req.RawText,
	})
	if err != nil {
		// Log the actual error for debugging (in production, be more careful about exposing errors)
		http.Error(w, "failed to save rubric: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Convert pgtype.UUID to uuid.UUID for response
	var rubricID uuid.UUID
	if rubric.RubricID.Valid {
		rubricID = rubric.RubricID.Bytes
	}

	var teacherIDResp uuid.UUID
	if rubric.TeacherID.Valid {
		teacherIDResp = rubric.TeacherID.Bytes
	}

	resp := RubricResponse{
		RubricID:    rubricID,
		TeacherID:   teacherIDResp,
		Title:       rubric.Title,
		Description: rubric.Description.String,
		RawText:     rubric.RawText,
		IsEnabled:   rubric.IsEnabled,
		CreatedAt:   rubric.CreatedAt,
		UpdatedAt:   rubric.UpdatedAt,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(resp)
}

func (h *RubricHandler) ListRubrics(w http.ResponseWriter, r *http.Request) {
	// Extract teacherId from query parameter
	teacherIdStr := r.URL.Query().Get("teacherId")
	if teacherIdStr == "" {
		http.Error(w, "teacherId is required", http.StatusBadRequest)
		return
	}

	// Validate UUID format
	teacherID, err := uuid.Parse(teacherIdStr)
	if err != nil {
		http.Error(w, "invalid teacherId format", http.StatusBadRequest)
		return
	}

	ctx := r.Context()

	// Convert to pgtype.UUID
	teacherIDPgtype := pgtype.UUID{
		Bytes: teacherID,
		Valid: true,
	}

	// Query database
	rubrics, err := h.q.ListRubricsByTeacher(ctx, teacherIDPgtype)
	if err != nil {
		http.Error(w, "failed to retrieve rubrics", http.StatusInternalServerError)
		return
	}

	// Convert to response format
	resp := make([]RubricResponse, len(rubrics))
	for i, rubric := range rubrics {
		var rubricID uuid.UUID
		if rubric.RubricID.Valid {
			rubricID = rubric.RubricID.Bytes
		}

		var teacherIDResp uuid.UUID
		if rubric.TeacherID.Valid {
			teacherIDResp = rubric.TeacherID.Bytes
		}

		resp[i] = RubricResponse{
			RubricID:    rubricID,
			TeacherID:   teacherIDResp,
			Title:       rubric.Title,
			Description: rubric.Description.String,
			RawText:     rubric.RawText,
			IsEnabled:   rubric.IsEnabled,
			CreatedAt:   rubric.CreatedAt,
			UpdatedAt:   rubric.UpdatedAt,
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(resp)
}
