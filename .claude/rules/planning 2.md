---
paths:
  - docs/superpowers/specs/**/*.md
  - docs/superpowers/plans/**/*.md
---

# Planning rules

- Treat `docs/superpowers/specs/` and `docs/superpowers/plans/` as the only source of truth for planning.
- Keep `specs/` for durable requirements and design decisions.
- Keep `plans/` for executable implementation steps and verification.
- When updating planning docs, prefer concrete next steps and validation points over abstract strategy language.
- If implementation status changes, update the corresponding implementation plan; if product or architecture decisions change, update the corresponding spec.
- Do not create or preserve authoritative planning docs under `plan/`.
