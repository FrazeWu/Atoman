package handlers

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

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

			protected.GET("/drafts", GetBlogDraft(db))
			protected.PUT("/drafts", PutBlogDraft(db))
			protected.DELETE("/drafts", DeleteBlogDraft(db))

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

type BlogDraftInput struct {
	ContextKey    string   `json:"context_key" binding:"required"`
	SourcePostID  string   `json:"source_post_id"`
	Title         string   `json:"title"`
	Content       string   `json:"content"`
	Summary       string   `json:"summary"`
	CoverURL      string   `json:"cover_url"`
	AllowComments *bool    `json:"allow_comments"`
	ChannelID     string   `json:"channel_id"`
	CollectionIDs []string `json:"collection_ids"`
}

type BlogDraftResponse struct {
	ID            uuid.UUID `json:"id"`
	UserID        uuid.UUID `json:"user_id"`
	ContextKey    string    `json:"context_key"`
	SourcePostID  *string   `json:"source_post_id,omitempty"`
	Title         string    `json:"title"`
	Content       string    `json:"content"`
	Summary       string    `json:"summary"`
	CoverURL      string    `json:"cover_url"`
	AllowComments bool      `json:"allow_comments"`
	ChannelID     *string   `json:"channel_id,omitempty"`
	CollectionIDs []string  `json:"collection_ids"`
	CreatedAt     any       `json:"created_at"`
	UpdatedAt     any       `json:"updated_at"`
}

// GetPosts returns a list of published posts, optionally filtered
func GetPosts(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var posts []model.Post
		query := db.Preload("User").Preload("Channel").Preload("Collections").Where("status = ?", "published")

		if userID := c.Query("user_id"); userID != "" {
			query = query.Where("user_id = ?", userID)
		}

		if channelID := c.Query("channel_id"); channelID != "" {
			query = query.Where("channel_id = ?", channelID)
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

		if err := db.Preload("User").Preload("Channel").Preload("Collections").First(&post, "id = ?", id).Error; err != nil {
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
			ChannelID:     channelID,
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

		if err := db.Preload("Channel").Preload("Collections").First(&post, "id = ?", post.ID).Error; err != nil {
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

		var targetChannelID *uuid.UUID
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
				c.JSON(http.StatusForbidden, gin.H{"error": "You don't have permission to move post to this channel"})
				return
			}

			targetChannelID = &parsedChannelID
			updates["channel_id"] = parsedChannelID
		} else {
			updates["channel_id"] = nil
		}

		selectedCollections := make([]model.Collection, 0, len(input.CollectionIDs))
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

			if targetChannelID == nil || collection.ChannelID != *targetChannelID {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Collection does not belong to selected channel"})
				return
			}

			selectedCollections = append(selectedCollections, collection)
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

		if targetChannelID != nil {
			defaultCollection, err := ensureDefaultCollection(db, *targetChannelID)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to ensure default collection"})
				return
			}

			collectionsToAssign := make([]model.Collection, 0, len(selectedCollections)+1)
			collectionsToAssign = append(collectionsToAssign, *defaultCollection)
			for _, collection := range selectedCollections {
				if collection.ID == defaultCollection.ID {
					continue
				}
				collectionsToAssign = append(collectionsToAssign, collection)
			}
			if err := db.Model(&post).Association("Collections").Replace(collectionsToAssign); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update post collections"})
				return
			}
		} else if err := db.Model(&post).Association("Collections").Clear(); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to clear post collections"})
			return
		}

		if err := db.Preload("Channel").Preload("Collections").First(&post, "id = ?", post.ID).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch updated post"})
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

func GetBlogDraft(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		contextKey := strings.TrimSpace(c.Query("context_key"))
		if contextKey == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "context_key required"})
			return
		}

		userIDVal, _ := c.Get("user_id")
		userID := userIDVal.(uuid.UUID)

		var draft model.BlogDraft
		if err := db.Where("user_id = ? AND context_key = ?", userID, contextKey).First(&draft).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Draft not found"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"data": buildBlogDraftResponse(draft), "message": "ok"})
	}
}

func PutBlogDraft(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var input BlogDraftInput
		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		userIDVal, _ := c.Get("user_id")
		userID := userIDVal.(uuid.UUID)

		sourcePostID, err := parseOptionalUUID(input.SourcePostID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid source_post_id"})
			return
		}

		channelID, err := parseOptionalUUID(input.ChannelID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid channel_id"})
			return
		}

		collectionIDs, err := normalizeUUIDList(input.CollectionIDs)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid collection_ids"})
			return
		}

		allowComments := true
		if input.AllowComments != nil {
			allowComments = *input.AllowComments
		}

		draft := model.BlogDraft{
			UserID:        userID,
			ContextKey:    strings.TrimSpace(input.ContextKey),
			SourcePostID:  sourcePostID,
			Title:         input.Title,
			Content:       input.Content,
			Summary:       input.Summary,
			CoverURL:      input.CoverURL,
			AllowComments: allowComments,
			ChannelID:     channelID,
			CollectionIDs: strings.Join(collectionIDs, ","),
		}
		upsertCols := []string{"source_post_id", "title", "content", "summary", "cover_url", "allow_comments", "channel_id", "collection_ids", "updated_at"}
		if err := db.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "user_id"}, {Name: "context_key"}},
			DoUpdates: clause.AssignmentColumns(upsertCols),
		}).Create(&draft).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save draft"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"data": buildBlogDraftResponse(draft), "message": "ok"})
	}
}

func DeleteBlogDraft(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		contextKey := strings.TrimSpace(c.Query("context_key"))
		if contextKey == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "context_key required"})
			return
		}

		userIDVal, _ := c.Get("user_id")
		userID := userIDVal.(uuid.UUID)

		if err := db.Where("user_id = ? AND context_key = ?", userID, contextKey).Delete(&model.BlogDraft{}).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete draft"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "ok"})
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

		if post.ChannelID == nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Post is not assigned to a channel"})
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

		if collection.ChannelID != *post.ChannelID {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Collection does not belong to post channel"})
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

		if post.ChannelID == nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Post is not assigned to a channel"})
			return
		}

		var collection model.Collection
		if err := db.First(&collection, "id = ?", collectionID).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Collection not found"})
			return
		}

		if collection.ChannelID != *post.ChannelID {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Collection does not belong to post channel"})
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

func parseOptionalUUID(raw string) (*uuid.UUID, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return nil, nil
	}
	parsed, err := uuid.Parse(trimmed)
	if err != nil {
		return nil, err
	}
	return &parsed, nil
}

func normalizeUUIDList(values []string) ([]string, error) {
	if len(values) == 0 {
		return []string{}, nil
	}

	normalized := make([]string, 0, len(values))
	seen := make(map[string]struct{}, len(values))
	for _, value := range values {
		parsed, err := uuid.Parse(strings.TrimSpace(value))
		if err != nil {
			return nil, err
		}
		stringID := parsed.String()
		if _, exists := seen[stringID]; exists {
			continue
		}
		seen[stringID] = struct{}{}
		normalized = append(normalized, stringID)
	}
	return normalized, nil
}

func buildBlogDraftResponse(draft model.BlogDraft) BlogDraftResponse {
	var sourcePostID *string
	if draft.SourcePostID != nil {
		value := draft.SourcePostID.String()
		sourcePostID = &value
	}

	var channelID *string
	if draft.ChannelID != nil {
		value := draft.ChannelID.String()
		channelID = &value
	}

	collectionIDs := []string{}
	for _, collectionID := range strings.Split(draft.CollectionIDs, ",") {
		trimmed := strings.TrimSpace(collectionID)
		if trimmed == "" {
			continue
		}
		collectionIDs = append(collectionIDs, trimmed)
	}

	return BlogDraftResponse{
		ID:            draft.ID,
		UserID:        draft.UserID,
		ContextKey:    draft.ContextKey,
		SourcePostID:  sourcePostID,
		Title:         draft.Title,
		Content:       draft.Content,
		Summary:       draft.Summary,
		CoverURL:      draft.CoverURL,
		AllowComments: draft.AllowComments,
		ChannelID:     channelID,
		CollectionIDs: collectionIDs,
		CreatedAt:     draft.CreatedAt,
		UpdatedAt:     draft.UpdatedAt,
	}
}
