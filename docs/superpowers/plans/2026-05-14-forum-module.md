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

## Task 8：验收

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
