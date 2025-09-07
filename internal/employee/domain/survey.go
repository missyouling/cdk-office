package domain

import (
	"time"
)

// EmployeeSurvey represents an employee survey
type EmployeeSurvey struct {
	ID            string    `json:"id" gorm:"primaryKey"`
	Title         string    `json:"title" gorm:"size:200"`
	Description   string    `json:"description" gorm:"type:text"`
	SurveyType    string    `json:"survey_type" gorm:"size:50"`
	CreatedBy     string    `json:"created_by" gorm:"size:50"`
	StartDate     time.Time `json:"start_date"`
	EndDate       time.Time `json:"end_date"`
	Status        string    `json:"status" gorm:"size:20"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// SurveyResponse represents a response to an employee survey
type SurveyResponse struct {
	ID            string    `json:"id" gorm:"primaryKey"`
	SurveyID      string    `json:"survey_id" gorm:"index"`
	EmployeeID    string    `json:"employee_id" gorm:"index"`
	Responses     string    `json:"responses" gorm:"type:jsonb"`
	SubmittedAt   time.Time `json:"submitted_at"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// SurveyQuestion represents a question in an employee survey
type SurveyQuestion struct {
	ID            string    `json:"id" gorm:"primaryKey"`
	SurveyID      string    `json:"survey_id" gorm:"index"`
	QuestionText  string    `json:"question_text" gorm:"type:text"`
	QuestionType  string    `json:"question_type" gorm:"size:50"`
	Options       string    `json:"options" gorm:"type:jsonb"`
	Required      bool      `json:"required"`
	OrderNumber   int       `json:"order_number"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}