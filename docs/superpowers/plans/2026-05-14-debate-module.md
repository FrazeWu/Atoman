# Debate Module Completion Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task.

**Goal:** 补全 Debate 模块的两类缺失功能：① 证据来源独立字段（Evidence 类型论点的 source_url / source_title / source_excerpt）；② 管理员折叠/锁定单个论点（超出讨论边界时管理员可折叠）。

**Architecture:**
- 后端已完整：Debate CRUD、Argument 树（support/oppose/neutral/evidence/question/counter 类型）、论点投票、论点相互引用（many2many）、议题结论（conclude/reopen）、投票结论（vote to conclude）
- 前端已完整：DebateHomeView（议题列表）、DebateTopicView（论点树渲染 + ArgumentNode 组件 + 投票 + 总结区 + admin conclude/reopen）
- 缺失 ① `Argument` 模型无证据来源字段（规格：`source_url` + `source_title` + `source_excerpt` 必须展示）
- 缺失 ② `Argument` 无管理员折叠状态；后端无 fold/unfold endpoint；ArgumentNode 无折叠操作按钮

**已完整的部分（无需修改）：**
- `debate_handler.go`: 所有路由和 handler
- `debate.go` 模型: Debate, Argument, DebateVote, ArgumentType 常量
- `DebateHomeView.vue`, `DebateTopicView.vue`, `ArgumentNode.vue`

---

## 文件清单

### 修改
| 文件 | 改动 |
|------|------|
| `server/internal/model/debate.go` | `Argument` struct 添加 `SourceURL`、`SourceTitle`、`SourceExcerpt` 字段，`IsFolded bool` 字段 |
| `server/internal/handlers/debate_handler.go` | `CreateArgument` / `UpdateArgument` 处理新 source 字段；添加 `FoldArgument` / `UnfoldArgument` handler + 路由 |
| `web/src/views/debate/DebateTopicView.vue` | 创建/编辑论点时显示 evidence 类型的 source 输入字段 |
| `web/src/components/debate/ArgumentNode.vue` | 渲染 evidence source 信息；admin 折叠/展开按钮 |
| `web/src/types.ts` | `Argument` 类型添加新字段 |

---

## Task 1：后端模型 — Argument 添加 source 和 IsFolded 字段

**Files:**
- Modify: `server/internal/model/debate.go`

- [ ] **Step 1: 读取 Argument struct 当前最后一个字段位置**

```bash
grep -n "IsConcluded\|Conclusion\|CreatedAt\|UpdatedAt" server/internal/model/debate.go | tail -10
```

- [ ] **Step 2: 在 Argument struct 中添加新字段**

找到 `IsConcluded` 和 `Conclusion` 字段区域，在 `Conclusion string` 后、`CreatedAt` 前添加：

```go
// Evidence source fields (only used when ArgumentType == "evidence")
SourceURL     string `json:"source_url" gorm:"type:varchar(2048);default:''"`
SourceTitle   string `json:"source_title" gorm:"type:varchar(512);default:''"`
SourceExcerpt string `json:"source_excerpt" gorm:"type:text;default:''"`

// Admin moderation
IsFolded bool   `json:"is_folded" gorm:"default:false"`
FoldNote  string `json:"fold_note" gorm:"type:text;default:''"` // admin note for why folded
```

- [ ] **Step 3: 编译验证**

```bash
cd server && go build ./...
```

- [ ] **Step 4: Commit**

```bash
git add server/internal/model/debate.go
git commit -m "feat(debate): add source fields and IsFolded to Argument model"
```

---

## Task 2：后端迁移 — AutoMigrate 自动处理新字段

GORM AutoMigrate 会自动添加缺失列，无需显式迁移文件。只需验证服务器启动时迁移成功。

- [ ] **Step 1: 确认 Argument 在 AutoMigrate 列表中**

```bash
grep -n "Argument\|debate" server/cmd/start_server/main.go | head -10
```

期望：`&model.Argument{}` 在 AutoMigrate 列表中。若不在则添加。

- [ ] **Step 2: Commit（若有变更）**

```bash
git add server/cmd/start_server/main.go
git commit -m "feat(debate): ensure Argument in AutoMigrate"
```

---

## Task 3：后端 handler — CreateArgument / UpdateArgument 处理 source 字段

**Files:**
- Modify: `server/internal/handlers/debate_handler.go`

- [ ] **Step 1: 读取 CreateArgument 的 input struct**

```bash
grep -n "CreateArgumentInput\|struct {" server/internal/handlers/debate_handler.go | head -20
```

读取 `CreateArgument` handler 的 input struct 完整定义（约 415–465 行）。

- [ ] **Step 2: 在 CreateArgument input struct 中添加 source 字段**

找到类似：

```go
var input struct {
    Content      string `json:"content" binding:"required"`
    ArgumentType string `json:"argument_type" binding:"required"`
    ParentID     string `json:"parent_id"`
}
```

添加 source 字段（optional，仅 evidence 类型使用）：

```go
var input struct {
    Content       string `json:"content" binding:"required"`
    ArgumentType  string `json:"argument_type" binding:"required"`
    ParentID      string `json:"parent_id"`
    SourceURL     string `json:"source_url"`
    SourceTitle   string `json:"source_title"`
    SourceExcerpt string `json:"source_excerpt"`
}
```

- [ ] **Step 3: 在创建 Argument 的赋值语句中添加新字段**

在 `argument := model.Argument{...}` 赋值块中添加：

```go
SourceURL:     input.SourceURL,
SourceTitle:   input.SourceTitle,
SourceExcerpt: input.SourceExcerpt,
```

- [ ] **Step 4: 同样更新 UpdateArgument**

找到 `UpdateArgument` handler 中的 input struct 和 db.Model 更新语句，添加：

```go
// In input struct:
SourceURL     string `json:"source_url"`
SourceTitle   string `json:"source_title"`
SourceExcerpt string `json:"source_excerpt"`

// In updates map or direct assignment:
"source_url":     input.SourceURL,
"source_title":   input.SourceTitle,
"source_excerpt": input.SourceExcerpt,
```

- [ ] **Step 5: 添加 FoldArgument / UnfoldArgument handler**

在文件末尾添加：

```go
// FoldArgument hides an argument from display (admin only).
// Route: POST /api/debate/arguments/:id/fold
func FoldArgument(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		var req struct {
			FoldNote string `json:"fold_note"`
		}
		c.ShouldBindJSON(&req)

		if err := db.Model(&model.Argument{}).Where("id = ?", id).Updates(map[string]interface{}{
			"is_folded": true,
			"fold_note": req.FoldNote,
		}).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fold"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "folded"})
	}
}

// UnfoldArgument makes a previously folded argument visible again (admin only).
// Route: DELETE /api/debate/arguments/:id/fold
func UnfoldArgument(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		if err := db.Model(&model.Argument{}).Where("id = ?", id).Updates(map[string]interface{}{
			"is_folded": false,
			"fold_note": "",
		}).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to unfold"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "unfolded"})
	}
}
```

- [ ] **Step 6: 注册 fold/unfold 路由**

在 `SetupDebateRoutes` 的 protected 区域，vote 路由的下方添加：

```go
protected.POST("/arguments/:id/fold", middleware.RequireAuth(), FoldArgument(db))
protected.DELETE("/arguments/:id/fold", middleware.RequireAuth(), UnfoldArgument(db))
```

**注：** fold/unfold 在 handler 内部无 admin 检查，因此路由层也应做鉴权，或在 handler 中加管理员检查。推荐在 handler 中加：

```go
// At top of FoldArgument handler body:
// Admin check (reuse pattern from existing handlers)
userID := c.MustGet("userID").(uuid.UUID)
var user model.User
if db.First(&user, "uuid = ?", userID).Error != nil || user.Role != "admin" {
    c.JSON(http.StatusForbidden, gin.H{"error": "admin only"})
    return
}
```

- [ ] **Step 7: 编译验证**

```bash
cd server && go build ./...
```

- [ ] **Step 8: Commit**

```bash
git add server/internal/handlers/debate_handler.go
git commit -m "feat(debate): handle source fields in CreateArgument/UpdateArgument; add fold/unfold endpoints"
```

---

## Task 4：前端类型 — Argument 添加新字段

**Files:**
- Modify: `web/src/types.ts`

- [ ] **Step 1: 找到 Argument 类型定义**

```bash
grep -n "interface Argument\|type Argument" web/src/types.ts
```

- [ ] **Step 2: 添加新字段**

在 `Argument` interface 中添加（在 `is_concluded` 等字段后）：

```typescript
source_url?: string
source_title?: string
source_excerpt?: string
is_folded?: boolean
fold_note?: string
```

- [ ] **Step 3: 类型检查**

```bash
cd web && bun run type-check 2>&1 | tail -5
```

- [ ] **Step 4: Commit**

```bash
git add web/src/types.ts
git commit -m "feat(debate): add source and is_folded fields to Argument type"
```

---

## Task 5：前端 — DebateTopicView.vue evidence source 输入

**Files:**
- Modify: `web/src/views/debate/DebateTopicView.vue`

**背景：** 创建/编辑论点表单中，当 `argument_type === 'evidence'` 时，显示证据来源的三个输入字段。

- [ ] **Step 1: 在创建论点表单中添加 source 字段**

找到论点类型选择区域（`<select>` 或 `<label>论点类型</label>`），在 `AEditor`（论点内容）下方条件渲染 source 字段：

```vue
<!-- Evidence source fields - only shown for evidence type -->
<template v-if="newArgForm.argumentType === 'evidence'">
  <div class="a-field" style="margin-top:.5rem">
    <label class="a-field-label">来源 URL <span style="color:var(--a-color-danger)">*</span></label>
    <input v-model="newArgForm.sourceUrl" type="url" class="a-input" placeholder="https://..." />
  </div>
  <div class="a-field" style="margin-top:.5rem">
    <label class="a-field-label">来源标题</label>
    <input v-model="newArgForm.sourceTitle" type="text" class="a-input" placeholder="文章/报告标题" />
  </div>
  <div class="a-field" style="margin-top:.5rem">
    <label class="a-field-label">来源摘要</label>
    <textarea v-model="newArgForm.sourceExcerpt" class="a-textarea" rows="2" placeholder="相关引文……" />
  </div>
</template>
```

- [ ] **Step 2: 在 newArgForm 响应式对象中添加新字段**

找到 `newArgForm` 的定义（`const newArgForm = reactive({...})`），添加：

```typescript
sourceUrl: '',
sourceTitle: '',
sourceExcerpt: '',
```

- [ ] **Step 3: 在提交创建论点的 API 调用中传递新字段**

在 `createArgument` 函数的 payload 中添加：

```typescript
source_url: newArgForm.sourceUrl,
source_title: newArgForm.sourceTitle,
source_excerpt: newArgForm.sourceExcerpt,
```

- [ ] **Step 4: 同样更新编辑论点表单**

找到 `editArgForm` 的定义，添加 `sourceUrl / sourceTitle / sourceExcerpt` 字段；在填充编辑表单时将论点的已有来源数据填入；在提交更新时传递新字段。

- [ ] **Step 5: 类型检查**

```bash
cd web && bun run type-check 2>&1 | tail -5
```

- [ ] **Step 6: Commit**

```bash
git add web/src/views/debate/DebateTopicView.vue
git commit -m "feat(debate): add evidence source input fields to argument creation form"
```

---

## Task 6：前端 — ArgumentNode.vue 渲染 evidence source + admin 折叠

**Files:**
- Modify: `web/src/components/debate/ArgumentNode.vue`

- [ ] **Step 1: 读取 ArgumentNode.vue 现有结构**

```bash
wc -l web/src/components/debate/ArgumentNode.vue
grep -n "evidence\|vote\|fold\|admin\|source" web/src/components/debate/ArgumentNode.vue | head -20
```

- [ ] **Step 2: 添加 evidence source 展示**

在 argument 内容（`{{ argument.content }}`）下方，添加 evidence 类型的来源卡片：

```vue
<!-- Evidence source card -->
<div
  v-if="argument.argument_type === 'evidence' && argument.source_url"
  class="evidence-source-card"
  style="margin-top:.5rem;padding:.5rem .75rem;border:1px solid var(--a-border);border-radius:.375rem;font-size:.75rem"
>
  <a :href="argument.source_url" target="_blank" rel="noopener noreferrer" class="evidence-source-title" style="font-weight:700;display:block;margin-bottom:.2rem">
    {{ argument.source_title || argument.source_url }}
  </a>
  <p v-if="argument.source_excerpt" class="evidence-source-excerpt" style="color:var(--a-color-muted);margin:0;font-style:italic">
    "{{ argument.source_excerpt }}"
  </p>
</div>
```

- [ ] **Step 3: 添加折叠状态遮罩**

在 argument 主体内容的外层 `<div>` 上添加折叠覆盖：

```vue
<!-- Folded overlay (admin collapsed) -->
<div v-if="argument.is_folded" class="folded-mask" style="padding:.5rem .75rem;color:var(--a-color-muted);font-size:.75rem;font-style:italic">
  <span>[此论点已被管理员折叠</span>
  <span v-if="argument.fold_note">：{{ argument.fold_note }}</span>
  <span>]</span>
  <button
    v-if="authStore.user?.role === 'admin'"
    style="margin-left:.5rem;font-size:.65rem"
    @click="unfold"
  >展开</button>
</div>

<!-- Normal content (hidden when folded, except for admin) -->
<div v-if="!argument.is_folded || authStore.user?.role === 'admin'" :style="argument.is_folded ? 'opacity:.4' : ''">
  <!-- existing content here -->
</div>
```

- [ ] **Step 4: 添加管理员折叠按钮**

在 argument 操作栏（投票/回复按钮旁），添加管理员专属操作：

```vue
<button
  v-if="authStore.user?.role === 'admin' && !argument.is_folded"
  class="arg-action-btn"
  style="color:var(--a-color-danger);font-size:.65rem"
  @click="fold"
>折叠</button>
```

- [ ] **Step 5: 添加 fold/unfold 函数**

在 `<script setup>` 中添加（参照 `useAuthStore` 和 api 的现有调用方式）：

```typescript
const emit = defineEmits<{
  (e: 'reload'): void
}>()

async function fold() {
  const note = prompt('折叠原因（可选）：') || ''
  await api.post(`/api/debate/arguments/${props.argument.id}/fold`, { fold_note: note })
  emit('reload')
}

async function unfold() {
  await api.delete(`/api/debate/arguments/${props.argument.id}/fold`)
  emit('reload')
}
```

并在 `DebateTopicView.vue` 中监听 `ArgumentNode` 的 `reload` 事件来刷新论点列表。

- [ ] **Step 6: 类型检查**

```bash
cd web && bun run type-check 2>&1 | tail -5
```

- [ ] **Step 7: Commit**

```bash
git add web/src/components/debate/ArgumentNode.vue web/src/views/debate/DebateTopicView.vue
git commit -m "feat(debate): show evidence source in ArgumentNode; add admin fold/unfold"
```

---

## Task 7：验收

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

**证据来源：**
1. 进入某议题 → 点击"添加论点" → 选择类型"证据" → 出现 URL/标题/摘要三个输入框
2. 填写后提交 → 论点显示在列表中，下方出现来源卡片（链接 + 标题 + 引文）

**管理员折叠：**
1. 以管理员身份进入议题 → 每个论点旁出现"折叠"按钮
2. 点击折叠 → 论点显示为灰色折叠提示
3. 普通用户看到折叠提示，无法展开
4. 管理员可点击"展开"恢复

- [ ] **Step 4: Final commit**

```bash
git add -A
git commit -m "feat(debate): complete Debate module - evidence sources, admin fold/unfold"
```
