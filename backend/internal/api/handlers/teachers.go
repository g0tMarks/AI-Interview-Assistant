package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/g0tMarks/AI-Interview-Assistant/backend/internal/db"
	"github.com/g0tMarks/AI-Interview-Assistant/backend/internal/validation"
)

type TeacherHandler struct {
	q *db.Queries
}

func NewTeacherHandler(q *db.Queries) *TeacherHandler {
	return &TeacherHandler{q: q}
}

type RegisterTeacherRequest struct {
	Email    string `json:"email"`
	FullName string `json:"fullName"`
	Password string `json:"password"`
}

type TeacherResponse struct {
	TeacherID uuid.UUID          `json:"teacherId"`
	Email     string             `json:"email"`
	FullName  string             `json:"fullName"`
	IsEnabled bool               `json:"isEnabled"`
	CreatedAt pgtype.Timestamptz `json:"createdAt"`
	UpdatedAt pgtype.Timestamptz `json:"updatedAt"`
}

func (h *TeacherHandler) RegisterTeacher(w http.ResponseWriter, r *http.Request) {
	var req RegisterTeacherRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid JSON body", http.StatusBadRequest)
		return
	}

	// Trim and validate fullName
	req.FullName = strings.TrimSpace(req.FullName)
	if req.FullName == "" {
		http.Error(w, "fullName is required", http.StatusBadRequest)
		return
	}

	// Validate email format
	if err := validation.ValidateEmail(req.Email); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Validate password strength
	if err := validation.ValidatePassword(req.Password); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	ctx := r.Context()

	// Check if email already exists
	_, err := h.q.GetTeacherByEmail(ctx, req.Email)
	if err == nil {
		// Email already exists
		http.Error(w, "email address is already registered", http.StatusConflict)
		return
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		// Unexpected database error
		http.Error(w, "failed to check email availability", http.StatusInternalServerError)
		return
	}
	// If err == pgx.ErrNoRows, email doesn't exist, which is what we want

	// Hash password
	hashedPassword, err := validation.HashPassword(req.Password)
	if err != nil {
		http.Error(w, "failed to process password", http.StatusInternalServerError)
		return
	}

	// Convert hashed password to pgtype.Text
	passwordHash := pgtype.Text{
		String: hashedPassword,
		Valid:  true,
	}

	// Create teacher
	teacher, err := h.q.CreateTeacher(ctx, db.CreateTeacherParams{
		Email:        req.Email,
		FullName:     req.FullName,
		PasswordHash: passwordHash,
	})
	if err != nil {
		// Check for unique constraint violation (in case of race condition)
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			// PostgreSQL error code 23505 is unique_violation
			if pgErr.Code == "23505" {
				http.Error(w, "email address is already registered", http.StatusConflict)
				return
			}
		}
		// Other database errors
		http.Error(w, "failed to create teacher account", http.StatusInternalServerError)
		return
	}

	// Convert pgtype.UUID to uuid.UUID for response
	var teacherID uuid.UUID
	if teacher.TeacherID.Valid {
		teacherID = teacher.TeacherID.Bytes
	}

	// Build response (exclude passwordHash)
	resp := TeacherResponse{
		TeacherID: teacherID,
		Email:     teacher.Email,
		FullName:  teacher.FullName,
		IsEnabled: teacher.IsEnabled,
		CreatedAt: teacher.CreatedAt,
		UpdatedAt: teacher.UpdatedAt,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(resp)
}
