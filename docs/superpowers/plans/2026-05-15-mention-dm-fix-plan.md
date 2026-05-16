# @ 提及与私信链路修复 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 修复 Forum `@` 提及与私信 username 链路，让 followers 提示、`@username` 插入、旧 mention 兼容解析、`/inbox?tab=dm&user=:username` 打开和发送消息都稳定可用。

**Architecture:** 前端统一以 `username` 作为 `@` 与 DM 的唯一目标标识；`AEditor` 只插入 `@username`，Inbox/DM store 只传 `username`。后端通过 followers 限定候选来源、mention 解析兼容 `@username` 与旧 Markdown mention，并把 DM 查人链路统一收敛到稳健的 username 解析函数。

**Tech Stack:** Vue 3.5 / Pinia / Vue Router 4 / TypeScript 5.9 / Go / Gin / GORM / PostgreSQL / WebSocket

---

## File Structure

- Modify: `server/internal/handlers/user_handler.go`
  - 扩展用户搜索接口，支持当前登录用户 followers 范围的 mention 候选查询
- Modify: `server/internal/service/forum_mention_parser.go`
  - 统一解析 `@username` 与 `[@显示名](/user/username)` 两种 mention 格式
- Modify: `server/internal/handlers/dm_handler.go`
  - 引入统一的 username 规范化与查找逻辑，修复 Inbox 深链会话打开与发送消息
- Modify: `web/src/components/shared/AEditor.vue`
  - mention 候选改为 followers，插入格式改成 `@username`
- Modify: `web/src/stores/dm.ts`
  - 把 username 无效错误显式抛给页面，并保证 query 深链能反馈失败
- Modify: `web/src/views/feed/InboxPage.vue`
  - 读取 `/inbox?tab=dm&user=:username` 后，失败时显示明确错误

---

### Task 1: 后端 followers mention 候选接口

**Files:**
- Modify: `server/internal/handlers/user_handler.go:408-445`
- Verify: `server/internal/handlers/user_handler.go:16-40`

- [ ] **Step 1: 阅读并定位现有用户搜索路由**

Run:
```bash
rg -n "users.GET\(\"/search\"|func SearchUsers\(" server/internal/handlers/user_handler.go
```

Expected: 能看到 `users.GET("/search", SearchUsers(db))` 与 `func SearchUsers(db *gorm.DB)` 的定义位置。

- [ ] **Step 2: 将 SearchUsers 改成支持 mention 场景的 followers 过滤**

把 `SearchUsers` 的实现替换为下面这段，保留原有 `q/limit`，新增 `scope=mention` 时只返回“关注我的用户”：

```go
// SearchUsers returns users matching the query string.
// GET /api/users/search?q=<query>&limit=<n>&scope=mention
func SearchUsers(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		q := strings.TrimSpace(c.Query("q"))
		scope := strings.TrimSpace(c.Query("scope"))
		limit := 5
		if l, err := strconv.Atoi(c.Query("limit")); err == nil && l > 0 && l <= 20 {
			limit = l
		}

		type UserResult struct {
			UUID        string `json:"uuid"`
			Username    string `json:"username"`
			DisplayName string `json:"display_name"`
			AvatarURL   string `json:"avatar_url"`
		}

		query := db.Model(&model.User{}).
			Select("Users.uuid, Users.username, Users.display_name, Users.avatar_url").
			Where("Users.is_active = ?", true).
			Limit(limit)

		if scope == "mention" {
			userIDVal, ok := c.Get("user_id")
			if !ok {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
				return
			}
			userID := userIDVal.(uuid.UUID)
			query = query.Joins("JOIN follows ON follows.follower_id = Users.uuid").Where("follows.following_id = ?", userID)
		}

		if q != "" {
			like := "%" + q + "%"
			query = query.Where("Users.username ILIKE ? OR Users.display_name ILIKE ?", like, like)
		}

		var results []UserResult
		if err := query.Scan(&results).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "search failed"})
			return
		}
		if results == nil {
			results = []UserResult{}
		}
		c.JSON(http.StatusOK, gin.H{"data": results})
	}
}
```

- [ ] **Step 3: 补上 `strings` import 并确认无重复 import**

确保 `server/internal/handlers/user_handler.go` 的 import 至少包含：

```go
import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"

	"atoman/internal/middleware"
	"atoman/internal/model"
)
```

- [ ] **Step 4: 运行后端构建验证接口改动**

Run:
```bash
cd server && go build ./...
```

Expected: 无编译错误。

- [ ] **Step 5: Commit**

```bash
git add server/internal/handlers/user_handler.go
git commit -m "fix(user): filter mention search candidates to followers"
```

---

### Task 2: 后端 mention 解析兼容新旧格式

**Files:**
- Modify: `server/internal/service/forum_mention_parser.go:1-57`

- [ ] **Step 1: 阅读当前 mention 正则实现**

Run:
```bash
sed -n '1,120p' server/internal/service/forum_mention_parser.go
```

Expected: 能看到当前只解析 `@username` 的 `mentionRe`。

- [ ] **Step 2: 用统一 username 提取逻辑替换现有解析器**

把整个文件替换成下面内容：

```go
package service

import (
	"regexp"
	"strings"

	"gorm.io/gorm"

	"atoman/internal/model"
)

var (
	codeBlockRe = regexp.MustCompile("(?s)```[\\s\\S]*?```|`[^`]+`")
	plainMentionRe = regexp.MustCompile(`@([A-Za-z0-9_-]{2,32})`)
	markdownMentionRe = regexp.MustCompile(`\[[^\]]*\]\(/user/([A-Za-z0-9_-]{2,32})\)`)
)

func ParseMentions(db *gorm.DB, content string) ([]model.User, error) {
	stripped := codeBlockRe.ReplaceAllString(content, "")
	usernames := extractMentionUsernames(stripped)
	if len(usernames) == 0 {
		return nil, nil
	}

	var users []model.User
	if err := db.Where("LOWER(username) IN ?", usernames).Find(&users).Error; err != nil {
		return nil, err
	}
	return users, nil
}

func extractMentionUsernames(content string) []string {
	seen := make(map[string]struct{})
	usernames := make([]string, 0)

	for _, match := range plainMentionRe.FindAllStringSubmatch(content, -1) {
		addMentionUsername(seen, &usernames, match[1])
	}
	for _, match := range markdownMentionRe.FindAllStringSubmatch(content, -1) {
		addMentionUsername(seen, &usernames, match[1])
	}
	return usernames
}

func addMentionUsername(seen map[string]struct{}, usernames *[]string, raw string) {
	username := strings.ToLower(strings.TrimSpace(raw))
	if username == "" {
		return
	}
	if _, ok := seen[username]; ok {
		return
	}
	seen[username] = struct{}{}
	*usernames = append(*usernames, username)
}
```

- [ ] **Step 3: 运行后端构建验证 mention 解析改动**

Run:
```bash
cd server && go build ./...
```

Expected: 无编译错误。

- [ ] **Step 4: Commit**

```bash
git add server/internal/service/forum_mention_parser.go
git commit -m "fix(forum): support plain and markdown mention parsing"
```

---

### Task 3: 后端 DM username 查找稳健化

**Files:**
- Modify: `server/internal/handlers/dm_handler.go:343-349`
- Modify: `server/internal/handlers/dm_handler.go:105-149`
- Modify: `server/internal/handlers/dm_handler.go:222-230`

- [ ] **Step 1: 阅读当前 DM username 查找函数和调用点**

Run:
```bash
rg -n "findUserByUsername|c.Param\(\"username\"\)" server/internal/handlers/dm_handler.go
```

Expected: 能看到 `getMessages`、`sendMessage`、`markRead` 都依赖 `findUserByUsername`。

- [ ] **Step 2: 添加 username 规范化函数**

在 `mustGetUserUUID` 之前加入：

```go
func normalizeUsername(raw string) string {
	return strings.TrimSpace(raw)
}
```

- [ ] **Step 3: 替换 findUserByUsername 实现为稳健匹配**

把 `findUserByUsername` 替换为：

```go
func (h *dmHandler) findUserByUsername(c *gin.Context, username string) (*model.User, bool) {
	normalized := normalizeUsername(username)
	if normalized == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_username", "message": "用户名不能为空"})
		return nil, false
	}

	var user model.User
	if err := h.db.Preload("Settings").Where("LOWER(username) = LOWER(?)", normalized).First(&user).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user_not_found", "message": "用户不存在"})
		return nil, false
	}
	return &user, true
}
```

- [ ] **Step 4: 确保三个调用点继续透传原始 param，由统一函数处理规范化**

三个调用点应保持这种形式：

```go
other, ok := h.findUserByUsername(c, c.Param("username"))
if !ok {
	return
}
```

分别检查：
- `getMessages(...)`
- `sendMessage(...)`
- `markRead(...)`

- [ ] **Step 5: 运行后端构建验证 DM username 修复**

Run:
```bash
cd server && go build ./...
```

Expected: 无编译错误。

- [ ] **Step 6: Commit**

```bash
git add server/internal/handlers/dm_handler.go
git commit -m "fix(dm): normalize username lookup for inbox conversations"
```

---

### Task 4: 前端 AEditor mention 候选改为 followers 并插入 `@username`

**Files:**
- Modify: `web/src/components/shared/AEditor.vue:561-622`

- [ ] **Step 1: 阅读当前 mention 获取与插入逻辑**

Run:
```bash
sed -n '560,630p' web/src/components/shared/AEditor.vue
```

Expected: 能看到 `/api/users/search` 请求和 `[@显示名](/user/username)` 插入逻辑。

- [ ] **Step 2: 把候选接口改成 mention followers 范围**

把 `fetchMentionUsers` 改成：

```ts
async function fetchMentionUsers(q: string) {
  try {
    const headers: Record<string, string> = {}
    if (authStore.token) headers.Authorization = `Bearer ${authStore.token}`
    const res = await fetch(`/api/users/search?scope=mention&q=${encodeURIComponent(q)}&limit=5`, { headers })
    if (!res.ok) return
    const data = await res.json()
    mention.value.results = data.data || []
    mention.value.visible = mention.value.results.length > 0
    mention.value.index = 0
  } catch { /* ignore */ }
}
```

- [ ] **Step 3: 把选中 mention 的插入格式改成 `@username`**

把 `applyMention` 改成：

```ts
function applyMention(user: MentionUser) {
  if (!cmView) return
  const pos = cmView.state.selection.main.head
  const insertText = `@${user.username}`
  cmView.dispatch({
    changes: { from: mention.value.startPos, to: pos, insert: insertText },
    selection: { anchor: mention.value.startPos + insertText.length },
  })
  cmView.focus()
  closeMention()
}
```

- [ ] **Step 4: 运行前端类型检查验证 mention 编辑器改动**

Run:
```bash
cd web && bun run type-check
```

Expected: `vue-tsc --noEmit` 通过。

- [ ] **Step 5: Commit**

```bash
git add web/src/components/shared/AEditor.vue
git commit -m "fix(editor): use followers mention candidates and insert username mentions"
```

---

### Task 5: 前端 Inbox 显式显示私信目标用户错误

**Files:**
- Modify: `web/src/stores/dm.ts`
- Modify: `web/src/views/feed/InboxPage.vue`

- [ ] **Step 1: 在 dm store 中显式抛出后端错误**

把 `openConversation` 改成如下实现，确保 username 无效时页面能拿到错误：

```ts
const openConversation = async (username: string, page = 1) => {
  if (!authStore.token) return
  loading.value = true
  activeConversation.value = username
  try {
    const res = await fetch(`${api.dm.conversation(username)}?page=${page}`, { headers: authHeaders() })
    const data = await res.json().catch(() => ({}))
    if (!res.ok) {
      throw new Error(data.message || data.error || '获取私信消息失败')
    }
    messages.value = data.data || []
    total.value = data.total || 0
    await markRead(username)
  } finally {
    loading.value = false
  }
}
```

- [ ] **Step 2: 在 Inbox 页面增加会话打开错误状态**

在 `web/src/views/feed/InboxPage.vue` 的状态定义区加入：

```ts
const dmOpenError = ref('')
```

- [ ] **Step 3: 让 loadTab 与 openConversation 捕获并显示错误**

把 `loadTab` 中 DM 分支改成：

```ts
if (activeTab.value === 'dm') {
  dmOpenError.value = ''
  await dmStore.fetchConversations()
  const user = typeof route.query.user === 'string' ? route.query.user : ''
  if (user) {
    try {
      await openConversation(user)
    } catch (error) {
      dmOpenError.value = error instanceof Error ? error.message : '打开私信失败'
    }
  }
  return
}
```

把 `openConversation` 改成：

```ts
const openConversation = async (username: string) => {
  dmOpenError.value = ''
  await router.replace({ path: '/inbox', query: { tab: 'dm', user: username } })
  await dmStore.openConversation(username)
}
```

- [ ] **Step 4: 在 DM 详情区渲染错误态**

把 DM 详情区的空状态片段替换成：

```vue
<AEmpty
  v-else
  :title="dmOpenError ? '无法打开会话' : '选择一个会话'"
  :description="dmOpenError || '点击左侧私信会话开始聊天。'"
/>
```

- [ ] **Step 5: 运行前端类型检查验证 Inbox 错误态改动**

Run:
```bash
cd web && bun run type-check
```

Expected: `vue-tsc --noEmit` 通过。

- [ ] **Step 6: Commit**

```bash
git add web/src/stores/dm.ts web/src/views/feed/InboxPage.vue
git commit -m "fix(inbox): surface invalid username errors for dm deep links"
```

---

### Task 6: 端到端验证 mention 与私信链路

**Files:**
- Verify: `server/internal/handlers/user_handler.go`
- Verify: `server/internal/service/forum_mention_parser.go`
- Verify: `server/internal/handlers/dm_handler.go`
- Verify: `web/src/components/shared/AEditor.vue`
- Verify: `web/src/stores/dm.ts`
- Verify: `web/src/views/feed/InboxPage.vue`

- [ ] **Step 1: 运行后端构建**

Run:
```bash
cd server && go build ./...
```

Expected: 无编译错误。

- [ ] **Step 2: 运行前端类型检查**

Run:
```bash
cd web && bun run type-check
```

Expected: `vue-tsc --noEmit` 通过。

- [ ] **Step 3: 手动验证 mention 候选范围**

Run:
```bash
cd web && bun run dev
```

Manual check:
- 登录 A 用户
- 在支持 `AEditor` 的输入区输入 `@`
- 只应看到“关注 A 的用户”候选，而不是任意用户

Expected: 候选列表范围正确。

- [ ] **Step 4: 手动验证 mention 插入与通知**

Manual check:
- 在编辑器里选中某个候选
- 验证正文里插入的是 `@username`，不是 `[@显示名](/user/username)`
- 发布/提交内容
- 被提及用户打开收件箱，能收到 mention 通知

Expected: mention 通知成功送达。

- [ ] **Step 5: 手动验证历史 Markdown mention 兼容**

Manual check:
- 构造一条内容包含：

```markdown
[@alice](/user/alice)
```

- 提交后确认 `alice` 仍能收到 mention 通知

Expected: 历史格式仍兼容。

- [ ] **Step 6: 手动验证用户主页私信链路**

Manual check:
- 打开某个用户主页
- 点击“发私信”
- 验证跳转到 `/inbox?tab=dm&user=:username`
- 输入消息并发送

Expected: 不再出现 `user not found`，会话能正常打开并发送。

- [ ] **Step 7: 手动验证无效 username 错误态**

Manual check:
- 手动访问：

```text
/inbox?tab=dm&user=definitely-not-exist-user
```

Expected: 页面显示明确错误，例如“用户不存在”或“无法打开会话”，而不是静默空白。

- [ ] **Step 8: Commit**

```bash
git add server/internal/handlers/user_handler.go \
        server/internal/service/forum_mention_parser.go \
        server/internal/handlers/dm_handler.go \
        web/src/components/shared/AEditor.vue \
        web/src/stores/dm.ts \
        web/src/views/feed/InboxPage.vue
git commit -m "fix(forum,dm): restore mention notifications and username-based dm flow"
```

---

## Self-Review

### Spec coverage
- followers mention 候选：Task 1 + Task 4
- 插入统一为 `@username`：Task 4
- 后端兼容 `@username` 与旧 Markdown mention：Task 2
- 修复 `/inbox?tab=dm&user=:username` 会话打开与发送：Task 3 + Task 5 + Task 6
- 无效 username 显式失败：Task 3 + Task 5 + Task 6
- DM 推送/未读变化：沿用现有发送逻辑，Task 6 做端到端验证

### Placeholder scan
- 无 TBD/TODO
- 所有代码步骤给出明确代码
- 所有验证步骤给出明确命令与预期

### Type consistency
- username 作为前后端唯一目标标识保持一致
- mention 新格式统一为 `@username`
- 历史格式统一由后端兼容解析
