-- name: CreateMisconceptionTag :one
INSERT INTO app.misconception_tags (
    name,
    display_name,
    description,
    concept_tag_id,
    curriculum_descriptor_id
)
VALUES (
    @name,
    @display_name,
    @description,
    @concept_tag_id::uuid,
    @curriculum_descriptor_id::uuid
)
RETURNING *;

-- name: GetMisconceptionTagByID :one
SELECT *
FROM app.misconception_tags
WHERE misconception_tag_id = @misconception_tag_id::uuid;

-- name: GetMisconceptionTagByName :one
SELECT *
FROM app.misconception_tags
WHERE curriculum_descriptor_id = @curriculum_descriptor_id::uuid
  AND name = @name;

-- name: ListMisconceptionTagsByDescriptor :many
SELECT *
FROM app.misconception_tags
WHERE curriculum_descriptor_id = @curriculum_descriptor_id::uuid
ORDER BY name ASC;

-- name: ListMisconceptionTagsByConcept :many
SELECT *
FROM app.misconception_tags
WHERE concept_tag_id = @concept_tag_id::uuid
ORDER BY name ASC;

-- name: UpdateMisconceptionTag :one
UPDATE app.misconception_tags
SET display_name = @display_name,
    description = @description,
    updated_at = NOW()
WHERE misconception_tag_id = @misconception_tag_id::uuid
RETURNING *;

