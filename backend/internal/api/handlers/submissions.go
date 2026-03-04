package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/g0tMarks/AI-Interview-Assistant/backend/internal/db"
	"github.com/g0tMarks/AI-Interview-Assistant/backend/internal/safety"
	"github.com/g0tMarks/AI-Interview-Assistant/backend/internal/services"
)

// SubmissionHandler handles submission, artifacts, viva, and authorship report endpoints.
type SubmissionHandler struct {
	q                *db.Queries
	llm              services.LLMService
	interviewHandler *InterviewHandler
}

// NewSubmissionHandler creates a SubmissionHandler.
func NewSubmissionHandler(q *db.Queries, llm services.LLMService, interviewHandler *InterviewHandler) *SubmissionHandler {
	return &SubmissionHandler{q: q, llm: llm, interviewHandler: interviewHandler}
}

// CreateSubmissionRequest is the body for POST /submissions.
type CreateSubmissionRequest struct {
	StudentID uuid.UUID `json:"studentId"`
	RubricID  uuid.UUID `json:"rubricId"`
	Title     string    `json:"title"`
	Notes     string    `json:"notes"`
	Status    string    `json:"status"`
}

// SubmissionResponse is the response for a single submission.
type SubmissionResponse struct {
	SubmissionID uuid.UUID `json:"submissionId"`
	StudentID    uuid.UUID `json:"studentId"`
	RubricID     uuid.UUID `json:"rubricId"`
	Status       string    `json:"status"`
	Title        *string   `json:"title,omitempty"`
	Notes        *string   `json:"notes,omitempty"`
	CreatedAt    string    `json:"createdAt"`
	UpdatedAt    string    `json:"updatedAt"`
}

func (h *SubmissionHandler) CreateSubmission(w http.ResponseWriter, r *http.Request) {
	var req CreateSubmissionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid JSON body", http.StatusBadRequest)
		return
	}
	if req.StudentID == uuid.Nil || req.RubricID == uuid.Nil {
		http.Error(w, "studentId and rubricId are required", http.StatusBadRequest)
		return
	}

	status := "draft"
	if req.Status != "" {
		status = req.Status
		if status != "draft" && status != "submitted" && status != "viva_in_progress" && status != "viva_completed" && status != "report_ready" {
			http.Error(w, "invalid status", http.StatusBadRequest)
			return
		}
	}

	ctx := r.Context()
	_, err := h.q.GetRubricByID(ctx, pgtype.UUID{Bytes: req.RubricID, Valid: true})
	if err != nil {
		if err == pgx.ErrNoRows {
			http.Error(w, "rubric not found", http.StatusNotFound)
			return
		}
		http.Error(w, "failed to verify rubric", http.StatusInternalServerError)
		return
	}
	_, err = h.q.GetStudentByID(ctx, pgtype.UUID{Bytes: req.StudentID, Valid: true})
	if err != nil {
		if err == pgx.ErrNoRows {
			http.Error(w, "student not found", http.StatusNotFound)
			return
		}
		http.Error(w, "failed to verify student", http.StatusInternalServerError)
		return
	}

	sub, err := h.q.CreateSubmission(ctx, db.CreateSubmissionParams{
		StudentID: pgtype.UUID{Bytes: req.StudentID, Valid: true},
		RubricID:  pgtype.UUID{Bytes: req.RubricID, Valid: true},
		Status:    status,
		Title:     pgtype.Text{String: req.Title, Valid: req.Title != ""},
		Notes:     pgtype.Text{String: req.Notes, Valid: req.Notes != ""},
	})
	if err != nil {
		http.Error(w, "failed to create submission: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(submissionToResponse(sub))
}

func (h *SubmissionHandler) GetSubmission(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "invalid submission id", http.StatusBadRequest)
		return
	}
	ctx := r.Context()
	sub, err := h.q.GetSubmissionByID(ctx, pgtype.UUID{Bytes: id, Valid: true})
	if err != nil {
		if err == pgx.ErrNoRows {
			http.Error(w, "submission not found", http.StatusNotFound)
			return
		}
		http.Error(w, "failed to get submission", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(submissionToResponse(sub))
}

func submissionToResponse(sub db.AppSubmission) SubmissionResponse {
	var title, notes *string
	if sub.Title.Valid {
		title = &sub.Title.String
	}
	if sub.Notes.Valid {
		notes = &sub.Notes.String
	}
	return SubmissionResponse{
		SubmissionID: sub.SubmissionID.Bytes,
		StudentID:    sub.StudentID.Bytes,
		RubricID:     sub.RubricID.Bytes,
		Status:       sub.Status,
		Title:        title,
		Notes:        notes,
		CreatedAt:    sub.CreatedAt.Time.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:    sub.UpdatedAt.Time.Format("2006-01-02T15:04:05Z07:00"),
	}
}

// CreateSubmissionArtifactRequest is the body for POST /submissions/{id}/artifacts.
type CreateSubmissionArtifactRequest struct {
	ArtifactType string          `json:"artifactType"`
	Payload      json.RawMessage `json:"payload"`
	OrderIndex   *int            `json:"orderIndex"`
}

func (h *SubmissionHandler) CreateArtifact(w http.ResponseWriter, r *http.Request) {
	subIDStr := chi.URLParam(r, "id")
	subID, err := uuid.Parse(subIDStr)
	if err != nil {
		http.Error(w, "invalid submission id", http.StatusBadRequest)
		return
	}
	var req CreateSubmissionArtifactRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid JSON body", http.StatusBadRequest)
		return
	}
	validTypes := map[string]bool{"main_text": true, "draft_checkpoint": true, "revision_note": true, "citation_source": true, "file_ref": true}
	if !validTypes[req.ArtifactType] {
		http.Error(w, "invalid artifactType", http.StatusBadRequest)
		return
	}
	if req.Payload == nil {
		req.Payload = []byte("{}")
	}

	ctx := r.Context()
	_, err = h.q.GetSubmissionByID(ctx, pgtype.UUID{Bytes: subID, Valid: true})
	if err != nil {
		if err == pgx.ErrNoRows {
			http.Error(w, "submission not found", http.StatusNotFound)
			return
		}
		http.Error(w, "failed to get submission", http.StatusInternalServerError)
		return
	}

	orderIndex := 0
	if req.OrderIndex != nil {
		orderIndex = *req.OrderIndex
	}

	art, err := h.q.CreateSubmissionArtifact(ctx, db.CreateSubmissionArtifactParams{
		SubmissionID: pgtype.UUID{Bytes: subID, Valid: true},
		ArtifactType: req.ArtifactType,
		Payload:      req.Payload,
		OrderIndex:   int32(orderIndex),
	})
	if err != nil {
		http.Error(w, "failed to create artifact: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(artifactToResponse(art))
}

func (h *SubmissionHandler) ListArtifacts(w http.ResponseWriter, r *http.Request) {
	subIDStr := chi.URLParam(r, "id")
	subID, err := uuid.Parse(subIDStr)
	if err != nil {
		http.Error(w, "invalid submission id", http.StatusBadRequest)
		return
	}
	ctx := r.Context()
	list, err := h.q.ListArtifactsBySubmission(ctx, pgtype.UUID{Bytes: subID, Valid: true})
	if err != nil {
		http.Error(w, "failed to list artifacts", http.StatusInternalServerError)
		return
	}
	out := make([]ArtifactResponse, len(list))
	for i := range list {
		out[i] = artifactToResponse(list[i])
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(out)
}

type ArtifactResponse struct {
	ArtifactID   uuid.UUID       `json:"artifactId"`
	SubmissionID uuid.UUID       `json:"submissionId"`
	ArtifactType string          `json:"artifactType"`
	Payload      json.RawMessage `json:"payload"`
	OrderIndex   int             `json:"orderIndex"`
	CreatedAt    string          `json:"createdAt"`
}

func artifactToResponse(a db.AppSubmissionArtifact) ArtifactResponse {
	return ArtifactResponse{
		ArtifactID:   a.ArtifactID.Bytes,
		SubmissionID: a.SubmissionID.Bytes,
		ArtifactType: a.ArtifactType,
		Payload:      a.Payload,
		OrderIndex:   int(a.OrderIndex),
		CreatedAt:    a.CreatedAt.Time.Format("2006-01-02T15:04:05Z07:00"),
	}
}

// StartViva creates or reuses an interview for this submission (viva). Uses first available plan for the submission's rubric.
func (h *SubmissionHandler) StartViva(w http.ResponseWriter, r *http.Request) {
	subIDStr := chi.URLParam(r, "id")
	subID, err := uuid.Parse(subIDStr)
	if err != nil {
		http.Error(w, "invalid submission id", http.StatusBadRequest)
		return
	}
	ctx := r.Context()
	sub, err := h.q.GetSubmissionByID(ctx, pgtype.UUID{Bytes: subID, Valid: true})
	if err != nil {
		if err == pgx.ErrNoRows {
			http.Error(w, "submission not found", http.StatusNotFound)
			return
		}
		http.Error(w, "failed to get submission", http.StatusInternalServerError)
		return
	}

	// Already have a viva?
	existing, err := h.q.GetInterviewBySubmissionID(ctx, pgtype.UUID{Bytes: subID, Valid: true})
	if err == nil {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(interviewToMinimalResponse(existing))
		return
	}
	if err != pgx.ErrNoRows {
		http.Error(w, "failed to get viva", http.StatusInternalServerError)
		return
	}

	// Get or create an interview plan for this rubric
	plans, err := h.q.ListPlansByRubric(ctx, sub.RubricID)
	if err != nil || len(plans) == 0 {
		http.Error(w, "no interview plan found for this rubric; create one first via POST /interview-templates", http.StatusBadRequest)
		return
	}
	plan := plans[0]
	planID := plan.InterviewPlanID.Bytes

	// Create interview linked to submission
	rubric, _ := h.q.GetRubricByID(ctx, sub.RubricID)
	teacherID := rubric.TeacherID

	inv, err := h.q.CreateInterview(ctx, db.CreateInterviewParams{
		InterviewPlanID: pgtype.UUID{Bytes: planID, Valid: true},
		TeacherID:      teacherID,
		StudentID:      sub.StudentID,
		Simulated:      false,
		StudentName:    pgtype.Text{},
		Status:         "in_progress",
		SubmissionID:   pgtype.UUID{Bytes: subID, Valid: true},
	})
	if err != nil {
		http.Error(w, "failed to create viva: "+err.Error(), http.StatusInternalServerError)
		return
	}

	_ = h.q.UpdateSubmissionStatus(ctx, db.UpdateSubmissionStatusParams{
		SubmissionID: pgtype.UUID{Bytes: subID, Valid: true},
		Status:       "viva_in_progress",
	})

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(interviewToMinimalResponse(inv))
}

func interviewToMinimalResponse(inv db.AppInterview) map[string]interface{} {
	out := map[string]interface{}{
		"interviewId":     inv.InterviewID.Bytes,
		"interviewPlanId": inv.InterviewPlanID.Bytes,
		"status":          inv.Status,
		"startedAt":       inv.StartedAt.Time.Format("2006-01-02T15:04:05Z07:00"),
	}
	if inv.CompletedAt.Valid {
		out["completedAt"] = inv.CompletedAt.Time.Format("2006-01-02T15:04:05Z07:00")
	}
	if inv.SubmissionID.Valid {
		out["submissionId"] = inv.SubmissionID.Bytes
	}
	return out
}

func (h *SubmissionHandler) GetViva(w http.ResponseWriter, r *http.Request) {
	subIDStr := chi.URLParam(r, "id")
	subID, err := uuid.Parse(subIDStr)
	if err != nil {
		http.Error(w, "invalid submission id", http.StatusBadRequest)
		return
	}
	ctx := r.Context()
	inv, err := h.q.GetInterviewBySubmissionID(ctx, pgtype.UUID{Bytes: subID, Valid: true})
	if err != nil {
		if err == pgx.ErrNoRows {
			http.Error(w, "viva not found for this submission", http.StatusNotFound)
			return
		}
		http.Error(w, "failed to get viva", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(interviewToMinimalResponse(inv))
}

// VivaMessages appends a message to the submission's viva (creates message on the viva interview).
func (h *SubmissionHandler) VivaMessages(w http.ResponseWriter, r *http.Request) {
	subIDStr := chi.URLParam(r, "id")
	subID, err := uuid.Parse(subIDStr)
	if err != nil {
		http.Error(w, "invalid submission id", http.StatusBadRequest)
		return
	}
	ctx := r.Context()
	inv, err := h.q.GetInterviewBySubmissionID(ctx, pgtype.UUID{Bytes: subID, Valid: true})
	if err != nil {
		if err == pgx.ErrNoRows {
			http.Error(w, "viva not found; start viva first", http.StatusNotFound)
			return
		}
		http.Error(w, "failed to get viva", http.StatusInternalServerError)
		return
	}

	var req CreateInterviewMessageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid JSON body", http.StatusBadRequest)
		return
	}
	if req.Sender != "ai" && req.Sender != "user" {
		http.Error(w, "sender must be either 'ai' or 'user'", http.StatusBadRequest)
		return
	}
	if req.Content == "" {
		http.Error(w, "content is required", http.StatusBadRequest)
		return
	}
	sanitized, err := safety.SanitizeUserMessage(req.Content)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	questionIDPg := pgtype.UUID{}
	if req.InterviewQuestionID != nil {
		questionIDPg.Bytes = *req.InterviewQuestionID
		questionIDPg.Valid = true
	}
	msg, err := h.q.CreateInterviewMessage(ctx, db.CreateInterviewMessageParams{
		InterviewID:         inv.InterviewID,
		Sender:              req.Sender,
		InterviewQuestionID: questionIDPg,
		Content:             sanitized,
	})
	if err != nil {
		http.Error(w, "failed to create message: "+err.Error(), http.StatusInternalServerError)
		return
	}

	var questionIDResp *uuid.UUID
	if msg.InterviewQuestionID.Valid {
		id := uuid.UUID(msg.InterviewQuestionID.Bytes)
		questionIDResp = &id
	}
	resp := InterviewMessageResponse{
		InterviewMessageID:  msg.InterviewMessageID.Bytes,
		InterviewID:         msg.InterviewID.Bytes,
		Sender:              msg.Sender,
		InterviewQuestionID: questionIDResp,
		Content:             msg.Content,
		CreatedAt:           msg.CreatedAt,
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(resp)
}

func (h *SubmissionHandler) ListVivaMessages(w http.ResponseWriter, r *http.Request) {
	subIDStr := chi.URLParam(r, "id")
	subID, err := uuid.Parse(subIDStr)
	if err != nil {
		http.Error(w, "invalid submission id", http.StatusBadRequest)
		return
	}
	ctx := r.Context()
	inv, err := h.q.GetInterviewBySubmissionID(ctx, pgtype.UUID{Bytes: subID, Valid: true})
	if err != nil {
		if err == pgx.ErrNoRows {
			http.Error(w, "viva not found", http.StatusNotFound)
			return
		}
		http.Error(w, "failed to get viva", http.StatusInternalServerError)
		return
	}

	msgs, err := h.q.ListMessagesByInterview(ctx, inv.InterviewID)
	if err != nil {
		http.Error(w, "failed to list messages: "+err.Error(), http.StatusInternalServerError)
		return
	}
	resp := make([]InterviewMessageResponse, len(msgs))
	for i, msg := range msgs {
		var questionIDResp *uuid.UUID
		if msg.InterviewQuestionID.Valid {
			id := uuid.UUID(msg.InterviewQuestionID.Bytes)
			questionIDResp = &id
		}
		resp[i] = InterviewMessageResponse{
			InterviewMessageID:  msg.InterviewMessageID.Bytes,
			InterviewID:         msg.InterviewID.Bytes,
			Sender:              msg.Sender,
			InterviewQuestionID: questionIDResp,
			Content:             msg.Content,
			CreatedAt:           msg.CreatedAt,
		}
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(resp)
}

// RunAuthorship computes and stores an authorship report for the submission.
func (h *SubmissionHandler) RunAuthorship(w http.ResponseWriter, r *http.Request) {
	subIDStr := chi.URLParam(r, "id")
	subID, err := uuid.Parse(subIDStr)
	if err != nil {
		http.Error(w, "invalid submission id", http.StatusBadRequest)
		return
	}
	ctx := r.Context()
	sub, err := h.q.GetSubmissionByID(ctx, pgtype.UUID{Bytes: subID, Valid: true})
	if err != nil {
		if err == pgx.ErrNoRows {
			http.Error(w, "submission not found", http.StatusNotFound)
			return
		}
		http.Error(w, "failed to get submission", http.StatusInternalServerError)
		return
	}

	artifacts, err := h.q.ListArtifactsBySubmission(ctx, pgtype.UUID{Bytes: subID, Valid: true})
	if err != nil {
		http.Error(w, "failed to list artifacts", http.StatusInternalServerError)
		return
	}
	var submissionSummary string
	var artifactIDs []string
	for _, a := range artifacts {
		artifactIDs = append(artifactIDs, uuid.UUID(a.ArtifactID.Bytes).String())
		if len(a.Payload) > 0 {
			var pl struct {
				Text string `json:"text"`
			}
			_ = json.Unmarshal(a.Payload, &pl)
			if pl.Text != "" {
				submissionSummary += pl.Text + "\n\n"
			}
		}
	}
	if submissionSummary == "" {
		submissionSummary = "(no artifact text)"
	}

	var interviewID string
	var transcript string
	inv, err := h.q.GetInterviewBySubmissionID(ctx, pgtype.UUID{Bytes: subID, Valid: true})
	if err == nil {
		interviewID = uuid.UUID(inv.InterviewID.Bytes).String()
		msgs, _ := h.q.ListMessagesByInterview(ctx, inv.InterviewID)
		for _, m := range msgs {
			transcript += "[" + m.Sender + "] " + m.Content + "\n"
		}
	}
	if transcript == "" {
		transcript = "(no viva messages)"
	}

	var studentProfile *services.StudentProfilePayload
	profileRec, err := h.q.GetLatestStudentProfileByStudent(ctx, sub.StudentID)
	if err == nil {
		var p services.StudentProfilePayload
		if jsonErr := json.Unmarshal(profileRec.Profile, &p); jsonErr == nil {
			studentProfile = &p
		}
	}

	rubric, _ := h.q.GetRubricByID(ctx, sub.RubricID)
	opts := services.GenerateAuthorshipReportOpts{
		RubricTitle:       rubric.Title,
		SubmissionSummary: submissionSummary,
		Transcript:        transcript,
		InterviewID:       interviewID,
		ArtifactIDs:       artifactIDs,
		StudentProfile:    studentProfile,
	}
	reportPayload, err := h.llm.GenerateAuthorshipReport(ctx, opts)
	if err != nil {
		http.Error(w, "failed to generate report: "+err.Error(), http.StatusInternalServerError)
		return
	}
	reportJSON, err := reportPayload.ToJSONB()
	if err != nil {
		http.Error(w, "failed to serialize report", http.StatusInternalServerError)
		return
	}

	var interviewIDPg pgtype.UUID
	if interviewID != "" {
		parsed, _ := uuid.Parse(interviewID)
		interviewIDPg = pgtype.UUID{Bytes: parsed, Valid: true}
	}
	rec, err := h.q.CreateAuthorshipReport(ctx, db.CreateAuthorshipReportParams{
		SubmissionID: pgtype.UUID{Bytes: subID, Valid: true},
		InterviewID:  interviewIDPg,
		Report:       reportJSON,
	})
	if err != nil {
		http.Error(w, "failed to save report: "+err.Error(), http.StatusInternalServerError)
		return
	}

	_ = h.q.UpdateSubmissionStatus(ctx, db.UpdateSubmissionStatusParams{
		SubmissionID: pgtype.UUID{Bytes: subID, Valid: true},
		Status:       "report_ready",
	})

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"reportId":      rec.ReportID.Bytes,
		"submissionId":  subID,
		"report":        reportPayload,
		"createdAt":     rec.CreatedAt.Time.Format("2006-01-02T15:04:05Z07:00"),
	})
}

// GetAuthorship returns the latest authorship report for the submission.
func (h *SubmissionHandler) GetAuthorship(w http.ResponseWriter, r *http.Request) {
	subIDStr := chi.URLParam(r, "id")
	subID, err := uuid.Parse(subIDStr)
	if err != nil {
		http.Error(w, "invalid submission id", http.StatusBadRequest)
		return
	}
	ctx := r.Context()
	rec, err := h.q.GetLatestAuthorshipReportBySubmission(ctx, pgtype.UUID{Bytes: subID, Valid: true})
	if err != nil {
		if err == pgx.ErrNoRows {
			http.Error(w, "no authorship report found for this submission", http.StatusNotFound)
			return
		}
		http.Error(w, "failed to get report", http.StatusInternalServerError)
		return
	}
	var report services.AuthorshipReportPayload
	if err := json.Unmarshal(rec.Report, &report); err != nil {
		http.Error(w, "failed to decode report", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"reportId":     rec.ReportID.Bytes,
		"submissionId": rec.SubmissionID.Bytes,
		"report":       report,
		"createdAt":    rec.CreatedAt.Time.Format("2006-01-02T15:04:05Z07:00"),
	})
}
