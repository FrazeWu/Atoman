# Debate Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Evolve the existing debate feature from a flat topic-plus-quoted-argument flow into the planned structured debate system with positions, argument tree semantics, evidence, conclusions, and moderation.

**Architecture:** Treat `plan/debate_task_plan.md` as the product scope contract, `plan/debate_findings.md` as the domain-model and interaction decision record, and `plan/debate_progress.md` as the implementation-order recommendation. Build on the existing `Debate` / `Argument` backend and Vue debate views incrementally: first normalize the domain model around topic-position-argument structure, then layer evidence, summaries, moderation, and richer verification.

**Tech Stack:** Vue 3, TypeScript, Pinia, Vue Router, Go, Gin, GORM, SQLite/PostgreSQL, Playwright

---

## Upstream Sources

- `plan/debate_task_plan.md`
- `plan/debate_findings.md`
- `plan/debate_progress.md`
- Reference style only: `docs/superpowers/plans/2026-05-14-blog-implementation.md`

## Normalized Current State

The planning documents describe a structured debate system centered on `DebateTopic`, `DebatePosition`, `DebateArgument`, `DebateEvidence`, voting, conclusions, and moderation. The current codebase already ships a partially implemented debate module, but it does not yet match that target architecture.

What already exists in code:
- backend topic CRUD under `server/internal/handlers/debate_handler.go`
- backend argument CRUD, argument-to-argument references, debate references, voting, and conclude/reopen flows
- backend current models in `server/internal/model/debate.go`
- frontend route entries in `web/src/router.ts`
- frontend store in `web/src/stores/debate.ts`
- frontend views in `web/src/views/debate/DebateHomeView.vue` and `web/src/views/debate/DebateTopicView.vue`
- frontend argument card component in `web/src/components/debate/ArgumentNode.vue`

What is missing relative to the plan:
- no explicit `DebatePosition` model or UI
- no explicit `DebateEvidence` model or UI
- current `parent_id` is treated as a quoted argument, not a true structured argument tree contract
- no dedicated moderation action model or moderation UI
- no conclusion history / iterative summary model
- no homepage sections for hot/latest/controversial/concluded topics
- no automated frontend browser coverage in this worktree yet

## Confirmed Scope To Preserve

- Debate is not a forum clone; it is a structured argument system.
- Core entities remain topic, position, argument, evidence, vote, conclusion, moderation.
- Voting expresses stance/support and must not be treated as proof of truth.
- Summaries and conclusions are iterative and should preserve evolution.
- Governance must support low-quality-content control, folding, locking, pinning, and dispute handling.
- The implementation order should still start with topic detail and multi-position structure, then argument flow, then voting/summary, then evidence/governance enhancements.

## Implementation Strategy

1. Reconcile the current shipped debate code against the planning contract.
2. Introduce missing backend models and APIs for positions, evidence, moderation, and iterative conclusions.
3. Refactor the frontend detail page from a flat argument list into a position-aware structured debate workspace.
4. Add verification coverage at three layers:
   - Go build and targeted handler/model tests
   - Vue type-check and production build
   - Playwright browser validation for the main debate flows

## Critical Implementation Files To Modify Or Add

### Backend
- Modify: `server/internal/model/debate.go`
- Modify: `server/internal/handlers/debate_handler.go`
- Modify: `server/cmd/start_server/main.go`
- Add: `server/internal/handlers/debate_handler_test.go`
- Add: `server/internal/model/debate_test.go`

### Frontend
- Modify: `web/src/types.ts`
- Modify: `web/src/stores/debate.ts`
- Modify: `web/src/router.ts`
- Modify: `web/src/views/debate/DebateHomeView.vue`
- Modify: `web/src/views/debate/DebateTopicView.vue`
- Modify: `web/src/components/debate/ArgumentNode.vue`
- Add: `web/src/components/debate/PositionColumn.vue`
- Add: `web/src/components/debate/EvidenceList.vue`
- Add: `web/src/components/debate/ConclusionTimeline.vue`
- Add: `web/src/components/debate/ModerationPanel.vue`

### Browser Verification
- Add: `web/playwright.config.ts`
- Add: `web/tests/debate.spec.ts`

## Task 1: Reconcile the current debate implementation with the planning contract

**Files:**
- Inspect: `plan/debate_task_plan.md`
- Inspect: `plan/debate_findings.md`
- Inspect: `plan/debate_progress.md`
- Inspect: `server/internal/model/debate.go`
- Inspect: `server/internal/handlers/debate_handler.go`
- Inspect: `web/src/types.ts`
- Inspect: `web/src/stores/debate.ts`
- Inspect: `web/src/views/debate/DebateHomeView.vue`
- Inspect: `web/src/views/debate/DebateTopicView.vue`
- Inspect: `web/src/components/debate/ArgumentNode.vue`

- [ ] **Step 1: Write a gap checklist in working notes**

Use this exact checklist:
```text
Current code supports:
- topic CRUD
- argument CRUD
- argument voting
- conclude/reopen
- argument references
- debate references

Current code does not yet support:
- explicit positions
- explicit evidence entities
- iterative conclusion records
- moderation action records
- homepage sectioned ranking
- browser automation coverage
```

Expected: a one-page reconciliation note that prevents accidental reimplementation of existing features.

- [ ] **Step 2: Run the frontend type-check baseline**

Run:
```bash
cd /Users/fafa/Documents/projects/Atoman/.claude/worktrees/agent-a9dac96e4ba6d6edb/web && bun run type-check
```

Expected: PASS. If it fails, record the existing failure before starting debate edits.

- [ ] **Step 3: Run the frontend build baseline**

Run:
```bash
cd /Users/fafa/Documents/projects/Atoman/.claude/worktrees/agent-a9dac96e4ba6d6edb/web && bun run build
```

Expected: PASS. If it fails, record the existing failure before starting debate edits.

- [ ] **Step 4: Run the backend build baseline**

Run:
```bash
cd /Users/fafa/Documents/projects/Atoman/.claude/worktrees/agent-a9dac96e4ba6d6edb/server && go build ./...
```

Expected: PASS. If it fails, record the existing failure before starting debate edits.

- [ ] **Step 5: Commit the reconciliation checkpoint**

```bash
git add plan/debate_task_plan.md plan/debate_findings.md plan/debate_progress.md

git commit -m "docs: capture debate implementation baseline"
```

Expected: a clean checkpoint before schema and behavior changes begin.

## Task 2: Add explicit structured debate domain models on the backend

**Files:**
- Modify: `server/internal/model/debate.go`
- Modify: `server/cmd/start_server/main.go`
- Test: `server/internal/model/debate_test.go`

- [ ] **Step 1: Write the failing model test for the structured entities**

Create `server/internal/model/debate_test.go` with this test skeleton:
```go
package model

import "testing"

func TestDebateStructuredEntitiesExposeExpectedFields(t *testing.T) {
	position := DebatePosition{}
	evidence := DebateEvidence{}
	conclusion := DebateConclusion{}
	action := DebateModerationAction{}

	if position.Title != "" {
		t.Fatal("expected zero-value position title")
	}
	if evidence.SourceType != "" {
		t.Fatal("expected zero-value evidence source type")
	}
	if conclusion.Summary != "" {
		t.Fatal("expected zero-value conclusion summary")
	}
	if action.ActionType != "" {
		t.Fatal("expected zero-value moderation action type")
	}
}
```

Expected: FAIL because the structs do not exist yet.

- [ ] **Step 2: Run the model test to verify it fails**

Run:
```bash
cd /Users/fafa/Documents/projects/Atoman/.claude/worktrees/agent-a9dac96e4ba6d6edb/server && go test ./internal/model -run TestDebateStructuredEntitiesExposeExpectedFields -v
```

Expected: FAIL with undefined `DebatePosition`, `DebateEvidence`, `DebateConclusion`, and `DebateModerationAction`.

- [ ] **Step 3: Add the missing domain structs with conservative fields**

Extend `server/internal/model/debate.go` with these structs and enums:
```go
type DebateTopicStatus string

const (
	DebateTopicStatusOpen     DebateTopicStatus = "open"
	DebateTopicStatusLocked   DebateTopicStatus = "locked"
	DebateTopicStatusArchived DebateTopicStatus = "archived"
	DebateTopicStatusClosed   DebateTopicStatus = "concluded"
)

type DebatePosition struct {
	Base
	DebateID     uuid.UUID `json:"debate_id" gorm:"type:uuid;not null;index"`
	Title        string    `json:"title" gorm:"not null"`
	Description  string    `json:"description" gorm:"type:text"`
	SortOrder    int       `json:"sort_order" gorm:"default:0"`
	CreatedBy    uuid.UUID `json:"created_by" gorm:"type:uuid;not null;index"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

func (DebatePosition) TableName() string { return "debate_positions" }

type DebateEvidence struct {
	Base
	ArgumentID       uuid.UUID `json:"argument_id" gorm:"type:uuid;not null;index"`
	SourceType       string    `json:"source_type" gorm:"type:varchar(32);not null"`
	SourceRef        string    `json:"source_ref" gorm:"type:text"`
	Title            string    `json:"title" gorm:"not null"`
	Excerpt          string    `json:"excerpt" gorm:"type:text"`
	CredibilityNote  string    `json:"credibility_note" gorm:"type:text"`
	CreatedBy        uuid.UUID `json:"created_by" gorm:"type:uuid;not null;index"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

func (DebateEvidence) TableName() string { return "debate_evidences" }

type DebateConclusion struct {
	Base
	DebateID        uuid.UUID `json:"debate_id" gorm:"type:uuid;not null;index"`
	Summary         string    `json:"summary" gorm:"type:text;not null"`
	OpenQuestions   string    `json:"open_questions" gorm:"type:text"`
	VersionNumber   int       `json:"version_number" gorm:"default:1"`
	CreatedBy       uuid.UUID `json:"created_by" gorm:"type:uuid;not null;index"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

func (DebateConclusion) TableName() string { return "debate_conclusions" }

type DebateModerationAction struct {
	Base
	DebateID        *uuid.UUID `json:"debate_id,omitempty" gorm:"type:uuid;index"`
	ArgumentID      *uuid.UUID `json:"argument_id,omitempty" gorm:"type:uuid;index"`
	ActionType      string     `json:"action_type" gorm:"type:varchar(32);not null"`
	Reason          string     `json:"reason" gorm:"type:text"`
	CreatedBy       uuid.UUID  `json:"created_by" gorm:"type:uuid;not null;index"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

func (DebateModerationAction) TableName() string { return "debate_moderation_actions" }
```

Also extend `Argument` minimally with:
```go
PositionID *uuid.UUID `json:"position_id" gorm:"type:uuid;index"`
Depth      int        `json:"depth" gorm:"default:0"`
IsFolded   bool       `json:"is_folded" gorm:"default:false"`
```

Expected: the backend model layer now matches the planned structured entities without removing existing working behavior.

- [ ] **Step 4: Register the new models in auto-migration**

Add these entries to `db.AutoMigrate(...)` inside `server/cmd/start_server/main.go` immediately after the existing debate models:
```go
&model.DebatePosition{},
&model.DebateEvidence{},
&model.DebateConclusion{},
&model.DebateModerationAction{},
```

Expected: local dev startup can materialize the new schema automatically.

- [ ] **Step 5: Run the model test and backend build**

Run:
```bash
cd /Users/fafa/Documents/projects/Atoman/.claude/worktrees/agent-a9dac96e4ba6d6edb/server && go test ./internal/model -run TestDebateStructuredEntitiesExposeExpectedFields -v && go build ./...
```

Expected: PASS.

- [ ] **Step 6: Commit the schema foundation**

```bash
git add /Users/fafa/Documents/projects/Atoman/.claude/worktrees/agent-a9dac96e4ba6d6edb/server/internal/model/debate.go \
  /Users/fafa/Documents/projects/Atoman/.claude/worktrees/agent-a9dac96e4ba6d6edb/server/internal/model/debate_test.go \
  /Users/fafa/Documents/projects/Atoman/.claude/worktrees/agent-a9dac96e4ba6d6edb/server/cmd/start_server/main.go

git commit -m "feat: add structured debate domain models"
```

Expected: the planned entities exist before API refactors begin.

## Task 3: Add backend APIs for positions, evidences, conclusions, and moderation

**Files:**
- Modify: `server/internal/handlers/debate_handler.go`
- Test: `server/internal/handlers/debate_handler_test.go`

- [ ] **Step 1: Write the failing handler test for position creation**

Create `server/internal/handlers/debate_handler_test.go` with this initial test:
```go
package handlers

import "testing"

func TestCreateDebatePositionHandlerExists(t *testing.T) {
	if CreateDebatePosition == nil {
		t.Fatal("expected CreateDebatePosition handler factory to exist")
	}
}
```

Expected: FAIL because the handler factory does not exist yet.

- [ ] **Step 2: Run the handler test to verify it fails**

Run:
```bash
cd /Users/fafa/Documents/projects/Atoman/.claude/worktrees/agent-a9dac96e4ba6d6edb/server && go test ./internal/handlers -run TestCreateDebatePositionHandlerExists -v
```

Expected: FAIL with undefined `CreateDebatePosition`.

- [ ] **Step 3: Add route entries for the missing structured resources**

Inside `SetupDebateRoutes(...)`, add these protected routes:
```go
protected.POST("/topics/:id/positions", CreateDebatePosition(db))
protected.GET("/topics/:id/positions", GetDebatePositions(db))
protected.POST("/arguments/:id/evidences", CreateDebateEvidence(db))
protected.GET("/arguments/:id/evidences", GetDebateEvidences(db))
protected.POST("/topics/:id/conclusions", CreateDebateConclusion(db))
protected.GET("/topics/:id/conclusions", GetDebateConclusions(db))
protected.POST("/moderation/:id/action", CreateDebateModerationAction(db))
```

Expected: the route map now reflects the planned API surface.

- [ ] **Step 4: Implement minimal handler inputs and persistence flows**

Add these input shapes and factories:
```go
type CreateDebatePositionInput struct {
	Title       string `json:"title" binding:"required"`
	Description string `json:"description"`
	SortOrder   int    `json:"sort_order"`
}

type CreateDebateEvidenceInput struct {
	SourceType      string `json:"source_type" binding:"required"`
	SourceRef       string `json:"source_ref"`
	Title           string `json:"title" binding:"required"`
	Excerpt         string `json:"excerpt"`
	CredibilityNote string `json:"credibility_note"`
}

type CreateDebateConclusionInput struct {
	Summary       string `json:"summary" binding:"required"`
	OpenQuestions string `json:"open_questions"`
}

type CreateDebateModerationActionInput struct {
	ActionType string `json:"action_type" binding:"required"`
	Reason     string `json:"reason"`
}
```

Implement each handler with the same conservative pattern used elsewhere in the file:
- verify target debate/argument exists
- read `user_id` from context
- create the model row
- return `201` with `{"data": ...}`
- use `200` with `{"data": ...}` for GET list endpoints

Expected: missing structured resources become available without rewriting working topic/argument behavior.

- [ ] **Step 5: Thread `position_id` through argument creation and update**

Extend `CreateArgumentInput` to include:
```go
PositionID *uuid.UUID `json:"position_id"`
```

Then assign `PositionID: input.PositionID` when creating an argument, and include `"position_id": input.PositionID` in argument updates.

Expected: arguments can now belong to explicit positions.

- [ ] **Step 6: Run targeted handler tests and backend build**

Run:
```bash
cd /Users/fafa/Documents/projects/Atoman/.claude/worktrees/agent-a9dac96e4ba6d6edb/server && go test ./internal/handlers -run TestCreateDebatePositionHandlerExists -v && go build ./...
```

Expected: PASS.

- [ ] **Step 7: Commit the API layer**

```bash
git add /Users/fafa/Documents/projects/Atoman/.claude/worktrees/agent-a9dac96e4ba6d6edb/server/internal/handlers/debate_handler.go \
  /Users/fafa/Documents/projects/Atoman/.claude/worktrees/agent-a9dac96e4ba6d6edb/server/internal/handlers/debate_handler_test.go

git commit -m "feat: add structured debate APIs"
```

Expected: backend APIs now match the first-pass planned contract.

## Task 4: Expand shared frontend types and store methods for structured debate data

**Files:**
- Modify: `web/src/types.ts`
- Modify: `web/src/stores/debate.ts`

- [ ] **Step 1: Write the failing frontend type-check by using missing interfaces in the store**

At the top of `web/src/stores/debate.ts`, change the import to include the new types:
```ts
import type {
  Debate,
  Argument,
  DebateVote,
  VoteHistory,
  DebatePosition,
  DebateEvidence,
  DebateConclusion,
  DebateModerationAction,
} from '@/types'
```

Expected: type-check fails because the new interfaces do not exist yet.

- [ ] **Step 2: Run type-check to verify it fails**

Run:
```bash
cd /Users/fafa/Documents/projects/Atoman/.claude/worktrees/agent-a9dac96e4ba6d6edb/web && bun run type-check
```

Expected: FAIL with missing exported debate interfaces from `web/src/types.ts`.

- [ ] **Step 3: Add the missing debate interfaces in `web/src/types.ts`**

Add these interfaces below the existing debate types:
```ts
export interface DebatePosition {
  id: string
  debate_id: string
  title: string
  description: string
  sort_order: number
  created_by: string
  created_at: string
  updated_at: string
}

export interface DebateEvidence {
  id: string
  argument_id: string
  source_type: 'url' | 'quote' | 'internal_post' | 'internal_music' | 'file'
  source_ref: string
  title: string
  excerpt: string
  credibility_note: string
  created_by: string
  created_at: string
  updated_at: string
}

export interface DebateConclusion {
  id: string
  debate_id: string
  summary: string
  open_questions: string
  version_number: number
  created_by: string
  created_at: string
  updated_at: string
}

export interface DebateModerationAction {
  id: string
  debate_id?: string
  argument_id?: string
  action_type: 'fold' | 'pin' | 'lock' | 'mark_disputed' | 'restore'
  reason: string
  created_by: string
  created_at: string
  updated_at: string
}
```

Also extend `Argument` with:
```ts
position_id?: string
position?: DebatePosition
is_folded?: boolean
```

Expected: the frontend can represent the planned backend contract.

- [ ] **Step 4: Add store state for positions, evidences, conclusions, and moderation actions**

In `useDebateStore`, add:
```ts
const positions = ref<DebatePosition[]>([])
const evidencesByArgument = ref<Record<string, DebateEvidence[]>>({})
const conclusions = ref<DebateConclusion[]>([])
const moderationActions = ref<DebateModerationAction[]>([])
```

Expected: the store can hold the structured data required by the planned UI.

- [ ] **Step 5: Add minimal fetch/create methods for the new resources**

Implement these methods using the same fetch style already used in the store:
```ts
fetchPositions(debateId: string)
createPosition(debateId: string, payload: { title: string; description: string; sort_order: number })
fetchEvidences(argumentId: string)
createEvidence(argumentId: string, payload: { source_type: string; source_ref: string; title: string; excerpt: string; credibility_note: string })
fetchConclusions(debateId: string)
createConclusion(debateId: string, payload: { summary: string; open_questions: string })
createModerationAction(targetId: string, payload: { action_type: string; reason: string })
```

Expected: the frontend store exposes every missing API surface before the view refactor starts.

- [ ] **Step 6: Run type-check**

Run:
```bash
cd /Users/fafa/Documents/projects/Atoman/.claude/worktrees/agent-a9dac96e4ba6d6edb/web && bun run type-check
```

Expected: PASS.

- [ ] **Step 7: Commit the shared contract layer**

```bash
git add /Users/fafa/Documents/projects/Atoman/.claude/worktrees/agent-a9dac96e4ba6d6edb/web/src/types.ts \
  /Users/fafa/Documents/projects/Atoman/.claude/worktrees/agent-a9dac96e4ba6d6edb/web/src/stores/debate.ts

git commit -m "feat: add structured debate frontend contracts"
```

Expected: frontend data contracts are stable before component work begins.

## Task 5: Refactor the debate topic page into a position-aware structured debate workspace

**Files:**
- Add: `web/src/components/debate/PositionColumn.vue`
- Add: `web/src/components/debate/EvidenceList.vue`
- Add: `web/src/components/debate/ConclusionTimeline.vue`
- Add: `web/src/components/debate/ModerationPanel.vue`
- Modify: `web/src/views/debate/DebateTopicView.vue`
- Modify: `web/src/components/debate/ArgumentNode.vue`

- [ ] **Step 1: Add the failing template usage for `PositionColumn`**

In `web/src/views/debate/DebateTopicView.vue`, replace the current single argument-list mount with this placeholder usage:
```vue
<div class="debate-position-grid">
  <PositionColumn
    v-for="position in positions"
    :key="position.id"
    :position="position"
    :arguments="argumentsByPosition[position.id] || []"
    :debate="debate!"
  />
</div>
```

Expected: type-check fails because `PositionColumn`, `positions`, and `argumentsByPosition` do not exist yet.

- [ ] **Step 2: Run type-check to verify it fails**

Run:
```bash
cd /Users/fafa/Documents/projects/Atoman/.claude/worktrees/agent-a9dac96e4ba6d6edb/web && bun run type-check
```

Expected: FAIL with missing component and computed properties.

- [ ] **Step 3: Create `PositionColumn.vue`**

Use this initial component shape:
```vue
<template>
  <section class="a-card">
    <header class="mb-4 border-b-2 border-black pb-3">
      <h3 class="text-xl font-black tracking-tight">{{ position.title }}</h3>
      <p v-if="position.description" class="mt-2 text-sm text-gray-700">{{ position.description }}</p>
    </header>

    <div class="space-y-4">
      <ArgumentNode
        v-for="argument in arguments"
        :key="argument.id"
        :argument="argument"
        :debate="debate"
      />
    </div>
  </section>
</template>

<script setup lang="ts">
import ArgumentNode from '@/components/debate/ArgumentNode.vue'
import type { Debate, DebatePosition, Argument } from '@/types'

defineProps<{
  position: DebatePosition
  arguments: Argument[]
  debate: Debate
}>()
</script>
```

Expected: the topic page gets a dedicated position column abstraction instead of keeping all structure in one large file.

- [ ] **Step 4: Add evidence, conclusion, and moderation side-panel components**

Create minimal first-pass components with these responsibilities:
- `EvidenceList.vue`: render evidence title, source type, excerpt, credibility note
- `ConclusionTimeline.vue`: render conclusion version, summary, open questions, timestamps
- `ModerationPanel.vue`: render action history and action buttons for admin/moderator controls

Use this exact prop contract:
```ts
// EvidenceList.vue
items: DebateEvidence[]

// ConclusionTimeline.vue
items: DebateConclusion[]

// ModerationPanel.vue
items: DebateModerationAction[]
canModerate: boolean
```

Expected: the topic page can compose structured side-panel views instead of embedding everything inline.

- [ ] **Step 5: Refactor `DebateTopicView.vue` to load and render structured data**

Add these computed values and load steps:
```ts
const positions = computed(() => debateStore.positions)
const conclusions = computed(() => debateStore.conclusions)
const moderationActions = computed(() => debateStore.moderationActions)

const argumentsByPosition = computed(() => {
  const grouped: Record<string, Argument[]> = {}
  for (const argument of argumentsList.value) {
    const key = argument.position_id || 'unassigned'
    grouped[key] ||= []
    grouped[key].push(argument)
  }
  return grouped
})
```

Inside the initial page loader, fetch in this order:
```ts
await debateStore.fetchDebate(id)
await debateStore.fetchPositions(id)
await debateStore.fetchArguments(id)
await debateStore.fetchConclusions(id)
```

Then render the page in three regions:
```text
- header and topic summary
- center position columns
- right-side evidence / conclusions / moderation stack
```

Expected: the detail page reflects the planned structured debate interaction model.

- [ ] **Step 6: Refactor `ArgumentNode.vue` for tree-oriented semantics**

Make these changes:
```text
- rename the visible “引用” action label to “回复/引用” only if it still maps to parent-based response behavior
- show position-aware context when present
- show folded-state UI when `argument.is_folded` is true
- reserve the evidence area for explicit evidence entities instead of overloading quoted arguments
```

Use token-driven styles where possible instead of introducing more hardcoded colors.

Expected: argument cards match the planned semantics more closely without breaking the current route flow.

- [ ] **Step 7: Run type-check and build**

Run:
```bash
cd /Users/fafa/Documents/projects/Atoman/.claude/worktrees/agent-a9dac96e4ba6d6edb/web && bun run type-check && bun run build
```

Expected: PASS.

- [ ] **Step 8: Commit the structured topic UI**

```bash
git add /Users/fafa/Documents/projects/Atoman/.claude/worktrees/agent-a9dac96e4ba6d6edb/web/src/views/debate/DebateTopicView.vue \
  /Users/fafa/Documents/projects/Atoman/.claude/worktrees/agent-a9dac96e4ba6d6edb/web/src/components/debate/ArgumentNode.vue \
  /Users/fafa/Documents/projects/Atoman/.claude/worktrees/agent-a9dac96e4ba6d6edb/web/src/components/debate/PositionColumn.vue \
  /Users/fafa/Documents/projects/Atoman/.claude/worktrees/agent-a9dac96e4ba6d6edb/web/src/components/debate/EvidenceList.vue \
  /Users/fafa/Documents/projects/Atoman/.claude/worktrees/agent-a9dac96e4ba6d6edb/web/src/components/debate/ConclusionTimeline.vue \
  /Users/fafa/Documents/projects/Atoman/.claude/worktrees/agent-a9dac96e4ba6d6edb/web/src/components/debate/ModerationPanel.vue

git commit -m "feat: add structured debate topic workspace"
```

Expected: the main debate page now reflects the target product shape.

## Task 6: Upgrade the debate homepage to match the planned discovery model

**Files:**
- Modify: `web/src/views/debate/DebateHomeView.vue`
- Modify: `web/src/stores/debate.ts`

- [ ] **Step 1: Write the failing UI expectation in working notes**

Use this exact target list:
```text
Homepage sections required:
- 热门议题
- 最新议题
- 高争议议题
- 已形成结论的议题
```

Expected: a visible acceptance target before page edits begin.

- [ ] **Step 2: Extend the store with derived collections**

Add these computed helpers or derived methods in `web/src/stores/debate.ts`:
```ts
const latestDebates = computed(() => [...debates.value].sort((a, b) => Date.parse(b.created_at) - Date.parse(a.created_at)))
const concludedDebates = computed(() => debates.value.filter(item => item.status === 'concluded'))
const controversialDebates = computed(() => [...debates.value].sort((a, b) => b.argument_count - a.argument_count))
const hotDebates = computed(() => [...debates.value].sort((a, b) => (b.view_count + b.vote_count) - (a.view_count + a.vote_count)))
```

Expected: the homepage can render its planned sections without inventing a new backend endpoint yet.

- [ ] **Step 3: Refactor `DebateHomeView.vue` into sectioned discovery blocks**

Replace the current single-card grid mental model with four sections using the store collections above:
```text
- 热门议题
- 最新议题
- 高争议议题
- 已形成结论的议题
```

Keep the existing create-debate modal and filters, but move them below the main discovery sections.

Expected: the homepage aligns with the planned “discovery first” structure.

- [ ] **Step 4: Run type-check and build**

Run:
```bash
cd /Users/fafa/Documents/projects/Atoman/.claude/worktrees/agent-a9dac96e4ba6d6edb/web && bun run type-check && bun run build
```

Expected: PASS.

- [ ] **Step 5: Commit the homepage redesign**

```bash
git add /Users/fafa/Documents/projects/Atoman/.claude/worktrees/agent-a9dac96e4ba6d6edb/web/src/views/debate/DebateHomeView.vue \
  /Users/fafa/Documents/projects/Atoman/.claude/worktrees/agent-a9dac96e4ba6d6edb/web/src/stores/debate.ts

git commit -m "feat: align debate homepage with structured discovery design"
```

Expected: the homepage now matches the planned entry-point behavior.

## Task 7: Add browser verification for the critical debate flows

**Files:**
- Add: `web/playwright.config.ts`
- Add: `web/tests/debate.spec.ts`

- [ ] **Step 1: Create the Playwright config**

Create `web/playwright.config.ts` with this minimal config:
```ts
import { defineConfig } from '@playwright/test'

export default defineConfig({
  testDir: './tests',
  use: {
    baseURL: 'http://127.0.0.1:4173',
    trace: 'on-first-retry',
  },
  webServer: {
    command: 'bun run build && bun run preview --host 127.0.0.1 --port 4173',
    port: 4173,
    reuseExistingServer: true,
    timeout: 120000,
  },
})
```

Expected: browser verification has a reproducible entry point.

- [ ] **Step 2: Create the debate smoke test**

Create `web/tests/debate.spec.ts` with this test:
```ts
import { test, expect } from '@playwright/test'

test('debate home and topic pages render structured sections', async ({ page }) => {
  await page.goto('/debate')
  await expect(page.getByText('辩论')).toBeVisible()
  await expect(page.getByText('热门议题')).toBeVisible()
  await expect(page.getByText('最新议题')).toBeVisible()

  const firstCard = page.locator('[class*="a-card"]').first()
  await expect(firstCard).toBeVisible()
  await firstCard.click()

  await expect(page.getByText('论点列表')).toBeVisible()
})
```

Expected: FAIL first if the page has not been refactored to the structured headings yet.

- [ ] **Step 3: Run the Playwright test and make it pass**

Run:
```bash
cd /Users/fafa/Documents/projects/Atoman/.claude/worktrees/agent-a9dac96e4ba6d6edb/web && bun run test:e2e --grep "debate home and topic pages render structured sections"
```

Expected: PASS after the homepage and topic page are updated.

- [ ] **Step 4: Commit the browser coverage**

```bash
git add /Users/fafa/Documents/projects/Atoman/.claude/worktrees/agent-a9dac96e4ba6d6edb/web/playwright.config.ts \
  /Users/fafa/Documents/projects/Atoman/.claude/worktrees/agent-a9dac96e4ba6d6edb/web/tests/debate.spec.ts

git commit -m "test: add debate browser smoke coverage"
```

Expected: the main debate user path is no longer verified by memory alone.

## Task 8: Final end-to-end verification and plan closeout

**Files:**
- Verify: `server/internal/model/debate.go`
- Verify: `server/internal/handlers/debate_handler.go`
- Verify: `server/cmd/start_server/main.go`
- Verify: `web/src/types.ts`
- Verify: `web/src/stores/debate.ts`
- Verify: `web/src/views/debate/DebateHomeView.vue`
- Verify: `web/src/views/debate/DebateTopicView.vue`
- Verify: `web/src/components/debate/ArgumentNode.vue`
- Verify: `web/src/components/debate/PositionColumn.vue`
- Verify: `web/src/components/debate/EvidenceList.vue`
- Verify: `web/src/components/debate/ConclusionTimeline.vue`
- Verify: `web/src/components/debate/ModerationPanel.vue`
- Verify: `web/playwright.config.ts`
- Verify: `web/tests/debate.spec.ts`

- [ ] **Step 1: Run backend tests and build**

Run:
```bash
cd /Users/fafa/Documents/projects/Atoman/.claude/worktrees/agent-a9dac96e4ba6d6edb/server && go test ./... && go build ./...
```

Expected: PASS.

- [ ] **Step 2: Run frontend type-check and build**

Run:
```bash
cd /Users/fafa/Documents/projects/Atoman/.claude/worktrees/agent-a9dac96e4ba6d6edb/web && bun run type-check && bun run build
```

Expected: PASS.

- [ ] **Step 3: Run debate browser verification**

Run:
```bash
cd /Users/fafa/Documents/projects/Atoman/.claude/worktrees/agent-a9dac96e4ba6d6edb/web && bun run test:e2e --grep "debate"
```

Expected: PASS.

- [ ] **Step 4: Manual verification in the browser**

Verify all of the following exactly:
```text
- /debate shows 热门议题, 最新议题, 高争议议题, 已形成结论的议题
- authenticated users can still open the create-debate modal
- /debate/:id loads the topic header and structured position columns
- adding an argument still works and attaches to a position when selected
- evidence entries render in the side panel for the selected argument or debate context
- conclusion timeline shows versioned summary entries
- admin or moderator controls expose fold / lock / mark-disputed style actions
- existing conclude / reopen flows still work
```

Expected: the feature matches both the planning documents and the implemented route flow.

- [ ] **Step 5: Commit the completed feature**

```bash
git add /Users/fafa/Documents/projects/Atoman/.claude/worktrees/agent-a9dac96e4ba6d6edb/server/internal/model/debate.go \
  /Users/fafa/Documents/projects/Atoman/.claude/worktrees/agent-a9dac96e4ba6d6edb/server/internal/model/debate_test.go \
  /Users/fafa/Documents/projects/Atoman/.claude/worktrees/agent-a9dac96e4ba6d6edb/server/internal/handlers/debate_handler.go \
  /Users/fafa/Documents/projects/Atoman/.claude/worktrees/agent-a9dac96e4ba6d6edb/server/internal/handlers/debate_handler_test.go \
  /Users/fafa/Documents/projects/Atoman/.claude/worktrees/agent-a9dac96e4ba6d6edb/server/cmd/start_server/main.go \
  /Users/fafa/Documents/projects/Atoman/.claude/worktrees/agent-a9dac96e4ba6d6edb/web/src/types.ts \
  /Users/fafa/Documents/projects/Atoman/.claude/worktrees/agent-a9dac96e4ba6d6edb/web/src/stores/debate.ts \
  /Users/fafa/Documents/projects/Atoman/.claude/worktrees/agent-a9dac96e4ba6d6edb/web/src/views/debate/DebateHomeView.vue \
  /Users/fafa/Documents/projects/Atoman/.claude/worktrees/agent-a9dac96e4ba6d6edb/web/src/views/debate/DebateTopicView.vue \
  /Users/fafa/Documents/projects/Atoman/.claude/worktrees/agent-a9dac96e4ba6d6edb/web/src/components/debate/ArgumentNode.vue \
  /Users/fafa/Documents/projects/Atoman/.claude/worktrees/agent-a9dac96e4ba6d6edb/web/src/components/debate/PositionColumn.vue \
  /Users/fafa/Documents/projects/Atoman/.claude/worktrees/agent-a9dac96e4ba6d6edb/web/src/components/debate/EvidenceList.vue \
  /Users/fafa/Documents/projects/Atoman/.claude/worktrees/agent-a9dac96e4ba6d6edb/web/src/components/debate/ConclusionTimeline.vue \
  /Users/fafa/Documents/projects/Atoman/.claude/worktrees/agent-a9dac96e4ba6d6edb/web/src/components/debate/ModerationPanel.vue \
  /Users/fafa/Documents/projects/Atoman/.claude/worktrees/agent-a9dac96e4ba6d6edb/web/playwright.config.ts \
  /Users/fafa/Documents/projects/Atoman/.claude/worktrees/agent-a9dac96e4ba6d6edb/web/tests/debate.spec.ts

git commit -m "feat: complete structured debate system"
```

Expected: a single explicit closeout checkpoint after all verification passes.

## Verification

After execution, verify all of the following:

1. `server/internal/model/debate.go` contains explicit models for positions, evidence, conclusions, and moderation actions.
2. `server/internal/handlers/debate_handler.go` exposes routes and handlers for those resources.
3. `web/src/types.ts` and `web/src/stores/debate.ts` can represent and fetch the new structured resources.
4. `web/src/views/debate/DebateHomeView.vue` presents the planned discovery sections.
5. `web/src/views/debate/DebateTopicView.vue` renders a position-aware structured debate layout.
6. `cd /Users/fafa/Documents/projects/Atoman/.claude/worktrees/agent-a9dac96e4ba6d6edb/server && go test ./... && go build ./...` passes.
7. `cd /Users/fafa/Documents/projects/Atoman/.claude/worktrees/agent-a9dac96e4ba6d6edb/web && bun run type-check && bun run build` passes.
8. `cd /Users/fafa/Documents/projects/Atoman/.claude/worktrees/agent-a9dac96e4ba6d6edb/web && bun run test:e2e --grep "debate"` passes.
9. Manual browser verification confirms topic creation, topic viewing, argument creation, structured position rendering, evidence display, conclusion timeline, moderation controls, and conclude/reopen behavior.

## Notes For The Implementer

- Do not delete the existing conclude/reopen flow while adding iterative conclusions; preserve it as the topic-level status transition until a better replacement is deliberately designed.
- Prefer additive schema changes over destructive rewrites so existing local SQLite development data keeps booting.
- Keep Vue components focused; do not let `DebateTopicView.vue` absorb all structured-debate logic.
- Follow the project UI rules from `CLAUDE.md`: use `A*` components first, prefer design tokens over hardcoded colors, and avoid expanding inline layout styles.
- If a future requirement adds trust scoring or moderator roles, create a separate plan rather than overloading this one.
