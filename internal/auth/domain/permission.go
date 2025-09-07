package domain

import (
	"time"
)

// Permission represents a permission in the system
type Permission struct {
	ID          string    `json:"id" gorm:"primaryKey"`
	Name        string    `json:"name" gorm:"size:100;uniqueIndex"`
	Resource    string    `json:"resource" gorm:"size:100;index"`
	Action      string    `json:"action" gorm:"size:50;index"`
	Description string    `json:"description" gorm:"size:255"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// Role represents a role in the system
type Role struct {
	ID          string    `json:"id" gorm:"primaryKey"`
	Name        string    `json:"name" gorm:"size:50;uniqueIndex"`
	Description string    `json:"description" gorm:"size:255"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// RolePermission represents the relationship between roles and permissions
type RolePermission struct {
	ID           string    `json:"id" gorm:"primaryKey"`
	RoleID       string    `json:"role_id" gorm:"index"`
	PermissionID string    `json:"permission_id" gorm:"index"`
	CreatedAt    time.Time `json:"created_at"`
}