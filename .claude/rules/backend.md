---
paths:
  - server/**/*.go
---

# Backend rules

- Read the surrounding handler, service, model, and migration code before changing backend behavior.
- Keep API, model, and persistence changes consistent across handlers, services, models, and migrations when the task requires all layers.
- Prefer the existing Gin, GORM, and project-specific patterns over introducing new backend abstractions.
- For schema or data-shape changes, check whether related request handlers, models, and migration code must change together.
- Verify backend changes with the smallest relevant command, such as `go build ./...` or a targeted run path.
- Do not add speculative endpoints, config knobs, or compatibility layers unless explicitly requested.
