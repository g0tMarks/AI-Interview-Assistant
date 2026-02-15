# Next Steps Plan – Interview Assistant

**Generated**: 2025-02-15  
**Purpose**: Align roadmap with codebase state and your ordered list. Use this to decide what to do next and in what order.

---

## 1. Current State vs Your List

### Already in place
| Area | Status |
|------|--------|
| **Schema** | `teachers`, `students`, `classes`, `roster`, `rubrics`, `rubric_criteria`, `interview_plans`, `interviews`, `interview_messages`, `interview_summaries`, `criterion_evidence`, curriculum/concept/misconception tables |
| **SQLC** | Configured; generated code for all existing tables including students/classes/roster |
| **Rubrics** | POST /rubrics, GET /rubrics?teacherId=uuid |
| **Teachers** | POST /teachers/register (no login/JWT) |
| **Interview templates** | POST /interview-templates (LLM-generated instructions from rubric) |
| **Interviews** | POST /interviews, GET /interviews/{id} (uses `student_name` text only; no `student_id`) |
| **interview_messages** | **Table + SQLC** (`CreateInterviewMessage`, `ListMessagesByInterview`) — **no HTTP endpoints** |
| **Summaries / criterion_evidence** | **Tables + SQLC** (Create/Get/Update summary, criterion_evidence) — **no HTTP endpoints** |
| **Integration test** | Teacher → rubric → template → interview → 2 messages via DB → GET interview; **does not** drive /next, engine, or results API |
| **Students / classes / roster** | CRUD handlers and routes (POST/GET/PATCH students, classes, roster) |
| **Student auth + JWT** | POST /auth/student/login (class code + email → JWT); RequireStudentAuth middleware; GET /student/me (protected) |

### Not present (from your list)
- **Uploads** or **file storage abstraction**
- **Text extraction** (PDF/DOCX → raw text)
- **Rubric parser** (raw → structured JSON) or **schema validation**
- **Rubric version editing** (teacher corrects parser) — no UpdateRubric / PATCH rubric
- **Interview messages HTTP API** (POST/GET messages for an interview)
- **Interview engine v1** and **GET/POST /interviews/{id}/next**
- **Final evaluation + results endpoint** and stored scoring JSON API
- **Golden-path integration test** covering full flow (through /next and results)
- **Rate limits** or **prompt injection hardening**
- **Bulk interview creation** for a class
- **Results listing/export** for teacher
- **Voice (push-to-talk) + STT**
- **Microsoft Entra OIDC SSO**

---

## 2. Recommended Order (Matches Your List + Dependencies)

Do these in sequence so each step has the right foundation.

| # | Step | Notes / dependencies |
|---|------|----------------------|
| **1** | **Students / classes / roster tables + SQLC + CRUD** | Add `app.students`, `app.classes`, `app.roster` (e.g. class_id, student_id, role). Run sqlc generate; add queries and CRUD handlers. Optionally link `app.interviews.student_id` to `app.students` later. |
| **2** | **Student auth MVP (magic link or class code) + JWT middleware** | Magic link (email token) or class-code flow; issue JWT. Add middleware to validate JWT and set identity on context. Protect student-facing routes. |
| **3** | **Uploads + file storage abstraction** | Multipart upload endpoint(s); abstraction (e.g. interface) for store (local/S3). Used by rubric file upload and later by other assets. |
| **4** | **Text extraction for PDF/DOCX → raw text** | Use uploads + storage; extract text; store in `rubrics.raw_text` (or temp) and return in API. Library: e.g. unidoc (PDF), gooxml or similar (DOCX). |
| **5** | **Rubric parser (raw → structured JSON) + schema validation** | Parse raw text → criteria (name, description, weight, levels). Define JSON schema; validate output. Can be LLM-based or rule-based; store result as rubric_criteria. |
| **6** | **Rubric version editing endpoint** | PATCH /rubrics/{id} and/or PATCH /rubrics/{id}/criteria (or replace criteria). Teacher can fix parser mistakes. Requires UpdateRubric / update criteria in SQLC if not present. |
| **7** | **Interview_messages table + endpoints** | Table and SQLC exist. Add: POST /interviews/{id}/messages, GET /interviews/{id}/messages. Used by engine and frontend. |
| **8** | **Interview engine v1 + /interviews/{id}/next** | Implement “next question / next step” logic (from plan + branches + messages); optional LLM for classification. Expose as POST /interviews/{id}/next (and/or GET for idempotent “current next”). |
| **9** | **Final evaluation + results endpoint + stored scoring JSON** | After interview completion, run evaluation (LLM or rules) → fill `interview_summaries` + `criterion_evidence`; store scoring JSON (e.g. in summary or dedicated column). Add GET /interviews/{id}/results (and optionally GET /interviews/{id}/summary). |
| **10** | **Golden-path integration test** | Single test: create teacher → (optional class/student) → rubric → template → interview → call /next until done → trigger evaluation → GET results; assert status, summary, and scoring shape. |
| **11** | **Rate limits + prompt injection hardening** | Global or per-route rate limits; sanitize/validate user content before sending to LLM and in storage. |
| **12** | **Bulk interview creation for a class** | Endpoint (e.g. POST /classes/{id}/interviews/bulk) using plan + roster to create N interviews (one per student or selected list). Depends on classes/roster. |
| **13** | **Results listing/export for teacher** | List results by teacher/class/interview plan; export (CSV/JSON). Depends on results endpoint and optionally on classes. |
| **14** | **Voice (push-to-talk) + STT** | Push-to-talk UI; send audio to STT; feed transcript into interview (e.g. as user messages). Backend: STT integration and possibly WebSocket or chunked HTTP. |
| **15** | **Microsoft Entra OIDC SSO** | Replace or complement teacher (and optionally student) auth with Entra OIDC; map identity to teachers (and students if applicable). |

---

## 3. What to Do Next (Concrete)

**Step #1 – Students / classes / roster + SQLC + CRUD** — ✅ **Done**

**Step #2 – Student auth MVP + JWT middleware** — ✅ **Done**

1. **Auth package** (`backend/internal/auth/`): `jwt.go` (IssueStudentToken, ValidateStudentToken, StudentClaims), `context.go` (WithStudentID, StudentIDFromContext).
2. **Middleware** (`backend/internal/api/middleware/auth.go`): `RequireStudentAuth(jwtSecret)` — validates `Authorization: Bearer <token>` and sets student ID on context.
3. **Auth handler** (`backend/internal/api/handlers/auth.go`): `POST /auth/student/login` — body `{ "classCode", "email" }`; student must exist and be on class roster; returns `{ "token": "<jwt>" }`.
4. **Protected route**: `GET /student/me` — requires valid student JWT; returns current student profile.
5. **Config**: `JWT_SECRET` env (defaults to dev placeholder if unset). Dependencies include `JWTSecret`.

**Usage:** Set `JWT_SECRET` in production. Student flow: teacher creates class (gets `class_code`), adds students to roster; student calls POST /auth/student/login with class code + email, then uses the token in `Authorization: Bearer <token>` for GET /student/me and future student routes.

After that, proceed in order: **#3** (uploads), then **#4**, and so on.

---

## 4. Optional: Link interviews to students

Once students and roster exist, consider:
- Add `student_id UUID REFERENCES app.students(student_id)` to `app.interviews` (nullable for backward compatibility).
- When creating an interview for a class, set `student_id` from roster; keep `student_name` as optional display override.

---

## 5. File Reference

| Topic | Files |
|-------|--------|
| Schema | `backend/schema/schema.sql` |
| SQLC config | `backend/schema/sqlc.yaml` |
| Queries | `backend/internal/db/queries/*.sql` |
| Router | `backend/internal/api/router.go` |
| Server deps | `backend/internal/api/server.go` |
| Student auth | `backend/internal/auth/`, `backend/internal/api/middleware/auth.go`, `backend/internal/api/handlers/auth.go` |
| Integration test | `backend/internal/api/handlers/integration_test.go` |

Use this plan as the single checklist; update the “Current state” section as you complete each item.
