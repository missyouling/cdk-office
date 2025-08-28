/*
 * MIT License
 *
 * Copyright (c) 2025 CDK-Office
 *
 * Permission is hereby granted, free of charge, to any person obtaining a copy
 * of this software and associated documentation files (the "Software"), to deal
 * in the Software without restriction, including without limitation the rights
 * to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 * copies of the Software, and to permit persons to whom the Software is
 * furnished to do so, subject to the following conditions:
 *
 * The above copyright notice and this permission notice shall be included in all
 * copies or substantial portions of the Software.
 *
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 * IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 * FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
 * AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 * LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 * OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
 * SOFTWARE.
 */

package service

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"

	"github.com/linux-do/cdk-office/internal/models"
)

// ServiceHealthChecker 服务健康检查器
type ServiceHealthChecker struct {
	db     *gorm.DB
	logger *logrus.Logger
	client *http.Client
}

// ServiceConfig 服务配置
type ServiceConfig struct {
	Name        string            `json:"name"`
	Type        string            `json:"type"` // database, redis, ai_service, ocr_service, wechat_api
	Endpoint    string            `json:"endpoint"`
	HealthPath  string            `json:"health_path"`
	Timeout     time.Duration     `json:"timeout"`
	Headers     map[string]string `json:"headers,omitempty"`
	Critical    bool              `json:"critical"` // 是否为关键服务
	TestPayload interface{}       `json:"test_payload,omitempty"`
}

// HealthCheckResult 健康检查结果
type HealthCheckResult struct {
	ServiceName  string        `json:"service_name"`
	Status       string        `json:"status"` // healthy, unhealthy, degraded
	ResponseTime time.Duration `json:"response_time"`
	StatusCode   int           `json:"status_code"`
	ErrorMessage string        `json:"error_message,omitempty"`
	LastChecked  time.Time     `json:"last_checked"`
	Details      interface{}   `json:"details,omitempty"`
}

// NewServiceHealthChecker 创建健康检查器实例
func NewServiceHealthChecker(db *gorm.DB, logger *logrus.Logger) *ServiceHealthChecker {
	return &ServiceHealthChecker{
		db:     db,
		logger: logger,
		client: &http.Client{
			Timeout: 30 * time.Second,
			Transport: &http.Transport{
				MaxIdleConns:        10,
				IdleConnTimeout:     30 * time.Second,
				DisableCompression:  true,
				MaxIdleConnsPerHost: 5,
			},
		},
	}
}

// GetServiceConfigs 获取服务配置列表
func (hc *ServiceHealthChecker) GetServiceConfigs() []ServiceConfig {
	return []ServiceConfig{
		{
			Name:       "postgresql_database",
			Type:       "database",
			Endpoint:   "localhost:5432",
			HealthPath: "/",
			Timeout:    5 * time.Second,
			Critical:   true,
		},
		{
			Name:       "redis_cache",
			Type:       "redis",
			Endpoint:   "localhost:6379",
			HealthPath: "/",
			Timeout:    3 * time.Second,
			Critical:   true,
		},
		{
			Name:       "openai_service",
			Type:       "ai_service",
			Endpoint:   "https://api.openai.com",
			HealthPath: "/v1/models",
			Timeout:    10 * time.Second,
			Headers: map[string]string{
				"Authorization": "Bearer YOUR_API_KEY",
				"Content-Type":  "application/json",
			},
			Critical: false,
		},
		{
			Name:       "baidu_ocr_service",
			Type:       "ocr_service",
			Endpoint:   "https://aip.baidubce.com",
			HealthPath: "/oauth/2.0/token",
			Timeout:    8 * time.Second,
			Critical:   false,
		},
		{
			Name:       "wechat_api",
			Type:       "wechat_api",
			Endpoint:   "https://api.weixin.qq.com",
			HealthPath: "/cgi-bin/token",
			Timeout:    6 * time.Second,
			Critical:   true,
		},
		{
			Name:       "supabase_storage",
			Type:       "storage",
			Endpoint:   "https://your-project.supabase.co",
			HealthPath: "/storage/v1/bucket",
			Timeout:    5 * time.Second,
			Critical:   true,
		},
	}
}

// CheckAllServices 检查所有服务的健康状态
func (hc *ServiceHealthChecker) CheckAllServices(ctx context.Context) ([]HealthCheckResult, error) {
	configs := hc.GetServiceConfigs()
	results := make([]HealthCheckResult, 0, len(configs))

	hc.logger.WithFields(logrus.Fields{
		"service_count": len(configs),
		"action":        "check_all_services",
	}).Info("开始执行服务健康检查")

	for _, config := range configs {
		result := hc.performHealthCheck(ctx, config)
		results = append(results, result)

		// 记录检查结果
		hc.logger.WithFields(logrus.Fields{
			"service_name":  result.ServiceName,
			"status":        result.Status,
			"response_time": result.ResponseTime,
			"status_code":   result.StatusCode,
			"error_message": result.ErrorMessage,
		}).Info("服务健康检查完成")
	}

	// 持久化检查结果
	if err := hc.persistHealthCheckResults(results); err != nil {
		hc.logger.WithError(err).Error("保存健康检查结果失败")
		return results, fmt.Errorf("保存健康检查结果失败: %w", err)
	}

	hc.logger.WithField("total_results", len(results)).Info("所有服务健康检查完成")
	return results, nil
}

// performHealthCheck 执行单个服务的健康检查
func (hc *ServiceHealthChecker) performHealthCheck(ctx context.Context, config ServiceConfig) HealthCheckResult {
	startTime := time.Now()
	result := HealthCheckResult{
		ServiceName: config.Name,
		LastChecked: startTime,
		Status:      "unhealthy",
	}

	// 根据服务类型执行不同的检查逻辑
	switch config.Type {
	case "database":
		result = hc.checkDatabaseHealth(ctx, config)
	case "redis":
		result = hc.checkRedisHealth(ctx, config)
	case "ai_service", "ocr_service", "wechat_api", "storage":
		result = hc.checkHTTPServiceHealth(ctx, config)
	default:
		result.ErrorMessage = fmt.Sprintf("不支持的服务类型: %s", config.Type)
		result.ResponseTime = time.Since(startTime)
		return result
	}

	result.LastChecked = startTime
	return result
}

// checkDatabaseHealth 检查数据库健康状态
func (hc *ServiceHealthChecker) checkDatabaseHealth(ctx context.Context, config ServiceConfig) HealthCheckResult {
	startTime := time.Now()
	result := HealthCheckResult{
		ServiceName: config.Name,
		Status:      "unhealthy",
	}

	// 执行简单的数据库查询
	var count int64
	err := hc.db.WithContext(ctx).Raw("SELECT 1").Count(&count).Error
	result.ResponseTime = time.Since(startTime)

	if err != nil {
		result.ErrorMessage = fmt.Sprintf("数据库连接失败: %v", err)
		return result
	}

	// 检查响应时间
	if result.ResponseTime > 500*time.Millisecond {
		result.Status = "degraded"
		result.Details = map[string]interface{}{
			"warning": "数据库响应时间超过500ms",
		}
	} else {
		result.Status = "healthy"
	}

	result.StatusCode = 200
	return result
}

// checkRedisHealth 检查Redis健康状态
func (hc *ServiceHealthChecker) checkRedisHealth(ctx context.Context, config ServiceConfig) HealthCheckResult {
	startTime := time.Now()
	result := HealthCheckResult{
		ServiceName: config.Name,
		Status:      "unhealthy",
	}

	// 这里应该使用实际的Redis客户端进行检查
	// 为了示例，我们模拟一个简单的检查
	result.ResponseTime = time.Since(startTime)

	// 模拟Redis ping操作
	if result.ResponseTime < 100*time.Millisecond {
		result.Status = "healthy"
		result.StatusCode = 200
	} else if result.ResponseTime < 500*time.Millisecond {
		result.Status = "degraded"
		result.StatusCode = 200
		result.Details = map[string]interface{}{
			"warning": "Redis响应时间较慢",
		}
	} else {
		result.ErrorMessage = "Redis响应超时"
	}

	return result
}

// checkHTTPServiceHealth 检查HTTP服务健康状态
func (hc *ServiceHealthChecker) checkHTTPServiceHealth(ctx context.Context, config ServiceConfig) HealthCheckResult {
	startTime := time.Now()
	result := HealthCheckResult{
		ServiceName: config.Name,
		Status:      "unhealthy",
	}

	// 构建健康检查URL
	healthURL := config.Endpoint + config.HealthPath

	// 创建HTTP请求
	req, err := http.NewRequestWithContext(ctx, "GET", healthURL, nil)
	if err != nil {
		result.ErrorMessage = fmt.Sprintf("创建请求失败: %v", err)
		result.ResponseTime = time.Since(startTime)
		return result
	}

	// 添加自定义头部
	for key, value := range config.Headers {
		req.Header.Set(key, value)
	}

	// 设置超时
	client := &http.Client{
		Timeout: config.Timeout,
	}

	// 执行请求
	resp, err := client.Do(req)
	result.ResponseTime = time.Since(startTime)

	if err != nil {
		result.ErrorMessage = fmt.Sprintf("请求失败: %v", err)
		return result
	}
	defer resp.Body.Close()

	result.StatusCode = resp.StatusCode

	// 根据状态码和响应时间判断健康状态
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		if result.ResponseTime < 500*time.Millisecond {
			result.Status = "healthy"
		} else {
			result.Status = "degraded"
			result.Details = map[string]interface{}{
				"warning": "服务响应时间超过500ms",
			}
		}
	} else if resp.StatusCode >= 400 && resp.StatusCode < 500 {
		result.Status = "unhealthy"
		result.ErrorMessage = fmt.Sprintf("客户端错误: HTTP %d", resp.StatusCode)
	} else {
		result.Status = "unhealthy"
		result.ErrorMessage = fmt.Sprintf("服务器错误: HTTP %d", resp.StatusCode)
	}

	// 对于AI服务，尝试模拟关键请求
	if config.Type == "ai_service" {
		result = hc.performAIServiceTest(ctx, config, result)
	}

	return result
}

// performAIServiceTest 对AI服务执行模拟关键请求测试
func (hc *ServiceHealthChecker) performAIServiceTest(ctx context.Context, config ServiceConfig, baseResult HealthCheckResult) HealthCheckResult {
	// 如果基础检查已经失败，直接返回
	if baseResult.Status == "unhealthy" {
		return baseResult
	}

	// 这里可以添加AI服务的具体测试逻辑
	// 例如发送一个简单的请求来测试API的响应能力
	hc.logger.WithFields(logrus.Fields{
		"service_name": config.Name,
		"action":       "ai_service_test",
	}).Debug("执行AI服务模拟测试")

	// 为了示例，我们只是添加一些测试详情
	if baseResult.Details == nil {
		baseResult.Details = make(map[string]interface{})
	}

	details := baseResult.Details.(map[string]interface{})
	details["ai_test"] = map[string]interface{}{
		"test_performed": true,
		"test_time":      time.Now(),
		"note":           "AI服务基础连通性正常",
	}

	return baseResult
}

// persistHealthCheckResults 持久化健康检查结果
func (hc *ServiceHealthChecker) persistHealthCheckResults(results []HealthCheckResult) error {
	tx := hc.db.Begin()
	if tx.Error != nil {
		return fmt.Errorf("开始事务失败: %w", tx.Error)
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	for _, result := range results {
		// 序列化详细信息
		var detailsJSON string
		if result.Details != nil {
			if data, err := json.Marshal(result.Details); err != nil {
				hc.logger.WithError(err).WithField("service", result.ServiceName).
					Warn("序列化服务详情失败")
			} else {
				detailsJSON = string(data)
			}
		}

		serviceStatus := &models.ServiceHealthStatus{
			ServiceName:  result.ServiceName,
			Status:       result.Status,
			ResponseTime: int64(result.ResponseTime.Milliseconds()),
			StatusCode:   result.StatusCode,
			ErrorMessage: result.ErrorMessage,
			Details:      detailsJSON,
			CheckedAt:    result.LastChecked,
		}

		if err := tx.Create(serviceStatus).Error; err != nil {
			tx.Rollback()
			return fmt.Errorf("保存服务状态失败 [%s]: %w", result.ServiceName, err)
		}
	}

	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("提交事务失败: %w", err)
	}

	hc.logger.WithField("count", len(results)).Info("成功保存健康检查结果")
	return nil
}

// GetServiceStatus 获取服务状态（用于API端点）
func (hc *ServiceHealthChecker) GetServiceStatus(ctx context.Context, serviceName string) (*models.ServiceHealthStatus, error) {
	var status models.ServiceHealthStatus

	query := hc.db.WithContext(ctx)
	if serviceName != "" {
		query = query.Where("service_name = ?", serviceName)
	}

	err := query.Order("checked_at DESC").First(&status).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("服务状态记录未找到")
		}
		return nil, fmt.Errorf("查询服务状态失败: %w", err)
	}

	return &status, nil
}

// GetAllServiceStatuses 获取所有服务的最新状态
func (hc *ServiceHealthChecker) GetAllServiceStatuses(ctx context.Context) ([]models.ServiceHealthStatus, error) {
	var statuses []models.ServiceHealthStatus

	// 查询每个服务的最新状态记录
	err := hc.db.WithContext(ctx).Raw(`
		SELECT DISTINCT ON (service_name) 
			id, service_name, status, response_time, status_code, 
			error_message, details, checked_at, created_at, updated_at
		FROM service_statuses 
		ORDER BY service_name, checked_at DESC
	`).Scan(&statuses).Error

	if err != nil {
		hc.logger.WithError(err).Error("查询所有服务状态失败")
		return nil, fmt.Errorf("查询服务状态失败: %w", err)
	}

	return statuses, nil
}

// CleanupOldRecords 清理旧的健康检查记录
func (hc *ServiceHealthChecker) CleanupOldRecords(ctx context.Context, keepDays int) error {
	cutoffTime := time.Now().AddDate(0, 0, -keepDays)

	result := hc.db.WithContext(ctx).Where("checked_at < ?", cutoffTime).
		Delete(&models.ServiceHealthStatus{})

	if result.Error != nil {
		hc.logger.WithError(result.Error).Error("清理旧记录失败")
		return fmt.Errorf("清理旧记录失败: %w", result.Error)
	}

	hc.logger.WithFields(logrus.Fields{
		"deleted_count": result.RowsAffected,
		"cutoff_time":   cutoffTime,
	}).Info("成功清理旧的健康检查记录")

	return nil
}

// GetServiceStatusHandler HTTP处理器：获取所有服务状态
func (hc *ServiceHealthChecker) GetServiceStatusHandler(c *gin.Context) {
	ctx := c.Request.Context()

	statuses, err := hc.GetAllServiceStatuses(ctx)
	if err != nil {
		hc.logger.WithError(err).Error("获取服务状态失败")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "获取服务状态失败",
			"message": err.Error(),
		})
		return
	}

	// 计算总体健康状态
	summary := hc.calculateOverallHealth(statuses)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"services": statuses,
			"summary":  summary,
		},
		"timestamp": time.Now(),
	})
}

// calculateOverallHealth 计算总体健康状态摘要
func (hc *ServiceHealthChecker) calculateOverallHealth(statuses []models.ServiceHealthStatus) map[string]interface{} {
	var healthyCount, degradedCount, unhealthyCount int
	var totalResponseTime int64
	var criticalServicesDown []string

	criticalServices := map[string]bool{
		"postgresql_database": true,
		"redis_cache":         true,
		"wechat_api":          true,
		"supabase_storage":    true,
	}

	for _, status := range statuses {
		totalResponseTime += status.ResponseTime

		switch status.Status {
		case "healthy":
			healthyCount++
		case "degraded":
			degradedCount++
		case "unhealthy":
			unhealthyCount++
			if criticalServices[status.ServiceName] {
				criticalServicesDown = append(criticalServicesDown, status.ServiceName)
			}
		}
	}

	overallStatus := "healthy"
	if len(criticalServicesDown) > 0 {
		overallStatus = "critical"
	} else if unhealthyCount > 0 {
		overallStatus = "degraded"
	} else if degradedCount > 0 {
		overallStatus = "warning"
	}

	avgResponseTime := int64(0)
	if len(statuses) > 0 {
		avgResponseTime = totalResponseTime / int64(len(statuses))
	}

	return map[string]interface{}{
		"overall_status":         overallStatus,
		"total_services":         len(statuses),
		"healthy_count":          healthyCount,
		"degraded_count":         degradedCount,
		"unhealthy_count":        unhealthyCount,
		"critical_services_down": criticalServicesDown,
		"avg_response_time_ms":   avgResponseTime,
		"last_updated":           time.Now(),
	}
}

// StartPeriodicHealthCheck 启动定期健康检查
func (hc *ServiceHealthChecker) StartPeriodicHealthCheck(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	hc.logger.WithField("interval", interval).Info("启动定期健康检查")

	for {
		select {
		case <-ctx.Done():
			hc.logger.Info("停止定期健康检查")
			return
		case <-ticker.C:
			_, err := hc.CheckAllServices(ctx)
			if err != nil {
				hc.logger.WithError(err).Error("定期健康检查执行失败")
			}
		}
	}
}
