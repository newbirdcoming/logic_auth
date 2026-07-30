package model

import "time"

type Role struct {
	ID          uint64       `gorm:"primaryKey;autoIncrement" json:"id"`
	Name        string       `gorm:"size:64;uniqueIndex;not null" json:"name"`
	Description string       `gorm:"size:256" json:"description,omitempty"`
	IsSystem    int8         `gorm:"default:0;not null" json:"is_system"`
	CreatedAt   time.Time    `json:"created_at"`
	Permissions []Permission `gorm:"many2many:role_permissions;" json:"permissions,omitempty"`
}

func (Role) TableName() string { return "roles" }

type Permission struct {
	ID        uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	Code      string    `gorm:"size:128;uniqueIndex;not null" json:"code"`     // user:create
	Name      string    `gorm:"size:128;not null" json:"name"`                  // 创建用户
	Module    string    `gorm:"size:64;not null;index" json:"module"`           // user
	CreatedAt time.Time `json:"created_at"`
}

func (Permission) TableName() string { return "permissions" }
