package main

import (
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"atoman/internal/handlers"
	"atoman/internal/middleware"
	"atoman/internal/model"
	"atoman/internal/service"
	"atoman/internal/storage"
)

func main() {
	log.Println("Starting Atoman Backend Server...")

	if err := godotenv.Load(".env"); err != nil {
		if err := godotenv.Load(".env.dev"); err != nil {
			log.Println("No .env file found, using system environment variables")
		} else {
			log.Println("Loaded .env.dev")
		}
	} else {
		log.Println("Loaded .env")
	}

	if os.Getenv("GIN_MODE") == "release" {
		gin.SetMode(gin.ReleaseMode)
		log.Println("Running in production mode")
	} else {
		log.Println("Running in development mode")
	}

	dbType := os.Getenv("DATABASE_TYPE")
	if dbType == "" {
		log.Fatal("DATABASE_TYPE environment variable is required (mysql/postgres)")
	}

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
	if dbURL == "" {
		log.Fatal("DATABASE_URL environment variable is required")
	}
	}

	log.Printf("Connecting to %s database: %s", dbType, dbURL)

	var dialector gorm.Dialector
	switch dbType {
	case "mysql":
		dialector = mysql.Open(dbURL)
	case "postgres", "postgresql":
		dialector = postgres.Open(dbURL)
	default:
		log.Fatal("Unsupported DATABASE_TYPE: ", dbType)
	}

	db, err := gorm.Open(dialector, &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database: ", err)
	}
	log.Println("Database connected successfully")

	log.Println("Running database migrations...")
	db.AutoMigrate(
		&model.User{},
		&model.Follow{},
		&model.UserSettings{},
		&model.Channel{},
		&model.Collection{},
		&model.Post{},
		&model.PostCollection{},
		&model.Comment{},
		&model.Like{},
		&model.BookmarkFolder{},
		&model.Bookmark{},
		&model.FeedSource{},
		&model.Subscription{},
		&model.OrbitItem{},
		&model.Notification{},
		&model.Artist{},
		&model.Album{},
		&model.AlbumArtist{},
		&model.Song{},
		&model.SongArtist{},
		&model.AlbumCorrection{},
		&model.SongCorrection{},
	)
	// Manually ensure deleted_at column exists for feed_sources if AutoMigrate fails
	if dbType == "postgres" || dbType == "postgresql" {
		db.Exec("ALTER TABLE feed_sources ADD COLUMN IF NOT EXISTS deleted_at timestamp with time zone")
		db.Exec("CREATE INDEX IF NOT EXISTS idx_feed_sources_deleted_at ON feed_sources(deleted_at)")
	}
	log.Println("Database migrations completed")

	log.Println("Initializing S3 client...")
	s3Client, err := storage.InitS3Client()
	if err != nil {
		log.Fatal("Failed to create S3 client: ", err)
	}
	log.Println("S3 client initialized")

	log.Println("Validating S3 connection...")
	if err := storage.ValidateS3Connection(s3Client); err != nil {
		log.Fatal("Failed to validate S3 connection: ", err)
	}
	log.Println("S3 connection validated")

	log.Println("Starting background RSS cron worker...")
	service.StartRSSCron(db)

	log.Println("Initializing Casbin Enforcer...")
	if err := middleware.InitCasbin(db); err != nil {
		log.Fatal("Failed to initialize Casbin: ", err)
	}

	r := gin.Default()

	r.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

	// Add global Optional Auth and Casbin Middleware
	r.Use(middleware.OptionalAuthMiddleware())
	r.Use(middleware.CasbinMiddleware())

	handlers.SetupAuthRoutes(r, db)
	handlers.SetupUserRoutes(r, db)
	handlers.SetupBlogChannelRoutes(r, db)
	handlers.SetupBlogPostRoutes(r, db)
	handlers.SetupBlogInteractionRoutes(r, db)
	handlers.SetupFeedRoutes(r, db)
	handlers.SetupNotificationRoutes(r, db)
	handlers.SetupSongRoutes(r, db, s3Client)
	handlers.SetupAlbumRoutes(r, db, s3Client)
	handlers.SetupArtistRoutes(r, db)
	handlers.SetupCorrectionRoutes(r, db, s3Client)
	handlers.SetupAdminRoutes(r, db, s3Client)

	r.Static("/uploads", "./uploads")

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server starting on port %s", port)
	r.Run(":" + port)
}
