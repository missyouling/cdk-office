package domain

import (
	"time"
)

// Document represents a document in the system
type Document struct {
	ID          string    `json:"id" gorm:"primaryKey"`
	Title       string    `json:"title" gorm:"size:200"`
	Description string    `json:"description" gorm:"type:text"`
	FilePath    string    `json:"file_path" gorm:"size:500"`
	FileSize    int64     `json:"file_size"`
	MimeType    string    `json:"mime_type" gorm:"size:100"`
	OwnerID     string    `json:"owner_id" gorm:"index"`
	TeamID      string    `json:"team_id" gorm:"index"`
	Status      string    `json:"status" gorm:"size:20"`
	Tags        string    `json:"tags" gorm:"type:jsonb"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// DocumentVersion represents a version of a document
type DocumentVersion struct {
	ID         string    `json:"id" gorm:"primaryKey"`
	DocumentID string    `json:"document_id" gorm:"index"`
	Version    int       `json:"version"`
	FilePath   string    `json:"file_path" gorm:"size:500"`
	FileSize   int64     `json:"file_size"`
	CreatedAt  time.Time `json:"created_at"`
}