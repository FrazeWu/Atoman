# Forum Module Completion Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task.

**Goal:** 补全 Forum 模块的三类缺失功能：① 精华帖（Featured）支持——模型字段 + 后端排序 + 管理员标记 + 前端 Tab；② 举报功能——模型 + 后端接收/阈值折叠 + 前端举报按钮；③ 分区申请——用户提交 + 管理员审核 + 前端入口。

**Architecture:**
- 后端已完整：Category CRUD、Topic CRUD（含 pin/close）、Reply 嵌套（ParentReplyID）、Like/Bookmark、Draft、Search
- 前端已完整：ForumHomeView（分类侧边栏 + 排序 Tab latest/top/active/new/bookmarked）、ForumTopicView（回复树、点赞）、ForumNewTopicView、ForumSearchView
- 缺失：Featured 字段/排序、举报模型和端点、分区申请模型和端点

**已完整的部分（无需修改）：**
- `forum_handler.go`: Topic/Reply CRUD, Pin/Close, Like/Bookmark, Draft, Search
- `forum.go` 模型: ForumCategory, ForumTopic, ForumReply, ForumLike, ForumBookmark, ForumDraft
- `ForumHomeView.vue`: 分类导航 + 多排序 Tab
- `ForumTopicView.vue`: 帖子详情 + 回复区 + 点赞
- `ForumNewTopicView.vue`: 发帖 + 草稿

---

## 文件清单

### 修改
| 文件 | 改动 |
|------|------|
| `server/internal/model/forum.go` | `ForumTopic` 添加 `Featured bool`；添加 `ForumReport`、`CategoryRequest` 模型 |
| `server/cmd/start_server/main.go` | AutoMigrate 注册新模型 |
| `server/internal/handlers/forum_handler.go` | 添加 featured toggle、report endpoint、category request endpoint；GetForumTopics 支持 `sort=featured` |
| `web/src/views/forum/ForumHomeView.vue` | 添加"精华"排序 Tab + 分区申请入口 |
| `web/src/views/forum/ForumTopicView.vue` | 添加举报按钮 + admin featured/unfeatured 按钮 |

---

## Task 1：后端模型 — Featured、ForumReport、CategoryRequest

**Files:**
- Modify: `server/internal/model/forum.go`

- [ ] **Step 1: 在 ForumTopic struct 中添加 Featured 字段**

找到 `ForumTopic` struct 中 `Pinned` 字段的下一行，添加：

```go
Featured     bool           `json:"featured" gorm:"default:false"`
```

（位置：`Pinned bool` 和 `Closed bool` 之间）

- [ ] **Step 2: 在文件末尾添加 ForumReport 模型**

```go
// ForumReport represents a user's report on a topic or reply.
// When a topic's report count reaches the auto-collapse threshold, it is soft-hidden.
type ForumReport struct {
	Base
	UserID     uuid.UUID `json:"user_id" gorm:"type:uuid;not null;index"`
	TargetType string    `json:"target_type" gorm:"not null"` // "topic" | "reply"
	TargetID   uuid.UUID `json:"target_id" gorm:"type:uuid;not null;index"`
	Reason     string    `json:"reason" gorm:"not null"` // spam | off-topic | harassment | other
	Note       string    `json:"note" gorm:"type:text"`  // optional user note
}

func (ForumReport) TableName() string { return "forum_reports" }
```

- [ ] **Step 3: 添加 CategoryRequest 模型**

```go
// CategoryRequest represents a user's request to create a new forum category.
// Workflow: user submits → admin reviews → approve (creates category) or reject.
type CategoryRequest struct {
	Base
	UserID      uuid.UUID `json:"user_id" gorm:"type:uuid;not null;index"`
	User        *User     `json:"user,omitempty" gorm:"foreignKey:UserID;references:UUID"`
	Name        string    `json:"name" gorm:"not null"`
	Description string    `json:"description" gorm:"type:text"`
	Reason      string    `json:"reason" gorm:"type:text"` // why this category is needed
	Status      string    `json:"status" gorm:"default:'pending'"`  // pending | approved | rejected
	ReviewedBy  *uuid.UUID `json:"reviewed_by" gorm:"type:uuid"`
	ReviewNote  string    `json:"review_note" gorm:"type:text"`
}

func (CategoryRequest) TableName() string { return "category_requests" }
```

- [ ] **Step 4: Commit**

```bash
cd server && go build ./...
git add server/internal/model/forum.go
git commit -m "feat(forum): add Featured field, ForumReport and CategoryRequest models"
```

---

## Task 2：后端迁移 — 注册新模型

**Files:**
- Modify: `server/cmd/start_server/main.go`

- [ ] **Step 1: 找到 AutoMigrate 中 Forum 相关模型**

```bash
grep -n "ForumTopic\|ForumCategory\|forum" server/cmd/start_server/main.go | head -10
```

- [ ] **Step 2: 在 ForumDraft 或最后一个 forum 模型后添加新模型**

```go
&model.ForumReport{},
&model.CategoryRequest{},
```

- [ ] **Step 3: 编译验证**

```bash
cd server && go build ./...
```

- [ ] **Step 4: Commit**

```bash
git add server/cmd/start_server/main.go
git commit -m "feat(forum): register ForumReport and CategoryRequest in AutoMigrate"
```

---

## Task 3：后端 handler — Featured 功能

**Files:**
- Modify: `server/internal/handlers/forum_handler.go`

**背景：** 添加三个功能：① `GET /api/forum/topics` 支持 `?sort=featured`；② 新增 `POST /api/forum/topics/:id/feature`（管理员标记精华）；③ `DELETE /api/forum/topics/:id/feature`（取消精华）。

- [ ] **Step 1: 更新 GetForumTopics — 支持 sort=featured**

在 `GetForumTopics` handler 中，找到 `sort` 参数处理逻辑（判断 latest/top/active 等），添加：

```go
case "featured":
    query = query.Where("featured = ?", true).Order("created_at DESC")
```

- [ ] **Step 2: 添加 ToggleFeatured handler**

在文件末尾添加：

```go
// FeatureForumTopic marks a topic as featured (admin only).
// Route: POST /api/forum/topics/:id/feature
func FeatureForumTopic(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !isAdmin(c) {
			c.JSON(http.StatusForbidden, gin.H{"error": "admin only"})
			return
		}
		id := c.Param("id")
		if err := db.Model(&model.ForumTopic{}).Where("id = ?", id).Update("featured", true).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "featured"})
	}
}

// UnfeatureForumTopic removes featured status from a topic (admin only).
// Route: DELETE /api/forum/topics/:id/feature
func UnfeatureForumTopic(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !isAdmin(c) {
			c.JSON(http.StatusForbidden, gin.H{"error": "admin only"})
			return
		}
		id := c.Param("id")
		if err := db.Model(&model.ForumTopic{}).Where("id = ?", id).Update("featured", false).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "unfeatured"})
	}
}
```

- [ ] **Step 3: 在 SetupForumRoutes 中注册路由**

在 protected 路由区域中，pin/close 的下方添加：

```go
protected.POST("/topics/:id/feature", FeatureForumTopic(db))
protected.DELETE("/topics/:id/feature", UnfeatureForumTopic(db))
```

- [ ] **Step 4: 确认 isAdmin 函数存在**

```bash
grep -n "func isAdmin\|isAdmin(" server/internal/handlers/forum_handler.go | head -5
```

若不存在，添加：

```go
func isAdmin(c *gin.Context) bool {
	userID, exists := c.Get("userID")
	if !exists {
		return false
	}
	// Check role from context (set by auth middleware)
	// Alternatively read from DB if needed
	role, _ := c.Get("userRole")
	_ = userID
	return role == "admin"
}
```

若 role 不在 context 中，参考 `admin_handler.go` 中的 admin 鉴权方式，保持与现有代码一致。

- [ ] **Step 5: 编译验证**

```bash
cd server && go build ./...
```

- [ ] **Step 6: Commit**

```bash
git add server/internal/handlers/forum_handler.go
git commit -m "feat(forum): add featured sort support and admin feature/unfeature endpoints"
```

---

## Task 4：后端 handler — 举报功能

**Files:**
- Modify: `server/internal/handlers/forum_handler.go`

**背景：** 用户可对帖子或回复举报；当某帖子举报数 ≥ 阈值时，自动将帖子标记为折叠（管理员可手动恢复）。阈值：10 次。

- [ ] **Step 1: 添加 ReportForumContent handler**

```go
// ReportForumContent submits a report for a topic or reply.
// Route: POST /api/forum/report
// Payload: { target_type: "topic"|"reply", target_id: UUID, reason: string, note: string }
// Side effect: if topic report count >= 10, auto-close the topic (soft hide).
func ReportForumContent(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.MustGet("userID").(uuid.UUID)

		var req struct {
			TargetType string `json:"target_type" binding:"required,oneof=topic reply"`
			TargetID   string `json:"target_id" binding:"required"`
			Reason     string `json:"reason" binding:"required,oneof=spam off-topic harassment other"`
			Note       string `json:"note"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		targetUUID, err := uuid.Parse(req.TargetID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid target_id"})
			return
		}

		// Prevent duplicate reports from the same user
		var existing model.ForumReport
		if db.Where("user_id = ? AND target_type = ? AND target_id = ?", userID, req.TargetType, targetUUID).First(&existing).Error == nil {
			c.JSON(http.StatusConflict, gin.H{"error": "already reported"})
			return
		}

		report := model.ForumReport{
			UserID:     userID,
			TargetType: req.TargetType,
			TargetID:   targetUUID,
			Reason:     req.Reason,
			Note:       req.Note,
		}
		if err := db.Create(&report).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create report"})
			return
		}

		// Auto-collapse topic when report count reaches threshold (10)
		const threshold = 10
		if req.TargetType == "topic" {
			var count int64
			db.Model(&model.ForumReport{}).Where("target_type = ? AND target_id = ?", "topic", targetUUID).Count(&count)
			if count >= threshold {
				db.Model(&model.ForumTopic{}).Where("id = ?", targetUUID).Update("closed", true)
			}
		}

		c.JSON(http.StatusCreated, gin.H{"message": "reported"})
	}
}
```

- [ ] **Step 2: 注册路由**

在 protected 路由区域中添加：

```go
protected.POST("/report", ReportForumContent(db))
```

- [ ] **Step 3: 编译验证**

```bash
cd server && go build ./...
```

- [ ] **Step 4: Commit**

```bash
git add server/internal/handlers/forum_handler.go
git commit -m "feat(forum): add report endpoint with auto-collapse threshold"
```

---

## Task 5：后端 handler — 分区申请

**Files:**
- Modify: `server/internal/handlers/forum_handler.go`

- [ ] **Step 1: 添加分区申请相关 handler（3 个）**

```go
// CreateCategoryRequest submits a request to create a new forum category.
// Route: POST /api/forum/category-requests
func CreateCategoryRequest(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.MustGet("userID").(uuid.UUID)
		var req struct {
			Name        string `json:"name" binding:"required"`
			Description string `json:"description"`
			Reason      string `json:"reason"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		cr := model.CategoryRequest{
			UserID:      userID,
			Name:        req.Name,
			Description: req.Description,
			Reason:      req.Reason,
			Status:      "pending",
		}
		if err := db.Create(&cr).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed"})
			return
		}
		c.JSON(http.StatusCreated, gin.H{"data": cr})
	}
}

// GetCategoryRequests lists pending category requests (admin only).
// Route: GET /api/forum/category-requests
func GetCategoryRequests(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !isAdmin(c) {
			c.JSON(http.StatusForbidden, gin.H{"error": "admin only"})
			return
		}
		var requests []model.CategoryRequest
		db.Where("status = ?", "pending").Preload("User").Order("created_at ASC").Find(&requests)
		c.JSON(http.StatusOK, gin.H{"data": requests})
	}
}

// ReviewCategoryRequest approves or rejects a category request (admin only).
// Route: POST /api/forum/category-requests/:id/review
// Payload: { action: "approve"|"reject", review_note: string, color: string }
func ReviewCategoryRequest(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !isAdmin(c) {
			c.JSON(http.StatusForbidden, gin.H{"error": "admin only"})
			return
		}
		id := c.Param("id")
		var req struct {
			Action     string `json:"action" binding:"required,oneof=approve reject"`
			ReviewNote string `json:"review_note"`
			Color      string `json:"color"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		adminID := c.MustGet("userID").(uuid.UUID)
		var cr model.CategoryRequest
		if err := db.First(&cr, "id = ?", id).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
			return
		}

		cr.Status = req.Action + "d" // "approved" | "rejected"
		cr.ReviewedBy = &adminID
		cr.ReviewNote = req.ReviewNote
		db.Save(&cr)

		// If approved, create the category
		if req.Action == "approve" {
			color := req.Color
			if color == "" {
				color = "#6366f1"
			}
			cat := model.ForumCategory{
				Name:        cr.Name,
				Description: cr.Description,
				Color:       color,
			}
			if err := db.Create(&cat).Error; err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "approved but failed to create category"})
				return
			}
			c.JSON(http.StatusOK, gin.H{"data": cr, "category": cat})
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": cr})
	}
}
```

- [ ] **Step 2: 注册路由**

在 public 路由区域（protected 外）不需要登录的：
在 protected 内添加：

```go
protected.POST("/category-requests", CreateCategoryRequest(db))
```

在 admin 路由或 protected + isAdmin 校验中添加（可以共用 protected，因为 handler 内部做 isAdmin 检查）：

```go
protected.GET("/category-requests", GetCategoryRequests(db))
protected.POST("/category-requests/:id/review", ReviewCategoryRequest(db))
```

- [ ] **Step 3: 编译验证**

```bash
cd server && go build ./...
```

- [ ] **Step 4: Commit**

```bash
git add server/internal/handlers/forum_handler.go
git commit -m "feat(forum): add category request submit/review endpoints"
```

---

## Task 6：前端 — ForumHomeView.vue 精华 Tab + 分区申请入口

**Files:**
- Modify: `web/src/views/forum/ForumHomeView.vue`

- [ ] **Step 1: 添加"精华"排序 Tab**

找到 `tabOptions` 对象（`const tabOptions: Record<TabKey, string> = { latest: '最新', top: '最热', active: '最活跃', new: '新帖', bookmarked: '收藏' }`），更新 `TabKey` 类型并添加 featured：

```typescript
type TabKey = 'latest' | 'top' | 'active' | 'new' | 'bookmarked' | 'featured'

const tabOptions: Record<TabKey, string> = {
  latest: '最新',
  top: '最热',
  active: '最活跃',
  new: '新帖',
  bookmarked: '收藏',
  featured: '精华',
}

const sortMap: Record<TabKey, string> = {
  latest: 'latest',
  top: 'top',
  active: 'active',
  new: 'latest',
  bookmarked: 'latest',
  featured: 'featured',
}
```

- [ ] **Step 2: 添加分区申请入口**

在左侧分类侧边栏底部，"管理员新建分区"按钮下方（或末尾），添加登录用户可见的"申请新分区"入口：

```vue
<button
  v-if="authStore.isAuthenticated"
  class="sidebar-item"
  style="font-size:.7rem;color:var(--a-color-muted)"
  @click="showRequestModal = true"
>+ 申请新分区</button>
```

- [ ] **Step 3: 添加分区申请 Modal**

在模板末尾添加：

```vue
<!-- Category Request Modal -->
<div v-if="showRequestModal" class="a-modal-backdrop" @click.self="showRequestModal = false">
  <div class="a-modal" style="max-width:32rem">
    <h2 class="a-modal-title" style="font-size:1rem;font-weight:900;margin-bottom:1rem">申请新分区</h2>
    <form @submit.prevent="submitCategoryRequest">
      <div class="a-field">
        <label class="a-label">分区名称 <span style="color:var(--a-color-danger)">*</span></label>
        <input v-model="requestForm.name" class="a-input" required placeholder="例：科技、生活方式……" />
      </div>
      <div class="a-field" style="margin-top:.5rem">
        <label class="a-label">分区描述</label>
        <textarea v-model="requestForm.description" class="a-textarea" rows="2" />
      </div>
      <div class="a-field" style="margin-top:.5rem">
        <label class="a-label">申请理由</label>
        <textarea v-model="requestForm.reason" class="a-textarea" rows="2" placeholder="为什么需要这个分区？" />
      </div>
      <div style="display:flex;gap:.5rem;margin-top:1rem;justify-content:flex-end">
        <button type="button" class="a-btn a-btn-ghost" @click="showRequestModal = false">取消</button>
        <button type="submit" class="a-btn a-btn-primary" :disabled="requestSubmitting">
          {{ requestSubmitting ? '提交中…' : '提交申请' }}
        </button>
      </div>
    </form>
  </div>
</div>
```

在 `<script setup>` 中添加：

```typescript
const showRequestModal = ref(false)
const requestSubmitting = ref(false)
const requestForm = reactive({ name: '', description: '', reason: '' })

async function submitCategoryRequest() {
  requestSubmitting.value = true
  try {
    await api.post('/api/forum/category-requests', { ...requestForm })
    showRequestModal.value = false
    requestForm.name = ''
    requestForm.description = ''
    requestForm.reason = ''
    alert('申请已提交，管理员审核后会创建该分区。')
  } catch (e: any) {
    alert('提交失败：' + (e?.message || '未知错误'))
  } finally {
    requestSubmitting.value = false
  }
}
```

- [ ] **Step 4: 类型检查**

```bash
cd web && bun run type-check 2>&1 | tail -5
```

- [ ] **Step 5: Commit**

```bash
git add web/src/views/forum/ForumHomeView.vue
git commit -m "feat(forum): add featured sort tab and category request modal to ForumHomeView"
```

---

## Task 7：前端 — ForumTopicView.vue 举报按钮 + admin 精华标记

**Files:**
- Modify: `web/src/views/forum/ForumTopicView.vue`

- [ ] **Step 1: 添加举报按钮（帖子级别）**

在帖子操作区（like/bookmark 按钮旁），添加举报入口（登录用户可见，非作者）：

```vue
<button
  v-if="authStore.isAuthenticated && topic.user_id !== authStore.user?.id"
  class="action-btn action-report"
  @click="showReportModal = true"
  title="举报"
>举报</button>
```

- [ ] **Step 2: 添加举报 Modal**

```vue
<div v-if="showReportModal" class="a-modal-backdrop" @click.self="showReportModal = false">
  <div class="a-modal" style="max-width:28rem">
    <h2 class="a-modal-title" style="font-size:1rem;font-weight:900;margin-bottom:1rem">举报内容</h2>
    <form @submit.prevent="submitReport">
      <div class="a-field">
        <label class="a-label">举报原因</label>
        <select v-model="reportReason" class="a-select" required>
          <option value="">请选择……</option>
          <option value="spam">垃圾信息</option>
          <option value="off-topic">偏题</option>
          <option value="harassment">骚扰/攻击</option>
          <option value="other">其他</option>
        </select>
      </div>
      <div class="a-field" style="margin-top:.5rem">
        <label class="a-label">补充说明（可选）</label>
        <textarea v-model="reportNote" class="a-textarea" rows="2" />
      </div>
      <div style="display:flex;gap:.5rem;margin-top:1rem;justify-content:flex-end">
        <button type="button" class="a-btn a-btn-ghost" @click="showReportModal = false">取消</button>
        <button type="submit" class="a-btn a-btn-danger" :disabled="reportSubmitting">
          {{ reportSubmitting ? '提交中…' : '提交举报' }}
        </button>
      </div>
    </form>
  </div>
</div>
```

- [ ] **Step 3: 添加管理员精华标记按钮**

在管理员操作区（pin/close 按钮旁），添加：

```vue
<button
  v-if="authStore.user?.role === 'admin'"
  class="action-btn"
  :class="topic.featured ? 'action-active' : ''"
  @click="toggleFeatured"
>{{ topic.featured ? '取消精华' : '设为精华' }}</button>
```

- [ ] **Step 4: 在 script setup 中添加对应逻辑**

```typescript
const showReportModal = ref(false)
const reportReason = ref('')
const reportNote = ref('')
const reportSubmitting = ref(false)

async function submitReport() {
  if (!reportReason.value) return
  reportSubmitting.value = true
  try {
    await api.post('/api/forum/report', {
      target_type: 'topic',
      target_id: topic.value?.id,
      reason: reportReason.value,
      note: reportNote.value,
    })
    showReportModal.value = false
    reportReason.value = ''
    reportNote.value = ''
    alert('举报已提交')
  } catch (e: any) {
    const msg = e?.response?.data?.error || e?.message || '提交失败'
    alert(msg === 'already reported' ? '你已举报过此内容' : `举报失败：${msg}`)
  } finally {
    reportSubmitting.value = false
  }
}

async function toggleFeatured() {
  if (!topic.value) return
  const method = topic.value.featured ? 'delete' : 'post'
  await api[method](`/api/forum/topics/${topic.value.id}/feature`)
  topic.value.featured = !topic.value.featured
}
```

- [ ] **Step 5: 类型检查**

```bash
cd web && bun run type-check 2>&1 | tail -5
```

- [ ] **Step 6: Commit**

```bash
git add web/src/views/forum/ForumTopicView.vue
git commit -m "feat(forum): add report modal and admin featured toggle to ForumTopicView"
```

---

## Task 8：后端 — SiteSettings 模型 + 迁移注册

**Files:**
- Create: `server/internal/model/site_settings.go`
- Modify: `server/cmd/start_server/main.go`
- Modify: `server/internal/service/forum_migrate.go`

**背景：** 需要存储可配置的站点参数（如 `forum.solved_auto_threshold`），用于 Solved 自动标记阈值。

- [ ] **Step 1: 创建 SiteSettings 模型**

创建 `server/internal/model/site_settings.go`：

```go
package model

import "time"

// SiteSetting stores administrator-configurable key/value parameters.
// Keys are namespaced by module: "forum.xxx", "timeline.xxx", etc.
type SiteSetting struct {
	Key         string    `json:"key" gorm:"primaryKey"`
	Value       string    `json:"value" gorm:"not null"`
	Description string    `json:"description"`
	UpdatedAt   time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

func (SiteSetting) TableName() string { return "site_settings" }
```

- [ ] **Step 2: 在 AutoMigrate 中注册新模型**

找到 `server/cmd/start_server/main.go` 中 `AutoMigrate` 调用，在 forum 模型列表后添加：

```go
&model.ForumReport{},
&model.CategoryRequest{},
&model.SiteSetting{},
```

- [ ] **Step 3: 在 forum_migrate.go 的 PostgreSQL 分支中补充新列迁移**

在 `ALTER TABLE forum_topics ADD COLUMN IF NOT EXISTS tags TEXT` 之后添加：

```go
db.Exec(`ALTER TABLE forum_topics ADD COLUMN IF NOT EXISTS featured BOOLEAN DEFAULT FALSE`)
db.Exec(`ALTER TABLE forum_topics ADD COLUMN IF NOT EXISTS is_solved BOOLEAN DEFAULT FALSE`)
db.Exec(`ALTER TABLE forum_topics ADD COLUMN IF NOT EXISTS solved_reply_id UUID`)
db.Exec(`ALTER TABLE forum_replies ADD COLUMN IF NOT EXISTS depth INT DEFAULT 0`)
db.Exec(`ALTER TABLE forum_replies ADD COLUMN IF NOT EXISTS is_solved BOOLEAN DEFAULT FALSE`)
```

- [ ] **Step 4: 种子数据（幂等）**

在 `cmd/start_server/main.go` 的 AutoMigrate 之后添加：

```go
db.Exec(`INSERT INTO site_settings (key, value, description, updated_at)
VALUES ('forum.solved_auto_threshold', '10', '回复点赞数达到该值时自动标记为解决方案', NOW())
ON CONFLICT (key) DO NOTHING`)
```

- [ ] **Step 5: 更新 ForumTopic 模型字段**

在 `server/internal/model/forum.go` 的 `ForumTopic` struct 中，`Pinned bool` 后添加：

```go
Featured      bool       `json:"featured" gorm:"default:false"`
IsSolved      bool       `json:"is_solved" gorm:"default:false"`
SolvedReplyID *uuid.UUID `json:"solved_reply_id" gorm:"type:uuid"`
```

在 `ForumReply` struct 的 `FloorNumber int` 后添加：

```go
Depth    int  `json:"depth" gorm:"default:0"`
IsSolved bool `json:"is_solved" gorm:"default:false"`
```

- [ ] **Step 6: 编译验证**

```bash
cd /Users/fafa/Documents/projects/Atoman/server && go build ./...
```

- [ ] **Step 7: Commit**

```bash
git add server/internal/model/site_settings.go server/internal/model/forum.go \
        server/cmd/start_server/main.go server/internal/service/forum_migrate.go
git commit -m "feat(forum): add SiteSettings model, IsSolved/Featured/Depth fields, DB migrations"
```

---

## Task 9：后端 handler — Solved 标记 + 深度校验

**Files:**
- Modify: `server/internal/handlers/forum_handler.go`

**背景：**
- `POST /api/forum/replies/:id/solve`：楼主或管理员标记该回复为解决方案；同时检查点赞数是否达到阈值
- `DELETE /api/forum/replies/:id/solve`：楼主或管理员取消标记
- `CreateForumReply` 中添加 depth 校验：若 parent 的 depth=1，则返回 400

- [ ] **Step 1: 读取当前 CreateForumReply handler**

```bash
grep -n "CreateForumReply\|ParentReplyID\|parent_reply_id" /Users/fafa/Documents/projects/Atoman/server/internal/handlers/forum_handler.go | head -10
```

- [ ] **Step 2: 在 CreateForumReply 中添加 Depth 校验和赋值**

找到 CreateForumReply handler 中创建 `ForumReply` 对象的位置，在创建前添加：

```go
var depth int
if req.ParentReplyID != nil {
    var parent model.ForumReply
    if err := db.First(&parent, "id = ?", req.ParentReplyID).Error; err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "parent reply not found"})
        return
    }
    if parent.Depth >= 1 {
        c.JSON(http.StatusBadRequest, gin.H{"error": "nesting limit reached: max 2 levels"})
        return
    }
    depth = parent.Depth + 1
}
```

并在 `ForumReply` struct 初始化时加上 `Depth: depth`。

- [ ] **Step 3: 添加 MarkReplySolved handler**

```go
// MarkReplySolved marks a reply as the solution for its topic.
// Only the topic author or an admin can call this.
// A topic can have only one solved reply; marking a new one replaces the previous.
// Route: POST /api/forum/replies/:id/solve
func MarkReplySolved(db *gorm.DB) gin.HandlerFunc {
    return func(c *gin.Context) {
        replyID := c.Param("id")
        userID := c.MustGet("userID").(uuid.UUID)

        var reply model.ForumReply
        if err := db.Preload("Topic").First(&reply, "id = ?", replyID).Error; err != nil {
            c.JSON(http.StatusNotFound, gin.H{"error": "reply not found"})
            return
        }

        // Only topic author or admin
        isAuthor := reply.Topic != nil && reply.Topic.UserID == userID
        if !isAuthor && !isAdmin(c) {
            c.JSON(http.StatusForbidden, gin.H{"error": "only topic author or admin"})
            return
        }

        // Clear previous solved reply on this topic
        db.Model(&model.ForumReply{}).Where("topic_id = ? AND is_solved = ?", reply.TopicID, true).Update("is_solved", false)

        // Mark this reply as solved
        db.Model(&reply).Update("is_solved", true)
        replyUUID := reply.ID
        db.Model(&model.ForumTopic{}).Where("id = ?", reply.TopicID).Updates(map[string]interface{}{
            "is_solved":       true,
            "solved_reply_id": replyUUID,
        })

        c.JSON(http.StatusOK, gin.H{"message": "marked as solution"})
    }
}

// UnmarkReplySolved removes the solution mark from a reply.
// Route: DELETE /api/forum/replies/:id/solve
func UnmarkReplySolved(db *gorm.DB) gin.HandlerFunc {
    return func(c *gin.Context) {
        replyID := c.Param("id")
        userID := c.MustGet("userID").(uuid.UUID)

        var reply model.ForumReply
        if err := db.Preload("Topic").First(&reply, "id = ?", replyID).Error; err != nil {
            c.JSON(http.StatusNotFound, gin.H{"error": "reply not found"})
            return
        }

        isAuthor := reply.Topic != nil && reply.Topic.UserID == userID
        if !isAuthor && !isAdmin(c) {
            c.JSON(http.StatusForbidden, gin.H{"error": "only topic author or admin"})
            return
        }

        db.Model(&reply).Update("is_solved", false)
        db.Model(&model.ForumTopic{}).Where("id = ? AND solved_reply_id = ?", reply.TopicID, reply.ID).Updates(map[string]interface{}{
            "is_solved":       false,
            "solved_reply_id": nil,
        })

        c.JSON(http.StatusOK, gin.H{"message": "solution mark removed"})
    }
}
```

- [ ] **Step 4: 在 ToggleForumTopicLike / ToggleReplyLike 中触发自动 Solved**

找到点赞 handler（`ToggleForumTopicLike` 或类似函数），在点赞计数更新后，若是 reply 被点赞，添加自动触发逻辑：

```go
// Auto-mark as solved if like count reaches threshold
var threshold int
var setting model.SiteSetting
if db.First(&setting, "key = ?", "forum.solved_auto_threshold").Error == nil {
    fmt.Sscanf(setting.Value, "%d", &threshold)
}
if threshold > 0 {
    var updatedReply model.ForumReply
    db.First(&updatedReply, "id = ?", replyID)
    if updatedReply.LikeCount >= threshold && !updatedReply.IsSolved {
        db.Model(&updatedReply).Update("is_solved", true)
        replyUUID := updatedReply.ID
        db.Model(&model.ForumTopic{}).Where("id = ? AND is_solved = ?", updatedReply.TopicID, false).Updates(map[string]interface{}{
            "is_solved":       true,
            "solved_reply_id": replyUUID,
        })
    }
}
```

- [ ] **Step 5: 注册路由**

在 protected 路由区域添加：

```go
protected.POST("/replies/:id/solve", MarkReplySolved(db))
protected.DELETE("/replies/:id/solve", UnmarkReplySolved(db))
```

- [ ] **Step 6: 添加 SiteSettings admin 接口**

```go
// GetSiteSettings returns all site settings (admin only).
// Route: GET /api/admin/settings
func GetSiteSettings(db *gorm.DB) gin.HandlerFunc {
    return func(c *gin.Context) {
        var settings []model.SiteSetting
        db.Order("key ASC").Find(&settings)
        c.JSON(http.StatusOK, gin.H{"data": settings})
    }
}

// UpdateSiteSetting updates a site setting value (admin only).
// Route: PUT /api/admin/settings/:key
func UpdateSiteSetting(db *gorm.DB) gin.HandlerFunc {
    return func(c *gin.Context) {
        key := c.Param("key")
        var req struct {
            Value string `json:"value" binding:"required"`
        }
        if err := c.ShouldBindJSON(&req); err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
            return
        }
        if err := db.Model(&model.SiteSetting{}).Where("key = ?", key).Update("value", req.Value).Error; err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "failed"})
            return
        }
        c.JSON(http.StatusOK, gin.H{"message": "updated"})
    }
}
```

在 `admin_handler.go` 的路由注册处（或 `forum_handler.go` 末尾）添加，确认挂在 `/api/admin/` 前缀下。

- [ ] **Step 7: 编译验证**

```bash
cd /Users/fafa/Documents/projects/Atoman/server && go build ./...
```

- [ ] **Step 8: Commit**

```bash
git add server/internal/handlers/forum_handler.go server/internal/handlers/admin_handler.go
git commit -m "feat(forum): add Solved endpoints, depth enforcement, SiteSettings admin API"
```

---

## Task 10：前端 — ForumNewTopicView Tags 输入

**Files:**
- Modify: `web/src/views/forum/ForumNewTopicView.vue`

**背景：** 发帖时允许输入自由标签，以 chip 形式显示，回车或逗号分隔确认。

- [ ] **Step 1: 读取当前 ForumNewTopicView 表单**

```bash
grep -n "content\|title\|form\|submit" /Users/fafa/Documents/projects/Atoman/web/src/views/forum/ForumNewTopicView.vue | head -20
```

- [ ] **Step 2: 在 form 数据中添加 tags 字段**

在 reactive form 对象中添加：

```typescript
tags: [] as string[],
tagInput: '',
```

- [ ] **Step 3: 在模板 category/title 字段下方添加 Tags 输入区**

```vue
<!-- Tags -->
<div class="a-field">
  <label class="a-label">标签（可选）</label>
  <div class="tag-input-wrap" style="display:flex;flex-wrap:wrap;gap:.25rem;align-items:center;border:1px solid var(--a-border-color);border-radius:.375rem;padding:.375rem .5rem;min-height:2.25rem">
    <span
      v-for="tag in form.tags"
      :key="tag"
      style="display:inline-flex;align-items:center;gap:.25rem;background:var(--a-color-surface-2);border-radius:.25rem;padding:.1rem .4rem;font-size:.75rem"
    >
      {{ tag }}
      <button type="button" style="line-height:1;background:none;border:none;cursor:pointer;padding:0;color:var(--a-color-muted)" @click="removeTag(tag)">×</button>
    </span>
    <input
      v-model="form.tagInput"
      type="text"
      placeholder="输入标签后按回车"
      style="border:none;outline:none;background:transparent;font-size:.875rem;min-width:8rem;flex:1"
      @keydown.enter.prevent="addTag"
      @keydown.comma.prevent="addTag"
    />
  </div>
  <p style="font-size:.7rem;color:var(--a-color-muted);margin:.25rem 0 0">按回车或逗号分隔，最多 5 个标签</p>
</div>
```

- [ ] **Step 4: 添加 addTag / removeTag 函数**

```typescript
function addTag() {
  const tag = form.tagInput.replace(/,/g, '').trim()
  if (!tag || form.tags.includes(tag) || form.tags.length >= 5) {
    form.tagInput = ''
    return
  }
  form.tags.push(tag)
  form.tagInput = ''
}

function removeTag(tag: string) {
  form.tags = form.tags.filter(t => t !== tag)
}
```

- [ ] **Step 5: 确认提交时 tags 包含在 payload 中**

找到提交函数，确认 `tags: form.tags` 包含在 POST body 中。

- [ ] **Step 6: 类型检查**

```bash
cd /Users/fafa/Documents/projects/Atoman/web && bun run type-check 2>&1 | tail -5
```

- [ ] **Step 7: Commit**

```bash
git add web/src/views/forum/ForumNewTopicView.vue
git commit -m "feat(forum): add tags input to ForumNewTopicView"
```

---

## Task 11：前端 — ForumHomeView Tags 过滤 + ForumTopicView Solved UI + 子回复折叠

**Files:**
- Modify: `web/src/views/forum/ForumHomeView.vue`
- Modify: `web/src/views/forum/ForumTopicView.vue`

### Part A：ForumHomeView — Tags 过滤

- [ ] **Step 1: 添加 activeTag 状态**

```typescript
const activeTag = ref<string | null>(null)
```

- [ ] **Step 2: 帖子卡片上的标签 chip 可点击**

在帖子卡片中，标签 chip 绑定点击：

```vue
<span
  v-for="tag in topic.tags"
  :key="tag"
  class="topic-tag"
  :class="{ active: activeTag === tag }"
  style="cursor:pointer;font-size:.7rem;padding:.1rem .35rem;border-radius:.25rem;background:var(--a-color-surface-2)"
  @click.stop="activeTag = activeTag === tag ? null : tag"
>{{ tag }}</span>
```

- [ ] **Step 3: 在 fetchTopics 请求中附加 tag 参数**

```typescript
if (activeTag.value) params.tag = activeTag.value
```

- [ ] **Step 4: 显示当前过滤 tag 的 pill + 清除按钮**

在 Tab 栏下方条件渲染：

```vue
<div v-if="activeTag" style="display:flex;align-items:center;gap:.5rem;font-size:.8rem;padding:.25rem 0">
  <span>标签筛选：</span>
  <span style="background:var(--a-color-primary);color:#fff;padding:.1rem .5rem;border-radius:.25rem">{{ activeTag }}</span>
  <button class="a-btn a-btn-ghost" style="padding:.1rem .35rem;font-size:.75rem" @click="activeTag = null">清除</button>
</div>
```

### Part B：ForumTopicView — Solved UI

- [ ] **Step 5: 在回复列表中显示 Solved 徽章**

在回复卡片中，若 `reply.is_solved`，显示绿色徽章：

```vue
<span v-if="reply.is_solved" style="font-size:.7rem;background:#10b981;color:#fff;padding:.1rem .4rem;border-radius:.25rem;margin-left:.5rem">✓ 解决方案</span>
```

- [ ] **Step 6: 楼主/管理员可点击"标为解决"按钮**

在回复操作区添加：

```vue
<button
  v-if="canMarkSolved(reply)"
  class="action-btn"
  :class="{ 'action-active': reply.is_solved }"
  @click="toggleSolved(reply)"
>{{ reply.is_solved ? '取消解决' : '标为解决' }}</button>
```

在 `<script setup>` 添加：

```typescript
function canMarkSolved(reply: ForumReply): boolean {
  if (!authStore.isAuthenticated) return false
  const isTopicAuthor = topic.value?.user_id === authStore.user?.id
  const isAdminUser = authStore.user?.role === 'admin'
  return isTopicAuthor || isAdminUser
}

async function toggleSolved(reply: ForumReply) {
  const method = reply.is_solved ? 'delete' : 'post'
  await api[method](`/api/forum/replies/${reply.id}/solve`)
  reply.is_solved = !reply.is_solved
  if (topic.value) topic.value.is_solved = !reply.is_solved ? false : true
}
```

- [ ] **Step 7: 帖子标题旁显示"已解决"pill**

在帖子标题旁添加：

```vue
<span v-if="topic.is_solved" style="font-size:.75rem;background:#10b981;color:#fff;padding:.1rem .5rem;border-radius:.25rem;margin-left:.5rem">已解决</span>
```

### Part C：ForumTopicView — 子回复折叠（默认显示前 2 条）

- [ ] **Step 8: 为顶层回复添加子回复折叠状态**

在 `<script setup>` 添加展开状态 map：

```typescript
const expandedReplies = reactive<Record<string, boolean>>({})

function isExpanded(parentId: string): boolean {
  return !!expandedReplies[parentId]
}

function toggleExpand(parentId: string) {
  expandedReplies[parentId] = !expandedReplies[parentId]
}

function getSubReplies(parentId: string): ForumReply[] {
  return replies.value.filter(r => r.parent_reply_id === parentId)
}
```

- [ ] **Step 9: 修改回复渲染逻辑**

子回复列表默认只显示前 2 条，其余折叠：

```vue
<!-- 子回复区域 -->
<template v-if="getSubReplies(reply.id).length > 0">
  <div
    v-for="sub in (isExpanded(reply.id) ? getSubReplies(reply.id) : getSubReplies(reply.id).slice(0, 2))"
    :key="sub.id"
    class="sub-reply"
    style="margin-left:2rem;border-left:2px solid var(--a-border-color);padding-left:.75rem;margin-top:.5rem"
  >
    <!-- 子回复内容（复用已有结构） -->
  </div>
  <button
    v-if="getSubReplies(reply.id).length > 2"
    class="a-btn a-btn-ghost"
    style="margin-left:2rem;font-size:.75rem;padding:.2rem .5rem;margin-top:.25rem"
    @click="toggleExpand(reply.id)"
  >
    {{ isExpanded(reply.id) ? '收起' : `展开全部 ${getSubReplies(reply.id).length} 条回复` }}
  </button>
</template>
```

- [ ] **Step 10: 类型检查**

```bash
cd /Users/fafa/Documents/projects/Atoman/web && bun run type-check 2>&1 | tail -5
```

- [ ] **Step 11: Commit**

```bash
git add web/src/views/forum/ForumHomeView.vue web/src/views/forum/ForumTopicView.vue
git commit -m "feat(forum): add tags filter, solved UI, and sub-reply collapse"
```

---

## Task 12：验收

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

1. `/forum` → Tab 栏中出现"精华"Tab，点击后只显示 featured=true 的帖子
2. 以管理员身份进入帖子详情页 → 出现"设为精华"按钮 → 点击 → 帖子在精华 Tab 中出现
3. 登录用户进入他人帖子详情 → 出现"举报"按钮 → 选择原因提交 → 201 成功
4. 同一用户对同一内容再次举报 → 提示"已举报过此内容"
5. `/forum` → 点击"申请新分区" → 填写表单提交 → 201 成功
6. 管理员访问 `/api/forum/category-requests`（或管理界面）→ 审批 → 分区自动创建

- [ ] **Step 4: Final commit**

```bash
git add -A
git commit -m "feat(forum): complete Forum module - featured, report, category request"
```
