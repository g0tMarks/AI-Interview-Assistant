-- name: CreateInterviewPlan :one
INSERT INTO app.interview_plans (rubric_id, title, instructions, config, status)
VALUES (
    @rubric_id::uuid,
    @title,
    @instructions,
    COALESCE(@config::jsonb, '{}'::jsonb),
    @status::app.interview_status
)
RETURNING *;

-- name: GetInterviewPlanByID :one
SELECT *
FROM app.interview_plans
WHERE interview_plan_id = @interview_plan_id::uuid;

-- name: ListPlansByRubric :many
SELECT *
FROM app.interview_plans
WHERE rubric_id = @rubric_id::uuid
ORDER BY created_at DESC;

-- name: UpdateInterviewPlanStatus :exec
UPDATE app.interview_plans
SET status = @status::app.interview_status,
    updated_at = NOW()
WHERE interview_plan_id = @interview_plan_id::uuid;
