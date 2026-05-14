# Music Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Revalidate and complete the music module as an album-centric wiki music library with artist pages, revision history, governance states, lyric annotations, and stable album-based playback.

**Architecture:** Treat `plan/music_task_plan.md` as the authoritative scope and phase ledger, `plan/music_findings.md` as the product and governance decision record, `plan/music_progress.md` as implementation evidence plus prior regression history, and `plan/music_pages.md` as the UI route and interaction contract. Implementation should preserve the existing album-centered data model in Go, expose the required wiki APIs through Gin handlers, and align Vue routes/views/player behavior with the confirmed music information architecture.

**Tech Stack:** Vue 3, TypeScript, Vite, Vue Router, Pinia, Playwright, Go, Gin, GORM, SQLite/PostgreSQL

---

## Upstream Sources

- `plan/music_task_plan.md`
- `plan/music_findings.md`
- `plan/music_progress.md`
- `plan/music_pages.md`
- Style reference only: `docs/superpowers/plans/2026-05-14-blog-implementation.md`

## Normalized Current State

The upstream music planning set describes a feature that has already been designed and, according to the planning ledger, fully implemented across Phases A-F. The current repository also contains substantial supporting code: album, song, artist, revision, artist-alias, artist-merge, and lyric-annotation models on the backend; album, artist, history, discussion, admin review, upload, and player-related views on the frontend.

For execution purposes, do **not** assume the upstream claims are fully current. Instead, use this plan to revalidate and harden the album-centric wiki behavior end to end. The key risk is not missing architecture; it is drift between the original design, the current routes/types/API surface, and the actual browser/runtime behavior.

## Confirmed Scope To Preserve

- Album is the primary governance and presentation object.
- A single song is modeled as a single-track album, not as an isolated top-level workflow.
- Logged-in users can edit open/disputed music entries in wiki style.
- Every meaningful edit should be captured in revision history rather than overwriting the past.
- Artist pages are first-class wiki entities with aliases and history.
- Entry governance states center on `open`, `confirmed`, and `disputed`.
- Lyric annotations are line-based and tied to songs but surfaced inside the album listening/reading experience.
- Playback should operate on album queues and remain stable across route transitions.

## Verified Architecture Decisions From Upstream

- **Core object model:** `Album` is the main entry; `Song` is nested structure; `Artist` is related but also wiki-editable.
- **Lifecycle model:** new entries start open, may be confirmed by admins, and can return to disputed when contested.
- **Revision model:** rollback creates a new revision rather than deleting or overwriting history.
- **Alias model:** artist aliases participate in search and preserve alternate names.
- **Merge model:** duplicate artists merge into a chosen primary artist and redirect.
- **Listening model:** queue behavior should be album-based, not cross-album mixed playback.
- **Annotation model:** lyric notes are tied to `song_id` and `line_number`, with per-user authored commentary.

## Critical Implementation Files To Revalidate

### Frontend
- `web/src/router.ts`
- `web/src/types.ts`
- `web/src/composables/useApi.ts`
- `web/src/stores/player.ts`
- `web/src/views/music/HomeView.vue`
- `web/src/views/music/AlbumDetailView.vue`
- `web/src/views/music/EditAlbumView.vue`
- `web/src/views/music/UploadView.vue`
- `web/src/views/music/AlbumHistoryView.vue`
- `web/src/views/music/AlbumDiscussionView.vue`
- `web/src/views/music/ArtistDetailView.vue`
- `web/src/views/music/ArtistEditView.vue`
- `web/src/views/music/ArtistHistoryView.vue`
- `web/src/views/music/AddArtistView.vue`
- `web/src/views/music/AdminReviewView.vue`

### Backend
- `server/cmd/start_server/main.go`
- `server/internal/model/music.go`
- `server/internal/model/revision.go`
- `server/internal/service/revision_service.go`
- `server/internal/handlers/albums_handler.go`
- `server/internal/handlers/artists_handler.go`
- `server/internal/handlers/artist_wiki_handler.go`
- `server/internal/handlers/revision_handler.go`

### Test Targets To Add Or Reuse
- `tests/music-album-flow.spec.ts`
- `tests/music-artist-wiki.spec.ts`
- `tests/music-governance.spec.ts`
- `tests/music-player-regression.spec.ts`
- `server/internal/handlers/albums_handler_test.go`
- `server/internal/handlers/artist_wiki_handler_test.go`
- `server/internal/service/revision_service_test.go`

## Task 1: Revalidate the music baseline before changing behavior

**Files:**
- Inspect: `plan/music_task_plan.md`
- Inspect: `plan/music_findings.md`
- Inspect: `plan/music_progress.md`
- Inspect: `plan/music_pages.md`
- Test: `web/src/router.ts`
- Test: `server/cmd/start_server/main.go`

- [ ] **Step 1: Reconcile upstream completion claims against the current codebase**

Capture these working notes before editing any code:
```text
- which phases upstream says are complete
- which music routes/views/models already exist in the repo
- which behaviors still need proof in tests or browser checks
- any mismatches between the route contract in music_pages and the current router
```

Expected: a short discrepancy list that tells you whether you are implementing missing behavior or correcting drift.

- [ ] **Step 2: Run the frontend type-check**

Run:
```bash
cd /Users/fafa/Documents/projects/Atoman/.claude/worktrees/agent-a2e4df7d462a55426/web && bun run type-check
```

Expected: successful TypeScript verification with no music-specific type errors.

- [ ] **Step 3: Run the frontend production build**

Run:
```bash
cd /Users/fafa/Documents/projects/Atoman/.claude/worktrees/agent-a2e4df7d462a55426/web && bun run build
```

Expected: successful build confirming the current music views and router compile end to end.

- [ ] **Step 4: Run the backend build**

Run:
```bash
cd /Users/fafa/Documents/projects/Atoman/.claude/worktrees/agent-a2e4df7d462a55426/server && go build ./...
```

Expected: successful Go build confirming the model, migration, revision, and handler surface compiles.

- [ ] **Step 5: Verify the current route map against the page contract**

Compare the actual routes in `web/src/router.ts` against the desired contract from `plan/music_pages.md`:
```text
- /music
- /music/artists
- /music/artists/new
- /music/artists/:id
- /music/artists/:id/edit
- /music/artists/:id/history
- /music/artists/:id/discussion
- /music/albums/new
- /music/albums/:id
- /music/albums/:id/edit
- /music/albums/:id/history
- /music/albums/:id/discussion
- /music/albums/:id/proposals
- /music/albums/:id/proposals/new
- /music/admin or /music/admin/review
```

Expected: an explicit list of missing routes, route aliases, and naming drift such as `artistName` vs `artistId` or `/music/contribute` vs `/music/albums/new`.

## Task 2: Lock down the album-centric data and revision model

**Files:**
- Modify: `server/internal/model/music.go`
- Modify: `server/internal/model/revision.go`
- Modify: `server/internal/service/revision_service.go`
- Modify: `server/cmd/start_server/main.go`
- Test: `server/internal/service/revision_service_test.go`

- [ ] **Step 1: Write the failing backend test for album revision snapshots**

Create `server/internal/service/revision_service_test.go` with a test shaped like this:
```go
func TestAlbumRevisionSnapshotIncludesAlbumAndSongs(t *testing.T) {
	db := newTestDB(t)
	album := seedAlbumWithSongs(t, db)

	svc := NewRevisionService(db)
	rev, err := svc.CreateAlbumRevision(album.ID, testUserID(t), "fix track order")
	if err != nil {
		t.Fatalf("CreateAlbumRevision returned error: %v", err)
	}

	if rev.EntityType != "album" {
		t.Fatalf("expected entity type album, got %s", rev.EntityType)
	}
	if !strings.Contains(rev.SnapshotJSON, "songs") {
		t.Fatalf("expected snapshot to include songs payload")
	}
	if !strings.Contains(rev.SnapshotJSON, "title") {
		t.Fatalf("expected snapshot to include album metadata")
	}
}
```

Expected: the test either fails because the snapshot helper is incomplete or passes and documents the contract.

- [ ] **Step 2: Run the single backend test**

Run:
```bash
cd /Users/fafa/Documents/projects/Atoman/.claude/worktrees/agent-a2e4df7d462a55426/server && go test ./internal/service -run TestAlbumRevisionSnapshotIncludesAlbumAndSongs -v
```

Expected: FAIL if the revision snapshot contract is missing, otherwise PASS with the contract documented.

- [ ] **Step 3: Implement or tighten minimal revision snapshot logic**

Ensure the revision service preserves this shape for album revisions:
```go
type AlbumRevisionSnapshot struct {
	Album struct {
		ID          string `json:"id"`
		Title       string `json:"title"`
		AlbumType   string `json:"album_type"`
		EntryStatus string `json:"entry_status"`
	} `json:"album"`
	Songs []struct {
		ID          string `json:"id"`
		Title       string `json:"title"`
		TrackNumber int    `json:"track_number"`
		Lyrics      string `json:"lyrics"`
		AudioURL    string `json:"audio_url"`
	} `json:"songs"`
}
```

Expected: the revision payload is deterministic enough for history, diff, and rollback views.

- [ ] **Step 4: Re-run the targeted revision test**

Run:
```bash
cd /Users/fafa/Documents/projects/Atoman/.claude/worktrees/agent-a2e4df7d462a55426/server && go test ./internal/service -run TestAlbumRevisionSnapshotIncludesAlbumAndSongs -v
```

Expected: PASS.

- [ ] **Step 5: Add a rollback-as-new-revision test**

Extend `server/internal/service/revision_service_test.go` with:
```go
func TestAlbumRollbackCreatesNewRevision(t *testing.T) {
	db := newTestDB(t)
	album := seedAlbumWithSongs(t, db)
	svc := NewRevisionService(db)

	original, _ := svc.CreateAlbumRevision(album.ID, testUserID(t), "initial")
	mutateAlbumTitle(t, db, album.ID, "Renamed Album")
	_, _ = svc.CreateAlbumRevision(album.ID, testUserID(t), "rename album")

	reverted, err := svc.RevertAlbumToRevision(album.ID, original.ID, testUserID(t), "revert title")
	if err != nil {
		t.Fatalf("revert failed: %v", err)
	}
	if reverted.ID == original.ID {
		t.Fatalf("expected revert to create a new revision")
	}
}
```

Expected: the test proves rollback appends history instead of mutating it.

- [ ] **Step 6: Run the service test package**

Run:
```bash
cd /Users/fafa/Documents/projects/Atoman/.claude/worktrees/agent-a2e4df7d462a55426/server && go test ./internal/service -v
```

Expected: PASS for revision service coverage.

- [ ] **Step 7: Commit**

```bash
git add /Users/fafa/Documents/projects/Atoman/.claude/worktrees/agent-a2e4df7d462a55426/server/internal/model/music.go /Users/fafa/Documents/projects/Atoman/.claude/worktrees/agent-a2e4df7d462a55426/server/internal/model/revision.go /Users/fafa/Documents/projects/Atoman/.claude/worktrees/agent-a2e4df7d462a55426/server/internal/service/revision_service.go /Users/fafa/Documents/projects/Atoman/.claude/worktrees/agent-a2e4df7d462a55426/server/internal/service/revision_service_test.go /Users/fafa/Documents/projects/Atoman/.claude/worktrees/agent-a2e4df7d462a55426/server/cmd/start_server/main.go

git commit -m "test: lock music revision history contract"
```

## Task 3: Complete album wiki APIs for create, edit, history, and discussion

**Files:**
- Modify: `server/internal/handlers/albums_handler.go`
- Modify: `server/internal/handlers/revision_handler.go`
- Modify: `server/cmd/start_server/main.go`
- Test: `server/internal/handlers/albums_handler_test.go`

- [ ] **Step 1: Write the failing album create/edit handler test**

Create `server/internal/handlers/albums_handler_test.go` with a table-driven test that checks:
```go
func TestCreateAlbumCreatesSingleWhenOneSongProvided(t *testing.T) {
	router, db := newAlbumTestRouter(t)
	body := `{
	  "title": "Signal Lost",
	  "artist_ids": ["11111111-1111-1111-1111-111111111111"],
	  "songs": [{"title": "Signal Lost", "track_number": 1, "audio_url": "/uploads/signal.mp3"}]
	}`

	resp := performJSON(router, "POST", "/api/albums", body, authHeaderFor(t, db))
	assertStatus(t, resp.Code, http.StatusCreated)
	assertJSONContains(t, resp.Body.String(), `"album_type":"single"`)
}
```

Expected: FAIL if album type inference or response shape is missing.

- [ ] **Step 2: Run the targeted album handler test**

Run:
```bash
cd /Users/fafa/Documents/projects/Atoman/.claude/worktrees/agent-a2e4df7d462a55426/server && go test ./internal/handlers -run TestCreateAlbumCreatesSingleWhenOneSongProvided -v
```

Expected: FAIL first if the create flow is incomplete.

- [ ] **Step 3: Implement minimal album create/edit logic to satisfy the wiki contract**

Preserve these request/behavior rules in `albums_handler.go`:
```text
- create album with at least one song
- infer album_type=single when song count is 1
- infer album_type=ep when song count is 2-6 unless client overrides
- infer album_type=album when song count is 7+ unless client overrides
- require edit reason for update requests
- create a revision after successful create or update
```

Expected: API behavior matches the upstream product contract.

- [ ] **Step 4: Add a failing history endpoint test**

Extend `server/internal/handlers/albums_handler_test.go` with:
```go
func TestAlbumHistoryListsRevisions(t *testing.T) {
	router, db := newAlbumTestRouter(t)
	album := seedAlbumWithRevisionHistory(t, db)

	resp := performJSON(router, "GET", "/api/albums/"+album.ID.String()+"/revisions", "", authHeaderFor(t, db))
	assertStatus(t, resp.Code, http.StatusOK)
	assertJSONContains(t, resp.Body.String(), "revision")
}
```

Expected: FAIL if the revision route is absent or returns the wrong entity scope.

- [ ] **Step 5: Implement or fix history and discussion routes registration**

Ensure `main.go` registers the album wiki routes needed by the plan:
```go
music := api.Group("/albums")
{
	music.POST("", handlers.CreateAlbum)
	music.GET(":id", handlers.GetAlbum)
	music.PUT(":id", handlers.UpdateAlbum)
	music.GET(":id/revisions", handlers.ListAlbumRevisions)
	music.GET(":id/revisions/:revisionID", handlers.GetAlbumRevision)
	music.POST(":id/revert/:revisionID", handlers.RevertAlbumRevision)
	music.GET(":id/discussion", handlers.ListAlbumDiscussion)
	music.POST(":id/discussion", handlers.CreateAlbumDiscussionPost)
}
```

Expected: every route required by album detail/history/discussion views exists in one consistent namespace.

- [ ] **Step 6: Run the handler package**

Run:
```bash
cd /Users/fafa/Documents/projects/Atoman/.claude/worktrees/agent-a2e4df7d462a55426/server && go test ./internal/handlers -v
```

Expected: PASS for album creation, update, history, and discussion coverage.

- [ ] **Step 7: Commit**

```bash
git add /Users/fafa/Documents/projects/Atoman/.claude/worktrees/agent-a2e4df7d462a55426/server/internal/handlers/albums_handler.go /Users/fafa/Documents/projects/Atoman/.claude/worktrees/agent-a2e4df7d462a55426/server/internal/handlers/revision_handler.go /Users/fafa/Documents/projects/Atoman/.claude/worktrees/agent-a2e4df7d462a55426/server/internal/handlers/albums_handler_test.go /Users/fafa/Documents/projects/Atoman/.claude/worktrees/agent-a2e4df7d462a55426/server/cmd/start_server/main.go

git commit -m "feat: complete album wiki api flow"
```

## Task 4: Complete artist wiki APIs for aliases, history, and merge readiness

**Files:**
- Modify: `server/internal/handlers/artists_handler.go`
- Modify: `server/internal/handlers/artist_wiki_handler.go`
- Modify: `server/internal/model/music.go`
- Test: `server/internal/handlers/artist_wiki_handler_test.go`

- [ ] **Step 1: Write the failing alias CRUD test**

Create `server/internal/handlers/artist_wiki_handler_test.go` with:
```go
func TestArtistAliasCRUD(t *testing.T) {
	router, db := newArtistTestRouter(t)
	artist := seedArtist(t, db, "Björk")

	createResp := performJSON(router, "POST", "/api/artists/"+artist.ID.String()+"/aliases", `{"alias":"Bjork"}`, authHeaderFor(t, db))
	assertStatus(t, createResp.Code, http.StatusCreated)

	listResp := performJSON(router, "GET", "/api/artists/"+artist.ID.String()+"/aliases", "", authHeaderFor(t, db))
	assertStatus(t, listResp.Code, http.StatusOK)
	assertJSONContains(t, listResp.Body.String(), "Bjork")
}
```

Expected: FAIL if alias routes or payload handling are incomplete.

- [ ] **Step 2: Run the targeted alias test**

Run:
```bash
cd /Users/fafa/Documents/projects/Atoman/.claude/worktrees/agent-a2e4df7d462a55426/server && go test ./internal/handlers -run TestArtistAliasCRUD -v
```

Expected: FAIL first if alias CRUD is missing or misrouted.

- [ ] **Step 3: Implement minimal alias and artist revision support**

Preserve these behaviors in `artist_wiki_handler.go` and related helpers:
```text
- GET artist detail returns aliases and current entry_status
- POST/PUT artist edit creates artist revisions
- GET artist history returns wiki-style revision timeline
- alias create/delete operates on artist-scoped routes
- merged artist records can redirect to the primary artist detail route
```

Expected: the backend can support the artist detail, edit, and history views defined in `plan/music_pages.md`.

- [ ] **Step 4: Add a failing merged-artist redirect test**

Extend `server/internal/handlers/artist_wiki_handler_test.go` with:
```go
func TestMergedArtistRedirectsToPrimaryEntry(t *testing.T) {
	router, db := newArtistTestRouter(t)
	target, source := seedMergedArtistPair(t, db)

	resp := performJSON(router, "GET", "/api/artists/"+source.ID.String(), "", authHeaderFor(t, db))
	assertStatus(t, resp.Code, http.StatusOK)
	assertJSONContains(t, resp.Body.String(), target.ID.String())
	assertJSONContains(t, resp.Body.String(), "redirect")
}
```

Expected: FAIL if merged artists are silently returned as normal records instead of redirect-aware responses.

- [ ] **Step 5: Run the full artist handler coverage**

Run:
```bash
cd /Users/fafa/Documents/projects/Atoman/.claude/worktrees/agent-a2e4df7d462a55426/server && go test ./internal/handlers -run 'TestArtist' -v
```

Expected: PASS for artist wiki create/edit/history/alias/merge behavior.

- [ ] **Step 6: Commit**

```bash
git add /Users/fafa/Documents/projects/Atoman/.claude/worktrees/agent-a2e4df7d462a55426/server/internal/handlers/artists_handler.go /Users/fafa/Documents/projects/Atoman/.claude/worktrees/agent-a2e4df7d462a55426/server/internal/handlers/artist_wiki_handler.go /Users/fafa/Documents/projects/Atoman/.claude/worktrees/agent-a2e4df7d462a55426/server/internal/handlers/artist_wiki_handler_test.go /Users/fafa/Documents/projects/Atoman/.claude/worktrees/agent-a2e4df7d462a55426/server/internal/model/music.go

git commit -m "feat: complete artist wiki backend flow"
```

## Task 5: Complete governance state flow for open, confirmed, and disputed entries

**Files:**
- Modify: `server/internal/handlers/albums_handler.go`
- Modify: `server/internal/handlers/artist_wiki_handler.go`
- Modify: `web/src/types.ts`
- Modify: `web/src/composables/useApi.ts`
- Test: `tests/music-governance.spec.ts`

- [ ] **Step 1: Write the failing browser test for governance actions**

Create `tests/music-governance.spec.ts` with a flow like:
```ts
import { test, expect } from '@playwright/test'

test('admin can confirm and dispute an album entry', async ({ page }) => {
  await page.goto('/music/admin/review')
  await page.getByText('Open').click()
  await page.getByRole('link', { name: /signal lost/i }).click()
  await page.getByRole('button', { name: /confirm/i }).click()
  await expect(page.getByText(/confirmed|已确认/i)).toBeVisible()
  await page.getByRole('button', { name: /dispute|争议/i }).click()
  await expect(page.getByText(/disputed|争议中/i)).toBeVisible()
})
```

Expected: FAIL if the admin surface or status badge transitions are broken.

- [ ] **Step 2: Run the targeted governance browser test**

Run:
```bash
cd /Users/fafa/Documents/projects/Atoman/.claude/worktrees/agent-a2e4df7d462a55426/web && bun run test:e2e -- ../tests/music-governance.spec.ts
```

Expected: FAIL first until the admin state flow is wired correctly.

- [ ] **Step 3: Normalize frontend status typing and API endpoints**

Update `web/src/types.ts` so album and artist wiki states are explicit:
```ts
export type MusicEntryStatus = 'open' | 'confirmed' | 'disputed'
```

Then make sure `Album` and `Artist` use that type for `entry_status` rather than a generic string where possible.

Expected: the frontend can make state-dependent rendering decisions without stringly typed drift.

- [ ] **Step 4: Add or fix governance API helpers**

Extend `web/src/composables/useApi.ts` with helpers shaped like:
```ts
albumEntryStatus: (id: number | string) => `${apiUrl}/albums/${id}/entry-status`,
artistEntryStatus: (id: number | string) => `${apiUrl}/artists/${id}/entry-status`,
adminMusicReview: `${apiUrl}/music/admin/review`,
```

Expected: views stop hand-building status endpoints and use one consistent API surface.

- [ ] **Step 5: Implement minimal admin transition handlers**

Preserve this transition contract:
```text
- open -> confirmed by admin
- confirmed -> disputed by admin
- disputed -> confirmed by admin
- non-admin users cannot trigger transitions directly
```

Expected: governance flow matches the simplified upstream operating model.

- [ ] **Step 6: Re-run the governance browser test**

Run:
```bash
cd /Users/fafa/Documents/projects/Atoman/.claude/worktrees/agent-a2e4df7d462a55426/web && bun run test:e2e -- ../tests/music-governance.spec.ts
```

Expected: PASS.

- [ ] **Step 7: Commit**

```bash
git add /Users/fafa/Documents/projects/Atoman/.claude/worktrees/agent-a2e4df7d462a55426/server/internal/handlers/albums_handler.go /Users/fafa/Documents/projects/Atoman/.claude/worktrees/agent-a2e4df7d462a55426/server/internal/handlers/artist_wiki_handler.go /Users/fafa/Documents/projects/Atoman/.claude/worktrees/agent-a2e4df7d462a55426/web/src/types.ts /Users/fafa/Documents/projects/Atoman/.claude/worktrees/agent-a2e4df7d462a55426/web/src/composables/useApi.ts /Users/fafa/Documents/projects/Atoman/.claude/worktrees/agent-a2e4df7d462a55426/tests/music-governance.spec.ts

git commit -m "feat: wire music entry governance states"
```

## Task 6: Align router and views with the album/artist page contract

**Files:**
- Modify: `web/src/router.ts`
- Modify: `web/src/views/music/HomeView.vue`
- Modify: `web/src/views/music/AlbumDetailView.vue`
- Modify: `web/src/views/music/EditAlbumView.vue`
- Modify: `web/src/views/music/UploadView.vue`
- Modify: `web/src/views/music/ArtistDetailView.vue`
- Modify: `web/src/views/music/ArtistEditView.vue`
- Modify: `web/src/views/music/ArtistHistoryView.vue`
- Modify: `web/src/views/music/AlbumHistoryView.vue`
- Modify: `web/src/views/music/AlbumDiscussionView.vue`
- Modify: `web/src/views/music/AdminReviewView.vue`
- Test: `tests/music-album-flow.spec.ts`
- Test: `tests/music-artist-wiki.spec.ts`

- [ ] **Step 1: Write the failing album flow browser test**

Create `tests/music-album-flow.spec.ts` with:
```ts
import { test, expect } from '@playwright/test'

test('user can create a single-track album and open history/discussion views', async ({ page }) => {
  await page.goto('/music/albums/new')
  await page.getByLabel(/album title|专辑名/i).fill('Signal Lost')
  await page.getByLabel(/song title|歌名/i).fill('Signal Lost')
  await page.getByRole('button', { name: /save|create|创建/i }).click()
  await expect(page).toHaveURL(/\/music\/albums\//)
  await expect(page.getByText(/single/i)).toBeVisible()
  await page.getByRole('link', { name: /history|历史/i }).click()
  await expect(page).toHaveURL(/\/history$/)
  await page.getByRole('link', { name: /discussion|讨论/i }).click()
  await expect(page).toHaveURL(/\/discussion$/)
})
```

Expected: FAIL first if route aliases or view wiring are incomplete.

- [ ] **Step 2: Write the failing artist wiki browser test**

Create `tests/music-artist-wiki.spec.ts` with:
```ts
import { test, expect } from '@playwright/test'

test('user can open artist detail, edit, and history views', async ({ page }) => {
  await page.goto('/music/artists/new')
  await page.getByLabel(/artist name|艺术家名/i).fill('Björk')
  await page.getByRole('button', { name: /save|create|创建/i }).click()
  await expect(page).toHaveURL(/\/music\/artists\//)
  await page.getByRole('link', { name: /edit|编辑/i }).click()
  await expect(page).toHaveURL(/\/edit$/)
  await page.getByRole('link', { name: /history|历史/i }).click()
  await expect(page).toHaveURL(/\/history$/)
})
```

Expected: FAIL first if artist route naming is inconsistent.

- [ ] **Step 3: Run the targeted browser tests**

Run:
```bash
cd /Users/fafa/Documents/projects/Atoman/.claude/worktrees/agent-a2e4df7d462a55426/web && bun run test:e2e -- ../tests/music-album-flow.spec.ts ../tests/music-artist-wiki.spec.ts
```

Expected: FAIL until route and view parity is restored.

- [ ] **Step 4: Normalize the music route structure**

Update `web/src/router.ts` toward this route set:
```ts
{ path: '/music', component: () => import('@/views/music/HomeView.vue') },
{ path: '/music/artists', component: () => import('@/views/music/HomeView.vue') },
{ path: '/music/artists/new', component: () => import('@/views/music/AddArtistView.vue'), meta: { requiresAuth: true } },
{ path: '/music/artists/:artistId', component: () => import('@/views/music/ArtistDetailView.vue') },
{ path: '/music/artists/:artistId/edit', component: () => import('@/views/music/ArtistEditView.vue'), meta: { requiresAuth: true } },
{ path: '/music/artists/:artistId/history', component: () => import('@/views/music/ArtistHistoryView.vue') },
{ path: '/music/albums/new', component: () => import('@/views/music/UploadView.vue'), meta: { requiresAuth: true } },
{ path: '/music/albums/:albumId', component: () => import('@/views/music/AlbumDetailView.vue') },
{ path: '/music/albums/:albumId/edit', component: () => import('@/views/music/EditAlbumView.vue'), meta: { requiresAuth: true } },
{ path: '/music/albums/:albumId/history', component: () => import('@/views/music/AlbumHistoryView.vue') },
{ path: '/music/albums/:albumId/discussion', component: () => import('@/views/music/AlbumDiscussionView.vue') },
{ path: '/music/admin/review', component: () => import('@/views/music/AdminReviewView.vue'), meta: { requiresAuth: true, requiresAdmin: true } },
```

Keep legacy aliases only if you need them for compatibility, and make the album/artist canonical routes match the design contract.

Expected: route semantics become predictable across the views and tests.

- [ ] **Step 5: Align the views with state-aware page behavior**

Verify and fix these UI rules from `plan/music_pages.md`:
```text
- album detail shows title, artists, album type, release date, last editor, entry status badge, actions
- album edit requires edit reason on update flow
- history view renders a revision timeline with revert metadata
- discussion view renders thread list and posting form
- artist detail shows aliases, biography fields, works list, and actions based on entry_status
- admin review view exposes open and disputed filters
```

Expected: the page surface matches the intended music wiki information architecture.

- [ ] **Step 6: Re-run the browser tests**

Run:
```bash
cd /Users/fafa/Documents/projects/Atoman/.claude/worktrees/agent-a2e4df7d462a55426/web && bun run test:e2e -- ../tests/music-album-flow.spec.ts ../tests/music-artist-wiki.spec.ts
```

Expected: PASS.

- [ ] **Step 7: Commit**

```bash
git add /Users/fafa/Documents/projects/Atoman/.claude/worktrees/agent-a2e4df7d462a55426/web/src/router.ts /Users/fafa/Documents/projects/Atoman/.claude/worktrees/agent-a2e4df7d462a55426/web/src/views/music/HomeView.vue /Users/fafa/Documents/projects/Atoman/.claude/worktrees/agent-a2e4df7d462a55426/web/src/views/music/AlbumDetailView.vue /Users/fafa/Documents/projects/Atoman/.claude/worktrees/agent-a2e4df7d462a55426/web/src/views/music/EditAlbumView.vue /Users/fafa/Documents/projects/Atoman/.claude/worktrees/agent-a2e4df7d462a55426/web/src/views/music/UploadView.vue /Users/fafa/Documents/projects/Atoman/.claude/worktrees/agent-a2e4df7d462a55426/web/src/views/music/ArtistDetailView.vue /Users/fafa/Documents/projects/Atoman/.claude/worktrees/agent-a2e4df7d462a55426/web/src/views/music/ArtistEditView.vue /Users/fafa/Documents/projects/Atoman/.claude/worktrees/agent-a2e4df7d462a55426/web/src/views/music/ArtistHistoryView.vue /Users/fafa/Documents/projects/Atoman/.claude/worktrees/agent-a2e4df7d462a55426/web/src/views/music/AlbumHistoryView.vue /Users/fafa/Documents/projects/Atoman/.claude/worktrees/agent-a2e4df7d462a55426/web/src/views/music/AlbumDiscussionView.vue /Users/fafa/Documents/projects/Atoman/.claude/worktrees/agent-a2e4df7d462a55426/web/src/views/music/AdminReviewView.vue /Users/fafa/Documents/projects/Atoman/.claude/worktrees/agent-a2e4df7d462a55426/tests/music-album-flow.spec.ts /Users/fafa/Documents/projects/Atoman/.claude/worktrees/agent-a2e4df7d462a55426/tests/music-artist-wiki.spec.ts

git commit -m "feat: align music routes and wiki views"
```

## Task 7: Complete lyric annotation flow inside the listening experience

**Files:**
- Modify: `server/internal/model/music.go`
- Modify: `server/internal/handlers/albums_handler.go`
- Modify: `web/src/types.ts`
- Modify: `web/src/views/music/AlbumDetailView.vue`
- Test: `tests/music-album-flow.spec.ts`
- Test: `server/internal/handlers/albums_handler_test.go`

- [ ] **Step 1: Write the failing backend test for lyric annotations by line number**

Add to `server/internal/handlers/albums_handler_test.go`:
```go
func TestSongLyricAnnotationsAreGroupedByLine(t *testing.T) {
	router, db := newAlbumTestRouter(t)
	song := seedSongWithLyrics(t, db)

	performJSON(router, "POST", "/api/songs/"+song.ID.String()+"/annotations", `{"line_number":2,"content":"This line references isolation."}`, authHeaderFor(t, db))
	resp := performJSON(router, "GET", "/api/songs/"+song.ID.String()+"/annotations", "", authHeaderFor(t, db))

	assertStatus(t, resp.Code, http.StatusOK)
	assertJSONContains(t, resp.Body.String(), `"line_number":2`)
}
```

Expected: FAIL if song annotation routes are missing or unscoped.

- [ ] **Step 2: Run the targeted annotation test**

Run:
```bash
cd /Users/fafa/Documents/projects/Atoman/.claude/worktrees/agent-a2e4df7d462a55426/server && go test ./internal/handlers -run TestSongLyricAnnotationsAreGroupedByLine -v
```

Expected: FAIL first if the lyric annotation flow is incomplete.

- [ ] **Step 3: Implement or fix annotation endpoints and response typing**

Preserve these API/UI rules:
```text
- annotations are keyed by song and line number
- detail view can fetch all annotations for the current song
- entries include author username and created time
- multiple users can annotate the same lyric line without overwriting one another
```

Expected: the album detail page has the data needed for Genius-style line annotation presentation.

- [ ] **Step 4: Extend the browser test for annotation interaction**

Append to `tests/music-album-flow.spec.ts` a scenario that verifies:
```text
- open album detail
- expand lyrics for a track
- click a lyric line with an annotation affordance
- submit a line note
- see the username and annotation content appear in the side panel or inline note list
```

Expected: the annotation UX is covered by an observable browser assertion rather than assumption.

- [ ] **Step 5: Run the targeted backend and browser tests**

Run:
```bash
cd /Users/fafa/Documents/projects/Atoman/.claude/worktrees/agent-a2e4df7d462a55426/server && go test ./internal/handlers -run TestSongLyricAnnotationsAreGroupedByLine -v
cd /Users/fafa/Documents/projects/Atoman/.claude/worktrees/agent-a2e4df7d462a55426/web && bun run test:e2e -- ../tests/music-album-flow.spec.ts
```

Expected: PASS.

- [ ] **Step 6: Commit**

```bash
git add /Users/fafa/Documents/projects/Atoman/.claude/worktrees/agent-a2e4df7d462a55426/server/internal/model/music.go /Users/fafa/Documents/projects/Atoman/.claude/worktrees/agent-a2e4df7d462a55426/server/internal/handlers/albums_handler.go /Users/fafa/Documents/projects/Atoman/.claude/worktrees/agent-a2e4df7d462a55426/server/internal/handlers/albums_handler_test.go /Users/fafa/Documents/projects/Atoman/.claude/worktrees/agent-a2e4df7d462a55426/web/src/types.ts /Users/fafa/Documents/projects/Atoman/.claude/worktrees/agent-a2e4df7d462a55426/web/src/views/music/AlbumDetailView.vue /Users/fafa/Documents/projects/Atoman/.claude/worktrees/agent-a2e4df7d462a55426/tests/music-album-flow.spec.ts

git commit -m "feat: add line-based lyric annotations"
```

## Task 8: Revalidate album-based playback and route-transition stability

**Files:**
- Modify: `web/src/stores/player.ts`
- Modify: `web/src/views/music/HomeView.vue`
- Modify: `web/src/views/music/AlbumDetailView.vue`
- Test: `tests/music-player-regression.spec.ts`

- [ ] **Step 1: Write the failing playback regression test**

Create `tests/music-player-regression.spec.ts` with:
```ts
import { test, expect } from '@playwright/test'

test('playback continues when navigating from music list to album detail', async ({ page }) => {
  await page.goto('/music')
  await page.getByRole('button', { name: /play/i }).first().click()
  await expect(page.getByTestId('global-player')).toContainText(/playing|pause/i)
  await page.getByRole('link').filter({ hasText: /album|专辑/i }).first().click()
  await expect(page).toHaveURL(/\/music\/albums\//)
  await expect(page.getByTestId('global-player')).toContainText(/playing|pause/i)
})
```

Expected: FAIL if route navigation still resets playback state.

- [ ] **Step 2: Run the targeted playback browser test**

Run:
```bash
cd /Users/fafa/Documents/projects/Atoman/.claude/worktrees/agent-a2e4df7d462a55426/web && bun run test:e2e -- ../tests/music-player-regression.spec.ts
```

Expected: FAIL first if playback state is still rehydrated incorrectly.

- [ ] **Step 3: Preserve one-time restoration and album-queue behavior in the player store**

Keep or reintroduce these invariants in `web/src/stores/player.ts`:
```text
- playback restoration from localStorage only happens once per app boot
- existing in-memory playback state wins over stale persisted state during route changes
- playAlbum(queue, startIndex) or equivalent album queue method remains the canonical playback entrypoint
- next/previous navigation stays inside the active album queue
```

Expected: playback remains stable while users navigate between list/detail/history pages.

- [ ] **Step 4: Re-run the targeted playback test**

Run:
```bash
cd /Users/fafa/Documents/projects/Atoman/.claude/worktrees/agent-a2e4df7d462a55426/web && bun run test:e2e -- ../tests/music-player-regression.spec.ts
```

Expected: PASS.

- [ ] **Step 5: Run a focused browser sanity pass by hand**

Manually verify these checkpoints:
```text
- /music list play button starts playback
- entering album detail does not pause audio
- returning to /music keeps the current track and queue
- refreshing the page restores playback only when no newer in-memory state exists
```

Expected: manual validation matches the regression history recorded in `plan/music_progress.md`.

- [ ] **Step 6: Commit**

```bash
git add /Users/fafa/Documents/projects/Atoman/.claude/worktrees/agent-a2e4df7d462a55426/web/src/stores/player.ts /Users/fafa/Documents/projects/Atoman/.claude/worktrees/agent-a2e4df7d462a55426/web/src/views/music/HomeView.vue /Users/fafa/Documents/projects/Atoman/.claude/worktrees/agent-a2e4df7d462a55426/web/src/views/music/AlbumDetailView.vue /Users/fafa/Documents/projects/Atoman/.claude/worktrees/agent-a2e4df7d462a55426/tests/music-player-regression.spec.ts

git commit -m "fix: stabilize music playback across route changes"
```

## Task 9: Final end-to-end verification and browser checklist

**Files:**
- Test: `web/src/views/music/**`
- Test: `server/internal/**/music*`
- Test: `tests/music-album-flow.spec.ts`
- Test: `tests/music-artist-wiki.spec.ts`
- Test: `tests/music-governance.spec.ts`
- Test: `tests/music-player-regression.spec.ts`

- [ ] **Step 1: Run backend tests for music-related packages**

Run:
```bash
cd /Users/fafa/Documents/projects/Atoman/.claude/worktrees/agent-a2e4df7d462a55426/server && go test ./internal/service ./internal/handlers -v
```

Expected: PASS for revision, album, artist, and annotation coverage.

- [ ] **Step 2: Run frontend static verification again**

Run:
```bash
cd /Users/fafa/Documents/projects/Atoman/.claude/worktrees/agent-a2e4df7d462a55426/web && bun run type-check
cd /Users/fafa/Documents/projects/Atoman/.claude/worktrees/agent-a2e4df7d462a55426/web && bun run build
```

Expected: PASS.

- [ ] **Step 3: Run the music browser suite**

Run:
```bash
cd /Users/fafa/Documents/projects/Atoman/.claude/worktrees/agent-a2e4df7d462a55426/web && bun run test:e2e -- ../tests/music-album-flow.spec.ts ../tests/music-artist-wiki.spec.ts ../tests/music-governance.spec.ts ../tests/music-player-regression.spec.ts
```

Expected: PASS.

- [ ] **Step 4: Perform manual browser checks that automated tests do not fully cover**

Verify all of the following in a real browser session:
```text
- album list shows filters and disputed badge behavior where designed
- album detail shows artist link, type badge, last editor, and action buttons appropriate to status
- confirmed entries surface a suggestion/proposal path instead of direct editing if that flow is implemented
- artist detail shows aliases and linked works
- admin review page can filter open vs disputed items without runtime errors
- lyric annotation UI shows multiple annotations on the same line without replacing prior notes
- playback queue remains album-scoped rather than mixing unrelated tracks
```

Expected: the remaining user-facing wiki/listening contract is verified beyond build/test status.

- [ ] **Step 5: Record any remaining gaps explicitly before claiming completion**

If any of these are still partial, write them down as follow-up issues instead of silently assuming completeness:
```text
- proposal flow for confirmed albums
- artist discussion page parity
- true diff visualization between revisions
- richer lyric-side-panel UI
```

Expected: no overclaiming; either the feature is complete or the remaining gaps are named.

- [ ] **Step 6: Commit**

```bash
git add /Users/fafa/Documents/projects/Atoman/.claude/worktrees/agent-a2e4df7d462a55426/tests/music-album-flow.spec.ts /Users/fafa/Documents/projects/Atoman/.claude/worktrees/agent-a2e4df7d462a55426/tests/music-artist-wiki.spec.ts /Users/fafa/Documents/projects/Atoman/.claude/worktrees/agent-a2e4df7d462a55426/tests/music-governance.spec.ts /Users/fafa/Documents/projects/Atoman/.claude/worktrees/agent-a2e4df7d462a55426/tests/music-player-regression.spec.ts

git commit -m "test: verify music wiki and playback flows"
```

## Verification

After execution, verify all of the following:

1. `plan/music_task_plan.md`, `plan/music_findings.md`, `plan/music_progress.md`, and `plan/music_pages.md` no longer need to be reopened to understand the implementation order.
2. `cd /Users/fafa/Documents/projects/Atoman/.claude/worktrees/agent-a2e4df7d462a55426/web && bun run type-check` passes.
3. `cd /Users/fafa/Documents/projects/Atoman/.claude/worktrees/agent-a2e4df7d462a55426/web && bun run build` passes.
4. `cd /Users/fafa/Documents/projects/Atoman/.claude/worktrees/agent-a2e4df7d462a55426/server && go build ./...` passes.
5. `cd /Users/fafa/Documents/projects/Atoman/.claude/worktrees/agent-a2e4df7d462a55426/server && go test ./internal/service ./internal/handlers -v` passes.
6. `cd /Users/fafa/Documents/projects/Atoman/.claude/worktrees/agent-a2e4df7d462a55426/web && bun run test:e2e -- ../tests/music-album-flow.spec.ts ../tests/music-artist-wiki.spec.ts ../tests/music-governance.spec.ts ../tests/music-player-regression.spec.ts` passes.
7. Browser validation covers album creation, artist editing/history, governance transitions, lyric annotation interaction, admin review filtering, and the playback regression path from list to detail.
8. Any incomplete proposal-flow or diff-visualization work is written down explicitly instead of implied as done.

## Commit

```bash
git add /Users/fafa/Documents/projects/Atoman/.claude/worktrees/agent-a2e4df7d462a55426/docs/superpowers/plans/2026-05-14-music-implementation.md
git commit -m "docs: add music superpowers implementation plan"
```
