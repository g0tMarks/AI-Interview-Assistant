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
| **Interviews** | POST /interviews, GET /interviews/{id} (supports `student_id` linked to `app.students`; `student_name` remains optional display override) |
| **interview_messages** | **Table + SQLC** (`CreateInterviewMessage`, `ListMessagesByInterview`) — **no HTTP endpoints** |
| **Summaries / criterion_evidence** | **Tables + SQLC** (Create/Get/Update summary, criterion_evidence) — **no HTTP endpoints** |
| **Integration test** | Teacher → rubric → template → interview → 2 messages via DB → GET interview; **does not** drive /next, engine, or results API |
| **Students / classes / roster** | CRUD handlers and routes (POST/GET/PATCH students, classes, roster) |
| **Student auth + JWT** | POST /auth/student/login (class code + email → JWT); RequireStudentAuth middleware; GET /student/me (protected) |
| **Uploads (local storage)** | POST /uploads (multipart file); GET /uploads/{key} (download). Local disk store under `UPLOADS_DIR`. |
| **Bulk student roster upload (.xlsx)** | POST /classes/{id}/roster/upload (multipart form with .xlsx file). Parses Excel with first name, last name, email columns; creates students or matches by email; adds to roster. Returns summary (created, added, skipped, errors). Uses excelize library. |
| **Text extraction** (PDF/DOCX → raw text) | POST /rubrics/upload (multipart form with PDF/DOCX file). Extracts text using **ledongthuc/pdf** (PDF, BSD-3-Clause) and **mydocx** (DOCX, MIT); no commercial license required. Stores in `rubrics.raw_text` and returns in API response. |

### Not present (from your list)
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
| **4** | **Bulk student roster upload (.xlsx)** | Teacher uploads .xlsx with columns: first name, last name, email. Parse file (e.g. excelize), create students or match by email, add all to specified class roster. Endpoint e.g. POST /classes/{id}/roster/upload. Depends on uploads (multipart) and classes/roster. |
| **5** | **Text extraction for PDF/DOCX → raw text** | ✅ **Done** — Use uploads + storage; extract text; store in `rubrics.raw_text` and return in API. PDF: ledongthuc/pdf; DOCX: mydocx (no commercial license). |
| **6** | **Rubric parser (raw → structured JSON) + schema validation** | Parse raw text → criteria (name, description, weight, levels). Define JSON schema; validate output. Can be LLM-based or rule-based; store result as rubric_criteria. |
| **7** | **Rubric version editing endpoint** | PATCH /rubrics/{id} and/or PATCH /rubrics/{id}/criteria (or replace criteria). Teacher can fix parser mistakes. Requires UpdateRubric / update criteria in SQLC if not present. |
| **8** | **Interview_messages table + endpoints** | Table and SQLC exist. Add: POST /interviews/{id}/messages, GET /interviews/{id}/messages. Used by engine and frontend. |
| **9** | **Interview engine v1 + /interviews/{id}/next** | Implement “next question / next step” logic (from plan + branches + messages); optional LLM for classification. Expose as POST /interviews/{id}/next (and/or GET for idempotent “current next”). |
| **10** | **Final evaluation + results endpoint + stored scoring JSON** | After interview completion, run evaluation (LLM or rules) → fill `interview_summaries` + `criterion_evidence`; store scoring JSON (e.g. in summary or dedicated column). Add GET /interviews/{id}/results (and optionally GET /interviews/{id}/summary). |
| **11** | **Golden-path integration test** | Single test: create teacher → (optional class/student) → rubric → template → interview → call /next until done → trigger evaluation → GET results; assert status, summary, and scoring shape. |
| **12** | **Rate limits + prompt injection hardening** | Global or per-route rate limits; sanitize/validate user content before sending to LLM and in storage. |
| **13** | **Bulk interview creation for a class** | Endpoint (e.g. POST /classes/{id}/interviews/bulk) using plan + roster to create N interviews (one per student or selected list). Depends on classes/roster. |
| **14** | **Results listing/export for teacher** | List results by teacher/class/interview plan; export (CSV/JSON). Depends on results endpoint and optionally on classes. |
| **15** | **Voice (push-to-talk) + STT** | Push-to-talk UI; send audio to STT; feed transcript into interview (e.g. as user messages). Backend: STT integration and possibly WebSocket or chunked HTTP. |
| **16** | **Microsoft Entra OIDC SSO** | Replace or complement teacher (and optionally student) auth with Entra OIDC; map identity to teachers (and students if applicable). |

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

**Step #3 – Uploads + file storage abstraction** — ✅ **Done**

1. **Storage package** (`backend/internal/storage/`): `Store` interface + `LocalStore` implementation (writes `<key>.data` + `<key>.json`).
2. **Endpoints**: `POST /uploads` (multipart field `file`), `GET /uploads/{key}` (download).
3. **Config**: `UPLOADS_DIR` (default `./uploads`), `UPLOADS_MAX_BYTES` (default 25 MiB), injected via `api.Dependencies`.
4. **API test**: `backend/api_test/test-uploads.sh`.

After that, proceed in order: **#4** (bulk student roster upload), then **#5** (text extraction), and so on.

**Step #4 – Bulk student roster upload (.xlsx)** — ✅ **Done**

1. **Handler** (`backend/internal/api/handlers/roster.go`): `UploadRoster` method — parses multipart form, validates .xlsx file, reads Excel using excelize library.
2. **Column detection**: Automatically detects columns (supports variations: "first name"/"firstname"/"first_name"/"fname", "last name"/"lastname"/"last_name"/"lname", "email"/"e-mail"/"email address").
3. **Student processing**: For each row, gets existing student by email or creates new one; adds to class roster (skips if already in roster).
4. **Error handling**: Validates file format, handles missing columns, invalid emails, duplicates gracefully.
5. **Response**: Returns summary JSON with `createdCount`, `addedToRosterCount`, `skippedCount`, `errorCount`, and `errors` array.
6. **Route**: `POST /classes/{id}/roster/upload` added to router.
7. **Dependency**: Added `github.com/xuri/excelize/v2` to go.mod.
8. **Test script**: `backend/api_test/test-roster-upload.sh` for testing.

**Usage:** POST multipart form to `/classes/{class-id}/roster/upload` with `file` field containing .xlsx file with header row (first name, last name, email) and data rows.

---

## 4. Link interviews to students — ✅ **Done**

**Completed**: 2025-02-19

1. **Schema update** (`backend/schema/schema.sql`): Added `student_id UUID REFERENCES app.students(student_id) ON DELETE SET NULL` to `app.interviews` table (nullable for backward compatibility).

2. **SQL queries** (`backend/internal/db/queries/interviews.sql`): Updated `CreateInterview` query to include `student_id` in INSERT statement.

3. **Generated code**: Regenerated sqlc code; `AppInterview` model and queries now include `student_id` field.

4. **API handler** (`backend/internal/api/handlers/interviews.go`):
   - Updated `CreateInterviewRequest` to accept optional `classId` and `studentId` fields.
   - Added logic to verify student is in class roster when both `classId` and `studentId` are provided.
   - Updated `InterviewResponse` to include `studentId` field.
   - Updated both `CreateInterview` and `GetInterview` handlers to handle `studentId` in responses.

**Behavior**:
- If both `classId` and `studentId` are provided: verifies student is in the class roster before creating interview.
- If only `studentId` is provided: uses it directly (no roster verification).
- If neither is provided: `student_id` remains null (backward compatible).
- `studentName` remains optional as a display override.

**Usage**: When creating an interview for a class, include both `classId` and `studentId` in the request body. The API will verify the student is enrolled in the class before creating the interview.

**Step #5 – Text extraction for PDF/DOCX → raw text** — ✅ **Done**

**Completed**: 2025-02-19 (updated 2025-02-19 with license-free libraries)

1. **Dependencies** (no commercial license required):
   - **PDF**: `github.com/ledongthuc/pdf` (BSD-3-Clause). Uses a temp file and `GetPlainText()` for extraction.
   - **DOCX**: `github.com/xavier268/mydocx` (MIT). Uses `ExtractTextBytes()` for in-memory extraction.
   - UniDoc (unipdf/unioffice) was removed; it required a commercial license and was causing "license required" or empty text.

2. **Extraction package** (`backend/internal/extraction/extraction.go`):
   - `ExtractTextFromPDF()` - Writes upload to a temp `.pdf` file, opens with ledongthuc/pdf, extracts plain text, removes temp file.
   - `ExtractTextFromDOCX()` - Extracts text from DOCX using mydocx (paragraphs/tables flattened to a single string).
   - `ExtractText()` - Detects format from content type or filename and calls the appropriate extractor.
   - Error handling for unsupported formats and empty documents; PDF empty-text error suggests image-only/OCR.

3. **API handler** (`backend/internal/api/handlers/rubrics.go`):
   - `UploadRubricFile` accepts multipart form: `file` (PDF/DOCX), `teacherId` (required), `title` (optional), `description` (optional).
   - Extracts text, creates rubric with `rubrics.raw_text`, returns created rubric including `rawText`.

4. **Route**: `POST /rubrics/upload` registered in router.

5. **Test script**: `backend/api_test/test-rubric-upload.sh` — uploads a file, optionally registers a teacher, shows rawText length and preview.

**Usage**: POST multipart form to `/rubrics/upload` with `file` (PDF or DOCX), `teacherId`, and optional `title` and `description`. The API extracts text, creates a rubric, and returns it with the extracted text in `rawText`.

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
| Roster upload | `backend/internal/api/handlers/roster.go` (UploadRoster method), `backend/api_test/test-roster-upload.sh` |
| Text extraction | `backend/internal/extraction/extraction.go`, `backend/internal/api/handlers/rubrics.go` (UploadRubricFile), `backend/api_test/test-rubric-upload.sh` |
| Integration test | `backend/internal/api/handlers/integration_test.go` |

Use this plan as the single checklist; update the “Current state” section as you complete each item.
