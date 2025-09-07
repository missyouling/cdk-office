package handler

import (
	"net/http"

	"cdk-office/internal/document/service"
	"github.com/gin-gonic/gin"
)

// CategoryHandlerInterface defines the interface for category handler
type CategoryHandlerInterface interface {
	CreateCategory(c *gin.Context)
	GetCategory(c *gin.Context)
	UpdateCategory(c *gin.Context)
	DeleteCategory(c *gin.Context)
	ListCategories(c *gin.Context)
	AssignDocumentToCategory(c *gin.Context)
	RemoveDocumentFromCategory(c *gin.Context)
	GetDocumentCategories(c *gin.Context)
}

// CategoryHandler implements the CategoryHandlerInterface
type CategoryHandler struct {
	categoryService service.CategoryServiceInterface
}

// NewCategoryHandler creates a new instance of CategoryHandler
func NewCategoryHandler() *CategoryHandler {
	return &CategoryHandler{
		categoryService: service.NewCategoryService(),
	}
}

// CreateCategoryRequest represents the request for creating a category
type CreateCategoryRequest struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
	ParentID    string `json:"parent_id"`
}

// UpdateCategoryRequest represents the request for updating a category
type UpdateCategoryRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

// AssignDocumentToCategoryRequest represents the request for assigning a document to a category
type AssignDocumentToCategoryRequest struct {
	DocumentID string `json:"document_id" binding:"required"`
	CategoryID string `json:"category_id" binding:"required"`
}

// RemoveDocumentFromCategoryRequest represents the request for removing a document from a category
type RemoveDocumentFromCategoryRequest struct {
	DocumentID string `json:"document_id" binding:"required"`
	CategoryID string `json:"category_id" binding:"required"`
}

// CreateCategory handles creating a new category
func (h *CategoryHandler) CreateCategory(c *gin.Context) {
	var req CreateCategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Call service to create category
	category, err := h.categoryService.CreateCategory(c.Request.Context(), req.Name, req.Description, req.ParentID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, category)
}

// GetCategory handles retrieving a category by ID
func (h *CategoryHandler) GetCategory(c *gin.Context) {
	categoryID := c.Param("id")
	if categoryID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "category id is required"})
		return
	}

	// Call service to get category
	category, err := h.categoryService.GetCategory(c.Request.Context(), categoryID)
	if err != nil {
		if err.Error() == "category not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "category not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, category)
}

// UpdateCategory handles updating a category
func (h *CategoryHandler) UpdateCategory(c *gin.Context) {
	categoryID := c.Param("id")
	if categoryID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "category id is required"})
		return
	}

	var req UpdateCategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Call service to update category
	if err := h.categoryService.UpdateCategory(c.Request.Context(), categoryID, req.Name, req.Description); err != nil {
		if err.Error() == "category not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "category not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "category updated successfully"})
}

// DeleteCategory handles deleting a category
func (h *CategoryHandler) DeleteCategory(c *gin.Context) {
	categoryID := c.Param("id")
	if categoryID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "category id is required"})
		return
	}

	// Call service to delete category
	if err := h.categoryService.DeleteCategory(c.Request.Context(), categoryID); err != nil {
		if err.Error() == "category not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "category not found"})
			return
		}
		if err.Error() == "cannot delete category with child categories" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "cannot delete category with child categories"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "category deleted successfully"})
}

// ListCategories handles listing categories
func (h *CategoryHandler) ListCategories(c *gin.Context) {
	// Get parent_id from query parameters (optional)
	parentID := c.Query("parent_id")

	// Call service to list categories
	categories, err := h.categoryService.ListCategories(c.Request.Context(), parentID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, categories)
}

// AssignDocumentToCategory handles assigning a document to a category
func (h *CategoryHandler) AssignDocumentToCategory(c *gin.Context) {
	var req AssignDocumentToCategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Call service to assign document to category
	if err := h.categoryService.AssignDocumentToCategory(c.Request.Context(), req.DocumentID, req.CategoryID); err != nil {
		if err.Error() == "document not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "document not found"})
			return
		}
		if err.Error() == "category not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "category not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "document assigned to category successfully"})
}

// RemoveDocumentFromCategory handles removing a document from a category
func (h *CategoryHandler) RemoveDocumentFromCategory(c *gin.Context) {
	var req RemoveDocumentFromCategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Call service to remove document from category
	if err := h.categoryService.RemoveDocumentFromCategory(c.Request.Context(), req.DocumentID, req.CategoryID); err != nil {
		if err.Error() == "document-category relation not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "document-category relation not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "document removed from category successfully"})
}

// GetDocumentCategories handles retrieving all categories for a document
func (h *CategoryHandler) GetDocumentCategories(c *gin.Context) {
	documentID := c.Param("document_id")
	if documentID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "document id is required"})
		return
	}

	// Call service to get document categories
	categories, err := h.categoryService.GetDocumentCategories(c.Request.Context(), documentID)
	if err != nil {
		if err.Error() == "document not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "document not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, categories)
}