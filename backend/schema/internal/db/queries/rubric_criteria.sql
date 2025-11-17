-- name: CreateRubricCriterion :one
INSERT INTO app.rubric_criteria (rubric_id, name, description, weight, order_index, levels)
VALUES (
    @rubric_id::uuid,
    @name,
    @description,
    @weight,
    @order_index,
    @levels::jsonb
)
RETURNING *;

-- name: ListCriteriaByRubric :many
SELECT *
FROM app.rubric_criteria
WHERE rubric_id = @rubric_id::uuid
ORDER BY order_index ASC, created_at ASC;

-- name: DeleteCriteriaByRubric :exec
DELETE FROM app.rubric_criteria
WHERE rubric_id = @rubric_id::uuid;
