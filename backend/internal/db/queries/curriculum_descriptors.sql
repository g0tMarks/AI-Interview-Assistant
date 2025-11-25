-- name: CreateCurriculumDescriptor :one
INSERT INTO app.curriculum_descriptors (
    code,
    subject,
    level_band,
    strand,
    substrand,
    description,
    elaborations,
    achievement_standard_excerpt,
    metadata
)
VALUES (
    @code,
    @subject,
    @level_band,
    @strand,
    @substrand,
    @description,
    @elaborations::jsonb,
    @achievement_standard_excerpt,
    COALESCE(@metadata::jsonb, '{}'::jsonb)
)
RETURNING *;

-- name: GetCurriculumDescriptorByID :one
SELECT *
FROM app.curriculum_descriptors
WHERE curriculum_descriptor_id = @curriculum_descriptor_id::uuid;

-- name: GetCurriculumDescriptorByCode :one
SELECT *
FROM app.curriculum_descriptors
WHERE code = @code;

-- name: ListCurriculumDescriptors :many
SELECT *
FROM app.curriculum_descriptors
WHERE (@subject::text IS NULL OR subject = @subject)
  AND (@level_band::text IS NULL OR level_band = @level_band)
ORDER BY code ASC;

-- name: UpdateCurriculumDescriptor :one
UPDATE app.curriculum_descriptors
SET code = @code,
    subject = @subject,
    level_band = @level_band,
    strand = @strand,
    substrand = @substrand,
    description = @description,
    elaborations = @elaborations::jsonb,
    achievement_standard_excerpt = @achievement_standard_excerpt,
    metadata = @metadata::jsonb,
    updated_at = NOW()
WHERE curriculum_descriptor_id = @curriculum_descriptor_id::uuid
RETURNING *;

