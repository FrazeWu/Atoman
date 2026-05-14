# Blog Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Consolidate the Blog/Studio upstream planning artifacts into one execution-ready implementation plan that reflects confirmed design decisions, current implementation status, and remaining verification work.

**Architecture:** Treat `plan/blog_task_plan.md` as the scope baseline and phase ledger, `plan/blog_findings.md` as the technical decision record, and `plan/blog_progress.md` as implementation evidence plus prior verification history. The generated document should preserve actionable work and verified context while filtering out stale planning chatter and outdated in-progress notes.

**Tech Stack:** Vue 3, TypeScript, Vite, Pinia, Vue Router, Go, Gin, GORM, Tiptap, Yjs

---

## Upstream Sources

- `plan/blog_task_plan.md`
- `plan/blog_findings.md`
- `plan/blog_progress.md`
- Related dependency plans referenced upstream:
  - `plan/video_task_plan.md`
  - `plan/podcast_task_plan.md`

## Normalized Current State

The upstream documents agree on the intended Blog-to-Studio redesign and the core architecture: channel-based authorship, collection-based curation, collaborative publishing, Tiptap-based editing, and future-ready real-time collaboration. `plan/blog_task_plan.md` claims the full redesign is complete, while `plan/blog_progress.md` preserves an older snapshot where implementation was still mid-flight and browser verification was still pending. For execution purposes, treat the feature as **largely implemented but requiring fresh verification and reconciliation against the current codebase** before making any new behavioral changes.

## Confirmed Scope To Preserve

- Posts belong to `channel_id`, not `user_id`.
- `Channel.slug` drives the primary channel routes.
- `/blog` is a discovery surface, not a creator dashboard.
- Channel pages, channel manage pages, and user profile channel lists are part of the baseline redesign.
- Collaboration uses invite-based membership, version history, and eventual Yjs-backed real-time editing.
- The long-term editor direction is Tiptap with Markdown round-trip support and custom embedded post/music/video nodes.

## Verified Architecture Decisions From Upstream

- **Identity model:** user account is the consumer identity; channel is the creator identity.
- **Collection model:** channel-owned collections may reference content across channels and content types.
- **Editor strategy:** Tiptap is the long-term editor foundation; Markdown remains the persisted format.
- **Embed model:** `postEmbed`, `musicEmbed`, and `videoEmbed` are block-level atomic nodes whose persisted attrs only keep `id`.
- **Serialization model:** custom Markdown directive parsing and custom serializer are required for round-trip stability.
- **Collaboration model:** invite flow, version history, and Yjs/WebSocket collaboration are part of the design contract.

## Critical Implementation Files To Revalidate

### Frontend
- `web/src/views/blog/PostEditorView.vue`
- `web/src/components/blog/MarkdownEditor.vue`
- `web/src/components/blog/TiptapMarkdownEditor.vue`
- `web/src/components/blog/editor/markdown/parseMarkdownToHtml.ts`
- `web/src/components/blog/editor/markdown/serializeTiptapToMarkdown.ts`
- `web/src/components/blog/editor/tiptap/nodes/PostEmbed.ts`
- `web/src/views/blog/PostDetailView.vue`
- `web/src/views/blog/ChannelView.vue`
- `web/src/views/blog/BlogHomeView.vue`
- `web/src/router.ts`
- `web/src/types.ts`
- `web/src/composables/useApi.ts`

### Backend
- `server/cmd/start_server/main.go`
- `server/internal/model/user.go`
- `server/internal/service/forum_migrate.go`
- `server/internal/handlers/blog_interaction_handler.go`

## Task 1: Revalidate the current Blog/Studio baseline

**Files:**
- Inspect: `plan/blog_task_plan.md`
- Inspect: `plan/blog_findings.md`
- Inspect: `plan/blog_progress.md`
- Test: `web/src/views/blog/**`
- Test: `web/src/components/blog/**`
- Test: `server/cmd/start_server/main.go`

- [ ] **Step 1: Review the upstream completion claims**

Read the upstream files and record three things in working notes:
```text
- what the redesign says is complete
- which features are explicitly still follow-up work
- where progress notes appear stale or contradictory
```

Expected: a short reconciliation list you can compare against the current codebase.

- [ ] **Step 2: Run the frontend type-check**

Run:
```bash
cd web && bun run type-check
```

Expected: successful TypeScript verification with no blog-specific type regressions.

- [ ] **Step 3: Run the frontend production build**

Run:
```bash
cd web && bun run build
```

Expected: successful build confirming the Blog/Studio surfaces still compile end-to-end.

- [ ] **Step 4: Run the backend build**

Run:
```bash
cd server && go build ./...
```

Expected: successful Go build confirming model and route changes referenced by the upstream plan still compile.

- [ ] **Step 5: Compare actual code state against the upstream plan**

Inspect the listed frontend and backend files and record whether these features are present:
```text
- channel-based routing
- channel manage page
- discovery-style /blog page
- Tiptap editor mount
- postEmbed parsing and rendering path
```

Expected: confirmation that the upstream plan is still aligned with the current repository before any new edits are attempted.

## Task 2: Revalidate editor behavior and embedded content flow

**Files:**
- Test: `web/src/views/blog/PostEditorView.vue`
- Test: `web/src/components/blog/TiptapMarkdownEditor.vue`
- Test: `web/src/components/blog/editor/markdown/parseMarkdownToHtml.ts`
- Test: `web/src/components/blog/editor/markdown/serializeTiptapToMarkdown.ts`
- Test: `web/src/components/blog/editor/tiptap/nodes/PostEmbed.ts`
- Test: `web/src/views/blog/PostDetailView.vue`

- [ ] **Step 1: Start the frontend dev server**

Run:
```bash
cd web && bun run dev
```

Expected: the Vite development server starts successfully for browser validation.

- [ ] **Step 2: Open the Blog/Studio editor flow in the browser**

Manually verify these checkpoints:
```text
- create or edit a blog post
- confirm the editor surface is the Tiptap-based flow, not the old Vditor-only surface
- confirm mode switching behavior still exists if upstream says SV/WYSIWYG was delivered
```

Expected: the editor experience matches the documented redesign.

- [ ] **Step 3: Verify `postEmbed` insertion in the editor**

Manually verify these checkpoints:
```text
- trigger the toolbar flow that inserts a post reference
- confirm the expected directive or embedded node appears in editor state
- save the post successfully
```

Expected: embedded post references can still be created using the current editor.

- [ ] **Step 4: Verify the reading path for embedded posts**

Manually verify these checkpoints:
```text
- open the saved post in the reading view
- confirm the post embed renders as a card rather than raw directive text
- confirm title/summary/channel metadata load successfully
```

Expected: the read path matches the upstream documented `postEmbed` behavior.

- [ ] **Step 5: Record any remaining browser-only gaps**

Capture findings such as:
```text
- save flow regression
- mismatch between raw Markdown and rendered node
- missing metadata fetch
- hydration/layout problems in the editor shell
```

Expected: any unresolved issues are explicit follow-up items, not hidden assumptions.

## Task 3: Revalidate channel-based navigation and management surfaces

**Files:**
- Test: `web/src/views/blog/BlogHomeView.vue`
- Test: `web/src/views/blog/ChannelView.vue`
- Test: `web/src/router.ts`
- Test: `web/src/composables/useApi.ts`
- Test: related channel management views and route targets

- [ ] **Step 1: Verify `/blog` behaves as a discovery page**

Manually verify these checkpoints:
```text
- `/blog` renders as a content discovery view
- there is no old creator-dashboard assumption in the page shell
- filter/sort behavior matches the documented redesign if present
```

Expected: `/blog` matches the upstream “discovery, not dashboard” decision.

- [ ] **Step 2: Verify channel routes and channel identity pages**

Manually verify these checkpoints:
```text
- `/channel/:slug` loads correctly
- channel header, collections, and mixed content areas render
- slug-based routing resolves through current API helpers
```

Expected: the creator-identity model is still visible in the live UI.

- [ ] **Step 3: Verify channel management access and routing**

Manually verify these checkpoints:
```text
- `/channel/:slug/manage` is routable
- management sections render without runtime errors
- content-management and settings surfaces still mount correctly
```

Expected: the management route documented upstream still works in the current codebase.

- [ ] **Step 4: Verify user profile channel-list behavior**

Manually verify these checkpoints:
```text
- user-facing profile surfaces show channel-oriented output
- the redesign did not regress back to the old user-owned-post mental model
```

Expected: the consumer-vs-creator identity split is preserved in the UI.

## Task 4: Revalidate collaboration and version-history surfaces

**Files:**
- Test: `web/src/views/blog/PostEditorView.vue`
- Test: collaboration-related blog editor components
- Test: `server/cmd/start_server/main.go`
- Test: any collaboration or WebSocket integration files currently used by the editor flow

- [ ] **Step 1: Verify collaboration UI surfaces still render**

Manually verify these checkpoints:
```text
- collaboration section is visible where upstream expects it
- version-history section is visible where upstream expects it
- no obvious runtime errors occur when opening these panels
```

Expected: documented collaboration and history surfaces still exist.

- [ ] **Step 2: Verify non-realtime collaboration workflow if present**

Manually verify these checkpoints:
```text
- invite-link related UI can be reached
- acceptance or collaborator state UI still renders if available in current code
```

Expected: the invite-based collaboration baseline has not regressed.

- [ ] **Step 3: Verify realtime-collaboration assumptions without overclaiming**

If a full second browser/client setup is not available, verify only what is actually observable:
```text
- the realtime collaboration code path mounts without crashing
- the editor opens when collaboration-related code is enabled
- any WebSocket bootstrap errors are captured explicitly
```

Expected: the plan reports verified facts, not optimistic assumptions about Yjs behavior.

## Task 5: Produce a reusable synthesis rule for future features

**Files:**
- Inspect: `plan/forum_task_plan.md`
- Inspect: `plan/music_task_plan.md`
- Inspect: `plan/music_pages.md`
- Inspect: `plan/ui_system_task_plan.md`
- Modify: `docs/superpowers/plans/2026-05-14-blog-implementation.md`

- [ ] **Step 1: Add a reuse rule section to the generated plan**

Use this exact section:
```md
## Reuse Rule For Other Features

To generate a superpowers implementation plan for another feature, reuse the same synthesis flow:
1. Read `plan/<feature>_task_plan.md` for scope and phase order.
2. Read `plan/<feature>_findings.md` for technical decisions when present.
3. Read `plan/<feature>_progress.md` for implementation evidence and prior validation when present.
4. Read any feature-specific supporting design file such as `plan/music_pages.md` when it exists.
5. Save the normalized result to `docs/superpowers/plans/YYYY-MM-DD-<feature>-implementation.md`.
```

Expected: the plan can serve as both a concrete blog execution plan and a pattern for future generated plans.

- [ ] **Step 2: Note upstream feature-shape differences the workflow must tolerate**

Capture examples like:
```text
- `music` has extra design material in `plan/music_pages.md`
- `forum`, `debate`, and `timeline` have full planning trios but may be at different implementation depths
- `editor`, `podcast`, and `video` may only have task-plan sources
```

Expected: future plan generation does not assume every feature has identical upstream artifacts.

## Reuse Rule For Other Features

To generate a superpowers implementation plan for another feature, reuse the same synthesis flow:
1. Read `plan/<feature>_task_plan.md` for scope and phase order.
2. Read `plan/<feature>_findings.md` for technical decisions when present.
3. Read `plan/<feature>_progress.md` for implementation evidence and prior validation when present.
4. Read any feature-specific supporting design file such as `plan/music_pages.md` when it exists.
5. Save the normalized result to `docs/superpowers/plans/YYYY-MM-DD-<feature>-implementation.md`.

## Verification

After execution, verify all of the following:

1. `docs/superpowers/plans/2026-05-14-blog-implementation.md` remains self-contained and executable without reopening upstream source files.
2. `cd web && bun run type-check` passes.
3. `cd web && bun run build` passes.
4. `cd server && go build ./...` passes.
5. Browser validation covers editor flow, embedded post rendering, discovery page behavior, channel routing, and management surfaces.
6. Any unverified collaboration or realtime behavior is explicitly documented rather than assumed complete.

## Commit

```bash
git add docs/superpowers/plans/2026-05-14-blog-implementation.md
git commit -m "docs: add blog superpowers implementation plan"
```
