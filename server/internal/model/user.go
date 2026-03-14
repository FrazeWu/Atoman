package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type User struct {
	UUID        uuid.UUID      `json:"uuid" gorm:"type:uuid;primaryKey"`
	ID          uint           `json:"id" gorm:"unique;autoIncrement"` // Frontend identifier
	Username    string         `json:"username" gorm:"unique;not null;column:username"`
	Email       string         `json:"email" gorm:"unique;not null;column:email"`
	Password    string         `json:"-" gorm:"not null;column:password"`
	Role        string         `json:"role" gorm:"default:'user';column:role"` // user / moderator / admin
	DisplayName string         `json:"display_name" gorm:"column:display_name"`
	AvatarURL   string         `json:"avatar_url" gorm:"column:avatar_url"`
	Bio         string         `json:"bio" gorm:"type:text;column:bio"`
	Website     string         `json:"website" gorm:"column:website"`
	Location    string         `json:"location" gorm:"column:location"`
	IsActive    bool           `json:"is_active" gorm:"default:true;column:is_active"`
	CreatedAt   time.Time      `json:"created_at" gorm:"column:created_at"`
	UpdatedAt   time.Time      `json:"updated_at" gorm:"column:updated_at"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"`
}

func (u *User) BeforeCreate(tx *gorm.DB) error {
	if u.UUID == uuid.Nil {
		id, err := uuid.NewV7()
		if err != nil {
			return err
		}
		u.UUID = id
	}
	return nil
}

func (User) TableName() string {
	return "Users"
}

type Follow struct {
	FollowerID  uuid.UUID `json:"follower_id" gorm:"type:uuid;primaryKey"`
	FollowingID uuid.UUID `json:"following_id" gorm:"type:uuid;primaryKey"`
	CreatedAt   time.Time `json:"created_at"`
}

func (Follow) TableName() string {
	return "follows"
}

type UserSettings struct {
	UserID             uuid.UUID `json:"user_id" gorm:"type:uuid;primaryKey"`
	EmailNotifications bool      `json:"email_notifications" gorm:"default:true"`
	PrivateProfile     bool      `json:"private_profile" gorm:"default:false"`
}

func (UserSettings) TableName() string {
	return "user_settings"
}

type Notification struct {
	Base
	UserID     uuid.UUID  `json:"user_id" gorm:"type:uuid;not null;index"` // 接收者
	User       *User      `json:"user,omitempty" gorm:"foreignKey:UserID;references:UUID"`
	Type       string     `json:"type" gorm:"not null"` // comment / like / bookmark / system
	Content    string     `json:"content" gorm:"type:text;not null"`
	TargetType string     `json:"target_type" gorm:"type:text"` // post / comment / nil
	TargetID   *uuid.UUID `json:"target_id" gorm:"type:uuid"`
	ReadAt     *time.Time `json:"read_at"`
}

func (Notification) TableName() string {
	return "notifications"
}
