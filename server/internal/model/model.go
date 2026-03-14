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

type Channel struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	UserID      uint      `json:"user_id" gorm:"not null;index"`
	User        *User     `json:"user,omitempty" gorm:"foreignKey:UserID"`
	Name        string    `json:"name" gorm:"not null"`
	Description string    `json:"description" gorm:"type:text"`
	CoverURL    string    `json:"cover_url" gorm:"type:text"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func (Channel) TableName() string { return "channels" }

type Collection struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	ChannelID   uint      `json:"channel_id" gorm:"not null;index"`
	Channel     *Channel  `json:"channel,omitempty" gorm:"foreignKey:ChannelID"`
	Name        string    `json:"name" gorm:"not null"`
	Description string    `json:"description" gorm:"type:text"`
	CoverURL    string    `json:"cover_url" gorm:"type:text"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func (Collection) TableName() string { return "collections" }

type Post struct {
	ID            uint         `json:"id" gorm:"primaryKey"`
	UserID        uint         `json:"user_id" gorm:"not null;index"`
	User          *User        `json:"user,omitempty" gorm:"foreignKey:UserID"`
	Title         string       `json:"title" gorm:"not null"`
	Content       string       `json:"content" gorm:"type:text;not null"`
	Summary       string       `json:"summary" gorm:"type:text"`
	CoverURL      string       `json:"cover_url" gorm:"type:text"`
	Status        string       `json:"status" gorm:"default:'draft'"` // draft / published
	AllowComments bool         `json:"allow_comments" gorm:"default:true"`
	Pinned        bool         `json:"pinned" gorm:"default:false"`
	Collections   []Collection `json:"collections,omitempty" gorm:"many2many:post_collections;"`
	CreatedAt     time.Time    `json:"created_at"`
	UpdatedAt     time.Time    `json:"updated_at"`
}

func (Post) TableName() string { return "posts" }

type PostCollection struct {
	PostID       uint `json:"post_id" gorm:"primaryKey"`
	CollectionID uint `json:"collection_id" gorm:"primaryKey"`
}

func (PostCollection) TableName() string { return "post_collections" }

type Comment struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	PostID    uint      `json:"post_id" gorm:"not null;index"`
	Post      *Post     `json:"post,omitempty" gorm:"foreignKey:PostID"`
	UserID    uint      `json:"user_id" gorm:"not null;index"`
	User      *User     `json:"user,omitempty" gorm:"foreignKey:UserID"`
	Content   string    `json:"content" gorm:"type:text;not null"`
	Status    string    `json:"status" gorm:"default:'visible'"` // visible / hidden
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (Comment) TableName() string { return "comments" }

type Like struct {
	ID         uint      `json:"id" gorm:"primaryKey"`
	UserID     uint      `json:"user_id" gorm:"not null;index"`
	User       *User     `json:"user,omitempty" gorm:"foreignKey:UserID"`
	TargetType string    `json:"target_type" gorm:"not null"` // post / comment
	TargetID   uint      `json:"target_id" gorm:"not null"`
	CreatedAt  time.Time `json:"created_at"`
}

func (Like) TableName() string { return "likes" }

type BookmarkFolder struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	UserID    uint      `json:"user_id" gorm:"not null;index"`
	User      *User     `json:"user,omitempty" gorm:"foreignKey:UserID"`
	Name      string    `json:"name" gorm:"not null"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (BookmarkFolder) TableName() string { return "bookmark_folders" }

type Bookmark struct {
	ID               uint            `json:"id" gorm:"primaryKey"`
	UserID           uint            `json:"user_id" gorm:"not null;index"`
	User             *User           `json:"user,omitempty" gorm:"foreignKey:UserID"`
	PostID           uint            `json:"post_id" gorm:"not null;index"`
	Post             *Post           `json:"post,omitempty" gorm:"foreignKey:PostID"`
	BookmarkFolderID *uint           `json:"bookmark_folder_id" gorm:"index"`
	BookmarkFolder   *BookmarkFolder `json:"bookmark_folder,omitempty" gorm:"foreignKey:BookmarkFolderID"`
	CreatedAt        time.Time       `json:"created_at"`
}

func (Bookmark) TableName() string { return "bookmarks" }

type FeedSource struct {
	ID            uint       `json:"id" gorm:"primaryKey"`
	UserID        uint       `json:"user_id" gorm:"not null;index"` // 订阅者
	User          *User      `json:"user,omitempty" gorm:"foreignKey:UserID"`
	SourceType    string     `json:"source_type" gorm:"not null"` // internal_user | internal_channel | internal_collection | external_rss
	SourceID      *uint      `json:"source_id"`                   // 站内资源 ID（外部 RSS 时为 null）
	RssURL        string     `json:"rss_url" gorm:"type:text"`    // 外部 RSS URL（站内时为空）
	Title         string     `json:"title"`                       // 用户自定义名称
	LastFetchedAt *time.Time `json:"last_fetched_at"`
	CreatedAt     time.Time  `json:"created_at"`
}

func (FeedSource) TableName() string { return "feed_sources" }

// OrbitItem 仅存外部 RSS 抓取条目，站内内容动态 JOIN Post 表
type OrbitItem struct {
	ID           uint        `json:"id" gorm:"primaryKey"`
	FeedSourceID uint        `json:"feed_source_id" gorm:"not null;index"`
	FeedSource   *FeedSource `json:"feed_source,omitempty" gorm:"foreignKey:FeedSourceID"`
	GUID         string      `json:"guid" gorm:"not null"` // RSS item guid 或 link，用于去重
	Title        string      `json:"title"`
	Link         string      `json:"link" gorm:"type:text"`
	Summary      string      `json:"summary" gorm:"type:text"`
	Author       string      `json:"author"`
	PublishedAt  time.Time   `json:"published_at"`
	FetchedAt    time.Time   `json:"fetched_at"`
}

func (OrbitItem) TableName() string { return "orbit_items" }

type Notification struct {
	ID         uint       `json:"id" gorm:"primaryKey"`
	UserID     uint       `json:"user_id" gorm:"not null;index"` // 接收者
	User       *User      `json:"user,omitempty" gorm:"foreignKey:UserID"`
	Type       string     `json:"type" gorm:"not null"` // comment / like / bookmark / system
	Content    string     `json:"content" gorm:"type:text;not null"`
	TargetType string     `json:"target_type" gorm:"type:text"` // post / comment / nil
	TargetID   *uint      `json:"target_id"`
	ReadAt     *time.Time `json:"read_at"`
	CreatedAt  time.Time  `json:"created_at"`
}

func (Notification) TableName() string { return "notifications" }

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
