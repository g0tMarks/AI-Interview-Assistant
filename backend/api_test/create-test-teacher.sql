-- Create a test teacher for testing rubrics endpoint
-- Run this with: make psql (then copy/paste) or: docker exec -i test-db psql -U postgres -d test-db < create-test-teacher.sql

INSERT INTO app.teachers (email, full_name, password_hash)
VALUES ('test@example.com', 'Test Teacher', NULL)
ON CONFLICT (email) DO NOTHING
RETURNING teacher_id, email, full_name;

-- To see all teachers:
-- SELECT teacher_id, email, full_name FROM app.teachers;

