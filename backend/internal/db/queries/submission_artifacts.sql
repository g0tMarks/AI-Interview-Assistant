-- name: CreateSubmissionArtifact :one
INSERT INTO app.submission_artifacts (submission_id, artifact_type, payload, order_index)
VALUES (
    @submission_id::uuid,
    @artifact_type::app.artifact_type,
    COALESCE(@payload::jsonb, '{}'::jsonb),
    COALESCE(@order_index, 0)
)
RETURNING *;

-- name: GetSubmissionArtifactByID :one
SELECT *
FROM app.submission_artifacts
WHERE artifact_id = @artifact_id::uuid;

-- name: ListArtifactsBySubmission :many
SELECT *
FROM app.submission_artifacts
WHERE submission_id = @submission_id::uuid
ORDER BY order_index ASC, created_at ASC;
