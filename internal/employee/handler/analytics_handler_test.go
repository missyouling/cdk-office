package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"cdk-office/internal/employee/service"
	"cdk-office/internal/shared/testutils"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockAnalyticsService is a mock implementation of AnalyticsServiceInterface
type MockAnalyticsService struct {
	mock.Mock
}

func (m *MockAnalyticsService) GetEmployeeCountByDepartment(ctx context.Context, teamID string) ([]*service.DepartmentEmployeeCount, error) {
	args := m.Called(ctx, teamID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*service.DepartmentEmployeeCount), args.Error(1)
}

func (m *MockAnalyticsService) GetEmployeeCountByPosition(ctx context.Context, teamID string) ([]*service.PositionEmployeeCount, error) {
	args := m.Called(ctx, teamID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*service.PositionEmployeeCount), args.Error(1)
}

func (m *MockAnalyticsService) GetEmployeeLifecycleStats(ctx context.Context, teamID string) (*service.EmployeeLifecycleStats, error) {
	args := m.Called(ctx, teamID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*service.EmployeeLifecycleStats), args.Error(1)
}

func (m *MockAnalyticsService) GetEmployeeAgeDistribution(ctx context.Context, teamID string) ([]*service.AgeGroupCount, error) {
	args := m.Called(ctx, teamID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*service.AgeGroupCount), args.Error(1)
}

func (m *MockAnalyticsService) GetEmployeePerformanceStats(ctx context.Context, teamID string) (*service.EmployeePerformanceStats, error) {
	args := m.Called(ctx, teamID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*service.EmployeePerformanceStats), args.Error(1)
}

func (m *MockAnalyticsService) GetTerminationAnalysis(ctx context.Context, teamID string) (*service.TerminationAnalysis, error) {
	args := m.Called(ctx, teamID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*service.TerminationAnalysis), args.Error(1)
}

func (m *MockAnalyticsService) GetEmployeeTurnoverRate(ctx context.Context, teamID string) (*service.EmployeeTurnoverRate, error) {
	args := m.Called(ctx, teamID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*service.EmployeeTurnoverRate), args.Error(1)
}

func (m *MockAnalyticsService) GetSurveyAnalysis(ctx context.Context, teamID string) (*service.SurveyAnalysis, error) {
	args := m.Called(ctx, teamID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*service.SurveyAnalysis), args.Error(1)
}

// TestNewAnalyticsHandler tests the NewAnalyticsHandler function
func TestNewAnalyticsHandler(t *testing.T) {
	handler := NewAnalyticsHandler()
	assert.NotNil(t, handler)
	assert.NotNil(t, handler.analyticsService)
}

// TestGetEmployeeCountByDepartment tests the GetEmployeeCountByDepartment handler
func TestGetEmployeeCountByDepartment(t *testing.T) {
	// Set up test environment
	gin.SetMode(gin.TestMode)

	// Create mock service
	mockService := new(MockAnalyticsService)

	// Create handler with mock service
	handler := &AnalyticsHandler{
		analyticsService: mockService,
	}

	// Create test router
	router := gin.New()
	router.GET("/analytics/department", handler.GetEmployeeCountByDepartment)

	// Test successful retrieval
	t.Run("SuccessfulRetrieval", func(t *testing.T) {
		// Prepare test data
		teamID := "team_123"
		expectedResults := []*service.DepartmentEmployeeCount{
			{DepartmentName: "Engineering", EmployeeCount: 10},
			{DepartmentName: "Marketing", EmployeeCount: 5},
			{DepartmentName: "Sales", EmployeeCount: 8},
		}

		// Mock service response
		mockService.On("GetEmployeeCountByDepartment", mock.Anything, teamID).Return(expectedResults, nil).Once()

		// Create request
		req, _ := http.NewRequest(http.MethodGet, "/analytics/department?team_id="+teamID, nil)

		// Create response recorder
		w := httptest.NewRecorder()

		// Perform request
		router.ServeHTTP(w, req)

		// Assert response
		assert.Equal(t, http.StatusOK, w.Code)

		// Parse response
		var response []*service.DepartmentEmployeeCount
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, expectedResults, response)

		// Assert mock expectations
		mockService.AssertExpectations(t)
	})

	// Test missing team ID
	t.Run("MissingTeamID", func(t *testing.T) {
		// Create request without team ID
		req, _ := http.NewRequest(http.MethodGet, "/analytics/department", nil)

		// Create response recorder
		w := httptest.NewRecorder()

		// Perform request
		router.ServeHTTP(w, req)

		// Assert response
		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "team id is required")
	})

	// Test service error
	t.Run("ServiceError", func(t *testing.T) {
		// Prepare test data
		teamID := "team_123"

		// Mock service response
		mockService.On("GetEmployeeCountByDepartment", mock.Anything, teamID).Return([]*service.DepartmentEmployeeCount(nil), testutils.NewError("internal error")).Once()

		// Create request
		req, _ := http.NewRequest(http.MethodGet, "/analytics/department?team_id="+teamID, nil)

		// Create response recorder
		w := httptest.NewRecorder()

		// Perform request
		router.ServeHTTP(w, req)

		// Assert response
		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Contains(t, w.Body.String(), "internal error")

		// Assert mock expectations
		mockService.AssertExpectations(t)
	})
}

// TestGetEmployeeCountByPosition tests the GetEmployeeCountByPosition handler
func TestGetEmployeeCountByPosition(t *testing.T) {
	// Set up test environment
	gin.SetMode(gin.TestMode)

	// Create mock service
	mockService := new(MockAnalyticsService)

	// Create handler with mock service
	handler := &AnalyticsHandler{
		analyticsService: mockService,
	}

	// Create test router
	router := gin.New()
	router.GET("/analytics/position", handler.GetEmployeeCountByPosition)

	// Test successful retrieval
	t.Run("SuccessfulRetrieval", func(t *testing.T) {
		// Prepare test data
		teamID := "team_123"
		expectedResults := []*service.PositionEmployeeCount{
			{Position: "Engineer", EmployeeCount: 10},
			{Position: "Manager", EmployeeCount: 3},
			{Position: "Designer", EmployeeCount: 2},
			{Position: "Analyst", EmployeeCount: 5},
		}

		// Mock service response
		mockService.On("GetEmployeeCountByPosition", mock.Anything, teamID).Return(expectedResults, nil).Once()

		// Create request
		req, _ := http.NewRequest(http.MethodGet, "/analytics/position?team_id="+teamID, nil)

		// Create response recorder
		w := httptest.NewRecorder()

		// Perform request
		router.ServeHTTP(w, req)

		// Assert response
		assert.Equal(t, http.StatusOK, w.Code)

		// Parse response
		var response []*service.PositionEmployeeCount
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, expectedResults, response)

		// Assert mock expectations
		mockService.AssertExpectations(t)
	})

	// Test missing team ID
	t.Run("MissingTeamID", func(t *testing.T) {
		// Create request without team ID
		req, _ := http.NewRequest(http.MethodGet, "/analytics/position", nil)

		// Create response recorder
		w := httptest.NewRecorder()

		// Perform request
		router.ServeHTTP(w, req)

		// Assert response
		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "team id is required")
	})

	// Test service error
	t.Run("ServiceError", func(t *testing.T) {
		// Prepare test data
		teamID := "team_123"

		// Mock service response
		mockService.On("GetEmployeeCountByPosition", mock.Anything, teamID).Return([]*service.PositionEmployeeCount(nil), testutils.NewError("internal error")).Once()

		// Create request
		req, _ := http.NewRequest(http.MethodGet, "/analytics/position?team_id="+teamID, nil)

		// Create response recorder
		w := httptest.NewRecorder()

		// Perform request
		router.ServeHTTP(w, req)

		// Assert response
		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Contains(t, w.Body.String(), "internal error")

		// Assert mock expectations
		mockService.AssertExpectations(t)
	})
}

// TestGetEmployeeLifecycleStats tests the GetEmployeeLifecycleStats handler
func TestGetEmployeeLifecycleStats(t *testing.T) {
	// Set up test environment
	gin.SetMode(gin.TestMode)

	// Create mock service
	mockService := new(MockAnalyticsService)

	// Create handler with mock service
	handler := &AnalyticsHandler{
		analyticsService: mockService,
	}

	// Create test router
	router := gin.New()
	router.GET("/analytics/lifecycle", handler.GetEmployeeLifecycleStats)

	// Test successful retrieval
	t.Run("SuccessfulRetrieval", func(t *testing.T) {
		// Prepare test data
		teamID := "team_123"
		expectedStats := &service.EmployeeLifecycleStats{
			TotalEmployees:      100,
			ActiveEmployees:     90,
			TerminatedEmployees: 5,
			PromotionCount:      2,
			TransferCount:       3,
		}

		// Mock service response
		mockService.On("GetEmployeeLifecycleStats", mock.Anything, teamID).Return(expectedStats, nil).Once()

		// Create request
		req, _ := http.NewRequest(http.MethodGet, "/analytics/lifecycle?team_id="+teamID, nil)

		// Create response recorder
		w := httptest.NewRecorder()

		// Perform request
		router.ServeHTTP(w, req)

		// Assert response
		assert.Equal(t, http.StatusOK, w.Code)

		// Parse response
		var response service.EmployeeLifecycleStats
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, *expectedStats, response)

		// Assert mock expectations
		mockService.AssertExpectations(t)
	})

	// Test missing team ID
	t.Run("MissingTeamID", func(t *testing.T) {
		// Create request without team ID
		req, _ := http.NewRequest(http.MethodGet, "/analytics/lifecycle", nil)

		// Create response recorder
		w := httptest.NewRecorder()

		// Perform request
		router.ServeHTTP(w, req)

		// Assert response
		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "team id is required")
	})

	// Test service error
	t.Run("ServiceError", func(t *testing.T) {
		// Prepare test data
		teamID := "team_123"

		// Mock service response
		mockService.On("GetEmployeeLifecycleStats", mock.Anything, teamID).Return((*service.EmployeeLifecycleStats)(nil), testutils.NewError("internal error")).Once()

		// Create request
		req, _ := http.NewRequest(http.MethodGet, "/analytics/lifecycle?team_id="+teamID, nil)

		// Create response recorder
		w := httptest.NewRecorder()

		// Perform request
		router.ServeHTTP(w, req)

		// Assert response
		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Contains(t, w.Body.String(), "internal error")

		// Assert mock expectations
		mockService.AssertExpectations(t)
	})
}

// TestGetEmployeeAgeDistribution tests the GetEmployeeAgeDistribution handler
func TestGetEmployeeAgeDistribution(t *testing.T) {
	// Set up test environment
	gin.SetMode(gin.TestMode)

	// Create mock service
	mockService := new(MockAnalyticsService)

	// Create handler with mock service
	handler := &AnalyticsHandler{
		analyticsService: mockService,
	}

	// Create test router
	router := gin.New()
	router.GET("/analytics/age", handler.GetEmployeeAgeDistribution)

	// Test successful retrieval
	t.Run("SuccessfulRetrieval", func(t *testing.T) {
		// Prepare test data
		teamID := "team_123"
		expectedResults := []*service.AgeGroupCount{
			{AgeGroup: "20-30", EmployeeCount: 30},
			{AgeGroup: "31-40", EmployeeCount: 40},
			{AgeGroup: "41-50", EmployeeCount: 20},
			{AgeGroup: "51-60", EmployeeCount: 10},
		}

		// Mock service response
		mockService.On("GetEmployeeAgeDistribution", mock.Anything, teamID).Return(expectedResults, nil).Once()

		// Create request
		req, _ := http.NewRequest(http.MethodGet, "/analytics/age?team_id="+teamID, nil)

		// Create response recorder
		w := httptest.NewRecorder()

		// Perform request
		router.ServeHTTP(w, req)

		// Assert response
		assert.Equal(t, http.StatusOK, w.Code)

		// Parse response
		var response []*service.AgeGroupCount
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, expectedResults, response)

		// Assert mock expectations
		mockService.AssertExpectations(t)
	})

	// Test missing team ID
	t.Run("MissingTeamID", func(t *testing.T) {
		// Create request without team ID
		req, _ := http.NewRequest(http.MethodGet, "/analytics/age", nil)

		// Create response recorder
		w := httptest.NewRecorder()

		// Perform request
		router.ServeHTTP(w, req)

		// Assert response
		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "team id is required")
	})

	// Test service error
	t.Run("ServiceError", func(t *testing.T) {
		// Prepare test data
		teamID := "team_123"

		// Mock service response
		mockService.On("GetEmployeeAgeDistribution", mock.Anything, teamID).Return([]*service.AgeGroupCount(nil), testutils.NewError("internal error")).Once()

		// Create request
		req, _ := http.NewRequest(http.MethodGet, "/analytics/age?team_id="+teamID, nil)

		// Create response recorder
		w := httptest.NewRecorder()

		// Perform request
		router.ServeHTTP(w, req)

		// Assert response
		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Contains(t, w.Body.String(), "internal error")

		// Assert mock expectations
		mockService.AssertExpectations(t)
	})
}

// TestGetEmployeePerformanceStats tests the GetEmployeePerformanceStats handler
func TestGetEmployeePerformanceStats(t *testing.T) {
	// Set up test environment
	gin.SetMode(gin.TestMode)

	// Create mock service
	mockService := new(MockAnalyticsService)

	// Create handler with mock service
	handler := &AnalyticsHandler{
		analyticsService: mockService,
	}

	// Create test router
	router := gin.New()
	router.GET("/analytics/performance", handler.GetEmployeePerformanceStats)

	// Test successful retrieval
	t.Run("SuccessfulRetrieval", func(t *testing.T) {
		// Prepare test data
		teamID := "team_123"
		expectedStats := &service.EmployeePerformanceStats{
			AverageScore:     4.2,
			HighPerformers:   25,
			LowPerformers:    5,
			TotalReviews:     100,
		}

		// Mock service response
		mockService.On("GetEmployeePerformanceStats", mock.Anything, teamID).Return(expectedStats, nil).Once()

		// Create request
		req, _ := http.NewRequest(http.MethodGet, "/analytics/performance?team_id="+teamID, nil)

		// Create response recorder
		w := httptest.NewRecorder()

		// Perform request
		router.ServeHTTP(w, req)

		// Assert response
		assert.Equal(t, http.StatusOK, w.Code)

		// Parse response
		var response service.EmployeePerformanceStats
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, *expectedStats, response)

		// Assert mock expectations
		mockService.AssertExpectations(t)
	})

	// Test missing team ID
	t.Run("MissingTeamID", func(t *testing.T) {
		// Create request without team ID
		req, _ := http.NewRequest(http.MethodGet, "/analytics/performance", nil)

		// Create response recorder
		w := httptest.NewRecorder()

		// Perform request
		router.ServeHTTP(w, req)

		// Assert response
		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "team id is required")
	})

	// Test service error
	t.Run("ServiceError", func(t *testing.T) {
		// Prepare test data
		teamID := "team_123"

		// Mock service response
		mockService.On("GetEmployeePerformanceStats", mock.Anything, teamID).Return((*service.EmployeePerformanceStats)(nil), testutils.NewError("internal error")).Once()

		// Create request
		req, _ := http.NewRequest(http.MethodGet, "/analytics/performance?team_id="+teamID, nil)

		// Create response recorder
		w := httptest.NewRecorder()

		// Perform request
		router.ServeHTTP(w, req)

		// Assert response
		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Contains(t, w.Body.String(), "internal error")

		// Assert mock expectations
		mockService.AssertExpectations(t)
	})
}

// TestGetTerminationAnalysis tests the GetTerminationAnalysis handler
func TestGetTerminationAnalysis(t *testing.T) {
	// Set up test environment
	gin.SetMode(gin.TestMode)

	// Create mock service
	mockService := new(MockAnalyticsService)

	// Create handler with mock service
	handler := &AnalyticsHandler{
		analyticsService: mockService,
	}

	// Create test router
	router := gin.New()
	router.GET("/analytics/termination", handler.GetTerminationAnalysis)

	// Test successful retrieval
	t.Run("SuccessfulRetrieval", func(t *testing.T) {
		// Prepare test data
		teamID := "team_123"
		expectedAnalysis := &service.TerminationAnalysis{
			TotalTerminations: 10,
			TerminationReasons: []*service.ReasonCount{
				{Reason: "Resignation", Count: 5},
				{Reason: "Retirement", Count: 2},
				{Reason: "Dismissal", Count: 3},
			},
		}

		// Mock service response
		mockService.On("GetTerminationAnalysis", mock.Anything, teamID).Return(expectedAnalysis, nil).Once()

		// Create request
		req, _ := http.NewRequest(http.MethodGet, "/analytics/termination?team_id="+teamID, nil)

		// Create response recorder
		w := httptest.NewRecorder()

		// Perform request
		router.ServeHTTP(w, req)

		// Assert response
		assert.Equal(t, http.StatusOK, w.Code)

		// Parse response
		var response service.TerminationAnalysis
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, *expectedAnalysis, response)

		// Assert mock expectations
		mockService.AssertExpectations(t)
	})

	// Test missing team ID
	t.Run("MissingTeamID", func(t *testing.T) {
		// Create request without team ID
		req, _ := http.NewRequest(http.MethodGet, "/analytics/termination", nil)

		// Create response recorder
		w := httptest.NewRecorder()

		// Perform request
		router.ServeHTTP(w, req)

		// Assert response
		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "team id is required")
	})

	// Test service error
	t.Run("ServiceError", func(t *testing.T) {
		// Prepare test data
		teamID := "team_123"

		// Mock service response
		mockService.On("GetTerminationAnalysis", mock.Anything, teamID).Return((*service.TerminationAnalysis)(nil), testutils.NewError("internal error")).Once()

		// Create request
		req, _ := http.NewRequest(http.MethodGet, "/analytics/termination?team_id="+teamID, nil)

		// Create response recorder
		w := httptest.NewRecorder()

		// Perform request
		router.ServeHTTP(w, req)

		// Assert response
		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Contains(t, w.Body.String(), "internal error")

		// Assert mock expectations
		mockService.AssertExpectations(t)
	})
}

// TestGetEmployeeTurnoverRate tests the GetEmployeeTurnoverRate handler
func TestGetEmployeeTurnoverRate(t *testing.T) {
	// Set up test environment
	gin.SetMode(gin.TestMode)

	// Create mock service
	mockService := new(MockAnalyticsService)

	// Create handler with mock service
	handler := &AnalyticsHandler{
		analyticsService: mockService,
	}

	// Create test router
	router := gin.New()
	router.GET("/analytics/turnover", handler.GetEmployeeTurnoverRate)

	// Test successful retrieval
	t.Run("SuccessfulRetrieval", func(t *testing.T) {
		// Prepare test data
		teamID := "team_123"
		expectedRate := &service.EmployeeTurnoverRate{
			TurnoverRate:    12.5,
			PromotionRate:   5.0,
			TransferRate:    3.0,
			TotalEmployees:  100,
		}

		// Mock service response
		mockService.On("GetEmployeeTurnoverRate", mock.Anything, teamID).Return(expectedRate, nil).Once()

		// Create request
		req, _ := http.NewRequest(http.MethodGet, "/analytics/turnover?team_id="+teamID, nil)

		// Create response recorder
		w := httptest.NewRecorder()

		// Perform request
		router.ServeHTTP(w, req)

		// Assert response
		assert.Equal(t, http.StatusOK, w.Code)

		// Parse response
		var response service.EmployeeTurnoverRate
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, *expectedRate, response)

		// Assert mock expectations
		mockService.AssertExpectations(t)
	})

	// Test missing team ID
	t.Run("MissingTeamID", func(t *testing.T) {
		// Create request without team ID
		req, _ := http.NewRequest(http.MethodGet, "/analytics/turnover", nil)

		// Create response recorder
		w := httptest.NewRecorder()

		// Perform request
		router.ServeHTTP(w, req)

		// Assert response
		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "team id is required")
	})

	// Test service error
	t.Run("ServiceError", func(t *testing.T) {
		// Prepare test data
		teamID := "team_123"

		// Mock service response
		mockService.On("GetEmployeeTurnoverRate", mock.Anything, teamID).Return((*service.EmployeeTurnoverRate)(nil), testutils.NewError("internal error")).Once()

		// Create request
		req, _ := http.NewRequest(http.MethodGet, "/analytics/turnover?team_id="+teamID, nil)

		// Create response recorder
		w := httptest.NewRecorder()

		// Perform request
		router.ServeHTTP(w, req)

		// Assert response
		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Contains(t, w.Body.String(), "internal error")

		// Assert mock expectations
		mockService.AssertExpectations(t)
	})
}

// TestGetSurveyAnalysis tests the GetSurveyAnalysis handler
func TestGetSurveyAnalysis(t *testing.T) {
	// Set up test environment
	gin.SetMode(gin.TestMode)

	// Create mock service
	mockService := new(MockAnalyticsService)

	// Create handler with mock service
	handler := &AnalyticsHandler{
		analyticsService: mockService,
	}

	// Create test router
	router := gin.New()
	router.GET("/analytics/survey", handler.GetSurveyAnalysis)

	// Test successful retrieval
	t.Run("SuccessfulRetrieval", func(t *testing.T) {
		// Prepare test data
		teamID := "team_123"
		expectedAnalysis := &service.SurveyAnalysis{
			TotalSurveys:    20,
			TotalResponses:  17,
			ResponseRate:    85.0,
			AverageScore:    4.2,
		}

		// Mock service response
		mockService.On("GetSurveyAnalysis", mock.Anything, teamID).Return(expectedAnalysis, nil).Once()

		// Create request
		req, _ := http.NewRequest(http.MethodGet, "/analytics/survey?team_id="+teamID, nil)

		// Create response recorder
		w := httptest.NewRecorder()

		// Perform request
		router.ServeHTTP(w, req)

		// Assert response
		assert.Equal(t, http.StatusOK, w.Code)

		// Parse response
		var response service.SurveyAnalysis
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, *expectedAnalysis, response)

		// Assert mock expectations
		mockService.AssertExpectations(t)
	})

	// Test missing team ID
	t.Run("MissingTeamID", func(t *testing.T) {
		// Create request without team ID
		req, _ := http.NewRequest(http.MethodGet, "/analytics/survey", nil)

		// Create response recorder
		w := httptest.NewRecorder()

		// Perform request
		router.ServeHTTP(w, req)

		// Assert response
		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "team id is required")
	})

	// Test service error
	t.Run("ServiceError", func(t *testing.T) {
		// Prepare test data
		teamID := "team_123"

		// Mock service response
		mockService.On("GetSurveyAnalysis", mock.Anything, teamID).Return((*service.SurveyAnalysis)(nil), testutils.NewError("internal error")).Once()

		// Create request
		req, _ := http.NewRequest(http.MethodGet, "/analytics/survey?team_id="+teamID, nil)

		// Create response recorder
		w := httptest.NewRecorder()

		// Perform request
		router.ServeHTTP(w, req)

		// Assert response
		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Contains(t, w.Body.String(), "internal error")

		// Assert mock expectations
		mockService.AssertExpectations(t)
	})
}