# Learner API Contracts

All endpoints are prefixed `/api/`.
All requests that modify progress MUST include `X-Learner-Token: <uuid>` header.
All responses are `Content-Type: application/json`.

---

## Courses

### GET /api/courses

List published courses. Filterable by language and level.

**Query params**: `language` (code), `level`

**Response 200**:
```json
[
  {
    "ID": 42,
    "language_code": "it",
    "level": "A1",
    "title": "Greetings and Introductions",
    "description": "...",
    "skills": ["reading", "listening"],
    "lesson_count": 3
  }
]
```

---

### GET /api/courses/:id

Get course overview with lesson list (no content blocks or exercises — those load per lesson).

**Response 200**:
```json
{
  "ID": 42,
  "title": "...",
  "level": "A1",
  "skills": ["reading"],
  "lessons": [
    { "ID": 1, "title": "Lesson 1 — Greetings", "position": 0 }
  ]
}
```

**Response 404**: `{ "error": "course not found" }`

---

## Lessons

### GET /api/courses/:courseId/lessons/:lessonId

Get full lesson content including content blocks and exercises.
If `X-Learner-Token` header is present, also returns the learner's most recent attempt
for this lesson (if any).

**Response 200**:
```json
{
  "ID": 1,
  "title": "Lesson 1 — Greetings",
  "content_blocks": [
    {
      "ID": 10,
      "type": "grammar-note",
      "content": { "title": "...", "body": "..." },
      "position": 0
    }
  ],
  "exercises": [
    {
      "ID": 20,
      "type": "multiple-choice",
      "question": "How do you say 'Good morning'?",
      "options": ["Buonasera", "Buongiorno", "Ciao", "Arrivederci"],
      "skills": ["reading"],
      "position": 0
    }
  ],
  "current_attempt": {
    "ID": 5,
    "current_position": 2,
    "answers": { "20": "1" },
    "completed": false
  }
}
```

Note: `correct_answer` and `explanation` are NOT returned here — they are returned
only in the submit response to prevent client-side cheating.

---

## Lesson Attempts

### POST /api/courses/:courseId/lessons/:lessonId/attempts

Start a new lesson attempt. Requires `X-Learner-Token`.

**Response 201**:
```json
{ "ID": 7, "lesson_id": 1, "current_position": 0, "completed": false }
```

---

### PUT /api/courses/:courseId/lessons/:lessonId/attempts/:attemptId/answer

Submit an answer for one exercise. Returns feedback immediately.

**Request**:
```json
{ "exercise_id": 20, "answer": "1" }
```

**Response 200**:
```json
{
  "correct": true,
  "correct_answer": "1",
  "explanation": null,
  "next_position": 3
}
```

**Response 200 (wrong)**:
```json
{
  "correct": false,
  "correct_answer": "1",
  "explanation": "'Buongiorno' is used in the morning; 'Buonasera' is for the evening.",
  "next_position": 3
}
```

---

### POST /api/courses/:courseId/lessons/:lessonId/attempts/:attemptId/complete

Mark the attempt as completed and calculate final score.

**Response 200**:
```json
{
  "ID": 7,
  "score": 8,
  "total": 10,
  "completed": true,
  "completed_at": "2026-05-22T14:30:00Z"
}
```
