package models

import "time"

type Role struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Name      string    `gorm:"size:255;uniqueIndex" json:"name"`
	GuardName string    `gorm:"size:255" json:"guard_name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (Role) TableName() string { return "roles" }

type Permission struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Name      string    `gorm:"size:255;uniqueIndex" json:"name"`
	GuardName string    `gorm:"size:255" json:"guard_name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (Permission) TableName() string { return "permissions" }

type ModelHasRole struct {
	RoleID    uint   `gorm:"primaryKey" json:"role_id"`
	ModelType string `gorm:"primaryKey;size:255" json:"model_type"`
	ModelID   string `gorm:"primaryKey;size:36" json:"model_id"`
}

func (ModelHasRole) TableName() string { return "model_has_roles" }
