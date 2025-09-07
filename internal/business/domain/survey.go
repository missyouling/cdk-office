package domain

import (
	"time"
)

// Survey represents a survey in the system
type Survey struct {
	ID          string    `json:"id" gorm:"primaryKey"`
	TeamID      string    `json:"team_id" gorm:"index"`
	Title       string    `json:"title" gorm:"size:200"`
	Description string    `json:"description" gorm:"type:text"`
	Status      string    `json:"status" gorm:"size:20"`
	CreatedBy   string    `json:"created_by" gorm:"size:50"`
	Questions   string    `json:"questions" gorm:"type:jsonb"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// SurveyResponse represents a survey response in the system
type SurveyResponse struct {
	ID         string    `json:"id" gorm:"primaryKey"`
	SurveyID   string    `json:"survey_id" gorm:"index"`
	UserID     string    `json:"user_id" gorm:"size:50;index"`
	Answers    string    `json:"answers" gorm:"type:jsonb"`
	CreatedAt  time.Time `json:"created_at"`
}