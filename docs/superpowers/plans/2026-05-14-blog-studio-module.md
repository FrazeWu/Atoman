# Blog / Studio Module Completion Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task.

**Goal:** 完成 Blog / Studio 模块的剩余功能：频道内容类型 Tab 导航（全部/文章/视频/播客）、频道主页展示集成 Video 和 Podcast 内容、以及确认所有交互（评论、点赞、收藏）的正确性。

**Architecture:** 博客后端已完整（Channel/Post CRUD、Comment、Like、Bookmark、BlogDraft、RSS）。本计划主要补前端侧的内容类型 Tab 导航，以及在 Video 和 Podcast 模块上线后将其内容整合进频道主页。同时包含一个小型后端补丁：将 `/channel/:slug/rss/article` 端点从 Feed 模块提取为正式 Blog 模块端点（与 Feed 计划共用 Task 1 逻辑，若已执行 Feed 计划则跳过）。

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
| `web/src/views/blog/ChannelView.vue` | 添加内容类型 Tab（全部/文章/视频/播客），集成 Video/Podcast 内容列表 |

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

## Task 2：前端 — ChannelView.vue 内容类型 Tab 导航

**Files:**
- Modify: `web/src/views/blog/ChannelView.vue`

**背景：** 规格要求频道主页分 Tab：全部 / 文章 / 视频 / 播客。当前 ChannelView.vue 只显示文章列表。本 Task 添加 Tab 切换，让每个 Tab 加载对应类型内容。

**约束：**
- Video 和 Podcast 模块后端 API 已在其各自实施计划中建立（`/api/videos?channel_id=...` 和 `/api/podcast/shows/:slug/episodes`）
- Tab 显示策略：始终显示全部/文章 Tab；视频/播客 Tab 仅在该频道有该类内容时显示（初始化时静默拉取 count，count > 0 时展示）

- [ ] **Step 1: 读取 ChannelView.vue 当前渲染文章的代码段**

找到文章列表渲染区域（`v-for="post in posts"`），理解其位置和相邻结构。

- [ ] **Step 2: 添加 Tab 状态和内容加载逻辑**

在 `<script setup>` 中添加以下响应式状态和函数：

```typescript
import type { Video, PodcastEpisode } from '@/types'

// --- 内容类型 Tab ---
type ContentTab = 'all' | 'post' | 'video' | 'podcast'
const activeTab = ref<ContentTab>('all')

// Video/Podcast 内容（仅在切换到对应 Tab 时加载）
const videos = ref<Video[]>([])
const episodes = ref<PodcastEpisode[]>([])
const loadingVideos = ref(false)
const loadingEpisodes = ref(false)
// Tab 可见性（根据内容存在性控制）
const hasVideos = ref(false)
const hasPodcast = ref(false)

async function checkVideoAndPodcastExistence() {
  if (!channel.value?.id) return
  // Videos: 检查是否有任何视频
  try {
    const vRes = await api.get<Video[]>(`/api/videos?channel_id=${channel.value.id}`)
    hasVideos.value = vRes.data.length > 0
  } catch {}
  // Episodes: 检查是否有播客单集
  try {
    if (!channel.value?.slug) return
    const pRes = await api.get<{ episodes: PodcastEpisode[] }>(`/api/podcast/shows/${channel.value.slug}/episodes`)
    hasPodcast.value = pRes.data.episodes.length > 0
  } catch {}
}

async function loadVideos() {
  if (!channel.value?.id) return
  loadingVideos.value = true
  try {
    const res = await api.get<Video[]>(`/api/videos?channel_id=${channel.value.id}&sort=latest`)
    videos.value = res.data
  } finally {
    loadingVideos.value = false
  }
}

async function loadEpisodes() {
  if (!channel.value?.slug) return
  loadingEpisodes.value = true
  try {
    const res = await api.get<{ episodes: PodcastEpisode[] }>(`/api/podcast/shows/${channel.value.slug}/episodes`)
    episodes.value = res.data.episodes
  } finally {
    loadingEpisodes.value = false
  }
}

async function switchTab(tab: ContentTab) {
  activeTab.value = tab
  if (tab === 'video' && videos.value.length === 0) await loadVideos()
  if (tab === 'podcast' && episodes.value.length === 0) await loadEpisodes()
}
```

在 `onMounted` 或频道加载完成后调用 `checkVideoAndPodcastExistence()`。

- [ ] **Step 3: 在模板中添加 Tab 导航**

在频道头部下方、内容区域上方插入 Tab 导航：

```vue
<!-- Content type tabs -->
<div class="channel-tabs" v-if="channel">
  <button
    v-for="tab in availableTabs"
    :key="tab.value"
    class="channel-tab"
    :class="{ active: activeTab === tab.value }"
    @click="switchTab(tab.value)"
  >{{ tab.label }}</button>
</div>
```

在 `<script setup>` 中添加 `availableTabs` 计算属性：

```typescript
const availableTabs = computed(() => {
  const tabs: { label: string; value: ContentTab }[] = [
    { label: '全部', value: 'all' },
    { label: '文章', value: 'post' },
  ]
  if (hasVideos.value) tabs.push({ label: '视频', value: 'video' })
  if (hasPodcast.value) tabs.push({ label: '播客', value: 'podcast' })
  return tabs
})
```

- [ ] **Step 4: 在模板中条件渲染各 Tab 内容**

现有文章列表区域（`v-for="post in posts"`）加上 `v-if`：

```vue
<!-- 文章内容（全部/文章 Tab） -->
<div v-if="activeTab === 'all' || activeTab === 'post'">
  <!-- existing posts grid/list code stays here -->
</div>

<!-- 视频内容 -->
<div v-else-if="activeTab === 'video'">
  <div v-if="loadingVideos" class="a-skeleton" style="height:10rem" />
  <AEmpty v-else-if="!videos.length" title="暂无视频" />
  <div v-else class="a-grid-2">
    <VideoCard v-for="v in videos" :key="v.id" :video="v" />
  </div>
</div>

<!-- 播客内容 -->
<div v-else-if="activeTab === 'podcast'">
  <div v-if="loadingEpisodes" class="a-skeleton" style="height:10rem" />
  <AEmpty v-else-if="!episodes.length" title="暂无播客" />
  <ul v-else class="episode-list">
    <li v-for="ep in episodes" :key="ep.id" class="episode-item">
      <router-link :to="`/podcast/${ep.id}`" class="episode-title">{{ ep.post?.title }}</router-link>
      <span class="episode-num" v-if="ep.episode_number">第 {{ ep.episode_number }} 集</span>
    </li>
  </ul>
</div>
```

添加所需 import：

```typescript
import VideoCard from '@/components/shared/VideoCard.vue'
```

- [ ] **Step 5: 添加 Tab 样式**

在 `<style scoped>` 中添加：

```css
.channel-tabs {
  @apply flex gap-2 mb-4;
}
.channel-tab {
  @apply px-3 py-1 text-sm font-medium border border-transparent rounded hover:border-neutral-400 transition-colors;
}
.channel-tab.active {
  @apply border-neutral-900 dark:border-neutral-100;
}
.episode-list { @apply list-none p-0 flex flex-col gap-2; }
.episode-item { @apply flex items-center gap-3 py-2 border-b border-neutral-100 dark:border-neutral-800; }
.episode-title { @apply text-sm no-underline hover:underline; }
.episode-num { @apply text-xs text-neutral-400 ml-auto; }
```

- [ ] **Step 6: 类型检查**

```bash
cd web && bun run type-check 2>&1 | tail -5
```

期望：无新增错误（若 Video/Podcast 计划未执行，Video/PodcastEpisode 类型可能暂缺，可临时注释对应 import）。

- [ ] **Step 7: Commit**

```bash
git add web/src/views/blog/ChannelView.vue
git commit -m "feat(blog): add content type tabs (post/video/podcast) to ChannelView"
```

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
