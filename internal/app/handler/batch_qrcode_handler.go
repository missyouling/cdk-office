package handler

import (
	"net/http"
	"strconv"

	"cdk-office/internal/app/service"
	"github.com/gin-gonic/gin"
)

// BatchQRCodeHandlerInterface defines the interface for batch QR code handler
type BatchQRCodeHandlerInterface interface {
	CreateBatchQRCode(c *gin.Context)
	UpdateBatchQRCode(c *gin.Context)
	DeleteBatchQRCode(c *gin.Context)
	ListBatchQRCodes(c *gin.Context)
	GetBatchQRCode(c *gin.Context)
	GenerateBatchQRCodes(c *gin.Context)
}

// BatchQRCodeHandler implements the BatchQRCodeHandlerInterface
type BatchQRCodeHandler struct {
	batchService service.BatchQRCodeServiceInterface
}

// NewBatchQRCodeHandler creates a new instance of BatchQRCodeHandler
func NewBatchQRCodeHandler() *BatchQRCodeHandler {
	return &BatchQRCodeHandler{
		batchService: service.NewBatchQRCodeService(),
	}
}

// CreateBatchQRCodeRequest represents the request for creating a batch QR code
type CreateBatchQRCodeRequest struct {
	AppID       string            `json:"app_id" binding:"required"`
	Name        string            `json:"name" binding:"required"`
	Description string            `json:"description"`
	Prefix      string            `json:"prefix"`
	Count       int               `json:"count" binding:"required"`
	Type        string            `json:"type" binding:"required"` // static or dynamic
	URLTemplate string            `json:"url_template"`
	Config      map[string]string `json:"config"`
	CreatedBy   string            `json:"created_by" binding:"required"`
}

// UpdateBatchQRCodeRequest represents the request for updating a batch QR code
type UpdateBatchQRCodeRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Prefix      string `json:"prefix"`
	URLTemplate string `json:"url_template"`
	Config      string `json:"config"`
}

// ListBatchQRCodesRequest represents the request for listing batch QR codes
type ListBatchQRCodesRequest struct {
	AppID string `form:"app_id" binding:"required"`
	Page  int    `form:"page"`
	Size  int    `form:"size"`
}

// CreateBatchQRCode handles creating a new batch QR code
func (h *BatchQRCodeHandler) CreateBatchQRCode(c *gin.Context) {
	var req CreateBatchQRCodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Call service to create batch QR code
	batch, err := h.batchService.CreateBatchQRCode(c.Request.Context(), &service.CreateBatchQRCodeRequest{
		AppID:       req.AppID,
		Name:        req.Name,
		Description: req.Description,
		Prefix:      req.Prefix,
		Count:       req.Count,
		Type:        req.Type,
		URLTemplate: req.URLTemplate,
		Config:      req.Config,
		CreatedBy:   req.CreatedBy,
	})
	if err != nil {
		if err.Error() == "invalid QR code type" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid QR code type"})
			return
		}
		if err.Error() == "invalid count, must be between 1 and 10000" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid count, must be between 1 and 10000"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, batch)
}

// UpdateBatchQRCode handles updating an existing batch QR code
func (h *BatchQRCodeHandler) UpdateBatchQRCode(c *gin.Context) {
	batchID := c.Param("id")
	if batchID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "batch id is required"})
		return
	}

	var req UpdateBatchQRCodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Call service to update batch QR code
	if err := h.batchService.UpdateBatchQRCode(c.Request.Context(), batchID, &service.UpdateBatchQRCodeRequest{
		Name:        req.Name,
		Description: req.Description,
		Prefix:      req.Prefix,
		URLTemplate: req.URLTemplate,
		Config:      req.Config,
	}); err != nil {
		if err.Error() == "batch QR code not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "batch QR code not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "batch QR code updated successfully"})
}

// DeleteBatchQRCode handles deleting a batch QR code
func (h *BatchQRCodeHandler) DeleteBatchQRCode(c *gin.Context) {
	batchID := c.Param("id")
	if batchID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "batch id is required"})
		return
	}

	// Call service to delete batch QR code
	if err := h.batchService.DeleteBatchQRCode(c.Request.Context(), batchID); err != nil {
		if err.Error() == "batch QR code not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "batch QR code not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "batch QR code deleted successfully"})
}

// ListBatchQRCodes handles listing batch QR codes with pagination
func (h *BatchQRCodeHandler) ListBatchQRCodes(c *gin.Context) {
	// Parse query parameters
	appID := c.Query("app_id")
	if appID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "app id is required"})
		return
	}

	page := 1
	size := 10

	if pageStr := c.Query("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}

	if sizeStr := c.Query("size"); sizeStr != "" {
		if s, err := strconv.Atoi(sizeStr); err == nil && s > 0 && s <= 100 {
			size = s
		}
	}

	// Call service to list batch QR codes
	batches, total, err := h.batchService.ListBatchQRCodes(c.Request.Context(), appID, page, size)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Build response
	response := ListBatchQRCodesResponse{
		Items: batches,
		Total: total,
		Page:  page,
		Size:  size,
	}

	c.JSON(http.StatusOK, response)
}

// GetBatchQRCode handles retrieving a batch QR code by ID
func (h *BatchQRCodeHandler) GetBatchQRCode(c *gin.Context) {
	batchID := c.Param("id")
	if batchID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "batch id is required"})
		return
	}

	// Call service to get batch QR code
	batch, err := h.batchService.GetBatchQRCode(c.Request.Context(), batchID)
	if err != nil {
		if err.Error() == "batch QR code not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "batch QR code not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, batch)
}

// GenerateBatchQRCodes handles generating QR codes for a batch
func (h *BatchQRCodeHandler) GenerateBatchQRCodes(c *gin.Context) {
	batchID := c.Param("id")
	if batchID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "batch id is required"})
		return
	}

	// Call service to generate batch QR codes
	qrCodes, err := h.batchService.GenerateBatchQRCodes(c.Request.Context(), batchID)
	if err != nil {
		if err.Error() == "batch QR code not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "batch QR code not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, qrCodes)
}

// ListBatchQRCodesResponse represents the response for listing batch QR codes
type ListBatchQRCodesResponse struct {
	Items []*service.BatchQRCode `json:"items"`
	Total int64                  `json:"total"`
	Page  int                    `json:"page"`
	Size  int                    `json:"size"`
}