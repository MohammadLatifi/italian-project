# Data Model: Admin Course Builder

**Date**: 2026-05-22
**Research**: research.md

---

## Entity Relationship Overview

```
Language (1) ──< Course (1) ──< Lesson (1) ──< ContentBlock
                                           └──< Exercise
                                (1) ──< LessonAttempt

Language (1) ──< Word          (existing, gains language_id FK)
Language (1) ──< Paragraph     (existing, gains language_id FK)
Language (1) ──< GrammarSuggestion
```

---

## New Entities

### Language

Lookup table for supported languages. Adding a row is the only action needed to support a new language.

| Column      | Type         | Constraints              | Notes                      |
|-------------|--------------|--------------------------|----------------------------|
| id          | uint         | PK, auto-increment       | GORM default               |
| code        | varchar(10)  | NOT NULL, UNIQUE         | ISO 639-1 code: `it`, `fa` |
| name        | varchar(100) | NOT NULL                 | Display name: "Italian"    |
| created_at  | datetime     | NOT NULL                 | GORM auto                  |
| updated_at  | datetime     | NOT NULL                 | GORM auto                  |

**Seed data**: `{code: "it", name: "Italian"}` inserted on first migration.

---

### Course

Top-level content unit. Belongs to a Language. Contains ordered Lessons.

| Column      | Type         | Constraints              | Notes                                         |
|-------------|--------------|--------------------------|-----------------------------------------------|
| id          | uint         | PK, auto-increment       | GORM default                                  |
| language_id | uint         | FK → languages.id        | NOT NULL                                      |
| level       | enum         | NOT NULL                 | A1, A2, B1-1, B1-2, B2, C1, C2               |
| title       | varchar(255) | NOT NULL                 | Short display title                           |
| topic       | text         | NOT NULL                 | Admin-provided topic prompt                   |
| description | text         |                          | Optional summary for learner                  |
| status      | enum         | NOT NULL, default: draft | draft, published, archived                    |
| skills      | json         | NOT NULL                 | StringArray: ["reading","writing",...]        |
| created_at  | datetime     | NOT NULL                 | GORM auto                                     |
| updated_at  | datetime     | NOT NULL                 | GORM auto                                     |
| deleted_at  | datetime     | nullable                 | GORM soft-delete                              |

**State transitions**: draft → published → archived. Archived cannot be re-published without
explicitly updating status (no automatic transitions).

---

### Lesson

A named section within a Course. Has an explicit `position` for ordering.

| Column     | Type         | Constraints       | Notes              |
|------------|--------------|-------------------|--------------------|
| id         | uint         | PK, auto-increment|                    |
| course_id  | uint         | FK → courses.id   | NOT NULL           |
| title      | varchar(255) | NOT NULL          |                    |
| position   | int          | NOT NULL, default 0 | Ordering within course |
| created_at | datetime     | NOT NULL          | GORM auto          |
| updated_at | datetime     | NOT NULL          | GORM auto          |
| deleted_at | datetime     | nullable          | GORM soft-delete   |

---

### ContentBlock

A unit of lesson content. Ordered within a Lesson by `position`.

| Column     | Type     | Constraints        | Notes                                                        |
|------------|----------|--------------------|--------------------------------------------------------------|
| id         | uint     | PK, auto-increment |                                                              |
| lesson_id  | uint     | FK → lessons.id    | NOT NULL                                                     |
| type       | enum     | NOT NULL           | vocabulary-list, grammar-note, conversation-example, reading-passage |
| content    | json     | NOT NULL           | Shape depends on type (see Content Shapes below)             |
| position   | int      | NOT NULL, default 0|                                                              |
| created_at | datetime | NOT NULL           | GORM auto                                                    |
| updated_at | datetime | NOT NULL           | GORM auto                                                    |

**Content shapes by type**:

```json
// vocabulary-list
{ "word_ids": [1, 2, 3], "notes": "optional context note" }

// grammar-note
{ "title": "Present tense of essere", "body": "Markdown text" }

// conversation-example
{ "lines": [{"speaker": "A", "text": "Ciao!", "translation": "Hi!"}] }

// reading-passage
{ "text": "Italian text...", "translation": "English translation..." }
```

---

### Exercise

An interactive item within a Lesson. Ordered by `position`.

| Column         | Type     | Constraints        | Notes                                            |
|----------------|----------|--------------------|--------------------------------------------------|
| id             | uint     | PK, auto-increment |                                                  |
| lesson_id      | uint     | FK → lessons.id    | NOT NULL                                         |
| type           | enum     | NOT NULL           | multiple-choice, fill-in-the-blank, conversation-reconstruction |
| question       | text     | NOT NULL           | Prompt shown to learner                          |
| options        | json     |                    | StringArray — answer options (multiple-choice only) |
| correct_answer | text     | NOT NULL           | Index string for MC ("2"), blank fill for FIB    |
| explanation    | text     |                    | Shown after wrong answer                         |
| skills         | json     | NOT NULL           | StringArray: ["reading"], ["listening","speaking"] |
| position       | int      | NOT NULL, default 0|                                                  |
| created_at     | datetime | NOT NULL           | GORM auto                                        |
| updated_at     | datetime | NOT NULL           | GORM auto                                        |

---

### LessonAttempt

Records one learner's progress through one Lesson.

| Column           | Type         | Constraints         | Notes                                          |
|------------------|--------------|---------------------|------------------------------------------------|
| id               | uint         | PK, auto-increment  |                                                |
| lesson_id        | uint         | FK → lessons.id     | NOT NULL                                       |
| learner_token    | varchar(64)  | NOT NULL            | UUID from client localStorage                  |
| current_position | int          | NOT NULL, default 0 | Index of current exercise                      |
| answers          | json         |                     | Map of exercise_id → answer given              |
| score            | int          | nullable            | Set on completion (correct count / total)      |
| completed        | bool         | NOT NULL, default false |                                            |
| started_at       | datetime     | NOT NULL            | GORM auto (created_at)                         |
| completed_at     | datetime     | nullable            | Set when completed=true                        |

**Index**: `(lesson_id, learner_token)` — allows upsert-style resume.
Learner can retake a lesson: a new `LessonAttempt` row is created; previous attempts are preserved.

---

### GrammarSuggestion

An AI-generated suggestion of a missing grammar topic for a language/level pair.

| Column      | Type         | Constraints         | Notes                          |
|-------------|--------------|---------------------|--------------------------------|
| id          | uint         | PK, auto-increment  |                                |
| language_id | uint         | FK → languages.id   | NOT NULL                       |
| level       | enum         | NOT NULL            | Same CEFR enum as Course       |
| topic       | text         | NOT NULL            | Suggested grammar topic text   |
| rationale   | text         |                     | Why AI thinks this is missing  |
| status      | enum         | NOT NULL, default: pending | pending, accepted, dismissed |
| created_at  | datetime     | NOT NULL            | GORM auto                      |
| updated_at  | datetime     | NOT NULL            | GORM auto                      |

---

## Existing Entity Amendments

### Word (existing — `models/words.go`)

Add one nullable column:

| Column      | Type | Constraints          | Notes                                   |
|-------------|------|----------------------|-----------------------------------------|
| language_id | uint | FK → languages.id, nullable | NULL = Italian (backward compat) |

### Paragraph (existing — `models/paragraph.go`)

Add one nullable column:

| Column      | Type | Constraints          | Notes                                   |
|-------------|------|----------------------|-----------------------------------------|
| language_id | uint | FK → languages.id, nullable | NULL = Italian (backward compat) |

---

## GORM Migration Notes

- All new tables created via `db.AutoMigrate(...)` in `database/database.go`.
- Seed: after Language table is created, insert `{code:"it", name:"Italian"}` if empty.
- The nullable `language_id` on `Word` and `Paragraph` means zero migration pain for
  the existing 2000+ rows — they continue to work and are implicitly Italian.
