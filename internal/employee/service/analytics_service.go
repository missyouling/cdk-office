package service

import (
	"context"
	"errors"
	"time"

	"cdk-office/internal/employee/domain"
	"cdk-office/internal/shared/database"
	"cdk-office/pkg/logger"
	"gorm.io/gorm"
)

// AnalyticsServiceInterface defines the interface for employee analytics service
type AnalyticsServiceInterface interface {
	GetEmployeeCountByDepartment(ctx context.Context, teamID string) ([]*DepartmentEmployeeCount, error)
	GetEmployeeCountByPosition(ctx context.Context, teamID string) ([]*PositionEmployeeCount, error)
	GetEmployeeLifecycleStats(ctx context.Context, teamID string) (*EmployeeLifecycleStats, error)
	GetEmployeeAgeDistribution(ctx context.Context, teamID string) ([]*AgeGroupCount, error)
	GetEmployeePerformanceStats(ctx context.Context, teamID string) (*EmployeePerformanceStats, error)
	GetTerminationAnalysis(ctx context.Context, teamID string) (*TerminationAnalysis, error)
	GetEmployeeTurnoverRate(ctx context.Context, teamID string) (*EmployeeTurnoverRate, error)
	GetSurveyAnalysis(ctx context.Context, teamID string) (*SurveyAnalysis, error)
}

// AnalyticsService implements the AnalyticsServiceInterface
type AnalyticsService struct {
	db *gorm.DB
}

// NewAnalyticsService creates a new instance of AnalyticsService
func NewAnalyticsService() *AnalyticsService {
	return &AnalyticsService{
		db: database.GetDB(),
	}
}

// DepartmentEmployeeCount represents employee count by department
type DepartmentEmployeeCount struct {
	DepartmentName string `json:"department_name"`
	EmployeeCount  int64  `json:"employee_count"`
}

// PositionEmployeeCount represents employee count by position
type PositionEmployeeCount struct {
	Position      string `json:"position"`
	EmployeeCount int64  `json:"employee_count"`
}

// EmployeeLifecycleStats represents employee lifecycle statistics
type EmployeeLifecycleStats struct {
	TotalEmployees   int64 `json:"total_employees"`
	ActiveEmployees  int64 `json:"active_employees"`
	TerminatedEmployees int64 `json:"terminated_employees"`
	PromotionCount   int64 `json:"promotion_count"`
	TransferCount    int64 `json:"transfer_count"`
}

// AgeGroupCount represents employee count by age group
type AgeGroupCount struct {
	AgeGroup      string `json:"age_group"`
	EmployeeCount int64  `json:"employee_count"`
}

// EmployeePerformanceStats represents employee performance statistics
type EmployeePerformanceStats struct {
	AverageScore     float64 `json:"average_score"`
	HighPerformers   int64   `json:"high_performers"`
	LowPerformers    int64   `json:"low_performers"`
	TotalReviews     int64   `json:"total_reviews"`
}

// TerminationAnalysis represents termination analysis data
type TerminationAnalysis struct {
	TotalTerminations int64            `json:"total_terminations"`
	TerminationReasons []*ReasonCount  `json:"termination_reasons"`
}

// ReasonCount represents count by reason
type ReasonCount struct {
	Reason string `json:"reason"`
	Count  int64  `json:"count"`
}

// EmployeeTurnoverRate represents employee turnover rate
type EmployeeTurnoverRate struct {
	TurnoverRate    float64 `json:"turnover_rate"`
	PromotionRate   float64 `json:"promotion_rate"`
	TransferRate    float64 `json:"transfer_rate"`
	TotalEmployees  int64   `json:"total_employees"`
}

// SurveyAnalysis represents survey analysis data
type SurveyAnalysis struct {
	TotalSurveys    int64   `json:"total_surveys"`
	TotalResponses  int64   `json:"total_responses"`
	ResponseRate    float64 `json:"response_rate"`
	AverageScore    float64 `json:"average_score"`
}

// GetEmployeeCountByDepartment retrieves employee count by department
func (s *AnalyticsService) GetEmployeeCountByDepartment(ctx context.Context, teamID string) ([]*DepartmentEmployeeCount, error) {
	var results []*DepartmentEmployeeCount

	// Execute query to get employee count by department
	if err := s.db.Model(&domain.Employee{}).
		Select("d.name as department_name, count(e.id) as employee_count").
		Joins("e LEFT JOIN departments d ON e.dept_id = d.id").
		Where("e.team_id = ?", teamID).
		Group("d.name").
		Scan(&results).Error; err != nil {
		logger.Error("failed to get employee count by department", "error", err)
		return nil, errors.New("failed to get employee count by department")
	}

	return results, nil
}

// GetEmployeeCountByPosition retrieves employee count by position
func (s *AnalyticsService) GetEmployeeCountByPosition(ctx context.Context, teamID string) ([]*PositionEmployeeCount, error) {
	var results []*PositionEmployeeCount

	// Execute query to get employee count by position
	if err := s.db.Model(&domain.Employee{}).
		Select("position, count(id) as employee_count").
		Where("team_id = ?", teamID).
		Group("position").
		Scan(&results).Error; err != nil {
		logger.Error("failed to get employee count by position", "error", err)
		return nil, errors.New("failed to get employee count by position")
	}

	return results, nil
}

// GetEmployeeLifecycleStats retrieves employee lifecycle statistics
func (s *AnalyticsService) GetEmployeeLifecycleStats(ctx context.Context, teamID string) (*EmployeeLifecycleStats, error) {
	var stats EmployeeLifecycleStats

	// Get total employees
	if err := s.db.Model(&domain.Employee{}).
		Where("team_id = ?", teamID).
		Count(&stats.TotalEmployees).Error; err != nil {
		logger.Error("failed to count total employees", "error", err)
		return nil, errors.New("failed to get employee lifecycle stats")
	}

	// Get active employees
	if err := s.db.Model(&domain.Employee{}).
		Where("team_id = ? AND status = ?", teamID, "active").
		Count(&stats.ActiveEmployees).Error; err != nil {
		logger.Error("failed to count active employees", "error", err)
		return nil, errors.New("failed to get employee lifecycle stats")
	}

	// Get terminated employees
	if err := s.db.Model(&domain.Employee{}).
		Where("team_id = ? AND status = ?", teamID, "terminated").
		Count(&stats.TerminatedEmployees).Error; err != nil {
		logger.Error("failed to count terminated employees", "error", err)
		return nil, errors.New("failed to get employee lifecycle stats")
	}

	// Get promotion count
	if err := s.db.Model(&EmployeeLifecycleEvent{}).
		Joins("e LEFT JOIN employees emp ON e.employee_id = emp.id").
		Where("emp.team_id = ? AND e.event_type = ?", teamID, "promotion").
		Count(&stats.PromotionCount).Error; err != nil {
		logger.Error("failed to count promotions", "error", err)
		return nil, errors.New("failed to get employee lifecycle stats")
	}

	// Get transfer count
	if err := s.db.Model(&EmployeeLifecycleEvent{}).
		Joins("e LEFT JOIN employees emp ON e.employee_id = emp.id").
		Where("emp.team_id = ? AND e.event_type = ?", teamID, "transfer").
		Count(&stats.TransferCount).Error; err != nil {
		logger.Error("failed to count transfers", "error", err)
		return nil, errors.New("failed to get employee lifecycle stats")
	}

	return &stats, nil
}

// GetEmployeeAgeDistribution retrieves employee age distribution
func (s *AnalyticsService) GetEmployeeAgeDistribution(ctx context.Context, teamID string) ([]*AgeGroupCount, error) {
	var results []*AgeGroupCount

	// Execute query to get employee age distribution
	// Note: This is a simplified implementation that may need to be adjusted based on the actual database schema
	if err := s.db.Model(&domain.Employee{}).
		Select(`
			CASE 
				WHEN EXTRACT(YEAR FROM AGE(birth_date)) < 25 THEN '<25'
				WHEN EXTRACT(YEAR FROM AGE(birth_date)) BETWEEN 25 AND 34 THEN '25-34'
				WHEN EXTRACT(YEAR FROM AGE(birth_date)) BETWEEN 35 AND 44 THEN '35-44'
				WHEN EXTRACT(YEAR FROM AGE(birth_date)) BETWEEN 45 AND 54 THEN '45-54'
				ELSE '55+'
			END as age_group,
			count(id) as employee_count
		`).
		Where("team_id = ?", teamID).
		Group("age_group").
		Order("age_group").
		Scan(&results).Error; err != nil {
		logger.Error("failed to get employee age distribution", "error", err)
		return nil, errors.New("failed to get employee age distribution")
	}

	return results, nil
}

// GetEmployeePerformanceStats retrieves employee performance statistics
func (s *AnalyticsService) GetEmployeePerformanceStats(ctx context.Context, teamID string) (*EmployeePerformanceStats, error) {
	var stats EmployeePerformanceStats

	// Get average performance score
	if err := s.db.Model(&domain.PerformanceReview{}).
		Joins("pr LEFT JOIN employees e ON pr.employee_id = e.id").
		Where("e.team_id = ?", teamID).
		Select("AVG(score) as average_score, COUNT(id) as total_reviews").
		Scan(&stats).Error; err != nil {
		logger.Error("failed to get employee performance stats", "error", err)
		return nil, errors.New("failed to get employee performance stats")
	}

	// Get high performers (score >= 4.0)
	if err := s.db.Model(&domain.PerformanceReview{}).
		Joins("pr LEFT JOIN employees e ON pr.employee_id = e.id").
		Where("e.team_id = ? AND score >= ?", teamID, 4.0).
		Count(&stats.HighPerformers).Error; err != nil {
		logger.Error("failed to count high performers", "error", err)
		return nil, errors.New("failed to get employee performance stats")
	}

	// Get low performers (score <= 2.0)
	if err := s.db.Model(&domain.PerformanceReview{}).
		Joins("pr LEFT JOIN employees e ON pr.employee_id = e.id").
		Where("e.team_id = ? AND score <= ?", teamID, 2.0).
		Count(&stats.LowPerformers).Error; err != nil {
		logger.Error("failed to count low performers", "error", err)
		return nil, errors.New("failed to get employee performance stats")
	}

	return &stats, nil
}

// GetTerminationAnalysis retrieves termination analysis
func (s *AnalyticsService) GetTerminationAnalysis(ctx context.Context, teamID string) (*TerminationAnalysis, error) {
	var analysis TerminationAnalysis

	// Get total terminations
	if err := s.db.Model(&domain.TerminationRecord{}).
		Joins("tr LEFT JOIN employees e ON tr.employee_id = e.id").
		Where("e.team_id = ?", teamID).
		Count(&analysis.TotalTerminations).Error; err != nil {
		logger.Error("failed to count terminations", "error", err)
		return nil, errors.New("failed to get termination analysis")
	}

	// Get termination reasons
	if err := s.db.Model(&domain.TerminationRecord{}).
		Select("reason, count(id) as count").
		Joins("tr LEFT JOIN employees e ON tr.employee_id = e.id").
		Where("e.team_id = ?", teamID).
		Group("reason").
		Scan(&analysis.TerminationReasons).Error; err != nil {
		logger.Error("failed to get termination reasons", "error", err)
		return nil, errors.New("failed to get termination analysis")
	}

	return &analysis, nil
}

// GetEmployeeTurnoverRate retrieves employee turnover rate
func (s *AnalyticsService) GetEmployeeTurnoverRate(ctx context.Context, teamID string) (*EmployeeTurnoverRate, error) {
	var rate EmployeeTurnoverRate
	var tempCount int64

	// Get total employees
	if err := s.db.Model(&domain.Employee{}).
		Where("team_id = ?", teamID).
		Count(&rate.TotalEmployees).Error; err != nil {
		logger.Error("failed to count total employees", "error", err)
		return nil, errors.New("failed to get employee turnover rate")
	}

	// Get termination count for the last year
	oneYearAgo := time.Now().AddDate(-1, 0, 0)
	if err := s.db.Model(&domain.TerminationRecord{}).
		Joins("tr LEFT JOIN employees e ON tr.employee_id = e.id").
		Where("e.team_id = ? AND tr.termination_date >= ?", teamID, oneYearAgo).
		Count(&tempCount).Error; err != nil {
		logger.Error("failed to count terminations", "error", err)
		return nil, errors.New("failed to get employee turnover rate")
	}
	rate.TurnoverRate = float64(tempCount)

	// Calculate turnover rate as percentage
	if rate.TotalEmployees > 0 {
		rate.TurnoverRate = (rate.TurnoverRate / float64(rate.TotalEmployees)) * 100
	}

	// Get promotion count for the last year
	if err := s.db.Model(&EmployeeLifecycleEvent{}).
		Joins("e LEFT JOIN employees emp ON e.employee_id = emp.id").
		Where("emp.team_id = ? AND e.event_type = ? AND e.created_at >= ?", teamID, "promotion", oneYearAgo).
		Count(&tempCount).Error; err != nil {
		logger.Error("failed to count promotions", "error", err)
		return nil, errors.New("failed to get employee turnover rate")
	}
	rate.PromotionRate = float64(tempCount)

	// Calculate promotion rate as percentage
	if rate.TotalEmployees > 0 {
		rate.PromotionRate = (rate.PromotionRate / float64(rate.TotalEmployees)) * 100
	}

	// Get transfer count for the last year
	if err := s.db.Model(&EmployeeLifecycleEvent{}).
		Joins("e LEFT JOIN employees emp ON e.employee_id = emp.id").
		Where("emp.team_id = ? AND e.event_type = ? AND e.created_at >= ?", teamID, "transfer", oneYearAgo).
		Count(&tempCount).Error; err != nil {
		logger.Error("failed to count transfers", "error", err)
		return nil, errors.New("failed to get employee turnover rate")
	}
	rate.TransferRate = float64(tempCount)

	// Calculate transfer rate as percentage
	if rate.TotalEmployees > 0 {
		rate.TransferRate = (rate.TransferRate / float64(rate.TotalEmployees)) * 100
	}

	return &rate, nil
}

// GetSurveyAnalysis retrieves survey analysis
func (s *AnalyticsService) GetSurveyAnalysis(ctx context.Context, teamID string) (*SurveyAnalysis, error) {
	var analysis SurveyAnalysis

	// Get total surveys
	if err := s.db.Model(&domain.EmployeeSurvey{}).
		Where("created_by IN (SELECT user_id FROM employees WHERE team_id = ?)", teamID).
		Count(&analysis.TotalSurveys).Error; err != nil {
		logger.Error("failed to count surveys", "error", err)
		return nil, errors.New("failed to get survey analysis")
	}

	// Get total responses
	if err := s.db.Model(&domain.SurveyResponse{}).
		Joins("sr LEFT JOIN employee_surveys es ON sr.survey_id = es.id").
		Where("es.created_by IN (SELECT user_id FROM employees WHERE team_id = ?)", teamID).
		Count(&analysis.TotalResponses).Error; err != nil {
		logger.Error("failed to count survey responses", "error", err)
		return nil, errors.New("failed to get survey analysis")
	}

	// Calculate response rate
	if analysis.TotalSurveys > 0 {
		analysis.ResponseRate = (float64(analysis.TotalResponses) / float64(analysis.TotalSurveys)) * 100
	}

	// Get average survey score (simplified implementation)
	// In a real implementation, this would depend on how survey responses are scored
	analysis.AverageScore = 0.0

	return &analysis, nil
}