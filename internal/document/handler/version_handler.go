package handler

import (
	"net/http"

	"cdk-office/internal/document/service"
	"github.com/gin-gonic/gin"
)

// VersionHandlerInterface defines the interface for document version handler
type VersionHandlerInterface interface {
	CreateVersion(c *gin.Context)
	GetVersion(c *gin.Context)
	ListVersions(c *gin.Context)
	GetLatestVersion(c *gin.Context)
	RestoreVersion(c *gin.Context)
}

// VersionHandler implements the VersionHandlerInterface
type VersionHandler struct {
	versionService service.VersionServiceInterface
}

// NewVersionHandler creates a new instance of VersionHandler
func NewVersionHandler() *VersionHandler {
	return &VersionHandler{
		versionService: service.NewVersionService(),
	}
}

// CreateVersionRequest represents the request for creating a document version
type CreateVersionRequest struct {
	DocumentID string `json:"document_id" binding:"required"`
	FilePath   string `json:"file_path" binding:"required"`
	FileSize   int64  `json:"file_size" binding:"required"`
}

// RestoreVersionRequest represents the request for restoring a document version
type RestoreVersionRequest struct {
	VersionID string `json:"version_id" binding:"required"`
}

// CreateVersion handles creating a new version of a document
func (h *VersionHandler) CreateVersion(c *gin.Context) {
	var req CreateVersionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Call service to create version
	version, err := h.versionService.CreateVersion(c.Request.Context(), req.DocumentID, req.FilePath, req.FileSize)
	if err != nil {
		if err.Error() == "document not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "document not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, version)
}

// GetVersion handles retrieving a specific version of a document
func (h *VersionHandler) GetVersion(c *gin.Context) {
	versionID := c.Param("id")
	if versionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "version id is required"})
		return
	}

	// Call service to get version
	version, err := h.versionService.GetVersion(c.Request.Context(), versionID)
	if err != nil {
		if err.Error() == "version not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "version not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, version)
}

// ListVersions handles listing all versions of a document
func (h *VersionHandler) ListVersions(c *gin.Context) {
	documentID := c.Param("document_id")
	if documentID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "document id is required"})
		return
	}

	// Call service to list versions
	versions, err := h.versionService.ListVersions(c.Request.Context(), documentID)
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

// GetLatestVersion handles retrieving the latest version of a document
func (h *VersionHandler) GetLatestVersion(c *gin.Context) {
	documentID := c.Param("document_id")
	if documentID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "document id is required"})
		return
	}

	// Call service to get latest version
	version, err := h.versionService.GetLatestVersion(c.Request.Context(), documentID)
	if err != nil {
		if err.Error() == "document not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "document not found"})
			return
		}
		if err.Error() == "no versions found for document" {
			c.JSON(http.StatusNotFound, gin.H{"error": "no versions found for document"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, version)
}

// RestoreVersion handles restoring a document to a specific version
func (h *VersionHandler) RestoreVersion(c *gin.Context) {
	var req RestoreVersionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Call service to restore version
	if err := h.versionService.RestoreVersion(c.Request.Context(), req.VersionID); err != nil {
		if err.Error() == "version not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "version not found"})
			return
		}
		if err.Error() == "document not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "document not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "version restored successfully"})
}