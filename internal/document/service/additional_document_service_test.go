package service

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"cdk-office/internal/document/domain"
	"cdk-office/internal/shared/testutils"
)

// TestDocumentServiceAdditional tests additional scenarios for the DocumentService
func TestDocumentServiceAdditional(t *testing.T) {
	// Set up test environment
	testDB := testutils.SetupTestDB()

	// Create document service with database connection
	documentService := &DocumentService{
		db: testDB,
	}

	// Test Upload with all fields
	t.Run("UploadWithAllFields", func(t *testing.T) {
		ctx := context.Background()
		req := &UploadRequest{
			Title:       "Complete Test Document",
			Description: "A test document with all fields filled",
			FilePath:    "/path/to/complete/document.pdf",
			FileSize:    2048,
			MimeType:    "application/pdf",
			OwnerID:     "user_456",
			TeamID:      "team_456",
			Tags:        "test,document,complete",
		}

		doc, err := documentService.Upload(ctx, req)

		assert.NoError(t, err)
		assert.NotNil(t, doc)
		assert.Equal(t, "Complete Test Document", doc.Title)
		assert.Equal(t, "A test document with all fields filled", doc.Description)
		assert.Equal(t, "/path/to/complete/document.pdf", doc.FilePath)
		assert.Equal(t, int64(2048), doc.FileSize)
		assert.Equal(t, "application/pdf", doc.MimeType)
		assert.Equal(t, "user_456", doc.OwnerID)
		assert.Equal(t, "team_456", doc.TeamID)
		assert.Equal(t, "test,document,complete", doc.Tags)
		assert.Equal(t, "active", doc.Status)
	})

	// Test GetDocument with non-existent ID
	t.Run("GetDocumentNotFound", func(t *testing.T) {
		ctx := context.Background()

		doc, err := documentService.GetDocument(ctx, "non-existent-id")

		assert.Error(t, err)
		assert.Nil(t, doc)
		assert.Equal(t, "document not found", err.Error())
	})

	// Test UpdateDocument with non-existent ID
	t.Run("UpdateDocumentNotFound", func(t *testing.T) {
		ctx := context.Background()
		req := &UpdateRequest{
			Title: "Updated Document",
		}

		err := documentService.UpdateDocument(ctx, "non-existent-id", req)

		assert.Error(t, err)
		assert.Equal(t, "document not found", err.Error())
	})

	// Test DeleteDocument with non-existent ID
	t.Run("DeleteDocumentNotFound", func(t *testing.T) {
		ctx := context.Background()

		err := documentService.DeleteDocument(ctx, "non-existent-id")

		assert.Error(t, err)
		assert.Equal(t, "document not found", err.Error())
	})

	// Test GetDocumentVersions with non-existent document
	t.Run("GetDocumentVersionsNotFound", func(t *testing.T) {
		ctx := context.Background()

		versions, err := documentService.GetDocumentVersions(ctx, "non-existent-id")

		assert.Error(t, err)
		assert.Nil(t, versions)
		assert.Equal(t, "document not found", err.Error())
	})

	// Test UpdateDocument with all fields
	t.Run("UpdateDocumentAllFields", func(t *testing.T) {
		ctx := context.Background()

		// Create a document
		createReq := &UploadRequest{
			Title:    "Update All Fields Document",
			FilePath: "/path/to/update/document.txt",
			FileSize: 1024,
			MimeType: "text/plain",
			OwnerID:  "user_789",
			TeamID:   "team_789",
		}

		doc, err := documentService.Upload(ctx, createReq)
		assert.NoError(t, err)
		assert.NotNil(t, doc)

		// Update all fields
		updateReq := &UpdateRequest{
			Title:       "Fully Updated Document",
			Description: "Updated description",
			Status:      "archived",
			Tags:        "updated,test",
		}

		err = documentService.UpdateDocument(ctx, doc.ID, updateReq)
		assert.NoError(t, err)

		// Verify the update
		updatedDoc, err := documentService.GetDocument(ctx, doc.ID)
		assert.NoError(t, err)
		assert.NotNil(t, updatedDoc)
		assert.Equal(t, "Fully Updated Document", updatedDoc.Title)
		assert.Equal(t, "Updated description", updatedDoc.Description)
		assert.Equal(t, "archived", updatedDoc.Status)
		assert.Equal(t, "updated,test", updatedDoc.Tags)
	})

	// Test GetDocumentVersions
	t.Run("GetDocumentVersions", func(t *testing.T) {
		ctx := context.Background()

		// Create a document
		createReq := &UploadRequest{
			Title:    "Versions Test Document",
			FilePath: "/path/to/versions/document.txt",
			FileSize: 1024,
			MimeType: "text/plain",
			OwnerID:  "user_101",
			TeamID:   "team_101",
		}

		doc, err := documentService.Upload(ctx, createReq)
		assert.NoError(t, err)
		assert.NotNil(t, doc)

		// Get document versions
		versions, err := documentService.GetDocumentVersions(ctx, doc.ID)
		assert.NoError(t, err)
		assert.NotNil(t, versions)
		assert.Len(t, versions, 1)
		assert.Equal(t, doc.ID, versions[0].DocumentID)
		assert.Equal(t, 1, versions[0].Version)
	})

	// Test multiple document operations
	t.Run("MultipleDocumentOperations", func(t *testing.T) {
		ctx := context.Background()

		// Create multiple documents
		documentsData := []struct {
			title    string
			filePath string
			fileSize int64
			mimeType string
			ownerID  string
			teamID   string
		}{
			{"Multi Test 1", "/path/to/multi1.txt", 1024, "text/plain", "user_201", "team_201"},
			{"Multi Test 2", "/path/to/multi2.pdf", 2048, "application/pdf", "user_202", "team_202"},
			{"Multi Test 3", "/path/to/multi3.doc", 4096, "application/msword", "user_203", "team_203"},
		}

		var createdDocuments []*domain.Document
		for _, data := range documentsData {
			req := &UploadRequest{
				Title:    data.title,
				FilePath: data.filePath,
				FileSize: data.fileSize,
				MimeType: data.mimeType,
				OwnerID:  data.ownerID,
				TeamID:   data.teamID,
			}

			doc, err := documentService.Upload(ctx, req)
			assert.NoError(t, err)
			assert.NotNil(t, doc)
			createdDocuments = append(createdDocuments, doc)
		}

		// Update all documents
		for _, doc := range createdDocuments {
			updateReq := &UpdateRequest{
				Description: "Updated description for " + doc.Title,
				Status:      "archived",
			}

			err := documentService.UpdateDocument(ctx, doc.ID, updateReq)
			assert.NoError(t, err)
		}

		// Verify updates
		for _, doc := range createdDocuments {
			updatedDoc, err := documentService.GetDocument(ctx, doc.ID)
			assert.NoError(t, err)
			assert.NotNil(t, updatedDoc)
			assert.Contains(t, updatedDoc.Description, "Updated description")
			assert.Equal(t, "archived", updatedDoc.Status)
		}

		// Get versions for all documents
		for _, doc := range createdDocuments {
			versions, err := documentService.GetDocumentVersions(ctx, doc.ID)
			assert.NoError(t, err)
			assert.NotNil(t, versions)
			assert.Len(t, versions, 1)
		}

		// Delete all documents
		for _, doc := range createdDocuments {
			err := documentService.DeleteDocument(ctx, doc.ID)
			assert.NoError(t, err)
		}

		// Verify deletions
		for _, doc := range createdDocuments {
			_, err := documentService.GetDocument(ctx, doc.ID)
			assert.Error(t, err)
			assert.Equal(t, "document not found", err.Error())

			versions, err := documentService.GetDocumentVersions(ctx, doc.ID)
			assert.Error(t, err)
			assert.Nil(t, versions)
			assert.Equal(t, "document not found", err.Error())
		}
	})
}