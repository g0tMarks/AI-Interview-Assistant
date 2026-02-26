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

// BulkCreateInterviewsRequest is the body for POST /classes/{id}/interviews/bulk.
// If studentIds is empty or omitted, the handler will create one interview per
// student currently on the class roster.
type BulkCreateInterviewsRequest struct {
	InterviewPlanID uuid.UUID   `json:"interviewPlanId"`
	TeacherID       uuid.UUID   `json:"teacherId"`
	StudentIDs      []uuid.UUID `json:"studentIds,omitempty"`
	Simulated       *bool       `json:"simulated,omitempty"`
}

// BulkCreateInterviewsResponse summarizes the result of the bulk operation.
type BulkCreateInterviewsResponse struct {
	ClassID      uuid.UUID `json:"classId"`
	CreatedCount int       `json:"createdCount"`
	SkippedCount int       `json:"skippedCount"`
	ErrorCount   int       `json:"errorCount"`
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

// BulkCreateInterviews handles POST /classes/{id}/interviews/bulk.
// It uses the specified interview plan and teacher, plus either the provided
// list of studentIds or the full class roster, to create one interview per student.
func (h *ClassHandler) BulkCreateInterviews(w http.ResponseWriter, r *http.Request) {
	classIDStr := chi.URLParam(r, "id")
	classID, err := uuid.Parse(classIDStr)
	if err != nil {
		http.Error(w, "invalid class id", http.StatusBadRequest)
		return
	}

	var req BulkCreateInterviewsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid JSON body", http.StatusBadRequest)
		return
	}

	if req.InterviewPlanID == uuid.Nil {
		http.Error(w, "interviewPlanId is required", http.StatusBadRequest)
		return
	}
	if req.TeacherID == uuid.Nil {
		http.Error(w, "teacherId is required", http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	classIDPg := pgtype.UUID{Bytes: classID, Valid: true}

	// Verify class exists
	if _, err := h.q.GetClassByID(ctx, classIDPg); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			http.Error(w, "class not found", http.StatusNotFound)
			return
		}
		http.Error(w, "failed to get class", http.StatusInternalServerError)
		return
	}

	// Verify interview plan exists
	planIDPg := pgtype.UUID{Bytes: req.InterviewPlanID, Valid: true}
	if _, err := h.q.GetInterviewPlanByID(ctx, planIDPg); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			http.Error(w, "interview plan not found", http.StatusNotFound)
			return
		}
		http.Error(w, "failed to get interview plan", http.StatusInternalServerError)
		return
	}

	// Determine which students to target.
	var studentIDs []uuid.UUID
	if len(req.StudentIDs) > 0 {
		studentIDs = req.StudentIDs
	} else {
		// Use full roster for the class.
		rows, err := h.q.ListRosterByClass(ctx, classIDPg)
		if err != nil {
			http.Error(w, "failed to list roster", http.StatusInternalServerError)
			return
		}
		studentIDs = make([]uuid.UUID, 0, len(rows))
		for _, row := range rows {
			if row.StudentID.Valid {
				studentIDs = append(studentIDs, row.StudentID.Bytes)
			}
		}
	}

	if len(studentIDs) == 0 {
		http.Error(w, "no students to create interviews for", http.StatusBadRequest)
		return
	}

	simulated := false
	if req.Simulated != nil {
		simulated = *req.Simulated
	}

	teacherIDPg := pgtype.UUID{Bytes: req.TeacherID, Valid: true}

	respSummary := BulkCreateInterviewsResponse{
		ClassID: classID,
	}

	for _, sid := range studentIDs {
		if sid == uuid.Nil {
			respSummary.ErrorCount++
			continue
		}

		studentIDPg := pgtype.UUID{Bytes: sid, Valid: true}

		// Optionally, we could skip students not actually in the roster when
		// studentIds are supplied explicitly. For now, rely on DB constraints.
		_, err := h.q.CreateInterview(ctx, db.CreateInterviewParams{
			InterviewPlanID: planIDPg,
			TeacherID:       teacherIDPg,
			StudentID:       studentIDPg,
			Simulated:       simulated,
			StudentName:     pgtype.Text{},
			Status:          "in_progress",
			SubmissionID:    pgtype.UUID{}, // null for bulk class interviews
		})
		if err != nil {
			respSummary.ErrorCount++
			continue
		}
		respSummary.CreatedCount++
	}

	w.Header().Set("Content-Type", "application/json")
	if respSummary.CreatedCount > 0 {
		w.WriteHeader(http.StatusCreated)
	} else {
		// No interview was created; treat as a bad request.
		w.WriteHeader(http.StatusBadRequest)
	}
	_ = json.NewEncoder(w).Encode(respSummary)
}
