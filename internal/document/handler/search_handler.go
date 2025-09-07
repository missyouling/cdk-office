package handler

import (
	"net/http"
	"strconv"

	"cdk-office/internal/document/domain"
	"cdk-office/internal/document/service"
	"github.com/gin-gonic/gin"
)

// SearchHandlerInterface defines the interface for document search handler
type SearchHandlerInterface interface {
	SearchDocuments(c *gin.Context)
}

// SearchHandler implements the SearchHandlerInterface
type SearchHandler struct {
	searchService service.SearchServiceInterface
}

// NewSearchHandler creates a new instance of SearchHandler
func NewSearchHandler() *SearchHandler {
	return &SearchHandler{
		searchService: service.NewSearchService(),
	}
}

// SearchDocumentsRequest represents the request for searching documents
type SearchDocumentsRequest struct {
	Query  string `form:"q"`
	TeamID string `form:"team_id"`
	Page   int    `form:"page"`
	Size   int    `form:"size"`
}

// SearchDocumentsResponse represents the response for searching documents
type SearchDocumentsResponse struct {
	Items []*domain.Document `json:"items"`
	Total int64              `json:"total"`
	Page  int                `json:"page"`
	Size  int                `json:"size"`
}

// SearchDocuments handles searching for documents
func (h *SearchHandler) SearchDocuments(c *gin.Context) {
	// Parse query parameters
	var req SearchDocumentsRequest
	req.Query = c.Query("q")
	req.TeamID = c.Query("team_id")
	
	// Parse page and size parameters
	pageStr := c.Query("page")
	sizeStr := c.Query("size")
	
	var err error
	if pageStr != "" {
		req.Page, err = strconv.Atoi(pageStr)
		if err != nil || req.Page < 1 {
			req.Page = 1
		}
	} else {
		req.Page = 1
	}
	
	if sizeStr != "" {
		req.Size, err = strconv.Atoi(sizeStr)
		if err != nil || req.Size < 1 || req.Size > 100 {
			req.Size = 10
		}
	} else {
		req.Size = 10
	}

	// Call service to search documents
	documents, total, err := h.searchService.SearchDocuments(c.Request.Context(), req.Query, req.TeamID, req.Page, req.Size)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Build response
	response := SearchDocumentsResponse{
		Items: documents,
		Total: total,
		Page:  req.Page,
		Size:  req.Size,
	}

	c.JSON(http.StatusOK, response)
}