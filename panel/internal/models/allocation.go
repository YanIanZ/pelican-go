package models

import (
	"time"

	"gorm.io/gorm"
)

type Allocation struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	NodeID    uint           `gorm:"index" json:"node_id"`
	ServerID  *uint          `gorm:"index" json:"server_id"`
	IP        string         `gorm:"size:45" json:"ip"`
	Alias     *string        `gorm:"size:255" json:"alias"`
	Port      int            `json:"port"`
	Notes     *string        `gorm:"type:text" json:"notes"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	Node   Node    `gorm:"foreignKey:NodeID" json:"node,omitempty"`
	Server *Server `gorm:"foreignKey:ServerID" json:"server,omitempty"`
}

func (Allocation) TableName() string { return "allocations" }
