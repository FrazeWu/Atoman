package handlers

import (
	"net/http"

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
		query := db.Order("name ASC")
		if q != "" {
			query = query.Where("LOWER(name) LIKE LOWER(?)", "%"+q+"%")
		}
		var artists []model.Artist
		if err := query.Find(&artists).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch artists"})
			return
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
