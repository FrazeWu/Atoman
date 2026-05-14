# Editor Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Normalize and finish the shared editor v2 migration so every supported surface uses the intended SV/plain modes correctly, blog collaboration stays stable on CodeMirror + Yjs, and verification covers embeds, uploads, mentions, and cross-feature reuse.

**Architecture:** Treat `plan/editor/plan.md` as the only upstream intent document, then reconcile it against the current repository before making any behavior changes. The current codebase already contains a live CodeMirror-based `AEditor`, blog collaboration wiring, embed rendering, and upload support, so execution should be a gap-fixing and regression-hardening pass rather than a greenfield rewrite.

**Tech Stack:** Vue 3, TypeScript, Vite, CodeMirror 6, marked, Yjs, y-codemirror.next, y-websocket, Playwright, Go, Gin

---

## Upstream Sources

- Primary upstream input: `plan/editor/plan.md`
- Style reference only: `docs/superpowers/plans/2026-05-14-blog-implementation.md`

## Missing Upstream Inputs

The editor feature does **not** currently have the full planning trio used by some other modules.

- Missing: `plan/editor/findings.md`
- Missing: `plan/editor/progress.md`
- Present: `plan/editor/plan.md`

Because those findings/progress inputs are absent, do **not** assume the task checklist in `plan/editor/plan.md` reflects the current implementation state. Use the repository itself as the source of truth for reconciliation.

## Normalized Current State

The repository already contains a large portion of the editor v2 target state:

- `web/src/components/shared/AEditor.vue` already implements `mode="sv"` with CodeMirror 6 and `mode="plain"` with a textarea.
- Blog collaboration is already wired through `enableCollab`, `collabRoomId`, `y-codemirror.next`, and `y-websocket`.
- The Go backend already exposes a y-websocket-compatible relay at `/api/collab/ws/:roomID` in `server/internal/collab/hub.go`.
- Blog image upload is already wired through `/api/blog/upload-image` in `server/internal/handlers/blog_upload_handler.go`.
- Blog embeds are already persisted and rendered with the directive form `:::post{id="uuid"}:::`, `:::music{id="uuid"}:::`, and `:::video{id="uuid"}:::`.
- `web/src/views/blog/PostEditorView.vue` already mounts `AEditor` in `sv` mode with `enableEmbeds` and `enableCollab` for edit sessions.
- The old Tiptap directory expected by the upstream checklist is already gone from `web/src/components/blog/`.

Execution should therefore start by reconciling stale upstream assumptions before changing any code.

## Verified Divergences From The Upstream Checklist

These divergences must be handled intentionally instead of overwritten accidentally:

1. **Embed syntax divergence**
   - Upstream plan target: `{{embed:post:uuid}}`
   - Current repository contract: `:::post{id="uuid"}:::` / `:::music{...}` / `:::video{...}`
   - Recommended plan behavior: keep the existing `:::` directive contract unless product explicitly reopens the persisted Markdown format decision.

2. **Collaboration dependency divergence**
   - Upstream plan suggests removing `y-websocket` if possible.
   - Current repository imports `WebsocketProvider` from `y-websocket` in `AEditor.vue`, and the backend relay in `server/internal/collab/hub.go` is explicitly y-websocket-compatible.
   - Recommended plan behavior: keep `y-websocket` unless you also replace both the client provider and the server relay contract in one coordinated change.

3. **Cleanup-state divergence**
   - Upstream plan still contains cleanup tasks for Tiptap-related frontend files.
   - Current repository no longer contains the old `blog/editor/tiptap/` directory or the old serializer/parser files named in earlier blog planning.
   - Recommended plan behavior: verify absence, then remove only truly stale references that still remain in package manifests or build config.

## Confirmed Scope To Preserve

- Shared editor contract remains `mode: 'sv' | 'plain'`.
- Blog editor uses `sv` mode with embeds and collaboration.
- Forum and music discussion surfaces use `sv` mode with mentions and without collaboration.
- Debate argument editing stays on `plain` mode.
- Image upload must work for toolbar upload, paste, and drag/drop.
- Upload placeholder replacement must remain deterministic for concurrent uploads.
- Scroll synchronization must remain bidirectional without infinite event loops.
- Collaboration must remain room-based and tied to blog post IDs.

## Critical Implementation Files

### Frontend

- `web/src/components/shared/AEditor.vue`
- `web/src/composables/useMarkdownRenderer.ts`
- `web/src/views/blog/PostEditorView.vue`
- `web/src/views/blog/PostDetailView.vue`
- `web/src/views/forum/ForumNewTopicView.vue`
- `web/src/views/forum/ForumTopicView.vue`
- `web/src/views/music/AlbumDiscussionView.vue`
- `web/src/views/music/SongDiscussionView.vue`
- `web/src/views/debate/DebateTopicView.vue`
- `web/package.json`
- `web/vite.config.ts`
- `web/playwright.config.ts`
- `web/tests/e2e/global-setup.ts`
- Create: `web/tests/e2e/editor-regression.spec.ts`

### Backend

- `server/internal/handlers/blog_upload_handler.go`
- `server/internal/collab/hub.go`
- `server/cmd/start_server/main.go`

## File Responsibilities

- `web/src/components/shared/AEditor.vue`
  - Owns the reusable editor runtime: SV/plain mode split, toolbar, mentions, image upload entry points, scroll sync, and collaboration mounting.
- `web/src/composables/useMarkdownRenderer.ts`
  - Owns preview-time Markdown rendering and directive-to-card replacement.
- `web/src/views/blog/PostEditorView.vue`
  - Owns the blog authoring shell, autosave, blog-specific editor props, and collab room selection.
- `web/src/views/blog/PostDetailView.vue`
  - Owns read-path embed metadata fetch and final rendered post content.
- `web/src/views/forum/*.vue`, `web/src/views/music/*.vue`, `web/src/views/debate/DebateTopicView.vue`
  - Own the non-blog caller contracts that must continue to work after any shared editor cleanup.
- `server/internal/handlers/blog_upload_handler.go`
  - Owns authenticated blog image upload behavior and returned URLs.
- `server/internal/collab/hub.go`
  - Owns the WebSocket room relay semantics used by the editor collaboration client.
- `web/tests/e2e/editor-regression.spec.ts`
  - Will own end-to-end regression coverage for shared editor behavior after the implementation pass.

## Task 1: Reconcile the upstream plan with the live editor contract

**Files:**
- Inspect: `plan/editor/plan.md`
- Inspect: `web/src/components/shared/AEditor.vue`
- Inspect: `web/src/composables/useMarkdownRenderer.ts`
- Inspect: `web/package.json`
- Inspect: `web/vite.config.ts`
- Inspect: `server/internal/collab/hub.go`
- Inspect: `server/internal/handlers/blog_upload_handler.go`

- [ ] **Step 1: Record the current contract before editing anything**

Write down this exact reconciliation checklist in working notes:

```text
- AEditor already supports mode="sv" and mode="plain"
- blog collaboration already uses y-codemirror.next + y-websocket
- blog embed syntax is :::kind{id="uuid"}:::
- blog upload endpoint is /api/blog/upload-image
- server relay path is /api/collab/ws/:roomID
- old Tiptap directory is already absent
```

Expected: the engineer begins from verified facts instead of redoing already-landed work.

- [ ] **Step 2: Decide whether persisted embed syntax is changing or staying**

Use this decision rule exactly:

```text
If product/design has NOT explicitly reopened embed Markdown syntax, keep the existing :::post{id="uuid"}::: contract and do not migrate stored content.
If product/design HAS explicitly reopened embed Markdown syntax, plan a full writer + preview + read-path + existing-content migration in one batch.
```

Expected: no one changes only the insertion path or only the render path.

- [ ] **Step 3: Confirm the collaboration dependency contract**

Use this decision rule exactly:

```text
Keep y-websocket in web/package.json while AEditor imports WebsocketProvider from y-websocket and server/internal/collab/hub.go stays y-websocket-compatible.
Only remove y-websocket if both sides are replaced together and equivalent room relay behavior is verified.
```

Expected: dependency cleanup does not silently break collaboration.

- [ ] **Step 4: Run frontend type-check on the current baseline**

Run:

```bash
cd web && bun run type-check
```

Expected: successful TypeScript verification before any editor changes begin.

- [ ] **Step 5: Run the frontend build on the current baseline**

Run:

```bash
cd web && bun run build
```

Expected: successful production build confirming the current editor stack compiles.

- [ ] **Step 6: Run the backend build on the current baseline**

Run:

```bash
cd server && go build ./...
```

Expected: successful Go build confirming upload and collaboration routes still compile.

## Task 2: Harden `AEditor` as the single shared editor runtime

**Files:**
- Modify: `web/src/components/shared/AEditor.vue`
- Inspect: `web/src/views/blog/PostEditorView.vue`
- Inspect: `web/src/views/forum/ForumNewTopicView.vue`
- Inspect: `web/src/views/forum/ForumTopicView.vue`
- Inspect: `web/src/views/music/AlbumDiscussionView.vue`
- Inspect: `web/src/views/music/SongDiscussionView.vue`
- Inspect: `web/src/views/debate/DebateTopicView.vue`

- [ ] **Step 1: Freeze the public prop contract before refactoring internals**

The target caller contract is:

```ts
interface AEditorProps {
  modelValue: string
  mode: 'sv' | 'plain'
  placeholder?: string
  noBorder?: boolean
  enableImageUpload?: boolean
  enableMentions?: boolean
  enableEmbeds?: boolean
  enableCollab?: boolean
  collabRoomId?: string
}
```

Expected: internal cleanup does not leak new required props into forum/music/debate callers.

- [ ] **Step 2: Recheck mode isolation behavior**

Verify these exact invariants in `AEditor.vue` after any edits:

```text
- plain mode renders only the textarea path
- sv mode owns CodeMirror, toolbar, preview, upload, mentions, and collab logic
- toggling plain-mode callers does not initialize CodeMirror or Yjs
```

Expected: debate plain-mode editing stays lightweight.

- [ ] **Step 3: Recheck scroll synchronization guardrails**

Keep or restore this behavior:

```text
- left scroll updates right preview via ratio mapping
- right preview scroll updates left editor via ratio mapping
- a syncing flag prevents recursive scroll loops
```

Expected: bidirectional sync remains stable on long documents.

- [ ] **Step 4: Recheck image upload behavior in one place**

`AEditor.vue` must continue to support all three entry points:

```text
- toolbar button → file input → upload
- paste image from clipboard → upload
- drag/drop image file → upload
```

Expected: every image entry point uses the same placeholder and replacement logic.

- [ ] **Step 5: Recheck concurrent upload placeholder replacement**

Preserve the current unique-placeholder rule:

```text
Insert placeholder: ![上传中-<unique-id>]()
On success replace the exact matching placeholder with ![图片](<url>)
On failure remove only the matching placeholder
```

Expected: two simultaneous uploads do not overwrite each other.

- [ ] **Step 6: Recheck mention behavior for SV-only callers**

Verify these exact conditions:

```text
- mention lookup runs only when enableMentions is true
- dropdown keyboard navigation handles ArrowUp, ArrowDown, Enter/Tab, Escape
- applying a mention inserts a Markdown link, not raw display text
```

Expected: forum and music mention flows keep working after shared editor cleanup.

- [ ] **Step 7: Recheck collaboration mount conditions**

Verify this exact rule in `AEditor.vue`:

```text
Only mount Yjs collaboration when enableCollab === true and collabRoomId is a non-empty string.
Otherwise create a normal EditorState seeded from modelValue.
```

Expected: non-blog callers never accidentally attempt to open collaboration sockets.

## Task 3: Reconcile blog-specific write path, preview path, and read path as one contract

**Files:**
- Modify: `web/src/views/blog/PostEditorView.vue`
- Modify: `web/src/components/shared/AEditor.vue`
- Modify: `web/src/composables/useMarkdownRenderer.ts`
- Modify: `web/src/views/blog/PostDetailView.vue`
- Modify: `server/internal/handlers/blog_upload_handler.go`
- Inspect: `server/cmd/start_server/main.go`

- [ ] **Step 1: Keep the blog caller minimal and explicit**

`PostEditorView.vue` should continue to mount the shared editor with the blog-only features turned on and nothing else:

```vue
<AEditor
  v-model="form.content"
  mode="sv"
  :no-border="true"
  :enable-embeds="true"
  :enable-collab="isEdit"
  :collab-room-id="isEdit ? String(route.params.id || '') : undefined"
/>
```

Expected: blog-specific behavior stays in the caller contract, not hidden in the shared editor.

- [ ] **Step 2: Preserve one embed syntax across insertion, preview, and read rendering**

Preferred contract for this implementation plan:

```text
writer insertion in AEditor.vue: :::post{id="uuid"}::: / :::music{...}::: / :::video{...}:::
preview replacement in useMarkdownRenderer.ts: same syntax
read-path metadata fetch in PostDetailView.vue: same syntax
```

Expected: the same persisted Markdown renders the same way in preview and final post detail.

- [ ] **Step 3: If embed syntax is intentionally changed, change every dependent path together**

If product explicitly requires `{{embed:post:uuid}}`, the minimum synchronized change list is:

```text
- AEditor.vue insertion helpers
- useMarkdownRenderer.ts directive/parser regexes
- PostDetailView.vue extractEmbedIds regexes
- any existing post-content migration/backfill plan
- manual verification of old vs new content behavior
```

Expected: no half-migration is allowed.

- [ ] **Step 4: Recheck image upload contract end-to-end**

The end-to-end path is:

```text
AEditor uploadImage(file)
→ POST /api/blog/upload-image
→ server/internal/handlers/blog_upload_handler.go validates file type/size
→ local or S3 storage writes file
→ handler returns { url }
→ AEditor replaces placeholder Markdown with final image Markdown
```

Expected: every successful upload yields a usable Markdown image URL in the editor document.

- [ ] **Step 5: Recheck read-path embed hydration**

`PostDetailView.vue` must continue to:

```text
- scan persisted content for post/music/video embed IDs
- fetch metadata for those IDs
- pass hydrated maps into renderMarkdown(...)
- render missing embeds as fallback cards instead of raw directive text
```

Expected: saved blog content degrades gracefully when referenced content is unavailable.

- [ ] **Step 6: Recheck collaboration assumptions without overclaiming**

Use these rules exactly during verification:

```text
- verify that editing an existing post opens the collaboration socket path /api/collab/ws/:roomID
- verify that a second browser window can join the same post room
- if real-time sync fails, capture the observable failure mode and the browser console error
- do not claim collaboration is fixed unless cross-window edits are visibly synchronized
```

Expected: collaboration status is reported from observed behavior, not from code optimism.

## Task 4: Reconcile all non-blog callers after shared-editor cleanup

**Files:**
- Modify: `web/src/views/forum/ForumNewTopicView.vue`
- Modify: `web/src/views/forum/ForumTopicView.vue`
- Modify: `web/src/views/music/AlbumDiscussionView.vue`
- Modify: `web/src/views/music/SongDiscussionView.vue`
- Modify: `web/src/views/debate/DebateTopicView.vue`

- [ ] **Step 1: Keep forum creation and reply flows on SV mode with mentions only**

Forum caller expectations:

```text
- mode="sv"
- enableMentions=true
- enableEmbeds=false
- enableCollab=false
```

Expected: forum users keep Markdown + mentions without blog-only controls.

- [ ] **Step 2: Keep music discussion flows on SV mode with mentions only**

Music caller expectations:

```text
- mode="sv"
- enableMentions=true
- enableEmbeds=false
- enableCollab=false
```

Expected: album and song discussions continue sharing the same editor behavior as forum discussions.

- [ ] **Step 3: Keep debate argument editing on plain mode**

Debate caller expectations:

```text
- mode="plain"
- no CodeMirror mount
- no toolbar
- no embed controls
- no collaboration socket
```

Expected: debate editing remains a simple textarea-based input path.

- [ ] **Step 4: Remove any caller props that no longer exist**

Search for and eliminate stale caller usage of removed editor props or legacy assumptions.

Run:

```bash
rg -n "mode=\"wysiwyg\"|enableCollab|enableEmbeds|collabRoomId|postId" web/src
```

Expected: only valid current props remain, and no caller references old wysiwyg mode.

## Task 5: Add regression coverage and remove only verified leftovers

**Files:**
- Create: `web/tests/e2e/editor-regression.spec.ts`
- Modify: `web/playwright.config.ts` only if test discovery/config actually needs it
- Modify: `web/tests/e2e/global-setup.ts` only if fixture bootstrapping is required for stable editor tests
- Modify: `web/package.json` only if stale dependencies are truly unused
- Modify: `web/vite.config.ts` only if editor-chunk references need cleanup

- [ ] **Step 1: Add an editor regression smoke spec**

Create a Playwright spec that covers the shared editor at the feature level. Minimum scenarios:

```ts
import { test, expect } from '@playwright/test'

test.describe('editor regression', () => {
  test('forum SV editor still accepts Markdown input', async ({ page }) => {
    // navigate to the forum topic creation surface
    // type Markdown into the editor
    // assert visible content updates and submit path remains enabled
  })

  test('debate plain editor stays textarea-based', async ({ page }) => {
    // navigate to a debate argument modal or page
    // assert plain textarea path is rendered
    // assert SV toolbar controls are absent
  })
})
```

Expected: at least one SV caller and one plain caller have automated regression coverage.

- [ ] **Step 2: Add a blog editor manual verification checklist if automation cannot be stabilized yet**

If the current dev dataset is too unstable for reliable blog E2E automation, keep the manual checklist in the implementation notes and do **not** fake a flaky browser test.

Use this exact manual checklist:

```text
1. Open /post/new?channel=<valid-channel-id>
2. Confirm the editor mounts in SV mode
3. Type Markdown and confirm the preview updates
4. Upload one image from toolbar and confirm placeholder replacement
5. Paste one image and confirm placeholder replacement
6. Insert one embed and save
7. Open the saved post and confirm the embed renders as a card
8. Open /post/<id>/edit in two windows and confirm observed collaboration behavior
```

Expected: blog-specific behavior is still fully verified even if only part of it is automated.

- [ ] **Step 3: Remove only stale references proven to be unused**

Run these exact audits after code changes:

```bash
rg -n "@tiptap|parseMarkdownToHtml|serializeTiptapToMarkdown|mode=\"wysiwyg\"" web
rg -n "y-websocket" web
```

Interpretation rules:

```text
- zero results for @tiptap and wysiwyg references are required
- y-websocket results are allowed if collaboration still depends on WebsocketProvider
- do not remove y-websocket just to make grep output empty
```

Expected: cleanup is evidence-based rather than aesthetic.

- [ ] **Step 4: Re-run the full verification suite after cleanup**

Run:

```bash
cd web && bun run type-check
cd web && bun run build
cd server && go build ./...
cd web && bun run test:e2e
```

Expected: compilation, build, backend, and browser regressions all pass after the final editor adjustments.

## Exact Verification Steps

Run these steps in order after the implementation work is complete.

### 1. Frontend static verification

```bash
cd web && bun run type-check
cd web && bun run build
```

Expected:

```text
- type-check exits 0
- build exits 0
- no editor-related TypeScript or bundling regressions
```

### 2. Backend compile verification

```bash
cd server && go build ./...
```

Expected:

```text
- build exits 0
- upload and collaboration handlers compile cleanly
```

### 3. Stale-reference audit

```bash
rg -n "@tiptap|parseMarkdownToHtml|serializeTiptapToMarkdown|mode=\"wysiwyg\"" web
rg -n "y-websocket" web
```

Expected:

```text
- no Tiptap or wysiwyg references remain
- y-websocket remains only if collaboration still uses WebsocketProvider
```

### 4. Shared-editor manual verification

Verify all of the following in the browser:

```text
- SV mode shows toolbar + CodeMirror + live preview
- plain mode shows textarea only
- scroll sync works both directions on a long document
- image upload works from toolbar
- image upload works from clipboard paste
- image upload works from drag/drop
- concurrent uploads replace only their own placeholders
- mentions appear only on callers with enableMentions=true
```

### 5. Blog-specific manual verification

Verify all of the following in the browser:

```text
- /post/new?channel=<valid-channel-id> mounts the blog editor
- the title field and body editor both remain editable
- embed insertion writes the expected persisted Markdown syntax
- preview renders embeds without showing raw directive text
- saved post detail renders embed cards with metadata or fallback cards
- /post/<id>/edit opens the collaboration-enabled editor path
- collaboration behavior is confirmed with two browser windows or explicitly marked unverified
```

### 6. Cross-feature caller verification

Verify all of the following in the browser:

```text
- forum new-topic editor still works in SV mode with mentions
- forum reply editor still works in SV mode with mentions
- music album discussion editor still works in SV mode with mentions
- music song discussion editor still works in SV mode with mentions
- debate argument editor still works in plain mode without SV toolbar
```

## Commit

```bash
git add web/src/components/shared/AEditor.vue web/src/composables/useMarkdownRenderer.ts web/src/views/blog/PostEditorView.vue web/src/views/blog/PostDetailView.vue web/src/views/forum/ForumNewTopicView.vue web/src/views/forum/ForumTopicView.vue web/src/views/music/AlbumDiscussionView.vue web/src/views/music/SongDiscussionView.vue web/src/views/debate/DebateTopicView.vue web/package.json web/vite.config.ts web/playwright.config.ts web/tests/e2e/global-setup.ts web/tests/e2e/editor-regression.spec.ts server/internal/handlers/blog_upload_handler.go server/internal/collab/hub.go server/cmd/start_server/main.go docs/superpowers/plans/2026-05-14-editor-implementation.md
git commit -m "docs: add editor superpowers implementation plan"
```

## Reuse Note For Sparse Plan Inputs

When a feature only has a task-plan source and lacks findings/progress documents, use this exact synthesis rule:

1. Read the task-plan source for target behavior and phase order.
2. Inspect the live repository to determine what is already implemented.
3. Record every divergence between the stale checklist and current code.
4. Preserve working contracts unless product explicitly reopens them.
5. Convert the result into a self-contained implementation plan with exact verification steps.
