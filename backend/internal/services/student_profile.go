package services

import "encoding/json"

// StudentProfilePayload is the structured profile stored in DB (JSONB) and returned by the API.
type StudentProfilePayload struct {
	WritingFeatures   WritingFeatures   `json:"writing_features"`
	ReasoningFeatures ReasoningFeatures `json:"reasoning_features"`
	VoiceMarkers      VoiceMarkers      `json:"voice_markers"`
	Provenance        ProfileProvenance `json:"provenance"`
}

type WritingFeatures struct {
	AvgSentenceLength        float64  `json:"avg_sentence_length"`
	LexicalDiversity         float64  `json:"lexical_diversity"`
	ClauseComplexity         string   `json:"clause_complexity"`
	RhetoricalStructure      []string `json:"rhetorical_structure_patterns"`
	CommonErrors             []string `json:"common_errors"`
	ArgumentDepthMarkers     []string `json:"argument_depth_markers"`
}

type ReasoningFeatures struct {
	CausalLanguageUse         []string `json:"causal_language_use"`
	EvidenceIntegration       []string `json:"evidence_integration_patterns"`
	ParagraphCohesionPatterns []string `json:"paragraph_cohesion_patterns"`
}

type VoiceMarkers struct {
	FrequentPhrases      []string `json:"frequent_phrases"`
	PreferredConnectives []string `json:"preferred_connectives"`
	ToneIndicators       []string `json:"tone_indicators"`
}

type ProfileProvenance struct {
	SubmissionIDs []string `json:"submission_ids"`
	ArtifactIDs   []string `json:"artifact_ids"`
	SampleCount   int      `json:"sample_count"`
	GeneratedAt   string   `json:"generated_at"`
}

// GenerateStudentProfileOpts controls generation of a student profile from multiple samples.
type GenerateStudentProfileOpts struct {
	StudentDisplayName string
	Samples            []StudentWritingSample
}

// StudentWritingSample is an individual writing sample used to build a profile.
type StudentWritingSample struct {
	SubmissionID string
	ArtifactID   string
	Text         string
	Context      string
}

// ToJSONB returns the payload as JSON bytes for storing in DB.
func (p *StudentProfilePayload) ToJSONB() ([]byte, error) {
	return json.Marshal(p)
}

