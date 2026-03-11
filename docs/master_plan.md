# Microviva — Master Plan

## Authenticating Student Thinking in the Age of AI

Microviva is an AI-assisted assessment platform designed to help educators verify student understanding and authorship in a world where generative AI can easily produce high-quality written work.

Rather than attempting to detect AI usage, Microviva focuses on **evidence of thinking**. It combines writing baseline analysis, authorship consistency reasoning, and targeted micro-viva conversations to help teachers confirm that submitted work reflects a student's own understanding.

This document provides the **10,000-foot overview** of the system, including:

- why the product exists  
- who it serves  
- what success looks like  
- the guiding principles behind the platform  

For detailed implementation guidance, see **design_guidelines.md**.

---

# 1. The Problem

Generative AI has fundamentally changed written assessment.

Students can now generate essays, reports, and explanations that are:

- grammatically correct  
- conceptually structured  
- stylistically polished  

Traditional academic integrity tools attempt to solve this by **detecting AI usage**.

However this approach has major limitations:

- AI detectors are unreliable  
- False positives damage trust  
- Detection tools cannot prove authorship  
- The arms race between detectors and models is unwinnable  

More importantly:

**The real problem is not AI use.  
The real problem is verifying student thinking.**

Teachers need to answer a simple question:

> Does this work genuinely represent this student’s understanding?

Existing systems do not provide a reliable or defensible way to answer that question.

---

# 2. A New Reality: AI is Part of the Learning Environment

AI is rapidly becoming a normal part of the learning process.

Different assessments allow different levels of AI assistance.

In some tasks:

- AI may be prohibited.

In others:

- AI may be used for brainstorming.
- AI may assist drafting.
- AI may be part of collaborative knowledge work.

The challenge for educators is not simply preventing AI use.

The challenge is designing assessments where **student thinking remains visible and verifiable**.

Microviva supports this by incorporating an **AI Assistance Level** into the assessment workflow.

---

# 3. AI Assistance Levels in Assessment

Before authorship comparison occurs, teachers specify the **AI Assistance Level** for the task.

This concept is inspired by research such as the *AI Assessment Scale* developed by Leon Furze and collaborators.

The scale recognises that AI can play different roles in learning depending on the design of the assessment.

Microviva incorporates a similar framework.

### Example AI Assistance Levels

| Level | Description |
|------|-------------|
| **Level 1 — No AI** | AI tools are not permitted |
| **Level 2 — AI Planning** | AI may be used for brainstorming or planning |
| **Level 3 — AI Collaboration** | AI may assist drafting but students must significantly edit |
| **Level 4 — AI Co-Creation** | AI can act as a collaborative writing partner |
| **Level 5 — AI Exploration** | AI may generate content and the student evaluates or critiques it |

The selected level provides **context for the authorship analysis system**.

Instead of asking:

> Is this writing identical to the student's baseline?

Microviva asks:

> Is this writing consistent with the student’s baseline **given the level of AI assistance permitted for this task?**

This allows the system to interpret differences appropriately.

---

# 4. The Core Idea

Microviva reframes the authorship problem.

Instead of asking:

> “Was AI used?”

Microviva asks:

> “Is this work consistent with this student’s demonstrated thinking?”

The platform builds **longitudinal writing profiles** for students using supervised in-class writing samples.

When new work is submitted, Microviva analyzes the submission relative to the student’s established writing patterns.

If significant differences appear, the system recommends a **micro-viva** — a short oral conversation designed to confirm conceptual ownership.

This creates a closed loop:

baseline writing
↓
submission comparison
↓
context-aware authorship reasoning
↓
micro-viva follow-up
↓
teacher judgement


Microviva does not replace teacher judgement.

It **supports it with structured evidence**.

---

# 5. Who Microviva Is For

Microviva is designed primarily for **secondary schools (Years 7–12)** and the educators responsible for assessing student learning.

## Teachers

Teachers need tools that help them:

- verify student understanding
- investigate suspicious submissions
- confidently assess work completed outside the classroom
- maintain academic integrity without relying on unreliable detection tools

Microviva provides structured insight without requiring teachers to become AI experts.

---

## Faculty Leaders

Heads of department and curriculum leaders need systems that:

- scale across classes
- align with assessment policy
- support consistent integrity practices
- reduce time spent investigating suspected misconduct

Microviva creates a transparent, auditable workflow for authenticity verification.

---

## School Leadership

School leadership teams are increasingly concerned about:

- academic integrity
- AI use in assessment
- maintaining trust in grades

Microviva offers a constructive alternative to bans and detection systems.

It enables schools to **adapt assessment practices to the AI era**.

---

# 6. Product Vision

Microviva aims to become the **assessment integrity infrastructure for AI-era education**.

The platform brings together three complementary systems.

---

## 1. Writing Baseline Profiles

Students produce supervised writing samples during class.

Microviva analyzes these samples to build a **baseline writing fingerprint** that captures patterns such as:

- sentence structure
- vocabulary usage
- stylistic markers
- error patterns
- argument structure

These baselines provide the reference point for authorship comparisons.

---

## 2. Context-Aware Authorship Analysis

When a student submits an assignment, Microviva compares the work to their baseline profile.

The system analyzes:

- stylistic consistency
- structural differences
- semantic distance
- feature deviations

Importantly, these comparisons are interpreted **in light of the AI Assistance Level chosen for the assessment**.

The output is not an “AI probability.”

Instead, Microviva produces a **structured explanation of observed differences**.

---

## 3. Micro-Viva Conversations

If discrepancies appear, Microviva can generate **targeted viva questions**.

A micro-viva is a short oral conversation (2–5 minutes) where the student explains their reasoning.

The goal is not interrogation.

It is to confirm conceptual understanding.

Microviva helps generate questions that focus on the most important parts of the student's submission.

---

# 7. What Success Looks Like

Microviva succeeds when:

### Teachers regain confidence in assessment

Teachers should feel able to say:

> “I know this student understands their work.”

---

### AI stops being a threat to assessment

Instead of banning AI, schools can adopt **verification workflows** that maintain integrity.

---

### Micro-vivas become normal practice

Short oral follow-ups become a standard part of authentic assessment.

---

### Schools trust the system

The platform must produce insights that are:

- explainable
- evidence-based
- transparent
- respectful of teacher judgement

---

### Microviva defines a new category

The long-term ambition is for Microviva to define a new category:

**AI-Era Assessment Integrity Platforms**

---

# 8. Guiding Principles

## Teacher judgement is central

Microviva does not make disciplinary decisions.

Teachers remain responsible for interpreting evidence.

---

## Avoid AI detection framing

Microviva does not claim to detect AI usage.

Instead it focuses on **authorship consistency and conceptual understanding**.

---

## Evidence over probability

Outputs should emphasize:

- observed differences
- reasoning
- context

rather than opaque probability scores.

---

## Pedagogy first

Technology should support good assessment practice.

Microviva encourages:

- baseline writing tasks
- reflective assessment
- oral explanation of thinking

---

## Workflow over analytics

Teachers do not want dashboards.

They want tools that integrate into existing assessment workflows.

Microviva prioritizes actionable recommendations.

---

# 9. System Architecture Overview

At a high level the system consists of three layers.

Frontend (Next.js)
↓
Backend API (Go / Chi)
↓
AI Services
↓
LLM Providers


The backend manages:

- authentication
- assessment workflows
- data persistence
- orchestration of AI analysis

AI services perform:

- stylometric feature extraction
- embedding comparisons
- reasoning generation
- interview question generation

For detailed architecture see:

**architecture.md**

---

# 10. Development Strategy

## Stage 1 — Internal Validation

Build a working system that:

- stores writing samples
- builds baseline profiles
- compares submissions
- produces authorship insight reports

Test the system with real classroom data.

Goal: validate usefulness for teachers.

---

## Stage 2 — Micro-Viva Integration

Add:

- targeted viva question generation
- viva transcript analysis
- conceptual alignment assessment

Goal: close the authorship verification loop.

---

## Stage 3 — School Pilots

Deploy Microviva in a small number of partner schools.

Focus on:

- teacher feedback
- workflow usability
- policy alignment

Goal: confirm real-world value.

---

## Stage 4 — Productisation

Develop:

- teacher dashboards
- reporting tools
- integrations
- security features

Goal: launch Microviva as a commercial SaaS platform.

---

# 11. Relationship to Other Documents

This document provides the **strategic overview**.

More detailed documentation lives elsewhere:

| Document | Purpose |
|--------|--------|
| **architecture.md** | Current system architecture |
| **design_guidelines.md** | Engineering principles and implementation decisions |
| **next-steps-plan.md** | Development roadmap and task tracking |
| **schema.sql** | Database schema |

---

# 12. Long-Term Vision

Education is entering a new phase.

AI will increasingly assist students in generating written work.

This does not mean assessment must collapse.

Instead, assessment must evolve.

The future of assessment will emphasize:

- thinking
- reasoning
- explanation
- authorship

Microviva is an attempt to build the infrastructure for that future.