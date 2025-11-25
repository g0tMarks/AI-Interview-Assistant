-- name: LinkInterviewPlanToCurriculumDescriptor :one
INSERT INTO app.interview_plan_curriculum_descriptors (
    interview_plan_id,
    curriculum_descriptor_id,
    primary_alignment
)
VALUES (
    @interview_plan_id::uuid,
    @curriculum_descriptor_id::uuid,
    COALESCE(@primary_alignment, FALSE)
)
RETURNING *;

-- name: GetCurriculumDescriptorsByPlanID :many
SELECT cd.*
FROM app.curriculum_descriptors cd
INNER JOIN app.interview_plan_curriculum_descriptors ipcd
    ON cd.curriculum_descriptor_id = ipcd.curriculum_descriptor_id
WHERE ipcd.interview_plan_id = @interview_plan_id::uuid
ORDER BY ipcd.primary_alignment DESC, cd.code ASC;

-- name: GetInterviewPlansByCurriculumDescriptorID :many
SELECT ip.*
FROM app.interview_plans ip
INNER JOIN app.interview_plan_curriculum_descriptors ipcd
    ON ip.interview_plan_id = ipcd.interview_plan_id
WHERE ipcd.curriculum_descriptor_id = @curriculum_descriptor_id::uuid
ORDER BY ip.created_at DESC;

-- name: UnlinkInterviewPlanFromCurriculumDescriptor :exec
DELETE FROM app.interview_plan_curriculum_descriptors
WHERE interview_plan_id = @interview_plan_id::uuid
  AND curriculum_descriptor_id = @curriculum_descriptor_id::uuid;

-- name: DeleteAllLinksForPlan :exec
DELETE FROM app.interview_plan_curriculum_descriptors
WHERE interview_plan_id = @interview_plan_id::uuid;

