package evaluation

// CriterionForEval is a criterion's name and description passed to EvaluateInterview.
type CriterionForEval struct {
	Name        string
	Description string
	LevelsJSON  string // optional, e.g. `{"A":"...","B":"..."}`
}

// EvalOutput is the structured output from the LLM for interview evaluation.
type EvalOutput struct {
	OverallSummary     string              `json:"overallSummary"`
	Strengths          string              `json:"strengths"`
	AreasForGrowth     string              `json:"areasForGrowth"`
	SuggestedNextSteps string              `json:"suggestedNextSteps"`
	Criteria           []EvalCriterion     `json:"criteria"`
}

// EvalCriterion is per-criterion evaluation from the LLM.
type EvalCriterion struct {
	CriterionName   string   `json:"criterionName"`
	Level           string   `json:"level"`
	EvidenceText    string   `json:"evidenceText"`
	ModelConfidence *float64 `json:"modelConfidence,omitempty"`
}

// ScoringJSON is stored in interview_summaries.raw_llm_output for the results API.
type ScoringJSON struct {
	OverallSummary     string                   `json:"overallSummary"`
	Strengths          string                   `json:"strengths"`
	AreasForGrowth     string                   `json:"areasForGrowth"`
	SuggestedNextSteps string                   `json:"suggestedNextSteps"`
	Scores             map[string]string `json:"scores"` // criterion name -> level
	Criteria           []EvalCriterion    `json:"criteria"`
}
