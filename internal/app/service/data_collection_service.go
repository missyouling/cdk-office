package service

import (
	"context"
	"errors"
	"time"

	"cdk-office/internal/shared/database"
	"cdk-office/pkg/logger"
	"gorm.io/gorm"
)

// DataCollectionServiceInterface defines the interface for data collection service
type DataCollectionServiceInterface interface {
	CreateDataCollection(ctx context.Context, req *CreateDataCollectionRequest) (*DataCollection, error)
	UpdateDataCollection(ctx context.Context, collectionID string, req *UpdateDataCollectionRequest) error
	DeleteDataCollection(ctx context.Context, collectionID string) error
	ListDataCollections(ctx context.Context, appID string, page, size int) ([]*DataCollection, int64, error)
	GetDataCollection(ctx context.Context, collectionID string) (*DataCollection, error)
	SubmitDataEntry(ctx context.Context, req *SubmitDataEntryRequest) (*DataCollectionEntry, error)
	ListDataEntries(ctx context.Context, collectionID string, page, size int) ([]*DataCollectionEntry, int64, error)
	ExportDataEntries(ctx context.Context, collectionID string) ([]*DataCollectionEntry, error)
}

// DataCollectionService implements the DataCollectionServiceInterface
type DataCollectionService struct {
	db *gorm.DB
}

// NewDataCollectionService creates a new instance of DataCollectionService
func NewDataCollectionService() *DataCollectionService {
	return &DataCollectionService{
		db: database.GetDB(),
	}
}

// CreateDataCollectionRequest represents the request for creating a data collection
type CreateDataCollectionRequest struct {
	AppID       string `json:"app_id" binding:"required"`
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
	Schema      string `json:"schema" binding:"required"`
	Config      string `json:"config"`
	CreatedBy   string `json:"created_by" binding:"required"`
}

// UpdateDataCollectionRequest represents the request for updating a data collection
type UpdateDataCollectionRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Schema      string `json:"schema"`
	Config      string `json:"config"`
	IsActive    *bool  `json:"is_active"`
}

// SubmitDataEntryRequest represents the request for submitting a data entry
type SubmitDataEntryRequest struct {
	CollectionID string `json:"collection_id" binding:"required"`
	Data         string `json:"data" binding:"required"`
	CreatedBy    string `json:"created_by" binding:"required"`
}

// DataCollection represents the data collection entity
type DataCollection struct {
	ID          string    `json:"id"`
	AppID       string    `json:"app_id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Schema      string    `json:"schema"`
	Config      string    `json:"config"`
	IsActive    bool      `json:"is_active"`
	CreatedBy   string    `json:"created_by"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// DataCollectionEntry represents a data entry in a collection
type DataCollectionEntry struct {
	ID          string    `json:"id"`
	CollectionID string    `json:"collection_id"`
	Data        string    `json:"data"`
	CreatedBy   string    `json:"created_by"`
	CreatedAt   time.Time `json:"created_at"`
}

// CreateDataCollection creates a new data collection
func (s *DataCollectionService) CreateDataCollection(ctx context.Context, req *CreateDataCollectionRequest) (*DataCollection, error) {
	// Create new data collection
	collection := &DataCollection{
		ID:          generateCollectionID(),
		AppID:       req.AppID,
		Name:        req.Name,
		Description: req.Description,
		Schema:      req.Schema,
		Config:      req.Config,
		IsActive:    true,
		CreatedBy:   req.CreatedBy,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// Save data collection to database
	if err := s.db.Table("data_collections").Create(collection).Error; err != nil {
		logger.Error("failed to create data collection", "error", err)
		return nil, errors.New("failed to create data collection")
	}

	return collection, nil
}

// UpdateDataCollection updates an existing data collection
func (s *DataCollectionService) UpdateDataCollection(ctx context.Context, collectionID string, req *UpdateDataCollectionRequest) error {
	// Find data collection by ID
	var collection DataCollection
	if err := s.db.Table("data_collections").Where("id = ?", collectionID).First(&collection).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("data collection not found")
		}
		logger.Error("failed to find data collection", "error", err)
		return errors.New("failed to update data collection")
	}

	// Update data collection fields
	if req.Name != "" {
		collection.Name = req.Name
	}

	if req.Description != "" {
		collection.Description = req.Description
	}

	if req.Schema != "" {
		collection.Schema = req.Schema
	}

	if req.Config != "" {
		collection.Config = req.Config
	}

	if req.IsActive != nil {
		collection.IsActive = *req.IsActive
	}

	collection.UpdatedAt = time.Now()

	// Save updated data collection to database
	if err := s.db.Table("data_collections").Save(&collection).Error; err != nil {
		logger.Error("failed to update data collection", "error", err)
		return errors.New("failed to update data collection")
	}

	return nil
}

// DeleteDataCollection deletes a data collection
func (s *DataCollectionService) DeleteDataCollection(ctx context.Context, collectionID string) error {
	// Find data collection by ID
	var collection DataCollection
	if err := s.db.Table("data_collections").Where("id = ?", collectionID).First(&collection).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("data collection not found")
		}
		logger.Error("failed to find data collection", "error", err)
		return errors.New("failed to delete data collection")
	}

	// Delete data entries associated with this collection
	if err := s.db.Table("data_collection_entries").Where("collection_id = ?", collectionID).Delete(&DataCollectionEntry{}).Error; err != nil {
		logger.Error("failed to delete data collection entries", "error", err)
		return errors.New("failed to delete data collection entries")
	}

	// Delete data collection from database
	if err := s.db.Table("data_collections").Delete(&collection).Error; err != nil {
		logger.Error("failed to delete data collection", "error", err)
		return errors.New("failed to delete data collection")
	}

	return nil
}

// ListDataCollections lists data collections with pagination
func (s *DataCollectionService) ListDataCollections(ctx context.Context, appID string, page, size int) ([]*DataCollection, int64, error) {
	// Validate pagination parameters
	if page < 1 {
		page = 1
	}
	if size < 1 || size > 100 {
		size = 10
	}

	// Build query
	dbQuery := s.db.Table("data_collections").Where("app_id = ?", appID)

	// Count total results
	var total int64
	if err := dbQuery.Count(&total).Error; err != nil {
		logger.Error("failed to count data collections", "error", err)
		return nil, 0, errors.New("failed to list data collections")
	}

	// Apply pagination
	offset := (page - 1) * size
	dbQuery = dbQuery.Offset(offset).Limit(size).Order("created_at desc")

	// Execute query
	var collections []*DataCollection
	if err := dbQuery.Find(&collections).Error; err != nil {
		logger.Error("failed to list data collections", "error", err)
		return nil, 0, errors.New("failed to list data collections")
	}

	return collections, total, nil
}

// GetDataCollection retrieves a data collection by ID
func (s *DataCollectionService) GetDataCollection(ctx context.Context, collectionID string) (*DataCollection, error) {
	var collection DataCollection
	if err := s.db.Table("data_collections").Where("id = ?", collectionID).First(&collection).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("data collection not found")
		}
		logger.Error("failed to find data collection", "error", err)
		return nil, errors.New("failed to get data collection")
	}

	return &collection, nil
}

// SubmitDataEntry submits a new data entry to a collection
func (s *DataCollectionService) SubmitDataEntry(ctx context.Context, req *SubmitDataEntryRequest) (*DataCollectionEntry, error) {
	// Verify data collection exists and is active
	var collection DataCollection
	if err := s.db.Table("data_collections").Where("id = ? AND is_active = ?", req.CollectionID, true).First(&collection).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("data collection not found or inactive")
		}
		logger.Error("failed to find data collection", "error", err)
		return nil, errors.New("failed to submit data entry")
	}

	// Create new data entry
	entry := &DataCollectionEntry{
		ID:          generateEntryID(),
		CollectionID: req.CollectionID,
		Data:        req.Data,
		CreatedBy:   req.CreatedBy,
		CreatedAt:   time.Now(),
	}

	// Save data entry to database
	if err := s.db.Table("data_collection_entries").Create(entry).Error; err != nil {
		logger.Error("failed to create data entry", "error", err)
		return nil, errors.New("failed to submit data entry")
	}

	return entry, nil
}

// ListDataEntries lists data entries with pagination
func (s *DataCollectionService) ListDataEntries(ctx context.Context, collectionID string, page, size int) ([]*DataCollectionEntry, int64, error) {
	// Validate pagination parameters
	if page < 1 {
		page = 1
	}
	if size < 1 || size > 100 {
		size = 10
	}

	// Verify data collection exists
	var collection DataCollection
	if err := s.db.Table("data_collections").Where("id = ?", collectionID).First(&collection).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, 0, errors.New("data collection not found")
		}
		logger.Error("failed to find data collection", "error", err)
		return nil, 0, errors.New("failed to list data entries")
	}

	// Build query
	dbQuery := s.db.Table("data_collection_entries").Where("collection_id = ?", collectionID)

	// Count total results
	var total int64
	if err := dbQuery.Count(&total).Error; err != nil {
		logger.Error("failed to count data entries", "error", err)
		return nil, 0, errors.New("failed to list data entries")
	}

	// Apply pagination
	offset := (page - 1) * size
	dbQuery = dbQuery.Offset(offset).Limit(size).Order("created_at desc")

	// Execute query
	var entries []*DataCollectionEntry
	if err := dbQuery.Find(&entries).Error; err != nil {
		logger.Error("failed to list data entries", "error", err)
		return nil, 0, errors.New("failed to list data entries")
	}

	return entries, total, nil
}

// ExportDataEntries exports all data entries from a collection
func (s *DataCollectionService) ExportDataEntries(ctx context.Context, collectionID string) ([]*DataCollectionEntry, error) {
	// Verify data collection exists
	var collection DataCollection
	if err := s.db.Table("data_collections").Where("id = ?", collectionID).First(&collection).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("data collection not found")
		}
		logger.Error("failed to find data collection", "error", err)
		return nil, errors.New("failed to export data entries")
	}

	// Get all data entries
	var entries []*DataCollectionEntry
	if err := s.db.Table("data_collection_entries").Where("collection_id = ?", collectionID).Order("created_at asc").Find(&entries).Error; err != nil {
		logger.Error("failed to export data entries", "error", err)
		return nil, errors.New("failed to export data entries")
	}

	return entries, nil
}

// generateCollectionID generates a unique collection ID
func generateCollectionID() string {
	// In a real application, use a proper ID generation library like uuid
	return "collection_" + time.Now().Format("20060102150405")
}

// generateEntryID generates a unique entry ID
func generateEntryID() string {
	// In a real application, use a proper ID generation library like uuid
	return "entry_" + time.Now().Format("20060102150405")
}