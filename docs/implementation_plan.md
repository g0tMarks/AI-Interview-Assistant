## Microviva – Implementation Sequencing Roadmap

**Purpose**: High-level, ordered roadmap for evolving Microviva into an AI-era assessment integrity platform centred on baseline writing, context-aware authorship analysis, and micro-vivas. This is intentionally coarse-grained (phases and major streams), not a task checklist.

---

## Phase 1 – Baseline & Authorship Core (Stage 1: Internal Validation)

1. **Student baseline writing collection**
   - Enable teachers to run supervised in-class writing tasks and capture samples per student. Writing tasks will be handwritten on paper, then scanned in and converted to structured data - possibly JSON, using OCR.
   - Make it easy to schedule, run, and manage these baseline activities within normal assessment workflows. Possibly provide a .pdf document that contains a template page for each student - with the student UUID, and clear isntructions to write in blue or black pen to make it easy to scan and digitise.
2. **Baseline writing profiles**
   - Build and store longitudinal writing fingerprints from baseline samples (stylistic markers, error patterns, argument structure, etc.).
   - Validate that these profiles are stable enough to act as a reference for later comparisons.
3. **Submission ingestion and comparison**
   - Allow teachers and students to submit new student work (file upload or LMS-style integrations later).
   - Implement comparison of submissions against baselines, conditioned on the chosen AI Assistance Level for the task.
4. **Authorship insight reports**
   - Generate structured, explainable authorship insight reports (observed differences, reasoning, context) rather than AI “probabilities”.
   - Test this loop with real classroom data to validate usefulness for teachers.

---

## Phase 2 – Micro-Viva Loop (Stage 2: Micro-Viva Integration)

5. **Targeted micro-viva question generation**
   - From suspicious or “interesting” submissions, generate short, focused viva questions that probe key parts of the work.
   - Respect AI Assistance Levels when framing questions (e.g. focus on reasoning, not just wording differences).
   - Create an option for teachers to select that enables micro-viva questions to be generated for every student.
6. **Micro-viva delivery and capture**
   - Provide a simple workflow for teachers to conduct 2–5 minute micro-vivas (text, audio, or in-person support with prompts).
   - Capture student explanations (transcripts or notes) and attach them to the relevant submission and baseline.
7. **Conceptual alignment assessment**
   - Analyse viva transcripts alongside the original submission and baseline to assess conceptual ownership and understanding.
   - Surface this as additional evidence in the authorship insight report, not as a separate or opaque score.

---

## Phase 3 – School Pilots (Stage 3)

8. **Pilot-ready workflows for teachers**
   - Smooth the end-to-end flow: set up classes, collect baselines, ingest submissions, review comparison reports, trigger micro-vivas, and record outcomes.
   - Minimise cognitive load and training required for pilot teachers.
9. **Policy and leadership alignment**
   - Ensure workflows, terminology, and outputs align with school academic integrity policies (avoid “AI detection” framing).
   - Provide clear, auditable records that faculty leaders and school leadership can understand and trust.
10. **Feedback loops from pilots**
   - Run targeted pilots with a small number of partner schools.
   - Collect feedback on usability, interpretability of evidence, and impact on teacher confidence; feed this back into product decisions.

---

## Phase 4 – Productisation (Stage 4) and Beyond

11. **Productisation of the platform**
   - Develop the minimal reporting and oversight views leadership needs (cohort-level patterns, trends, and usage), while keeping workflow first.
   - Add integrations, security features, and tenancy/identity options appropriate for school-scale deployment (e.g. SSO, data governance).
12. **Operational robustness and scaling**
   - Strengthen rate limiting, observability, monitoring, and privacy controls so the system can run reliably across multiple schools.
   - Codify migration/versioning strategies for schemas and models as baselines and analysis methods evolve.
13. **Long-term evolution of assessment integrity**
   - Iterate on baseline collection patterns, authorship reasoning techniques, and micro-viva designs as AI and assessment practices change.
   - Keep the product tightly aligned to the guiding principles in the master plan: teacher judgement central, evidence over probability, pedagogy and workflow first.

