# Blog / Studio Module Completion Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task.

**Goal:** 完成 Blog / Studio 模块的剩余功能：确认频道 RSS 端点存在，以及确认所有交互（评论、点赞、收藏）的正确性。

**Architecture:** 博客后端已完整（Channel/Post CRUD、Comment、Like、Bookmark、BlogDraft、RSS）。Video 和 Podcast 作为独立模块存在，不合并进博客频道主页。本计划仅补一个小型后端补丁：确认 `/channel/:slug/rss/article` 端点存在（与 Feed 计划共用 Task 1 逻辑，若已执行 Feed 计划则跳过）。

**已完整的部分（无需修改）：**
- `blog_channel_handler.go`: Channel/Collection CRUD
- `blog_post_handler.go`: Post CRUD（含 draft、pin、collection 分配）
- `blog_interaction_handler.go`: Comment、Like、Bookmark 全部完整
- `blog_upload_handler.go`: 图片上传
- `PostEditorView.vue`: AEditor SV 模式 + Yjs 协作 + 草稿管理
- `PostDetailView.vue`: Markdown 渲染 + CommentSection + Like 按钮 + Bookmark
- `ChannelView.vue`: 频道主页（合集侧边栏 + 文章列表）+ 订阅按钮

---

## 文件清单

### 修改
| 文件 | 改动 |
|------|------|
| `server/internal/handlers/blog_channel_handler.go` | 注册 `/channels/slug/:slug/rss/article`（若 Feed 计划已执行则跳过） |

---

## Task 1：后端 — 确认频道文章 RSS 端点存在

**背景：** Feed 计划 Task 1 已添加此端点。请先检查是否已存在，已存在则跳过。

- [ ] **Step 1: 检查端点是否已注册**

```bash
grep -n "rss/article" server/internal/handlers/blog_channel_handler.go
```

若输出包含 `rss/article`，本 Task 完成，直接进入 Task 2。

- [ ] **Step 2（仅在缺失时）: 参照 Feed 计划 Task 1 补充端点**

按 `docs/superpowers/plans/2026-05-14-feed-module.md` Task 1 执行，在 `blog_channel_handler.go` 中注册并实现 `GetChannelArticleRSS`。

- [ ] **Step 3: Commit**

```bash
cd server && go build ./...
git add server/internal/handlers/blog_channel_handler.go
git commit -m "feat(blog): ensure channel article RSS endpoint exists"
```

---

## Task 2：~~前端 — ChannelView.vue 内容类型 Tab 导航~~（已取消）

> **决策（2026-05-15）：** Video 和 Podcast 作为独立模块保持独立，不合并进博客频道主页。ChannelView.vue 无需添加视频/播客 Tab，本 Task 取消。

---

## Task 3：审计 — 交互功能完整性验证

**Files:**
- 只读审计，不修改文件（除非发现 bug）

**背景：** Comment、Like、Bookmark 后端已完整实现，PostDetailView.vue 已有 CommentSection 组件和 Like 按钮。本 Task 做冒烟验证确认其正确工作。

- [ ] **Step 1: 验证评论功能**

```bash
# 检查 CommentSection 是否被 PostDetailView 正确使用
grep -n "CommentSection" web/src/views/blog/PostDetailView.vue
```

期望：找到 `<CommentSection :post-id="post.id"` 的调用。

- [ ] **Step 2: 验证点赞功能**

```bash
grep -n "toggleLike\|likesCount\|liked" web/src/views/blog/PostDetailView.vue
```

期望：找到 `toggleLike` 函数和 `liked` / `likesCount` 状态。

- [ ] **Step 3: 验证书签/收藏功能**

```bash
grep -rn "bookmark\|Bookmark" web/src/views/blog/BookmarkView.vue | head -10
grep -n "bookmark\|Bookmark" web/src/views/blog/PostDetailView.vue | head -5
```

期望：BookmarkView.vue 有完整的 bookmark 列表，PostDetailView 有书签按钮。

- [ ] **Step 4: 若发现缺失则按以下方式修复**

**如果 PostDetailView.vue 缺少 Bookmark 按钮：** 在点赞按钮旁边添加：

```vue
<button
  @click="toggleBookmark"
  class="a-toggle-btn"
  :class="{ 'a-toggle-btn-active': bookmarked }"
>
  {{ bookmarked ? '★ 已收藏' : '☆ 收藏' }}
</button>
```

对应 script 添加：

```typescript
const bookmarked = ref(false)

async function loadBookmarkStatus() {
  if (!authStore.isAuthenticated || !post.value) return
  const res = await api.get<{ data: any[] }>('/api/blog/bookmarks')
  bookmarked.value = res.data.data.some((b: any) => b.post_id === post.value?.id)
}

async function toggleBookmark() {
  if (!authStore.isAuthenticated) return
  if (bookmarked.value) {
    const bookmarks = (await api.get<{ data: any[] }>('/api/blog/bookmarks')).data.data
    const bk = bookmarks.find((b: any) => b.post_id === post.value?.id)
    if (bk) await api.delete(`/api/blog/bookmarks/${bk.id}`)
    bookmarked.value = false
  } else {
    await api.post('/api/blog/bookmarks', { post_id: post.value?.id })
    bookmarked.value = true
  }
}
```

- [ ] **Step 5: Commit（仅在有修复时）**

```bash
git add web/src/views/blog/PostDetailView.vue
git commit -m "fix(blog): add bookmark toggle to PostDetailView"
```

---

## Task 4：验收

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

1. `/channel/:slug` → 看到 Tab 导航（至少"全部"和"文章"）
2. 若频道有 Video，"视频" Tab 出现；切换后显示视频卡片列表
3. 若频道有播客，"播客" Tab 出现；切换后显示单集列表
4. `/channel/:slug` → 点击 RSS → 粘贴到播客客户端或 RSS 阅读器，能解析出文章标题列表
5. 文章详情页：评论、点赞、收藏三个功能均可正常使用
6. BookmarkView `/blog/bookmarks` 能显示已收藏的文章

- [ ] **Step 4: Final commit**

```bash
git add -A
git commit -m "feat(blog): complete Blog/Studio module - content tabs, RSS, interaction audit"
```
