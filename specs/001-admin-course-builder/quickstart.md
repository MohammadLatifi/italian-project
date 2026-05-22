# Quickstart: Admin Course Builder

## Prerequisites

- Go 1.25+, Node 20+, MySQL 8+
- `.env` populated (see `.env-example`)
- Existing database with Italian words already seeded

## 1. Run Database Migration

Start the backend — GORM auto-migrate adds all new tables on startup:

```bash
go run server.go
```

On first run you should see no errors and the following new tables created:
`languages`, `courses`, `lessons`, `content_blocks`, `exercises`, `lesson_attempts`, `grammar_suggestions`

The `it` (Italian) language row is seeded automatically.

## 2. Set Admin Credentials

Add to your `.env`:

```
ADMIN_USERNAME=admin
ADMIN_PASSWORD=your-secure-password
JWT_SECRET=your-random-secret-string
```

## 3. Verify Admin Login

```bash
curl -X POST http://localhost:3000/api/admin/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"your-secure-password"}'
# → {"token":"eyJ..."}
```

## 4. Generate Your First Course

```bash
curl -X POST http://localhost:3000/api/admin/courses/generate \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <token>" \
  -d '{
    "language_code": "it",
    "level": "A1",
    "skills": ["reading", "listening"],
    "topic": "Greetings and daily expressions"
  }'
# → {"ID":1,"status":"draft","title":"...","lesson_count":3}
```

## 5. Publish the Course

```bash
curl -X POST http://localhost:3000/api/admin/courses/1/publish \
  -H "Authorization: Bearer <token>"
# → {"ID":1,"status":"published"}
```

## 6. Verify Learner Access

```bash
curl http://localhost:3000/api/courses?language=it
# → [...] (includes the course just published)
```

## 7. Start the Frontend

```bash
cd frontend && npm install && npm run dev
```

Navigate to `http://localhost:5173/courses` to see the published course.
Navigate to `http://localhost:5173/admin` to access the admin panel.

## 8. Check Grammar Coverage

```bash
curl "http://localhost:3000/api/admin/grammar-suggestions?language=it&level=A1" \
  -H "Authorization: Bearer <token>"
# → [...] list of AI-suggested missing topics
```

## 9. MCP Server (for AI-assisted generation)

The MCP server exposes the new course tools automatically on rebuild:

```bash
go build -o mcp ./cmd/mcp/main.go && ./mcp
```

Confirm new tools are registered by checking MCP tool list in Claude's MCP panel.
New tools: `get_courses`, `create_course_draft`, `analyze_grammar_coverage`
