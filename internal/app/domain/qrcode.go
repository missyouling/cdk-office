package domain

import (
	"time"
)

// QRCode represents a QR code entity
type QRCode struct {
	ID        string    `json:"id" gorm:"primaryKey"`
	AppID     string    `json:"app_id" gorm:"index"`
	Name      string    `json:"name"`
	Content   string    `json:"content"`
	Type      string    `json:"type"` // static or dynamic
	URL       string    `json:"url"`
	ImagePath string    `json:"image_path"`
	CreatedBy string    `json:"created_by"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}