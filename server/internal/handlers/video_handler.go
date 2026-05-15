package handlers

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"

	"atoman/internal/middleware"
	"atoman/internal/model"
)

func SetupVideoRoutes(router *gin.Engine, db *gorm.DB, s3Client *s3.S3) {
	v := router.Group("/api/videos")
	{
		v.GET("", GetVideos(db))
		v.GET("/:id", GetVideo(db))
		v.GET("/:id/recommended", GetRecommendedVideos(db))
		v.POST("/:id/view", IncrementVideoView(db))
		v.POST("", middleware.AuthMiddleware(), CreateVideo(db))
		v.PUT("/:id", middleware.AuthMiddleware(), UpdateVideo(db))
		v.DELETE("/:id", middleware.AuthMiddleware(), DeleteVideo(db))
	}
	// Per-channel Video RSS feed
	router.GET("/api/channels/slug/:slug/rss/video", GetVideoRSS(db))
}

// GetVideos returns published videos. Supports ?channel_id=&tag=&sort=latest|popular&limit=40
func GetVideos(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		channelID := c.Query("channel_id")
		tag := c.Query("tag")
		sort := c.DefaultQuery("sort", "latest")

		q := db.Model(&model.Video{}).
			Where("videos.status = ?", "published").
			Where("videos.visibility = ?", "public").
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
			q = q.Order("videos.view_count DESC")
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

// IncrementVideoView adds 1 to view_count. No auth required.
func IncrementVideoView(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		db.Model(&model.Video{}).Where("id = ?", id).
			UpdateColumn("view_count", gorm.Expr("view_count + 1"))
		c.JSON(http.StatusOK, gin.H{"ok": true})
	}
}

// CreateVideo creates a new video.
func CreateVideo(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.MustGet("userID").(uuid.UUID)
		var input struct {
			ChannelID    *uuid.UUID `json:"channel_id"`
			Title        string     `json:"title" binding:"required"`
			Description  string     `json:"description"`
			StorageType  string     `json:"storage_type"`
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

// UpdateVideo updates a video's fields.
func UpdateVideo(db *gorm.DB) gin.HandlerFunc {
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

// GetRecommendedVideos returns up to 8 recommended videos based on same channel (score 60) and same tags (score 40).
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

		var channelCandidates, tagCandidates []model.Video
		if source.ChannelID != nil {
			db.Model(&model.Video{}).
				Where("channel_id = ? AND id <> ? AND status = ? AND visibility = ?",
					source.ChannelID, id, "published", "public").
				Preload("Tags").Preload("Channel").Limit(20).Find(&channelCandidates)
		}
		if len(tagIDs) > 0 {
			db.Model(&model.Video{}).
				Joins("JOIN video_tag_relations vtr ON vtr.video_id = videos.id").
				Where("vtr.tag_id IN ? AND videos.id <> ? AND videos.status = ? AND videos.visibility = ?",
					tagIDs, id, "published", "public").
				Preload("Tags").Preload("Channel").Limit(20).Find(&tagCandidates)
		}

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

		var results []model.Video
		if len(seen) == 0 {
			// Fallback: latest public videos
			db.Model(&model.Video{}).
				Where("id <> ? AND status = ? AND visibility = ?", id, "published", "public").
				Order("created_at DESC").Preload("Channel").Preload("Tags").Limit(8).Find(&results)
			c.JSON(http.StatusOK, results)
			return
		}

		type scoredID struct {
			id    uuid.UUID
			score int
		}
		var ranked []scoredID
		for vid, score := range scores {
			ranked = append(ranked, scoredID{vid, score})
		}
		for i := 1; i < len(ranked); i++ {
			for j := i; j > 0 && ranked[j].score > ranked[j-1].score; j-- {
				ranked[j], ranked[j-1] = ranked[j-1], ranked[j]
			}
		}
		if len(ranked) > 8 {
			ranked = ranked[:8]
		}
		for _, r := range ranked {
			results = append(results, seen[r.id])
		}

		c.JSON(http.StatusOK, results)
	}
}

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

		scheme := c.Request.Header.Get("X-Forwarded-Proto")
		if scheme == "" {
			scheme = "https"
		}
		siteURL := fmt.Sprintf("%s://%s", scheme, c.Request.Host)

		c.Header("Content-Type", "application/rss+xml; charset=utf-8")
		c.String(http.StatusOK, buildVideoRSS(channel, videos, siteURL))
	}
}

func buildVideoRSS(ch model.Channel, videos []model.Video, siteURL string) string {
	var items strings.Builder
	for _, v := range videos {
		pubDate := v.CreatedAt.Format(time.RFC1123Z)
		enclosure := ""
		if v.StorageType == "local" {
			enclosure = fmt.Sprintf(`<enclosure url="%s" type="video/mp4"/>`, v.VideoURL)
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
