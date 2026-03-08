# Architecture Overview

## Summary

AI-Interview-Assistant is a full-stack platform for AI-powered oral interview assessment ("Microviva"). It enables educators to create structured, rubric-driven interviews that verify student understanding and validate academic authorship.

---

## High-Level Architecture

```
┌─────────────────────────────────────────────────────────┐
│                        Frontend                         │
│              Next.js 16 (TypeScript/React)               │
│         Landing page, waitlist, teacher/student UI       │
└────────────────────────┬────────────────────────────────┘
                         │ HTTP / REST
┌────────────────────────▼────────────────────────────────┐
│                    Backend API                          │
│              Go + Chi v5 (Port 8080)                    │
│         JWT Auth · Rate Limiting · File Uploads          │
└────┬───────────────────┬────────────────────────────────┘
     │                   │
     ▼                   ▼
┌─────────┐     ┌────────────────────┐
│Postgres │     │  LLM Service       │
│   15    │     │  OpenAI / Anthropic│
└─────────┘     └────────────────────┘
```

---

## Components

### Frontend (`/frontend`)

| Item | Detail |
|------|--------|
| Framework | Next.js 16.1.6 (App Router) |
| Language | TypeScript / React 18 |
| Styling | Tailwind CSS |
| Storage | Upstash Redis (waitlist) |

**Key files:**
- `app/page.tsx` — Landing page (hero, value props, waitlist)
- `app/layout.tsx` — Root layout with metadata
- `app/api/waitlist/route.ts` — Edge function: email waitlist via Redis
- `components/` — Nav, Hero, InsightStrip, ValueProps, WaitlistForm

---

### Backend (`/backend`)

| Item | Detail |
|------|--------|
| Language | Go 1.25.4 |
| Router | Chi v5 |
| Database | PostgreSQL 15 via pgx/v5 |
| DB Access | SQLC (type-safe generated queries) |
| Port | 8080 |

**Entry point:** `backend/cmd/api/main.go`
- Loads `.env`
- Opens PostgreSQL connection
- Initialises LLM service (OpenAI or Anthropic)
- Starts HTTP server

#### Internal Packages

| Package | Responsibility |
|---------|---------------|
| `internal/api` | Server and router setup |
| `internal/api/handlers` | HTTP request handlers (one file per domain) |
| `internal/api/middleware` | Rate limiting, JWT auth |
| `internal/db` | SQLC-generated type-safe database access |
| `internal/db/queries` | Raw SQL query definitions |
| `internal/services` | LLM integration, authorship, student profiles |
| `internal/engine` | Interview branching logic |
| `internal/evaluation` | Interview scoring and summary generation |
| `internal/auth` | JWT token creation and validation |
| `internal/storage` | File upload abstraction (local filesystem) |
| `internal/rubricparser` | Rubric schema types and validation |
| `internal/extraction` | Text extraction from PDF, DOCX, Excel |
| `internal/safety` | Input sanitisation |
| `internal/validation` | Email and password validation |

---

### Database (`/backend/schema`)

All tables live in the `app` schema. Key tables:

| Table | Purpose |
|-------|---------|
| `teachers` | Educator accounts |
| `students` | Student accounts |
| `classes` | Teacher-owned classes |
| `roster` | Student ↔ class membership (M:N) |
| `rubrics` | Assessment rubrics |
| `rubric_criteria` | Scoring criteria within a rubric |
| `interview_plans` | Interview plans linked to rubrics |
| `interview_questions` | Questions in a plan |
| `interview_question_branches` | Branching logic: response category → next question |
| `interviews` | Interview instances (student × template) |
| `interview_messages` | Conversation turns (AI / student) |
| `interview_summaries` | Evaluation output with per-criterion evidence |
| `student_writing_samples` | Baseline and assessment writing samples |
| `student_profile_versions` | Writing fingerprints per student/semester |
| `submissions` | Assignment submissions for authorship workflow |
| `submission_artifacts` | Supporting files attached to submissions |
| `authorship_reports` | LLM-generated authorship confidence reports |
| `authorship_reviews` | Teacher review and decision records |
| `authorship_vivas` | Oral follow-up viva sessions |

Schema is applied with `make createdb`. Migrations live in `schema/migrations/`.

---

### LLM Service (`internal/services/llm.go`)

The `LLMService` interface abstracts all model calls. The `OpenAIService` implementation supports both OpenAI and Anthropic (selected by which API key is present in the environment).

| Method | Purpose |
|--------|---------|
| `GenerateInterviewInstructions` | Produces AI interviewer guidance from a rubric |
| `ParseRubric` | Extracts structured criteria and question plan from raw rubric text |
| `ClassifyResponse` | Categorises a student answer (strong / partial / incorrect / misconception / don't know) |
| `EvaluateInterview` | Generates a structured evaluation from a completed transcript |
| `GenerateAuthorshipReport` | Assesses authorship confidence and surfaces risk signals |
| `GenerateStudentProfile` | Creates a writing-style fingerprint from baseline samples |

All prompts include explicit injection-resistance instructions. User messages are capped at 4,000 runes and sanitised before being sent to the model.

---

### Interview Engine (`internal/engine/engine.go`)

`ComputeNext(interviewID)` drives question progression:

1. Load interview state and question plan.
2. Classify the most recent student response.
3. Follow the branch mapping (response category → next question).
4. Fall back to linear order if no branch is defined.
5. Return `NextResult` with status: `next_question`, `waiting_for_user`, or `done`.

---

### Evaluation Runner (`internal/evaluation/evaluation.go`)

`Run(interviewID)` is called when an interview completes:

1. Loads the full transcript and rubric criteria.
2. Calls `LLMService.EvaluateInterview`.
3. Persists the overall summary and per-criterion evidence.
4. Idempotent — skips if a summary already exists.

---

## Key Workflows

### Interview Flow

```
Teacher creates rubric
       ↓
LLM parses rubric → criteria + question plan
       ↓
Teacher creates interview template
       ↓
Interview instance created (student / class)
       ↓
Student answers questions
       ↓
Engine classifies response → selects next question (branching)
       ↓
Interview completes → Evaluation runner scores with LLM
       ↓
Summary + per-criterion evidence stored
```

### Authorship Verification Flow

```
Student submits assignment artifact
       ↓
Student writing profile generated (if baseline exists)
       ↓
Teacher initiates authorship viva (oral follow-up)
       ↓
Student answers questions about submission
       ↓
LLM generates authorship report
(confidence level, evidence signals, risk flags)
       ↓
Teacher reviews report and records final decision
```

### File Upload & Extraction

- Accepted formats: PDF, DOCX, Excel
- Libraries: `ledongthuc/pdf`, `xavier268/mydocx`, `xuri/excelize`
- Storage: local filesystem via `storage.Store` interface
- SHA-256 checksums computed on upload

---

## Security

| Concern | Approach |
|---------|---------|
| Authentication | JWT — student logs in with class code + email |
| Rate limiting | 100 requests / minute per IP (global) |
| Input sanitisation | Max 4,000 runes, control characters stripped |
| LLM prompt injection | Explicit resistance instructions in every prompt |
| Password storage | bcrypt via `golang.org/x/crypto` |
| Transactions | Full pgx transaction support for multi-step writes |

---

## Environment Variables

| Variable | Purpose | Default |
|----------|---------|---------|
| `DATABASE_URL` | PostgreSQL connection string | `postgres://postgres:mysecretpassword@localhost:5432/test-db?sslmode=disable` |
| `OPENAI_API_KEY` | OpenAI API key | — |
| `ANTHROPIC_API_KEY` | Anthropic API key (fallback) | — |
| `ANTHROPIC_MODEL` | Anthropic model name | `claude-sonnet-4-6` |
| `OPENAI_BASE_URL` | Custom OpenAI-compatible base URL | — |
| `JWT_SECRET` | JWT signing secret | `dev-secret-change-in-production` |
| `UPLOADS_DIR` | File upload directory | `./uploads` |
| `UPLOADS_MAX_BYTES` | Max upload size | 25 MiB |
| `APPENV` | Runtime environment | `development` |

---

## Development

```bash
# Start local Postgres
make postgres

# Apply schema
make createdb

# Regenerate SQLC types
make sqlc

# Run backend
cd backend && go run ./cmd/api

# Run frontend
cd frontend && npm run dev
```

Tests live alongside source files. Integration tests require a running Postgres instance.
