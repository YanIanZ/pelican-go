package models

import "time"

type ServerVariable struct {
	ID            uint      `gorm:"primaryKey" json:"id"`
	ServerID      uint      `gorm:"index" json:"server_id"`
	VariableID    uint      `gorm:"index" json:"variable_id"`
	VariableValue string    `gorm:"type:text" json:"variable_value"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`

	Server      Server      `gorm:"foreignKey:ServerID" json:"server,omitempty"`
	EggVariable EggVariable `gorm:"foreignKey:VariableID" json:"egg_variable,omitempty"`
}

func (ServerVariable) TableName() string { return "server_variables" }
