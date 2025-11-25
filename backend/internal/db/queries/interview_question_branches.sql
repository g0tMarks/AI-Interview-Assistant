-- name: CreateInterviewQuestionBranch :one
INSERT INTO app.interview_question_branches (
    parent_question_id,
    response_category,
    misconception_tag_id,
    next_question_id,
    follow_up_prompt_override,
    terminate_interview,
    order_index
)
VALUES (
    @parent_question_id::uuid,
    @response_category::app.response_category,
    @misconception_tag_id::uuid,
    @next_question_id::uuid,
    @follow_up_prompt_override,
    COALESCE(@terminate_interview, FALSE),
    @order_index
)
RETURNING *;

-- name: GetBranchesByParentQuestionID :many
SELECT *
FROM app.interview_question_branches
WHERE parent_question_id = @parent_question_id::uuid
ORDER BY order_index ASC, created_at ASC;

-- name: GetBranchByCategory :one
SELECT *
FROM app.interview_question_branches
WHERE parent_question_id = @parent_question_id::uuid
  AND response_category = @response_category::app.response_category
  AND (misconception_tag_id = @misconception_tag_id::uuid OR (@misconception_tag_id::uuid IS NULL AND misconception_tag_id IS NULL))
ORDER BY order_index ASC
LIMIT 1;

-- name: DeleteBranchesByParentQuestionID :exec
DELETE FROM app.interview_question_branches
WHERE parent_question_id = @parent_question_id::uuid;

-- name: UpdateInterviewQuestionBranch :one
UPDATE app.interview_question_branches
SET next_question_id = @next_question_id::uuid,
    follow_up_prompt_override = @follow_up_prompt_override,
    terminate_interview = @terminate_interview,
    order_index = @order_index,
    updated_at = NOW()
WHERE interview_question_branch_id = @interview_question_branch_id::uuid
RETURNING *;

