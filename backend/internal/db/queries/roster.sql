-- name: AddToRoster :one
INSERT INTO app.roster (class_id, student_id)
VALUES (@class_id::uuid, @student_id::uuid)
RETURNING *;

-- name: RemoveFromRoster :exec
DELETE FROM app.roster
WHERE class_id = @class_id::uuid
  AND student_id = @student_id::uuid;

-- name: ListRosterByClass :many
SELECT r.class_id, r.student_id, r.joined_at,
       s.email AS student_email, s.display_name AS student_display_name
FROM app.roster r
JOIN app.students s ON s.student_id = r.student_id
WHERE r.class_id = @class_id::uuid
ORDER BY s.display_name ASC, r.joined_at ASC;

-- name: ListClassesByStudent :many
SELECT c.class_id, c.teacher_id, c.name, c.class_code, c.created_at, c.updated_at,
       r.joined_at
FROM app.roster r
JOIN app.classes c ON c.class_id = r.class_id
WHERE r.student_id = @student_id::uuid
ORDER BY c.name ASC;

-- name: IsStudentInClass :one
SELECT EXISTS(
    SELECT 1 FROM app.roster
    WHERE class_id = @class_id::uuid AND student_id = @student_id::uuid
) AS in_class;
