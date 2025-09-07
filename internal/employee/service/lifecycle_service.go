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

// LifecycleServiceInterface defines the interface for employee lifecycle service
type LifecycleServiceInterface interface {
	PromoteEmployee(ctx context.Context, empID string, newPosition string) error
	TransferEmployee(ctx context.Context, empID, newDeptID string) error
	TerminateEmployee(ctx context.Context, empID string, terminationDate time.Time, reason string) error
	GetEmployeeLifecycleHistory(ctx context.Context, empID string) ([]*EmployeeLifecycleEvent, error)
}

// LifecycleService implements the LifecycleServiceInterface
type LifecycleService struct {
	db *gorm.DB
}

// NewLifecycleService creates a new instance of LifecycleService
func NewLifecycleService() *LifecycleService {
	return &LifecycleService{
		db: database.GetDB(),
	}
}

// EmployeeLifecycleEvent represents an event in an employee's lifecycle
type EmployeeLifecycleEvent struct {
	ID             string    `json:"id" gorm:"primaryKey"`
	EmployeeID     string    `json:"employee_id" gorm:"index"`
	EventType      string    `json:"event_type" gorm:"size:50"`
	OldValue       string    `json:"old_value" gorm:"size:255"`
	NewValue       string    `json:"new_value" gorm:"size:255"`
	EffectiveDate  time.Time `json:"effective_date"`
	Reason         string    `json:"reason" gorm:"type:text"`
	CreatedAt      time.Time `json:"created_at"`
}

// PromoteEmployee promotes an employee to a new position
func (s *LifecycleService) PromoteEmployee(ctx context.Context, empID string, newPosition string) error {
	// Find employee by ID
	var employee domain.Employee
	if err := s.db.Where("id = ?", empID).First(&employee).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("employee not found")
		}
		logger.Error("failed to find employee", "error", err)
		return errors.New("failed to promote employee")
	}

	// Store old position
	oldPosition := employee.Position

	// Update employee position
	employee.Position = newPosition
	employee.UpdatedAt = time.Now()

	// Save updated employee to database
	if err := s.db.Save(&employee).Error; err != nil {
		logger.Error("failed to update employee position", "error", err)
		return errors.New("failed to promote employee")
	}

	// Create lifecycle event
	event := &EmployeeLifecycleEvent{
		ID:            utils.GenerateLifecycleID(),
		EmployeeID:    empID,
		EventType:     "promotion",
		OldValue:      oldPosition,
		NewValue:      newPosition,
		EffectiveDate: time.Now(),
		Reason:        "Promotion to " + newPosition,
		CreatedAt:     time.Now(),
	}

	if err := s.db.Create(event).Error; err != nil {
		logger.Error("failed to create lifecycle event", "error", err)
		// Don't return error here as the promotion was successful
	}

	return nil
}

// TransferEmployee transfers an employee to a new department
func (s *LifecycleService) TransferEmployee(ctx context.Context, empID, newDeptID string) error {
	// Find employee by ID
	var employee domain.Employee
	if err := s.db.Where("id = ?", empID).First(&employee).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("employee not found")
		}
		logger.Error("failed to find employee", "error", err)
		return errors.New("failed to transfer employee")
	}

	// Check if new department exists
	var deptCount int64
	if err := s.db.Model(&domain.Department{}).Where("id = ?", newDeptID).Count(&deptCount).Error; err != nil {
		logger.Error("failed to count departments", "error", err)
		return errors.New("failed to transfer employee")
	}

	if deptCount == 0 {
		return errors.New("department not found")
	}

	// Store old department
	oldDeptID := employee.DeptID

	// Update employee department
	employee.DeptID = newDeptID
	employee.UpdatedAt = time.Now()

	// Save updated employee to database
	if err := s.db.Save(&employee).Error; err != nil {
		logger.Error("failed to update employee department", "error", err)
		return errors.New("failed to transfer employee")
	}

	// Create lifecycle event
	event := &EmployeeLifecycleEvent{
		ID:            utils.GenerateLifecycleID(),
		EmployeeID:    empID,
		EventType:     "transfer",
		OldValue:      oldDeptID,
		NewValue:      newDeptID,
		EffectiveDate: time.Now(),
		Reason:        "Transfer to department " + newDeptID,
		CreatedAt:     time.Now(),
	}

	if err := s.db.Create(event).Error; err != nil {
		logger.Error("failed to create lifecycle event", "error", err)
		// Don't return error here as the transfer was successful
	}

	return nil
}

// TerminateEmployee terminates an employee
func (s *LifecycleService) TerminateEmployee(ctx context.Context, empID string, terminationDate time.Time, reason string) error {
	// Find employee by ID
	var employee domain.Employee
	if err := s.db.Where("id = ?", empID).First(&employee).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("employee not found")
		}
		logger.Error("failed to find employee", "error", err)
		return errors.New("failed to terminate employee")
	}

	// Update employee status
	employee.Status = "terminated"
	employee.UpdatedAt = time.Now()

	// Save updated employee to database
	if err := s.db.Save(&employee).Error; err != nil {
		logger.Error("failed to update employee status", "error", err)
		return errors.New("failed to terminate employee")
	}

	// Create lifecycle event
	event := &EmployeeLifecycleEvent{
		ID:            utils.GenerateLifecycleID(),
		EmployeeID:    empID,
		EventType:     "termination",
		OldValue:      employee.Status,
		NewValue:      "terminated",
		EffectiveDate: terminationDate,
		Reason:        reason,
		CreatedAt:     time.Now(),
	}

	if err := s.db.Create(event).Error; err != nil {
		logger.Error("failed to create lifecycle event", "error", err)
		// Don't return error here as the termination was successful
	}

	return nil
}

// GetEmployeeLifecycleHistory retrieves an employee's lifecycle history
func (s *LifecycleService) GetEmployeeLifecycleHistory(ctx context.Context, empID string) ([]*EmployeeLifecycleEvent, error) {
	// Check if employee exists
	var employee domain.Employee
	if err := s.db.Where("id = ?", empID).First(&employee).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("employee not found")
		}
		logger.Error("failed to find employee", "error", err)
		return nil, errors.New("failed to get employee lifecycle history")
	}

	// Get lifecycle events
	var events []*EmployeeLifecycleEvent
	if err := s.db.Where("employee_id = ?", empID).Order("created_at desc").Find(&events).Error; err != nil {
		logger.Error("failed to find lifecycle events", "error", err)
		return nil, errors.New("failed to get employee lifecycle history")
	}

	return events, nil
}

