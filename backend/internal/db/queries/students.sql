-- name: CreateStudent :one
INSERT INTO app.students (email, display_name)
VALUES (@email, @display_name)
RETURNING *;

-- name: GetStudentByID :one
SELECT *
FROM app.students
WHERE student_id = @student_id::uuid;

-- name: GetStudentByEmail :one
SELECT *
FROM app.students
WHERE email = @email;

-- name: ListStudents :many
SELECT *
FROM app.students
ORDER BY display_name ASC, created_at DESC;

-- name: UpdateStudent :one
UPDATE app.students
SET display_name = @display_name,
    updated_at   = NOW()
WHERE student_id = @student_id::uuid
RETURNING *;
