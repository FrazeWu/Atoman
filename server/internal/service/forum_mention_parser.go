package service

import (
	"regexp"
	"strings"

	"gorm.io/gorm"

	"atoman/internal/model"
)

var (
	// codeBlockRe strips fenced code blocks and inline code to avoid false-positive @mentions
	codeBlockRe = regexp.MustCompile("(?s)```[\\s\\S]*?```|`[^`]+`")

	// mentionRe matches @username patterns (2–32 chars: letters, digits, underscores, hyphens)
	mentionRe = regexp.MustCompile(`@([A-Za-z0-9_-]{2,32})`)
)

// ParseMentions extracts @username patterns from content, looks them up in the DB,
// and returns the matched User records. Code blocks are stripped first to avoid
// mentioning users inside code examples.
func ParseMentions(db *gorm.DB, content string) ([]model.User, error) {
	// Strip code blocks and inline code
	stripped := codeBlockRe.ReplaceAllString(content, "")

	// Extract unique usernames
	matches := mentionRe.FindAllStringSubmatch(stripped, -1)
	if len(matches) == 0 {
		return nil, nil
	}

	seen := make(map[string]bool)
	usernames := make([]string, 0, len(matches))
	for _, m := range matches {
		uname := strings.ToLower(m[1])
		if !seen[uname] {
			seen[uname] = true
			usernames = append(usernames, m[1]) // preserve original case for DB lookup
		}
	}

	// Look up users by username (case-insensitive)
	var users []model.User
	if err := db.Where("LOWER(username) IN ?", lowercaseAll(usernames)).Find(&users).Error; err != nil {
		return nil, err
	}
	return users, nil
}

func lowercaseAll(ss []string) []string {
	out := make([]string, len(ss))
	for i, s := range ss {
		out[i] = strings.ToLower(s)
	}
	return out
}
