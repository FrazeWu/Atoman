package handlers

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"atoman/internal/middleware"
	"atoman/internal/model"
)

type ArtistInput struct {
	Name string `json:"name" binding:"required"`
	Bio  string `json:"bio"`
}

func SetupArtistRoutes(router *gin.Engine, db *gorm.DB) {
	artists := router.Group("/api/artists")
	{
		artists.GET("", GetArtistsHandler(db))
		artists.POST("", middleware.AuthMiddleware(), CreateArtistHandler(db))
	}
}

func GetArtistsHandler(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		q := c.Query("q")
		var artists []model.Artist
		if q != "" {
			like := "%" + strings.ToLower(q) + "%"
			if err := db.Raw(`SELECT DISTINCT "Artists".* FROM "Artists"
				LEFT JOIN artist_aliases ON artist_aliases.artist_id = "Artists".id
				WHERE LOWER("Artists".name) LIKE ? OR LOWER(artist_aliases.alias) LIKE ?
				ORDER BY "Artists".name ASC`, like, like).Scan(&artists).Error; err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch artists"})
				return
			}
		} else {
			if err := db.Order("name ASC").Find(&artists).Error; err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch artists"})
				return
			}
		}
		c.JSON(http.StatusOK, artists)
	}
}

func CreateArtistHandler(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var input ArtistInput
		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		var existingArtist model.Artist
		result := db.Where("name = ?", input.Name).First(&existingArtist)
		if result.Error == nil {
			c.JSON(http.StatusConflict, gin.H{"error": "Artist already exists", "id": existingArtist.ID})
			return
		}

		artist := model.Artist{
			Name: input.Name,
			Bio:  input.Bio,
		}

		if err := db.Create(&artist).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create artist"})
			return
		}

		c.JSON(http.StatusCreated, artist)
	}
}
