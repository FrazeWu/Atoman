package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// SetupUserRoutes configures user-related routes
func SetupUserRoutes(router *gin.Engine, db *gorm.DB) {
	users := router.Group("/api/users")
	{
		// These will be implemented in Wave 2
		users.GET("/:id/profile", func(c *gin.Context) {
			c.JSON(http.StatusNotImplemented, gin.H{"message": "not implemented yet"})
		})
		users.PUT("/me", func(c *gin.Context) {
			c.JSON(http.StatusNotImplemented, gin.H{"message": "not implemented yet"})
		})
		users.POST("/:id/follow", func(c *gin.Context) {
			c.JSON(http.StatusNotImplemented, gin.H{"message": "not implemented yet"})
		})
		users.DELETE("/:id/follow", func(c *gin.Context) {
			c.JSON(http.StatusNotImplemented, gin.H{"message": "not implemented yet"})
		})
		users.GET("/:id/followers", func(c *gin.Context) {
			c.JSON(http.StatusNotImplemented, gin.H{"message": "not implemented yet"})
		})
		users.GET("/:id/following", func(c *gin.Context) {
			c.JSON(http.StatusNotImplemented, gin.H{"message": "not implemented yet"})
		})
	}
}
