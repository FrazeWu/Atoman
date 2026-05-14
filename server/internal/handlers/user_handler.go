package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"

	"atoman/internal/middleware"
	"atoman/internal/model"
)

// SetupUserRoutes configures user-related routes
func SetupUserRoutes(router *gin.Engine, db *gorm.DB) {
	// Blog explore route
	router.GET("/api/blog/explore", ExplorePosts(db))

	users := router.Group("/api/users")
	{
		// Public routes — lookup by username (must come before /:id routes)
		users.GET("/search", SearchUsers(db))
		users.GET("/by-username/:username", GetUserByUsername(db))
		users.GET("/:id/profile", GetUserProfile(db))
		users.GET("/:id/followers", GetUserFollowers(db))
		users.GET("/:id/following", GetUserFollowing(db))

		// Protected routes
		protected := users.Group("")
		protected.Use(middleware.AuthMiddleware())
		{
			protected.GET("/me", GetCurrentUser(db))
			protected.PUT("/me", UpdateUserProfile(db))
			protected.GET("/me/settings", GetUserSettings(db))
			protected.PUT("/me/settings", UpdateUserSettings(db))

			protected.POST("/:id/follow", FollowUser(db))
			protected.DELETE("/:id/follow", UnfollowUser(db))
		}
	}
}

// UserProfileInput represents the request body for updating user profile
type UserProfileInput struct {
	DisplayName string `json:"display_name"`
	AvatarURL   string `json:"avatar_url"`
	Bio         string `json:"bio"`
	Website     string `json:"website"`
	Location    string `json:"location"`
}

// UserSettingsInput represents the request body for updating user settings
type UserSettingsInput struct {
	PrivateProfile *bool `json:"private_profile"`
}

// ExplorePostResponse represents a post in the explore feed
type ExplorePostResponse struct {
	model.Post
	LikesCount    int64 `json:"likes_count"`
	CommentsCount int64 `json:"comments_count"`
}

// ExplorePosts returns a paginated list of published posts with interaction counts
func ExplorePosts(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
		limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
		offset := (page - 1) * limit

		var posts []model.Post
		if err := db.Preload("User").
			Where("status = ?", "published").
			Order("created_at DESC").
			Limit(limit).
			Offset(offset).
			Find(&posts).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch explore posts"})
			return
		}

		// Get counts for each post
		var response []ExplorePostResponse
		for _, post := range posts {
			var likesCount int64
			var commentsCount int64

			db.Model(&model.Like{}).Where("target_type = ? AND target_id = ?", "post", post.ID).Count(&likesCount)
			db.Model(&model.Comment{}).Where("post_id = ? AND status = ?", post.ID, "visible").Count(&commentsCount)

			response = append(response, ExplorePostResponse{
				Post:          post,
				LikesCount:    likesCount,
				CommentsCount: commentsCount,
			})
		}

		c.JSON(http.StatusOK, gin.H{"data": response, "message": "ok"})
	}
}

// GetCurrentUser returns the authenticated user's own full profile
func GetCurrentUser(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userIDVal, _ := c.Get("user_id")
		userID := userIDVal.(uuid.UUID)

		var user model.User
		if err := db.Where("uuid = ?", userID).First(&user).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"data": user, "message": "ok"})
	}
}

// GetUserByUsername looks up a user by their username (public)
func GetUserByUsername(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		username := c.Param("username")
		var user model.User

		if err := db.Where("username = ?", username).First(&user).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
			return
		}

		var followersCount, followingCount, postsCount int64
		db.Model(&model.Follow{}).Where("following_id = ?", user.UUID).Count(&followersCount)
		db.Model(&model.Follow{}).Where("follower_id = ?", user.UUID).Count(&followingCount)
		db.Model(&model.Post{}).Where("user_id = ? AND status = ?", user.UUID, "published").Count(&postsCount)

		c.JSON(http.StatusOK, gin.H{
			"data": gin.H{
				"id":              user.ID,
				"uuid":            user.UUID,
				"username":        user.Username,
				"display_name":    user.DisplayName,
				"avatar_url":      user.AvatarURL,
				"bio":             user.Bio,
				"website":         user.Website,
				"role":            user.Role,
				"created_at":      user.CreatedAt,
				"followers_count": followersCount,
				"following_count": followingCount,
				"posts_count":     postsCount,
			},
			"message": "ok",
		})
	}
}

// GetUserProfile returns public profile information for a user
func GetUserProfile(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		var user model.User

		if err := db.Where("uuid = ? OR username = ?", id, id).First(&user).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
			return
		}

		// Get counts
		var followersCount int64
		var followingCount int64
		var postsCount int64

		db.Model(&model.Follow{}).Where("following_id = ?", user.UUID).Count(&followersCount)
		db.Model(&model.Follow{}).Where("follower_id = ?", user.UUID).Count(&followingCount)
		db.Model(&model.Post{}).Where("user_id = ? AND status = ?", user.UUID, "published").Count(&postsCount)

		// Get user's channels
		var channels []model.Channel
		db.Where("user_id = ?", user.UUID).Find(&channels)

		c.JSON(http.StatusOK, gin.H{
			"data": gin.H{
				"user": gin.H{
					"id":           user.ID,
					"uuid":         user.UUID,
					"username":     user.Username,
					"display_name": user.DisplayName,
					"avatar_url":   user.AvatarURL,
					"bio":          user.Bio,
					"website":      user.Website,
					"location":     user.Location,
					"created_at":   user.CreatedAt,
				},
				"stats": gin.H{
					"followers_count": followersCount,
					"following_count": followingCount,
					"posts_count":     postsCount,
				},
				"channels": channels,
			},
			"message": "ok",
		})
	}
}

// UpdateUserProfile updates the authenticated user's profile
func UpdateUserProfile(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var input UserProfileInput
		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		userIDVal, _ := c.Get("user_id")
		userID := userIDVal.(uuid.UUID)

		var user model.User
		if err := db.Where("uuid = ?", userID).First(&user).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
			return
		}

		if err := db.Model(&user).Updates(model.User{
			DisplayName: input.DisplayName,
			AvatarURL:   input.AvatarURL,
			Bio:         input.Bio,
			Website:     input.Website,
			Location:    input.Location,
		}).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update profile"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"data": user, "message": "ok"})
	}
}

// GetUserSettings returns the authenticated user's settings
func GetUserSettings(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userIDVal, _ := c.Get("user_id")
		userID := userIDVal.(uuid.UUID)

		var settings model.UserSettings
		if err := db.Where("user_id = ?", userID).First(&settings).Error; err != nil {
			// If settings don't exist, create default
			settings = model.UserSettings{UserID: userID}
			db.Create(&settings)
		}

		c.JSON(http.StatusOK, gin.H{"data": settings, "message": "ok"})
	}
}

// UpdateUserSettings updates the authenticated user's settings
func UpdateUserSettings(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var input UserSettingsInput
		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		userIDVal, _ := c.Get("user_id")
		userID := userIDVal.(uuid.UUID)

		var settings model.UserSettings
		if err := db.Where("user_id = ?", userID).First(&settings).Error; err != nil {
			settings = model.UserSettings{UserID: userID}
			db.Create(&settings)
		}

		updates := map[string]interface{}{}
		if input.PrivateProfile != nil {
			updates["private_profile"] = *input.PrivateProfile
		}

		if len(updates) > 0 {
			if err := db.Model(&settings).Updates(updates).Error; err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update settings"})
				return
			}
		}

		c.JSON(http.StatusOK, gin.H{"data": settings, "message": "ok"})
	}
}

// FollowUser creates a follow relationship
func FollowUser(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		targetIDStr := c.Param("id")
		targetID, err := uuid.Parse(targetIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user UUID"})
			return
		}

		userIDVal, _ := c.Get("user_id")
		userID := userIDVal.(uuid.UUID)

		if userID == targetID {
			c.JSON(http.StatusBadRequest, gin.H{"error": "You cannot follow yourself"})
			return
		}

		// Check if target user exists
		var targetUser model.User
		if err := db.Where("uuid = ?", targetID).First(&targetUser).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
			return
		}

		follow := model.Follow{
			FollowerID:  userID,
			FollowingID: targetID,
		}

		if err := db.Where(model.Follow{FollowerID: userID, FollowingID: targetID}).FirstOrCreate(&follow).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to follow user"})
			return
		}


		c.JSON(http.StatusOK, gin.H{"message": "ok"})
	}
}

// UnfollowUser removes a follow relationship
func UnfollowUser(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		targetID := c.Param("id")
		userIDVal, _ := c.Get("user_id")
		userID := userIDVal.(uuid.UUID)

		if err := db.Where("follower_id = ? AND following_id = ?", userID, targetID).Delete(&model.Follow{}).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to unfollow user"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "ok"})
	}
}

// GetUserFollowers returns a list of users following the specified user
func GetUserFollowers(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		var follows []model.Follow

		if err := db.Where("following_id = ?", id).Find(&follows).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch followers"})
			return
		}

		// Get user details for followers
		var followerIDs []uuid.UUID
		for _, f := range follows {
			followerIDs = append(followerIDs, f.FollowerID)
		}

		var users []model.User
		if len(followerIDs) > 0 {
			db.Where("uuid IN ?", followerIDs).Find(&users)
		}

		c.JSON(http.StatusOK, gin.H{"data": users, "message": "ok"})
	}
}

// GetUserFollowing returns a list of users the specified user is following
func GetUserFollowing(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		var follows []model.Follow

		if err := db.Where("follower_id = ?", id).Find(&follows).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch following"})
			return
		}

		// Get user details for following
		var followingIDs []uuid.UUID
		for _, f := range follows {
			followingIDs = append(followingIDs, f.FollowingID)
		}

		var users []model.User
		if len(followingIDs) > 0 {
			db.Where("uuid IN ?", followingIDs).Find(&users)
		}

		c.JSON(http.StatusOK, gin.H{"data": users, "message": "ok"})
	}
}

// SearchUsers returns users matching the query string (for @mention autocomplete)
// GET /api/users/search?q=<query>&limit=<n>
func SearchUsers(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		q := c.Query("q")
		limit := 5
		if l, err := strconv.Atoi(c.Query("limit")); err == nil && l > 0 && l <= 20 {
			limit = l
		}

		type UserResult struct {
			UUID        string `json:"uuid"`
			Username    string `json:"username"`
			DisplayName string `json:"display_name"`
			AvatarURL   string `json:"avatar_url"`
		}

		var results []UserResult
		query := db.Model(&model.User{}).
			Select("uuid, username, display_name, avatar_url").
			Where("is_active = ?", true).
			Limit(limit)

		if q != "" {
			like := "%" + q + "%"
			query = query.Where("username ILIKE ? OR display_name ILIKE ?", like, like)
		}

		if err := query.Scan(&results).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "search failed"})
			return
		}

		if results == nil {
			results = []UserResult{}
		}
		c.JSON(http.StatusOK, gin.H{"data": results})
	}
}
