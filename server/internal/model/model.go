package model

import (
	"time"
)

type User struct {
	ID          uint      `json:"id" gorm:"primaryKey;column:id"`
	Username    string    `json:"username" gorm:"unique;not null;column:username"`
	Email       string    `json:"email" gorm:"unique;not null;column:email"`
	Password    string    `json:"-" gorm:"not null;column:password"`
	Role        string    `json:"role" gorm:"default:'user';column:role"` // user / moderator / admin
	DisplayName string    `json:"display_name" gorm:"column:display_name"`
	AvatarURL   string    `json:"avatar_url" gorm:"column:avatar_url"`
	Bio         string    `json:"bio" gorm:"type:text;column:bio"`
	Website     string    `json:"website" gorm:"column:website"`
	Location    string    `json:"location" gorm:"column:location"`
	IsActive    bool      `json:"is_active" gorm:"default:true;column:is_active"`
	CreatedAt   time.Time `json:"created_at" gorm:"column:created_at"`
	UpdatedAt   time.Time `json:"updated_at" gorm:"column:updated_at"`
}

func (User) TableName() string {
	return "Users"
}

type Follow struct {
	FollowerID  uint      `json:"follower_id" gorm:"primaryKey"`
	FollowingID uint      `json:"following_id" gorm:"primaryKey"`
	CreatedAt   time.Time `json:"created_at"`
}

func (Follow) TableName() string {
	return "follows"
}

type UserSettings struct {
	UserID             uint `json:"user_id" gorm:"primaryKey"`
	EmailNotifications bool `json:"email_notifications" gorm:"default:true"`
	PrivateProfile     bool `json:"private_profile" gorm:"default:false"`
}

func (UserSettings) TableName() string {
	return "user_settings"
}

type Artist struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	Name      string    `json:"name" gorm:"unique;not null"`
	Bio       string    `json:"bio" gorm:"type:text"`
	ImageURL  string    `json:"image_url"`
	Albums    []Album   `json:"albums,omitempty" gorm:"many2many:album_artists;"`
	Songs     []Song    `json:"songs,omitempty" gorm:"many2many:song_artists;"`
	CreatedAt time.Time `json:"created_at" gorm:"column:created_at"`
	UpdatedAt time.Time `json:"updated_at" gorm:"column:updated_at"`
}

func (Artist) TableName() string {
	return "Artists"
}

type Album struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	Title       string    `json:"title" gorm:"not null"`
	Year        int       `json:"year"`
	ReleaseDate time.Time `json:"release_date" gorm:"type:date"`
	CoverURL    string    `json:"cover_url"`
	CoverSource string    `json:"cover_source" gorm:"default:'local'"`
	Status      string    `json:"status" gorm:"default:'pending'"`
	UploadedBy  *uint     `json:"uploaded_by"`
	User        *User     `json:"user,omitempty" gorm:"foreignKey:UploadedBy"`
	Artists     []Artist  `json:"artists,omitempty" gorm:"many2many:album_artists;"`
	Songs       []Song    `json:"songs,omitempty" gorm:"foreignKey:AlbumID"`
	CreatedAt   time.Time `json:"created_at" gorm:"column:created_at"`
	UpdatedAt   time.Time `json:"updated_at" gorm:"column:updated_at"`
}

func (Album) TableName() string {
	return "Albums"
}

type AlbumArtist struct {
	AlbumID   uint      `json:"album_id" gorm:"primaryKey"`
	ArtistID  uint      `json:"artist_id" gorm:"primaryKey"`
	Role      string    `json:"role" gorm:"default:'primary'"`
	CreatedAt time.Time `json:"created_at" gorm:"column:created_at"`
	UpdatedAt time.Time `json:"updated_at" gorm:"column:updated_at"`
}

func (AlbumArtist) TableName() string {
	return "album_artists"
}

type Song struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	Title       string    `json:"title" gorm:"not null"`
	ReleaseDate time.Time `json:"release_date" gorm:"type:date"`
	TrackNumber int       `json:"track_number"`
	Lyrics      string    `json:"lyrics" gorm:"type:text"`
	AudioURL    string    `json:"audio_url" gorm:"not null"`
	AudioSource string    `json:"audio_source" gorm:"default:'local'"`
	CoverURL    string    `json:"cover_url"`
	CoverSource string    `json:"cover_source" gorm:"default:'local'"`
	BatchID     string    `json:"batch_id" gorm:"index"`
	Status      string    `json:"status" gorm:"default:'pending'"`
	AlbumID     *uint     `json:"album_id"`
	Album       *Album    `json:"album,omitempty"`
	Artists     []Artist  `json:"artists,omitempty" gorm:"many2many:song_artists;"`
	UploadedBy  *uint     `json:"uploaded_by"`
	User        *User     `json:"user,omitempty" gorm:"foreignKey:UploadedBy"`
	CreatedAt   time.Time `json:"created_at" gorm:"column:created_at"`
	UpdatedAt   time.Time `json:"updated_at" gorm:"column:updated_at"`
}

func (Song) TableName() string {
	return "Songs"
}

type SongArtist struct {
	SongID    uint      `json:"song_id" gorm:"primaryKey"`
	ArtistID  uint      `json:"artist_id" gorm:"primaryKey"`
	Role      string    `json:"role" gorm:"default:'primary'"`
	CreatedAt time.Time `json:"created_at" gorm:"column:created_at"`
	UpdatedAt time.Time `json:"updated_at" gorm:"column:updated_at"`
}

func (SongArtist) TableName() string {
	return "song_artists"
}

type AlbumCorrection struct {
	ID      uint   `json:"id" gorm:"primaryKey"`
	AlbumID uint   `json:"album_id" gorm:"not null"`
	Album   *Album `json:"album,omitempty" gorm:"foreignKey:AlbumID"`
	UserID  *uint  `json:"user_id"`
	User    *User  `json:"user,omitempty" gorm:"foreignKey:UserID"`
	Status  string `json:"status" gorm:"default:'pending'"`

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
	CreatedAt      time.Time  `json:"created_at" gorm:"column:created_at"`
	ApprovedAt     *time.Time `json:"approved_at"`
	ApprovedBy     *uint      `json:"approved_by"`
	ApprovedByUser *User      `json:"approved_by_user,omitempty" gorm:"foreignKey:ApprovedBy"`
	RejectedAt     *time.Time `json:"rejected_at"`
	RejectedBy     *uint      `json:"rejected_by"`
	RejectedByUser *User      `json:"rejected_by_user,omitempty" gorm:"foreignKey:RejectedBy"`
}

func (AlbumCorrection) TableName() string {
	return "album_corrections"
}

type SongCorrection struct {
	ID     uint   `json:"id" gorm:"primaryKey"`
	SongID uint   `json:"song_id" gorm:"not null"`
	Song   *Song  `json:"song,omitempty" gorm:"foreignKey:SongID"`
	UserID *uint  `json:"user_id"`
	User   *User  `json:"user,omitempty" gorm:"foreignKey:UserID"`
	Status string `json:"status" gorm:"default:'pending'"`

	FieldName      string `json:"field_name" gorm:"not null"`
	CurrentValue   string `json:"current_value" gorm:"type:text"`
	CorrectedValue string `json:"corrected_value" gorm:"type:text;not null"`

	Reason         string     `json:"reason" gorm:"type:text"`
	CreatedAt      time.Time  `json:"created_at" gorm:"column:created_at"`
	ApprovedAt     *time.Time `json:"approved_at"`
	ApprovedBy     *uint      `json:"approved_by"`
	ApprovedByUser *User      `json:"approved_by_user,omitempty" gorm:"foreignKey:ApprovedBy"`
	RejectedAt     *time.Time `json:"rejected_at"`
	RejectedBy     *uint      `json:"rejected_by"`
	RejectedByUser *User      `json:"rejected_by_user,omitempty" gorm:"foreignKey:RejectedBy"`
}

func (SongCorrection) TableName() string {
	return "song_corrections"
}
