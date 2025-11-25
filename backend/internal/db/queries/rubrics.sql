-- name: CreateRubric :one
INSERT INTO app.rubrics (teacher_id, title, description, raw_text)
VALUES (@teacher_id::uuid, @title, @description, @raw_text)
RETURNING *;

-- name: GetRubricByID :one
SELECT *
FROM app.rubrics
WHERE rubric_id = @rubric_id::uuid;

-- name: ListRubricsByTeacher :many
SELECT *
FROM app.rubrics
WHERE teacher_id = @teacher_id::uuid
ORDER BY created_at DESC;

-- name: DisableRubric :exec
UPDATE app.rubrics
SET is_enabled = FALSE,
    updated_at = NOW()
WHERE rubric_id = @rubric_id::uuid;
