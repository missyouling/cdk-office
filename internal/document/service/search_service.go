package service

import (
	"context"
	"errors"
	"strings"

	"cdk-office/internal/document/domain"
	"cdk-office/internal/shared/database"
	"cdk-office/pkg/logger"
	"gorm.io/gorm"
)

// SearchServiceInterface defines the interface for document search service
type SearchServiceInterface interface {
	SearchDocuments(ctx context.Context, query string, teamID string, page, size int) ([]*domain.Document, int64, error)
}

// SearchService implements the SearchServiceInterface
type SearchService struct {
	db *gorm.DB
}

// NewSearchService creates a new instance of SearchService
func NewSearchService() *SearchService {
	return &SearchService{
		db: database.GetDB(),
	}
}

// SearchDocuments searches for documents based on a query
func (s *SearchService) SearchDocuments(ctx context.Context, query string, teamID string, page, size int) ([]*domain.Document, int64, error) {
	// Validate pagination parameters
	if page < 1 {
		page = 1
	}
	if size < 1 || size > 100 {
		size = 10
	}

	// Build the query
	dbQuery := s.db.Model(&domain.Document{})

	// Add team filter if provided
	if teamID != "" {
		dbQuery = dbQuery.Where("team_id = ?", teamID)
	}

	// Add search conditions if query is provided
	if query != "" {
		// Split query into terms
		terms := strings.Fields(query)
		
		// Build search conditions
		for _, term := range terms {
			// Escape special characters in term
			escapedTerm := strings.ReplaceAll(term, "%", "\\%")
			escapedTerm = strings.ReplaceAll(escapedTerm, "_", "\\_")
			
			// Add search condition for title, description, and tags
			dbQuery = dbQuery.Where(
				"title ILIKE ? OR description ILIKE ? OR tags ILIKE ?",
				"%"+escapedTerm+"%",
				"%"+escapedTerm+"%",
				"%"+escapedTerm+"%",
			)
		}
	}

	// Count total results
	var total int64
	if err := dbQuery.Count(&total).Error; err != nil {
		logger.Error("failed to count search results", "error", err)
		return nil, 0, errors.New("failed to search documents")
	}

	// Apply pagination
	offset := (page - 1) * size
	dbQuery = dbQuery.Offset(offset).Limit(size)

	// Execute query
	var documents []*domain.Document
	if err := dbQuery.Find(&documents).Error; err != nil {
		logger.Error("failed to search documents", "error", err)
		return nil, 0, errors.New("failed to search documents")
	}

	return documents, total, nil
}