# Feature Specification: Admin Course Builder

**Feature Branch**: `001-admin-course-builder`

**Created**: 2026-05-22

**Status**: Draft

**Input**: User description: "admin course builder with AI generation, interactive learner experience,
language-agnostic design, CEFR levels, and 4-skill coverage"

---

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Admin Generates a Course (Priority: P1)

The admin opens the admin panel, selects a language (e.g., Italian), a CEFR level (e.g., B1-1),
and the target language skills (e.g., Reading + Listening). They provide a course topic
(e.g., "Daily routines") and trigger AI generation. The system produces a structured course
including vocabulary sections, grammar explanations, conversation examples, and exercises.
The admin reviews and publishes the course.

**Why this priority**: This is the entire value proposition of the feature. Without the ability
to generate and publish courses, nothing else has meaning.

**Independent Test**: An admin can trigger generation of a single course for one language/level/skill
combination and see a fully formed, reviewable draft — without any learner interaction being needed.

**Acceptance Scenarios**:

1. **Given** the admin is on the course creation page, **When** they select language=Italian,
   level=A1, skills=[Reading], topic="Greetings", and trigger generation, **Then** a draft course
   appears containing at least one vocabulary section, one grammar note, one conversation example,
   and one quiz.

2. **Given** a generated draft course, **When** the admin clicks Publish, **Then** the course
   becomes visible to learners and its status changes to "Published".

3. **Given** the admin selects skills that include Speaking or Listening, **When** the course
   is generated, **Then** it contains at least one exercise tagged with that skill.

---

### User Story 2 - Learner Takes an Interactive Course Lesson (Priority: P1)

A learner opens a published course, selects a lesson, and progresses through it interactively.
They read content, answer multiple-choice and fill-in-the-blank questions, receive immediate
feedback on each answer, and see a completion score at the end of the lesson.

**Why this priority**: The interactive learner experience is the distinguishing value over
static reading material. Without it the platform is just a document viewer.

**Independent Test**: A learner can open a single published lesson, complete all its exercises,
and receive a pass/fail score — without any admin actions required after publishing.

**Acceptance Scenarios**:

1. **Given** a published lesson with a quiz, **When** the learner selects the correct answer,
   **Then** they see instant positive feedback and can advance to the next question.

2. **Given** a published lesson with a quiz, **When** the learner selects a wrong answer,
   **Then** they see the correct answer with an explanation and can continue.

3. **Given** a learner completes all exercises in a lesson, **When** the last exercise is
   submitted, **Then** they see a summary with their score and the option to move to the
   next lesson.

4. **Given** a learner returns to a lesson they already started, **When** they open it,
   **Then** their progress is preserved and they can resume where they left off.

---

### User Story 3 - AI Suggests Missing Grammar Topics (Priority: P2)

The admin opens a grammar coverage dashboard for a language/level combination. The system
(powered by AI) analyses existing published courses and identifies grammar topics not yet
covered. It presents a list of suggested topics. The admin selects a suggestion and approves
AI generation of a new course for that topic.

**Why this priority**: This closes the curriculum gap loop and is the key differentiator from
a purely manual admin workflow, but learners can use the platform without it.

**Independent Test**: With at least one published course present, the admin can view a list
of AI-suggested missing topics for the same language and level — and trigger course creation
for one of them — without any learner involvement.

**Acceptance Scenarios**:

1. **Given** the admin views the grammar coverage dashboard for Italian B1-1, **When** the
   page loads, **Then** at least one suggested missing topic is displayed with a brief rationale.

2. **Given** a displayed suggestion, **When** the admin clicks "Generate course for this topic",
   **Then** the course draft creation flow begins pre-filled with that topic.

3. **Given** a suggestion the admin does not want, **When** they dismiss it, **Then** it no
   longer appears in the suggestion list for that session.

---

### User Story 4 - Admin Manages Existing Courses (Priority: P3)

The admin can view all courses in a list, filter by language/level/skill/status, edit a
published or draft course, and archive (unpublish) a course so learners can no longer access it.

**Why this priority**: Necessary for long-term content health but does not block learner
or generation workflows.

**Independent Test**: An admin can change the status of a published course to Archived and
confirm learners can no longer access it — without affecting other courses.

**Acceptance Scenarios**:

1. **Given** the admin is on the course list, **When** they filter by level=A1 and
   language=Italian, **Then** only matching courses are shown.

2. **Given** a published course, **When** the admin archives it, **Then** its status becomes
   "Archived" and it disappears from the learner-facing course list.

3. **Given** a draft course, **When** the admin edits the course topic and saves, **Then** the
   updated topic is reflected without triggering a full re-generation.

---

### Edge Cases

- What happens when AI generation fails mid-course (partial content produced)?
- What if a learner's session expires mid-lesson — is their progress saved?
- What if the admin selects a language that has no vocabulary in the database yet?
- What if two admins attempt to edit the same course simultaneously?
- What happens when a learner retakes a lesson — do scores accumulate or overwrite?

---

## Requirements *(mandatory)*

### Functional Requirements

**Course Management**

- **FR-001**: The system MUST support multiple languages, identified by a language code
  (e.g., `it`, `fa`). Adding a new language MUST NOT require code changes.
- **FR-002**: Each course MUST be associated with exactly one language, one CEFR level
  (A1, A2, B1-1, B1-2, B2, C1, C2), and one or more of the four language skills
  (Reading, Writing, Listening, Speaking).
- **FR-003**: A course MUST have at least one lesson. A lesson MUST contain at least one
  content section and one interactive exercise.
- **FR-004**: Course status MUST be one of: Draft, Published, Archived. Only Published
  courses are visible to learners.
- **FR-005**: The system MUST allow the admin to manually edit any field of a draft course
  before publishing.

**AI-Powered Generation**

- **FR-006**: The admin MUST be able to trigger AI generation of a complete course draft
  by providing: language, level, target skills, and a course topic.
- **FR-007**: The AI generation MUST use existing vocabulary words and paragraphs already
  stored in the database as source material where applicable.
- **FR-008**: The system MUST expose MCP tools that the AI uses to read vocabulary,
  paragraphs, and existing course content during generation.
- **FR-009**: The system MUST expose documented prompts and tool schemas so that the AI
  can generate courses consistently without free-form hallucination of structure.
- **FR-010**: The AI grammar coverage checker MUST analyse published courses for a given
  language/level and return a list of suggested missing grammar topics with brief rationale.
- **FR-011**: The admin MUST be able to approve a suggested topic and immediately trigger
  course generation for it.

**Interactive Learner Experience**

- **FR-012**: The system MUST support at least three exercise types: multiple-choice,
  fill-in-the-blank, and conversation reconstruction (ordering dialogue lines).
- **FR-013**: The system MUST provide immediate feedback after each exercise answer,
  including the correct answer and a brief explanation when the learner is wrong.
- **FR-014**: Learner progress (current position in a lesson, answers given) MUST be
  persisted so the learner can resume after leaving.
- **FR-015**: The system MUST record a score for each completed lesson attempt.
- **FR-016**: Each exercise MUST be tagged with one or more language skills so that
  skill-specific progress can be tracked.

**Admin Panel**

- **FR-017**: The admin panel MUST be protected by authentication. Unauthenticated users
  MUST NOT access any admin functionality.
- **FR-018**: The admin MUST be able to filter the course list by language, level, skill,
  and status.
- **FR-019**: The admin MUST be able to archive (unpublish) any published course.

### Key Entities

- **Language**: A supported language (code, name, e.g., `it` / Italian). Vocabulary and
  courses belong to a language.
- **Course**: The top-level content unit. Has a language, level, skills, topic, status,
  and ordered list of lessons.
- **Lesson**: A named section within a course. Contains an ordered list of content blocks
  and exercises.
- **ContentBlock**: A unit of lesson content — types include: vocabulary-list, grammar-note,
  conversation-example, reading-passage.
- **Exercise**: An interactive item within a lesson — types include: multiple-choice,
  fill-in-the-blank, conversation-reconstruction. Each exercise carries a skill tag.
- **LessonAttempt**: Records a learner's progress through a specific lesson: current
  position, answers given, score, completion status.
- **GrammarSuggestion**: An AI-generated suggestion of a missing grammar topic for a
  language/level pair. Has a status: pending, accepted, dismissed.

---

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: An admin can go from "no course" to a fully reviewed draft course for any
  language/level/skill/topic combination in under 5 minutes.
- **SC-002**: A learner can complete a lesson end-to-end (open → last exercise → score)
  without any page that takes more than 3 seconds to respond.
- **SC-003**: 100% of published course content is associated with a CEFR level and at
  least one language skill — no unclassified content is ever visible to learners.
- **SC-004**: The grammar coverage tool identifies at least 5 suggested missing topics
  for any language/level pair that has fewer than 10 published courses.
- **SC-005**: Adding a new language (e.g., Persian) requires zero code changes — only
  a new language record in the database.
- **SC-006**: Learner progress is never lost: if a learner closes the browser mid-lesson
  and returns, they resume at the same position with the same answers preserved.

---

## Assumptions

- The admin panel has a single admin role (no multi-role permissions needed for v1).
  All authenticated admins have full access.
- Learner accounts are out of scope for v1 — learner progress is stored per browser
  session or a simple anonymous identifier, not a registered user account.
- The existing vocabulary database (2000+ Italian words) is the primary source material
  the AI uses when generating Italian courses; other languages will need their own vocabulary
  loaded before generation is practical.
- Speech synthesis (text-to-speech) for Listening exercises is handled client-side using
  the browser's built-in Web Speech API, as already used by the existing frontend.
- The platform hosts one instance (not multi-tenant); language isolation is by data record,
  not by deployment.
- Video content is out of scope for v1.
- Offline / downloadable course support is out of scope for v1.
