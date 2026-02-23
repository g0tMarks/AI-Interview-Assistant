-- name: CreateInterviewPlan :one
INSERT INTO app.interview_plans (rubric_id, title, instructions, config, status, curriculum_subject, curriculum_level_band)
VALUES (
    @rubric_id::uuid,
    @title,
    @instructions,
    COALESCE(@config::jsonb, '{}'::jsonb),
    @status::app.interview_status,
    @curriculum_subject,
    @curriculum_level_band
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

-- name: DeletePlansByRubric :exec
DELETE FROM app.interview_plans
WHERE rubric_id = @rubric_id::uuid;

-- name: UpdateInterviewPlanStatus :exec
UPDATE app.interview_plans
SET status = @status::app.interview_status,
    updated_at = NOW()
WHERE interview_plan_id = @interview_plan_id::uuid;

-- name: UpdateInterviewPlan :one
UPDATE app.interview_plans
SET title = @title,
    instructions = @instructions,
    config = COALESCE(@config::jsonb, config),
    status = @status::app.interview_status,
    curriculum_subject = @curriculum_subject,
    curriculum_level_band = @curriculum_level_band,
    updated_at = NOW()
WHERE interview_plan_id = @interview_plan_id::uuid
RETURNING *;
