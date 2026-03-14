package handlers

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/xml"
	"fmt"
	"net/http"
	"os"
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

// SetupFeedRoutes configures feed and subscription routes
func SetupFeedRoutes(router *gin.Engine, db *gorm.DB) {
	feed := router.Group("/api/feed")
	{
		// Public routes
		feed.GET("/rss/:username", GetUserRSS(db))

		// Protected routes
		protected := feed.Group("")
		protected.Use(middleware.AuthMiddleware())
		{
			protected.POST("/subscriptions", CreateSubscription(db))
			protected.DELETE("/subscriptions/:id", DeleteSubscription(db))
			protected.GET("/subscriptions", GetSubscriptions(db))
			protected.GET("/timeline", GetTimeline(db))
		}
	}
}

// SetupNotificationRoutes configures notification routes
func SetupNotificationRoutes(router *gin.Engine, db *gorm.DB) {
	notifications := router.Group("/api/notifications")
	notifications.Use(middleware.AuthMiddleware())
	{
		notifications.GET("", GetNotifications(db))
		notifications.PUT("/:id/read", MarkNotificationRead(db))
		notifications.PUT("/read-all", MarkAllNotificationsRead(db))
		notifications.GET("/unread-count", GetUnreadNotificationCount(db))
	}
}

// SubscriptionInput represents the request body for subscribing
type SubscriptionInput struct {
	TargetType string     `json:"target_type" binding:"required,oneof=internal_user internal_channel internal_collection external_rss"`
	TargetID   *uuid.UUID `json:"target_id"` // Required for internal types
	RssURL     string     `json:"rss_url"`   // Required for external_rss
	Title      string     `json:"title"`     // Optional custom title
}

// CreateSubscription creates a new feed source subscription
func CreateSubscription(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var input SubscriptionInput
		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Validate inputs based on type
		if input.TargetType != "external_rss" && input.TargetID == nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "target_id is required for internal subscriptions"})
			return
		}
		if input.TargetType == "external_rss" && input.RssURL == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "rss_url is required for external subscriptions"})
			return
		}

		userIDVal, _ := c.Get("user_id")
		userID := userIDVal.(uuid.UUID)

		// 1. Generate unique hash for this source to allow cross-user indexing
		var sourceHash string
		if input.TargetType == "external_rss" {
			h := sha256.New()
			h.Write([]byte(strings.TrimSpace(input.RssURL)))
			sourceHash = hex.EncodeToString(h.Sum(nil))
		} else {
			// For internal types: hash by type and ID
			sourceHash = fmt.Sprintf("%s:%s", input.TargetType, input.TargetID.String())
			h := sha256.New()
			h.Write([]byte(sourceHash))
			sourceHash = hex.EncodeToString(h.Sum(nil))
		}

		// 2. Find or Create the global FeedSource
		var source model.FeedSource
		if err := db.Where("hash = ?", sourceHash).First(&source).Error; err != nil {
			// If not found, create new source
			source = model.FeedSource{
				SourceType: input.TargetType,
				SourceID:   input.TargetID,
				RssURL:     input.RssURL,
				Hash:       sourceHash,
			}

			// For internal sources, pre-fill title
			if input.TargetType == "internal_user" {
				var user model.User
				if err := db.Where("uuid = ?", input.TargetID).First(&user).Error; err == nil {
					source.Title = user.Username
				}
			} else if input.TargetType == "internal_channel" {
				var channel model.Channel
				if err := db.First(&channel, input.TargetID).Error; err == nil {
					source.Title = channel.Name
				}
			} else if input.TargetType == "internal_collection" {
				var collection model.Collection
				if err := db.First(&collection, input.TargetID).Error; err == nil {
					source.Title = collection.Name
				}
			} else if input.TargetType == "external_rss" {
				// Fetch title for RSS
				_, sourceTitle, err := service.FetchAndParseRSS(input.RssURL)
				if err == nil {
					source.Title = sourceTitle
				}
			}

			if err := db.Create(&source).Error; err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create feed source"})
				return
			}

			// Trigger first fetch if RSS
			if input.TargetType == "external_rss" {
				go service.SyncSingleRSS(db, source)
			}
		}

		// 3. Check if user already subscribed to this source
		var existingSub model.Subscription
		if err := db.Where("user_id = ? AND feed_source_id = ?", userID, source.ID).First(&existingSub).Error; err == nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Already subscribed to this source"})
			return
		}

		// 4. Create the User Subscription
		subscription := model.Subscription{
			UserID:       userID,
			FeedSourceID: source.ID,
			Title:        input.Title, // User-defined custom title
		}

		if err := db.Create(&subscription).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create subscription"})
			return
		}

		c.JSON(http.StatusCreated, gin.H{"data": subscription, "message": "ok"})
	}
}

// DeleteSubscription removes a user's subscription to a feed source
func DeleteSubscription(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		userIDVal, _ := c.Get("user_id")
		userID := userIDVal.(uuid.UUID)

		if err := db.Where("id = ? AND user_id = ?", id, userID).Delete(&model.Subscription{}).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete subscription"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "ok"})
	}
}

// GetSubscriptions returns the authenticated user's subscriptions with source info
func GetSubscriptions(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userIDVal, _ := c.Get("user_id")
		userID := userIDVal.(uuid.UUID)

		var subscriptions []model.Subscription
		if err := db.Preload("FeedSource").Where("user_id = ?", userID).Find(&subscriptions).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch subscriptions"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"data": subscriptions, "message": "ok"})
	}
}

// TimelineItem represents a unified item in the feed timeline
type TimelineItem struct {
	Type        string           `json:"type"` // "post" or "orbit_item"
	Post        *model.Post      `json:"post,omitempty"`
	OrbitItem   *model.OrbitItem `json:"orbit_item,omitempty"`
	PublishedAt time.Time        `json:"published_at"`
}

// GetTimeline returns a unified timeline of internal posts and external RSS items
func GetTimeline(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userIDVal, _ := c.Get("user_id")
		userID := userIDVal.(uuid.UUID)

		page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
		limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
		offset := (page - 1) * limit

		sourceType := c.Query("source_type")
		sourceID := c.Query("source_id")

		var userSubscriptions []model.Subscription
		query := db.Where("user_id = ?", userID)

		if sourceType != "" {
			query = query.Joins("JOIN feed_sources ON feed_sources.id = subscriptions.feed_source_id").
				Where("feed_sources.source_type = ?", sourceType)
		}
		if sourceID != "" {
			query = query.Where("subscriptions.id = ?", sourceID)
		}

		if err := query.Preload("FeedSource").Find(&userSubscriptions).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch subscriptions"})
			return
		}

		if len(userSubscriptions) == 0 {
			c.JSON(http.StatusOK, gin.H{"data": []TimelineItem{}, "message": "ok"})
			return
		}

		var userIDs []uuid.UUID
		var channelIDs []uuid.UUID
		var collectionIDs []uuid.UUID
		var feedSourceIDs []uuid.UUID

		for _, sub := range userSubscriptions {
			fs := sub.FeedSource
			if fs == nil {
				continue
			}
			switch fs.SourceType {
			case "internal_user":
				if fs.SourceID != nil {
					userIDs = append(userIDs, *fs.SourceID)
				}
			case "internal_channel":
				if fs.SourceID != nil {
					channelIDs = append(channelIDs, *fs.SourceID)
				}
			case "internal_collection":
				if fs.SourceID != nil {
					collectionIDs = append(collectionIDs, *fs.SourceID)
				}
			case "external_rss":
				feedSourceIDs = append(feedSourceIDs, fs.ID)
			}
		}

		// Fetch internal posts
		var posts []model.Post

		// Build complex OR condition for internal sources
		var orConditions []string
		var orArgs []interface{}

		if len(userIDs) > 0 {
			orConditions = append(orConditions, "user_id IN ?")
			orArgs = append(orArgs, userIDs)
		}

		if len(channelIDs) > 0 {
			// Get collections for these channels
			var channelCollections []model.Collection
			db.Where("channel_id IN ?", channelIDs).Find(&channelCollections)
			for _, c := range channelCollections {
				collectionIDs = append(collectionIDs, c.ID)
			}
		}

		if len(collectionIDs) > 0 {
			// Get post IDs for these collections
			var postCollections []model.PostCollection
			db.Where("collection_id IN ?", collectionIDs).Find(&postCollections)
			var postIDs []uuid.UUID
			for _, pc := range postCollections {
				postIDs = append(postIDs, pc.PostID)
			}
			if len(postIDs) > 0 {
				orConditions = append(orConditions, "id IN ?")
				orArgs = append(orArgs, postIDs)
			}
		}

		if len(orConditions) > 0 {
			combinedQuery := "(" + strings.Join(orConditions, " OR ") + ")"
			db.Preload("User").Where("status = ?", "published").Where(combinedQuery, orArgs...).Find(&posts)
		}

		// Fetch external items
		var orbitItems []model.OrbitItem
		if len(feedSourceIDs) > 0 {
			db.Preload("FeedSource").Where("feed_source_id IN ?", feedSourceIDs).Order("published_at DESC").Limit(limit).Offset(offset).Find(&orbitItems)
		}

		// Combine and sort
		var timeline []TimelineItem

		for i := range posts {
			timeline = append(timeline, TimelineItem{
				Type:        "post",
				Post:        &posts[i],
				PublishedAt: posts[i].CreatedAt,
			})
		}

		for i := range orbitItems {
			timeline = append(timeline, TimelineItem{
				Type:        "orbit_item",
				OrbitItem:   &orbitItems[i],
				PublishedAt: orbitItems[i].PublishedAt,
			})
		}

		c.JSON(http.StatusOK, gin.H{"data": timeline, "message": "ok"})
	}
}

// RSS structs
type RSS struct {
	XMLName xml.Name   `xml:"rss"`
	Version string     `xml:"version,attr"`
	Channel RSSChannel `xml:"channel"`
}

type RSSChannel struct {
	Title       string    `xml:"title"`
	Link        string    `xml:"link"`
	Description string    `xml:"description"`
	Items       []RSSItem `xml:"item"`
}

type RSSItem struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	Description string `xml:"description"`
	PubDate     string `xml:"pubDate"`
	GUID        string `xml:"guid"`
}

// GetUserRSS generates an RSS feed for a user's published posts
func GetUserRSS(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		username := c.Param("username")

		var user model.User
		if err := db.Where("username = ?", username).First(&user).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
			return
		}

		var posts []model.Post
		if err := db.Where("user_id = ? AND status = ?", user.UUID, "published").Order("created_at DESC").Limit(50).Find(&posts).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch posts"})
			return
		}

		baseURL := os.Getenv("BASE_URL")
		if baseURL == "" {
			baseURL = "http://localhost:8080"
		}

		rss := RSS{
			Version: "2.0",
			Channel: RSSChannel{
				Title:       user.DisplayName + " 的博客 - Atoman",
				Link:        baseURL + "/blog/@" + user.Username,
				Description: user.DisplayName + " 的博客订阅",
				Items:       []RSSItem{},
			},
		}

		for _, post := range posts {
			itemURL := baseURL + "/blog/posts/" + post.ID.String()
			rss.Channel.Items = append(rss.Channel.Items, RSSItem{
				Title:       post.Title,
				Link:        itemURL,
				Description: post.Summary,
				PubDate:     post.CreatedAt.Format(time.RFC1123Z),
				GUID:        itemURL,
			})
		}

		c.Header("Content-Type", "application/rss+xml")
		c.XML(http.StatusOK, rss)
	}
}

// GetNotifications returns the authenticated user's notifications
func GetNotifications(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userIDVal, _ := c.Get("user_id")
		userID := userIDVal.(uuid.UUID)

		page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
		limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
		offset := (page - 1) * limit

		unreadOnly := c.Query("unread") == "true"

		var notifications []model.Notification
		query := db.Where("user_id = ?", userID)

		if unreadOnly {
			query = query.Where("read_at IS NULL")
		}

		if err := query.Order("created_at DESC").Limit(limit).Offset(offset).Find(&notifications).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch notifications"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"data": notifications, "message": "ok"})
	}
}

// MarkNotificationRead marks a specific notification as read
func MarkNotificationRead(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		userIDVal, _ := c.Get("user_id")
		userID := userIDVal.(uuid.UUID)

		now := time.Now()
		if err := db.Model(&model.Notification{}).Where("id = ? AND user_id = ?", id, userID).Update("read_at", now).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to mark notification as read"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "ok"})
	}
}

// MarkAllNotificationsRead marks all notifications as read for the user
func MarkAllNotificationsRead(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userIDVal, _ := c.Get("user_id")
		userID := userIDVal.(uuid.UUID)

		now := time.Now()
		if err := db.Model(&model.Notification{}).Where("user_id = ? AND read_at IS NULL", userID).Update("read_at", now).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to mark all notifications as read"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "ok"})
	}
}

// GetUnreadNotificationCount returns the count of unread notifications
func GetUnreadNotificationCount(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userIDVal, _ := c.Get("user_id")
		userID := userIDVal.(uuid.UUID)

		var count int64
		if err := db.Model(&model.Notification{}).Where("user_id = ? AND read_at IS NULL", userID).Count(&count).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get unread count"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"unread_count": count, "message": "ok"})
	}
}
