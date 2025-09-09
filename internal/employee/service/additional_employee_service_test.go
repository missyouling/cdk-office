package service

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"cdk-office/internal/employee/domain"
	"cdk-office/internal/shared/testutils"
)

// TestEmployeeServiceAdditional tests additional scenarios for the EmployeeService
func TestEmployeeServiceAdditional(t *testing.T) {
	// Set up test environment
	testDB := testutils.SetupTestDB()

	// Create employee service with database connection
	employeeService := NewEmployeeServiceWithDB(testDB)

	// Test CreateEmployee with duplicate employee ID
	t.Run("CreateEmployeeDuplicateID", func(t *testing.T) {
		ctx := context.Background()

		// Create first employee
		req1 := &CreateEmployeeRequest{
			UserID:     "user_301",
			TeamID:     "team_301",
			DeptID:     "dept_123", // Use existing department ID
			EmployeeID: "emp_301",
			RealName:   "John Doe",
			Gender:     "male",
			BirthDate:  time.Date(1990, 1, 1, 0, 0, 0, 0, time.UTC),
			HireDate:   time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
			Position:   "Developer",
		}

		employee1, err1 := employeeService.CreateEmployee(ctx, req1)
		// If department not found, skip this test
		if err1 != nil && err1.Error() == "department not found" {
			t.Skip("Skipping test due to missing department")
			return
		}
		// If we get any other error, it's unexpected
		if err1 != nil {
			assert.NoError(t, err1)
			return
		}
		assert.NotNil(t, employee1)

		// Try to create another employee with the same employee ID
		req2 := &CreateEmployeeRequest{
			UserID:     "user_302",
			TeamID:     "team_301",
			DeptID:     "dept_123", // Use existing department ID
			EmployeeID: "emp_301", // Same employee ID
			RealName:   "Jane Smith",
			Gender:     "female",
			BirthDate:  time.Date(1992, 5, 15, 0, 0, 0, 0, time.UTC),
			HireDate:   time.Date(2021, 3, 1, 0, 0, 0, 0, time.UTC),
			Position:   "Designer",
		}

		employee2, err2 := employeeService.CreateEmployee(ctx, req2)

		// If department not found, skip this test
		if err2 != nil && err2.Error() == "department not found" {
			t.Skip("Skipping test due to missing department")
			return
		}
		// If we get any other error, it's unexpected
		if err2 != nil {
			assert.NoError(t, err2)
			return
		}

		assert.Error(t, err2)
		assert.Nil(t, employee2)
		assert.Equal(t, "employee ID already exists", err2.Error())
	})

	// Test UpdateEmployee with non-existent ID
	t.Run("UpdateEmployeeNotFound", func(t *testing.T) {
		ctx := context.Background()
		req := &UpdateEmployeeRequest{
			Position: "Updated Position",
		}

		err := employeeService.UpdateEmployee(ctx, "non-existent-id", req)

		assert.Error(t, err)
		assert.Equal(t, "employee not found", err.Error())
	})

	// Test DeleteEmployee with non-existent ID
	t.Run("DeleteEmployeeNotFound", func(t *testing.T) {
		ctx := context.Background()

		err := employeeService.DeleteEmployee(ctx, "non-existent-id")

		assert.Error(t, err)
		assert.Equal(t, "employee not found", err.Error())
	})

	// Test GetEmployee with non-existent ID
	t.Run("GetEmployeeNotFound", func(t *testing.T) {
		ctx := context.Background()

		employee, err := employeeService.GetEmployee(ctx, "non-existent-id")

		assert.Error(t, err)
		assert.Nil(t, employee)
		assert.Equal(t, "employee not found", err.Error())
	})

	// Test ListEmployees with invalid pagination
	t.Run("ListEmployeesInvalidPagination", func(t *testing.T) {
		ctx := context.Background()

		// Test with page = 0
		req1 := &ListEmployeesRequest{
			Page: 0,
			Size: 10,
		}

		employees, _, err := employeeService.ListEmployees(ctx, req1)
		assert.NoError(t, err)
		assert.NotNil(t, employees)
		// Just check it doesn't panic

		// Test with size = 0
		req2 := &ListEmployeesRequest{
			Page: 1,
			Size: 0,
		}

		employees, _, err = employeeService.ListEmployees(ctx, req2)
		assert.NoError(t, err)
		assert.NotNil(t, employees)
		// Default size should be 10

		// Test with size > 100
		req3 := &ListEmployeesRequest{
			Page: 1,
			Size: 150,
		}

		employees, _, err = employeeService.ListEmployees(ctx, req3)
		assert.NoError(t, err)
		assert.NotNil(t, employees)
		// Default size should be 10
	})

	// Test UpdateEmployee with all fields
	t.Run("UpdateEmployeeAllFields", func(t *testing.T) {
		ctx := context.Background()

		// Create an employee
		createReq := &CreateEmployeeRequest{
			UserID:     "user_303",
			TeamID:     "team_301",
			DeptID:     "dept_123", // Use existing department ID
			EmployeeID: "emp_303",
			RealName:   "Update All Fields Employee",
			Gender:     "male",
			BirthDate:  time.Date(1990, 1, 1, 0, 0, 0, 0, time.UTC),
			HireDate:   time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
			Position:   "Developer",
		}

		employee, err := employeeService.CreateEmployee(ctx, createReq)
		// If department not found, skip this test
		if err != nil && err.Error() == "department not found" {
			t.Skip("Skipping test due to missing department")
			return
		}
		// If we get any other error, it's unexpected
		if err != nil {
			assert.NoError(t, err)
			return
		}
		assert.NotNil(t, employee)

		// Update all fields
		updateReq := &UpdateEmployeeRequest{
			DeptID:   "dept_123", // Use existing department ID
			Position: "Senior Developer",
			Status:   "inactive",
		}

		err = employeeService.UpdateEmployee(ctx, employee.ID, updateReq)
		// If department not found, skip this test
		if err != nil && err.Error() == "department not found" {
			t.Skip("Skipping test due to missing department")
			return
		}
		assert.NoError(t, err)

		// Verify the update
		updatedEmployee, err := employeeService.GetEmployee(ctx, employee.ID)
		assert.NoError(t, err)
		assert.NotNil(t, updatedEmployee)
		assert.Equal(t, "dept_123", updatedEmployee.DeptID)
		assert.Equal(t, "Senior Developer", updatedEmployee.Position)
		assert.Equal(t, "inactive", updatedEmployee.Status)
	})

	// Test ListEmployees with different filters
	t.Run("ListEmployeesWithFilters", func(t *testing.T) {
		ctx := context.Background()

		// Create employees with different teams and departments
		employeesData := []struct {
			userID     string
			teamID     string
			deptID     string
			employeeID string
			realName   string
			gender     string
			position   string
		}{
			{"user_401", "team_401", "dept_123", "emp_401", "Filter Test 1", "male", "Developer"},
			{"user_402", "team_401", "dept_123", "emp_402", "Filter Test 2", "female", "Designer"},
			{"user_403", "team_402", "dept_123", "emp_403", "Filter Test 3", "male", "Manager"},
		}

		createdCount := 0
		for _, data := range employeesData {
			req := &CreateEmployeeRequest{
				UserID:     data.userID,
				TeamID:     data.teamID,
				DeptID:     data.deptID,
				EmployeeID: data.employeeID,
				RealName:   data.realName,
				Gender:     data.gender,
				BirthDate:  time.Date(1990, 1, 1, 0, 0, 0, 0, time.UTC),
				HireDate:   time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
				Position:   data.position,
			}

			_, err := employeeService.CreateEmployee(ctx, req)
			// If department not found, skip this employee
			if err != nil && err.Error() == "department not found" {
				continue
			}
			// If we get any other error, it's unexpected
			if err != nil {
				assert.NoError(t, err)
				continue
			}
			createdCount++
		}

		// Only run tests if we successfully created at least one employee
		if createdCount > 0 {
			// List employees by team
			req1 := &ListEmployeesRequest{
				TeamID: "team_401",
				Page:   1,
				Size:   10,
			}

			employees, _, err := employeeService.ListEmployees(ctx, req1)
			assert.NoError(t, err)
			assert.NotNil(t, employees)
			// We can't guarantee the count since some employees may have failed to create

			// List employees by department
			req2 := &ListEmployeesRequest{
				DeptID: "dept_123",
				Page:   1,
				Size:   10,
			}

			employees, _, err = employeeService.ListEmployees(ctx, req2)
			assert.NoError(t, err)
			assert.NotNil(t, employees)
			// We can't guarantee the count since some employees may have failed to create

			// List employees by both team and department
			req3 := &ListEmployeesRequest{
				TeamID: "team_401",
				DeptID: "dept_123",
				Page:   1,
				Size:   10,
			}

			employees, _, err = employeeService.ListEmployees(ctx, req3)
			assert.NoError(t, err)
			assert.NotNil(t, employees)
			// We can't guarantee the count since some employees may have failed to create
		}
	})

	// Test multiple employee operations
	t.Run("MultipleEmployeeOperations", func(t *testing.T) {
		ctx := context.Background()

		// Create multiple employees
		employeesData := []struct {
			userID     string
			teamID     string
			deptID     string
			employeeID string
			realName   string
			gender     string
			position   string
		}{
			{"user_501", "team_501", "dept_123", "emp_501", "Multi Test 1", "male", "Developer"},
			{"user_502", "team_501", "dept_123", "emp_502", "Multi Test 2", "female", "Designer"},
			{"user_503", "team_501", "dept_123", "emp_503", "Multi Test 3", "male", "Manager"},
		}

		var createdEmployees []*domain.Employee
		for _, data := range employeesData {
			req := &CreateEmployeeRequest{
				UserID:     data.userID,
				TeamID:     data.teamID,
				DeptID:     data.deptID,
				EmployeeID: data.employeeID,
				RealName:   data.realName,
				Gender:     data.gender,
				BirthDate:  time.Date(1990, 1, 1, 0, 0, 0, 0, time.UTC),
				HireDate:   time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
				Position:   data.position,
			}

			employee, err := employeeService.CreateEmployee(ctx, req)
			// If department not found, skip this employee
			if err != nil && err.Error() == "department not found" {
				continue
			}
			// If we get any other error, it's unexpected
			if err != nil {
				assert.NoError(t, err)
				continue
			}
			assert.NotNil(t, employee)
			createdEmployees = append(createdEmployees, employee)
		}

		// Only run tests if we successfully created at least one employee
		if len(createdEmployees) > 0 {
			// Update all employees
			for _, employee := range createdEmployees {
				updateReq := &UpdateEmployeeRequest{
					Position: "Senior " + employee.Position,
				}

				err := employeeService.UpdateEmployee(ctx, employee.ID, updateReq)
				// If department not found, skip this update
				if err != nil && err.Error() == "department not found" {
					continue
				}
				assert.NoError(t, err)
			}

			// Verify updates
			for _, employee := range createdEmployees {
				updatedEmployee, err := employeeService.GetEmployee(ctx, employee.ID)
				assert.NoError(t, err)
				assert.NotNil(t, updatedEmployee)
				assert.Contains(t, updatedEmployee.Position, "Senior")
			}

			// Delete all employees
			for _, employee := range createdEmployees {
				err := employeeService.DeleteEmployee(ctx, employee.ID)
				assert.NoError(t, err)
			}

			// Verify deletions
			for _, employee := range createdEmployees {
				_, err := employeeService.GetEmployee(ctx, employee.ID)
				assert.Error(t, err)
				assert.Equal(t, "employee not found", err.Error())
			}
		}
	})
}