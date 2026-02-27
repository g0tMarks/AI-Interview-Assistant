package engine

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/g0tMarks/AI-Interview-Assistant/backend/internal/db"
	"github.com/g0tMarks/AI-Interview-Assistant/backend/internal/evaluation"
	"github.com/g0tMarks/AI-Interview-Assistant/backend/internal/rubricparser"
	"github.com/g0tMarks/AI-Interview-Assistant/backend/internal/services"
)

// fakeQueries returns fixed data for engine tests.
type fakeQueries struct {
	interview *db.AppInterview
	questions []db.AppInterviewQuestion
	messages  []db.AppInterviewMessage
	branch    *db.AppInterviewQuestionBranch
	getErr    error // e.g. pgx.ErrNoRows for "not found"
}

func (f *fakeQueries) GetInterviewByID(_ context.Context, _ pgtype.UUID) (db.AppInterview, error) {
	if f.getErr != nil {
		return db.AppInterview{}, f.getErr
	}
	if f.interview == nil {
		return db.AppInterview{}, pgx.ErrNoRows
	}
	return *f.interview, nil
}

func (f *fakeQueries) ListQuestionsByPlan(_ context.Context, _ pgtype.UUID) ([]db.AppInterviewQuestion, error) {
	return f.questions, nil
}

func (f *fakeQueries) ListMessagesByInterview(_ context.Context, _ pgtype.UUID) ([]db.AppInterviewMessage, error) {
	return f.messages, nil
}

func (f *fakeQueries) GetBranchByCategory(_ context.Context, _ db.GetBranchByCategoryParams) (db.AppInterviewQuestionBranch, error) {
	if f.branch == nil {
		return db.AppInterviewQuestionBranch{}, pgx.ErrNoRows
	}
	return *f.branch, nil
}

// mockLLM returns a fixed category for ClassifyResponse; other methods are no-ops.
type mockLLM struct {
	category string
	err     error
}

func (m *mockLLM) GenerateInterviewInstructions(context.Context, string, string) (string, error) {
	return "", nil
}

func (m *mockLLM) ParseRubric(context.Context, string, string) (*rubricparser.ParseRubricOutput, error) {
	return nil, nil
}

func (m *mockLLM) ClassifyResponse(_ context.Context, _, _ string) (string, error) {
	if m.err != nil {
		return "", m.err
	}
	if m.category != "" {
		return m.category, nil
	}
	return services.ResponseCategoryPartial, nil
}

func (m *mockLLM) EvaluateInterview(_ context.Context, _ string, _ []evaluation.CriterionForEval, _ string) (*evaluation.EvalOutput, error) {
	return nil, nil
}

func (m *mockLLM) GenerateAuthorshipReport(context.Context, services.GenerateAuthorshipReportOpts) (*services.AuthorshipReportPayload, error) {
	return nil, nil
}

func (m *mockLLM) GenerateStudentProfile(context.Context, services.GenerateStudentProfileOpts) (*services.StudentProfilePayload, error) {
	return nil, nil
}

var _ services.LLMService = (*mockLLM)(nil)

// pgUUID returns a valid pgtype.UUID from a uuid.UUID.
func pgUUID(u uuid.UUID) pgtype.UUID {
	return pgtype.UUID{Bytes: u, Valid: true}
}

func TestComputeNext_InterviewNotFound(t *testing.T) {
	fq := &fakeQueries{interview: nil}
	eng := NewEngine(fq, &mockLLM{})
	ctx := context.Background()

	_, err := eng.ComputeNext(ctx, uuid.MustParse("00000000-0000-0000-0000-000000000001"))
	if err == nil {
		t.Fatal("expected error when interview not found")
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		t.Errorf("expected pgx.ErrNoRows, got %v", err)
	}
}

func TestComputeNext_InterviewCompleted(t *testing.T) {
	planID := uuid.New()
	inv := &db.AppInterview{
		InterviewID:     pgUUID(uuid.New()),
		InterviewPlanID: pgUUID(planID),
		Status:          "completed",
	}
	fq := &fakeQueries{interview: inv}
	eng := NewEngine(fq, &mockLLM{})
	ctx := context.Background()

	result, err := eng.ComputeNext(ctx, inv.InterviewID.Bytes)
	if err != nil {
		t.Fatal(err)
	}
	if result.Status != NextStatusDone {
		t.Errorf("expected status done, got %s", result.Status)
	}
}

func TestComputeNext_NoMessages_ReturnsFirstQuestion(t *testing.T) {
	planID := uuid.New()
	q1ID := uuid.New()
	q2ID := uuid.New()
	inv := &db.AppInterview{
		InterviewID:     pgUUID(uuid.New()),
		InterviewPlanID: pgUUID(planID),
		Status:          "in_progress",
	}
	questions := []db.AppInterviewQuestion{
		{InterviewQuestionID: pgUUID(q2ID), OrderIndex: 1, IsActive: true, Prompt: "Second"},
		{InterviewQuestionID: pgUUID(q1ID), OrderIndex: 0, IsActive: true, Prompt: "First"},
	}
	fq := &fakeQueries{interview: inv, questions: questions}
	eng := NewEngine(fq, &mockLLM{})
	ctx := context.Background()

	result, err := eng.ComputeNext(ctx, inv.InterviewID.Bytes)
	if err != nil {
		t.Fatal(err)
	}
	if result.Status != NextStatusNextQuestion {
		t.Errorf("expected next_question, got %s", result.Status)
	}
	if result.NextQuestionID == nil || *result.NextQuestionID != q1ID {
		t.Errorf("expected first question id %v, got %v", q1ID, result.NextQuestionID)
	}
	if result.Prompt != "First" {
		t.Errorf("expected prompt First, got %q", result.Prompt)
	}
}

func TestComputeNext_NoQuestions_ReturnsDone(t *testing.T) {
	planID := uuid.New()
	inv := &db.AppInterview{
		InterviewID:     pgUUID(uuid.New()),
		InterviewPlanID: pgUUID(planID),
		Status:          "in_progress",
	}
	fq := &fakeQueries{interview: inv, questions: nil}
	eng := NewEngine(fq, &mockLLM{})
	ctx := context.Background()

	result, err := eng.ComputeNext(ctx, inv.InterviewID.Bytes)
	if err != nil {
		t.Fatal(err)
	}
	if result.Status != NextStatusDone {
		t.Errorf("expected done, got %s", result.Status)
	}
}

func TestComputeNext_LastMessageAI_ReturnsWaitingForUser(t *testing.T) {
	planID := uuid.New()
	qID := uuid.New()
	inv := &db.AppInterview{
		InterviewID:     pgUUID(uuid.New()),
		InterviewPlanID: pgUUID(planID),
		Status:          "in_progress",
	}
	questions := []db.AppInterviewQuestion{
		{InterviewQuestionID: pgUUID(qID), OrderIndex: 0, IsActive: true, Prompt: "Q1"},
	}
	messages := []db.AppInterviewMessage{
		{Sender: "ai", InterviewQuestionID: pgUUID(qID), Content: "Q1"},
	}
	fq := &fakeQueries{interview: inv, questions: questions, messages: messages}
	eng := NewEngine(fq, &mockLLM{})
	ctx := context.Background()

	result, err := eng.ComputeNext(ctx, inv.InterviewID.Bytes)
	if err != nil {
		t.Fatal(err)
	}
	if result.Status != NextStatusWaitingForUser {
		t.Errorf("expected waiting_for_user, got %s", result.Status)
	}
	if result.WaitingForQuestionID == nil || *result.WaitingForQuestionID != qID {
		t.Errorf("expected waiting for question %v, got %v", qID, result.WaitingForQuestionID)
	}
}

func TestComputeNext_LastMessageUser_NoBranch_LinearFallback(t *testing.T) {
	planID := uuid.New()
	q1ID := uuid.New()
	q2ID := uuid.New()
	inv := &db.AppInterview{
		InterviewID:     pgUUID(uuid.New()),
		InterviewPlanID: pgUUID(planID),
		Status:          "in_progress",
	}
	questions := []db.AppInterviewQuestion{
		{InterviewQuestionID: pgUUID(q1ID), OrderIndex: 0, IsActive: true, Prompt: "Q1"},
		{InterviewQuestionID: pgUUID(q2ID), OrderIndex: 1, IsActive: true, Prompt: "Q2"},
	}
	messages := []db.AppInterviewMessage{
		{Sender: "ai", InterviewQuestionID: pgUUID(q1ID), Content: "Q1"},
		{Sender: "user", InterviewQuestionID: pgUUID(q1ID), Content: "My answer"},
	}
	fq := &fakeQueries{interview: inv, questions: questions, messages: messages, branch: nil}
	eng := NewEngine(fq, &mockLLM{category: services.ResponseCategoryPartial})
	ctx := context.Background()

	result, err := eng.ComputeNext(ctx, inv.InterviewID.Bytes)
	if err != nil {
		t.Fatal(err)
	}
	if result.Status != NextStatusNextQuestion {
		t.Errorf("expected next_question, got %s", result.Status)
	}
	if result.NextQuestionID == nil || *result.NextQuestionID != q2ID {
		t.Errorf("expected next question %v, got %v", q2ID, result.NextQuestionID)
	}
	if result.ClassifiedCategory != services.ResponseCategoryPartial {
		t.Errorf("expected classified partial, got %q", result.ClassifiedCategory)
	}
}

func TestComputeNext_LastMessageUser_BranchTerminate_ReturnsDone(t *testing.T) {
	planID := uuid.New()
	q1ID := uuid.New()
	inv := &db.AppInterview{
		InterviewID:     pgUUID(uuid.New()),
		InterviewPlanID: pgUUID(planID),
		Status:          "in_progress",
	}
	questions := []db.AppInterviewQuestion{
		{InterviewQuestionID: pgUUID(q1ID), OrderIndex: 0, IsActive: true, Prompt: "Q1"},
	}
	messages := []db.AppInterviewMessage{
		{Sender: "ai", InterviewQuestionID: pgUUID(q1ID), Content: "Q1"},
		{Sender: "user", InterviewQuestionID: pgUUID(q1ID), Content: "I don't know"},
	}
	branch := &db.AppInterviewQuestionBranch{
		ParentQuestionID:   pgUUID(q1ID),
		ResponseCategory:   services.ResponseCategoryDontKnow,
		TerminateInterview: true,
	}
	fq := &fakeQueries{interview: inv, questions: questions, messages: messages, branch: branch}
	eng := NewEngine(fq, &mockLLM{category: services.ResponseCategoryDontKnow})
	ctx := context.Background()

	result, err := eng.ComputeNext(ctx, inv.InterviewID.Bytes)
	if err != nil {
		t.Fatal(err)
	}
	if result.Status != NextStatusDone {
		t.Errorf("expected done, got %s", result.Status)
	}
	if result.ClassifiedCategory != services.ResponseCategoryDontKnow {
		t.Errorf("expected classified dont_know, got %q", result.ClassifiedCategory)
	}
}

func TestComputeNext_LastMessageUser_BranchHasNextQuestion_ReturnsThatQuestion(t *testing.T) {
	planID := uuid.New()
	q1ID := uuid.New()
	q2ID := uuid.New()
	inv := &db.AppInterview{
		InterviewID:     pgUUID(uuid.New()),
		InterviewPlanID: pgUUID(planID),
		Status:          "in_progress",
	}
	questions := []db.AppInterviewQuestion{
		{InterviewQuestionID: pgUUID(q1ID), OrderIndex: 0, IsActive: true, Prompt: "Q1"},
		{InterviewQuestionID: pgUUID(q2ID), OrderIndex: 1, IsActive: true, Prompt: "Q2 follow-up"},
	}
	messages := []db.AppInterviewMessage{
		{Sender: "ai", InterviewQuestionID: pgUUID(q1ID), Content: "Q1"},
		{Sender: "user", InterviewQuestionID: pgUUID(q1ID), Content: "Partial answer"},
	}
	branch := &db.AppInterviewQuestionBranch{
		ParentQuestionID:       pgUUID(q1ID),
		ResponseCategory:       services.ResponseCategoryPartial,
		NextQuestionID:         pgUUID(q2ID),
		FollowUpPromptOverride: pgtype.Text{String: "Custom follow-up", Valid: true},
		TerminateInterview:     false,
	}
	fq := &fakeQueries{interview: inv, questions: questions, messages: messages, branch: branch}
	eng := NewEngine(fq, &mockLLM{category: services.ResponseCategoryPartial})
	ctx := context.Background()

	result, err := eng.ComputeNext(ctx, inv.InterviewID.Bytes)
	if err != nil {
		t.Fatal(err)
	}
	if result.Status != NextStatusNextQuestion {
		t.Errorf("expected next_question, got %s", result.Status)
	}
	if result.NextQuestionID == nil || *result.NextQuestionID != q2ID {
		t.Errorf("expected next question %v, got %v", q2ID, result.NextQuestionID)
	}
	if result.PromptOverride != "Custom follow-up" {
		t.Errorf("expected prompt override Custom follow-up, got %q", result.PromptOverride)
	}
}

func TestComputeNext_LastMessageUser_NoMoreQuestions_ReturnsDone(t *testing.T) {
	planID := uuid.New()
	q1ID := uuid.New()
	inv := &db.AppInterview{
		InterviewID:     pgUUID(uuid.New()),
		InterviewPlanID: pgUUID(planID),
		Status:          "in_progress",
	}
	questions := []db.AppInterviewQuestion{
		{InterviewQuestionID: pgUUID(q1ID), OrderIndex: 0, IsActive: true, Prompt: "Q1"},
	}
	messages := []db.AppInterviewMessage{
		{Sender: "ai", InterviewQuestionID: pgUUID(q1ID), Content: "Q1"},
		{Sender: "user", InterviewQuestionID: pgUUID(q1ID), Content: "Answer"},
	}
	fq := &fakeQueries{interview: inv, questions: questions, messages: messages}
	eng := NewEngine(fq, &mockLLM{category: services.ResponseCategoryStrong})
	ctx := context.Background()

	result, err := eng.ComputeNext(ctx, inv.InterviewID.Bytes)
	if err != nil {
		t.Fatal(err)
	}
	if result.Status != NextStatusDone {
		t.Errorf("expected done, got %s", result.Status)
	}
}

func TestComputeNext_ClassifyError_ReturnsError(t *testing.T) {
	planID := uuid.New()
	q1ID := uuid.New()
	inv := &db.AppInterview{
		InterviewID:     pgUUID(uuid.New()),
		InterviewPlanID: pgUUID(planID),
		Status:          "in_progress",
	}
	questions := []db.AppInterviewQuestion{
		{InterviewQuestionID: pgUUID(q1ID), OrderIndex: 0, IsActive: true, Prompt: "Q1"},
	}
	messages := []db.AppInterviewMessage{
		{Sender: "ai", InterviewQuestionID: pgUUID(q1ID), Content: "Q1"},
		{Sender: "user", InterviewQuestionID: pgUUID(q1ID), Content: "Answer"},
	}
	fq := &fakeQueries{interview: inv, questions: questions, messages: messages}
	classifyErr := errors.New("LLM unavailable")
	eng := NewEngine(fq, &mockLLM{err: classifyErr})
	ctx := context.Background()

	_, err := eng.ComputeNext(ctx, inv.InterviewID.Bytes)
	if err == nil {
		t.Fatal("expected error when ClassifyResponse fails")
	}
	if !errors.Is(err, classifyErr) {
		t.Errorf("expected classify error, got %v", err)
	}
}
