package service

import (
	"fmt"
	"log"

	"gorm.io/gorm"

	"atoman/internal/model"
)

// RunForumMigrations handles ltree extension setup and adds new columns to
// existing forum tables. It is idempotent and safe to run on every startup.
func RunForumMigrations(db *gorm.DB) error {
	dialect := db.Dialector.Name()
	log.Printf("Running forum migrations (dialect: %s)...", dialect)

	if dialect == "postgres" || dialect == "pgx" {
		// Enable ltree extension
		if err := db.Exec("CREATE EXTENSION IF NOT EXISTS ltree").Error; err != nil {
			log.Printf("WARN: could not create ltree extension: %v", err)
		}

		// Add new columns to forum_topics
		// tags stored as JSON text array (compatible with StringSlice custom type)
		db.Exec(`ALTER TABLE forum_topics ADD COLUMN IF NOT EXISTS tags TEXT DEFAULT '[]'`)
		db.Exec(`ALTER TABLE forum_topics ADD COLUMN IF NOT EXISTS last_reply_at TIMESTAMPTZ`)

		// Add path column as LTREE (requires extension) and floor_number
		db.Exec(`ALTER TABLE forum_replies ADD COLUMN IF NOT EXISTS path LTREE`)
		db.Exec(`ALTER TABLE forum_replies ADD COLUMN IF NOT EXISTS floor_number INT DEFAULT 0`)

		// Create GIST index for ltree path (fast subtree queries)
		db.Exec(`CREATE INDEX IF NOT EXISTS idx_forum_replies_path ON forum_replies USING GIST(path)`)
		db.Exec(`CREATE INDEX IF NOT EXISTS idx_forum_replies_path_btree ON forum_replies USING BTREE(path)`)
	} else {
		// SQLite fallback — path is plain TEXT, GIST not supported

		// Note: SQLite ALTER TABLE ADD COLUMN cannot use IF NOT EXISTS before SQLite 3.25.0
		// Since we want this to be resilient, we just ignore errors on ADD COLUMN for SQLite
		db.Exec(`ALTER TABLE forum_topics ADD COLUMN tags TEXT DEFAULT '[]'`)
		db.Exec(`ALTER TABLE forum_topics ADD COLUMN last_reply_at DATETIME`)
		db.Exec(`ALTER TABLE forum_replies ADD COLUMN path TEXT`)
		db.Exec(`ALTER TABLE forum_replies ADD COLUMN floor_number INTEGER DEFAULT 0`)
		db.Exec(`CREATE INDEX IF NOT EXISTS idx_forum_replies_path ON forum_replies (path)`)
	}

	// AutoMigrate new tables that use standard GORM-safe types
	if err := db.AutoMigrate(
		&model.ForumBookmark{},
		&model.ForumDraft{},
		&model.ActivityLog{},
	); err != nil {
		return fmt.Errorf("forum AutoMigrate failed: %w", err)
	}

	// Backfill path and floor_number for existing replies (idempotent)
	runForumPathBackfill(db, dialect)

	log.Println("Forum migrations completed")
	return nil
}

// runForumPathBackfill assigns floor_number and path to existing replies
// that were created before the path column was added.
func runForumPathBackfill(db *gorm.DB, dialect string) {
	// Step 1: Assign floor_number to replies that don't have one yet
	// Uses a correlated subcount: floor = number of earlier replies in same topic
	if dialect == "postgres" || dialect == "pgx" {
		db.Exec(`
			UPDATE forum_replies r
			SET floor_number = sub.rn
			FROM (
				SELECT id, ROW_NUMBER() OVER (PARTITION BY topic_id ORDER BY created_at ASC) AS rn
				FROM forum_replies
				WHERE deleted_at IS NULL
			) sub
			WHERE r.id = sub.id AND r.floor_number = 0
		`)
	} else {
		// SQLite: use correlated subquery
		db.Exec(`
			UPDATE forum_replies
			SET floor_number = (
				SELECT COUNT(*) FROM forum_replies r2
				WHERE r2.topic_id = forum_replies.topic_id
				  AND r2.created_at <= forum_replies.created_at
				  AND r2.deleted_at IS NULL
			)
			WHERE floor_number = 0 OR floor_number IS NULL
		`)
	}

	// Step 2: Set path for root replies (no parent)
	if dialect == "postgres" || dialect == "pgx" {
		db.Exec(`
			UPDATE forum_replies
			SET path = (to_char(floor_number, 'FM000000'))::ltree
			WHERE parent_reply_id IS NULL AND path IS NULL
		`)
	} else {
		db.Exec(`
			UPDATE forum_replies
			SET path = printf('%06d', floor_number)
			WHERE parent_reply_id IS NULL AND (path IS NULL OR path = '')
		`)
	}

	// Step 3: Iterative resolution for nested replies (up to 15 levels)
	for i := 0; i < 15; i++ {
		var res *gorm.DB
		if dialect == "postgres" || dialect == "pgx" {
			res = db.Exec(`
				UPDATE forum_replies r
				SET path = (p.path::text || '.' || to_char(r.floor_number, 'FM000000'))::ltree
				FROM forum_replies p
				WHERE r.parent_reply_id = p.id
				  AND p.path IS NOT NULL
				  AND r.path IS NULL
			`)
		} else {
			res = db.Exec(`
				UPDATE forum_replies
				SET path = (
					SELECT p.path || '.' || printf('%06d', forum_replies.floor_number)
					FROM forum_replies p
					WHERE p.id = forum_replies.parent_reply_id
					  AND p.path IS NOT NULL AND p.path != ''
				)
				WHERE parent_reply_id IS NOT NULL AND (path IS NULL OR path = '')
			`)
		}
		if res.RowsAffected == 0 {
			break
		}
	}
}
