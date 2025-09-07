package domain

import (
	"time"
)

// Application represents an application in the system
type Application struct {
	ID          string    `json:"id" gorm:"primaryKey"`
	TeamID      string    `json:"team_id" gorm:"index"`
	Name        string    `json:"name" gorm:"size:100;uniqueIndex:idx_team_app"`
	Description string    `json:"description" gorm:"type:text"`
	Type        string    `json:"type" gorm:"size:50"`
	Config      string    `json:"config" gorm:"type:jsonb"`
	IsActive    bool      `json:"is_active"`
	CreatedBy   string    `json:"created_by" gorm:"size:50"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// QRCode represents a QR code in the system
type QRCode struct {
	ID          string    `json:"id" gorm:"primaryKey"`
	AppID       string    `json:"app_id" gorm:"index"`
	Name        string    `json:"name" gorm:"size:100"`
	Content     string    `json:"content" gorm:"type:text"`
	Type        string    `json:"type" gorm:"size:20"` // static or dynamic
	URL         string    `json:"url" gorm:"size:500"`
	ImagePath   string    `json:"image_path" gorm:"size:500"`
	CreatedBy   string    `json:"created_by" gorm:"size:50"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// FormData represents a form in the system
type FormData struct {
	ID          string    `json:"id" gorm:"primaryKey"`
	AppID       string    `json:"app_id" gorm:"index"`
	Name        string    `json:"name" gorm:"size:100"`
	Description string    `json:"description" gorm:"type:text"`
	Schema      string    `json:"schema" gorm:"type:jsonb"`
	IsActive    bool      `json:"is_active"`
	CreatedBy   string    `json:"created_by" gorm:"size:50"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// FormDataEntry represents a form data entry in the system
type FormDataEntry struct {
	ID        string    `json:"id" gorm:"primaryKey"`
	FormID    string    `json:"form_id" gorm:"index"`
	Data      string    `json:"data" gorm:"type:jsonb"`
	CreatedBy string    `json:"created_by" gorm:"size:50"`
	CreatedAt time.Time `json:"created_at"`
}

// AppPermission represents an application permission in the system
type AppPermission struct {
	ID          string    `json:"id" gorm:"primaryKey"`
	AppID       string    `json:"app_id" gorm:"index"`
	Name        string    `json:"name" gorm:"size:100"`
	Description string    `json:"description" gorm:"type:text"`
	Permission  string    `json:"permission" gorm:"size:20"` // read, write, delete, manage
	CreatedBy   string    `json:"created_by" gorm:"size:50"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// AppUserPermission represents the relationship between user and permission
type AppUserPermission struct {
	ID           string    `json:"id" gorm:"primaryKey"`
	AppID        string    `json:"app_id" gorm:"index"`
	UserID       string    `json:"user_id" gorm:"index"`
	PermissionID string    `json:"permission_id" gorm:"index"`
	AssignedBy   string    `json:"assigned_by" gorm:"size:50"`
	CreatedAt    time.Time `json:"created_at"`
}

// BatchQRCode represents a batch of QR codes in the system
type BatchQRCode struct {
	ID          string    `json:"id" gorm:"primaryKey"`
	AppID       string    `json:"app_id" gorm:"index"`
	Name        string    `json:"name" gorm:"size:100"`
	Description string    `json:"description" gorm:"type:text"`
	Status      string    `json:"status" gorm:"size:20"` // pending, generating, completed, failed
	TotalCount  int       `json:"total_count"`
	CreatedBy   string    `json:"created_by" gorm:"size:50"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// BatchQRCodeItem represents an item in a batch of QR codes
type BatchQRCodeItem struct {
	ID        string    `json:"id" gorm:"primaryKey"`
	BatchID   string    `json:"batch_id" gorm:"index"`
	Name      string    `json:"name" gorm:"size:100"`
	Content   string    `json:"content" gorm:"type:text"`
	URL       string    `json:"url" gorm:"size:500"`
	ImagePath string    `json:"image_path" gorm:"size:500"`
	Status    string    `json:"status" gorm:"size:20"` // pending, generating, completed, failed
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// FormDesign represents a form design in the system
type FormDesign struct {
	ID          string    `json:"id" gorm:"primaryKey"`
	AppID       string    `json:"app_id" gorm:"index"`
	Name        string    `json:"name" gorm:"size:100"`
	Description string    `json:"description" gorm:"type:text"`
	Schema      string    `json:"schema" gorm:"type:jsonb"`
	Config      string    `json:"config" gorm:"type:jsonb"`
	IsActive    bool      `json:"is_active"`
	IsPublished bool      `json:"is_published"`
	CreatedBy   string    `json:"created_by" gorm:"size:50"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// DataCollection represents a data collection in the system
type DataCollection struct {
	ID          string    `json:"id" gorm:"primaryKey"`
	AppID       string    `json:"app_id" gorm:"index"`
	Name        string    `json:"name" gorm:"size:100"`
	Description string    `json:"description" gorm:"type:text"`
	Schema      string    `json:"schema" gorm:"type:jsonb"`
	Config      string    `json:"config" gorm:"type:jsonb"`
	IsActive    bool      `json:"is_active"`
	CreatedBy   string    `json:"created_by" gorm:"size:50"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// DataCollectionEntry represents a data entry in a collection
type DataCollectionEntry struct {
	ID          string    `json:"id" gorm:"primaryKey"`
	CollectionID string    `json:"collection_id" gorm:"index"`
	Data        string    `json:"data" gorm:"type:jsonb"`
	CreatedBy   string    `json:"created_by" gorm:"size:50"`
	CreatedAt   time.Time `json:"created_at"`
}