-- Enable UUID extension (matches your other projects)
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Use a dedicated schema
CREATE SCHEMA IF NOT EXISTS app;

-- ENUM types
CREATE TYPE app.message_sender AS ENUM ('ai', 'user');
CREATE TYPE app.interview_status AS ENUM ('draft', 'in_progress', 'completed');

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
    weight              NUMERIC(5,2) NOT NULL DEFAULT 1.0,m
    order_index         INT NOT NULL DEFAULT 0,
    levels              JSONB,   -- { "A": "...", "B": "..." } etc.
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Interview plans
CREATE TABLE app.interview_plans (
    interview_plan_id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    rubric_id         UUID NOT NULL REFERENCES app.rubrics(rubric_id) ON DELETE CASCADE,
    title             TEXT NOT NULL,
    instructions      TEXT,
    config            JSONB NOT NULL DEFAULT '{}'::jsonb,
    status            app.interview_status NOT NULL DEFAULT 'draft',
    created_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at        TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Interview questions
CREATE TABLE app.interview_questions (
    interview_question_id  UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    interview_plan_id      UUID NOT NULL REFERENCES app.interview_plans(interview_plan_id) ON DELETE CASCADE,
    rubric_criterion_id    UUID REFERENCES app.rubric_criteria(rubric_criterion_id) ON DELETE SET NULL,
    prompt                 TEXT NOT NULL,
    question_type          TEXT NOT NULL DEFAULT 'open', -- future-proof
    order_index            INT NOT NULL DEFAULT 0,
    follow_up_to_id        UUID REFERENCES app.interview_questions(interview_question_id) ON DELETE SET NULL,
    follow_up_condition    TEXT,
    is_active              BOOLEAN NOT NULL DEFAULT TRUE,
    created_at             TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at             TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Interviews (simulation now, real students later)m
CREATE TABLE app.interviews (
    interview_id      UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    interview_plan_id UUID NOT NULL REFERENCES app.interview_plans(interview_plan_id) ON DELETE CASCADE,
    teacher_id        UUID REFERENCES app.teachers(teacher_id) ON DELETE SET NULL,
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

-- Indexing strategy: index all FKs and common filters
CREATE INDEX IF NOT EXISTS idx_rubrics_teacher_id
    ON app.rubrics(teacher_id);

CREATE INDEX IF NOT EXISTS idx_rubric_criteria_rubric_id
    ON app.rubric_criteria(rubric_id);

CREATE INDEX IF NOT EXISTS idx_interview_plans_rubric_id
    ON app.interview_plans(rubric_id);

CREATE INDEX IF NOT EXISTS idx_interview_questions_plan_id
    ON app.interview_questions(interview_plan_id);

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
