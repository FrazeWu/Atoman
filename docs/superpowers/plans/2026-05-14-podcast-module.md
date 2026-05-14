# Podcast Module Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 实现 Podcast 模块，支持创作者通过频道发布播客节目（Show = Channel，Episode = Post 类型扩展），复用底部播放条收听，输出符合播客规范的 RSS feed（含 `<enclosure>`）。

**Architecture:** Show 直接复用已有 `Channel` 模型，Episode 扩展 `Post` 模型（新增 `audio_url`、`episode_number`、`season_number`、`episode_cover_url` 等字段到独立的 `PodcastEpisode` 表，关联 `post_id`）。后端新增 `podcast_handler.go`，前端新增发现页、节目页、单集页、发布页。播放复用现有底部播放条（`AudioPlayer`）。

**Tech Stack:** Go 1.21 / Gin / GORM / PostgreSQL；Vue 3.5 / TypeScript 5.9 / Tailwind CSS v4；S3/MinIO（音频文件本地上传）

---

## 文件清单

### 新增
| 文件 | 职责 |
|------|------|
| `server/internal/model/podcast.go` | PodcastEpisode 模型（扩展 Post） |
| `server/internal/handlers/podcast_handler.go` | Episode CRUD、节目 RSS |
| `web/src/views/podcast/PodcastHomeView.vue` | 节目发现页 `/podcast` |
| `web/src/views/podcast/PodcastShowView.vue` | 节目详情页 `/podcast/show/:channelSlug` |
| `web/src/views/podcast/PodcastEpisodeView.vue` | 单集收听页 `/podcast/:episodeId` |
| `web/src/views/podcast/PodcastEditorView.vue` | 单集发布/编辑页 |

### 修改
| 文件 | 改动 |
|------|------|
| `server/cmd/start_server/main.go` | 注册 PodcastEpisode 到 AutoMigrate，注册路由 |
| `web/src/router.ts` | 添加 podcast 路由 |
| `web/src/types.ts` | 添加 `PodcastEpisode` 类型 |

---

## Task 1：后端数据模型

**Files:**
- Create: `server/internal/model/podcast.go`
- Modify: `server/cmd/start_server/main.go`（AutoMigrate）

- [ ] **Step 1: 创建 model 文件**

```go
// server/internal/model/podcast.go
package model

import "github.com/google/uuid"

// PodcastEpisode extends Post with audio-specific fields.
// Show (节目) = Channel; Episode (单集) = Post + PodcastEpisode.
// The relationship: PodcastEpisode.PostID -> Post.ID (one-to-one).
type PodcastEpisode struct {
	Base
	PostID        uuid.UUID  `json:"post_id" gorm:"type:uuid;not null;uniqueIndex"`
	Post          *Post      `json:"post,omitempty" gorm:"foreignKey:PostID"`
	ChannelID     uuid.UUID  `json:"channel_id" gorm:"type:uuid;not null;index"`
	Channel       *Channel   `json:"channel,omitempty" gorm:"foreignKey:ChannelID"`
	// Audio file: always local upload (S3/MinIO)
	AudioURL      string     `json:"audio_url" gorm:"type:text;not null"`
	DurationSec   int        `json:"duration_sec" gorm:"default:0"`
	// Episode cover: optional; falls back to channel cover in RSS
	EpisodeCoverURL string   `json:"episode_cover_url" gorm:"type:text"`
	// Episode ordering
	SeasonNumber  int        `json:"season_number" gorm:"default:1"`
	EpisodeNumber int        `json:"episode_number" gorm:"default:0"`
	// Visibility mirrors Post.Status: draft | published
}

func (PodcastEpisode) TableName() string { return "podcast_episodes" }
```

- [ ] **Step 2: 注册 AutoMigrate**

在 `server/cmd/start_server/main.go` 的 AutoMigrate 列表中（紧跟 `&model.VideoTagRelation{}` 或 Video 模型之后）添加：

```go
			&model.PodcastEpisode{},
```

- [ ] **Step 3: 验证编译**

```bash
cd server && go build ./...
```

期望：无错误。

- [ ] **Step 4: 验证迁移**

```bash
go run cmd/start_server/main.go 2>&1 | grep -E "migration|podcast"
```

期望：migration 完成，无 Fatal。

- [ ] **Step 5: Commit**

```bash
git add server/internal/model/podcast.go server/cmd/start_server/main.go
git commit -m "feat(podcast): add PodcastEpisode model"
```

---

## Task 2：后端 Handler — Episode CRUD

**Files:**
- Create: `server/internal/handlers/podcast_handler.go`
- Modify: `server/cmd/start_server/main.go`（路由注册）

- [ ] **Step 1: 创建 Handler 文件**

```go
// server/internal/handlers/podcast_handler.go
package handlers

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"

	"atoman/internal/middleware"
	"atoman/internal/model"
	"atoman/internal/storage"
)

func SetupPodcastRoutes(router *gin.Engine, db *gorm.DB, s3Client *storage.S3Client) {
	p := router.Group("/api/podcast")
	{
		// List all published episodes (discover page)
		p.GET("/episodes", GetPodcastEpisodes(db))
		// Episodes for a specific channel (show page)
		p.GET("/shows/:channelSlug/episodes", GetShowEpisodes(db))
		// Single episode detail
		p.GET("/episodes/:id", GetPodcastEpisode(db))
		// Auth-required: create, update, delete
		p.POST("/episodes", middleware.RequireAuth(), CreatePodcastEpisode(db, s3Client))
		p.PUT("/episodes/:id", middleware.RequireAuth(), UpdatePodcastEpisode(db))
		p.DELETE("/episodes/:id", middleware.RequireAuth(), DeletePodcastEpisode(db))
	}
	// Per-channel podcast RSS
	router.GET("/api/channels/:slug/rss/podcast", GetPodcastRSS(db))
}

// GetPodcastEpisodes lists all published episodes across all shows.
func GetPodcastEpisodes(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var episodes []model.PodcastEpisode
		db.Preload("Post").Preload("Channel").
			Joins("JOIN posts ON posts.id = podcast_episodes.post_id AND posts.status = 'published'").
			Order("podcast_episodes.created_at DESC").
			Limit(40).Find(&episodes)
		c.JSON(http.StatusOK, episodes)
	}
}

// GetShowEpisodes returns all published episodes for a specific channel (show).
func GetShowEpisodes(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		slug := c.Param("channelSlug")
		var channel model.Channel
		if err := db.Where("slug = ?", slug).First(&channel).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "show not found"})
			return
		}
		var episodes []model.PodcastEpisode
		db.Where("podcast_episodes.channel_id = ?", channel.ID).
			Preload("Post").Preload("Channel").
			Joins("JOIN posts ON posts.id = podcast_episodes.post_id AND posts.status = 'published'").
			Order("podcast_episodes.season_number ASC, podcast_episodes.episode_number ASC").
			Find(&episodes)
		c.JSON(http.StatusOK, gin.H{"channel": channel, "episodes": episodes})
	}
}

// GetPodcastEpisode returns a single episode by ID.
func GetPodcastEpisode(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		var ep model.PodcastEpisode
		if err := db.Preload("Post").Preload("Channel").
			First(&ep, "podcast_episodes.id = ?", id).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "episode not found"})
			return
		}
		c.JSON(http.StatusOK, ep)
	}
}

// CreatePodcastEpisode creates a Post and linked PodcastEpisode in one transaction.
func CreatePodcastEpisode(db *gorm.DB, s3Client *storage.S3Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.MustGet("userID").(uuid.UUID)
		var input struct {
			ChannelID       string `json:"channel_id" binding:"required"`
			Title           string `json:"title" binding:"required"`
			Shownotes       string `json:"shownotes"`
			AudioURL        string `json:"audio_url" binding:"required"`
			DurationSec     int    `json:"duration_sec"`
			EpisodeCoverURL string `json:"episode_cover_url"`
			SeasonNumber    int    `json:"season_number"`
			EpisodeNumber   int    `json:"episode_number"`
			Visibility      string `json:"visibility"`
			Status          string `json:"status"`
		}
		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		chID, err := uuid.Parse(input.ChannelID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid channel_id"})
			return
		}

		status := input.Status
		if status == "" {
			status = "draft"
		}
		visibility := input.Visibility
		if visibility == "" {
			visibility = "public"
		}

		var ep model.PodcastEpisode
		txErr := db.Transaction(func(tx *gorm.DB) error {
			post := model.Post{
				UserID:    userID,
				ChannelID: &chID,
				Title:     strings.TrimSpace(input.Title),
				Content:   input.Shownotes,
				Status:    status,
			}
			if err := tx.Create(&post).Error; err != nil {
				return err
			}
			ep = model.PodcastEpisode{
				PostID:          post.ID,
				ChannelID:       chID,
				AudioURL:        input.AudioURL,
				DurationSec:     input.DurationSec,
				EpisodeCoverURL: input.EpisodeCoverURL,
				SeasonNumber:    maxInt(input.SeasonNumber, 1),
				EpisodeNumber:   input.EpisodeNumber,
			}
			return tx.Create(&ep).Error
		})
		if txErr != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": txErr.Error()})
			return
		}

		db.Preload("Post").Preload("Channel").First(&ep, "podcast_episodes.id = ?", ep.ID)
		c.JSON(http.StatusCreated, ep)
	}
}

// UpdatePodcastEpisode updates episode metadata and shownotes.
func UpdatePodcastEpisode(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.MustGet("userID").(uuid.UUID)
		id := c.Param("id")

		var ep model.PodcastEpisode
		if err := db.Preload("Post").First(&ep, "podcast_episodes.id = ?", id).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "episode not found"})
			return
		}
		if ep.Post == nil || ep.Post.UserID != userID {
			c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
			return
		}

		var input struct {
			Title           *string `json:"title"`
			Shownotes       *string `json:"shownotes"`
			EpisodeCoverURL *string `json:"episode_cover_url"`
			DurationSec     *int    `json:"duration_sec"`
			SeasonNumber    *int    `json:"season_number"`
			EpisodeNumber   *int    `json:"episode_number"`
			Status          *string `json:"status"`
		}
		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		postUpdates := map[string]interface{}{}
		if input.Title != nil {
			postUpdates["title"] = strings.TrimSpace(*input.Title)
		}
		if input.Shownotes != nil {
			postUpdates["content"] = *input.Shownotes
		}
		if input.Status != nil {
			postUpdates["status"] = *input.Status
		}
		if len(postUpdates) > 0 {
			db.Model(ep.Post).Updates(postUpdates)
		}

		epUpdates := map[string]interface{}{}
		if input.EpisodeCoverURL != nil {
			epUpdates["episode_cover_url"] = *input.EpisodeCoverURL
		}
		if input.DurationSec != nil {
			epUpdates["duration_sec"] = *input.DurationSec
		}
		if input.SeasonNumber != nil {
			epUpdates["season_number"] = *input.SeasonNumber
		}
		if input.EpisodeNumber != nil {
			epUpdates["episode_number"] = *input.EpisodeNumber
		}
		if len(epUpdates) > 0 {
			db.Model(&ep).Updates(epUpdates)
		}

		db.Preload("Post").Preload("Channel").First(&ep, "podcast_episodes.id = ?", ep.ID)
		c.JSON(http.StatusOK, ep)
	}
}

// DeletePodcastEpisode soft-deletes the episode and its associated Post.
func DeletePodcastEpisode(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.MustGet("userID").(uuid.UUID)
		id := c.Param("id")

		var ep model.PodcastEpisode
		if err := db.Preload("Post").First(&ep, "podcast_episodes.id = ?", id).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "episode not found"})
			return
		}
		if ep.Post == nil || ep.Post.UserID != userID {
			c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
			return
		}

		db.Delete(&ep)
		db.Delete(ep.Post)
		c.JSON(http.StatusOK, gin.H{"message": "deleted"})
	}
}

// GetPodcastRSS returns a standards-compliant podcast RSS with <enclosure> tags.
func GetPodcastRSS(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		slug := c.Param("slug")
		var channel model.Channel
		if err := db.Where("slug = ?", slug).First(&channel).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "channel not found"})
			return
		}

		var episodes []model.PodcastEpisode
		db.Where("podcast_episodes.channel_id = ?", channel.ID).
			Preload("Post").
			Joins("JOIN posts ON posts.id = podcast_episodes.post_id AND posts.status = 'published'").
			Order("podcast_episodes.season_number ASC, podcast_episodes.episode_number ASC").
			Limit(100).Find(&episodes)

		scheme := c.Request.Header.Get("X-Forwarded-Proto")
		if scheme == "" {
			scheme = "https"
		}
		siteURL := fmt.Sprintf("%s://%s", scheme, c.Request.Host)

		c.Header("Content-Type", "application/rss+xml; charset=utf-8")
		c.String(http.StatusOK, buildPodcastRSS(channel, episodes, siteURL))
	}
}

func buildPodcastRSS(ch model.Channel, episodes []model.PodcastEpisode, siteURL string) string {
	coverURL := ch.CoverURL
	if coverURL == "" {
		coverURL = siteURL + "/default-podcast-cover.png"
	}

	var items strings.Builder
	for _, ep := range episodes {
		if ep.Post == nil {
			continue
		}
		pubDate := ep.CreatedAt.Format(time.RFC1123Z)
		epCover := ep.EpisodeCoverURL
		if epCover == "" {
			epCover = coverURL
		}
		items.WriteString(fmt.Sprintf(`
    <item>
      <title><![CDATA[%s]]></title>
      <link>%s/podcast/%s</link>
      <guid isPermaLink="false">%s</guid>
      <pubDate>%s</pubDate>
      <description><![CDATA[%s]]></description>
      <enclosure url="%s" length="0" type="audio/mpeg"/>
      <itunes:image href="%s"/>
      <itunes:duration>%d</itunes:duration>
      <itunes:episode>%d</itunes:episode>
      <itunes:season>%d</itunes:season>
    </item>`,
			ep.Post.Title, siteURL, ep.ID, ep.ID, pubDate,
			ep.Post.Content, ep.AudioURL, epCover,
			ep.DurationSec, ep.EpisodeNumber, ep.SeasonNumber))
	}

	return fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<rss version="2.0"
     xmlns:itunes="http://www.itunes.com/dtds/podcast-1.0.dtd"
     xmlns:content="http://purl.org/rss/1.0/modules/content/">
  <channel>
    <title><![CDATA[%s]]></title>
    <link>%s/podcast/show/%s</link>
    <description><![CDATA[%s]]></description>
    <itunes:image href="%s"/>
    <language>zh-cn</language>
    %s
  </channel>
</rss>`, ch.Name, siteURL, ch.Slug, ch.Description, coverURL, items.String())
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
```

- [ ] **Step 2: 注册路由**

在 `server/cmd/start_server/main.go` 中，`handlers.SetupVideoRoutes(...)` 之后添加：

```go
		handlers.SetupPodcastRoutes(r, db, s3Client)
```

- [ ] **Step 3: 验证编译**

```bash
cd server && go build ./...
```

期望：无错误。

- [ ] **Step 4: Commit**

```bash
git add server/internal/handlers/podcast_handler.go server/cmd/start_server/main.go
git commit -m "feat(podcast): add podcast CRUD handler, RSS with enclosure tags"
```

---

## Task 3：前端 — TypeScript 类型

**Files:**
- Modify: `web/src/types.ts`

- [ ] **Step 1: 添加 PodcastEpisode 类型**

在 `web/src/types.ts` 末尾添加：

```typescript
export interface PodcastEpisode {
  id: string
  post_id: string
  post?: Post
  channel_id: string
  channel?: Channel
  audio_url: string
  duration_sec: number
  episode_cover_url: string
  season_number: number
  episode_number: number
  created_at: string
  updated_at: string
}
```

- [ ] **Step 2: 类型检查**

```bash
cd web && bun run type-check 2>&1 | tail -5
```

期望：无新增类型错误。

- [ ] **Step 3: Commit**

```bash
git add web/src/types.ts
git commit -m "feat(podcast): add PodcastEpisode TypeScript type"
```

---

## Task 4：前端 — 发现页 + 节目页

**Files:**
- Create: `web/src/views/podcast/PodcastHomeView.vue`
- Create: `web/src/views/podcast/PodcastShowView.vue`
- Modify: `web/src/router.ts`

- [ ] **Step 1: 创建 PodcastHomeView**

```vue
<!-- web/src/views/podcast/PodcastHomeView.vue -->
<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useApi } from '@/composables/useApi'
import type { PodcastEpisode } from '@/types'

const api = useApi()
const episodes = ref<PodcastEpisode[]>([])
const loading = ref(false)

onMounted(async () => {
  loading.value = true
  try {
    const res = await api.get<PodcastEpisode[]>('/api/podcast/episodes')
    episodes.value = res.data
  } finally {
    loading.value = false
  }
})

function fmtDuration(sec: number) {
  if (!sec) return ''
  const h = Math.floor(sec / 3600)
  const m = Math.floor((sec % 3600) / 60)
  const s = sec % 60
  if (h > 0) return `${h}:${m.toString().padStart(2, '0')}:${s.toString().padStart(2, '0')}`
  return `${m}:${s.toString().padStart(2, '0')}`
}
</script>

<template>
  <div class="ph-wrap">
    <h1 class="ph-title">播客</h1>

    <div v-if="loading" class="ph-loading">加载中…</div>
    <div v-else-if="episodes.length === 0" class="ph-empty">暂无节目</div>

    <ul v-else class="ph-list">
      <li v-for="ep in episodes" :key="ep.id" class="ph-item">
        <img
          :src="ep.episode_cover_url || ep.channel?.cover_url || ''"
          class="ph-cover"
          :alt="ep.post?.title"
        />
        <div class="ph-body">
          <router-link :to="`/podcast/${ep.id}`" class="ph-ep-title">{{ ep.post?.title }}</router-link>
          <router-link
            v-if="ep.channel"
            :to="`/podcast/show/${ep.channel.slug}`"
            class="ph-show-name"
          >{{ ep.channel.name }}</router-link>
          <span v-if="ep.duration_sec" class="ph-duration">{{ fmtDuration(ep.duration_sec) }}</span>
        </div>
      </li>
    </ul>
  </div>
</template>

<style scoped>
.ph-wrap { @apply max-w-3xl mx-auto px-4 py-8; }
.ph-title { @apply text-2xl font-bold mb-6; }
.ph-loading, .ph-empty { @apply text-center py-16 text-neutral-400; }
.ph-list { @apply flex flex-col gap-3 list-none p-0; }
.ph-item { @apply flex gap-3 items-start border-b border-neutral-100 dark:border-neutral-800 pb-3; }
.ph-cover { @apply w-16 h-16 rounded object-cover shrink-0; }
.ph-body { @apply flex flex-col gap-0.5; }
.ph-ep-title { @apply text-sm font-medium no-underline hover:underline text-neutral-900 dark:text-neutral-100; }
.ph-show-name { @apply text-xs text-neutral-500 no-underline hover:underline; }
.ph-duration { @apply text-xs text-neutral-400; }
</style>
```

- [ ] **Step 2: 创建 PodcastShowView**

```vue
<!-- web/src/views/podcast/PodcastShowView.vue -->
<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import { useApi } from '@/composables/useApi'
import type { PodcastEpisode, Channel } from '@/types'

const route = useRoute()
const api = useApi()
const channel = ref<Channel | null>(null)
const episodes = ref<PodcastEpisode[]>([])
const loading = ref(true)

onMounted(async () => {
  const slug = route.params.channelSlug as string
  try {
    const res = await api.get<{ channel: Channel; episodes: PodcastEpisode[] }>(
      `/api/podcast/shows/${slug}/episodes`
    )
    channel.value = res.data.channel
    episodes.value = res.data.episodes
  } finally {
    loading.value = false
  }
})
</script>

<template>
  <div v-if="loading" class="ps-loading">加载中…</div>
  <div v-else-if="!channel" class="ps-notfound">节目不存在</div>
  <div v-else class="ps-wrap">
    <header class="ps-header">
      <img :src="channel.cover_url" class="ps-cover" :alt="channel.name" />
      <div>
        <h1 class="ps-name">{{ channel.name }}</h1>
        <p class="ps-desc">{{ channel.description }}</p>
        <a :href="`/api/channels/${channel.slug}/rss/podcast`" class="ps-rss">RSS 订阅</a>
      </div>
    </header>

    <div v-if="episodes.length === 0" class="ps-empty">暂无单集</div>
    <ul v-else class="ps-list">
      <li v-for="ep in episodes" :key="ep.id" class="ps-ep">
        <div class="ps-ep-meta">
          <span v-if="ep.episode_number" class="ps-ep-num">第 {{ ep.episode_number }} 集</span>
        </div>
        <router-link :to="`/podcast/${ep.id}`" class="ps-ep-title">
          {{ ep.post?.title }}
        </router-link>
      </li>
    </ul>
  </div>
</template>

<style scoped>
.ps-loading, .ps-notfound { @apply text-center py-24 text-neutral-400; }
.ps-wrap { @apply max-w-3xl mx-auto px-4 py-8; }
.ps-header { @apply flex gap-6 mb-8; }
.ps-cover { @apply w-32 h-32 rounded object-cover shrink-0; }
.ps-name { @apply text-2xl font-bold; }
.ps-desc { @apply text-sm text-neutral-500 mt-1; }
.ps-rss { @apply text-xs text-neutral-400 hover:underline mt-2 inline-block; }
.ps-empty { @apply text-neutral-400 py-8; }
.ps-list { @apply list-none p-0 flex flex-col gap-2; }
.ps-ep { @apply flex items-baseline gap-3 py-2 border-b border-neutral-100 dark:border-neutral-800; }
.ps-ep-num { @apply text-xs text-neutral-400 w-12 shrink-0; }
.ps-ep-title { @apply text-sm no-underline hover:underline text-neutral-800 dark:text-neutral-200; }
</style>
```

- [ ] **Step 3: 添加路由**

在 `web/src/router.ts` 中添加（以及占位文件）：

```typescript
{ path: '/podcast', component: () => import('@/views/podcast/PodcastHomeView.vue') },
{ path: '/podcast/show/:channelSlug', component: () => import('@/views/podcast/PodcastShowView.vue') },
{ path: '/podcast/:id', component: () => import('@/views/podcast/PodcastEpisodeView.vue') },
{ path: '/podcast/new', component: () => import('@/views/podcast/PodcastEditorView.vue'), meta: { requiresAuth: true } },
{ path: '/podcast/:id/edit', component: () => import('@/views/podcast/PodcastEditorView.vue'), meta: { requiresAuth: true } },
```

创建占位文件：

```vue
<!-- web/src/views/podcast/PodcastEpisodeView.vue -->
<script setup lang="ts">
// Task 5 中实现
</script>
<template><div>Loading…</div></template>
```

```vue
<!-- web/src/views/podcast/PodcastEditorView.vue -->
<script setup lang="ts">
// Task 6 中实现
</script>
<template><div>Loading…</div></template>
```

- [ ] **Step 4: 类型检查**

```bash
cd web && bun run type-check 2>&1 | tail -5
```

期望：无新增错误。

- [ ] **Step 5: Commit**

```bash
git add web/src/views/podcast/ web/src/router.ts
git commit -m "feat(podcast): add PodcastHomeView, PodcastShowView, and routes"
```

---

## Task 5：前端 — 单集收听页

**Files:**
- Modify: `web/src/views/podcast/PodcastEpisodeView.vue`

- [ ] **Step 1: 实现单集页**

播放使用原生 `<audio>` 控件，不依赖底部播放条（底部播放条的 API 绑定由二期实现）。

```vue
<!-- web/src/views/podcast/PodcastEpisodeView.vue -->
<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import { useApi } from '@/composables/useApi'
import type { PodcastEpisode } from '@/types'

const route = useRoute()
const api = useApi()
const ep = ref<PodcastEpisode | null>(null)
const loading = ref(true)

onMounted(async () => {
  const id = route.params.id as string
  try {
    const res = await api.get<PodcastEpisode>(`/api/podcast/episodes/${id}`)
    ep.value = res.data
  } finally {
    loading.value = false
  }
})
</script>

<template>
  <div v-if="loading" class="pev-loading">加载中…</div>
  <div v-else-if="!ep" class="pev-notfound">单集不存在</div>
  <div v-else class="pev-wrap">
    <!-- 封面 + 标题 -->
    <div class="pev-header">
      <img
        :src="ep.episode_cover_url || ep.channel?.cover_url || ''"
        class="pev-cover"
        :alt="ep.post?.title"
      />
      <div>
        <h1 class="pev-title">{{ ep.post?.title }}</h1>
        <router-link v-if="ep.channel" :to="`/podcast/show/${ep.channel.slug}`" class="pev-show">
          {{ ep.channel.name }}
        </router-link>
        <div class="pev-meta">
          <span v-if="ep.episode_number">第 {{ ep.episode_number }} 集</span>
          <span v-if="ep.season_number > 1">第 {{ ep.season_number }} 季</span>
        </div>
      </div>
    </div>

    <!-- 播放器 -->
    <audio :src="ep.audio_url" controls class="pev-player" />

    <!-- Shownotes -->
    <div v-if="ep.post?.content" class="pev-notes">
      <h2 class="pev-notes-title">节目说明</h2>
      <div class="pev-notes-body">{{ ep.post.content }}</div>
    </div>
  </div>
</template>

<style scoped>
.pev-loading, .pev-notfound { @apply text-center py-24 text-neutral-400; }
.pev-wrap { @apply max-w-2xl mx-auto px-4 py-8; }
.pev-header { @apply flex gap-6 mb-6; }
.pev-cover { @apply w-28 h-28 rounded object-cover shrink-0; }
.pev-title { @apply text-xl font-bold; }
.pev-show { @apply text-sm text-neutral-500 no-underline hover:underline mt-1 block; }
.pev-meta { @apply flex gap-3 text-xs text-neutral-400 mt-1; }
.pev-player { @apply w-full mt-4; }
.pev-notes { @apply mt-8; }
.pev-notes-title { @apply text-base font-semibold mb-2; }
.pev-notes-body { @apply text-sm text-neutral-600 dark:text-neutral-400 whitespace-pre-wrap; }
</style>
```

- [ ] **Step 2: 类型检查**

```bash
cd web && bun run type-check 2>&1 | tail -5
```

期望：无新增错误。

- [ ] **Step 3: Commit**

```bash
git add web/src/views/podcast/PodcastEpisodeView.vue
git commit -m "feat(podcast): implement PodcastEpisodeView with audio player"
```

---

## Task 6：前端 — 单集发布/编辑页

**Files:**
- Modify: `web/src/views/podcast/PodcastEditorView.vue`

- [ ] **Step 1: 实现 PodcastEditorView**

```vue
<!-- web/src/views/podcast/PodcastEditorView.vue -->
<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useApi } from '@/composables/useApi'
import type { PodcastEpisode, Channel } from '@/types'

const route = useRoute()
const router = useRouter()
const api = useApi()

const isEdit = computed(() => !!route.params.id)
const saving = ref(false)
const error = ref('')
const channels = ref<Channel[]>([])

const form = ref({
  channel_id: '',
  title: '',
  shownotes: '',
  audio_url: '',
  duration_sec: 0,
  episode_cover_url: '',
  season_number: 1,
  episode_number: 0,
  status: 'draft' as 'draft' | 'published',
})

async function loadChannels() {
  const res = await api.get<Channel[]>('/api/channels/mine')
  channels.value = res.data
  if (!form.value.channel_id && channels.value.length > 0) {
    form.value.channel_id = channels.value[0].id
  }
}

async function loadEpisode() {
  const id = route.params.id as string
  const res = await api.get<PodcastEpisode>(`/api/podcast/episodes/${id}`)
  const ep = res.data
  form.value = {
    channel_id: ep.channel_id,
    title: ep.post?.title || '',
    shownotes: ep.post?.content || '',
    audio_url: ep.audio_url,
    duration_sec: ep.duration_sec,
    episode_cover_url: ep.episode_cover_url,
    season_number: ep.season_number,
    episode_number: ep.episode_number,
    status: (ep.post?.status as 'draft' | 'published') || 'draft',
  }
}

onMounted(async () => {
  await loadChannels()
  if (isEdit.value) await loadEpisode()
})

async function save(publish = false) {
  saving.value = true
  error.value = ''
  const payload = { ...form.value, status: publish ? 'published' : form.value.status }
  try {
    if (isEdit.value) {
      await api.put(`/api/podcast/episodes/${route.params.id}`, payload)
      router.push(`/podcast/${route.params.id}`)
    } else {
      const res = await api.post<PodcastEpisode>('/api/podcast/episodes', payload)
      router.push(`/podcast/${res.data.id}`)
    }
  } catch (e: any) {
    error.value = e?.response?.data?.error || '保存失败'
  } finally {
    saving.value = false
  }
}
</script>

<template>
  <div class="pe-wrap">
    <h1 class="pe-title">{{ isEdit ? '编辑单集' : '发布新单集' }}</h1>

    <form class="pe-form" @submit.prevent="save()">
      <!-- 节目 (频道) -->
      <label class="pe-label">节目（频道）*
        <select v-model="form.channel_id" required class="pe-select">
          <option v-for="ch in channels" :key="ch.id" :value="ch.id">{{ ch.name }}</option>
        </select>
      </label>

      <!-- 标题 -->
      <label class="pe-label">单集标题 *
        <input v-model="form.title" required class="pe-input" placeholder="单集标题" />
      </label>

      <!-- 音频链接（本地上传后的 S3 URL，需先通过上传接口获取） -->
      <label class="pe-label">音频文件 URL *
        <input v-model="form.audio_url" required class="pe-input" placeholder="https://..." />
        <span class="pe-hint">请先通过上传接口获取音频 URL，再填入此处</span>
      </label>

      <!-- 时长 -->
      <label class="pe-label">时长（秒）
        <input v-model.number="form.duration_sec" type="number" min="0" class="pe-input" />
      </label>

      <!-- 单集封面 -->
      <label class="pe-label">单集封面 URL（可选，不填则使用节目封面）
        <input v-model="form.episode_cover_url" class="pe-input" placeholder="https://..." />
      </label>

      <!-- 季/集编号 -->
      <div class="pe-row">
        <label class="pe-label">季
          <input v-model.number="form.season_number" type="number" min="1" class="pe-input" />
        </label>
        <label class="pe-label">集
          <input v-model.number="form.episode_number" type="number" min="0" class="pe-input" />
        </label>
      </div>

      <!-- Shownotes（纯文本 + URL + 加粗） -->
      <label class="pe-label">节目说明（Shownotes）
        <textarea v-model="form.shownotes" class="pe-textarea" rows="6"
          placeholder="支持 URL 和 **加粗** 语法" />
      </label>

      <p v-if="error" class="pe-error">{{ error }}</p>

      <div class="pe-actions">
        <button type="submit" :disabled="saving" class="pe-btn pe-btn--secondary">
          {{ saving ? '保存中…' : '保存草稿' }}
        </button>
        <button type="button" :disabled="saving" class="pe-btn pe-btn--primary" @click="save(true)">
          发布
        </button>
      </div>
    </form>
  </div>
</template>

<style scoped>
.pe-wrap { @apply max-w-2xl mx-auto px-4 py-8; }
.pe-title { @apply text-2xl font-bold mb-6; }
.pe-form { @apply flex flex-col gap-4; }
.pe-label { @apply flex flex-col gap-1 text-sm font-medium; }
.pe-input, .pe-select, .pe-textarea {
  @apply mt-1 px-3 py-2 border border-neutral-300 dark:border-neutral-600 rounded text-sm
         bg-white dark:bg-neutral-900 focus:outline-none focus:border-neutral-600;
}
.pe-textarea { @apply resize-y; }
.pe-hint { @apply text-xs text-neutral-400 mt-0.5; }
.pe-row { @apply flex gap-4; }
.pe-row .pe-label { @apply flex-1; }
.pe-error { @apply text-red-500 text-sm; }
.pe-actions { @apply flex gap-3 pt-2; }
.pe-btn { @apply px-4 py-2 rounded text-sm font-medium disabled:opacity-50 transition-colors; }
.pe-btn--primary { @apply bg-neutral-900 text-white hover:bg-neutral-700 dark:bg-neutral-100 dark:text-neutral-900; }
.pe-btn--secondary { @apply border border-neutral-300 hover:border-neutral-600; }
</style>
```

- [ ] **Step 2: 类型检查**

```bash
cd web && bun run type-check 2>&1 | tail -5
```

期望：无新增错误。

- [ ] **Step 3: Commit**

```bash
git add web/src/views/podcast/PodcastEditorView.vue
git commit -m "feat(podcast): implement PodcastEditorView"
```

---

## Task 7：验收

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

- [ ] **Step 3: 手动冒烟测试（dev 环境）**

1. 访问 `/podcast` — 发现页（空列表或已有单集）
2. 访问 `/podcast/new`（已登录）— 应看到发布表单
3. 填写标题、音频 URL，点发布 → 跳转到 `/podcast/:id` 收听页
4. 收听页应显示 `<audio>` 播放器，Shownotes 在下方
5. 访问 `/podcast/show/:channelSlug` — 节目页显示该频道所有单集
6. 访问 `/api/channels/:slug/rss/podcast` — 返回含 `<enclosure>` 的 RSS XML
7. 将 RSS URL 粘贴到本地播客客户端（如 Overcast/Pocket Casts）验证解析正确

- [ ] **Step 4: Final commit**

```bash
git add -A
git commit -m "feat(podcast): complete Podcast module - models, CRUD, RSS, frontend views"
```

---

## 二期待做（不在本计划内）

| 功能 | 说明 |
|------|------|
| 底部播放条接入 | 复用 `AudioPlayer.vue`，单集页点播放加入全局队列 |
| Feed 模块接入播客 RSS | FeedSource 支持解析 `<enclosure>` 类型 |
| 跨节目收听队列 | 播放条管理多节目队列 |
