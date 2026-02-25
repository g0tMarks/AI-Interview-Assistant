package handlers

import (
	"encoding/csv"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
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

// TeacherResultResponse is a flattened view of an interview result for listing/export.
type TeacherResultResponse struct {
	InterviewID        uuid.UUID           `json:"interviewId"`
	InterviewPlanID    uuid.UUID           `json:"interviewPlanId"`
	TeacherID          uuid.UUID           `json:"teacherId"`
	StudentID          *uuid.UUID          `json:"studentId,omitempty"`
	ClassID            *uuid.UUID          `json:"classId,omitempty"`
	Status             string              `json:"status"`
	StartedAt          pgtype.Timestamptz  `json:"startedAt"`
	CompletedAt        *pgtype.Timestamptz `json:"completedAt,omitempty"`
	StudentEmail       string              `json:"studentEmail,omitempty"`
	StudentDisplayName string              `json:"studentDisplayName,omitempty"`
	OverallSummary     string              `json:"overallSummary"`
	Strengths          string              `json:"strengths"`
	AreasForGrowth     string              `json:"areasForGrowth"`
	SuggestedNextSteps string              `json:"suggestedNextSteps"`
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

// ListResults lists completed interview results for a teacher and a specific interview plan.
// Optional filters:
//   - classId: only include interviews where the student is on the given class roster
//   - format=csv: return text/csv instead of JSON
//
// Route: GET /teachers/{id}/results?interviewPlanId=...&classId=...?&format=csv|json
func (h *TeacherHandler) ListResults(w http.ResponseWriter, r *http.Request) {
	teacherIDStr := chi.URLParam(r, "id")
	teacherID, err := uuid.Parse(teacherIDStr)
	if err != nil {
		http.Error(w, "invalid teacher id", http.StatusBadRequest)
		return
	}

	q := r.URL.Query()
	planIDStr := strings.TrimSpace(q.Get("interviewPlanId"))
	if planIDStr == "" {
		http.Error(w, "interviewPlanId query parameter is required", http.StatusBadRequest)
		return
	}
	planID, err := uuid.Parse(planIDStr)
	if err != nil {
		http.Error(w, "invalid interviewPlanId", http.StatusBadRequest)
		return
	}

	classIDStr := strings.TrimSpace(q.Get("classId"))
	var classID uuid.UUID
	var haveClassFilter bool
	if classIDStr != "" {
		classID, err = uuid.Parse(classIDStr)
		if err != nil {
			http.Error(w, "invalid classId", http.StatusBadRequest)
			return
		}
		haveClassFilter = true
	}

	format := strings.ToLower(strings.TrimSpace(q.Get("format")))

	ctx := r.Context()

	// Load interviews for this plan.
	planIDPg := pgtype.UUID{Bytes: planID, Valid: true}
	invs, err := h.q.ListInterviewsByPlan(ctx, planIDPg)
	if err != nil {
		http.Error(w, "failed to list interviews", http.StatusInternalServerError)
		return
	}

	results := make([]TeacherResultResponse, 0, len(invs))

	for _, inv := range invs {
		// Only include interviews owned by this teacher.
		if !inv.TeacherID.Valid || inv.TeacherID.Bytes != teacherID {
			continue
		}
		if inv.Status != "completed" {
			continue
		}

		interviewID := uuid.UUID(inv.InterviewID.Bytes)
		interviewIDPg := pgtype.UUID{Bytes: interviewID, Valid: true}

		// Optional class filter: require that the student is in the specified class.
		var classIDPtr *uuid.UUID
		if haveClassFilter {
			if !inv.StudentID.Valid {
				continue
			}
			inClass, err := h.q.IsStudentInClass(ctx, db.IsStudentInClassParams{
				ClassID:   pgtype.UUID{Bytes: classID, Valid: true},
				StudentID: inv.StudentID,
			})
			if err != nil {
				http.Error(w, "failed to verify student in class", http.StatusInternalServerError)
				return
			}
			if !inClass {
				continue
			}
			classCopy := classID
			classIDPtr = &classCopy
		}

		// Load summary; skip if none.
		summary, err := h.q.GetSummaryByInterviewID(ctx, interviewIDPg)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				continue
			}
			http.Error(w, "failed to get summary", http.StatusInternalServerError)
			return
		}

		// Optional student details.
		var studentIDPtr *uuid.UUID
		var studentEmail, studentDisplayName string
		if inv.StudentID.Valid {
			sid := uuid.UUID(inv.StudentID.Bytes)
			studentIDPtr = &sid

			student, err := h.q.GetStudentByID(ctx, inv.StudentID)
			if err == nil {
				studentEmail = student.Email
				studentDisplayName = student.DisplayName
			}
		}

		var completedAtPtr *pgtype.Timestamptz
		if inv.CompletedAt.Valid {
			tmp := inv.CompletedAt
			completedAtPtr = &tmp
		}

		overall := ""
		if summary.OverallSummary.Valid {
			overall = summary.OverallSummary.String
		}
		strengths := ""
		if summary.Strengths.Valid {
			strengths = summary.Strengths.String
		}
		areas := ""
		if summary.AreasForGrowth.Valid {
			areas = summary.AreasForGrowth.String
		}
		nextSteps := ""
		if summary.SuggestedNextSteps.Valid {
			nextSteps = summary.SuggestedNextSteps.String
		}

		results = append(results, TeacherResultResponse{
			InterviewID:        interviewID,
			InterviewPlanID:    planID,
			TeacherID:          teacherID,
			StudentID:          studentIDPtr,
			ClassID:            classIDPtr,
			Status:             inv.Status,
			StartedAt:          inv.StartedAt,
			CompletedAt:        completedAtPtr,
			StudentEmail:       studentEmail,
			StudentDisplayName: studentDisplayName,
			OverallSummary:     overall,
			Strengths:          strengths,
			AreasForGrowth:     areas,
			SuggestedNextSteps: nextSteps,
		})
	}

	if format == "csv" {
		w.Header().Set("Content-Type", "text/csv; charset=utf-8")
		w.Header().Set("Content-Disposition", "attachment; filename=\"teacher-results.csv\"")

		writer := csv.NewWriter(w)
		defer writer.Flush()

		_ = writer.Write([]string{
			"interview_id",
			"interview_plan_id",
			"teacher_id",
			"class_id",
			"student_id",
			"student_email",
			"student_display_name",
			"status",
			"started_at",
			"completed_at",
			"overall_summary",
			"strengths",
			"areas_for_growth",
			"suggested_next_steps",
		})

		for _, rres := range results {
			var classIDStrOut, studentIDStr, completedAtStr string
			if rres.ClassID != nil {
				classIDStrOut = rres.ClassID.String()
			}
			if rres.StudentID != nil {
				studentIDStr = rres.StudentID.String()
			}
			if rres.CompletedAt != nil && rres.CompletedAt.Valid {
				completedAtStr = rres.CompletedAt.Time.Format(time.RFC3339)
			}
			startedAtStr := ""
			if rres.StartedAt.Valid {
				startedAtStr = rres.StartedAt.Time.Format(time.RFC3339)
			}

			_ = writer.Write([]string{
				rres.InterviewID.String(),
				rres.InterviewPlanID.String(),
				rres.TeacherID.String(),
				classIDStrOut,
				studentIDStr,
				rres.StudentEmail,
				rres.StudentDisplayName,
				rres.Status,
				startedAtStr,
				completedAtStr,
				rres.OverallSummary,
				rres.Strengths,
				rres.AreasForGrowth,
				rres.SuggestedNextSteps,
			})
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(results)
}
