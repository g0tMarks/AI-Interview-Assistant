package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/g0tMarks/AI-Interview-Assistant/backend/internal/db"
	"github.com/g0tMarks/AI-Interview-Assistant/backend/internal/services"
)

// StudentProfileHandler handles generation and retrieval of student writing profiles.
type StudentProfileHandler struct {
	q   *db.Queries
	llm services.LLMService
}

// NewStudentProfileHandler creates a StudentProfileHandler.
func NewStudentProfileHandler(q *db.Queries, llm services.LLMService) *StudentProfileHandler {
	return &StudentProfileHandler{q: q, llm: llm}
}

// RunStudentProfile aggregates a student's submissions and artifacts and generates a profile.
// POST /students/{id}/profile/run
func (h *StudentProfileHandler) RunStudentProfile(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	studentID, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "invalid student id", http.StatusBadRequest)
		return
	}
	ctx := r.Context()

	// Ensure student exists.
	stu, err := h.q.GetStudentByID(ctx, pgtype.UUID{Bytes: studentID, Valid: true})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			http.Error(w, "student not found", http.StatusNotFound)
			return
		}
		http.Error(w, "failed to get student", http.StatusInternalServerError)
		return
	}

	// Collect all submissions for the student.
	subs, err := h.q.ListSubmissionsByStudent(ctx, pgtype.UUID{Bytes: studentID, Valid: true})
	if err != nil {
		http.Error(w, "failed to list submissions", http.StatusInternalServerError)
		return
	}
	if len(subs) == 0 {
		http.Error(w, "no submissions for student", http.StatusBadRequest)
		return
	}

	var samples []services.StudentWritingSample
	var submissionIDs []string
	var artifactIDs []string

	for _, sub := range subs {
		subUUID := uuid.UUID(sub.SubmissionID.Bytes)
		subIDStr := subUUID.String()
		submissionIDs = append(submissionIDs, subIDStr)

		arts, err := h.q.ListArtifactsBySubmission(ctx, sub.SubmissionID)
		if err != nil {
			http.Error(w, "failed to list artifacts", http.StatusInternalServerError)
			return
		}

		for _, a := range arts {
			artUUID := uuid.UUID(a.ArtifactID.Bytes)
			aid := artUUID.String()
			artifactIDs = append(artifactIDs, aid)

			var pl struct {
				Text string `json:"text"`
			}
			if err := json.Unmarshal(a.Payload, &pl); err != nil || pl.Text == "" {
				continue
			}

			// Optional: use rubric title as context if available.
			rubric, _ := h.q.GetRubricByID(ctx, sub.RubricID)
			context := rubric.Title

			samples = append(samples, services.StudentWritingSample{
				SubmissionID: subIDStr,
				ArtifactID:   aid,
				Text:         pl.Text,
				Context:      context,
			})
		}
	}

	if len(samples) == 0 {
		http.Error(w, "no text content in artifacts for this student", http.StatusBadRequest)
		return
	}

	opts := services.GenerateStudentProfileOpts{
		StudentDisplayName: stu.DisplayName,
		Samples:            samples,
	}

	profilePayload, err := h.llm.GenerateStudentProfile(ctx, opts)
	if err != nil {
		http.Error(w, "failed to generate student profile: "+err.Error(), http.StatusInternalServerError)
		return
	}
	profilePayload.Provenance.SubmissionIDs = submissionIDs
	profilePayload.Provenance.ArtifactIDs = artifactIDs
	profilePayload.Provenance.SampleCount = len(samples)

	blob, err := profilePayload.ToJSONB()
	if err != nil {
		http.Error(w, "failed to serialize profile", http.StatusInternalServerError)
		return
	}

	row, err := h.q.CreateStudentProfile(ctx, db.CreateStudentProfileParams{
		StudentID: pgtype.UUID{Bytes: studentID, Valid: true},
		Profile:   blob,
	})
	if err != nil {
		http.Error(w, "failed to save profile", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"studentProfileId": row.StudentProfileID.Bytes,
		"studentId":        row.StudentID.Bytes,
		"profile":          profilePayload,
		"createdAt":        row.CreatedAt.Time.Format(time.RFC3339),
	})
}

// GetStudentProfile returns the latest profile for a student.
// GET /students/{id}/profile
func (h *StudentProfileHandler) GetStudentProfile(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	studentID, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "invalid student id", http.StatusBadRequest)
		return
	}
	ctx := r.Context()

	rec, err := h.q.GetLatestStudentProfileByStudent(ctx, pgtype.UUID{Bytes: studentID, Valid: true})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			http.Error(w, "no profile found for student", http.StatusNotFound)
			return
		}
		http.Error(w, "failed to get profile", http.StatusInternalServerError)
		return
	}

	var profile services.StudentProfilePayload
	if err := json.Unmarshal(rec.Profile, &profile); err != nil {
		http.Error(w, "failed to decode profile", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"studentProfileId": rec.StudentProfileID.Bytes,
		"studentId":        rec.StudentID.Bytes,
		"profile":          profile,
		"createdAt":        rec.CreatedAt.Time.Format(time.RFC3339),
	})
}

