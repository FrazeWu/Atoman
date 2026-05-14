package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/glebarez/sqlite"
	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"atoman/internal/model"
)

// NOTE: Uses aws-sdk-go (v1) instead of v2 because that's what the project uses

var sqliteSourceTables = map[string]string{
	"Users":                    "Users",
	"user_settings":            "user_settings",
	"Artists":                  "Artists",
	"Albums":                   "Albums",
	"Songs":                    "Songs",
	"channels":                 "channels",
	"collections":              "collections",
	"posts":                    "posts",
	"comments":                 "comments",
	"likes":                    "likes",
	"bookmark_folders":         "bookmark_folders",
	"bookmarks":                "bookmarks",
	"feed_sources":             "feed_sources",
	"feed_items":               "feed_items",
	"feed_item_reads":          "feed_item_reads",
	"feed_item_stars":          "feed_item_stars",
	"reading_list_items":       "reading_list_items",
	"subscription_groups":      "subscription_groups",
	"subscriptions":            "subscriptions",
	"forum_categories":         "forum_categories",
	"forum_topics":             "forum_topics",
	"forum_replies":            "forum_replies",
	"forum_likes":              "forum_likes",
	"forum_bookmarks":          "forum_bookmarks",
	"forum_drafts":             "forum_drafts",
	"activity_logs":            "activity_logs",
	"debates":                  "debates",
	"arguments":                "arguments",
	"debate_votes":             "debate_votes",
	"vote_histories":           "vote_histories",
	"debate_conclude_votes":    "debate_conclude_votes",
	"email_verification_codes": "email_verification_codes",
	"timeline_events":          "timeline_events",
	"timeline_persons":         "timeline_persons",
	"person_locations":         "person_locations",
	"revisions":                "revisions",
	"edit_conflicts":           "edit_conflicts",
	"content_protections":      "content_protections",
	"discussions":              "discussions",
	"artist_aliases":           "artist_aliases",
	"artist_merges":            "artist_merges",
	"lyric_annotations":        "lyric_annotations",
	"album_corrections":        "album_corrections",
	"song_corrections":         "song_corrections",
	"album_artists":            "album_artists",
	"song_artists":             "song_artists",
	"post_collections":         "post_collections",
}

var boolColumns = map[string]map[string]bool{
	"Users":                    {"is_active": true},
	"user_settings":            {"private_profile": true},
	"channels":                 {"is_default": true},
	"collections":              {"is_default": true},
	"posts":                    {"allow_comments": true, "pinned": true},
	"blog_drafts":              {"allow_comments": true},
	"forum_topics":             {"pinned": true, "closed": true},
	"arguments":                {"is_concluded": true},
	"timeline_events":          {"is_public": true},
	"timeline_persons":         {"is_public": true},
	"artist_aliases":           {"is_main_name": true},
	"revisions":                {"is_current": true},
	"email_verification_codes": {"used": true},
}

var criticalTables = map[string]bool{
	"Users":               true,
	"Artists":             true,
	"Albums":              true,
	"Songs":               true,
	"feed_sources":        true,
	"subscription_groups": true,
	"subscriptions":       true,
	"feed_items":          true,
}

var associationAllowedColumns = map[string]map[string]bool{
	"album_artists":    {"album_id": true, "artist_id": true},
	"song_artists":     {"song_id": true, "artist_id": true},
	"post_collections": {"post_id": true, "collection_id": true},
}

func normalizeBoolLike(v any) any {
	switch x := v.(type) {
	case bool:
		return x
	case int64:
		return x != 0
	case int:
		return x != 0
	case float64:
		return x != 0
	case []byte:
		s := string(x)
		return s == "1" || strings.EqualFold(s, "true")
	case string:
		if x == "" {
			return false
		}
		if i, err := strconv.ParseInt(x, 10, 64); err == nil {
			return i != 0
		}
		return strings.EqualFold(x, "true")
	default:
		return v
	}
}

func shouldNormalizeToS3(s3Prefix string) bool {
	return strings.TrimSpace(s3Prefix) != ""
}

func normalizeRecord(tableName string, record map[string]any, s3Prefix string) map[string]any {
	if cols, ok := boolColumns[tableName]; ok {
		for col := range cols {
			if v, exists := record[col]; exists {
				record[col] = normalizeBoolLike(v)
			}
		}
	}

	if avatar, ok := record["avatar_url"].(string); ok && strings.HasPrefix(avatar, "/uploads/") && shouldNormalizeToS3(s3Prefix) {
		record["avatar_url"] = s3Prefix + "/" + strings.TrimPrefix(avatar, "/uploads/")
	}
	if audio, ok := record["audio_url"].(string); ok && strings.HasPrefix(audio, "/uploads/") && shouldNormalizeToS3(s3Prefix) {
		record["audio_url"] = s3Prefix + "/" + strings.TrimPrefix(audio, "/uploads/")
		if _, hasSource := record["audio_source"]; hasSource {
			record["audio_source"] = "s3"
		}
	}
	if cover, ok := record["cover_url"].(string); ok && strings.HasPrefix(cover, "/uploads/") && shouldNormalizeToS3(s3Prefix) {
		record["cover_url"] = s3Prefix + "/" + strings.TrimPrefix(cover, "/uploads/")
		if _, hasSource := record["cover_source"]; hasSource {
			record["cover_source"] = "s3"
		}
	}
	return record
}

func filterAssociationRecord(tableName string, record map[string]any) map[string]any {
	allowed := associationAllowedColumns[tableName]
	filtered := make(map[string]any, len(allowed))
	for k := range allowed {
		if v, ok := record[k]; ok {
			filtered[k] = v
		}
	}
	return filtered
}

func main() {
	log.Println("Starting Migration: SQLite + Local Files -> PostgreSQL + MinIO")

	if err := godotenv.Load(".env.dev"); err != nil {
		log.Fatalf("Failed to load .env.dev: %v", err)
	}

	sqliteDB, err := gorm.Open(sqlite.Open("dev.sqlite"), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to SQLite: %v", err)
	}
	log.Println("Connected to SQLite (dev.sqlite)")

	pgURL := os.Getenv("DATABASE_URL")
	if pgURL == "" {
		log.Fatal("DATABASE_URL not found in .env.dev")
	}
	pgDB, err := gorm.Open(postgres.Open(pgURL), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to PostgreSQL: %v", err)
	}
	log.Println("Connected to PostgreSQL")

	s3Endpoint := os.Getenv("S3_ENDPOINT")
	s3Region := os.Getenv("S3_REGION")
	if s3Region == "" {
		s3Region = "us-east-1"
	}

	sess, err := session.NewSession(&aws.Config{
		Region:           aws.String(s3Region),
		Endpoint:         aws.String(s3Endpoint),
		Credentials:      credentials.NewStaticCredentials(os.Getenv("AWS_ACCESS_KEY_ID"), os.Getenv("AWS_SECRET_ACCESS_KEY"), ""),
		S3ForcePathStyle: aws.Bool(true),
	})
	if err != nil {
		log.Fatalf("Unable to load AWS session: %v", err)
	}
	s3Client := s3.New(sess)
	bucketName := os.Getenv("S3_BUCKET")
	if _, err := s3Client.HeadBucket(&s3.HeadBucketInput{Bucket: aws.String(bucketName)}); err != nil {
		log.Printf("Bucket %s missing, creating it...", bucketName)
		if _, createErr := s3Client.CreateBucket(&s3.CreateBucketInput{Bucket: aws.String(bucketName)}); createErr != nil {
			log.Fatalf("Failed to create bucket %s: %v", bucketName, createErr)
		}
	}
	log.Printf("Connected to MinIO (Bucket: %s)", bucketName)

	log.Println("--- Starting File Uploads ---")
	uploadCount := 0
	uploadsDir := "uploads"

	err = filepath.Walk(uploadsDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() || strings.HasPrefix(info.Name(), ".") {
			return nil
		}

		relPath, err := filepath.Rel(uploadsDir, path)
		if err != nil {
			return err
		}
		s3Key := filepath.ToSlash(relPath)

		file, err := os.Open(path)
		if err != nil {
			log.Printf("Failed to open file %s: %v", path, err)
			return nil
		}
		defer file.Close()

		_, err = s3Client.PutObject(&s3.PutObjectInput{
			Bucket: aws.String(bucketName),
			Key:    aws.String(s3Key),
			Body:   file,
			ACL:    aws.String("public-read"),
		})
		if err != nil {
			log.Printf("Failed to upload %s: %v", s3Key, err)
			return nil
		}

		uploadCount++
		if uploadCount%10 == 0 {
			log.Printf("Uploaded %d files...", uploadCount)
		}
		return nil
	})
	if err != nil && !os.IsNotExist(err) {
		log.Printf("Error walking uploads directory: %v", err)
	} else if os.IsNotExist(err) {
		log.Println("No uploads directory found, skipping file migration.")
	} else {
		log.Printf("Successfully uploaded %d files to MinIO.", uploadCount)
	}

	log.Println("--- Starting Database Migration ---")
	log.Println("AutoMigrating PostgreSQL schema...")
	if err := pgDB.AutoMigrate(
		&model.User{}, &model.UserSettings{}, &model.Artist{}, &model.Album{},
		&model.Song{}, &model.SongCorrection{}, &model.AlbumCorrection{},
		&model.Channel{}, &model.Collection{}, &model.Post{}, &model.BlogDraft{},
		&model.Comment{}, &model.Like{}, &model.Bookmark{}, &model.BookmarkFolder{},
		&model.FeedSource{}, &model.Subscription{}, &model.FeedItem{},
		&model.FeedItemRead{}, &model.FeedItemStar{}, &model.ReadingListItem{},
		&model.SubscriptionGroup{}, &model.ForumCategory{},
		&model.ForumTopic{}, &model.ForumReply{}, &model.ForumLike{},
		&model.ForumBookmark{}, &model.ForumDraft{}, &model.ActivityLog{},
		&model.Debate{}, &model.Argument{}, &model.DebateVote{}, &model.VoteHistory{},
		&model.DebateConcludeVote{}, &model.EmailVerificationCode{},
		&model.TimelineEvent{}, &model.TimelinePerson{}, &model.PersonLocation{},
		&model.Revision{}, &model.EditConflict{}, &model.ContentProtection{},
		&model.Discussion{}, &model.ArtistAlias{}, &model.ArtistMerge{},
		&model.LyricAnnotation{},
	); err != nil {
		log.Fatalf("Failed to automigrate PG schema: %v", err)
	}

	migrateTable := func(targetTableName string) error {
		sourceTableName := sqliteSourceTables[targetTableName]
		if sourceTableName == "" {
			sourceTableName = targetTableName
		}
		log.Printf("Migrating table: %s <- %s", targetTableName, sourceTableName)

		var records []map[string]any
		if err := sqliteDB.Table(sourceTableName).Find(&records).Error; err != nil {
			if criticalTables[targetTableName] {
				return fmt.Errorf("load %s: %w", sourceTableName, err)
			}
			log.Printf("  Skipping %s: %v", sourceTableName, err)
			return nil
		}

		if len(records) == 0 {
			log.Printf("  0 records in %s", sourceTableName)
			return nil
		}

		s3Prefix := os.Getenv("S3_URL_PREFIX")
		normalized := make([]map[string]any, 0, len(records))
		for _, record := range records {
			record = normalizeRecord(sourceTableName, record, s3Prefix)
			if _, isAssoc := associationAllowedColumns[targetTableName]; isAssoc {
				record = filterAssociationRecord(targetTableName, record)
			}
			normalized = append(normalized, record)
		}

		result := pgDB.Table(targetTableName).CreateInBatches(normalized, 100)
		if result.Error != nil {
			if criticalTables[targetTableName] {
				return fmt.Errorf("insert %s: %w", targetTableName, result.Error)
			}
			log.Printf("  Error migrating %s: %v", targetTableName, result.Error)
			return nil
		}

		var targetCount int64
		if err := pgDB.Table(targetTableName).Count(&targetCount).Error; err != nil {
			if criticalTables[targetTableName] {
				return fmt.Errorf("verify %s: %w", targetTableName, err)
			}
			log.Printf("  Failed to verify %s count: %v", targetTableName, err)
		} else {
			log.Printf("  Successfully migrated %d records to %s (target count: %d)", result.RowsAffected, targetTableName, targetCount)
		}
		return nil
	}

	tables := []string{
		"Users", "user_settings",
		"Artists", "Albums", "Songs",
		"channels", "collections", "posts",
		"bookmark_folders", "comments", "likes", "bookmarks",
		"feed_sources", "subscription_groups", "feed_items", "subscriptions", "feed_item_reads", "feed_item_stars", "reading_list_items",
		"forum_categories", "forum_topics", "forum_replies", "forum_likes", "forum_bookmarks", "forum_drafts", "activity_logs",
		"debates", "arguments", "debate_votes", "vote_histories", "debate_conclude_votes",
		"timeline_events", "timeline_persons", "person_locations",
		"revisions", "edit_conflicts", "content_protections", "discussions",
		"artist_aliases", "artist_merges", "lyric_annotations",
		"album_corrections", "song_corrections", "email_verification_codes",
	}

	for _, table := range tables {
		if err := migrateTable(table); err != nil {
			log.Fatalf("Critical migration failed for %s: %v", table, err)
		}
	}

	assocTables := []string{"album_artists", "song_artists", "post_collections"}
	for _, table := range assocTables {
		if err := migrateTable(table); err != nil {
			log.Fatalf("Association migration failed for %s: %v", table, err)
		}
	}

	log.Println("--- Migration Complete! ---")
}
