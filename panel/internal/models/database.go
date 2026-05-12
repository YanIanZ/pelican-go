package models

import (
	"time"

	"gorm.io/gorm"
)

type DatabaseHost struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	Name      string         `gorm:"size:255" json:"name"`
	Host      string         `gorm:"size:255" json:"host"`
	Port      int            `gorm:"default:3306" json:"port"`
	Username  string         `gorm:"size:255" json:"username"`
	Password  string         `gorm:"size:255" json:"-"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	Databases []Database `gorm:"foreignKey:DatabaseHostID" json:"databases,omitempty"`
}

func (DatabaseHost) TableName() string { return "database_hosts" }

type Database struct {
	ID             uint           `gorm:"primaryKey" json:"id"`
	ServerID       uint           `gorm:"index" json:"server_id"`
	DatabaseHostID uint           `gorm:"index" json:"database_host_id"`
	DatabaseName   string         `gorm:"size:255" json:"database_name"`
	Username       string         `gorm:"size:255" json:"username"`
	Password       string         `gorm:"size:255" json:"-"`
	Remote         string         `gorm:"size:50;default:%" json:"remote"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
	DeletedAt      gorm.DeletedAt `gorm:"index" json:"-"`

	Server       Server       `gorm:"foreignKey:ServerID" json:"server,omitempty"`
	DatabaseHost DatabaseHost `gorm:"foreignKey:DatabaseHostID" json:"database_host,omitempty"`
}

func (Database) TableName() string { return "databases" }
