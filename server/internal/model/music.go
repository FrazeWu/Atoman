package model

import (
	"time"

	"github.com/google/uuid"
)

type Artist struct {
	Base
	Name     string  `json:"name" gorm:"unique;not null"`
	Bio      string  `json:"bio" gorm:"type:text"`
	ImageURL string  `json:"image_url"`
	Albums   []Album `json:"albums,omitempty" gorm:"many2many:album_artists;"`
	Songs    []Song  `json:"songs,omitempty" gorm:"many2many:song_artists;"`
}

func (Artist) TableName() string {
	return "Artists"
}

type Album struct {
	Base
	Title       string     `json:"title" gorm:"not null"`
	Year        int        `json:"year"`
	ReleaseDate time.Time  `json:"release_date" gorm:"type:date"`
	CoverURL    string     `json:"cover_url"`
	CoverSource string     `json:"cover_source" gorm:"default:'local'"`
	Status      string     `json:"status" gorm:"default:'open'"`
	UploadedBy  *uuid.UUID `json:"uploaded_by" gorm:"type:uuid"`
	User        *User      `json:"user,omitempty" gorm:"foreignKey:UploadedBy;references:UUID"`
	Artists     []Artist   `json:"artists,omitempty" gorm:"many2many:album_artists;"`
	Songs       []Song     `json:"songs,omitempty" gorm:"foreignKey:AlbumID"`
}

func (Album) TableName() string {
	return "Albums"
}

type AlbumArtist struct {
	AlbumID   uuid.UUID `json:"album_id" gorm:"type:uuid;primaryKey"`
	ArtistID  uuid.UUID `json:"artist_id" gorm:"type:uuid;primaryKey"`
	Role      string    `json:"role" gorm:"default:'primary'"`
	CreatedAt time.Time `json:"created_at" gorm:"column:created_at"`
	UpdatedAt time.Time `json:"updated_at" gorm:"column:updated_at"`
}

func (AlbumArtist) TableName() string {
	return "album_artists"
}

type Song struct {
	Base
	Title       string     `json:"title" gorm:"not null"`
	ReleaseDate time.Time  `json:"release_date" gorm:"type:date"`
	TrackNumber int        `json:"track_number"`
	Lyrics      string     `json:"lyrics" gorm:"type:text"`
	AudioURL    string     `json:"audio_url" gorm:"not null"`
	AudioSource string     `json:"audio_source" gorm:"default:'local'"`
	CoverURL    string     `json:"cover_url"`
	CoverSource string     `json:"cover_source" gorm:"default:'local'"`
	BatchID     string     `json:"batch_id" gorm:"index"`
	Status      string     `json:"status" gorm:"default:'open'"`
	AlbumID     *uuid.UUID `json:"album_id" gorm:"type:uuid"`
	Album       *Album     `json:"album,omitempty"`
	Artists     []Artist   `json:"artists,omitempty" gorm:"many2many:song_artists;"`
	UploadedBy  *uuid.UUID `json:"uploaded_by" gorm:"type:uuid"`
	User        *User      `json:"user,omitempty" gorm:"foreignKey:UploadedBy;references:UUID"`
}

func (Song) TableName() string {
	return "Songs"
}

type SongArtist struct {
	SongID    uuid.UUID `json:"song_id" gorm:"type:uuid;primaryKey"`
	ArtistID  uuid.UUID `json:"artist_id" gorm:"type:uuid;primaryKey"`
	Role      string    `json:"role" gorm:"default:'primary'"`
	CreatedAt time.Time `json:"created_at" gorm:"column:created_at"`
	UpdatedAt time.Time `json:"updated_at" gorm:"column:updated_at"`
}

func (SongArtist) TableName() string {
	return "song_artists"
}

type AlbumCorrection struct {
	Base
	AlbumID uuid.UUID  `json:"album_id" gorm:"type:uuid;not null"`
	Album   *Album     `json:"album,omitempty" gorm:"foreignKey:AlbumID"`
	UserID  *uuid.UUID `json:"user_id" gorm:"type:uuid"`
	User    *User      `json:"user,omitempty" gorm:"foreignKey:UserID;references:UUID"`
	Status  string     `json:"status" gorm:"default:'pending'"`

	OriginalTitle       string     `json:"original_title"`
	OriginalCoverURL    string     `json:"original_cover_url" gorm:"type:text"`
	OriginalReleaseDate *time.Time `json:"original_release_date" gorm:"type:date"`
	OriginalArtistIDs   string     `json:"original_artist_ids" gorm:"type:text"`

	CorrectedTitle       string     `json:"corrected_title"`
	CorrectedCoverURL    string     `json:"corrected_cover_url" gorm:"type:text"`
	CorrectedCoverSource string     `json:"corrected_cover_source" gorm:"default:'local'"`
	CorrectedReleaseDate *time.Time `json:"corrected_release_date" gorm:"type:date"`
	CorrectedArtistIDs   string     `json:"corrected_artist_ids" gorm:"type:text"`

	Reason         string     `json:"reason" gorm:"type:text"`
	ApprovedAt     *time.Time `json:"approved_at"`
	ApprovedBy     *uuid.UUID `json:"approved_by" gorm:"type:uuid"`
	ApprovedByUser *User      `json:"approved_by_user,omitempty" gorm:"foreignKey:ApprovedBy;references:UUID"`
	RejectedAt     *time.Time `json:"rejected_at"`
	RejectedBy     *uuid.UUID `json:"rejected_by" gorm:"type:uuid"`
	RejectedByUser *User      `json:"rejected_by_user,omitempty" gorm:"foreignKey:RejectedBy;references:UUID"`
}

func (AlbumCorrection) TableName() string {
	return "album_corrections"
}

type SongCorrection struct {
	Base
	SongID uuid.UUID  `json:"song_id" gorm:"type:uuid;not null"`
	Song   *Song      `json:"song,omitempty" gorm:"foreignKey:SongID"`
	UserID *uuid.UUID `json:"user_id" gorm:"type:uuid"`
	User   *User      `json:"user,omitempty" gorm:"foreignKey:UserID;references:UUID"`
	Status string     `json:"status" gorm:"default:'pending'"`

	FieldName      string `json:"field_name" gorm:"not null"`
	CurrentValue   string `json:"current_value" gorm:"type:text"`
	CorrectedValue string `json:"corrected_value" gorm:"type:text;not null"`

	Reason         string     `json:"reason" gorm:"type:text"`
	ApprovedAt     *time.Time `json:"approved_at"`
	ApprovedBy     *uuid.UUID `json:"approved_by" gorm:"type:uuid"`
	ApprovedByUser *User      `json:"approved_by_user,omitempty" gorm:"foreignKey:ApprovedBy;references:UUID"`
	RejectedAt     *time.Time `json:"rejected_at"`
	RejectedBy     *uuid.UUID `json:"rejected_by" gorm:"type:uuid"`
	RejectedByUser *User      `json:"rejected_by_user,omitempty" gorm:"foreignKey:RejectedBy;references:UUID"`
}

func (SongCorrection) TableName() string {
	return "song_corrections"
}
