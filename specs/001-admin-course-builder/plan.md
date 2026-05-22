# Implementation Plan: Admin Course Builder

**Branch**: `001-admin-course-builder` | **Date**: 2026-05-22 | **Spec**: [spec.md](spec.md)

**Input**: Feature specification from `specs/001-admin-course-builder/spec.md`

## Summary

Build an AI-powered admin panel that generates interactive language courses (Go/Fiber backend +
React/TypeScript frontend) stored in MySQL. Courses are language-agnostic, levelled by CEFR
(A1→C2), and tagged by the four language skills (reading, writing, listening, speaking).
Claude generates course drafts via the existing MCP server pattern; learners interact with
courses through quizzes and exercises with real-time feedback and anonymous progress tracking.

## Technical Context

**Language/Version**: Go 1.25+, TypeScript (React + Vite)

**Primary Dependencies**: Fiber v3, GORM v1.31+, golang-jwt/jwt v5 (new — admin auth),
React + Vite (existing frontend)

**Storage**: MySQL 8+ via GORM — 7 new tables, 2 existing tables amended (see data-model.md)

**Testing**: Go standard `testing` package; manual integration via quickstart.md

**Target Platform**: Web application — Go HTTP server + React SPA, same as existing project

**Project Type**: Web application (backend API + frontend SPA + MCP server extension)

**Performance Goals**: Admin generation ≤ 30s (Claude API call, acceptable for admin-only);
learner page responses ≤ 3s (spec SC-002)

**Constraints**: Simplicity First — no job queue, no background workers, no Redis.
GORM auto-migrate only (no raw SQL migrations). Single admin role.

**Scale/Scope**: Personal project — single admin, < 100 learners initially.

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-checked after Phase 1 design.*

| Principle | Status | Notes |
|-----------|--------|-------|
| I. Simplicity First | ✅ PASS | Synchronous generation, no job queue, no extra layers |
| II. MCP-Native AI Integration | ✅ PASS | Course writes go through `create_course_draft` MCP tool; admin REST handler calls Claude which calls MCP |
| III. Persistent Data Integrity | ✅ PASS | All entities use GORM soft-delete; LessonAttempt rows preserved on retake |
| IV. Clean REST Contracts | ✅ PASS | JSON in/out, proper status codes, no business logic in routes/ |
| V. Learning-Focused Frontend | ✅ PASS | Interactive exercise components, speech synthesis carried forward, no perf regression on existing pages |

**Post-Phase-1 re-check**: All gates still pass. One new dependency (`golang-jwt/jwt`) is
justified — smallest auth surface for single-admin JWT (Decision 1 in research.md).

## Project Structure

### Documentation (this feature)

```text
specs/001-admin-course-builder/
├── plan.md                   # This file
├── research.md               # Phase 0 decisions
├── data-model.md             # Phase 1 entity definitions
├── quickstart.md             # Phase 1 developer guide
├── contracts/
│   ├── admin-api.md          # Admin REST contract
│   ├── learner-api.md        # Learner REST contract
│   └── mcp-tools.md          # MCP tool schemas + system prompt
└── tasks.md                  # Phase 2 output (/speckit-tasks)
```

### Source Code (repository root — web application)

```text
models/
├── language.go               # Language entity
├── course.go                 # Course, Lesson, ContentBlock, Exercise
├── lesson_attempt.go         # LessonAttempt, GrammarSuggestion
├── words.go                  # (existing — add language_id)
└── paragraph.go              # (existing — add language_id)

handlers/
├── admin_auth.go             # POST /api/admin/login → JWT
├── admin_course.go           # Admin course CRUD + AI generation
├── admin_grammar.go          # Grammar suggestion endpoints
├── course.go                 # Learner course read endpoints
├── lesson_attempt.go         # Learner progress endpoints
├── generate.go               # (existing)
├── word.go                   # (existing)
└── paragraph.go              # (existing)

routes/
└── routes.go                 # Extended with admin + learner routes + JWT middleware

cmd/mcp/
└── main.go                   # Extended: get_courses, create_course_draft, analyze_grammar_coverage

database/
└── database.go               # Extended: AutoMigrate new models, seed Italian language

frontend/src/
├── pages/
│   ├── admin/
│   │   ├── AdminLoginPage.tsx
│   │   ├── AdminCoursesPage.tsx
│   │   ├── AdminCourseDetailPage.tsx
│   │   ├── AdminGeneratePage.tsx
│   │   └── AdminGrammarPage.tsx
│   ├── LearnCoursesPage.tsx
│   └── LearnLessonPage.tsx
├── components/
│   └── exercises/
│       ├── MultipleChoice.tsx
│       ├── FillInTheBlank.tsx
│       └── ConversationReconstruction.tsx
├── api.ts                    # (existing — extend with new API calls)
├── types.ts                  # (existing — extend with new types)
└── App.tsx                   # (existing — add new routes)
```

**Structure Decision**: Web application (Option 2 adapted) — Go backend at repo root,
React frontend in `frontend/`. Admin and learner sections share the same backend but
admin routes are JWT-protected. This matches the existing project layout exactly.
