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
	"github.com/g0tMarks/AI-Interview-Assistant/backend/internal/engine"
	"github.com/g0tMarks/AI-Interview-Assistant/backend/internal/evaluation"
	"github.com/g0tMarks/AI-Interview-Assistant/backend/internal/safety"
)

type InterviewHandler struct {
	q    *db.Queries
	eng  *engine.Engine
	eval *evaluation.Runner
}

func NewInterviewHandler(q *db.Queries, eng *engine.Engine, eval *evaluation.Runner) *InterviewHandler {
	return &InterviewHandler{q: q, eng: eng, eval: eval}
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

type CreateInterviewMessageRequest struct {
	Sender              string     `json:"sender"`
	InterviewQuestionID *uuid.UUID `json:"interviewQuestionId,omitempty"`
	Content             string     `json:"content"`
}

type InterviewMessageResponse struct {
	InterviewMessageID  uuid.UUID          `json:"interviewMessageId"`
	InterviewID         uuid.UUID          `json:"interviewId"`
	Sender              string             `json:"sender"`
	InterviewQuestionID *uuid.UUID         `json:"interviewQuestionId,omitempty"`
	Content             string             `json:"content"`
	CreatedAt           pgtype.Timestamptz `json:"createdAt"`
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

// CreateMessage handles POST /interviews/{id}/messages
func (h *InterviewHandler) CreateMessage(w http.ResponseWriter, r *http.Request) {
	// Extract interview id from path parameter
	idStr := chi.URLParam(r, "id")
	if idStr == "" {
		http.Error(w, "interview ID is required", http.StatusBadRequest)
		return
	}

	interviewID, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "invalid interview ID format", http.StatusBadRequest)
		return
	}

	var req CreateInterviewMessageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid JSON body", http.StatusBadRequest)
		return
	}

	req.Sender = strings.TrimSpace(req.Sender)
	if req.Sender != "ai" && req.Sender != "user" {
		http.Error(w, "sender must be either 'ai' or 'user'", http.StatusBadRequest)
		return
	}
	req.Content = strings.TrimSpace(req.Content)
	if req.Content == "" {
		http.Error(w, "content is required", http.StatusBadRequest)
		return
	}
	// Sanitize and bound user-supplied content before storing or using it later in LLM prompts.
	sanitized, err := safety.SanitizeUserMessage(req.Content)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	req.Content = sanitized

	ctx := r.Context()

	// Verify interview exists
	interviewIDPg := pgtype.UUID{
		Bytes: interviewID,
		Valid: true,
	}
	_, err = h.q.GetInterviewByID(ctx, interviewIDPg)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			http.Error(w, "interview not found", http.StatusNotFound)
			return
		}
		http.Error(w, "failed to verify interview", http.StatusInternalServerError)
		return
	}

	// Convert optional interviewQuestionId
	questionIDPg := pgtype.UUID{}
	if req.InterviewQuestionID != nil {
		questionIDPg.Bytes = *req.InterviewQuestionID
		questionIDPg.Valid = true
	}

	msg, err := h.q.CreateInterviewMessage(ctx, db.CreateInterviewMessageParams{
		InterviewID:         interviewIDPg,
		Sender:              req.Sender,
		InterviewQuestionID: questionIDPg,
		Content:             req.Content,
	})
	if err != nil {
		http.Error(w, "failed to create message: "+err.Error(), http.StatusInternalServerError)
		return
	}

	var messageID uuid.UUID
	if msg.InterviewMessageID.Valid {
		messageID = msg.InterviewMessageID.Bytes
	}

	var interviewIDResp uuid.UUID
	if msg.InterviewID.Valid {
		interviewIDResp = msg.InterviewID.Bytes
	}

	var questionIDResp *uuid.UUID
	if msg.InterviewQuestionID.Valid {
		id := uuid.UUID(msg.InterviewQuestionID.Bytes)
		questionIDResp = &id
	}

	resp := InterviewMessageResponse{
		InterviewMessageID:  messageID,
		InterviewID:         interviewIDResp,
		Sender:              msg.Sender,
		InterviewQuestionID: questionIDResp,
		Content:             msg.Content,
		CreatedAt:           msg.CreatedAt,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(resp)
}

// ListMessages handles GET /interviews/{id}/messages
func (h *InterviewHandler) ListMessages(w http.ResponseWriter, r *http.Request) {
	// Extract interview id from path parameter
	idStr := chi.URLParam(r, "id")
	if idStr == "" {
		http.Error(w, "interview ID is required", http.StatusBadRequest)
		return
	}

	interviewID, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "invalid interview ID format", http.StatusBadRequest)
		return
	}

	ctx := r.Context()

	interviewIDPg := pgtype.UUID{
		Bytes: interviewID,
		Valid: true,
	}

	// Verify interview exists
	_, err = h.q.GetInterviewByID(ctx, interviewIDPg)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			http.Error(w, "interview not found", http.StatusNotFound)
			return
		}
		http.Error(w, "failed to verify interview", http.StatusInternalServerError)
		return
	}

	msgs, err := h.q.ListMessagesByInterview(ctx, interviewIDPg)
	if err != nil {
		http.Error(w, "failed to list messages: "+err.Error(), http.StatusInternalServerError)
		return
	}

	resp := make([]InterviewMessageResponse, len(msgs))
	for i, msg := range msgs {
		var messageID uuid.UUID
		if msg.InterviewMessageID.Valid {
			messageID = msg.InterviewMessageID.Bytes
		}

		var interviewIDResp uuid.UUID
		if msg.InterviewID.Valid {
			interviewIDResp = msg.InterviewID.Bytes
		}

		var questionIDResp *uuid.UUID
		if msg.InterviewQuestionID.Valid {
			id := uuid.UUID(msg.InterviewQuestionID.Bytes)
			questionIDResp = &id
		}

		resp[i] = InterviewMessageResponse{
			InterviewMessageID:  messageID,
			InterviewID:         interviewIDResp,
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

// NextResponse is the JSON shape for GET/POST /interviews/{id}/next.
type NextResponse struct {
	Status               string     `json:"status"`
	NextQuestionID       *uuid.UUID `json:"nextQuestionId,omitempty"`
	Prompt               string     `json:"prompt,omitempty"`
	PromptOverride       string     `json:"promptOverride,omitempty"`
	WaitingForQuestionID *uuid.UUID `json:"waitingForQuestionId,omitempty"`
	ClassifiedCategory   string     `json:"classifiedCategory,omitempty"`
}

// GetNext handles GET /interviews/{id}/next — idempotent "current next" step.
func (h *InterviewHandler) GetNext(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	if idStr == "" {
		http.Error(w, "interview ID is required", http.StatusBadRequest)
		return
	}
	interviewID, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "invalid interview ID format", http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	result, err := h.eng.ComputeNext(ctx, interviewID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			http.Error(w, "interview not found", http.StatusNotFound)
			return
		}
		http.Error(w, "failed to compute next: "+err.Error(), http.StatusInternalServerError)
		return
	}

	resp := nextResultToResponse(result)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(resp)
}

// PostNext handles POST /interviews/{id}/next — advance: compute next, and if next is a question, create AI message; if done, mark interview completed.
func (h *InterviewHandler) PostNext(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	if idStr == "" {
		http.Error(w, "interview ID is required", http.StatusBadRequest)
		return
	}
	interviewID, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "invalid interview ID format", http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	result, err := h.eng.ComputeNext(ctx, interviewID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			http.Error(w, "interview not found", http.StatusNotFound)
			return
		}
		http.Error(w, "failed to compute next: "+err.Error(), http.StatusInternalServerError)
		return
	}

	if result.Status == engine.NextStatusNextQuestion && result.NextQuestionID != nil {
		interviewIDPg := pgtype.UUID{Bytes: interviewID, Valid: true}
		questionIDPg := pgtype.UUID{Bytes: *result.NextQuestionID, Valid: true}
		content := result.Prompt
		if result.PromptOverride != "" {
			content = result.PromptOverride
		}
		_, err = h.q.CreateInterviewMessage(ctx, db.CreateInterviewMessageParams{
			InterviewID:         interviewIDPg,
			Sender:             "ai",
			InterviewQuestionID: questionIDPg,
			Content:            content,
		})
		if err != nil {
			http.Error(w, "failed to create AI message: "+err.Error(), http.StatusInternalServerError)
			return
		}
	}

	if result.Status == engine.NextStatusDone {
		interviewIDPg := pgtype.UUID{Bytes: interviewID, Valid: true}
		err = h.q.UpdateInterviewStatus(ctx, db.UpdateInterviewStatusParams{
			InterviewID: interviewIDPg,
			Status:      "completed",
		})
		if err != nil {
			http.Error(w, "failed to complete interview: "+err.Error(), http.StatusInternalServerError)
			return
		}
	}

	resp := nextResultToResponse(result)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(resp)
}

func nextResultToResponse(r *engine.NextResult) NextResponse {
	resp := NextResponse{
		Status:               string(r.Status),
		NextQuestionID:       r.NextQuestionID,
		Prompt:               r.Prompt,
		PromptOverride:       r.PromptOverride,
		WaitingForQuestionID: r.WaitingForQuestionID,
		ClassifiedCategory:   r.ClassifiedCategory,
	}
	return resp
}

// ResultsResponse is the JSON shape for GET /interviews/{id}/results.
type ResultsResponse struct {
	InterviewID       uuid.UUID                 `json:"interviewId"`
	Status            string                     `json:"status"`
	OverallSummary   string                     `json:"overallSummary"`
	Strengths         string                     `json:"strengths"`
	AreasForGrowth    string                     `json:"areasForGrowth"`
	SuggestedNextSteps string                    `json:"suggestedNextSteps"`
	CriterionEvidence []CriterionEvidenceResponse `json:"criterionEvidence"`
	Scoring           *evaluation.ScoringJSON    `json:"scoring,omitempty"`
}

// CriterionEvidenceResponse is one criterion's evidence in the results.
type CriterionEvidenceResponse struct {
	CriterionEvidenceID uuid.UUID  `json:"criterionEvidenceId"`
	RubricCriterionID   uuid.UUID  `json:"rubricCriterionId"`
	Level               string     `json:"level"`
	EvidenceText        string     `json:"evidenceText"`
	ModelConfidence     *float64   `json:"modelConfidence,omitempty"`
}

// GetResults handles GET /interviews/{id}/results. If the interview is completed but not yet evaluated, runs evaluation then returns results.
func (h *InterviewHandler) GetResults(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	if idStr == "" {
		http.Error(w, "interview ID is required", http.StatusBadRequest)
		return
	}
	interviewID, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "invalid interview ID format", http.StatusBadRequest)
		return
	}
	ctx := r.Context()

	interviewIDPg := pgtype.UUID{Bytes: interviewID, Valid: true}
	inv, err := h.q.GetInterviewByID(ctx, interviewIDPg)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			http.Error(w, "interview not found", http.StatusNotFound)
			return
		}
		http.Error(w, "failed to get interview: "+err.Error(), http.StatusInternalServerError)
		return
	}
	if inv.Status != "completed" {
		http.Error(w, "interview is not completed", http.StatusBadRequest)
		return
	}

	if h.eval != nil {
		if err := h.eval.Run(ctx, interviewID); err != nil {
			http.Error(w, "evaluation failed: "+err.Error(), http.StatusInternalServerError)
			return
		}
	}

	summary, err := h.q.GetSummaryByInterviewID(ctx, interviewIDPg)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			http.Error(w, "results not available", http.StatusNotFound)
			return
		}
		http.Error(w, "failed to get summary: "+err.Error(), http.StatusInternalServerError)
		return
	}

	evidenceList, err := h.q.GetCriterionEvidenceBySummaryID(ctx, summary.InterviewSummaryID)
	if err != nil {
		http.Error(w, "failed to get criterion evidence: "+err.Error(), http.StatusInternalServerError)
		return
	}

	overallSummary := ""
	if summary.OverallSummary.Valid {
		overallSummary = summary.OverallSummary.String
	}
	strengths := ""
	if summary.Strengths.Valid {
		strengths = summary.Strengths.String
	}
	areasForGrowth := ""
	if summary.AreasForGrowth.Valid {
		areasForGrowth = summary.AreasForGrowth.String
	}
	suggestedNextSteps := ""
	if summary.SuggestedNextSteps.Valid {
		suggestedNextSteps = summary.SuggestedNextSteps.String
	}

	evidenceResp := make([]CriterionEvidenceResponse, len(evidenceList))
	for i, e := range evidenceList {
		level := ""
		if e.Level.Valid {
			level = e.Level.String
		}
		evidenceText := ""
		if e.EvidenceText.Valid {
			evidenceText = e.EvidenceText.String
		}
		var criterionID, evidenceID uuid.UUID
		if e.RubricCriterionID.Valid {
			criterionID = e.RubricCriterionID.Bytes
		}
		if e.CriterionEvidenceID.Valid {
			evidenceID = e.CriterionEvidenceID.Bytes
		}
		var confidence *float64
		if e.ModelConfidence.Valid {
			f := numericToFloat64(e.ModelConfidence)
			confidence = &f
		}
		evidenceResp[i] = CriterionEvidenceResponse{
			CriterionEvidenceID: evidenceID,
			RubricCriterionID:   criterionID,
			Level:               level,
			EvidenceText:        evidenceText,
			ModelConfidence:     confidence,
		}
	}

	var scoring *evaluation.ScoringJSON
	if len(summary.RawLlmOutput) > 0 {
		if err := json.Unmarshal(summary.RawLlmOutput, &scoring); err == nil {
			// use parsed scoring for consistent API shape
		}
	}

	resp := ResultsResponse{
		InterviewID:        interviewID,
		Status:             inv.Status,
		OverallSummary:    overallSummary,
		Strengths:          strengths,
		AreasForGrowth:     areasForGrowth,
		SuggestedNextSteps: suggestedNextSteps,
		CriterionEvidence:  evidenceResp,
		Scoring:            scoring,
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(resp)
}

// numericToFloat64 converts pgtype.Numeric to float64 (value = Int * 10^Exp, e.g. 85 and -2 -> 0.85).
func numericToFloat64(n pgtype.Numeric) float64 {
	if !n.Valid || n.Int == nil {
		return 0
	}
	f, _ := n.Int.Float64()
	exp := n.Exp
	for exp < 0 {
		f *= 0.1
		exp++
	}
	for exp > 0 {
		f *= 10
		exp--
	}
	return f
}

// SummaryResponse is the JSON shape for GET /interviews/{id}/summary.
type SummaryResponse struct {
	InterviewSummaryID uuid.UUID `json:"interviewSummaryId"`
	InterviewID        uuid.UUID `json:"interviewId"`
	OverallSummary    string    `json:"overallSummary"`
	Strengths         string    `json:"strengths"`
	AreasForGrowth    string    `json:"areasForGrowth"`
	SuggestedNextSteps string   `json:"suggestedNextSteps"`
}

// GetSummary handles GET /interviews/{id}/summary. Returns the summary only (no criterion evidence).
func (h *InterviewHandler) GetSummary(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	if idStr == "" {
		http.Error(w, "interview ID is required", http.StatusBadRequest)
		return
	}
	interviewID, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "invalid interview ID format", http.StatusBadRequest)
		return
	}
	ctx := r.Context()

	interviewIDPg := pgtype.UUID{Bytes: interviewID, Valid: true}
	summary, err := h.q.GetSummaryByInterviewID(ctx, interviewIDPg)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			http.Error(w, "summary not found", http.StatusNotFound)
			return
		}
		http.Error(w, "failed to get summary: "+err.Error(), http.StatusInternalServerError)
		return
	}

	var summaryID, invID uuid.UUID
	if summary.InterviewSummaryID.Valid {
		summaryID = summary.InterviewSummaryID.Bytes
	}
	if summary.InterviewID.Valid {
		invID = summary.InterviewID.Bytes
	}
	overallSummary := ""
	if summary.OverallSummary.Valid {
		overallSummary = summary.OverallSummary.String
	}
	strengths := ""
	if summary.Strengths.Valid {
		strengths = summary.Strengths.String
	}
	areasForGrowth := ""
	if summary.AreasForGrowth.Valid {
		areasForGrowth = summary.AreasForGrowth.String
	}
	suggestedNextSteps := ""
	if summary.SuggestedNextSteps.Valid {
		suggestedNextSteps = summary.SuggestedNextSteps.String
	}

	resp := SummaryResponse{
		InterviewSummaryID:  summaryID,
		InterviewID:         invID,
		OverallSummary:     overallSummary,
		Strengths:          strengths,
		AreasForGrowth:     areasForGrowth,
		SuggestedNextSteps: suggestedNextSteps,
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(resp)
}

