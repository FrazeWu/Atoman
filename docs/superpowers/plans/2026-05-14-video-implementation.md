# Video Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build the Studio video module as a first-class channel-owned content type with discovery, playback, management, typed RSS, and comment flows, while reusing existing channel, auth, follow, upload, and editor infrastructure.

**Architecture:** Introduce a dedicated Video domain in both backend and frontend instead of overloading blog posts. Keep video-specific data, routes, and views in their own files, then integrate them into the existing Studio channel shell, follow graph, RSS/feed helpers, and blog embed pipeline. The current codebase already has placeholder `video` embed support in the shared Markdown/editor layer, so implementation should preserve that directive format and replace placeholder resolution with real video data.

**Tech Stack:** Vue 3, TypeScript, Vite, Pinia, Vue Router, Go, Gin, GORM, SQLite/Postgres, local storage or S3-compatible storage

---

## Upstream Sources

- Primary planning input: `plan/video/plan.md`
- Style reference only: `docs/superpowers/plans/2026-05-14-blog-implementation.md`

## Missing Upstream Inputs

The repository currently does **not** provide the full planning trio for Video. This plan is therefore based on the task plan alone and makes its assumptions explicit.

Missing inputs:
- `plan/video_findings.md` — not present
- `plan/video_progress.md` — not present
- Any video-specific page map / IA document equivalent to `plan/music_pages.md` — not present

Planning consequence:
- No existing implementation evidence can be trusted for Video because no progress log exists.
- No prior technical decision log exists beyond the decisions embedded in `plan/video/plan.md`.
- Any question still marked “待讨论” in the task plan must stay out of MVP scope unless implementation work explicitly re-opens it.

## Normalized Current State

The upstream Video task plan says product-detail confirmation is complete and implementation planning is the next step. The current repository state is consistent with that claim: there is **no** first-class Video module yet in routes, API handlers, models, or views.

What already exists today:
- Channel-based creator identity already exists and is the correct ownership model for video:
  - `server/internal/model/feed.go`
  - `server/internal/handlers/blog_channel_handler.go`
  - `web/src/views/blog/ChannelView.vue`
- Follow relationships already exist and can power follower-only visibility:
  - `server/internal/model/user.go`
  - `server/internal/handlers/user_handler.go`
- Local/S3 upload patterns already exist and can be reused for video files, subtitles, and covers:
  - `server/internal/handlers/blog_upload_handler.go`
  - `server/internal/handlers/songs_handler.go`
- Forum-style reply UX already exists and is the closest comment interaction model:
  - `server/internal/handlers/forum_handler.go`
  - `web/src/views/forum/ForumTopicView.vue`
- Shared editor/Markdown infrastructure already recognizes `video` embeds, but only as placeholders:
  - `web/src/components/shared/AEditor.vue`
  - `web/src/composables/useMarkdownRenderer.ts`
  - `web/src/views/blog/PostDetailView.vue`

What does **not** exist yet:
- No `Video` model in `server/internal/model/`
- No `/api/videos` backend routes
- No `/video` frontend routes or views
- No dedicated player page
- No typed channel video RSS endpoint
- No real video embed resolution in blog post rendering

## Confirmed Scope To Preserve

The following scope is directly confirmed by `plan/video/plan.md` and should be treated as the implementation contract:

- Video is a standalone Studio submodule with its own discovery page, player page, and manage page.
- Video belongs to the Studio channel system; videos are channel-owned content.
- Video creation supports two source modes:
  - local upload
  - external link (YouTube, Bilibili, etc.)
- Subtitle upload is optional.
- Description formatting is intentionally limited to:
  - URLs
  - bold text
- Publication states must support:
  - draft
  - published
  - scheduled publish behavior
- Visibility must support:
  - public
  - followers only
  - private
- Channel home pages must expose tabs for:
  - all
  - posts
  - videos
  - podcasts
- Typed RSS must exist at: `/channel/:slug/rss/video`
- Player page layout should follow the YouTube mental model:
  - primary player + title/meta
  - right-side recommended videos
  - comments below
- Comments should reuse the forum-style plain-text + image experience rather than inventing a third comment system.

## Out Of Scope For MVP Unless Reopened

These are explicitly unresolved in the upstream task plan and should **not** silently expand MVP scope:

- Recommendation algorithm sophistication (same-channel vs tag-based vs popularity-based ranking)
- Playback analytics such as view count depth or completion rate
- Video playlists / collections / series handling

## Working Assumptions Required Because Findings/Progress Are Missing

To keep this plan executable, the following implementation assumptions are locked in unless a later design round changes them:

1. **Video remains a first-class domain model, not a `Post` subtype.**
   This keeps player, source-mode, subtitles, scheduling, and visibility rules from leaking into blog post code.

2. **Scheduling uses timestamps, not a separate public lifecycle enum.**
   The persisted model should support draft vs published while also storing `scheduled_for` and `published_at`. Public queries should treat future-scheduled content as not yet public.

3. **Limited description formatting is implemented with a dedicated formatter/helper, not the full blog editor.**
   The blog editor is more powerful than the confirmed product requirements.

4. **Follower-only visibility is enforced via existing `follows` relationships between the viewer and the channel owner.**
   No new relationship model should be introduced.

5. **Recommendation MVP is deterministic and simple.**
   Default order: same channel first, then most recent published public videos, excluding the current video.

6. **Forum-style comments means reusing the same interaction pattern, not sharing the exact forum tables.**
   The UI may reuse `AEditor`/Markdown/image behavior, but video comments should stay attached to videos rather than forum topics.

## Critical Implementation Files To Create Or Modify

### Backend

**Create:**
- `server/internal/model/video.go`
- `server/internal/handlers/video_handler.go`
- `server/internal/handlers/video_upload_handler.go`
- `server/internal/service/video_description.go`

**Modify:**
- `server/cmd/start_server/main.go`
- `server/internal/handlers/feed_handler.go`
- `server/internal/handlers/blog_channel_handler.go`
- `server/internal/model/feed.go`
- `server/internal/model/user.go` (only if helper/query additions are needed for visibility checks)

### Frontend

**Create:**
- `web/src/views/video/VideoHomeView.vue`
- `web/src/views/video/VideoDetailView.vue`
- `web/src/views/video/VideoManageView.vue`
- `web/src/views/video/VideoEditorView.vue`
- `web/src/components/video/VideoCard.vue`
- `web/src/components/video/VideoSourcePlayer.vue`
- `web/src/components/video/VideoDescription.vue`
- `web/src/components/video/VideoCommentSection.vue`
- `web/src/components/video/VideoForm.vue`

**Modify:**
- `web/src/router.ts`
- `web/src/types.ts`
- `web/src/composables/useApi.ts`
- `web/src/views/blog/ChannelView.vue`
- `web/src/views/blog/ChannelManageDetailView.vue`
- `web/src/views/blog/PostDetailView.vue`
- `web/src/composables/useMarkdownRenderer.ts` (only if small real-data rendering adjustments are needed)
- `web/src/components/shared/AEditor.vue` (only if the existing `VIDEO` embed insertion flow needs UX polish; do not change the directive format)

## Task 1: Establish the backend Video domain and publication rules

**Files:**
- Create: `server/internal/model/video.go`
- Modify: `server/cmd/start_server/main.go`
- Inspect: `server/internal/model/feed.go`
- Inspect: `server/internal/model/user.go`

- [ ] **Step 1: Add first-class Video models**

Create dedicated models for:
- `Video`
- `VideoComment`

The `Video` model should include at least these responsibilities:
- channel ownership (`channel_id`)
- creator ownership (`user_id`)
- title
- limited-format description source
- cover image URL
- source mode (`local` or `external`)
- local media URL
- external video URL
- optional subtitle URL
- status (`draft` or `published`)
- visibility (`public`, `followers`, `private`)
- `scheduled_for`
- `published_at`
- `allow_comments`

The `VideoComment` model should include at least:
- `video_id`
- `user_id`
- raw text/Markdown content
- visibility/status fields if moderation is required

Expected: Video-specific persistence exists without modifying blog post tables to carry video-specific fields.

- [ ] **Step 2: Register migrations in server startup**

Add the new video models to the existing `db.AutoMigrate(...)` sequence in `server/cmd/start_server/main.go`.

Expected: Video schema creation happens through the same migration path already used by blog, feed, forum, and music modules.

- [ ] **Step 3: Lock in read/write lifecycle rules**

Implement backend query rules so that:
- owners can read their own drafts, future-scheduled items, and private videos
- anonymous/public readers only see videos where:
  - `status = published`
  - `scheduled_for IS NULL` or `scheduled_for <= now`
  - visibility allows access
- follower-only videos require an authenticated viewer who follows the channel owner

Expected: scheduling and visibility are enforced at the data-access layer rather than only in the UI.

- [ ] **Step 4: Verify the backend still compiles**

Run:
```bash
cd /Users/fafa/Documents/projects/Atoman/server && go build ./...
```

Expected: successful Go build with no undefined model references or migration errors.

## Task 2: Add backend video CRUD, media upload, comments, and RSS endpoints

**Files:**
- Create: `server/internal/handlers/video_handler.go`
- Create: `server/internal/handlers/video_upload_handler.go`
- Create: `server/internal/service/video_description.go`
- Modify: `server/cmd/start_server/main.go`
- Modify: `server/internal/handlers/feed_handler.go`
- Modify: `server/internal/handlers/blog_channel_handler.go`

- [ ] **Step 1: Add public and protected video endpoints**

Create a dedicated route setup function that exposes:
- public discovery/list endpoint
- public detail endpoint
- owner manage list endpoint
- create/update/delete endpoints
- explicit publish/unpublish behavior if needed for parity with blog flows

Recommended endpoint surface:
- `GET /api/videos`
- `GET /api/videos/:id`
- `GET /api/videos/channel/:channelID`
- `GET /api/videos/channel/slug/:slug`
- `GET /api/videos/manage`
- `POST /api/videos`
- `PUT /api/videos/:id`
- `DELETE /api/videos/:id`

Expected: Video behavior lives behind its own handler file and does not bloat blog post routes.

- [ ] **Step 2: Add upload handling for local video, cover, and subtitle files**

Reuse the storage approach already used by blog images and music uploads.

Support at least these cases:
- cover upload
- local video file upload
- optional subtitle upload
- external-link mode with no local video file

Expected: local uploads work with both local filesystem mode and S3 mode, and external-link videos skip unnecessary upload requirements.

- [ ] **Step 3: Add limited description formatting validation**

In `server/internal/service/video_description.go`, centralize the logic that accepts only the confirmed subset of formatting.

Minimum behavior:
- preserve plain text
- preserve safe URLs
- preserve bold markers
- reject or strip unsupported rich formatting if needed

Expected: the description contract is enforced once and reused by create/update handlers.

- [ ] **Step 4: Add dedicated video comment endpoints with forum-style UX constraints**

Provide comment endpoints attached to videos rather than to posts or forum topics.

Recommended endpoint surface:
- `GET /api/videos/:id/comments`
- `POST /api/videos/:id/comments`
- `DELETE /api/video-comments/:id`

Behavior to preserve:
- plain-text / Markdown-like authoring
- image-compatible authoring path via the existing editor/upload capabilities
- authenticated creation
- deletion by comment owner, video owner, or admin

Expected: the UX can mirror forum discussions without forcing videos into the forum domain model.

- [ ] **Step 5: Add typed channel RSS for video-only output**

Expose:
- `GET /channel/:slug/rss/video`

Reuse the existing RSS generation patterns from `server/internal/handlers/feed_handler.go`, but limit the feed items to channel-owned videos that are publicly visible and effectively published.

Expected: users can subscribe to just the video stream for a channel.

- [ ] **Step 6: Verify backend compile after route wiring**

Run:
```bash
cd /Users/fafa/Documents/projects/Atoman/server && go build ./...
```

Expected: successful Go build with routes wired into `server/cmd/start_server/main.go`.

## Task 3: Add frontend types, API helpers, and route entries

**Files:**
- Modify: `web/src/types.ts`
- Modify: `web/src/composables/useApi.ts`
- Modify: `web/src/router.ts`

- [ ] **Step 1: Add typed frontend interfaces for Video and VideoComment**

Extend `web/src/types.ts` with types that match the backend response shapes for:
- `Video`
- `VideoComment`
- any small supporting enums/unions such as source mode and visibility

Expected: video views can be written with proper typing instead of `any` payloads.

- [ ] **Step 2: Add dedicated API helper entries**

Add `video` API helpers to `web/src/composables/useApi.ts` for:
- discovery/list
- detail
- create/update/delete
- manage list
- upload endpoints
- comments
- typed RSS URL helper if useful in the UI

Expected: video fetches use the same centralized helper style already used by blog, feed, auth, and users.

- [ ] **Step 3: Register frontend routes**

Add route entries in `web/src/router.ts` for at least:
- `/video`
- `/video/manage`
- `/video/new`
- `/video/:id`
- `/video/:id/edit`

Use direct import/lazy import behavior consistent with the repository router rules in `CLAUDE.md`.

Expected: the module has a routable shell before view implementation begins.

- [ ] **Step 4: Verify frontend typing and build still pass**

Run:
```bash
cd /Users/fafa/Documents/projects/Atoman/web && bun run type-check
```

Then run:
```bash
cd /Users/fafa/Documents/projects/Atoman/web && bun run build
```

Expected: TypeScript and production build both succeed with the new route and type additions.

## Task 4: Build the discovery, editor, manage, and player views

**Files:**
- Create: `web/src/views/video/VideoHomeView.vue`
- Create: `web/src/views/video/VideoDetailView.vue`
- Create: `web/src/views/video/VideoManageView.vue`
- Create: `web/src/views/video/VideoEditorView.vue`
- Create: `web/src/components/video/VideoCard.vue`
- Create: `web/src/components/video/VideoSourcePlayer.vue`
- Create: `web/src/components/video/VideoDescription.vue`
- Create: `web/src/components/video/VideoCommentSection.vue`
- Create: `web/src/components/video/VideoForm.vue`

- [ ] **Step 1: Build the public discovery page**

Implement `/video` as a discovery surface, not an owner dashboard.

The page should support at minimum:
- newest published public videos
- reusable `VideoCard` display
- channel/title/meta visibility
- empty/loading states matching the project design language

Expected: video discovery is independent from `/blog` while following the same black/white archive UI system.

- [ ] **Step 2: Build the create/edit form around the confirmed product contract**

Implement `VideoEditorView.vue` and/or `VideoForm.vue` so owners can:
- choose source mode: local upload vs external link
- upload/select cover
- upload optional subtitle file
- enter title
- enter limited-format description
- choose visibility
- choose draft vs publish intent
- set scheduled publish time
- toggle comment availability
- choose target channel

Expected: the form supports the entire confirmed MVP contract without adding playlist, analytics, or advanced recommendation controls.

- [ ] **Step 3: Build the owner manage page**

Implement `/video/manage` for authenticated owners.

The page should show at minimum:
- draft vs published state
- future-scheduled status labeling
- visibility label
- source mode label
- links to edit/view
- empty/loading states

Expected: creators can review and manage their own video inventory without going through the public discovery page.

- [ ] **Step 4: Build the player page with the confirmed layout**

Implement `/video/:id` as a YouTube-style page:
- main player area at the top/left
- title and channel/meta near the player
- recommended videos sidebar on the right
- comments below the player

For player behavior:
- local uploads should render through HTML5 video playback
- external links should embed or link out using the safest provider-compatible approach the implementation chooses
- subtitle attachment should render only when available

Expected: the page matches the task-plan layout without requiring a complex ranking system.

- [ ] **Step 5: Verify the frontend compiles after view creation**

Run:
```bash
cd /Users/fafa/Documents/projects/Atoman/web && bun run type-check
```

Then run:
```bash
cd /Users/fafa/Documents/projects/Atoman/web && bun run build
```

Expected: all new video views and components compile successfully.

## Task 5: Integrate Video into channel pages and real blog embed resolution

**Files:**
- Modify: `web/src/views/blog/ChannelView.vue`
- Modify: `web/src/views/blog/ChannelManageDetailView.vue`
- Modify: `web/src/views/blog/PostDetailView.vue`
- Modify: `web/src/components/shared/AEditor.vue` (only if the embed insertion UX needs small adjustments)
- Modify: `web/src/composables/useMarkdownRenderer.ts` (only if needed for final links/labels)

- [ ] **Step 1: Add channel-level content tabs**

Update channel pages so the public-facing channel surface can switch between:
- all
- posts
- videos
- podcasts

For this implementation pass, podcast may remain a non-active placeholder if the route/module does not yet exist, but the tab structure for video must be real and functional.

Expected: channel pages reflect the confirmed multi-content Studio model rather than a posts-only layout.

- [ ] **Step 2: Add channel management navigation for videos**

Update channel management surfaces so owners can reach video creation and video management from the existing channel management entry points.

Expected: video is visible as a first-class creator workflow inside Studio management, not an orphan route.

- [ ] **Step 3: Replace placeholder video embed resolution in blog post rendering**

Today `web/src/views/blog/PostDetailView.vue` creates placeholder embed cards for `:::video{id="..."}:::` directives.

Replace that placeholder path with real data fetching from the new video API so embeds show:
- actual title
- actual summary/description snippet
- actual channel/meta text
- actual video detail URL

Expected: existing `video` embed directive support becomes production-ready without changing the directive syntax.

- [ ] **Step 4: Keep the shared editor directive contract stable**

Do **not** invent a second embed format. Preserve the current `:::video{id="UUID"}:::` directive shape already recognized by:
- `web/src/components/shared/AEditor.vue`
- `web/src/composables/useMarkdownRenderer.ts`
- `web/src/views/blog/PostDetailView.vue`

Expected: blog content that references videos remains compatible with the existing shared editor philosophy.

- [ ] **Step 5: Verify embed and channel integration in the browser**

After the module is implemented, manually verify:
- a real video appears in the channel video tab
- a real blog post using the `video` directive renders an actual video card instead of placeholder text
- clicking the embed card opens the video detail page

Expected: video is integrated into both the Studio shell and the blog embedding path.

## Task 6: Exact end-to-end verification sequence

**Files:**
- Test: backend Video routes and models
- Test: frontend Video routes and views
- Test: channel integration in `web/src/views/blog/ChannelView.vue`
- Test: embed integration in `web/src/views/blog/PostDetailView.vue`

- [ ] **Step 1: Run static verification commands**

Run:
```bash
cd /Users/fafa/Documents/projects/Atoman/server && go build ./...
```

Run:
```bash
cd /Users/fafa/Documents/projects/Atoman/web && bun run type-check
```

Run:
```bash
cd /Users/fafa/Documents/projects/Atoman/web && bun run build
```

Expected: all three commands pass.

- [ ] **Step 2: Start the backend with explicit local dev env vars if needed**

Run:
```bash
cd /Users/fafa/Documents/projects/Atoman/server && JWT_SECRET=dev DATABASE_TYPE=sqlite DATABASE_URL=dev.sqlite STORAGE_TYPE=local GIN_MODE=debug go run cmd/start_server/main.go
```

Expected: the server starts, migrates the new video tables, and serves API routes locally.

- [ ] **Step 3: Start the frontend dev server**

Run:
```bash
cd /Users/fafa/Documents/projects/Atoman/web && bun run dev
```

Expected: the Vite dev server starts successfully.

- [ ] **Step 4: Verify creator flows in the browser**

Manual checklist:
- log in as a creator-capable account
- ensure a default channel exists
- create one local-upload video draft with cover and no subtitle
- create one external-link video with optional subtitle left empty
- edit one of the videos and set a future schedule time
- confirm the manage page labels draft/published/scheduled states correctly

Expected: core authoring and management workflows behave correctly.

- [ ] **Step 5: Verify public reader flows in the browser**

Manual checklist:
- open `/video` and confirm discovery cards render
- open a published public video at `/video/:id`
- confirm the player renders the correct source mode
- confirm the right-side recommendation list excludes the current video
- confirm comments render below the player
- confirm the channel link on the player page resolves correctly

Expected: public discovery and playback match the intended YouTube-style mental model.

- [ ] **Step 6: Verify channel and RSS integration**

Manual checklist:
- open `/channel/:slug`
- switch to the video tab
- confirm channel videos list renders only video content in that tab
- open `/channel/:slug/rss/video`
- confirm the response is RSS XML and only includes published public videos from that channel

Expected: typed channel video RSS and channel-tab integration both work.

- [ ] **Step 7: Verify visibility rules explicitly**

Manual checklist:
- mark one video `private` and confirm it is inaccessible to another user or anonymous visitor
- mark one video `followers` and confirm it is hidden from anonymous visitors
- if a second user is available, make that user follow the owner and confirm the follower-only video becomes visible

Expected: visibility rules are enforced by the backend and reflected by the UI.

- [ ] **Step 8: Verify blog embed integration explicitly**

Manual checklist:
- create or edit a blog post that includes `:::video{id="<real-video-uuid>"}:::`
- open the post detail page
- confirm the embed card shows real video data rather than placeholder text
- click the embed and confirm it opens the real video page

Expected: the existing `video` directive path is fully wired to real module data.

- [ ] **Step 9: Record any intentionally unverified follow-up items instead of overclaiming**

If the implementation cannot verify any of the following in one session, record them explicitly as follow-up work rather than silently assuming success:
- second-user follower-only visibility behavior
- cross-browser subtitle rendering
- recommendation-quality tuning beyond the deterministic MVP ordering
- analytics/playback metrics
- playlist/series support

Expected: completion notes distinguish between verified MVP behavior and deferred follow-up scope.

## Final Acceptance Checklist

The Video implementation should be considered complete for MVP only when all of the following are true:

1. There is a dedicated backend Video model and API surface.
2. There is a dedicated frontend `/video` route family.
3. Owners can create and manage local-upload and external-link videos.
4. Scheduling and visibility rules are enforced server-side.
5. Channel pages expose a working video tab.
6. `/channel/:slug/rss/video` returns valid channel-specific video RSS.
7. Blog `video` embed directives resolve to real video data.
8. `go build ./...`, `bun run type-check`, and `bun run build` all pass.
9. Any unverified edge cases are explicitly documented as follow-up items.

## Commit

```bash
git add docs/superpowers/plans/2026-05-14-video-implementation.md
git commit -m "docs: add video superpowers implementation plan"
```
