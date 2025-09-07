package service

import (
	"context"
	"errors"
	"time"

	"cdk-office/internal/document/domain"
	"cdk-office/internal/shared/cache"
	"cdk-office/internal/shared/database"
	"cdk-office/internal/shared/utils"
	"cdk-office/pkg/logger"
	"gorm.io/gorm"
)

// DocumentServiceInterface defines the interface for document service
type DocumentServiceInterface interface {
	Upload(ctx context.Context, req *UploadRequest) (*domain.Document, error)
	GetDocument(ctx context.Context, docID string) (*domain.Document, error)
	UpdateDocument(ctx context.Context, docID string, req *UpdateRequest) error
	DeleteDocument(ctx context.Context, docID string) error
	GetDocumentVersions(ctx context.Context, docID string) ([]*domain.DocumentVersion, error)
}

// DocumentService implements the DocumentServiceInterface
type DocumentService struct {
	db *gorm.DB
}

// NewDocumentService creates a new instance of DocumentService
func NewDocumentService() *DocumentService {
	return &DocumentService{
		db: database.GetDB(),
	}
}

// UploadRequest represents the request for uploading a document
type UploadRequest struct {
	Title       string `json:"title" binding:"required"`
	Description string `json:"description"`
	FilePath    string `json:"file_path" binding:"required"`
	FileSize    int64  `json:"file_size" binding:"required"`
	MimeType    string `json:"mime_type" binding:"required"`
	OwnerID     string `json:"owner_id" binding:"required"`
	TeamID      string `json:"team_id" binding:"required"`
	Tags        string `json:"tags"`
}

// UpdateRequest represents the request for updating a document
type UpdateRequest struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Status      string `json:"status"`
	Tags        string `json:"tags"`
}

// Upload uploads a new document
func (s *DocumentService) Upload(ctx context.Context, req *UploadRequest) (*domain.Document, error) {
	// Create new document
	document := &domain.Document{
		ID:          utils.GenerateDocumentID(),
		Title:       req.Title,
		Description: req.Description,
		FilePath:    req.FilePath,
		FileSize:    req.FileSize,
		MimeType:    req.MimeType,
		OwnerID:     req.OwnerID,
		TeamID:      req.TeamID,
		Status:      "active",
		Tags:        req.Tags,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// Save document to database
	if err := s.db.Create(document).Error; err != nil {
		logger.Error("failed to create document", "error", err)
		return nil, errors.New("failed to upload document")
	}

	// Create first version of the document
	version := &domain.DocumentVersion{
		ID:         utils.GenerateDocumentVersionID(),
		DocumentID: document.ID,
		Version:    1,
		FilePath:   req.FilePath,
		FileSize:   req.FileSize,
		CreatedAt:  time.Now(),
	}

	if err := s.db.Create(version).Error; err != nil {
		logger.Error("failed to create document version", "error", err)
		// Rollback document creation
		s.db.Delete(document)
		return nil, errors.New("failed to upload document")
	}

	return document, nil
}

// GetDocument retrieves a document by ID
func (s *DocumentService) GetDocument(ctx context.Context, docID string) (*domain.Document, error) {
	// Try to get document from cache first
	cacheKey := "document:" + docID
	var document domain.Document
	
	// Check if document exists in cache
	exists, err := cache.Exists(cacheKey)
	if err == nil && exists {
		// Get document from cache
		if err := cache.Get(cacheKey, &document); err == nil {
			return &document, nil
		}
	}

	// Get document from database
	if err := s.db.Where("id = ?", docID).First(&document).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("document not found")
		}
		logger.Error("failed to find document", "error", err)
		return nil, errors.New("failed to get document")
	}

	// Cache the document for 10 minutes
	cache.Set(cacheKey, &document, 10*time.Minute)

	return &document, nil
}

// UpdateDocument updates a document
func (s *DocumentService) UpdateDocument(ctx context.Context, docID string, req *UpdateRequest) error {
	// Find document by ID
	var document domain.Document
	if err := s.db.Where("id = ?", docID).First(&document).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("document not found")
		}
		logger.Error("failed to find document", "error", err)
		return errors.New("failed to update document")
	}

	// Update document fields
	if req.Title != "" {
		document.Title = req.Title
	}
	if req.Description != "" {
		document.Description = req.Description
	}
	if req.Status != "" {
		document.Status = req.Status
	}
	if req.Tags != "" {
		document.Tags = req.Tags
	}
	document.UpdatedAt = time.Now()

	// Save updated document to database
	if err := s.db.Save(&document).Error; err != nil {
		logger.Error("failed to update document", "error", err)
		return errors.New("failed to update document")
	}

	// Invalidate cache
	cacheKey := "document:" + docID
	cache.Delete(cacheKey)

	return nil
}

// DeleteDocument deletes a document
func (s *DocumentService) DeleteDocument(ctx context.Context, docID string) error {
	// Find document by ID
	var document domain.Document
	if err := s.db.Where("id = ?", docID).First(&document).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("document not found")
		}
		logger.Error("failed to find document", "error", err)
		return errors.New("failed to delete document")
	}

	// Delete document from database
	if err := s.db.Delete(&document).Error; err != nil {
		logger.Error("failed to delete document", "error", err)
		return errors.New("failed to delete document")
	}

	// Delete all versions of the document
	if err := s.db.Where("document_id = ?", docID).Delete(&domain.DocumentVersion{}).Error; err != nil {
		logger.Error("failed to delete document versions", "error", err)
		// Note: In a real application, you might want to handle this error more gracefully
		// For now, we'll just log it and continue
	}

	// Invalidate cache
	cacheKey := "document:" + docID
	cache.Delete(cacheKey)

	return nil
}

// GetDocumentVersions retrieves all versions of a document
func (s *DocumentService) GetDocumentVersions(ctx context.Context, docID string) ([]*domain.DocumentVersion, error) {
	// Check if document exists
	var document domain.Document
	if err := s.db.Where("id = ?", docID).First(&document).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("document not found")
		}
		logger.Error("failed to find document", "error", err)
		return nil, errors.New("failed to get document versions")
	}

	// Try to get document versions from cache first
	cacheKey := "document_versions:" + docID
	var versions []*domain.DocumentVersion
	
	// Check if document versions exist in cache
	exists, err := cache.Exists(cacheKey)
	if err == nil && exists {
		// Get document versions from cache
		if err := cache.Get(cacheKey, &versions); err == nil {
			return versions, nil
		}
	}

	// Get document versions from database
	if err := s.db.Where("document_id = ?", docID).Order("version asc").Find(&versions).Error; err != nil {
		logger.Error("failed to find document versions", "error", err)
		return nil, errors.New("failed to get document versions")
	}

	// Cache the document versions for 10 minutes
	cache.Set(cacheKey, &versions, 10*time.Minute)

	return versions, nil
}