package rubricparser

// ParsedCriterion is one criterion extracted from rubric raw text (LLM output).
type ParsedCriterion struct {
	Name        string             `json:"name"`
	Description string             `json:"description"`
	Weight      float64            `json:"weight"`
	OrderIndex  int                `json:"orderIndex"`
	Levels      map[string]string  `json:"levels,omitempty"` // e.g. {"A": "Full marks", "B": "Partial"}
}

// ParsedQuestion is one question in the initial question plan (LLM output).
type ParsedQuestion struct {
	Prompt        string `json:"prompt"`
	OrderIndex    int    `json:"orderIndex"`
	CriterionName string `json:"criterionName,omitempty"` // optional: link to criterion by name
}

// ParsedQuestionPlan is the initial question plan (LLM output).
type ParsedQuestionPlan struct {
	Title        string          `json:"title"`
	Instructions string          `json:"instructions"`
	Questions    []ParsedQuestion `json:"questions"`
}

// ParseRubricOutput is the full LLM one-shot parse result.
type ParseRubricOutput struct {
	Criteria    []ParsedCriterion `json:"criteria"`
	QuestionPlan ParsedQuestionPlan `json:"questionPlan"`
}
