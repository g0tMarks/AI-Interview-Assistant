-- name: CreateClass :one
INSERT INTO app.classes (teacher_id, name, class_code)
VALUES (@teacher_id::uuid, @name, @class_code)
RETURNING *;

-- name: GetClassByID :one
SELECT *
FROM app.classes
WHERE class_id = @class_id::uuid;

-- name: GetClassByCode :one
SELECT *
FROM app.classes
WHERE class_code = @class_code;

-- name: ListClassesByTeacher :many
SELECT *
FROM app.classes
WHERE teacher_id = @teacher_id::uuid
ORDER BY name ASC, created_at DESC;

-- name: UpdateClass :one
UPDATE app.classes
SET name       = @name,
    class_code = @class_code,
    updated_at  = NOW()
WHERE class_id = @class_id::uuid
RETURNING *;

-- name: DeleteClass :exec
DELETE FROM app.classes
WHERE class_id = @class_id::uuid;
