package services

import (
	"context"
	"encoding/json"
)

// AuthorshipReportPayload is the structured report stored in DB (JSONB) and returned by the API.
type AuthorshipReportPayload struct {
	OverallAssessment OverallAssessment   `json:"overall_assessment"`
	EvidenceSignals   []EvidenceSignal    `json:"evidence_signals"`
	RiskFlags         []RiskFlag          `json:"risk_flags"`
	RecommendedFollowups []RecommendedFollowup `json:"recommended_followups"`
	RubricAlignment   map[string]string   `json:"rubric_alignment,omitempty"`
	Provenance        Provenance          `json:"provenance"`
}

type OverallAssessment struct {
	Level      string  `json:"level"`      // confident, moderate, low, concern
	Confidence float64 `json:"confidence"` // 0.0–1.0
	Summary    string  `json:"summary"`
}

type EvidenceSignal struct {
	Signal                 string   `json:"signal"`
	Strength                string   `json:"strength"` // strong, moderate, weak
	Explanation             string   `json:"explanation"`
	SupportingQuotesOrRefs  []string `json:"supporting_quotes_or_refs,omitempty"`
}

type RiskFlag struct {
	Flag     string `json:"flag"`
	Severity string `json:"severity"` // high, medium, low
	Details  string `json:"details"`
}

type RecommendedFollowup struct {
	Question string `json:"question"`
	Why      string `json:"why"`
}

type Provenance struct {
	SubmissionArtifactIDs []string `json:"submission_artifact_ids,omitempty"`
	InterviewID           string  `json:"interview_id,omitempty"`
	ReportGeneratedAt     string  `json:"report_generated_at"`
}

// AuthorshipReportGenerator generates an authorship report from submission content and viva transcript.
// Implementations can use LLM or rule-based logic; abstracted so it can be swapped.
type AuthorshipReportGenerator interface {
	GenerateAuthorshipReport(ctx context.Context, opts GenerateAuthorshipReportOpts) (*AuthorshipReportPayload, error)
}

type GenerateAuthorshipReportOpts struct {
	RubricTitle       string
	SubmissionSummary string // concatenated or summarized artifact text
	Transcript        string // viva messages as text
	InterviewID       string
	ArtifactIDs       []string
}

// ToJSONB returns the payload as JSON bytes for storing in DB.
func (p *AuthorshipReportPayload) ToJSONB() ([]byte, error) {
	return json.Marshal(p)
}
