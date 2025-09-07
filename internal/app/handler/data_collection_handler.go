package handler

import (
	"net/http"
	"strconv"

	"cdk-office/internal/app/service"
	"github.com/gin-gonic/gin"
)

// DataCollectionHandlerInterface defines the interface for data collection handler
type DataCollectionHandlerInterface interface {
	CreateDataCollection(c *gin.Context)
	UpdateDataCollection(c *gin.Context)
	DeleteDataCollection(c *gin.Context)
	ListDataCollections(c *gin.Context)
	GetDataCollection(c *gin.Context)
	SubmitDataEntry(c *gin.Context)
	ListDataEntries(c *gin.Context)
	ExportDataEntries(c *gin.Context)
}

// DataCollectionHandler implements the DataCollectionHandlerInterface
type DataCollectionHandler struct {
	dataService service.DataCollectionServiceInterface
}

// NewDataCollectionHandler creates a new instance of DataCollectionHandler
func NewDataCollectionHandler() *DataCollectionHandler {
	return &DataCollectionHandler{
		dataService: service.NewDataCollectionService(),
	}
}

// CreateDataCollectionRequest represents the request for creating a data collection
type CreateDataCollectionRequest struct {
	AppID       string `json:"app_id" binding:"required"`
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
	Schema      string `json:"schema" binding:"required"`
	Config      string `json:"config"`
	CreatedBy   string `json:"created_by" binding:"required"`
}

// UpdateDataCollectionRequest represents the request for updating a data collection
type UpdateDataCollectionRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Schema      string `json:"schema"`
	Config      string `json:"config"`
	IsActive    *bool  `json:"is_active"`
}

// SubmitDataEntryRequest represents the request for submitting a data entry
type SubmitDataEntryRequest struct {
	CollectionID string `json:"collection_id" binding:"required"`
	Data         string `json:"data" binding:"required"`
	CreatedBy    string `json:"created_by" binding:"required"`
}

// ListDataCollectionsRequest represents the request for listing data collections
type ListDataCollectionsRequest struct {
	AppID string `form:"app_id" binding:"required"`
	Page  int    `form:"page"`
	Size  int    `form:"size"`
}

// ListDataEntriesRequest represents the request for listing data entries
type ListDataEntriesRequest struct {
	CollectionID string `form:"collection_id" binding:"required"`
	Page         int    `form:"page"`
	Size         int    `form:"size"`
}

// CreateDataCollection handles creating a new data collection
func (h *DataCollectionHandler) CreateDataCollection(c *gin.Context) {
	var req CreateDataCollectionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Call service to create data collection
	collection, err := h.dataService.CreateDataCollection(c.Request.Context(), &service.CreateDataCollectionRequest{
		AppID:       req.AppID,
		Name:        req.Name,
		Description: req.Description,
		Schema:      req.Schema,
		Config:      req.Config,
		CreatedBy:   req.CreatedBy,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, collection)
}

// UpdateDataCollection handles updating an existing data collection
func (h *DataCollectionHandler) UpdateDataCollection(c *gin.Context) {
	collectionID := c.Param("id")
	if collectionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "collection id is required"})
		return
	}

	var req UpdateDataCollectionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Call service to update data collection
	if err := h.dataService.UpdateDataCollection(c.Request.Context(), collectionID, &service.UpdateDataCollectionRequest{
		Name:        req.Name,
		Description: req.Description,
		Schema:      req.Schema,
		Config:      req.Config,
		IsActive:    req.IsActive,
	}); err != nil {
		if err.Error() == "data collection not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "data collection not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "data collection updated successfully"})
}

// DeleteDataCollection handles deleting a data collection
func (h *DataCollectionHandler) DeleteDataCollection(c *gin.Context) {
	collectionID := c.Param("id")
	if collectionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "collection id is required"})
		return
	}

	// Call service to delete data collection
	if err := h.dataService.DeleteDataCollection(c.Request.Context(), collectionID); err != nil {
		if err.Error() == "data collection not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "data collection not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "data collection deleted successfully"})
}

// ListDataCollections handles listing data collections with pagination
func (h *DataCollectionHandler) ListDataCollections(c *gin.Context) {
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

	// Call service to list data collections
	collections, total, err := h.dataService.ListDataCollections(c.Request.Context(), appID, page, size)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Build response
	response := ListDataCollectionsResponse{
		Items: collections,
		Total: total,
		Page:  page,
		Size:  size,
	}

	c.JSON(http.StatusOK, response)
}

// GetDataCollection handles retrieving a data collection by ID
func (h *DataCollectionHandler) GetDataCollection(c *gin.Context) {
	collectionID := c.Param("id")
	if collectionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "collection id is required"})
		return
	}

	// Call service to get data collection
	collection, err := h.dataService.GetDataCollection(c.Request.Context(), collectionID)
	if err != nil {
		if err.Error() == "data collection not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "data collection not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, collection)
}

// SubmitDataEntry handles submitting a new data entry to a collection
func (h *DataCollectionHandler) SubmitDataEntry(c *gin.Context) {
	var req SubmitDataEntryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Call service to submit data entry
	entry, err := h.dataService.SubmitDataEntry(c.Request.Context(), &service.SubmitDataEntryRequest{
		CollectionID: req.CollectionID,
		Data:         req.Data,
		CreatedBy:    req.CreatedBy,
	})
	if err != nil {
		if err.Error() == "data collection not found or inactive" {
			c.JSON(http.StatusNotFound, gin.H{"error": "data collection not found or inactive"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, entry)
}

// ListDataEntries handles listing data entries with pagination
func (h *DataCollectionHandler) ListDataEntries(c *gin.Context) {
	// Parse query parameters
	collectionID := c.Query("collection_id")
	if collectionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "collection id is required"})
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

	// Call service to list data entries
	entries, total, err := h.dataService.ListDataEntries(c.Request.Context(), collectionID, page, size)
	if err != nil {
		if err.Error() == "data collection not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "data collection not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Build response
	response := ListDataEntriesResponse{
		Items: entries,
		Total: total,
		Page:  page,
		Size:  size,
	}

	c.JSON(http.StatusOK, response)
}

// ExportDataEntries handles exporting all data entries from a collection
func (h *DataCollectionHandler) ExportDataEntries(c *gin.Context) {
	collectionID := c.Param("id")
	if collectionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "collection id is required"})
		return
	}

	// Call service to export data entries
	entries, err := h.dataService.ExportDataEntries(c.Request.Context(), collectionID)
	if err != nil {
		if err.Error() == "data collection not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "data collection not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, entries)
}

// ListDataCollectionsResponse represents the response for listing data collections
type ListDataCollectionsResponse struct {
	Items []*service.DataCollection `json:"items"`
	Total int64                     `json:"total"`
	Page  int                       `json:"page"`
	Size  int                       `json:"size"`
}

// ListDataEntriesResponse represents the response for listing data entries
type ListDataEntriesResponse struct {
	Items []*service.DataCollectionEntry `json:"items"`
	Total int64                          `json:"total"`
	Page  int                            `json:"page"`
	Size  int                            `json:"size"`
}