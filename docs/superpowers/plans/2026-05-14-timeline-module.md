# Timeline Module Completion Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task.

**Goal:** 补全 Timeline 模块的三类缺失功能：① BCE（公元前）日期支持——后端 `parseDateTime` 需处理负值年份，前端 `formatDatetime` 和 `formatTickLabel` 需展示"公元前 N 年"，筛选年份输入支持负值；② 版本历史（轻量维基化）——每次事件编辑生成 revision 快照，可查看历史、管理员可回滚；③ 事件可见性控制——`is_public=false` 时仅本人和管理员可见，前端创建/编辑表单暴露此开关。

**Architecture:**
- 后端已完整：`TimelineEvent` CRUD（含 `is_public`、`source`、`location`、`lat/lng`、`tags`）、`TimelinePerson` + `PersonLocation` CRUD、年份范围过滤、分类过滤
- 前端已完整：`TimelineHomeView.vue`（时间轴视图 + 创建/编辑/删除事件 + 人物关联 + 对比视图 + 地图 popup）、`PersonListView.vue`、`PersonMapView.vue`
- `DatetimePicker.vue` 已支持 BCE 年份（负值输入、`-0500` 格式输出，line 223 注释确认）
- 缺失 ① `parseDateTime` 在后端无法解析 `-0500-01-01` 格式的负年份字符串
- 缺失 ② 前端 `formatDatetime` 对负年份显示为"公元前 N 年"（JS `Date` 年份 ≤ 0 需用不同策略解析）
- 缺失 ③ `is_public` 字段在创建/编辑表单中无 UI 开关
- 缺失 ④ 版本历史模型和端点（`TimelineRevision`）

**已完整的部分（无需修改）：**
- `timeline_handler.go`: Event/Person/Location CRUD、年份/分类过滤、地理坐标
- `timeline.go` 模型: TimelineEvent, TimelinePerson, PersonLocation
- `TimelineHomeView.vue`: 完整时间轴 + 地图 + 人物
- `DatetimePicker.vue`: 已支持负年份输入

---

## 文件清单

### 新增
| 文件 | 职责 |
|------|------|
| （无） | 无需新增文件 |

### 修改
| 文件 | 改动 |
|------|------|
| `server/internal/model/timeline.go` | 添加 `TimelineRevision` 模型 |
| `server/cmd/start_server/main.go` | 注册 `TimelineRevision` 到 AutoMigrate |
| `server/internal/handlers/timeline_handler.go` | `parseDateTime` 支持负年份；`CreateTimelineEvent` / `UpdateTimelineEvent` 保存 revision 快照；新增 `GET /events/:id/history`、`POST /events/:id/revert/:revision_id` |
| `web/src/views/timeline/TimelineHomeView.vue` | `formatDatetime` / `formatTickLabel` 支持负年份展示；创建/编辑表单添加 `is_public` 开关；事件详情面板添加"历史版本"入口 |
| `web/src/types.ts` | 添加 `TimelineRevision` 类型；`TimelineEvent` 添加 `is_public` 字段 |

---

## Task 1：后端模型 — TimelineRevision

**Files:**
- Modify: `server/internal/model/timeline.go`

**背景：** 规格要求每次编辑生成版本快照（轻量维基化），状态只分两态：draft/published（TimelineEvent 已有 `is_public` 可做此用途，无需单独 status 字段）。

- [ ] **Step 1: 在 timeline.go 末尾添加 TimelineRevision 模型**

在 `PersonLocation` 的 `TableName()` 下方添加：

```go
// TimelineRevision is a snapshot of a TimelineEvent at a specific point in time.
// Created automatically on every update. Admins can revert to any revision.
type TimelineRevision struct {
	Base
	EventID   uuid.UUID `json:"event_id" gorm:"type:uuid;not null;index"`
	EditorID  uuid.UUID `json:"editor_id" gorm:"type:uuid;not null;index"`
	Editor    *User     `json:"editor,omitempty" gorm:"foreignKey:EditorID;references:UUID"`
	// Snapshot fields (duplicated from TimelineEvent)
	Title       string  `json:"title"`
	Description string  `json:"description" gorm:"type:text"`
	Content     string  `json:"content" gorm:"type:text"`
	EventDate   string  `json:"event_date"`   // stored as ISO string to preserve BCE
	EndDate     string  `json:"end_date"`
	Location    string  `json:"location"`
	Source      string  `json:"source"`
	Category    string  `json:"category"`
	IsPublic    bool    `json:"is_public"`
}

func (TimelineRevision) TableName() string { return "timeline_revisions" }
```

- [ ] **Step 2: 编译验证**

```bash
cd server && go build ./...
```

- [ ] **Step 3: Commit**

```bash
git add server/internal/model/timeline.go
git commit -m "feat(timeline): add TimelineRevision model"
```

---

## Task 2：后端迁移 — 注册 TimelineRevision

**Files:**
- Modify: `server/cmd/start_server/main.go`

- [ ] **Step 1: 找到 AutoMigrate 中 Timeline 相关模型**

```bash
grep -n "TimelineEvent\|TimelinePerson\|PersonLocation" server/cmd/start_server/main.go | head -10
```

- [ ] **Step 2: 在最后一个 Timeline 模型后添加**

```go
&model.TimelineRevision{},
```

- [ ] **Step 3: 编译验证**

```bash
cd server && go build ./...
```

- [ ] **Step 4: Commit**

```bash
git add server/cmd/start_server/main.go
git commit -m "feat(timeline): register TimelineRevision in AutoMigrate"
```

---

## Task 3：后端 — parseDateTime 支持负年份（BCE）

**Files:**
- Modify: `server/internal/handlers/timeline_handler.go`

**背景：** Go `time.Parse` 的标准格式不支持负年份。`DatetimePicker.vue` 输出的 BCE 日期格式为 `-0500-01-01T00:00`（5 位带符号年份）。需在 `parseDateTime` 前先检测并处理负年份字符串。

- [ ] **Step 1: 读取当前 parseDateTime 函数（第 17–31 行）**

```bash
sed -n '17,31p' server/internal/handlers/timeline_handler.go
```

- [ ] **Step 2: 将 parseDateTime 替换为支持负年份的版本**

找到：

```go
// parseDateTime 尝试多种格式解析时间，支持精确到小时分钟
func parseDateTime(s string) (time.Time, error) {
	formats := []string{
		"2006-01-02T15:04",
		"2006-01-02T15:04:05",
		time.RFC3339,
		"2006-01-02 15:04",
		"2006-01-02",
	}
	for _, f := range formats {
		if t, err := time.Parse(f, s); err == nil {
			return t, nil
		}
	}
	return time.Time{}, &time.ParseError{Value: s}
}
```

替换为：

```go
// parseDateTime parses a datetime string with support for BCE (negative year) dates.
// BCE format (from DatetimePicker): "-0500-01-01T00:00" or "-0500-01-01"
// Standard formats: "2006-01-02T15:04", "2006-01-02T15:04:05", RFC3339, "2006-01-02"
func parseDateTime(s string) (time.Time, error) {
	// Handle BCE: string starts with "-" followed by 4-digit year
	if len(s) > 1 && s[0] == '-' {
		rest := s[1:] // strip leading minus
		// Parse the absolute part
		formats := []string{"0001-01-02T15:04", "0001-01-02T15:04:05", "0001-01-02 15:04", "0001-01-02"}
		for _, f := range formats {
			if t, err := time.Parse(f, rest); err == nil {
				// Negate the year using AddDate
				negYear := -t.Year() - 1 // convert to BCE (astronomical year 0 = 1 BCE)
				result := time.Date(negYear, t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second(), 0, time.UTC)
				return result, nil
			}
		}
	}
	formats := []string{
		"2006-01-02T15:04",
		"2006-01-02T15:04:05",
		time.RFC3339,
		"2006-01-02 15:04",
		"2006-01-02",
	}
	for _, f := range formats {
		if t, err := time.Parse(f, s); err == nil {
			return t, nil
		}
	}
	return time.Time{}, fmt.Errorf("cannot parse datetime: %q", s)
}
```

注意需要在 `import` 中添加 `"fmt"`（若未引入）。

- [ ] **Step 3: 检查 GetTimelineEvents 中年份范围过滤是否支持负值**

```bash
sed -n '77,100p' server/internal/handlers/timeline_handler.go
```

确认 `strconv.Atoi(yearStart)` 能正确解析负数字符串（Go `strconv.Atoi` 本身支持负数，无需修改）。

- [ ] **Step 4: 编译验证**

```bash
cd server && go build ./...
```

- [ ] **Step 5: Commit**

```bash
git add server/internal/handlers/timeline_handler.go
git commit -m "feat(timeline): support BCE (negative year) in parseDateTime"
```

---

## Task 4：后端 — 编辑时自动保存 revision 快照

**Files:**
- Modify: `server/internal/handlers/timeline_handler.go`

**背景：** 在 `UpdateTimelineEvent` handler 中，保存成功后追加一条 `TimelineRevision` 记录作为变更快照。

- [ ] **Step 1: 读取 UpdateTimelineEvent handler（约第 192–255 行）**

```bash
sed -n '192,260p' server/internal/handlers/timeline_handler.go
```

- [ ] **Step 2: 在 UpdateTimelineEvent 保存成功后插入 revision**

在 `db.Save` 或 `db.Model(...).Updates(...)` 调用成功后、返回 `c.JSON` 前，添加快照插入：

```go
// Save revision snapshot
editorID := c.MustGet("userID").(uuid.UUID)
revision := model.TimelineRevision{
    EventID:     event.ID,
    EditorID:    editorID,
    Title:       event.Title,
    Description: event.Description,
    Content:     event.Content,
    EventDate:   event.EventDate.Format(time.RFC3339),
    Location:    event.Location,
    Source:      event.Source,
    Category:    event.Category,
    IsPublic:    event.IsPublic,
}
if event.EndDate != nil {
    revision.EndDate = event.EndDate.Format(time.RFC3339)
}
db.Create(&revision) // non-fatal: ignore revision save errors
```

- [ ] **Step 3: 添加 GetEventHistory handler**

在文件末尾添加：

```go
// GetEventHistory returns the edit history (revisions) of a timeline event.
// Route: GET /api/timeline/events/:id/history
func GetEventHistory(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		var revisions []model.TimelineRevision
		db.Where("event_id = ?", id).
			Preload("Editor").
			Order("created_at DESC").
			Find(&revisions)
		c.JSON(http.StatusOK, gin.H{"data": revisions})
	}
}

// RevertEventToRevision restores a timeline event to a previous revision (admin only).
// Route: POST /api/timeline/events/:id/revert/:revision_id
func RevertEventToRevision(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		revisionID := c.Param("revision_id")

		// Admin check
		var user model.User
		editorID := c.MustGet("userID").(uuid.UUID)
		if db.First(&user, "uuid = ?", editorID).Error != nil || user.Role != "admin" {
			c.JSON(http.StatusForbidden, gin.H{"error": "admin only"})
			return
		}

		var rev model.TimelineRevision
		if err := db.First(&rev, "id = ? AND event_id = ?", revisionID, id).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "revision not found"})
			return
		}

		// Parse the snapshot dates
		eventDate, _ := parseDateTime(rev.EventDate)
		updates := map[string]interface{}{
			"title":       rev.Title,
			"description": rev.Description,
			"content":     rev.Content,
			"event_date":  eventDate,
			"location":    rev.Location,
			"source":      rev.Source,
			"category":    rev.Category,
			"is_public":   rev.IsPublic,
		}
		if rev.EndDate != "" {
			if endDate, err := parseDateTime(rev.EndDate); err == nil {
				updates["end_date"] = endDate
			}
		}

		if err := db.Model(&model.TimelineEvent{}).Where("id = ?", id).Updates(updates).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "revert failed"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "reverted"})
	}
}
```

- [ ] **Step 4: 注册路由**

在 `SetupTimelineRoutes` 中，public 区域（无需登录）添加：

```go
tl.GET("/events/:id/history", GetEventHistory(db))
```

在 protected 区域添加：

```go
protected.POST("/events/:id/revert/:revision_id", RevertEventToRevision(db))
```

- [ ] **Step 5: 编译验证**

```bash
cd server && go build ./...
```

- [ ] **Step 6: Commit**

```bash
git add server/internal/handlers/timeline_handler.go
git commit -m "feat(timeline): save revision on update; add history and revert endpoints"
```

---

## Task 5：前端类型 — 添加 TimelineRevision，补全 TimelineEvent

**Files:**
- Modify: `web/src/types.ts`

- [ ] **Step 1: 找到 TimelineEvent 类型**

```bash
grep -n "TimelineEvent\|interface.*Timeline" web/src/types.ts
```

- [ ] **Step 2: 若 is_public 字段缺失，添加到 TimelineEvent interface**

```typescript
is_public?: boolean
```

- [ ] **Step 3: 添加 TimelineRevision 类型**

在 `TimelineEvent` 类型下方添加：

```typescript
export interface TimelineRevision {
  id: string
  event_id: string
  editor_id: string
  editor?: User
  title: string
  description?: string
  content?: string
  event_date: string
  end_date?: string
  location?: string
  source?: string
  category?: string
  is_public?: boolean
  created_at: string
}
```

- [ ] **Step 4: 类型检查**

```bash
cd web && bun run type-check 2>&1 | tail -5
```

- [ ] **Step 5: Commit**

```bash
git add web/src/types.ts
git commit -m "feat(timeline): add TimelineRevision type; ensure is_public in TimelineEvent"
```

---

## Task 6：前端 — formatDatetime 支持 BCE 显示

**Files:**
- Modify: `web/src/views/timeline/TimelineHomeView.vue`

**背景：** JS `new Date("-0500-01-01")` 在各浏览器行为不一致，负年份应手动解析。`formatDatetime` 需展示"公元前 500 年"，`formatTickLabel` 需展示"−500"。

- [ ] **Step 1: 读取 formatDatetime 当前实现（约第 601–614 行）**

确认当前用 `new Date(value)` 解析，负年份会得到错误结果。

- [ ] **Step 2: 替换 formatDatetime 为支持 BCE 的版本**

找到 `const formatDatetime = (value: string) => {` 块，替换为：

```typescript
const formatDatetime = (value: string) => {
  if (!value) return ''

  // Handle BCE dates: string starts with "-"
  if (value.startsWith('-')) {
    const yearMatch = value.match(/^-(\d{4})/)
    if (yearMatch) {
      const absYear = parseInt(yearMatch[1], 10)
      const datePart = value.slice(1) // strip "-"
      const hasTime = datePart.includes('T') && datePart.split('T')[1] !== '00:00'
      const timeSuffix = hasTime ? ` ${datePart.split('T')[1].slice(0, 5)}` : ''
      return `公元前 ${absYear} 年${timeSuffix}`
    }
    return value
  }

  const date = new Date(value)
  if (Number.isNaN(date.getTime())) return value.slice(0, 16)

  const hours = date.getHours()
  const minutes = date.getMinutes()
  const dateLabel = value.slice(0, 10)

  if (hours === 0 && minutes === 0) return dateLabel
  return `${dateLabel} ${String(hours).padStart(2, '0')}:${String(minutes).padStart(2, '0')}`
}
```

- [ ] **Step 3: 更新 formatTickLabel 支持负年份**

找到 `const formatTickLabel = (timestamp: number) => {` 块。

`formatTickLabel` 接收的是毫秒时间戳，JS `Date` 年份 ≤ 0 时 `getFullYear()` 返回负数（astronmical year）。添加负年份处理：

```typescript
const year = date.getFullYear()
const displayYear = year <= 0 ? `公元前 ${Math.abs(year - 1)}` : String(year)
```

并在 return 语句中把 `year` 替换为 `displayYear`：

```typescript
if (totalDays > 365 * 2) return displayYear
if (totalDays > 60) return `${displayYear}-${month}`
return `${displayYear}-${month}-${day}`
```

- [ ] **Step 4: 更新筛选输入区 — yearStart/yearEnd 允许负值**

找到筛选区的年份输入：

```vue
<input v-model.number="yearStart" type="number" placeholder="如 1800" class="a-input" style="width:120px" />
```

添加提示支持负值：

```vue
<input v-model.number="yearStart" type="number" placeholder="如 -500 或 1800" class="a-input" style="width:120px" />
```

同样更新 `yearEnd` 的 `placeholder`。

- [ ] **Step 5: 类型检查**

```bash
cd web && bun run type-check 2>&1 | tail -5
```

- [ ] **Step 6: Commit**

```bash
git add web/src/views/timeline/TimelineHomeView.vue
git commit -m "feat(timeline): support BCE date display in formatDatetime and formatTickLabel"
```

---

## Task 7：前端 — 创建/编辑表单 is_public 开关

**Files:**
- Modify: `web/src/views/timeline/TimelineHomeView.vue`

**背景：** `TimelineEvent` 有 `is_public` 字段（默认 `true`），但当前表单无 UI 开关，只能默认公开。

- [ ] **Step 1: 在 form 响应式对象中找到 is_public 字段**

```bash
grep -n "form.*is_public\|is_public.*form\|isPublic" web/src/views/timeline/TimelineHomeView.vue | head -5
```

若 `form` 中无 `is_public` 字段，在 `form` 的 reactive 定义中添加：

```typescript
is_public: true,
```

- [ ] **Step 2: 在表单中添加公开/私有开关**

在 `tags` 输入字段下方（表单末尾、提交按钮前）添加：

```vue
<div class="a-field">
  <label class="a-field-label">可见性</label>
  <label style="display:flex;align-items:center;gap:.5rem;font-size:.875rem;cursor:pointer">
    <input type="checkbox" v-model="form.is_public" />
    {{ form.is_public ? '公开（所有人可见）' : '仅自己可见' }}
  </label>
</div>
```

- [ ] **Step 3: 确认创建/编辑时传递 is_public**

找到 `createEvent` 或 `updateEvent` 调用处，确认 payload 包含 `is_public: form.is_public`。若缺失则添加。

- [ ] **Step 4: 类型检查**

```bash
cd web && bun run type-check 2>&1 | tail -5
```

- [ ] **Step 5: Commit**

```bash
git add web/src/views/timeline/TimelineHomeView.vue
git commit -m "feat(timeline): add is_public toggle to event create/edit form"
```

---

## Task 8：前端 — 事件详情面板添加历史版本入口

**Files:**
- Modify: `web/src/views/timeline/TimelineHomeView.vue`

**背景：** 点击事件后打开的详情面板（`detailEvent`）应显示"历史版本"按钮，展开后列出所有 revision，管理员可点击某版本执行回滚。

- [ ] **Step 1: 添加 revision 状态和加载函数**

在 `<script setup>` 中添加：

```typescript
import type { TimelineRevision } from '@/types'

const showHistory = ref(false)
const revisions = ref<TimelineRevision[]>([])
const revisionsLoading = ref(false)

async function loadHistory(eventId: string) {
  revisionsLoading.value = true
  showHistory.value = true
  try {
    const res = await fetch(`/api/timeline/events/${eventId}/history`)
    const data = await res.json()
    revisions.value = data.data || []
  } finally {
    revisionsLoading.value = false
  }
}

async function revertToRevision(eventId: string, revisionId: string) {
  if (!confirm('确定要回滚到此版本吗？当前内容将被覆盖。')) return
  const token = authStore.token
  await fetch(`/api/timeline/events/${eventId}/revert/${revisionId}`, {
    method: 'POST',
    headers: { Authorization: `Bearer ${token}` },
  })
  // Reload events
  await store.fetchEvents({})
  showHistory.value = false
}
```

- [ ] **Step 2: 在详情面板中添加历史版本区域**

在详情面板的编辑/删除按钮旁，添加"历史版本"按钮：

```vue
<button
  class="a-btn a-btn-ghost"
  style="font-size:.75rem"
  @click="detailEvent && loadHistory(detailEvent.id)"
>历史版本</button>
```

在详情面板内，条件渲染版本列表：

```vue
<div v-if="showHistory" class="history-panel" style="margin-top:1rem;border-top:var(--a-border);padding-top:.75rem">
  <h4 style="font-size:.75rem;font-weight:900;margin:0 0 .5rem 0">历史版本</h4>
  <div v-if="revisionsLoading" class="a-skeleton" style="height:3rem" />
  <p v-else-if="!revisions.length" style="font-size:.75rem;color:var(--a-color-muted)">暂无版本记录</p>
  <ul v-else style="list-style:none;padding:0;margin:0;display:flex;flex-direction:column;gap:.25rem">
    <li
      v-for="rev in revisions"
      :key="rev.id"
      style="font-size:.75rem;display:flex;align-items:center;gap:.5rem"
    >
      <span style="color:var(--a-color-muted)">{{ rev.created_at.slice(0, 16) }}</span>
      <span>{{ rev.editor?.display_name || rev.editor?.username || '未知' }}</span>
      <button
        v-if="authStore.user?.role === 'admin'"
        class="a-btn a-btn-ghost"
        style="font-size:.65rem;padding:.15rem .35rem"
        @click="detailEvent && revertToRevision(detailEvent.id, rev.id)"
      >回滚</button>
    </li>
  </ul>
</div>
```

- [ ] **Step 3: 类型检查**

```bash
cd web && bun run type-check 2>&1 | tail -5
```

- [ ] **Step 4: Commit**

```bash
git add web/src/views/timeline/TimelineHomeView.vue
git commit -m "feat(timeline): add revision history panel with admin revert to event detail"
```

---

## Task 9：验收

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

**BCE 日期：**
1. 创建事件 → 打开 DatetimePicker → 在年份输入框输入 `-500` → 选择日期 → 提交
2. 事件在时间轴上出现，显示"公元前 500 年"
3. 时间轴刻度在跨 BCE 范围时正确显示"公元前 N"标签
4. 筛选框输入年份 `-500` 到 `500` → 结果包含 BCE 事件

**is_public 开关：**
1. 创建事件时取消勾选"公开" → 提交 → 其他用户不可见，本人可见

**版本历史：**
1. 更新某事件内容 → 在详情面板点击"历史版本" → 出现版本列表（至少一条）
2. 管理员点击"回滚" → 确认 → 事件内容恢复为历史版本

- [ ] **Step 4: Final commit**

```bash
git add -A
git commit -m "feat(timeline): complete Timeline module - BCE support, is_public UI, revision history"
```
