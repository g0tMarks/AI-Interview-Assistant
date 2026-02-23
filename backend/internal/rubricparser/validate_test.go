package rubricparser

import (
	"testing"
)

func TestValidate(t *testing.T) {
	tests := []struct {
		name    string
		out     *ParseRubricOutput
		wantErr bool
	}{
		{
			name:    "nil",
			out:     nil,
			wantErr: true,
		},
		{
			name: "empty criteria",
			out: &ParseRubricOutput{
				Criteria: nil,
				QuestionPlan: ParsedQuestionPlan{
					Questions: []ParsedQuestion{{Prompt: "Q1", OrderIndex: 0}},
				},
			},
			wantErr: true,
		},
		{
			name: "criterion empty name",
			out: &ParseRubricOutput{
				Criteria: []ParsedCriterion{{Name: "", Description: "d", Weight: 1}},
				QuestionPlan: ParsedQuestionPlan{
					Questions: []ParsedQuestion{{Prompt: "Q1", OrderIndex: 0}},
				},
			},
			wantErr: true,
		},
		{
			name: "duplicate criterion name",
			out: &ParseRubricOutput{
				Criteria: []ParsedCriterion{
					{Name: "A", Description: "d", Weight: 1},
					{Name: "A", Description: "d2", Weight: 1},
				},
				QuestionPlan: ParsedQuestionPlan{
					Questions: []ParsedQuestion{{Prompt: "Q1", OrderIndex: 0}},
				},
			},
			wantErr: true,
		},
		{
			name: "negative weight",
			out: &ParseRubricOutput{
				Criteria: []ParsedCriterion{{Name: "A", Description: "d", Weight: -1}},
				QuestionPlan: ParsedQuestionPlan{
					Questions: []ParsedQuestion{{Prompt: "Q1", OrderIndex: 0}},
				},
			},
			wantErr: true,
		},
		{
			name: "questions nil",
			out: &ParseRubricOutput{
				Criteria:     []ParsedCriterion{{Name: "A", Description: "d", Weight: 1}},
				QuestionPlan: ParsedQuestionPlan{Questions: nil},
			},
			wantErr: true,
		},
		{
			name: "empty prompt",
			out: &ParseRubricOutput{
				Criteria: []ParsedCriterion{{Name: "A", Description: "d", Weight: 1}},
				QuestionPlan: ParsedQuestionPlan{
					Questions: []ParsedQuestion{{Prompt: "", OrderIndex: 0}},
				},
			},
			wantErr: true,
		},
		{
			name: "valid",
			out: &ParseRubricOutput{
				Criteria: []ParsedCriterion{
					{Name: "C1", Description: "First", Weight: 1},
					{Name: "C2", Description: "Second", Weight: 0.5},
				},
				QuestionPlan: ParsedQuestionPlan{
					Title:        "Plan",
					Instructions: "Instructions",
					Questions: []ParsedQuestion{
						{Prompt: "Question one?", OrderIndex: 0},
						{Prompt: "Question two?", OrderIndex: 1},
					},
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Validate(tt.out)
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
