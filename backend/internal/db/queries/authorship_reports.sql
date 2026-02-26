-- name: CreateAuthorshipReport :one
INSERT INTO app.authorship_reports (submission_id, interview_id, report)
VALUES (
    @submission_id::uuid,
    @interview_id::uuid,
    @report::jsonb
)
RETURNING *;

-- name: GetAuthorshipReportByID :one
SELECT *
FROM app.authorship_reports
WHERE report_id = @report_id::uuid;

-- name: GetLatestAuthorshipReportBySubmission :one
SELECT *
FROM app.authorship_reports
WHERE submission_id = @submission_id::uuid
ORDER BY created_at DESC
LIMIT 1;

-- name: ListAuthorshipReportsBySubmission :many
SELECT *
FROM app.authorship_reports
WHERE submission_id = @submission_id::uuid
ORDER BY created_at DESC;
