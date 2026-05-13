package handlers

import (
	"log"
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
	"atoman/internal/storage"
)

type AlbumInput struct {
	Title       string `form:"title"`
	Artist      string `form:"artist"`
	Year        int    `form:"year"`
	ReleaseDate string `form:"release_date"`
	CoverURL    string `form:"cover_url"`
}

func SetupAlbumRoutes(router *gin.Engine, db *gorm.DB, s3Client *s3.S3) {
	albums := router.Group("/api/albums")
	{
		albums.GET("", GetAlbumsHandler(db))
		albums.GET("/:id", GetAlbumHandler(db))
		albums.POST("", middleware.AuthMiddleware(), middleware.AdminMiddleware(db), CreateAlbumHandler(db, s3Client))
		// REMOVED AdminMiddleware - now allows all authenticated users to edit
		albums.PUT("/:id", middleware.AuthMiddleware(), UpdateAlbumHandler(db, s3Client))
		albums.DELETE("/:id", middleware.AuthMiddleware(), middleware.AdminMiddleware(db), DeleteAlbumHandler(db, s3Client))
	}
}

func GetAlbumsHandler(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var albums []model.Album
		if err := db.Where("status NOT IN ?", []string{"closed", "rejected", "draft"}).Preload("Artists").Order("release_date ASC, title ASC").Find(&albums).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch albums"})
			return
		}
		for i := range albums {
			albums[i].Status = normalizeMusicStatus(albums[i].Status)
		}
		c.JSON(http.StatusOK, albums)
	}
}

func GetAlbumHandler(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		var album model.Album
		if err := db.Preload("Artists").First(&album, "id = ?", id).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				c.JSON(http.StatusNotFound, gin.H{"error": "Album not found"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch album"})
			return
		}
		album.Status = normalizeMusicStatus(album.Status)
		c.JSON(http.StatusOK, album)
	}
}

func CreateAlbumHandler(db *gorm.DB, s3Client *s3.S3) gin.HandlerFunc {
	return func(c *gin.Context) {
		var input AlbumInput
		if err := c.ShouldBind(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		input.Title = strings.TrimSpace(input.Title)
		artistNames := splitArtistNames(input.Artist)
		if input.Title == "" || len(artistNames) == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "title and artist are required"})
			return
		}

		releaseDate := time.Now()
		if input.ReleaseDate != "" {
			parsedDate, err := time.Parse("2006-01-02", input.ReleaseDate)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "release_date must be YYYY-MM-DD"})
				return
			}
			releaseDate = parsedDate
		} else if input.Year > 0 {
			releaseDate = time.Date(input.Year, 1, 1, 0, 0, 0, 0, time.UTC)
		}

		year := input.Year
		if year == 0 {
			year = releaseDate.Year()
		}

		var userID *uuid.UUID
		if idVal, exists := c.Get("user_id"); exists {
			if uid, ok := idVal.(uuid.UUID); ok {
				userID = &uid
			}
		}

		coverURL := strings.TrimSpace(input.CoverURL)
		coverSource := ""
		if coverURL != "" {
			if strings.HasPrefix(coverURL, "/uploads/") {
				coverSource = "local"
			} else {
				coverSource = "s3"
			}
		}

		coverFile, coverHeader, err := c.Request.FormFile("cover")
		if err == nil {
			defer coverFile.Close()

			safeArtist := storage.SanitizeName(artistNames[0])
			safeAlbum := storage.SanitizeName(input.Title)
			coverKey := "music/" + safeArtist + "/" + safeAlbum + "/cover_" + coverHeader.Filename

			if s3Client != nil && os.Getenv("STORAGE_TYPE") == "s3" {
				_, err = s3Client.PutObject(&s3.PutObjectInput{
					Bucket: aws.String(os.Getenv("S3_BUCKET")),
					Key:    aws.String(coverKey),
					Body:   coverFile,
					ACL:    aws.String("public-read"),
				})
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to upload cover to S3"})
					return
				}

				coverURL = os.Getenv("S3_URL_PREFIX") + "/" + coverKey
				coverSource = "s3"
			} else {
				_, localURL, err := storage.SaveFileLocally(coverFile, "cover_"+coverHeader.Filename, safeArtist, input.Title)
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save cover locally"})
					return
				}
				coverURL = localURL
				coverSource = "local"
			}
		}

		tx := db.Begin()

		var existing model.Album
		if err := tx.Where("title = ? AND year = ?", input.Title, year).First(&existing).Error; err == nil {
			tx.Rollback()
			c.JSON(http.StatusConflict, gin.H{"error": "Album already exists", "id": existing.ID})
			return
		} else if err != gorm.ErrRecordNotFound {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check album"})
			return
		}

		album := model.Album{
			Title:       input.Title,
			Year:        year,
			ReleaseDate: releaseDate,
			CoverURL:    coverURL,
			CoverSource: coverSource,
			Status:      "open",
			UploadedBy:  userID,
		}
		if album.CoverSource == "" {
			album.CoverSource = "local"
		}

		if err := tx.Create(&album).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create album"})
			return
		}

		for _, name := range artistNames {
			var artist model.Artist
			if err := tx.FirstOrCreate(&artist, model.Artist{Name: name}).Error; err != nil {
				tx.Rollback()
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process artist"})
				return
			}
			if err := tx.Model(&album).Association("Artists").Append(&artist); err != nil {
				tx.Rollback()
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to link album to artist"})
				return
			}
		}

		if err := tx.Commit().Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create album"})
			return
		}

		db.Preload("Artists").First(&album, "id = ?", album.ID)
		c.JSON(http.StatusCreated, album)
	}
}

func splitArtistNames(value string) []string {
	parts := strings.Split(value, ",")
	names := make([]string, 0, len(parts))
	seen := make(map[string]bool, len(parts))
	for _, part := range parts {
		name := strings.TrimSpace(part)
		key := strings.ToLower(name)
		if name != "" && !seen[key] {
			seen[key] = true
			names = append(names, name)
		}
	}
	return names
}

func UpdateAlbumHandler(db *gorm.DB, s3Client *s3.S3) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")

		var album model.Album
		if err := db.Preload("Artists").First(&album, "id = ?", id).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				c.JSON(http.StatusNotFound, gin.H{"error": "Album not found"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch album"})
			return
		}

		var input AlbumInput
		if err := c.ShouldBind(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Get user info from context
		userID, userExists := c.Get("user_id")
		userRole := "anonymous"
		if roleVal, exists := c.Get("role"); exists {
			if role, ok := roleVal.(string); ok {
				userRole = role
			}
		}

		// Check ownership or admin permission
		if userRole != "admin" {
			if !userExists {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
				return
			}
			// Check if user owns this album
			if album.UploadedBy != nil && *album.UploadedBy != userID.(uuid.UUID) {
				c.JSON(http.StatusForbidden, gin.H{"error": "You can only edit your own albums"})
				return
			}
			// If album has no UploadedBy (legacy data), only admin can edit
			if album.UploadedBy == nil {
				c.JSON(http.StatusForbidden, gin.H{"error": "Cannot edit legacy albums without owner information"})
				return
			}
		}

		coverFile, coverHeader, err := c.Request.FormFile("cover")
		if err == nil {
			defer coverFile.Close()

			safeArtist := "Unknown Artist"
			if len(album.Artists) > 0 && album.Artists[0].Name != "" {
				safeArtist = strings.ReplaceAll(album.Artists[0].Name, "/", "-")
			}
			safeAlbum := strings.ReplaceAll(album.Title, "/", "-")
			if safeAlbum == "" {
				safeAlbum = "Unknown Album"
			}

			if s3Client != nil && os.Getenv("STORAGE_TYPE") == "s3" {
				coverKey := "music/" + safeArtist + "/" + safeAlbum + "/cover_" + coverHeader.Filename

				_, err = s3Client.PutObject(&s3.PutObjectInput{
					Bucket: aws.String(os.Getenv("S3_BUCKET")),
					Key:    aws.String(coverKey),
					Body:   coverFile,
					ACL:    aws.String("public-read"),
				})
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to upload cover to S3"})
					return
				}

				newCoverURL := os.Getenv("S3_URL_PREFIX") + "/" + coverKey

				if album.CoverURL != "" && album.CoverURL != newCoverURL {
					if album.CoverSource == "s3" {
						oldCoverKey := strings.TrimPrefix(album.CoverURL, os.Getenv("S3_URL_PREFIX")+"/")
						if err := storage.DeleteS3Object(s3Client, oldCoverKey); err != nil {
							log.Printf("Failed to delete old album cover %s from S3: %v", oldCoverKey, err)
						}
					} else if album.CoverSource == "local" {
						oldLocalPath := storage.GetLocalPathFromURL(album.CoverURL)
						storage.DeleteLocalFile(oldLocalPath)
					}
				}

				album.CoverURL = newCoverURL
				album.CoverSource = "s3"
			} else {
				_, localURL, err := storage.SaveFileLocally(coverFile, "cover_"+coverHeader.Filename, safeArtist, album.Title)
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save cover locally"})
					return
				}

				if album.CoverURL != "" && album.CoverURL != localURL {
					if album.CoverSource == "local" {
						oldLocalPath := storage.GetLocalPathFromURL(album.CoverURL)
						storage.DeleteLocalFile(oldLocalPath)
					}
				}

				album.CoverURL = localURL
				album.CoverSource = "local"
			}
		}

		if input.Title != "" {
			album.Title = input.Title
		}
		if input.Year != 0 {
			album.Year = input.Year
		}

		// Handle ReleaseDate
		if input.ReleaseDate != "" {
			timeZone, err := time.LoadLocation("Asia/Shanghai")
			if err != nil {
				timeZone = time.UTC
			}
			parsedDate, err := time.ParseInLocation("2006-01-02", input.ReleaseDate, timeZone)
			if err == nil {
				album.ReleaseDate = parsedDate
			}
		}

		// Handle Artist - find or create artist and update association
		if input.Artist != "" {
			var artist model.Artist
			result := db.Where("name = ?", input.Artist).First(&artist)
			if result.Error == gorm.ErrRecordNotFound {
				artist = model.Artist{Name: input.Artist}
				if err := db.Create(&artist).Error; err != nil {
					log.Printf("Failed to create artist: %v", err)
				} else {
					db.Model(&album).Association("Artists").Replace(&artist)
				}
			} else if result.Error == nil {
				db.Model(&album).Association("Artists").Replace(&artist)
			}
		}

		if err := db.Save(&album).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update album"})
			return
		}

		c.JSON(http.StatusOK, album)
	}
}

func DeleteAlbumHandler(db *gorm.DB, s3Client *s3.S3) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		var album model.Album
		if err := db.First(&album, "id = ?", id).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Album not found"})
			return
		}

		if err := db.Delete(&album).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete album"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Album deleted successfully"})
	}
}
