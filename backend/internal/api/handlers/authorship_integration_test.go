package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/g0tMarks/AI-Interview-Assistant/backend/internal/db"
	"github.com/g0tMarks/AI-Interview-Assistant/backend/internal/evaluation"
	"github.com/g0tMarks/AI-Interview-Assistant/backend/internal/rubricparser"
	"github.com/g0tMarks/AI-Interview-Assistant/backend/internal/services"
)

// mockLLM implements services.LLMService with fixed responses for authorship tests.
type mockLLM struct{}

func (m *mockLLM) GenerateInterviewInstructions(_ context.Context, _, _ string) (string, error) {
	return "mock instructions", nil
}

func (m *mockLLM) ParseRubric(_ context.Context, _, _ string) (*rubricparser.ParseRubricOutput, error) {
	return &rubricparser.ParseRubricOutput{}, nil
}

func (m *mockLLM) ClassifyResponse(_ context.Context, _, _ string) (string, error) {
	return "strong", nil
}

func (m *mockLLM) EvaluateInterview(_ context.Context, _ string, _ []evaluation.CriterionForEval, _ string) (*evaluation.EvalOutput, error) {
	return &evaluation.EvalOutput{}, nil
}

func (m *mockLLM) GenerateStudentProfile(_ context.Context, _ services.GenerateStudentProfileOpts) (*services.StudentProfilePayload, error) {
	return &services.StudentProfilePayload{
		WritingFeatures: services.WritingFeatures{
			AvgSentenceLength: 18.5,
			LexicalDiversity:  0.72,
			ClauseComplexity:  "moderate",
		},
		VoiceMarkers: services.VoiceMarkers{
			FrequentPhrases:      []string{"in conclusion", "furthermore"},
			PreferredConnectives: []string{"however", "therefore"},
			ToneIndicators:       []string{"analytical", "formal"},
		},
		Provenance: services.ProfileProvenance{
			GeneratedAt: "2026-01-01T00:00:00Z",
		},
	}, nil
}

func (m *mockLLM) GenerateAuthorshipReport(_ context.Context, _ services.GenerateAuthorshipReportOpts) (*services.AuthorshipReportPayload, error) {
	return &services.AuthorshipReportPayload{
		OverallAssessment: services.OverallAssessment{
			Level:      "confident",
			Confidence: 0.85,
			Summary:    "Strong evidence of student authorship based on submission and writing baseline.",
		},
		EvidenceSignals: []services.EvidenceSignal{
			{Signal: "consistent style", Strength: "strong", Explanation: "Matches baseline profile"},
		},
		RiskFlags:            []services.RiskFlag{},
		RecommendedFollowups: []services.RecommendedFollowup{},
		Provenance: services.Provenance{
			ReportGeneratedAt: "2026-01-01T00:00:00Z",
		},
	}, nil
}

func TestAuthorshipGoldenPath(t *testing.T) {
	conn, queries := setupTestDB(t)
	defer teardownTestDB(t, conn)

	ctx := context.Background()
	teacherID := createTestTeacher(t, queries, ctx)

	// Create rubric
	rubric, err := queries.CreateRubric(ctx, db.CreateRubricParams{
		TeacherID: pgtype.UUID{Bytes: teacherID, Valid: true},
		Title:     "Authorship Test Rubric",
		RawText:   "Demonstrate understanding of the topic.",
	})
	if err != nil {
		t.Fatalf("Failed to create rubric: %v", err)
	}

	// Create student
	studentEmail := fmt.Sprintf("student-%s@example.com", uuid.New().String())
	student, err := queries.CreateStudent(ctx, db.CreateStudentParams{
		Email:       studentEmail,
		DisplayName: "Test Student",
	})
	if err != nil {
		t.Fatalf("Failed to create student: %v", err)
	}
	studentID := uuid.UUID(student.StudentID.Bytes)

	// Create submission
	submission, err := queries.CreateSubmission(ctx, db.CreateSubmissionParams{
		StudentID: student.StudentID,
		RubricID:  rubric.RubricID,
		Status:    "draft",
		Title:     pgtype.Text{String: "Test Submission", Valid: true},
	})
	if err != nil {
		t.Fatalf("Failed to create submission: %v", err)
	}
	submissionID := uuid.UUID(submission.SubmissionID.Bytes)

	// Create artifact with text payload
	artifactPayload, _ := json.Marshal(map[string]string{"text": "This is the student's essay demonstrating understanding of the topic."})
	_, err = queries.CreateSubmissionArtifact(ctx, db.CreateSubmissionArtifactParams{
		SubmissionID: submission.SubmissionID,
		ArtifactType: "text",
		Payload:      artifactPayload,
		OrderIndex:   0,
	})
	if err != nil {
		t.Fatalf("Failed to create artifact: %v", err)
	}

	llm := &mockLLM{}
	studentProfileHandler := NewStudentProfileHandler(queries, llm)
	submissionHandler := NewSubmissionHandler(queries, llm, nil)

	// Step 1: POST /students/{id}/profile/run
	t.Log("Step 1: Generating student profile...")
	{
		req := httptest.NewRequest("POST", "/students/"+studentID.String()+"/profile/run", nil)
		req = withChiParam(req, "id", studentID.String())
		w := httptest.NewRecorder()
		studentProfileHandler.RunStudentProfile(w, req)
		if w.Code != http.StatusCreated {
			t.Fatalf("Expected 201 from profile/run, got %d: %s", w.Code, w.Body.String())
		}
		var resp map[string]interface{}
		if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
			t.Fatalf("Failed to decode profile response: %v", err)
		}
		if resp["studentProfileId"] == nil {
			t.Fatal("Expected studentProfileId in profile response")
		}
		t.Logf("Profile created: %v", resp["studentProfileId"])
	}

	// Step 2: POST /submissions/{id}/authorship/run
	t.Log("Step 2: Generating authorship report...")
	var reportID interface{}
	{
		req := httptest.NewRequest("POST", "/submissions/"+submissionID.String()+"/authorship/run", nil)
		req = withChiParam(req, "id", submissionID.String())
		w := httptest.NewRecorder()
		submissionHandler.RunAuthorship(w, req)
		if w.Code != http.StatusCreated {
			t.Fatalf("Expected 201 from authorship/run, got %d: %s", w.Code, w.Body.String())
		}
		var resp map[string]interface{}
		if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
			t.Fatalf("Failed to decode authorship response: %v", err)
		}
		if resp["reportId"] == nil {
			t.Fatal("Expected reportId in authorship response")
		}
		reportID = resp["reportId"]

		report, ok := resp["report"].(map[string]interface{})
		if !ok {
			t.Fatal("Expected report object in response")
		}
		assessment, ok := report["overall_assessment"].(map[string]interface{})
		if !ok {
			t.Fatal("Expected overall_assessment in report")
		}
		level, _ := assessment["level"].(string)
		if level == "" {
			t.Fatal("Expected non-empty overall_assessment.level")
		}
		t.Logf("Report created (id=%v, level=%s)", reportID, level)
	}

	// Step 3: GET /submissions/{id}/authorship
	t.Log("Step 3: Retrieving authorship report...")
	{
		req := httptest.NewRequest("GET", "/submissions/"+submissionID.String()+"/authorship", nil)
		req = withChiParam(req, "id", submissionID.String())
		w := httptest.NewRecorder()
		submissionHandler.GetAuthorship(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("Expected 200 from GET authorship, got %d: %s", w.Code, w.Body.String())
		}
		var resp map[string]interface{}
		if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
			t.Fatalf("Failed to decode GET authorship response: %v", err)
		}
		if resp["report"] == nil {
			t.Fatal("Expected report in GET authorship response")
		}
		t.Log("Authorship report retrieved successfully")
	}
}

// withChiParam injects a chi URL parameter into the request context.
func withChiParam(r *http.Request, key, value string) *http.Request {
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add(key, value)
	return r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))
}
