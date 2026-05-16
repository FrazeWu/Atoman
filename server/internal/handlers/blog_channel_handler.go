package handlers

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"atoman/internal/middleware"
	"atoman/internal/model"
)

// SetupBlogChannelRoutes configures blog channel and collection routes
func SetupBlogChannelRoutes(router *gin.Engine, db *gorm.DB) {
	blog := router.Group("/api/blog")
	{
		// Public routes
		blog.GET("/channels", middleware.OptionalAuthMiddleware(), GetChannels(db))
		blog.GET("/channels/:id", middleware.OptionalAuthMiddleware(), GetChannel(db))
		blog.GET("/channels/:id/collections", middleware.OptionalAuthMiddleware(), GetChannelCollections(db))
		blog.GET("/channels/slug/:slug", middleware.OptionalAuthMiddleware(), GetChannelBySlug(db))
		blog.GET("/channels/slug/:slug/collections", middleware.OptionalAuthMiddleware(), GetChannelCollectionsBySlug(db))
		blog.GET("/collections", middleware.OptionalAuthMiddleware(), GetUserCollections(db))
		blog.GET("/collections/:id", middleware.OptionalAuthMiddleware(), GetCollection(db))
		blog.GET("/channels/slug/:slug/rss/article", GetChannelArticleRSS(db))

		// Protected routes
		protected := blog.Group("")
		protected.Use(middleware.AuthMiddleware())
		{
			protected.POST("/channels/ensure-default", EnsureDefaultChannel(db))
			protected.POST("/channels", CreateChannel(db))
			protected.PUT("/channels/:id", UpdateChannel(db))
			protected.DELETE("/channels/:id", DeleteChannel(db))

			protected.POST("/channels/:id/collections", CreateCollection(db))
			protected.PUT("/collections/:id", UpdateCollection(db))
			protected.DELETE("/collections/:id", DeleteCollection(db))
		}
	}
}

// ChannelInput represents the request body for creating/updating a channel
type ChannelInput struct {
	Name        string `json:"name" binding:"required"`
	Slug        string `json:"slug"`
	Description string `json:"description"`
	CoverURL    string `json:"cover_url"`
}

// CollectionInput represents the request body for creating/updating a collection
type CollectionInput struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
	CoverURL    string `json:"cover_url"`
}

type DeleteChannelInput struct {
	Password        string `json:"password" binding:"required"`
	MoveContent     bool   `json:"move_content"`
	TargetChannelID string `json:"target_channel_id"`
}

// EnsureDefaultChannel creates a default channel for the authenticated user if they don't have one
func EnsureDefaultChannel(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userIDVal, _ := c.Get("user_id")
		userID := userIDVal.(uuid.UUID)

		var user model.User
		if err := db.First(&user, "uuid = ?", userID).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
			return
		}

		channel, err := EnsureDefaultChannelForUser(db, userID, user.Username)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to ensure default channel: " + err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"data": channel, "message": "ok"})
	}
}

// GetChannels returns a list of channels, optionally filtered by user_id
func GetChannels(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var channels []model.Channel
		query := db.Preload("User")

		if userID := c.Query("user_id"); userID != "" {
			if _, err := uuid.Parse(userID); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user_id"})
				return
			}
			query = query.Where("user_id = ?", userID)
		}

		if err := query.Find(&channels).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch channels"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"data": channels, "message": "ok"})
	}
}

// GetChannel returns a specific channel by ID
func GetChannel(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		var channel model.Channel

		if err := db.Preload("User").First(&channel, "id = ?", id).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Channel not found"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"data": channel, "message": "ok"})
	}
}

// GetChannelBySlug returns a specific channel by slug
func GetChannelBySlug(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		slug := strings.TrimSpace(c.Param("slug"))
		var channel model.Channel

		if err := db.Preload("User").First(&channel, "slug = ?", slug).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Channel not found"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"data": channel, "message": "ok"})
	}
}

// CreateChannel creates a new channel for the authenticated user
func CreateChannel(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var input ChannelInput
		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		input.Name = normalizeName(input.Name)
		input.Description = strings.TrimSpace(input.Description)
		if input.Name == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Channel name is required"})
			return
		}

		exists, err := channelNameExists(db, input.Name, nil)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to validate channel name"})
			return
		}
		if exists {
			c.JSON(http.StatusConflict, gin.H{"error": "Channel name already exists"})
			return
		}

		userIDVal, _ := c.Get("user_id")
		userID := userIDVal.(uuid.UUID)

		slugSource := input.Slug
		if strings.TrimSpace(slugSource) == "" {
			slugSource = input.Name
		}
		slug, err := uniqueChannelSlug(db, slugSource, nil)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate channel slug"})
			return
		}

		channel := model.Channel{
			UserID:      userID,
			Name:        input.Name,
			Slug:        slug,
			Description: input.Description,
			CoverURL:    input.CoverURL,
		}

		if err := db.Create(&channel).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create channel"})
			return
		}

		if _, err := ensureDefaultCollection(db, channel.ID); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create default collection"})
			return
		}

		c.JSON(http.StatusCreated, gin.H{"data": channel, "message": "ok"})
	}
}

// UpdateChannel updates an existing channel (only by owner)
func UpdateChannel(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		var input ChannelInput
		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		input.Name = normalizeName(input.Name)
		input.Description = strings.TrimSpace(input.Description)
		if input.Name == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Channel name is required"})
			return
		}

		var channel model.Channel
		if err := db.First(&channel, "id = ?", id).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Channel not found"})
			return
		}

		userIDVal, _ := c.Get("user_id")
		userID := userIDVal.(uuid.UUID)

		if channel.UserID != userID {
			c.JSON(http.StatusForbidden, gin.H{"error": "You don't have permission to update this channel"})
			return
		}

		excludeID := channel.ID
		exists, err := channelNameExists(db, input.Name, &excludeID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to validate channel name"})
			return
		}
		if exists {
			c.JSON(http.StatusConflict, gin.H{"error": "Channel name already exists"})
			return
		}

		slugSource := input.Slug
		if strings.TrimSpace(slugSource) == "" {
			slugSource = input.Name
		}
		slug, err := uniqueChannelSlug(db, slugSource, &excludeID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate channel slug"})
			return
		}

		if err := db.Model(&channel).Updates(model.Channel{
			Name:        input.Name,
			Slug:        slug,
			Description: input.Description,
			CoverURL:    input.CoverURL,
		}).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update channel"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"data": channel, "message": "ok"})
	}
}

// DeleteChannel deletes a channel (only by owner)
func DeleteChannel(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		var channel model.Channel

		if err := db.First(&channel, "id = ?", id).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Channel not found"})
			return
		}

		if channel.IsDefault {
			c.JSON(http.StatusForbidden, gin.H{"error": "Cannot delete default channel"})
			return
		}

		var input DeleteChannelInput
		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		userIDVal, _ := c.Get("user_id")
		userID := userIDVal.(uuid.UUID)

		if channel.UserID != userID {
			c.JSON(http.StatusForbidden, gin.H{"error": "You don't have permission to delete this channel"})
			return
		}

		var user model.User
		if err := db.First(&user, "uuid = ?", userID).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to verify user"})
			return
		}

		if input.Password == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Password is required"})
			return
		}

		if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(input.Password)); err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Password is incorrect"})
			return
		}

		var targetChannel *model.Channel
		if input.MoveContent {
			if strings.TrimSpace(input.TargetChannelID) == "" {
				c.JSON(http.StatusBadRequest, gin.H{"error": "target_channel_id is required when move_content is true"})
				return
			}

			targetID, err := uuid.Parse(strings.TrimSpace(input.TargetChannelID))
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid target channel UUID"})
				return
			}

			if targetID == channel.ID {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Target channel must be different from source channel"})
				return
			}

			var target model.Channel
			if err := db.First(&target, "id = ?", targetID).Error; err != nil {
				c.JSON(http.StatusNotFound, gin.H{"error": "Target channel not found"})
				return
			}

			if target.UserID != userID {
				c.JSON(http.StatusForbidden, gin.H{"error": "You don't have permission to move content to this channel"})
				return
			}

			targetChannel = &target
		}

		err := db.Transaction(func(tx *gorm.DB) error {
			var sourceCollections []model.Collection
			if err := tx.Where("channel_id = ?", channel.ID).Find(&sourceCollections).Error; err != nil {
				return err
			}

			sourceCollectionIDs := make([]uuid.UUID, 0, len(sourceCollections))
			for _, collection := range sourceCollections {
				sourceCollectionIDs = append(sourceCollectionIDs, collection.ID)
			}

			if input.MoveContent && targetChannel != nil && len(sourceCollectionIDs) > 0 {
				defaultCollection, err := ensureDefaultCollection(tx, targetChannel.ID)
				if err != nil {
					return err
				}

				var postCollections []model.PostCollection
				if err := tx.Where("collection_id IN ?", sourceCollectionIDs).Find(&postCollections).Error; err != nil {
					return err
				}

				seenPosts := make(map[uuid.UUID]bool)
				for _, relation := range postCollections {
					if seenPosts[relation.PostID] {
						continue
					}
					seenPosts[relation.PostID] = true

					if err := tx.Model(&model.Post{}).Where("id = ?", relation.PostID).Update("channel_id", targetChannel.ID).Error; err != nil {
						return err
					}

					postCollection := model.PostCollection{
						PostID:       relation.PostID,
						CollectionID: defaultCollection.ID,
					}

					if err := tx.Where("post_id = ? AND collection_id = ?", relation.PostID, defaultCollection.ID).
						FirstOrCreate(&postCollection).Error; err != nil {
						return err
					}
				}
			} else if input.MoveContent && targetChannel != nil {
				if err := tx.Model(&model.Post{}).Where("channel_id = ?", channel.ID).Update("channel_id", targetChannel.ID).Error; err != nil {
					return err
				}
			} else {
				if err := tx.Model(&model.Post{}).Where("channel_id = ?", channel.ID).Update("channel_id", nil).Error; err != nil {
					return err
				}
			}

			if len(sourceCollectionIDs) > 0 {
				if err := tx.Where("collection_id IN ?", sourceCollectionIDs).Delete(&model.PostCollection{}).Error; err != nil {
					return err
				}
			}

			if err := tx.Where("channel_id = ?", channel.ID).Delete(&model.Collection{}).Error; err != nil {
				return err
			}

			if err := tx.Delete(&channel).Error; err != nil {
				return err
			}

			return nil
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete channel"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "ok"})
	}
}

// GetCollection returns a single collection by ID
func GetCollection(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		var collection model.Collection

		if err := db.Preload("Channel").First(&collection, "id = ?", id).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Collection not found"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"data": collection, "message": "ok"})
	}
}

// GetChannelCollections returns all collections for a specific channel
func GetChannelCollections(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		channelID := c.Param("id")
		var collections []model.Collection

		if err := db.Where("channel_id = ?", channelID).Find(&collections).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch collections"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"data": collections, "message": "ok"})
	}
}

// GetChannelCollectionsBySlug returns all collections for a specific channel slug
func GetChannelCollectionsBySlug(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		slug := strings.TrimSpace(c.Param("slug"))
		var channel model.Channel
		if err := db.First(&channel, "slug = ?", slug).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Channel not found"})
			return
		}

		var collections []model.Collection
		if err := db.Where("channel_id = ?", channel.ID).Find(&collections).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch collections"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"data": collections, "message": "ok"})
	}
}

// GetUserCollections returns all collections for the authenticated user with channel names
func GetUserCollections(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userIDVal, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}
		userID := userIDVal.(uuid.UUID)

		// Get all channels for this user
		var userChannels []model.Channel
		if err := db.Where("user_id = ?", userID).Find(&userChannels).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch channels"})
			return
		}

		channelIDs := make([]uuid.UUID, len(userChannels))
		channelMap := make(map[uuid.UUID]string)
		for i, ch := range userChannels {
			channelIDs[i] = ch.ID
			channelMap[ch.ID] = ch.Name
		}

		// Get all collections for these channels
		var collections []model.Collection
		if len(channelIDs) > 0 {
			if err := db.Where("channel_id IN ?", channelIDs).Order("created_at DESC").Find(&collections).Error; err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch collections"})
				return
			}
		}

		// Add channel_name to each collection
		type CollectionWithChannel struct {
			model.Collection
			ChannelName string `json:"channel_name"`
		}

		result := make([]CollectionWithChannel, len(collections))
		for i, col := range collections {
			result[i] = CollectionWithChannel{
				Collection:  col,
				ChannelName: channelMap[col.ChannelID],
			}
		}

		c.JSON(http.StatusOK, gin.H{"data": result, "message": "ok"})
	}
}

// CreateCollection creates a new collection in a channel (only by channel owner)
func CreateCollection(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		channelIDStr := c.Param("id")
		channelID, err := uuid.Parse(channelIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid channel UUID"})
			return
		}

		var input CollectionInput
		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		input.Name = normalizeName(input.Name)
		input.Description = strings.TrimSpace(input.Description)
		if input.Name == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Collection name is required"})
			return
		}

		// Verify channel exists and belongs to user
		var channel model.Channel
		if err := db.First(&channel, "id = ?", channelID).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Channel not found"})
			return
		}

		userIDVal, _ := c.Get("user_id")
		userID := userIDVal.(uuid.UUID)

		if channel.UserID != userID {
			c.JSON(http.StatusForbidden, gin.H{"error": "You don't have permission to add collections to this channel"})
			return
		}

		exists, err := collectionNameExists(db, channelID, input.Name, nil)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to validate collection name"})
			return
		}
		if exists {
			c.JSON(http.StatusConflict, gin.H{"error": "Collection name already exists in this channel"})
			return
		}

		collection := model.Collection{
			ChannelID:   channelID,
			Name:        input.Name,
			Description: input.Description,
			CoverURL:    input.CoverURL,
		}

		if err := db.Create(&collection).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create collection"})
			return
		}

		c.JSON(http.StatusCreated, gin.H{"data": collection, "message": "ok"})
	}
}

// UpdateCollection updates an existing collection (only by channel owner)
func UpdateCollection(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		var input CollectionInput
		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		input.Name = normalizeName(input.Name)
		input.Description = strings.TrimSpace(input.Description)
		if input.Name == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Collection name is required"})
			return
		}

		var collection model.Collection
		if err := db.Preload("Channel").First(&collection, "id = ?", id).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Collection not found"})
			return
		}

		userIDVal, _ := c.Get("user_id")
		userID := userIDVal.(uuid.UUID)

		if collection.Channel.UserID != userID {
			c.JSON(http.StatusForbidden, gin.H{"error": "You don't have permission to update this collection"})
			return
		}

		// Prevent renaming default collection
		if collection.IsDefault && input.Name != collection.Name {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot rename default collection"})
			return
		}

		excludeID := collection.ID
		exists, err := collectionNameExists(db, collection.ChannelID, input.Name, &excludeID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to validate collection name"})
			return
		}
		if exists {
			c.JSON(http.StatusConflict, gin.H{"error": "Collection name already exists in this channel"})
			return
		}

		if err := db.Model(&collection).Updates(model.Collection{
			Name:        input.Name,
			Description: input.Description,
			CoverURL:    input.CoverURL,
		}).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update collection"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"data": collection, "message": "ok"})
	}
}

// DeleteCollection deletes a collection (only by channel owner)
func DeleteCollection(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		var collection model.Collection

		if err := db.Preload("Channel").First(&collection, "id = ?", id).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Collection not found"})
			return
		}

		userIDVal, _ := c.Get("user_id")
		userID := userIDVal.(uuid.UUID)

		if collection.Channel.UserID != userID {
			c.JSON(http.StatusForbidden, gin.H{"error": "You don't have permission to delete this collection"})
			return
		}

		if err := db.Delete(&collection).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete collection"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "ok"})
	}
}

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
