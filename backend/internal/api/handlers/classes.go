package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/g0tMarks/AI-Interview-Assistant/backend/internal/db"
)

type ClassHandler struct {
	q *db.Queries
}

func NewClassHandler(q *db.Queries) *ClassHandler {
	return &ClassHandler{q: q}
}

type CreateClassRequest struct {
	TeacherID uuid.UUID `json:"teacherId"`
	Name      string    `json:"name"`
	ClassCode string    `json:"classCode"`
}

type ClassResponse struct {
	ClassID   uuid.UUID          `json:"classId"`
	TeacherID uuid.UUID          `json:"teacherId"`
	Name      string             `json:"name"`
	ClassCode string             `json:"classCode"`
	CreatedAt pgtype.Timestamptz `json:"createdAt"`
	UpdatedAt pgtype.Timestamptz `json:"updatedAt"`
}

type UpdateClassRequest struct {
	Name      string `json:"name"`
	ClassCode string `json:"classCode"`
}

func classToResponse(c db.AppClass) ClassResponse {
	var classID, teacherID uuid.UUID
	if c.ClassID.Valid {
		classID = c.ClassID.Bytes
	}
	if c.TeacherID.Valid {
		teacherID = c.TeacherID.Bytes
	}
	return ClassResponse{
		ClassID:   classID,
		TeacherID: teacherID,
		Name:      c.Name,
		ClassCode: c.ClassCode,
		CreatedAt: c.CreatedAt,
		UpdatedAt: c.UpdatedAt,
	}
}

func (h *ClassHandler) CreateClass(w http.ResponseWriter, r *http.Request) {
	var req CreateClassRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid JSON body", http.StatusBadRequest)
		return
	}
	req.Name = strings.TrimSpace(req.Name)
	req.ClassCode = strings.TrimSpace(req.ClassCode)
	if req.TeacherID == uuid.Nil {
		http.Error(w, "teacherId is required", http.StatusBadRequest)
		return
	}
	if req.Name == "" {
		http.Error(w, "name is required", http.StatusBadRequest)
		return
	}
	if req.ClassCode == "" {
		http.Error(w, "classCode is required", http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	teacherID := pgtype.UUID{Bytes: req.TeacherID, Valid: true}
	class, err := h.q.CreateClass(ctx, db.CreateClassParams{
		TeacherID: teacherID,
		Name:      req.Name,
		ClassCode: req.ClassCode,
	})
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			http.Error(w, "classCode is already in use", http.StatusConflict)
			return
		}
		http.Error(w, "failed to create class", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(classToResponse(class))
}

func (h *ClassHandler) GetClass(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "invalid class id", http.StatusBadRequest)
		return
	}
	ctx := r.Context()
	pid := pgtype.UUID{Bytes: id, Valid: true}
	class, err := h.q.GetClassByID(ctx, pid)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			http.Error(w, "class not found", http.StatusNotFound)
			return
		}
		http.Error(w, "failed to get class", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(classToResponse(class))
}

func (h *ClassHandler) ListClasses(w http.ResponseWriter, r *http.Request) {
	teacherIDStr := r.URL.Query().Get("teacherId")
	if teacherIDStr == "" {
		http.Error(w, "teacherId query parameter is required", http.StatusBadRequest)
		return
	}
	teacherID, err := uuid.Parse(teacherIDStr)
	if err != nil {
		http.Error(w, "invalid teacherId", http.StatusBadRequest)
		return
	}
	ctx := r.Context()
	tid := pgtype.UUID{Bytes: teacherID, Valid: true}
	classes, err := h.q.ListClassesByTeacher(ctx, tid)
	if err != nil {
		http.Error(w, "failed to list classes", http.StatusInternalServerError)
		return
	}
	out := make([]ClassResponse, len(classes))
	for i := range classes {
		out[i] = classToResponse(classes[i])
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(out)
}

func (h *ClassHandler) UpdateClass(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "invalid class id", http.StatusBadRequest)
		return
	}
	var req UpdateClassRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid JSON body", http.StatusBadRequest)
		return
	}
	req.Name = strings.TrimSpace(req.Name)
	req.ClassCode = strings.TrimSpace(req.ClassCode)
	if req.Name == "" {
		http.Error(w, "name is required", http.StatusBadRequest)
		return
	}
	if req.ClassCode == "" {
		http.Error(w, "classCode is required", http.StatusBadRequest)
		return
	}
	ctx := r.Context()
	pid := pgtype.UUID{Bytes: id, Valid: true}
	class, err := h.q.UpdateClass(ctx, db.UpdateClassParams{
		Name:      req.Name,
		ClassCode: req.ClassCode,
		ClassID:   pid,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			http.Error(w, "class not found", http.StatusNotFound)
			return
		}
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			http.Error(w, "classCode is already in use", http.StatusConflict)
			return
		}
		http.Error(w, "failed to update class", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(classToResponse(class))
}

func (h *ClassHandler) DeleteClass(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "invalid class id", http.StatusBadRequest)
		return
	}
	ctx := r.Context()
	pid := pgtype.UUID{Bytes: id, Valid: true}
	err = h.q.DeleteClass(ctx, pid)
	if err != nil {
		http.Error(w, "failed to delete class", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
