-- name: CreateInterviewQuestion :one
INSERT INTO app.interview_questions (
    interview_plan_id,
    rubric_criterion_id,
    prompt,
    question_type,
    order_index,
    follow_up_to_id,
    follow_up_condition
)
VALUES (
    @interview_plan_id::uuid,
    @rubric_criterion_id::uuid,
    @prompt,
    @question_type,
    @order_index,
    @follow_up_to_id::uuid,
    @follow_up_condition
)
RETURNING *;

-- name: ListQuestionsByPlan :many
SELECT *
FROM app.interview_questions
WHERE interview_plan_id = @interview_plan_id::uuid
ORDER BY order_index ASC, created_at ASC;

-- name: DeleteQuestionsByPlan :exec
DELETE FROM app.interview_questions
WHERE interview_plan_id = @interview_plan_id::uuid;
