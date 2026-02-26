-- name: CreateInterview :one
INSERT INTO app.interviews (
    interview_plan_id,
    teacher_id,
    student_id,
    simulated,
    student_name,
    status,
    submission_id
)
VALUES (
    @interview_plan_id::uuid,
    @teacher_id::uuid,
    @student_id::uuid,
    @simulated,
    @student_name,
    @status::app.interview_status,
    @submission_id::uuid
)
RETURNING *;

-- name: GetInterviewByID :one
SELECT *
FROM app.interviews
WHERE interview_id = @interview_id::uuid;

-- name: UpdateInterviewStatus :exec
UPDATE app.interviews
SET status = @status::app.interview_status,
    completed_at = CASE
        WHEN @status::app.interview_status = 'completed' THEN NOW()
        ELSE completed_at
    END
WHERE interview_id = @interview_id::uuid;

-- name: ListInterviewsByPlan :many
SELECT *
FROM app.interviews
WHERE interview_plan_id = @interview_plan_id::uuid
ORDER BY started_at DESC;

-- name: GetInterviewBySubmissionID :one
SELECT *
FROM app.interviews
WHERE submission_id = @submission_id::uuid
ORDER BY started_at DESC
LIMIT 1;

-- name: LinkInterviewToSubmission :exec
UPDATE app.interviews
SET submission_id = @submission_id::uuid
WHERE interview_id = @interview_id::uuid;
