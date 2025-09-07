package domain

import (
	"time"
)

// BusinessPermission represents a business permission in the system
type BusinessPermission struct {
	ID          string    `json:"id" gorm:"primaryKey"`
	Name        string    `json:"name" gorm:"size:100;uniqueIndex"`
	Description string    `json:"description" gorm:"type:text"`
	Resource    string    `json:"resource" gorm:"size:100;index"`
	Action      string    `json:"action" gorm:"size:50;index"`
	CreatedBy   string    `json:"created_by" gorm:"size:50"`
	TeamID      string    `json:"team_id" gorm:"index"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}