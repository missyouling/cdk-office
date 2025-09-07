package handler

import (
	"net/http"
	"time"

	"cdk-office/internal/employee/service"
	"github.com/gin-gonic/gin"
)

// LifecycleHandlerInterface defines the interface for employee lifecycle handler
type LifecycleHandlerInterface interface {
	PromoteEmployee(c *gin.Context)
	TransferEmployee(c *gin.Context)
	TerminateEmployee(c *gin.Context)
	GetEmployeeLifecycleHistory(c *gin.Context)
}

// LifecycleHandler implements the LifecycleHandlerInterface
type LifecycleHandler struct {
	lifecycleService service.LifecycleServiceInterface
}

// NewLifecycleHandler creates a new instance of LifecycleHandler
func NewLifecycleHandler() *LifecycleHandler {
	return &LifecycleHandler{
		lifecycleService: service.NewLifecycleService(),
	}
}

// PromoteEmployeeRequest represents the request for promoting an employee
type PromoteEmployeeRequest struct {
	EmployeeID string `json:"employee_id" binding:"required"`
	NewPosition string `json:"new_position" binding:"required"`
}

// TransferEmployeeRequest represents the request for transferring an employee
type TransferEmployeeRequest struct {
	EmployeeID string `json:"employee_id" binding:"required"`
	NewDeptID  string `json:"new_dept_id" binding:"required"`
}

// TerminateEmployeeRequest represents the request for terminating an employee
type TerminateEmployeeRequest struct {
	EmployeeID      string `json:"employee_id" binding:"required"`
	TerminationDate string `json:"termination_date" binding:"required"`
	Reason          string `json:"reason" binding:"required"`
}

// PromoteEmployee handles promoting an employee to a new position
func (h *LifecycleHandler) PromoteEmployee(c *gin.Context) {
	var req PromoteEmployeeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Call service to promote employee
	if err := h.lifecycleService.PromoteEmployee(c.Request.Context(), req.EmployeeID, req.NewPosition); err != nil {
		if err.Error() == "employee not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "employee not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "employee promoted successfully"})
}

// TransferEmployee handles transferring an employee to a new department
func (h *LifecycleHandler) TransferEmployee(c *gin.Context) {
	var req TransferEmployeeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Call service to transfer employee
	if err := h.lifecycleService.TransferEmployee(c.Request.Context(), req.EmployeeID, req.NewDeptID); err != nil {
		if err.Error() == "employee not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "employee not found"})
			return
		}
		if err.Error() == "department not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "department not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "employee transferred successfully"})
}

// TerminateEmployee handles terminating an employee
func (h *LifecycleHandler) TerminateEmployee(c *gin.Context) {
	var req TerminateEmployeeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Parse termination date
	terminationDate, err := time.Parse("2006-01-02", req.TerminationDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid termination_date format"})
		return
	}

	// Call service to terminate employee
	if err := h.lifecycleService.TerminateEmployee(c.Request.Context(), req.EmployeeID, terminationDate, req.Reason); err != nil {
		if err.Error() == "employee not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "employee not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "employee terminated successfully"})
}

// GetEmployeeLifecycleHistory handles retrieving an employee's lifecycle history
func (h *LifecycleHandler) GetEmployeeLifecycleHistory(c *gin.Context) {
	empID := c.Param("id")
	if empID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "employee id is required"})
		return
	}

	// Call service to get employee lifecycle history
	events, err := h.lifecycleService.GetEmployeeLifecycleHistory(c.Request.Context(), empID)
	if err != nil {
		if err.Error() == "employee not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "employee not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, events)
}