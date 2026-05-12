package models

import (
	"time"

	"gorm.io/gorm"
)

type Mount struct {
	ID            uint           `gorm:"primaryKey" json:"id"`
	UUID          string         `gorm:"type:char(36);uniqueIndex" json:"uuid"`
	Name          string         `gorm:"size:255" json:"name"`
	Description   *string        `gorm:"type:text" json:"description"`
	Source        string         `gorm:"size:255" json:"source"`
	Target        string         `gorm:"size:255" json:"target"`
	ReadOnly      bool           `gorm:"default:false" json:"read_only"`
	UserMountable bool           `gorm:"default:false" json:"user_mountable"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"-"`
}

func (Mount) TableName() string { return "mounts" }

type MountEgg struct {
	MountID   uint      `gorm:"primaryKey" json:"mount_id"`
	EggID     uint      `gorm:"primaryKey" json:"egg_id"`
	CreatedAt time.Time `json:"created_at"`
}

func (MountEgg) TableName() string { return "mount_egg" }

type MountNode struct {
	MountID   uint      `gorm:"primaryKey" json:"mount_id"`
	NodeID    uint      `gorm:"primaryKey" json:"node_id"`
	CreatedAt time.Time `json:"created_at"`
}

func (MountNode) TableName() string { return "mount_node" }
