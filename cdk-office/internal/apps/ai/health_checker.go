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

package ai

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/linux-do/cdk-office/internal/models"
	"gorm.io/gorm"
)

// ServiceHealthChecker 服务健康检查器
type ServiceHealthChecker struct {
	db             *gorm.DB
	serviceManager *ServiceManager
	checkers       map[string]ServiceChecker
	mutex          sync.RWMutex
}

// ServiceChecker 服务检查接口
type ServiceChecker interface {
	CheckHealth(config *models.AIServiceConfig) (*HealthCheckResult, error)
	GetServiceType() string
}

// HealthCheckResult 健康检查结果
type HealthCheckResult struct {
	Status       string                 `json:"status"`
	ResponseTime time.Duration          `json:"response_time"`
	Error        error                  `json:"error,omitempty"`
	Details      map[string]interface{} `json:"details,omitempty"`
}

// NewServiceHealthChecker 创建新的服务健康检查器
func NewServiceHealthChecker(db *gorm.DB, serviceManager *ServiceManager) *ServiceHealthChecker {
	checker := &ServiceHealthChecker{
		db:             db,
		serviceManager: serviceManager,
		checkers:       make(map[string]ServiceChecker),
	}

	// 注册检查器
	checker.RegisterChecker(&AIChatChecker{})
	checker.RegisterChecker(&AIEmbeddingChecker{})
	checker.RegisterChecker(&AITranslationChecker{})

	return checker
}

// RegisterChecker 注册服务检查器
func (h *ServiceHealthChecker) RegisterChecker(checker ServiceChecker) {
	h.mutex.Lock()
	defer h.mutex.Unlock()
	h.checkers[checker.GetServiceType()] = checker
}

// StartHealthCheckRoutine 启动健康检查例程
func (h *ServiceHealthChecker) StartHealthCheckRoutine() {
	ticker := time.NewTicker(5 * time.Minute) // 每5分钟检查一次
	defer ticker.Stop()

	log.Println("Service health check routine started")

	for {
		select {
		case <-ticker.C:
			h.checkAllServices()
		}
	}
}

// checkAllServices 检查所有服务
func (h *ServiceHealthChecker) checkAllServices() {
	services, err := h.serviceManager.GetServiceList("")
	if err != nil {
		log.Printf("Failed to get service list for health check: %v", err)
		return
	}

	for _, service := range services {
		if service.IsEnabled {
			go h.checkServiceHealth(service)
		}
	}
}

// checkServiceHealth 检查单个服务健康状态
func (h *ServiceHealthChecker) checkServiceHealth(config *models.AIServiceConfig) {
	start := time.Now()

	// 获取对应的检查器
	h.mutex.RLock()
	checker, exists := h.checkers[config.ServiceType]
	h.mutex.RUnlock()

	if !exists {
		log.Printf("No health checker found for service type: %s", config.ServiceType)
		return
	}

	// 执行健康检查
	result, err := checker.CheckHealth(config)
	if err != nil {
		log.Printf("Health check failed for service %s: %v", config.ServiceName, err)
		result = &HealthCheckResult{
			Status:       "unavailable",
			ResponseTime: time.Since(start),
			Error:        err,
		}
	}

	// 更新服务状态
	h.updateServiceStatus(config, result)

	// 如果服务不可用，触发降级
	if result.Status == "unavailable" {
		log.Printf("Service %s is unavailable, triggering fallback", config.ServiceName)
		if err := h.serviceManager.TriggerFallback(config.ID); err != nil {
			log.Printf("Failed to trigger fallback for service %s: %v", config.ServiceName, err)
		}
	}
}

// updateServiceStatus 更新服务状态
func (h *ServiceHealthChecker) updateServiceStatus(config *models.AIServiceConfig, result *HealthCheckResult) {
	status := &models.ServiceStatus{}
	h.db.Where("service_id = ?", config.ID).FirstOrCreate(status, models.ServiceStatus{
		ServiceID:   config.ID,
		ServiceType: config.ServiceType,
	})

	status.Status = result.Status
	status.ResponseTime = int64(result.ResponseTime.Milliseconds())
	status.LastCheckAt = time.Now()

	if result.Error != nil {
		status.ErrorCount++
		status.LastError = result.Error.Error()

		// 计算成功率
		var totalChecks int64
		h.db.Model(&models.ServiceStatus{}).Where("service_id = ?", config.ID).Count(&totalChecks)
		if totalChecks > 0 {
			successRate := float64(totalChecks-int64(status.ErrorCount)) / float64(totalChecks)
			status.SuccessRate = successRate
		}
	} else {
		// 重置错误计数（如果检查成功）
		if status.ErrorCount > 0 {
			status.ErrorCount = 0
			status.LastError = ""
		}
		status.SuccessRate = 1.0
	}

	h.db.Save(status)
}

// CheckServiceNow 立即检查指定服务
func (h *ServiceHealthChecker) CheckServiceNow(serviceID string) (*HealthCheckResult, error) {
	services, err := h.serviceManager.GetServiceList("")
	if err != nil {
		return nil, err
	}

	var targetService *models.AIServiceConfig
	for _, service := range services {
		if service.ID == serviceID {
			targetService = service
			break
		}
	}

	if targetService == nil {
		return nil, fmt.Errorf("service not found: %s", serviceID)
	}

	h.mutex.RLock()
	checker, exists := h.checkers[targetService.ServiceType]
	h.mutex.RUnlock()

	if !exists {
		return nil, fmt.Errorf("no health checker found for service type: %s", targetService.ServiceType)
	}

	return checker.CheckHealth(targetService)
}

// AIChatChecker AI对话服务检查器
type AIChatChecker struct{}

func (c *AIChatChecker) GetServiceType() string {
	return "ai_chat"
}

func (c *AIChatChecker) CheckHealth(config *models.AIServiceConfig) (*HealthCheckResult, error) {
	start := time.Now()

	// 创建客户端
	client := createClientForConfig(config)

	// 发送测试请求
	testReq := map[string]interface{}{
		"messages": []map[string]string{
			{"role": "user", "content": "ping"},
		},
		"max_tokens": 5,
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(config.Timeout)*time.Second)
	defer cancel()

	_, err := client.Chat(ctx, testReq)
	responseTime := time.Since(start)

	status := "healthy"
	if err != nil {
		if responseTime > time.Duration(config.Timeout)*time.Second {
			status = "degraded" // 响应慢但可用
		} else {
			status = "unavailable" // 完全不可用
		}
		return &HealthCheckResult{
			Status:       status,
			ResponseTime: responseTime,
			Error:        err,
		}, nil
	}

	// 根据响应时间判断服务状态
	if responseTime > time.Duration(config.Timeout/2)*time.Second {
		status = "degraded"
	}

	return &HealthCheckResult{
		Status:       status,
		ResponseTime: responseTime,
		Details: map[string]interface{}{
			"response_time_ms": responseTime.Milliseconds(),
			"threshold_ms":     config.Timeout * 1000,
		},
	}, nil
}

// AIEmbeddingChecker AI向量化服务检查器
type AIEmbeddingChecker struct{}

func (c *AIEmbeddingChecker) GetServiceType() string {
	return "ai_embedding"
}

func (c *AIEmbeddingChecker) CheckHealth(config *models.AIServiceConfig) (*HealthCheckResult, error) {
	start := time.Now()

	client := createClientForConfig(config)

	testReq := map[string]interface{}{
		"input": "test",
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(config.Timeout)*time.Second)
	defer cancel()

	_, err := client.Embedding(ctx, testReq)
	responseTime := time.Since(start)

	status := "healthy"
	if err != nil {
		if responseTime > time.Duration(config.Timeout)*time.Second {
			status = "degraded"
		} else {
			status = "unavailable"
		}
		return &HealthCheckResult{
			Status:       status,
			ResponseTime: responseTime,
			Error:        err,
		}, nil
	}

	if responseTime > time.Duration(config.Timeout/2)*time.Second {
		status = "degraded"
	}

	return &HealthCheckResult{
		Status:       status,
		ResponseTime: responseTime,
		Details: map[string]interface{}{
			"response_time_ms": responseTime.Milliseconds(),
		},
	}, nil
}

// AITranslationChecker AI翻译服务检查器
type AITranslationChecker struct{}

func (c *AITranslationChecker) GetServiceType() string {
	return "ai_translation"
}

func (c *AITranslationChecker) CheckHealth(config *models.AIServiceConfig) (*HealthCheckResult, error) {
	start := time.Now()

	client := createClientForConfig(config)

	testReq := map[string]interface{}{
		"text": "hello",
		"from": "en",
		"to":   "zh",
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(config.Timeout)*time.Second)
	defer cancel()

	_, err := client.Translate(ctx, testReq)
	responseTime := time.Since(start)

	status := "healthy"
	if err != nil {
		if responseTime > time.Duration(config.Timeout)*time.Second {
			status = "degraded"
		} else {
			status = "unavailable"
		}
		return &HealthCheckResult{
			Status:       status,
			ResponseTime: responseTime,
			Error:        err,
		}, nil
	}

	if responseTime > time.Duration(config.Timeout/2)*time.Second {
		status = "degraded"
	}

	return &HealthCheckResult{
		Status:       status,
		ResponseTime: responseTime,
		Details: map[string]interface{}{
			"response_time_ms": responseTime.Milliseconds(),
		},
	}, nil
}

// createClientForConfig 根据配置创建客户端
func createClientForConfig(config *models.AIServiceConfig) AIClient {
	switch config.Provider {
	case "openai":
		return NewOpenAIClient(config)
	case "baidu":
		return NewBaiduAIClient(config)
	case "tencent":
		return NewTencentAIClient(config)
	case "aliyun":
		return NewAliyunAIClient(config)
	default:
		return NewGenericAIClient(config)
	}
}
