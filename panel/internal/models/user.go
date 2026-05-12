package models

import (
	"time"

	"gorm.io/gorm"
)

type User struct {
	ID                  uint           `gorm:"primaryKey" json:"id"`
	UUID                string         `gorm:"type:char(36);uniqueIndex" json:"uuid"`
	Email               string         `gorm:"size:255;uniqueIndex" json:"email"`
	Username            string         `gorm:"size:255;uniqueIndex" json:"username"`
	Password            string         `gorm:"size:255" json:"-"`
	RememberToken       *string        `gorm:"size:100" json:"-"`
	Language            string         `gorm:"size:5;default:en" json:"language"`
	Timezone            string         `gorm:"size:50;default:UTC" json:"timezone"`
	ExternalID          *string        `gorm:"size:255;uniqueIndex" json:"external_id"`
	IsManagedExternally bool           `gorm:"default:false" json:"is_managed_externally"`
	OAuth               string         `gorm:"type:json;default:'[]'" json:"-"`
	Customization       *string        `gorm:"type:json" json:"customization"`
	MFAAppSecret        *string        `json:"-"`
	MFAAppRecoveryCodes *string        `gorm:"type:json" json:"-"`
	MFAEmailEnabled     bool           `gorm:"default:false" json:"mfa_email_enabled"`
	CreatedAt           time.Time      `json:"created_at"`
	UpdatedAt           time.Time      `json:"updated_at"`
	DeletedAt           gorm.DeletedAt `gorm:"index" json:"-"`

	Servers  []Server  `gorm:"foreignKey:OwnerID" json:"servers,omitempty"`
	APIKeys  []APIKey  `gorm:"foreignKey:UserID" json:"api_keys,omitempty"`
	SSHKeys  []SSHKey  `gorm:"foreignKey:UserID" json:"ssh_keys,omitempty"`
	Subusers []Subuser `gorm:"foreignKey:UserID" json:"subusers,omitempty"`
}

func (User) TableName() string { return "users" }
