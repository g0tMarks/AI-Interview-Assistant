-- Student writing samples

-- name: CreateStudentWritingSample :one
INSERT INTO app.student_writing_samples (
    student_id,
    semester,
    assignment_name,
    source_type,
    written_at,
    text_content,
    is_teacher_verified,
    raw_features
)
VALUES (
    @student_id::uuid,
    @semester,
    @assignment_name,
    @source_type::app.writing_source_type,
    @written_at,
    @text_content,
    @is_teacher_verified,
    @raw_features
)
RETURNING *;

-- name: ListBaselineSamplesForStudentSemester :many
SELECT *
FROM app.student_writing_samples
WHERE student_id = @student_id::uuid
  AND semester = @semester
  AND source_type IN ('baseline_in_class', 'baseline_supervised');

-- name: GetStudentWritingSampleByID :one
SELECT *
FROM app.student_writing_samples
WHERE student_writing_sample_id = @student_writing_sample_id::uuid;

-- Profiles and baseline links

-- name: CreateStudentProfileVersion :one
INSERT INTO app.student_profile_versions (
    student_id,
    semester,
    version,
    profile_status,
    feature_summary,
    model_version
)
VALUES (
    @student_id::uuid,
    @semester,
    @version,
    @profile_status::app.profile_status,
    @feature_summary,
    @model_version
)
RETURNING *;

-- name: SetStudentProfileStatus :exec
UPDATE app.student_profile_versions
SET profile_status = @profile_status::app.profile_status
WHERE student_profile_version_id = @student_profile_version_id::uuid;

-- name: GetActiveProfileForStudentSemester :one
SELECT *
FROM app.student_profile_versions
WHERE student_id = @student_id::uuid
  AND semester = @semester
  AND profile_status = 'active'
ORDER BY version DESC
LIMIT 1;

-- name: AddBaselineSampleToProfile :exec
INSERT INTO app.student_profile_baseline_samples (
    student_profile_version_id,
    student_writing_sample_id
)
VALUES (
    @student_profile_version_id::uuid,
    @student_writing_sample_id::uuid
)
ON CONFLICT DO NOTHING;

-- name: ListBaselineSamplesForProfile :many
SELECT s.*
FROM app.student_profile_baseline_samples b
JOIN app.student_writing_samples s
  ON s.student_writing_sample_id = b.student_writing_sample_id
WHERE b.student_profile_version_id = @student_profile_version_id::uuid
ORDER BY s.created_at ASC;

-- Authorship analyses

-- name: CreateAuthorshipAnalysis :one
INSERT INTO app.authorship_analyses (
    student_id,
    student_writing_sample_id,
    student_profile_version_id,
    discrepancy_level,
    stylometric_distance,
    embedding_distance,
    feature_deltas,
    explanation,
    recommendation,
    model_version
)
VALUES (
    @student_id::uuid,
    @student_writing_sample_id::uuid,
    @student_profile_version_id::uuid,
    @discrepancy_level::app.discrepancy_level,
    @stylometric_distance,
    @embedding_distance,
    @feature_deltas,
    @explanation,
    @recommendation::app.authorship_recommendation,
    @model_version
)
RETURNING *;

-- name: GetAuthorshipAnalysisByID :one
SELECT *
FROM app.authorship_analyses
WHERE authorship_analysis_id = @authorship_analysis_id::uuid;

-- name: ListAuthorshipAnalysesForSample :many
SELECT *
FROM app.authorship_analyses
WHERE student_writing_sample_id = @student_writing_sample_id::uuid
ORDER BY created_at DESC;

-- Reviews

-- name: CreateAuthorshipReview :one
INSERT INTO app.authorship_reviews (
    authorship_analysis_id,
    teacher_id,
    review_status,
    teacher_decision,
    notes
)
VALUES (
    @authorship_analysis_id::uuid,
    @teacher_id::uuid,
    @review_status::app.review_status,
    @teacher_decision::app.teacher_decision,
    @notes
)
RETURNING *;

-- name: GetAuthorshipReviewByID :one
SELECT *
FROM app.authorship_reviews
WHERE authorship_review_id = @authorship_review_id::uuid;

-- name: ListPendingAuthorshipReviewsForTeacher :many
SELECT ar.*
FROM app.authorship_reviews ar
JOIN app.authorship_analyses aa
  ON ar.authorship_analysis_id = aa.authorship_analysis_id
WHERE ar.teacher_id = @teacher_id::uuid
  AND ar.review_status = 'pending'
ORDER BY aa.created_at DESC;

-- name: UpdateAuthorshipReview :one
UPDATE app.authorship_reviews
SET review_status    = @review_status::app.review_status,
    teacher_decision = @teacher_decision::app.teacher_decision,
    notes           = @notes,
    reviewed_at     = CASE
        WHEN @review_status::app.review_status IN ('reviewed', 'resolved') THEN NOW()
        ELSE reviewed_at
    END
WHERE authorship_review_id = @authorship_review_id::uuid
RETURNING *;

-- Vivas

-- name: CreateAuthorshipViva :one
INSERT INTO app.authorship_vivas (
    authorship_analysis_id,
    generated_questions,
    transcript,
    conceptual_alignment_summary,
    alignment_level,
    alignment_notes
)
VALUES (
    @authorship_analysis_id::uuid,
    @generated_questions,
    @transcript,
    @conceptual_alignment_summary,
    @alignment_level::app.alignment_level,
    @alignment_notes
)
RETURNING *;

-- name: GetAuthorshipVivaByID :one
SELECT *
FROM app.authorship_vivas
WHERE authorship_viva_id = @authorship_viva_id::uuid;

-- name: GetAuthorshipVivaByAnalysisID :one
SELECT *
FROM app.authorship_vivas
WHERE authorship_analysis_id = @authorship_analysis_id::uuid;

