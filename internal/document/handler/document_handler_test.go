package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"cdk-office/internal/document/domain"
	"cdk-office/internal/document/service"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// mockDocumentService is a mock implementation of the DocumentServiceInterface
type mockDocumentService struct {
	documents map[string]*domain.Document
	versions  map[string][]*domain.DocumentVersion
	nextID    int
}

func newMockDocumentService() *mockDocumentService {
	return &mockDocumentService{
		documents: make(map[string]*domain.Document),
		versions:  make(map[string][]*domain.DocumentVersion),
		nextID:    1,
	}
}

func (m *mockDocumentService) Upload(ctx context.Context, req *service.UploadRequest) (*domain.Document, error) {
	// Create new document
	doc := &domain.Document{
		ID:          "doc_" + string(rune(m.nextID+'0')),
		Title:       req.Title,
		Description: req.Description,
		FilePath:    req.FilePath,
		FileSize:    req.FileSize,
		MimeType:    req.MimeType,
		OwnerID:     req.OwnerID,
		TeamID:      req.TeamID,
		Tags:        req.Tags,
		Status:      "active",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	m.documents[doc.ID] = doc
	m.nextID++
	return doc, nil
}

func (m *mockDocumentService) GetDocument(ctx context.Context, docID string) (*domain.Document, error) {
	doc, exists := m.documents[docID]
	if !exists {
		return nil, errors.New("document not found")
	}
	return doc, nil
}

func (m *mockDocumentService) UpdateDocument(ctx context.Context, docID string, req *service.UpdateRequest) error {
	doc, exists := m.documents[docID]
	if !exists {
		return errors.New("document not found")
	}

	if req.Title != "" {
		doc.Title = req.Title
	}
	if req.Description != "" {
		doc.Description = req.Description
	}
	if req.Status != "" {
		doc.Status = req.Status
	}
	if req.Tags != "" {
		doc.Tags = req.Tags
	}

	doc.UpdatedAt = time.Now()
	return nil
}

func (m *mockDocumentService) DeleteDocument(ctx context.Context, docID string) error {
	_, exists := m.documents[docID]
	if !exists {
		return errors.New("document not found")
	}

	delete(m.documents, docID)
	return nil
}

func (m *mockDocumentService) GetDocumentVersions(ctx context.Context, docID string) ([]*domain.DocumentVersion, error) {
	_, exists := m.documents[docID]
	if !exists {
		return nil, errors.New("document not found")
	}

	versions, exists := m.versions[docID]
	if !exists {
		return []*domain.DocumentVersion{}, nil
	}

	return versions, nil
}

// TestDocumentHandler tests the DocumentHandler
func TestDocumentHandler(t *testing.T) {
	// Set up test environment
	gin.SetMode(gin.TestMode)

	// Create mock service
	mockService := newMockDocumentService()

	// Create document handler with mock service
	docHandler := NewDocumentHandlerWithService(mockService)

	// Test Upload
	t.Run("Upload", func(t *testing.T) {
		// Create test request
		reqBody := UploadRequest{
			Title:    "Test Document",
			FilePath: "/path/to/test.pdf",
			FileSize: 1024,
			MimeType: "application/pdf",
			OwnerID:  "user_123",
			TeamID:   "team_123",
		}
		jsonReq, _ := json.Marshal(reqBody)

		// Create HTTP request and response recorder
		req, _ := http.NewRequest("POST", "/documents", bytes.NewBuffer(jsonReq))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		// Create gin context and call handler
		c, _ := gin.CreateTestContext(w)
		c.Request = req
		docHandler.Upload(c)

		// Assert response
		assert.Equal(t, http.StatusOK, w.Code)
	})

	// Test Upload with invalid JSON
	t.Run("UploadInvalidJSON", func(t *testing.T) {
		// Create HTTP request with invalid JSON
		req, _ := http.NewRequest("POST", "/documents", bytes.NewBuffer([]byte("{invalid json}")))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		// Create gin context and call handler
		c, _ := gin.CreateTestContext(w)
		c.Request = req
		docHandler.Upload(c)

		// Assert response
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	// Test GetDocument
	t.Run("GetDocument", func(t *testing.T) {
		// First create a document
		uploadReq := &service.UploadRequest{
			Title:    "Get Test Document",
			FilePath: "/path/to/test.pdf",
			FileSize: 1024,
			MimeType: "application/pdf",
			OwnerID:  "user_123",
			TeamID:   "team_123",
		}
		doc, _ := mockService.Upload(context.Background(), uploadReq)

		// Create HTTP request and response recorder
		req, _ := http.NewRequest("GET", "/documents/"+doc.ID, nil)
		w := httptest.NewRecorder()

		// Create gin context and call handler
		c, _ := gin.CreateTestContext(w)
		c.Request = req
		c.AddParam("id", doc.ID)
		docHandler.GetDocument(c)

		// Assert response
		assert.Equal(t, http.StatusOK, w.Code)
	})

	// Test GetDocument with non-existent ID
	t.Run("GetDocumentNotFound", func(t *testing.T) {
		// Create HTTP request and response recorder
		req, _ := http.NewRequest("GET", "/documents/non-existent", nil)
		w := httptest.NewRecorder()

		// Create gin context and call handler
		c, _ := gin.CreateTestContext(w)
		c.Request = req
		c.AddParam("id", "non-existent")
		docHandler.GetDocument(c)

		// Assert response
		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	// Test UpdateDocument
	t.Run("UpdateDocument", func(t *testing.T) {
		// First create a document
		uploadReq := &service.UploadRequest{
			Title:    "Update Test Document",
			FilePath: "/path/to/test.pdf",
			FileSize: 1024,
			MimeType: "application/pdf",
			OwnerID:  "user_123",
			TeamID:   "team_123",
		}
		doc, _ := mockService.Upload(context.Background(), uploadReq)

		// Create update request
		updateReqBody := UpdateRequest{
			Title: "Updated Document",
		}
		jsonReq, _ := json.Marshal(updateReqBody)

		// Create HTTP request and response recorder
		req, _ := http.NewRequest("PUT", "/documents/"+doc.ID, bytes.NewBuffer(jsonReq))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		// Create gin context and call handler
		c, _ := gin.CreateTestContext(w)
		c.Request = req
		c.AddParam("id", doc.ID)
		docHandler.UpdateDocument(c)

		// Assert response
		assert.Equal(t, http.StatusOK, w.Code)
	})

	// Test UpdateDocument with non-existent ID
	t.Run("UpdateDocumentNotFound", func(t *testing.T) {
		// Create update request
		updateReqBody := UpdateRequest{
			Title: "Updated Document",
		}
		jsonReq, _ := json.Marshal(updateReqBody)

		// Create HTTP request and response recorder
		req, _ := http.NewRequest("PUT", "/documents/non-existent", bytes.NewBuffer(jsonReq))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		// Create gin context and call handler
		c, _ := gin.CreateTestContext(w)
		c.Request = req
		c.AddParam("id", "non-existent")
		docHandler.UpdateDocument(c)

		// Assert response
		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	// Test UpdateDocument with invalid JSON
	t.Run("UpdateDocumentInvalidJSON", func(t *testing.T) {
		// Create HTTP request with invalid JSON
		req, _ := http.NewRequest("PUT", "/documents/doc_1", bytes.NewBuffer([]byte("{invalid json}")))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		// Create gin context and call handler
		c, _ := gin.CreateTestContext(w)
		c.Request = req
		c.AddParam("id", "doc_1")
		docHandler.UpdateDocument(c)

		// Assert response
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	// Test DeleteDocument
	t.Run("DeleteDocument", func(t *testing.T) {
		// First create a document
		uploadReq := &service.UploadRequest{
			Title:    "Delete Test Document",
			FilePath: "/path/to/test.pdf",
			FileSize: 1024,
			MimeType: "application/pdf",
			OwnerID:  "user_123",
			TeamID:   "team_123",
		}
		doc, _ := mockService.Upload(context.Background(), uploadReq)

		// Create HTTP request and response recorder
		req, _ := http.NewRequest("DELETE", "/documents/"+doc.ID, nil)
		w := httptest.NewRecorder()

		// Create gin context and call handler
		c, _ := gin.CreateTestContext(w)
		c.Request = req
		c.AddParam("id", doc.ID)
		docHandler.DeleteDocument(c)

		// Assert response
		assert.Equal(t, http.StatusOK, w.Code)
	})

	// Test DeleteDocument with non-existent ID
	t.Run("DeleteDocumentNotFound", func(t *testing.T) {
		// Create HTTP request and response recorder
		req, _ := http.NewRequest("DELETE", "/documents/non-existent", nil)
		w := httptest.NewRecorder()

		// Create gin context and call handler
		c, _ := gin.CreateTestContext(w)
		c.Request = req
		c.AddParam("id", "non-existent")
		docHandler.DeleteDocument(c)

		// Assert response
		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	// Test GetDocumentVersions
	t.Run("GetDocumentVersions", func(t *testing.T) {
		// First create a document
		uploadReq := &service.UploadRequest{
			Title:    "Versions Test Document",
			FilePath: "/path/to/test.pdf",
			FileSize: 1024,
			MimeType: "application/pdf",
			OwnerID:  "user_123",
			TeamID:   "team_123",
		}
		doc, _ := mockService.Upload(context.Background(), uploadReq)

		// Create HTTP request and response recorder
		req, _ := http.NewRequest("GET", "/documents/"+doc.ID+"/versions", nil)
		w := httptest.NewRecorder()

		// Create gin context and call handler
		c, _ := gin.CreateTestContext(w)
		c.Request = req
		c.AddParam("id", doc.ID)
		docHandler.GetDocumentVersions(c)

		// Assert response
		assert.Equal(t, http.StatusOK, w.Code)
	})

	// Test GetDocumentVersions with non-existent ID
	t.Run("GetDocumentVersionsNotFound", func(t *testing.T) {
		// Create HTTP request and response recorder
		req, _ := http.NewRequest("GET", "/documents/non-existent/versions", nil)
		w := httptest.NewRecorder()

		// Create gin context and call handler
		c, _ := gin.CreateTestContext(w)
		c.Request = req
		c.AddParam("id", "non-existent")
		docHandler.GetDocumentVersions(c)

		// Assert response
		assert.Equal(t, http.StatusNotFound, w.Code)
	})
}