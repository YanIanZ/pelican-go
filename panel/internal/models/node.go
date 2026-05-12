package models

import (
	"time"

	"gorm.io/gorm"
)

type Node struct {
	ID                 uint           `gorm:"primaryKey" json:"id"`
	UUID               string         `gorm:"type:char(36);uniqueIndex" json:"uuid"`
	Public             bool           `gorm:"default:true" json:"public"`
	Name               string         `gorm:"size:255" json:"name"`
	FQDN               string         `gorm:"size:255" json:"fqdn"`
	Scheme             string         `gorm:"size:5;default:https" json:"scheme"`
	BehindProxy        bool           `gorm:"default:false" json:"behind_proxy"`
	Memory             int64          `gorm:"default:0" json:"memory"`
	MemoryOverallocate int64          `gorm:"default:0" json:"memory_overallocate"`
	Disk               int64          `gorm:"default:0" json:"disk"`
	DiskOverallocate   int64          `gorm:"default:0" json:"disk_overallocate"`
	CPU                int64          `gorm:"default:0" json:"cpu"`
	CPUOverallocate    int64          `gorm:"default:0" json:"cpu_overallocate"`
	UploadSize         int64          `gorm:"default:100" json:"upload_size"`
	DaemonBase         string         `gorm:"size:255;default:/var/lib/pelican/volumes" json:"daemon_base"`
	DaemonToken        string         `gorm:"size:255" json:"-"`
	DaemonTokenID      string         `gorm:"size:255" json:"-"`
	DaemonListen       int            `gorm:"default:8080" json:"daemon_listen"`
	DaemonSFTP         int            `gorm:"default:2022" json:"daemon_sftp"`
	DaemonSFTPAlias    *string        `gorm:"size:255" json:"daemon_sftp_alias"`
	DaemonConnect      int            `gorm:"default:8080" json:"daemon_connect"`
	Description        *string        `gorm:"type:text" json:"description"`
	MaintenanceMode    bool           `gorm:"default:false" json:"maintenance_mode"`
	Tags               string         `gorm:"type:json;default:'[]'" json:"tags"`
	CreatedAt          time.Time      `json:"created_at"`
	UpdatedAt          time.Time      `json:"updated_at"`
	DeletedAt          gorm.DeletedAt `gorm:"index" json:"-"`

	Servers     []Server     `gorm:"foreignKey:NodeID" json:"servers,omitempty"`
	Allocations []Allocation `gorm:"foreignKey:NodeID" json:"allocations,omitempty"`
}

func (Node) TableName() string { return "nodes" }
