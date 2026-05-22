<!--
SYNC IMPACT REPORT
==================
Version change: (new) → 1.0.0
Modified principles: N/A (initial ratification)
Added sections: Core Principles (I–V), Technology Stack, Development Workflow, Governance
Removed sections: N/A
Templates requiring updates:
  ✅ .specify/templates/plan-template.md — Constitution Check gate references principles I–V
  ✅ .specify/templates/spec-template.md — no structural changes required; principles are compatible
  ✅ .specify/templates/tasks-template.md — no structural changes required; task phases align
Deferred TODOs: none
-->

# Italian Project Constitution

## Core Principles

### I. Simplicity First

Every feature MUST start at the simplest viable implementation.
Abstractions are introduced only when duplication across three or more concrete cases
justifies them. Premature generalization, unnecessary layers, and speculative design
are prohibited. If a straight function call suffices, no interface is created.

**Rationale**: This is a personal learning tool maintained by one developer.
Complexity compounds quickly without a team to manage it.

### II. MCP-Native AI Integration

The MCP server (`cmd/mcp/main.go`) is the ONLY sanctioned interface for AI-assistant
interactions with the database. All bulk vocabulary and paragraph ingestion MUST go
through MCP tools (`add_words`, `add_paragraphs`). Direct database writes from AI
agents outside of MCP are prohibited.

**Rationale**: Keeps the AI integration surface area small, auditable, and versioned
independently from the REST API.

### III. Persistent Data Integrity

All application state MUST be stored in MySQL via GORM models. Soft-deletes (GORM's
`DeletedAt`) MUST be preserved — hard deletes require explicit justification.
Practice-tracking fields (`PracticeCount`, `HowManyFalse`, `LastPracticeFalse`) are
authoritative learning metrics and MUST NOT be reset silently.

**Rationale**: Vocabulary and practice history are the core value of the application;
data loss directly harms the learning workflow.

### IV. Clean REST Contracts

The Go/Fiber REST API MUST follow these rules:
- JSON in, JSON out for all endpoints.
- HTTP status codes MUST accurately reflect outcomes (201 Created, 404 Not Found, etc.).
- Request binding errors MUST return 400 with a machine-readable `{"error": "..."}` body.
- No business logic in route registration (`routes/`); logic lives in `handlers/`.

**Rationale**: The React frontend and any future clients depend on a predictable API
surface.

### V. Learning-Focused Frontend

The React/TypeScript frontend MUST prioritize the learner's workflow over engineering
elegance. Pages MUST be independently navigable. Speech synthesis (`useSpeech` hook)
MUST be available wherever Italian text is displayed. Feature additions to the frontend
MUST not degrade page-load performance for existing vocabulary lists.

**Rationale**: The frontend is the primary daily-use interface; slow or broken UX
directly disrupts practice sessions.

## Technology Stack

| Layer | Choice | Version |
|-------|--------|---------|
| Backend language | Go | 1.25+ |
| HTTP framework | Fiber | v3 |
| ORM | GORM | v1.31+ |
| Database | MySQL | 8+ |
| MCP SDK | mark3labs/mcp-go | v0.49+ |
| Frontend language | TypeScript | — |
| Frontend framework | React + Vite | — |

New dependencies MUST be justified against an existing dependency before being added.
The dependency list MUST stay minimal.

## Development Workflow

- Environment configuration MUST use `.env` (never committed; see `.env-example`).
- The MCP server binary (`mcp`) and the REST API (`server.go`) are built and run
  independently.
- Database schema changes MUST be handled via GORM auto-migrate; raw SQL migrations
  require explicit justification.
- All changes MUST be committed to the `master` branch on
  `github.com/MohammadLatifi/italian-project` under the personal account
  (`mohammadlatifi1993@gmail.com`).

## Governance

This constitution supersedes all other written or implied development practices for
this project. Amendments require:

1. Identifying which principle or section is affected.
2. Updating this file with a new version number (semantic versioning).
3. Updating the Sync Impact Report comment at the top of this file.
4. Committing the change with message:
   `docs: amend constitution to vX.Y.Z (<reason>)`

Versioning policy:
- **MAJOR**: Removal or incompatible redefinition of a principle.
- **MINOR**: New principle or section added.
- **PATCH**: Wording clarification, typo fix, non-semantic refinement.

**Version**: 1.0.0 | **Ratified**: 2026-05-22 | **Last Amended**: 2026-05-22
