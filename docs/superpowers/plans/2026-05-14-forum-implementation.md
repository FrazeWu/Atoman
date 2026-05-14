# Forum Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Consolidate the forum upstream planning artifacts into one execution-ready implementation plan that preserves the confirmed product contract, revalidates the already-built forum surfaces, and makes any remaining implementation gaps explicit before further feature work.

**Architecture:** Treat `plan/forum_task_plan.md` as the scope baseline and decision ledger, `plan/forum_findings.md` as the product/data/API design record, and `plan/forum_progress.md` as historical implementation guidance plus proof that the planning phase ended before fresh build verification. The current codebase already contains a working forum stack across Vue views, a Pinia store, Gin handlers, and GORM models, so execution should proceed as a revalidation-and-gap-closure plan rather than a greenfield build.

**Tech Stack:** Vue 3, TypeScript, Vite, Pinia, Vue Router, Go, Gin, GORM, SQLite/PostgreSQL, shared `A*` UI components, shared `AEditor`

---

## Upstream Sources

- `plan/forum_task_plan.md`
- `plan/forum_findings.md`
- `plan/forum_progress.md`
- Current implementation files referenced by this plan:
  - `web/src/views/forum/ForumHomeView.vue`
  - `web/src/views/forum/ForumTopicView.vue`
  - `web/src/views/forum/ForumNewTopicView.vue`
  - `web/src/views/forum/ForumSearchView.vue`
  - `web/src/components/forum/ForumReplyNode.vue`
  - `web/src/stores/forum.ts`
  - `web/src/router.ts`
  - `web/src/types.ts`
  - `server/internal/model/forum.go`
  - `server/internal/handlers/forum_handler.go`
  - `server/internal/service/forum_migrate.go`
  - `server/internal/service/forum_mention_parser.go`

## Normalized Current State

The upstream planning trio says forum product definition is complete through “阶段 8：产品细节确认完成，可进入实现规划”, with explicit decisions for category structure, topic/reply behavior, governance, editing rules, question-mode behavior, tagging, reporting, and notifications. Unlike a pure plan-only module, the current repository already contains a substantial forum implementation: routes exist for `/forum`, `/forum/search`, `/forum/new`, and `/topic/:id`; the frontend includes category/topic/reply/search flows; the backend includes category/topic/reply CRUD plus like/bookmark/pin/close/draft endpoints; and the data model includes topic tags, reply paths, draft persistence, and mention parsing.

For execution purposes, treat the forum module as **implemented enough to require systematic verification against the upstream contract**. The work now is to confirm what matches the plan, identify where the implementation diverges from the product decisions, and tighten any missing surfaces before making claims of completion.

## Confirmed Scope To Preserve

- Forum uses a **single-level category structure**; no subcategories.
- Topic composition is **category + title + rich body + free-form tags**.
- Topic authoring uses **Tiptap/WYSIWYG through the shared `AEditor` surface**.
- Replies stay **lighter-weight than topics** and must support quoting, mentions, and image insertion expectations from upstream.
- Users must be authenticated; **anonymous posting is not allowed**.
- Core interaction baseline includes **like, bookmark, pin, close, search, and reply sorting**.
- Question-mode expectations include **best-answer capability without auto-closing the topic**.
- Governance expectations include **reporting, moderation, auto-collapse threshold behavior, anti-spam/rate limiting, and category-request workflow**.
- Notification linkage must preserve **reply-driven and `@username` mention-driven notifications**.
- No reputation/level system is part of the forum scope.
- No category-subscription feature should be added.

## Verified Architecture Decisions From Upstream And Code

- **Primary content model:** `ForumCategory`, `ForumTopic`, and `ForumReply` are the core persisted forum entities.
- **Interaction model:** likes and bookmarks are implemented directly; subscriptions were planned upstream but are not present in the currently verified handler/store surface.
- **Reply relationship model:** `ForumReply.parent_reply_id` is used as quoted-reply linkage, while `path` and `floor_number` support ordered threaded display and migration backfill.
- **Draft model:** topic and reply drafts are supported both locally and through backend draft endpoints.
- **Mention model:** `server/internal/service/forum_mention_parser.go` strips code blocks and extracts `@username` mentions for notification workflows.
- **UI routing model:** current forum identity is topic-centric using `/topic/:id` for detail and `/forum` for discovery/search/category filtering.
- **Current editor model:** both topic creation and reply creation currently mount the shared `AEditor`, even though the upstream contract says replies should stay text-plus-image lightweight.

## Critical Implementation Files To Revalidate

### Frontend
- `web/src/views/forum/ForumHomeView.vue`
- `web/src/views/forum/ForumTopicView.vue`
- `web/src/views/forum/ForumNewTopicView.vue`
- `web/src/views/forum/ForumSearchView.vue`
- `web/src/components/forum/ForumReplyNode.vue`
- `web/src/stores/forum.ts`
- `web/src/router.ts`
- `web/src/types.ts`
- `web/src/components/shared/AEditor.vue`

### Backend
- `server/internal/model/forum.go`
- `server/internal/handlers/forum_handler.go`
- `server/internal/service/forum_migrate.go`
- `server/internal/service/forum_mention_parser.go`
- `server/cmd/start_server/main.go`
- `server/migrate_db.go`

## Known Gaps To Validate Explicitly

These are not assumptions; they are checkpoints that require proof during execution:

- Upstream planned **best-answer / question-mode** behavior, but it is not visible in the currently verified forum view/store/handler files.
- Upstream planned **reporting / moderation review / auto-collapse threshold** behavior, but the currently verified handler surface only exposes pin/close/admin-category actions.
- Upstream planned **category request workflow**, but no verified route or page for applying for a new category was found in the reviewed files.
- Upstream planned **reply text + image** constraints, but the current reply form uses shared `AEditor`, so actual behavior must be verified rather than inferred.
- Upstream findings mention **subscription**, while task-plan decisions later state **no category subscription**; execution must preserve the later confirmed task-plan decision and only verify existing topic-level behavior if present.

## Task 1: Reconcile the upstream forum contract with the current codebase

**Files:**
- Inspect: `plan/forum_task_plan.md`
- Inspect: `plan/forum_findings.md`
- Inspect: `plan/forum_progress.md`
- Inspect: `web/src/views/forum/ForumHomeView.vue`
- Inspect: `web/src/views/forum/ForumTopicView.vue`
- Inspect: `web/src/views/forum/ForumNewTopicView.vue`
- Inspect: `web/src/stores/forum.ts`
- Inspect: `server/internal/model/forum.go`
- Inspect: `server/internal/handlers/forum_handler.go`

- [ ] **Step 1: Build a contract checklist from upstream planning**

Read the three upstream files and write a working checklist using this exact structure:
```text
Implemented and must preserve:
- single-level categories
- topic creation/editing
- quoting and @mention notifications
- latest / hottest / featured-or-equivalent sorting

Planned but not yet proven in code:
- best-answer behavior
- reporting and auto-collapse workflow
- category request workflow
- anti-spam / rate limiting

Potential contradictions to resolve:
- reply editor richness vs lightweight reply requirement
- subscription language in findings vs no category subscription in confirmed decisions
```

Expected: a reconciliation list that distinguishes hard requirements from unverified planning carryovers.

- [ ] **Step 2: Compare data and API surfaces against the checklist**

Inspect the verified model and handler files and record whether these surfaces exist:
```text
- category CRUD or admin creation path
- topic CRUD
- reply CRUD
- like and bookmark endpoints
- pin and close moderation endpoints
- draft persistence endpoints
- best-answer endpoint
- report endpoint
- moderation review endpoint
```

Expected: a clear map of which upstream promises are implemented today and which remain absent.

- [ ] **Step 3: Compare frontend surfaces against the same checklist**

Inspect the verified forum views/store and record whether these UI flows exist:
```text
- category filtering
- sort switching
- topic composer
- tag entry
- reply quoting
- topic like/bookmark actions
- search results flow
- best-answer controls
- report controls
- category request entry point
```

Expected: a code-backed UI inventory you can validate in the browser later.

## Task 2: Revalidate build health before any forum changes

**Files:**
- Test: `web/src/views/forum/**`
- Test: `web/src/stores/forum.ts`
- Test: `server/internal/model/forum.go`
- Test: `server/internal/handlers/forum_handler.go`
- Test: `server/internal/service/forum_migrate.go`

- [ ] **Step 1: Run the frontend type-check**

Run:
```bash
cd "/Users/fafa/Documents/projects/Atoman/.claude/worktrees/agent-a79bd39272f2a0a58/web" && bun run type-check
```

Expected: successful TypeScript verification with no forum-specific type regressions.

- [ ] **Step 2: Run the frontend production build**

Run:
```bash
cd "/Users/fafa/Documents/projects/Atoman/.claude/worktrees/agent-a79bd39272f2a0a58/web" && bun run build
```

Expected: successful build confirming all forum views and shared editor usage compile end-to-end.

- [ ] **Step 3: Run the backend build**

Run:
```bash
cd "/Users/fafa/Documents/projects/Atoman/.claude/worktrees/agent-a79bd39272f2a0a58/server" && go build ./...
```

Expected: successful Go build confirming current forum handlers, services, and models compile together.

- [ ] **Step 4: Start the application stack needed for browser verification**

Run the smallest setup that exposes both web and API locally. Prefer the project’s existing local flow:
```bash
cd "/Users/fafa/Documents/projects/Atoman/.claude/worktrees/agent-a79bd39272f2a0a58/web" && bun run dev
```

and in another terminal/session:
```bash
cd "/Users/fafa/Documents/projects/Atoman/.claude/worktrees/agent-a79bd39272f2a0a58/server" && go run cmd/start_server/main.go
```

Expected: both web and backend start cleanly so browser checks can target the current forum stack.

## Task 3: Revalidate discovery, category, and search behavior

**Files:**
- Test: `web/src/views/forum/ForumHomeView.vue`
- Test: `web/src/views/forum/ForumSearchView.vue`
- Test: `web/src/stores/forum.ts`
- Test: `web/src/router.ts`
- Test: `server/internal/handlers/forum_handler.go`

- [ ] **Step 1: Verify `/forum` discovery behavior**

Manually verify these checkpoints:
```text
- the page loads without runtime errors
- category list renders from live API data
- topic list renders with reply/view/activity metadata
- sort switching works for latest / top / active or the current implemented set
- category filtering updates the topic list correctly
- tag filtering updates the topic list correctly
```

Expected: the forum home behaves as the public discovery surface described upstream.

- [ ] **Step 2: Verify the search entry points**

Manually verify these checkpoints:
```text
- typing a query on `/forum` can reach search behavior
- `/forum/search?q=...` loads directly
- result count and topic rows render correctly
- clicking a result opens the topic detail route
```

Expected: the search flow is navigable both from direct URL and user interaction.

- [ ] **Step 3: Record any mismatch between planned and shipped sorting semantics**

Capture findings using this structure:
```text
- upstream required: 最新 / 最热 / 精华
- current UI shows: <actual tabs seen>
- current API supports: <actual sort params seen>
- follow-up needed: yes/no with concrete reason
```

Expected: no ambiguity remains about whether “featured” is fully implemented or only planned.

## Task 4: Revalidate topic creation, editing expectations, and draft behavior

**Files:**
- Test: `web/src/views/forum/ForumNewTopicView.vue`
- Test: `web/src/components/shared/AEditor.vue`
- Test: `web/src/stores/forum.ts`
- Test: `server/internal/handlers/forum_handler.go`
- Test: `server/internal/model/forum.go`

- [ ] **Step 1: Verify authenticated access to topic creation**

Manually verify these checkpoints:
```text
- unauthenticated navigation to `/forum/new` redirects to login
- authenticated navigation opens the topic composer
- category loading state behaves correctly when categories are unavailable
```

Expected: the no-anonymous-posting contract is enforced at the route and API layers.

- [ ] **Step 2: Verify topic composition flow**

Manually verify these checkpoints:
```text
- category selection works
- title and body validation both trigger correctly
- tags can be added and removed
- the shared editor supports the intended topic authoring experience
- successful submission redirects to `/topic/:id`
```

Expected: the core topic publishing flow matches the upstream baseline.

- [ ] **Step 3: Verify draft persistence behavior**

Manually verify these checkpoints:
```text
- draft autosave occurs while composing
- reloading the page restores saved draft content
- clearing the draft removes the restored content state
- if backend draft endpoints are wired into the current flow, confirm cross-session persistence; otherwise mark backend draft support as present but not yet mounted in UI
```

Expected: local draft protection is proven, and backend-draft integration status is documented accurately.

- [ ] **Step 4: Verify topic editing expectations without overclaiming**

Inspect and test only what exists:
```text
- whether authors can edit their own topic through an exposed UI
- whether edits visibly mark content as edited
- whether delete is user-accessible or API-only in the current build
```

Expected: the plan distinguishes current UX reality from upstream intentions around editing transparency and self-deletion.

## Task 5: Revalidate topic detail, replies, quoting, and mention behavior

**Files:**
- Test: `web/src/views/forum/ForumTopicView.vue`
- Test: `web/src/components/forum/ForumReplyNode.vue`
- Test: `web/src/stores/forum.ts`
- Test: `server/internal/handlers/forum_handler.go`
- Test: `server/internal/service/forum_mention_parser.go`

- [ ] **Step 1: Verify topic detail rendering**

Manually verify these checkpoints:
```text
- `/topic/:id` loads from the topic list and direct URL
- title, tags, author, timestamps, view count, and reply count display correctly
- markdown body renders successfully in the reading view
- closed or pinned state badges appear when relevant
```

Expected: the topic detail page is stable and information-complete.

- [ ] **Step 2: Verify reply creation and lightweight-reply expectations**

Manually verify these checkpoints:
```text
- authenticated users can submit replies
- closed topics block reply submission
- reply draft restore works
- the actual reply editor behavior is documented precisely: plain text only, markdown-capable, rich editor, image insertion present/absent
```

Expected: the implementation is described truthfully, especially where it may differ from the original lightweight-reply requirement.

- [ ] **Step 3: Verify quote-reply behavior**

Manually verify these checkpoints:
```text
- clicking quote on an existing reply focuses the reply form
- quoted reply metadata is displayed before submission
- submitted reply preserves the quoted-reply relationship
- quoted preview renders correctly in the reply list
```

Expected: the reply threading model based on quoted linkage is observable in the live UI.

- [ ] **Step 4: Verify mention-notification behavior if the app exposes notifications locally**

Manually verify these checkpoints:
```text
- include an `@username` mention in a reply or topic where possible
- confirm the content saves successfully
- inspect the notification surface for a resulting mention or reply notification
- if end-to-end confirmation is not feasible, verify at minimum that the parser-backed code path is reachable and document the missing proof point explicitly
```

Expected: mention handling is either verified end-to-end or honestly marked as partially verified.

## Task 6: Revalidate interaction and moderation-baseline behavior

**Files:**
- Test: `web/src/views/forum/ForumTopicView.vue`
- Test: `web/src/views/forum/ForumHomeView.vue`
- Test: `web/src/stores/forum.ts`
- Test: `server/internal/handlers/forum_handler.go`
- Test: `server/internal/model/forum.go`

- [ ] **Step 1: Verify like and bookmark flows**

Manually verify these checkpoints:
```text
- topic like toggles from the detail page
- reply like toggles from the reply list
- bookmark toggles from the topic detail page
- updated counters or button states reflect the mutation immediately
```

Expected: the baseline engagement flows work in both API and UI.

- [ ] **Step 2: Verify admin-only moderation flows that are currently implemented**

Manually verify these checkpoints with an admin user if available:
```text
- pin toggling works
- close toggling works
- closed topics visibly stop new replies
- non-admin users cannot access the same moderation actions if the UI exposes them
```

Expected: the implemented moderation baseline is proven instead of assumed.

- [ ] **Step 3: Audit planned-but-unverified governance features**

Create an explicit gap report using this exact structure:
```text
Verified implemented:
- pin / close

Not found in current verified code:
- report entry point
- moderation review queue
- auto-collapse threshold
- user block / content fold controls
- anti-spam rate limiting
- category request workflow
- best-answer controls
```

Expected: downstream work can be scoped from facts rather than inherited planning optimism.

## Task 7: Revalidate migration and persistence assumptions

**Files:**
- Test: `server/internal/model/forum.go`
- Test: `server/internal/service/forum_migrate.go`
- Test: `server/migrate_db.go`
- Test: `server/cmd/start_server/main.go`

- [ ] **Step 1: Verify migration startup path is still wired**

Inspect startup code and confirm whether forum migration execution is called during local boot. Record findings in this structure:
```text
- migration entrypoint file:
- forum migration invoked: yes/no
- draft tables covered: yes/no
- path / floor backfill covered: yes/no
```

Expected: no one has to guess whether forum-specific schema setup actually runs.

- [ ] **Step 2: Validate topic/reply persistence invariants with a fresh manual data pass**

Using the running app, create a topic and at least two replies, one quoted. Then confirm:
```text
- reply_count increments correctly
- last_reply_at updates correctly
- floor numbering is reflected in UI order
- quoted reply relation survives page reload
```

Expected: the essential persistence rules behave correctly under live usage.

- [ ] **Step 3: Note database-specific caveats explicitly**

Capture findings such as:
```text
- SQLite path behavior uses text fallback
- PostgreSQL path behavior uses ltree when available
- tag search semantics differ between LIKE fallback and PostgreSQL array-style querying
```

Expected: future debugging starts with the known SQLite/PostgreSQL differences already documented.

## Task 8: Produce a reusable synthesis rule for future feature plans

**Files:**
- Inspect: `plan/blog_task_plan.md`
- Inspect: `plan/debate_task_plan.md`
- Inspect: `plan/music_task_plan.md`
- Inspect: `plan/timeline_task_plan.md`
- Modify: `docs/superpowers/plans/2026-05-14-forum-implementation.md`

- [ ] **Step 1: Add a reuse rule section to the generated plan**

Use this exact section:
```md
## Reuse Rule For Other Features

To generate a superpowers implementation plan for another feature, reuse the same synthesis flow:
1. Read `plan/<feature>_task_plan.md` for scope and phase order.
2. Read `plan/<feature>_findings.md` for technical and product decisions when present.
3. Read `plan/<feature>_progress.md` for implementation evidence and prior validation when present.
4. Read any feature-specific supporting design files when they exist.
5. Reconcile upstream planning claims against the current codebase before assuming the feature is greenfield or complete.
6. Save the normalized result to `docs/superpowers/plans/YYYY-MM-DD-<feature>-implementation.md`.
```

Expected: the document doubles as a concrete forum plan and a reliable pattern for future modules.

- [ ] **Step 2: Note feature-shape differences this workflow must tolerate**

Capture examples like:
```text
- `forum`, `debate`, and `timeline` may have planning trios but very different implementation depth
- `music` may include extra design artifacts like `plan/music_pages.md`
- `video`, `podcast`, or `editor` may only have task-plan sources and need more code discovery before plan normalization
```

Expected: future plan generation does not assume every module has identical upstream artifacts or maturity.

## Reuse Rule For Other Features

To generate a superpowers implementation plan for another feature, reuse the same synthesis flow:
1. Read `plan/<feature>_task_plan.md` for scope and phase order.
2. Read `plan/<feature>_findings.md` for technical and product decisions when present.
3. Read `plan/<feature>_progress.md` for implementation evidence and prior validation when present.
4. Read any feature-specific supporting design files when they exist.
5. Reconcile upstream planning claims against the current codebase before assuming the feature is greenfield or complete.
6. Save the normalized result to `docs/superpowers/plans/YYYY-MM-DD-<feature>-implementation.md`.

## Verification

After execution, verify all of the following:

1. `docs/superpowers/plans/2026-05-14-forum-implementation.md` remains self-contained and executable without reopening upstream planning files.
2. `cd "/Users/fafa/Documents/projects/Atoman/.claude/worktrees/agent-a79bd39272f2a0a58/web" && bun run type-check` passes.
3. `cd "/Users/fafa/Documents/projects/Atoman/.claude/worktrees/agent-a79bd39272f2a0a58/web" && bun run build` passes.
4. `cd "/Users/fafa/Documents/projects/Atoman/.claude/worktrees/agent-a79bd39272f2a0a58/server" && go build ./...` passes.
5. Browser validation covers forum discovery, category filtering, search, topic creation, reply creation, quoting, likes, bookmarks, and closed-topic behavior.
6. The final execution notes clearly separate verified forum behavior from planned-but-unimplemented governance features.
7. Any best-answer, reporting, anti-spam, category-request, or moderation-queue gaps are documented explicitly rather than silently treated as complete.

## Commit

```bash
git add docs/superpowers/plans/2026-05-14-forum-implementation.md
git commit -m "docs: add forum superpowers implementation plan"
```
