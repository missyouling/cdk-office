package service

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"cdk-office/internal/app/domain"
	"cdk-office/internal/shared/database"
	"cdk-office/internal/shared/utils"
	"cdk-office/pkg/logger"
	"github.com/yougg/go-qrcode"
	"gorm.io/gorm"
)

// BatchQRCodeServiceInterface defines the interface for batch QR code service
type BatchQRCodeServiceInterface interface {
	CreateBatchQRCode(ctx context.Context, req *CreateBatchQRCodeRequest) (*BatchQRCode, error)
	UpdateBatchQRCode(ctx context.Context, batchID string, req *UpdateBatchQRCodeRequest) error
	DeleteBatchQRCode(ctx context.Context, batchID string) error
	ListBatchQRCodes(ctx context.Context, appID string, page, size int) ([]*BatchQRCode, int64, error)
	GetBatchQRCode(ctx context.Context, batchID string) (*BatchQRCode, error)
	GenerateBatchQRCodes(ctx context.Context, batchID string) ([]*domain.QRCode, error)
}

// BatchQRCodeService implements the BatchQRCodeServiceInterface
type BatchQRCodeService struct {
	db *gorm.DB
}

// NewBatchQRCodeService creates a new instance of BatchQRCodeService
func NewBatchQRCodeService() *BatchQRCodeService {
	return &BatchQRCodeService{
		db: database.GetDB(),
	}
}

// CreateBatchQRCodeRequest represents the request for creating a batch QR code
type CreateBatchQRCodeRequest struct {
	AppID       string            `json:"app_id" binding:"required"`
	Name        string            `json:"name" binding:"required"`
	Description string            `json:"description"`
	Prefix      string            `json:"prefix"`
	Count       int               `json:"count" binding:"required"`
	Type        string            `json:"type" binding:"required"` // static or dynamic
	URLTemplate string            `json:"url_template"`
	Config      map[string]string `json:"config"`
	CreatedBy   string            `json:"created_by" binding:"required"`
}

// UpdateBatchQRCodeRequest represents the request for updating a batch QR code
type UpdateBatchQRCodeRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Prefix      string `json:"prefix"`
	URLTemplate string `json:"url_template"`
	Config      string `json:"config"`
}

// BatchQRCode represents the batch QR code entity
type BatchQRCode struct {
	ID          string    `json:"id"`
	AppID       string    `json:"app_id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Prefix      string    `json:"prefix"`
	Count       int       `json:"count"`
	Type        string    `json:"type"` // static or dynamic
	URLTemplate string    `json:"url_template"`
	Config      string    `json:"config"`
	Status      string    `json:"status"` // pending, generating, completed, failed
	CreatedBy   string    `json:"created_by"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// CreateBatchQRCode creates a new batch QR code
func (s *BatchQRCodeService) CreateBatchQRCode(ctx context.Context, req *CreateBatchQRCodeRequest) (*BatchQRCode, error) {
	// Validate QR code type
	if req.Type != "static" && req.Type != "dynamic" {
		return nil, errors.New("invalid QR code type")
	}

	// Validate count
	if req.Count <= 0 || req.Count > 10000 {
		return nil, errors.New("invalid count, must be between 1 and 10000")
	}

	// Create new batch QR code
	batch := &BatchQRCode{
		ID:          utils.GenerateBatchID(),
		AppID:       req.AppID,
		Name:        req.Name,
		Description: req.Description,
		Prefix:      req.Prefix,
		Count:       req.Count,
		Type:        req.Type,
		URLTemplate: req.URLTemplate,
		Status:      "pending",
		CreatedBy:   req.CreatedBy,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// Save batch QR code to database
	if err := s.db.Table("batch_qr_codes").Create(batch).Error; err != nil {
		logger.Error("failed to create batch QR code", "error", err)
		return nil, errors.New("failed to create batch QR code")
	}

	return batch, nil
}

// UpdateBatchQRCode updates an existing batch QR code
func (s *BatchQRCodeService) UpdateBatchQRCode(ctx context.Context, batchID string, req *UpdateBatchQRCodeRequest) error {
	// Find batch QR code by ID
	var batch BatchQRCode
	if err := s.db.Table("batch_qr_codes").Where("id = ?", batchID).First(&batch).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("batch QR code not found")
		}
		logger.Error("failed to find batch QR code", "error", err)
		return errors.New("failed to update batch QR code")
	}

	// Update batch QR code fields
	if req.Name != "" {
		batch.Name = req.Name
	}

	if req.Description != "" {
		batch.Description = req.Description
	}

	if req.Prefix != "" {
		batch.Prefix = req.Prefix
	}

	if req.URLTemplate != "" {
		batch.URLTemplate = req.URLTemplate
	}

	if req.Config != "" {
		batch.Config = req.Config
	}

	batch.UpdatedAt = time.Now()

	// Save updated batch QR code to database
	if err := s.db.Table("batch_qr_codes").Save(&batch).Error; err != nil {
		logger.Error("failed to update batch QR code", "error", err)
		return errors.New("failed to update batch QR code")
	}

	return nil
}

// DeleteBatchQRCode deletes a batch QR code
func (s *BatchQRCodeService) DeleteBatchQRCode(ctx context.Context, batchID string) error {
	// Find batch QR code by ID
	var batch BatchQRCode
	if err := s.db.Table("batch_qr_codes").Where("id = ?", batchID).First(&batch).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("batch QR code not found")
		}
		logger.Error("failed to find batch QR code", "error", err)
		return errors.New("failed to delete batch QR code")
	}

	// Delete batch QR code from database
	if err := s.db.Table("batch_qr_codes").Delete(&batch).Error; err != nil {
		logger.Error("failed to delete batch QR code", "error", err)
		return errors.New("failed to delete batch QR code")
	}

	return nil
}

// ListBatchQRCodes lists batch QR codes with pagination
func (s *BatchQRCodeService) ListBatchQRCodes(ctx context.Context, appID string, page, size int) ([]*BatchQRCode, int64, error) {
	// Validate pagination parameters
	if page < 1 {
		page = 1
	}
	if size < 1 || size > 100 {
		size = 10
	}

	// Build query
	dbQuery := s.db.Table("batch_qr_codes").Where("app_id = ?", appID)

	// Count total results
	var total int64
	if err := dbQuery.Count(&total).Error; err != nil {
		logger.Error("failed to count batch QR codes", "error", err)
		return nil, 0, errors.New("failed to list batch QR codes")
	}

	// Apply pagination
	offset := (page - 1) * size
	dbQuery = dbQuery.Offset(offset).Limit(size).Order("created_at desc")

	// Execute query
	var batches []*BatchQRCode
	if err := dbQuery.Find(&batches).Error; err != nil {
		logger.Error("failed to list batch QR codes", "error", err)
		return nil, 0, errors.New("failed to list batch QR codes")
	}

	return batches, total, nil
}

// GetBatchQRCode retrieves a batch QR code by ID
func (s *BatchQRCodeService) GetBatchQRCode(ctx context.Context, batchID string) (*BatchQRCode, error) {
	var batch BatchQRCode
	if err := s.db.Table("batch_qr_codes").Where("id = ?", batchID).First(&batch).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("batch QR code not found")
		}
		logger.Error("failed to find batch QR code", "error", err)
		return nil, errors.New("failed to get batch QR code")
	}

	return &batch, nil
}

// GenerateBatchQRCodes generates QR codes for a batch
func (s *BatchQRCodeService) GenerateBatchQRCodes(ctx context.Context, batchID string) ([]*domain.QRCode, error) {
	// Find batch QR code by ID
	var batch BatchQRCode
	if err := s.db.Table("batch_qr_codes").Where("id = ?", batchID).First(&batch).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("batch QR code not found")
		}
		logger.Error("failed to find batch QR code", "error", err)
		return nil, errors.New("failed to generate batch QR codes")
	}

	// Update batch status to generating
	batch.Status = "generating"
	batch.UpdatedAt = time.Now()
	if err := s.db.Table("batch_qr_codes").Save(&batch).Error; err != nil {
		logger.Error("failed to update batch status", "error", err)
	}

	// Generate QR codes
	var qrCodes []*domain.QRCode
	for i := 1; i <= batch.Count; i++ {
		// Create QR code name
		name := batch.Name
		if batch.Prefix != "" {
			name = batch.Prefix + "_" + name
		}
		name = fmt.Sprintf("%s_%d", name, i)

		// Create QR code content
		content := batch.URLTemplate
		if content == "" {
			content = fmt.Sprintf("https://example.com/%s/%d", batch.ID, i)
		} else {
			content = strings.ReplaceAll(batch.URLTemplate, "{index}", fmt.Sprintf("%d", i))
		}

		// Create QR code URL
		url := batch.URLTemplate
		if url == "" {
			url = fmt.Sprintf("https://example.com/%s/%d", batch.ID, i)
		} else {
			url = strings.ReplaceAll(batch.URLTemplate, "{index}", fmt.Sprintf("%d", i))
		}

		// Create new QR code
		qrCode := &domain.QRCode{
			ID:        utils.GenerateQRCodeID(),
			AppID:     batch.AppID,
			Name:      name,
			Content:   content,
			Type:      batch.Type,
			URL:       url,
			CreatedBy: batch.CreatedBy,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		// Save QR code to database
		if err := s.db.Create(qrCode).Error; err != nil {
			logger.Error("failed to create QR code", "error", err)
			// Update batch status to failed
			batch.Status = "failed"
			batch.UpdatedAt = time.Now()
			if err := s.db.Table("batch_qr_codes").Save(&batch).Error; err != nil {
				logger.Error("failed to update batch status", "error", err)
			}
			return nil, errors.New("failed to generate batch QR codes")
		}

		qrCodes = append(qrCodes, qrCode)
	}

	// Generate QR code images
	qrService := NewQRCodeService()
	for _, qrCode := range qrCodes {
		imagePath, err := qrService.GenerateQRCodeImage(ctx, qrCode.ID)
		if err != nil {
			logger.Error("failed to generate QR code image", "error", err)
			// Continue with other QR codes even if one fails
			continue
		}

		// Update QR code with image path
		qrCode.ImagePath = imagePath
		if err := s.db.Save(qrCode).Error; err != nil {
			logger.Error("failed to update QR code with image path", "error", err)
		}
	}

	// Update batch status to completed
	batch.Status = "completed"
	batch.UpdatedAt = time.Now()
	if err := s.db.Table("batch_qr_codes").Save(&batch).Error; err != nil {
		logger.Error("failed to update batch status", "error", err)
	}

	return qrCodes, nil
}



// generateQRCodeImage generates a QR code image and returns the image path
func (s *BatchQRCodeService) generateQRCodeImage(qrCode *domain.QRCode) (string, error) {
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