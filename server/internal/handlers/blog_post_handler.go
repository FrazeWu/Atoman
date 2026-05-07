package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"

	"atoman/internal/middleware"
	"atoman/internal/model"
)

// SetupBlogPostRoutes configures blog post routes
func SetupBlogPostRoutes(router *gin.Engine, db *gorm.DB) {
	blog := router.Group("/api/blog")
	{
		blog.GET("/posts", GetPosts(db))
		blog.GET("/posts/:id", middleware.OptionalAuthMiddleware(), GetPost(db))

		// Protected routes
		protected := blog.Group("")
		protected.Use(middleware.AuthMiddleware())
		{
			protected.POST("/posts", CreatePost(db))
			protected.PUT("/posts/:id", UpdatePost(db))
			protected.DELETE("/posts/:id", DeletePost(db))

			protected.POST("/posts/:id/publish", PublishPost(db))
			protected.POST("/posts/:id/unpublish", UnpublishPost(db))
			protected.POST("/posts/:id/pin", PinPost(db))
			protected.POST("/posts/:id/unpin", UnpinPost(db))

			protected.GET("/posts/drafts", GetDrafts(db))

			protected.POST("/posts/:id/collections", AddPostToCollection(db))
			protected.DELETE("/posts/:id/collections/:collection_id", RemovePostFromCollection(db))
		}
	}
}

// PostInput represents the request body for creating/updating a post
type PostInput struct {
	Title         string   `json:"title" binding:"required"`
	Content       string   `json:"content" binding:"required"`
	Summary       string   `json:"summary"`
	CoverURL      string   `json:"cover_url"`
	AllowComments *bool    `json:"allow_comments"`
	Status        string   `json:"status"` // "draft" or "published"
	ChannelID     string   `json:"channel_id"`
	CollectionIDs []string `json:"collection_ids"`
}

// CollectionActionInput represents the request body for adding a post to a collection
type CollectionActionInput struct {
	CollectionID uuid.UUID `json:"collection_id" binding:"required"`
}

// GetPosts returns a list of published posts, optionally filtered
func GetPosts(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var posts []model.Post
		query := db.Preload("User").Preload("Collections").Where("status = ?", "published")

		if userID := c.Query("user_id"); userID != "" {
			query = query.Where("user_id = ?", userID)
		}

		if collectionID := c.Query("collection_id"); collectionID != "" {
			query = query.Joins("JOIN post_collections pc ON pc.post_id = posts.id").
				Where("pc.collection_id = ?", collectionID)
		}

		if err := query.Order("pinned DESC, created_at DESC").Find(&posts).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch posts"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"data": posts, "message": "ok"})
	}
}

// GetPost returns a specific post by ID
func GetPost(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		var post model.Post

		if err := db.Preload("User").Preload("Collections").First(&post, "id = ?", id).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Post not found"})
			return
		}

		// If post is draft, only owner can view it
		if post.Status == "draft" {
			// Try to get user from context (might not exist if not logged in)
			userIDVal, exists := c.Get("user_id")
			if !exists || userIDVal.(uuid.UUID) != post.UserID {
				c.JSON(http.StatusForbidden, gin.H{"error": "You don't have permission to view this draft"})
				return
			}
		}

		c.JSON(http.StatusOK, gin.H{"data": post, "message": "ok"})
	}
}

// CreatePost creates a new post for the authenticated user
func CreatePost(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var input PostInput
		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		userIDVal, _ := c.Get("user_id")
		userID := userIDVal.(uuid.UUID)

		var channelID *uuid.UUID
		var selectedCollections []model.Collection
		if input.ChannelID != "" {
			parsedChannelID, err := uuid.Parse(input.ChannelID)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid channel UUID"})
				return
			}

			var channel model.Channel
			if err := db.First(&channel, "id = ?", parsedChannelID).Error; err != nil {
				c.JSON(http.StatusNotFound, gin.H{"error": "Channel not found"})
				return
			}

			if channel.UserID != userID {
				c.JSON(http.StatusForbidden, gin.H{"error": "You don't have permission to create post in this channel"})
				return
			}

			channelID = &parsedChannelID

			for _, collectionIDStr := range input.CollectionIDs {
				collectionID, err := uuid.Parse(collectionIDStr)
				if err != nil {
					c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid collection UUID"})
					return
				}

				var collection model.Collection
				if err := db.Preload("Channel").First(&collection, "id = ?", collectionID).Error; err != nil {
					c.JSON(http.StatusNotFound, gin.H{"error": "Collection not found"})
					return
				}

				if collection.Channel.UserID != userID {
					c.JSON(http.StatusForbidden, gin.H{"error": "You don't have permission to assign this collection"})
					return
				}

				if collection.ChannelID != parsedChannelID {
					c.JSON(http.StatusBadRequest, gin.H{"error": "Collection does not belong to selected channel"})
					return
				}

				selectedCollections = append(selectedCollections, collection)
			}
		}

		allowComments := true
		if input.AllowComments != nil {
			allowComments = *input.AllowComments
		}

		status := "draft"
		if input.Status == "published" {
			status = "published"
		}

		post := model.Post{
			UserID:        userID,
			Title:         input.Title,
			Content:       input.Content,
			Summary:       input.Summary,
			CoverURL:      input.CoverURL,
			Status:        status,
			AllowComments: allowComments,
		}

		if err := db.Create(&post).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create post"})
			return
		}

		if channelID != nil {
			defaultCollection, err := ensureDefaultCollection(db, *channelID)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to ensure default collection"})
				return
			}

			if err := db.Model(&post).Association("Collections").Append(defaultCollection); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to attach default collection"})
				return
			}

			for _, collection := range selectedCollections {
				if collection.ID == defaultCollection.ID {
					continue
				}

				if err := db.Model(&post).Association("Collections").Append(&collection); err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to assign collection"})
					return
				}
			}
		}

		if err := db.Preload("Collections").First(&post, "id = ?", post.ID).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch created post"})
			return
		}

		c.JSON(http.StatusCreated, gin.H{"data": post, "message": "ok"})
	}
}

// UpdatePost updates an existing post (only by owner)
func UpdatePost(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		var input PostInput
		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		var post model.Post
		if err := db.First(&post, "id = ?", id).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Post not found"})
			return
		}

		userIDVal, _ := c.Get("user_id")
		userID := userIDVal.(uuid.UUID)

		if post.UserID != userID {
			c.JSON(http.StatusForbidden, gin.H{"error": "You don't have permission to update this post"})
			return
		}

		updates := map[string]interface{}{
			"title":     input.Title,
			"content":   input.Content,
			"summary":   input.Summary,
			"cover_url": input.CoverURL,
		}

		if input.Status == "published" || input.Status == "draft" {
			updates["status"] = input.Status
		}

		if input.AllowComments != nil {
			updates["allow_comments"] = *input.AllowComments
		}

		if err := db.Model(&post).Updates(updates).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update post"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"data": post, "message": "ok"})
	}
}

// DeletePost deletes a post (only by owner)
func DeletePost(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		var post model.Post

		if err := db.First(&post, "id = ?", id).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Post not found"})
			return
		}

		userIDVal, _ := c.Get("user_id")
		userID := userIDVal.(uuid.UUID)

		if post.UserID != userID {
			c.JSON(http.StatusForbidden, gin.H{"error": "You don't have permission to delete this post"})
			return
		}

		if err := db.Delete(&post).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete post"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "ok"})
	}
}

// PublishPost changes post status to published
func PublishPost(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		updatePostStatus(c, db, "published")
	}
}

// UnpublishPost changes post status to draft
func UnpublishPost(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		updatePostStatus(c, db, "draft")
	}
}

// PinPost sets post as pinned
func PinPost(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		updatePostPin(c, db, true)
	}
}

// UnpinPost removes pinned status
func UnpinPost(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		updatePostPin(c, db, false)
	}
}

// GetDrafts returns a list of drafts for the authenticated user
func GetDrafts(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userIDVal, _ := c.Get("user_id")
		userID := userIDVal.(uuid.UUID)

		var posts []model.Post
		if err := db.Preload("Collections").Where("user_id = ? AND status = ?", userID, "draft").Order("updated_at DESC").Find(&posts).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch drafts"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"data": posts, "message": "ok"})
	}
}

// AddPostToCollection adds a post to a collection
func AddPostToCollection(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		postID := c.Param("id")
		var input CollectionActionInput
		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		var post model.Post
		if err := db.First(&post, "id = ?", postID).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Post not found"})
			return
		}

		userIDVal, _ := c.Get("user_id")
		userID := userIDVal.(uuid.UUID)

		if post.UserID != userID {
			c.JSON(http.StatusForbidden, gin.H{"error": "You don't have permission to modify this post"})
			return
		}

		// Verify collection exists and belongs to user's channel
		var collection model.Collection
		if err := db.Preload("Channel").First(&collection, "id = ?", input.CollectionID).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Collection not found"})
			return
		}

		if collection.Channel.UserID != userID {
			c.JSON(http.StatusForbidden, gin.H{"error": "You don't have permission to add to this collection"})
			return
		}

		// Add to collection
		if err := db.Model(&post).Association("Collections").Append(&collection); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add post to collection"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "ok"})
	}
}

// RemovePostFromCollection removes a post from a collection
func RemovePostFromCollection(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		postID := c.Param("id")
		collectionID := c.Param("collection_id")

		var post model.Post
		if err := db.First(&post, "id = ?", postID).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Post not found"})
			return
		}

		userIDVal, _ := c.Get("user_id")
		userID := userIDVal.(uuid.UUID)

		if post.UserID != userID {
			c.JSON(http.StatusForbidden, gin.H{"error": "You don't have permission to modify this post"})
			return
		}

		var collection model.Collection
		if err := db.First(&collection, "id = ?", collectionID).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Collection not found"})
			return
		}

		// Remove from collection
		if err := db.Model(&post).Association("Collections").Delete(&collection); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to remove post from collection"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "ok"})
	}
}

// Helper functions

func updatePostStatus(c *gin.Context, db *gorm.DB, status string) {
	id := c.Param("id")
	var post model.Post

	if err := db.First(&post, "id = ?", id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Post not found"})
		return
	}

	userIDVal, _ := c.Get("user_id")
	userID := userIDVal.(uuid.UUID)

	if post.UserID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "You don't have permission to modify this post"})
		return
	}

	if err := db.Model(&post).Update("status", status).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update post status"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "ok"})
}

func updatePostPin(c *gin.Context, db *gorm.DB, pinned bool) {
	id := c.Param("id")
	var post model.Post

	if err := db.First(&post, "id = ?", id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Post not found"})
		return
	}

	userIDVal, _ := c.Get("user_id")
	userID := userIDVal.(uuid.UUID)

	if post.UserID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "You don't have permission to modify this post"})
		return
	}

	if err := db.Model(&post).Update("pinned", pinned).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update post pin status"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "ok"})
}
