-- name: CreateCriterionEvidence :one
INSERT INTO app.criterion_evidence (
    interview_summary_id,
    rubric_criterion_id,
    level,
    evidence_text,
    model_confidence
)
VALUES (
    @interview_summary_id::uuid,
    @rubric_criterion_id::uuid,
    @level,
    @evidence_text,
    @model_confidence
)
RETURNING *;

-- name: GetCriterionEvidenceBySummaryID :many
SELECT *
FROM app.criterion_evidence
WHERE interview_summary_id = @interview_summary_id::uuid
ORDER BY created_at ASC;

-- name: GetCriterionEvidenceByCriterionID :many
SELECT *
FROM app.criterion_evidence
WHERE rubric_criterion_id = @rubric_criterion_id::uuid
ORDER BY created_at ASC;

-- name: DeleteCriterionEvidenceBySummaryID :exec
DELETE FROM app.criterion_evidence
WHERE interview_summary_id = @interview_summary_id::uuid;

