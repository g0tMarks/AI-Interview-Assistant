# Microviva — User Journeys

This document maps the core user journeys for Microviva across three primary user groups:

- Teachers
- Students
- School administrators

It is intended to clarify:

- step-by-step flows
- navigation and screen progression
- feature usage across roles
- where key product decisions appear in the experience

This is a product-facing document, not a technical spec. For implementation details, see `architecture.md`, `next-steps-plan.md`, and later `design_guidelines.md`.

---

# 1. Product Roles

## Teachers
Teachers create classes, configure assessments, upload baseline and assessment artefacts, run authentication analysis, review insights, and optionally conduct or launch micro-vivas.

## Students
Students log in, view assigned interviews or viva tasks, submit work where relevant, and complete micro-viva conversations.

## School Administrators
School administrators manage school-level settings, staff access, policy defaults, assessment governance, and reporting.

---

# 2. Core Product Navigation Model

At a high level, the product navigation should centre around role-specific home dashboards.

## Teacher Navigation
Primary areas:

- Dashboard
- Classes
- Assessments
- Baselines
- authentication Reviews
- Interviews / Micro-vivas
- Reports
- Settings

## Student Navigation
Primary areas:

- Home
- My Tasks
- My Interviews
- My Submissions
- Support / Instructions
- Profile

## School Administrator Navigation
Primary areas:

- Dashboard
- Staff
- Classes
- School Policy
- Assessment Settings
- Reporting
- Integrations
- Settings

---

# 3. Teacher User Journeys

# Journey T1 — Teacher Registers and Creates Their First Class

## Goal
A new teacher creates an account, sets up their workspace, and creates the first class.

## Entry Point
Landing page → Sign up

## Flow

### Step 1 — Register
Teacher enters:

- name
- school email
- password or SSO

System creates teacher account.

### Step 2 — Verify account
Teacher verifies email or signs in via approved SSO provider.

### Step 3 — First-time onboarding
Teacher sees a short onboarding flow:

- What subjects do you teach?
- Which year levels do you teach?
- Do you want to start with interviews, authentication, or both?
- Are you part of an existing school workspace?

### Step 4 — Arrive at Teacher Dashboard
Dashboard should immediately show the next recommended action:

- Create your first class
- Import students
- Create your first assessment
- Upload baseline writing

### Step 5 — Create class
Teacher clicks `Create Class` and enters:

- class name
- year level
- subject
- optional class code

### Step 6 — Add students
Teacher chooses one of:

- manual add
- bulk roster upload
- school sync / integration
- invite via class code

### Step 7 — Class created
Teacher lands on Class Overview.

## Navigation Notes
After registration, the system should always drive teachers toward one meaningful first action. The product should not leave them on an empty dashboard with no direction.

---

# Journey T2 — Teacher Creates an Assessment

## Goal
Teacher creates an assessment that can later be used for authentication comparison, interviews, or both.

## Entry Point
Teacher dashboard or class page → `Create Assessment`

## Flow

### Step 1 — Start assessment setup
Teacher selects class and clicks `Create Assessment`.

### Step 2 — Enter core details
Teacher enters:

- assessment title
- subject
- year level
- due date
- assessment type
- optional description

### Step 3 — Select assessment mode
Teacher chooses one or more:

- Writing baseline only
- authentication comparison
- Micro-viva follow-up
- Rubric-driven interview

### Step 4 — Set AI Assistance Level
Teacher selects where the task sits on the AI assistance spectrum.

This step is critical.

Teacher chooses the level that best matches the task design, for example:

- No AI
- AI for planning
- AI for drafting support
- AI collaboration
- AI exploration / evaluation

### Step 5 — Add rubric or instructions
Teacher either:

- uploads rubric file
- pastes raw rubric text
- creates rubric manually
- skips for later

### Step 6 — Review summary
Teacher sees a summary screen:

- assessment details
- AI Assistance Level
- class attached
- workflow enabled
- rubric status

### Step 7 — Publish assessment
Teacher publishes or saves as draft.

## Navigation Notes
This flow should live primarily under `Assessments`, but also be accessible from a class page. AI Assistance Level must be framed as part of assessment design, not as a compliance checkbox.

---

# Journey T3 — Teacher Builds Student Baselines

## Goal
Teacher creates supervised baseline writing profiles for students.

## Entry Point
Class page or assessment page → `Build Baselines`

## Flow

### Step 1 — Choose baseline source
Teacher selects one or more supervised samples:

- in-class writing task
- handwritten scan converted to text
- previous verified classroom writing

### Step 2 — Upload or confirm samples
Teacher uploads files or confirms already stored writing samples.

### Step 3 — Mark verification status
Teacher confirms that these are trusted baseline samples.

### Step 4 — Assign semester / profile window
Teacher selects the relevant semester or time period for the baseline.

### Step 5 — Generate student profiles
System processes each student’s baseline samples and creates writing profiles.

### Step 6 — Review baseline completeness
Teacher sees status by student:

- Ready
- Needs more samples
- Low-confidence baseline
- No verified baseline

### Step 7 — Confirm baseline set
Teacher confirms profile generation.

## Navigation Notes
This should be available under both `Baselines` and individual class/assessment flows. Teachers should be able to build baselines once, then reuse them across multiple assessments in the same semester.

---

# Journey T4 — Teacher Uploads or Reviews Student Submission for authentication Analysis

## Goal
Teacher compares a submitted assessment against a student baseline.

## Entry Point
Assessment page → `Run Authentication Analysis`

## Flow

### Step 1 — Open assessment
Teacher enters the relevant assessment workspace.

### Step 2 — View submission list
Teacher sees students and status:

- submitted
- baseline ready
- no baseline
- analysis pending
- reviewed

### Step 3 — Select a submission
Teacher opens an individual student submission or bulk-selects multiple submissions.

### Step 4 — Confirm context
System shows:

- assessment title
- AI Assistance Level
- baseline profile used
- submission date
- teacher-selected settings

### Step 5 — Run analysis
System compares:

- submission text
- baseline profile
- AI Assistance Level expectations

### Step 6 — View structured result
Teacher sees a report with sections such as:

- overall discrepancy level
- observed differences
- context-aware interpretation
- suggested next step

Examples of recommendations:

- No action needed
- Reflection suggested
- Micro-viva recommended

### Step 7 — Decide what to do next
Teacher chooses:

- mark as reviewed
- request reflection
- launch micro-viva
- add notes
- escalate for follow-up

## Navigation Notes
The analysis view must feel like an assessment workflow, not a forensic dashboard. Teachers should be guided toward action, not buried in raw metrics.

---

# Journey T5 — Teacher Launches a Micro-viva Follow-up

## Goal
Teacher uses the platform to initiate a targeted oral follow-up after authentication analysis.

## Entry Point
Authentication analysis result → `Launch Micro-viva`

## Flow

### Step 1 — Review recommendation
Teacher sees why a viva is recommended.

### Step 2 — Generate viva questions
System generates targeted questions based on:

- submission content
- baseline comparison
- rubric or task expectations
- AI Assistance Level

### Step 3 — Review / edit questions
Teacher can:

- accept generated questions
- edit wording
- remove questions
- add their own

### Step 4 — Choose delivery method
Teacher chooses:

- live in-class micro-viva
- student self-completed oral response
- scheduled teacher-led follow-up

### Step 5 — Assign or start viva
If asynchronous, system sends task to student.
If live, teacher begins session immediately.

### Step 6 — Record outcomes
System stores:

- transcript or response
- conceptual alignment notes
- teacher observations
- outcome recommendation

### Step 7 — Final teacher decision
Teacher records one of:

- authentic confirmed
- authentic with support
- follow-up required
- concern unresolved

## Navigation Notes
Micro-viva should sit naturally inside the authentication review journey, not as a separate disconnected module.

---

# Journey T6 — Teacher Creates and Runs a Rubric-driven Interview

## Goal
Teacher creates a full interview flow from a rubric.

## Entry Point
Assessments or Interviews → `Create Interview`

## Flow

### Step 1 — Upload or select rubric
Teacher uploads rubric or selects an existing one.

### Step 2 — Parse rubric
System extracts criteria and suggested question plan.

### Step 3 — Review and edit
Teacher edits:

- criteria
- interview instructions
- questions
- branch logic

### Step 4 — Create interview template
Teacher saves the plan as a reusable interview template.

### Step 5 — Assign interview
Teacher selects:

- one student
- a group
- entire class

### Step 6 — Student completes interview
Student answers questions through the interview interface.

### Step 7 — Review results
Teacher views:

- transcript
- evaluation summary
- criterion evidence
- suggested next steps

## Navigation Notes
This journey is broader than the authentication flow. It belongs under `Interviews`, but should also be connected to specific assessments where relevant.

---

# Journey T7 — Teacher Reviews Results Across a Class

## Goal
Teacher gets a class-wide view of progress, authentication reviews, and follow-up needs.

## Entry Point
Class page or Reports → `View Results`

## Flow

### Step 1 — Open class results
Teacher chooses class and assessment.

### Step 2 — View summary table
Teacher sees per-student status, for example:

- submission received
- baseline ready
- authentication reviewed
- viva completed
- final decision recorded

### Step 3 — Filter results
Teacher filters by:

- no baseline
- analysis pending
- viva recommended
- review unresolved
- complete

### Step 4 — Open student record
Teacher clicks into any student to review details.

### Step 5 — Export or report
Teacher exports results or downloads a teacher summary.

## Navigation Notes
Teachers need fast triage. The class-level reporting view should help them move from overview to action quickly.

---

# 4. Student User Journeys

# Journey S1 — Student Logs In for the First Time

## Goal
Student accesses the platform and reaches their task dashboard.

## Entry Point
Student invitation, class code, direct school link, or SSO

## Flow

### Step 1 — Start sign-in
Student chooses login method:

- school SSO
- class code + email
- magic link

### Step 2 — Verify identity
System verifies student membership in the relevant class or school.

### Step 3 — First-time welcome
Student sees a simple orientation:

- what Microviva is
- what they may be asked to do
- how interviews and micro-vivas work
- privacy / expectations summary

### Step 4 — Arrive at Student Home
Home shows:

- upcoming tasks
- interviews due
- micro-viva follow-ups
- completed tasks

## Navigation Notes
Student navigation must remain minimal and calm. Students should not feel they are entering an investigative system.

---

# Journey S2 — Student Completes a Micro-viva Task

## Goal
Student completes a short oral or written explanation task linked to a submission.

## Entry Point
Student Home or My Tasks → `Complete Micro-viva`

## Flow

### Step 1 — Open assigned task
Student sees:

- assessment title
- instructions
- due date
- expected response type

### Step 2 — Read explanation
Student is told that the micro-viva is a short opportunity to explain their thinking.

### Step 3 — Start viva
Student begins one of:

- text response
- audio response
- interactive interview

### Step 4 — Answer questions
Student responds to the targeted questions.

### Step 5 — Submit
System confirms completion.

### Step 6 — Return to dashboard
Task status updates to complete.

## Navigation Notes
The tone here matters. The product language should communicate reflection and explanation, not suspicion.

---

# Journey S3 — Student Completes a Full Interview

## Goal
Student completes a rubric-driven interview.

## Entry Point
My Tasks or My Interviews → `Start Interview`

## Flow

### Step 1 — Open interview
Student sees the task title and expected duration.

### Step 2 — Begin interview
System presents question one.

### Step 3 — Respond in sequence
Student answers questions; system advances based on branching logic.

### Step 4 — Finish interview
System confirms completion.

### Step 5 — Return to home
Interview appears as completed.

## Navigation Notes
This flow should be highly focused, with minimal distractions and very clear progress cues.

---

# Journey S4 — Student Views Their Task History

## Goal
Student can see what they have completed and what remains outstanding.

## Entry Point
My Tasks or Home

## Flow

### Step 1 — Open task list
Student sees grouped sections:

- To do
- In progress
- Completed

### Step 2 — Open an item
Student views instructions or prior completion status.

### Step 3 — Return to dashboard
Student navigates back to Home or My Tasks.

## Navigation Notes
Students should never see teacher-facing authentication analysis or internal review language.

---

# 5. School Administrator User Journeys

# Journey A1 — Administrator Sets Up School Workspace

## Goal
Administrator configures the school environment.

## Entry Point
Admin invitation or school onboarding

## Flow

### Step 1 — Register or accept invite
Administrator creates account or joins workspace.

### Step 2 — Verify school ownership / authority
System verifies school domain or approval path.

### Step 3 — Configure school profile
Administrator enters:

- school name
- domain
- year level range
- branding
- contact details

### Step 4 — Set authentication method
Administrator configures:

- SSO
- staff login defaults
- student login method

### Step 5 — Invite staff
Administrator invites teachers and leaders.

### Step 6 — Review setup checklist
System shows setup progress:

- staff invited
- classes synced or ready
- policy defaults set
- integrations configured

## Navigation Notes
The administrator onboarding should feel operational and implementation-focused, not pedagogical.

---

# Journey A2 — Administrator Configures School-wide AI and Assessment Policy Defaults

## Goal
Administrator defines default settings that guide how teachers configure assessments.

## Entry Point
School Policy → `Assessment Defaults`

## Flow

### Step 1 — Open policy settings
Administrator enters policy configuration.

### Step 2 — Define AI Assistance defaults
Administrator sets school-wide options, such as:

- allowed AI Assistance Levels
- naming conventions
- recommended defaults by year level
- whether teachers can override defaults

### Step 3 — Define review workflow defaults
Administrator configures:

- whether teacher review is required after analysis
- whether viva recommendations are automatic or optional
- whether administrator visibility is enabled for unresolved cases

### Step 4 — Save settings
Defaults are applied across the workspace.

## Navigation Notes
This is where research-informed policy meets product controls. It must be powerful but understandable.

---

# Journey A3 — Administrator Monitors Adoption and Usage

## Goal
Administrator tracks whether the school is actually using the system.

## Entry Point
Admin Dashboard

## Flow

### Step 1 — Open dashboard
Administrator sees school-level overview.

### Step 2 — Review adoption metrics
Examples:

- active teachers
- active classes
- assessments created
- baselines generated
- authentication reviews completed
- micro-vivas completed

### Step 3 — Review risk and support signals
Examples:

- classes with no baseline usage
- unresolved follow-ups
- heavy manual review workload

### Step 4 — Take action
Administrator can:

- contact staff
- provide support
- export report
- update defaults

## Navigation Notes
This view should support school implementation and governance, not disciplinary surveillance.

---

# Journey A4 — Administrator Reviews School-level Reporting

## Goal
Administrator can produce summaries for leadership, governance, or implementation reviews.

## Entry Point
Reporting

## Flow

### Step 1 — Select reporting scope
Administrator selects:

- school-wide
- faculty
- year level
- date range

### Step 2 — Generate report
System prepares selected metrics and summaries.

### Step 3 — Export
Administrator exports to PDF, CSV, or internal summary view.

## Navigation Notes
Reports should focus on adoption, workflow completion, and policy alignment rather than student-level accusations.

---

# 6. Shared Workflow Journeys

# Journey X1 — Teacher Invites Student into a Follow-up Workflow

## Goal
Bridge the teacher and student experiences cleanly.

## Flow

### Step 1 — Teacher records follow-up requirement
From analysis result, teacher clicks `Request Follow-up`.

### Step 2 — Student receives task
Student dashboard updates with a new required item.

### Step 3 — Student completes task
Student finishes reflection, viva, or interview.

### Step 4 — Teacher reviews completion
Teacher receives notification and opens result.

### Step 5 — Teacher records final decision
Case is resolved.

---

# Journey X2 — School-wide Rollout

## Goal
Show the likely multi-user sequence during a real school launch.

## Flow

### Step 1 — Administrator sets up school workspace
### Step 2 — Administrator invites teachers
### Step 3 — Teachers create classes and import students
### Step 4 — Teachers create assessments
### Step 5 — Teachers upload or confirm baseline writing
### Step 6 — Students submit work or complete interviews
### Step 7 — Teachers review authentication analysis
### Step 8 — Selected students complete micro-vivas
### Step 9 — Teachers record decisions
### Step 10 — Administrator monitors adoption and policy alignment

---

# 7. Recommended Product Structure by Screen

## Teacher Screens

- Teacher Dashboard
- Class List
- Class Overview
- Assessment List
- Assessment Setup
- Baseline Builder
- Submission Review
- authentication Analysis
- Micro-viva Builder
- Interview Template Editor
- Results / Reports
- Teacher Settings

## Student Screens

- Student Home
- Task List
- Task Detail
- Interview Screen
- Micro-viva Screen
- Submission Status
- Student Profile

## Administrator Screens

- Admin Dashboard
- Staff Management
- School Policy
- Assessment Defaults
- Reporting
- Integrations
- Workspace Settings

---

# 8. Product Sequencing Recommendations

For MVP, not every journey needs to be fully built at once.

## MVP Priority Journeys

### Teacher
- T1 Register and create class
- T2 Create assessment
- T3 Build baseline
- T4 Run authentication analysis
- T5 Launch micro-viva
- T7 Review class results

### Student
- S1 First login
- S2 Complete micro-viva
- S3 Complete interview

### Administrator
- A1 School setup
- A2 Policy defaults

These journeys define the core value loop.

---

# 9. Core Value Loop

The core value loop for Microviva should be clear across the whole product:

1. Teacher sets up class and assessment
2. Teacher defines AI Assistance Level
3. Teacher builds or confirms baseline writing
4. Student work is submitted or uploaded
5. System compares submission to baseline in context
6. Teacher receives structured insight
7. Student completes micro-viva if needed
8. Teacher records judgement
9. School gains a defensible authenticity workflow

If the product does this well, it will feel coherent.

If it does not, it will feel like disconnected tools.

---

# 10. Design Intent

The experience of Microviva should feel:

- calm
- professional
- pedagogically grounded
- transparent
- supportive of teacher judgement

It should not feel like:

- a plagiarism detector
- a policing system
- an AI surveillance product

The user journeys above should be used to guide navigation, feature prioritisation, and screen design decisions.
