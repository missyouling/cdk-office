package test

import (
	"context"
	"testing"
	"time"

	"cdk-office/internal/employee/domain"
	"cdk-office/internal/employee/service"
	"cdk-office/internal/shared/testutils"
	"github.com/stretchr/testify/assert"
)

func TestEmployeeService_CreateEmployee(t *testing.T) {
	db := testutils.SetupTestDB()
	defer db.Migrator().DropTable(&domain.Employee{}, &domain.Department{})

	employeeService := service.NewEmployeeServiceWithDB(db)

	// Create a test department first
	dept := &domain.Department{
		ID:   "dept-001",
		Name: "技术部",
	}
	err := db.Create(dept).Error
	assert.NoError(t, err)

	tests := []struct {
		name          string
		request       *service.CreateEmployeeRequest
		expectError   bool
		errorContains string
	}{
		{
			name: "successful employee creation",
			request: &service.CreateEmployeeRequest{
				UserID:     "user-001",
				TeamID:     "team-001",
				DeptID:     "dept-001",
				EmployeeID: "emp-001",
				RealName:   "张三",
				Gender:     "男",
				BirthDate:  time.Now().AddDate(-30, 0, 0),
				HireDate:   time.Now(),
				Position:   "软件工程师",
			},
			expectError: false,
		},
		{
			name: "duplicate employee ID",
			request: &service.CreateEmployeeRequest{
				UserID:     "user-003",
				TeamID:     "team-003",
				DeptID:     "dept-001",
				EmployeeID: "emp-001", // Same as first test
				RealName:   "李四",
				Gender:     "男",
				BirthDate:  time.Now().AddDate(-28, 0, 0),
				HireDate:   time.Now(),
				Position:   "设计师",
			},
			expectError:   true,
			errorContains: "already exists",
		},
		{
			name: "non-existent department",
			request: &service.CreateEmployeeRequest{
				UserID:     "user-004",
				TeamID:     "team-004",
				DeptID:     "dept-999", // Non-existent department
				EmployeeID: "emp-004",
				RealName:   "王五",
				Gender:     "女",
				BirthDate:  time.Now().AddDate(-26, 0, 0),
				HireDate:   time.Now(),
				Position:   "测试工程师",
			},
			expectError:   true,
			errorContains: "department not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			employee, err := employeeService.CreateEmployee(context.Background(), tt.request)

			if tt.expectError {
				assert.Error(t, err)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
			} else {
				assert.NoError(t, err)
				assert.NotEmpty(t, employee.ID)
				assert.Equal(t, tt.request.UserID, employee.UserID)
				assert.Equal(t, tt.request.TeamID, employee.TeamID)
				assert.Equal(t, tt.request.DeptID, employee.DeptID)
				assert.Equal(t, tt.request.EmployeeID, employee.EmployeeID)
				assert.Equal(t, tt.request.RealName, employee.RealName)
				assert.Equal(t, tt.request.Gender, employee.Gender)
				assert.Equal(t, tt.request.Position, employee.Position)
				assert.Equal(t, "active", employee.Status)
				assert.NotZero(t, employee.CreatedAt)
				assert.NotZero(t, employee.UpdatedAt)
			}
		})
	}
}

func TestEmployeeService_GetEmployee(t *testing.T) {
	db := testutils.SetupTestDB()
	defer db.Migrator().DropTable(&domain.Employee{}, &domain.Department{})

	employeeService := service.NewEmployeeServiceWithDB(db)

	// Create a test department
	dept := &domain.Department{
		ID:   "dept-001",
		Name: "技术部",
	}
	err := db.Create(dept).Error
	assert.NoError(t, err)

	// Create test employee
	request := &service.CreateEmployeeRequest{
		UserID:     "user-001",
		TeamID:     "team-001",
		DeptID:     "dept-001",
		EmployeeID: "emp-001",
		RealName:   "测试员工",
		Gender:     "男",
		BirthDate:  time.Now().AddDate(-30, 0, 0),
		HireDate:   time.Now(),
		Position:   "测试工程师",
	}
	createdEmployee, err := employeeService.CreateEmployee(context.Background(), request)
	assert.NoError(t, err)

	tests := []struct {
		name          string
		employeeID    string
		expectError   bool
		errorContains string
	}{
		{
			name:        "get existing employee",
			employeeID:  createdEmployee.ID,
			expectError: false,
		},
		{
			name:          "get non-existent employee",
			employeeID:    "non-existent-id",
			expectError:   true,
			errorContains: "employee not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			employee, err := employeeService.GetEmployee(context.Background(), tt.employeeID)

			if tt.expectError {
				assert.Error(t, err)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
			} else {
				assert.NoError(t, err)
				assert.Equal(t, createdEmployee.ID, employee.ID)
				assert.Equal(t, createdEmployee.UserID, employee.UserID)
				assert.Equal(t, createdEmployee.TeamID, employee.TeamID)
				assert.Equal(t, createdEmployee.DeptID, employee.DeptID)
				assert.Equal(t, createdEmployee.EmployeeID, employee.EmployeeID)
				assert.Equal(t, createdEmployee.RealName, employee.RealName)
				assert.Equal(t, createdEmployee.Gender, employee.Gender)
				assert.Equal(t, createdEmployee.Position, employee.Position)
				assert.Equal(t, createdEmployee.Status, employee.Status)
			}
		})
	}
}

func TestEmployeeService_UpdateEmployee(t *testing.T) {
	db := testutils.SetupTestDB()
	defer db.Migrator().DropTable(&domain.Employee{}, &domain.Department{})

	employeeService := service.NewEmployeeServiceWithDB(db)

	// Create test departments
	dept1 := &domain.Department{
		ID:   "dept-001",
		Name: "技术部",
	}
	err := db.Create(dept1).Error
	assert.NoError(t, err)

	dept2 := &domain.Department{
		ID:   "dept-002",
		Name: "产品部",
	}
	err = db.Create(dept2).Error
	assert.NoError(t, err)

	// Create test employee
	request := &service.CreateEmployeeRequest{
		UserID:     "user-001",
		TeamID:     "team-001",
		DeptID:     "dept-001",
		EmployeeID: "emp-001",
		RealName:   "原始姓名",
		Gender:     "男",
		BirthDate:  time.Now().AddDate(-30, 0, 0),
		HireDate:   time.Now(),
		Position:   "原始职位",
	}
	createdEmployee, err := employeeService.CreateEmployee(context.Background(), request)
	assert.NoError(t, err)

	tests := []struct {
		name          string
		employeeID    string
		request       *service.UpdateEmployeeRequest
		expectError   bool
		errorContains string
	}{
		{
			name:       "successful update with department change",
			employeeID: createdEmployee.ID,
			request: &service.UpdateEmployeeRequest{
				DeptID:   "dept-002",
				Position: "更新职位",
				Status:   "active",
			},
			expectError: false,
		},
		{
			name:       "successful update with position only",
			employeeID: createdEmployee.ID,
			request: &service.UpdateEmployeeRequest{
				Position: "新职位",
			},
			expectError: false,
		},
		{
			name:          "update non-existent employee",
			employeeID:    "non-existent-id",
			request:       &service.UpdateEmployeeRequest{Position: "新职位"},
			expectError:   true,
			errorContains: "employee not found",
		},
		{
			name:       "update with non-existent department",
			employeeID: createdEmployee.ID,
			request: &service.UpdateEmployeeRequest{
				DeptID: "dept-999",
			},
			expectError:   true,
			errorContains: "department not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := employeeService.UpdateEmployee(context.Background(), tt.employeeID, tt.request)

			if tt.expectError {
				assert.Error(t, err)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
			} else {
				assert.NoError(t, err)
				
				// Verify the update by getting the employee
				updatedEmployee, err := employeeService.GetEmployee(context.Background(), tt.employeeID)
				assert.NoError(t, err)
				
				// Check updated fields
				if tt.request.DeptID != "" {
					assert.Equal(t, tt.request.DeptID, updatedEmployee.DeptID)
				}
				if tt.request.Position != "" {
					assert.Equal(t, tt.request.Position, updatedEmployee.Position)
				}
				if tt.request.Status != "" {
					assert.Equal(t, tt.request.Status, updatedEmployee.Status)
				}
				
				// UpdatedAt should be changed
				assert.True(t, updatedEmployee.UpdatedAt.After(createdEmployee.UpdatedAt))
			}
		})
	}
}

func TestEmployeeService_DeleteEmployee(t *testing.T) {
	db := testutils.SetupTestDB()
	defer db.Migrator().DropTable(&domain.Employee{}, &domain.Department{})

	employeeService := service.NewEmployeeServiceWithDB(db)

	// Create a test department
	dept := &domain.Department{
		ID:   "dept-001",
		Name: "技术部",
	}
	err := db.Create(dept).Error
	assert.NoError(t, err)

	// Create test employee
	request := &service.CreateEmployeeRequest{
		UserID:     "user-001",
		TeamID:     "team-001",
		DeptID:     "dept-001",
		EmployeeID: "emp-001",
		RealName:   "待删除员工",
		Gender:     "男",
		BirthDate:  time.Now().AddDate(-30, 0, 0),
		HireDate:   time.Now(),
		Position:   "临时职位",
	}
	createdEmployee, err := employeeService.CreateEmployee(context.Background(), request)
	assert.NoError(t, err)

	tests := []struct {
		name          string
		employeeID    string
		expectError   bool
		errorContains string
	}{
		{
			name:        "delete existing employee",
			employeeID:  createdEmployee.ID,
			expectError: false,
		},
		{
			name:          "delete non-existent employee",
			employeeID:    "non-existent-id",
			expectError:   true,
			errorContains: "employee not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := employeeService.DeleteEmployee(context.Background(), tt.employeeID)

			if tt.expectError {
				assert.Error(t, err)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
			} else {
				assert.NoError(t, err)
				
				// Verify employee is deleted
				_, err := employeeService.GetEmployee(context.Background(), tt.employeeID)
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "employee not found")
			}
		})
	}
}

func TestEmployeeService_ListEmployees(t *testing.T) {
	db := testutils.SetupTestDB()
	defer db.Migrator().DropTable(&domain.Employee{}, &domain.Department{})

	employeeService := service.NewEmployeeServiceWithDB(db)

	// Create a test department
	dept := &domain.Department{
		ID:   "dept-001",
		Name: "技术部",
	}
	err := db.Create(dept).Error
	assert.NoError(t, err)

	// Create test employees
	employees := []*service.CreateEmployeeRequest{
		{
			UserID:     "user-001",
			TeamID:     "team-001",
			DeptID:     "dept-001",
			EmployeeID: "emp-001",
			RealName:   "员工1",
			Gender:     "男",
			BirthDate:  time.Now().AddDate(-30, 0, 0),
			HireDate:   time.Now(),
			Position:   "工程师",
		},
		{
			UserID:     "user-002",
			TeamID:     "team-001",
			DeptID:     "dept-001",
			EmployeeID: "emp-002",
			RealName:   "员工2",
			Gender:     "女",
			BirthDate:  time.Now().AddDate(-25, 0, 0),
			HireDate:   time.Now(),
			Position:   "设计师",
		},
		{
			UserID:     "user-003",
			TeamID:     "team-002",
			DeptID:     "dept-001",
			EmployeeID: "emp-003",
			RealName:   "员工3",
			Gender:     "男",
			BirthDate:  time.Now().AddDate(-28, 0, 0),
			HireDate:   time.Now(),
			Position:   "产品经理",
		},
	}

	for _, req := range employees {
		_, err := employeeService.CreateEmployee(context.Background(), req)
		assert.NoError(t, err)
	}

	tests := []struct {
		name          string
		request       *service.ListEmployeesRequest
		expectedCount int
		totalCount    int64
		expectError   bool
	}{
		{
			name: "list first page",
			request: &service.ListEmployeesRequest{
				Page: 1,
				Size: 2,
			},
			expectedCount: 2,
			totalCount:    3,
			expectError:   false,
		},
		{
			name: "list second page",
			request: &service.ListEmployeesRequest{
				Page: 2,
				Size: 2,
			},
			expectedCount: 1,
			totalCount:    3,
			expectError:   false,
		},
		{
			name: "list with large page size",
			request: &service.ListEmployeesRequest{
				Page: 1,
				Size: 10,
			},
			expectedCount: 3,
			totalCount:    3,
			expectError:   false,
		},
		{
			name: "list with zero page",
			request: &service.ListEmployeesRequest{
				Page: 0,
				Size: 10,
			},
			expectedCount: 3,
			totalCount:    3,
			expectError:   false,
		},
		{
			name: "list with team filter",
			request: &service.ListEmployeesRequest{
				TeamID: "team-001",
				Page:   1,
				Size:   10,
			},
			expectedCount: 2,
			totalCount:    2, // Only 2 employees in team-001
			expectError:   false,
		},
		{
			name: "list with department filter",
			request: &service.ListEmployeesRequest{
				DeptID: "dept-001",
				Page:   1,
				Size:   10,
			},
			expectedCount: 3,
			totalCount:    3, // All 3 employees in dept-001
			expectError:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, total, err := employeeService.ListEmployees(context.Background(), tt.request)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedCount, len(result))
				assert.Equal(t, tt.totalCount, total)
			}
		})
	}
}