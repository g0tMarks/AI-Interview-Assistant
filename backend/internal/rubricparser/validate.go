package rubricparser

import (
	"fmt"
	"strings"
)

// Validate checks the parsed output and returns a clear error if invalid.
func Validate(out *ParseRubricOutput) error {
	if out == nil {
		return fmt.Errorf("parse output is nil")
	}
	if len(out.Criteria) == 0 {
		return fmt.Errorf("criteria list is empty")
	}
	seenNames := make(map[string]bool)
	for i, c := range out.Criteria {
		name := strings.TrimSpace(c.Name)
		if name == "" {
			return fmt.Errorf("criterion at index %d has empty name", i)
		}
		if seenNames[name] {
			return fmt.Errorf("duplicate criterion name %q", name)
		}
		seenNames[name] = true
		if c.Weight < 0 {
			return fmt.Errorf("criterion %q has negative weight", name)
		}
	}
	if out.QuestionPlan.Questions == nil {
		return fmt.Errorf("questionPlan.questions is missing or null")
	}
	for i, q := range out.QuestionPlan.Questions {
		prompt := strings.TrimSpace(q.Prompt)
		if prompt == "" {
			return fmt.Errorf("question at index %d has empty prompt", i)
		}
	}
	return nil
}
