-- name: CreateInterviewSummary :one
INSERT INTO app.interview_summaries (
    interview_id,
    overall_summary,
    strengths,
    areas_for_growth,
    suggested_next_steps,
    raw_llm_output
)
VALUES (
    @interview_id::uuid,
    @overall_summary,
    @strengths,
    @areas_for_growth,
    @suggested_next_steps,
    @raw_llm_output::jsonb
)
RETURNING *;

-- name: GetSummaryByInterviewID :one
SELECT *
FROM app.interview_summaries
WHERE interview_id = @interview_id::uuid;

-- name: UpdateInterviewSummary :one
UPDATE app.interview_summaries
SET overall_summary      = @overall_summary,
    strengths            = @strengths,
    areas_for_growth     = @areas_for_growth,
    suggested_next_steps = @suggested_next_steps,
    raw_llm_output       = @raw_llm_output::jsonb
WHERE interview_id = @interview_id::uuid
RETURNING *;
