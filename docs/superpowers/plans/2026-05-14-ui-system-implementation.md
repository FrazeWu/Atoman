# UI System Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Finish the remaining UI-system convergence work so the existing Archive / Brutalist design tokens and shared `A*` primitives are used consistently across the remaining high-traffic frontend surfaces.

**Architecture:** Treat `plan/ui_system_task_plan.md` as the scope baseline, `plan/ui_system_findings.md` as the design decision record, and `plan/ui_system_progress.md` as implementation evidence. Preserve the current foundation in `web/src/style.css` and `web/src/components/ui/*`, then finish the rollout through targeted migrations: remove stale plan assumptions, standardize `AModal` consumption, extract repeated inline layout styles into scoped classes, normalize legacy button/status patterns, and re-verify the result with both static scans and browser checks.

**Tech Stack:** Vue 3, TypeScript, Vite, Vue Router, Pinia, scoped CSS, global CSS variables in `web/src/style.css`

---

## Upstream Sources

- `plan/ui_system_task_plan.md`
- `plan/ui_system_findings.md`
- `plan/ui_system_progress.md`
- Style reference only: `docs/superpowers/plans/2026-05-14-blog-implementation.md`

## Normalized Current State

The upstream ui-system documents agree on the design language and rollout order: Archive / Brutalist styling, black/white high contrast, `0` radius, `2px` black borders, hard shadows only for floating objects, global design tokens first, shared `A*` primitives second, page migrations third. The implementation is **already far beyond initial planning**: `web/src/style.css` contains the first token set, the first-wave shared components exist in `web/src/components/ui/`, and several page batches were already migrated.

The remaining work is not “build the design system from scratch.” It is **finish the convergence and remove drift**. Current repo inspection shows four important normalization facts:

1. `ASelect`, `AModal`, `ABtn`, `AInput`, `ATextarea`, `ADropdown`, and `APopover` already exist.
2. `DatetimePicker.vue` already composes `ASelect` internally, so the old task-plan item about replacing internal native `<select>` elements is stale.
3. `AlbumHistoryView.vue`, `SongHistoryView.vue`, and `BlogManageView.vue` no longer contain raw `<select>` tags, so those old “pending select migration” items are also stale.
4. The real remaining drift is concentrated in:
   - hand-composed modal header/body/footer markup around `AModal`
   - widespread inline layout styles in feed/blog/timeline views
   - legacy button/status patterns such as `outline`, raw `a-btn-*` classes, and ad-hoc red danger styles

## Confirmed Scope To Preserve

- Keep the existing Archive / Brutalist direction unchanged: black/white, high contrast, hard structure, no rounded corners.
- Keep the current token location in `web/src/style.css`; do not fork token ownership into ad-hoc page-level color constants.
- Keep inputs flat and restrained; only floating objects such as buttons, dropdowns, and modals should retain hard-shadow elevation.
- Keep `AButton` semantics centered on `variant="primary|secondary|danger|ghost"` and `size="sm|md|lg"`.
- Keep `ASelect` aligned with `AInput` visually and behaviorally.
- Keep `AModal` as the canonical modal shell; page files should not rebuild modal chrome unless a shared API gap is first proven.
- Preserve intentional content-semantic colors such as code highlighting, warnings, moderation status, and timeline/debate content meaning. Only UI-structure colors must converge to tokens.

## Verified Current Repo Facts

### Shared foundation already present
- `web/src/style.css`
- `web/src/components/ui/ABtn.vue`
- `web/src/components/ui/AInput.vue`
- `web/src/components/ui/ATextarea.vue`
- `web/src/components/ui/ASelect.vue`
- `web/src/components/ui/AModal.vue`
- `web/src/components/ui/ADropdown.vue`
- `web/src/components/ui/APopover.vue`
- `web/src/components/ui/AConfirm.vue`
- `web/src/components/ui/DatetimePicker.vue`

### Current `AModal` API already supports convergence
`web/src/components/ui/AModal.vue` already provides:
- `modelValue`
- `size`
- `title`
- `closable`
- default slot for body
- `#footer` slot for actions

That means the next step should be migrating consumers back to the shared shell before changing the modal primitive itself.

### Remaining modal-shell drift found in current code
These files still hand-compose modal chrome or modal footer wrappers around `AModal`:
- `web/src/views/timeline/TimelineHomeView.vue`
- `web/src/views/timeline/PersonMapView.vue`
- `web/src/views/blog/BlogManageView.vue`
- `web/src/views/blog/ChannelManageView.vue`
- `web/src/views/blog/CollectionManageView.vue`
- `web/src/views/blog/BookmarkView.vue`
- `web/src/views/feed/FeedView.vue`

### Remaining inline-layout drift found in current code
Inline `style="..."` usage is still widespread in at least these target areas:
- `web/src/views/feed/FeedView.vue`
- `web/src/views/feed/FeedStatsView.vue`
- `web/src/views/feed/FeedStarredView.vue`
- `web/src/views/feed/FeedReadingListView.vue`
- `web/src/views/feed/FeedItemDetailView.vue`
- `web/src/views/blog/BlogHomeView.vue`
- `web/src/views/blog/BlogManageView.vue`
- `web/src/views/blog/BlogSettingsView.vue`
- `web/src/views/blog/BookmarkView.vue`
- `web/src/views/blog/ChannelManageDetailView.vue`
- `web/src/views/blog/ChannelManageView.vue`
- `web/src/views/blog/ChannelView.vue`
- `web/src/views/blog/CollectionManageView.vue`
- `web/src/views/blog/CollectionView.vue`
- `web/src/views/blog/ExploreView.vue`
- `web/src/views/blog/PostDetailView.vue`
- `web/src/views/blog/ProfileView.vue`
- `web/src/views/timeline/TimelineHomeView.vue`
- `web/src/views/timeline/PersonListView.vue`
- `web/src/views/timeline/PersonMapView.vue`

### Remaining legacy button/status drift found in current code
Representative targets from the current repo scan:
- `web/src/views/timeline/TimelineHomeView.vue`
- `web/src/views/timeline/PersonMapView.vue`
- `web/src/views/timeline/PersonListView.vue`
- `web/src/views/blog/BlogHomeView.vue`
- `web/src/views/blog/BookmarkView.vue`
- `web/src/views/blog/ChannelManageView.vue`
- `web/src/views/blog/ChannelManageDetailView.vue`
- `web/src/views/blog/CollectionView.vue`
- `web/src/views/feed/FeedView.vue`
- `web/src/views/feed/FeedStatsView.vue`
- `web/src/views/feed/FeedItemDetailView.vue`
- `web/src/views/feed/FeedStarredView.vue`
- `web/src/views/feed/FeedReadingListView.vue`
- `web/src/components/debate/ArgumentNode.vue`
- `web/src/components/forum/ForumReplyNode.vue`
- `web/src/components/blog/CommentSection.vue`

## Critical Implementation Files To Revalidate

### Foundation
- `web/src/style.css`
- `web/src/components/ui/ABtn.vue`
- `web/src/components/ui/ASelect.vue`
- `web/src/components/ui/AModal.vue`
- `web/src/components/ui/DatetimePicker.vue`

### Blog / creator surfaces
- `web/src/views/blog/BlogHomeView.vue`
- `web/src/views/blog/BlogManageView.vue`
- `web/src/views/blog/BookmarkView.vue`
- `web/src/views/blog/ChannelManageView.vue`
- `web/src/views/blog/ChannelManageDetailView.vue`
- `web/src/views/blog/ChannelView.vue`
- `web/src/views/blog/CollectionManageView.vue`
- `web/src/views/blog/CollectionView.vue`
- `web/src/views/blog/PostDetailView.vue`
- `web/src/views/blog/ProfileView.vue`

### Feed surfaces
- `web/src/views/feed/FeedView.vue`
- `web/src/views/feed/FeedStatsView.vue`
- `web/src/views/feed/FeedStarredView.vue`
- `web/src/views/feed/FeedReadingListView.vue`
- `web/src/views/feed/FeedItemDetailView.vue`

### Timeline surfaces
- `web/src/views/timeline/TimelineHomeView.vue`
- `web/src/views/timeline/PersonListView.vue`
- `web/src/views/timeline/PersonMapView.vue`

### Secondary semantic/button cleanup targets
- `web/src/components/debate/ArgumentNode.vue`
- `web/src/components/forum/ForumReplyNode.vue`
- `web/src/components/blog/CommentSection.vue`

## Task 1: Revalidate the current UI-system baseline and freeze the real scope

**Files:**
- Inspect: `plan/ui_system_task_plan.md`
- Inspect: `plan/ui_system_findings.md`
- Inspect: `plan/ui_system_progress.md`
- Inspect: `web/src/style.css`
- Inspect: `web/src/components/ui/ABtn.vue`
- Inspect: `web/src/components/ui/ASelect.vue`
- Inspect: `web/src/components/ui/AModal.vue`
- Inspect: `web/src/components/ui/DatetimePicker.vue`
- Test: `web/src/views/blog/**`
- Test: `web/src/views/feed/**`
- Test: `web/src/views/timeline/**`

- [ ] **Step 1: Reconcile stale upstream checklist items before editing**

Record these facts in working notes and treat them as the authoritative implementation baseline:
```text
- shared tokens already exist in web/src/style.css
- the first-wave A* components already exist in web/src/components/ui/
- DatetimePicker already uses ASelect internally
- AlbumHistoryView, SongHistoryView, and BlogManageView no longer contain raw <select> tags
- the remaining work is convergence, not greenfield component creation
```

Expected: the engineer starts from the real code state, not from stale checkboxes.

- [ ] **Step 2: Run the frontend type-check before any edits**

Run:
```bash
cd "/Users/fafa/Documents/projects/Atoman/.claude/worktrees/agent-af35196b94d880c20/web" && bun run type-check
```

Expected: the current baseline passes, or any pre-existing failures are captured before UI-system edits begin.

- [ ] **Step 3: Run the frontend production build before any edits**

Run:
```bash
cd "/Users/fafa/Documents/projects/Atoman/.claude/worktrees/agent-af35196b94d880c20/web" && bun run build
```

Expected: the current baseline builds successfully before migration work starts.

- [ ] **Step 4: Run static drift scans and save the hit list**

Run:
```bash
rg -n "<select|a-modal-header|a-modal-footer|style=\"" "/Users/fafa/Documents/projects/Atoman/.claude/worktrees/agent-af35196b94d880c20/web/src/views/blog" "/Users/fafa/Documents/projects/Atoman/.claude/worktrees/agent-af35196b94d880c20/web/src/views/feed" "/Users/fafa/Documents/projects/Atoman/.claude/worktrees/agent-af35196b94d880c20/web/src/views/timeline"
```

Run:
```bash
rg -n "<ABtn[^\n]*outline|class=\"a-btn-(outline|primary|secondary|danger)|\bdanger\b" "/Users/fafa/Documents/projects/Atoman/.claude/worktrees/agent-af35196b94d880c20/web/src/views/blog" "/Users/fafa/Documents/projects/Atoman/.claude/worktrees/agent-af35196b94d880c20/web/src/views/feed" "/Users/fafa/Documents/projects/Atoman/.claude/worktrees/agent-af35196b94d880c20/web/src/views/timeline" "/Users/fafa/Documents/projects/Atoman/.claude/worktrees/agent-af35196b94d880c20/web/src/components/blog" "/Users/fafa/Documents/projects/Atoman/.claude/worktrees/agent-af35196b94d880c20/web/src/components/debate" "/Users/fafa/Documents/projects/Atoman/.claude/worktrees/agent-af35196b94d880c20/web/src/components/forum"
```

Expected: a concrete migration hit list grouped into modal-shell, inline-layout, and legacy-semantic drift.

## Task 2: Standardize modal consumers around the shared `AModal` shell

**Files:**
- Modify: `web/src/views/blog/BlogManageView.vue`
- Modify: `web/src/views/blog/ChannelManageView.vue`
- Modify: `web/src/views/blog/CollectionManageView.vue`
- Modify: `web/src/views/blog/BookmarkView.vue`
- Modify: `web/src/views/feed/FeedView.vue`
- Modify: `web/src/views/timeline/TimelineHomeView.vue`
- Modify: `web/src/views/timeline/PersonMapView.vue`
- Revalidate: `web/src/components/ui/AModal.vue`

- [ ] **Step 1: Migrate simple modal consumers first**

Start with files that already use `AModal` but still manually render title/body spacing:
```text
- web/src/views/blog/BlogManageView.vue
- web/src/views/blog/ChannelManageView.vue
- web/src/views/blog/CollectionManageView.vue
- web/src/views/blog/BookmarkView.vue
- web/src/views/feed/FeedView.vue
```

For each file:
```text
- move the title text into AModal's title prop when possible
- keep body content in the default slot
- keep action buttons in the #footer slot
- remove duplicated heading margins and repeated footer wrapper spacing when AModal already provides the shell
- prefer AInput / ATextarea / ASelect labels over hand-styled label blocks where the component API already supports labels
```

Expected: simple consumers stop rebuilding modal chrome and become thinner wrappers over `AModal`.

- [ ] **Step 2: Migrate the timeline modal-heavy files second**

Apply the same convergence to:
```text
- web/src/views/timeline/TimelineHomeView.vue
- web/src/views/timeline/PersonMapView.vue
```

Pay special attention to:
```text
- .a-modal-header blocks inside the page file
- .a-modal-body wrappers duplicated inside the page file
- .a-modal-footer wrappers nested inside the #footer slot
- hand-written close buttons that duplicate AModal's closable behavior
```

Expected: timeline modals rely on the shared shell instead of page-local modal chrome.

- [ ] **Step 3: Only widen `AModal` if two or more consumers prove a real API gap**

Before editing `web/src/components/ui/AModal.vue`, verify whether a missing shared behavior is truly shared:
```text
- custom title alignment
- optional footer spacing rule
- closable control nuance
- size limitations shared by multiple consumers
```

Expected: `AModal` changes are driven by real repeated needs, not by one-off page structure.

- [ ] **Step 4: Re-run the modal drift scan**

Run:
```bash
rg -n "a-modal-header|a-modal-footer" "/Users/fafa/Documents/projects/Atoman/.claude/worktrees/agent-af35196b94d880c20/web/src/views/blog" "/Users/fafa/Documents/projects/Atoman/.claude/worktrees/agent-af35196b94d880c20/web/src/views/feed" "/Users/fafa/Documents/projects/Atoman/.claude/worktrees/agent-af35196b880c20/web/src/views/timeline"
```

Expected: zero results in migrated consumers, or a short explicit defer list with justification.

- [ ] **Step 5: Commit the modal-shell convergence**

Run:
```bash
git add web/src/views/blog/BlogManageView.vue web/src/views/blog/ChannelManageView.vue web/src/views/blog/CollectionManageView.vue web/src/views/blog/BookmarkView.vue web/src/views/feed/FeedView.vue web/src/views/timeline/TimelineHomeView.vue web/src/views/timeline/PersonMapView.vue web/src/components/ui/AModal.vue
```

Run:
```bash
git commit -m "refactor: standardize modal consumers on AModal"
```

Expected: modal-shell cleanup lands as one reviewable change.

## Task 3: Extract repeated inline layout styles into scoped classes

**Files:**
- Modify: `web/src/views/feed/FeedView.vue`
- Modify: `web/src/views/feed/FeedStatsView.vue`
- Modify: `web/src/views/feed/FeedStarredView.vue`
- Modify: `web/src/views/feed/FeedReadingListView.vue`
- Modify: `web/src/views/feed/FeedItemDetailView.vue`
- Modify: `web/src/views/blog/BlogHomeView.vue`
- Modify: `web/src/views/blog/BlogManageView.vue`
- Modify: `web/src/views/blog/BookmarkView.vue`
- Modify: `web/src/views/blog/ChannelManageDetailView.vue`
- Modify: `web/src/views/blog/ChannelManageView.vue`
- Modify: `web/src/views/blog/ChannelView.vue`
- Modify: `web/src/views/blog/CollectionManageView.vue`
- Modify: `web/src/views/blog/CollectionView.vue`
- Modify: `web/src/views/blog/PostDetailView.vue`
- Modify: `web/src/views/blog/ProfileView.vue`
- Modify: `web/src/views/timeline/TimelineHomeView.vue`
- Modify: `web/src/views/timeline/PersonListView.vue`
- Modify: `web/src/views/timeline/PersonMapView.vue`
- Revalidate: `web/src/style.css`

- [ ] **Step 1: Clean up feed pages first**

Process this batch first:
```text
- web/src/views/feed/FeedView.vue
- web/src/views/feed/FeedStatsView.vue
- web/src/views/feed/FeedStarredView.vue
- web/src/views/feed/FeedReadingListView.vue
- web/src/views/feed/FeedItemDetailView.vue
```

For each file:
```text
- replace repeated inline flex/gap/padding/margin/width styles with named scoped classes
- keep semantic color decisions unchanged unless they are clearly UI-structure colors
- prefer existing utility-like global classes only when they already exist; otherwise add scoped classes in the SFC
```

Expected: feed pages stop accumulating layout rules in templates.

- [ ] **Step 2: Clean up blog management and profile surfaces second**

Process this batch second:
```text
- web/src/views/blog/BlogHomeView.vue
- web/src/views/blog/BlogManageView.vue
- web/src/views/blog/BookmarkView.vue
- web/src/views/blog/ChannelManageDetailView.vue
- web/src/views/blog/ChannelManageView.vue
- web/src/views/blog/ChannelView.vue
- web/src/views/blog/CollectionManageView.vue
- web/src/views/blog/CollectionView.vue
- web/src/views/blog/PostDetailView.vue
- web/src/views/blog/ProfileView.vue
```

Expected: blog-facing pages move to readable scoped layout classes without changing route behavior or content semantics.

- [ ] **Step 3: Clean up timeline pages third**

Process this batch third:
```text
- web/src/views/timeline/TimelineHomeView.vue
- web/src/views/timeline/PersonListView.vue
- web/src/views/timeline/PersonMapView.vue
```

Expected: timeline layout becomes easier to maintain without disturbing OpenLayers behavior or timeline-specific semantic colors.

- [ ] **Step 4: Re-run the inline-style scan for migrated files only**

Run:
```bash
rg -n "style=\"" "/Users/fafa/Documents/projects/Atoman/.claude/worktrees/agent-af35196b94d880c20/web/src/views/feed" "/Users/fafa/Documents/projects/Atoman/.claude/worktrees/agent-af35196b94d880c20/web/src/views/blog/BlogHomeView.vue" "/Users/fafa/Documents/projects/Atoman/.claude/worktrees/agent-af35196b94d880c20/web/src/views/blog/BlogManageView.vue" "/Users/fafa/Documents/projects/Atoman/.claude/worktrees/agent-af35196b94d880c20/web/src/views/blog/BookmarkView.vue" "/Users/fafa/Documents/projects/Atoman/.claude/worktrees/agent-af35196b94d880c20/web/src/views/blog/ChannelManageDetailView.vue" "/Users/fafa/Documents/projects/Atoman/.claude/worktrees/agent-af35196b94d880c20/web/src/views/blog/ChannelManageView.vue" "/Users/fafa/Documents/projects/Atoman/.claude/worktrees/agent-af35196b94d880c20/web/src/views/blog/ChannelView.vue" "/Users/fafa/Documents/projects/Atoman/.claude/worktrees/agent-af35196b94d880c20/web/src/views/blog/CollectionManageView.vue" "/Users/fafa/Documents/projects/Atoman/.claude/worktrees/agent-af35196b94d880c20/web/src/views/blog/CollectionView.vue" "/Users/fafa/Documents/projects/Atoman/.claude/worktrees/agent-af35196b94d880c20/web/src/views/blog/PostDetailView.vue" "/Users/fafa/Documents/projects/Atoman/.claude/worktrees/agent-af35196b94d880c20/web/src/views/blog/ProfileView.vue" "/Users/fafa/Documents/projects/Atoman/.claude/worktrees/agent-af35196b94d880c20/web/src/views/timeline"
```

Expected: only intentional exceptions remain, and each exception is documented inline in the PR or progress notes.

- [ ] **Step 5: Commit the layout extraction**

Run:
```bash
git add web/src/views/feed web/src/views/blog/BlogHomeView.vue web/src/views/blog/BlogManageView.vue web/src/views/blog/BookmarkView.vue web/src/views/blog/ChannelManageDetailView.vue web/src/views/blog/ChannelManageView.vue web/src/views/blog/ChannelView.vue web/src/views/blog/CollectionManageView.vue web/src/views/blog/CollectionView.vue web/src/views/blog/PostDetailView.vue web/src/views/blog/ProfileView.vue web/src/views/timeline
```

Run:
```bash
git commit -m "refactor: extract shared page layouts into scoped UI classes"
```

Expected: template cleanup lands separately from semantic/button work.

## Task 4: Normalize legacy button semantics and remaining token drift

**Files:**
- Modify: `web/src/views/timeline/TimelineHomeView.vue`
- Modify: `web/src/views/timeline/PersonMapView.vue`
- Modify: `web/src/views/timeline/PersonListView.vue`
- Modify: `web/src/views/blog/BlogHomeView.vue`
- Modify: `web/src/views/blog/BookmarkView.vue`
- Modify: `web/src/views/blog/ChannelManageView.vue`
- Modify: `web/src/views/blog/ChannelManageDetailView.vue`
- Modify: `web/src/views/blog/CollectionView.vue`
- Modify: `web/src/views/feed/FeedView.vue`
- Modify: `web/src/views/feed/FeedStatsView.vue`
- Modify: `web/src/views/feed/FeedItemDetailView.vue`
- Modify: `web/src/views/feed/FeedStarredView.vue`
- Modify: `web/src/views/feed/FeedReadingListView.vue`
- Modify: `web/src/components/debate/ArgumentNode.vue`
- Modify: `web/src/components/forum/ForumReplyNode.vue`
- Modify: `web/src/components/blog/CommentSection.vue`
- Revalidate: `web/src/components/ui/ABtn.vue`
- Revalidate: `web/src/style.css`

- [ ] **Step 1: Replace legacy `outline` usage with explicit variants in migrated files**

Convert remaining `ABtn outline` usage in the current UI-system scope to:
```text
- variant="secondary" for the standard white/black bordered action
- variant="ghost" only when a lighter bordered action is truly intended
- variant="danger" for destructive actions
```

Expected: page and component code stop depending on compatibility-only props for new work.

- [ ] **Step 2: Replace raw `a-btn-*` class buttons when shared button semantics already fit**

Prioritize these files:
```text
- web/src/views/blog/CollectionView.vue
- web/src/views/blog/ChannelManageDetailView.vue
- web/src/views/feed/FeedStatsView.vue
- web/src/views/feed/FeedItemDetailView.vue
- web/src/views/feed/FeedStarredView.vue
- web/src/views/feed/FeedReadingListView.vue
```

Expected: common actions render through `ABtn` or documented intentional exceptions, rather than page-local legacy classes.

- [ ] **Step 3: Normalize ad-hoc danger styles into token-backed semantics**

Review these patterns specifically:
```text
- inline red border/text combinations in blog management views
- danger-only local classes such as .danger-btn, .reply-btn-danger, .card-action-danger, .tl-inline-link.danger
- hard-coded red styles inside feed action buttons
```

Expected: destructive UI states map to shared token semantics where possible, while content-semantic warning/danger states remain explicit and documented.

- [ ] **Step 4: Re-check whether `ABtn` itself needs a shared API change**

Only edit `web/src/components/ui/ABtn.vue` if the cleanup proves a real repeated need across multiple consumers, such as:
```text
- a missing size
- a missing loading-text pattern
- a justified slot/layout behavior used repeatedly
```

Expected: `ABtn` stays minimal unless the repo now clearly needs a shared expansion.

- [ ] **Step 5: Re-run the legacy-button scan**

Run:
```bash
rg -n "<ABtn[^\n]*outline|class=\"a-btn-(outline|primary|secondary|danger)" "/Users/fafa/Documents/projects/Atoman/.claude/worktrees/agent-af35196b94d880c20/web/src/views/blog" "/Users/fafa/Documents/projects/Atoman/.claude/worktrees/agent-af35196b94d880c20/web/src/views/feed" "/Users/fafa/Documents/projects/Atoman/.claude/worktrees/agent-af35196b94d880c20/web/src/views/timeline" "/Users/fafa/Documents/projects/Atoman/.claude/worktrees/agent-af35196b94d880c20/web/src/components/blog" "/Users/fafa/Documents/projects/Atoman/.claude/worktrees/agent-af35196b94d880c20/web/src/components/debate" "/Users/fafa/Documents/projects/Atoman/.claude/worktrees/agent-af35196b94d880c20/web/src/components/forum"
```

Expected: zero results in the migrated scope, or a small explicit exception list with rationale.

- [ ] **Step 6: Commit the semantic/button cleanup**

Run:
```bash
git add web/src/views/timeline/TimelineHomeView.vue web/src/views/timeline/PersonMapView.vue web/src/views/timeline/PersonListView.vue web/src/views/blog/BlogHomeView.vue web/src/views/blog/BookmarkView.vue web/src/views/blog/ChannelManageView.vue web/src/views/blog/ChannelManageDetailView.vue web/src/views/blog/CollectionView.vue web/src/views/feed/FeedView.vue web/src/views/feed/FeedStatsView.vue web/src/views/feed/FeedItemDetailView.vue web/src/views/feed/FeedStarredView.vue web/src/views/feed/FeedReadingListView.vue web/src/components/debate/ArgumentNode.vue web/src/components/forum/ForumReplyNode.vue web/src/components/blog/CommentSection.vue web/src/components/ui/ABtn.vue web/src/style.css
```

Run:
```bash
git commit -m "refactor: normalize remaining UI button and token semantics"
```

Expected: semantic convergence lands independently from layout and modal refactors.

## Task 5: Re-scope the old `DatetimePicker` follow-up and document the rule

**Files:**
- Inspect: `web/src/components/ui/DatetimePicker.vue`
- Test: `web/src/views/timeline/TimelineHomeView.vue`
- Test: `web/src/views/timeline/PersonMapView.vue`
- Test: any other current consumer of `DatetimePicker.vue`

- [ ] **Step 1: Explicitly close the stale native-select task**

Record this decision in working notes or implementation summary:
```text
The old task-plan item about replacing native <select> elements inside DatetimePicker is obsolete because DatetimePicker already uses ASelect for hour/minute selection.
```

Expected: future workers do not repeat already-completed select-migration work.

- [ ] **Step 2: Evaluate the real remaining consistency question inside `DatetimePicker`**

Review whether these internal controls should stay custom or be further unified:
```text
- dtp-year-input
- dtp-nav-btn
- dtp-clear
- dtp-confirm
```

Decision rule:
```text
- keep internal controls custom if they are part of the widget's compound behavior and already match the token system
- only switch them to shared primitives if doing so simplifies the widget without harming keyboard/mouse interaction
```

Expected: `DatetimePicker` gets a clear boundary instead of a vague “maybe unify later” note.

- [ ] **Step 3: Verify the picker in real modal forms**

Manually verify these routes and flows after any `DatetimePicker` edits:
```text
- /timeline -> create/edit event modal
- /timeline/persons/:id -> add/edit location modal
- /timeline/persons/:id -> edit person modal
```

Expected: date/time selection, open/close behavior, confirm/clear actions, and modal interaction still work correctly.

## Task 6: Final verification sweep

**Files:**
- Test: `web/src/style.css`
- Test: `web/src/components/ui/ABtn.vue`
- Test: `web/src/components/ui/ASelect.vue`
- Test: `web/src/components/ui/AModal.vue`
- Test: `web/src/components/ui/DatetimePicker.vue`
- Test: all files changed in Tasks 2–5

- [ ] **Step 1: Run final static verification commands**

Run:
```bash
cd "/Users/fafa/Documents/projects/Atoman/.claude/worktrees/agent-af35196b94d880c20/web" && bun run type-check
```

Run:
```bash
cd "/Users/fafa/Documents/projects/Atoman/.claude/worktrees/agent-af35196b94d880c20/web" && bun run build
```

Expected: final UI-system changes compile cleanly.

- [ ] **Step 2: Start the local app for manual verification**

Run in one terminal:
```bash
cd "/Users/fafa/Documents/projects/Atoman/.claude/worktrees/agent-af35196b94d880c20/server" && go run cmd/start_server/main.go
```

Run in a second terminal:
```bash
cd "/Users/fafa/Documents/projects/Atoman/.claude/worktrees/agent-af35196b94d880c20/web" && bun run dev
```

Expected: backend API and frontend dev server both start successfully.

- [ ] **Step 3: Verify blog management flows in the browser**

Using an authenticated session, verify:
```text
- /blog/manage opens both creation modals and their actions render through the shared shell
- /channels supports create/edit/delete channel flows without duplicated modal chrome
- /collections supports create/edit/delete collection flows without duplicated modal chrome
- /blog/bookmarks supports creating a folder with the updated modal/button semantics
- /channel/:slug/manage still loads and destructive actions still read clearly as danger actions
```

Expected: creator/blog surfaces behave the same but are visually implemented through the shared system.

- [ ] **Step 4: Verify feed flows in the browser**

Using an authenticated session, verify:
```text
- /feed opens the add-subscription modal and group selection still works
- /feed/stats renders correctly after layout extraction
- /feed/starred and /feed/reading-list still navigate and render their primary/back actions correctly
- /feed/item/:id still renders the detail page with the expected navigation/action buttons
```

Expected: feed flows keep behavior while dropping inline layout and legacy button drift.

- [ ] **Step 5: Verify timeline flows in the browser**

Verify:
```text
- /timeline opens event detail, create event, and create person modals with the shared shell
- /timeline/persons opens search/list actions with the updated button semantics
- /timeline/persons/:id opens add/edit location and edit person modals correctly
- DatetimePicker open/close, confirm, and clear behavior still works inside timeline modals
```

Expected: timeline remains functional after the heaviest modal/layout cleanup.

- [ ] **Step 6: Re-run the post-migration drift scans**

Run:
```bash
rg -n "<select|a-modal-header|a-modal-footer|style=\"" "/Users/fafa/Documents/projects/Atoman/.claude/worktrees/agent-af35196b94d880c20/web/src/views/blog" "/Users/fafa/Documents/projects/Atoman/.claude/worktrees/agent-af35196b94d880c20/web/src/views/feed" "/Users/fafa/Documents/projects/Atoman/.claude/worktrees/agent-af35196b94d880c20/web/src/views/timeline"
```

Run:
```bash
rg -n "<ABtn[^\n]*outline|class=\"a-btn-(outline|primary|secondary|danger)" "/Users/fafa/Documents/projects/Atoman/.claude/worktrees/agent-af35196b94d880c20/web/src/views/blog" "/Users/fafa/Documents/projects/Atoman/.claude/worktrees/agent-af35196b94d880c20/web/src/views/feed" "/Users/fafa/Documents/projects/Atoman/.claude/worktrees/agent-af35196b94d880c20/web/src/views/timeline" "/Users/fafa/Documents/projects/Atoman/.claude/worktrees/agent-af35196b94d880c20/web/src/components/blog" "/Users/fafa/Documents/projects/Atoman/.claude/worktrees/agent-af35196b94d880c20/web/src/components/debate" "/Users/fafa/Documents/projects/Atoman/.claude/worktrees/agent-af35196b94d880c20/web/src/components/forum"
```

Expected: no unexpected structural drift remains in the declared migration scope.

## Verification Summary

The implementation is only complete when all of the following are true:

1. `web/src/style.css` remains the single source of truth for shared UI-structure tokens.
2. `AModal` consumers no longer rebuild `.a-modal-header` / `.a-modal-footer` shells in the migrated scope.
3. Target feed/blog/timeline views no longer rely on large inline layout blocks for ordinary page structure.
4. Legacy `outline` and raw `a-btn-*` usage is eliminated from the migrated scope, or explicitly documented as an intentional compatibility exception.
5. `DatetimePicker` has a documented boundary: internal compound controls are either intentionally custom or intentionally migrated, but not left ambiguous.
6. `cd "/Users/fafa/Documents/projects/Atoman/.claude/worktrees/agent-af35196b94d880c20/web" && bun run type-check` passes.
7. `cd "/Users/fafa/Documents/projects/Atoman/.claude/worktrees/agent-af35196b94d880c20/web" && bun run build` passes.
8. Browser validation covers `/blog/manage`, `/channels`, `/collections`, `/blog/bookmarks`, `/feed`, `/feed/stats`, `/feed/item/:id`, `/timeline`, and `/timeline/persons/:id`.

## Recommended Commit Sequence

```bash
git commit -m "refactor: standardize modal consumers on AModal"
git commit -m "refactor: extract shared page layouts into scoped UI classes"
git commit -m "refactor: normalize remaining UI button and token semantics"
```

## Execution Handoff

This plan is self-contained and should be executable without reopening the upstream `plan/ui_system_*.md` files.
