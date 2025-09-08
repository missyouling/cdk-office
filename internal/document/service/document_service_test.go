package service

import (
	"context"
	"testing"
	
	"github.com/stretchr/testify/assert"
	"cdk-office/internal/shared/testutils"
)

// TestDocumentService tests the DocumentService
func TestDocumentService(t *testing.T) {
	// Set up test environment
	// logger.InitTestLogger()
	
	// Initialize the database connection for testing
	testDB := testutils.SetupTestDB()
	
	// Create document service with database connection
	documentService := &DocumentService{
		db: testDB,
	}
	
	// Test Upload
	t.Run("Upload", func(t *testing.T) {
		// Prepare test data
		ctx := context.Background()
		req := &UploadRequest{
			Title:    "Test Document",
			FilePath: "/path/to/test/document.txt",
			FileSize: 1024,
			MimeType: "text/plain",
			OwnerID:  "user_123",
			TeamID:   "team_123",
		}
		
		// Call the method under test
		doc, err := documentService.Upload(ctx, req)
		
		// Assert results
		assert.NoError(t, err)
		assert.NotNil(t, doc)
	})
	
	// Test UpdateDocument
	t.Run("UpdateDocument", func(t *testing.T) {
		// First create a document to update
		ctx := context.Background()
		createReq := &UploadRequest{
			Title:    "Test Document 2",
			FilePath: "/path/to/test/document2.txt",
			FileSize: 2048,
			MimeType: "text/plain",
			OwnerID:  "user_123",
			TeamID:   "team_123",
		}
		
		doc, err := documentService.Upload(ctx, createReq)
		assert.NoError(t, err)
		assert.NotNil(t, doc)
		
		// Now update the document
		updateReq := &UpdateRequest{
			Title: "Updated Document",
		}
		
		err = documentService.UpdateDocument(ctx, doc.ID, updateReq)
		assert.NoError(t, err)
	})
	
	// Test DeleteDocument
	t.Run("DeleteDocument", func(t *testing.T) {
		// First create a document to delete
		ctx := context.Background()
		createReq := &UploadRequest{
			Title:    "Test Document 3",
			FilePath: "/path/to/test/document3.txt",
			FileSize: 3072,
			MimeType: "text/plain",
			OwnerID:  "user_123",
			TeamID:   "team_123",
		}
		
		doc, err := documentService.Upload(ctx, createReq)
		assert.NoError(t, err)
		assert.NotNil(t, doc)
		
		// Now delete the document
		err = documentService.DeleteDocument(ctx, doc.ID)
		assert.NoError(t, err)
	})
}