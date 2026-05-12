package models

import (
	"time"

	"gorm.io/gorm"
)

type SSHKey struct {
	ID          uint           `gorm:"primaryKey" json:"id"`
	UserID      uint           `gorm:"index" json:"user_id"`
	Name        string         `gorm:"size:255" json:"name"`
	Fingerprint string         `gorm:"size:255;uniqueIndex" json:"fingerprint"`
	PublicKey   string         `gorm:"type:text" json:"public_key"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`

	User User `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

func (SSHKey) TableName() string { return "ssh_keys" }
