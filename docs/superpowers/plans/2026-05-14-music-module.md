# Music Module Completion Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task.

**Goal:** 完成 Music 模块的维基化闭环：① 添加 `ArtistCorrection` 模型和相关 handler，让 Artist 与 Album 共享相同的修改建议流程；② 在 `AlbumDetailView.vue` 和 `ArtistDetailView.vue` 中根据 `entry_status` 区分"编辑"与"提交修改建议"两种操作模式；③ 更新 `AdminReviewView.vue` 纳入艺人修改建议的管理审核界面。

**Architecture:**
- `entry_status` 三种状态：`open`（开放编辑）/ `confirmed`（已确认，需提交建议）/ `disputed`（争议中）
- 已存在：`AlbumCorrection` 和 `SongCorrection` 模型、`corrections_handler.go`、`admin_handler.go` 中的 Album Correction 审核端点
- 缺失：`ArtistCorrection` 模型 + 端点 + 管理审核端点
- 前端规则：`admin` 始终看到编辑按钮；非 admin 在 `entry_status === 'confirmed'` 时只看到"提交修改建议"按钮，其他状态看编辑按钮

**已完整的部分（无需修改）：**
- `entry_status_handler.go`: `POST /api/albums/:id/status`, `POST /api/artists/:id/status`, `GET /api/admin/music/entries`
- `corrections_handler.go`: `POST /api/corrections/album`, `POST /api/corrections/song`
- `admin_handler.go`: album correction 的 GET/APPROVE/REJECT 端点
- `AlbumDetailView.vue`: 专辑歌曲列表、播放功能、歌词查看
- `ArtistDetailView.vue`: 艺人基本信息、别名、entry_status 角标、管理状态切换
- `AdminReviewView.vue`: 已实现 pending albums、album corrections、song corrections 审核

---

## 文件清单

### 修改
| 文件 | 改动 |
|------|------|
| `server/internal/model/music.go` | 添加 `ArtistCorrection` struct |
| `server/cmd/start_server/main.go` | 在 AutoMigrate 列表中注册 `ArtistCorrection` |
| `server/internal/handlers/corrections_handler.go` | 添加 `CreateArtistCorrection` handler + 路由注册 |
| `server/internal/handlers/admin_handler.go` | 添加艺人修改建议的 GET/APPROVE/REJECT handler + 路由注册 |
| `web/src/views/music/AlbumDetailView.vue` | 直接调用 `/api/albums/:id` 获取 entry_status；条件显示编辑/建议按钮 |
| `web/src/views/music/ArtistDetailView.vue` | 条件显示编辑/建议按钮 |
| `web/src/views/music/AdminReviewView.vue` | 添加艺人修改建议审核面板 |
| `web/src/types.ts` | 添加 `ArtistCorrection` 类型 |

### 可能新增（若尚无通用 Proposal Modal）
| 文件 | 职责 |
|------|------|
| `web/src/components/music/CorrectionProposalModal.vue` | 提交修改建议的模态框（专辑和艺人共用） |

---

## Task 1：后端模型 — 添加 ArtistCorrection

**Files:**
- Modify: `server/internal/model/music.go`

- [ ] **Step 1: 读取 music.go 末尾，找到 AlbumCorrection 定义**

确认 `AlbumCorrection` 的完整字段，参照其结构实现 `ArtistCorrection`。

- [ ] **Step 2: 在 music.go 末尾（AlbumCorrection 下方）添加 ArtistCorrection**

```go
// ArtistCorrection is a proposed change to a confirmed Artist entry, submitted by users.
// Status: pending | approved | rejected
type ArtistCorrection struct {
	Base
	ArtistID    uuid.UUID  `json:"artist_id" gorm:"type:uuid;not null"`
	Artist      *Artist    `json:"artist,omitempty" gorm:"foreignKey:ArtistID"`
	UserID      *uuid.UUID `json:"user_id" gorm:"type:uuid"`
	User        *User      `json:"user,omitempty" gorm:"foreignKey:UserID;references:UUID"`
	Description string     `json:"description" gorm:"type:text;not null"` // 修改说明
	Reason      string     `json:"reason" gorm:"type:text"`               // 修改理由
	Status      string     `json:"status" gorm:"default:'pending'"`        // pending|approved|rejected
	ApprovedBy  *uuid.UUID `json:"approved_by" gorm:"type:uuid"`
	ApprovedAt  *time.Time `json:"approved_at"`
}

func (ArtistCorrection) TableName() string { return "artist_corrections" }
```

确认 `time` 包已在文件顶部 import（若无则添加）。

- [ ] **Step 3: Commit**

```bash
cd server && go build ./...
git add server/internal/model/music.go
git commit -m "feat(music): add ArtistCorrection model"
```

---

## Task 2：后端迁移 — 注册 ArtistCorrection 到 AutoMigrate

**Files:**
- Modify: `server/cmd/start_server/main.go`

- [ ] **Step 1: 找到 AutoMigrate 调用**

```bash
grep -n "AutoMigrate\|AlbumCorrection" server/cmd/start_server/main.go | head -20
```

- [ ] **Step 2: 在 AlbumCorrection（或 SongCorrection）后面添加 ArtistCorrection**

找到类似：

```go
&model.AlbumCorrection{},
```

在其下方添加：

```go
&model.ArtistCorrection{},
```

- [ ] **Step 3: 编译 + 验证**

```bash
cd server && go build ./...
```

期望：无错误。

- [ ] **Step 4: 运行迁移（本地 dev 环境）**

```bash
go run cmd/start_server/main.go &
# Server starts and AutoMigrate creates artist_corrections table
```

验证：

```bash
# 若使用 SQLite dev 模式
sqlite3 atoman_dev.db ".tables" | grep artist_corrections
```

- [ ] **Step 5: Commit**

```bash
git add server/cmd/start_server/main.go
git commit -m "feat(music): register ArtistCorrection in AutoMigrate"
```

---

## Task 3：后端 handler — 提交艺人修改建议

**Files:**
- Modify: `server/internal/handlers/corrections_handler.go`

- [ ] **Step 1: 读取 CreateAlbumCorrection 实现**

```bash
grep -n "CreateAlbumCorrection\|func Create" server/internal/handlers/corrections_handler.go | head -20
```

读取 `CreateAlbumCorrection` 函数体，参照其结构实现 `CreateArtistCorrection`。

- [ ] **Step 2: 添加 CreateArtistCorrection handler**

在文件末尾添加（参照 `CreateAlbumCorrection` 的模式）：

```go
// CreateArtistCorrection submits a proposed change for a confirmed artist entry.
// Route: POST /api/corrections/artist
// Auth: Required
func CreateArtistCorrection(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.MustGet("userID").(uuid.UUID)

		var req struct {
			ArtistID    string `json:"artist_id" binding:"required"`
			Description string `json:"description" binding:"required"`
			Reason      string `json:"reason"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		artistUUID, err := uuid.Parse(req.ArtistID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid artist_id"})
			return
		}

		// Verify the artist exists
		var artist model.Artist
		if err := db.First(&artist, "id = ?", artistUUID).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "artist not found"})
			return
		}

		correction := model.ArtistCorrection{
			ArtistID:    artistUUID,
			UserID:      &userID,
			Description: req.Description,
			Reason:      req.Reason,
			Status:      "pending",
		}
		if err := db.Create(&correction).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create correction"})
			return
		}

		c.JSON(http.StatusCreated, gin.H{"data": correction})
	}
}
```

- [ ] **Step 3: 注册路由**

找到 `SetupCorrectionRoutes`（或 corrections 路由注册函数），在 `POST /artist` 或类似位置添加：

```go
corrections.POST("/artist", middleware.RequireAuth(), CreateArtistCorrection(db))
```

若无单独的 SetupCorrectionRoutes 函数，则在主路由注册文件（`cmd/start_server/main.go` 或 router setup）中添加：

```go
router.POST("/api/corrections/artist", middleware.RequireAuth(), handlers.CreateArtistCorrection(db))
```

- [ ] **Step 4: 编译验证**

```bash
cd server && go build ./...
```

- [ ] **Step 5: Commit**

```bash
git add server/internal/handlers/corrections_handler.go
git commit -m "feat(music): add POST /api/corrections/artist endpoint"
```

---

## Task 4：后端 handler — 管理员审核艺人修改建议

**Files:**
- Modify: `server/internal/handlers/admin_handler.go`

- [ ] **Step 1: 读取已有 album correction 审核 handler 的结构**

```bash
grep -n "PendingAlbumCorrection\|ApproveAlbumCorrection\|RejectAlbumCorrection" server/internal/handlers/admin_handler.go
```

读取对应函数体，参照其结构实现艺人版本。

- [ ] **Step 2: 添加三个 handler（GET 列表 + APPROVE + REJECT）**

在 `admin_handler.go` 末尾添加：

```go
// GetPendingArtistCorrections returns all pending artist correction proposals.
// Route: GET /api/admin/pending-artist-corrections
// Auth: Admin only
func GetPendingArtistCorrections(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var corrections []model.ArtistCorrection
		db.Where("status = ?", "pending").
			Preload("Artist").
			Preload("User").
			Order("created_at ASC").
			Find(&corrections)
		c.JSON(http.StatusOK, gin.H{"data": corrections})
	}
}

// ApproveArtistCorrection marks an artist correction as approved.
// Route: POST /api/admin/approve-artist-correction/:id
// Auth: Admin only
func ApproveArtistCorrection(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		adminID := c.MustGet("userID").(uuid.UUID)
		now := time.Now()

		var correction model.ArtistCorrection
		if err := db.First(&correction, "id = ?", id).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "correction not found"})
			return
		}
		correction.Status = "approved"
		correction.ApprovedBy = &adminID
		correction.ApprovedAt = &now
		if err := db.Save(&correction).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": correction})
	}
}

// RejectArtistCorrection marks an artist correction as rejected.
// Route: POST /api/admin/reject-artist-correction/:id
// Auth: Admin only
func RejectArtistCorrection(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")

		var correction model.ArtistCorrection
		if err := db.First(&correction, "id = ?", id).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "correction not found"})
			return
		}
		correction.Status = "rejected"
		if err := db.Save(&correction).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": correction})
	}
}
```

- [ ] **Step 3: 注册路由**

在路由注册处找到 album correction 的三个 admin 路由，在其下方添加：

```go
admin.GET("/pending-artist-corrections", GetPendingArtistCorrections(db))
admin.POST("/approve-artist-correction/:id", ApproveArtistCorrection(db))
admin.POST("/reject-artist-correction/:id", RejectArtistCorrection(db))
```

- [ ] **Step 4: 确认 time 包已 import**

若 `admin_handler.go` 中已有 `time.Time` 使用则无需操作。

- [ ] **Step 5: 编译验证**

```bash
cd server && go build ./...
```

- [ ] **Step 6: Commit**

```bash
git add server/internal/handlers/admin_handler.go
git commit -m "feat(music): add admin artist correction review endpoints"
```

---

## Task 5：前端类型 — 添加 ArtistCorrection 类型

**Files:**
- Modify: `web/src/types.ts`

- [ ] **Step 1: 找到 AlbumCorrection 类型定义**

```bash
grep -n "AlbumCorrection\|SongCorrection" web/src/types.ts
```

- [ ] **Step 2: 在其下方添加 ArtistCorrection 类型**

```typescript
export interface ArtistCorrection {
  id: string
  artist_id: string
  artist?: Artist
  user_id?: string
  user?: User
  description: string
  reason?: string
  status: 'pending' | 'approved' | 'rejected'
  approved_by?: string
  approved_at?: string
  created_at: string
  updated_at: string
}
```

- [ ] **Step 3: 类型检查**

```bash
cd web && bun run type-check 2>&1 | tail -5
```

- [ ] **Step 4: Commit**

```bash
git add web/src/types.ts
git commit -m "feat(music): add ArtistCorrection type to types.ts"
```

---

## Task 6：前端 — CorrectionProposalModal 共用组件

**Files:**
- Create: `web/src/components/music/CorrectionProposalModal.vue`

**背景：** 专辑和艺人的"提交修改建议"使用同一个 Modal，通过 `type: 'album' | 'artist'` 和 `targetId` prop 区分。

- [ ] **Step 1: 检查是否已有类似组件**

```bash
ls web/src/components/music/
```

若存在 `CorrectionModal.vue` 或 `ProposalModal.vue` 等，参考其实现并按需改造，不要重复创建。

- [ ] **Step 2: 创建 CorrectionProposalModal.vue**

```vue
<template>
  <div v-if="show" class="a-modal-backdrop" @click.self="$emit('close')">
    <div class="a-modal" style="max-width:36rem">
      <h2 class="a-modal-title" style="font-size:1rem;font-weight:900;margin-bottom:1rem">
        提交修改建议
      </h2>
      <form @submit.prevent="submit">
        <div class="a-field">
          <label class="a-label">修改说明 <span style="color:var(--a-color-danger)">*</span></label>
          <textarea
            v-model="description"
            class="a-textarea"
            placeholder="描述你希望修改的内容……"
            rows="4"
            required
          />
        </div>
        <div class="a-field" style="margin-top:.75rem">
          <label class="a-label">修改理由（可选）</label>
          <textarea
            v-model="reason"
            class="a-textarea"
            placeholder="为什么需要这个修改？"
            rows="2"
          />
        </div>
        <div style="display:flex;gap:.5rem;margin-top:1.25rem;justify-content:flex-end">
          <button type="button" class="a-btn a-btn-ghost" @click="$emit('close')">取消</button>
          <button type="submit" class="a-btn a-btn-primary" :disabled="submitting">
            {{ submitting ? '提交中…' : '提交建议' }}
          </button>
        </div>
      </form>
      <p v-if="errorMsg" style="color:var(--a-color-danger);font-size:.75rem;margin-top:.5rem">
        {{ errorMsg }}
      </p>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { useApi } from '@/composables/useApi'

const props = defineProps<{
  show: boolean
  type: 'album' | 'artist'
  targetId: string
}>()

const emit = defineEmits<{
  (e: 'close'): void
  (e: 'submitted'): void
}>()

const api = useApi()
const description = ref('')
const reason = ref('')
const submitting = ref(false)
const errorMsg = ref('')

async function submit() {
  if (!description.value.trim()) return
  submitting.value = true
  errorMsg.value = ''
  try {
    const payload: Record<string, string> = {
      description: description.value,
      reason: reason.value,
    }
    if (props.type === 'album') {
      payload['album_id'] = props.targetId
      await api.post('/api/corrections/album', payload)
    } else {
      payload['artist_id'] = props.targetId
      await api.post('/api/corrections/artist', payload)
    }
    description.value = ''
    reason.value = ''
    emit('submitted')
    emit('close')
  } catch (e: any) {
    errorMsg.value = e?.message || '提交失败，请稍后再试'
  } finally {
    submitting.value = false
  }
}
</script>
```

- [ ] **Step 3: 类型检查**

```bash
cd web && bun run type-check 2>&1 | tail -5
```

- [ ] **Step 4: Commit**

```bash
git add web/src/components/music/CorrectionProposalModal.vue
git commit -m "feat(music): create shared CorrectionProposalModal component"
```

---

## Task 7：前端 — AlbumDetailView.vue 条件编辑/建议按钮

**Files:**
- Modify: `web/src/views/music/AlbumDetailView.vue`

**背景：**
- `AlbumDetailView.vue` 通过 `player.songs` 获取专辑数据，无法直接获得 `entry_status`
- 需要在 `onMounted` 额外调用 `/api/albums/:id` 获取完整 album（含 entry_status）
- 按钮逻辑：`canDirectEdit = isAdmin || entry_status !== 'confirmed'`；若不能直接编辑则显示"提交修改建议"

- [ ] **Step 1: 读取 AlbumDetailView.vue 的 script 部分**

```bash
grep -n "canEdit\|protection\|entry_status\|onMounted\|albumId\|route.params" web/src/views/music/AlbumDetailView.vue | head -30
```

- [ ] **Step 2: 添加 entry_status 状态和专辑直接获取**

在 `<script setup>` 中添加：

```typescript
import { useAuthStore } from '@/stores/auth'
import CorrectionProposalModal from '@/components/music/CorrectionProposalModal.vue'

const authStore = useAuthStore()
const entryStatus = ref<'open' | 'confirmed' | 'disputed'>('open')
const showProposalModal = ref(false)
const albumIdStr = computed(() => String(route.params.albumId || route.params.id || ''))

async function fetchAlbumEntryStatus() {
  if (!albumIdStr.value) return
  try {
    const res = await api.get<{ data: { entry_status: string } }>(`/api/albums/${albumIdStr.value}`)
    entryStatus.value = (res.data.data?.entry_status || 'open') as 'open' | 'confirmed' | 'disputed'
  } catch {}
}
```

在 `onMounted` 中调用 `fetchAlbumEntryStatus()`。

添加计算属性：

```typescript
const canDirectEdit = computed(() =>
  authStore.user?.role === 'admin' || entryStatus.value !== 'confirmed'
)
```

- [ ] **Step 3: 更新模板的编辑/建议按钮**

找到当前"编辑"按钮（RouterLink 或 button），将其替换为：

```vue
<!-- Edit vs Proposal button based on entry_status -->
<RouterLink
  v-if="canDirectEdit"
  :to="`/music/albums/${albumIdStr}/edit`"
  class="a-btn a-btn-ghost"
>编辑</RouterLink>
<button
  v-else
  class="a-btn a-btn-ghost"
  @click="showProposalModal = true"
>提交修改建议</button>

<!-- Entry status badge -->
<span
  v-if="entryStatus === 'confirmed'"
  class="status-badge status-confirmed"
  style="font-size:.65rem;padding:.15rem .5rem"
>已确认</span>
<span
  v-else-if="entryStatus === 'disputed'"
  class="status-badge status-disputed"
  style="font-size:.65rem;padding:.15rem .5rem"
>争议</span>

<!-- Proposal modal -->
<CorrectionProposalModal
  :show="showProposalModal"
  type="album"
  :target-id="albumIdStr"
  @close="showProposalModal = false"
  @submitted="showProposalModal = false"
/>
```

- [ ] **Step 4: 类型检查**

```bash
cd web && bun run type-check 2>&1 | tail -5
```

- [ ] **Step 5: Commit**

```bash
git add web/src/views/music/AlbumDetailView.vue
git commit -m "feat(music): conditional edit/proposal button in AlbumDetailView based on entry_status"
```

---

## Task 8：前端 — ArtistDetailView.vue 条件编辑/建议按钮

**Files:**
- Modify: `web/src/views/music/ArtistDetailView.vue`

**背景：** `artist.entry_status` 已在 ArtistDetailView.vue 中渲染为角标，`artist` 对象已包含 `entry_status` 字段，只需在按钮逻辑中利用它。

- [ ] **Step 1: 读取 ArtistDetailView.vue 的 nav actions 区域**

定位：

```vue
<!-- Nav actions -->
<div class="artist-nav">
  <RouterLink :to="`/music/artists/${artistId}/edit`" class="nav-btn">编辑</RouterLink>
  ...
```

（已在上文研究中确认在文件约 46–50 行）

- [ ] **Step 2: 确认 authStore 已被引入**

```bash
grep -n "useAuthStore\|authStore" web/src/views/music/ArtistDetailView.vue | head -5
```

若未引入则在 `<script setup>` 顶部添加：

```typescript
import { useAuthStore } from '@/stores/auth'
const authStore = useAuthStore()
```

- [ ] **Step 3: 添加 canDirectEdit 计算属性**

```typescript
import CorrectionProposalModal from '@/components/music/CorrectionProposalModal.vue'

const showArtistProposalModal = ref(false)

const canDirectEditArtist = computed(() =>
  authStore.user?.role === 'admin' || artist.value?.entry_status !== 'confirmed'
)
```

- [ ] **Step 4: 替换模板中的编辑按钮**

将：

```vue
<RouterLink :to="`/music/artists/${artistId}/edit`" class="nav-btn">编辑</RouterLink>
```

替换为：

```vue
<RouterLink
  v-if="canDirectEditArtist"
  :to="`/music/artists/${artistId}/edit`"
  class="nav-btn"
>编辑</RouterLink>
<button
  v-else
  class="nav-btn"
  @click="showArtistProposalModal = true"
>提交修改建议</button>

<CorrectionProposalModal
  :show="showArtistProposalModal"
  type="artist"
  :target-id="String(artistId)"
  @close="showArtistProposalModal = false"
  @submitted="showArtistProposalModal = false"
/>
```

- [ ] **Step 5: 类型检查**

```bash
cd web && bun run type-check 2>&1 | tail -5
```

- [ ] **Step 6: Commit**

```bash
git add web/src/views/music/ArtistDetailView.vue
git commit -m "feat(music): conditional edit/proposal button in ArtistDetailView based on entry_status"
```

---

## Task 9：前端 — AdminReviewView.vue 添加艺人修改建议面板

**Files:**
- Modify: `web/src/views/music/AdminReviewView.vue`

**背景：** `AdminReviewView.vue` 已有 `pending-album-corrections` 面板，参照其结构添加 `pending-artist-corrections` 面板。

- [ ] **Step 1: 读取 AdminReviewView.vue 的 album corrections 面板实现**

```bash
grep -n "AlbumCorrection\|album-correction\|pendingAlbumCorrections" web/src/views/music/AdminReviewView.vue | head -20
```

读取完整的 album corrections panel（数据加载 + 渲染 + 审批/拒绝操作）。

- [ ] **Step 2: 添加 ArtistCorrection 数据状态**

在 `<script setup>` 中参照 album corrections 模式添加：

```typescript
import type { ArtistCorrection } from '@/types'

const artistCorrections = ref<ArtistCorrection[]>([])
const artistCorrectionsLoading = ref(false)

async function loadArtistCorrections() {
  artistCorrectionsLoading.value = true
  try {
    const res = await api.get<{ data: ArtistCorrection[] }>('/api/admin/pending-artist-corrections')
    artistCorrections.value = res.data.data || []
  } finally {
    artistCorrectionsLoading.value = false
  }
}

async function approveArtistCorrection(id: string) {
  await api.post(`/api/admin/approve-artist-correction/${id}`)
  await loadArtistCorrections()
}

async function rejectArtistCorrection(id: string) {
  await api.post(`/api/admin/reject-artist-correction/${id}`)
  await loadArtistCorrections()
}
```

在 `onMounted` 中调用 `loadArtistCorrections()`。

- [ ] **Step 3: 在模板中添加艺人修改建议面板**

参照已有 album corrections 面板，在其下方添加类似结构：

```vue
<!-- Artist Corrections -->
<section class="review-section">
  <h2 class="review-section-title">艺人修改建议</h2>
  <div v-if="artistCorrectionsLoading" class="a-skeleton" style="height:4rem" />
  <p v-else-if="!artistCorrections.length" class="review-empty">暂无待审核的艺人修改建议</p>
  <ul v-else class="review-list">
    <li v-for="item in artistCorrections" :key="item.id" class="review-item">
      <div class="review-item-header">
        <RouterLink :to="`/music/artists/${item.artist_id}`" class="review-item-title">
          {{ item.artist?.name || item.artist_id }}
        </RouterLink>
        <span class="review-item-user">由 {{ item.user?.display_name || item.user?.username || '匿名' }} 提交</span>
      </div>
      <p class="review-item-desc">{{ item.description }}</p>
      <p v-if="item.reason" class="review-item-reason" style="color:var(--a-color-muted);font-size:.75rem">
        理由：{{ item.reason }}
      </p>
      <div class="review-item-actions">
        <button class="a-btn a-btn-primary" style="font-size:.75rem" @click="approveArtistCorrection(item.id)">批准</button>
        <button class="a-btn a-btn-danger" style="font-size:.75rem" @click="rejectArtistCorrection(item.id)">拒绝</button>
      </div>
    </li>
  </ul>
</section>
```

- [ ] **Step 4: 类型检查**

```bash
cd web && bun run type-check 2>&1 | tail -5
```

- [ ] **Step 5: Commit**

```bash
git add web/src/views/music/AdminReviewView.vue
git commit -m "feat(music): add artist corrections panel to AdminReviewView"
```

---

## Task 10：验收

- [ ] **Step 1: 后端构建**

```bash
cd server && go build ./...
```

- [ ] **Step 2: 前端构建**

```bash
cd web && bun run type-check && bun run build
```

期望：0 type errors，build 成功。

- [ ] **Step 3: 手动冒烟测试**

**艺人修改建议流程：**
1. 以管理员身份访问某艺人页面 → 将其状态改为 `confirmed`
2. 退出管理员，用普通用户登录 → 访问同一艺人页面 → 看到"提交修改建议"按钮（而非"编辑"）
3. 点击按钮 → 填写说明 → 提交 → 201 成功
4. 管理员登录 → 访问 `/admin/music-review` → 在"艺人修改建议"面板看到该条记录 → 批准 → 记录消失

**专辑修改建议流程：**
1. 将某专辑状态改为 `confirmed`
2. 普通用户访问专辑详情页 → 看到"提交修改建议"按钮 + `entry_status` 角标
3. 提交建议 → 管理员审核 → 批准/拒绝

**开放状态正常编辑：**
1. `entry_status = 'open'` 的艺人/专辑 → 普通用户仍看到"编辑"按钮

- [ ] **Step 4: Final commit**

```bash
git add -A
git commit -m "feat(music): complete Music module - ArtistCorrection, entry_status gating, admin review"
```
