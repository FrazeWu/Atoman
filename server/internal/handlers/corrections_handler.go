package handlers

import (
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"

	"atoman/internal/middleware"
	"atoman/internal/model"
)

type SongCorrectionInput struct {
	SongID         uuid.UUID `json:"song_id" binding:"required"`
	FieldName      string    `json:"field_name" binding:"required"`
	CurrentValue   string    `json:"current_value"`
	CorrectedValue string    `json:"corrected_value" binding:"required"`
	Reason         string    `json:"reason"`
}

type AlbumCorrectionFormInput struct {
	AlbumID              uuid.UUID             `form:"album_id" binding:"required"`
	CorrectedTitle       string                `form:"corrected_title"`
	CorrectedReleaseDate string                `form:"corrected_release_date"`
	Reason               string                `form:"reason"`
	Cover                *multipart.FileHeader `form:"cover"`
}

func SetupCorrectionRoutes(router *gin.Engine, db *gorm.DB, s3Client *s3.S3) {
	corrections := router.Group("/api/corrections")
	{
		corrections.POST("/song", middleware.AuthMiddleware(), CreateSongCorrectionHandler(db))
		corrections.POST("/album", middleware.AuthMiddleware(), CreateAlbumCorrectionHandler(db, s3Client))
		corrections.POST("/artist", middleware.AuthMiddleware(), CreateArtistCorrectionHandler(db))
	}
}

func CreateSongCorrectionHandler(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var uid uuid.UUID
		var userRole string
		if idVal, exists := c.Get("user_id"); exists {
			uid = idVal.(uuid.UUID)
		}
		if roleVal, exists := c.Get("role"); exists {
			if role, ok := roleVal.(string); ok {
				userRole = role
			}
		}

		var input SongCorrectionInput

		if err := c.ShouldBindJSON(&input); err != nil {
			log.Printf("Song correction input error: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		status := "pending"
		var approvedBy *uuid.UUID
		var approvedAt *time.Time
		if userRole == "admin" {
			status = "approved"
			approvedBy = &uid
			now := time.Now()
			approvedAt = &now
		}

		correction := model.SongCorrection{
			SongID:         input.SongID,
			FieldName:      input.FieldName,
			CurrentValue:   input.CurrentValue,
			CorrectedValue: input.CorrectedValue,
			Reason:         input.Reason,
			UserID:         &uid,
			Status:         status,
			ApprovedBy:     approvedBy,
			ApprovedAt:     approvedAt,
		}

		if err := db.Create(&correction).Error; err != nil {
			log.Printf("Failed to create song correction: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to submit song correction"})
			return
		}

		c.JSON(http.StatusCreated, gin.H{
			"message": "Song correction submitted successfully",
			"id":      correction.ID,
			"status":  status,
		})
	}
}

func CreateAlbumCorrectionHandler(db *gorm.DB, s3Client *s3.S3) gin.HandlerFunc {
	return func(c *gin.Context) {
		var uid uuid.UUID
		var userRole string
		if idVal, exists := c.Get("user_id"); exists {
			uid = idVal.(uuid.UUID)
		}
		if roleVal, exists := c.Get("role"); exists {
			if role, ok := roleVal.(string); ok {
				userRole = role
			}
		}

		var input AlbumCorrectionFormInput
		if err := c.ShouldBind(&input); err != nil {
			log.Printf("Album correction input error: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		var originalAlbum model.Album
		if err := db.Preload("Artists").First(&originalAlbum, "id = ?", input.AlbumID).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Original album not found"})
			return
		}

		hasChanges := input.CorrectedTitle != "" || input.CorrectedReleaseDate != "" || input.Cover != nil

		if !hasChanges {
			c.JSON(http.StatusBadRequest, gin.H{"error": "No changes provided"})
			return
		}

		if hasChanges && input.Reason == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Reason is required for album corrections"})
			return
		}

		status := "pending"
		var approvedBy *uuid.UUID
		var approvedAt *time.Time
		if userRole == "admin" {
			status = "approved"
			approvedBy = &uid
			now := time.Now()
			approvedAt = &now
		}

		correction := model.AlbumCorrection{
			AlbumID:          input.AlbumID,
			UserID:           &uid,
			Status:           status,
			Reason:           input.Reason,
			OriginalTitle:    originalAlbum.Title,
			OriginalCoverURL: originalAlbum.CoverURL,
			ApprovedBy:       approvedBy,
			ApprovedAt:       approvedAt,
		}

		if !originalAlbum.ReleaseDate.IsZero() {
			correction.OriginalReleaseDate = &originalAlbum.ReleaseDate
		}

		if input.CorrectedTitle != "" {
			correction.CorrectedTitle = input.CorrectedTitle
		}

		if input.CorrectedReleaseDate != "" {
			parsedDate, err := time.Parse("2006-01-02", input.CorrectedReleaseDate)
			if err == nil {
				correction.CorrectedReleaseDate = &parsedDate
			}
		}

		if input.Cover != nil {
			log.Printf("Uploading new cover for album %v", input.AlbumID)
			src, err := input.Cover.Open()
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to open cover file"})
				return
			}
			defer src.Close()

			safeAlbum := strings.ReplaceAll(originalAlbum.Title, "/", "-")
			coverKey := "album_covers/pending/" + safeAlbum + "/" + input.Cover.Filename

			_, err = s3Client.PutObject(&s3.PutObjectInput{
				Bucket: aws.String(os.Getenv("S3_BUCKET")),
				Key:    aws.String(coverKey),
				Body:   src,
				ACL:    aws.String("public-read"),
			})
			if err != nil {
				log.Printf("Failed to upload cover to S3: %v", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to upload cover to S3"})
				return
			}
			uploadedCoverURL := os.Getenv("S3_URL_PREFIX") + "/" + coverKey

			correction.CorrectedCoverURL = uploadedCoverURL
			correction.CorrectedCoverSource = "s3"
		}

		if err := db.Create(&correction).Error; err != nil {
			log.Printf("Failed to create album correction: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to submit album correction"})
			return
		}

		c.JSON(http.StatusCreated, gin.H{
			"message": "Album correction submitted successfully",
			"id":      correction.ID,
			"status":  status,
		})
	}
}

// CreateArtistCorrectionHandler submits a proposed change for a confirmed artist entry.
// Route: POST /api/corrections/artist
// Auth: Required
func CreateArtistCorrectionHandler(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		uid, exists := c.Get("userID")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}
		userID := uid.(uuid.UUID)

		var req struct {
			ArtistID    string `json:"artist_id" binding:"required"`
			Description string `json:"description" binding:"required"`
			Reason      string `json:"reason"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		artistUUID, err := uuid.Parse(req.ArtistID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid artist_id"})
			return
		}

		var artist model.Artist
		if err := db.First(&artist, "id = ?", artistUUID).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "artist not found"})
			return
		}

		correction := model.ArtistCorrection{
			ArtistID:    artistUUID,
			UserID:      &userID,
			Description: req.Description,
			Reason:      req.Reason,
			Status:      "pending",
		}
		if err := db.Create(&correction).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create correction"})
			return
		}

		c.JSON(http.StatusCreated, gin.H{"data": correction})
	}
}
