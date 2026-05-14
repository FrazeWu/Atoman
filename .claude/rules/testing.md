---
paths:
  - web/**/*test*
  - web/**/*spec*
  - server/**/*test*
  - server/**/*spec*
---

# Testing rules

- Prefer the smallest verification step that proves the requested behavior changed correctly.
- When fixing a bug, prefer reproducing the failing path before claiming success.
- Do not weaken tests just to make them pass.
- Keep test changes scoped to the requested behavior and any direct fallout from that change.
- If no automated test exists for the affected area, say so clearly and use the most relevant build, type-check, or manual verification available.
