package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"cdk-office/internal/employee/domain"
	"cdk-office/internal/employee/service"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// mockEmployeeService is a mock implementation of the EmployeeServiceInterface
type mockEmployeeService struct {
	employees map[string]*domain.Employee
	nextID    int
}

func newMockEmployeeService() *mockEmployeeService {
	return &mockEmployeeService{
		employees: make(map[string]*domain.Employee),
		nextID:    1,
	}
}

func (m *mockEmployeeService) CreateEmployee(ctx context.Context, req *service.CreateEmployeeRequest) (*domain.Employee, error) {
	// Check if employee ID already exists
	for _, emp := range m.employees {
		if emp.EmployeeID == req.EmployeeID {
			return nil, errors.New("employee ID already exists")
		}
	}

	// Create new employee
	emp := &domain.Employee{
		ID:         "emp_" + string(rune(m.nextID+'0')),
		UserID:     req.UserID,
		TeamID:     req.TeamID,
		DeptID:     req.DeptID,
		EmployeeID: req.EmployeeID,
		RealName:   req.RealName,
		Gender:     req.Gender,
		BirthDate:  req.BirthDate,
		HireDate:   req.HireDate,
		Position:   req.Position,
		Status:     "active",
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	m.employees[emp.ID] = emp
	m.nextID++
	return emp, nil
}

func (m *mockEmployeeService) UpdateEmployee(ctx context.Context, empID string, req *service.UpdateEmployeeRequest) error {
	emp, exists := m.employees[empID]
	if !exists {
		return errors.New("employee not found")
	}

	if req.DeptID != "" {
		emp.DeptID = req.DeptID
	}
	if req.Position != "" {
		emp.Position = req.Position
	}
	if req.Status != "" {
		emp.Status = req.Status
	}

	emp.UpdatedAt = time.Now()
	return nil
}

func (m *mockEmployeeService) DeleteEmployee(ctx context.Context, empID string) error {
	_, exists := m.employees[empID]
	if !exists {
		return errors.New("employee not found")
	}

	delete(m.employees, empID)
	return nil
}

func (m *mockEmployeeService) ListEmployees(ctx context.Context, req *service.ListEmployeesRequest) ([]*domain.Employee, int64, error) {
	var employees []*domain.Employee
	for _, emp := range m.employees {
		if req.TeamID != "" && emp.TeamID != req.TeamID {
			continue
		}
		if req.DeptID != "" && emp.DeptID != req.DeptID {
			continue
		}
		employees = append(employees, emp)
	}

	// Simple pagination
	start := (req.Page - 1) * req.Size
	end := start + req.Size
	if start >= len(employees) {
		return []*domain.Employee{}, int64(len(employees)), nil
	}
	if end > len(employees) {
		end = len(employees)
	}

	return employees[start:end], int64(len(employees)), nil
}

func (m *mockEmployeeService) GetEmployee(ctx context.Context, empID string) (*domain.Employee, error) {
	emp, exists := m.employees[empID]
	if !exists {
		return nil, errors.New("employee not found")
	}
	return emp, nil
}

// TestNewEmployeeHandler tests the NewEmployeeHandler function
func TestNewEmployeeHandler(t *testing.T) {
	handler := NewEmployeeHandler()
	assert.NotNil(t, handler)
	// Note: We can't directly check the service field as it's not exported
}

// TestNewEmployeeHandlerWithService tests the NewEmployeeHandlerWithService function
func TestNewEmployeeHandlerWithService(t *testing.T) {
	mockService := newMockEmployeeService()
	handler := NewEmployeeHandlerWithService(mockService)
	assert.NotNil(t, handler)
	assert.Equal(t, mockService, handler.employeeService)
}

// TestEmployeeHandler tests the EmployeeHandler
func TestEmployeeHandler(t *testing.T) {
	// Set up test environment
	gin.SetMode(gin.TestMode)

	// Create mock service
	mockService := newMockEmployeeService()

	// Create employee handler with mock service
	empHandler := NewEmployeeHandlerWithService(mockService)

	// Test CreateEmployee
	t.Run("CreateEmployee", func(t *testing.T) {
		// Create test request
		reqBody := CreateEmployeeRequest{
			UserID:     "user_123",
			TeamID:     "team_123",
			DeptID:     "dept_123",
			EmployeeID: "emp001",
			RealName:   "张三",
			Gender:     "男",
			BirthDate:  "1990-01-01",
			HireDate:   "2020-01-01",
			Position:   "软件工程师",
		}
		jsonReq, _ := json.Marshal(reqBody)

		// Create HTTP request and response recorder
		req, _ := http.NewRequest("POST", "/employees", bytes.NewBuffer(jsonReq))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		// Create gin context and call handler
		c, _ := gin.CreateTestContext(w)
		c.Request = req
		empHandler.CreateEmployee(c)

		// Assert response
		assert.Equal(t, http.StatusOK, w.Code)
	})

	// Test CreateEmployee with employee ID already exists
	t.Run("CreateEmployeeIDAlreadyExists", func(t *testing.T) {
		// First create an employee
		createReq := &service.CreateEmployeeRequest{
			UserID:     "user_123",
			TeamID:     "team_123",
			DeptID:     "dept_123",
			EmployeeID: "emp002",
			RealName:   "李四",
			Gender:     "女",
			BirthDate:  time.Now().AddDate(-30, 0, 0),
			HireDate:   time.Now().AddDate(-2, 0, 0),
			Position:   "产品经理",
		}
		_, _ = mockService.CreateEmployee(context.Background(), createReq)

		// Try to create another employee with the same ID
		reqBody := CreateEmployeeRequest{
			UserID:     "user_456",
			TeamID:     "team_456",
			DeptID:     "dept_456",
			EmployeeID: "emp002", // Same ID
			RealName:   "王五",
			Gender:     "男",
			BirthDate:  "1990-01-01",
			HireDate:   "2020-01-01",
			Position:   "设计师",
		}
		jsonReq, _ := json.Marshal(reqBody)

		// Create HTTP request and response recorder
		req, _ := http.NewRequest("POST", "/employees", bytes.NewBuffer(jsonReq))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		// Create gin context and call handler
		c, _ := gin.CreateTestContext(w)
		c.Request = req
		empHandler.CreateEmployee(c)

		// Assert response
		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "employee ID already exists")
	})

	// Test CreateEmployee with department not found
	t.Run("CreateEmployeeDepartmentNotFound", func(t *testing.T) {
		// Create a mock service that returns "department not found" error
		mockServiceWithDeptError := &mockEmployeeServiceWithError{
			createEmployeeError: errors.New("department not found"),
		}

		handler := NewEmployeeHandlerWithService(mockServiceWithDeptError)

		// Create test request
		reqBody := CreateEmployeeRequest{
			UserID:     "user_123",
			TeamID:     "team_123",
			DeptID:     "dept_123",
			EmployeeID: "emp003",
			RealName:   "赵六",
			Gender:     "女",
			BirthDate:  "1990-01-01",
			HireDate:   "2020-01-01",
			Position:   "测试工程师",
		}
		jsonReq, _ := json.Marshal(reqBody)

		// Create HTTP request and response recorder
		req, _ := http.NewRequest("POST", "/employees", bytes.NewBuffer(jsonReq))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		// Create gin context and call handler
		c, _ := gin.CreateTestContext(w)
		c.Request = req
		handler.CreateEmployee(c)

		// Assert response
		assert.Equal(t, http.StatusNotFound, w.Code)
		assert.Contains(t, w.Body.String(), "department not found")
	})

	// Test CreateEmployee with invalid JSON
	t.Run("CreateEmployeeInvalidJSON", func(t *testing.T) {
		// Create HTTP request with invalid JSON
		req, _ := http.NewRequest("POST", "/employees", bytes.NewBuffer([]byte("{invalid json}")))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		// Create gin context and call handler
		c, _ := gin.CreateTestContext(w)
		c.Request = req
		empHandler.CreateEmployee(c)

		// Assert response
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	// Test CreateEmployee with invalid date format
	t.Run("CreateEmployeeInvalidDateFormat", func(t *testing.T) {
		// Create test request with invalid date format
		reqBody := CreateEmployeeRequest{
			UserID:     "user_123",
			TeamID:     "team_123",
			DeptID:     "dept_123",
			EmployeeID: "emp004",
			RealName:   "孙七",
			Gender:     "男",
			BirthDate:  "invalid-date",
			HireDate:   "2020-01-01",
			Position:   "运维工程师",
		}
		jsonReq, _ := json.Marshal(reqBody)

		// Create HTTP request and response recorder
		req, _ := http.NewRequest("POST", "/employees", bytes.NewBuffer(jsonReq))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		// Create gin context and call handler
		c, _ := gin.CreateTestContext(w)
		c.Request = req
		empHandler.CreateEmployee(c)

		// Assert response
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	// Test CreateEmployee with invalid hire date format
	t.Run("CreateEmployeeInvalidHireDateFormat", func(t *testing.T) {
		// Create test request with invalid hire date format
		reqBody := CreateEmployeeRequest{
			UserID:     "user_123",
			TeamID:     "team_123",
			DeptID:     "dept_123",
			EmployeeID: "emp005",
			RealName:   "周八",
			Gender:     "女",
			BirthDate:  "1990-01-01",
			HireDate:   "invalid-date",
			Position:   "人事专员",
		}
		jsonReq, _ := json.Marshal(reqBody)

		// Create HTTP request and response recorder
		req, _ := http.NewRequest("POST", "/employees", bytes.NewBuffer(jsonReq))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		// Create gin context and call handler
		c, _ := gin.CreateTestContext(w)
		c.Request = req
		empHandler.CreateEmployee(c)

		// Assert response
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	// Test CreateEmployee with service error
	t.Run("CreateEmployeeServiceError", func(t *testing.T) {
		// Create a mock service that returns an error
		mockServiceWithError := &mockEmployeeServiceWithError{
			createEmployeeError: errors.New("internal server error"),
		}

		handler := NewEmployeeHandlerWithService(mockServiceWithError)

		// Create test request
		reqBody := CreateEmployeeRequest{
			UserID:     "user_123",
			TeamID:     "team_123",
			DeptID:     "dept_123",
			EmployeeID: "emp006",
			RealName:   "吴九",
			Gender:     "男",
			BirthDate:  "1990-01-01",
			HireDate:   "2020-01-01",
			Position:   "财务专员",
		}
		jsonReq, _ := json.Marshal(reqBody)

		// Create HTTP request and response recorder
		req, _ := http.NewRequest("POST", "/employees", bytes.NewBuffer(jsonReq))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		// Create gin context and call handler
		c, _ := gin.CreateTestContext(w)
		c.Request = req
		handler.CreateEmployee(c)

		// Assert response
		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Contains(t, w.Body.String(), "internal server error")
	})

	// Test UpdateEmployee
	t.Run("UpdateEmployee", func(t *testing.T) {
		// First create an employee
		createReq := &service.CreateEmployeeRequest{
			UserID:     "user_123",
			TeamID:     "team_123",
			DeptID:     "dept_123",
			EmployeeID: "emp007",
			RealName:   "郑十",
			Gender:     "女",
			BirthDate:  time.Now().AddDate(-30, 0, 0),
			HireDate:   time.Now().AddDate(-2, 0, 0),
			Position:   "软件工程师",
		}
		emp, _ := mockService.CreateEmployee(context.Background(), createReq)

		// Create update request
		updateReqBody := UpdateEmployeeRequest{
			Position: "高级软件工程师",
		}
		jsonReq, _ := json.Marshal(updateReqBody)

		// Create HTTP request and response recorder
		req, _ := http.NewRequest("PUT", "/employees/"+emp.ID, bytes.NewBuffer(jsonReq))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		// Create gin context and call handler
		c, _ := gin.CreateTestContext(w)
		c.Request = req
		c.AddParam("id", emp.ID)
		empHandler.UpdateEmployee(c)

		// Assert response
		assert.Equal(t, http.StatusOK, w.Code)
	})

	// Test UpdateEmployee with non-existent ID
	t.Run("UpdateEmployeeNotFound", func(t *testing.T) {
		// Create update request
		updateReqBody := UpdateEmployeeRequest{
			Position: "高级软件工程师",
		}
		jsonReq, _ := json.Marshal(updateReqBody)

		// Create HTTP request and response recorder
		req, _ := http.NewRequest("PUT", "/employees/non-existent", bytes.NewBuffer(jsonReq))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		// Create gin context and call handler
		c, _ := gin.CreateTestContext(w)
		c.Request = req
		c.AddParam("id", "non-existent")
		empHandler.UpdateEmployee(c)

		// Assert response
		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	// Test UpdateEmployee with invalid JSON
	t.Run("UpdateEmployeeInvalidJSON", func(t *testing.T) {
		// Create HTTP request with invalid JSON
		req, _ := http.NewRequest("PUT", "/employees/emp_1", bytes.NewBuffer([]byte("{invalid json}")))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		// Create gin context and call handler
		c, _ := gin.CreateTestContext(w)
		c.Request = req
		c.AddParam("id", "emp_1")
		empHandler.UpdateEmployee(c)

		// Assert response
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	// Test UpdateEmployee with missing ID
	t.Run("UpdateEmployeeMissingID", func(t *testing.T) {
		// Create update request
		updateReqBody := UpdateEmployeeRequest{
			Position: "高级软件工程师",
		}
		jsonReq, _ := json.Marshal(updateReqBody)

		// Create HTTP request and response recorder
		req, _ := http.NewRequest("PUT", "/employees/", bytes.NewBuffer(jsonReq))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		// Create gin context and call handler
		c, _ := gin.CreateTestContext(w)
		c.Request = req
		// Don't add param ID
		empHandler.UpdateEmployee(c)

		// Assert response
		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "employee id is required")
	})

	// Test UpdateEmployee with department not found
	t.Run("UpdateEmployeeDepartmentNotFound", func(t *testing.T) {
		// First create an employee
		createReq := &service.CreateEmployeeRequest{
			UserID:     "user_123",
			TeamID:     "team_123",
			DeptID:     "dept_123",
			EmployeeID: "emp008",
			RealName:   "钱十一",
			Gender:     "男",
			BirthDate:  time.Now().AddDate(-30, 0, 0),
			HireDate:   time.Now().AddDate(-2, 0, 0),
			Position:   "软件工程师",
		}
		emp, _ := mockService.CreateEmployee(context.Background(), createReq)

		// Create a mock service that returns "department not found" error
		mockServiceWithDeptError := &mockEmployeeServiceWithError{
			updateEmployeeError: errors.New("department not found"),
		}

		handler := NewEmployeeHandlerWithService(mockServiceWithDeptError)

		// Create update request
		updateReqBody := UpdateEmployeeRequest{
			DeptID: "dept_456",
		}
		jsonReq, _ := json.Marshal(updateReqBody)

		// Create HTTP request and response recorder
		req, _ := http.NewRequest("PUT", "/employees/"+emp.ID, bytes.NewBuffer(jsonReq))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		// Create gin context and call handler
		c, _ := gin.CreateTestContext(w)
		c.Request = req
		c.AddParam("id", emp.ID)
		handler.UpdateEmployee(c)

		// Assert response
		assert.Equal(t, http.StatusNotFound, w.Code)
		assert.Contains(t, w.Body.String(), "department not found")
	})

	// Test UpdateEmployee with service error
	t.Run("UpdateEmployeeServiceError", func(t *testing.T) {
		// First create an employee
		createReq := &service.CreateEmployeeRequest{
			UserID:     "user_123",
			TeamID:     "team_123",
			DeptID:     "dept_123",
			EmployeeID: "emp009",
			RealName:   "孙十二",
			Gender:     "女",
			BirthDate:  time.Now().AddDate(-30, 0, 0),
			HireDate:   time.Now().AddDate(-2, 0, 0),
			Position:   "软件工程师",
		}
		emp, _ := mockService.CreateEmployee(context.Background(), createReq)

		// Create a mock service that returns an error
		mockServiceWithError := &mockEmployeeServiceWithError{
			updateEmployeeError: errors.New("internal server error"),
		}

		handler := NewEmployeeHandlerWithService(mockServiceWithError)

		// Create update request
		updateReqBody := UpdateEmployeeRequest{
			Position: "高级软件工程师",
		}
		jsonReq, _ := json.Marshal(updateReqBody)

		// Create HTTP request and response recorder
		req, _ := http.NewRequest("PUT", "/employees/"+emp.ID, bytes.NewBuffer(jsonReq))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		// Create gin context and call handler
		c, _ := gin.CreateTestContext(w)
		c.Request = req
		c.AddParam("id", emp.ID)
		handler.UpdateEmployee(c)

		// Assert response
		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Contains(t, w.Body.String(), "internal server error")
	})

	// Test DeleteEmployee
	t.Run("DeleteEmployee", func(t *testing.T) {
		// First create an employee
		createReq := &service.CreateEmployeeRequest{
			UserID:     "user_123",
			TeamID:     "team_123",
			DeptID:     "dept_123",
			EmployeeID: "emp010",
			RealName:   "李十三",
			Gender:     "男",
			BirthDate:  time.Now().AddDate(-28, 0, 0),
			HireDate:   time.Now().AddDate(-1, 0, 0),
			Position:   "UI设计师",
		}
		emp, _ := mockService.CreateEmployee(context.Background(), createReq)

		// Create HTTP request and response recorder
		req, _ := http.NewRequest("DELETE", "/employees/"+emp.ID, nil)
		w := httptest.NewRecorder()

		// Create gin context and call handler
		c, _ := gin.CreateTestContext(w)
		c.Request = req
		c.AddParam("id", emp.ID)
		empHandler.DeleteEmployee(c)

		// Assert response
		assert.Equal(t, http.StatusOK, w.Code)
	})

	// Test DeleteEmployee with non-existent ID
	t.Run("DeleteEmployeeNotFound", func(t *testing.T) {
		// Create HTTP request and response recorder
		req, _ := http.NewRequest("DELETE", "/employees/non-existent", nil)
		w := httptest.NewRecorder()

		// Create gin context and call handler
		c, _ := gin.CreateTestContext(w)
		c.Request = req
		c.AddParam("id", "non-existent")
		empHandler.DeleteEmployee(c)

		// Assert response
		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	// Test DeleteEmployee with missing ID
	t.Run("DeleteEmployeeMissingID", func(t *testing.T) {
		// Create HTTP request and response recorder
		req, _ := http.NewRequest("DELETE", "/employees/", nil)
		w := httptest.NewRecorder()

		// Create gin context and call handler
		c, _ := gin.CreateTestContext(w)
		c.Request = req
		// Don't add param ID
		empHandler.DeleteEmployee(c)

		// Assert response
		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "employee id is required")
	})

	// Test DeleteEmployee with service error
	t.Run("DeleteEmployeeServiceError", func(t *testing.T) {
		// First create an employee
		createReq := &service.CreateEmployeeRequest{
			UserID:     "user_123",
			TeamID:     "team_123",
			DeptID:     "dept_123",
			EmployeeID: "emp011",
			RealName:   "周十四",
			Gender:     "女",
			BirthDate:  time.Now().AddDate(-28, 0, 0),
			HireDate:   time.Now().AddDate(-1, 0, 0),
			Position:   "UI设计师",
		}
		emp, _ := mockService.CreateEmployee(context.Background(), createReq)

		// Create a mock service that returns an error
		mockServiceWithError := &mockEmployeeServiceWithError{
			deleteEmployeeError: errors.New("internal server error"),
		}

		handler := NewEmployeeHandlerWithService(mockServiceWithError)

		// Create HTTP request and response recorder
		req, _ := http.NewRequest("DELETE", "/employees/"+emp.ID, nil)
		w := httptest.NewRecorder()

		// Create gin context and call handler
		c, _ := gin.CreateTestContext(w)
		c.Request = req
		c.AddParam("id", emp.ID)
		handler.DeleteEmployee(c)

		// Assert response
		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Contains(t, w.Body.String(), "internal server error")
	})

	// Test ListEmployees
	t.Run("ListEmployees", func(t *testing.T) {
		// Create a few employees
		for i := 1; i <= 3; i++ {
			createReq := &service.CreateEmployeeRequest{
				UserID:     "user_" + string(rune(i+'0')),
				TeamID:     "team_list",
				DeptID:     "dept_list",
				EmployeeID: "emp0" + string(rune(i+'0')),
				RealName:   "员工" + string(rune(i+'0')),
				Gender:     "男",
				BirthDate:  time.Now().AddDate(-25-i, 0, 0),
				HireDate:   time.Now().AddDate(-i, 0, 0),
				Position:   "职位" + string(rune(i+'0')),
			}
			_, _ = mockService.CreateEmployee(context.Background(), createReq)
		}

		// Create HTTP request and response recorder
		req, _ := http.NewRequest("GET", "/employees?team_id=team_list", nil)
		w := httptest.NewRecorder()

		// Create gin context and call handler
		c, _ := gin.CreateTestContext(w)
		c.Request = req
		empHandler.ListEmployees(c)

		// Assert response
		assert.Equal(t, http.StatusOK, w.Code)
	})

	// Test ListEmployees with pagination
	t.Run("ListEmployeesWithPagination", func(t *testing.T) {
		// Create HTTP request with pagination parameters
		req, _ := http.NewRequest("GET", "/employees?team_id=team_list&page=1&size=2", nil)
		w := httptest.NewRecorder()

		// Create gin context and call handler
		c, _ := gin.CreateTestContext(w)
		c.Request = req
		empHandler.ListEmployees(c)

		// Assert response
		assert.Equal(t, http.StatusOK, w.Code)
	})

	// Test ListEmployees with invalid page
	t.Run("ListEmployeesInvalidPage", func(t *testing.T) {
		// Create HTTP request with invalid page
		req, _ := http.NewRequest("GET", "/employees?team_id=team_list&page=invalid", nil)
		w := httptest.NewRecorder()

		// Create gin context and call handler
		c, _ := gin.CreateTestContext(w)
		c.Request = req
		empHandler.ListEmployees(c)

		// Assert response
		assert.Equal(t, http.StatusOK, w.Code) // Should default to page 1
	})

	// Test ListEmployees with invalid size
	t.Run("ListEmployeesInvalidSize", func(t *testing.T) {
		// Create HTTP request with invalid size
		req, _ := http.NewRequest("GET", "/employees?team_id=team_list&size=invalid", nil)
		w := httptest.NewRecorder()

		// Create gin context and call handler
		c, _ := gin.CreateTestContext(w)
		c.Request = req
		empHandler.ListEmployees(c)

		// Assert response
		assert.Equal(t, http.StatusOK, w.Code) // Should default to size 10
	})

	// Test ListEmployees with service error
	t.Run("ListEmployeesServiceError", func(t *testing.T) {
		// Create a mock service that returns an error
		mockServiceWithError := &mockEmployeeServiceWithError{
			listEmployeesError: errors.New("internal server error"),
		}

		handler := NewEmployeeHandlerWithService(mockServiceWithError)

		// Create HTTP request and response recorder
		req, _ := http.NewRequest("GET", "/employees?team_id=team_list", nil)
		w := httptest.NewRecorder()

		// Create gin context and call handler
		c, _ := gin.CreateTestContext(w)
		c.Request = req
		handler.ListEmployees(c)

		// Assert response
		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Contains(t, w.Body.String(), "internal server error")
	})

	// Test GetEmployee
	t.Run("GetEmployee", func(t *testing.T) {
		// First create an employee
		createReq := &service.CreateEmployeeRequest{
			UserID:     "user_123",
			TeamID:     "team_123",
			DeptID:     "dept_123",
			EmployeeID: "emp012",
			RealName:   "吴十五",
			Gender:     "男",
			BirthDate:  time.Now().AddDate(-32, 0, 0),
			HireDate:   time.Now().AddDate(-3, 0, 0),
			Position:   "测试工程师",
		}
		emp, _ := mockService.CreateEmployee(context.Background(), createReq)

		// Create HTTP request and response recorder
		req, _ := http.NewRequest("GET", "/employees/"+emp.ID, nil)
		w := httptest.NewRecorder()

		// Create gin context and call handler
		c, _ := gin.CreateTestContext(w)
		c.Request = req
		c.AddParam("id", emp.ID)
		empHandler.GetEmployee(c)

		// Assert response
		assert.Equal(t, http.StatusOK, w.Code)
	})

	// Test GetEmployee with non-existent ID
	t.Run("GetEmployeeNotFound", func(t *testing.T) {
		// Create HTTP request and response recorder
		req, _ := http.NewRequest("GET", "/employees/non-existent", nil)
		w := httptest.NewRecorder()

		// Create gin context and call handler
		c, _ := gin.CreateTestContext(w)
		c.Request = req
		c.AddParam("id", "non-existent")
		empHandler.GetEmployee(c)

		// Assert response
		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	// Test GetEmployee with missing ID
	t.Run("GetEmployeeMissingID", func(t *testing.T) {
		// Create HTTP request and response recorder
		req, _ := http.NewRequest("GET", "/employees/", nil)
		w := httptest.NewRecorder()

		// Create gin context and call handler
		c, _ := gin.CreateTestContext(w)
		c.Request = req
		// Don't add param ID
		empHandler.GetEmployee(c)

		// Assert response
		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "employee id is required")
	})

	// Test GetEmployee with service error
	t.Run("GetEmployeeServiceError", func(t *testing.T) {
		// Create a mock service that returns an error
		mockServiceWithError := &mockEmployeeServiceWithError{
			getEmployeeError: errors.New("internal server error"),
		}

		handler := NewEmployeeHandlerWithService(mockServiceWithError)

		// Create HTTP request and response recorder
		req, _ := http.NewRequest("GET", "/employees/emp_123", nil)
		w := httptest.NewRecorder()

		// Create gin context and call handler
		c, _ := gin.CreateTestContext(w)
		c.Request = req
		c.AddParam("id", "emp_123")
		handler.GetEmployee(c)

		// Assert response
		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Contains(t, w.Body.String(), "internal server error")
	})
}

// mockEmployeeServiceWithError is a mock implementation of the EmployeeServiceInterface that returns errors
type mockEmployeeServiceWithError struct {
	createEmployeeError error
	updateEmployeeError error
	deleteEmployeeError error
	listEmployeesError  error
	getEmployeeError    error
}

func (m *mockEmployeeServiceWithError) CreateEmployee(ctx context.Context, req *service.CreateEmployeeRequest) (*domain.Employee, error) {
	return nil, m.createEmployeeError
}

func (m *mockEmployeeServiceWithError) UpdateEmployee(ctx context.Context, empID string, req *service.UpdateEmployeeRequest) error {
	return m.updateEmployeeError
}

func (m *mockEmployeeServiceWithError) DeleteEmployee(ctx context.Context, empID string) error {
	return m.deleteEmployeeError
}

func (m *mockEmployeeServiceWithError) ListEmployees(ctx context.Context, req *service.ListEmployeesRequest) ([]*domain.Employee, int64, error) {
	return nil, 0, m.listEmployeesError
}

func (m *mockEmployeeServiceWithError) GetEmployee(ctx context.Context, empID string) (*domain.Employee, error) {
	return nil, m.getEmployeeError
}