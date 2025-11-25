-- name: LinkQuestionToConceptTag :one
INSERT INTO app.interview_question_concept_tags (
    interview_question_id,
    concept_tag_id
)
VALUES (
    @interview_question_id::uuid,
    @concept_tag_id::uuid
)
RETURNING *;

-- name: GetConceptTagsByQuestionID :many
SELECT ct.*
FROM app.concept_tags ct
INNER JOIN app.interview_question_concept_tags iqct
    ON ct.concept_tag_id = iqct.concept_tag_id
WHERE iqct.interview_question_id = @interview_question_id::uuid
ORDER BY ct.name ASC;

-- name: GetQuestionsByConceptTagID :many
SELECT iq.*
FROM app.interview_questions iq
INNER JOIN app.interview_question_concept_tags iqct
    ON iq.interview_question_id = iqct.interview_question_id
WHERE iqct.concept_tag_id = @concept_tag_id::uuid
ORDER BY iq.order_index ASC;

-- name: UnlinkQuestionFromConceptTag :exec
DELETE FROM app.interview_question_concept_tags
WHERE interview_question_id = @interview_question_id::uuid
  AND concept_tag_id = @concept_tag_id::uuid;

-- name: DeleteAllConceptTagLinksForQuestion :exec
DELETE FROM app.interview_question_concept_tags
WHERE interview_question_id = @interview_question_id::uuid;

