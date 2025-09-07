package domain

import (
	"time"
)

// User represents a user in the system
type User struct {
	ID          string    `json:"id" gorm:"primaryKey"`
	Username    string    `json:"username" gorm:"uniqueIndex;size:50"`
	Email       string    `json:"email" gorm:"uniqueIndex;size:100"`
	Phone       string    `json:"phone" gorm:"size:20"`
	Password    string    `json:"-" gorm:"size:255"`
	RealName    string    `json:"real_name" gorm:"size:50"`
	IDCard      string    `json:"id_card" gorm:"size:18"`
	Role        string    `json:"role" gorm:"size:20"`
	Status      string    `json:"status" gorm:"size:20"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// UserRole represents a user role
type UserRole struct {
	ID          string    `json:"id" gorm:"primaryKey"`
	UserID      string    `json:"user_id" gorm:"index"`
	Role        string    `json:"role" gorm:"size:20"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}