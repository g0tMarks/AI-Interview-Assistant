-- name: CreateTeacher :one
INSERT INTO app.teachers (email, full_name, password_hash)
VALUES (@email, @full_name, @password_hash)
RETURNING *;

-- name: GetTeacherByID :one
SELECT *
FROM app.teachers
WHERE teacher_id = @teacher_id::uuid;

-- name: GetTeacherByEmail :one
SELECT *
FROM app.teachers
WHERE email = @email;

-- name: ListTeachers :many
SELECT *
FROM app.teachers
ORDER BY created_at DESC;
