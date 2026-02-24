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

-- name: UpdateRubric :one
UPDATE app.rubrics
SET title = COALESCE(sqlc.narg('title'), title),
    description = COALESCE(sqlc.narg('description'), description),
    raw_text = COALESCE(sqlc.narg('raw_text'), raw_text),
    updated_at = NOW()
WHERE rubric_id = @rubric_id::uuid
RETURNING *;

-- name: DisableRubric :exec
UPDATE app.rubrics
SET is_enabled = FALSE,
    updated_at = NOW()
WHERE rubric_id = @rubric_id::uuid;
