package service

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"cdk-office/internal/employee/domain"
	"cdk-office/internal/shared/testutils"
)

// TestAnalyticsService tests the AnalyticsService
func TestAnalyticsService(t *testing.T) {
	// Set up test environment
	testDB := testutils.SetupTestDB()

	// Create analytics service with database connection
	analyticsService := &AnalyticsService{
		db: testDB,
	}

	// Create a department for testing
	department := &domain.Department{
		ID:   "dept_123",
		Name: "Test Department",
	}
	err := testDB.Create(department).Error
	assert.NoError(t, err)

	// Create some employees for testing
	employees := []*domain.Employee{
		{
			ID:        "emp_1",
			UserID:    "user_1",
			TeamID:    "team_123",
			DeptID:    "dept_123",
			EmployeeID: "emp_001",
			RealName:  "John Doe",
			Gender:    "male",
			BirthDate: time.Now().AddDate(-30, 0, 0),
			HireDate:  time.Now().AddDate(-2, 0, 0),
			Position:  "Developer",
			Status:    "active",
		},
		{
			ID:        "emp_2",
			UserID:    "user_2",
			TeamID:    "team_123",
			DeptID:    "dept_123",
			EmployeeID: "emp_002",
			RealName:  "Jane Smith",
			Gender:    "female",
			BirthDate: time.Now().AddDate(-25, 0, 0),
			HireDate:  time.Now().AddDate(-1, 0, 0),
			Position:  "Designer",
			Status:    "active",
		},
		{
			ID:        "emp_3",
			UserID:    "user_3",
			TeamID:    "team_123",
			DeptID:    "dept_123",
			EmployeeID: "emp_003",
			RealName:  "Bob Johnson",
			Gender:    "male",
			BirthDate: time.Now().AddDate(-40, 0, 0),
			HireDate:  time.Now().AddDate(-3, 0, 0),
			Position:  "Manager",
			Status:    "terminated",
		},
	}

	for _, emp := range employees {
		err := testDB.Create(emp).Error
		assert.NoError(t, err)
	}

	// Create some performance reviews
	reviews := []*domain.PerformanceReview{
		{
			ID:         "rev_1",
			EmployeeID: "emp_1",
			ReviewerID: "user_2",
			Score:      4.5,
			Comments:   "Good performance",
		},
		{
			ID:         "rev_2",
			EmployeeID: "emp_2",
			ReviewerID: "user_1",
			Score:      3.8,
			Comments:   "Satisfactory performance",
		},
	}

	for _, rev := range reviews {
		err := testDB.Create(rev).Error
		assert.NoError(t, err)
	}

	// Create some termination records
	termination := &domain.TerminationRecord{
		ID:            "term_1",
		EmployeeID:    "emp_3",
		TerminationDate: time.Now().AddDate(0, -6, 0),
		Reason:        "Performance",
	}
	err = testDB.Create(termination).Error
	assert.NoError(t, err)

	// Create some lifecycle events
	events := []*EmployeeLifecycleEvent{
		{
			ID:         "event_1",
			EmployeeID: "emp_1",
			EventType:  "promotion",
			CreatedAt:  time.Now().AddDate(0, -6, 0),
		},
		{
			ID:         "event_2",
			EmployeeID: "emp_2",
			EventType:  "transfer",
			CreatedAt:  time.Now().AddDate(0, -3, 0),
		},
	}

	for _, event := range events {
		err := testDB.Create(event).Error
		assert.NoError(t, err)
	}

	// Create some surveys and responses
	survey := &domain.EmployeeSurvey{
		ID:          "survey_1",
		Title:       "Employee Satisfaction",
		Description: "Annual employee satisfaction survey",
		CreatedBy:   "user_1",
		Status:      "active",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	err = testDB.Create(survey).Error
	assert.NoError(t, err)

	response := &domain.SurveyResponse{
		ID:          "resp_1",
		SurveyID:    "survey_1",
		EmployeeID:  "emp_1",
		Responses:   `{"q1": "satisfied", "q2": 4}`,
		SubmittedAt: time.Now(),
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	err = testDB.Create(response).Error
	assert.NoError(t, err)

	// Test GetEmployeeCountByDepartment
	t.Run("GetEmployeeCountByDepartment", func(t *testing.T) {
		ctx := context.Background()

		results, err := analyticsService.GetEmployeeCountByDepartment(ctx, "team_123")

		assert.NoError(t, err)
		assert.NotNil(t, results)
		// We expect at least one department with employees
		assert.GreaterOrEqual(t, len(results), 1)
	})

	// Test GetEmployeeCountByPosition
	t.Run("GetEmployeeCountByPosition", func(t *testing.T) {
		ctx := context.Background()

		results, err := analyticsService.GetEmployeeCountByPosition(ctx, "team_123")

		assert.NoError(t, err)
		assert.NotNil(t, results)
		// We expect at least one position with employees
		assert.GreaterOrEqual(t, len(results), 1)
	})

	// Test GetEmployeeLifecycleStats
	t.Run("GetEmployeeLifecycleStats", func(t *testing.T) {
		ctx := context.Background()

		stats, err := analyticsService.GetEmployeeLifecycleStats(ctx, "team_123")

		assert.NoError(t, err)
		assert.NotNil(t, stats)
		assert.Equal(t, int64(3), stats.TotalEmployees)
		assert.Equal(t, int64(2), stats.ActiveEmployees)
		assert.Equal(t, int64(1), stats.TerminatedEmployees)
	})

	// Test GetEmployeeAgeDistribution
	t.Run("GetEmployeeAgeDistribution", func(t *testing.T) {
		ctx := context.Background()

		results, err := analyticsService.GetEmployeeAgeDistribution(ctx, "team_123")

		assert.NoError(t, err)
		assert.NotNil(t, results)
		// We expect at least one age group with employees
		assert.GreaterOrEqual(t, len(results), 1)
	})

	// Test GetEmployeePerformanceStats
	t.Run("GetEmployeePerformanceStats", func(t *testing.T) {
		ctx := context.Background()

		stats, err := analyticsService.GetEmployeePerformanceStats(ctx, "team_123")

		assert.NoError(t, err)
		assert.NotNil(t, stats)
		assert.Greater(t, stats.AverageScore, 0.0)
		assert.Equal(t, int64(1), stats.HighPerformers)
		assert.Equal(t, int64(0), stats.LowPerformers)
		assert.Equal(t, int64(2), stats.TotalReviews)
	})

	// Test GetTerminationAnalysis
	t.Run("GetTerminationAnalysis", func(t *testing.T) {
		ctx := context.Background()

		analysis, err := analyticsService.GetTerminationAnalysis(ctx, "team_123")

		assert.NoError(t, err)
		assert.NotNil(t, analysis)
		assert.Equal(t, int64(1), analysis.TotalTerminations)
		assert.GreaterOrEqual(t, len(analysis.TerminationReasons), 1)
	})

	// Test GetEmployeeTurnoverRate
	t.Run("GetEmployeeTurnoverRate", func(t *testing.T) {
		ctx := context.Background()

		rate, err := analyticsService.GetEmployeeTurnoverRate(ctx, "team_123")

		assert.NoError(t, err)
		assert.NotNil(t, rate)
		assert.Equal(t, int64(3), rate.TotalEmployees)
		assert.GreaterOrEqual(t, rate.TurnoverRate, 0.0)
		assert.GreaterOrEqual(t, rate.PromotionRate, 0.0)
		assert.GreaterOrEqual(t, rate.TransferRate, 0.0)
	})

	// Test GetSurveyAnalysis
	t.Run("GetSurveyAnalysis", func(t *testing.T) {
		ctx := context.Background()

		analysis, err := analyticsService.GetSurveyAnalysis(ctx, "team_123")

		assert.NoError(t, err)
		assert.NotNil(t, analysis)
		assert.Equal(t, int64(1), analysis.TotalSurveys)
		assert.Equal(t, int64(1), analysis.TotalResponses)
		assert.Greater(t, analysis.ResponseRate, 0.0)
		assert.Equal(t, 0.0, analysis.AverageScore) // As implemented in the service
	})
}