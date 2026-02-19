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
	InterviewPlanID uuid.UUID  `json:"interviewPlanId"`
	TeacherID       uuid.UUID  `json:"teacherId"`
	ClassID         *uuid.UUID `json:"classId,omitempty"`
	StudentID       *uuid.UUID `json:"studentId,omitempty"`
	Simulated       *bool      `json:"simulated"`
	StudentName     string     `json:"studentName"`
	Status          string     `json:"status"`
}

type InterviewResponse struct {
	InterviewID     uuid.UUID          `json:"interviewId"`
	InterviewPlanID uuid.UUID          `json:"interviewPlanId"`
	TeacherID       *uuid.UUID         `json:"teacherId"`
	StudentID       *uuid.UUID         `json:"studentId"`
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

	// Determine student_id: if classId is provided, verify student is in roster
	studentIDPgtype := pgtype.UUID{}
	if req.ClassID != nil && req.StudentID != nil {
		// Verify student is in the class roster
		classIDPgtype := pgtype.UUID{
			Bytes: *req.ClassID,
			Valid: true,
		}
		studentIDPgtypeCheck := pgtype.UUID{
			Bytes: *req.StudentID,
			Valid: true,
		}
		inClass, err := h.q.IsStudentInClass(ctx, db.IsStudentInClassParams{
			ClassID:   classIDPgtype,
			StudentID: studentIDPgtypeCheck,
		})
		if err != nil {
			http.Error(w, "failed to verify student in class: "+err.Error(), http.StatusInternalServerError)
			return
		}
		if !inClass {
			http.Error(w, "student is not in the specified class", http.StatusBadRequest)
			return
		}
		// Student is in class, use the provided student_id
		studentIDPgtype = studentIDPgtypeCheck
	} else if req.StudentID != nil {
		// Direct student_id provided without classId
		studentIDPgtype = pgtype.UUID{
			Bytes: *req.StudentID,
			Valid: true,
		}
	}
	// If neither classId+studentId nor studentId is provided, studentIDPgtype remains invalid (null)

	interview, err := h.q.CreateInterview(ctx, db.CreateInterviewParams{
		InterviewPlanID: interviewPlanIDPgtype,
		TeacherID:       teacherIDPgtype,
		StudentID:       studentIDPgtype,
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

	var studentIDResp *uuid.UUID
	if interview.StudentID.Valid {
		uid := uuid.UUID(interview.StudentID.Bytes)
		studentIDResp = &uid
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
		StudentID:       studentIDResp,
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

	var studentIDResp *uuid.UUID
	if interview.StudentID.Valid {
		uid := uuid.UUID(interview.StudentID.Bytes)
		studentIDResp = &uid
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
		StudentID:       studentIDResp,
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

