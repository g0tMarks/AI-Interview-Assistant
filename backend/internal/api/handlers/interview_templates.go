package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/g0tMarks/AI-Interview-Assistant/backend/internal/db"
	"github.com/g0tMarks/AI-Interview-Assistant/backend/internal/services"
)

type InterviewTemplateHandler struct {
	q          *db.Queries
	llmService services.LLMService
}

func NewInterviewTemplateHandler(q *db.Queries, llmService services.LLMService) *InterviewTemplateHandler {
	return &InterviewTemplateHandler{
		q:          q,
		llmService: llmService,
	}
}

type CreateInterviewTemplateRequest struct {
	RubricID            uuid.UUID       `json:"rubricId"`
	Title               string          `json:"title"`
	Instructions        string          `json:"instructions"`
	Config              json.RawMessage `json:"config"`
	Status              string          `json:"status"`
	CurriculumSubject   string          `json:"curriculumSubject"`
	CurriculumLevelBand string          `json:"curriculumLevelBand"`
}

type InterviewTemplateResponse struct {
	InterviewPlanID     uuid.UUID          `json:"interviewPlanId"`
	RubricID            uuid.UUID          `json:"rubricId"`
	Title               string             `json:"title"`
	Instructions        *string            `json:"instructions"`
	Config              json.RawMessage    `json:"config"`
	Status              string             `json:"status"`
	CurriculumSubject   *string            `json:"curriculumSubject"`
	CurriculumLevelBand *string            `json:"curriculumLevelBand"`
	CreatedAt           pgtype.Timestamptz `json:"createdAt"`
	UpdatedAt           pgtype.Timestamptz `json:"updatedAt"`
}

func (h *InterviewTemplateHandler) CreateInterviewTemplate(w http.ResponseWriter, r *http.Request) {
	var req CreateInterviewTemplateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid JSON body", http.StatusBadRequest)
		return
	}

	// Validate required fields
	req.Title = strings.TrimSpace(req.Title)
	if req.Title == "" {
		http.Error(w, "title is required", http.StatusBadRequest)
		return
	}
	if req.RubricID == uuid.Nil {
		http.Error(w, "rubricId is required", http.StatusBadRequest)
		return
	}

	// Validate status enum
	status := req.Status
	if status == "" {
		status = "draft" // default
	} else if status != "draft" && status != "in_progress" && status != "completed" {
		http.Error(w, "invalid status: must be one of draft, in_progress, completed", http.StatusBadRequest)
		return
	}

	ctx := r.Context()

	// Verify rubric exists and fetch it
	rubricIDPgtype := pgtype.UUID{
		Bytes: req.RubricID,
		Valid: true,
	}
	rubric, err := h.q.GetRubricByID(ctx, rubricIDPgtype)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			http.Error(w, "rubric not found", http.StatusNotFound)
			return
		}
		http.Error(w, "failed to verify rubric", http.StatusInternalServerError)
		return
	}

	// Generate interview instructions using LLM if not provided
	instructionsText := req.Instructions
	if instructionsText == "" && h.llmService != nil {
		// Call LLM to generate instructions based on the rubric
		generatedInstructions, err := h.llmService.GenerateInterviewInstructions(ctx, rubric.Title, rubric.RawText)
		if err != nil {
			// Log the error but don't fail the request - allow manual instructions
			// In production, you might want to handle this differently
			// For now, we'll continue without LLM-generated instructions
			// You could also return an error if LLM is required
		} else {
			instructionsText = generatedInstructions
		}
	}

	// Handle config default
	configBytes := []byte("{}")
	if len(req.Config) > 0 {
		// Validate JSON
		var configMap map[string]interface{}
		if err := json.Unmarshal(req.Config, &configMap); err != nil {
			http.Error(w, "invalid config: must be valid JSON", http.StatusBadRequest)
			return
		}
		configBytes = req.Config
	}

	// Convert to database params
	instructions := pgtype.Text{}
	if instructionsText != "" {
		instructions.String = instructionsText
		instructions.Valid = true
	}

	curriculumSubject := pgtype.Text{}
	if req.CurriculumSubject != "" {
		curriculumSubject.String = req.CurriculumSubject
		curriculumSubject.Valid = true
	}

	curriculumLevelBand := pgtype.Text{}
	if req.CurriculumLevelBand != "" {
		curriculumLevelBand.String = req.CurriculumLevelBand
		curriculumLevelBand.Valid = true
	}

	plan, err := h.q.CreateInterviewPlan(ctx, db.CreateInterviewPlanParams{
		RubricID:            rubricIDPgtype,
		Title:               req.Title,
		Instructions:        instructions,
		Config:              configBytes,
		Status:              status,
		CurriculumSubject:   curriculumSubject,
		CurriculumLevelBand: curriculumLevelBand,
	})
	if err != nil {
		http.Error(w, "failed to create interview template: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Convert to response
	var interviewPlanID uuid.UUID
	if plan.InterviewPlanID.Valid {
		interviewPlanID = plan.InterviewPlanID.Bytes
	}

	var rubricIDResp uuid.UUID
	if plan.RubricID.Valid {
		rubricIDResp = plan.RubricID.Bytes
	}

	var instructionsResp *string
	if plan.Instructions.Valid {
		instructionsResp = &plan.Instructions.String
	}

	var curriculumSubjectResp *string
	if plan.CurriculumSubject.Valid {
		curriculumSubjectResp = &plan.CurriculumSubject.String
	}

	var curriculumLevelBandResp *string
	if plan.CurriculumLevelBand.Valid {
		curriculumLevelBandResp = &plan.CurriculumLevelBand.String
	}

	resp := InterviewTemplateResponse{
		InterviewPlanID:     interviewPlanID,
		RubricID:            rubricIDResp,
		Title:               plan.Title,
		Instructions:        instructionsResp,
		Config:              plan.Config,
		Status:              plan.Status,
		CurriculumSubject:   curriculumSubjectResp,
		CurriculumLevelBand: curriculumLevelBandResp,
		CreatedAt:           plan.CreatedAt,
		UpdatedAt:           plan.UpdatedAt,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(resp)
}
