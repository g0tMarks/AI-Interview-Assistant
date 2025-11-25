-- name: CreateInterviewMessage :one
INSERT INTO app.interview_messages (
    interview_id,
    sender,
    interview_question_id,
    content
)
VALUES (
    @interview_id::uuid,
    @sender::app.message_sender,
    @interview_question_id::uuid,
    @content
)
RETURNING *;

-- name: ListMessagesByInterview :many
SELECT *
FROM app.interview_messages
WHERE interview_id = @interview_id::uuid
ORDER BY created_at ASC;
