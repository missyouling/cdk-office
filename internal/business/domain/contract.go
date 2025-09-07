package domain

import (
	"time"
)

// Contract represents a contract in the system
type Contract struct {
	ID          string    `json:"id" gorm:"primaryKey"`
	TeamID      string    `json:"team_id" gorm:"index"`
	Title       string    `json:"title" gorm:"size:200"`
	Description string    `json:"description" gorm:"type:text"`
	Content     string    `json:"content" gorm:"type:text"`
	Status      string    `json:"status" gorm:"size:20"`
	CreatedBy   string    `json:"created_by" gorm:"size:50"`
	Signers     string    `json:"signers" gorm:"type:jsonb"`
	SignedBy    string    `json:"signed_by" gorm:"type:jsonb"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}