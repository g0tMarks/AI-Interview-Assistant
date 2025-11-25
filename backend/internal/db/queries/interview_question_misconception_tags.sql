-- name: LinkQuestionToMisconceptionTag :one
INSERT INTO app.interview_question_misconception_tags (
    interview_question_id,
    misconception_tag_id
)
VALUES (
    @interview_question_id::uuid,
    @misconception_tag_id::uuid
)
RETURNING *;

-- name: GetMisconceptionTagsByQuestionID :many
SELECT mt.*
FROM app.misconception_tags mt
INNER JOIN app.interview_question_misconception_tags iqmt
    ON mt.misconception_tag_id = iqmt.misconception_tag_id
WHERE iqmt.interview_question_id = @interview_question_id::uuid
ORDER BY mt.name ASC;

-- name: GetQuestionsByMisconceptionTagID :many
SELECT iq.*
FROM app.interview_questions iq
INNER JOIN app.interview_question_misconception_tags iqmt
    ON iq.interview_question_id = iqmt.interview_question_id
WHERE iqmt.misconception_tag_id = @misconception_tag_id::uuid
ORDER BY iq.order_index ASC;

-- name: UnlinkQuestionFromMisconceptionTag :exec
DELETE FROM app.interview_question_misconception_tags
WHERE interview_question_id = @interview_question_id::uuid
  AND misconception_tag_id = @misconception_tag_id::uuid;

-- name: DeleteAllMisconceptionTagLinksForQuestion :exec
DELETE FROM app.interview_question_misconception_tags
WHERE interview_question_id = @interview_question_id::uuid;

