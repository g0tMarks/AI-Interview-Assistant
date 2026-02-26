package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/g0tMarks/AI-Interview-Assistant/backend/internal/db"
	"github.com/g0tMarks/AI-Interview-Assistant/backend/internal/engine"
	"github.com/g0tMarks/AI-Interview-Assistant/backend/internal/evaluation"
	"github.com/g0tMarks/AI-Interview-Assistant/backend/internal/rubricparser"
	"github.com/g0tMarks/AI-Interview-Assistant/backend/internal/services"
	"github.com/g0tMarks/AI-Interview-Assistant/backend/internal/validation"
)

// setupTestDB initializes a database connection for testing
func setupTestDB(t *testing.T) (*pgx.Conn, *db.Queries) {
	requireDB := os.Getenv("REQUIRE_TEST_DB") == "1"

	dbURI := os.Getenv("TEST_DATABASE_URL")
	if dbURI == "" {
		dbURI = os.Getenv("DATABASE_URL")
	}
	if dbURI == "" {
		dbURI = "postgres://postgres:mysecretpassword@localhost:5432/test-db?sslmode=disable"
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, err := pgx.Connect(ctx, dbURI)
	if err != nil {
		if requireDB {
			t.Fatalf("Failed to connect to test database: %v", err)
		}
		t.Skipf("Skipping integration test (test DB unavailable): %v", err)
	}

	// Ping to verify connection
	if err := conn.Ping(ctx); err != nil {
		conn.Close(ctx)
		if requireDB {
			t.Fatalf("Failed to ping test database: %v", err)
		}
		t.Skipf("Skipping integration test (test DB unavailable): %v", err)
	}

	queries := db.New(conn)
	return conn, queries
}

// teardownTestDB closes the database connection
func teardownTestDB(t *testing.T, conn *pgx.Conn) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := conn.Close(ctx); err != nil {
		t.Logf("Warning: Failed to close database connection: %v", err)
	}
}

// createTestTeacher creates a test teacher and returns the teacher ID
func createTestTeacher(t *testing.T, queries *db.Queries, ctx context.Context) uuid.UUID {
	// Generate unique email to avoid conflicts
	email := fmt.Sprintf("test-teacher-%s@example.com", uuid.New().String())
	fullName := "Test Teacher"
	password := "TestPassword123!"

	// Hash password
	hashedPassword, err := validation.HashPassword(password)
	if err != nil {
		t.Fatalf("Failed to hash password: %v", err)
	}

	passwordHash := pgtype.Text{
		String: hashedPassword,
		Valid:  true,
	}

	teacher, err := queries.CreateTeacher(ctx, db.CreateTeacherParams{
		Email:        email,
		FullName:     fullName,
		PasswordHash: passwordHash,
	})
	if err != nil {
		t.Fatalf("Failed to create test teacher: %v", err)
	}

	if !teacher.TeacherID.Valid {
		t.Fatal("Teacher ID is not valid")
	}

	return teacher.TeacherID.Bytes
}

// cleanupTestData cleans up test data (messages, interview, template, rubric, teacher)
func cleanupTestData(t *testing.T, queries *db.Queries, ctx context.Context, teacherID uuid.UUID) {
	// Note: Due to CASCADE constraints, deleting the teacher will delete
	// rubrics, templates, interviews, and messages automatically.
	// But we'll clean up explicitly for clarity and to avoid issues if CASCADE is removed.

	// Get teacher's rubrics first
	teacherIDPgtype := pgtype.UUID{
		Bytes: teacherID,
		Valid: true,
	}

	rubrics, err := queries.ListRubricsByTeacher(ctx, teacherIDPgtype)
	if err != nil {
		t.Logf("Warning: Failed to list rubrics for cleanup: %v", err)
		return
	}

	// Delete rubrics (this will cascade to templates, interviews, messages)
	for _, rubric := range rubrics {
		if rubric.RubricID.Valid {
			err := queries.DisableRubric(ctx, rubric.RubricID)
			if err != nil {
				t.Logf("Warning: Failed to disable rubric %v: %v", rubric.RubricID.Bytes, err)
			}
		}
	}

	// Note: We don't delete the teacher itself as it might be reused in other tests
	// In a real scenario, you might want to delete it or use transactions
}

// TestCreateRubricTemplateInterviewFlow is the main integration test
func TestCreateRubricTemplateInterviewFlow(t *testing.T) {
	// Setup
	conn, queries := setupTestDB(t)
	defer teardownTestDB(t, conn)

	ctx := context.Background()
	teacherID := createTestTeacher(t, queries, ctx)
	defer cleanupTestData(t, queries, ctx, teacherID)

	// Initialize handlers
	llmService := services.NewOpenAIService()
	rubricHandler := NewRubricHandler(queries, nil, nil)
	templateHandler := NewInterviewTemplateHandler(queries, llmService)
	interviewEngine := engine.NewEngine(queries, llmService)
	evalRunner := evaluation.NewRunner(queries, llmService)
	interviewHandler := NewInterviewHandler(queries, interviewEngine, evalRunner)

	// Task 2: Create Rubric
	t.Log("Task 2: Creating rubric...")
	createRubricReq := CreateRubricRequest{
		TeacherID:   teacherID,
		Title:       "Test Rubric",
		Description: "Test rubric description",
		RawText:     "This is a test rubric content. Students should demonstrate understanding of key concepts.",
	}

	reqBody, err := json.Marshal(createRubricReq)
	if err != nil {
		t.Fatalf("Failed to marshal rubric request: %v", err)
	}

	req := httptest.NewRequest("POST", "/rubrics", bytes.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	rubricHandler.CreateRubric(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("Expected status 201, got %d. Response: %s", w.Code, w.Body.String())
	}

	var rubricResp RubricResponse
	if err := json.NewDecoder(w.Body).Decode(&rubricResp); err != nil {
		t.Fatalf("Failed to decode rubric response: %v", err)
	}

	if rubricResp.RubricID == uuid.Nil {
		t.Fatal("Rubric ID is nil")
	}
	if rubricResp.TeacherID != teacherID {
		t.Fatalf("Expected teacher ID %v, got %v", teacherID, rubricResp.TeacherID)
	}
	if rubricResp.Title != createRubricReq.Title {
		t.Fatalf("Expected title %s, got %s", createRubricReq.Title, rubricResp.Title)
	}
	if rubricResp.RawText != createRubricReq.RawText {
		t.Fatalf("Expected rawText %s, got %s", createRubricReq.RawText, rubricResp.RawText)
	}

	rubricID := rubricResp.RubricID
	t.Logf("Created rubric with ID: %v", rubricID)

	// Task 3: Create Interview Template
	t.Log("Task 3: Creating interview template...")
	createTemplateReq := CreateInterviewTemplateRequest{
		RubricID:            rubricID,
		Title:               "Test Interview Template",
		Instructions:        "", // Empty to trigger LLM generation (if available)
		Config:              json.RawMessage(`{}`),
		Status:              "draft",
		CurriculumSubject:   "Mathematics",
		CurriculumLevelBand: "7-8",
	}

	reqBody, err = json.Marshal(createTemplateReq)
	if err != nil {
		t.Fatalf("Failed to marshal template request: %v", err)
	}

	req = httptest.NewRequest("POST", "/interview-templates", bytes.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()

	templateHandler.CreateInterviewTemplate(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("Expected status 201, got %d. Response: %s", w.Code, w.Body.String())
	}

	var templateResp InterviewTemplateResponse
	if err := json.NewDecoder(w.Body).Decode(&templateResp); err != nil {
		t.Fatalf("Failed to decode template response: %v", err)
	}

	if templateResp.InterviewPlanID == uuid.Nil {
		t.Fatal("Interview Plan ID is nil")
	}
	if templateResp.RubricID != rubricID {
		t.Fatalf("Expected rubric ID %v, got %v", rubricID, templateResp.RubricID)
	}
	if templateResp.Title != createTemplateReq.Title {
		t.Fatalf("Expected title %s, got %s", createTemplateReq.Title, templateResp.Title)
	}

	interviewPlanID := templateResp.InterviewPlanID
	t.Logf("Created interview template with ID: %v", interviewPlanID)

	// Task 4: Create Interview
	t.Log("Task 4: Creating interview...")
	createInterviewReq := CreateInterviewRequest{
		InterviewPlanID: interviewPlanID,
		TeacherID:       teacherID,
		Simulated:       boolPtr(true),
		StudentName:     "Test Student",
		Status:          "in_progress",
	}

	reqBody, err = json.Marshal(createInterviewReq)
	if err != nil {
		t.Fatalf("Failed to marshal interview request: %v", err)
	}

	req = httptest.NewRequest("POST", "/interviews", bytes.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()

	interviewHandler.CreateInterview(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("Expected status 201, got %d. Response: %s", w.Code, w.Body.String())
	}

	var interviewResp InterviewResponse
	if err := json.NewDecoder(w.Body).Decode(&interviewResp); err != nil {
		t.Fatalf("Failed to decode interview response: %v", err)
	}

	if interviewResp.InterviewID == uuid.Nil {
		t.Fatal("Interview ID is nil")
	}
	if interviewResp.InterviewPlanID != interviewPlanID {
		t.Fatalf("Expected interview plan ID %v, got %v", interviewPlanID, interviewResp.InterviewPlanID)
	}
	if interviewResp.Status != "in_progress" {
		t.Fatalf("Expected status 'in_progress', got %s", interviewResp.Status)
	}
	if !interviewResp.Simulated {
		t.Fatal("Expected simulated to be true")
	}
	if interviewResp.StudentName == nil || *interviewResp.StudentName != createInterviewReq.StudentName {
		t.Fatalf("Expected student name %s, got %v", createInterviewReq.StudentName, interviewResp.StudentName)
	}
	if !interviewResp.StartedAt.Valid {
		t.Fatal("Expected startedAt to be set")
	}

	interviewID := interviewResp.InterviewID
	t.Logf("Created interview with ID: %v", interviewID)

	// Task 5: Add Messages to Interview
	t.Log("Task 5: Adding messages to interview...")
	interviewIDPgtype := pgtype.UUID{
		Bytes: interviewID,
		Valid: true,
	}

	// Create first message (AI)
	message1, err := queries.CreateInterviewMessage(ctx, db.CreateInterviewMessageParams{
		InterviewID:         interviewIDPgtype,
		Sender:              "ai",
		InterviewQuestionID: pgtype.UUID{Valid: false}, // null
		Content:             "Hello, how are you?",
	})
	if err != nil {
		t.Fatalf("Failed to create first message: %v", err)
	}

	if !message1.InterviewMessageID.Valid {
		t.Fatal("Message 1 ID is not valid")
	}
	if message1.Sender != "ai" {
		t.Fatalf("Expected sender 'ai', got %s", message1.Sender)
	}
	if message1.Content != "Hello, how are you?" {
		t.Fatalf("Expected content 'Hello, how are you?', got %s", message1.Content)
	}
	if !message1.CreatedAt.Valid {
		t.Fatal("Message 1 createdAt is not valid")
	}

	t.Logf("Created message 1 with ID: %v", message1.InterviewMessageID.Bytes)

	// Small delay to ensure different timestamps
	time.Sleep(10 * time.Millisecond)

	// Create second message (User)
	message2, err := queries.CreateInterviewMessage(ctx, db.CreateInterviewMessageParams{
		InterviewID:         interviewIDPgtype,
		Sender:              "user",
		InterviewQuestionID: pgtype.UUID{Valid: false}, // null
		Content:             "I'm doing well, thank you!",
	})
	if err != nil {
		t.Fatalf("Failed to create second message: %v", err)
	}

	if !message2.InterviewMessageID.Valid {
		t.Fatal("Message 2 ID is not valid")
	}
	if message2.Sender != "user" {
		t.Fatalf("Expected sender 'user', got %s", message2.Sender)
	}
	if message2.Content != "I'm doing well, thank you!" {
		t.Fatalf("Expected content 'I'm doing well, thank you!', got %s", message2.Content)
	}
	if !message2.CreatedAt.Valid {
		t.Fatal("Message 2 createdAt is not valid")
	}

	t.Logf("Created message 2 with ID: %v", message2.InterviewMessageID.Bytes)

	// Task 6: Retrieve Interview and Assert Structure
	t.Log("Task 6: Retrieving interview and asserting structure...")

	// Test GetInterview handler via HTTP with chi router context
	req = httptest.NewRequest("GET", fmt.Sprintf("/interviews/%s", interviewID.String()), nil)
	w = httptest.NewRecorder()

	// Set up chi router context for URL params
	r := chi.NewRouter()
	r.Get("/interviews/{id}", interviewHandler.GetInterview)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Expected status 200, got %d. Response: %s", w.Code, w.Body.String())
	}

	var retrievedInterviewResp InterviewResponse
	if err := json.NewDecoder(w.Body).Decode(&retrievedInterviewResp); err != nil {
		t.Fatalf("Failed to decode interview response: %v", err)
	}

	// Also retrieve directly from DB for additional verification
	retrievedInterview, err := queries.GetInterviewByID(ctx, interviewIDPgtype)
	if err != nil {
		t.Fatalf("Failed to retrieve interview from DB: %v", err)
	}

	// Assert interview structure from HTTP response
	if retrievedInterviewResp.InterviewID != interviewID {
		t.Fatalf("Expected interview ID %v, got %v", interviewID, retrievedInterviewResp.InterviewID)
	}
	if retrievedInterviewResp.InterviewPlanID != interviewPlanID {
		t.Fatalf("Expected interview plan ID %v, got %v", interviewPlanID, retrievedInterviewResp.InterviewPlanID)
	}
	if retrievedInterviewResp.Status != "in_progress" {
		t.Fatalf("Expected status 'in_progress', got %s", retrievedInterviewResp.Status)
	}
	if !retrievedInterviewResp.Simulated {
		t.Fatal("Expected simulated to be true")
	}
	if retrievedInterviewResp.StudentName == nil || *retrievedInterviewResp.StudentName != createInterviewReq.StudentName {
		t.Fatalf("Expected student name %s, got %v", createInterviewReq.StudentName, retrievedInterviewResp.StudentName)
	}
	if !retrievedInterviewResp.StartedAt.Valid {
		t.Fatal("Expected startedAt to be set")
	}
	if retrievedInterviewResp.CompletedAt != nil {
		t.Fatal("Expected completedAt to be null")
	}

	// Also verify DB structure matches
	if !retrievedInterview.InterviewID.Valid || retrievedInterview.InterviewID.Bytes != interviewID {
		t.Fatalf("DB: Expected interview ID %v, got %v", interviewID, retrievedInterview.InterviewID.Bytes)
	}
	if !retrievedInterview.InterviewPlanID.Valid || retrievedInterview.InterviewPlanID.Bytes != interviewPlanID {
		t.Fatalf("DB: Expected interview plan ID %v, got %v", interviewPlanID, retrievedInterview.InterviewPlanID.Bytes)
	}
	if retrievedInterview.Status != "in_progress" {
		t.Fatalf("DB: Expected status 'in_progress', got %s", retrievedInterview.Status)
	}
	if !retrievedInterview.Simulated {
		t.Fatal("DB: Expected simulated to be true")
	}
	if !retrievedInterview.StartedAt.Valid {
		t.Fatal("DB: Expected startedAt to be set")
	}
	if retrievedInterview.CompletedAt.Valid {
		t.Fatal("DB: Expected completedAt to be null")
	}

	// Retrieve messages
	messages, err := queries.ListMessagesByInterview(ctx, interviewIDPgtype)
	if err != nil {
		t.Fatalf("Failed to retrieve messages: %v", err)
	}

	// Assert message structure
	if len(messages) != 2 {
		t.Fatalf("Expected 2 messages, got %d", len(messages))
	}

	// Messages should be ordered by createdAt ASC
	if messages[0].Sender != "ai" || messages[0].Content != "Hello, how are you?" {
		t.Fatalf("Expected first message to be from 'ai' with content 'Hello, how are you?', got sender '%s' with content '%s'",
			messages[0].Sender, messages[0].Content)
	}
	if messages[1].Sender != "user" || messages[1].Content != "I'm doing well, thank you!" {
		t.Fatalf("Expected second message to be from 'user' with content 'I'm doing well, thank you!', got sender '%s' with content '%s'",
			messages[1].Sender, messages[1].Content)
	}

	// Verify message IDs match
	if messages[0].InterviewMessageID.Bytes != message1.InterviewMessageID.Bytes {
		t.Fatalf("Expected first message ID %v, got %v", message1.InterviewMessageID.Bytes, messages[0].InterviewMessageID.Bytes)
	}
	if messages[1].InterviewMessageID.Bytes != message2.InterviewMessageID.Bytes {
		t.Fatalf("Expected second message ID %v, got %v", message2.InterviewMessageID.Bytes, messages[1].InterviewMessageID.Bytes)
	}

	// Verify timestamps are in ascending order
	if messages[0].CreatedAt.Time.After(messages[1].CreatedAt.Time) {
		t.Fatal("Messages are not ordered by createdAt ASC")
	}

	// Assert overall structure
	t.Log("All assertions passed!")
	t.Logf("Interview ID: %v", interviewID)
	t.Logf("Interview Plan ID: %v", interviewPlanID)
	t.Logf("Rubric ID: %v", rubricID)
	t.Logf("Teacher ID: %v", teacherID)
	t.Logf("Number of messages: %d", len(messages))
}

// mockLLMForGoldenPath implements LLMService for the golden-path test (no API key required).
type mockLLMForGoldenPath struct{}

func (m *mockLLMForGoldenPath) GenerateInterviewInstructions(ctx context.Context, rubricTitle, rubricRawText string) (string, error) {
	return "Mock instructions", nil
}
func (m *mockLLMForGoldenPath) ParseRubric(ctx context.Context, rubricTitle, rawText string) (*rubricparser.ParseRubricOutput, error) {
	return nil, fmt.Errorf("not implemented")
}
func (m *mockLLMForGoldenPath) ClassifyResponse(ctx context.Context, questionPrompt, userResponse string) (string, error) {
	return services.ResponseCategoryStrong, nil
}
func (m *mockLLMForGoldenPath) EvaluateInterview(ctx context.Context, rubricTitle string, criteria []evaluation.CriterionForEval, transcript string) (*evaluation.EvalOutput, error) {
	conf := 0.9
	return &evaluation.EvalOutput{
		OverallSummary:     "Student demonstrated understanding.",
		Strengths:           "Clear communication.",
		AreasForGrowth:      "Could add more detail.",
		SuggestedNextSteps:  "Practice extended responses.",
		Criteria: []evaluation.EvalCriterion{
			{CriterionName: "Understanding", Level: "B", EvidenceText: "Student answered correctly.", ModelConfidence: &conf},
		},
	}, nil
}

func (m *mockLLMForGoldenPath) GenerateAuthorshipReport(ctx context.Context, opts services.GenerateAuthorshipReportOpts) (*services.AuthorshipReportPayload, error) {
	return &services.AuthorshipReportPayload{
		OverallAssessment: services.OverallAssessment{
			Level:      "confident",
			Confidence: 0.85,
			Summary:    "Submission and viva are consistent with student authorship.",
		},
		EvidenceSignals: []services.EvidenceSignal{
			{Signal: "Consistency", Strength: "strong", Explanation: "Viva answers aligned with submission.", SupportingQuotesOrRefs: nil},
		},
		RiskFlags:         nil,
		RecommendedFollowups: []services.RecommendedFollowup{
			{Question: "Can you expand on section X?", Why: "To verify depth of understanding."},
		},
		RubricAlignment: map[string]string{"Understanding": "Addressed in viva."},
		Provenance: services.Provenance{
			SubmissionArtifactIDs: opts.ArtifactIDs,
			InterviewID:           opts.InterviewID,
			ReportGeneratedAt:     "2025-01-01T00:00:00Z",
		},
	}, nil
}

// TestGoldenPathFullFlow runs: teacher → rubric → template → add criterion + question + branch → interview → POST /next until done → GET results.
func TestGoldenPathFullFlow(t *testing.T) {
	conn, queries := setupTestDB(t)
	defer teardownTestDB(t, conn)

	ctx := context.Background()
	teacherID := createTestTeacher(t, queries, ctx)
	defer cleanupTestData(t, queries, ctx, teacherID)

	mockLLM := &mockLLMForGoldenPath{}
	rubricHandler := NewRubricHandler(queries, nil, nil)
	templateHandler := NewInterviewTemplateHandler(queries, mockLLM)
	interviewEngine := engine.NewEngine(queries, mockLLM)
	evalRunner := evaluation.NewRunner(queries, mockLLM)
	interviewHandler := NewInterviewHandler(queries, interviewEngine, evalRunner)

	// Create rubric
	createRubricReq := CreateRubricRequest{
		TeacherID:   teacherID,
		Title:       "Golden Path Rubric",
		Description: "For golden path test",
		RawText:     "Understands key concepts. Explains clearly.",
	}
	reqBody, _ := json.Marshal(createRubricReq)
	req := httptest.NewRequest("POST", "/rubrics", bytes.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	rubricHandler.CreateRubric(w, req)
	if w.Code != http.StatusCreated {
		t.Fatalf("create rubric: expected 201, got %d: %s", w.Code, w.Body.String())
	}
	var rubricResp RubricResponse
	if err := json.NewDecoder(w.Body).Decode(&rubricResp); err != nil {
		t.Fatalf("decode rubric: %v", err)
	}
	rubricID := rubricResp.RubricID

	// Create template (plan)
	createTemplateReq := CreateInterviewTemplateRequest{
		RubricID:            rubricID,
		Title:               "Golden Path Plan",
		Instructions:        "Ask one question.",
		Config:              json.RawMessage(`{}`),
		Status:              "in_progress",
		CurriculumSubject:   "",
		CurriculumLevelBand: "",
	}
	reqBody, _ = json.Marshal(createTemplateReq)
	req = httptest.NewRequest("POST", "/interview-templates", bytes.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	templateHandler.CreateInterviewTemplate(w, req)
	if w.Code != http.StatusCreated {
		t.Fatalf("create template: expected 201, got %d: %s", w.Code, w.Body.String())
	}
	var templateResp InterviewTemplateResponse
	if err := json.NewDecoder(w.Body).Decode(&templateResp); err != nil {
		t.Fatalf("decode template: %v", err)
	}
	interviewPlanID := templateResp.InterviewPlanID

	// Add one rubric criterion and one interview question + branch so /next returns one question then "done"
	rubricIDPg := pgtype.UUID{Bytes: rubricID, Valid: true}
	planIDPg := pgtype.UUID{Bytes: interviewPlanID, Valid: true}
	criterion, err := queries.CreateRubricCriterion(ctx, db.CreateRubricCriterionParams{
		RubricID:    rubricIDPg,
		Name:        "Understanding",
		Description: pgtype.Text{String: "Demonstrates understanding", Valid: true},
		Weight:      pgtype.Numeric{Int: big.NewInt(100), Exp: -2, Valid: true},
		OrderIndex:  0,
		Levels:      nil,
	})
	if err != nil {
		t.Fatalf("create criterion: %v", err)
	}
	criterionIDPg := criterion.RubricCriterionID
	question, err := queries.CreateInterviewQuestion(ctx, db.CreateInterviewQuestionParams{
		InterviewPlanID:   planIDPg,
		RubricCriterionID: criterionIDPg,
		Prompt:             "What is the main idea?",
		QuestionType:      "open",
		OrderIndex:         0,
		IsActive:           true,
		FollowUpToID:       pgtype.UUID{},
		FollowUpCondition:  pgtype.Text{},
	})
	if err != nil {
		t.Fatalf("create question: %v", err)
	}
	questionIDPg := question.InterviewQuestionID
	_, err = queries.CreateInterviewQuestionBranch(ctx, db.CreateInterviewQuestionBranchParams{
		ParentQuestionID:       questionIDPg,
		ResponseCategory:       "strong",
		MisconceptionTagID:     pgtype.UUID{},
		NextQuestionID:         pgtype.UUID{},
		FollowUpPromptOverride: pgtype.Text{},
		TerminateInterview:     true,
		OrderIndex:             0,
	})
	if err != nil {
		t.Fatalf("create branch: %v", err)
	}

	// Create interview
	createInterviewReq := CreateInterviewRequest{
		InterviewPlanID: interviewPlanID,
		TeacherID:       teacherID,
		Simulated:       boolPtr(true),
		StudentName:     "Student",
		Status:          "in_progress",
	}
	reqBody, _ = json.Marshal(createInterviewReq)
	req = httptest.NewRequest("POST", "/interviews", bytes.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	interviewHandler.CreateInterview(w, req)
	if w.Code != http.StatusCreated {
		t.Fatalf("create interview: expected 201, got %d: %s", w.Code, w.Body.String())
	}
	var interviewResp InterviewResponse
	if err := json.NewDecoder(w.Body).Decode(&interviewResp); err != nil {
		t.Fatalf("decode interview: %v", err)
	}
	interviewID := interviewResp.InterviewID

	// Route for interview sub-resources (so chi sets URL param "id")
	r := chi.NewRouter()
	r.Post("/interviews", interviewHandler.CreateInterview)
	r.Get("/interviews/{id}", interviewHandler.GetInterview)
	r.Post("/interviews/{id}/messages", interviewHandler.CreateMessage)
	r.Get("/interviews/{id}/messages", interviewHandler.ListMessages)
	r.Get("/interviews/{id}/next", interviewHandler.GetNext)
	r.Post("/interviews/{id}/next", interviewHandler.PostNext)
	r.Get("/interviews/{id}/results", interviewHandler.GetResults)

	// POST /next — should return next_question (first question)
	req = httptest.NewRequest("POST", "/interviews/"+interviewID.String()+"/next", nil)
	req = req.WithContext(ctx)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("first POST /next: expected 200, got %d: %s", w.Code, w.Body.String())
	}
	var nextResp NextResponse
	if err := json.NewDecoder(w.Body).Decode(&nextResp); err != nil {
		t.Fatalf("decode next: %v", err)
	}
	if nextResp.Status != "next_question" {
		t.Fatalf("first /next: expected status next_question, got %s", nextResp.Status)
	}

	// POST user message
	msgBody, _ := json.Marshal(CreateInterviewMessageRequest{Sender: "user", Content: "The main idea is that we need to understand key concepts."})
	req = httptest.NewRequest("POST", "/interviews/"+interviewID.String()+"/messages", bytes.NewReader(msgBody))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(ctx)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusCreated {
		t.Fatalf("POST message: expected 201, got %d: %s", w.Code, w.Body.String())
	}

	// POST /next — should return done and mark interview completed
	req = httptest.NewRequest("POST", "/interviews/"+interviewID.String()+"/next", nil)
	req = req.WithContext(ctx)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("second POST /next: expected 200, got %d: %s", w.Code, w.Body.String())
	}
	if err := json.NewDecoder(w.Body).Decode(&nextResp); err != nil {
		t.Fatalf("decode next: %v", err)
	}
	if nextResp.Status != "done" {
		t.Fatalf("second /next: expected status done, got %s", nextResp.Status)
	}

	// GET /interviews/{id} — assert status is completed
	req = httptest.NewRequest("GET", "/interviews/"+interviewID.String(), nil)
	req = req.WithContext(ctx)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("GET interview: expected 200, got %d: %s", w.Code, w.Body.String())
	}
	if err := json.NewDecoder(w.Body).Decode(&interviewResp); err != nil {
		t.Fatalf("decode interview: %v", err)
	}
	if interviewResp.Status != "completed" {
		t.Fatalf("interview status: expected completed, got %s", interviewResp.Status)
	}

	// GET /interviews/{id}/results — triggers evaluation, returns summary + scoring
	req = httptest.NewRequest("GET", "/interviews/"+interviewID.String()+"/results", nil)
	req = req.WithContext(ctx)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("GET results: expected 200, got %d: %s", w.Code, w.Body.String())
	}
	var resultsResp ResultsResponse
	if err := json.NewDecoder(w.Body).Decode(&resultsResp); err != nil {
		t.Fatalf("decode results: %v", err)
	}
	if resultsResp.InterviewID != interviewID {
		t.Fatalf("results interviewId: expected %v, got %v", interviewID, resultsResp.InterviewID)
	}
	if resultsResp.Status != "completed" {
		t.Fatalf("results status: expected completed, got %s", resultsResp.Status)
	}
	if resultsResp.OverallSummary == "" {
		t.Fatal("results: expected non-empty overallSummary")
	}
	if len(resultsResp.CriterionEvidence) == 0 {
		t.Fatal("results: expected at least one criterion evidence")
	}
	if resultsResp.Scoring == nil {
		t.Fatal("results: expected non-nil scoring")
	}
	if resultsResp.Scoring.OverallSummary == "" {
		t.Fatal("results.scoring: expected non-empty overallSummary")
	}
	if len(resultsResp.Scoring.Scores) == 0 {
		t.Fatal("results.scoring: expected non-empty scores map")
	}
	t.Log("Golden path test passed: full flow through /next and results.")
}

// TestAuthorshipFlow runs: create rubric → template → student → submission → artifacts → start viva → add messages → run authorship → get report.
func TestAuthorshipFlow(t *testing.T) {
	conn, queries := setupTestDB(t)
	defer teardownTestDB(t, conn)

	ctx := context.Background()
	teacherID := createTestTeacher(t, queries, ctx)
	defer cleanupTestData(t, queries, ctx, teacherID)

	mockLLM := &mockLLMForGoldenPath{}
	rubricHandler := NewRubricHandler(queries, nil, nil)
	templateHandler := NewInterviewTemplateHandler(queries, mockLLM)
	interviewEngine := engine.NewEngine(queries, mockLLM)
	evalRunner := evaluation.NewRunner(queries, mockLLM)
	interviewHandler := NewInterviewHandler(queries, interviewEngine, evalRunner)
	submissionHandler := NewSubmissionHandler(queries, mockLLM, interviewHandler)

	// Create rubric
	createRubricReq := CreateRubricRequest{
		TeacherID:   teacherID,
		Title:       "Authorship Test Rubric",
		Description: "For authorship flow",
		RawText:     "Student demonstrates original work.",
	}
	reqBody, _ := json.Marshal(createRubricReq)
	req := httptest.NewRequest("POST", "/rubrics", bytes.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	rubricHandler.CreateRubric(w, req)
	if w.Code != http.StatusCreated {
		t.Fatalf("create rubric: expected 201, got %d: %s", w.Code, w.Body.String())
	}
	var rubricResp RubricResponse
	if err := json.NewDecoder(w.Body).Decode(&rubricResp); err != nil {
		t.Fatalf("decode rubric: %v", err)
	}
	rubricID := rubricResp.RubricID

	// Create template so submission can start viva
	createTemplateReq := CreateInterviewTemplateRequest{
		RubricID:            rubricID,
		Title:               "Authorship Viva Plan",
		Instructions:        "Ask about the submission.",
		Config:              json.RawMessage(`{}`),
		Status:              "in_progress",
		CurriculumSubject:   "",
		CurriculumLevelBand: "",
	}
	reqBody, _ = json.Marshal(createTemplateReq)
	req = httptest.NewRequest("POST", "/interview-templates", bytes.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	templateHandler.CreateInterviewTemplate(w, req)
	if w.Code != http.StatusCreated {
		t.Fatalf("create template: expected 201, got %d: %s", w.Code, w.Body.String())
	}

	// Create student via DB
	student, err := queries.CreateStudent(ctx, db.CreateStudentParams{
		Email:        "authorship-student@test.com",
		DisplayName:  "Authorship Student",
	})
	if err != nil {
		t.Fatalf("create student: %v", err)
	}
	studentID := student.StudentID.Bytes

	// Create submission
	createSubReq := CreateSubmissionRequest{
		StudentID: studentID,
		RubricID:  rubricID,
		Title:     "My submission",
	}
	reqBody, _ = json.Marshal(createSubReq)
	req = httptest.NewRequest("POST", "/submissions", bytes.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	submissionHandler.CreateSubmission(w, req)
	if w.Code != http.StatusCreated {
		t.Fatalf("create submission: expected 201, got %d: %s", w.Code, w.Body.String())
	}
	var subResp SubmissionResponse
	if err := json.NewDecoder(w.Body).Decode(&subResp); err != nil {
		t.Fatalf("decode submission: %v", err)
	}
	submissionID := subResp.SubmissionID

	// Add artifact
	artifactBody := []byte(`{"artifactType":"main_text","payload":{"text":"This is my essay content for the task."}}`)
	req = httptest.NewRequest("POST", "/submissions/"+submissionID.String()+"/artifacts", bytes.NewReader(artifactBody))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	submissionHandler.CreateArtifact(w, req)
	if w.Code != http.StatusCreated {
		t.Fatalf("create artifact: expected 201, got %d: %s", w.Code, w.Body.String())
	}

	// Start viva
	req = httptest.NewRequest("POST", "/submissions/"+submissionID.String()+"/viva/start", nil)
	req = req.WithContext(ctx)
	w = httptest.NewRecorder()
	submissionHandler.StartViva(w, req)
	if w.Code != http.StatusCreated && w.Code != http.StatusOK {
		t.Fatalf("start viva: expected 201 or 200, got %d: %s", w.Code, w.Body.String())
	}

	// Add viva messages
	msgBody := []byte(`{"sender":"ai","content":"Can you explain your main argument?"}`)
	req = httptest.NewRequest("POST", "/submissions/"+submissionID.String()+"/viva/messages", bytes.NewReader(msgBody))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(ctx)
	w = httptest.NewRecorder()
	submissionHandler.VivaMessages(w, req)
	if w.Code != http.StatusCreated {
		t.Fatalf("viva message 1: expected 201, got %d: %s", w.Code, w.Body.String())
	}
	msgBody = []byte(`{"sender":"user","content":"I argued that the evidence supports my conclusion."}`)
	req = httptest.NewRequest("POST", "/submissions/"+submissionID.String()+"/viva/messages", bytes.NewReader(msgBody))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(ctx)
	w = httptest.NewRecorder()
	submissionHandler.VivaMessages(w, req)
	if w.Code != http.StatusCreated {
		t.Fatalf("viva message 2: expected 201, got %d: %s", w.Code, w.Body.String())
	}

	// Run authorship report
	req = httptest.NewRequest("POST", "/submissions/"+submissionID.String()+"/authorship/run", nil)
	req = req.WithContext(ctx)
	w = httptest.NewRecorder()
	submissionHandler.RunAuthorship(w, req)
	if w.Code != http.StatusCreated {
		t.Fatalf("run authorship: expected 201, got %d: %s", w.Code, w.Body.String())
	}

	// Get authorship report
	req = httptest.NewRequest("GET", "/submissions/"+submissionID.String()+"/authorship", nil)
	req = req.WithContext(ctx)
	w = httptest.NewRecorder()
	submissionHandler.GetAuthorship(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("get authorship: expected 200, got %d: %s", w.Code, w.Body.String())
	}
	var reportResp struct {
		ReportID     uuid.UUID                      `json:"reportId"`
		SubmissionID uuid.UUID                      `json:"submissionId"`
		Report       services.AuthorshipReportPayload `json:"report"`
		CreatedAt    string                         `json:"createdAt"`
	}
	if err := json.NewDecoder(w.Body).Decode(&reportResp); err != nil {
		t.Fatalf("decode authorship report: %v", err)
	}
	if reportResp.Report.OverallAssessment.Summary == "" {
		t.Fatal("expected non-empty report overall summary")
	}
	if reportResp.Report.OverallAssessment.Level != "confident" {
		t.Fatalf("expected level confident, got %s", reportResp.Report.OverallAssessment.Level)
	}
	t.Log("Authorship flow test passed.")
}

// Helper function to create a bool pointer
func boolPtr(b bool) *bool {
	return &b
}
