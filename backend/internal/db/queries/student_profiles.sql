-- name: CreateStudentProfile :one
INSERT INTO app.student_profiles (student_id, profile)
VALUES (@student_id::uuid, @profile::jsonb)
RETURNING *;

-- name: GetLatestStudentProfileByStudent :one
SELECT *
FROM app.student_profiles
WHERE student_id = @student_id::uuid
ORDER BY created_at DESC
LIMIT 1;

