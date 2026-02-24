package engine

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/g0tMarks/AI-Interview-Assistant/backend/internal/db"
	"github.com/g0tMarks/AI-Interview-Assistant/backend/internal/services"
)

// nextStepQueries is the subset of DB used by ComputeNext (allows tests to inject a fake).
type nextStepQueries interface {
	GetInterviewByID(ctx context.Context, interviewID pgtype.UUID) (db.AppInterview, error)
	ListQuestionsByPlan(ctx context.Context, interviewPlanID pgtype.UUID) ([]db.AppInterviewQuestion, error)
	ListMessagesByInterview(ctx context.Context, interviewID pgtype.UUID) ([]db.AppInterviewMessage, error)
	GetBranchByCategory(ctx context.Context, arg db.GetBranchByCategoryParams) (db.AppInterviewQuestionBranch, error)
}

// NextStatus is the state of the interview's "next" step.
type NextStatus string

const (
	NextStatusNextQuestion    NextStatus = "next_question"
	NextStatusWaitingForUser  NextStatus = "waiting_for_user"
	NextStatusDone            NextStatus = "done"
)

// NextResult is the result of computing the next step.
type NextResult struct {
	Status                NextStatus `json:"status"`
	NextQuestionID        *uuid.UUID `json:"nextQuestionId,omitempty"`
	Prompt                string     `json:"prompt,omitempty"`
	PromptOverride        string     `json:"promptOverride,omitempty"`
	WaitingForQuestionID  *uuid.UUID `json:"waitingForQuestionId,omitempty"`
	ClassifiedCategory    string     `json:"classifiedCategory,omitempty"`
}

// Engine computes the next interview step from plan, branches, and messages.
type Engine struct {
	Queries nextStepQueries
	LLM     services.LLMService
}

// NewEngine creates an engine with the given dependencies.
func NewEngine(queries nextStepQueries, llm services.LLMService) *Engine {
	return &Engine{Queries: queries, LLM: llm}
}

// ComputeNext returns the current "next" step for the interview (idempotent; no side effects).
func (e *Engine) ComputeNext(ctx context.Context, interviewID uuid.UUID) (*NextResult, error) {
	interviewIDPg := pgtype.UUID{Bytes: interviewID, Valid: true}

	inv, err := e.Queries.GetInterviewByID(ctx, interviewIDPg)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, err
		}
		return nil, err
	}

	if inv.Status == "completed" {
		return &NextResult{Status: NextStatusDone}, nil
	}

	planID := inv.InterviewPlanID
	if !planID.Valid {
		return &NextResult{Status: NextStatusDone}, nil
	}

	questions, err := e.Queries.ListQuestionsByPlan(ctx, planID)
	if err != nil {
		return nil, err
	}
	if len(questions) == 0 {
		return &NextResult{Status: NextStatusDone}, nil
	}

	messages, err := e.Queries.ListMessagesByInterview(ctx, interviewIDPg)
	if err != nil {
		return nil, err
	}

	// No messages: next is first question (lowest order_index).
	if len(messages) == 0 {
		q := firstQuestionByOrder(questions)
		if q == nil {
			return &NextResult{Status: NextStatusDone}, nil
		}
		return questionToNextResult(q, ""), nil
	}

	last := messages[len(messages)-1]
	if last.Sender == "ai" {
		// Waiting for user to answer the last question.
		var waitID *uuid.UUID
		if last.InterviewQuestionID.Valid {
			id := uuid.UUID(last.InterviewQuestionID.Bytes)
			waitID = &id
		}
		return &NextResult{
			Status:               NextStatusWaitingForUser,
			WaitingForQuestionID: waitID,
		}, nil
	}

	// Last message is from user. Determine which question they answered (from their message or previous AI).
	var answeredQuestionID uuid.UUID
	if last.InterviewQuestionID.Valid {
		answeredQuestionID = last.InterviewQuestionID.Bytes
	} else {
		// Find the last AI message to get the question it asked.
		for i := len(messages) - 1; i >= 0; i-- {
			if messages[i].Sender == "ai" && messages[i].InterviewQuestionID.Valid {
				answeredQuestionID = messages[i].InterviewQuestionID.Bytes
				break
			}
		}
	}

	// Classify user response (for branching).
	userContent := last.Content
	var questionPrompt string
	for _, q := range questions {
		if q.InterviewQuestionID.Valid && q.InterviewQuestionID.Bytes == answeredQuestionID {
			questionPrompt = q.Prompt
			break
		}
	}

	category, err := e.LLM.ClassifyResponse(ctx, questionPrompt, userContent)
	if err != nil {
		return nil, err
	}

	parentIDPg := pgtype.UUID{Bytes: answeredQuestionID, Valid: true}
	// Look up branch for this category; v1 uses misconception_tag_id = NULL.
	branch, err := e.Queries.GetBranchByCategory(ctx, db.GetBranchByCategoryParams{
		ParentQuestionID:   parentIDPg,
		ResponseCategory:   category,
		MisconceptionTagID: pgtype.UUID{},
	})
	if err == nil {
		if branch.TerminateInterview {
			return &NextResult{Status: NextStatusDone, ClassifiedCategory: category}, nil
		}
		if branch.NextQuestionID.Valid {
			nextID := uuid.UUID(branch.NextQuestionID.Bytes)
			var nextQ *db.AppInterviewQuestion
			for i := range questions {
				if questions[i].InterviewQuestionID.Valid && questions[i].InterviewQuestionID.Bytes == nextID {
					nextQ = &questions[i]
					break
				}
			}
			if nextQ != nil {
				promptOverride := ""
				if branch.FollowUpPromptOverride.Valid {
					promptOverride = branch.FollowUpPromptOverride.String
				}
				return &NextResult{
					Status:             NextStatusNextQuestion,
					NextQuestionID:     &nextID,
					Prompt:             nextQ.Prompt,
					PromptOverride:     promptOverride,
					ClassifiedCategory: category,
				}, nil
			}
		}
	}

	// No matching branch or no next_question_id: linear fallback — next question by order_index.
	currentOrder := orderIndexForQuestion(questions, answeredQuestionID)
	for _, q := range questions {
		if !q.InterviewQuestionID.Valid || !q.IsActive {
			continue
		}
		if q.OrderIndex > currentOrder {
			id := uuid.UUID(q.InterviewQuestionID.Bytes)
			return &NextResult{
				Status:             NextStatusNextQuestion,
				NextQuestionID:     &id,
				Prompt:             q.Prompt,
				ClassifiedCategory: category,
			}, nil
		}
	}

	return &NextResult{Status: NextStatusDone, ClassifiedCategory: category}, nil
}

func firstQuestionByOrder(questions []db.AppInterviewQuestion) *db.AppInterviewQuestion {
	var first *db.AppInterviewQuestion
	for i := range questions {
		if !questions[i].IsActive {
			continue
		}
		if first == nil || questions[i].OrderIndex < first.OrderIndex {
			first = &questions[i]
		}
	}
	return first
}

func orderIndexForQuestion(questions []db.AppInterviewQuestion, questionID uuid.UUID) int32 {
	for _, q := range questions {
		if q.InterviewQuestionID.Valid && q.InterviewQuestionID.Bytes == questionID {
			return q.OrderIndex
		}
	}
	return -1
}

func questionToNextResult(q *db.AppInterviewQuestion, promptOverride string) *NextResult {
	if q == nil {
		return &NextResult{Status: NextStatusDone}
	}
	prompt := q.Prompt
	var id *uuid.UUID
	if q.InterviewQuestionID.Valid {
		uid := uuid.UUID(q.InterviewQuestionID.Bytes)
		id = &uid
	}
	return &NextResult{
		Status:         NextStatusNextQuestion,
		NextQuestionID: id,
		Prompt:         prompt,
		PromptOverride: promptOverride,
	}
}
