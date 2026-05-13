package model

import (
	"time"

	"github.com/google/uuid"
)

type Channel struct {
	Base
	UserID      uuid.UUID `json:"user_id" gorm:"type:uuid;not null;index"`
	User        *User     `json:"user,omitempty" gorm:"foreignKey:UserID;references:UUID"`
	Name        string    `json:"name" gorm:"not null;uniqueIndex:idx_channels_name"`
	Slug        string    `json:"slug" gorm:"uniqueIndex"`
	Description string    `json:"description" gorm:"type:text"`
	CoverURL    string    `json:"cover_url" gorm:"type:text"`
	IsDefault   bool      `json:"is_default" gorm:"default:false;index"`
}

func (Channel) TableName() string { return "channels" }

type Collection struct {
	Base
	ChannelID   uuid.UUID `json:"channel_id" gorm:"type:uuid;not null;index;uniqueIndex:idx_collection_channel_name,priority:1"`
	Channel     *Channel  `json:"channel,omitempty" gorm:"foreignKey:ChannelID"`
	Name        string    `json:"name" gorm:"not null;uniqueIndex:idx_collection_channel_name,priority:2"`
	Description string    `json:"description" gorm:"type:text"`
	CoverURL    string    `json:"cover_url" gorm:"type:text"`
	IsDefault   bool      `json:"is_default" gorm:"default:false;index"`
}

func (Collection) TableName() string { return "collections" }

type Post struct {
	Base
	UserID        uuid.UUID    `json:"user_id" gorm:"type:uuid;not null;index"`
	User          *User        `json:"user,omitempty" gorm:"foreignKey:UserID;references:UUID"`
	ChannelID     *uuid.UUID   `json:"channel_id,omitempty" gorm:"type:uuid;index"`
	Channel       *Channel     `json:"channel,omitempty" gorm:"foreignKey:ChannelID"`
	Title         string       `json:"title" gorm:"not null"`
	Content       string       `json:"content" gorm:"type:text;not null"`
	Summary       string       `json:"summary" gorm:"type:text"`
	CoverURL      string       `json:"cover_url" gorm:"type:text"`
	Status        string       `json:"status" gorm:"default:'draft'"` // draft / published
	AllowComments bool         `json:"allow_comments" gorm:"default:true"`
	Pinned        bool         `json:"pinned" gorm:"default:false"`
	Collections   []Collection `json:"collections,omitempty" gorm:"many2many:post_collections;"`
}

func (Post) TableName() string { return "posts" }

type BlogDraft struct {
	Base
	UserID        uuid.UUID  `json:"user_id" gorm:"type:uuid;not null;index;uniqueIndex:idx_blog_drafts_user_context,priority:1"`
	ContextKey    string     `json:"context_key" gorm:"not null;uniqueIndex:idx_blog_drafts_user_context,priority:2"`
	SourcePostID  *uuid.UUID `json:"source_post_id,omitempty" gorm:"type:uuid;index"`
	Title         string     `json:"title"`
	Content       string     `json:"content" gorm:"type:text"`
	Summary       string     `json:"summary" gorm:"type:text"`
	CoverURL      string     `json:"cover_url" gorm:"type:text"`
	AllowComments bool       `json:"allow_comments" gorm:"default:true"`
	ChannelID     *uuid.UUID `json:"channel_id,omitempty" gorm:"type:uuid;index"`
	CollectionIDs string     `json:"collection_ids" gorm:"type:text"`
}

func (BlogDraft) TableName() string { return "blog_drafts" }

type PostCollection struct {
	PostID       uuid.UUID `json:"post_id" gorm:"type:uuid;primaryKey"`
	CollectionID uuid.UUID `json:"collection_id" gorm:"type:uuid;primaryKey"`
}

func (PostCollection) TableName() string { return "post_collections" }

type Comment struct {
	Base
	PostID  uuid.UUID `json:"post_id" gorm:"type:uuid;not null;index"`
	Post    *Post     `json:"post,omitempty" gorm:"foreignKey:PostID"`
	UserID  uuid.UUID `json:"user_id" gorm:"type:uuid;not null;index"`
	User    *User     `json:"user,omitempty" gorm:"foreignKey:UserID;references:UUID"`
	Content string    `json:"content" gorm:"type:text;not null"`
	Status  string    `json:"status" gorm:"default:'visible'"` // visible / hidden
}

func (Comment) TableName() string { return "comments" }

type Like struct {
	Base
	UserID     uuid.UUID `json:"user_id" gorm:"type:uuid;not null;index"`
	User       *User     `json:"user,omitempty" gorm:"foreignKey:UserID;references:UUID"`
	TargetType string    `json:"target_type" gorm:"not null"` // post / comment
	TargetID   uuid.UUID `json:"target_id" gorm:"type:uuid;not null"`
}

func (Like) TableName() string { return "likes" }

type BookmarkFolder struct {
	Base
	UserID uuid.UUID `json:"user_id" gorm:"type:uuid;not null;index"`
	User   *User     `json:"user,omitempty" gorm:"foreignKey:UserID;references:UUID"`
	Name   string    `json:"name" gorm:"not null"`
}

func (BookmarkFolder) TableName() string { return "bookmark_folders" }

type Bookmark struct {
	Base
	UserID           uuid.UUID       `json:"user_id" gorm:"type:uuid;not null;index"`
	User             *User           `json:"user,omitempty" gorm:"foreignKey:UserID;references:UUID"`
	PostID           uuid.UUID       `json:"post_id" gorm:"type:uuid;not null;index"`
	Post             *Post           `json:"post,omitempty" gorm:"foreignKey:PostID"`
	BookmarkFolderID *uuid.UUID      `json:"bookmark_folder_id" gorm:"type:uuid;index"`
	BookmarkFolder   *BookmarkFolder `json:"bookmark_folder,omitempty" gorm:"foreignKey:BookmarkFolderID"`
}

// FeedSource 存储全局唯一的订阅源元数据
type FeedSource struct {
	Base
	SourceType    string     `json:"source_type" gorm:"not null"` // internal_user | internal_channel | internal_collection | external_rss
	SourceID      *uuid.UUID `json:"source_id" gorm:"type:uuid"`  // 站内资源 ID（外部 RSS 时为 null）
	RssURL        string     `json:"rss_url" gorm:"type:text"`
	Hash          string     `json:"hash" gorm:"type:varchar(64);uniqueIndex"` // 唯一哈希
	Title         string     `json:"title"`                                    // 全局默认标题
	LastFetchedAt *time.Time `json:"last_fetched_at"`
}

func (FeedSource) TableName() string { return "feed_sources" }

// Subscription 存储用户与订阅源的多对多关系
type Subscription struct {
	Base
	UserID              uuid.UUID          `json:"user_id" gorm:"type:uuid;not null;index"`
	User                *User              `json:"user,omitempty" gorm:"foreignKey:UserID;references:UUID"`
	FeedSourceID        uuid.UUID          `json:"feed_source_id" gorm:"type:uuid;not null;index"`
	FeedSource          *FeedSource        `json:"feed_source,omitempty" gorm:"foreignKey:FeedSourceID"`
	Title               string             `json:"title"`
	SubscriptionGroupID *uuid.UUID         `json:"subscription_group_id" gorm:"type:uuid;index"`
	SubscriptionGroup   *SubscriptionGroup `json:"subscription,omitempty" gorm:"foreignKey:SubscriptionGroupID"`
	HealthStatus        string             `json:"health_status" gorm:"default:'healthy'"` // healthy | warning | error
	ErrorMessage        string             `json:"error_message" gorm:"type:text"`
	LastChecked         *time.Time         `json:"last_checked"`
}

func (Subscription) TableName() string { return "subscriptions" }

type FeedItem struct {
	Base
	FeedSourceID     uuid.UUID   `json:"feed_source_id" gorm:"type:uuid;not null;index"`
	FeedSource       *FeedSource `json:"feed_source,omitempty" gorm:"foreignKey:FeedSourceID"`
	GUID             string      `json:"guid" gorm:"not null"`
	Title            string      `json:"title"`
	Link             string      `json:"link" gorm:"type:text"`
	Summary          string      `json:"summary" gorm:"type:text"`
	Author           string      `json:"author"`
	PublishedAt      time.Time   `json:"published_at"`
	FetchedAt        time.Time   `json:"fetched_at"`
	EnclosureURL     string      `json:"enclosure_url" gorm:"type:text"`
	EnclosureType    string      `json:"enclosure_type"`
	Duration         string      `json:"duration"`
	ImageURL         string      `json:"image_url" gorm:"type:text"`
	IsDuplicate      bool        `json:"is_duplicate" gorm:"-"`
	DuplicateCount   int         `json:"duplicate_count" gorm:"-"`
	DuplicateOfID    *uuid.UUID  `json:"duplicate_of_id,omitempty" gorm:"-"`
	DuplicateSources []string    `json:"duplicate_sources,omitempty" gorm:"-"`
}

func (FeedItem) TableName() string { return "feed_items" }

type FeedItemRead struct {
	UserID     uuid.UUID `json:"user_id" gorm:"type:uuid;not null;primaryKey;index"`
	FeedItemID uuid.UUID `json:"feed_item_id" gorm:"type:uuid;not null;primaryKey;index"`
	ReadAt     time.Time `json:"read_at"`
}

func (FeedItemRead) TableName() string { return "feed_item_reads" }

type FeedItemStar struct {
	UserID     uuid.UUID `json:"user_id" gorm:"type:uuid;not null;primaryKey;index"`
	FeedItemID uuid.UUID `json:"feed_item_id" gorm:"type:uuid;not null;primaryKey;index"`
	StarredAt  time.Time `json:"starred_at"`
}

func (FeedItemStar) TableName() string { return "feed_item_stars" }

type ReadingListItem struct {
	UserID     uuid.UUID `json:"user_id" gorm:"type:uuid;not null;primaryKey;index"`
	FeedItemID uuid.UUID `json:"feed_item_id" gorm:"type:uuid;not null;primaryKey;index"`
	FeedItem   *FeedItem `json:"feed_item,omitempty" gorm:"foreignKey:FeedItemID"`
	CreatedAt  time.Time `json:"created_at"`
}

func (ReadingListItem) TableName() string { return "reading_list_items" }

type SubscriptionGroup struct {
	Base
	UserID uuid.UUID `json:"user_id" gorm:"type:uuid;not null;index"`
	Name   string    `json:"name" gorm:"not null"`
}

func (SubscriptionGroup) TableName() string { return "subscription_groups" }
