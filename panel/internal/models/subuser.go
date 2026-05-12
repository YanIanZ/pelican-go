package models

import (
	"time"

	"gorm.io/gorm"
)

type Subuser struct {
	ID          uint           `gorm:"primaryKey" json:"id"`
	UserID      uint           `gorm:"index" json:"user_id"`
	ServerID    uint           `gorm:"index" json:"server_id"`
	Permissions *string        `gorm:"type:json" json:"permissions"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`

	User   User   `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Server Server `gorm:"foreignKey:ServerID" json:"server,omitempty"`
}

func (Subuser) TableName() string { return "subusers" }
