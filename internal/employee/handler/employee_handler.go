package handler

import (
	"net/http"
	"strconv"
	"time"

	"cdk-office/internal/employee/domain"
	"cdk-office/internal/employee/service"
	"github.com/gin-gonic/gin"
)

// EmployeeHandlerInterface defines the interface for employee handler
type EmployeeHandlerInterface interface {
	CreateEmployee(c *gin.Context)
	UpdateEmployee(c *gin.Context)
	DeleteEmployee(c *gin.Context)
	ListEmployees(c *gin.Context)
	GetEmployee(c *gin.Context)
}

// EmployeeHandler implements the EmployeeHandlerInterface
type EmployeeHandler struct {
	employeeService service.EmployeeServiceInterface
}

// NewEmployeeHandler creates a new instance of EmployeeHandler
func NewEmployeeHandler() *EmployeeHandler {
	return &EmployeeHandler{
		employeeService: service.NewEmployeeService(),
	}
}

// NewEmployeeHandlerWithService creates a new instance of EmployeeHandler with a specific service
func NewEmployeeHandlerWithService(employeeService service.EmployeeServiceInterface) *EmployeeHandler {
	return &EmployeeHandler{
		employeeService: employeeService,
	}
}

// CreateEmployeeRequest represents the request for creating an employee
type CreateEmployeeRequest struct {
	UserID     string    `json:"user_id" binding:"required"`
	TeamID     string    `json:"team_id" binding:"required"`
	DeptID     string    `json:"dept_id" binding:"required"`
	EmployeeID string    `json:"employee_id" binding:"required"`
	RealName   string    `json:"real_name" binding:"required"`
	Gender     string    `json:"gender" binding:"required"`
	BirthDate  string    `json:"birth_date" binding:"required"`
	HireDate   string    `json:"hire_date" binding:"required"`
	Position   string    `json:"position" binding:"required"`
}

// UpdateEmployeeRequest represents the request for updating an employee
type UpdateEmployeeRequest struct {
	DeptID    string    `json:"dept_id"`
	Position  string    `json:"position"`
	Status    string    `json:"status"`
}

// ListEmployeesRequest represents the request for listing employees
type ListEmployeesRequest struct {
	TeamID  string `form:"team_id"`
	DeptID  string `form:"dept_id"`
	Page    int    `form:"page"`
	Size    int    `form:"size"`
}

// CreateEmployee handles creating a new employee
func (h *EmployeeHandler) CreateEmployee(c *gin.Context) {
	var req CreateEmployeeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Parse dates
	birthDate, err := time.Parse("2006-01-02", req.BirthDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid birth_date format"})
		return
	}

	hireDate, err := time.Parse("2006-01-02", req.HireDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid hire_date format"})
		return
	}

	// Call service to create employee
	employee, err := h.employeeService.CreateEmployee(c.Request.Context(), &service.CreateEmployeeRequest{
		UserID:     req.UserID,
		TeamID:     req.TeamID,
		DeptID:     req.DeptID,
		EmployeeID: req.EmployeeID,
		RealName:   req.RealName,
		Gender:     req.Gender,
		BirthDate:  birthDate,
		HireDate:   hireDate,
		Position:   req.Position,
	})
	if err != nil {
		if err.Error() == "employee ID already exists" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "employee ID already exists"})
			return
		}
		if err.Error() == "department not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "department not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, employee)
}

// UpdateEmployee handles updating an existing employee
func (h *EmployeeHandler) UpdateEmployee(c *gin.Context) {
	empID := c.Param("id")
	if empID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "employee id is required"})
		return
	}

	var req UpdateEmployeeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Call service to update employee
	if err := h.employeeService.UpdateEmployee(c.Request.Context(), empID, &service.UpdateEmployeeRequest{
		DeptID:    req.DeptID,
		Position:  req.Position,
		Status:    req.Status,
	}); err != nil {
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

	c.JSON(http.StatusOK, gin.H{"message": "employee updated successfully"})
}

// DeleteEmployee handles deleting an employee
func (h *EmployeeHandler) DeleteEmployee(c *gin.Context) {
	empID := c.Param("id")
	if empID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "employee id is required"})
		return
	}

	// Call service to delete employee
	if err := h.employeeService.DeleteEmployee(c.Request.Context(), empID); err != nil {
		if err.Error() == "employee not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "employee not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "employee deleted successfully"})
}

// ListEmployees handles listing employees with pagination
func (h *EmployeeHandler) ListEmployees(c *gin.Context) {
	// Parse query parameters
	var req ListEmployeesRequest
	req.TeamID = c.Query("team_id")
	req.DeptID = c.Query("dept_id")
	
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

	// Call service to list employees
	employees, total, err := h.employeeService.ListEmployees(c.Request.Context(), &service.ListEmployeesRequest{
		TeamID: req.TeamID,
		DeptID: req.DeptID,
		Page:   req.Page,
		Size:   req.Size,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Build response
	response := ListEmployeesResponse{
		Items: employees,
		Total: total,
		Page:  req.Page,
		Size:  req.Size,
	}

	c.JSON(http.StatusOK, response)
}

// GetEmployee handles retrieving an employee by ID
func (h *EmployeeHandler) GetEmployee(c *gin.Context) {
	empID := c.Param("id")
	if empID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "employee id is required"})
		return
	}

	// Call service to get employee
	employee, err := h.employeeService.GetEmployee(c.Request.Context(), empID)
	if err != nil {
		if err.Error() == "employee not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "employee not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, employee)
}

// ListEmployeesResponse represents the response for listing employees
type ListEmployeesResponse struct {
	Items []*domain.Employee `json:"items"`
	Total int64             `json:"total"`
	Page  int               `json:"page"`
	Size  int               `json:"size"`
}