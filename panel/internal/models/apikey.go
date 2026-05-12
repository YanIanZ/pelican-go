package models

import (
	"time"

	"gorm.io/gorm"
)

const (
	APIKeyTypeNone        = 0
	APIKeyTypeAccount     = 1
	APIKeyTypeApplication = 2
	IdentifierLength      = 16
	RandomLength          = 32
)

type APIKey struct {
	ID          uint           `gorm:"primaryKey" json:"id"`
	UserID      *uint          `gorm:"index" json:"user_id"`
	Token       string         `json:"-"`
	Identifier  string         `gorm:"type:char(16);uniqueIndex" json:"identifier"`
	KeyType     int            `gorm:"default:0" json:"key_type"`
	Memo        *string        `gorm:"size:500" json:"memo"`
	AllowedIPs  *string        `gorm:"type:json" json:"allowed_ips"`
	Permissions *string        `gorm:"type:json" json:"permissions"`
	LastUsedAt  *time.Time     `json:"last_used_at"`
	ExpiresAt   *time.Time     `json:"expires_at"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`

	User User `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

func (APIKey) TableName() string { return "api_keys" }
