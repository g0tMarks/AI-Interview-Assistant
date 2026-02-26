-- Authorship v1: submissions, artifacts, reports; extend interviews.
-- Run this on an existing DB that already has schema.sql applied.
-- For fresh installs, the same DDL is in schema.sql.

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
