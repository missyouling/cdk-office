package domain

import (
	"time"
)

// DocumentCategory represents a category for documents
type DocumentCategory struct {
	ID          string    `json:"id" gorm:"primaryKey"`
	Name        string    `json:"name" gorm:"size:100;uniqueIndex"`
	Description string    `json:"description" gorm:"type:text"`
	ParentID    string    `json:"parent_id" gorm:"index"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// DocumentCategoryRelation represents the relationship between documents and categories
type DocumentCategoryRelation struct {
	ID         string    `json:"id" gorm:"primaryKey"`
	DocumentID string    `json:"document_id" gorm:"index"`
	CategoryID string    `json:"category_id" gorm:"index"`
	CreatedAt  time.Time `json:"created_at"`
}