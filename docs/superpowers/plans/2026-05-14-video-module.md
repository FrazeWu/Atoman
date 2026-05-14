# Video Module Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 实现 Video 模块，支持创作者通过频道发布视频（本地上传或外链），提供发现页、播放页、管理页，以及混合推荐算法和频道合集集成。

**Architecture:** Video 复用 Studio 频道体系（`Channel`）和 `Collection` many2many 关系，模型结构对齐 `Post`。后端新增 `video_handler.go`，前端新增三个视图和一个卡片组件。推荐算法用 SQL 简单打分（同频道 60% + 同标签 40%），无需外部服务。

**Tech Stack:** Go 1.21 / Gin / GORM / PostgreSQL；Vue 3.5 / TypeScript 5.9 / Tailwind CSS v4；S3/MinIO（视频文件本地上传）

---

## 文件清单

### 新增
| 文件 | 职责 |
|------|------|
| `server/internal/model/video.go` | Video + VideoCollection + VideoTag 模型 |
| `server/internal/handlers/video_handler.go` | Video CRUD、发现页、推荐、RSS |
| `web/src/views/video/VideoHomeView.vue` | 视频发现页 `/video` |
| `web/src/views/video/VideoDetailView.vue` | 播放页 `/video/:id` |
| `web/src/views/video/VideoEditorView.vue` | 创建/编辑页 `/video/new` `/video/:id/edit` |
| `web/src/components/shared/VideoCard.vue` | 视频卡片（复用于发现页/推荐栏） |

### 修改
| 文件 | 改动 |
|------|------|
| `server/cmd/start_server/main.go` | 注册 Video 模型到 AutoMigrate，注册路由 |
| `web/src/router.ts` | 添加 video 路由 |
| `web/src/types.ts` | 添加 `Video` / `VideoTag` 类型 |

---

## Task 1：后端数据模型

**Files:**
- Create: `server/internal/model/video.go`
- Modify: `server/cmd/start_server/main.go:186-240`（AutoMigrate 列表）

- [ ] **Step 1: 创建 model 文件**

```go
// server/internal/model/video.go
package model

import "github.com/google/uuid"

// Video represents a video post published under a channel.
type Video struct {
	Base
	ChannelID    *uuid.UUID `json:"channel_id,omitempty" gorm:"type:uuid;index"`
	Channel      *Channel   `json:"channel,omitempty" gorm:"foreignKey:ChannelID"`
	UserID       uuid.UUID  `json:"user_id" gorm:"type:uuid;not null;index"`
	User         *User      `json:"user,omitempty" gorm:"foreignKey:UserID;references:UUID"`
	Title        string     `json:"title" gorm:"not null"`
	Description  string     `json:"description" gorm:"type:text"`
	// Storage: "local" (S3/MinIO) or "external" (YouTube, Bilibili, etc.)
	StorageType  string     `json:"storage_type" gorm:"not null;default:'external'"` // local | external
	VideoURL     string     `json:"video_url" gorm:"type:text;not null"`             // S3 key or external URL
	ThumbnailURL string     `json:"thumbnail_url" gorm:"type:text"`
	DurationSec  int        `json:"duration_sec" gorm:"default:0"`
	// Visibility: public | followers | private
	Visibility   string     `json:"visibility" gorm:"not null;default:'public'"`
	Status       string     `json:"status" gorm:"not null;default:'draft'"` // draft | published | scheduled
	ScheduledAt  *int64     `json:"scheduled_at,omitempty"`                 // Unix timestamp
	ViewCount    int        `json:"view_count" gorm:"default:0"`
	SubtitleURL  string     `json:"subtitle_url" gorm:"type:text"`
	Tags         []VideoTag `json:"tags,omitempty" gorm:"many2many:video_tag_relations;joinForeignKey:VideoID;joinReferences:TagID"`
	Collections  []Collection `json:"collections,omitempty" gorm:"many2many:video_collections;"`
}

func (Video) TableName() string { return "videos" }

// VideoTag is a reusable tag for video discovery.
type VideoTag struct {
	Base
	Name string `json:"name" gorm:"uniqueIndex;not null"`
}

func (VideoTag) TableName() string { return "video_tags" }

// VideoCollection is the join table between Video and Collection.
type VideoCollection struct {
	VideoID      uuid.UUID `json:"video_id" gorm:"type:uuid;primaryKey"`
	CollectionID uuid.UUID `json:"collection_id" gorm:"type:uuid;primaryKey"`
}

func (VideoCollection) TableName() string { return "video_collections" }

// VideoTagRelation is the join table between Video and VideoTag.
type VideoTagRelation struct {
	VideoID uuid.UUID `json:"video_id" gorm:"type:uuid;primaryKey"`
	TagID   uuid.UUID `json:"tag_id" gorm:"type:uuid;primaryKey"`
}

func (VideoTagRelation) TableName() string { return "video_tag_relations" }
```

- [ ] **Step 2: 注册 AutoMigrate**

在 `server/cmd/start_server/main.go` 的 `AutoMigrate` 调用中，在 `&model.ForumCategory{}` 之前添加：

```go
			&model.Video{},
			&model.VideoTag{},
			&model.VideoCollection{},
			&model.VideoTagRelation{},
```

- [ ] **Step 3: 验证编译**

```bash
cd server && go build ./...
```

期望输出：无错误，二进制生成成功。

- [ ] **Step 4: 验证迁移（dev 环境）**

```bash
go run cmd/start_server/main.go 2>&1 | grep -E "migration|video"
```

期望输出：`Running database migrations... Database migrations completed`，日志中无 Fatal。

- [ ] **Step 5: Commit**

```bash
git add server/internal/model/video.go server/cmd/start_server/main.go
git commit -m "feat(video): add Video, VideoTag, VideoCollection models"
```

---

## Task 2：后端 Handler — CRUD 基础

**Files:**
- Create: `server/internal/handlers/video_handler.go`
- Modify: `server/cmd/start_server/main.go`（路由注册）

- [ ] **Step 1: 创建 Handler 文件（路由注册 + 基础 CRUD）**

```go
// server/internal/handlers/video_handler.go
package handlers

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"

	"atoman/internal/middleware"
	"atoman/internal/model"
	"atoman/internal/storage"
)

func SetupVideoRoutes(router *gin.Engine, db *gorm.DB, s3Client *storage.S3Client) {
	v := router.Group("/api/videos")
	{
		v.GET("", GetVideos(db))
		v.GET("/:id", GetVideo(db))
		v.GET("/:id/recommended", GetRecommendedVideos(db))
		v.POST("", middleware.RequireAuth(), CreateVideo(db, s3Client))
		v.PUT("/:id", middleware.RequireAuth(), UpdateVideo(db, s3Client))
		v.DELETE("/:id", middleware.RequireAuth(), DeleteVideo(db))
		v.POST("/:id/view", IncrementVideoView(db))
	}
	// Per-channel RSS
	router.GET("/api/channels/:slug/rss/video", GetVideoRSS(db))
}

// GetVideos returns published videos, supports ?channel_id=&tag=&sort=latest|popular
func GetVideos(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		channelID := c.Query("channel_id")
		tag := c.Query("tag")
		sort := c.DefaultQuery("sort", "latest")

		q := db.Model(&model.Video{}).
			Where("status = ?", "published").
			Where("visibility = ?", "public").
			Preload("Channel").
			Preload("Tags")

		if channelID != "" {
			q = q.Where("channel_id = ?", channelID)
		}
		if tag != "" {
			q = q.Joins("JOIN video_tag_relations vtr ON vtr.video_id = videos.id").
				Joins("JOIN video_tags vt ON vt.id = vtr.tag_id AND vt.name = ?", tag)
		}
		if sort == "popular" {
			q = q.Order("view_count DESC")
		} else {
			q = q.Order("videos.created_at DESC")
		}

		var videos []model.Video
		if err := q.Limit(40).Find(&videos).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, videos)
	}
}

// GetVideo returns a single video by ID.
func GetVideo(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		var video model.Video
		if err := db.Preload("Channel").Preload("Tags").Preload("Collections").
			First(&video, "id = ?", id).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "video not found"})
			return
		}
		c.JSON(http.StatusOK, video)
	}
}

// CreateVideo creates a new video (draft or published).
func CreateVideo(db *gorm.DB, s3Client *storage.S3Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.MustGet("userID").(uuid.UUID)
		var input struct {
			ChannelID    *uuid.UUID `json:"channel_id"`
			Title        string     `json:"title" binding:"required"`
			Description  string     `json:"description"`
			StorageType  string     `json:"storage_type"` // local | external
			VideoURL     string     `json:"video_url" binding:"required"`
			ThumbnailURL string     `json:"thumbnail_url"`
			DurationSec  int        `json:"duration_sec"`
			Visibility   string     `json:"visibility"`
			Status       string     `json:"status"`
			Tags         []string   `json:"tags"`
		}
		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		storageType := input.StorageType
		if storageType == "" {
			storageType = "external"
		}
		visibility := input.Visibility
		if visibility == "" {
			visibility = "public"
		}
		status := input.Status
		if status == "" {
			status = "draft"
		}

		video := model.Video{
			UserID:       userID,
			ChannelID:    input.ChannelID,
			Title:        strings.TrimSpace(input.Title),
			Description:  input.Description,
			StorageType:  storageType,
			VideoURL:     input.VideoURL,
			ThumbnailURL: input.ThumbnailURL,
			DurationSec:  input.DurationSec,
			Visibility:   visibility,
			Status:       status,
		}

		if err := db.Create(&video).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// Attach tags
		if len(input.Tags) > 0 {
			if err := attachVideoTags(db, &video, input.Tags); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "tags failed: " + err.Error()})
				return
			}
		}

		db.Preload("Channel").Preload("Tags").First(&video, "id = ?", video.ID)
		c.JSON(http.StatusCreated, video)
	}
}

// UpdateVideo updates title, description, tags, visibility, status, thumbnail, etc.
func UpdateVideo(db *gorm.DB, s3Client *storage.S3Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.MustGet("userID").(uuid.UUID)
		id := c.Param("id")

		var video model.Video
		if err := db.First(&video, "id = ?", id).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "video not found"})
			return
		}
		if video.UserID != userID {
			c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
			return
		}

		var input struct {
			Title        *string  `json:"title"`
			Description  *string  `json:"description"`
			ThumbnailURL *string  `json:"thumbnail_url"`
			Visibility   *string  `json:"visibility"`
			Status       *string  `json:"status"`
			Tags         []string `json:"tags"`
		}
		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		updates := map[string]interface{}{}
		if input.Title != nil {
			updates["title"] = strings.TrimSpace(*input.Title)
		}
		if input.Description != nil {
			updates["description"] = *input.Description
		}
		if input.ThumbnailURL != nil {
			updates["thumbnail_url"] = *input.ThumbnailURL
		}
		if input.Visibility != nil {
			updates["visibility"] = *input.Visibility
		}
		if input.Status != nil {
			updates["status"] = *input.Status
		}

		if len(updates) > 0 {
			if err := db.Model(&video).Updates(updates).Error; err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
		}

		if input.Tags != nil {
			if err := db.Model(&video).Association("Tags").Unscoped().Clear(); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			if len(input.Tags) > 0 {
				if err := attachVideoTags(db, &video, input.Tags); err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
					return
				}
			}
		}

		db.Preload("Channel").Preload("Tags").First(&video, "id = ?", video.ID)
		c.JSON(http.StatusOK, video)
	}
}

// DeleteVideo soft-deletes a video.
func DeleteVideo(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.MustGet("userID").(uuid.UUID)
		id := c.Param("id")

		var video model.Video
		if err := db.First(&video, "id = ?", id).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "video not found"})
			return
		}
		if video.UserID != userID {
			c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
			return
		}

		db.Delete(&video)
		c.JSON(http.StatusOK, gin.H{"message": "deleted"})
	}
}

// IncrementVideoView adds 1 to view_count. No auth required.
func IncrementVideoView(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		db.Model(&model.Video{}).Where("id = ?", id).
			UpdateColumn("view_count", gorm.Expr("view_count + 1"))
		c.JSON(http.StatusOK, gin.H{"ok": true})
	}
}

// attachVideoTags upserts VideoTag rows and links them to the video.
func attachVideoTags(db *gorm.DB, video *model.Video, names []string) error {
	var tags []model.VideoTag
	for _, name := range names {
		name = strings.TrimSpace(name)
		if name == "" {
			continue
		}
		var tag model.VideoTag
		db.Where("name = ?", name).FirstOrCreate(&tag, model.VideoTag{Name: name})
		tags = append(tags, tag)
	}
	return db.Model(video).Association("Tags").Append(tags)
}
```

- [ ] **Step 2: 注册路由**

在 `server/cmd/start_server/main.go` 的路由注册区（`SetupTimelineRoutes` 之后）添加：

```go
		handlers.SetupVideoRoutes(r, db, s3Client)
```

同时在 import 区（已有 `s3Client`，无需单独添加）确认 `atoman/internal/handlers` 已被导入。

- [ ] **Step 3: 验证编译**

```bash
cd server && go build ./...
```

期望：无错误。

- [ ] **Step 4: 快速 smoke test**

启动开发服务器后执行：

```bash
curl -s http://localhost:8080/api/videos | jq type
```

期望输出：`"array"`

- [ ] **Step 5: Commit**

```bash
git add server/internal/handlers/video_handler.go server/cmd/start_server/main.go
git commit -m "feat(video): add video CRUD handler and route registration"
```

---

## Task 3：后端 — 推荐算法 + RSS

**Files:**
- Modify: `server/internal/handlers/video_handler.go`（补充 GetRecommendedVideos + GetVideoRSS）

- [ ] **Step 1: 添加推荐 handler**

在 `video_handler.go` 末尾追加：

```go
// GetRecommendedVideos returns up to 8 recommended videos.
// 策略：同频道 video 优先（权重 60），同标签 video 次之（权重 40），去掉自身，按得分降序。
func GetRecommendedVideos(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		var source model.Video
		if err := db.Preload("Tags").First(&source, "id = ?", id).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "video not found"})
			return
		}

		var tagIDs []uuid.UUID
		for _, t := range source.Tags {
			tagIDs = append(tagIDs, t.ID)
		}

		// Build scored candidate list via UNION + subquery in raw SQL for portability.
		// Same-channel: score = 60; same-tag: score += 40 per shared tag (capped at 40).
		type scored struct {
			model.Video
			Score int `json:"score"`
		}

		var results []model.Video

		q := db.Model(&model.Video{}).
			Where("videos.id <> ?", id).
			Where("videos.status = ?", "published").
			Where("videos.visibility = ?", "public")

		var channelCandidates, tagCandidates []model.Video

		if source.ChannelID != nil {
			db.Model(&model.Video{}).
				Where("channel_id = ? AND id <> ? AND status = ? AND visibility = ?",
					source.ChannelID, id, "published", "public").
				Preload("Tags").Limit(20).Find(&channelCandidates)
		}

		if len(tagIDs) > 0 {
			db.Model(&model.Video{}).
				Joins("JOIN video_tag_relations vtr ON vtr.video_id = videos.id").
				Where("vtr.tag_id IN ? AND videos.id <> ? AND videos.status = ? AND videos.visibility = ?",
					tagIDs, id, "published", "public").
				Preload("Tags").Limit(20).Find(&tagCandidates)
		}

		// Score and deduplicate
		scores := map[uuid.UUID]int{}
		seen := map[uuid.UUID]model.Video{}

		for _, v := range channelCandidates {
			scores[v.ID] += 60
			seen[v.ID] = v
		}
		for _, v := range tagCandidates {
			scores[v.ID] += 40
			seen[v.ID] = v
		}

		// If no candidates, fallback: latest public videos
		if len(seen) == 0 {
			q.Order("created_at DESC").Preload("Channel").Preload("Tags").Limit(8).Find(&results)
			c.JSON(http.StatusOK, results)
			return
		}

		// Sort by score desc, take top 8
		type scoredID struct {
			id    uuid.UUID
			score int
		}
		var ranked []scoredID
		for id, score := range scores {
			ranked = append(ranked, scoredID{id, score})
		}
		// Simple insertion sort (small N)
		for i := 1; i < len(ranked); i++ {
			for j := i; j > 0 && ranked[j].score > ranked[j-1].score; j-- {
				ranked[j], ranked[j-1] = ranked[j-1], ranked[j]
			}
		}
		if len(ranked) > 8 {
			ranked = ranked[:8]
		}
		for _, r := range ranked {
			v := seen[r.id]
			results = append(results, v)
		}

		// Preload Channel for results
		for i := range results {
			db.Preload("Channel").First(&results[i], "id = ?", results[i].ID)
		}

		c.JSON(http.StatusOK, results)
	}
}
```

- [ ] **Step 2: 添加 RSS handler**

```go
import (
	"fmt"
	"time"
)

// GetVideoRSS outputs a Media RSS feed for all published videos in a channel.
func GetVideoRSS(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		slug := c.Param("slug")
		var channel model.Channel
		if err := db.Where("slug = ?", slug).First(&channel).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "channel not found"})
			return
		}

		var videos []model.Video
		db.Where("channel_id = ? AND status = ? AND visibility = ?",
			channel.ID, "published", "public").
			Order("created_at DESC").Limit(100).Find(&videos)

		baseURL := c.Request.Header.Get("X-Forwarded-Proto")
		if baseURL == "" {
			baseURL = "https"
		}
		host := c.Request.Host
		siteURL := fmt.Sprintf("%s://%s", baseURL, host)

		c.Header("Content-Type", "application/rss+xml; charset=utf-8")
		c.String(http.StatusOK, buildVideoRSS(channel, videos, siteURL))
	}
}

func buildVideoRSS(ch model.Channel, videos []model.Video, siteURL string) string {
	var items strings.Builder
	for _, v := range videos {
		pubDate := v.CreatedAt.Format(time.RFC1123Z)
		videoURL := v.VideoURL
		enclosure := ""
		if v.StorageType == "local" {
			enclosure = fmt.Sprintf(`<enclosure url="%s" type="video/mp4"/>`, videoURL)
		}
		items.WriteString(fmt.Sprintf(`
    <item>
      <title><![CDATA[%s]]></title>
      <link>%s/video/%s</link>
      <guid>%s/video/%s</guid>
      <pubDate>%s</pubDate>
      <description><![CDATA[%s]]></description>
      %s
    </item>`, v.Title, siteURL, v.ID, siteURL, v.ID, pubDate, v.Description, enclosure))
	}

	return fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<rss version="2.0">
  <channel>
    <title><![CDATA[%s - Videos]]></title>
    <link>%s/channel/%s</link>
    <description><![CDATA[%s]]></description>
    %s
  </channel>
</rss>`, ch.Name, siteURL, ch.Slug, ch.Description, items.String())
}
```

注意：在文件顶部 import 块中补充 `"fmt"` 和 `"time"`。

- [ ] **Step 3: 验证编译**

```bash
cd server && go build ./...
```

期望：无错误。

- [ ] **Step 4: Commit**

```bash
git add server/internal/handlers/video_handler.go
git commit -m "feat(video): add recommended videos algorithm and RSS feed"
```

---

## Task 4：前端 — TypeScript 类型

**Files:**
- Modify: `web/src/types.ts`

- [ ] **Step 1: 添加 Video 相关类型**

在 `web/src/types.ts` 的末尾添加：

```typescript
export interface VideoTag {
  id: string
  name: string
}

export interface Video {
  id: string
  channel_id: string | null
  channel?: Channel
  user_id: string
  title: string
  description: string
  storage_type: 'local' | 'external'
  video_url: string
  thumbnail_url: string
  duration_sec: number
  visibility: 'public' | 'followers' | 'private'
  status: 'draft' | 'published' | 'scheduled'
  scheduled_at?: number
  view_count: number
  subtitle_url: string
  tags: VideoTag[]
  collections?: Collection[]
  created_at: string
  updated_at: string
}
```

- [ ] **Step 2: 类型检查**

```bash
cd web && bun run type-check 2>&1 | tail -5
```

期望输出：无新增类型错误。

- [ ] **Step 3: Commit**

```bash
git add web/src/types.ts
git commit -m "feat(video): add Video and VideoTag TypeScript types"
```

---

## Task 5：前端 — 视频发现页

**Files:**
- Create: `web/src/views/video/VideoHomeView.vue`
- Create: `web/src/components/shared/VideoCard.vue`
- Modify: `web/src/router.ts`

- [ ] **Step 1: 创建 VideoCard 组件**

```vue
<!-- web/src/components/shared/VideoCard.vue -->
<script setup lang="ts">
import type { Video } from '@/types'

defineProps<{ video: Video }>()

function formatDuration(sec: number): string {
  if (!sec) return ''
  const m = Math.floor(sec / 60)
  const s = sec % 60
  return `${m}:${s.toString().padStart(2, '0')}`
}
</script>

<template>
  <router-link :to="`/video/${video.id}`" class="video-card">
    <div class="video-card__thumb">
      <img v-if="video.thumbnail_url" :src="video.thumbnail_url" :alt="video.title" />
      <div v-else class="video-card__thumb-placeholder">VIDEO</div>
      <span v-if="video.duration_sec" class="video-card__duration">{{ formatDuration(video.duration_sec) }}</span>
    </div>
    <div class="video-card__body">
      <p class="video-card__title">{{ video.title }}</p>
      <p v-if="video.channel" class="video-card__channel">{{ video.channel.name }}</p>
      <p class="video-card__meta">{{ video.view_count }} 次播放</p>
    </div>
  </router-link>
</template>

<style scoped>
.video-card {
  @apply block rounded overflow-hidden border border-neutral-200 dark:border-neutral-700 hover:border-neutral-400 transition-colors no-underline;
}
.video-card__thumb {
  @apply relative aspect-video bg-neutral-100 dark:bg-neutral-800;
}
.video-card__thumb img {
  @apply w-full h-full object-cover;
}
.video-card__thumb-placeholder {
  @apply absolute inset-0 flex items-center justify-center text-neutral-400 text-xs font-mono;
}
.video-card__duration {
  @apply absolute bottom-1 right-1 bg-black/70 text-white text-xs px-1 rounded;
}
.video-card__body {
  @apply p-2;
}
.video-card__title {
  @apply text-sm font-medium line-clamp-2 text-neutral-900 dark:text-neutral-100;
}
.video-card__channel {
  @apply text-xs text-neutral-500 mt-0.5;
}
.video-card__meta {
  @apply text-xs text-neutral-400 mt-0.5;
}
</style>
```

- [ ] **Step 2: 创建 VideoHomeView**

```vue
<!-- web/src/views/video/VideoHomeView.vue -->
<script setup lang="ts">
import { ref, onMounted, watch } from 'vue'
import { useRouter } from 'vue-router'
import { useApi } from '@/composables/useApi'
import type { Video } from '@/types'
import VideoCard from '@/components/shared/VideoCard.vue'

const api = useApi()
const videos = ref<Video[]>([])
const loading = ref(false)
const sort = ref<'latest' | 'popular'>('latest')

async function fetchVideos() {
  loading.value = true
  try {
    const res = await api.get<Video[]>(`/api/videos?sort=${sort.value}`)
    videos.value = res.data
  } finally {
    loading.value = false
  }
}

onMounted(fetchVideos)
watch(sort, fetchVideos)
</script>

<template>
  <div class="video-home">
    <header class="video-home__header">
      <h1 class="video-home__title">视频</h1>
      <div class="video-home__sort">
        <button
          v-for="s in (['latest', 'popular'] as const)"
          :key="s"
          :class="['sort-btn', { active: sort === s }]"
          @click="sort = s"
        >{{ s === 'latest' ? '最新' : '最热' }}</button>
      </div>
    </header>

    <div v-if="loading" class="video-home__loading">加载中…</div>

    <div v-else-if="videos.length === 0" class="video-home__empty">暂无视频</div>

    <div v-else class="video-home__grid">
      <VideoCard v-for="v in videos" :key="v.id" :video="v" />
    </div>
  </div>
</template>

<style scoped>
.video-home { @apply max-w-6xl mx-auto px-4 py-8; }
.video-home__header { @apply flex items-center justify-between mb-6; }
.video-home__title { @apply text-2xl font-bold; }
.video-home__sort { @apply flex gap-2; }
.sort-btn { @apply px-3 py-1 text-sm border border-neutral-300 rounded hover:border-neutral-600 transition-colors; }
.sort-btn.active { @apply border-neutral-900 dark:border-neutral-100 font-medium; }
.video-home__loading, .video-home__empty { @apply text-center py-16 text-neutral-400; }
.video-home__grid { @apply grid grid-cols-2 sm:grid-cols-3 md:grid-cols-4 gap-4; }
</style>
```

- [ ] **Step 3: 添加路由**

在 `web/src/router.ts` 中添加：

```typescript
{ path: '/video', component: () => import('@/views/video/VideoHomeView.vue') },
{ path: '/video/new', component: () => import('@/views/video/VideoEditorView.vue'), meta: { requiresAuth: true } },
{ path: '/video/:id', component: () => import('@/views/video/VideoDetailView.vue') },
{ path: '/video/:id/edit', component: () => import('@/views/video/VideoEditorView.vue'), meta: { requiresAuth: true } },
```

（VideoDetailView 和 VideoEditorView 将在后续 Task 中创建，这里先添加路由，留空的视图文件在下一步创建。）

先创建占位文件以通过类型检查：

```vue
<!-- web/src/views/video/VideoDetailView.vue -->
<script setup lang="ts">
// Task 6 中实现
</script>
<template><div>Loading…</div></template>
```

```vue
<!-- web/src/views/video/VideoEditorView.vue -->
<script setup lang="ts">
// Task 7 中实现
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
git add web/src/views/video/ web/src/components/shared/VideoCard.vue web/src/router.ts
git commit -m "feat(video): add VideoHomeView, VideoCard component, and routes"
```

---

## Task 6：前端 — 播放详情页

**Files:**
- Modify: `web/src/views/video/VideoDetailView.vue`

- [ ] **Step 1: 实现 VideoDetailView**

```vue
<!-- web/src/views/video/VideoDetailView.vue -->
<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { useRoute } from 'vue-router'
import { useApi } from '@/composables/useApi'
import type { Video } from '@/types'
import VideoCard from '@/components/shared/VideoCard.vue'

const route = useRoute()
const api = useApi()
const video = ref<Video | null>(null)
const recommended = ref<Video[]>([])
const loading = ref(true)

// Detect if the URL is a YouTube or Bilibili embed
const embedSrc = computed(() => {
  const url = video.value?.video_url || ''
  const ytMatch = url.match(/(?:youtube\.com\/watch\?v=|youtu\.be\/)([A-Za-z0-9_-]{11})/)
  if (ytMatch) return `https://www.youtube.com/embed/${ytMatch[1]}`
  const biliMatch = url.match(/bilibili\.com\/video\/(BV[A-Za-z0-9]+)/)
  if (biliMatch) return `https://player.bilibili.com/player.html?bvid=${biliMatch[1]}&autoplay=0`
  return null
})

async function load() {
  const id = route.params.id as string
  loading.value = true
  try {
    const [vRes, rRes] = await Promise.all([
      api.get<Video>(`/api/videos/${id}`),
      api.get<Video[]>(`/api/videos/${id}/recommended`),
    ])
    video.value = vRes.data
    recommended.value = rRes.data
    // Fire-and-forget view count increment
    api.post(`/api/videos/${id}/view`)
  } finally {
    loading.value = false
  }
}

onMounted(load)
</script>

<template>
  <div v-if="loading" class="vd-loading">加载中…</div>

  <div v-else-if="!video" class="vd-notfound">视频不存在</div>

  <div v-else class="vd-layout">
    <!-- 左栏：播放器 + 信息 -->
    <div class="vd-main">
      <!-- 播放器 -->
      <div class="vd-player">
        <template v-if="embedSrc">
          <iframe
            :src="embedSrc"
            allow="accelerometer; autoplay; clipboard-write; encrypted-media; gyroscope; picture-in-picture"
            allowfullscreen
            class="vd-iframe"
          />
        </template>
        <template v-else-if="video.storage_type === 'local'">
          <video :src="video.video_url" controls class="vd-native-video" />
        </template>
        <template v-else>
          <div class="vd-external-fallback">
            <a :href="video.video_url" target="_blank" rel="noopener">在外部平台观看 →</a>
          </div>
        </template>
      </div>

      <!-- 标题 & 元信息 -->
      <div class="vd-info">
        <h1 class="vd-title">{{ video.title }}</h1>
        <div class="vd-meta">
          <router-link v-if="video.channel" :to="`/channel/${video.channel.slug}`" class="vd-channel">
            {{ video.channel.name }}
          </router-link>
          <span class="vd-views">{{ video.view_count }} 次播放</span>
        </div>
        <div class="vd-tags">
          <span v-for="tag in video.tags" :key="tag.id" class="vd-tag">{{ tag.name }}</span>
        </div>
        <div v-if="video.description" class="vd-desc" v-html="video.description" />
      </div>
    </div>

    <!-- 右栏：推荐视频 -->
    <aside class="vd-sidebar">
      <h2 class="vd-sidebar-title">推荐</h2>
      <div class="vd-recommended">
        <VideoCard v-for="v in recommended" :key="v.id" :video="v" />
      </div>
    </aside>
  </div>
</template>

<style scoped>
.vd-loading, .vd-notfound { @apply text-center py-24 text-neutral-400; }
.vd-layout { @apply max-w-7xl mx-auto px-4 py-8 flex gap-6; }
.vd-main { @apply flex-1 min-w-0; }
.vd-player { @apply relative w-full bg-black rounded overflow-hidden; }
.vd-iframe { @apply w-full aspect-video border-0; }
.vd-native-video { @apply w-full aspect-video; }
.vd-external-fallback { @apply aspect-video flex items-center justify-center text-white bg-neutral-900; }
.vd-info { @apply mt-4; }
.vd-title { @apply text-xl font-bold; }
.vd-meta { @apply flex items-center gap-4 mt-2 text-sm text-neutral-500; }
.vd-channel { @apply font-medium text-neutral-700 dark:text-neutral-300 no-underline hover:underline; }
.vd-tags { @apply flex flex-wrap gap-2 mt-3; }
.vd-tag { @apply text-xs px-2 py-0.5 bg-neutral-100 dark:bg-neutral-800 rounded; }
.vd-desc { @apply mt-4 text-sm text-neutral-600 dark:text-neutral-400 whitespace-pre-wrap; }
.vd-sidebar { @apply w-72 shrink-0; }
.vd-sidebar-title { @apply text-base font-semibold mb-3; }
.vd-recommended { @apply flex flex-col gap-3; }

@media (max-width: 768px) {
  .vd-layout { @apply flex-col; }
  .vd-sidebar { @apply w-full; }
  .vd-recommended { @apply grid grid-cols-2 gap-3; }
}
</style>
```

- [ ] **Step 2: 类型检查**

```bash
cd web && bun run type-check 2>&1 | tail -5
```

期望：无新增错误。

- [ ] **Step 3: Commit**

```bash
git add web/src/views/video/VideoDetailView.vue
git commit -m "feat(video): implement VideoDetailView with embed detection and recommended sidebar"
```

---

## Task 7：前端 — 视频创建/编辑页

**Files:**
- Modify: `web/src/views/video/VideoEditorView.vue`

- [ ] **Step 1: 实现 VideoEditorView**

```vue
<!-- web/src/views/video/VideoEditorView.vue -->
<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useApi } from '@/composables/useApi'
import { useAuthStore } from '@/stores/auth'
import type { Video, Channel } from '@/types'

const route = useRoute()
const router = useRouter()
const api = useApi()
const auth = useAuthStore()

const isEdit = computed(() => !!route.params.id)
const saving = ref(false)
const error = ref('')
const channels = ref<Channel[]>([])

const form = ref({
  channel_id: '' as string,
  title: '',
  description: '',
  storage_type: 'external' as 'local' | 'external',
  video_url: '',
  thumbnail_url: '',
  duration_sec: 0,
  visibility: 'public' as 'public' | 'followers' | 'private',
  status: 'draft' as 'draft' | 'published',
  tags: '' as string, // comma-separated
})

async function loadChannels() {
  const res = await api.get<Channel[]>('/api/channels/mine')
  channels.value = res.data
  if (!form.value.channel_id && channels.value.length > 0) {
    form.value.channel_id = channels.value[0].id
  }
}

async function loadVideo() {
  const id = route.params.id as string
  const res = await api.get<Video>(`/api/videos/${id}`)
  const v = res.data
  form.value = {
    channel_id: v.channel_id || '',
    title: v.title,
    description: v.description,
    storage_type: v.storage_type,
    video_url: v.video_url,
    thumbnail_url: v.thumbnail_url,
    duration_sec: v.duration_sec,
    visibility: v.visibility,
    status: v.status as 'draft' | 'published',
    tags: v.tags.map(t => t.name).join(', '),
  }
}

onMounted(async () => {
  await loadChannels()
  if (isEdit.value) await loadVideo()
})

async function save(publish = false) {
  saving.value = true
  error.value = ''
  const payload = {
    ...form.value,
    channel_id: form.value.channel_id || null,
    status: publish ? 'published' : form.value.status,
    tags: form.value.tags.split(',').map(t => t.trim()).filter(Boolean),
  }
  try {
    if (isEdit.value) {
      await api.put(`/api/videos/${route.params.id}`, payload)
      router.push(`/video/${route.params.id}`)
    } else {
      const res = await api.post<Video>('/api/videos', payload)
      router.push(`/video/${res.data.id}`)
    }
  } catch (e: any) {
    error.value = e?.response?.data?.error || '保存失败'
  } finally {
    saving.value = false
  }
}
</script>

<template>
  <div class="ve-wrap">
    <h1 class="ve-title">{{ isEdit ? '编辑视频' : '发布视频' }}</h1>

    <form class="ve-form" @submit.prevent="save()">
      <!-- 频道 -->
      <label class="ve-label">频道
        <select v-model="form.channel_id" class="ve-select">
          <option value="">不关联频道</option>
          <option v-for="ch in channels" :key="ch.id" :value="ch.id">{{ ch.name }}</option>
        </select>
      </label>

      <!-- 标题 -->
      <label class="ve-label">标题 *
        <input v-model="form.title" required class="ve-input" placeholder="视频标题" />
      </label>

      <!-- 来源类型 -->
      <label class="ve-label">视频来源
        <select v-model="form.storage_type" class="ve-select">
          <option value="external">外部链接（YouTube / Bilibili / 其他）</option>
          <option value="local">本地上传</option>
        </select>
      </label>

      <!-- 视频 URL -->
      <label class="ve-label">视频链接 *
        <input v-model="form.video_url" required class="ve-input"
          :placeholder="form.storage_type === 'external' ? 'https://youtube.com/watch?v=...' : 'S3 对象路径'" />
      </label>

      <!-- 封面 -->
      <label class="ve-label">封面图 URL
        <input v-model="form.thumbnail_url" class="ve-input" placeholder="https://..." />
      </label>

      <!-- 时长 -->
      <label class="ve-label">时长（秒）
        <input v-model.number="form.duration_sec" type="number" min="0" class="ve-input" />
      </label>

      <!-- 简介（URL + 加粗，plain textarea） -->
      <label class="ve-label">简介
        <textarea v-model="form.description" class="ve-textarea" rows="4"
          placeholder="支持 URL 和 **加粗** 语法" />
      </label>

      <!-- 标签 -->
      <label class="ve-label">标签（逗号分隔）
        <input v-model="form.tags" class="ve-input" placeholder="music, tutorial, vlog" />
      </label>

      <!-- 可见范围 -->
      <label class="ve-label">可见范围
        <select v-model="form.visibility" class="ve-select">
          <option value="public">公开</option>
          <option value="followers">仅关注者</option>
          <option value="private">私密</option>
        </select>
      </label>

      <p v-if="error" class="ve-error">{{ error }}</p>

      <div class="ve-actions">
        <button type="submit" :disabled="saving" class="ve-btn ve-btn--secondary">
          {{ saving ? '保存中…' : '保存草稿' }}
        </button>
        <button type="button" :disabled="saving" class="ve-btn ve-btn--primary" @click="save(true)">
          发布
        </button>
      </div>
    </form>
  </div>
</template>

<style scoped>
.ve-wrap { @apply max-w-2xl mx-auto px-4 py-8; }
.ve-title { @apply text-2xl font-bold mb-6; }
.ve-form { @apply flex flex-col gap-4; }
.ve-label { @apply flex flex-col gap-1 text-sm font-medium; }
.ve-input, .ve-select, .ve-textarea {
  @apply mt-1 px-3 py-2 border border-neutral-300 dark:border-neutral-600 rounded text-sm
         bg-white dark:bg-neutral-900 focus:outline-none focus:border-neutral-600;
}
.ve-textarea { @apply resize-y; }
.ve-error { @apply text-red-500 text-sm; }
.ve-actions { @apply flex gap-3 pt-2; }
.ve-btn { @apply px-4 py-2 rounded text-sm font-medium disabled:opacity-50 transition-colors; }
.ve-btn--primary { @apply bg-neutral-900 text-white hover:bg-neutral-700 dark:bg-neutral-100 dark:text-neutral-900; }
.ve-btn--secondary { @apply border border-neutral-300 hover:border-neutral-600; }
</style>
```

- [ ] **Step 2: 类型检查**

```bash
cd web && bun run type-check 2>&1 | tail -5
```

期望：无新增错误。

- [ ] **Step 3: Commit**

```bash
git add web/src/views/video/VideoEditorView.vue
git commit -m "feat(video): implement VideoEditorView with channel selection and tag input"
```

---

## Task 8：验收

- [ ] **Step 1: 后端构建验证**

```bash
cd server && go build ./...
```

期望：无错误。

- [ ] **Step 2: 前端构建验证**

```bash
cd web && bun run type-check && bun run build
```

期望：type-check 0 errors，build 成功。

- [ ] **Step 3: 手动冒烟测试（dev 环境）**

1. 访问 `/video` — 应看到发现页（空列表或已有视频）
2. 访问 `/video/new`（已登录）— 应看到发布表单
3. 填写标题、外部链接，保存草稿 → 点发布 → 应跳转到 `/video/:id`
4. 在 `/video/:id` 播放外链视频（YouTube）应显示 iframe
5. 推荐区应显示来自相同频道或标签的视频
6. 访问 `/api/channels/:slug/rss/video` — 应返回 RSS XML

- [ ] **Step 4: Final commit**

```bash
git add -A
git commit -m "feat(video): complete Video module - models, CRUD, recommendations, RSS, frontend views"
```
