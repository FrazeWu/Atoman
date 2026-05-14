package main

import (
	"log"
	"os"
	"strings"

	"github.com/google/uuid"

	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq" // PostgreSQL array type support
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"atoman/internal/collab"
	"atoman/internal/handlers"
	"atoman/internal/middleware"
	"atoman/internal/model"
	"atoman/internal/service"
	"atoman/internal/storage"
)

func ensureSoftDeleteColumns(db *gorm.DB) {
	softDeleteModels := []interface{}{
		&model.User{},
		&model.Artist{},
		&model.Album{},
		&model.Song{},
		&model.Channel{},
		&model.Collection{},
		&model.Post{},
		&model.Comment{},
		&model.FeedSource{},
		&model.FeedItem{},
		&model.AlbumCorrection{},
		&model.SongCorrection{},
	}

	for _, m := range softDeleteModels {
		if !db.Migrator().HasTable(m) {
			continue
		}
		if !db.Migrator().HasColumn(m, "deleted_at") {
			if err := db.Migrator().AddColumn(m, "DeletedAt"); err != nil {
				log.Printf("WARN: failed to add deleted_at for %T: %v", m, err)
			}
		}
	}
}

func backfillBlogChannelFields(db *gorm.DB) {
	var channels []model.Channel
	if err := db.Find(&channels).Error; err != nil {
		log.Printf("WARN: failed to load channels for backfill: %v", err)
		return
	}

	for _, channel := range channels {
		updates := map[string]interface{}{}
		if strings.TrimSpace(channel.Slug) == "" {
			base := strings.TrimSpace(channel.Name)
			if base == "" {
				base = "channel"
			}
			candidate := handlersSlugify(base)
			for {
				var count int64
				query := db.Model(&model.Channel{}).Where("slug = ?", candidate).Where("id <> ?", channel.ID)
				if err := query.Count(&count).Error; err != nil {
					log.Printf("WARN: failed to check slug uniqueness for channel %s: %v", channel.ID, err)
					break
				}
				if count == 0 {
					updates["slug"] = candidate
					break
				}
				candidate = candidate + "-" + uuid.NewString()[:8]
			}
		}
		if len(updates) > 0 {
			if err := db.Model(&model.Channel{}).Where("id = ?", channel.ID).Updates(updates).Error; err != nil {
				log.Printf("WARN: failed to backfill channel %s: %v", channel.ID, err)
			}
		}
	}

	var posts []model.Post
	if err := db.Preload("Collections").Find(&posts).Error; err != nil {
		log.Printf("WARN: failed to load posts for channel backfill: %v", err)
		return
	}

	for _, post := range posts {
		if post.ChannelID != nil {
			continue
		}
		if len(post.Collections) == 0 {
			continue
		}
		channelID := post.Collections[0].ChannelID
		if err := db.Model(&model.Post{}).Where("id = ?", post.ID).Update("channel_id", channelID).Error; err != nil {
			log.Printf("WARN: failed to backfill post %s channel_id: %v", post.ID, err)
		}
	}
}

func handlersSlugify(value string) string {
	slug := strings.ToLower(strings.TrimSpace(value))
	replacer := func(r rune) rune {
		switch {
		case r >= 'a' && r <= 'z':
			return r
		case r >= '0' && r <= '9':
			return r
		case r >= '一' && r <= '龥':
			return r
		default:
			return '-'
		}
	}
	mapped := strings.Map(replacer, slug)
	mapped = strings.Trim(mapped, "-")
	mapped = strings.Join(strings.FieldsFunc(mapped, func(r rune) bool { return r == '-' }), "-")
	if mapped == "" {
		return "channel"
	}
	return mapped
}

func main() {
	log.Println("Starting Atoman Backend Server...")

	if err := godotenv.Load(".env.dev"); err == nil {
		log.Println("Loaded .env.dev")
	} else if err := godotenv.Load(".env"); err == nil {
		log.Println("Loaded .env")
	} else {
		log.Println("No .env file found, using system environment variables")
	}

	if os.Getenv("GIN_MODE") == "release" {
		gin.SetMode(gin.ReleaseMode)
		log.Println("Running in production mode")
	} else {
		log.Println("Running in development mode")
	}

	if os.Getenv("JWT_SECRET") == "" {
		log.Fatal("JWT_SECRET environment variable is required")
	}

	dbType := os.Getenv("DATABASE_TYPE")
	if dbType == "" {
		log.Fatal("DATABASE_TYPE environment variable is required (postgres or sqlite)")
	}

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL environment variable is required")
	}

	log.Printf("Connecting to %s database: %s", dbType, dbURL)

	var dialector gorm.Dialector
	switch dbType {
	case "postgres", "postgresql":
		dialector = postgres.Open(dbURL)
	case "sqlite":
		dialector = sqlite.Open(dbURL)
	default:
		log.Fatal("Unsupported DATABASE_TYPE: ", dbType, " (expected: postgres or sqlite)")
	}

	db, err := gorm.Open(dialector, &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database: ", err)
	}
	log.Println("Database connected successfully")

	// Only run migrations and backfills in dev mode or manually
	if os.Getenv("GIN_MODE") != "release" {
		log.Println("Running database migrations...")
		if err := db.AutoMigrate(
			&model.User{},
			&model.UserSettings{},
			&model.Artist{},
			&model.Album{},
			&model.Song{},
			&model.SongCorrection{},
			&model.AlbumCorrection{},
			&model.Channel{},
			&model.Collection{},
			&model.Post{},
			&model.BlogDraft{},
			&model.Comment{},
			&model.Like{},
			&model.Bookmark{},
			&model.BookmarkFolder{},
			&model.FeedSource{},
			&model.Subscription{},
			&model.FeedItem{},
			&model.FeedItemRead{},
			&model.FeedItemStar{},
			&model.ReadingListItem{},
			&model.SubscriptionGroup{},
				&model.ForumCategory{},
			&model.ForumTopic{},
			&model.ForumReply{},
			&model.ForumLike{},
			&model.ForumBookmark{},
			&model.ForumDraft{},
			&model.ActivityLog{},
			&model.Debate{},
			&model.Argument{},
			&model.DebateVote{},
			&model.VoteHistory{},
			&model.DebateConcludeVote{},
			&model.EmailVerificationCode{},
			&model.TimelineEvent{},
			&model.TimelinePerson{},
			&model.PersonLocation{},
			// Revision / wiki system
			&model.Revision{},
			&model.EditConflict{},
			&model.ContentProtection{},
			&model.Discussion{},
			// Music wiki extensions
			&model.ArtistAlias{},
			&model.ArtistMerge{},
			&model.LyricAnnotation{},
		); err != nil {
			log.Fatal("Failed to run migrations: ", err)
		}
		log.Println("Database migrations completed")

		log.Println("Running blog channel field backfill...")
		backfillBlogChannelFields(db)

		ensureSoftDeleteColumns(db)

		// Run forum-specific migrations (ltree extension, new columns, backfill)
		if err := service.RunForumMigrations(db); err != nil {
			log.Printf("WARN: forum migrations had errors: %v", err)
		}
	} else {
		log.Println("Skipping automatic database migrations in production mode.")
	}

	// Initialize email service (without Redis)
	emailService := service.NewEmailServiceWithoutRedis(db)
	log.Println("Email service initialized (Redis disabled)")

	var s3Client *s3.S3
	if os.Getenv("STORAGE_TYPE") == "local" {
		log.Println("Storage mode: local (S3 disabled)")
	} else {
		var err error
		s3Client, err = storage.InitS3Client()
		if err != nil {
			log.Println("WARN: Failed to create S3 client:", err)
			log.Println("S3 storage disabled, falling back to local storage")
			s3Client = nil
		} else if err := storage.ValidateS3Connection(s3Client); err != nil {
			log.Println("WARN: Failed to validate S3 connection:", err)
			log.Println("S3 storage disabled, falling back to local storage")
			s3Client = nil
		} else {
			log.Println("S3 storage initialized")
		}
	}

	log.Println("Starting background RSS cron worker...")
	service.StartRSSCron(db)

	log.Println("Initializing Casbin Enforcer...")
	if err := middleware.InitCasbin(db); err != nil {
		log.Fatal("Failed to initialize Casbin: ", err)
	}

	r := gin.Default()

	// Configure allowed origins based on environment
	allowedOrigins := []string{
		"http://localhost:5173",
		"http://localhost:3000",
		"http://127.0.0.1:5173",
		"http://127.0.0.1:3000",
	}
	if env := os.Getenv("ENV"); env == "production" {
		// Add production domains from environment variable
		if prodOrigins := os.Getenv("ALLOWED_ORIGINS"); prodOrigins != "" {
			allowedOrigins = append(allowedOrigins, strings.Split(prodOrigins, ",")...)
		}
	}

	r.Use(func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")
		isAllowed := false
		for _, allowed := range allowedOrigins {
			if origin == allowed {
				isAllowed = true
				break
			}
		}

		if isAllowed {
			c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
		} else {
			// For development, allow all origins (but log a warning)
			if os.Getenv("ENV") != "production" {
				c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
			}
		}

		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With, X-Request-ID")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE, PATCH")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

	// Add global Optional Auth and Casbin Middleware
	r.Use(middleware.OptionalAuthMiddleware())
	r.Use(middleware.CasbinMiddleware())

	// Serve static files from uploads directory
	r.Static("/uploads", "./uploads")
	log.Println("Static files served from ./uploads directory")

	handlers.SetupAuthRoutes(r, db, emailService)
	handlers.SetupUserRoutes(r, db)
	handlers.SetupBlogChannelRoutes(r, db)
	handlers.SetupBlogPostRoutes(r, db)
	handlers.SetupBlogInteractionRoutes(r, db)
	handlers.SetupBlogUploadRoutes(r, db, s3Client)
	handlers.SetupFeedRoutes(r, db)
	handlers.SetupSongRoutes(r, db, s3Client)
	handlers.SetupAlbumRoutes(r, db, s3Client)
	handlers.SetupArtistRoutes(r, db)
	handlers.SetupArtistWikiRoutes(r, db)
	handlers.SetupCorrectionRoutes(r, db, s3Client)
	handlers.SetupEntryStatusRoutes(r, db)
	handlers.SetupLyricAnnotationRoutes(r, db)
	handlers.SetupForumRoutes(r, db)
	handlers.SetupDebateRoutes(r, db)
	handlers.SetupTimelineRoutes(r, db)

	// Revision system routes (wiki-style collaboration)
	handlers.SetupRevisionRoutes(r, db)
	handlers.SetupDiscussionRoutes(r, db)
	handlers.SetupProtectionRoutes(r, db)

	// Real-time collaborative editing (Yjs WebSocket relay)
	collabHub := collab.NewHub()
	collabGroup := r.Group("/api/collab")
	collabGroup.Use(middleware.OptionalAuthMiddleware())
	collabGroup.GET("/ws/:roomID", collabHub.ServeWS)

	// Admin routes
	handlers.SetupAdminRoutes(r, db, s3Client)

	// 404 handler - must be last
	r.NoRoute(func(c *gin.Context) {
		c.JSON(404, gin.H{"error": "Not found"})
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server starting on port %s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatal("Failed to start server: ", err)
	}
}
