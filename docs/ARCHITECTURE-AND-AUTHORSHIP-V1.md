# Architecture Map & Authorship v1 Design

## 1. Current Architecture Summary

### Entrypoints & routing
- **Entrypoint:** `backend/cmd/api/main.go` — loads `.env`, connects Postgres via `pgx.Connect(DATABASE_URL)`, builds `api.Dependencies`, calls `api.NewServer(deps)`, then `http.ListenAndServe(":8080", srv)`.
- **Router:** `backend/internal/api/router.go` — `NewRouter(deps Dependencies) http.Handler` builds `chi.NewRouter()`, applies middleware, instantiates handlers from `deps`, registers all routes.
- **Middleware:** `RateLimitIP(100, time.Minute)` global; `RequireStudentAuth(jwtSecret)` only on the `/student/me` group.

### Route map (relevant)
| Method + Path | Handler |
|---------------|---------|
| GET /health | HealthHandler.Health |
| POST /rubrics, GET /rubrics, PATCH /rubrics/{id}, POST /rubrics/{id}/parse, PUT /rubrics/{id}/criteria-and-plan, POST /rubrics/upload | RubricHandler |
| POST /teachers/register, GET /teachers/{id}/results | TeacherHandler |
| POST /interview-templates | InterviewTemplateHandler |
| POST /interviews, GET /interviews/{id}, POST/GET /interviews/{id}/messages, GET/POST /interviews/{id}/next, GET /interviews/{id}/results, GET /interviews/{id}/summary | InterviewHandler |
| POST /uploads, GET /uploads/{key} | UploadHandler |
| POST /auth/student/login | AuthHandler.StudentLogin |
| GET /student/me (auth required) | StudentHandler.GetMe |
| POST/GET /students, GET/PATCH /students/{id} | StudentHandler |
| POST/GET /classes, GET/PATCH/DELETE /classes/{id}, POST /classes/{id}/interviews/bulk | ClassHandler |
| GET/POST/DELETE roster, POST /classes/{id}/roster/upload | RosterHandler |

### Handlers & services
- **Handlers:** `backend/internal/api/handlers/*.go` — each gets `*db.Queries` and optionally `LLMService`, `TxBeginner`, `Storage`. No globals; DI only.
- **LLM:** `backend/internal/services/llm.go` — `LLMService` interface (GenerateInterviewInstructions, ParseRubric, ClassifyResponse, EvaluateInterview); `OpenAIService` implementation (OpenAI or Anthropic).

### Database
- **Schema:** All tables in schema `app`. Single file: `backend/schema/schema.sql` (no versioned migrations today). Applied via `make createdb` (docker exec psql -f schema.sql).
- **Existing tables (summary):**
  - **Identity:** `app.teachers`, `app.students`, `app.classes`, `app.roster`
  - **Rubrics:** `app.rubrics`, `app.rubric_criteria`
  - **Plans:** `app.interview_plans`, `app.interview_plan_curriculum_descriptors`, `app.curriculum_descriptors`, `app.concept_tags`, `app.misconception_tags`
  - **Questions:** `app.interview_questions`, `app.interview_question_concept_tags`, `app.interview_question_misconception_tags`, `app.interview_question_branches`
  - **Runs:** `app.interviews` (interview_plan_id, teacher_id, student_id, simulated, status, started_at, completed_at), `app.interview_messages`, `app.interview_summaries`, `app.criterion_evidence`
- **Enums:** `app.message_sender` (ai, user), `app.interview_status` (draft, in_progress, completed), `app.response_category` (strong, partial, incorrect, misconception, dont_know).

### SQLC
- **Config:** `backend/schema/sqlc.yaml` — schema: `./schema.sql`, queries: `../internal/db/queries`, package `db`, pgx/v5, overrides for uuid/timestamptz/app enums.
- **Queries:** One `.sql` file per entity under `backend/internal/db/queries/`. Conventions: `:one` single row, `:many` slice, `:exec` no rows; params cast as `@x::uuid`, `@y::app.interview_status`, `@z::jsonb`.

### Auth
- **Student only:** `backend/internal/auth/jwt.go` — `IssueStudentToken`, `ValidateStudentToken`, `StudentIDFromClaims`; `context.go` — `WithStudentID`, `StudentIDFromContext`. Middleware `RequireStudentAuth` on `/student/me` only. No teacher/admin auth in code.

---

## 2. Authorship v1 Design (minimal, reusing existing)

### Goals
- (a) Capture student work (text + optional files)
- (b) Capture process evidence (drafts, revision history, notes, citations)
- (c) Run a short viva/interview from rubric + submission
- (d) Store an authorship report: confidence, reasons, flags, follow-ups

### Reuse
- **Rubric ingestion + question generation:** Keep existing rubric CRUD and parse; use existing interview plan + questions as the viva template (or generate viva questions from rubric + submission content via LLM).
- **Interview tables/endpoints:** Link a viva to a **submission** by adding `submission_id` to `app.interviews` (nullable). Reuse `app.interviews`, `app.interview_messages`, `app.interview_summaries` / `app.criterion_evidence` for the viva run; new **authorship reports** table stores the authorship-specific outcome (separate from interview summary).

### New concepts
- **Submission:** One per student per assessment (rubric or “assessment” context). Holds high-level status and links to rubric (or plan).
- **Artifact:** A piece of evidence attached to a submission (main text, draft checkpoint, note, citation/link). Metadata in DB; large text in `payload` JSONB or referenced upload key.
- **Viva:** An interview run tied to a submission (reuse `app.interviews` with `submission_id`). Questions can come from existing plan or from a “viva plan” generated from rubric + submission.
- **Authorship report:** One per submission (or per run), JSONB payload with overall_assessment, evidence_signals, risk_flags, recommended_followups, rubric_alignment, provenance.

### ERD-level description
- **app.submissions** — student_id, rubric_id (or assessment reference), status, optional title/notes, created_at, updated_at. One submission per “attempt” per student per rubric.
- **app.submission_artifacts** — submission_id, artifact_type (enum: main_text, draft_checkpoint, revision_note, citation_source, file_ref), payload (JSONB: { "text", "upload_key", "label", "created_at_override" }), order_index, created_at.
- **app.interviews** — add optional **submission_id** UUID NULL FK to app.submissions. When set, this interview is the “viva” for that submission.
- **app.authorship_reports** — submission_id, report payload (JSONB), optional interview_id (viva used), created_at. “Latest” by created_at DESC per submission.

### Tables and fields

#### app.submission_status (enum)
- `draft` — being collected
- `submitted` — submitted for review
- `viva_in_progress` — viva started
- `viva_completed` — viva done
- `report_ready` — authorship report generated

#### app.submissions
| Column        | Type      | Constraints |
|---------------|-----------|-------------|
| submission_id | UUID      | PK, default uuid_generate_v4() |
| student_id    | UUID      | NOT NULL, FK app.students ON DELETE CASCADE |
| rubric_id     | UUID      | NOT NULL, FK app.rubrics ON DELETE CASCADE |
| status        | app.submission_status | NOT NULL, DEFAULT 'draft' |
| title         | TEXT      | NULL |
| notes         | TEXT      | NULL |
| created_at    | TIMESTAMPTZ | NOT NULL DEFAULT NOW() |
| updated_at    | TIMESTAMPTZ | NOT NULL DEFAULT NOW() |

#### app.artifact_type (enum)
- `main_text`, `draft_checkpoint`, `revision_note`, `citation_source`, `file_ref`

#### app.submission_artifacts
| Column         | Type      | Constraints |
|----------------|-----------|-------------|
| artifact_id    | UUID      | PK, default uuid_generate_v4() |
| submission_id  | UUID      | NOT NULL, FK app.submissions ON DELETE CASCADE |
| artifact_type | app.artifact_type | NOT NULL |
| payload        | JSONB      | NOT NULL DEFAULT '{}' (text, upload_key, label, etc.) |
| order_index    | INT       | NOT NULL DEFAULT 0 |
| created_at     | TIMESTAMPTZ | NOT NULL DEFAULT NOW() |

#### app.interviews (extension)
- Add: `submission_id UUID NULL REFERENCES app.submissions(submission_id) ON DELETE SET NULL`
- Index: `idx_interviews_submission_id`

#### app.authorship_reports
| Column           | Type      | Constraints |
|------------------|-----------|-------------|
| report_id        | UUID      | PK, default uuid_generate_v4() |
| submission_id    | UUID      | NOT NULL, FK app.submissions ON DELETE CASCADE |
| interview_id     | UUID      | NULL, FK app.interviews ON DELETE SET NULL (viva used) |
| report           | JSONB     | NOT NULL (full report structure) |
| created_at       | TIMESTAMPTZ | NOT NULL DEFAULT NOW() |

### Report JSONB structure (authorship_reports.report)
```json
{
  "overall_assessment": {
    "level": "confident|moderate|low|concern",
    "confidence": 0.0–1.0,
    "summary": "string"
  },
  "evidence_signals": [
    { "signal": "string", "strength": "strong|moderate|weak", "explanation": "string", "supporting_quotes_or_refs": ["string"] }
  ],
  "risk_flags": [
    { "flag": "string", "severity": "high|medium|low", "details": "string" }
  ],
  "recommended_followups": [
    { "question": "string", "why": "string" }
  ],
  "rubric_alignment": { "criterion_id_or_name": "brief note" },
  "provenance": {
    "submission_artifact_ids": ["uuid"],
    "interview_id": "uuid",
    "report_generated_at": "ISO8601"
  }
}
```

### Indexes and constraints
- `idx_submissions_student_id` ON app.submissions(student_id)
- `idx_submissions_rubric_id` ON app.submissions(rubric_id)
- `idx_submissions_status` ON app.submissions(status)
- `idx_submission_artifacts_submission_id` ON app.submission_artifacts(submission_id)
- `idx_authorship_reports_submission_id` ON app.authorship_reports(submission_id)
- `idx_authorship_reports_created_at` ON app.authorship_reports(submission_id, created_at DESC) — for “latest by submission”

### API (minimal, match existing style)
- **POST /submissions** — create submission (student_id, rubric_id, optional title/notes). Returns submission.
- **GET /submissions/{id}** — get submission by ID.
- **POST /submissions/{id}/artifacts** — add artifact (type, payload, order_index). Returns artifact.
- **GET /submissions/{id}/artifacts** — list artifacts for submission.
- **POST /submissions/{id}/viva/start** — create interview linked to submission (create or reuse plan from rubric; create interview with submission_id). Returns interview.
- **POST /submissions/{id}/viva/messages** — append transcript message (delegate to existing interview messages using the submission’s viva interview_id).
- **GET /submissions/{id}/viva** — get viva interview for submission (single interview per submission for v1).
- **POST /submissions/{id}/authorship/run** — compute and store authorship report (call AuthorshipService, insert into authorship_reports). Returns report.
- **GET /submissions/{id}/authorship** — get latest authorship report for submission.

### LLM abstraction
- Introduce **AuthorshipService** interface (e.g. `GenerateAuthorshipReport(ctx, submissionID, interviewID, rubricTitle, submissionText, transcript string) (*AuthorshipReportPayload, error)`). Implement with existing LLM client behind the scenes; no change to existing LLM interface. Handlers call AuthorshipService only for the new report flow.

---

## 3. Implementation Plan Checklist (ordered)

1. **DB migration** — Add enum `app.submission_status`, `app.artifact_type`; tables `app.submissions`, `app.submission_artifacts`, `app.authorship_reports`; add `submission_id` to `app.interviews`; add indexes. Apply via new migration file and append to schema.sql.
2. **SQLC** — Add `submissions.sql`, `submission_artifacts.sql`, `authorship_reports.sql`; add queries to `interviews.sql` (e.g. GetInterviewBySubmissionID, UpdateInterviewSubmissionID). Run `make generate`.
3. **sqlc.yaml** — Add overrides for new enums `app.submission_status`, `app.artifact_type` → Go string.
4. **AuthorshipService** — Define interface and report payload struct in Go; implement with LLM call that builds prompt from submission + transcript + rubric, parses JSON into report struct.
5. **Handlers** — SubmissionHandler (Create, Get); SubmissionArtifactsHandler (Create, List); SubmissionVivaHandler (Start, Get, Messages proxy); SubmissionAuthorshipHandler (Run, Get). Use chi URL params `{id}` for submission ID.
6. **Router** — Register POST/GET /submissions, POST/GET /submissions/{id}/artifacts, POST /submissions/{id}/viva/start, GET /submissions/{id}/viva, POST /submissions/{id}/viva/messages, POST /submissions/{id}/authorship/run, GET /submissions/{id}/authorship. Wire Dependencies (add AuthorshipService to deps if needed).
7. **Backwards compatibility** — Leave existing rubric/interview endpoints unchanged; no deletions; add comment on any deprecated path if needed.
8. **Integration test** — One test: create rubric → create submission → add artifacts → start viva → add messages → run authorship → fetch report. Use existing test DB and handler pattern from integration_test.go.
9. **Follow-up notes** — Document SSO, UI, and scoring improvements in a short “Follow-up work” section (no implementation).

---

## 4. Relevant files (reference)

| Area        | Path |
|------------|------|
| Entrypoint | backend/cmd/api/main.go |
| Router     | backend/internal/api/router.go |
| Server/Deps| backend/internal/api/server.go |
| Handlers   | backend/internal/api/handlers/*.go |
| LLM        | backend/internal/services/llm.go |
| Authorship | backend/internal/services/authorship.go |
| Schema     | backend/schema/schema.sql |
| Migration  | backend/schema/migrations/000001_authorship_v1.sql |
| SQLC config| backend/schema/sqlc.yaml |
| Queries    | backend/internal/db/queries/*.sql |
| Auth       | backend/internal/api/middleware/auth.go, backend/internal/auth/*.go |
| Engine/Eval| backend/internal/engine, backend/internal/evaluation |

---

## 5. Implementation checklist (completed)

- [x] DB migration: enums, tables (submissions, submission_artifacts, authorship_reports), interviews.submission_id, indexes.
- [x] schema.sql updated with same DDL for fresh installs.
- [x] SQLC: submissions.sql, submission_artifacts.sql, authorship_reports.sql; interviews.sql extended (CreateInterview + submission_id, GetInterviewBySubmissionID, LinkInterviewToSubmission); sqlc.yaml overrides for new enums.
- [x] Authorship report payload struct and LLMService.GenerateAuthorshipReport (OpenAIService implementation).
- [x] Handlers: submissions.go (CreateSubmission, GetSubmission, CreateArtifact, ListArtifacts, StartViva, GetViva, VivaMessages, ListVivaMessages, RunAuthorship, GetAuthorship).
- [x] Router: POST/GET /submissions, POST/GET /submissions/{id}/artifacts, POST /submissions/{id}/viva/start, GET /submissions/{id}/viva, POST/GET /submissions/{id}/viva/messages, POST /submissions/{id}/authorship/run, GET /submissions/{id}/authorship.
- [x] Backwards compatibility: existing rubric/interview endpoints unchanged; CreateInterview accepts optional submission_id (existing callers pass pgtype.UUID{}).
- [x] Integration test: TestAuthorshipFlow (rubric → template → student → submission → artifacts → start viva → messages → run authorship → get report).

---

## 6. Follow-up work (not implemented)

- **SSO / teacher auth:** No teacher or admin auth in code today; add auth middleware and identity for teacher-scoped submission lists.
- **UI:** Frontend for submission creation, artifact upload, viva conduction, and report view.
- **Scoring improvements:** Optional numeric score or band in the report; link report to rubric criteria more formally.
- **Viva question seeding:** Currently reuses first interview plan for the rubric; could generate viva questions from rubric + submission text via LLM.
- **Migrations runner:** Project uses single schema.sql + optional migration file; consider golang-migrate or similar for versioned migrations in production.
