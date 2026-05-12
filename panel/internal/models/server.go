package models

import (
	"time"

	"gorm.io/gorm"
)

type Server struct {
	ID              uint           `gorm:"primaryKey" json:"id"`
	UUID            string         `gorm:"type:char(36);uniqueIndex" json:"uuid"`
	UUIDShort       string         `gorm:"type:char(8);uniqueIndex" json:"uuid_short"`
	NodeID          uint           `gorm:"index" json:"node_id"`
	OwnerID         uint           `gorm:"index" json:"owner_id"`
	Name            string         `gorm:"size:255" json:"name"`
	Description     *string        `gorm:"type:text" json:"description"`
	ExternalID      *string        `gorm:"size:255" json:"external_id"`
	Memory          int64          `json:"memory"`
	Swap            int64          `json:"swap"`
	Disk            int64          `json:"disk"`
	IO              int            `gorm:"default:500" json:"io"`
	CPU             int64          `json:"cpu"`
	Threads         *string        `gorm:"size:255" json:"threads"`
	OOMKiller       bool           `gorm:"default:false" json:"oom_killer"`
	AllocationID    *uint          `json:"allocation_id"`
	EggID           uint           `gorm:"index" json:"egg_id"`
	Startup         string         `gorm:"type:text" json:"startup"`
	Image           string         `gorm:"size:255" json:"image"`
	SkipScripts     bool           `gorm:"default:false" json:"skip_scripts"`
	DatabaseLimit   *int           `json:"database_limit"`
	AllocationLimit *int           `json:"allocation_limit"`
	BackupLimit     int            `gorm:"default:0" json:"backup_limit"`
	Status          *string        `gorm:"size:50" json:"status"`
	InstalledAt     *time.Time     `json:"installed_at"`
	DockerLabels    *string        `gorm:"type:json" json:"docker_labels"`
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
	DeletedAt       gorm.DeletedAt `gorm:"index" json:"-"`

	Node        Node         `gorm:"foreignKey:NodeID" json:"node,omitempty"`
	Owner       User         `gorm:"foreignKey:OwnerID" json:"owner,omitempty"`
	Allocation  *Allocation  `gorm:"foreignKey:AllocationID" json:"allocation,omitempty"`
	Egg         Egg          `gorm:"foreignKey:EggID" json:"egg,omitempty"`
	Allocations []Allocation `gorm:"foreignKey:ServerID" json:"allocations,omitempty"`
	Subusers    []Subuser    `gorm:"foreignKey:ServerID" json:"subusers,omitempty"`
	Schedules   []Schedule   `gorm:"foreignKey:ServerID" json:"schedules,omitempty"`
	Databases   []Database   `gorm:"foreignKey:ServerID" json:"databases,omitempty"`
	Backups     []Backup     `gorm:"foreignKey:ServerID" json:"backups,omitempty"`
}

func (Server) TableName() string { return "servers" }
