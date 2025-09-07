package service

import (
	"context"
	"errors"
	"time"

	"cdk-office/internal/document/domain"
	"cdk-office/internal/shared/database"
	"cdk-office/pkg/logger"
	"gorm.io/gorm"
)

// VersionServiceInterface defines the interface for document version service
type VersionServiceInterface interface {
	CreateVersion(ctx context.Context, documentID, filePath string, fileSize int64) (*domain.DocumentVersion, error)
	GetVersion(ctx context.Context, versionID string) (*domain.DocumentVersion, error)
	ListVersions(ctx context.Context, documentID string) ([]*domain.DocumentVersion, error)
	GetLatestVersion(ctx context.Context, documentID string) (*domain.DocumentVersion, error)
	RestoreVersion(ctx context.Context, versionID string) error
}

// VersionService implements the VersionServiceInterface
type VersionService struct {
	db *gorm.DB
}

// NewVersionService creates a new instance of VersionService
func NewVersionService() *VersionService {
	return &VersionService{
		db: database.GetDB(),
	}
}

// CreateVersion creates a new version of a document
func (s *VersionService) CreateVersion(ctx context.Context, documentID, filePath string, fileSize int64) (*domain.DocumentVersion, error) {
	// Check if document exists
	var document domain.Document
	if err := s.db.Where("id = ?", documentID).First(&document).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("document not found")
		}
		logger.Error("failed to find document", "error", err)
		return nil, errors.New("failed to create version")
	}

	// Get the latest version number for this document
	var latestVersion domain.DocumentVersion
	var versionNumber int
	if err := s.db.Where("document_id = ?", documentID).Order("version desc").First(&latestVersion).Error; err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			logger.Error("failed to find latest version", "error", err)
			return nil, errors.New("failed to create version")
		}
		// If no previous version, start with version 1
		versionNumber = 1
	} else {
		// Increment version number
		versionNumber = latestVersion.Version + 1
	}

	// Create new version
	version := &domain.DocumentVersion{
		ID:         generateID(),
		DocumentID: documentID,
		Version:    versionNumber,
		FilePath:   filePath,
		FileSize:   fileSize,
		CreatedAt:  time.Now(),
	}

	// Save version to database
	if err := s.db.Create(version).Error; err != nil {
		logger.Error("failed to create version", "error", err)
		return nil, errors.New("failed to create version")
	}

	return version, nil
}

// GetVersion retrieves a specific version of a document
func (s *VersionService) GetVersion(ctx context.Context, versionID string) (*domain.DocumentVersion, error) {
	var version domain.DocumentVersion
	if err := s.db.Where("id = ?", versionID).First(&version).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("version not found")
		}
		logger.Error("failed to find version", "error", err)
		return nil, errors.New("failed to get version")
	}

	return &version, nil
}

// ListVersions lists all versions of a document
func (s *VersionService) ListVersions(ctx context.Context, documentID string) ([]*domain.DocumentVersion, error) {
	// Check if document exists
	var document domain.Document
	if err := s.db.Where("id = ?", documentID).First(&document).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("document not found")
		}
		logger.Error("failed to find document", "error", err)
		return nil, errors.New("failed to list versions")
	}

	// Get all versions for this document
	var versions []*domain.DocumentVersion
	if err := s.db.Where("document_id = ?", documentID).Order("version asc").Find(&versions).Error; err != nil {
		logger.Error("failed to find versions", "error", err)
		return nil, errors.New("failed to list versions")
	}

	return versions, nil
}

// GetLatestVersion retrieves the latest version of a document
func (s *VersionService) GetLatestVersion(ctx context.Context, documentID string) (*domain.DocumentVersion, error) {
	// Check if document exists
	var document domain.Document
	if err := s.db.Where("id = ?", documentID).First(&document).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("document not found")
		}
		logger.Error("failed to find document", "error", err)
		return nil, errors.New("failed to get latest version")
	}

	// Get the latest version for this document
	var version domain.DocumentVersion
	if err := s.db.Where("document_id = ?", documentID).Order("version desc").First(&version).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("no versions found for document")
		}
		logger.Error("failed to find latest version", "error", err)
		return nil, errors.New("failed to get latest version")
	}

	return &version, nil
}

// RestoreVersion restores a document to a specific version
func (s *VersionService) RestoreVersion(ctx context.Context, versionID string) error {
	// Get the version to restore
	var version domain.DocumentVersion
	if err := s.db.Where("id = ?", versionID).First(&version).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("version not found")
		}
		logger.Error("failed to find version", "error", err)
		return errors.New("failed to restore version")
	}

	// Get the document
	var document domain.Document
	if err := s.db.Where("id = ?", version.DocumentID).First(&document).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("document not found")
		}
		logger.Error("failed to find document", "error", err)
		return errors.New("failed to restore version")
	}

	// Update document with version details
	document.FilePath = version.FilePath
	document.FileSize = version.FileSize
	document.UpdatedAt = time.Now()

	// Save updated document to database
	if err := s.db.Save(&document).Error; err != nil {
		logger.Error("failed to update document", "error", err)
		return errors.New("failed to restore version")
	}

	// Create a new version to record the restoration
	// This creates a new version with the same content as the restored version
	newVersion := &domain.DocumentVersion{
		ID:         generateID(),
		DocumentID: document.ID,
		Version:    version.Version + 1,
		FilePath:   version.FilePath,
		FileSize:   version.FileSize,
		CreatedAt:  time.Now(),
	}

	if err := s.db.Create(newVersion).Error; err != nil {
		logger.Error("failed to create new version after restoration", "error", err)
		// Note: In a real application, you might want to handle this error more gracefully
		// For now, we'll just log it and continue
	}

	return nil
}