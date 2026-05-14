package handlers

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"

	"atoman/internal/middleware"
	"atoman/internal/model"
	"atoman/internal/storage"
)

func SetupAdminRoutes(router *gin.Engine, db *gorm.DB, s3Client *s3.S3) {
	admin := router.Group("/api/admin")
	admin.Use(middleware.AuthMiddleware())
	admin.Use(middleware.AdminMiddleware(db))
	{
		admin.GET("/pending", GetPendingSongsHandler(db))
		admin.POST("/approve/:id", ApproveSongHandler(db, s3Client))
		admin.POST("/reject/:id", RejectSongHandler(db, s3Client))

		admin.GET("/pending-song-corrections", GetPendingSongCorrectionsHandler(db))
		admin.POST("/approve-song-correction/:id", ApproveSongCorrectionHandler(db))
		admin.POST("/reject-song-correction/:id", RejectSongCorrectionHandler(db))

		admin.GET("/pending-albums", GetPendingAlbumsHandler(db))
		admin.POST("/approve-album/:id", ApproveAlbumHandler(db, s3Client))
		admin.POST("/reject-album/:id", RejectAlbumHandler(db, s3Client))

		admin.GET("/pending-album-corrections", GetPendingAlbumCorrectionsHandler(db))
		admin.POST("/approve-album-correction/:id", ApproveAlbumCorrectionHandler(db))
		admin.POST("/reject-album-correction/:id", RejectAlbumCorrectionHandler(db, s3Client))

			admin.GET("/pending-artist-corrections", GetPendingArtistCorrectionsHandler(db))
			admin.POST("/approve-artist-correction/:id", ApproveArtistCorrectionHandler(db))
			admin.POST("/reject-artist-correction/:id", RejectArtistCorrectionHandler(db))
	}
}

func canUploadToS3(s3Client *s3.S3) bool {
	return s3Client != nil && os.Getenv("S3_BUCKET") != "" && os.Getenv("S3_URL_PREFIX") != ""
}

func GetPendingSongsHandler(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var songs []model.Song
		if err := db.Where("status = ?", "pending").
			Preload("User").
			Preload("Album").
			Preload("Album.Artists").
			Preload("Artists").
			Order("created_at desc").
			Find(&songs).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch pending songs"})
			return
		}
		c.JSON(http.StatusOK, songs)
	}
}

func ApproveSongHandler(db *gorm.DB, s3Client *s3.S3) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")

		var song model.Song
		if err := db.First(&song, "id = ?", id).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Song not found"})
			return
		}

		if canUploadToS3(s3Client) && song.AudioSource == "local" && song.AudioURL != "" {
			localPath := storage.GetLocalPathFromURL(song.AudioURL)
			if localPath != "" {
				s3URL, err := storage.UploadLocalFileToS3(s3Client, localPath)
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to upload audio to S3"})
					return
				}
				song.AudioURL = s3URL
				song.AudioSource = "s3"
				storage.DeleteLocalFile(localPath)
			}
		}

		if canUploadToS3(s3Client) && song.CoverSource == "local" && song.CoverURL != "" {
			localPath := storage.GetLocalPathFromURL(song.CoverURL)
			if localPath != "" {
				s3URL, err := storage.UploadLocalFileToS3(s3Client, localPath)
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to upload cover to S3"})
					return
				}
				song.CoverURL = s3URL
				song.CoverSource = "s3"
				storage.DeleteLocalFile(localPath)
			}
		}

		song.Status = "approved"

		if err := db.Save(&song).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to approve song"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Song approved"})
	}
}

func RejectSongHandler(db *gorm.DB, s3Client *s3.S3) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")

		var song model.Song
		if err := db.Preload("Album").First(&song, "id = ?", id).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Song not found"})
			return
		}

		if song.AudioSource == "local" && song.AudioURL != "" {
			localPath := storage.GetLocalPathFromURL(song.AudioURL)
			storage.DeleteLocalFile(localPath)
		}

		if song.CoverSource == "local" && song.CoverURL != "" {
			localPath := storage.GetLocalPathFromURL(song.CoverURL)
			storage.DeleteLocalFile(localPath)
		}

		if err := storage.DeleteSongAndS3Objects(db, s3Client, &song); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete song and associated files"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Song rejected and deleted"})
	}
}

func GetPendingSongCorrectionsHandler(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var corrections []model.SongCorrection
		if err := db.Where("status = ?", "pending").
			Preload("User").
			Preload("Song").
			Order("created_at desc").
			Find(&corrections).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch pending song corrections"})
			return
		}
		c.JSON(http.StatusOK, corrections)
	}
}

func ApproveSongCorrectionHandler(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		adminIDVal, _ := c.Get("user_id")
		adminID := adminIDVal.(uuid.UUID)
		now := time.Now()

		var correction model.SongCorrection
		if err := db.Preload("Song").First(&correction, "id = ?", id).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Correction not found"})
			return
		}

		tx := db.Begin()

		if err := tx.Model(&model.SongCorrection{}).Where("id = ?", id).Updates(map[string]interface{}{
			"status":      "approved",
			"approved_by": adminID,
			"approved_at": now,
		}).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to approve correction"})
			return
		}

		song := correction.Song
		updated := false

		switch correction.FieldName {
		case "title":
			song.Title = correction.CorrectedValue
			updated = true
		case "lyrics":
			song.Lyrics = correction.CorrectedValue
			updated = true
		}

		if updated {
			if err := tx.Save(&song).Error; err != nil {
				tx.Rollback()
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to apply correction"})
				return
			}
		}

		tx.Commit()
		c.JSON(http.StatusOK, gin.H{"message": "Song correction approved and applied"})
	}
}

func RejectSongCorrectionHandler(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		adminIDVal, _ := c.Get("user_id")
		adminID := adminIDVal.(uuid.UUID)
		now := time.Now()

		if err := db.Model(&model.SongCorrection{}).Where("id = ?", id).Updates(map[string]interface{}{
			"status":      "rejected",
			"rejected_by": adminID,
			"rejected_at": now,
		}).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to reject correction"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "Song correction rejected"})
	}
}

func GetPendingAlbumsHandler(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var albums []model.Album
		if err := db.Where("status = ?", "pending").
			Preload("Artists").
			Preload("User").
			Preload("Songs").
			Order("created_at desc").
			Find(&albums).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch pending albums"})
			return
		}
		c.JSON(http.StatusOK, albums)
	}
}

func ApproveAlbumHandler(db *gorm.DB, s3Client *s3.S3) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")

		var album model.Album
		if err := db.Preload("Songs").First(&album, "id = ?", id).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Album not found"})
			return
		}

		// Upload album cover to S3 if local
		if canUploadToS3(s3Client) && album.CoverSource == "local" && album.CoverURL != "" {
			localPath := storage.GetLocalPathFromURL(album.CoverURL)
			if localPath != "" {
				s3URL, err := storage.UploadLocalFileToS3(s3Client, localPath)
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to upload cover to S3"})
					return
				}
				album.CoverURL = s3URL
				album.CoverSource = "s3"
				storage.DeleteLocalFile(localPath)
			}
		}

		// Upload all songs' local files to S3
		for i := range album.Songs {
			song := &album.Songs[i]
			if canUploadToS3(s3Client) && song.AudioSource == "local" && song.AudioURL != "" {
				localPath := storage.GetLocalPathFromURL(song.AudioURL)
				if localPath != "" {
					s3URL, err := storage.UploadLocalFileToS3(s3Client, localPath)
					if err != nil {
						log.Printf("Failed to upload song audio to S3: %v", err)
						continue
					}
					song.AudioURL = s3URL
					song.AudioSource = "s3"
					storage.DeleteLocalFile(localPath)
				}
			}
			if canUploadToS3(s3Client) && song.CoverSource == "local" && song.CoverURL != "" {
				localPath := storage.GetLocalPathFromURL(song.CoverURL)
				if localPath != "" {
					s3URL, err := storage.UploadLocalFileToS3(s3Client, localPath)
					if err != nil {
						log.Printf("Failed to upload song cover to S3: %v", err)
						continue
					}
					song.CoverURL = s3URL
					song.CoverSource = "s3"
					storage.DeleteLocalFile(localPath)
				}
			}
			song.Status = "approved"
			db.Save(song)
		}

		album.Status = "approved"

		if err := db.Save(&album).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to approve album"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Album approved"})
	}
}

func RejectAlbumHandler(db *gorm.DB, s3Client *s3.S3) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")

		var album model.Album
		if err := db.First(&album, "id = ?", id).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Album not found"})
			return
		}

		if album.CoverSource == "local" && album.CoverURL != "" {
			localPath := storage.GetLocalPathFromURL(album.CoverURL)
			storage.DeleteLocalFile(localPath)
		}

		if err := storage.DeleteAlbumAndS3Objects(db, s3Client, &album); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete album and associated files"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Album rejected and deleted"})
	}
}

func GetPendingAlbumCorrectionsHandler(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var corrections []model.AlbumCorrection
		if err := db.Where("status = ?", "pending").
			Preload("User").
			Preload("Album").
			Preload("Album.Artists").
			Order("created_at desc").
			Find(&corrections).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch pending album corrections"})
			return
		}
		c.JSON(http.StatusOK, corrections)
	}
}

func ApproveAlbumCorrectionHandler(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		adminIDVal, _ := c.Get("user_id")
		adminID := adminIDVal.(uuid.UUID)
		now := time.Now()

		var correction model.AlbumCorrection
		if err := db.Preload("Album").Preload("Album.Artists").First(&correction, "id = ?", id).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Correction not found"})
			return
		}

		tx := db.Begin()

		if err := tx.Model(&model.AlbumCorrection{}).Where("id = ?", id).Updates(map[string]interface{}{
			"status":      "approved",
			"approved_by": adminID,
			"approved_at": now,
		}).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to approve correction"})
			return
		}

		var album model.Album
		if err := tx.First(&album, "id = ?", correction.AlbumID).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusNotFound, gin.H{"error": "Album not found"})
			return
		}

		if correction.CorrectedTitle != "" {
			album.Title = correction.CorrectedTitle
		}
		if correction.CorrectedCoverURL != "" {
			album.CoverURL = correction.CorrectedCoverURL
			album.CoverSource = correction.CorrectedCoverSource
		}
		if correction.CorrectedReleaseDate != nil {
			album.ReleaseDate = *correction.CorrectedReleaseDate
		}

		if err := tx.Save(&album).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to apply album correction"})
			return
		}

		tx.Commit()
		c.JSON(http.StatusOK, gin.H{"message": "Album correction approved and applied"})
	}
}

func RejectAlbumCorrectionHandler(db *gorm.DB, s3Client *s3.S3) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		adminIDVal, _ := c.Get("user_id")
		adminID := adminIDVal.(uuid.UUID)
		now := time.Now()

		var correction model.AlbumCorrection
		if err := db.First(&correction, "id = ?", id).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Correction not found"})
			return
		}

		if err := db.Model(&model.AlbumCorrection{}).Where("id = ?", id).Updates(map[string]interface{}{
			"status":      "rejected",
			"rejected_by": adminID,
			"rejected_at": now,
		}).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to reject correction"})
			return
		}

		if correction.CorrectedCoverURL != "" && correction.CorrectedCoverSource == "s3" {
			log.Printf("Note: Should delete S3 object for rejected cover: %s", correction.CorrectedCoverURL)
		}

		c.JSON(http.StatusOK, gin.H{"message": "Album correction rejected"})
	}
}

func GetPendingArtistCorrectionsHandler(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var corrections []model.ArtistCorrection
		if err := db.Where("status = ?", "pending").
			Preload("Artist").
			Preload("User").
			Order("created_at asc").
			Find(&corrections).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch pending artist corrections"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": corrections})
	}
}

func ApproveArtistCorrectionHandler(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		adminIDVal, _ := c.Get("user_id")
		adminID := adminIDVal.(uuid.UUID)
		now := time.Now()

		var correction model.ArtistCorrection
		if err := db.First(&correction, "id = ?", id).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Correction not found"})
			return
		}

		if err := db.Model(&model.ArtistCorrection{}).Where("id = ?", id).Updates(map[string]interface{}{
			"status":      "approved",
			"approved_by": adminID,
			"approved_at": now,
		}).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to approve correction"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "Artist correction approved"})
	}
}

func RejectArtistCorrectionHandler(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")

		var correction model.ArtistCorrection
		if err := db.First(&correction, "id = ?", id).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Correction not found"})
			return
		}

		if err := db.Model(&model.ArtistCorrection{}).Where("id = ?", id).Update("status", "rejected").Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to reject correction"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "Artist correction rejected"})
	}
}
