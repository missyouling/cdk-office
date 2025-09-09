package handler

import (
	"net/http"

	"cdk-office/internal/document/service"
	"github.com/gin-gonic/gin"
)

// DocumentHandlerInterface defines the interface for document handler
type DocumentHandlerInterface interface {
	Upload(c *gin.Context)
	GetDocument(c *gin.Context)
	UpdateDocument(c *gin.Context)
	DeleteDocument(c *gin.Context)
	GetDocumentVersions(c *gin.Context)
}

// DocumentHandler implements the DocumentHandlerInterface
type DocumentHandler struct {
	documentService service.DocumentServiceInterface
}

// NewDocumentHandler creates a new instance of DocumentHandler
func NewDocumentHandler() *DocumentHandler {
	return &DocumentHandler{
		documentService: service.NewDocumentService(),
	}
}

// NewDocumentHandlerWithService creates a new instance of DocumentHandler with a specific service
func NewDocumentHandlerWithService(documentService service.DocumentServiceInterface) *DocumentHandler {
	return &DocumentHandler{
		documentService: documentService,
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

// Upload handles uploading a new document
func (h *DocumentHandler) Upload(c *gin.Context) {
	var req UploadRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Call service to upload document
	document, err := h.documentService.Upload(c.Request.Context(), &service.UploadRequest{
		Title:       req.Title,
		Description: req.Description,
		FilePath:    req.FilePath,
		FileSize:    req.FileSize,
		MimeType:    req.MimeType,
		OwnerID:     req.OwnerID,
		TeamID:      req.TeamID,
		Tags:        req.Tags,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, document)
}

// GetDocument handles retrieving a document by ID
func (h *DocumentHandler) GetDocument(c *gin.Context) {
	docID := c.Param("id")
	if docID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "document id is required"})
		return
	}

	// Call service to get document
	document, err := h.documentService.GetDocument(c.Request.Context(), docID)
	if err != nil {
		if err.Error() == "document not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "document not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, document)
}

// UpdateDocument handles updating a document
func (h *DocumentHandler) UpdateDocument(c *gin.Context) {
	docID := c.Param("id")
	if docID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "document id is required"})
		return
	}

	var req UpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Call service to update document
	if err := h.documentService.UpdateDocument(c.Request.Context(), docID, &service.UpdateRequest{
		Title:       req.Title,
		Description: req.Description,
		Status:      req.Status,
		Tags:        req.Tags,
	}); err != nil {
		if err.Error() == "document not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "document not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "document updated successfully"})
}

// DeleteDocument handles deleting a document
func (h *DocumentHandler) DeleteDocument(c *gin.Context) {
	docID := c.Param("id")
	if docID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "document id is required"})
		return
	}

	// Call service to delete document
	if err := h.documentService.DeleteDocument(c.Request.Context(), docID); err != nil {
		if err.Error() == "document not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "document not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "document deleted successfully"})
}

// GetDocumentVersions handles retrieving all versions of a document
func (h *DocumentHandler) GetDocumentVersions(c *gin.Context) {
	docID := c.Param("id")
	if docID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "document id is required"})
		return
	}

	// Call service to get document versions
	versions, err := h.documentService.GetDocumentVersions(c.Request.Context(), docID)
	if err != nil {
		if err.Error() == "document not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "document not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, versions)
}