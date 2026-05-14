# Podcast Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add the first podcast slice to Studio so channels can publish podcast episodes with uploaded audio, expose a podcast-specific RSS feed with enclosure metadata, and let listeners play episodes through the existing bottom audio player.

**Architecture:** Extend the existing Blog/Studio channel-plus-post model instead of creating a separate podcast domain. Represent podcast episodes as a content subtype of channel-owned content, reuse the existing upload and comment patterns, generate a dedicated public RSS route from the backend, and adapt the current music player store/UI so a podcast episode can be queued and played without breaking music playback.

**Tech Stack:** Vue 3, TypeScript, Vite, Pinia, Vue Router, Go, Gin, GORM, SQLite/PostgreSQL, existing AudioPlayer, existing Blog and Feed handlers

---

## Upstream Sources

Primary source used for this plan:
- `/Users/fafa/Documents/projects/Atoman/.claude/worktrees/agent-aa22d520ca6ee2b71/plan/podcast_task_plan.md`

Style reference used only for formatting expectations:
- `/Users/fafa/Documents/projects/Atoman/docs/superpowers/plans/2026-05-14-blog-implementation.md`

Missing upstream inputs explicitly acknowledged:
- `plan/podcast_findings.md` does not exist in the worktree.
- `plan/podcast_progress.md` does not exist in the worktree.

Because only a task-plan source exists, this implementation plan treats confirmed checklist items in `podcast_task_plan.md` as the source of truth, preserves unresolved questions as open verification items, and avoids claiming any pre-existing implementation or validation history.

## Normalized Current State

The repository already has the core primitives podcast should reuse:
- Studio/blog channels, collections, posts, and slug-based channel routing already exist.
- The frontend already mounts a global bottom audio player in `/Users/fafa/Documents/projects/Atoman/.claude/worktrees/agent-aa22d520ca6ee2b71/web/src/App.vue`.
- The current player state is song-centric in `/Users/fafa/Documents/projects/Atoman/.claude/worktrees/agent-aa22d520ca6ee2b71/web/src/stores/player.ts` and `/Users/fafa/Documents/projects/Atoman/.claude/worktrees/agent-aa22d520ca6ee2b71/web/src/components/AudioPlayer.vue`.
- Blog post CRUD, channel CRUD, comment flows, and feed subscription primitives already exist in the backend.
- RSS parsing for external feeds and internal feed subscription records already exist, but there is no evidence yet of a podcast-specific internal RSS publishing path.

That means podcast work should be planned as an extension of the current Studio/blog stack plus the shared player, not as a standalone subsystem.

## Confirmed Scope From `podcast_task_plan.md`

The upstream task plan explicitly confirms all of the following and this implementation plan preserves them as hard scope:
- Podcast is a Studio submodule.
- Show = channel identity.
- Episode = podcast content published inside a channel.
- Channels are not globally type-restricted; podcast is a content attribute, not a channel-only silo.
- Audio must be uploaded locally or through existing managed storage; no external audio URL option.
- Shownotes support only URL and bold formatting, not a full Markdown editor.
- No timestamp chapter support in the first version.
- Episode lifecycle includes draft and published.
- Scheduled publishing is required.
- Visibility includes public, followers-only, and private.
- Channel pages must expose type tabs: all / articles / videos / podcasts.
- Podcast RSS lives at `/channel/:slug/rss/podcast` and must include `<enclosure>`.
- Listening reuses the existing bottom playback bar and queues individual episodes.
- Comments are supported and should reuse the forum-style plain text plus image model.

## Open Questions To Preserve As Verification Targets

These are not blockers for planning, but they must not be silently assumed during implementation:
- Whether podcast RSS should be directly subscribable by the Feed module via first-class internal subscription flow.
- Whether the queue should support cross-show podcast playback semantics beyond one-episode-at-a-time enqueue.
- Whether show cover and episode cover should be separately managed in the first release or episode cover should fall back to channel cover.

## File Structure And Responsibility Map

### Backend files likely to modify
- `/Users/fafa/Documents/projects/Atoman/.claude/worktrees/agent-aa22d520ca6ee2b71/server/internal/model/feed.go`
  - Existing Studio/blog content models live here. Add podcast-capable content fields or carefully extract a compatible extension strategy.
- `/Users/fafa/Documents/projects/Atoman/.claude/worktrees/agent-aa22d520ca6ee2b71/server/cmd/start_server/main.go`
  - Register migrations for any new fields or models and register any new podcast routes.
- `/Users/fafa/Documents/projects/Atoman/.claude/worktrees/agent-aa22d520ca6ee2b71/server/internal/handlers/blog_post_handler.go`
  - Extend post creation/update/read flows if podcast episodes remain part of the post model.
- `/Users/fafa/Documents/projects/Atoman/.claude/worktrees/agent-aa22d520ca6ee2b71/server/internal/handlers/blog_channel_handler.go`
  - Extend channel read responses if podcast tabs, counts, or podcast RSS metadata are served from channel endpoints.
- `/Users/fafa/Documents/projects/Atoman/.claude/worktrees/agent-aa22d520ca6ee2b71/server/internal/handlers/blog_upload_handler.go`
  - Reuse or extend upload endpoints for episode audio and optional episode artwork.
- `/Users/fafa/Documents/projects/Atoman/.claude/worktrees/agent-aa22d520ca6ee2b71/server/internal/handlers/blog_interaction_handler.go`
  - Reuse comment/interaction handling if episode comments remain aligned with blog/forum comment behavior.
- `/Users/fafa/Documents/projects/Atoman/.claude/worktrees/agent-aa22d520ca6ee2b71/server/internal/handlers/feed_handler.go`
  - Potential extension point if podcast RSS becomes internally subscribable.
- `/Users/fafa/Documents/projects/Atoman/.claude/worktrees/agent-aa22d520ca6ee2b71/server/internal/service/rss_cron.go`
  - Only touch if internal podcast RSS generation or feed ingestion needs shared parsing/serialization helpers.

### Frontend files likely to modify
- `/Users/fafa/Documents/projects/Atoman/.claude/worktrees/agent-aa22d520ca6ee2b71/web/src/types.ts`
  - Add podcast episode, shownotes, visibility, scheduling, and player union types.
- `/Users/fafa/Documents/projects/Atoman/.claude/worktrees/agent-aa22d520ca6ee2b71/web/src/composables/useApi.ts`
  - Add podcast creation, listing, detail, upload, and RSS URL helpers.
- `/Users/fafa/Documents/projects/Atoman/.claude/worktrees/agent-aa22d520ca6ee2b71/web/src/router.ts`
  - Add any dedicated episode editor/detail routes if they are not embedded inside existing blog routes.
- `/Users/fafa/Documents/projects/Atoman/.claude/worktrees/agent-aa22d520ca6ee2b71/web/src/views/blog/ChannelView.vue`
  - Add the podcasts tab and episode listing within channel pages.
- `/Users/fafa/Documents/projects/Atoman/.claude/worktrees/agent-aa22d520ca6ee2b71/web/src/views/blog/PostEditorView.vue`
  - Reuse or branch editor behavior if podcast episode creation lives in the existing Studio editor shell.
- `/Users/fafa/Documents/projects/Atoman/.claude/worktrees/agent-aa22d520ca6ee2b71/web/src/components/blog/CommentSection.vue`
  - Reuse for episode discussion if detail pages share the same comment component.
- `/Users/fafa/Documents/projects/Atoman/.claude/worktrees/agent-aa22d520ca6ee2b71/web/src/stores/player.ts`
  - Generalize from `Song`-only playback to a typed audio item that can represent songs and podcast episodes.
- `/Users/fafa/Documents/projects/Atoman/.claude/worktrees/agent-aa22d520ca6ee2b71/web/src/components/AudioPlayer.vue`
  - Render shared playback UI for both songs and episodes without losing current music behavior.
- `/Users/fafa/Documents/projects/Atoman/.claude/worktrees/agent-aa22d520ca6ee2b71/web/src/App.vue`
  - Likely unchanged structurally, but part of verification because the global player mount must still work.

### New files likely to create
The exact final filenames can follow existing naming patterns, but implementation should prefer focused files rather than growing unrelated views. Likely additions include:
- Podcast-specific view components under `/Users/fafa/Documents/projects/Atoman/.claude/worktrees/agent-aa22d520ca6ee2b71/web/src/views/blog/` or a new `/web/src/views/podcast/` subtree if separation stays clean.
- Focused backend handler file(s) such as `server/internal/handlers/podcast_handler.go` if podcast logic becomes too large for `blog_post_handler.go`.
- Frontend E2E coverage under `/Users/fafa/Documents/projects/Atoman/.claude/worktrees/agent-aa22d520ca6ee2b71/web/tests/`.

## Implementation Strategy

Keep podcast within the Studio/blog mental model.
1. Introduce the minimum data model needed to distinguish podcast episodes from ordinary posts.
2. Extend backend CRUD and public read routes before building UI.
3. Reuse current player and comment systems by generalizing interfaces instead of duplicating playback or discussion code.
4. Ship podcast RSS once episode publication and public reads are stable.
5. Treat Feed-module integration as a follow-up slice unless the implementation naturally fits the existing internal subscription model without widening scope.

## Task 1: Lock the content model and API contract

**Files:**
- Modify: `/Users/fafa/Documents/projects/Atoman/.claude/worktrees/agent-aa22d520ca6ee2b71/server/internal/model/feed.go`
- Modify: `/Users/fafa/Documents/projects/Atoman/.claude/worktrees/agent-aa22d520ca6ee2b71/server/internal/handlers/blog_post_handler.go`
- Modify: `/Users/fafa/Documents/projects/Atoman/.claude/worktrees/agent-aa22d520ca6ee2b71/web/src/types.ts`
- Modify: `/Users/fafa/Documents/projects/Atoman/.claude/worktrees/agent-aa22d520ca6ee2b71/web/src/composables/useApi.ts`
- Test: `/Users/fafa/Documents/projects/Atoman/.claude/worktrees/agent-aa22d520ca6ee2b71/server/cmd/start_server/main.go`
- Test: `/Users/fafa/Documents/projects/Atoman/.claude/worktrees/agent-aa22d520ca6ee2b71/web/src/types.ts`

- [ ] **Step 1: Inspect the current post model and choose the smallest extension shape**

Review the existing `Post` model and decide explicitly between:
```text
A. Extending Post with podcast-specific fields such as content_type, audio_url, audio_source,
   shownotes, visibility, published_at, scheduled_for, and episode_cover_url.
B. Creating a dedicated PodcastEpisode model that still belongs to a channel and reuses comments/player.
```

Expected: a written decision in working notes that prefers the simpler option unless a dedicated model clearly avoids breaking blog assumptions.

- [ ] **Step 2: Update backend request/response structs to represent podcast metadata**

Add or extend API fields to support:
```text
- content_type: article | video | podcast
- audio_url / audio_source
- shownotes
- visibility: public | followers | private
- scheduled_for
- published_at
- optional episode-specific cover
```

Expected: create/update/read handlers can represent podcast episodes without overloading unrelated music-only fields.

- [ ] **Step 3: Add matching frontend types and API endpoints**

Update `/web/src/types.ts` and `/web/src/composables/useApi.ts` so the frontend has first-class types and helper URLs for:
```text
- podcast episode create/update/get/list
- channel podcast listing
- channel podcast RSS URL
- audio upload endpoint reuse or dedicated upload helper
```

Expected: no frontend code needs ad-hoc string concatenation for podcast APIs.

- [ ] **Step 4: Run backend compile verification**

Run:
```bash
cd "/Users/fafa/Documents/projects/Atoman/.claude/worktrees/agent-aa22d520ca6ee2b71/server" && go build ./...
```

Expected: PASS with no model or handler type errors.

- [ ] **Step 5: Run frontend type-check verification**

Run:
```bash
cd "/Users/fafa/Documents/projects/Atoman/.claude/worktrees/agent-aa22d520ca6ee2b71/web" && bun run type-check
```

Expected: PASS with no `types.ts` or API helper regressions.

## Task 2: Implement backend create/update/read flows for podcast episodes

**Files:**
- Modify: `/Users/fafa/Documents/projects/Atoman/.claude/worktrees/agent-aa22d520ca6ee2b71/server/internal/handlers/blog_post_handler.go`
- Modify: `/Users/fafa/Documents/projects/Atoman/.claude/worktrees/agent-aa22d520ca6ee2b71/server/internal/handlers/blog_channel_handler.go`
- Modify: `/Users/fafa/Documents/projects/Atoman/.claude/worktrees/agent-aa22d520ca6ee2b71/server/cmd/start_server/main.go`
- Test: podcast handler file(s) created or modified during implementation

- [ ] **Step 1: Add validation rules for podcast episode creation**

Enforce the confirmed upstream rules:
```text
- podcast episode must belong to a channel
- audio upload is required for podcast publication
- shownotes allow only URL and bold-safe formatting chosen for this feature
- scheduled publication accepts future timestamps
- visibility must be one of public / followers / private
```

Expected: invalid payloads fail early with field-specific errors.

- [ ] **Step 2: Implement list/detail filtering by content type and visibility**

Ensure the public and authenticated read paths can:
```text
- list only podcast episodes for a channel podcast tab or API filter
- hide followers-only/private episodes from unauthorized viewers
- expose scheduled items only to authorized owners before publication
```

Expected: the backend contract cleanly separates editorial state from public state.

- [ ] **Step 3: Register migrations and routes in server startup**

Update startup wiring so any new fields/models are migrated and any new handlers are reachable from the Gin router.

Expected: a local dev boot can serve podcast APIs without manual database patching.

- [ ] **Step 4: Run backend compile verification**

Run:
```bash
cd "/Users/fafa/Documents/projects/Atoman/.claude/worktrees/agent-aa22d520ca6ee2b71/server" && go build ./...
```

Expected: PASS with route and migration wiring intact.

- [ ] **Step 5: Smoke-test the create/read flow manually with local API requests**

Using the running local stack, verify:
```text
- create draft podcast episode
- update draft with uploaded audio metadata
- publish or schedule the episode
- fetch detail successfully
- list it from the owning channel’s podcast feed
```

Expected: draft and published podcast records behave like Studio content, not like detached music rows.

## Task 3: Reuse upload infrastructure for podcast audio

**Files:**
- Modify: `/Users/fafa/Documents/projects/Atoman/.claude/worktrees/agent-aa22d520ca6ee2b71/server/internal/handlers/blog_upload_handler.go`
- Modify: `/Users/fafa/Documents/projects/Atoman/.claude/worktrees/agent-aa22d520ca6ee2b71/web/src/composables/useApi.ts`
- Modify: podcast editor/view files introduced by implementation
- Test: upload endpoints and editor form flow

- [ ] **Step 1: Review current upload handler constraints and storage behavior**

Confirm how the existing blog upload path stores files and whether it already distinguishes by MIME type or object prefix.

Expected: a decision on whether podcast audio can safely reuse the existing handler or needs a dedicated endpoint with tighter file validation.

- [ ] **Step 2: Add podcast-safe upload validation**

Support at minimum:
```text
- accepted audio MIME types for podcast files
- reasonable max file size enforcement
- stored URL persisted in the episode record
```

Expected: image-only assumptions do not silently reject or corrupt audio uploads.

- [ ] **Step 3: Wire the editor form to upload and persist episode audio**

Verify the frontend can:
```text
- select a local audio file
- upload successfully
- persist the returned URL/source
- prevent publication when no audio URL is present
```

Expected: authors cannot publish a broken podcast item with missing media.

- [ ] **Step 4: Run backend and frontend verification**

Run:
```bash
cd "/Users/fafa/Documents/projects/Atoman/.claude/worktrees/agent-aa22d520ca6ee2b71/server" && go build ./...
cd "/Users/fafa/Documents/projects/Atoman/.claude/worktrees/agent-aa22d520ca6ee2b71/web" && bun run type-check
```

Expected: PASS for both commands.

- [ ] **Step 5: Browser-test audio upload end to end**

Manually verify:
```text
- create or edit an episode
- upload local audio
- save draft
- reload editor
- confirm audio metadata is still attached
```

Expected: uploaded episode media survives round-trip editing.

## Task 4: Build the podcast authoring UI inside Studio

**Files:**
- Modify: `/Users/fafa/Documents/projects/Atoman/.claude/worktrees/agent-aa22d520ca6ee2b71/web/src/views/blog/PostEditorView.vue`
- Modify or create: podcast-specific form components under `/Users/fafa/Documents/projects/Atoman/.claude/worktrees/agent-aa22d520ca6ee2b71/web/src/components/`
- Modify: `/Users/fafa/Documents/projects/Atoman/.claude/worktrees/agent-aa22d520ca6ee2b71/web/src/router.ts`
- Modify: `/Users/fafa/Documents/projects/Atoman/.claude/worktrees/agent-aa22d520ca6ee2b71/web/src/types.ts`
- Test: browser authoring flow

- [ ] **Step 1: Decide whether podcast uses a dedicated editor route or a mode within the existing post editor**

Choose the approach that minimizes duplication while keeping the form understandable.

Expected: one clear authoring entry point for podcast episodes.

- [ ] **Step 2: Add podcast-specific form controls**

The form must support:
```text
- episode title
- optional summary
- shownotes with only URL and bold-capable formatting
- draft / publish / schedule actions
- visibility selector
- channel selection
- audio upload status
- optional episode cover if implemented in this slice
```

Expected: all confirmed podcast planning fields are editable without exposing a full Markdown authoring surface for shownotes.

- [ ] **Step 3: Preserve existing design-system rules**

Use existing shared UI primitives where possible:
```text
- ABtn
- AInput
- ATextarea
- ASelect
- AModal
```

Expected: no new podcast-specific visual language drifts away from the archive-style UI conventions.

- [ ] **Step 4: Run frontend verification**

Run:
```bash
cd "/Users/fafa/Documents/projects/Atoman/.claude/worktrees/agent-aa22d520ca6ee2b71/web" && bun run type-check && bun run build
```

Expected: PASS for type-check and production build.

- [ ] **Step 5: Browser-test the author workflow**

Manually verify:
```text
- create a podcast draft
- save without publishing
- reopen and edit
- schedule publication
- publish immediately
- confirm validation errors are clear when audio is missing
```

Expected: the Studio author flow is usable without any hidden backend-only steps.

## Task 5: Add podcast playback support to the shared bottom player

**Files:**
- Modify: `/Users/fafa/Documents/projects/Atoman/.claude/worktrees/agent-aa22d520ca6ee2b71/web/src/stores/player.ts`
- Modify: `/Users/fafa/Documents/projects/Atoman/.claude/worktrees/agent-aa22d520ca6ee2b71/web/src/components/AudioPlayer.vue`
- Modify: podcast detail/list UI components that trigger playback
- Test: shared player flows in browser

- [ ] **Step 1: Generalize player state from `Song`-only to a shared audio item shape**

Create a common playback contract that can represent both music and podcast content, for example:
```text
- id
- title
- audio_url
- cover_url
- creator display text
- kind: song | podcast
```

Expected: podcast support does not fork the entire player store.

- [ ] **Step 2: Keep existing music behavior working while adding episode queueing**

Verify that player actions still support:
```text
- current song playback
- album queue playback
- podcast episode single enqueue or list queue
- play / pause / next / previous / seek / volume
```

Expected: podcast playback is additive, not a regression for music.

- [ ] **Step 3: Update player UI labels and metadata rendering**

Show sensible creator text for podcast episodes, such as channel name or author/channel combination, while preserving artist rendering for songs.

Expected: the bottom player never displays empty metadata for podcast items.

- [ ] **Step 4: Run frontend verification**

Run:
```bash
cd "/Users/fafa/Documents/projects/Atoman/.claude/worktrees/agent-aa22d520ca6ee2b71/web" && bun run type-check && bun run build
```

Expected: PASS with no player type regressions.

- [ ] **Step 5: Browser-test mixed playback**

Manually verify:
```text
- play a song from the music module
- switch to a podcast episode
- pause/resume
- seek through the episode
- return to music playback
```

Expected: shared player state remains stable across content kinds.

## Task 6: Add public channel podcast tab and episode detail surfaces

**Files:**
- Modify: `/Users/fafa/Documents/projects/Atoman/.claude/worktrees/agent-aa22d520ca6ee2b71/web/src/views/blog/ChannelView.vue`
- Modify or create: podcast list/detail views in `/Users/fafa/Documents/projects/Atoman/.claude/worktrees/agent-aa22d520ca6ee2b71/web/src/views/`
- Modify: `/Users/fafa/Documents/projects/Atoman/.claude/worktrees/agent-aa22d520ca6ee2b71/web/src/router.ts`
- Modify: `/Users/fafa/Documents/projects/Atoman/.claude/worktrees/agent-aa22d520ca6ee2b71/web/src/components/blog/CommentSection.vue` if needed
- Test: browser public-read flows

- [ ] **Step 1: Add the podcasts tab to channel pages**

Support the confirmed tab model:
```text
- 全部
- 文章
- 视频
- 播客
```

Expected: the channel page can filter down to podcast episodes without losing the existing all-content view.

- [ ] **Step 2: Build episode cards and detail rendering**

The public UI should expose:
```text
- title
- summary
- show/channel identity
- publish date
- cover art
- play action
- shownotes
- comment section
```

Expected: public podcast pages feel like Studio content with audio-first affordances.

- [ ] **Step 3: Respect visibility and scheduling in the UI**

Ensure hidden, followers-only, and future-scheduled episodes do not appear incorrectly in public listings.

Expected: frontend behavior matches the backend visibility contract.

- [ ] **Step 4: Run frontend verification**

Run:
```bash
cd "/Users/fafa/Documents/projects/Atoman/.claude/worktrees/agent-aa22d520ca6ee2b71/web" && bun run type-check && bun run build
```

Expected: PASS.

- [ ] **Step 5: Browser-test the public consumption flow**

Manually verify:
```text
- open /channel/:slug
- switch to the podcast tab
- open an episode detail page
- play the episode from the detail page
- add and read comments
```

Expected: listeners can discover and consume episodes from existing channel identity pages.

## Task 7: Generate podcast RSS at `/channel/:slug/rss/podcast`

**Files:**
- Modify: `/Users/fafa/Documents/projects/Atoman/.claude/worktrees/agent-aa22d520ca6ee2b71/server/internal/handlers/blog_channel_handler.go` or a dedicated RSS/podcast handler file
- Modify: `/Users/fafa/Documents/projects/Atoman/.claude/worktrees/agent-aa22d520ca6ee2b71/server/cmd/start_server/main.go`
- Modify: `/Users/fafa/Documents/projects/Atoman/.claude/worktrees/agent-aa22d520ca6ee2b71/web/src/composables/useApi.ts`
- Test: XML output in browser or curl

- [ ] **Step 1: Define the RSS output contract from confirmed scope**

The feed must include at minimum:
```xml
<rss>
  <channel>
    <title>...</title>
    <link>...</link>
    <description>...</description>
    <item>
      <title>...</title>
      <link>...</link>
      <guid>...</guid>
      <pubDate>...</pubDate>
      <description>...</description>
      <enclosure url="..." type="audio/..." length="..." />
    </item>
  </channel>
</rss>
```

Expected: a documented minimum contract that is sufficient for podcast clients even if richer iTunes tags come later.

- [ ] **Step 2: Implement the public route and serialize only published, visible podcast episodes**

Exclude:
```text
- drafts
- private episodes
- followers-only episodes unless public exposure is explicitly intended
- future-scheduled items not yet publishable
```

Expected: feed consumers only see externally valid podcast episodes.

- [ ] **Step 3: Add an exact browser/API verification for the route**

Verify:
```bash
curl -i "http://localhost:8080/channel/<slug>/rss/podcast"
```

Expected:
```text
- HTTP 200
- Content-Type indicates XML
- response body includes <rss, <channel, <item, and <enclosure
```

- [ ] **Step 4: Run backend compile verification**

Run:
```bash
cd "/Users/fafa/Documents/projects/Atoman/.claude/worktrees/agent-aa22d520ca6ee2b71/server" && go build ./...
```

Expected: PASS.

- [ ] **Step 5: Validate the feed with at least one real episode**

Manually verify:
```text
- publish one episode
- request the RSS URL
- confirm enclosure URL resolves to uploaded audio
- confirm title/link/pubDate match the episode
```

Expected: the feed is externally consumable, not just syntactically present.

## Task 8: Decide and implement the minimum Feed-module integration

**Files:**
- Modify: `/Users/fafa/Documents/projects/Atoman/.claude/worktrees/agent-aa22d520ca6ee2b71/server/internal/handlers/feed_handler.go` if needed
- Modify: `/Users/fafa/Documents/projects/Atoman/.claude/worktrees/agent-aa22d520ca6ee2b71/server/internal/model/feed.go` if needed
- Modify: feed-related frontend files only if a UI surface is added in this slice
- Test: subscription and timeline behavior if implemented

- [ ] **Step 1: Make an explicit scope call before coding**

Choose one of these and document it in working notes:
```text
A. Defer first-class Feed integration and rely on the public RSS URL only.
B. Add internal subscription support for podcast channel RSS in this slice.
```

Expected: no accidental partial integration.

- [ ] **Step 2: If integrating now, map podcast RSS onto existing feed source rules**

Reuse existing patterns in `CreateSubscription` and internal source types where possible.

Expected: Feed integration does not invent a second subscription model.

- [ ] **Step 3: Verify the chosen path**

If deferred:
```text
- ensure the public RSS URL is easy to copy from the channel UI
```

If implemented:
```text
- subscribe successfully
- trigger sync
- confirm items appear in timeline with enclosure metadata preserved
```

Expected: the decision is testable either way.

## Task 9: End-to-end verification and scope reconciliation

**Files:**
- Inspect: `/Users/fafa/Documents/projects/Atoman/.claude/worktrees/agent-aa22d520ca6ee2b71/plan/podcast_task_plan.md`
- Test: backend and frontend commands
- Test: browser flows across Studio, channel pages, player, and RSS

- [ ] **Step 1: Run all compile/build checks**

Run:
```bash
cd "/Users/fafa/Documents/projects/Atoman/.claude/worktrees/agent-aa22d520ca6ee2b71/server" && go build ./...
cd "/Users/fafa/Documents/projects/Atoman/.claude/worktrees/agent-aa22d520ca6ee2b71/web" && bun run type-check && bun run build
```

Expected: PASS for all commands.

- [ ] **Step 2: Run browser verification against the confirmed upstream checklist**

Verify all of the following:
```text
- draft and publish flow works
- scheduled publish can be configured
- visibility options are present and enforced
- channel page has the podcast tab
- bottom player can play an episode
- comments are available on episode detail
- RSS route returns podcast XML with enclosure
```

Expected: each upstream confirmed decision has a visible implementation checkpoint.

- [ ] **Step 3: Record the still-open product questions as follow-up work, not hidden debt**

Explicitly note whether the implemented slice includes:
```text
- Feed integration
- cross-show queueing behavior
- separate episode cover management
```

Expected: future work starts from facts instead of assumptions.

## Verification

After implementation, verify all of the following exactly:

1. `cd "/Users/fafa/Documents/projects/Atoman/.claude/worktrees/agent-aa22d520ca6ee2b71/server" && go build ./...` passes.
2. `cd "/Users/fafa/Documents/projects/Atoman/.claude/worktrees/agent-aa22d520ca6ee2b71/web" && bun run type-check` passes.
3. `cd "/Users/fafa/Documents/projects/Atoman/.claude/worktrees/agent-aa22d520ca6ee2b71/web" && bun run build` passes.
4. Creating a podcast draft without audio cannot be published.
5. Uploading a local audio file persists and survives editor reload.
6. `/channel/:slug` exposes a podcast tab and lists published podcast episodes only.
7. Opening an episode detail page allows playback through the shared bottom player.
8. Song playback still works after the player is generalized for podcasts.
9. `curl -i "http://localhost:8080/channel/<slug>/rss/podcast"` returns HTTP 200 and XML containing `<rss`, `<channel`, `<item`, and `<enclosure`.
10. Any deferred decisions from the original task plan are written down explicitly in implementation notes or follow-up planning.

## Commit

```bash
git add "/Users/fafa/Documents/projects/Atoman/.claude/worktrees/agent-aa22d520ca6ee2b71/docs/superpowers/plans/2026-05-14-podcast-implementation.md"
git commit -m "docs: add podcast superpowers implementation plan"
```
