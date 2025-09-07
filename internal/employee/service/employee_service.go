package service

import (
	"context"
	"errors"
	"time"

	"cdk-office/internal/employee/domain"
	"cdk-office/internal/shared/cache"
	"cdk-office/internal/shared/database"
	"cdk-office/internal/shared/utils"
	"cdk-office/pkg/logger"
	"gorm.io/gorm"
)

// EmployeeServiceInterface defines the interface for employee service
type EmployeeServiceInterface interface {
	CreateEmployee(ctx context.Context, req *CreateEmployeeRequest) (*domain.Employee, error)
	UpdateEmployee(ctx context.Context, empID string, req *UpdateEmployeeRequest) error
	DeleteEmployee(ctx context.Context, empID string) error
	ListEmployees(ctx context.Context, req *ListEmployeesRequest) ([]*domain.Employee, int64, error)
	GetEmployee(ctx context.Context, empID string) (*domain.Employee, error)
}

// EmployeeService implements the EmployeeServiceInterface
type EmployeeService struct {
	db *gorm.DB
}

// NewEmployeeService creates a new instance of EmployeeService
func NewEmployeeService() *EmployeeService {
	return &EmployeeService{
		db: database.GetDB(),
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
	BirthDate  time.Time `json:"birth_date" binding:"required"`
	HireDate   time.Time `json:"hire_date" binding:"required"`
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
	TeamID  string `json:"team_id"`
	DeptID  string `json:"dept_id"`
	Page    int    `json:"page"`
	Size    int    `json:"size"`
}

// CreateEmployee creates a new employee
func (s *EmployeeService) CreateEmployee(ctx context.Context, req *CreateEmployeeRequest) (*domain.Employee, error) {
	// Check if employee ID already exists
	var existingEmployee domain.Employee
	if err := s.db.Where("employee_id = ?", req.EmployeeID).First(&existingEmployee).Error; err == nil {
		return nil, errors.New("employee ID already exists")
	}

	// Check if user exists
	var userCount int64
	if err := s.db.Model(&domain.Employee{}).Where("user_id = ?", req.UserID).Count(&userCount).Error; err != nil {
		logger.Error("failed to count users", "error", err)
		return nil, errors.New("failed to create employee")
	}

	if userCount == 0 {
		// In a real application, you would check the actual user table
		// For now, we'll assume the user exists
	}

	// Check if department exists
	var deptCount int64
	if err := s.db.Model(&domain.Department{}).Where("id = ?", req.DeptID).Count(&deptCount).Error; err != nil {
		logger.Error("failed to count departments", "error", err)
		return nil, errors.New("failed to create employee")
	}

	if deptCount == 0 {
		return nil, errors.New("department not found")
	}

	// Create new employee
	employee := &domain.Employee{
		ID:         utils.GenerateEmployeeID(),
		UserID:     req.UserID,
		TeamID:     req.TeamID,
		DeptID:     req.DeptID,
		EmployeeID: req.EmployeeID,
		RealName:   req.RealName,
		Gender:     req.Gender,
		BirthDate:  req.BirthDate,
		HireDate:   req.HireDate,
		Position:   req.Position,
		Status:     "active", // Default status
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	// Save employee to database
	if err := s.db.Create(employee).Error; err != nil {
		logger.Error("failed to create employee", "error", err)
		return nil, errors.New("failed to create employee")
	}

	return employee, nil
}

// UpdateEmployee updates an existing employee
func (s *EmployeeService) UpdateEmployee(ctx context.Context, empID string, req *UpdateEmployeeRequest) error {
	// Find employee by ID
	var employee domain.Employee
	if err := s.db.Where("id = ?", empID).First(&employee).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("employee not found")
		}
		logger.Error("failed to find employee", "error", err)
		return errors.New("failed to update employee")
	}

	// Update employee fields
	if req.DeptID != "" {
		// Check if department exists
		var deptCount int64
		if err := s.db.Model(&domain.Department{}).Where("id = ?", req.DeptID).Count(&deptCount).Error; err != nil {
			logger.Error("failed to count departments", "error", err)
			return errors.New("failed to update employee")
		}

		if deptCount == 0 {
			return errors.New("department not found")
		}
		
		employee.DeptID = req.DeptID
	}
	
	if req.Position != "" {
		employee.Position = req.Position
	}
	
	if req.Status != "" {
		employee.Status = req.Status
	}
	
	employee.UpdatedAt = time.Now()

	// Save updated employee to database
	if err := s.db.Save(&employee).Error; err != nil {
		logger.Error("failed to update employee", "error", err)
		return errors.New("failed to update employee")
	}

	// Invalidate cache
	cacheKey := "employee:" + empID
	cache.Delete(cacheKey)

	return nil
}

// DeleteEmployee deletes an employee
func (s *EmployeeService) DeleteEmployee(ctx context.Context, empID string) error {
	// Find employee by ID
	var employee domain.Employee
	if err := s.db.Where("id = ?", empID).First(&employee).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("employee not found")
		}
		logger.Error("failed to find employee", "error", err)
		return errors.New("failed to delete employee")
	}

	// Delete employee from database
	if err := s.db.Delete(&employee).Error; err != nil {
		logger.Error("failed to delete employee", "error", err)
		return errors.New("failed to delete employee")
	}

	// Invalidate cache
	cacheKey := "employee:" + empID
	cache.Delete(cacheKey)

	return nil
}

// ListEmployees lists employees with pagination
func (s *EmployeeService) ListEmployees(ctx context.Context, req *ListEmployeesRequest) ([]*domain.Employee, int64, error) {
	// Validate pagination parameters
	if req.Page < 1 {
		req.Page = 1
	}
	if req.Size < 1 || req.Size > 100 {
		req.Size = 10
	}

	// Try to get employees list from cache first
	cacheKey := "employees_list"
	if req.TeamID != "" {
		cacheKey += ":team_" + req.TeamID
	}
	if req.DeptID != "" {
		cacheKey += ":dept_" + req.DeptID
	}
	cacheKey += ":page_" + string(rune(req.Page)) + ":size_" + string(rune(req.Size))

	var employees []*domain.Employee
	var total int64
	
	// Check if employees list exists in cache
	exists, err := cache.Exists(cacheKey + ":list")
	if err == nil && exists {
		// Get employees list from cache
		if err := cache.Get(cacheKey + ":list", &employees); err == nil {
			// Get total count from cache
			if err := cache.Get(cacheKey + ":total", &total); err == nil {
				return employees, total, nil
			}
		}
	}

	// Build query
	dbQuery := s.db.Model(&domain.Employee{})

	// Add filters
	if req.TeamID != "" {
		dbQuery = dbQuery.Where("team_id = ?", req.TeamID)
	}
	if req.DeptID != "" {
		dbQuery = dbQuery.Where("dept_id = ?", req.DeptID)
	}

	// Count total results
	if err := dbQuery.Count(&total).Error; err != nil {
		logger.Error("failed to count employees", "error", err)
		return nil, 0, errors.New("failed to list employees")
	}

	// Apply pagination
	offset := (req.Page - 1) * req.Size
	dbQuery = dbQuery.Offset(offset).Limit(req.Size).Order("created_at desc")

	// Execute query
	if err := dbQuery.Find(&employees).Error; err != nil {
		logger.Error("failed to list employees", "error", err)
		return nil, 0, errors.New("failed to list employees")
	}

	// Cache the employees list and total count for 5 minutes
	cache.Set(cacheKey + ":list", &employees, 5*time.Minute)
	cache.Set(cacheKey + ":total", &total, 5*time.Minute)

	return employees, total, nil
}

// GetEmployee retrieves an employee by ID
func (s *EmployeeService) GetEmployee(ctx context.Context, empID string) (*domain.Employee, error) {
	// Try to get employee from cache first
	cacheKey := "employee:" + empID
	var employee domain.Employee
	
	// Check if employee exists in cache
	exists, err := cache.Exists(cacheKey)
	if err == nil && exists {
		// Get employee from cache
		if err := cache.Get(cacheKey, &employee); err == nil {
			return &employee, nil
		}
	}

	// Get employee from database
	if err := s.db.Where("id = ?", empID).First(&employee).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("employee not found")
		}
		logger.Error("failed to find employee", "error", err)
		return nil, errors.New("failed to get employee")
	}

	// Cache the employee for 10 minutes
	cache.Set(cacheKey, &employee, 10*time.Minute)

	return &employee, nil
}