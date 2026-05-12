package models

import (
	"time"

	"gorm.io/gorm"
)

type Egg struct {
	ID                  uint           `gorm:"primaryKey" json:"id"`
	UUID                string         `gorm:"type:char(36);uniqueIndex" json:"uuid"`
	Name                string         `gorm:"size:255" json:"name"`
	Author              string         `gorm:"size:255" json:"author"`
	Description         *string        `gorm:"type:text" json:"description"`
	Features            *string        `gorm:"type:json" json:"features"`
	DockerImages        string         `gorm:"type:json" json:"docker_images"`
	ForceOutgoingIP     bool           `gorm:"default:false" json:"force_outgoing_ip"`
	FileDenylist        *string        `gorm:"type:json" json:"file_denylist"`
	ConfigFrom          *uint          `json:"config_from"`
	ConfigStop          *string        `gorm:"type:text" json:"config_stop"`
	ConfigStartup       *string        `gorm:"type:json" json:"config_startup"`
	ConfigLogs          *string        `gorm:"type:json" json:"config_logs"`
	ConfigFiles         *string        `gorm:"type:json" json:"config_files"`
	StartupCommands     string         `gorm:"type:json" json:"startup_commands"`
	UpdateURL           *string        `gorm:"size:255" json:"update_url"`
	ScriptIsPrivileged  bool           `gorm:"default:false" json:"script_is_privileged"`
	ScriptInstall       *string        `gorm:"type:text" json:"script_install"`
	ScriptEntry         string         `gorm:"size:255" json:"script_entry"`
	ScriptContainer     string         `gorm:"size:255" json:"script_container"`
	CopyScriptFrom      *uint          `json:"copy_script_from"`
	Tags                string         `gorm:"type:json;default:'[]'" json:"tags"`
	CreatedAt           time.Time      `json:"created_at"`
	UpdatedAt           time.Time      `json:"updated_at"`
	DeletedAt           gorm.DeletedAt `gorm:"index" json:"-"`

	Servers       []Server       `gorm:"foreignKey:EggID" json:"servers,omitempty"`
	EggVariables  []EggVariable  `gorm:"foreignKey:EggID" json:"egg_variables,omitempty"`
	Children      []Egg          `gorm:"foreignKey:ConfigFrom" json:"children,omitempty"`
}

func (Egg) TableName() string { return "eggs" }
