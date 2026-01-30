package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
	"gorm.io/driver/mysql"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// Models matching current schema (2026-01-18)
type User struct {
	ID        uint      `gorm:"primaryKey;column:id"`
	Username  string    `gorm:"unique;not null;column:username"`
	Email     string    `gorm:"unique;not null;column:email"`
	Password  string    `gorm:"not null;column:password"`
	Role      string    `gorm:"default:'user';column:role"`
	CreatedAt time.Time `gorm:"column:created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at"`
}

func (User) TableName() string {
	return "Users"
}

type Artist struct {
	ID        uint      `gorm:"primaryKey"`
	Name      string    `gorm:"unique;not null"`
	Bio       string    `gorm:"type:text"`
	ImageURL  string    `json:"image_url"`
	CreatedAt time.Time `gorm:"column:created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at"`
}

func (Artist) TableName() string {
	return "Artists"
}

type Album struct {
	ID          uint   `gorm:"primaryKey"`
	Title       string `gorm:"not null"`
	Year        int
	ReleaseDate time.Time `gorm:"type:date"`
	CoverURL    string
	CoverSource string `gorm:"default:'local'"`
	Status      string `gorm:"default:'pending'"`
	UploadedBy  *uint
	CreatedAt   time.Time `gorm:"column:created_at"`
	UpdatedAt   time.Time `gorm:"column:updated_at"`
}

func (Album) TableName() string {
	return "Albums"
}

type AlbumArtist struct {
	AlbumID   uint      `gorm:"primaryKey"`
	ArtistID  uint      `gorm:"primaryKey"`
	Role      string    `gorm:"default:'primary'"`
	CreatedAt time.Time `gorm:"column:created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at"`
}

func (AlbumArtist) TableName() string {
	return "album_artists"
}

type Song struct {
	ID          uint      `gorm:"primaryKey"`
	Title       string    `gorm:"not null"`
	ReleaseDate time.Time `gorm:"type:date"`
	TrackNumber int
	Lyrics      string `gorm:"type:text"`
	AudioURL    string `gorm:"not null"`
	AudioSource string `gorm:"default:'local'"`
	CoverURL    string
	CoverSource string `gorm:"default:'local'"`
	BatchID     string `gorm:"index"`
	Status      string `gorm:"default:'pending'"`
	AlbumID     *uint
	UploadedBy  *uint
	CreatedAt   time.Time `gorm:"column:created_at"`
	UpdatedAt   time.Time `gorm:"column:updated_at"`
}

func (Song) TableName() string {
	return "Songs"
}

type SongArtist struct {
	SongID    uint      `gorm:"primaryKey"`
	ArtistID  uint      `gorm:"primaryKey"`
	Role      string    `gorm:"default:'primary'"`
	CreatedAt time.Time `gorm:"column:created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at"`
}

func (SongArtist) TableName() string {
	return "song_artists"
}

type AlbumCorrection struct {
	ID                   uint `gorm:"primaryKey"`
	AlbumID              uint `gorm:"not null"`
	UserID               *uint
	Status               string `gorm:"default:'pending'"`
	OriginalTitle        string
	OriginalCoverURL     string     `gorm:"type:text"`
	OriginalReleaseDate  *time.Time `gorm:"type:date"`
	OriginalArtistIDs    string     `gorm:"type:text"`
	CorrectedTitle       string
	CorrectedCoverURL    string     `gorm:"type:text"`
	CorrectedCoverSource string     `gorm:"default:'local'"`
	CorrectedReleaseDate *time.Time `gorm:"type:date"`
	CorrectedArtistIDs   string     `gorm:"type:text"`
	Reason               string     `gorm:"type:text"`
	CreatedAt            time.Time  `gorm:"column:created_at"`
	ApprovedAt           *time.Time
	ApprovedBy           *uint
	RejectedAt           *time.Time
	RejectedBy           *uint
}

func (AlbumCorrection) TableName() string {
	return "album_corrections"
}

type SongCorrection struct {
	ID             uint `gorm:"primaryKey"`
	SongID         uint `gorm:"not null"`
	UserID         *uint
	Status         string `gorm:"default:'pending'"`
	FieldName      string `gorm:"not null"`
	CurrentValue   string `gorm:"type:text"`
	CorrectedValue string `gorm:"type:text;not null"`
	Reason         string `gorm:"type:text"`
	CreatedAt      time.Time
	ApprovedAt     *time.Time
	ApprovedBy     *uint
	RejectedAt     *time.Time
	RejectedBy     *uint
}

func (SongCorrection) TableName() string {
	return "song_corrections"
}

func main() {
	log.Println("========================================")
	log.Println("SQLite → MySQL Migration Tool v2.0")
	log.Println("Updated: 2026-01-18")
	log.Println("========================================")

	// Load environment
	if err := godotenv.Load(); err != nil {
		if err := godotenv.Load("../../.env"); err != nil {
			log.Println("⚠️  No .env file found, using environment variables")
		}
	}

	// Get SQLite path
	sqlitePath := os.Getenv("SQLITE_PATH")
	if sqlitePath == "" {
		sqlitePath = "../../database.sqlite"
	}

	// Build MySQL DSN
	mysqlDSN := os.Getenv("MYSQL_DSN")
	if mysqlDSN == "" {
		dbHost := os.Getenv("DB_HOST")
		dbPort := os.Getenv("DB_PORT")
		dbUser := os.Getenv("DB_USER")
		dbPassword := os.Getenv("DB_PASSWORD")
		dbName := os.Getenv("DB_NAME")

		if dbHost == "" || dbUser == "" || dbPassword == "" || dbName == "" {
			log.Fatal("❌ MySQL connection details not provided")
			log.Fatal("   Set DB_HOST, DB_USER, DB_PASSWORD, DB_NAME or MYSQL_DSN")
		}

		if dbPort == "" {
			dbPort = "3306"
		}

		mysqlDSN = fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
			dbUser, dbPassword, dbHost, dbPort, dbName)
	}

	// Connect to SQLite
	log.Printf("📂 Connecting to SQLite: %s", sqlitePath)
	sqliteDB, err := gorm.Open(sqlite.Open(sqlitePath), &gorm.Config{})
	if err != nil {
		log.Fatal("❌ Failed to connect to SQLite:", err)
	}
	log.Println("✅ SQLite connected")

	// Connect to MySQL
	log.Printf("🔌 Connecting to MySQL...")
	mysqlDB, err := gorm.Open(mysql.Open(mysqlDSN), &gorm.Config{})
	if err != nil {
		log.Fatal("❌ Failed to connect to MySQL:", err)
	}
	log.Println("✅ MySQL connected")

	// Create schema
	log.Println("🏗️  Creating MySQL schema...")
	if err := mysqlDB.AutoMigrate(
		&User{},
		&Artist{},
		&Album{},
		&AlbumArtist{},
		&Song{},
		&SongArtist{},
		&AlbumCorrection{},
		&SongCorrection{},
	); err != nil {
		log.Fatal("❌ Failed to create MySQL schema:", err)
	}
	log.Println("✅ MySQL schema created")

	// Migrate data
	totalMigrated := 0

	// Users
	log.Println("\n👤 Migrating Users...")
	var users []User
	if err := sqliteDB.Find(&users).Error; err != nil {
		log.Fatal("❌ Failed to read users from SQLite:", err)
	}
	if len(users) > 0 {
		if err := mysqlDB.Create(&users).Error; err != nil {
			log.Printf("⚠️  Warning: Failed to migrate some users: %v", err)
		} else {
			log.Printf("✅ Migrated %d users", len(users))
			totalMigrated += len(users)
		}
	} else {
		log.Println("   No users to migrate")
	}

	// Artists
	log.Println("\n🎤 Migrating Artists...")
	var artists []Artist
	if err := sqliteDB.Find(&artists).Error; err != nil {
		log.Fatal("❌ Failed to read artists from SQLite:", err)
	}
	if len(artists) > 0 {
		if err := mysqlDB.Create(&artists).Error; err != nil {
			log.Printf("⚠️  Warning: Failed to migrate some artists: %v", err)
		} else {
			log.Printf("✅ Migrated %d artists", len(artists))
			totalMigrated += len(artists)
		}
	} else {
		log.Println("   No artists to migrate")
	}

	// Albums
	log.Println("\n💿 Migrating Albums...")
	var albums []Album
	if err := sqliteDB.Find(&albums).Error; err != nil {
		log.Fatal("❌ Failed to read albums from SQLite:", err)
	}
	if len(albums) > 0 {
		// Fix invalid dates before migration
		for i := range albums {
			if albums[i].ReleaseDate.IsZero() || albums[i].ReleaseDate.Year() < 1900 {
				albums[i].ReleaseDate = time.Time{}
			}
		}
		if err := mysqlDB.Create(&albums).Error; err != nil {
			log.Printf("⚠️  Warning: Failed to migrate some albums: %v", err)
		} else {
			log.Printf("✅ Migrated %d albums", len(albums))
			totalMigrated += len(albums)
		}
	} else {
		log.Println("   No albums to migrate")
	}

	// Album-Artist relationships
	log.Println("\n🔗 Migrating Album-Artist relationships...")
	var albumArtists []AlbumArtist
	if err := sqliteDB.Table("album_artists").Find(&albumArtists).Error; err != nil {
		log.Printf("⚠️  Warning: Failed to read album_artists: %v", err)
	} else if len(albumArtists) > 0 {
		if err := mysqlDB.Table("album_artists").Create(&albumArtists).Error; err != nil {
			log.Printf("⚠️  Warning: Failed to migrate album_artists: %v", err)
		} else {
			log.Printf("✅ Migrated %d album-artist relationships", len(albumArtists))
			totalMigrated += len(albumArtists)
		}
	} else {
		log.Println("   No album-artist relationships to migrate")
	}

	// Songs
	log.Println("\n🎵 Migrating Songs...")
	var songs []Song
	if err := sqliteDB.Find(&songs).Error; err != nil {
		log.Fatal("❌ Failed to read songs from SQLite:", err)
	}
	if len(songs) > 0 {
		if err := mysqlDB.Create(&songs).Error; err != nil {
			log.Printf("⚠️  Warning: Failed to migrate some songs: %v", err)
		} else {
			log.Printf("✅ Migrated %d songs", len(songs))
			totalMigrated += len(songs)
		}
	} else {
		log.Println("   No songs to migrate")
	}

	// Song-Artist relationships
	log.Println("\n🔗 Migrating Song-Artist relationships...")
	var songArtists []SongArtist
	if err := sqliteDB.Table("song_artists").Find(&songArtists).Error; err != nil {
		log.Printf("⚠️  Warning: Failed to read song_artists: %v", err)
	} else if len(songArtists) > 0 {
		if err := mysqlDB.Table("song_artists").Create(&songArtists).Error; err != nil {
			log.Printf("⚠️  Warning: Failed to migrate song_artists: %v", err)
		} else {
			log.Printf("✅ Migrated %d song-artist relationships", len(songArtists))
			totalMigrated += len(songArtists)
		}
	} else {
		log.Println("   No song-artist relationships to migrate")
	}

	// Album Corrections
	log.Println("\n📝 Migrating Album Corrections...")
	var albumCorrections []AlbumCorrection
	if err := sqliteDB.Table("album_corrections").Find(&albumCorrections).Error; err != nil {
		log.Printf("⚠️  Warning: Failed to read album_corrections: %v", err)
	} else if len(albumCorrections) > 0 {
		if err := mysqlDB.Create(&albumCorrections).Error; err != nil {
			log.Printf("⚠️  Warning: Failed to migrate album corrections: %v", err)
		} else {
			log.Printf("✅ Migrated %d album corrections", len(albumCorrections))
			totalMigrated += len(albumCorrections)
		}
	} else {
		log.Println("   No album corrections to migrate")
	}

	// Song Corrections
	log.Println("\n📝 Migrating Song Corrections...")
	var songCorrections []SongCorrection
	if err := sqliteDB.Table("song_corrections").Find(&songCorrections).Error; err != nil {
		log.Printf("⚠️  Warning: Failed to read song_corrections: %v", err)
	} else if len(songCorrections) > 0 {
		if err := mysqlDB.Create(&songCorrections).Error; err != nil {
			log.Printf("⚠️  Warning: Failed to migrate song corrections: %v", err)
		} else {
			log.Printf("✅ Migrated %d song corrections", len(songCorrections))
			totalMigrated += len(songCorrections)
		}
	} else {
		log.Println("   No song corrections to migrate")
	}

	// Summary
	log.Println("\n========================================")
	log.Println("🎉 Migration completed successfully!")
	log.Printf("📊 Total records migrated: %d", totalMigrated)
	log.Println("========================================")
}
