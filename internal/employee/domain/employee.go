package domain

import (
	"time"
)

// Employee represents an employee in the system
type Employee struct {
	ID            string    `json:"id" gorm:"primaryKey"`
	UserID        string    `json:"user_id" gorm:"index"`
	TeamID        string    `json:"team_id" gorm:"index"`
	DeptID        string    `json:"dept_id" gorm:"index"`
	EmployeeID    string    `json:"employee_id" gorm:"size:50;uniqueIndex"`
	RealName      string    `json:"real_name" gorm:"size:50"`
	Gender        string    `json:"gender" gorm:"size:10"`
	BirthDate     time.Time `json:"birth_date"`
	HireDate      time.Time `json:"hire_date"`
	Position      string    `json:"position" gorm:"size:100"`
	Status        string    `json:"status" gorm:"size:20"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// Department represents a department in the organization
type Department struct {
	ID          string    `json:"id" gorm:"primaryKey"`
	Name        string    `json:"name" gorm:"size:100;uniqueIndex"`
	Description string    `json:"description" gorm:"type:text"`
	TeamID      string    `json:"team_id" gorm:"index"`
	ParentID    string    `json:"parent_id" gorm:"index"`
	Level       int       `json:"level"`
	SortOrder   int       `json:"sort_order"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}