---
paths:
  - web/**/*.vue
  - web/**/*.ts
  - web/**/*.tsx
---

# Frontend rules

- Do not create a separate page-level visual system or component visual language. Small scoped classes for page layout or composing existing components are allowed, but visual tokens, control appearance, and interaction styles must reuse the existing `A*` UI primitives from `web/src/components/ui` and the design tokens and global classes defined in `web/src/style.css` first.
- Prefer `ABtn`, `AInput`, `ATextarea`, `ASelect`, `AModal`, `ADropdown`, `APopover`, and `AConfirm` where applicable.
- Prefer `variant="primary|secondary|danger|ghost"` and `size="sm|md|lg"` for new `ABtn` usage.
- Keep labels above controls; do not rely on placeholders as labels.
- Prefer design tokens from `web/src/style.css` over hard-coded structural UI colors.
- Prefer scoped classes over large inline layout styles in templates.
- Do not introduce Naive UI.
