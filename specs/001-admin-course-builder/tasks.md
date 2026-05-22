---
description: "Task list for Admin Course Builder feature"
---

# Tasks: Admin Course Builder

**Input**: Design documents from `specs/001-admin-course-builder/`

**Prerequisites**: plan.md ✅, spec.md ✅, research.md ✅, data-model.md ✅, contracts/

**Tests**: Not requested — no test tasks included.

**Organization**: Tasks grouped by user story for independent implementation and testing.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: US1–US4 maps to spec.md user stories
- Paths are relative to repo root

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Project initialization, new dependency, base model files.

- [x] T001 Add `github.com/golang-jwt/jwt/v5` dependency by running `go get github.com/golang-jwt/jwt/v5` from repo root
- [x] T002 [P] Create `models/language.go` with Language struct (id, code varchar(10) unique, name varchar(100))
- [x] T003 [P] Create `models/course.go` with Course, Lesson, ContentBlock, Exercise structs per data-model.md; include CEFRLevel and CourseStatus typed string aliases; add StringArray type (same pattern as IntArray in `models/paragraph.go`)
- [x] T004 [P] Create `models/lesson_attempt.go` with LessonAttempt and GrammarSuggestion structs per data-model.md
- [x] T005 [P] Create `handlers/admin_auth.go` as empty file with package declaration (stub for Phase 3)
- [x] T006 [P] Create `handlers/admin_course.go` as empty file with package declaration (stub for Phase 3)
- [x] T007 [P] Create `handlers/admin_grammar.go` as empty file with package declaration (stub for Phase 5)
- [x] T008 [P] Create `handlers/course.go` as empty file with package declaration (stub for Phase 4)
- [x] T009 [P] Create `handlers/lesson_attempt.go` as empty file with package declaration (stub for Phase 4)

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Database migration, auth middleware, backward-compatible model amendments.
All user story work is blocked until this phase is complete.

⚠️ **CRITICAL**: No user story work begins until this phase is complete.

- [x] T010 Extend `database/database.go` to `AutoMigrate` all new models: Language, Course, Lesson, ContentBlock, Exercise, LessonAttempt, GrammarSuggestion; after migration seed the Italian language row `{code:"it", name:"Italian"}` if `languages` table is empty
- [x] T011 Add nullable `LanguageID *uint` field with `gorm:"index"` to Word struct in `models/words.go` (backward-compatible; existing rows retain NULL = Italian)
- [x] T012 Add nullable `LanguageID *uint` field with `gorm:"index"` to Paragraph struct in `models/paragraph.go` (same backward-compat pattern as T011)
- [x] T013 Create `handlers/middleware.go` with `AdminAuth` Fiber middleware: reads `Authorization: Bearer <token>` header, validates JWT using `JWT_SECRET` from env, returns 401 on invalid/missing token
- [x] T014 Extend `routes/routes.go` to register a protected `/api/admin` group using the `AdminAuth` middleware from T013; register `/api/admin/login` outside the protected group
- [x] T015 [P] Extend `frontend/src/types.ts` with new TypeScript interfaces: Language, Course, Lesson, ContentBlock, Exercise, LessonAttempt, GrammarSuggestion matching the Go model shapes from data-model.md
- [x] T016 [P] Extend `frontend/src/api.ts` with admin API functions: `adminLogin`, `listCourses`, `generateCourse`, `getCourse`, `publishCourse`, `archiveCourse`, `patchCourse`; add learner API functions: `listPublishedCourses`, `getCourseLessons`, `getLesson`, `startAttempt`, `submitAnswer`, `completeAttempt`; store admin JWT in `localStorage` under key `admin_token`; store learner UUID in `localStorage` under key `learner_token` (generate UUID v4 on first access)

**Checkpoint**: Compile backend (`go build ./...`) with no errors before continuing.

---

## Phase 3: User Story 1 — Admin Generates a Course (Priority: P1) 🎯 MVP

**Goal**: Admin can log in, trigger AI course generation for a language/level/skill/topic,
review the draft, and publish it.

**Independent Test**: Follow `specs/001-admin-course-builder/quickstart.md` steps 2–5.
A generated draft course with at least one lesson and one exercise must appear in the DB.

### Backend — US1

- [x] T017 Implement `POST /api/admin/login` in `handlers/admin_auth.go`: compare request body `{username, password}` against `ADMIN_USERNAME`/`ADMIN_PASSWORD` env vars, issue a signed JWT (1-week expiry) using `JWT_SECRET`, return `{token}`; return 401 on mismatch
- [x] T018 Implement `GET /api/admin/courses` in `handlers/admin_course.go`: query `courses` table with optional query params `language` (code), `level`, `skill` (JSON contains), `status`; join language code; return array per admin-api.md contract
- [x] T019 Implement `GET /api/admin/courses/:id` in `handlers/admin_course.go`: load course with its lessons, content_blocks, and exercises ordered by `position`; return 404 if not found
- [x] T020 Implement `PATCH /api/admin/courses/:id` in `handlers/admin_course.go`: allow updating `title`, `description`, `topic` on draft courses only; return 400 if course is not in draft status
- [x] T021 Implement `POST /api/admin/courses/:id/publish` in `handlers/admin_course.go`: validate course has ≥1 lesson with ≥1 exercise; set status to `published`; return 400 with descriptive error if validation fails
- [x] T022 Implement `POST /api/admin/courses/:id/archive` in `handlers/admin_course.go`: set status to `archived`; return updated course
- [x] T023 Implement `POST /api/admin/courses/generate` in `handlers/admin_course.go`: validate request body `{language_code, level, skills, topic}`; before calling Claude, query the top 100 vocabulary words for the target language from the DB and inject the word list (Italian word + English translation pairs) into the Claude user message as context so FR-007 is satisfied; call Claude API (same pattern as `handlers/generate.go`) using the system prompt from `specs/001-admin-course-builder/contracts/mcp-tools.md`; force tool_choice to `create_course_draft`; return the saved draft course ID and lesson count; return 502 on Claude API error
- [x] T024 Add `get_courses` MCP tool to `cmd/mcp/main.go`: accepts optional `language_code`, `level`, `status`, `limit` inputs; queries courses with lesson titles; returns JSON array per mcp-tools.md contract
- [x] T025 Add `create_course_draft` MCP tool to `cmd/mcp/main.go`: accepts full course tree JSON per mcp-tools.md schema; inserts Course + Lessons + ContentBlocks + Exercises in one DB transaction; returns success message with new course ID
- [x] T026 Register all admin course routes in `routes/routes.go` inside the protected admin group: GET/PATCH/POST for courses and generate endpoint

### Frontend — US1

- [x] T027 [P] [US1] Create `frontend/src/pages/admin/AdminLoginPage.tsx`: form with username/password fields; on submit call `adminLogin` from api.ts; on success store token and redirect to `/admin/courses`; show error on 401
- [x] T028 [P] [US1] Create `frontend/src/pages/admin/AdminCoursesPage.tsx`: fetch and display course list with filter controls (language, level, status dropdowns); each row links to detail page; requires admin token (redirect to login if missing)
- [x] T029 [US1] Create `frontend/src/pages/admin/AdminGeneratePage.tsx`: form with language selector, level selector (A1/A2/B1-1/B1-2/B2/C1/C2), skills checkboxes (reading/writing/listening/speaking), topic text input; on submit call `generateCourse`; show loading spinner during generation (can take up to 30s); on success redirect to course detail page
- [x] T030 [US1] Create `frontend/src/pages/admin/AdminCourseDetailPage.tsx`: display full course (lessons, content blocks, exercises); show Publish button (if draft) or Archive button (if published); show editable title/description fields; call appropriate API functions on action
- [x] T031 [US1] Add admin routes to `frontend/src/App.tsx`: `/admin` → redirect to `/admin/courses`, `/admin/login` → AdminLoginPage, `/admin/courses` → AdminCoursesPage, `/admin/courses/generate` → AdminGeneratePage, `/admin/courses/:id` → AdminCourseDetailPage

**Checkpoint**: Admin can log in, generate a draft course, view it, and publish it end-to-end.

---

## Phase 4: User Story 2 — Learner Takes an Interactive Lesson (Priority: P1)

**Goal**: Learner can browse published courses, open a lesson, complete exercises with
real-time feedback, and see a completion score. Progress is preserved across sessions.

**Independent Test**: Open a published course lesson in the browser, answer all exercises,
verify score appears at the end. Close browser, reopen — verify progress was preserved.

### Backend — US2

- [x] T032 Implement `GET /api/courses` in `handlers/course.go`: return published courses with optional `language`/`level` query params; include `lesson_count` per course
- [x] T033 Implement `GET /api/courses/:id` in `handlers/course.go`: return published course with lesson list (title + position only, no content); return 404 if not found or not published
- [x] T034 Implement `GET /api/courses/:courseId/lessons/:lessonId` in `handlers/course.go`: return lesson with full content_blocks and exercises; omit `correct_answer` and `explanation` from exercise response to prevent client-side cheating; if `X-Learner-Token` header present, include most recent incomplete attempt for this lesson
- [x] T035 Implement `POST /api/courses/:courseId/lessons/:lessonId/attempts` in `handlers/lesson_attempt.go`: create new LessonAttempt row with `learner_token` from `X-Learner-Token` header; return attempt ID and initial state; return 400 if header missing
- [x] T036 Implement `PUT /api/courses/:courseId/lessons/:lessonId/attempts/:attemptId/answer` in `handlers/lesson_attempt.go`: validate `exercise_id` belongs to lesson; look up `correct_answer` and `explanation`; append answer to attempt's `answers` JSON; update `current_position`; return `{correct, correct_answer, explanation, next_position}` per learner-api.md
- [x] T037 Implement `POST /api/courses/:courseId/lessons/:lessonId/attempts/:attemptId/complete` in `handlers/lesson_attempt.go`: calculate score (count correct answers / total exercises); set `completed=true`, `completed_at=now`, `score`; return final score per learner-api.md
- [x] T038 Register all learner course and attempt routes in `routes/routes.go`

### Frontend — US2

- [x] T039 [P] [US2] Create `frontend/src/components/blocks/GrammarNote.tsx`: renders `{title, body}` content JSON as a styled card with markdown-like body text
- [x] T040 [P] [US2] Create `frontend/src/components/blocks/VocabularyList.tsx`: renders `{word_ids, notes}` — fetches word data and displays Italian word + English translation list with the speech synthesis button (reuse `SpeakButton` component)
- [x] T041 [P] [US2] Create `frontend/src/components/blocks/ConversationExample.tsx`: renders `{lines: [{speaker, text, translation}]}` as a dialogue with alternating speaker styling and speech button per line
- [x] T042 [P] [US2] Create `frontend/src/components/blocks/ReadingPassage.tsx`: renders `{text, translation}` with a toggle to show/hide the English translation
- [x] T043 [US2] Create `frontend/src/components/blocks/BlockRenderer.tsx`: type→component map for all four block types; renders the correct component based on `block.type`; logs a warning for unknown types
- [x] T044 [P] [US2] Create `frontend/src/components/exercises/MultipleChoice.tsx`: renders question + radio button options; on selection highlights correct/incorrect; shows explanation on wrong answer; calls `onAnswer(exerciseId, selectedIndex)` prop
- [x] T045 [P] [US2] Create `frontend/src/components/exercises/FillInTheBlank.tsx`: renders question with blank; text input for answer; on submit compares (case-insensitive trim) against correct answer from API response; shows feedback; calls `onAnswer(exerciseId, value)` prop
- [x] T046 [P] [US2] Create `frontend/src/components/exercises/ConversationReconstruction.tsx`: renders shuffled dialogue lines as draggable cards; learner drags to correct order; on submit compares order to correct sequence; calls `onAnswer(exerciseId, orderedIds)` prop
- [x] T047 [US2] Create `frontend/src/components/exercises/ExerciseRenderer.tsx`: type→component map for all three exercise types; passes `onAnswer` callback down; manages per-exercise feedback state
- [x] T048 [US2] Create `frontend/src/pages/LearnCoursesPage.tsx`: fetch published courses with optional language/level filter; display as cards with level badge and skill tags; each card links to lesson list
- [x] T049 [US2] Create `frontend/src/pages/LearnLessonPage.tsx`: load lesson data (content blocks + exercises + current attempt if any); render content blocks via BlockRenderer in order; render exercises via ExerciseRenderer; on each answer call `submitAnswer` API; on last exercise call `completeAttempt` and show score summary; manage learner token via `localStorage`
- [x] T050 [US2] Add learner routes to `frontend/src/App.tsx`: `/courses` → LearnCoursesPage, `/courses/:courseId/lessons/:lessonId` → LearnLessonPage; add "Courses" nav link to site header

**Checkpoint**: Learner can open a published lesson, answer all exercises, see score, close and reopen to verify progress persisted.

---

## Phase 5: User Story 3 — AI Suggests Missing Grammar Topics (Priority: P2)

**Goal**: Admin can view AI-generated suggestions of missing grammar topics for a
language/level pair and approve one to immediately trigger course generation.

**Independent Test**: With ≥1 published course, visit `/admin/grammar?language=it&level=A1`
and see at least one suggested topic. Click Accept on a suggestion — a new draft course
appears in the course list.

### Backend — US3

- [x] T051 Add `analyze_grammar_coverage` MCP tool to `cmd/mcp/main.go`: accepts `language_code` and `level`; queries published course topics for that language/level; calls Claude to identify covered grammar topics and suggest ≥5 missing ones; persists results as `GrammarSuggestion` rows with `status=pending`; returns JSON with `covered_topics` and `suggested_missing` arrays per mcp-tools.md contract
- [x] T052 Implement `GET /api/admin/grammar-suggestions` in `handlers/admin_grammar.go`: accepts `language` and `level` query params (both required); if no `pending` suggestions exist for this combination, call the `analyze_grammar_coverage` MCP tool handler function (the same Go function registered in `cmd/mcp/main.go`) to generate and persist suggestions — do NOT duplicate DB+Claude logic inline (Principle II); return the pending suggestions array per admin-api.md
- [x] T053 Implement `POST /api/admin/grammar-suggestions/:id/accept` in `handlers/admin_grammar.go`: set suggestion status to `accepted`; use suggestion's language/level/topic to call the same generation logic as `POST /api/admin/courses/generate`; return new draft course object; return 404 if suggestion not found
- [x] T054 Implement `POST /api/admin/grammar-suggestions/:id/dismiss` in `handlers/admin_grammar.go`: set suggestion status to `dismissed`; return updated suggestion
- [x] T055 Register grammar suggestion routes in `routes/routes.go` inside the protected admin group

### Frontend — US3

- [x] T056 [P] [US3] Add grammar suggestion API functions to `frontend/src/api.ts`: `getGrammarSuggestions(language, level)`, `acceptSuggestion(id)`, `dismissSuggestion(id)`
- [x] T057 [US3] Create `frontend/src/pages/admin/AdminGrammarPage.tsx`: language + level selectors to load suggestions; display each suggestion with topic, rationale, Accept and Dismiss buttons; on Accept show loading then redirect to new course detail; on Dismiss remove from list
- [x] T058 [US3] Add grammar route to `frontend/src/App.tsx`: `/admin/grammar` → AdminGrammarPage; add "Grammar" nav link to admin navigation

**Checkpoint**: Admin can see AI-suggested topics and approve one to create a course.

---

## Phase 6: User Story 4 — Admin Manages Existing Courses (Priority: P3)

**Goal**: Admin can filter the course list, edit draft metadata, and archive published courses.
(Most backend endpoints were implemented in Phase 3; this phase adds filtering UI and covers
the archive/edit flows end-to-end.)

**Independent Test**: Publish a course, archive it, verify it disappears from `/courses`
(learner view). Confirm a second published course is unaffected.

- [x] T059 [US4] Extend `frontend/src/pages/admin/AdminCoursesPage.tsx` with working filter controls: language dropdown (fetches from `GET /api/languages`), level dropdown, status dropdown; filters update the course list on change without page reload
- [x] T060 [US4] Add `GET /api/languages` endpoint in a new `handlers/language.go` file and register it in `routes/routes.go` (public endpoint — no auth required); returns all language rows
- [x] T061 [US4] Add `GET /api/languages` API function to `frontend/src/api.ts`; populate the language dropdown in AdminCoursesPage from this endpoint
- [x] T062 [US4] Verify archive flow in `AdminCourseDetailPage.tsx`: after archiving, redirect to course list; confirm archived course no longer appears in learner `LearnCoursesPage.tsx` (status filter in `GET /api/courses` already handles this)

**Checkpoint**: Admin can filter all combinations of language/level/status. Archiving a course removes it from learner view.

---

## Phase N: Polish & Cross-Cutting Concerns

**Purpose**: Wiring, error states, navigation, and end-to-end smoke test.

- [x] T063 [P] Add `"ADMIN_USERNAME"`, `"ADMIN_PASSWORD"`, `"JWT_SECRET"` entries to `.env-example` with placeholder values
- [x] T064 [P] Update `frontend/src/App.tsx` site header navigation: add "Courses" link pointing to `/courses` for learner nav; add separate "Admin" link pointing to `/admin` that only shows when `admin_token` is present in localStorage
- [ ] T065 Run the full quickstart.md validation: start server, run migration, login as admin, generate one A1 Italian course, publish it, verify learner can access it, complete a lesson, verify score saved
- [x] T066 [P] Push updated `mcp` binary to GitHub after rebuilding: `go build -o mcp ./cmd/mcp/main.go`

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies — start immediately; T002–T009 can all run in parallel
- **Foundational (Phase 2)**: Depends on Phase 1 complete — BLOCKS all user stories; T010 must run before T011/T012
- **US1 (Phase 3)**: Depends on Phase 2 — backend tasks T017–T026 can start in parallel; frontend T027–T030 can start after T015/T016
- **US2 (Phase 4)**: Depends on Phase 2 — backend T032–T038 can start in parallel with US1 backend; block/exercise components T039–T047 can be built in parallel
- **US3 (Phase 5)**: Depends on Phase 3 complete (needs published courses to analyse)
- **US4 (Phase 6)**: Depends on Phase 3 complete (builds on course list UI)
- **Polish (Phase N)**: Depends on all desired stories complete

### Within Each User Story

- Models before handlers
- Handlers before route registration
- Route registration before frontend API calls
- Frontend components before pages

---

## Parallel Opportunities

### Phase 1 (all parallel after T001)
```
T002 language.go  |  T003 course.go  |  T004 lesson_attempt.go  |  T005–T009 handler stubs
```

### Phase 3 Backend (all parallel after Phase 2)
```
T017 login  |  T018 list courses  |  T019 get course  |  T020 patch  |  T021 publish  |  T022 archive
T023 generate  |  T024 get_courses MCP  |  T025 create_course_draft MCP
```

### Phase 4 Components (all parallel)
```
T039 GrammarNote  |  T040 VocabularyList  |  T041 ConversationExample  |  T042 ReadingPassage
T044 MultipleChoice  |  T045 FillInTheBlank  |  T046 ConversationReconstruction
```

---

## Implementation Strategy

### MVP First (US1 + US2 Only)

1. Complete Phase 1: Setup
2. Complete Phase 2: Foundational
3. Complete Phase 3: Admin generates and publishes a course
4. Complete Phase 4: Learner takes the course interactively
5. **STOP and VALIDATE**: Full quickstart.md end-to-end test
6. Deploy/demo with real Italian courses

### Incremental Delivery

1. Setup + Foundational → foundation ready
2. US1 → admin can generate + publish courses (MVP admin)
3. US2 → learner can study interactively (MVP learner)
4. US3 → AI curriculum gap analysis added
5. US4 → admin management polish

---

## Notes

- No test tasks — not requested in spec
- [P] tasks target different files and have no shared dependencies
- Each user story phase is independently completable and testable
- `correct_answer` is never sent to the client in the lesson GET response — only returned in the answer submission response
- Learner token UUID is generated and stored entirely client-side; server never creates it
- The `analyze_grammar_coverage` MCP tool (T051) requires Claude API access; ensure `ANTHROPIC_API_KEY` is set
