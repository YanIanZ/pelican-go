package models

import (
	"time"

	"gorm.io/gorm"
)

type Webhook struct {
	ID          uint           `gorm:"primaryKey" json:"id"`
	ServerID    uint           `gorm:"index" json:"server_id"`
	URL         string         `gorm:"size:2048" json:"url"`
	Description *string        `gorm:"type:text" json:"description"`
	Events      string         `gorm:"type:json" json:"events"`
	Enabled     bool           `gorm:"default:true" json:"enabled"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`

	Server Server `gorm:"foreignKey:ServerID" json:"server,omitempty"`
}

func (Webhook) TableName() string { return "webhooks" }
