package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"

	"atoman/internal/middleware"
	"atoman/internal/model"
)

func SetupLyricAnnotationRoutes(router *gin.Engine, db *gorm.DB) {
	songs := router.Group("/api/songs/:id")
	{
		songs.GET("/annotations", GetSongAnnotationsHandler(db))
		songs.POST("/annotations", middleware.AuthMiddleware(), CreateSongAnnotationHandler(db))
		songs.PUT("/annotations/:annotationId", middleware.AuthMiddleware(), UpdateSongAnnotationHandler(db))
		songs.DELETE("/annotations/:annotationId", middleware.AuthMiddleware(), DeleteSongAnnotationHandler(db))
	}
}

type AnnotationGroup struct {
	LineNumber  int                    `json:"line_number"`
	Annotations []model.LyricAnnotation `json:"annotations"`
}

func GetSongAnnotationsHandler(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		songID, err := uuid.Parse(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid song ID"})
			return
		}

		var annotations []model.LyricAnnotation
		if err := db.Where("song_id = ?", songID).Preload("User").
			Order("line_number ASC, created_at ASC").Find(&annotations).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch annotations"})
			return
		}

		// Group by line number
		groupMap := map[int]*AnnotationGroup{}
		var groupOrder []int
		for _, a := range annotations {
			if _, ok := groupMap[a.LineNumber]; !ok {
				groupMap[a.LineNumber] = &AnnotationGroup{LineNumber: a.LineNumber}
				groupOrder = append(groupOrder, a.LineNumber)
			}
			groupMap[a.LineNumber].Annotations = append(groupMap[a.LineNumber].Annotations, a)
		}

		groups := make([]AnnotationGroup, 0, len(groupOrder))
		for _, ln := range groupOrder {
			groups = append(groups, *groupMap[ln])
		}

		c.JSON(http.StatusOK, gin.H{"data": groups})
	}
}

func CreateSongAnnotationHandler(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		songID, err := uuid.Parse(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid song ID"})
			return
		}

		var input struct {
			LineNumber int    `json:"line_number" binding:"required"`
			Content    string `json:"content" binding:"required"`
		}
		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		userID := c.GetString("user_id")
		userUUID, _ := uuid.Parse(userID)

		annotation := model.LyricAnnotation{
			SongID:     songID,
			LineNumber: input.LineNumber,
			Content:    input.Content,
			UserID:     userUUID,
		}
		if err := db.Create(&annotation).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create annotation"})
			return
		}

		db.Preload("User").First(&annotation, "id = ?", annotation.ID)
		c.JSON(http.StatusCreated, gin.H{"data": annotation})
	}
}

func UpdateSongAnnotationHandler(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		songID, err := uuid.Parse(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid song ID"})
			return
		}
		annotationID, err := uuid.Parse(c.Param("annotationId"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid annotation ID"})
			return
		}

		userID := c.GetString("user_id")
		userUUID, _ := uuid.Parse(userID)

		var annotation model.LyricAnnotation
		if err := db.Where("id = ? AND song_id = ?", annotationID, songID).First(&annotation).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Annotation not found"})
			return
		}

		if annotation.UserID != userUUID {
			c.JSON(http.StatusForbidden, gin.H{"error": "You can only edit your own annotations"})
			return
		}

		var input struct {
			Content string `json:"content" binding:"required"`
		}
		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if err := db.Model(&annotation).Update("content", input.Content).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update annotation"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"data": annotation})
	}
}

func DeleteSongAnnotationHandler(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		songID, err := uuid.Parse(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid song ID"})
			return
		}
		annotationID, err := uuid.Parse(c.Param("annotationId"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid annotation ID"})
			return
		}

		userID := c.GetString("user_id")
		userUUID, _ := uuid.Parse(userID)
		userRole := c.GetString("role")

		var annotation model.LyricAnnotation
		if err := db.Where("id = ? AND song_id = ?", annotationID, songID).First(&annotation).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Annotation not found"})
			return
		}

		if annotation.UserID != userUUID && userRole != "admin" {
			c.JSON(http.StatusForbidden, gin.H{"error": "Permission denied"})
			return
		}

		if err := db.Delete(&annotation).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete annotation"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Annotation deleted"})
	}
}
