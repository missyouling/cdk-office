package test

import (
	"context"
	"testing"
	"time"

	"cdk-office/internal/document/service"
	"cdk-office/internal/shared/testutils"
	"github.com/stretchr/testify/assert"
)

func TestDocumentService_Upload(t *testing.T) {
	// Setup
	db := testutils.SetupTestDB()
	docService := service.NewDocumentServiceWithDB(db)
	ctx := context.Background()

	// Test cases
	tests := []struct {
		name          string
		request       *service.UploadRequest
		expectError   bool
		errorMessage  string
	}{
		{
			name: "Valid document upload",
			request: &service.UploadRequest{
				Title:       "Test Document",
				Description: "Test document description",
				FilePath:    "/path/to/document.pdf",
				FileSize:    1024,
				MimeType:    "application/pdf",
				OwnerID:     "user1",
				TeamID:      "team1",
				Tags:        "tag1,tag2",
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Execute
			doc, err := docService.Upload(ctx, tt.request)

			// Assert
			if tt.expectError {
				assert.Error(t, err)
				assert.Equal(t, tt.errorMessage, err.Error())
				assert.Nil(t, doc)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, doc)
				assert.NotEmpty(t, doc.ID)
				assert.Equal(t, tt.request.Title, doc.Title)
				assert.Equal(t, tt.request.Description, doc.Description)
				assert.Equal(t, tt.request.FilePath, doc.FilePath)
				assert.Equal(t, tt.request.FileSize, doc.FileSize)
				assert.Equal(t, tt.request.MimeType, doc.MimeType)
				assert.Equal(t, tt.request.OwnerID, doc.OwnerID)
				assert.Equal(t, tt.request.TeamID, doc.TeamID)
				assert.Equal(t, tt.request.Tags, doc.Tags)
				assert.Equal(t, "active", doc.Status)
				assert.WithinDuration(t, time.Now(), doc.CreatedAt, time.Second)
				assert.WithinDuration(t, time.Now(), doc.UpdatedAt, time.Second)

				// Check that a version was created
				versions, err := docService.GetDocumentVersions(ctx, doc.ID)
				assert.NoError(t, err)
				assert.Len(t, versions, 1)
				assert.Equal(t, doc.ID, versions[0].DocumentID)
				assert.Equal(t, 1, versions[0].Version)
				assert.Equal(t, tt.request.FilePath, versions[0].FilePath)
				assert.Equal(t, tt.request.FileSize, versions[0].FileSize)
			}
		})
	}
}

func TestDocumentService_GetDocument(t *testing.T) {
	// Setup
	db := testutils.SetupTestDB()
	docService := service.NewDocumentServiceWithDB(db)
	ctx := context.Background()

	// Create a document for testing
	uploadReq := &service.UploadRequest{
		Title:       "Test Document",
		Description: "Test document description",
		FilePath:    "/path/to/document.pdf",
		FileSize:    1024,
		MimeType:    "application/pdf",
		OwnerID:     "user1",
		TeamID:      "team1",
		Tags:        "tag1,tag2",
	}
	createdDoc, err := docService.Upload(ctx, uploadReq)
	assert.NoError(t, err)
	assert.NotNil(t, createdDoc)

	// Test cases
	tests := []struct {
		name         string
		docID        string
		expectError  bool
		errorMessage string
	}{
		{
			name:         "Get non-existent document",
			docID:        "non-existent-id",
			expectError:  true,
			errorMessage: "document not found",
		},
		{
			name:        "Get existing document",
			docID:       createdDoc.ID,
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Execute
			doc, err := docService.GetDocument(ctx, tt.docID)

			// Assert
			if tt.expectError {
				assert.Error(t, err)
				assert.Equal(t, tt.errorMessage, err.Error())
				assert.Nil(t, doc)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, doc)
				assert.Equal(t, createdDoc.ID, doc.ID)
				assert.Equal(t, createdDoc.Title, doc.Title)
				assert.Equal(t, createdDoc.Description, doc.Description)
				assert.Equal(t, createdDoc.FilePath, doc.FilePath)
				assert.Equal(t, createdDoc.FileSize, doc.FileSize)
				assert.Equal(t, createdDoc.MimeType, doc.MimeType)
				assert.Equal(t, createdDoc.OwnerID, doc.OwnerID)
				assert.Equal(t, createdDoc.TeamID, doc.TeamID)
				assert.Equal(t, createdDoc.Tags, doc.Tags)
				assert.Equal(t, createdDoc.Status, doc.Status)
			}
		})
	}
}

func TestDocumentService_UpdateDocument(t *testing.T) {
	// Setup
	db := testutils.SetupTestDB()
	docService := service.NewDocumentServiceWithDB(db)
	ctx := context.Background()

	// Create a document for testing
	uploadReq := &service.UploadRequest{
		Title:       "Test Document",
		Description: "Test document description",
		FilePath:    "/path/to/document.pdf",
		FileSize:    1024,
		MimeType:    "application/pdf",
		OwnerID:     "user1",
		TeamID:      "team1",
		Tags:        "tag1,tag2",
	}
	createdDoc, err := docService.Upload(ctx, uploadReq)
	assert.NoError(t, err)
	assert.NotNil(t, createdDoc)

	// Test cases
	tests := []struct {
		name          string
		docID         string
		request       *service.UpdateRequest
		expectError   bool
		errorMessage  string
	}{
		{
			name:  "Update non-existent document",
			docID: "non-existent-id",
			request: &service.UpdateRequest{
				Title: "Updated Title",
			},
			expectError:  true,
			errorMessage: "document not found",
		},
		{
			name:  "Valid document update",
			docID: createdDoc.ID,
			request: &service.UpdateRequest{
				Title:       "Updated Title",
				Description: "Updated description",
				Status:      "archived",
				Tags:        "updated,tag1,tag2",
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Execute
			err := docService.UpdateDocument(ctx, tt.docID, tt.request)

			// Assert
			if tt.expectError {
				assert.Error(t, err)
				assert.Equal(t, tt.errorMessage, err.Error())
			} else {
				assert.NoError(t, err)

				// Verify the update
				updatedDoc, getErr := docService.GetDocument(ctx, tt.docID)
				assert.NoError(t, getErr)
				assert.NotNil(t, updatedDoc)
				assert.Equal(t, tt.request.Title, updatedDoc.Title)
				assert.Equal(t, tt.request.Description, updatedDoc.Description)
				assert.Equal(t, tt.request.Status, updatedDoc.Status)
				assert.Equal(t, tt.request.Tags, updatedDoc.Tags)
			}
		})
	}
}

func TestDocumentService_DeleteDocument(t *testing.T) {
	// Setup
	db := testutils.SetupTestDB()
	docService := service.NewDocumentServiceWithDB(db)
	ctx := context.Background()

	// Create a document for testing
	uploadReq := &service.UploadRequest{
		Title:       "Test Document",
		Description: "Test document description",
		FilePath:    "/path/to/document.pdf",
		FileSize:    1024,
		MimeType:    "application/pdf",
		OwnerID:     "user1",
		TeamID:      "team1",
		Tags:        "tag1,tag2",
	}
	createdDoc, err := docService.Upload(ctx, uploadReq)
	assert.NoError(t, err)
	assert.NotNil(t, createdDoc)

	// Test cases
	tests := []struct {
		name         string
		docID        string
		expectError  bool
		errorMessage string
	}{
		{
			name:         "Delete non-existent document",
			docID:        "non-existent-id",
			expectError:  true,
			errorMessage: "document not found",
		},
		{
			name:        "Valid document deletion",
			docID:       createdDoc.ID,
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Execute
			err := docService.DeleteDocument(ctx, tt.docID)

			// Assert
			if tt.expectError {
				assert.Error(t, err)
				assert.Equal(t, tt.errorMessage, err.Error())
			} else {
				assert.NoError(t, err)

				// Verify the deletion
				_, getErr := docService.GetDocument(ctx, tt.docID)
				assert.Error(t, getErr)
				assert.Equal(t, "document not found", getErr.Error())

				// Verify versions are also deleted
				versions, getVersionsErr := docService.GetDocumentVersions(ctx, tt.docID)
				assert.Error(t, getVersionsErr)
				assert.Equal(t, "document not found", getVersionsErr.Error())
				assert.Nil(t, versions)
			}
		})
	}
}

func TestDocumentService_GetDocumentVersions(t *testing.T) {
	// Setup
	db := testutils.SetupTestDB()
	docService := service.NewDocumentServiceWithDB(db)
	ctx := context.Background()

	// Create a document for testing
	uploadReq := &service.UploadRequest{
		Title:       "Test Document",
		Description: "Test document description",
		FilePath:    "/path/to/document.pdf",
		FileSize:    1024,
		MimeType:    "application/pdf",
		OwnerID:     "user1",
		TeamID:      "team1",
		Tags:        "tag1,tag2",
	}
	createdDoc, err := docService.Upload(ctx, uploadReq)
	assert.NoError(t, err)
	assert.NotNil(t, createdDoc)

	// Test cases
	tests := []struct {
		name         string
		docID        string
		expectError  bool
		errorMessage string
	}{
		{
			name:         "Get versions for non-existent document",
			docID:        "non-existent-id",
			expectError:  true,
			errorMessage: "document not found",
		},
		{
			name:        "Get versions for existing document",
			docID:       createdDoc.ID,
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Execute
			versions, err := docService.GetDocumentVersions(ctx, tt.docID)

			// Assert
			if tt.expectError {
				assert.Error(t, err)
				assert.Equal(t, tt.errorMessage, err.Error())
				assert.Nil(t, versions)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, versions)
				assert.Len(t, versions, 1)
				assert.Equal(t, createdDoc.ID, versions[0].DocumentID)
				assert.Equal(t, 1, versions[0].Version)
				assert.Equal(t, uploadReq.FilePath, versions[0].FilePath)
				assert.Equal(t, uploadReq.FileSize, versions[0].FileSize)
			}
		})
	}
}