package model

import (
	"time"

	"github.com/google/uuid"
)

type Channel struct {
	Base
	UserID      uuid.UUID `json:"user_id" gorm:"type:uuid;not null;index"`
	User        *User     `json:"user,omitempty" gorm:"foreignKey:UserID;references:UUID"`
	Name        string    `json:"name" gorm:"not null"`
	Description string    `json:"description" gorm:"type:text"`
	CoverURL    string    `json:"cover_url" gorm:"type:text"`
}

func (Channel) TableName() string { return "channels" }

type Collection struct {
	Base
	ChannelID   uuid.UUID `json:"channel_id" gorm:"type:uuid;not null;index"`
	Channel     *Channel  `json:"channel,omitempty" gorm:"foreignKey:ChannelID"`
	Name        string    `json:"name" gorm:"not null"`
	Description string    `json:"description" gorm:"type:text"`
	CoverURL    string    `json:"cover_url" gorm:"type:text"`
}

func (Collection) TableName() string { return "collections" }

type Post struct {
	Base
	UserID        uuid.UUID    `json:"user_id" gorm:"type:uuid;not null;index"`
	User          *User        `json:"user,omitempty" gorm:"foreignKey:UserID;references:UUID"`
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

func (Bookmark) TableName() string { return "bookmarks" }

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
}

func (Subscription) TableName() string { return "subscriptions" }

type FeedItem struct {
	Base
	FeedSourceID  uuid.UUID   `json:"feed_source_id" gorm:"type:uuid;not null;index"`
	FeedSource    *FeedSource `json:"feed_source,omitempty" gorm:"foreignKey:FeedSourceID"`
	GUID          string      `json:"guid" gorm:"not null"`
	Title         string      `json:"title"`
	Link          string      `json:"link" gorm:"type:text"`
	Summary       string      `json:"summary" gorm:"type:text"`
	Author        string      `json:"author"`
	PublishedAt   time.Time   `json:"published_at"`
	FetchedAt     time.Time   `json:"fetched_at"`
	EnclosureURL  string      `json:"enclosure_url" gorm:"type:text"`
	EnclosureType string      `json:"enclosure_type"`
	Duration      string      `json:"duration"`
	ImageURL      string      `json:"image_url" gorm:"type:text"`
}

func (FeedItem) TableName() string { return "feed_items" }

type FeedItemRead struct {
	UserID     uuid.UUID `json:"user_id" gorm:"type:uuid;not null;primaryKey;index"`
	FeedItemID uuid.UUID `json:"feed_item_id" gorm:"type:uuid;not null;primaryKey;index"`
	ReadAt     time.Time `json:"read_at"`
}

func (FeedItemRead) TableName() string { return "feed_item_reads" }

type SubscriptionGroup struct {
	Base
	UserID uuid.UUID `json:"user_id" gorm:"type:uuid;not null;index"`
	Name   string    `json:"name" gorm:"not null"`
}

func (SubscriptionGroup) TableName() string { return "subscription_groups" }
