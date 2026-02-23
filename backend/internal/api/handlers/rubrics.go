package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/g0tMarks/AI-Interview-Assistant/backend/internal/db"
	"github.com/g0tMarks/AI-Interview-Assistant/backend/internal/extraction"
	"github.com/g0tMarks/AI-Interview-Assistant/backend/internal/rubricparser"
	"github.com/g0tMarks/AI-Interview-Assistant/backend/internal/services"
)

type RubricHandler struct {
	q           *db.Queries
	llmService  services.LLMService
	txBeginner  apiTxBeginner
}

type apiTxBeginner interface {
	Begin(ctx context.Context) (pgx.Tx, error)
}

func NewRubricHandler(q *db.Queries, llmService services.LLMService, txBeginner apiTxBeginner) *RubricHandler {
	return &RubricHandler{q: q, llmService: llmService, txBeginner: txBeginner}
}

type CreateRubricRequest struct {
	TeacherID   uuid.UUID `json:"teacherId"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	RawText     string    `json:"rawText"`
}

type RubricResponse struct {
	RubricID    uuid.UUID          `json:"rubricId"`
	TeacherID   uuid.UUID          `json:"teacherId"`
	Title       string             `json:"title"`
	Description string             `json:"description"`
	RawText     string             `json:"rawText"`
	IsEnabled   bool               `json:"isEnabled"`
	CreatedAt   pgtype.Timestamptz `json:"createdAt"`
	UpdatedAt   pgtype.Timestamptz `json:"updatedAt"`
}

func (h *RubricHandler) CreateRubric(w http.ResponseWriter, r *http.Request) {
	var req CreateRubricRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid JSON body", http.StatusBadRequest)
		return
	}

	// Minimal validation: we just ensure there's enough info to be useful
	if req.Title == "" {
		http.Error(w, "title is required", http.StatusBadRequest)
		return
	}
	if req.TeacherID == uuid.Nil {
		http.Error(w, "teacherId is required", http.StatusBadRequest)
		return
	}
	if req.RawText == "" {
		http.Error(w, "rawText is required", http.StatusBadRequest)
		return
	}

	ctx := r.Context()

	// Convert uuid.UUID to pgtype.UUID
	teacherID := pgtype.UUID{
		Bytes: req.TeacherID,
		Valid: true,
	}

	// Convert string to pgtype.Text for description
	description := pgtype.Text{}
	if req.Description != "" {
		description.String = req.Description
		description.Valid = true
	}

	rubric, err := h.q.CreateRubric(ctx, db.CreateRubricParams{
		TeacherID:   teacherID,
		Title:       req.Title,
		Description: description,
		RawText:     req.RawText,
	})
	if err != nil {
		// Log the actual error for debugging (in production, be more careful about exposing errors)
		http.Error(w, "failed to save rubric: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Convert pgtype.UUID to uuid.UUID for response
	var rubricID uuid.UUID
	if rubric.RubricID.Valid {
		rubricID = rubric.RubricID.Bytes
	}

	var teacherIDResp uuid.UUID
	if rubric.TeacherID.Valid {
		teacherIDResp = rubric.TeacherID.Bytes
	}

	resp := RubricResponse{
		RubricID:    rubricID,
		TeacherID:   teacherIDResp,
		Title:       rubric.Title,
		Description: rubric.Description.String,
		RawText:     rubric.RawText,
		IsEnabled:   rubric.IsEnabled,
		CreatedAt:   rubric.CreatedAt,
		UpdatedAt:   rubric.UpdatedAt,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(resp)
}

func (h *RubricHandler) ListRubrics(w http.ResponseWriter, r *http.Request) {
	// Extract teacherId from query parameter
	teacherIdStr := r.URL.Query().Get("teacherId")
	if teacherIdStr == "" {
		http.Error(w, "teacherId is required", http.StatusBadRequest)
		return
	}

	// Validate UUID format
	teacherID, err := uuid.Parse(teacherIdStr)
	if err != nil {
		http.Error(w, "invalid teacherId format", http.StatusBadRequest)
		return
	}

	ctx := r.Context()

	// Convert to pgtype.UUID
	teacherIDPgtype := pgtype.UUID{
		Bytes: teacherID,
		Valid: true,
	}

	// Query database
	rubrics, err := h.q.ListRubricsByTeacher(ctx, teacherIDPgtype)
	if err != nil {
		http.Error(w, "failed to retrieve rubrics", http.StatusInternalServerError)
		return
	}

	// Convert to response format
	resp := make([]RubricResponse, len(rubrics))
	for i, rubric := range rubrics {
		var rubricID uuid.UUID
		if rubric.RubricID.Valid {
			rubricID = rubric.RubricID.Bytes
		}

		var teacherIDResp uuid.UUID
		if rubric.TeacherID.Valid {
			teacherIDResp = rubric.TeacherID.Bytes
		}

		resp[i] = RubricResponse{
			RubricID:    rubricID,
			TeacherID:   teacherIDResp,
			Title:       rubric.Title,
			Description: rubric.Description.String,
			RawText:     rubric.RawText,
			IsEnabled:   rubric.IsEnabled,
			CreatedAt:   rubric.CreatedAt,
			UpdatedAt:   rubric.UpdatedAt,
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(resp)
}

// UploadRubricFile handles file uploads (PDF/DOCX), extracts text, and creates a rubric.
// Expects multipart/form-data with:
//   - file: the PDF or DOCX file
//   - teacherId: UUID of the teacher
//   - title: title for the rubric (optional, defaults to filename)
//   - description: description for the rubric (optional)
func (h *RubricHandler) UploadRubricFile(w http.ResponseWriter, r *http.Request) {
	// Parse multipart form (max 25 MB)
	if err := r.ParseMultipartForm(25 << 20); err != nil {
		http.Error(w, "failed to parse multipart form", http.StatusBadRequest)
		return
	}

	// Get file from form
	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "file is required", http.StatusBadRequest)
		return
	}
	defer func() { _ = file.Close() }()

	// Get teacherId from form
	teacherIdStr := r.FormValue("teacherId")
	if teacherIdStr == "" {
		http.Error(w, "teacherId is required", http.StatusBadRequest)
		return
	}

	teacherID, err := uuid.Parse(teacherIdStr)
	if err != nil {
		http.Error(w, "invalid teacherId format", http.StatusBadRequest)
		return
	}

	// Get optional title and description
	title := strings.TrimSpace(r.FormValue("title"))
	if title == "" {
		// Default to filename without extension
		filename := header.Filename
		if idx := strings.LastIndex(filename, "."); idx > 0 {
			title = filename[:idx]
		} else {
			title = filename
		}
	}

	description := strings.TrimSpace(r.FormValue("description"))

	// Read file into memory for extraction (needs to be seekable for PDF)
	fileBytes, err := io.ReadAll(file)
	if err != nil {
		http.Error(w, "failed to read file", http.StatusInternalServerError)
		return
	}

	// Extract text from file
	contentType := header.Header.Get("Content-Type")
	if contentType == "" {
		// Try to detect from filename
		filename := strings.ToLower(header.Filename)
		if strings.HasSuffix(filename, ".pdf") {
			contentType = "application/pdf"
		} else if strings.HasSuffix(filename, ".docx") {
			contentType = "application/vnd.openxmlformats-officedocument.wordprocessingml.document"
		} else if strings.HasSuffix(filename, ".doc") {
			contentType = "application/msword"
		}
	}

	reader := bytes.NewReader(fileBytes)
	extractedText, err := extraction.ExtractText(reader, contentType, header.Filename)
	if err != nil {
		if errors.Is(err, extraction.ErrUnsupportedFormat) {
			http.Error(w, "unsupported file format. Only PDF and DOCX files are supported", http.StatusBadRequest)
			return
		}
		if errors.Is(err, extraction.ErrEmptyDocument) {
			msg := "file contains no extractable text"
			if strings.HasSuffix(strings.ToLower(header.Filename), ".pdf") {
				msg += " (PDF may be image-only/scanned; use a PDF with a text layer or run OCR first)"
			}
			msg += "."
			http.Error(w, msg, http.StatusBadRequest)
			return
		}
		http.Error(w, "failed to extract text from file: "+err.Error(), http.StatusInternalServerError)
		return
	}

	ctx := r.Context()

	// Convert uuid.UUID to pgtype.UUID
	teacherIDPgtype := pgtype.UUID{
		Bytes: teacherID,
		Valid: true,
	}

	// Convert string to pgtype.Text for description
	descriptionPgtype := pgtype.Text{}
	if description != "" {
		descriptionPgtype.String = description
		descriptionPgtype.Valid = true
	}

	// Create rubric with extracted text
	rubric, err := h.q.CreateRubric(ctx, db.CreateRubricParams{
		TeacherID:   teacherIDPgtype,
		Title:       title,
		Description: descriptionPgtype,
		RawText:     extractedText,
	})
	if err != nil {
		http.Error(w, "failed to save rubric: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Convert pgtype.UUID to uuid.UUID for response
	var rubricID uuid.UUID
	if rubric.RubricID.Valid {
		rubricID = rubric.RubricID.Bytes
	}

	var teacherIDResp uuid.UUID
	if rubric.TeacherID.Valid {
		teacherIDResp = rubric.TeacherID.Bytes
	}

	resp := RubricResponse{
		RubricID:    rubricID,
		TeacherID:   teacherIDResp,
		Title:       rubric.Title,
		Description: rubric.Description.String,
		RawText:     rubric.RawText,
		IsEnabled:   rubric.IsEnabled,
		CreatedAt:   rubric.CreatedAt,
		UpdatedAt:   rubric.UpdatedAt,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(resp)
}

// ParseRubricResponse is the response for POST /rubrics/{id}/parse.
type ParseRubricResponse struct {
	RubricID        uuid.UUID `json:"rubricId"`
	CriteriaCount   int       `json:"criteriaCount"`
	InterviewPlanID uuid.UUID `json:"interviewPlanId"`
	QuestionCount   int       `json:"questionCount"`
}

// ParseRubric runs LLM one-shot parse, validates, and stores criteria + interview plan.
// Expects rubric ID in path (chi URL param "id").
func (h *RubricHandler) ParseRubric(w http.ResponseWriter, r *http.Request) {
	rubricIDStr := chi.URLParam(r, "id")
	if rubricIDStr == "" {
		http.Error(w, "rubric id is required", http.StatusBadRequest)
		return
	}
	rubricID, err := uuid.Parse(rubricIDStr)
	if err != nil {
		http.Error(w, "invalid rubric id", http.StatusBadRequest)
		return
	}

	if h.llmService == nil {
		http.Error(w, "LLM service not configured", http.StatusServiceUnavailable)
		return
	}
	if h.txBeginner == nil {
		http.Error(w, "database transactions not available", http.StatusServiceUnavailable)
		return
	}

	ctx := r.Context()
	rubricIDPg := pgtype.UUID{Bytes: rubricID, Valid: true}

	rubric, err := h.q.GetRubricByID(ctx, rubricIDPg)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			http.Error(w, "rubric not found", http.StatusNotFound)
			return
		}
		http.Error(w, "failed to get rubric", http.StatusInternalServerError)
		return
	}
	if strings.TrimSpace(rubric.RawText) == "" {
		http.Error(w, "rubric has no raw text to parse (upload a file or set rawText first)", http.StatusBadRequest)
		return
	}

	parsed, err := h.llmService.ParseRubric(ctx, rubric.Title, rubric.RawText)
	if err != nil {
		http.Error(w, "parse failed: "+err.Error(), http.StatusUnprocessableEntity)
		return
	}
	if err := rubricparser.Validate(parsed); err != nil {
		http.Error(w, "validation failed: "+err.Error(), http.StatusUnprocessableEntity)
		return
	}

	tx, err := h.txBeginner.Begin(ctx)
	if err != nil {
		http.Error(w, "failed to start transaction", http.StatusInternalServerError)
		return
	}
	defer func() { _ = tx.Rollback(ctx) }()
	q := h.q.WithTx(tx)

	// Replace existing: delete plans (cascades to questions), then criteria
	if err := q.DeletePlansByRubric(ctx, rubricIDPg); err != nil {
		http.Error(w, "failed to clear existing plans", http.StatusInternalServerError)
		return
	}
	if err := q.DeleteCriteriaByRubric(ctx, rubricIDPg); err != nil {
		http.Error(w, "failed to clear existing criteria", http.StatusInternalServerError)
		return
	}

	// Create criteria and collect IDs by name for linking questions
	criterionIDByName := make(map[string]pgtype.UUID)
	descPg := func(s string) pgtype.Text {
		t := pgtype.Text{}
		if s != "" {
			t.String = s
			t.Valid = true
		}
		return t
	}
	for i, c := range parsed.Criteria {
		weight := pgtype.Numeric{}
		_ = weight.Scan(c.Weight)
		levelsJSON := []byte("null")
		if len(c.Levels) > 0 {
			levelsJSON, _ = json.Marshal(c.Levels)
		}
		orderIdx := int32(i)
		if c.OrderIndex >= 0 {
			orderIdx = int32(c.OrderIndex)
		}
		created, err := q.CreateRubricCriterion(ctx, db.CreateRubricCriterionParams{
			RubricID:    rubricIDPg,
			Name:        strings.TrimSpace(c.Name),
			Description: descPg(c.Description),
			Weight:      weight,
			OrderIndex:  orderIdx,
			Levels:      levelsJSON,
		})
		if err != nil {
			http.Error(w, "failed to create criterion: "+err.Error(), http.StatusInternalServerError)
			return
		}
		criterionIDByName[strings.TrimSpace(c.Name)] = created.RubricCriterionID
	}

	// Create interview plan
	planTitle := strings.TrimSpace(parsed.QuestionPlan.Title)
	if planTitle == "" {
		planTitle = rubric.Title + " – Interview plan"
	}
	instructionsPg := pgtype.Text{}
	if parsed.QuestionPlan.Instructions != "" {
		instructionsPg.String = parsed.QuestionPlan.Instructions
		instructionsPg.Valid = true
	}
	plan, err := q.CreateInterviewPlan(ctx, db.CreateInterviewPlanParams{
		RubricID:            rubricIDPg,
		Title:               planTitle,
		Instructions:        instructionsPg,
		Config:              []byte("{}"),
		Status:              string(db.AppInterviewStatusDraft),
		CurriculumSubject:   pgtype.Text{},
		CurriculumLevelBand: pgtype.Text{},
	})
	if err != nil {
		http.Error(w, "failed to create interview plan: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Create questions (optionally link to criterion by name)
	for i, qu := range parsed.QuestionPlan.Questions {
		orderIdx := int32(i)
		if qu.OrderIndex >= 0 {
			orderIdx = int32(qu.OrderIndex)
		}
		criterionID := pgtype.UUID{Valid: false}
		if name := strings.TrimSpace(qu.CriterionName); name != "" {
			if id, ok := criterionIDByName[name]; ok {
				criterionID = id
			}
		}
		_, err := q.CreateInterviewQuestion(ctx, db.CreateInterviewQuestionParams{
			InterviewPlanID:   pgtype.UUID{Bytes: plan.InterviewPlanID.Bytes, Valid: true},
			RubricCriterionID: criterionID,
			Prompt:            strings.TrimSpace(qu.Prompt),
			QuestionType:      "open",
			OrderIndex:        orderIdx,
			IsActive:          true,
			FollowUpToID:      pgtype.UUID{},
			FollowUpCondition: pgtype.Text{},
		})
		if err != nil {
			http.Error(w, "failed to create question: "+err.Error(), http.StatusInternalServerError)
			return
		}
	}

	if err := tx.Commit(ctx); err != nil {
		http.Error(w, "failed to commit transaction", http.StatusInternalServerError)
		return
	}

	var planID uuid.UUID
	if plan.InterviewPlanID.Valid {
		planID = plan.InterviewPlanID.Bytes
	}
	resp := ParseRubricResponse{
		RubricID:        rubricID,
		CriteriaCount:   len(parsed.Criteria),
		InterviewPlanID: planID,
		QuestionCount:   len(parsed.QuestionPlan.Questions),
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(resp)
}
