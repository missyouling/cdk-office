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

// CategoryServiceInterface defines the interface for category service
type CategoryServiceInterface interface {
	CreateCategory(ctx context.Context, name, description, parentID string) (*domain.DocumentCategory, error)
	GetCategory(ctx context.Context, categoryID string) (*domain.DocumentCategory, error)
	UpdateCategory(ctx context.Context, categoryID, name, description string) error
	DeleteCategory(ctx context.Context, categoryID string) error
	ListCategories(ctx context.Context, parentID string) ([]*domain.DocumentCategory, error)
	AssignDocumentToCategory(ctx context.Context, documentID, categoryID string) error
	RemoveDocumentFromCategory(ctx context.Context, documentID, categoryID string) error
	GetDocumentCategories(ctx context.Context, documentID string) ([]*domain.DocumentCategory, error)
}

// CategoryService implements the CategoryServiceInterface
type CategoryService struct {
	db *gorm.DB
}

// NewCategoryService creates a new instance of CategoryService
func NewCategoryService() *CategoryService {
	return &CategoryService{
		db: database.GetDB(),
	}
}

// CreateCategory creates a new document category
func (s *CategoryService) CreateCategory(ctx context.Context, name, description, parentID string) (*domain.DocumentCategory, error) {
	// Check if parent category exists (if parentID is provided)
	if parentID != "" {
		var parentCategory domain.DocumentCategory
		if err := s.db.Where("id = ?", parentID).First(&parentCategory).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, errors.New("parent category not found")
			}
			logger.Error("failed to find parent category", "error", err)
			return nil, errors.New("failed to create category")
		}
	}

	// Create new category
	category := &domain.DocumentCategory{
		ID:          generateID(),
		Name:        name,
		Description: description,
		ParentID:    parentID,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// Save category to database
	if err := s.db.Create(category).Error; err != nil {
		logger.Error("failed to create category", "error", err)
		return nil, errors.New("failed to create category")
	}

	return category, nil
}

// GetCategory retrieves a category by ID
func (s *CategoryService) GetCategory(ctx context.Context, categoryID string) (*domain.DocumentCategory, error) {
	var category domain.DocumentCategory
	if err := s.db.Where("id = ?", categoryID).First(&category).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("category not found")
		}
		logger.Error("failed to find category", "error", err)
		return nil, errors.New("failed to get category")
	}

	return &category, nil
}

// UpdateCategory updates a category
func (s *CategoryService) UpdateCategory(ctx context.Context, categoryID, name, description string) error {
	// Find category by ID
	var category domain.DocumentCategory
	if err := s.db.Where("id = ?", categoryID).First(&category).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("category not found")
		}
		logger.Error("failed to find category", "error", err)
		return errors.New("failed to update category")
	}

	// Update category fields
	if name != "" {
		category.Name = name
	}
	if description != "" {
		category.Description = description
	}
	category.UpdatedAt = time.Now()

	// Save updated category to database
	if err := s.db.Save(&category).Error; err != nil {
		logger.Error("failed to update category", "error", err)
		return errors.New("failed to update category")
	}

	return nil
}

// DeleteCategory deletes a category
func (s *CategoryService) DeleteCategory(ctx context.Context, categoryID string) error {
	// Find category by ID
	var category domain.DocumentCategory
	if err := s.db.Where("id = ?", categoryID).First(&category).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("category not found")
		}
		logger.Error("failed to find category", "error", err)
		return errors.New("failed to delete category")
	}

	// Check if category has child categories
	var childCount int64
	if err := s.db.Model(&domain.DocumentCategory{}).Where("parent_id = ?", categoryID).Count(&childCount).Error; err != nil {
		logger.Error("failed to count child categories", "error", err)
		return errors.New("failed to delete category")
	}

	if childCount > 0 {
		return errors.New("cannot delete category with child categories")
	}

	// Delete category from database
	if err := s.db.Delete(&category).Error; err != nil {
		logger.Error("failed to delete category", "error", err)
		return errors.New("failed to delete category")
	}

	// Delete all document-category relations for this category
	if err := s.db.Where("category_id = ?", categoryID).Delete(&domain.DocumentCategoryRelation{}).Error; err != nil {
		logger.Error("failed to delete document-category relations", "error", err)
		// Note: In a real application, you might want to handle this error more gracefully
		// For now, we'll just log it and continue
	}

	return nil
}

// ListCategories lists all categories with a specific parent ID (or all root categories if parentID is empty)
func (s *CategoryService) ListCategories(ctx context.Context, parentID string) ([]*domain.DocumentCategory, error) {
	var categories []*domain.DocumentCategory

	// Build query based on parentID
	query := s.db
	if parentID != "" {
		query = query.Where("parent_id = ?", parentID)
	} else {
		query = query.Where("parent_id = ?", "")
	}

	// Execute query
	if err := query.Find(&categories).Error; err != nil {
		logger.Error("failed to list categories", "error", err)
		return nil, errors.New("failed to list categories")
	}

	return categories, nil
}

// AssignDocumentToCategory assigns a document to a category
func (s *CategoryService) AssignDocumentToCategory(ctx context.Context, documentID, categoryID string) error {
	// Check if document exists
	var documentCount int64
	if err := s.db.Model(&domain.Document{}).Where("id = ?", documentID).Count(&documentCount).Error; err != nil {
		logger.Error("failed to count documents", "error", err)
		return errors.New("failed to assign document to category")
	}

	if documentCount == 0 {
		return errors.New("document not found")
	}

	// Check if category exists
	var categoryCount int64
	if err := s.db.Model(&domain.DocumentCategory{}).Where("id = ?", categoryID).Count(&categoryCount).Error; err != nil {
		logger.Error("failed to count categories", "error", err)
		return errors.New("failed to assign document to category")
	}

	if categoryCount == 0 {
		return errors.New("category not found")
	}

	// Check if relation already exists
	var relationCount int64
	if err := s.db.Model(&domain.DocumentCategoryRelation{}).
		Where("document_id = ? AND category_id = ?", documentID, categoryID).
		Count(&relationCount).Error; err != nil {
		logger.Error("failed to count document-category relations", "error", err)
		return errors.New("failed to assign document to category")
	}

	if relationCount > 0 {
		// Relation already exists, nothing to do
		return nil
	}

	// Create document-category relation
	relation := &domain.DocumentCategoryRelation{
		ID:         generateID(),
		DocumentID: documentID,
		CategoryID: categoryID,
		CreatedAt:  time.Now(),
	}

	if err := s.db.Create(relation).Error; err != nil {
		logger.Error("failed to create document-category relation", "error", err)
		return errors.New("failed to assign document to category")
	}

	return nil
}

// RemoveDocumentFromCategory removes a document from a category
func (s *CategoryService) RemoveDocumentFromCategory(ctx context.Context, documentID, categoryID string) error {
	// Delete document-category relation
	result := s.db.Where("document_id = ? AND category_id = ?", documentID, categoryID).
		Delete(&domain.DocumentCategoryRelation{})

	if result.Error != nil {
		logger.Error("failed to delete document-category relation", "error", result.Error)
		return errors.New("failed to remove document from category")
	}

	if result.RowsAffected == 0 {
		return errors.New("document-category relation not found")
	}

	return nil
}

// GetDocumentCategories retrieves all categories for a document
func (s *CategoryService) GetDocumentCategories(ctx context.Context, documentID string) ([]*domain.DocumentCategory, error) {
	// Check if document exists
	var documentCount int64
	if err := s.db.Model(&domain.Document{}).Where("id = ?", documentID).Count(&documentCount).Error; err != nil {
		logger.Error("failed to count documents", "error", err)
		return nil, errors.New("failed to get document categories")
	}

	if documentCount == 0 {
		return nil, errors.New("document not found")
	}

	// Get category IDs for the document
	var relations []domain.DocumentCategoryRelation
	if err := s.db.Where("document_id = ?", documentID).Find(&relations).Error; err != nil {
		logger.Error("failed to find document-category relations", "error", err)
		return nil, errors.New("failed to get document categories")
	}

	// If no categories, return empty slice
	if len(relations) == 0 {
		return []*domain.DocumentCategory{}, nil
	}

	// Extract category IDs
	categoryIDs := make([]string, len(relations))
	for i, relation := range relations {
		categoryIDs[i] = relation.CategoryID
	}

	// Get categories
	var categories []*domain.DocumentCategory
	if err := s.db.Where("id IN ?", categoryIDs).Find(&categories).Error; err != nil {
		logger.Error("failed to find categories", "error", err)
		return nil, errors.New("failed to get document categories")
	}

	return categories, nil
}

// generateID generates a unique ID (simplified implementation)
func generateID() string {
	// In a real application, use a proper ID generation library like uuid
	return "cat_" + time.Now().Format("20060102150405")
}