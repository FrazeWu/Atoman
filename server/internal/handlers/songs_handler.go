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

// SongInput represents song creation request
type SongInput struct {
	Title       string `form:"title" binding:"required"`
	Artist      string `form:"artist" binding:"required"`
	Album       string `form:"album"`
	ReleaseDate string `form:"release_date"` // YYYY-MM-DD
	TrackNumber int    `form:"track_number"`
	Lyrics      string `form:"lyrics"`
	BatchID     string `form:"batch_id"`
	AudioURL    string `form:"audio_url"` // For reusing existing audio
	CoverURL    string `form:"cover_url"` // For reusing existing cover
}

// SetupSongRoutes configures song-related routes
func SetupSongRoutes(router *gin.Engine, db *gorm.DB, s3Client *s3.S3) {
	songs := router.Group("/api/songs")
	{
		songs.GET("", GetSongsHandler(db))
		songs.GET("/:id", GetSongHandler(db))
		songs.POST("", middleware.AuthMiddleware(), CreateSongHandler(db, s3Client))
		songs.PUT("/:id", middleware.AuthMiddleware(), UpdateSongHandler(db, s3Client))
		songs.DELETE("/:id", middleware.AuthMiddleware(), DeleteSongHandler(db, s3Client))
	}
}

// GetSongsHandler retrieves all approved songs
func GetSongsHandler(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var songs []model.Song

		if err := db.Where("status = ?", "approved").
			Preload("Album").
			Preload("Album.Artists").
			Preload("Artists").
			Order("release_date ASC, track_number ASC, title ASC").
			Find(&songs).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch songs"})
			return
		}

		response := make([]map[string]interface{}, len(songs))
		for i, song := range songs {
			artistName := "Unknown Artist"
			albumTitle := "Unknown Album"
			albumYear := 0
			releaseDate := song.ReleaseDate.Format("2006-01-02")
			coverURL := song.CoverURL

			if song.Album != nil {
				albumTitle = song.Album.Title
				albumYear = song.Album.Year
				// If album year is 0, use release date year
				if albumYear == 0 && !song.ReleaseDate.IsZero() {
					albumYear = song.ReleaseDate.Year()
				}
				if song.Album.CoverURL != "" {
					coverURL = song.Album.CoverURL
				}
				if len(song.Album.Artists) > 0 && song.Album.Artists[0].Name != "" {
					artistName = song.Album.Artists[0].Name
				}
			}

			response[i] = map[string]interface{}{
				"id":           song.ID,
				"title":        song.Title,
				"artist":       artistName,
				"album":        albumTitle,
				"album_id":     song.AlbumID,
				"year":         albumYear,
				"release_date": releaseDate,
				"lyrics":       song.Lyrics,
				"audio_url":    song.AudioURL,
				"cover_url":    coverURL,
			}
		}

		c.JSON(http.StatusOK, response)
	}
}

// GetSongHandler retrieves a single song by ID
func GetSongHandler(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")

		var song model.Song
		if err := db.Preload("Album").Preload("Album.Artists").Preload("Artists").First(&song, "id = ?", id).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Song not found"})
			return
		}

		artistName := "Unknown Artist"
		albumTitle := "Unknown Album"
		albumYear := 0
		releaseDate := song.ReleaseDate.Format("2006-01-02")
		coverURL := song.CoverURL

		if song.Album != nil {
			albumTitle = song.Album.Title
			albumYear = song.Album.Year
			// If album year is 0, use release date year
			if albumYear == 0 && !song.ReleaseDate.IsZero() {
				albumYear = song.ReleaseDate.Year()
			}
			if song.Album.CoverURL != "" {
				coverURL = song.Album.CoverURL
			}
			if len(song.Album.Artists) > 0 && song.Album.Artists[0].Name != "" {
				artistName = song.Album.Artists[0].Name
			}
		}

		response := map[string]interface{}{
			"id":           song.ID,
			"title":        song.Title,
			"artist":       artistName,
			"album":        albumTitle,
			"album_id":     song.AlbumID,
			"year":         albumYear,
			"release_date": releaseDate,
			"lyrics":       song.Lyrics,
			"audio_url":    song.AudioURL,
			"cover_url":    coverURL,
		}

		c.JSON(http.StatusOK, response)
	}
}

// CreateSongHandler creates a new song with optional audio upload
func CreateSongHandler(db *gorm.DB, s3Client *s3.S3) gin.HandlerFunc {
	return func(c *gin.Context) {
		var input SongInput

		if err := c.ShouldBind(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Parse ReleaseDate
		var releaseDate time.Time
		var err error
		if input.ReleaseDate != "" {
			releaseDate, err = time.Parse("2006-01-02", input.ReleaseDate)
			if err != nil {
				releaseDate = time.Now()
			}
		} else {
			releaseDate = time.Now()
		}

		// Check for duplicate song before uploading
		checkAlbum := input.Album
		if checkAlbum == "" {
			checkAlbum = "Unknown Album"
		}

		var existingCount int64
		if err := db.Table("Songs").
			Joins("JOIN Albums ON Albums.id = Songs.album_id").
			Joins("JOIN album_artists ON album_artists.album_id = Albums.id").
			Joins("JOIN Artists ON Artists.id = album_artists.artist_id").
			Where("Songs.title = ? AND Albums.title = ? AND Artists.name = ? AND Songs.status != 'rejected'",
				input.Title, checkAlbum, input.Artist).
			Count(&existingCount).Error; err != nil {
			log.Printf("Error checking for duplicates: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error checking duplicates"})
			return
		}

		if existingCount > 0 {
			log.Printf("Skipping duplicate song: %s - %s - %s", input.Title, checkAlbum, input.Artist)

			// Return success response with existing song info
			var existingSong model.Song
			db.Table("Songs").
				Joins("JOIN Albums ON Albums.id = Songs.album_id").
				Joins("JOIN album_artists ON album_artists.album_id = Albums.id").
				Joins("JOIN Artists ON Artists.id = album_artists.artist_id").
				Where("Songs.title = ? AND Albums.title = ? AND Artists.name = ? AND Songs.status != 'rejected'",
					input.Title, checkAlbum, input.Artist).
				First(&existingSong)

			c.JSON(http.StatusCreated, existingSong)
			return
		}

		// Determine user role first
		var userRole string
		if roleVal, exists := c.Get("role"); exists {
			if role, ok := roleVal.(string); ok {
				userRole = role
			}
		}

		// Handle File Upload Logic
		var audioURL string
		var audioSource string
		var coverURL string
		var coverSource string

		// Audio file handling
		if input.AudioURL != "" {
			audioURL = input.AudioURL
			if strings.HasPrefix(audioURL, "/uploads/") {
				audioSource = "local"
			} else {
				audioSource = "s3"
			}
		} else {
			file, header, err := c.Request.FormFile("audio")
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Audio file is required"})
				return
			}
			defer file.Close()

			if userRole == "admin" {
				// Admin: upload directly to S3
				safeArtist := strings.ReplaceAll(input.Artist, "/", "-")
				if safeArtist == "" {
					safeArtist = "Unknown Artist"
				}
				safeAlbum := strings.ReplaceAll(input.Album, "/", "-")
				if safeAlbum == "" {
					safeAlbum = "Unknown Album"
				}
				key := "music/" + safeArtist + "/" + safeAlbum + "/" + header.Filename
				_, err = s3Client.PutObject(&s3.PutObjectInput{
					Bucket: aws.String(os.Getenv("S3_BUCKET")),
					Key:    aws.String(key),
					Body:   file,
					ACL:    aws.String("public-read"),
				})
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to upload file to S3"})
					return
				}
				audioURL = os.Getenv("S3_URL_PREFIX") + "/" + key
				audioSource = "s3"
			} else {
				// Regular user: save locally
				_, localURL, err := storage.SaveFileLocally(file, header.Filename, input.Artist, input.Album)
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save file locally"})
					return
				}
				audioURL = localURL
				audioSource = "local"
			}
		}

		// Cover file handling (similar logic)
		if input.CoverURL != "" {
			coverURL = input.CoverURL
			if strings.HasPrefix(coverURL, "/uploads/") {
				coverSource = "local"
			} else {
				coverSource = "s3"
			}
		} else {
			coverFile, coverHeader, err := c.Request.FormFile("cover")
			if err == nil {
				defer coverFile.Close()

				if userRole == "admin" {
					safeArtist := strings.ReplaceAll(input.Artist, "/", "-")
					if safeArtist == "" {
						safeArtist = "Unknown Artist"
					}
					safeAlbum := strings.ReplaceAll(input.Album, "/", "-")
					if safeAlbum == "" {
						safeAlbum = "Unknown Album"
					}
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
					coverURL = os.Getenv("S3_URL_PREFIX") + "/" + coverKey
					coverSource = "s3"
				} else {
					_, localURL, err := storage.SaveFileLocally(coverFile, "cover_"+coverHeader.Filename, input.Artist, input.Album)
					if err != nil {
						c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save cover locally"})
						return
					}
					coverURL = localURL
					coverSource = "local"
				}
			}
		}

		// Transaction to ensure atomicity
		tx := db.Begin()

		// 1. Find or Create Artist
		var artist model.Artist
		if err := tx.FirstOrCreate(&artist, model.Artist{Name: input.Artist}).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process artist"})
			return
		}

		// 2. Find or Create Album
		var album model.Album
		albumTitle := input.Album
		if albumTitle == "" {
			albumTitle = "Unknown Album"
		}

		if err := tx.Where("title = ? AND year = ?", albumTitle, releaseDate.Year()).FirstOrCreate(&album, model.Album{Title: albumTitle, Year: releaseDate.Year(), ReleaseDate: releaseDate}).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process album"})
			return
		}
		if album.ID != uuid.Nil {
			var existingAssoc int64
			tx.Table("album_artists").Where("album_id = ? AND artist_id = ?", album.ID, artist.ID).Count(&existingAssoc)
			if existingAssoc == 0 {
				if err := tx.Model(&album).Association("Artists").Append(&artist); err != nil {
					tx.Rollback()
					c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to link album to artist"})
					return
				}
			}
		}

		if coverURL != "" && album.CoverURL == "" {
			album.CoverURL = coverURL
			album.CoverSource = coverSource
			if err := tx.Save(&album).Error; err != nil {
				tx.Rollback()
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update album cover"})
				return
			}
		}

		// 3. Create Song
		// Get User ID from context
		var userID *uuid.UUID
		if idVal, exists := c.Get("user_id"); exists {
			uid := idVal.(uuid.UUID)
			userID = &uid
		}

		status := "pending"
		if userRole == "admin" {
			status = "approved"
		}

		song := model.Song{
			Title:       input.Title,
			ReleaseDate: releaseDate,
			TrackNumber: input.TrackNumber,
			Lyrics:      input.Lyrics,
			AudioURL:    audioURL,
			AudioSource: audioSource,
			CoverURL:    coverURL,
			CoverSource: coverSource,
			Status:      status,
			AlbumID:     &album.ID,
			UploadedBy:  userID,
			BatchID:     input.BatchID,
		}

		if err := tx.Create(&song).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create song"})
			return
		}

		// 4. Associate Song with Artist (Many-to-Many)
		if err := tx.Model(&song).Association("Artists").Append(&artist); err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to associate artist"})
			return
		}

		tx.Commit()

		// Reload song with associations for response
		db.Preload("Album").Preload("Artists").First(&song, "id = ?", song.ID)
		c.JSON(http.StatusCreated, song)
	}
}

// UpdateSongHandler updates an existing song
func UpdateSongHandler(db *gorm.DB, s3Client *s3.S3) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")

		var song model.Song
		if err := db.Preload("Album").Preload("Album.Artists").First(&song, "id = ?", id).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Song not found"})
			return
		}

		var input SongInput
		if err := c.ShouldBind(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Parse ReleaseDate
		var releaseDate time.Time
		var err error
		if input.ReleaseDate != "" {
			releaseDate, err = time.Parse("2006-01-02", input.ReleaseDate)
			if err != nil {
				releaseDate = time.Now()
			}
		} else {
			releaseDate = time.Now()
		}

		// Handle Cover Upload
		var coverURL string
		var coverSource string

		// Determine user role
		var userRole string
		if roleVal, exists := c.Get("role"); exists {
			if role, ok := roleVal.(string); ok {
				userRole = role
			}
		}

		coverFile, coverHeader, err := c.Request.FormFile("cover")
		if err == nil {
			defer coverFile.Close()

			safeArtist := strings.ReplaceAll(input.Artist, "/", "-")
			if safeArtist == "" {
				safeArtist = "Unknown Artist"
			}
			safeAlbum := strings.ReplaceAll(input.Album, "/", "-")
			if safeAlbum == "" {
				safeAlbum = "Unknown Album"
			}

			if userRole == "admin" {
				// Admin: upload to S3
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

				coverURL = os.Getenv("S3_URL_PREFIX") + "/" + coverKey
				coverSource = "s3"
			} else {
				// Regular user: save locally
				_, localURL, err := storage.SaveFileLocally(coverFile, "cover_"+coverHeader.Filename, input.Artist, input.Album)
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save cover locally"})
					return
				}
				coverURL = localURL
				coverSource = "local"
			}
		}

		// Transaction to ensure atomicity
		tx := db.Begin()

		// 1. Find or Create Artist
		var artist model.Artist
		if err := tx.FirstOrCreate(&artist, model.Artist{Name: input.Artist}).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process artist"})
			return
		}

		// 2. Find or Create Album
		var album model.Album
		albumTitle := input.Album
		if albumTitle == "" {
			albumTitle = "Unknown Album"
		}

		if err := tx.Where("title = ? AND year = ?", albumTitle, releaseDate.Year()).FirstOrCreate(&album, model.Album{Title: albumTitle, Year: releaseDate.Year(), ReleaseDate: releaseDate}).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process album"})
			return
		}
		if album.ID != uuid.Nil {
			var existingAssoc int64
			tx.Table("album_artists").Where("album_id = ? AND artist_id = ?", album.ID, artist.ID).Count(&existingAssoc)
			if existingAssoc == 0 {
				if err := tx.Model(&album).Association("Artists").Append(&artist); err != nil {
					tx.Rollback()
					c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to link album to artist"})
					return
				}
			}
		}

		// Update album cover if new one is provided
		if coverURL != "" {
			album.CoverURL = coverURL
			album.CoverSource = coverSource
			if err := tx.Save(&album).Error; err != nil {
				tx.Rollback()
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update album cover"})
				return
			}
		}

		// 3. Update Song
		song.Title = input.Title
		song.ReleaseDate = releaseDate
		song.TrackNumber = input.TrackNumber
		song.Lyrics = input.Lyrics
		song.AlbumID = &album.ID

		if err := tx.Save(&song).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update song"})
			return
		}

		// 4. Update Artist Association
		if err := tx.Model(&song).Association("Artists").Clear(); err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to clear artist associations"})
			return
		}

		if err := tx.Model(&song).Association("Artists").Append(&artist); err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to associate artist"})
			return
		}

		tx.Commit()

		// Reload song with associations for response
		db.Preload("Album").Preload("Album.Artists").Preload("Artists").First(&song, "id = ?", song.ID)
		c.JSON(http.StatusOK, song)
	}
}

// DeleteSongHandler deletes a song
func DeleteSongHandler(db *gorm.DB, s3Client *s3.S3) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")

		var song model.Song
		if err := db.First(&song, "id = ?", id).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Song not found"})
			return
		}

		if err := db.Delete(&song).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete song"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Song deleted successfully"})
	}
}
