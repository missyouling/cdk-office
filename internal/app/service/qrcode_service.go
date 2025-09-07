package service

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"time"

	"cdk-office/internal/app/domain"
	"cdk-office/internal/shared/database"
	"cdk-office/internal/shared/utils"
	"cdk-office/pkg/logger"
	"github.com/yougg/go-qrcode"
	"gorm.io/gorm"
)

// QRCodeServiceInterface defines the interface for QR code service
type QRCodeServiceInterface interface {
	CreateQRCode(ctx context.Context, req *CreateQRCodeRequest) (*domain.QRCode, error)
	UpdateQRCode(ctx context.Context, qrCodeID string, req *UpdateQRCodeRequest) error
	DeleteQRCode(ctx context.Context, qrCodeID string) error
	ListQRCodes(ctx context.Context, appID string, page, size int) ([]*domain.QRCode, int64, error)
	GetQRCode(ctx context.Context, qrCodeID string) (*domain.QRCode, error)
	GenerateQRCodeImage(ctx context.Context, qrCodeID string) (string, error)
}

// QRCodeService implements the QRCodeServiceInterface
type QRCodeService struct {
	db *gorm.DB
}

// NewQRCodeService creates a new instance of QRCodeService
func NewQRCodeService() *QRCodeService {
	return &QRCodeService{
		db: database.GetDB(),
	}
}

// CreateQRCodeRequest represents the request for creating a QR code
type CreateQRCodeRequest struct {
	AppID     string `json:"app_id" binding:"required"`
	Name      string `json:"name" binding:"required"`
	Content   string `json:"content" binding:"required"`
	Type      string `json:"type" binding:"required"` // static or dynamic
	URL       string `json:"url"`
	CreatedBy string `json:"created_by" binding:"required"`
}

// UpdateQRCodeRequest represents the request for updating a QR code
type UpdateQRCodeRequest struct {
	Name    string `json:"name"`
	Content string `json:"content"`
	URL     string `json:"url"`
}

// CreateQRCode creates a new QR code
func (s *QRCodeService) CreateQRCode(ctx context.Context, req *CreateQRCodeRequest) (*domain.QRCode, error) {
	// Validate QR code type
	if req.Type != "static" && req.Type != "dynamic" {
		return nil, errors.New("invalid QR code type")
	}

	// Create new QR code
	qrCode := &domain.QRCode{
		ID:        utils.GenerateQRCodeID(),
		AppID:     req.AppID,
		Name:      req.Name,
		Content:   req.Content,
		Type:      req.Type,
		URL:       req.URL,
		CreatedBy: req.CreatedBy,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Save QR code to database
	if err := s.db.Create(qrCode).Error; err != nil {
		logger.Error("failed to create QR code", "error", err)
		return nil, errors.New("failed to create QR code")
	}

	// Generate QR code image
	imagePath, err := s.GenerateQRCodeImage(ctx, qrCode.ID)
	if err != nil {
		// Log error but don't fail the entire operation
		logger.Error("failed to generate QR code image", "error", err)
	} else {
		// Update QR code with image path
		qrCode.ImagePath = imagePath
		if err := s.db.Save(qrCode).Error; err != nil {
			logger.Error("failed to update QR code with image path", "error", err)
		}
	}

	return qrCode, nil
}

// UpdateQRCode updates an existing QR code
func (s *QRCodeService) UpdateQRCode(ctx context.Context, qrCodeID string, req *UpdateQRCodeRequest) error {
	// Find QR code by ID
	var qrCode domain.QRCode
	if err := s.db.Where("id = ?", qrCodeID).First(&qrCode).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("QR code not found")
		}
		logger.Error("failed to find QR code", "error", err)
		return errors.New("failed to update QR code")
	}

	// Update QR code fields
	if req.Name != "" {
		qrCode.Name = req.Name
	}
	
	if req.Content != "" {
		qrCode.Content = req.Content
	}
	
	if req.URL != "" {
		qrCode.URL = req.URL
	}
	
	qrCode.UpdatedAt = time.Now()

	// Save updated QR code to database
	if err := s.db.Save(&qrCode).Error; err != nil {
		logger.Error("failed to update QR code", "error", err)
		return errors.New("failed to update QR code")
	}

	// Regenerate QR code image
	imagePath, err := s.GenerateQRCodeImage(ctx, qrCode.ID)
	if err != nil {
		// Log error but don't fail the entire operation
		logger.Error("failed to regenerate QR code image", "error", err)
	} else {
		// Update QR code with new image path
		qrCode.ImagePath = imagePath
		if err := s.db.Save(&qrCode).Error; err != nil {
			logger.Error("failed to update QR code with new image path", "error", err)
		}
	}

	return nil
}

// DeleteQRCode deletes a QR code
func (s *QRCodeService) DeleteQRCode(ctx context.Context, qrCodeID string) error {
	// Find QR code by ID
	var qrCode domain.QRCode
	if err := s.db.Where("id = ?", qrCodeID).First(&qrCode).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("QR code not found")
		}
		logger.Error("failed to find QR code", "error", err)
		return errors.New("failed to delete QR code")
	}

	// Delete QR code from database
	if err := s.db.Delete(&qrCode).Error; err != nil {
		logger.Error("failed to delete QR code", "error", err)
		return errors.New("failed to delete QR code")
	}

	return nil
}

// ListQRCodes lists QR codes with pagination
func (s *QRCodeService) ListQRCodes(ctx context.Context, appID string, page, size int) ([]*domain.QRCode, int64, error) {
	// Validate pagination parameters
	if page < 1 {
		page = 1
	}
	if size < 1 || size > 100 {
		size = 10
	}

	// Build query
	dbQuery := s.db.Model(&domain.QRCode{}).Where("app_id = ?", appID)

	// Count total results
	var total int64
	if err := dbQuery.Count(&total).Error; err != nil {
		logger.Error("failed to count QR codes", "error", err)
		return nil, 0, errors.New("failed to list QR codes")
	}

	// Apply pagination
	offset := (page - 1) * size
	dbQuery = dbQuery.Offset(offset).Limit(size).Order("created_at desc")

	// Execute query
	var qrCodes []*domain.QRCode
	if err := dbQuery.Find(&qrCodes).Error; err != nil {
		logger.Error("failed to list QR codes", "error", err)
		return nil, 0, errors.New("failed to list QR codes")
	}

	return qrCodes, total, nil
}

// GetQRCode retrieves a QR code by ID
func (s *QRCodeService) GetQRCode(ctx context.Context, qrCodeID string) (*domain.QRCode, error) {
	var qrCode domain.QRCode
	if err := s.db.Where("id = ?", qrCodeID).First(&qrCode).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("QR code not found")
		}
		logger.Error("failed to find QR code", "error", err)
		return nil, errors.New("failed to get QR code")
	}

	return &qrCode, nil
}

// GenerateQRCodeImage generates a QR code image
func (s *QRCodeService) GenerateQRCodeImage(ctx context.Context, qrCodeID string) (string, error) {
	// Find QR code by ID
	var qrCode domain.QRCode
	if err := s.db.Where("id = ?", qrCodeID).First(&qrCode).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", errors.New("QR code not found")
		}
		logger.Error("failed to find QR code", "error", err)
		return "", errors.New("failed to generate QR code image")
	}

	// Create the directory for QR code images if it doesn't exist
	imageDir := "/tmp/qrcodes"
	if err := os.MkdirAll(imageDir, 0755); err != nil {
		logger.Error("failed to create QR code directory", "error", err)
		return "", errors.New("failed to create QR code directory")
	}

	// Generate QR code image using the go-qrcode library
	qr, err := qrcode.New(qrCode.Content, qrcode.High)
	if err != nil {
		logger.Error("failed to create QR code", "error", err)
		return "", errors.New("failed to generate QR code image")
	}

	// Save the QR code image to a file
	imagePath := filepath.Join(imageDir, qrCode.ID+".png")
	if err := qr.WriteFile(256, imagePath); err != nil {
		logger.Error("failed to save QR code image", "error", err)
		return "", errors.New("failed to save QR code image")
	}

	return imagePath, nil
}

