package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"

	"atoman/internal/middleware"
	"atoman/internal/model"
)

// SetupProtectionRoutes registers content protection routes
func SetupProtectionRoutes(router *gin.Engine, db *gorm.DB) {
	protected := router.Group("/api")
	{
		// Album protection
		albums := protected.Group("/albums/:id")
		{
			albums.GET("/protection", GetAlbumProtectionHandler(db))
			albums.POST("/protect", middleware.AuthMiddleware(), middleware.AdminMiddleware(db), SetAlbumProtectionHandler(db))
			albums.DELETE("/protect", middleware.AuthMiddleware(), middleware.AdminMiddleware(db), RemoveAlbumProtectionHandler(db))
		}

		// Song protection
		songs := protected.Group("/songs/:id")
		{
			songs.GET("/protection", GetSongProtectionHandler(db))
			songs.POST("/protect", middleware.AuthMiddleware(), middleware.AdminMiddleware(db), SetSongProtectionHandler(db))
			songs.DELETE("/protect", middleware.AuthMiddleware(), middleware.AdminMiddleware(db), RemoveSongProtectionHandler(db))
		}

		// Artist protection
		artists := protected.Group("/api/artists/:id")
		{
			artists.GET("/protection", GetArtistProtectionHandler(db))
			artists.POST("/protect", middleware.AuthMiddleware(), middleware.AdminMiddleware(db), SetArtistProtectionHandler(db))
			artists.DELETE("/protect", middleware.AuthMiddleware(), middleware.AdminMiddleware(db), RemoveArtistProtectionHandler(db))
		}
	}
}

type SetProtectionInput struct {
	ProtectionLevel string  `json:"protection_level" binding:"required,oneof=none semi full"`
	Reason          string  `json:"reason"`
	ExpiresAt       *string `json:"expires_at"` // ISO 8601 format
}

// GetAlbumProtectionHandler returns protection status for an album
func GetAlbumProtectionHandler(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		albumID, err := uuid.Parse(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid album ID"})
			return
		}

		var protection model.ContentProtection
		if err := db.Where("content_type = ? AND content_id = ?", "album", albumID).
			Preload("ProtectedUser").
			First(&protection).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				// Default: no protection
				c.JSON(http.StatusOK, gin.H{
					"data": gin.H{
						"protection_level": "none",
						"reason":           "",
						"protected_by":     nil,
						"expires_at":       nil,
					},
				})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch protection"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"data": gin.H{
				"protection_level": protection.ProtectionLevel,
				"reason":           protection.Reason,
				"protected_by":     protection.ProtectedBy,
				"protected_user":   protection.ProtectedUser,
				"expires_at":       protection.ExpiresAt,
				"created_at":       protection.CreatedAt,
			},
		})
	}
}

// GetSongProtectionHandler returns protection status for a song
func GetSongProtectionHandler(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		songID, err := uuid.Parse(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid song ID"})
			return
		}

		var protection model.ContentProtection
		if err := db.Where("content_type = ? AND content_id = ?", "song", songID).
			Preload("ProtectedUser").
			First(&protection).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				c.JSON(http.StatusOK, gin.H{
					"data": gin.H{
						"protection_level": "none",
						"reason":           "",
					},
				})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch protection"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"data": protection})
	}
}

// GetArtistProtectionHandler returns protection status for an artist
func GetArtistProtectionHandler(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		artistID, err := uuid.Parse(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid artist ID"})
			return
		}

		var protection model.ContentProtection
		if err := db.Where("content_type = ? AND content_id = ?", "artist", artistID).
			Preload("ProtectedUser").
			First(&protection).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				c.JSON(http.StatusOK, gin.H{
					"data": gin.H{
						"protection_level": "none",
					},
				})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch protection"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"data": protection})
	}
}

// SetAlbumProtectionHandler sets protection for an album
func SetAlbumProtectionHandler(db *gorm.DB) gin.HandlerFunc {
	return setProtectionHandler(db, "album")
}

// SetSongProtectionHandler sets protection for a song
func SetSongProtectionHandler(db *gorm.DB) gin.HandlerFunc {
	return setProtectionHandler(db, "song")
}

// SetArtistProtectionHandler sets protection for an artist
func SetArtistProtectionHandler(db *gorm.DB) gin.HandlerFunc {
	return setProtectionHandler(db, "artist")
}

func setProtectionHandler(db *gorm.DB, contentType string) gin.HandlerFunc {
	return func(c *gin.Context) {
		contentID, err := uuid.Parse(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
			return
		}

		var input SetProtectionInput
		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input: " + err.Error()})
			return
		}

		adminID := c.GetString("user_id")
		adminUUID, _ := uuid.Parse(adminID)

		// Parse expiration if provided
		var expiresAt *time.Time
		if input.ExpiresAt != nil {
			parsed, err := time.Parse(time.RFC3339, *input.ExpiresAt)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid expires_at format, use ISO 8601"})
				return
			}
			expiresAt = &parsed
		}

		// Check if protection already exists
		var protection model.ContentProtection
		err = db.Where("content_type = ? AND content_id = ?", contentType, contentID).
			First(&protection).Error

		if err == gorm.ErrRecordNotFound {
			// Create new protection
			protection = model.ContentProtection{
				ContentType:     contentType,
				ContentID:       contentID,
				ProtectionLevel: input.ProtectionLevel,
				ProtectedBy:     adminUUID,
				Reason:          input.Reason,
				ExpiresAt:       expiresAt,
			}
			if err := db.Create(&protection).Error; err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to set protection"})
				return
			}
		} else if err == nil {
			// Update existing protection
			updates := map[string]interface{}{
				"protection_level": input.ProtectionLevel,
				"protected_by":     adminUUID,
				"reason":           input.Reason,
				"expires_at":       expiresAt,
			}
			if err := db.Model(&protection).Updates(updates).Error; err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update protection"})
				return
			}
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
			return
		}

		db.Preload("ProtectedUser").First(&protection, protection.ID)

		c.JSON(http.StatusOK, gin.H{
			"message": "Protection set successfully",
			"data":    protection,
		})
	}
}

// RemoveAlbumProtectionHandler removes protection from an album
func RemoveAlbumProtectionHandler(db *gorm.DB) gin.HandlerFunc {
	return removeProtectionHandler(db, "album")
}

// RemoveSongProtectionHandler removes protection from a song
func RemoveSongProtectionHandler(db *gorm.DB) gin.HandlerFunc {
	return removeProtectionHandler(db, "song")
}

// RemoveArtistProtectionHandler removes protection from an artist
func RemoveArtistProtectionHandler(db *gorm.DB) gin.HandlerFunc {
	return removeProtectionHandler(db, "artist")
}

func removeProtectionHandler(db *gorm.DB, contentType string) gin.HandlerFunc {
	return func(c *gin.Context) {
		contentID, err := uuid.Parse(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
			return
		}

		result := db.Where("content_type = ? AND content_id = ?", contentType, contentID).
			Delete(&model.ContentProtection{})

		if result.Error != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to remove protection"})
			return
		}

		if result.RowsAffected == 0 {
			c.JSON(http.StatusNotFound, gin.H{"error": "No protection found"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Protection removed"})
	}
}

// GetProtectionStatus is a helper to get protection level for any content
// Used by revision handlers to check permissions
func GetProtectionStatus(db *gorm.DB, contentType string, contentID uuid.UUID) string {
	var protection model.ContentProtection
	err := db.Where("content_type = ? AND content_id = ?", contentType, contentID).
		First(&protection).Error

	if err != nil {
		return "none" // No protection = default
	}

	// Check if protection has expired
	if protection.ExpiresAt != nil && protection.ExpiresAt.Before(time.Now()) {
		return "none" // Expired
	}

	return protection.ProtectionLevel
}

// GetPendingRevisionsCount returns count of pending revisions for admin dashboard
func GetPendingRevisionsCount(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var count int64
		if err := db.Model(&model.Revision{}).Where("status = ?", "pending").Count(&count).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to count revisions"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"pending_revisions": count})
	}
}

// GetPendingDiscussionsCount returns count of discussions
func GetPendingDiscussionsCount(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var total, active, resolved int64

		db.Model(&model.Discussion{}).Count(&total)
		db.Model(&model.Discussion{}).Where("status = ?", "active").Count(&active)
		db.Model(&model.Discussion{}).Where("status = ?", "resolved").Count(&resolved)

		c.JSON(http.StatusOK, gin.H{
			"total":    total,
			"active":   active,
			"resolved": resolved,
		})
	}
}

// SetupStatusRoutes registers status management routes
func SetupStatusRoutes(router *gin.Engine, db *gorm.DB) {
	status := router.Group("/api")
	{
		// Album status
		albums := status.Group("/albums/:id")
		{
			albums.POST("/status", middleware.AuthMiddleware(), middleware.AdminMiddleware(db), UpdateAlbumStatusHandler(db))
		}

		// Song status
		songs := status.Group("/songs/:id")
		{
			songs.POST("/status", middleware.AuthMiddleware(), middleware.AdminMiddleware(db), UpdateSongStatusHandler(db))
		}
	}
}

type UpdateStatusInput struct {
	Status string `json:"status" binding:"required,oneof=draft pending verified"`
}

// UpdateAlbumStatusHandler updates album status (admin only)
func UpdateAlbumStatusHandler(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		albumID, err := uuid.Parse(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid album ID"})
			return
		}

		var input UpdateStatusInput
		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid status"})
			return
		}

		var album model.Album
		if err := db.First(&album, "id = ?", albumID).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Album not found"})
			return
		}

		if err := db.Model(&album).Update("status", input.Status).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update status"})
			return
		}

		// Create a status change revision
		adminID := c.GetString("user_id")
		adminUUID, _ := uuid.Parse(adminID)

		// Get current revision to form chain
		var currentRev model.Revision
		db.Where("content_type = ? AND content_id = ? AND is_current = ?", "album", albumID, true).
			Order("version_number DESC").First(&currentRev)

		statusChange := map[string]interface{}{
			"status": input.Status,
		}
		snapshot, _ := json.Marshal(statusChange)

		newRev := model.Revision{
			ContentType:     "album",
			ContentID:       albumID,
			VersionNumber:   currentRev.VersionNumber + 1,
			ContentSnapshot: snapshot,
			EditorID:        adminUUID,
			EditSummary:     "Status changed to: " + input.Status,
			EditType:        "edit",
			Status:          "approved",
			IsCurrent:       true,
		}

		if currentRev.ID != uuid.Nil {
			newRev.PreviousRevisionID = &currentRev.ID
			db.Model(&currentRev).Update("is_current", false)
		}

		db.Create(&newRev)

		c.JSON(http.StatusOK, gin.H{
			"message": "Status updated",
			"status":  input.Status,
		})
	}
}

// UpdateSongStatusHandler updates song status (admin only)
func UpdateSongStatusHandler(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		songID, err := uuid.Parse(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid song ID"})
			return
		}

		var input UpdateStatusInput
		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid status"})
			return
		}

		var song model.Song
		if err := db.First(&song, "id = ?", songID).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Song not found"})
			return
		}

		if err := db.Model(&song).Update("status", input.Status).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update status"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "Status updated",
			"status":  input.Status,
		})
	}
}