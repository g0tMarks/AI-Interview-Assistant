-- Enable UUID extension (matches your other projects)
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Use a dedicated schema
CREATE SCHEMA IF NOT EXISTS app;

-- ENUM types
CREATE TYPE app.message_sender AS ENUM ('ai', 'user');
CREATE TYPE app.interview_status AS ENUM ('draft', 'in_progress', 'completed');

-- Response category ENUM for branching logic
CREATE TYPE app.response_category AS ENUM (
    'strong',
    'partial',
    'incorrect',
    'misconception',
    'dont_know'
);

-- Teachers
CREATE TABLE app.teachers (
    teacher_id      UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    email           TEXT NOT NULL UNIQUE,
    full_name       TEXT NOT NULL,
    password_hash   TEXT,   -- optional if external auth
    is_enabled      BOOLEAN NOT NULL DEFAULT TRUE,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Students (identity for interview takers; may join classes via roster)
CREATE TABLE app.students (
    student_id   UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    email        TEXT NOT NULL UNIQUE,
    display_name TEXT NOT NULL,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Classes (teacher-owned; class_code used for student join / auth)
CREATE TABLE app.classes (
    class_id    UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    teacher_id  UUID NOT NULL REFERENCES app.teachers(teacher_id) ON DELETE CASCADE,
    name        TEXT NOT NULL,
    class_code  TEXT NOT NULL UNIQUE,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Roster (many-to-many: students in classes)
CREATE TABLE app.roster (
    class_id   UUID NOT NULL REFERENCES app.classes(class_id) ON DELETE CASCADE,
    student_id UUID NOT NULL REFERENCES app.students(student_id) ON DELETE CASCADE,
    joined_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (class_id, student_id)
);

-- Rubrics
CREATE TABLE app.rubrics (
    rubric_id       UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    teacher_id      UUID NOT NULL REFERENCES app.teachers(teacher_id) ON DELETE CASCADE,
    title           TEXT NOT NULL,
    description     TEXT,
    raw_text        TEXT NOT NULL, -- full rubric/task sheet
    is_enabled      BOOLEAN NOT NULL DEFAULT TRUE,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Rubric criteria
CREATE TABLE app.rubric_criteria (
    rubric_criterion_id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    rubric_id           UUID NOT NULL REFERENCES app.rubrics(rubric_id) ON DELETE CASCADE,
    name                TEXT NOT NULL,
    description         TEXT,
    weight              NUMERIC(5,2) NOT NULL DEFAULT 1.0,
    order_index         INT NOT NULL DEFAULT 0,
    levels              JSONB,   -- { "A": "...", "B": "..." } etc.
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

--------------------------------------------------------------------------------
-- Interview plans
--------------------------------------------------------------------------------

CREATE TABLE app.interview_plans (
    interview_plan_id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    rubric_id         UUID NOT NULL REFERENCES app.rubrics(rubric_id) ON DELETE CASCADE,
    title             TEXT NOT NULL,
    instructions      TEXT,
    config            JSONB NOT NULL DEFAULT '{}'::jsonb,
    status            app.interview_status NOT NULL DEFAULT 'draft',

    -- High-level curriculum alignment summaries for quick filtering
    curriculum_subject    TEXT,    -- denormalised from curriculum_descriptors.subject
    curriculum_level_band TEXT,    -- denormalised from curriculum_descriptors.level_band
    
    created_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at        TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

--------------------------------------------------------------------------------
-- Curriculum alignment tables (Victorian Curriculum / subject-agnostic)
--------------------------------------------------------------------------------

-- High-level curriculum descriptor, e.g. VCDTDI038, VCMNA350, VCELT458, etc.
CREATE TABLE app.curriculum_descriptors (
    curriculum_descriptor_id  UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    code                      TEXT NOT NULL,          -- e.g. 'VCDTDI038'
    subject                   TEXT NOT NULL,          -- e.g. 'Digital Technologies'
    level_band                TEXT NOT NULL,          -- e.g. '7-8', '9-10'
    strand                    TEXT,                   -- e.g. 'Digital Systems', 'Data and Information'
    substrand                 TEXT,                   -- optional sub-strand name
    description               TEXT NOT NULL,          -- official content description text
    elaborations              JSONB,                  -- array of elaboration strings, if you want them
    achievement_standard_excerpt  TEXT,               -- relevant bit of the achievement standard
    metadata                  JSONB NOT NULL DEFAULT '{}'::jsonb, -- for any extra flags/tags
    created_at                TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at                TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_curriculum_descriptors_code UNIQUE (code)
);

-- Link interview plans to one or more curriculum descriptors
-- (Many-to-many: one plan may hit multiple descriptors; one descriptor used by many plans)
CREATE TABLE app.interview_plan_curriculum_descriptors (
    interview_plan_id         UUID NOT NULL REFERENCES app.interview_plans(interview_plan_id) ON DELETE CASCADE,
    curriculum_descriptor_id  UUID NOT NULL REFERENCES app.curriculum_descriptors(curriculum_descriptor_id) ON DELETE CASCADE,
    primary_alignment         BOOLEAN NOT NULL DEFAULT FALSE, -- e.g. the main descriptor
    created_at                TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (interview_plan_id, curriculum_descriptor_id)
);

--------------------------------------------------------------------------------
-- Concept and misconception tagging for questions
--------------------------------------------------------------------------------

-- Concept tags like 'DHCP_process', 'IP_addressing', 'Pythagoras', etc.
CREATE TABLE app.concept_tags (
    concept_tag_id           UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name                     TEXT NOT NULL,          -- short slug, e.g. 'DHCP_process'
    display_name             TEXT,                   -- optional nice label
    description              TEXT,
    curriculum_descriptor_id UUID REFERENCES app.curriculum_descriptors(curriculum_descriptor_id)
                               ON DELETE SET NULL,   -- optional: concept anchored to a descriptor
    created_at               TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at               TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_concept_tags_unique_per_descriptor
        UNIQUE (curriculum_descriptor_id, name)
);

-- Misconceptions, optionally linked to a concept tag
CREATE TABLE app.misconception_tags (
    misconception_tag_id     UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name                     TEXT NOT NULL,         -- e.g. 'router_is_internet'
    display_name             TEXT,                  -- e.g. 'Router is the internet'
    description              TEXT,
    concept_tag_id           UUID REFERENCES app.concept_tags(concept_tag_id) ON DELETE SET NULL,
    curriculum_descriptor_id UUID REFERENCES app.curriculum_descriptors(curriculum_descriptor_id)
                               ON DELETE SET NULL,
    created_at               TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at               TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_misconception_tags_unique_per_descriptor
        UNIQUE (curriculum_descriptor_id, name)
);

--------------------------------------------------------------------------------
-- Interview questions
--------------------------------------------------------------------------------

-- Interview questions
CREATE TABLE app.interview_questions (
    interview_question_id  UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    interview_plan_id      UUID NOT NULL REFERENCES app.interview_plans(interview_plan_id) ON DELETE CASCADE,
    rubric_criterion_id    UUID REFERENCES app.rubric_criteria(rubric_criterion_id) ON DELETE SET NULL,
    
    prompt                 TEXT NOT NULL,
    question_type          TEXT NOT NULL DEFAULT 'open', -- future-proof
    order_index            INT NOT NULL DEFAULT 0,
    is_active              BOOLEAN NOT NULL DEFAULT TRUE,
    
    follow_up_to_id        UUID REFERENCES app.interview_questions(interview_question_id) ON DELETE SET NULL,
    follow_up_condition    TEXT,
    created_at             TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at             TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Link questions to concept tags (many-to-many)
CREATE TABLE app.interview_question_concept_tags (
    interview_question_id  UUID NOT NULL REFERENCES app.interview_questions(interview_question_id) ON DELETE CASCADE,
    concept_tag_id         UUID NOT NULL REFERENCES app.concept_tags(concept_tag_id) ON DELETE CASCADE,
    PRIMARY KEY (interview_question_id, concept_tag_id)
);

-- NEW: Optional link questions to misconception tags they are intended to probe
CREATE TABLE app.interview_question_misconception_tags (
    interview_question_id  UUID NOT NULL REFERENCES app.interview_questions(interview_question_id) ON DELETE CASCADE,
    misconception_tag_id   UUID NOT NULL REFERENCES app.misconception_tags(misconception_tag_id) ON DELETE CASCADE,
    PRIMARY KEY (interview_question_id, misconception_tag_id)
);

--------------------------------------------------------------------------------
-- Structured branching rules for each question using response categories
--------------------------------------------------------------------------------

CREATE TABLE app.interview_question_branches (
    interview_question_branch_id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),

    parent_question_id      UUID NOT NULL REFERENCES app.interview_questions(interview_question_id) ON DELETE CASCADE,

    -- How the model classifies the student's response to the parent question
    response_category       app.response_category NOT NULL,

    -- Optional: narrow to a specific misconception tag
    misconception_tag_id    UUID REFERENCES app.misconception_tags(misconception_tag_id) ON DELETE SET NULL,

    -- What question to ask next (child). Can be NULL if this rule ends the interview sequence.
    next_question_id        UUID REFERENCES app.interview_questions(interview_question_id) ON DELETE SET NULL,

    -- Optional override prompt, if you want a tailored follow-up wording
    follow_up_prompt_override TEXT,

    -- Allow marking that this branch should terminate the interview after this step
    terminate_interview     BOOLEAN NOT NULL DEFAULT FALSE,

    order_index             INT NOT NULL DEFAULT 0, -- if multiple branches match, which to prefer

    created_at              TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at              TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT uq_question_branch_category UNIQUE (parent_question_id, response_category, misconception_tag_id)
);

--------------------------------------------------------------------------------
-- Interviews (simulation now, real students later)
--------------------------------------------------------------------------------

CREATE TABLE app.interviews (
    interview_id      UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    interview_plan_id UUID NOT NULL REFERENCES app.interview_plans(interview_plan_id) ON DELETE CASCADE,
    teacher_id        UUID REFERENCES app.teachers(teacher_id) ON DELETE SET NULL,
    student_id        UUID REFERENCES app.students(student_id) ON DELETE SET NULL,
    simulated         BOOLEAN NOT NULL DEFAULT TRUE,
    student_name      TEXT,
    status            app.interview_status NOT NULL DEFAULT 'in_progress',
    started_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    completed_at      TIMESTAMPTZ
);

-- Interview messages (chat log)
CREATE TABLE app.interview_messages (
    interview_message_id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    interview_id         UUID NOT NULL REFERENCES app.interviews(interview_id) ON DELETE CASCADE,
    sender               app.message_sender NOT NULL,
    interview_question_id UUID REFERENCES app.interview_questions(interview_question_id) ON DELETE SET NULL,
    content              TEXT NOT NULL,
    created_at           TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Interview summaries
CREATE TABLE app.interview_summaries (
    interview_summary_id  UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    interview_id          UUID NOT NULL UNIQUE REFERENCES app.interviews(interview_id) ON DELETE CASCADE,
    overall_summary       TEXT,
    strengths             TEXT,
    areas_for_growth      TEXT,
    suggested_next_steps  TEXT,
    raw_llm_output        JSONB,
    created_at            TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Criterion-level evidence
CREATE TABLE app.criterion_evidence (
    criterion_evidence_id  UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    interview_summary_id   UUID NOT NULL REFERENCES app.interview_summaries(interview_summary_id) ON DELETE CASCADE,
    rubric_criterion_id    UUID NOT NULL REFERENCES app.rubric_criteria(rubric_criterion_id) ON DELETE CASCADE,
    level                  TEXT,          -- later could be its own ENUM
    evidence_text          TEXT,
    model_confidence       NUMERIC(3,2),  -- 0.00–1.00
    created_at             TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

--------------------------------------------------------------------------------
-- Indexing strategy: index all FKs and common filters
--------------------------------------------------------------------------------

CREATE INDEX IF NOT EXISTS idx_students_email
    ON app.students(email);

CREATE INDEX IF NOT EXISTS idx_classes_teacher_id
    ON app.classes(teacher_id);

CREATE INDEX IF NOT EXISTS idx_classes_class_code
    ON app.classes(class_code);

CREATE INDEX IF NOT EXISTS idx_roster_class_id
    ON app.roster(class_id);

CREATE INDEX IF NOT EXISTS idx_roster_student_id
    ON app.roster(student_id);

CREATE INDEX IF NOT EXISTS idx_rubrics_teacher_id
    ON app.rubrics(teacher_id);

CREATE INDEX IF NOT EXISTS idx_rubric_criteria_rubric_id
    ON app.rubric_criteria(rubric_id);

CREATE INDEX IF NOT EXISTS idx_interview_plans_rubric_id
    ON app.interview_plans(rubric_id);

CREATE INDEX IF NOT EXISTS idx_interview_plans_subject_level
    ON app.interview_plans(curriculum_subject, curriculum_level_band);

CREATE INDEX IF NOT EXISTS idx_curriculum_descriptors_code
    ON app.curriculum_descriptors(code);

CREATE INDEX IF NOT EXISTS idx_interview_plan_curriculum_descriptors_plan_id
    ON app.interview_plan_curriculum_descriptors(interview_plan_id);

CREATE INDEX IF NOT EXISTS idx_interview_plan_curriculum_descriptors_descriptor_id
    ON app.interview_plan_curriculum_descriptors(curriculum_descriptor_id);

CREATE INDEX IF NOT EXISTS idx_concept_tags_descriptor_id
    ON app.concept_tags(curriculum_descriptor_id);

CREATE INDEX IF NOT EXISTS idx_misconception_tags_descriptor_id
    ON app.misconception_tags(curriculum_descriptor_id);

CREATE INDEX IF NOT EXISTS idx_interview_questions_plan_id
    ON app.interview_questions(interview_plan_id);

CREATE INDEX IF NOT EXISTS idx_interview_question_concept_tags_question_id
    ON app.interview_question_concept_tags(interview_question_id);

CREATE INDEX IF NOT EXISTS idx_interview_question_concept_tags_concept_id
    ON app.interview_question_concept_tags(concept_tag_id);

CREATE INDEX IF NOT EXISTS idx_interview_question_branches_parent_id
    ON app.interview_question_branches(parent_question_id);

CREATE INDEX IF NOT EXISTS idx_interview_question_branches_next_question_id
    ON app.interview_question_branches(next_question_id);

CREATE INDEX IF NOT EXISTS idx_interviews_plan_id
    ON app.interviews(interview_plan_id);

CREATE INDEX IF NOT EXISTS idx_interviews_teacher_id
    ON app.interviews(teacher_id);

CREATE INDEX IF NOT EXISTS idx_interview_messages_interview_id
    ON app.interview_messages(interview_id);

CREATE INDEX IF NOT EXISTS idx_criterion_evidence_summary_id
    ON app.criterion_evidence(interview_summary_id);

CREATE INDEX IF NOT EXISTS idx_criterion_evidence_criterion_id
    ON app.criterion_evidence(rubric_criterion_id);

--------------------------------------------------------------------------------
-- Authorship v1: submissions, artifacts, reports; extend interviews
--------------------------------------------------------------------------------

CREATE TYPE app.submission_status AS ENUM (
    'draft',
    'submitted',
    'viva_in_progress',
    'viva_completed',
    'report_ready'
);

CREATE TYPE app.artifact_type AS ENUM (
    'main_text',
    'draft_checkpoint',
    'revision_note',
    'citation_source',
    'file_ref'
);

CREATE TABLE app.submissions (
    submission_id   UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    student_id      UUID NOT NULL REFERENCES app.students(student_id) ON DELETE CASCADE,
    rubric_id       UUID NOT NULL REFERENCES app.rubrics(rubric_id) ON DELETE CASCADE,
    status          app.submission_status NOT NULL DEFAULT 'draft',
    title           TEXT,
    notes           TEXT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE app.submission_artifacts (
    artifact_id     UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    submission_id   UUID NOT NULL REFERENCES app.submissions(submission_id) ON DELETE CASCADE,
    artifact_type   app.artifact_type NOT NULL,
    payload         JSONB NOT NULL DEFAULT '{}'::jsonb,
    order_index     INT NOT NULL DEFAULT 0,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

ALTER TABLE app.interviews
    ADD COLUMN IF NOT EXISTS submission_id UUID REFERENCES app.submissions(submission_id) ON DELETE SET NULL;

CREATE TABLE app.authorship_reports (
    report_id       UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    submission_id   UUID NOT NULL REFERENCES app.submissions(submission_id) ON DELETE CASCADE,
    interview_id    UUID REFERENCES app.interviews(interview_id) ON DELETE SET NULL,
    report          JSONB NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_submissions_student_id ON app.submissions(student_id);
CREATE INDEX IF NOT EXISTS idx_submissions_rubric_id ON app.submissions(rubric_id);
CREATE INDEX IF NOT EXISTS idx_submissions_status ON app.submissions(status);
CREATE INDEX IF NOT EXISTS idx_submission_artifacts_submission_id ON app.submission_artifacts(submission_id);
CREATE INDEX IF NOT EXISTS idx_interviews_submission_id ON app.interviews(submission_id);
CREATE INDEX IF NOT EXISTS idx_authorship_reports_submission_id ON app.authorship_reports(submission_id);
CREATE INDEX IF NOT EXISTS idx_authorship_reports_submission_created ON app.authorship_reports(submission_id, created_at DESC);
