package handler

import (
	"net/http"

	"cdk-office/internal/employee/service"
	"github.com/gin-gonic/gin"
)

// DepartmentHandlerInterface defines the interface for department handler
type DepartmentHandlerInterface interface {
	CreateDepartment(c *gin.Context)
	UpdateDepartment(c *gin.Context)
	DeleteDepartment(c *gin.Context)
	ListDepartments(c *gin.Context)
	GetDepartment(c *gin.Context)
}

// DepartmentHandler implements the DepartmentHandlerInterface
type DepartmentHandler struct {
	departmentService service.DepartmentServiceInterface
}

// NewDepartmentHandler creates a new instance of DepartmentHandler
func NewDepartmentHandler() *DepartmentHandler {
	return &DepartmentHandler{
		departmentService: service.NewDepartmentService(),
	}
}

// CreateDepartmentRequest represents the request for creating a department
type CreateDepartmentRequest struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
	TeamID      string `json:"team_id" binding:"required"`
	ParentID    string `json:"parent_id"`
}

// UpdateDepartmentRequest represents the request for updating a department
type UpdateDepartmentRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	ParentID    string `json:"parent_id"`
}

// CreateDepartment handles creating a new department
func (h *DepartmentHandler) CreateDepartment(c *gin.Context) {
	var req CreateDepartmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Call service to create department
	if err := h.departmentService.CreateDepartment(c.Request.Context(), &service.CreateDepartmentRequest{
		Name:        req.Name,
		Description: req.Description,
		TeamID:      req.TeamID,
		ParentID:    req.ParentID,
	}); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "department created successfully"})
}

// UpdateDepartment handles updating an existing department
func (h *DepartmentHandler) UpdateDepartment(c *gin.Context) {
	deptID := c.Param("id")
	if deptID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "department id is required"})
		return
	}

	var req UpdateDepartmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Call service to update department
	if err := h.departmentService.UpdateDepartment(c.Request.Context(), deptID, &service.UpdateDepartmentRequest{
		Name:        req.Name,
		Description: req.Description,
		ParentID:    req.ParentID,
	}); err != nil {
		if err.Error() == "department not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "department not found"})
			return
		}
		if err.Error() == "parent department not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "parent department not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "department updated successfully"})
}

// DeleteDepartment handles deleting a department
func (h *DepartmentHandler) DeleteDepartment(c *gin.Context) {
	deptID := c.Param("id")
	if deptID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "department id is required"})
		return
	}

	// Call service to delete department
	if err := h.departmentService.DeleteDepartment(c.Request.Context(), deptID); err != nil {
		if err.Error() == "department not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "department not found"})
			return
		}
		if err.Error() == "cannot delete department with child departments" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "cannot delete department with child departments"})
			return
		}
		if err.Error() == "cannot delete department with employees" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "cannot delete department with employees"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "department deleted successfully"})
}

// ListDepartments handles listing all departments for a team
func (h *DepartmentHandler) ListDepartments(c *gin.Context) {
	teamID := c.Query("team_id")
	if teamID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "team id is required"})
		return
	}

	// Call service to list departments
	departments, err := h.departmentService.ListDepartments(c.Request.Context(), teamID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, departments)
}

// GetDepartment handles retrieving a department by ID
func (h *DepartmentHandler) GetDepartment(c *gin.Context) {
	deptID := c.Param("id")
	if deptID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "department id is required"})
		return
	}

	// Call service to get department
	department, err := h.departmentService.GetDepartment(c.Request.Context(), deptID)
	if err != nil {
		if err.Error() == "department not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "department not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, department)
}