package main

import (
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"gorm.io/driver/mysql"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"amb/internal/handlers"
	"amb/internal/model"
	"amb/internal/storage"
)

func main() {
	log.Println("Starting All Kanye Backend Server...")

	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}
	log.Println("Environment variables loaded")

	// Check environment
	env := os.Getenv("ENV")
	if env == "production" {
		gin.SetMode(gin.ReleaseMode)
		log.Println("Running in production mode")
	} else {
		log.Println("Running in development mode")
	}

	// Database connection
	dbType := os.Getenv("DATABASE_TYPE")
	if dbType == "" {
		dbType = "sqlite"
	}

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		if dbType == "sqlite" {
			dbURL = "./database.sqlite"
		} else {
			log.Fatal("DATABASE_URL required for", dbType)
		}
	}

	log.Printf("Connecting to %s database: %s", dbType, dbURL)

	var dialector gorm.Dialector
	switch dbType {
	case "sqlite":
		dialector = sqlite.Open(dbURL)
	case "mysql":
		dialector = mysql.Open(dbURL)
	default:
		log.Fatal("Unsupported DATABASE_TYPE:", dbType)
	}

	db, err := gorm.Open(dialector, &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	log.Println("Database connected successfully")

	// Auto migrate
	log.Println("Skipping database migrations (tables exist)")
	db.AutoMigrate(
		&model.User{},
		&model.Artist{},
		&model.Album{},
		&model.AlbumArtist{},
		&model.Song{},
		&model.SongArtist{},
		&model.AlbumCorrection{},
		&model.SongCorrection{},
	)
	log.Println("Database migrations completed")

	// Initialize S3 client
	log.Println("Initializing S3 client...")
	s3Client, err := storage.InitS3Client()
	if err != nil {
		log.Fatal("Failed to create S3 client:", err)
	}
	log.Println("S3 client initialized")

	// Validate S3 connection
	log.Println("Validating S3 connection...")
	if err := storage.ValidateS3Connection(s3Client); err != nil {
		log.Fatal("Failed to validate S3 connection:", err)
	}
	log.Println("S3 connection validated")

	// Gin router
	log.Println("Setting up routes...")
	r := gin.Default()

	// CORS Middleware
	r.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*") // In production, replace * with specific origin
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

	// Setup routes
	handlers.SetupAuthRoutes(r, db)
	handlers.SetupSongRoutes(r, db, s3Client)
	handlers.SetupAlbumRoutes(r, db, s3Client)
	handlers.SetupCorrectionRoutes(r, db, s3Client)
	handlers.SetupAdminRoutes(r, db, s3Client)
	log.Println("Routes configured")

	r.Static("/uploads", "./uploads")
	log.Println("Static file serving configured for /uploads")

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("🚀 Server starting on port %s", port)
	log.Printf("API endpoints available at http://localhost:%s", port)
	r.Run(":" + port)
}
