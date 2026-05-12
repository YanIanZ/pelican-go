package models

import (
	"time"

	"gorm.io/gorm"
)

type EggVariable struct {
	ID            uint           `gorm:"primaryKey" json:"id"`
	EggID         uint           `gorm:"index" json:"egg_id"`
	Name          string         `gorm:"size:255" json:"name"`
	Description   *string        `gorm:"type:text" json:"description"`
	EnvVariable   string         `gorm:"size:255" json:"env_variable"`
	DefaultValue  string         `gorm:"size:255" json:"default_value"`
	UserViewable  bool           `gorm:"default:false" json:"user_viewable"`
	UserEditable  bool           `gorm:"default:false" json:"user_editable"`
	Required      bool           `gorm:"default:false" json:"required"`
	Rules         string         `gorm:"type:json" json:"rules"`
	SortOrder     int            `gorm:"default:0" json:"sort_order"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"-"`

	Egg Egg `gorm:"foreignKey:EggID" json:"egg,omitempty"`
}

func (EggVariable) TableName() string { return "egg_variables" }
