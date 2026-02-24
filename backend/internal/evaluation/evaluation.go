package evaluation

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/g0tMarks/AI-Interview-Assistant/backend/internal/db"
)

// EvalLLM is the subset of LLM needed for evaluation (avoids import cycle with services).
type EvalLLM interface {
	EvaluateInterview(ctx context.Context, rubricTitle string, criteria []CriterionForEval, transcript string) (*EvalOutput, error)
}

// Runner runs the full evaluation flow: load interview data, call LLM, persist summary and criterion evidence.
type Runner struct {
	Queries *db.Queries
	LLM     EvalLLM
}

// NewRunner creates an evaluation runner.
func NewRunner(queries *db.Queries, llm EvalLLM) *Runner {
	return &Runner{Queries: queries, LLM: llm}
}

// Run evaluates a completed interview and persists the summary and criterion evidence.
// It is idempotent: if a summary already exists for the interview, it returns nil (no error).
func (r *Runner) Run(ctx context.Context, interviewID uuid.UUID) error {
	interviewIDPg := pgtype.UUID{Bytes: interviewID, Valid: true}

	inv, err := r.Queries.GetInterviewByID(ctx, interviewIDPg)
	if err != nil {
		if err == pgx.ErrNoRows {
			return fmt.Errorf("interview not found")
		}
		return err
	}
	if inv.Status != "completed" {
		return fmt.Errorf("interview is not completed (status: %s)", inv.Status)
	}

	// Already evaluated?
	_, err = r.Queries.GetSummaryByInterviewID(ctx, interviewIDPg)
	if err == nil {
		return nil // already have a summary
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return err
	}

	plan, err := r.Queries.GetInterviewPlanByID(ctx, inv.InterviewPlanID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return fmt.Errorf("interview plan not found")
		}
		return err
	}
	if !plan.RubricID.Valid {
		return fmt.Errorf("interview plan has no rubric")
	}

	criteria, err := r.Queries.ListCriteriaByRubric(ctx, plan.RubricID)
	if err != nil {
		return err
	}
	if len(criteria) == 0 {
		return fmt.Errorf("rubric has no criteria")
	}

	messages, err := r.Queries.ListMessagesByInterview(ctx, interviewIDPg)
	if err != nil {
		return err
	}
	transcript := buildTranscript(messages)

	rubric, err := r.Queries.GetRubricByID(ctx, plan.RubricID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return fmt.Errorf("rubric not found")
		}
		return err
	}

	criteriaForEval := make([]CriterionForEval, len(criteria))
	nameToCriterionID := make(map[string]uuid.UUID)
	for i, c := range criteria {
		desc := ""
		if c.Description.Valid {
			desc = c.Description.String
		}
		levelsJSON := ""
		if len(c.Levels) > 0 && string(c.Levels) != "null" {
			levelsJSON = string(c.Levels)
		}
		criteriaForEval[i] = CriterionForEval{
			Name:        c.Name,
			Description: desc,
			LevelsJSON:  levelsJSON,
		}
		nameToCriterionID[c.Name] = c.RubricCriterionID.Bytes
	}

	out, err := r.LLM.EvaluateInterview(ctx, rubric.Title, criteriaForEval, transcript)
	if err != nil {
		return fmt.Errorf("LLM evaluation: %w", err)
	}

	// Build scoring JSON for raw_llm_output
	scores := make(map[string]string)
	for _, c := range out.Criteria {
		scores[c.CriterionName] = c.Level
	}
	scoring := ScoringJSON{
		OverallSummary:     out.OverallSummary,
		Strengths:          out.Strengths,
		AreasForGrowth:     out.AreasForGrowth,
		SuggestedNextSteps: out.SuggestedNextSteps,
		Scores:             scores,
		Criteria:           out.Criteria,
	}
	rawJSON, err := json.Marshal(scoring)
	if err != nil {
		return err
	}

	overallSummary := pgtype.Text{String: out.OverallSummary, Valid: out.OverallSummary != ""}
	strengths := pgtype.Text{String: out.Strengths, Valid: out.Strengths != ""}
	areasForGrowth := pgtype.Text{String: out.AreasForGrowth, Valid: out.AreasForGrowth != ""}
	suggestedNextSteps := pgtype.Text{String: out.SuggestedNextSteps, Valid: out.SuggestedNextSteps != ""}

	summary, err := r.Queries.CreateInterviewSummary(ctx, db.CreateInterviewSummaryParams{
		InterviewID:        interviewIDPg,
		OverallSummary:     overallSummary,
		Strengths:          strengths,
		AreasForGrowth:     areasForGrowth,
		SuggestedNextSteps: suggestedNextSteps,
		RawLlmOutput:       rawJSON,
	})
	if err != nil {
		return err
	}

	summaryIDPg := summary.InterviewSummaryID
	for _, c := range out.Criteria {
		criterionID, ok := nameToCriterionID[c.CriterionName]
		if !ok {
			continue // skip unknown criterion name from LLM
		}
		criterionIDPg := pgtype.UUID{Bytes: criterionID, Valid: true}
		level := pgtype.Text{String: c.Level, Valid: c.Level != ""}
		evidenceText := pgtype.Text{String: c.EvidenceText, Valid: c.EvidenceText != ""}
		confidence := pgtype.Numeric{Valid: false}
		if c.ModelConfidence != nil && *c.ModelConfidence >= 0 && *c.ModelConfidence <= 1 {
			confidence = floatToNumeric(*c.ModelConfidence)
		}
		_, err = r.Queries.CreateCriterionEvidence(ctx, db.CreateCriterionEvidenceParams{
			InterviewSummaryID: summaryIDPg,
			RubricCriterionID:  criterionIDPg,
			Level:              level,
			EvidenceText:       evidenceText,
			ModelConfidence:    confidence,
		})
		if err != nil {
			return err
		}
	}

	return nil
}

func buildTranscript(messages []db.AppInterviewMessage) string {
	var b strings.Builder
	for _, m := range messages {
		role := "AI"
		if m.Sender == "user" {
			role = "Student"
		}
		b.WriteString(role + ": " + m.Content + "\n")
	}
	return b.String()
}

func floatToNumeric(f float64) pgtype.Numeric {
	// 0.85 -> Int=85, Exp=-2
	scaled := int64(f * 100)
	return pgtype.Numeric{
		Int:   big.NewInt(scaled),
		Exp:   -2,
		Valid: true,
	}
}
