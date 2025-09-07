package handler

import (
	"net/http"
	"strconv"

	"cdk-office/internal/app/domain"
	"cdk-office/internal/app/service"
	"github.com/gin-gonic/gin"
)

// QRCodeHandlerInterface defines the interface for QR code handler
type QRCodeHandlerInterface interface {
	CreateQRCode(c *gin.Context)
	UpdateQRCode(c *gin.Context)
	DeleteQRCode(c *gin.Context)
	ListQRCodes(c *gin.Context)
	GetQRCode(c *gin.Context)
	GenerateQRCodeImage(c *gin.Context)
}

// QRCodeHandler implements the QRCodeHandlerInterface
type QRCodeHandler struct {
	qrCodeService service.QRCodeServiceInterface
}

// NewQRCodeHandler creates a new instance of QRCodeHandler
func NewQRCodeHandler() *QRCodeHandler {
	return &QRCodeHandler{
		qrCodeService: service.NewQRCodeService(),
	}
}

// CreateQRCodeRequest represents the request for creating a QR code
type CreateQRCodeRequest struct {
	AppID     string `json:"app_id" binding:"required"`
	Name      string `json:"name" binding:"required"`
	Content   string `json:"content" binding:"required"`
	Type      string `json:"type" binding:"required"` // static or dynamic
	URL       string `json:"url"`
	CreatedBy string `json:"created_by" binding:"required"`
}

// UpdateQRCodeRequest represents the request for updating a QR code
type UpdateQRCodeRequest struct {
	Name    string `json:"name"`
	Content string `json:"content"`
	URL     string `json:"url"`
}

// ListQRCodesRequest represents the request for listing QR codes
type ListQRCodesRequest struct {
	AppID string `form:"app_id" binding:"required"`
	Page  int    `form:"page"`
	Size  int    `form:"size"`
}

// CreateQRCode handles creating a new QR code
func (h *QRCodeHandler) CreateQRCode(c *gin.Context) {
	var req CreateQRCodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Call service to create QR code
	qrCode, err := h.qrCodeService.CreateQRCode(c.Request.Context(), &service.CreateQRCodeRequest{
		AppID:     req.AppID,
		Name:      req.Name,
		Content:   req.Content,
		Type:      req.Type,
		URL:       req.URL,
		CreatedBy: req.CreatedBy,
	})
	if err != nil {
		if err.Error() == "invalid QR code type" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid QR code type"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, qrCode)
}

// UpdateQRCode handles updating an existing QR code
func (h *QRCodeHandler) UpdateQRCode(c *gin.Context) {
	qrCodeID := c.Param("id")
	if qrCodeID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "QR code id is required"})
		return
	}

	var req UpdateQRCodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Call service to update QR code
	if err := h.qrCodeService.UpdateQRCode(c.Request.Context(), qrCodeID, &service.UpdateQRCodeRequest{
		Name:    req.Name,
		Content: req.Content,
		URL:     req.URL,
	}); err != nil {
		if err.Error() == "QR code not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "QR code not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "QR code updated successfully"})
}

// DeleteQRCode handles deleting a QR code
func (h *QRCodeHandler) DeleteQRCode(c *gin.Context) {
	qrCodeID := c.Param("id")
	if qrCodeID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "QR code id is required"})
		return
	}

	// Call service to delete QR code
	if err := h.qrCodeService.DeleteQRCode(c.Request.Context(), qrCodeID); err != nil {
		if err.Error() == "QR code not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "QR code not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "QR code deleted successfully"})
}

// ListQRCodes handles listing QR codes with pagination
func (h *QRCodeHandler) ListQRCodes(c *gin.Context) {
	// Parse query parameters
	appID := c.Query("app_id")
	if appID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "app id is required"})
		return
	}

	pageStr := c.Query("page")
	sizeStr := c.Query("size")

	page := 1
	size := 10

	var err error
	if pageStr != "" {
		page, err = strconv.Atoi(pageStr)
		if err != nil || page < 1 {
			page = 1
		}
	}

	if sizeStr != "" {
		size, err = strconv.Atoi(sizeStr)
		if err != nil || size < 1 || size > 100 {
			size = 10
		}
	}

	// Call service to list QR codes
	qrCodes, total, err := h.qrCodeService.ListQRCodes(c.Request.Context(), appID, page, size)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Build response
	response := ListQRCodesResponse{
		Items: qrCodes,
		Total: total,
		Page:  page,
		Size:  size,
	}

	c.JSON(http.StatusOK, response)
}

// GetQRCode handles retrieving a QR code by ID
func (h *QRCodeHandler) GetQRCode(c *gin.Context) {
	qrCodeID := c.Param("id")
	if qrCodeID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "QR code id is required"})
		return
	}

	// Call service to get QR code
	qrCode, err := h.qrCodeService.GetQRCode(c.Request.Context(), qrCodeID)
	if err != nil {
		if err.Error() == "QR code not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "QR code not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, qrCode)
}

// GenerateQRCodeImage handles generating a QR code image
func (h *QRCodeHandler) GenerateQRCodeImage(c *gin.Context) {
	qrCodeID := c.Param("id")
	if qrCodeID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "QR code id is required"})
		return
	}

	// Call service to generate QR code image
	imagePath, err := h.qrCodeService.GenerateQRCodeImage(c.Request.Context(), qrCodeID)
	if err != nil {
		if err.Error() == "QR code not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "QR code not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"image_path": imagePath})
}

// ListQRCodesResponse represents the response for listing QR codes
type ListQRCodesResponse struct {
	Items []*domain.QRCode `json:"items"`
	Total int64             `json:"total"`
	Page  int               `json:"page"`
	Size  int               `json:"size"`
}