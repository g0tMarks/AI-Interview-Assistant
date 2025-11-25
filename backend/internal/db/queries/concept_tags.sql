-- name: CreateConceptTag :one
INSERT INTO app.concept_tags (
    name,
    display_name,
    description,
    curriculum_descriptor_id
)
VALUES (
    @name,
    @display_name,
    @description,
    @curriculum_descriptor_id::uuid
)
RETURNING *;

-- name: GetConceptTagByID :one
SELECT *
FROM app.concept_tags
WHERE concept_tag_id = @concept_tag_id::uuid;

-- name: GetConceptTagByName :one
SELECT *
FROM app.concept_tags
WHERE curriculum_descriptor_id = @curriculum_descriptor_id::uuid
  AND name = @name;

-- name: ListConceptTagsByDescriptor :many
SELECT *
FROM app.concept_tags
WHERE curriculum_descriptor_id = @curriculum_descriptor_id::uuid
ORDER BY name ASC;

-- name: UpdateConceptTag :one
UPDATE app.concept_tags
SET display_name = @display_name,
    description = @description,
    updated_at = NOW()
WHERE concept_tag_id = @concept_tag_id::uuid
RETURNING *;

