-- name: CreateInterviewQuestion :one
INSERT INTO app.interview_questions (
    interview_plan_id,
    rubric_criterion_id,
    prompt,
    question_type,
    order_index,
    is_active,
    follow_up_to_id,
    follow_up_condition
)
VALUES (
    @interview_plan_id::uuid,
    @rubric_criterion_id::uuid,
    @prompt,
    @question_type,
    @order_index,
    COALESCE(@is_active, TRUE),
    @follow_up_to_id::uuid,
    @follow_up_condition
)
RETURNING *;

-- name: GetQuestionByID :one
SELECT *
FROM app.interview_questions
WHERE interview_question_id = @interview_question_id::uuid;

-- name: ListQuestionsByPlan :many
SELECT *
FROM app.interview_questions
WHERE interview_plan_id = @interview_plan_id::uuid
ORDER BY order_index ASC, created_at ASC;

-- name: UpdateInterviewQuestion :one
UPDATE app.interview_questions
SET prompt = @prompt,
    question_type = @question_type,
    order_index = @order_index,
    is_active = @is_active,
    rubric_criterion_id = @rubric_criterion_id::uuid,
    follow_up_to_id = @follow_up_to_id::uuid,
    follow_up_condition = @follow_up_condition,
    updated_at = NOW()
WHERE interview_question_id = @interview_question_id::uuid
RETURNING *;

-- name: DeleteQuestionsByPlan :exec
DELETE FROM app.interview_questions
WHERE interview_plan_id = @interview_plan_id::uuid;
