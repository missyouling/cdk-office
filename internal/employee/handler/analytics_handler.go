package handler

import (
	"net/http"

	"cdk-office/internal/employee/service"
	"github.com/gin-gonic/gin"
)

// AnalyticsHandlerInterface defines the interface for employee analytics handler
type AnalyticsHandlerInterface interface {
	GetEmployeeCountByDepartment(c *gin.Context)
	GetEmployeeCountByPosition(c *gin.Context)
	GetEmployeeLifecycleStats(c *gin.Context)
	GetEmployeeAgeDistribution(c *gin.Context)
	GetEmployeePerformanceStats(c *gin.Context)
	GetTerminationAnalysis(c *gin.Context)
	GetEmployeeTurnoverRate(c *gin.Context)
	GetSurveyAnalysis(c *gin.Context)
}

// AnalyticsHandler implements the AnalyticsHandlerInterface
type AnalyticsHandler struct {
	analyticsService service.AnalyticsServiceInterface
}

// NewAnalyticsHandler creates a new instance of AnalyticsHandler
func NewAnalyticsHandler() *AnalyticsHandler {
	return &AnalyticsHandler{
		analyticsService: service.NewAnalyticsService(),
	}
}

// GetEmployeeCountByDepartment handles retrieving employee count by department
func (h *AnalyticsHandler) GetEmployeeCountByDepartment(c *gin.Context) {
	teamID := c.Query("team_id")
	if teamID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "team id is required"})
		return
	}

	// Call service to get employee count by department
	results, err := h.analyticsService.GetEmployeeCountByDepartment(c.Request.Context(), teamID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, results)
}

// GetEmployeeCountByPosition handles retrieving employee count by position
func (h *AnalyticsHandler) GetEmployeeCountByPosition(c *gin.Context) {
	teamID := c.Query("team_id")
	if teamID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "team id is required"})
		return
	}

	// Call service to get employee count by position
	results, err := h.analyticsService.GetEmployeeCountByPosition(c.Request.Context(), teamID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, results)
}

// GetEmployeeLifecycleStats handles retrieving employee lifecycle statistics
func (h *AnalyticsHandler) GetEmployeeLifecycleStats(c *gin.Context) {
	teamID := c.Query("team_id")
	if teamID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "team id is required"})
		return
	}

	// Call service to get employee lifecycle stats
	stats, err := h.analyticsService.GetEmployeeLifecycleStats(c.Request.Context(), teamID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, stats)
}

// GetEmployeeAgeDistribution handles retrieving employee age distribution
func (h *AnalyticsHandler) GetEmployeeAgeDistribution(c *gin.Context) {
	teamID := c.Query("team_id")
	if teamID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "team id is required"})
		return
	}

	// Call service to get employee age distribution
	results, err := h.analyticsService.GetEmployeeAgeDistribution(c.Request.Context(), teamID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, results)
}

// GetEmployeePerformanceStats handles retrieving employee performance statistics
func (h *AnalyticsHandler) GetEmployeePerformanceStats(c *gin.Context) {
	teamID := c.Query("team_id")
	if teamID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "team id is required"})
		return
	}

	// Call service to get employee performance stats
	stats, err := h.analyticsService.GetEmployeePerformanceStats(c.Request.Context(), teamID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, stats)
}

// GetTerminationAnalysis handles retrieving termination analysis
func (h *AnalyticsHandler) GetTerminationAnalysis(c *gin.Context) {
	teamID := c.Query("team_id")
	if teamID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "team id is required"})
		return
	}

	// Call service to get termination analysis
	analysis, err := h.analyticsService.GetTerminationAnalysis(c.Request.Context(), teamID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, analysis)
}

// GetEmployeeTurnoverRate handles retrieving employee turnover rate
func (h *AnalyticsHandler) GetEmployeeTurnoverRate(c *gin.Context) {
	teamID := c.Query("team_id")
	if teamID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "team id is required"})
		return
	}

	// Call service to get employee turnover rate
	rate, err := h.analyticsService.GetEmployeeTurnoverRate(c.Request.Context(), teamID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, rate)
}

// GetSurveyAnalysis handles retrieving survey analysis
func (h *AnalyticsHandler) GetSurveyAnalysis(c *gin.Context) {
	teamID := c.Query("team_id")
	if teamID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "team id is required"})
		return
	}

	// Call service to get survey analysis
	analysis, err := h.analyticsService.GetSurveyAnalysis(c.Request.Context(), teamID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, analysis)
}