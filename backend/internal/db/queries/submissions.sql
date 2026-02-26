-- name: CreateSubmission :one
INSERT INTO app.submissions (student_id, rubric_id, status, title, notes)
VALUES (
    @student_id::uuid,
    @rubric_id::uuid,
    COALESCE(@status::app.submission_status, 'draft'::app.submission_status),
    @title,
    @notes
)
RETURNING *;

-- name: GetSubmissionByID :one
SELECT *
FROM app.submissions
WHERE submission_id = @submission_id::uuid;

-- name: UpdateSubmissionStatus :exec
UPDATE app.submissions
SET status = @status::app.submission_status,
    updated_at = NOW()
WHERE submission_id = @submission_id::uuid;

-- name: ListSubmissionsByStudent :many
SELECT *
FROM app.submissions
WHERE student_id = @student_id::uuid
ORDER BY created_at DESC;

-- name: ListSubmissionsByRubric :many
SELECT *
FROM app.submissions
WHERE rubric_id = @rubric_id::uuid
ORDER BY created_at DESC;
