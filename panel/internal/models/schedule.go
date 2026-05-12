package models

import (
	"time"

	"gorm.io/gorm"
)

type Schedule struct {
	ID             uint           `gorm:"primaryKey" json:"id"`
	ServerID       uint           `gorm:"index" json:"server_id"`
	Name           string         `gorm:"size:255" json:"name"`
	CronDayOfWeek  string         `gorm:"size:10" json:"cron_day_of_week"`
	CronDayOfMonth string         `gorm:"size:10" json:"cron_day_of_month"`
	CronHour       string         `gorm:"size:10" json:"cron_hour"`
	CronMinute     string         `gorm:"size:10" json:"cron_minute"`
	IsActive       bool           `gorm:"default:true" json:"is_active"`
	IsProcessing   bool           `gorm:"default:false" json:"is_processing"`
	LastRunAt      *time.Time     `json:"last_run_at"`
	NextRunAt      *time.Time     `json:"next_run_at"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
	DeletedAt      gorm.DeletedAt `gorm:"index" json:"-"`

	Server Server `gorm:"foreignKey:ServerID" json:"server,omitempty"`
	Tasks  []Task `gorm:"foreignKey:ScheduleID" json:"tasks,omitempty"`
}

func (Schedule) TableName() string { return "schedules" }

type Task struct {
	ID                 uint           `gorm:"primaryKey" json:"id"`
	ScheduleID         uint           `gorm:"index" json:"schedule_id"`
	SequenceID         int            `json:"sequence_id"`
	Action             string         `gorm:"size:50" json:"action"`
	Payload            string         `gorm:"type:text" json:"payload"`
	TimeOffset         int            `gorm:"default:0" json:"time_offset"`
	IsQueued           bool           `gorm:"default:false" json:"is_queued"`
	ContinuedOnFailure bool           `gorm:"default:false" json:"continued_on_failure"`
	CreatedAt          time.Time      `json:"created_at"`
	UpdatedAt          time.Time      `json:"updated_at"`
	DeletedAt          gorm.DeletedAt `gorm:"index" json:"-"`

	Schedule Schedule `gorm:"foreignKey:ScheduleID" json:"schedule,omitempty"`
}

func (Task) TableName() string { return "tasks" }
