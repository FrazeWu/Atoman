# Atoman

Atoman is a self-hostable platform for publishing and discussion: Studio (blog/video/podcast), RSS feeds, forum, debate, and music archive.

## Tech Stack

| Layer | Stack |
|---|---|
| Frontend | Vue 3.5, Vite, TypeScript 5.9 |
| Frontend State | Pinia 3, Vue Router 4 |
| Styling | Tailwind CSS v4 |
| Backend | Go, Gin, GORM, JWT |
| Database | PostgreSQL (prod), SQLite (dev) |
| Storage | S3-compatible storage, MinIO in dev |
| Infra | Docker Compose, Nginx, supervisord |

## Commands

```bash
# frontend
bun install
bun run dev
bun run build
bun run type-check
bun run lint

# backend
go build ./...
go run cmd/start_server/main.go
go run cmd/create_admin/main.go
```

## Directory Architecture

| Path | Responsibility |
|---|---|
| `web/` | Vue frontend |
| `server/` | Go backend |
| `docs/superpowers/specs/` | Durable requirements, design decisions, architecture knowledge |
| `docs/superpowers/plans/` | Execution-ready implementation plans and verification steps |
| `doc/` | Legacy or supplemental documentation |
| `.claude/` | Claude local config, skills, rules |

## Planning System

`docs/superpowers/specs/` and `docs/superpowers/plans/` are the only official planning system.

- Put durable requirements, product rules, architecture decisions, and long-lived design knowledge in `docs/superpowers/specs/`.
- Put executable implementation steps, sequencing, rollout details, verification commands, and browser validation checklists in `docs/superpowers/plans/`.
- Do not use `plan/` as an authoritative planning location.
- Do not create new planning docs under `doc/`.
- If a decision changes product behavior or architecture, update the relevant spec.
- If execution order or validation changes, update the relevant implementation plan.

## Action Principles

- For tasks with 3 or more meaningful steps, propose a short plan first and wait for confirmation before editing.
- Read the existing implementation before changing a file; match existing patterns unless the task requires a deliberate change.
- Treat `CLAUDE.md` as global guidance only. Put path-specific or module-specific rules in `.claude/rules/`.
- Verify affected behavior before calling work complete. For frontend changes, prefer running the app and checking the user flow. For backend changes, run the smallest relevant build or test command.
- Stop and ask when requirements are ambiguous, when multiple valid interpretations exist, or when repeated attempts are not converging.
- Keep changes minimal. Do not add speculative abstractions, fallback behavior, or cleanup unrelated code.


