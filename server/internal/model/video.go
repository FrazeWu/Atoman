package model

import "github.com/google/uuid"

// Video represents a video post published under a channel.
type Video struct {
	Base
	ChannelID   *uuid.UUID `json:"channel_id,omitempty" gorm:"type:uuid;index"`
	Channel     *Channel   `json:"channel,omitempty" gorm:"foreignKey:ChannelID"`
	UserID      uuid.UUID  `json:"user_id" gorm:"type:uuid;not null;index"`
	User        *User      `json:"user,omitempty" gorm:"foreignKey:UserID;references:UUID"`
	Title       string     `json:"title" gorm:"not null"`
	Description string     `json:"description" gorm:"type:text"`
	// StorageType: "local" (S3/MinIO) or "external" (YouTube, Bilibili, etc.)
	StorageType  string `json:"storage_type" gorm:"not null;default:'external'"` // local | external
	VideoURL     string `json:"video_url" gorm:"type:text;not null"`             // S3 key or external URL
	ThumbnailURL string `json:"thumbnail_url" gorm:"type:text"`
	DurationSec  int    `json:"duration_sec" gorm:"default:0"`
	// Visibility: public | followers | private
	Visibility  string       `json:"visibility" gorm:"not null;default:'public'"`
	Status      string       `json:"status" gorm:"not null;default:'draft'"` // draft | published
	ViewCount   int          `json:"view_count" gorm:"default:0"`
	Tags        []VideoTag   `json:"tags,omitempty" gorm:"many2many:video_tag_relations;joinForeignKey:VideoID;joinReferences:TagID"`
	Collections []Collection `json:"collections,omitempty" gorm:"many2many:video_collections;"`
}

func (Video) TableName() string { return "videos" }

// VideoTag is a reusable tag for video discovery.
type VideoTag struct {
	Base
	Name string `json:"name" gorm:"uniqueIndex;not null"`
}

func (VideoTag) TableName() string { return "video_tags" }

// VideoCollection is the join table between Video and Collection.
type VideoCollection struct {
	VideoID      uuid.UUID `json:"video_id" gorm:"type:uuid;primaryKey"`
	CollectionID uuid.UUID `json:"collection_id" gorm:"type:uuid;primaryKey"`
}

func (VideoCollection) TableName() string { return "video_collections" }

// VideoTagRelation is the join table between Video and VideoTag.
type VideoTagRelation struct {
	VideoID uuid.UUID `json:"video_id" gorm:"type:uuid;primaryKey"`
	TagID   uuid.UUID `json:"tag_id" gorm:"type:uuid;primaryKey"`
}

func (VideoTagRelation) TableName() string { return "video_tag_relations" }
