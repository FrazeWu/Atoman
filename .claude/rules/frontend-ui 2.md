---
paths:
  - web/**/*.vue
  - web/**/*.ts
  - web/**/*.tsx
---

# Frontend rules

- Use the existing `A*` UI primitives before creating page-local alternatives.
- Prefer `ABtn`, `AInput`, `ATextarea`, `ASelect`, `AModal`, `ADropdown`, `APopover`, and `AConfirm` where applicable.
- Prefer `variant="primary|secondary|danger|ghost"` and `size="sm|md|lg"` for new `ABtn` usage.
- Keep labels above controls; do not rely on placeholders as labels.
- Prefer design tokens from `web/src/style.css` over hard-coded structural UI colors.
- Prefer scoped classes over large inline layout styles in templates.
- Do not introduce Naive UI.
