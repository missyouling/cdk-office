package service

import (
	"context"
	"errors"
	"time"

	"cdk-office/internal/employee/domain"
	"cdk-office/internal/shared/database"
	"cdk-office/internal/shared/utils"
	"cdk-office/pkg/logger"
	"gorm.io/gorm"
)

// DepartmentServiceInterface defines the interface for department service
type DepartmentServiceInterface interface {
	CreateDepartment(ctx context.Context, req *CreateDepartmentRequest) error
	UpdateDepartment(ctx context.Context, deptID string, req *UpdateDepartmentRequest) error
	DeleteDepartment(ctx context.Context, deptID string) error
	ListDepartments(ctx context.Context, teamID string) ([]*domain.Department, error)
	GetDepartment(ctx context.Context, deptID string) (*domain.Department, error)
}

// DepartmentService implements the DepartmentServiceInterface
type DepartmentService struct {
	db *gorm.DB
}

// NewDepartmentService creates a new instance of DepartmentService
func NewDepartmentService() *DepartmentService {
	return &DepartmentService{
		db: database.GetDB(),
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

// CreateDepartment creates a new department
func (s *DepartmentService) CreateDepartment(ctx context.Context, req *CreateDepartmentRequest) error {
	// Check if parent department exists (if parentID is provided)
	if req.ParentID != "" {
		var parentDept domain.Department
		if err := s.db.Where("id = ?", req.ParentID).First(&parentDept).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return errors.New("parent department not found")
			}
			logger.Error("failed to find parent department", "error", err)
			return errors.New("failed to create department")
		}
	}

	// Determine department level
	level := 1
	if req.ParentID != "" {
		var parentDept domain.Department
		if err := s.db.Where("id = ?", req.ParentID).First(&parentDept).Error; err != nil {
			logger.Error("failed to find parent department", "error", err)
			return errors.New("failed to create department")
		}
		level = parentDept.Level + 1
	}

	// Create new department
	department := &domain.Department{
		ID:          utils.GenerateDepartmentID(),
		Name:        req.Name,
		Description: req.Description,
		TeamID:      req.TeamID,
		ParentID:    req.ParentID,
		Level:       level,
		SortOrder:   0, // Default sort order
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// Save department to database
	if err := s.db.Create(department).Error; err != nil {
		logger.Error("failed to create department", "error", err)
		return errors.New("failed to create department")
	}

	return nil
}

// UpdateDepartment updates an existing department
func (s *DepartmentService) UpdateDepartment(ctx context.Context, deptID string, req *UpdateDepartmentRequest) error {
	// Find department by ID
	var department domain.Department
	if err := s.db.Where("id = ?", deptID).First(&department).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("department not found")
		}
		logger.Error("failed to find department", "error", err)
		return errors.New("failed to update department")
	}

	// Update department fields
	if req.Name != "" {
		department.Name = req.Name
	}
	if req.Description != "" {
		department.Description = req.Description
	}
	if req.ParentID != "" {
		// Check if parent department exists
		var parentDept domain.Department
		if err := s.db.Where("id = ?", req.ParentID).First(&parentDept).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return errors.New("parent department not found")
			}
			logger.Error("failed to find parent department", "error", err)
			return errors.New("failed to update department")
		}
		
		department.ParentID = req.ParentID
		department.Level = parentDept.Level + 1
	}
	
	department.UpdatedAt = time.Now()

	// Save updated department to database
	if err := s.db.Save(&department).Error; err != nil {
		logger.Error("failed to update department", "error", err)
		return errors.New("failed to update department")
	}

	return nil
}

// DeleteDepartment deletes a department
func (s *DepartmentService) DeleteDepartment(ctx context.Context, deptID string) error {
	// Find department by ID
	var department domain.Department
	if err := s.db.Where("id = ?", deptID).First(&department).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("department not found")
		}
		logger.Error("failed to find department", "error", err)
		return errors.New("failed to delete department")
	}

	// Check if department has child departments
	var childCount int64
	if err := s.db.Model(&domain.Department{}).Where("parent_id = ?", deptID).Count(&childCount).Error; err != nil {
		logger.Error("failed to count child departments", "error", err)
		return errors.New("failed to delete department")
	}

	if childCount > 0 {
		return errors.New("cannot delete department with child departments")
	}

	// Check if department has employees
	var employeeCount int64
	if err := s.db.Model(&domain.Employee{}).Where("dept_id = ?", deptID).Count(&employeeCount).Error; err != nil {
		logger.Error("failed to count employees in department", "error", err)
		return errors.New("failed to delete department")
	}

	if employeeCount > 0 {
		return errors.New("cannot delete department with employees")
	}

	// Delete department from database
	if err := s.db.Delete(&department).Error; err != nil {
		logger.Error("failed to delete department", "error", err)
		return errors.New("failed to delete department")
	}

	return nil
}

// ListDepartments lists all departments for a team
func (s *DepartmentService) ListDepartments(ctx context.Context, teamID string) ([]*domain.Department, error) {
	var departments []*domain.Department

	// Build query
	query := s.db.Where("team_id = ?", teamID).Order("level asc, sort_order asc")

	// Execute query
	if err := query.Find(&departments).Error; err != nil {
		logger.Error("failed to list departments", "error", err)
		return nil, errors.New("failed to list departments")
	}

	return departments, nil
}

// GetDepartment retrieves a department by ID
func (s *DepartmentService) GetDepartment(ctx context.Context, deptID string) (*domain.Department, error) {
	var department domain.Department
	if err := s.db.Where("id = ?", deptID).First(&department).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("department not found")
		}
		logger.Error("failed to find department", "error", err)
		return nil, errors.New("failed to get department")
	}

	return &department, nil
}

