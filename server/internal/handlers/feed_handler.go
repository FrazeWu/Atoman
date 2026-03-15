package handlers

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/xml"
	"fmt"
	"net/http"
	"os"
	"sort"
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

func SetupFeedRoutes(router *gin.Engine, db *gorm.DB) {
	feed := router.Group("/api/feed")
	{
		feed.GET("/rss/:username", GetUserRSS(db))

		protected := feed.Group("")
		protected.Use(middleware.AuthMiddleware())
		{
			protected.POST("/subscriptions", CreateSubscription(db))
			protected.DELETE("/subscriptions/:id", DeleteSubscription(db))
			protected.GET("/subscriptions", GetSubscriptions(db))
			protected.GET("/timeline", GetTimeline(db))

			protected.POST("/timeline/mark-read", MarkItemsRead(db))
			protected.POST("/timeline/mark-all-read", MarkAllRead(db))

			protected.GET("/groups", GetSubscriptionGroups(db))
			protected.POST("/groups", CreateSubscriptionGroup(db))
			protected.PUT("/groups/:id", UpdateSubscriptionGroup(db))
			protected.DELETE("/groups/:id", DeleteSubscriptionGroup(db))
			protected.PUT("/subscriptions/:id/group", SetSubscriptionGroup(db))
		}
	}
}

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

const defaultSubscriptionGroupName = "默认分组"

func getOrCreateDefaultSubscriptionGroup(db *gorm.DB, userID uuid.UUID) (*model.SubscriptionGroup, error) {
	var canonical model.SubscriptionGroup

	err := db.Transaction(func(tx *gorm.DB) error {
		var groups []model.SubscriptionGroup
		if err := tx.Where("user_id = ? AND name = ?", userID, defaultSubscriptionGroupName).
			Order("created_at ASC").Find(&groups).Error; err != nil {
			return err
		}

		switch len(groups) {
		case 0:
			canonical = model.SubscriptionGroup{
				UserID: userID,
				Name:   defaultSubscriptionGroupName,
			}
			if err := tx.Create(&canonical).Error; err != nil {
				return err
			}
		case 1:
			canonical = groups[0]
		default:
			canonical = groups[0]
			duplicateIDs := make([]uuid.UUID, 0, len(groups)-1)
			for _, g := range groups[1:] {
				duplicateIDs = append(duplicateIDs, g.ID)
			}

			if err := tx.Model(&model.Subscription{}).
				Where("user_id = ? AND subscription_group_id IN ?", userID, duplicateIDs).
				Update("subscription_group_id", canonical.ID).Error; err != nil {
				return err
			}

			if err := tx.Where("user_id = ? AND id IN ?", userID, duplicateIDs).
				Delete(&model.SubscriptionGroup{}).Error; err != nil {
				return err
			}
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return &canonical, nil
}

type SubscriptionInput struct {
	TargetType string     `json:"target_type" binding:"required,oneof=internal_user internal_channel internal_collection external_rss"`
	TargetID   *uuid.UUID `json:"target_id"`
	RssURL     string     `json:"rss_url"`
	Title      string     `json:"title"`
}

func CreateSubscription(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var input SubscriptionInput
		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

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

		defaultGroup, err := getOrCreateDefaultSubscriptionGroup(db, userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to prepare default group"})
			return
		}

		var sourceHash string
		if input.TargetType == "external_rss" {
			h := sha256.New()
			h.Write([]byte(strings.TrimSpace(input.RssURL)))
			sourceHash = hex.EncodeToString(h.Sum(nil))
		} else {
			raw := fmt.Sprintf("%s:%s", input.TargetType, input.TargetID.String())
			h := sha256.New()
			h.Write([]byte(raw))
			sourceHash = hex.EncodeToString(h.Sum(nil))
		}

		var source model.FeedSource
		if err := db.Where("hash = ?", sourceHash).First(&source).Error; err != nil {
			source = model.FeedSource{
				SourceType: input.TargetType,
				SourceID:   input.TargetID,
				RssURL:     input.RssURL,
				Hash:       sourceHash,
			}

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
				_, sourceTitle, err := service.FetchAndParseRSS(input.RssURL)
				if err == nil {
					source.Title = sourceTitle
				}
			}

			if err := db.Create(&source).Error; err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create feed source"})
				return
			}

		}

		var existingSub model.Subscription
		if err := db.Where("user_id = ? AND feed_source_id = ?", userID, source.ID).First(&existingSub).Error; err == nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Already subscribed to this source"})
			return
		}

		subscription := model.Subscription{
			UserID:              userID,
			FeedSourceID:        source.ID,
			Title:               input.Title,
			SubscriptionGroupID: &defaultGroup.ID,
		}

		if err := db.Create(&subscription).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create subscription"})
			return
		}

		// Always trigger immediate RSS sync for external subscriptions,
		// even if the feed source already existed before this subscription.
		if input.TargetType == "external_rss" {
			go service.SyncSingleRSS(db, source)
		}

		c.JSON(http.StatusCreated, gin.H{"data": subscription, "message": "ok"})
	}
}

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

func GetSubscriptions(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userIDVal, _ := c.Get("user_id")
		userID := userIDVal.(uuid.UUID)

		defaultGroup, err := getOrCreateDefaultSubscriptionGroup(db, userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to prepare default group"})
			return
		}

		// Keep old data compatible: migrate NULL group subscriptions to default group.
		if err := db.Model(&model.Subscription{}).
			Where("user_id = ? AND subscription_group_id IS NULL", userID).
			Update("subscription_group_id", defaultGroup.ID).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to normalize subscriptions"})
			return
		}

		var subscriptions []model.Subscription
		if err := db.Preload("FeedSource").Preload("SubscriptionGroup").Where("user_id = ?", userID).Find(&subscriptions).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch subscriptions"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"data": subscriptions, "message": "ok"})
	}
}

type TimelineItem struct {
	Type        string          `json:"type"`
	Post        *model.Post     `json:"post,omitempty"`
	FeedItem    *model.FeedItem `json:"feed_item,omitempty"`
	PublishedAt time.Time       `json:"published_at"`
	IsRead      bool            `json:"is_read"`
}

func GetTimeline(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userIDVal, _ := c.Get("user_id")
		userID := userIDVal.(uuid.UUID)

		page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
		limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
		if limit > 100 {
			limit = 100
		}
		offset := (page - 1) * limit

		sourceType := c.Query("source_type")
		sourceID := c.Query("source_id")
		groupID := c.Query("group_id")

		var userSubscriptions []model.Subscription
		query := db.Where("subscriptions.user_id = ?", userID)

		if sourceType != "" {
			query = query.Joins("JOIN feed_sources ON feed_sources.id = subscriptions.feed_source_id").
				Where("feed_sources.source_type = ?", sourceType)
		}
		if sourceID != "" {
			query = query.Where("subscriptions.id = ?", sourceID)
		}
		if groupID != "" {
			query = query.Where("subscriptions.subscription_group_id = ?", groupID)
		}

		if err := query.Preload("FeedSource").Find(&userSubscriptions).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch subscriptions"})
			return
		}

		if len(userSubscriptions) == 0 {
			c.JSON(http.StatusOK, gin.H{"data": []TimelineItem{}, "total": 0, "page": page, "limit": limit, "message": "ok"})
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

		var posts []model.Post
		var orConditions []string
		var orArgs []interface{}

		if len(userIDs) > 0 {
			orConditions = append(orConditions, "user_id IN ?")
			orArgs = append(orArgs, userIDs)
		}

		if len(channelIDs) > 0 {
			var channelCollections []model.Collection
			db.Where("channel_id IN ?", channelIDs).Find(&channelCollections)
			for _, col := range channelCollections {
				collectionIDs = append(collectionIDs, col.ID)
			}
		}

		if len(collectionIDs) > 0 {
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
			combined := "(" + strings.Join(orConditions, " OR ") + ")"
			db.Preload("User").Where("status = ?", "published").Where(combined, orArgs...).Find(&posts)
		}

		var feedItems []model.FeedItem
		if len(feedSourceIDs) > 0 {
			db.Preload("FeedSource").Where("feed_source_id IN ?", feedSourceIDs).Order("published_at DESC").Find(&feedItems)
		}

		var readFeedItemIDs map[uuid.UUID]bool
		if len(feedItems) > 0 {
			var feedItemIDs []uuid.UUID
			for _, fi := range feedItems {
				feedItemIDs = append(feedItemIDs, fi.ID)
			}
			var reads []model.FeedItemRead
			db.Where("user_id = ? AND feed_item_id IN ?", userID, feedItemIDs).Find(&reads)
			readFeedItemIDs = make(map[uuid.UUID]bool, len(reads))
			for _, r := range reads {
				readFeedItemIDs[r.FeedItemID] = true
			}
		}

		var timeline []TimelineItem

		for i := range posts {
			timeline = append(timeline, TimelineItem{
				Type:        "post",
				Post:        &posts[i],
				PublishedAt: posts[i].CreatedAt,
				IsRead:      false,
			})
		}

		for i := range feedItems {
			timeline = append(timeline, TimelineItem{
				Type:        "feed_item",
				FeedItem:    &feedItems[i],
				PublishedAt: feedItems[i].PublishedAt,
				IsRead:      readFeedItemIDs[feedItems[i].ID],
			})
		}

		sort.Slice(timeline, func(i, j int) bool {
			return timeline[i].PublishedAt.After(timeline[j].PublishedAt)
		})

		total := len(timeline)
		start := offset
		if start > total {
			start = total
		}
		end := start + limit
		if end > total {
			end = total
		}
		paged := timeline[start:end]

		c.JSON(http.StatusOK, gin.H{
			"data":    paged,
			"total":   total,
			"page":    page,
			"limit":   limit,
			"message": "ok",
		})
	}
}

type MarkReadInput struct {
	FeedItemIDs []uuid.UUID `json:"feed_item_ids" binding:"required"`
}

func MarkItemsRead(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var input MarkReadInput
		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		userIDVal, _ := c.Get("user_id")
		userID := userIDVal.(uuid.UUID)
		now := time.Now()

		for _, itemID := range input.FeedItemIDs {
			read := model.FeedItemRead{
				UserID:     userID,
				FeedItemID: itemID,
				ReadAt:     now,
			}
			db.Where("user_id = ? AND feed_item_id = ?", userID, itemID).
				FirstOrCreate(&read)
		}

		c.JSON(http.StatusOK, gin.H{"message": "ok"})
	}
}

func MarkAllRead(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userIDVal, _ := c.Get("user_id")
		userID := userIDVal.(uuid.UUID)

		var userSubscriptions []model.Subscription
		db.Where("user_id = ?", userID).Preload("FeedSource").Find(&userSubscriptions)

		var feedSourceIDs []uuid.UUID
		for _, sub := range userSubscriptions {
			if sub.FeedSource != nil && sub.FeedSource.SourceType == "external_rss" {
				feedSourceIDs = append(feedSourceIDs, sub.FeedSource.ID)
			}
		}

		if len(feedSourceIDs) == 0 {
			c.JSON(http.StatusOK, gin.H{"message": "ok"})
			return
		}

		var unreadItems []model.FeedItem
		db.Where("feed_source_id IN ?", feedSourceIDs).Find(&unreadItems)

		now := time.Now()
		for _, item := range unreadItems {
			read := model.FeedItemRead{
				UserID:     userID,
				FeedItemID: item.ID,
				ReadAt:     now,
			}
			db.Where("user_id = ? AND feed_item_id = ?", userID, item.ID).
				FirstOrCreate(&read)
		}

		c.JSON(http.StatusOK, gin.H{"message": "ok"})
	}
}

func GetSubscriptionGroups(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userIDVal, _ := c.Get("user_id")
		userID := userIDVal.(uuid.UUID)

		if _, err := getOrCreateDefaultSubscriptionGroup(db, userID); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to prepare default group"})
			return
		}

		var groups []model.SubscriptionGroup
		if err := db.Where("user_id = ?", userID).Find(&groups).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch groups"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"data": groups, "message": "ok"})
	}
}

type GroupInput struct {
	Name string `json:"name" binding:"required"`
}

func CreateSubscriptionGroup(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var input GroupInput
		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		name := strings.TrimSpace(input.Name)
		if name == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Group name is required"})
			return
		}

		if name == defaultSubscriptionGroupName {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Default group already exists"})
			return
		}

		userIDVal, _ := c.Get("user_id")
		userID := userIDVal.(uuid.UUID)

		var existing model.SubscriptionGroup
		if err := db.Where("user_id = ? AND name = ?", userID, name).First(&existing).Error; err == nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Group name already exists"})
			return
		} else if err != gorm.ErrRecordNotFound {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to validate group name"})
			return
		}

		group := model.SubscriptionGroup{
			UserID: userID,
			Name:   name,
		}

		if err := db.Create(&group).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create group"})
			return
		}

		c.JSON(http.StatusCreated, gin.H{"data": group, "message": "ok"})
	}
}

func UpdateSubscriptionGroup(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		userIDVal, _ := c.Get("user_id")
		userID := userIDVal.(uuid.UUID)

		var input GroupInput
		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		name := strings.TrimSpace(input.Name)
		if name == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Group name is required"})
			return
		}

		var target model.SubscriptionGroup
		if err := db.Where("id = ? AND user_id = ?", id, userID).First(&target).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Group not found"})
			return
		}

		if target.Name == defaultSubscriptionGroupName {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Default group cannot be renamed"})
			return
		}

		if name == defaultSubscriptionGroupName {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Default group name is reserved"})
			return
		}

		var existing model.SubscriptionGroup
		if err := db.Where("user_id = ? AND name = ? AND id <> ?", userID, name, id).First(&existing).Error; err == nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Group name already exists"})
			return
		} else if err != gorm.ErrRecordNotFound {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to validate group name"})
			return
		}

		if err := db.Model(&model.SubscriptionGroup{}).Where("id = ? AND user_id = ?", id, userID).Update("name", name).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update group"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "ok"})
	}
}

func DeleteSubscriptionGroup(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		userIDVal, _ := c.Get("user_id")
		userID := userIDVal.(uuid.UUID)

		var targetGroup model.SubscriptionGroup
		if err := db.Where("id = ? AND user_id = ?", id, userID).First(&targetGroup).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Group not found"})
			return
		}

		if targetGroup.Name == defaultSubscriptionGroupName {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Default group cannot be deleted"})
			return
		}

		defaultGroup, err := getOrCreateDefaultSubscriptionGroup(db, userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to prepare default group"})
			return
		}

		db.Model(&model.Subscription{}).
			Where("subscription_group_id = ? AND user_id = ?", id, userID).
			Update("subscription_group_id", defaultGroup.ID)

		if err := db.Where("id = ? AND user_id = ?", id, userID).Delete(&model.SubscriptionGroup{}).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete group"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "ok"})
	}
}

type SetGroupInput struct {
	GroupID *uuid.UUID `json:"group_id"`
}

func SetSubscriptionGroup(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		subID := c.Param("id")
		userIDVal, _ := c.Get("user_id")
		userID := userIDVal.(uuid.UUID)

		var input SetGroupInput
		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		var targetGroupID *uuid.UUID
		if input.GroupID == nil {
			defaultGroup, err := getOrCreateDefaultSubscriptionGroup(db, userID)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to prepare default group"})
				return
			}
			targetGroupID = &defaultGroup.ID
		} else {
			targetGroupID = input.GroupID
		}

		if err := db.Model(&model.Subscription{}).
			Where("id = ? AND user_id = ?", subID, userID).
			Update("subscription_group_id", targetGroupID).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update subscription group"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "ok"})
	}
}

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
