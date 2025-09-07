package domain

import (
	"time"
)

// ExitInterview represents an exit interview for a terminated employee
type ExitInterview struct {
	ID            string    `json:"id" gorm:"primaryKey"`
	EmployeeID    string    `json:"employee_id" gorm:"index"`
	InterviewerID string    `json:"interviewer_id" gorm:"size:50"`
	ExitDate      time.Time `json:"exit_date"`
	Reason        string    `json:"reason" gorm:"size:100"`
	Comments      string    `json:"comments" gorm:"type:text"`
	Feedback      string    `json:"feedback" gorm:"type:text"`
	Rehireable    bool      `json:"rehireable"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// TerminationRecord represents a termination record for an employee
type TerminationRecord struct {
	ID            string    `json:"id" gorm:"primaryKey"`
	EmployeeID    string    `json:"employee_id" gorm:"index"`
	TerminationDate time.Time `json:"termination_date"`
	Reason        string    `json:"reason" gorm:"size:100"`
	Comments      string    `json:"comments" gorm:"type:text"`
	ExitType      string    `json:"exit_type" gorm:"size:50"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}