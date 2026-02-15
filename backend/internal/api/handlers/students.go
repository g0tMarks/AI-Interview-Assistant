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
	"github.com/g0tMarks/AI-Interview-Assistant/backend/internal/validation"
)

type StudentHandler struct {
	q *db.Queries
}

func NewStudentHandler(q *db.Queries) *StudentHandler {
	return &StudentHandler{q: q}
}

type CreateStudentRequest struct {
	Email       string `json:"email"`
	DisplayName string `json:"displayName"`
}

type StudentResponse struct {
	StudentID   uuid.UUID          `json:"studentId"`
	Email       string             `json:"email"`
	DisplayName string             `json:"displayName"`
	CreatedAt   pgtype.Timestamptz `json:"createdAt"`
	UpdatedAt   pgtype.Timestamptz `json:"updatedAt"`
}

type UpdateStudentRequest struct {
	DisplayName string `json:"displayName"`
}

func studentToResponse(s db.AppStudent) StudentResponse {
	var studentID uuid.UUID
	if s.StudentID.Valid {
		studentID = s.StudentID.Bytes
	}
	return StudentResponse{
		StudentID:   studentID,
		Email:       s.Email,
		DisplayName: s.DisplayName,
		CreatedAt:   s.CreatedAt,
		UpdatedAt:   s.UpdatedAt,
	}
}

func (h *StudentHandler) CreateStudent(w http.ResponseWriter, r *http.Request) {
	var req CreateStudentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid JSON body", http.StatusBadRequest)
		return
	}
	req.Email = strings.TrimSpace(req.Email)
	req.DisplayName = strings.TrimSpace(req.DisplayName)
	if req.Email == "" {
		http.Error(w, "email is required", http.StatusBadRequest)
		return
	}
	if req.DisplayName == "" {
		http.Error(w, "displayName is required", http.StatusBadRequest)
		return
	}
	if err := validation.ValidateEmail(req.Email); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	student, err := h.q.CreateStudent(ctx, db.CreateStudentParams{
		Email:       req.Email,
		DisplayName: req.DisplayName,
	})
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			http.Error(w, "email address is already registered", http.StatusConflict)
			return
		}
		http.Error(w, "failed to create student", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(studentToResponse(student))
}

func (h *StudentHandler) GetStudent(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "invalid student id", http.StatusBadRequest)
		return
	}
	ctx := r.Context()
	pid := pgtype.UUID{Bytes: id, Valid: true}
	student, err := h.q.GetStudentByID(ctx, pid)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			http.Error(w, "student not found", http.StatusNotFound)
			return
		}
		http.Error(w, "failed to get student", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(studentToResponse(student))
}

func (h *StudentHandler) ListStudents(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	students, err := h.q.ListStudents(ctx)
	if err != nil {
		http.Error(w, "failed to list students", http.StatusInternalServerError)
		return
	}
	out := make([]StudentResponse, len(students))
	for i := range students {
		out[i] = studentToResponse(students[i])
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(out)
}

func (h *StudentHandler) UpdateStudent(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "invalid student id", http.StatusBadRequest)
		return
	}
	var req UpdateStudentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid JSON body", http.StatusBadRequest)
		return
	}
	req.DisplayName = strings.TrimSpace(req.DisplayName)
	if req.DisplayName == "" {
		http.Error(w, "displayName is required", http.StatusBadRequest)
		return
	}
	ctx := r.Context()
	pid := pgtype.UUID{Bytes: id, Valid: true}
	student, err := h.q.UpdateStudent(ctx, db.UpdateStudentParams{
		DisplayName: req.DisplayName,
		StudentID:   pid,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			http.Error(w, "student not found", http.StatusNotFound)
			return
		}
		http.Error(w, "failed to update student", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(studentToResponse(student))
}
