package domain

import (
	"time"
)

// BusinessModule represents a business module in the system
type BusinessModule struct {
	ID          string    `json:"id" gorm:"primaryKey"`
	Name        string    `json:"name" gorm:"size:100;uniqueIndex"`
	Description string    `json:"description" gorm:"type:text"`
	Version     string    `json:"version" gorm:"size:20"`
	IsActive    bool      `json:"is_active"`
	Config      string    `json:"config" gorm:"type:jsonb"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// BusinessModulePermission represents permissions for a business module
type BusinessModulePermission struct {
	ID         string    `json:"id" gorm:"primaryKey"`
	ModuleID   string    `json:"module_id" gorm:"index"`
	RoleID     string    `json:"role_id" gorm:"index"`
	Permission string    `json:"permission" gorm:"size:50"`
	CreatedAt  time.Time `json:"created_at"`
}