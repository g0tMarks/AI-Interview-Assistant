package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/g0tMarks/AI-Interview-Assistant/backend/internal/auth"
	"github.com/g0tMarks/AI-Interview-Assistant/backend/internal/db"
	"github.com/g0tMarks/AI-Interview-Assistant/backend/internal/validation"
)

// AuthHandler handles authentication (e.g. student login via class code).
type AuthHandler struct {
	q         *db.Queries
	jwtSecret string
}

// NewAuthHandler returns an AuthHandler.
func NewAuthHandler(q *db.Queries, jwtSecret string) *AuthHandler {
	return &AuthHandler{q: q, jwtSecret: jwtSecret}
}

// StudentLoginRequest is the body for POST /auth/student/login.
type StudentLoginRequest struct {
	ClassCode string `json:"classCode"`
	Email     string `json:"email"`
}

// StudentLoginResponse is the response with the JWT.
type StudentLoginResponse struct {
	Token string `json:"token"`
}

// StudentLogin authenticates a student by class code and email, then returns a JWT.
// The student must already exist and be on the class roster (added by teacher).
func (h *AuthHandler) StudentLogin(w http.ResponseWriter, r *http.Request) {
	var req StudentLoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid JSON body", http.StatusBadRequest)
		return
	}
	req.ClassCode = strings.TrimSpace(req.ClassCode)
	req.Email = strings.TrimSpace(req.Email)
	if req.ClassCode == "" {
		http.Error(w, "classCode is required", http.StatusBadRequest)
		return
	}
	if req.Email == "" {
		http.Error(w, "email is required", http.StatusBadRequest)
		return
	}
	if err := validation.ValidateEmail(req.Email); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	ctx := r.Context()

	class, err := h.q.GetClassByCode(ctx, req.ClassCode)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			http.Error(w, "invalid class code", http.StatusUnauthorized)
			return
		}
		http.Error(w, "failed to look up class", http.StatusInternalServerError)
		return
	}

	student, err := h.q.GetStudentByEmail(ctx, req.Email)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			http.Error(w, "no student found with this email", http.StatusUnauthorized)
			return
		}
		http.Error(w, "failed to look up student", http.StatusInternalServerError)
		return
	}

	inClass, err := h.q.IsStudentInClass(ctx, db.IsStudentInClassParams{
		ClassID:   class.ClassID,
		StudentID: student.StudentID,
	})
	if err != nil {
		http.Error(w, "failed to verify roster", http.StatusInternalServerError)
		return
	}
	if !inClass {
		http.Error(w, "student is not in this class", http.StatusForbidden)
		return
	}

	var studentID uuid.UUID
	if student.StudentID.Valid {
		studentID = student.StudentID.Bytes
	} else {
		http.Error(w, "invalid student record", http.StatusInternalServerError)
		return
	}

	token, err := auth.IssueStudentToken(h.jwtSecret, studentID, auth.DefaultStudentTokenExpiry)
	if err != nil {
		http.Error(w, "failed to issue token", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(StudentLoginResponse{Token: token})
}
