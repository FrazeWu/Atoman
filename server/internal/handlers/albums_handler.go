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
		albums.PUT("/:id", middleware.AuthMiddleware(), middleware.AdminMiddleware(db), UpdateAlbumHandler(db, s3Client))
		albums.DELETE("/:id", middleware.AuthMiddleware(), middleware.AdminMiddleware(db), DeleteAlbumHandler(db, s3Client))
	}
}

func GetAlbumsHandler(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var albums []model.Album
		if err := db.Where("status = ?", "approved").Preload("Artists").Order("release_date ASC, title ASC").Find(&albums).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch albums"})
			return
		}
		c.JSON(http.StatusOK, albums)
	}
}

func GetAlbumHandler(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		var album model.Album
		if err := db.Preload("Artists").First(&album, "id = ?", id).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Album not found"})
			return
		}
		c.JSON(http.StatusOK, album)
	}
}

// Placeholder handlers
func CreateAlbumHandler(db *gorm.DB, s3Client *s3.S3) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusNotImplemented, gin.H{"error": "Album creation not yet implemented via direct API, use song upload or admin tools"})
	}
}

func UpdateAlbumHandler(db *gorm.DB, s3Client *s3.S3) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")

		var album model.Album
		if err := db.Preload("Artists").First(&album, "id = ?", id).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Album not found"})
			return
		}

		var input AlbumInput
		if err := c.ShouldBind(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		var userRole string
		if roleVal, exists := c.Get("role"); exists {
			if role, ok := roleVal.(string); ok {
				userRole = role
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

			if userRole == "admin" {
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
