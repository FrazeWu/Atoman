# Timeline Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Bring the Timeline feature from a partially implemented event-and-person explorer to the planned archive-style timeline system with aligned data model, browse flows, and source-aware verification.

**Architecture:** Treat `plan/timeline_task_plan.md` as the authoritative scope and phase ledger, `plan/timeline_findings.md` as the product and data-model decision record, and `plan/timeline_progress.md` as the prior sequencing guidance. The current repository already contains timeline backend models, handlers, store logic, routes, and three UI surfaces, so implementation should proceed as a reconciliation-and-completion effort: first verify what exists, then close the highest-value gaps in the planned first-release slice without widening scope into BCE support, public submissions, or a full knowledge graph.

**Tech Stack:** Vue 3, TypeScript, Vite, Pinia, Vue Router, OpenLayers, Go, Gin, GORM, SQLite/PostgreSQL

---

## Upstream Sources

- `plan/timeline_task_plan.md`
- `plan/timeline_findings.md`
- `plan/timeline_progress.md`
- Style reference only: `docs/superpowers/plans/2026-05-14-blog-implementation.md`

## Normalized Current State

The upstream planning docs agree on the Timeline module’s identity: it is not a blog-like content stream, but an archive-style event system centered on historical events, related people, places, tags, sources, and visual browsing. The recommended first-release slice is consistent across all three source files: event CRUD, a timeline view, event detail, person and place associations, source visibility, and search/filtering.

The current codebase is ahead of the planning docs in some areas and behind in others. Timeline code already exists in:

- backend model definitions: `server/internal/model/timeline.go`
- backend routes and CRUD handlers: `server/internal/handlers/timeline_handler.go`
- frontend store: `web/src/stores/timeline.ts`
- timeline entry view: `web/src/views/timeline/TimelineHomeView.vue`
- person list view: `web/src/views/timeline/PersonListView.vue`
- person map/trajectory view: `web/src/views/timeline/PersonMapView.vue`
- frontend routing: `web/src/router.ts`
- frontend shared types: `web/src/types.ts`
- backend migration wiring: `server/migrate_db.go`

However, the implementation does not yet match the richer planning contract. The current backend model supports event date, end date, free-text location, optional coordinates, free-text source, category, tags, people, and person-location trajectories, but it does not yet expose dedicated `TimelinePlace`, `TimelineSource`, `TimelineRelation`, or date-precision semantics described in the planning docs. The current frontend already offers an event comparison board, map plotting, event modals, person creation, and person trajectory editing, but it still needs systematic verification and likely tightening around data-shape alignment, source visibility, filtering/search completeness, and design-system consistency.

For execution purposes, treat the feature as **partially implemented and requiring baseline verification before any structural changes**.

## Confirmed Scope To Preserve

- Timeline is event-centered, not post-centered.
- Historical events must remain browsable by time first, then by theme and geography.
- Sources must remain visible in the UI, not hidden as backend-only metadata.
- Event date ranges are part of the first-release contract.
- Person and place context should enrich event browsing, not replace it.
- Mobile should degrade to readable stacked flows rather than desktop-only visualization.

## Explicitly Deferred Scope

Do not add these in the first implementation pass unless product direction changes:

- BCE / 公元前 date support
- user-submitted public drafts for moderation
- full event-to-event causal relation taxonomy
- dedicated map-first product mode beyond the current timeline-plus-map browse flow
- knowledge-graph exploration beyond lightweight related-entity linking

## Verified Repository Baseline To Reconcile

### Backend files already participating in timeline
- `server/internal/model/timeline.go`
- `server/internal/handlers/timeline_handler.go`
- `server/migrate_db.go`
- `server/cmd/start_server/main.go`

### Frontend files already participating in timeline
- `web/src/stores/timeline.ts`
- `web/src/views/timeline/TimelineHomeView.vue`
- `web/src/views/timeline/PersonListView.vue`
- `web/src/views/timeline/PersonMapView.vue`
- `web/src/router.ts`
- `web/src/types.ts`
- `web/src/components/ui/ABtn.vue`
- `web/src/components/ui/AInput.vue`
- `web/src/components/ui/ATextarea.vue`
- `web/src/components/ui/ASelect.vue`
- `web/src/components/ui/AModal.vue`
- `web/src/components/ui/AConfirm.vue`
- `web/src/components/ui/APageHeader.vue`
- `web/src/components/ui/DatetimePicker.vue`

### Likely new files or focused additions needed for the planned target
- `server/internal/model/timeline.go` for schema reconciliation and additional entities
- `server/internal/handlers/timeline_handler.go` for search, source, relation, and browse refinements
- `web/src/stores/timeline.ts` for aligned client-side API surface
- `web/src/types.ts` for aligned timeline type definitions
- `web/src/views/timeline/TimelineHomeView.vue` for event-first browse refinements
- `web/src/views/timeline/PersonListView.vue` for person browse cleanup
- `web/src/views/timeline/PersonMapView.vue` for source-aware location editing and display
- `web/src/views/timeline/TimelineEventDetailView.vue` if event detail is split out of the current modal approach
- `web/src/views/timeline/TimelinePlaceListView.vue` if place browsing is introduced as a dedicated surface in scope
- `web/tests/timeline/*.spec.ts` for end-to-end regression coverage if Playwright tests are added

## Design Constraints From Project Instructions

- Prefer existing `A*` UI primitives over raw custom controls.
- Prefer design tokens from `web/src/style.css` over hardcoded UI structure colors.
- Avoid introducing Naive UI for new timeline work.
- Follow Vue SFC order: template, script setup, scoped style.
- Keep new logic focused instead of widening existing large files unnecessarily.
- Do not change unrelated modules while implementing timeline.

## Task 1: Revalidate the current timeline baseline

**Files:**
- Inspect: `plan/timeline_task_plan.md`
- Inspect: `plan/timeline_findings.md`
- Inspect: `plan/timeline_progress.md`
- Inspect: `server/internal/model/timeline.go`
- Inspect: `server/internal/handlers/timeline_handler.go`
- Inspect: `web/src/stores/timeline.ts`
- Inspect: `web/src/views/timeline/TimelineHomeView.vue`
- Inspect: `web/src/views/timeline/PersonListView.vue`
- Inspect: `web/src/views/timeline/PersonMapView.vue`
- Inspect: `web/src/router.ts`
- Inspect: `web/src/types.ts`
- Test: `web/package.json`
- Test: `server/go.mod`

- [ ] **Step 1: Reconcile planning claims against current code**

Create working notes with these exact headings:
```text
- Confirmed in plan
- Present in code already
- Missing or only partially implemented
- Explicitly deferred
```

Expected: a short reconciliation list proving which first-release timeline capabilities already exist and which still need implementation.

- [ ] **Step 2: Run the frontend type-check**

Run:
```bash
cd /Users/fafa/Documents/projects/Atoman/.claude/worktrees/agent-a967772d6ce334c8a/web && bun run type-check
```

Expected: successful TypeScript verification with no timeline-specific type regressions.

- [ ] **Step 3: Run the frontend production build**

Run:
```bash
cd /Users/fafa/Documents/projects/Atoman/.claude/worktrees/agent-a967772d6ce334c8a/web && bun run build
```

Expected: successful Vite production build confirming the timeline routes compile end-to-end.

- [ ] **Step 4: Run the backend build**

Run:
```bash
cd /Users/fafa/Documents/projects/Atoman/.claude/worktrees/agent-a967772d6ce334c8a/server && go build ./...
```

Expected: successful Go build confirming the existing timeline model and route wiring compile.

- [ ] **Step 5: Verify the currently wired routes and entities**

Inspect and record whether these are already live:
```text
- /timeline route
- /timeline/persons route
- /timeline/persons/:id route
- event CRUD handlers
- person CRUD handlers
- person location CRUD handlers
- migration wiring for timeline tables
```

Expected: a concrete baseline checklist before any model or UI restructuring begins.

## Task 2: Align the data model with the planned first-release scope

**Files:**
- Modify: `server/internal/model/timeline.go`
- Modify: `server/internal/handlers/timeline_handler.go`
- Modify: `server/migrate_db.go`
- Modify: `web/src/types.ts`
- Modify: `web/src/stores/timeline.ts`
- Test: `server/internal/model/timeline.go`
- Test: `web/src/types.ts`

- [ ] **Step 1: Define the minimal first-release timeline entities to support**

Use this exact target shape as the implementation contract for first release:
```text
TimelineEvent
- id
- title
- summary or description
- content
- start_date
- end_date
- date_precision
- category
- importance
- source visibility fields
- optional coordinates and display location
- tags
- publication status / visibility

TimelinePerson
- id
- name
- aliases optional for later
- bio
- birth_date
- death_date
- tags
- visibility

TimelinePlace
- id
- name
- aliases optional for later
- latitude
- longitude
- region
- description

TimelineSource
- id
- title or citation label
- url or archive reference
- source_type
- note

Event relations
- event to person
- event to place
- event to source
```

Expected: a constrained first-release schema contract that matches the planning docs without jumping to full graph complexity.

- [ ] **Step 2: Choose the compatibility strategy before editing models**

Pick one and document it in working notes before coding:
```text
Option A: rename existing fields in place and backfill compatibility
Option B: preserve current fields and add new optional fields for planned semantics
Option C: introduce join tables/entities while keeping current event/person CRUD payloads backward compatible
```

Recommended: Option C, because it lets the current UI keep functioning while first-release archive semantics grow incrementally.

Expected: a declared migration strategy that avoids breaking the current timeline pages mid-implementation.

- [ ] **Step 3: Add only the entities needed for first-release archive browsing**

When editing the model, target these additions before anything else:
```go
// pseudo-shape only; final code must follow repo conventions
// add dedicated place/source models or join records only if they are used by handlers/UI in later tasks
```

Expected: the model supports event-person-place-source relationships needed by the plan’s first-release scope.

- [ ] **Step 4: Keep frontend types in lockstep with backend changes**

Update `web/src/types.ts` so the client types name the same fields as the API contract.
Use this validation checklist while editing:
```text
- event date field names match server JSON exactly
- person location source field exists in types if server returns it
- place/source/relation types are defined before store methods use them
- no later task references a type name missing from types.ts
```

Expected: no type mismatch between timeline store calls and backend JSON payloads.

- [ ] **Step 5: Rebuild backend and frontend after schema alignment**

Run:
```bash
cd /Users/fafa/Documents/projects/Atoman/.claude/worktrees/agent-a967772d6ce334c8a/server && go build ./...
cd /Users/fafa/Documents/projects/Atoman/.claude/worktrees/agent-a967772d6ce334c8a/web && bun run type-check
```

Expected: both commands pass after model and type updates.

## Task 3: Complete event-first timeline browsing and event detail

**Files:**
- Modify: `web/src/stores/timeline.ts`
- Modify: `web/src/views/timeline/TimelineHomeView.vue`
- Modify: `web/src/router.ts`
- Modify: `web/src/types.ts`
- Create or Modify: `web/src/views/timeline/TimelineEventDetailView.vue`
- Test: `web/src/views/timeline/TimelineHomeView.vue`

- [ ] **Step 1: Decide whether event detail remains modal-based or becomes route-based**

Use this rule:
```text
If event detail needs related people, related places, source list, and related-event navigation, promote it to a dedicated route.
If it remains only title/date/content/source and quick edit/delete, the modal can stay.
```

Recommended: promote to a route-based detail view so the archive experience can support deep linking and related-entity navigation.

Expected: one clear event-detail direction before adding more detail UI.

- [ ] **Step 2: Bring the timeline home view back to the documented first-release responsibilities**

The view must cover all of these after implementation:
```text
- event list or source panel driven by time ordering
- year/category filtering
- visible source information on event detail path
- event creation and editing for authenticated users
- timeline visualization as the primary browse surface
- mobile-readable fallback when compare/map density is too high
```

Expected: the page is still visually rich, but its behavior clearly matches the planning docs rather than a purely experimental compare tool.

- [ ] **Step 3: Add event detail fields that match the archive contract**

Ensure the event detail experience visibly presents:
```text
- title
- date or date range
- category
- display location
- source or source list
- tags
- body content
- related people if present
- related places if present
```

Expected: readers can understand and verify an event without reopening editor forms.

- [ ] **Step 4: Add a stable deep-link entry point for event detail**

If using a dedicated detail page, wire a route like this:
```text
/timeline/events/:id
```

Expected: event detail is directly shareable and reachable from timeline cards, map markers, and any future person/place association pages.

- [ ] **Step 5: Verify the browse flow in the browser**

Manually verify these checkpoints:
```text
- open /timeline
- filter by year range and category
- open at least one event detail
- confirm source information is visible
- create an event as an authenticated user
- edit the same event
- delete the same event
- confirm list/map/timeline state stays coherent after each mutation
```

Expected: the event-first user journey works from browse to detail to CRUD without losing archive context.

## Task 4: Complete person and place association flows

**Files:**
- Modify: `server/internal/handlers/timeline_handler.go`
- Modify: `web/src/stores/timeline.ts`
- Modify: `web/src/views/timeline/PersonListView.vue`
- Modify: `web/src/views/timeline/PersonMapView.vue`
- Create or Modify: `web/src/views/timeline/PlaceListView.vue`
- Create or Modify: `web/src/views/timeline/PlaceDetailView.vue`
- Modify: `web/src/router.ts`
- Modify: `web/src/types.ts`

- [ ] **Step 1: Preserve the existing person trajectory map, but align it with archive semantics**

The person detail flow must continue supporting:
```text
- ordered location history
- map plotting of locations
- source visible per location record
- create, edit, delete location entries for authorized users
```

Expected: person history remains useful and does not regress while broader timeline work lands.

- [ ] **Step 2: Add person-to-event navigation if missing**

A person page should include:
```text
- core identity block
- timeline or list of related events
- trajectory map
```

Expected: a reader can move from person context back into event context rather than being trapped in a map-only view.

- [ ] **Step 3: Introduce place browsing only if it is backed by actual event or location data**

Use this guardrail:
```text
Do not create decorative place pages with no event or person linkage.
Only add place routes if handlers and UI can show related events and/or person stays.
```

Expected: place pages, if added, are real archival navigation nodes instead of empty shells.

- [ ] **Step 4: Verify association loops in the browser**

Manually verify these checkpoints:
```text
- create a person
- add at least two locations with visible sources
- open the person map page
- click a mapped location and verify popup/source display
- navigate from event to related person if implemented
- navigate from person or place back to related events if implemented
```

Expected: event, person, and place context reinforce one another instead of behaving like isolated CRUD islands.

## Task 5: Add search, filtering, and source-aware retrieval endpoints

**Files:**
- Modify: `server/internal/handlers/timeline_handler.go`
- Modify: `web/src/stores/timeline.ts`
- Modify: `web/src/views/timeline/TimelineHomeView.vue`
- Modify: `web/src/views/timeline/PersonListView.vue`
- Modify: `web/src/types.ts`
- Modify: `web/src/composables/useApi.ts`

- [ ] **Step 1: Normalize the first-release read APIs around the planning docs**

Target the following API surface as the release contract, whether via exact endpoints or equivalent handler organization:
```text
GET /api/timeline/events
GET /api/timeline/events/:id
POST /api/timeline/events
PUT /api/timeline/events/:id
DELETE /api/timeline/events/:id
GET /api/timeline/persons
GET /api/timeline/persons/:id
GET /api/timeline/search
GET /api/timeline/visualization
```

Expected: timeline browsing has stable backend contracts for UI work and later testing.

- [ ] **Step 2: Add search only across entities that the UI can display well**

Search results should support these at minimum:
```text
- event title
- event description or summary
- person name
- place name if place entities are added
- category and tag filtering where practical
```

Expected: search returns useful archive discovery results without requiring a full-text overhaul.

- [ ] **Step 3: Ensure source-aware filters are visible if source quality is part of the returned payload**

If the API exposes source metadata such as type or citation count, the UI may add filters like:
```text
- has source
- source type
- category
- tag
- year range
```

Expected: users can narrow the archive while preserving the source-traceable character described in the planning docs.

- [ ] **Step 4: Centralize timeline endpoint builders if useful**

If `web/src/composables/useApi.ts` is expanded, add a timeline section only when the store can consume it immediately.

Expected: endpoint strings do not drift across multiple timeline files.

- [ ] **Step 5: Verify search and filter behavior manually**

Manually verify these checkpoints:
```text
- search by event title keyword
- search by person name keyword
- filter events by category
- filter events by year range
- confirm no-result state is readable
- confirm source display still appears after filtering/searching
```

Expected: the archive discovery loop works for both exploratory and targeted lookup behavior.

## Task 6: Tighten design-system consistency and mobile fallback

**Files:**
- Modify: `web/src/views/timeline/TimelineHomeView.vue`
- Modify: `web/src/views/timeline/PersonListView.vue`
- Modify: `web/src/views/timeline/PersonMapView.vue`
- Modify: timeline-specific extracted components if created during implementation
- Test: timeline view files above

- [ ] **Step 1: Replace new raw form controls with existing `A*` primitives where feasible**

Use this checklist while cleaning up UI:
```text
- ABtn for buttons
- AInput for single-line inputs
- ATextarea for multiline inputs
- ASelect for selects
- AModal for modals
- AConfirm for destructive confirmation
```

Expected: no unnecessary expansion of ad hoc controls in timeline screens.

- [ ] **Step 2: Pull layout styling out of new inline styles where practical**

Focus on newly added or heavily edited timeline code, especially repeated layout blocks.

Expected: new work follows the project instruction to prefer scoped classes over growing inline style usage.

- [ ] **Step 3: Make the mobile fallback intentionally readable**

Check these behaviors on a narrow viewport:
```text
- filters wrap without becoming unusable
- compare or timeline cards stack sensibly
- detail pages remain readable without map interaction
- person map flow still exposes core information even when map space is tight
```

Expected: timeline is still useful on mobile even if visualization density is reduced.

- [ ] **Step 4: Run one browser review focused only on UI consistency**

Manually verify these checkpoints:
```text
- page headers match archive-style UI
- buttons use consistent variants and sizes
- modal headers and footers are consistent
- hardcoded structure colors are minimized in newly touched code
```

Expected: the feature looks like part of Atoman, not a visually separate prototype.

## Task 7: Add regression coverage for the stabilized timeline slice

**Files:**
- Create: `web/tests/timeline/timeline-smoke.spec.ts`
- Create: `web/tests/timeline/person-map.spec.ts`
- Modify: `web/playwright.config.ts` if setup is required
- Test: timeline browser flows

- [ ] **Step 1: Add a smoke test only for stable, shippable timeline flows**

Cover these paths first:
```text
- visit /timeline
- confirm event list or empty state renders
- open an event detail path if seed data exists
- visit /timeline/persons
- confirm person list or empty state renders
```

Expected: a minimal e2e net for the feature’s public browse surfaces.

- [ ] **Step 2: Add one authenticated CRUD path only if test fixtures can support it reliably**

Candidate path:
```text
- log in with a fixture account
- create timeline event
- verify event appears
- delete event
```

Expected: one durable mutation test is better than many flaky ones.

- [ ] **Step 3: Add one person-map regression test only if OpenLayers timing is stable in CI**

Candidate path:
```text
- visit /timeline/persons/:id
- confirm location list renders
- confirm map container renders
```

Expected: map-page regressions are caught without overfitting to map pixel details.

- [ ] **Step 4: Run e2e tests if and only if the required local environment exists**

Run:
```bash
cd /Users/fafa/Documents/projects/Atoman/.claude/worktrees/agent-a967772d6ce334c8a/web && bun run test:e2e
```

Expected: passing timeline smoke tests, or an explicit note documenting why the local environment was not sufficient to run them.

## Task 8: Final verification and documentation handoff

**Files:**
- Inspect: all timeline files touched in earlier tasks
- Modify: `plan/timeline_progress.md` only if the team’s execution workflow requires progress logging during implementation

- [ ] **Step 1: Re-run the core verification commands**

Run:
```bash
cd /Users/fafa/Documents/projects/Atoman/.claude/worktrees/agent-a967772d6ce334c8a/web && bun run type-check
cd /Users/fafa/Documents/projects/Atoman/.claude/worktrees/agent-a967772d6ce334c8a/web && bun run build
cd /Users/fafa/Documents/projects/Atoman/.claude/worktrees/agent-a967772d6ce334c8a/server && go build ./...
```

Expected: all three commands pass on the final implementation state.

- [ ] **Step 2: Perform the exact browser verification walkthrough**

Walk through these steps in order:
```text
1. Open /timeline.
2. Confirm the page renders without console errors.
3. Filter by year range.
4. Filter by category.
5. Open event detail and verify source visibility.
6. Create an event.
7. Edit the event.
8. Delete the event.
9. Open /timeline/persons.
10. Create a person.
11. Open that person’s page.
12. Add two location records with sources.
13. Click each location in the list and confirm map focus changes.
14. Verify popups display location metadata.
15. If place pages exist, navigate from an event or person to a place page and back.
16. Verify mobile layout behavior in a narrow viewport.
```

Expected: the implemented first-release slice works end-to-end for event browsing, source visibility, and person trajectory context.

- [ ] **Step 3: Capture any still-deferred work explicitly**

If any of these remain unimplemented, record them clearly instead of implying completion:
```text
- BCE support
- dedicated place entity pages
- event-to-event causal relations
- public submission drafts
- advanced knowledge graph exploration
```

Expected: the release state is honest, with no hidden scope drift.

## Verification

After execution, verify all of the following:

1. `docs/superpowers/plans/2026-05-14-timeline-implementation.md` remains self-contained and executable without reopening upstream planning files.
2. `cd /Users/fafa/Documents/projects/Atoman/.claude/worktrees/agent-a967772d6ce334c8a/web && bun run type-check` passes.
3. `cd /Users/fafa/Documents/projects/Atoman/.claude/worktrees/agent-a967772d6ce334c8a/web && bun run build` passes.
4. `cd /Users/fafa/Documents/projects/Atoman/.claude/worktrees/agent-a967772d6ce334c8a/server && go build ./...` passes.
5. Browser validation covers `/timeline`, `/timeline/persons`, and `/timeline/persons/:id`.
6. Event detail visibly exposes source information.
7. Event CRUD works for an authenticated user.
8. Person creation and person-location CRUD work for an authenticated user.
9. Map and list interactions stay in sync on the person detail flow.
10. Any still-deferred archive features are explicitly documented rather than implied complete.

## Commit

```bash
git add /Users/fafa/Documents/projects/Atoman/.claude/worktrees/agent-a967772d6ce334c8a/docs/superpowers/plans/2026-05-14-timeline-implementation.md
git commit -m "docs: add timeline superpowers implementation plan"
```
