package model

import "time"

type User struct {
	ID            uint64     `gorm:"primaryKey;autoIncrement" json:"id"`
	Username      string     `gorm:"size:64;uniqueIndex;not null" json:"username"`
	Email         string     `gorm:"size:128;uniqueIndex;not null" json:"email"`
	Phone         string     `gorm:"size:20;uniqueIndex" json:"phone,omitempty"`
	PasswordHash  string     `gorm:"size:256;not null" json:"-"`
	Nickname      string     `gorm:"size:64" json:"nickname,omitempty"`
	AvatarURL     string     `gorm:"size:512" json:"avatar_url,omitempty"`
	Status        int8       `gorm:"default:1;not null" json:"status"`  // 1:正常 0:禁用 -1:删除
	EmailVerified int8       `gorm:"default:0;not null" json:"email_verified"`
	PhoneVerified int8       `gorm:"default:0;not null" json:"phone_verified"`
	LastLoginAt   *time.Time `json:"last_login_at,omitempty"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
	Roles         []Role     `gorm:"many2many:user_roles;" json:"roles,omitempty"`
}

func (User) TableName() string { return "users" }
