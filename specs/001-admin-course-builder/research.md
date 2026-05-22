# Research: Admin Course Builder

**Phase 0 — All NEEDS CLARIFICATION resolved**
**Date**: 2026-05-22

---

## Decision 1: Admin Authentication

**Decision**: JWT (HMAC-SHA256) issued on single-credential login; credentials stored in `.env`.
No database user table for v1.

**Rationale**: One admin user. Adding a `golang-jwt/jwt` dependency is the smallest possible
surface area. Sessions require a store; OAuth2 requires a provider. JWT + `.env` credential
is auditable and reversible with zero schema changes.

**Alternatives considered**:
- HTTP Basic Auth: rejected — no token expiry, harder to extend later.
- Session cookies with server-side store: rejected — requires Redis or DB table, violates
  Simplicity First for a single-user panel.
- OAuth2/SSO: rejected — third-party dependency, overkill for personal project.

---

## Decision 2: Learner Progress (Anonymous Tracking)

**Decision**: Client generates a UUID v4 on first visit, stores it in `localStorage`, and sends
it as an `X-Learner-Token` request header. Server stores `learner_token` on `LessonAttempt`
records. No user accounts, no cookies, no server-side session.

**Rationale**: Spec explicitly defers learner accounts to post-v1. `localStorage` UUID gives
persistent identity across browser restarts without any authentication infrastructure.

**Alternatives considered**:
- Cookie-based anonymous session: rejected — same data as localStorage but adds CORS/SameSite
  complexity.
- Registered user accounts: deferred to post-v1 per spec assumption.

---

## Decision 3: Course Generation (Synchronous Claude API)

**Decision**: Admin triggers generation via a REST endpoint. The handler calls the Claude API
synchronously with a structured tool_choice (same pattern as `handlers/generate.go`), and
the full course JSON is returned and persisted as a draft before the HTTP response is sent.

**Rationale**: Follows the existing `GenerateParagraph` pattern (Principle I: Simplicity First,
Principle II: MCP-Native AI Integration respected for bulk reads; the REST endpoint is for
admin-triggered single-run generation). No background job infrastructure needed.

**Alternatives considered**:
- Background job + polling: rejected — requires job queue and status endpoint, violates
  Simplicity First. Course generation takes 10–30s which is acceptable behind a loading spinner
  for an admin-only action.
- Streaming response (SSE): considered but deferred — useful UX enhancement but complicates
  the handler significantly for v1.

---

## Decision 4: Exercise Storage

**Decision**: Each exercise is a row in an `exercises` table with typed columns:
`type` (enum), `question` (text), `options` (JSON array of strings), `correct_answer` (text),
`explanation` (text), `skills` (JSON array of strings), `position` (int).

**Rationale**: Flat typed rows with JSON for variable-length arrays. Avoids polymorphic table
design or EAV anti-pattern. All exercise types share the same shape; `options` is empty for
fill-in-the-blank, `correct_answer` is the index for multiple-choice.

**Alternatives considered**:
- Separate table per exercise type: rejected — too many tables, complicated joins.
- Single `content` JSONB blob: rejected — untyped, unqueryable for analytics.

---

## Decision 5: MCP Tools for Course Generation

**Decision**: Add three new MCP tools to `cmd/mcp/main.go`:
- `get_courses` — reads published/draft courses (filterable by language, level).
- `create_course_draft` — inserts a full course tree (course + lessons + content blocks + exercises) atomically.
- `analyze_grammar_coverage` — returns list of grammar topics found in existing courses for a given language/level, plus AI-suggested missing topics.

**Rationale**: Principle II requires all AI→DB writes to go through MCP. The admin REST endpoint
for generation will call out to the Claude API, and Claude will call back into the MCP server
to write the course. This keeps the write path auditable and consistent with the existing
`add_words`/`add_paragraphs` pattern.

**Alternatives considered**:
- Admin handler writes directly to DB after Claude returns JSON: acceptable for paragraphs
  (small) but for a full course tree it bypasses the MCP contract. Rejected to stay consistent
  with Principle II.

---

## Decision 6: Language-Agnostic Design

**Decision**: Add a `languages` table. All course data references `language_id`. The existing
`Word` and `Paragraph` models get a nullable `language_id` column (default: Italian's ID)
so backward compatibility is maintained — existing 2000 words continue to work without a migration
that forces data backfill.

**Rationale**: Spec SC-005 requires zero code changes to add a new language. A `languages`
lookup table achieves this; adding Persian is a single `INSERT INTO languages`.

**Alternatives considered**:
- Language as a string column (e.g., `language_code VARCHAR(10)`): simpler but no referential
  integrity and no place to store language metadata (name, direction, etc.).
- Hard-coded enum: rejected — requires code change to add language, violates SC-005.

---

## Decision 7: CEFR Level Representation

**Decision**: CEFR level stored as a MySQL ENUM: `('A1','A2','B1-1','B1-2','B2','C1','C2')`.
Go model uses a typed string alias (`type CEFRLevel string`).

**Rationale**: Fixed set of values, enum enforces validity at DB level, human-readable in queries.
The `B1-1`/`B1-2` split is non-standard CEFR but is the user's explicit requirement.

---

## Decision 8: Skills Representation

**Decision**: Skills stored as a JSON array string per course and per exercise:
`["reading","writing","listening","speaking"]`. Go model uses `StringArray` (same pattern
as existing `IntArray`).

**Rationale**: A course or exercise can target multiple skills. JSON array in MySQL text column
avoids a many-to-many join table for this simple use case, consistent with existing `IntArray`
pattern.
