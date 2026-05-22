# MCP Tool Contracts: Course Builder

These tools are added to `cmd/mcp/main.go` and used by Claude during admin-triggered
course generation. They follow the same pattern as the existing `add_words`/`get_words` tools.

---

## get_courses

Read existing courses from the database.

**Input schema**:
```json
{
  "language_code": { "type": "string", "description": "Filter by language code (e.g. 'it')" },
  "level":         { "type": "string", "description": "Filter by CEFR level (e.g. 'B1-1')" },
  "status":        { "type": "string", "description": "Filter by status: draft|published|archived" },
  "limit":         { "type": "number", "description": "Max records (default 20)" }
}
```

**Returns**: JSON array of courses with their lessons (titles only, no content).

---

## create_course_draft

Insert a fully structured course draft in one atomic operation.
Claude calls this to persist the generated course it has assembled.

**Input schema**:
```json
{
  "course": {
    "type": "object",
    "required": true,
    "properties": {
      "language_code": { "type": "string" },
      "level":         { "type": "string" },
      "title":         { "type": "string" },
      "topic":         { "type": "string" },
      "description":   { "type": "string" },
      "skills":        { "type": "array", "items": { "type": "string" } },
      "lessons": {
        "type": "array",
        "items": {
          "title": { "type": "string" },
          "content_blocks": {
            "type": "array",
            "items": {
              "type": { "type": "string" },
              "content": { "type": "object" }
            }
          },
          "exercises": {
            "type": "array",
            "items": {
              "type":           { "type": "string" },
              "question":       { "type": "string" },
              "options":        { "type": "array", "items": { "type": "string" } },
              "correct_answer": { "type": "string" },
              "explanation":    { "type": "string" },
              "skills":         { "type": "array", "items": { "type": "string" } }
            }
          }
        }
      }
    }
  }
}
```

**Returns**: `"Successfully created course draft ID=<id> with <N> lessons"`

---

## analyze_grammar_coverage

Examine existing published courses for a language/level and return covered grammar topics
plus AI-suggested missing ones. Suggestions are persisted as `GrammarSuggestion` records.

**Input schema**:
```json
{
  "language_code": { "type": "string", "required": true },
  "level":         { "type": "string", "required": true }
}
```

**Returns**: JSON with two arrays:
```json
{
  "covered_topics": ["Present tense", "Articles", "Plural nouns"],
  "suggested_missing": [
    {
      "topic": "Passato prossimo",
      "rationale": "Essential A2 past tense; no course covers it yet."
    }
  ]
}
```

---

## Prompt Documentation (System Prompt for Course Generation)

The following system prompt is used by `handlers/admin_course.go` when calling Claude
to generate a course. It instructs Claude to use the MCP tools above.

```
You are an expert language course designer.
You have access to three MCP tools:
  - get_courses: read existing courses to avoid duplication
  - create_course_draft: persist your generated course
  - analyze_grammar_coverage: check what grammar has been covered

When generating a course:
1. Call get_courses to understand what already exists at this level.
2. Design 2–4 lessons. Each lesson must have:
   - 1–2 content blocks (grammar-note, vocabulary-list, or conversation-example)
   - 2–4 exercises (mix of multiple-choice and fill-in-the-blank)
   - Each exercise must be tagged with the relevant skill(s) from: reading, writing, listening, speaking
3. Call create_course_draft with the complete course structure.
4. Return only the course ID from the tool response.

Rules:
- Every exercise must have a correct_answer and an explanation for wrong answers.
- For multiple-choice, correct_answer is the 0-based index as a string ("0", "1", "2", "3").
- For fill-in-the-blank, correct_answer is the expected word or phrase.
- Content must match the requested CEFR level vocabulary and grammar complexity.
- Never invent vocabulary word IDs; only reference word_ids that exist in the database.
```
