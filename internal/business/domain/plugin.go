package domain

import (
	"time"
)

// Plugin represents a plugin in the system
type Plugin struct {
	ID          string    `json:"id" gorm:"primaryKey"`
	Name        string    `json:"name" gorm:"size:100;uniqueIndex"`
	Description string    `json:"description" gorm:"type:text"`
	Version     string    `json:"version" gorm:"size:20"`
	Status      string    `json:"status" gorm:"size:20"`
	CreatedBy   string    `json:"created_by" gorm:"size:50"`
	TeamID      string    `json:"team_id" gorm:"index"`
	Config      string    `json:"config" gorm:"type:jsonb"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}