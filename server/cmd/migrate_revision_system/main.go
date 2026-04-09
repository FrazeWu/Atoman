package main

import (
	"log"
	"os"

	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"atoman/internal/migrations"
)

func main() {
	// Load environment
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	// Connect to database
	dbType := os.Getenv("DB_TYPE")
	if dbType == "" {
		dbType = "sqlite"
	}

	var db *gorm.DB
	var err error

	if dbType == "postgres" {
		dsn := os.Getenv("DATABASE_URL")
		db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	} else {
		dbPath := os.Getenv("DB_PATH")
		if dbPath == "" {
			dbPath = "./atoman.db"
		}
		db, err = gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	}

	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	log.Println("Connected to database successfully")

	// Check if user wants to rollback
	if len(os.Args) > 1 && os.Args[1] == "rollback" {
		log.Println("Rolling back revision system migration...")
		if err := migrations.RollbackRevisionSystem(db); err != nil {
			log.Fatalf("Rollback failed: %v", err)
		}
		log.Println("Rollback completed successfully")
		return
	}

	// Run migration
	log.Println("Starting revision system migration...")
	if err := migrations.MigrateToRevisionSystem(db); err != nil {
		log.Fatalf("Migration failed: %v", err)
	}

	log.Println("✅ Migration completed successfully!")
	log.Println("")
	log.Println("Next steps:")
	log.Println("  1. Restart the server to enable new revision routes")
	log.Println("  2. Test the revision system with: GET /api/albums/:id/revisions")
	log.Println("  3. Create a test edit with: POST /api/albums/:id/edit")
	log.Println("")
	log.Println("To rollback: go run cmd/migrate_revision_system/main.go rollback")
}
