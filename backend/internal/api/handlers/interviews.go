package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/g0tMarks/AI-Interview-Assistant/backend/internal/db"
)

type InterviewHandler struct {
	q *db.Queries
}

func NewInterviewHandler(q *db.Queries) *InterviewHandler {
	return &InterviewHandler{q: q}
}

type CreateInterviewRequest struct {
	InterviewPlanID uuid.UUID `json:"interviewPlanId"`
	TeacherID       uuid.UUID `json:"teacherId"`
	Simulated       *bool     `json:"simulated"`
	StudentName     string    `json:"studentName"`
	Status          string    `json:"status"`
}

type InterviewResponse struct {
	InterviewID     uuid.UUID          `json:"interviewId"`
	InterviewPlanID uuid.UUID          `json:"interviewPlanId"`
	TeacherID       *uuid.UUID         `json:"teacherId"`
	Simulated       bool               `json:"simulated"`
	StudentName     *string            `json:"studentName"`
	Status          string             `json:"status"`
	StartedAt       pgtype.Timestamptz `json:"startedAt"`
	CompletedAt     *pgtype.Timestamptz `json:"completedAt"`
}

func (h *InterviewHandler) CreateInterview(w http.ResponseWriter, r *http.Request) {
	var req CreateInterviewRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid JSON body", http.StatusBadRequest)
		return
	}

	// Validate required fields
	if req.InterviewPlanID == uuid.Nil {
		http.Error(w, "interviewPlanId is required", http.StatusBadRequest)
		return
	}

	// Validate status enum
	status := req.Status
	if status == "" {
		status = "in_progress" // default
	} else if status != "draft" && status != "in_progress" && status != "completed" {
		http.Error(w, "invalid status: must be one of draft, in_progress, completed", http.StatusBadRequest)
		return
	}

	// Default simulated
	simulated := true
	if req.Simulated != nil {
		simulated = *req.Simulated
	}

	ctx := r.Context()

	// Verify interview plan exists
	planIDPgtype := pgtype.UUID{
		Bytes: req.InterviewPlanID,
		Valid: true,
	}
	_, err := h.q.GetInterviewPlanByID(ctx, planIDPgtype)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			http.Error(w, "interview plan not found", http.StatusNotFound)
			return
		}
		http.Error(w, "failed to verify interview plan", http.StatusInternalServerError)
		return
	}

	// Convert to database params
	interviewPlanIDPgtype := pgtype.UUID{
		Bytes: req.InterviewPlanID,
		Valid: true,
	}

	teacherIDPgtype := pgtype.UUID{}
	if req.TeacherID != uuid.Nil {
		teacherIDPgtype.Bytes = req.TeacherID
		teacherIDPgtype.Valid = true
	}

	studentNamePgtype := pgtype.Text{}
	if req.StudentName != "" {
		studentNamePgtype.String = strings.TrimSpace(req.StudentName)
		studentNamePgtype.Valid = true
	}

	interview, err := h.q.CreateInterview(ctx, db.CreateInterviewParams{
		InterviewPlanID: interviewPlanIDPgtype,
		TeacherID:       teacherIDPgtype,
		Simulated:       simulated,
		StudentName:     studentNamePgtype,
		Status:           status,
	})
	if err != nil {
		http.Error(w, "failed to create interview: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Convert to response
	var interviewID uuid.UUID
	if interview.InterviewID.Valid {
		interviewID = interview.InterviewID.Bytes
	}

	var interviewPlanIDResp uuid.UUID
	if interview.InterviewPlanID.Valid {
		interviewPlanIDResp = interview.InterviewPlanID.Bytes
	}

	var teacherIDResp *uuid.UUID
	if interview.TeacherID.Valid {
		uid := uuid.UUID(interview.TeacherID.Bytes)
		teacherIDResp = &uid
	}

	var studentNameResp *string
	if interview.StudentName.Valid {
		studentNameResp = &interview.StudentName.String
	}

	var completedAtResp *pgtype.Timestamptz
	if interview.CompletedAt.Valid {
		completedAtResp = &interview.CompletedAt
	}

	resp := InterviewResponse{
		InterviewID:     interviewID,
		InterviewPlanID: interviewPlanIDResp,
		TeacherID:       teacherIDResp,
		Simulated:       interview.Simulated,
		StudentName:     studentNameResp,
		Status:          interview.Status,
		StartedAt:       interview.StartedAt,
		CompletedAt:     completedAtResp,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(resp)
}

func (h *InterviewHandler) GetInterview(w http.ResponseWriter, r *http.Request) {
	// Extract id from path parameter
	idStr := chi.URLParam(r, "id")
	if idStr == "" {
		http.Error(w, "interview ID is required", http.StatusBadRequest)
		return
	}

	// Validate UUID format
	interviewID, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "invalid interview ID format", http.StatusBadRequest)
		return
	}

	ctx := r.Context()

	// Convert to pgtype.UUID
	interviewIDPgtype := pgtype.UUID{
		Bytes: interviewID,
		Valid: true,
	}

	// Query database
	interview, err := h.q.GetInterviewByID(ctx, interviewIDPgtype)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			http.Error(w, "interview not found", http.StatusNotFound)
			return
		}
		http.Error(w, "failed to retrieve interview", http.StatusInternalServerError)
		return
	}

	// Convert to response format
	var interviewIDResp uuid.UUID
	if interview.InterviewID.Valid {
		interviewIDResp = interview.InterviewID.Bytes
	}

	var interviewPlanIDResp uuid.UUID
	if interview.InterviewPlanID.Valid {
		interviewPlanIDResp = interview.InterviewPlanID.Bytes
	}

	var teacherIDResp *uuid.UUID
	if interview.TeacherID.Valid {
		uid := uuid.UUID(interview.TeacherID.Bytes)
		teacherIDResp = &uid
	}

	var studentNameResp *string
	if interview.StudentName.Valid {
		studentNameResp = &interview.StudentName.String
	}

	var completedAtResp *pgtype.Timestamptz
	if interview.CompletedAt.Valid {
		completedAtResp = &interview.CompletedAt
	}

	resp := InterviewResponse{
		InterviewID:     interviewIDResp,
		InterviewPlanID: interviewPlanIDResp,
		TeacherID:       teacherIDResp,
		Simulated:       interview.Simulated,
		StudentName:     studentNameResp,
		Status:          interview.Status,
		StartedAt:       interview.StartedAt,
		CompletedAt:     completedAtResp,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(resp)
}

