package models

import (
	"time"

	"gorm.io/gorm"
)

type Backup struct {
	ID           uint           `gorm:"primaryKey" json:"id"`
	UUID         string         `gorm:"type:char(36);uniqueIndex" json:"uuid"`
	ServerID     uint           `gorm:"index" json:"server_id"`
	IsSuccessful *bool          `json:"is_successful"`
	IsLocked     bool           `gorm:"default:false" json:"is_locked"`
	Name         string         `gorm:"size:255" json:"name"`
	FileName     *string        `gorm:"size:255" json:"file_name"`
	Disk         string         `gorm:"size:50" json:"disk"`
	Size         int64          `gorm:"default:0" json:"size"`
	Checksum     *string        `gorm:"size:255" json:"checksum"`
	CompletedAt  *time.Time     `json:"completed_at"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`

	Server Server `gorm:"foreignKey:ServerID" json:"server,omitempty"`
}

func (Backup) TableName() string { return "backups" }
