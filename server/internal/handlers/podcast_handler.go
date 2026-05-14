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

func SetupPodcastRoutes(router *gin.Engine, db *gorm.DB, _ *s3.S3) {
	p := router.Group("/api/podcast")
	{
		p.GET("/episodes", GetPodcastEpisodes(db))
		p.GET("/shows/:channelSlug/episodes", GetShowEpisodes(db))
		p.GET("/episodes/:id", GetPodcastEpisode(db))
		p.POST("/episodes", middleware.AuthMiddleware(), CreatePodcastEpisode(db))
		p.PUT("/episodes/:id", middleware.AuthMiddleware(), UpdatePodcastEpisode(db))
		p.DELETE("/episodes/:id", middleware.AuthMiddleware(), DeletePodcastEpisode(db))
	}
	router.GET("/api/channels/:slug/rss/podcast", GetPodcastRSS(db))
}

// GetPodcastEpisodes lists all published episodes across all shows.
func GetPodcastEpisodes(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var episodes []model.PodcastEpisode
		db.Preload("Post").Preload("Channel").
			Joins("JOIN posts ON posts.id = podcast_episodes.post_id AND posts.status = 'published' AND posts.deleted_at IS NULL").
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
			Joins("JOIN posts ON posts.id = podcast_episodes.post_id AND posts.status = 'published' AND posts.deleted_at IS NULL").
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
func CreatePodcastEpisode(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		idVal, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}
		userID := idVal.(uuid.UUID)

		var input struct {
			ChannelID       string `json:"channel_id" binding:"required"`
			Title           string `json:"title" binding:"required"`
			Shownotes       string `json:"shownotes"`
			AudioURL        string `json:"audio_url" binding:"required"`
			DurationSec     int    `json:"duration_sec"`
			EpisodeCoverURL string `json:"episode_cover_url"`
			SeasonNumber    int    `json:"season_number"`
			EpisodeNumber   int    `json:"episode_number"`
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
		seasonNum := input.SeasonNumber
		if seasonNum < 1 {
			seasonNum = 1
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
				SeasonNumber:    seasonNum,
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
		idVal, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}
		userID := idVal.(uuid.UUID)
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
		idVal, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}
		userID := idVal.(uuid.UUID)
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
			Joins("JOIN posts ON posts.id = podcast_episodes.post_id AND posts.status = 'published' AND posts.deleted_at IS NULL").
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
