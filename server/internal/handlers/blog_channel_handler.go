package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"

	"atoman/internal/middleware"
	"atoman/internal/model"
)

// SetupBlogChannelRoutes configures blog channel and collection routes
func SetupBlogChannelRoutes(router *gin.Engine, db *gorm.DB) {
	blog := router.Group("/api/blog")
	{
		// Public routes
		blog.GET("/channels", GetChannels(db))
		blog.GET("/channels/:id", GetChannel(db))
		blog.GET("/channels/:id/collections", GetChannelCollections(db))

		// Protected routes
		protected := blog.Group("")
		protected.Use(middleware.AuthMiddleware())
		{
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
	Description string `json:"description"`
	CoverURL    string `json:"cover_url"`
}

// CollectionInput represents the request body for creating/updating a collection
type CollectionInput struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
	CoverURL    string `json:"cover_url"`
}

// GetChannels returns a list of channels, optionally filtered by user_id
func GetChannels(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var channels []model.Channel
		query := db.Preload("User")

		if userID := c.Query("user_id"); userID != "" {
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

// CreateChannel creates a new channel for the authenticated user
func CreateChannel(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var input ChannelInput
		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		userIDVal, _ := c.Get("user_id")
		userID := userIDVal.(uuid.UUID)

		channel := model.Channel{
			UserID:      userID,
			Name:        input.Name,
			Description: input.Description,
			CoverURL:    input.CoverURL,
		}

		if err := db.Create(&channel).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create channel"})
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

		if err := db.Model(&channel).Updates(model.Channel{
			Name:        input.Name,
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

		userIDVal, _ := c.Get("user_id")
		userID := userIDVal.(uuid.UUID)

		if channel.UserID != userID {
			c.JSON(http.StatusForbidden, gin.H{"error": "You don't have permission to delete this channel"})
			return
		}

		if err := db.Delete(&channel).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete channel"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "ok"})
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
