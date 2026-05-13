package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"

	"atoman/internal/middleware"
	"atoman/internal/model"
	"atoman/internal/service"
)

func SetupArtistWikiRoutes(router *gin.Engine, db *gorm.DB) {
	revisionService := service.NewRevisionService(db)

	artists := router.Group("/api/artists")
	{
		artists.GET("/:id", GetArtistByIDHandler(db))
		artists.PUT("/:id", middleware.AuthMiddleware(), UpdateArtistHandler(db, revisionService))
		artists.GET("/:id/revisions", GetArtistRevisionsHandler(revisionService))
		artists.GET("/:id/revisions/:version", GetArtistRevisionHandler(revisionService))
		artists.POST("/:id/edit", middleware.AuthMiddleware(), CreateArtistRevisionHandler(db, revisionService))
		artists.POST("/:id/revert/:version", middleware.AuthMiddleware(), RevertArtistHandler(revisionService))
		artists.GET("/:id/aliases", GetArtistAliasesHandler(db))
		artists.POST("/:id/aliases", middleware.AuthMiddleware(), AddArtistAliasHandler(db))
		artists.DELETE("/:id/aliases/:aliasId", middleware.AuthMiddleware(), DeleteArtistAliasHandler(db))
	}

	admin := router.Group("/api/admin/artists")
	admin.Use(middleware.AuthMiddleware(), middleware.AdminMiddleware(db))
	{
		admin.POST("/:id/merge", MergeArtistsHandler(db))
	}
}

func GetArtistByIDHandler(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		artistID, err := uuid.Parse(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid artist ID"})
			return
		}

		var artist model.Artist
		if err := db.Preload("Aliases").Preload("Albums").Preload("Albums.Artists").
			First(&artist, "id = ?", artistID).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Artist not found"})
			return
		}

		if artist.RedirectTo != nil {
			c.JSON(http.StatusOK, gin.H{
				"data":        artist,
				"redirect_to": artist.RedirectTo,
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{"data": artist})
	}
}

func UpdateArtistHandler(db *gorm.DB, revisionService *service.RevisionService) gin.HandlerFunc {
	return func(c *gin.Context) {
		artistID, err := uuid.Parse(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid artist ID"})
			return
		}

		var input struct {
			Name        string `json:"name"`
			Bio         string `json:"bio"`
			Nationality string `json:"nationality"`
			BirthYear   int    `json:"birth_year"`
			DeathYear   int    `json:"death_year"`
			Members     string `json:"members"`
			ImageURL    string `json:"image_url"`
			EditSummary string `json:"edit_summary"`
		}
		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		var artist model.Artist
		if err := db.First(&artist, "id = ?", artistID).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Artist not found"})
			return
		}

		updates := map[string]interface{}{}
		if input.Name != "" {
			updates["name"] = input.Name
		}
		if input.Bio != "" {
			updates["bio"] = input.Bio
		}
		if input.Nationality != "" {
			updates["nationality"] = input.Nationality
		}
		if input.BirthYear != 0 {
			updates["birth_year"] = input.BirthYear
		}
		if input.DeathYear != 0 {
			updates["death_year"] = input.DeathYear
		}
		if input.Members != "" {
			updates["members"] = input.Members
		}
		if input.ImageURL != "" {
			updates["image_url"] = input.ImageURL
		}

		if len(updates) == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "No changes provided"})
			return
		}

		if err := db.Model(&artist).Updates(updates).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update artist"})
			return
		}

		userID := c.GetString("user_id")
		editorUUID, _ := uuid.Parse(userID)
		userRole := c.GetString("role")
		autoApprove := userRole == "admin"

		var latestRev model.Revision
		var baseVersion int
		if err := db.Where("content_id = ? AND content_type = ?", artistID, "artist").
			Order("version_number DESC").First(&latestRev).Error; err == nil {
			baseVersion = latestRev.VersionNumber
		}

		if baseVersion > 0 {
			_, conflicts, err := revisionService.CreateRevision(
				"artist",
				artistID,
				editorUUID,
				updates,
				input.EditSummary,
				baseVersion,
				autoApprove,
			)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			if len(conflicts) > 0 {
				c.JSON(http.StatusConflict, gin.H{"error": "Edit conflicts detected", "conflicts": conflicts})
				return
			}
		}

		if err := db.Preload("Aliases").Preload("Albums").First(&artist, "id = ?", artistID).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to reload artist"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"data": artist})
	}
}

func GetArtistRevisionsHandler(revisionService *service.RevisionService) gin.HandlerFunc {
	return func(c *gin.Context) {
		artistID, err := uuid.Parse(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid artist ID"})
			return
		}

		limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
		offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
		revisions, total, err := revisionService.GetRevisions("artist", artistID, limit, offset)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch revisions"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"data": revisions, "total": total, "limit": limit, "offset": offset})
	}
}

func GetArtistRevisionHandler(revisionService *service.RevisionService) gin.HandlerFunc {
	return func(c *gin.Context) {
		artistID, err := uuid.Parse(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid artist ID"})
			return
		}

		version, err := strconv.Atoi(c.Param("version"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid version number"})
			return
		}

		var revision model.Revision
		if err := revisionService.GetDB().
			Where("content_id = ? AND content_type = ? AND version_number = ?", artistID, "artist", version).
			Preload("Editor").
			First(&revision).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Revision not found"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"data": revision})
	}
}

func CreateArtistRevisionHandler(db *gorm.DB, revisionService *service.RevisionService) gin.HandlerFunc {
	return func(c *gin.Context) {
		artistID, err := uuid.Parse(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid artist ID"})
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
		autoApprove := userRole == "admin"

		var protection model.ContentProtection
		if err := db.Where("content_id = ? AND content_type = ?", artistID, "artist").
			First(&protection).Error; err == nil {
			if protection.ProtectionLevel == "full" && userRole != "admin" {
				c.JSON(http.StatusForbidden, gin.H{"error": "This artist is fully protected"})
				return
			}
			if protection.ProtectionLevel == "semi" {
				autoApprove = false
			}
		}

		revision, conflicts, err := revisionService.CreateRevision(
			"artist", artistID, editorUUID, input.Changes, input.EditSummary, input.BaseRevision, autoApprove,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		if len(conflicts) > 0 {
			c.JSON(http.StatusConflict, gin.H{"error": "Edit conflicts detected", "conflicts": conflicts})
			return
		}

		c.JSON(http.StatusOK, gin.H{"data": revision, "message": statusMessage(autoApprove)})
	}
}

func RevertArtistHandler(revisionService *service.RevisionService) gin.HandlerFunc {
	return func(c *gin.Context) {
		artistID, err := uuid.Parse(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid artist ID"})
			return
		}

		version, _ := strconv.Atoi(c.Param("version"))
		var input struct {
			EditSummary string `json:"edit_summary"`
		}
		c.ShouldBindJSON(&input)

		userID := c.GetString("user_id")
		editorUUID, _ := uuid.Parse(userID)
		revision, err := revisionService.RevertToRevision("artist", artistID, version, editorUUID, input.EditSummary)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"data": revision, "message": "Artist reverted successfully"})
	}
}

func GetArtistAliasesHandler(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		artistID, err := uuid.Parse(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid artist ID"})
			return
		}

		var aliases []model.ArtistAlias
		if err := db.Where("artist_id = ?", artistID).Find(&aliases).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch aliases"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"data": aliases})
	}
}

func AddArtistAliasHandler(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		artistID, err := uuid.Parse(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid artist ID"})
			return
		}

		var input struct {
			Alias      string `json:"alias" binding:"required"`
			IsMainName bool   `json:"is_main_name"`
		}
		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		alias := model.ArtistAlias{
			ArtistID:   artistID,
			Alias:      input.Alias,
			IsMainName: input.IsMainName,
		}
		if err := db.Create(&alias).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create alias"})
			return
		}

		c.JSON(http.StatusCreated, gin.H{"data": alias})
	}
}

func DeleteArtistAliasHandler(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		artistID, err := uuid.Parse(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid artist ID"})
			return
		}

		aliasID, err := uuid.Parse(c.Param("aliasId"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid alias ID"})
			return
		}

		if err := db.Where("id = ? AND artist_id = ?", aliasID, artistID).
			Delete(&model.ArtistAlias{}).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete alias"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Alias deleted"})
	}
}

func MergeArtistsHandler(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		targetID, err := uuid.Parse(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid target artist ID"})
			return
		}

		var input struct {
			SourceID uuid.UUID `json:"source_id" binding:"required"`
		}
		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		userID := c.GetString("user_id")
		mergedByUUID, _ := uuid.Parse(userID)

		if err := db.Transaction(func(tx *gorm.DB) error {
			if err := tx.Exec("UPDATE album_artists SET artist_id = ? WHERE artist_id = ?", targetID, input.SourceID).Error; err != nil {
				return err
			}
			if err := tx.Exec("UPDATE song_artists SET artist_id = ? WHERE artist_id = ?", targetID, input.SourceID).Error; err != nil {
				return err
			}
			if err := tx.Exec(
				"UPDATE revisions SET content_id = ? WHERE content_id = ? AND content_type = 'artist'",
				targetID,
				input.SourceID,
			).Error; err != nil {
				return err
			}
			if err := tx.Exec(
				"UPDATE discussions SET content_id = ? WHERE content_id = ? AND content_type = 'artist'",
				targetID,
				input.SourceID,
			).Error; err != nil {
				return err
			}

			merge := model.ArtistMerge{
				SourceArtistID: input.SourceID,
				TargetArtistID: targetID,
				MergedBy:       mergedByUUID,
				MergedAt:       time.Now(),
			}
			if err := tx.Create(&merge).Error; err != nil {
				return err
			}

			if err := tx.Model(&model.Artist{}).Where("id = ?", input.SourceID).
				Update("redirect_to", targetID).Error; err != nil {
				return err
			}

			return nil
		}); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to merge artists"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Artists merged successfully"})
	}
}
