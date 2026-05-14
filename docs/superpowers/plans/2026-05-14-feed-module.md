# Feed Module Completion Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task.

**Goal:** 完成 Feed 订阅模块的剩余功能：为博客频道添加独立 RSS 端点（`/channel/:slug/rss/article`）、补齐 FeedView.vue 的 OPML 导入/导出 UI 和健康检查显示，确保平台内频道订阅与 RSS 阅读器形成完整闭环。

**Architecture:** Feed 后端已完整（订阅管理、时间线、星标、稍后阅读、OPML、健康检查、平台频道/合集订阅均已实现）。本计划主要补两类缺口：① 博客频道 RSS 端点（博客 handler 缺失，但属于 Feed 体验的基础）；② FeedView.vue 中 OPML 和健康检查的 UI 集成。

**已完整的部分（无需修改）：**
- `feed_handler.go`: 全部路由和 handler
- `feed.go` 模型：`FeedSource`, `Subscription`, `FeedItem`, `FeedItemRead`, `FeedItemStar`, `ReadingListItem`, `SubscriptionGroup`
- `FeedView.vue`：主界面（订阅组、时间线、星标、稍后阅读）
- `FeedStarredView.vue`, `FeedReadingListView.vue`, `FeedItemDetailView.vue`, `FeedStatsView.vue`
- `ChannelView.vue`：`subscribeToChannel` / `unsubscribeFromChannel` 已接入 feedStore
- `CollectionView.vue`：`subscribeToCollection` / `unsubscribeFromCollection` 已接入

---

## 文件清单

### 新增
| 文件 | 职责 |
|------|------|
| （无） | Feed 后端 handler 已完整 |

### 修改
| 文件 | 改动 |
|------|------|
| `server/internal/handlers/blog_channel_handler.go` | 新增 `GetChannelArticleRSS` handler + 注册路由 `/channel/:slug/rss/article` |
| `web/src/views/blog/ChannelView.vue` | 将 RSS URL 从 `api.feed.rss(username)` 改为 `/channel/:slug/rss/article` |
| `web/src/views/feed/FeedView.vue` | 补 OPML 导入/导出 UI 及健康检查角标显示（仅在缺失时添加） |

---

## Task 1：后端 — 频道文章 RSS 端点

**Files:**
- Modify: `server/internal/handlers/blog_channel_handler.go`

**背景：** `feed_handler.go` 已有 `/api/feed/rss/:username`（用户级 RSS），但规格要求每个频道有独立的 `/channel/:slug/rss/article`。该端点输出该频道所有已发布文章的标准 RSS 2.0 XML。

- [ ] **Step 1: 在 blog_channel_handler.go 顶部补充 import**

确认 `blog_channel_handler.go` 已有如下 import（若缺失则添加）：

```go
import (
    "fmt"
    "strings"
    "time"
    // existing imports...
)
```

- [ ] **Step 2: 注册路由**

在 `SetupBlogChannelRoutes` 函数中，public routes 区域最后添加：

```go
		// Channel-specific article RSS feed
		blog.GET("/channels/slug/:slug/rss/article", GetChannelArticleRSS(db))
```

- [ ] **Step 3: 实现 handler**

在文件末尾添加：

```go
// GetChannelArticleRSS outputs a standard RSS 2.0 feed for a channel's published articles.
// Route: GET /api/blog/channels/slug/:slug/rss/article
func GetChannelArticleRSS(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		slug := c.Param("slug")
		var channel model.Channel
		if err := db.Preload("User").Where("slug = ?", slug).First(&channel).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "channel not found"})
			return
		}

		var posts []model.Post
		db.Where("channel_id = ? AND status = ?", channel.ID, "published").
			Preload("User").
			Order("created_at DESC").
			Limit(50).Find(&posts)

		scheme := c.Request.Header.Get("X-Forwarded-Proto")
		if scheme == "" {
			scheme = "https"
		}
		siteURL := fmt.Sprintf("%s://%s", scheme, c.Request.Host)

		c.Header("Content-Type", "application/rss+xml; charset=utf-8")
		c.String(http.StatusOK, buildArticleRSS(channel, posts, siteURL))
	}
}

func buildArticleRSS(ch model.Channel, posts []model.Post, siteURL string) string {
	var items strings.Builder
	for _, p := range posts {
		pubDate := p.CreatedAt.Format(time.RFC1123Z)
		summary := p.Summary
		if summary == "" && len(p.Content) > 280 {
			summary = p.Content[:280] + "…"
		} else if summary == "" {
			summary = p.Content
		}
		authorName := ""
		if p.User != nil {
			authorName = p.User.DisplayName
			if authorName == "" {
				authorName = p.User.Username
			}
		}
		items.WriteString(fmt.Sprintf(`
    <item>
      <title><![CDATA[%s]]></title>
      <link>%s/post/%s</link>
      <guid isPermaLink="true">%s/post/%s</guid>
      <pubDate>%s</pubDate>
      <description><![CDATA[%s]]></description>
      <author>%s</author>
    </item>`, p.Title, siteURL, p.ID, siteURL, p.ID, pubDate, summary, authorName))
	}

	return fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<rss version="2.0">
  <channel>
    <title><![CDATA[%s]]></title>
    <link>%s/channel/%s</link>
    <description><![CDATA[%s]]></description>
    <language>zh-cn</language>
    <lastBuildDate>%s</lastBuildDate>
    %s
  </channel>
</rss>`, ch.Name, siteURL, ch.Slug, ch.Description,
		time.Now().Format(time.RFC1123Z), items.String())
}
```

- [ ] **Step 4: 验证编译**

```bash
cd server && go build ./...
```

期望：无错误。

- [ ] **Step 5: smoke test**

```bash
curl -s "http://localhost:8080/api/blog/channels/slug/<your-test-slug>/rss/article" | head -5
```

期望：返回 `<?xml version="1.0"...>` 开头的 XML。

- [ ] **Step 6: Commit**

```bash
git add server/internal/handlers/blog_channel_handler.go
git commit -m "feat(feed/blog): add /channel/:slug/rss/article endpoint"
```

---

## Task 2：前端 — ChannelView.vue 修复 RSS URL

**Files:**
- Modify: `web/src/views/blog/ChannelView.vue`

**背景：** 当前 `channelRssUrl` 计算属性指向 `api.feed.rss(channel.value.user?.username)` (用户级 RSS)，应改为频道级 RSS 端点。

- [ ] **Step 1: 找到 channelRssUrl 计算属性**

在 `ChannelView.vue` 中找到：

```javascript
const channelRssUrl = computed(() => {
  // ...
  return api.feed.rss(channel.value.user?.username || '')
})
```

- [ ] **Step 2: 替换为频道级 RSS URL**

将计算属性改为：

```typescript
const channelRssUrl = computed(() => {
  if (!channel.value?.slug) return ''
  const base = import.meta.env.VITE_API_URL || '/api'
  return `${base}/blog/channels/slug/${channel.value.slug}/rss/article`
})
```

- [ ] **Step 3: 类型检查**

```bash
cd web && bun run type-check 2>&1 | tail -5
```

期望：无新增错误。

- [ ] **Step 4: Commit**

```bash
git add web/src/views/blog/ChannelView.vue
git commit -m "feat(feed/blog): update ChannelView RSS URL to per-channel article feed"
```

---

## Task 3：前端 — FeedView.vue OPML 导入/导出 UI

**Files:**
- Modify: `web/src/views/feed/FeedView.vue`

**背景：** 后端已有 `POST /api/feed/opml/import` 和 `GET /api/feed/opml/export`，需确认前端 UI 是否已存在；若缺失则在侧边栏底部添加 OPML 操作入口。

- [ ] **Step 1: 检查 FeedView.vue 中是否已有 OPML UI**

```bash
grep -n "opml\|OPML" web/src/views/feed/FeedView.vue
```

若已有 OPML 相关代码，则跳过本 Task，进入 Task 4。

- [ ] **Step 2（仅在缺失时执行）: 在侧边栏底部添加 OPML 操作区**

在 FeedView.vue 左侧侧边栏（subscription group 列表下方）找到最后一个分组操作区，在其后添加：

```vue
<!-- OPML Import / Export -->
<div style="border-top:var(--a-border);padding:0.75rem 1.25rem;display:flex;gap:0.5rem;flex-direction:column">
  <p style="font-weight:900;font-size:.65rem;text-transform:uppercase;letter-spacing:.08em;color:var(--a-color-muted);margin:0 0 .25rem 0">OPML</p>
  <label style="cursor:pointer;font-size:.75rem;font-weight:700;color:var(--a-color-fg)">
    <input
      type="file"
      accept=".opml,.xml"
      style="display:none"
      @change="handleOpmlImport"
    />
    导入订阅源
  </label>
  <a
    :href="`${apiBase}/feed/opml/export`"
    :download="`atoman-subscriptions.opml`"
    style="font-size:.75rem;font-weight:700;color:var(--a-color-fg);text-decoration:none"
  >导出订阅源</a>
</div>
```

- [ ] **Step 3（仅在缺失时执行）: 添加 handleOpmlImport 函数**

在 `<script setup>` 中添加：

```typescript
const apiBase = import.meta.env.VITE_API_URL || '/api'

const handleOpmlImport = async (event: Event) => {
  const file = (event.target as HTMLInputElement).files?.[0]
  if (!file) return
  const formData = new FormData()
  formData.append('opml', file)
  try {
    const res = await fetch(`${apiBase}/feed/opml/import`, {
      method: 'POST',
      headers: { Authorization: `Bearer ${authStore.token}` },
      body: formData,
    })
    if (res.ok) {
      await loadSubscriptions()
      // Toast feedback (若有 AToast 则调用，否则 alert)
      alert('OPML 导入成功')
    } else {
      const data = await res.json()
      alert(`导入失败: ${data.error || '未知错误'}`)
    }
  } catch (e) {
    alert('导入失败，请检查网络')
  }
}
```

- [ ] **Step 4: 类型检查**

```bash
cd web && bun run type-check 2>&1 | tail -5
```

- [ ] **Step 5: Commit**

```bash
git add web/src/views/feed/FeedView.vue
git commit -m "feat(feed): add OPML import/export UI to FeedView sidebar"
```

---

## Task 4：前端 — 订阅健康状态显示

**Files:**
- Modify: `web/src/views/feed/FeedView.vue`

**背景：** `Subscription` 模型有 `health_status`（healthy/warning/error）和 `error_message` 字段，后端有 `POST /api/feed/subscriptions/:id/health` 健康检查接口，但前端侧边栏订阅列表条目可能未显示健康状态角标。

- [ ] **Step 1: 检查是否已有健康状态显示**

```bash
grep -n "health\|warning\|error_message" web/src/views/feed/FeedView.vue
```

若已有，则跳过本 Task。

- [ ] **Step 2（仅在缺失时）: 在订阅条目中添加健康状态角标**

找到侧边栏中渲染每个 subscription 的元素（`v-for="sub in group.subscriptions"` 或类似结构），在订阅名称旁添加：

```vue
<span
  v-if="sub.health_status && sub.health_status !== 'healthy'"
  :title="sub.error_message || ''"
  :style="`
    display:inline-block;
    width:.45rem;height:.45rem;border-radius:9999px;margin-left:.35rem;vertical-align:middle;
    background:${sub.health_status === 'error' ? 'var(--a-color-danger)' : 'var(--a-color-warning, #f59e0b)'};
  `"
/>
```

- [ ] **Step 3: Commit**

```bash
git add web/src/views/feed/FeedView.vue
git commit -m "feat(feed): show health status indicator on subscriptions"
```

---

## Task 5：验收

- [ ] **Step 1: 后端构建**

```bash
cd server && go build ./...
```

期望：无错误。

- [ ] **Step 2: 前端构建**

```bash
cd web && bun run type-check && bun run build
```

期望：0 type errors，build 成功。

- [ ] **Step 3: 手动冒烟测试**

1. 访问 `/channel/:slug` → 点击 RSS 按钮 → 复制的 URL 应为 `/api/blog/channels/slug/:slug/rss/article`
2. 直接访问该 URL → 返回合法 RSS 2.0 XML
3. 将 RSS URL 粘贴到 FeedView → "+ 添加订阅" → 应能成功添加为外部 RSS 订阅源
4. FeedView 侧边栏底部有 OPML 导入/导出入口
5. 订阅列表中错误状态的订阅（`health_status !== 'healthy'`）旁显示橙/红色小圆点

- [ ] **Step 4: Final commit**

```bash
git add -A
git commit -m "feat(feed): complete Feed module - channel RSS, OPML UI, health indicators"
```
