# Admin API Contracts

All endpoints are prefixed `/api/admin/`.
All protected endpoints require `Authorization: Bearer <jwt>` header.
All responses are `Content-Type: application/json`.

---

## Authentication

### POST /api/admin/login

Login with admin credentials.

**Request**:
```json
{ "username": "admin", "password": "secret" }
```

**Response 200**:
```json
{ "token": "<jwt>" }
```

**Response 401**:
```json
{ "error": "invalid credentials" }
```

---

## Courses

### GET /api/admin/courses

List courses with optional filters.

**Query params**: `language` (code), `level`, `skill`, `status`

**Response 200**:
```json
[
  {
    "ID": 1,
    "language_id": 1,
    "language_code": "it",
    "level": "A1",
    "title": "Greetings and Introductions",
    "topic": "Daily greetings",
    "status": "published",
    "skills": ["reading", "listening"],
    "CreatedAt": "2026-05-22T10:00:00Z",
    "UpdatedAt": "2026-05-22T10:00:00Z"
  }
]
```

---

### POST /api/admin/courses/generate

Trigger AI generation of a course draft. Synchronous — response returns when draft is saved.

**Request**:
```json
{
  "language_code": "it",
  "level": "A1",
  "skills": ["reading", "listening"],
  "topic": "Greetings and Introductions"
}
```

**Response 201**:
```json
{ "ID": 42, "status": "draft", "title": "...", "lesson_count": 3 }
```

**Response 400**:
```json
{ "error": "level must be one of A1 A2 B1-1 B1-2 B2 C1 C2" }
```

**Response 502**:
```json
{ "error": "Claude API error: ..." }
```

---

### GET /api/admin/courses/:id

Get full course including lessons, content blocks, and exercises.

**Response 200**:
```json
{
  "ID": 42,
  "level": "A1",
  "title": "...",
  "status": "draft",
  "skills": ["reading"],
  "lessons": [
    {
      "ID": 1,
      "title": "Lesson 1",
      "position": 0,
      "content_blocks": [...],
      "exercises": [...]
    }
  ]
}
```

**Response 404**: `{ "error": "course not found" }`

---

### PATCH /api/admin/courses/:id

Update editable fields of a draft course. Only `title`, `description`, `topic` are editable
post-generation without re-generating.

**Request** (all fields optional):
```json
{ "title": "Updated title", "description": "A short intro course." }
```

**Response 200**: Updated course object.
**Response 400**: `{ "error": "only draft courses can be edited" }`

---

### POST /api/admin/courses/:id/publish

Publish a draft course. Validates at least one lesson with one exercise exists.

**Response 200**: `{ "ID": 42, "status": "published" }`
**Response 400**: `{ "error": "course must have at least one lesson with one exercise" }`

---

### POST /api/admin/courses/:id/archive

Archive a published course. Learners can no longer access it.

**Response 200**: `{ "ID": 42, "status": "archived" }`

---

## Grammar Suggestions

### GET /api/admin/grammar-suggestions

Get AI-suggested missing grammar topics for a language/level.
Triggers a Claude API call if no pending suggestions exist for this combination.

**Query params**: `language` (code, required), `level` (required)

**Response 200**:
```json
[
  {
    "ID": 1,
    "language_id": 1,
    "level": "B1-1",
    "topic": "Imperfetto vs Passato Prossimo",
    "rationale": "No course covers the contrast between these past tenses at this level.",
    "status": "pending"
  }
]
```

---

### POST /api/admin/grammar-suggestions/:id/accept

Accept a suggestion and immediately begin course generation for it.
Equivalent to calling `POST /api/admin/courses/generate` with the suggestion's data.

**Response 201**: New draft course object (same shape as generate response).
**Response 404**: `{ "error": "suggestion not found" }`

---

### POST /api/admin/grammar-suggestions/:id/dismiss

Dismiss a suggestion so it no longer appears.

**Response 200**: `{ "ID": 1, "status": "dismissed" }`
