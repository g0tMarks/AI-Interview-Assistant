# Interview Assistant – All Tasks Consolidated

This document consolidates all PRDs and implementation tasks from `.cursor/tasks/` into a single reference.

---

# Part 1: Core API Endpoints

## 1.1 PRD: Core API Endpoints

**Source**: `1. PRD-API-Endpoints.md`

### Problem Statement

The AI Interview Assistant backend currently has a basic health check endpoint and a rubric creation endpoint, but lacks the core API endpoints needed for the full interview workflow. Teachers need to be able to:

- List and retrieve rubrics they've created
- Create interview templates (interview plans) that define the structure and questions for interviews
- Create new interviews based on templates
- Retrieve interview details to view progress and results

### Target Users

- **Primary Users**: Teachers who use the AI Interview Assistant to conduct diagnostic or formative assessment interviews with students
- **Secondary Users**: System administrators who may need to manage or audit interview data

### User Stories

- **Rubric Management**: View all my rubrics in a list; see details of a specific rubric.
- **Interview Template Creation**: Create an interview template linked to a rubric; specify curriculum alignment (subject, level band).
- **Interview Management**: Create a new interview from a template; retrieve interview details by ID.

### Endpoints Summary

| Endpoint | Status | Purpose |
|----------|--------|---------|
| POST /rubrics | ✅ Implemented | Create a new rubric |
| GET /rubrics | ⚠️ Needs implementation | List all rubrics for a teacher |
| POST /interview-templates | ⚠️ Needs implementation | Create interview template linked to rubric |
| POST /interviews | ⚠️ Needs implementation | Create interview from template |
| GET /interviews/{id} | ⚠️ Needs implementation | Retrieve interview by ID |

*(Full request/response specs are in the individual PRD sections below.)*

### Non-Goals

- Authentication/Authorization; Pagination; Filtering/Sorting; Interview Updates; Interview Messages; Interview Summaries; Nested Resources; Bulk Operations; Soft Deletes

---

## 1.2 Implementation Tasks: Core API Endpoints

**Source**: `1.Tasks.md`

### Task 1: Update Dependencies Structure — ✅ Complete

- Update `api.Dependencies` to include `db.Queries`. Files: `backend/internal/api/server.go`, `main.go`.

### Task 2: Register POST /rubrics Route — ✅ Complete

- Register `CreateRubric` in `router.go`: `r.Post("/rubrics", rubricHandler.CreateRubric)`.

### Task 3: Implement GET /rubrics Handler — ⚠️ Pending

- Add `ListRubrics` to `RubricHandler` in `handlers/rubrics.go`. Query param `teacherId` (required, UUID). Call `ListRubricsByTeacher`, return 200 with array; 400 for invalid/missing teacherId, 500 for DB errors.

### Task 4: Register GET /rubrics Route — ⚠️ Pending

- In `router.go`: `r.Get("/rubrics", rubricHandler.ListRubrics)`.

### Task 5: Create Interview Template Handler — ⚠️ Pending

- Create `handlers/interview_templates.go`: `InterviewTemplateHandler`, `CreateInterviewTemplateRequest`/`InterviewTemplateResponse`, `CreateInterviewTemplate` with validation, rubric existence check, defaults (status="draft", config={}).

### Task 6: Register POST /interview-templates Route — ⚠️ Pending

- In `router.go`: `r.Post("/interview-templates", templateHandler.CreateInterviewTemplate)`.

### Task 7: Create Interview Handler — ⚠️ Pending

- Create `handlers/interviews.go`: `InterviewHandler`, `CreateInterviewRequest`/`InterviewResponse`, `CreateInterview` and `GetInterview` with validation, plan existence check, defaults (simulated=true, status="in_progress").

### Task 8: Register Interview Routes — ⚠️ Pending

- In `router.go`: POST `/interviews`, GET `/interviews/{id}`.

### Task 9: Update main.go to Pass Queries — ⚠️ Pending

- Ensure `main.go` passes `queries` in `api.Dependencies`.

### Task 10: Testing & Validation — ⚠️ Pending

- Manual test cases for GET/POST /rubrics, POST /interview-templates, POST /interviews, GET /interviews/{id} (success, validation, not found, server errors).

**Implementation Order**: 1 → 9 → 2 → 3 → 4 → 5 → 6 → 7 → 8 → 10.

---

# Part 2: Create Teacher Account

## 2.1 PRD: Create Teacher Account API Endpoint

**Source**: `2. PRD-API-Endpoints-CreateTeacherAccount.md`

### Problem Statement

Teachers need a way to create accounts. Currently there is no registration endpoint. The system needs a secure registration endpoint using email and password (SSO to be added later).

### Endpoint: POST /teachers/register

- **Request**: `email` (required), `fullName` (required), `password` (required).
- **Response**: 201 Created with `teacherId`, `email`, `fullName`, `isEnabled`, `createdAt`, `updatedAt` (no password hash).
- **Validation**: Email RFC 5322, unique; fullName non-empty (trimmed); password: min 8 chars, 1 upper, 1 lower, 1 number, 1 special.
- **Errors**: 400 (validation), 409 (email exists), 500 (DB/hashing).

### Non-Goals

SSO; Email verification; Password reset; 2FA; Account updates/deletion; Login/session; etc.

---

## 2.2 Implementation Tasks: Create Teacher Account

**Source**: `2. Tasks.md`

### Task 1: Add Password Validation Utility — ✅ Complete

- `backend/internal/utils/password.go` (or validation package): `ValidatePassword(password string) error` (length, upper, lower, number, special).

### Task 2: Add Email Validation Utility — ✅ Complete

- `ValidateEmail(email string) error` (RFC 5322).

### Task 3: Add Password Hashing Utility — ✅ Complete

- Same file: `HashPassword(password string) (string, error)` using bcrypt (cost 10–12).

### Task 4: Create Teacher Handler — ✅ Complete

- `handlers/teachers.go`: `TeacherHandler`, `RegisterTeacherRequest`/`TeacherResponse`, `RegisterTeacher` with validation, email uniqueness (409), hash password, never return hash.

### Task 5: Register POST /teachers/register Route — ✅ Complete

- `r.Post("/teachers/register", teacherHandler.RegisterTeacher)`.

### Task 6: Update go.mod Dependencies — ✅ Complete

- Ensure `golang.org/x/crypto/bcrypt` available.

### Task 7: Testing & Validation — ⚠️ Pending

- Manual tests: success, validation (400), conflict (409), server errors (500), security (hash stored, never returned).

**Order**: 6 → 1 → 2 → 3 → 4 → 5 → 7.

---

# Part 3: GET /rubrics

## 3.1 PRD: GET /rubrics

**Source**: `3. PRD-GET-rubrics.md`

- **Method/Path**: GET `/rubrics`
- **Query**: `teacherId` (required, UUID).
- **Response**: 200 OK with array of rubrics, ordered by `created_at DESC`. Empty array if none.
- **Errors**: 400 (missing/invalid teacherId), 500 (DB).
- **Implementation**: `ListRubricsByTeacher`, convert pgtype to standard types for JSON.

---

## 3.2 Implementation Tasks: GET /rubrics

**Source**: `3. Tasks.md`

### Task 1: Implement GET /rubrics Handler — ✅ Complete

- Add `ListRubrics` to `RubricHandler`; validate teacherId; call `ListRubricsByTeacher`; convert to `[]RubricResponse`; 200/400/500.

### Task 2: Register GET /rubrics Route — ✅ Complete

- `r.Get("/rubrics", rubricHandler.ListRubrics)`.

### Task 3: Testing & Validation — ✅ Complete

- Success (200, empty array for no rubrics), validation (400), server (500).

---

# Part 4: POST /interview-templates

## 4.1 PRD: POST /interview-templates

**Source**: `4. PRD-POST-interview-templates.md`

- **Method/Path**: POST `/interview-templates`
- **Request**: `rubricId` (required), `title` (required), `instructions`, `config` (JSON, default {}), `status` (draft|in_progress|completed, default "draft"), `curriculumSubject`, `curriculumLevelBand`.
- **Response**: 201 Created with full template object.
- **Validation**: rubricId valid UUID and existing rubric; title non-empty; status enum; config valid JSON.
- **Errors**: 400 (validation), 404 (rubric not found), 500.

---

## 4.2 Implementation Tasks: POST /interview-templates

**Source**: `4. Tasks.md`

### Task 1: Create Interview Template Handler — ✅ Complete

- `handlers/interview_templates.go`: `InterviewTemplateHandler`, request/response structs, `CreateInterviewTemplate` with validation, rubric existence check, config default {}.

### Task 2: Register POST /interview-templates Route — ✅ Complete

- `r.Post("/interview-templates", templateHandler.CreateInterviewTemplate)`.

### Task 3: Testing & Validation — ✅ Complete

- Success, validation (400), not found (404), server (500).

---

# Part 5: POST /interviews

## 5.1 PRD: POST /interviews

**Source**: `5. PRD-POST-interviews.md`

- **Method/Path**: POST `/interviews`
- **Request**: `interviewPlanId` (required), `teacherId`, `simulated` (default true), `studentName`, `status` (default "in_progress").
- **Response**: 201 Created with interview object (`interviewId`, `startedAt`, `completedAt` null initially).
- **Validation**: interviewPlanId valid and existing plan; status enum; teacherId valid if provided.
- **Errors**: 400, 404 (plan not found), 500.

---

## 5.2 Implementation Tasks: POST /interviews

**Source**: `5. Tasks.md`

### Task 1: Create Interview Handler — ✅ Complete

- `handlers/interviews.go`: `InterviewHandler`, `CreateInterviewRequest`/`InterviewResponse`, `CreateInterview` (and later `GetInterview`) with validation, plan existence check, defaults.

### Task 2: Register POST /interviews Route — ✅ Complete

- `r.Post("/interviews", interviewHandler.CreateInterview)`.

### Task 3: Testing & Validation — ✅ Complete

- Success, validation, not found, server; nullable fields; simulated/status defaults.

---

# Part 6: GET /interviews/{id}

## 6.1 PRD: GET /interviews/{id}

**Source**: `6. PRD-GET-interviews-id.md`

- **Method/Path**: GET `/interviews/{id}`
- **Path**: `id` (required, UUID).
- **Response**: 200 OK with interview object (nullable: teacherId, studentName, completedAt).
- **Errors**: 400 (invalid UUID), 404 (not found), 500.
- **Implementation**: `GetInterviewByID`; extract id via `chi.URLParam(r, "id")`.

---

## 6.2 Implementation Tasks: GET /interviews/{id}

**Source**: `6. Tasks.md`

### Task 1: Implement GET /interviews/{id} Handler — ✅ Complete

- Add `GetInterview` to `InterviewHandler`; parse/validate id; call `GetInterviewByID`; convert to `InterviewResponse`; 200/400/404/500.

### Task 2: Register GET /interviews/{id} Route — ✅ Complete

- `r.Get("/interviews/{id}", interviewHandler.GetInterview)`.

### Task 3: Testing & Validation — ✅ Complete

- Success, invalid UUID (400), not found (404), server (500); nullable fields.

---

# Part 7: Integration Test (Rubric + Template + Interview + Messages)

## 7.1 Implementation Tasks: Integration Test

**Source**: `7. Integration-test-1.md`

### Task 1: Set Up Test Infrastructure — ✅ Complete

- Test file, DB connection, helpers: `setupTestDB`, `teardownTestDB`, `createTestTeacher`.

### Task 2: Implement Test - Create Rubric — ✅ Complete

- Create teacher, then rubric via handler/API; assert 201, rubricId, store for next step.

### Task 3: Implement Test - Create Interview Template — ✅ Complete

- Use rubricId; create template; assert 201, interviewPlanId; store for next step.

### Task 4: Implement Test - Create Interview — ✅ Complete

- Use interviewPlanId; create interview; assert 201, interviewId, startedAt; store for next step.

### Task 5: Implement Test - Add Messages to Interview — ✅ Complete

- Use interviewId; create two messages (e.g. via `CreateInterviewMessage` or handler); assert created and linked.

### Task 6: Implement Test - Retrieve Interview and Assert Structure — ✅ Complete

- Get interview via `GetInterview`; assert full structure; get messages via `ListMessagesByInterview`; assert count and order.

### Task 7: Add Test Cleanup and Error Handling — ✅ Complete

- Defer cleanup (or rely on CASCADE); clear errors with `t.Fatalf`/`t.Errorf`; isolation; `t.Logf` for debugging.

### Task 8: Run and Validate Integration Test — ⏳ Ready to Run

- Run: `go test -v ./backend/internal/api/handlers -run TestCreateRubricTemplateInterviewFlow`; document DB and env requirements.

**Order**: 1 → 2 → 3 → 4 → 5 → 6 → 7 → 8.

---

# Quick Reference: Status Summary

| Area | Tasks | Status |
|------|--------|--------|
| Core API (deps, routes, handlers) | 1.Tasks 1–10 | Mix: 1–2 ✅; 3–10 ⚠️ |
| Create Teacher Account | 2.Tasks 1–7 | 1–6 ✅; 7 ⚠️ |
| GET /rubrics | 3.Tasks 1–3 | ✅ Complete |
| POST /interview-templates | 4.Tasks 1–3 | ✅ Complete |
| POST /interviews | 5.Tasks 1–3 | ✅ Complete |
| GET /interviews/{id} | 6.Tasks 1–3 | ✅ Complete |
| Integration test | 7.Tasks 1–8 | 1–7 ✅; 8 ⏳ Run |

---

# General Notes (All Handlers)

- Follow patterns in `handlers/rubrics.go`.
- Use pgtype for DB; convert to standard types for JSON.
- Validate UUIDs with `github.com/google/uuid`.
- Handle config/JSONB carefully; default config to `{}` where specified.
- Return appropriate HTTP status codes; user-friendly errors without exposing internals.
- Chi router: path params via `chi.URLParam(r, "id")`; query via `r.URL.Query().Get("teacherId")`.
