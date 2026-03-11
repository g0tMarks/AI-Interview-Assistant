## Microviva — AI-Era Assessment Integrity Platform

Microviva is an AI-assisted assessment platform that helps educators verify student understanding and authenticate their thinking in a world where generative AI can easily produce high‑quality written work.

Instead of trying to **detect AI usage**, Microviva focuses on **evidence of thinking**. It enables teachers to preset the level of AI they allow in each assessment, then combines baseline writing analysis, context‑aware reasoning, and targeted micro‑viva conversations so teachers can confidently judge whether submitted work genuinely reflects a student’s understanding.

---

## Why This Exists

- **Core problem**: Generative AI means that for teachers to assess a students thinking, it has to be done in class on paper. No digital work produced by students can be a trustworthy representation of their ability. Existing tools cannot reliably answer: **“Does this work genuinely represent this student’s thinking?”**
- **Limitations of AI detectors**:
  - Unreliable and easy to game
  - Produce harmful false positives
  - Cannot prove authorship
- **Microviva’s reframing**: Shift from “Was AI used?” to **“Is this work consistent with this student’s demonstrated thinking, given the allowed level of AI assistance?”**

---

## Who Microviva Is For

- **Teachers**: Verify understanding, investigate suspicious submissions, and assess out‑of‑class work without becoming AI experts.
- **Faculty leaders**: Scale integrity practices across classes and align with assessment policy.
- **School leadership**: Maintain trust in grades and adapt assessment practices to the AI era without blanket bans on AI tools.

Target context: **secondary schools (Years 7–12)**.

---

## Core Concepts

- **AI Assistance Levels**  
  Tasks explicitly declare an AI assistance level (e.g. No AI, AI for planning only, AI‑assisted drafting, AI co‑creation). Authorship analysis is interpreted in this context rather than assuming all AI use is misconduct.

- **Writing Baseline Profiles**  
  Students complete supervised in‑class writing tasks. Microviva builds a longitudinal baseline “fingerprint” capturing sentence structure, vocabulary, stylistic markers, error patterns, and argument structure.

- **Context‑Aware Authorship Analysis**  
  New submissions are compared against each student’s baseline and the allowed AI assistance level. The system produces **structured explanations of differences**, not opaque “AI probability” scores.

- **Micro‑Viva Conversations**  
  When discrepancies matter, Microviva generates short, targeted oral questions (2–5 minutes) that help teachers confirm conceptual understanding instead of running adversarial interrogations.

---

## What Success Looks Like

- **Teachers regain confidence in assessment**  
  They can say: “I know this student understands their work.”

- **AI becomes a normal part of learning, not a threat**  
  Schools adopt verification workflows instead of AI bans and detection arms races.

- **Micro‑vivas become routine**  
  Short oral follow‑ups become a standard tool for authentic assessment.

- **Insights are explainable and defensible**  
  Outputs are transparent, evidence‑based, and always subordinate to teacher judgement.

---

## Guiding Principles

- **Teacher judgement is central** – Microviva surfaces evidence; it does not make disciplinary decisions.
- **Avoid “AI detection” framing** – Focus on authorship consistency and conceptual understanding.
- **Evidence over probability** – Explain observed differences and reasoning, not black‑box scores.
- **Pedagogy first** – Encourage baseline writing, reflective assessment, and oral explanation.
- **Workflow over dashboards** – Prioritise tools that fit real assessment workflows over vanity analytics.

---

## High-Level Architecture

At a high level, the system has three layers:

- **Frontend**: Next.js web app for teachers and school staff
- **Backend API**: Go (Chi) service handling authentication, assessment workflows, persistence, and AI orchestration
- **AI Services**: Stylometric feature extraction, embedding comparisons, reasoning generation, and micro‑viva question generation, backed by external LLM providers

For detailed architecture and implementation notes, see:

- `docs/architecture.md`
- `docs/master_plan.md` (strategic overview, this document’s source)
- `docs/design_guidelines.md`
- `.cursor/tasks/Next-Steps-Plan.md`

---

## Development Roadmap (Summarised)

1. **Stage 1 – Internal validation**  
   Store writing samples, build baselines, compare submissions, and generate authorship insight reports. Validate usefulness with real classroom data.

2. **Stage 2 – Micro‑viva integration**  
   Add targeted viva question generation, viva transcript analysis, and conceptual alignment assessment.

3. **Stage 3 – School pilots**  
   Deploy to partner schools, focus on teacher feedback, workflow usability, and policy alignment.

4. **Stage 4 – Productisation**  
   Harden for SaaS: dashboards, reporting, integrations, security, and operational readiness.

---

## Long-Term Vision

Microviva aims to define a new category of **AI‑era assessment integrity platforms**: infrastructure that lets schools embrace AI in learning while still centering thinking, reasoning, explanation, and authorship as the core of assessment.