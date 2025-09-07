package domain

import (
	"time"
)

// PerformanceReview represents an employee performance review
type PerformanceReview struct {
	ID            string    `json:"id" gorm:"primaryKey"`
	EmployeeID    string    `json:"employee_id" gorm:"index"`
	ReviewerID    string    `json:"reviewer_id" gorm:"size:50"`
	ReviewDate    time.Time `json:"review_date"`
	ReviewPeriod  string    `json:"review_period" gorm:"size:50"`
	Score         float64   `json:"score"`
	Comments      string    `json:"comments" gorm:"type:text"`
	Goals         string    `json:"goals" gorm:"type:text"`
	Improvements  string    `json:"improvements" gorm:"type:text"`
	Status        string    `json:"status" gorm:"size:20"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// PerformanceMetric represents a performance metric for an employee
type PerformanceMetric struct {
	ID            string    `json:"id" gorm:"primaryKey"`
	EmployeeID    string    `json:"employee_id" gorm:"index"`
	MetricName    string    `json:"metric_name" gorm:"size:100"`
	TargetValue   float64   `json:"target_value"`
	ActualValue   float64   `json:"actual_value"`
	Weight        float64   `json:"weight"`
	MeasurementDate time.Time `json:"measurement_date"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// PerformanceGoal represents a performance goal for an employee
type PerformanceGoal struct {
	ID            string    `json:"id" gorm:"primaryKey"`
	EmployeeID    string    `json:"employee_id" gorm:"index"`
	GoalName      string    `json:"goal_name" gorm:"size:200"`
	Description   string    `json:"description" gorm:"type:text"`
	StartDate     time.Time `json:"start_date"`
	EndDate       time.Time `json:"end_date"`
	TargetValue   float64   `json:"target_value"`
	ActualValue   float64   `json:"actual_value"`
	Status        string    `json:"status" gorm:"size:20"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}