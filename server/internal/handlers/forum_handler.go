package handlers

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"

	"atoman/internal/middleware"
	"atoman/internal/model"
	"atoman/internal/service"
)

func SetupForumRoutes(router *gin.Engine, db *gorm.DB) {
	forum := router.Group("/api/forum")
	{
		// Public / optional-auth routes
		forum.GET("/categories", GetForumCategories(db))
		forum.GET("/topics", GetForumTopics(db))
		forum.GET("/topics/:id", GetForumTopic(db))
		forum.GET("/topics/:id/replies", GetForumReplies(db))
		forum.GET("/search", SearchForumTopics(db))

		protected := forum.Group("")
		protected.Use(middleware.AuthMiddleware())
		{
			// Topics
			protected.POST("/topics", CreateForumTopic(db))
			protected.PUT("/topics/:id", UpdateForumTopic(db))
			protected.DELETE("/topics/:id", DeleteForumTopic(db))
			protected.POST("/topics/:id/like", ToggleForumTopicLike(db))
			protected.POST("/topics/:id/bookmark", ToggleForumTopicBookmark(db))
			protected.POST("/topics/:id/pin", PinForumTopic(db))
			protected.POST("/topics/:id/close", CloseForumTopic(db))

			// Replies
			protected.POST("/topics/:id/replies", CreateForumReply(db))
			protected.PUT("/replies/:id", UpdateForumReply(db))
			protected.DELETE("/replies/:id", DeleteForumReply(db))
			protected.POST("/replies/:id/like", ToggleForumReplyLike(db))

			// Drafts
			protected.GET("/drafts", GetForumDraft(db))
			protected.PUT("/drafts", PutForumDraft(db))
			protected.DELETE("/drafts", DeleteForumDraft(db))

			// Admin
			protected.POST("/categories", CreateForumCategory(db))
		}
	}
}

// ─── Categories ────────────────────────────────────────────────────────────────

func GetForumCategories(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var categories []model.ForumCategory
		if err := db.Find(&categories).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch categories"})
			return
		}
		for i := range categories {
			var count int64
			db.Model(&model.ForumTopic{}).Where("category_id = ? AND deleted_at IS NULL", categories[i].ID).Count(&count)
			categories[i].TopicCount = int(count)
		}
		c.JSON(http.StatusOK, gin.H{"data": categories})
	}
}

func CreateForumCategory(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		role, _ := c.Get("role")
		if role != "admin" {
			c.JSON(http.StatusForbidden, gin.H{"error": "Admin only"})
			return
		}
		var input struct {
			Name        string `json:"name" binding:"required"`
			Description string `json:"description"`
			Color       string `json:"color"`
		}
		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		cat := model.ForumCategory{Name: input.Name, Description: input.Description, Color: input.Color}
		if cat.Color == "" {
			cat.Color = "#000000"
		}
		if err := db.Create(&cat).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create category"})
			return
		}
		c.JSON(http.StatusCreated, gin.H{"data": cat})
	}
}

// ─── Topics ────────────────────────────────────────────────────────────────────

func GetForumTopics(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
		limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
		if page < 1 {
			page = 1
		}
		if limit < 1 || limit > 100 {
			limit = 20
		}
		offset := (page - 1) * limit

		sort := c.DefaultQuery("sort", "latest")
		categoryID := c.Query("category_id")
		tag := c.Query("tag")
		search := c.Query("search")

		query := db.Model(&model.ForumTopic{}).Preload("User").Preload("Category")
		if categoryID != "" {
			query = query.Where("category_id = ?", categoryID)
		}
		if tag != "" {
			dialect := db.Dialector.Name()
			if dialect == "postgres" || dialect == "pgx" {
				query = query.Where("? = ANY(tags)", tag)
			} else {
				query = query.Where("tags LIKE ?", "%"+tag+"%")
			}
		}
		if search != "" {
			query = query.Where("title ILIKE ? OR content ILIKE ?",
				"%"+search+"%", "%"+search+"%")
		}

		var total int64
		query.Count(&total)

		orderClause := "pinned DESC, created_at DESC"
		if sort == "top" {
			orderClause = "pinned DESC, like_count DESC, reply_count DESC"
		} else if sort == "active" {
			orderClause = "pinned DESC, COALESCE(last_reply_at, created_at) DESC"
		}

		var topics []model.ForumTopic
		if err := query.Order(orderClause).Limit(limit).Offset(offset).Find(&topics).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch topics"})
			return
		}

		enrichTopicsForUser(db, c, topics)

		c.JSON(http.StatusOK, gin.H{
			"data":  topics,
			"total": total,
			"page":  page,
			"limit": limit,
		})
	}
}

func GetForumTopic(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := uuid.Parse(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid topic id"})
			return
		}
		var topic model.ForumTopic
		if err := db.Preload("User").Preload("Category").First(&topic, "id = ?", id).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Topic not found"})
			return
		}

		// Increment view count (deduplicated per authenticated user per hour)
		currentUserID, hasUser := c.Get("user_id")
		if hasUser {
			uid, _ := currentUserID.(uuid.UUID)
			var recentView int64
			db.Model(&model.ActivityLog{}).
				Where("user_id = ? AND action = 'view_topic' AND target_id = ? AND created_at > ?",
					uid, topic.ID, time.Now().Add(-1*time.Hour)).
				Count(&recentView)
			if recentView == 0 {
				db.Model(&topic).UpdateColumn("view_count", gorm.Expr("view_count + 1"))
				topic.ViewCount++
				service.LogActivity(db, uid, "view_topic", "topic", topic.ID)
			}
			// Set like / bookmark status
			var like model.ForumLike
			if db.Where("user_id = ? AND target_type = ? AND target_id = ?", uid, "topic", topic.ID).First(&like).Error == nil {
				topic.IsLiked = true
			}
			var bookmark model.ForumBookmark
			if db.Where("user_id = ? AND topic_id = ?", uid, topic.ID).First(&bookmark).Error == nil {
				topic.IsBookmarked = true
			}
		} else {
			// Anonymous: always increment
			db.Model(&topic).UpdateColumn("view_count", gorm.Expr("view_count + 1"))
			topic.ViewCount++
		}

		c.JSON(http.StatusOK, gin.H{"data": topic})
	}
}

func CreateForumTopic(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var input struct {
			CategoryID string   `json:"category_id" binding:"required"`
			Title      string   `json:"title" binding:"required"`
			Content    string   `json:"content" binding:"required"`
			Tags       []string `json:"tags"`
		}
		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		catID, err := uuid.Parse(input.CategoryID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid category_id"})
			return
		}
		var cat model.ForumCategory
		if db.First(&cat, "id = ?", catID).Error != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Category not found"})
			return
		}
		userID, _ := c.Get("user_id")
		uid, _ := userID.(uuid.UUID)

		tags := model.StringSlice(input.Tags)
		if tags == nil {
			tags = model.StringSlice{}
		}

		topic := model.ForumTopic{
			UserID:     uid,
			CategoryID: catID,
			Title:      input.Title,
			Content:    input.Content,
			Tags:       tags,
		}
		if err := db.Create(&topic).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create topic"})
			return
		}
		db.Preload("User").Preload("Category").First(&topic, "id = ?", topic.ID)

		// Log activity
		service.LogActivity(db, uid, "create_topic", "topic", topic.ID)

		c.JSON(http.StatusCreated, gin.H{"data": topic})
	}
}

func UpdateForumTopic(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := uuid.Parse(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid topic id"})
			return
		}
		var topic model.ForumTopic
		if db.First(&topic, "id = ?", id).Error != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Topic not found"})
			return
		}
		userID, _ := c.Get("user_id")
		uid, _ := userID.(uuid.UUID)
		role, _ := c.Get("role")
		if topic.UserID != uid && role != "admin" {
			c.JSON(http.StatusForbidden, gin.H{"error": "Forbidden"})
			return
		}
		var input struct {
			Title   string   `json:"title"`
			Content string   `json:"content"`
			Tags    []string `json:"tags"`
		}
		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if input.Title != "" {
			topic.Title = input.Title
		}
		if input.Content != "" {
			topic.Content = input.Content
		}
		if input.Tags != nil {
			topic.Tags = model.StringSlice(input.Tags)
		}
		db.Save(&topic)
		c.JSON(http.StatusOK, gin.H{"data": topic})
	}
}

func DeleteForumTopic(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := uuid.Parse(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid topic id"})
			return
		}
		var topic model.ForumTopic
		if db.First(&topic, "id = ?", id).Error != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Topic not found"})
			return
		}
		userID, _ := c.Get("user_id")
		uid, _ := userID.(uuid.UUID)
		role, _ := c.Get("role")
		if topic.UserID != uid && role != "admin" {
			c.JSON(http.StatusForbidden, gin.H{"error": "Forbidden"})
			return
		}
		db.Delete(&topic)
		c.JSON(http.StatusOK, gin.H{"message": "Deleted"})
	}
}

func ToggleForumTopicLike(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := uuid.Parse(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid topic id"})
			return
		}
		userID, _ := c.Get("user_id")
		uid, _ := userID.(uuid.UUID)

		var like model.ForumLike
		if db.Where("user_id = ? AND target_type = ? AND target_id = ?", uid, "topic", id).First(&like).Error == nil {
			db.Delete(&like)
			db.Model(&model.ForumTopic{}).Where("id = ?", id).UpdateColumn("like_count", gorm.Expr("like_count - 1"))
			c.JSON(http.StatusOK, gin.H{"liked": false})
		} else {
			db.Create(&model.ForumLike{UserID: uid, TargetType: "topic", TargetID: id})
			db.Model(&model.ForumTopic{}).Where("id = ?", id).UpdateColumn("like_count", gorm.Expr("like_count + 1"))
			c.JSON(http.StatusOK, gin.H{"liked": true})
		}
	}
}

func ToggleForumTopicBookmark(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := uuid.Parse(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid topic id"})
			return
		}
		userID, _ := c.Get("user_id")
		uid, _ := userID.(uuid.UUID)

		var bookmark model.ForumBookmark
		if db.Where("user_id = ? AND topic_id = ?", uid, id).First(&bookmark).Error == nil {
			db.Delete(&bookmark)
			c.JSON(http.StatusOK, gin.H{"bookmarked": false})
		} else {
			db.Create(&model.ForumBookmark{UserID: uid, TopicID: id})
			c.JSON(http.StatusOK, gin.H{"bookmarked": true})
		}
	}
}

func PinForumTopic(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		role, _ := c.Get("role")
		if role != "admin" {
			c.JSON(http.StatusForbidden, gin.H{"error": "Admin only"})
			return
		}
		id, err := uuid.Parse(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid topic id"})
			return
		}
		var topic model.ForumTopic
		if db.First(&topic, "id = ?", id).Error != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Topic not found"})
			return
		}
		topic.Pinned = !topic.Pinned
		db.Save(&topic)
		c.JSON(http.StatusOK, gin.H{"pinned": topic.Pinned})
	}
}

func CloseForumTopic(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		role, _ := c.Get("role")
		if role != "admin" {
			c.JSON(http.StatusForbidden, gin.H{"error": "Admin only"})
			return
		}
		id, err := uuid.Parse(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid topic id"})
			return
		}
		var topic model.ForumTopic
		if db.First(&topic, "id = ?", id).Error != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Topic not found"})
			return
		}
		topic.Closed = !topic.Closed
		db.Save(&topic)
		c.JSON(http.StatusOK, gin.H{"closed": topic.Closed})
	}
}

// ─── Replies ───────────────────────────────────────────────────────────────────

func GetForumReplies(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		topicID, err := uuid.Parse(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid topic id"})
			return
		}

		sort := c.DefaultQuery("sort", "oldest") // oldest | best

		var allReplies []model.ForumReply
		query := db.Preload("User").Where("topic_id = ? AND deleted_at IS NULL", topicID)
		if sort == "best" {
			query = query.Order("like_count DESC, floor_number ASC")
		} else {
			query = query.Order("floor_number ASC")
		}
		if err := query.Find(&allReplies).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch replies"})
			return
		}

		// Enrich with per-user like status
		currentUserID, hasUser := c.Get("user_id")
		if hasUser {
			uid, _ := currentUserID.(uuid.UUID)
			replyIDs := make([]uuid.UUID, len(allReplies))
			for i, r := range allReplies {
				replyIDs[i] = r.ID
			}
			var likes []model.ForumLike
			db.Where("user_id = ? AND target_type = ? AND target_id IN ?", uid, "reply", replyIDs).Find(&likes)
			likedSet := map[uuid.UUID]bool{}
			for _, l := range likes {
				likedSet[l.TargetID] = true
			}
			for i := range allReplies {
				allReplies[i].IsLiked = likedSet[allReplies[i].ID]
			}
		}

		c.JSON(http.StatusOK, gin.H{"data": allReplies})
	}
}

func CreateForumReply(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		topicID, err := uuid.Parse(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid topic id"})
			return
		}
		var topic model.ForumTopic
		if db.First(&topic, "id = ?", topicID).Error != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Topic not found"})
			return
		}
		if topic.Closed {
			c.JSON(http.StatusForbidden, gin.H{"error": "Topic is closed"})
			return
		}

		var input struct {
			Content       string  `json:"content" binding:"required"`
			ParentReplyID *string `json:"parent_reply_id"`
		}
		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		userID, _ := c.Get("user_id")
		uid, _ := userID.(uuid.UUID)

		var parentUID *uuid.UUID
		if input.ParentReplyID != nil && *input.ParentReplyID != "" {
			pid, err := uuid.Parse(*input.ParentReplyID)
			if err == nil {
				parentUID = &pid
			}
		}

		if parentUID != nil {
			var quotedReply model.ForumReply
			if err := db.Select("id").First(&quotedReply, "id = ? AND topic_id = ? AND deleted_at IS NULL", *parentUID, topicID).Error; err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Quoted reply not found"})
				return
			}
		}

		// Calculate flat floor order; parent_reply_id is now quote metadata only.
		path, floor, err := service.BuildReplyPath(db, topicID, parentUID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to compute reply path"})
			return
		}

		reply := model.ForumReply{
			TopicID:       topicID,
			UserID:        uid,
			ParentReplyID: parentUID,
			Content:       input.Content,
			Path:          path,
			FloorNumber:   floor,
		}
		if err := db.Create(&reply).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create reply"})
			return
		}

		// Update topic counters
		now := time.Now()
		db.Model(&topic).Updates(map[string]interface{}{
			"reply_count":   gorm.Expr("reply_count + 1"),
			"last_reply_at": now,
		})

		db.Preload("User").First(&reply, "id = ?", reply.ID)


		// 3. Log activity
		service.LogActivity(db, uid, "create_reply", "reply", reply.ID)

		c.JSON(http.StatusCreated, gin.H{"data": reply})
	}
}

func UpdateForumReply(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := uuid.Parse(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid reply id"})
			return
		}
		var reply model.ForumReply
		if db.First(&reply, "id = ?", id).Error != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Reply not found"})
			return
		}
		userID, _ := c.Get("user_id")
		uid, _ := userID.(uuid.UUID)
		role, _ := c.Get("role")
		if reply.UserID != uid && role != "admin" {
			c.JSON(http.StatusForbidden, gin.H{"error": "Forbidden"})
			return
		}
		var input struct {
			Content string `json:"content" binding:"required"`
		}
		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		reply.Content = input.Content
		db.Save(&reply)
		c.JSON(http.StatusOK, gin.H{"data": reply})
	}
}

func DeleteForumReply(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := uuid.Parse(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid reply id"})
			return
		}
		var reply model.ForumReply
		if db.First(&reply, "id = ?", id).Error != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Reply not found"})
			return
		}
		userID, _ := c.Get("user_id")
		uid, _ := userID.(uuid.UUID)
		role, _ := c.Get("role")
		if reply.UserID != uid && role != "admin" {
			c.JSON(http.StatusForbidden, gin.H{"error": "Forbidden"})
			return
		}
		db.Model(&model.ForumReply{}).Where("parent_reply_id = ?", reply.ID).Update("parent_reply_id", nil)
		db.Delete(&reply)
		db.Model(&model.ForumTopic{}).Where("id = ?", reply.TopicID).
			UpdateColumn("reply_count", gorm.Expr("reply_count - 1"))
		c.JSON(http.StatusOK, gin.H{"message": "Deleted"})
	}
}

func ToggleForumReplyLike(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := uuid.Parse(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid reply id"})
			return
		}
		userID, _ := c.Get("user_id")
		uid, _ := userID.(uuid.UUID)

		var like model.ForumLike
		if db.Where("user_id = ? AND target_type = ? AND target_id = ?", uid, "reply", id).First(&like).Error == nil {
			db.Delete(&like)
			db.Model(&model.ForumReply{}).Where("id = ?", id).UpdateColumn("like_count", gorm.Expr("like_count - 1"))
			c.JSON(http.StatusOK, gin.H{"liked": false})
		} else {
			db.Create(&model.ForumLike{UserID: uid, TargetType: "reply", TargetID: id})
			db.Model(&model.ForumReply{}).Where("id = ?", id).UpdateColumn("like_count", gorm.Expr("like_count + 1"))
			// Notify reply owner
			var replyOwner model.ForumReply
			if db.First(&replyOwner, "id = ?", id).Error == nil && replyOwner.UserID != uid {
				service.LogActivity(db, replyOwner.UserID, "receive_like", "reply", id)
			}
			c.JSON(http.StatusOK, gin.H{"liked": true})
		}
	}
}

// ─── Search ────────────────────────────────────────────────────────────────────

func SearchForumTopics(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		q := strings.TrimSpace(c.Query("q"))
		if q == "" {
			c.JSON(http.StatusOK, gin.H{"data": []model.ForumTopic{}, "total": 0})
			return
		}
		page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
		limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
		if page < 1 {
			page = 1
		}
		if limit < 1 || limit > 100 {
			limit = 20
		}
		offset := (page - 1) * limit

		var topics []model.ForumTopic
		var total int64
		query := db.Model(&model.ForumTopic{}).Preload("User").Preload("Category").
			Where("(title ILIKE ? OR content ILIKE ?) AND deleted_at IS NULL",
				"%"+q+"%", "%"+q+"%")
		query.Count(&total)
		if err := query.Order("created_at DESC").Limit(limit).Offset(offset).Find(&topics).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Search failed"})
			return
		}
		enrichTopicsForUser(db, c, topics)
		c.JSON(http.StatusOK, gin.H{"data": topics, "total": total, "page": page, "limit": limit, "q": q})
	}
}

// ─── Drafts ────────────────────────────────────────────────────────────────────

func GetForumDraft(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		contextKey := c.Query("context_key")
		if contextKey == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "context_key required"})
			return
		}
		userID, _ := c.Get("user_id")
		uid, _ := userID.(uuid.UUID)

		var draft model.ForumDraft
		if err := db.Where("user_id = ? AND context_key = ?", uid, contextKey).First(&draft).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Draft not found"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": draft})
	}
}

func PutForumDraft(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var input struct {
			ContextKey string `json:"context_key" binding:"required"`
			Title      string `json:"title"`
			Content    string `json:"content"`
			Tags       string `json:"tags"`
		}
		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		userID, _ := c.Get("user_id")
		uid, _ := userID.(uuid.UUID)

		var draft model.ForumDraft
		result := db.Where("user_id = ? AND context_key = ?", uid, input.ContextKey).First(&draft)
		if result.Error != nil {
			// Create new
			draft = model.ForumDraft{
				UserID:     uid,
				ContextKey: input.ContextKey,
				Title:      input.Title,
				Content:    input.Content,
				Tags:       input.Tags,
			}
			db.Create(&draft)
		} else {
			// Update existing
			draft.Title = input.Title
			draft.Content = input.Content
			draft.Tags = input.Tags
			db.Save(&draft)
		}
		c.JSON(http.StatusOK, gin.H{"data": draft})
	}
}

func DeleteForumDraft(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		contextKey := c.Query("context_key")
		if contextKey == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "context_key required"})
			return
		}
		userID, _ := c.Get("user_id")
		uid, _ := userID.(uuid.UUID)
		db.Where("user_id = ? AND context_key = ?", uid, contextKey).Delete(&model.ForumDraft{})
		c.JSON(http.StatusOK, gin.H{"message": "Draft deleted"})
	}
}

// ─── Helpers ───────────────────────────────────────────────────────────────────

// enrichTopicsForUser bulk-loads like/bookmark status for the authenticated user.
func enrichTopicsForUser(db *gorm.DB, c *gin.Context, topics []model.ForumTopic) {
	currentUserID, hasUser := c.Get("user_id")
	if !hasUser || len(topics) == 0 {
		return
	}
	uid, _ := currentUserID.(uuid.UUID)

	topicIDs := make([]uuid.UUID, len(topics))
	for i, t := range topics {
		topicIDs[i] = t.ID
	}

	// Likes
	var likes []model.ForumLike
	db.Where("user_id = ? AND target_type = ? AND target_id IN ?", uid, "topic", topicIDs).Find(&likes)
	likedSet := map[uuid.UUID]bool{}
	for _, l := range likes {
		likedSet[l.TargetID] = true
	}

	// Bookmarks
	var bookmarks []model.ForumBookmark
	db.Where("user_id = ? AND topic_id IN ?", uid, topicIDs).Find(&bookmarks)
	bookmarkSet := map[uuid.UUID]bool{}
	for _, b := range bookmarks {
		bookmarkSet[b.TopicID] = true
	}

	for i := range topics {
		topics[i].IsLiked = likedSet[topics[i].ID]
		topics[i].IsBookmarked = bookmarkSet[topics[i].ID]
	}
}

// truncate shortens a string to maxLen runes.
func truncate(s string, maxLen int) string {
	runes := []rune(s)
	if len(runes) <= maxLen {
		return s
	}
	return string(runes[:maxLen]) + "..."
}

// getDisplayName returns display_name or username for a User pointer.
func getDisplayName(u *model.User) string {
	if u == nil {
		return "匿名"
	}
	if u.DisplayName != "" {
		return u.DisplayName
	}
	return u.Username
}
