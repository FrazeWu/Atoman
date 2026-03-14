package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"

	"atoman/internal/middleware"
	"atoman/internal/model"
)

// SetupBlogInteractionRoutes configures comment, like, and bookmark routes
func SetupBlogInteractionRoutes(router *gin.Engine, db *gorm.DB) {
	blog := router.Group("/api/blog")
	{
		// Public routes
		blog.GET("/posts/:id/comments", GetPostComments(db))
		blog.GET("/posts/:id/likes/count", GetPostLikesCount(db))

		// Protected routes
		protected := blog.Group("")
		protected.Use(middleware.AuthMiddleware())
		{
			protected.POST("/posts/:id/comments", CreateComment(db))
			protected.DELETE("/comments/:id", DeleteComment(db))

			protected.POST("/likes", ToggleLike(db, true))
			protected.DELETE("/likes", ToggleLike(db, false))

			protected.GET("/bookmarks", GetBookmarks(db))
			protected.POST("/bookmarks", CreateBookmark(db))
			protected.DELETE("/bookmarks/:id", DeleteBookmark(db))

			protected.GET("/bookmark-folders", GetBookmarkFolders(db))
			protected.POST("/bookmark-folders", CreateBookmarkFolder(db))
			protected.DELETE("/bookmark-folders/:id", DeleteBookmarkFolder(db))
		}
	}
}

// CommentInput represents the request body for creating a comment
type CommentInput struct {
	Content string `json:"content" binding:"required"`
}

// LikeInput represents the request body for liking/unliking
type LikeInput struct {
	TargetType string    `json:"target_type" binding:"required,oneof=post comment"`
	TargetID   uuid.UUID `json:"target_id" binding:"required"`
}

// BookmarkInput represents the request body for bookmarking
type BookmarkInput struct {
	PostID           uuid.UUID  `json:"post_id" binding:"required"`
	BookmarkFolderID *uuid.UUID `json:"bookmark_folder_id"`
}

// BookmarkFolderInput represents the request body for creating a bookmark folder
type BookmarkFolderInput struct {
	Name string `json:"name" binding:"required"`
}

// GetPostComments returns all visible comments for a post
func GetPostComments(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		postID := c.Param("id")
		var comments []model.Comment

		if err := db.Preload("User").Where("post_id = ? AND status = ?", postID, "visible").Order("created_at ASC").Find(&comments).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch comments"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"data": comments, "message": "ok"})
	}
}

// CreateComment creates a new comment on a post
func CreateComment(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		postID := c.Param("id")
		var input CommentInput
		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		var post model.Post
		if err := db.First(&post, postID).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Post not found"})
			return
		}

		if !post.AllowComments {
			c.JSON(http.StatusForbidden, gin.H{"error": "Comments are disabled for this post"})
			return
		}

		userIDVal, _ := c.Get("user_id")
		userID := userIDVal.(uuid.UUID)

		comment := model.Comment{
			PostID:  post.ID,
			UserID:  userID,
			Content: input.Content,
			Status:  "visible",
		}

		if err := db.Create(&comment).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create comment"})
			return
		}

		// Create notification for post owner (if not commenting on own post)
		if post.UserID != userID {
			notification := model.Notification{
				UserID:     post.UserID,
				Type:       "comment",
				Content:    "有人评论了你的文章",
				TargetType: "post",
				TargetID:   &post.ID,
			}
			db.Create(&notification)
		}

		c.JSON(http.StatusCreated, gin.H{"data": comment, "message": "ok"})
	}
}

// DeleteComment deletes a comment (by comment owner or post owner)
func DeleteComment(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		var comment model.Comment

		if err := db.Preload("Post").First(&comment, id).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Comment not found"})
			return
		}

		userIDVal, _ := c.Get("user_id")
		userID := userIDVal.(uuid.UUID)

		// Check if user is comment owner or post owner
		if comment.UserID != userID && comment.Post.UserID != userID {
			c.JSON(http.StatusForbidden, gin.H{"error": "You don't have permission to delete this comment"})
			return
		}

		if err := db.Delete(&comment).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete comment"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "ok"})
	}
}

// ToggleLike handles liking and unliking
func ToggleLike(db *gorm.DB, isLike bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		var input LikeInput
		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		userIDVal, _ := c.Get("user_id")
		userID := userIDVal.(uuid.UUID)

		var targetOwnerID uuid.UUID
		if input.TargetType == "post" {
			var post model.Post
			if err := db.First(&post, input.TargetID).Error; err != nil {
				c.JSON(http.StatusNotFound, gin.H{"error": "Post not found"})
				return
			}
			targetOwnerID = post.UserID
		} else {
			var comment model.Comment
			if err := db.First(&comment, input.TargetID).Error; err != nil {
				c.JSON(http.StatusNotFound, gin.H{"error": "Comment not found"})
				return
			}
			targetOwnerID = comment.UserID
		}

		if isLike {
			like := model.Like{
				UserID:     userID,
				TargetType: input.TargetType,
				TargetID:   input.TargetID,
			}

			// Use FirstOrCreate to prevent duplicate likes
			if err := db.Where(model.Like{UserID: userID, TargetType: input.TargetType, TargetID: input.TargetID}).FirstOrCreate(&like).Error; err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to like"})
				return
			}

			// Create notification if liking someone else's content
			if targetOwnerID != userID {
				notification := model.Notification{
					UserID:     targetOwnerID,
					Type:       "like",
					Content:    "有人点赞了你的" + func() string { if input.TargetType == "post" { return "文章" }; return "评论" }(),
					TargetType: input.TargetType,
					TargetID:   &input.TargetID,
				}
				db.Create(&notification)
			}
		} else {
			if err := db.Where("user_id = ? AND target_type = ? AND target_id = ?", userID, input.TargetType, input.TargetID).Delete(&model.Like{}).Error; err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to unlike"})
				return
			}
		}

		c.JSON(http.StatusOK, gin.H{"message": "ok"})
	}
}

// GetPostLikesCount returns the number of likes for a post
func GetPostLikesCount(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		postID := c.Param("id")
		var count int64

		if err := db.Model(&model.Like{}).Where("target_type = ? AND target_id = ?", "post", postID).Count(&count).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get likes count"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"data": gin.H{"count": count}, "message": "ok"})
	}
}

// GetBookmarks returns the authenticated user's bookmarks
func GetBookmarks(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userIDVal, _ := c.Get("user_id")
		userID := userIDVal.(uuid.UUID)

		var bookmarks []model.Bookmark
		query := db.Preload("Post").Preload("Post.User").Where("user_id = ?", userID)

		if folderID := c.Query("folder_id"); folderID != "" {
			query = query.Where("bookmark_folder_id = ?", folderID)
		}

		if err := query.Order("created_at DESC").Find(&bookmarks).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch bookmarks"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"data": bookmarks, "message": "ok"})
	}
}

// CreateBookmark creates a new bookmark
func CreateBookmark(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var input BookmarkInput
		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		var post model.Post
		if err := db.First(&post, input.PostID).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Post not found"})
			return
		}

		userIDVal, _ := c.Get("user_id")
		userID := userIDVal.(uuid.UUID)

		if input.BookmarkFolderID != nil {
			var folder model.BookmarkFolder
			if err := db.Where("id = ? AND user_id = ?", *input.BookmarkFolderID, userID).First(&folder).Error; err != nil {
				c.JSON(http.StatusNotFound, gin.H{"error": "Bookmark folder not found or doesn't belong to you"})
				return
			}
		}

		bookmark := model.Bookmark{
			UserID:           userID,
			PostID:           input.PostID,
			BookmarkFolderID: input.BookmarkFolderID,
		}

		if err := db.Where(model.Bookmark{UserID: userID, PostID: input.PostID}).FirstOrCreate(&bookmark).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create bookmark"})
			return
		}

		// Create notification if bookmarking someone else's post
		if post.UserID != userID {
			notification := model.Notification{
				UserID:     post.UserID,
				Type:       "bookmark",
				Content:    "有人收藏了你的文章",
				TargetType: "post",
				TargetID:   &post.ID,
			}
			db.Create(&notification)
		}

		c.JSON(http.StatusCreated, gin.H{"data": bookmark, "message": "ok"})
	}
}

// DeleteBookmark deletes a bookmark
func DeleteBookmark(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		userIDVal, _ := c.Get("user_id")
		userID := userIDVal.(uuid.UUID)

		if err := db.Where("id = ? AND user_id = ?", id, userID).Delete(&model.Bookmark{}).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete bookmark"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "ok"})
	}
}

// GetBookmarkFolders returns the authenticated user's bookmark folders
func GetBookmarkFolders(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userIDVal, _ := c.Get("user_id")
		userID := userIDVal.(uuid.UUID)

		var folders []model.BookmarkFolder
		if err := db.Where("user_id = ?", userID).Order("created_at DESC").Find(&folders).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch bookmark folders"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"data": folders, "message": "ok"})
	}
}

// CreateBookmarkFolder creates a new bookmark folder
func CreateBookmarkFolder(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var input BookmarkFolderInput
		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		userIDVal, _ := c.Get("user_id")
		userID := userIDVal.(uuid.UUID)

		folder := model.BookmarkFolder{
			UserID: userID,
			Name:   input.Name,
		}

		if err := db.Create(&folder).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create bookmark folder"})
			return
		}

		c.JSON(http.StatusCreated, gin.H{"data": folder, "message": "ok"})
	}
}

// DeleteBookmarkFolder deletes a bookmark folder
func DeleteBookmarkFolder(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		userIDVal, _ := c.Get("user_id")
		userID := userIDVal.(uuid.UUID)

		// Start transaction to delete folder and update bookmarks
		err := db.Transaction(func(tx *gorm.DB) error {
			// Set folder_id to null for all bookmarks in this folder
			if err := tx.Model(&model.Bookmark{}).Where("bookmark_folder_id = ? AND user_id = ?", id, userID).Update("bookmark_folder_id", nil).Error; err != nil {
				return err
			}

			// Delete the folder
			if err := tx.Where("id = ? AND user_id = ?", id, userID).Delete(&model.BookmarkFolder{}).Error; err != nil {
				return err
			}

			return nil
		})

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete bookmark folder"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "ok"})
	}
}
