package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"

	"atoman/internal/middleware"
	"atoman/internal/model"
	"atoman/internal/service"
)

// SetupRevisionRoutes registers revision-related routes
func SetupRevisionRoutes(router *gin.Engine, db *gorm.DB) {
	revisionService := service.NewRevisionService(db)

	// Album revisions
	albums := router.Group("/api/albums/:id")
	{
		albums.GET("/revisions", GetAlbumRevisionsHandler(revisionService))
		albums.GET("/revisions/:version", GetAlbumRevisionHandler(revisionService))
		albums.GET("/revisions/diff", GetAlbumRevisionDiffHandler(revisionService))
		albums.POST("/edit", middleware.AuthMiddleware(), CreateAlbumRevisionHandler(db, revisionService))
		albums.POST("/revert/:version", middleware.AuthMiddleware(), RevertAlbumHandler(revisionService))
	}

	// Song revisions
	songs := router.Group("/api/songs/:id")
	{
		songs.GET("/revisions", GetSongRevisionsHandler(revisionService))
		songs.GET("/revisions/:version", GetSongRevisionHandler(revisionService))
		songs.GET("/revisions/diff", GetSongRevisionDiffHandler(revisionService))
		songs.POST("/edit", middleware.AuthMiddleware(), CreateSongRevisionHandler(db, revisionService))
		songs.POST("/revert/:version", middleware.AuthMiddleware(), RevertSongHandler(revisionService))
	}

	// Admin approval endpoints
	admin := router.Group("/api/admin/revisions")
	admin.Use(middleware.AuthMiddleware(), middleware.AdminMiddleware(db))
	{
		admin.POST("/:id/approve", ApproveRevisionHandler(revisionService))
		admin.POST("/:id/reject", RejectRevisionHandler(revisionService))
	}
}

type CreateRevisionInput struct {
	BaseRevision int                    `json:"base_revision" binding:"required"`
	Changes      map[string]interface{} `json:"changes" binding:"required"`
	EditSummary  string                 `json:"edit_summary" binding:"required"`
}

// GetAlbumRevisionsHandler returns revision history for an album
func GetAlbumRevisionsHandler(revisionService *service.RevisionService) gin.HandlerFunc {
	return func(c *gin.Context) {
		albumID, err := uuid.Parse(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid album ID"})
			return
		}

		limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
		offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

		revisions, total, err := revisionService.GetRevisions("album", albumID, limit, offset)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch revisions"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"data":  revisions,
			"total": total,
			"limit": limit,
			"offset": offset,
		})
	}
}

// GetAlbumRevisionHandler returns a specific revision
func GetAlbumRevisionHandler(revisionService *service.RevisionService) gin.HandlerFunc {
	return func(c *gin.Context) {
		albumID, err := uuid.Parse(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid album ID"})
			return
		}

		version, err := strconv.Atoi(c.Param("version"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid version number"})
			return
		}

		var revision model.Revision
		if err := revisionService.GetDB().
			Where("content_id = ? AND content_type = ? AND version_number = ?", albumID, "album", version).
			Preload("Editor").
			Preload("Reviewer").
			First(&revision).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Revision not found"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"data": revision})
	}
}

// GetAlbumRevisionDiffHandler compares two versions
func GetAlbumRevisionDiffHandler(revisionService *service.RevisionService) gin.HandlerFunc {
	return func(c *gin.Context) {
		albumID, err := uuid.Parse(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid album ID"})
			return
		}

		v1, err := strconv.Atoi(c.Query("v1"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid v1 parameter"})
			return
		}

		v2, err := strconv.Atoi(c.Query("v2"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid v2 parameter"})
			return
		}

		diff, err := revisionService.GetRevisionDiff("album", albumID, v1, v2)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"data": diff})
	}
}

// CreateAlbumRevisionHandler creates a new album revision
func CreateAlbumRevisionHandler(db *gorm.DB, revisionService *service.RevisionService) gin.HandlerFunc {
	return func(c *gin.Context) {
		albumID, err := uuid.Parse(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid album ID"})
			return
		}

		var input CreateRevisionInput
		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Get user info
		userID := c.GetString("user_id")
		editorUUID, _ := uuid.Parse(userID)

		// Check if user is admin for auto-approval
		userRole := c.GetString("role")
		autoApprove := (userRole == "admin")

		// Check protection level
		var protection model.ContentProtection
		protectionLevel := "none"
		if err := db.Where("content_id = ? AND content_type = ?", albumID, "album").
			First(&protection).Error; err == nil {
			protectionLevel = protection.ProtectionLevel
		}

		// Apply protection rules
		if protectionLevel == "full" && userRole != "admin" {
			c.JSON(http.StatusForbidden, gin.H{"error": "This album is fully protected. Only admins can edit."})
			return
		}

		if protectionLevel == "semi" {
			autoApprove = false // Force approval for semi-protected content
		}

		// Create revision
		revision, conflicts, err := revisionService.CreateRevision(
			"album",
			albumID,
			editorUUID,
			input.Changes,
			input.EditSummary,
			input.BaseRevision,
			autoApprove,
		)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// If conflicts exist, return them
		if len(conflicts) > 0 {
			c.JSON(http.StatusConflict, gin.H{
				"error":     "Edit conflicts detected",
				"conflicts": conflicts,
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"data":    revision,
			"message": statusMessage(autoApprove),
		})
	}
}

// RevertAlbumHandler reverts album to a previous version
func RevertAlbumHandler(revisionService *service.RevisionService) gin.HandlerFunc {
	return func(c *gin.Context) {
		albumID, err := uuid.Parse(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid album ID"})
			return
		}

		version, err := strconv.Atoi(c.Param("version"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid version number"})
			return
		}

		var input struct {
			EditSummary string `json:"edit_summary"`
		}
		c.ShouldBindJSON(&input)

		userID := c.GetString("user_id")
		editorUUID, _ := uuid.Parse(userID)

		revision, err := revisionService.RevertToRevision(
			"album",
			albumID,
			version,
			editorUUID,
			input.EditSummary,
		)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"data":    revision,
			"message": "Album reverted successfully",
		})
	}
}

// Song handlers (similar structure to album handlers)
func GetSongRevisionsHandler(revisionService *service.RevisionService) gin.HandlerFunc {
	return func(c *gin.Context) {
		songID, err := uuid.Parse(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid song ID"})
			return
		}

		limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
		offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

		revisions, total, err := revisionService.GetRevisions("song", songID, limit, offset)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch revisions"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"data":  revisions,
			"total": total,
		})
	}
}

func GetSongRevisionHandler(revisionService *service.RevisionService) gin.HandlerFunc {
	return func(c *gin.Context) {
		songID, err := uuid.Parse(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid song ID"})
			return
		}

		version, err := strconv.Atoi(c.Param("version"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid version number"})
			return
		}

		var revision model.Revision
		if err := revisionService.GetDB().
			Where("content_id = ? AND content_type = ? AND version_number = ?", songID, "song", version).
			Preload("Editor").
			Preload("Reviewer").
			First(&revision).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Revision not found"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"data": revision})
	}
}

func GetSongRevisionDiffHandler(revisionService *service.RevisionService) gin.HandlerFunc {
	return func(c *gin.Context) {
		songID, err := uuid.Parse(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid song ID"})
			return
		}

		v1, _ := strconv.Atoi(c.Query("v1"))
		v2, _ := strconv.Atoi(c.Query("v2"))

		diff, err := revisionService.GetRevisionDiff("song", songID, v1, v2)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"data": diff})
	}
}

func CreateSongRevisionHandler(db *gorm.DB, revisionService *service.RevisionService) gin.HandlerFunc {
	return func(c *gin.Context) {
		songID, err := uuid.Parse(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid song ID"})
			return
		}

		var input CreateRevisionInput
		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		userID := c.GetString("user_id")
		editorUUID, _ := uuid.Parse(userID)
		userRole := c.GetString("role")
		autoApprove := (userRole == "admin")

		var protection model.ContentProtection
		if err := db.Where("content_id = ? AND content_type = ?", songID, "song").
			First(&protection).Error; err == nil {
			if protection.ProtectionLevel == "full" && userRole != "admin" {
				c.JSON(http.StatusForbidden, gin.H{"error": "This song is fully protected"})
				return
			}
			if protection.ProtectionLevel == "semi" {
				autoApprove = false
			}
		}

		revision, conflicts, err := revisionService.CreateRevision(
			"song",
			songID,
			editorUUID,
			input.Changes,
			input.EditSummary,
			input.BaseRevision,
			autoApprove,
		)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		if len(conflicts) > 0 {
			c.JSON(http.StatusConflict, gin.H{
				"error":     "Edit conflicts detected",
				"conflicts": conflicts,
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"data":    revision,
			"message": statusMessage(autoApprove),
		})
	}
}

func RevertSongHandler(revisionService *service.RevisionService) gin.HandlerFunc {
	return func(c *gin.Context) {
		songID, err := uuid.Parse(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid song ID"})
			return
		}

		version, _ := strconv.Atoi(c.Param("version"))

		var input struct {
			EditSummary string `json:"edit_summary"`
		}
		c.ShouldBindJSON(&input)

		userID := c.GetString("user_id")
		editorUUID, _ := uuid.Parse(userID)

		revision, err := revisionService.RevertToRevision(
			"song",
			songID,
			version,
			editorUUID,
			input.EditSummary,
		)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"data":    revision,
			"message": "Song reverted successfully",
		})
	}
}

// ApproveRevisionHandler approves a pending revision
func ApproveRevisionHandler(revisionService *service.RevisionService) gin.HandlerFunc {
	return func(c *gin.Context) {
		revisionID, err := uuid.Parse(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid revision ID"})
			return
		}

		var input struct {
			ReviewNotes string `json:"review_notes"`
		}
		c.ShouldBindJSON(&input)

		reviewerID := c.GetString("user_id")
		reviewerUUID, _ := uuid.Parse(reviewerID)

		if err := revisionService.ApproveRevision(revisionID, reviewerUUID, input.ReviewNotes); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Revision approved successfully"})
	}
}

// RejectRevisionHandler rejects a pending revision
func RejectRevisionHandler(revisionService *service.RevisionService) gin.HandlerFunc {
	return func(c *gin.Context) {
		revisionID, err := uuid.Parse(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid revision ID"})
			return
		}

		var input struct {
			ReviewNotes string `json:"review_notes" binding:"required"`
		}
		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Review notes required"})
			return
		}

		reviewerID := c.GetString("user_id")
		reviewerUUID, _ := uuid.Parse(reviewerID)

		if err := revisionService.RejectRevision(revisionID, reviewerUUID, input.ReviewNotes); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Revision rejected"})
	}
}

func statusMessage(autoApprove bool) string {
	if autoApprove {
		return "Changes saved and approved automatically"
	}
	return "Changes saved and pending approval"
}
