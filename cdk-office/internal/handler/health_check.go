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

package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	"github.com/linux-do/cdk-office/internal/models"
	"github.com/linux-do/cdk-office/internal/service"
)

// HealthCheckHandler 健康检查处理器
type HealthCheckHandler struct {
	healthChecker *service.ServiceHealthChecker
	logger        *logrus.Logger
}

// NewHealthCheckHandler 创建健康检查处理器
func NewHealthCheckHandler(healthChecker *service.ServiceHealthChecker, logger *logrus.Logger) *HealthCheckHandler {
	return &HealthCheckHandler{
		healthChecker: healthChecker,
		logger:        logger,
	}
}

// GetServiceStatus 获取所有服务状态
// @Summary 获取所有服务健康状态
// @Description 返回系统中所有服务的当前健康状态信息
// @Tags 系统管理
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{} "成功返回服务状态"
// @Failure 500 {object} map[string]interface{} "内部服务器错误"
// @Router /api/admin/service-status [get]
// @Security BearerAuth
func (h *HealthCheckHandler) GetServiceStatus(c *gin.Context) {
	ctx := c.Request.Context()

	h.logger.WithFields(logrus.Fields{
		"action":    "get_service_status",
		"client_ip": c.ClientIP(),
	}).Info("获取服务状态请求")

	statuses, err := h.healthChecker.GetAllServiceStatuses(ctx)
	if err != nil {
		h.logger.WithError(err).Error("获取服务状态失败")
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "获取服务状态失败",
			"message": err.Error(),
		})
		return
	}

	// 计算总体健康状态
	summary := h.calculateServiceSummary(statuses)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"services": statuses,
			"summary":  summary,
		},
		"timestamp": time.Now(),
	})
}

// GetServiceStatusByName 获取指定服务的状态
// @Summary 获取指定服务健康状态
// @Description 返回指定服务的当前健康状态信息
// @Tags 系统管理
// @Accept json
// @Produce json
// @Param service_name path string true "服务名称"
// @Success 200 {object} map[string]interface{} "成功返回服务状态"
// @Failure 404 {object} map[string]interface{} "服务未找到"
// @Failure 500 {object} map[string]interface{} "内部服务器错误"
// @Router /api/admin/service-status/{service_name} [get]
// @Security BearerAuth
func (h *HealthCheckHandler) GetServiceStatusByName(c *gin.Context) {
	ctx := c.Request.Context()
	serviceName := c.Param("service_name")

	h.logger.WithFields(logrus.Fields{
		"action":       "get_service_status_by_name",
		"service_name": serviceName,
		"client_ip":    c.ClientIP(),
	}).Info("获取特定服务状态请求")

	status, err := h.healthChecker.GetServiceStatus(ctx, serviceName)
	if err != nil {
		h.logger.WithError(err).WithField("service_name", serviceName).Error("获取服务状态失败")
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "服务状态未找到",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    status,
	})
}

// TriggerHealthCheck 手动触发健康检查
// @Summary 手动触发健康检查
// @Description 立即执行一次完整的服务健康检查
// @Tags 系统管理
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{} "健康检查执行成功"
// @Failure 500 {object} map[string]interface{} "健康检查执行失败"
// @Router /api/admin/service-status/check [post]
// @Security BearerAuth
func (h *HealthCheckHandler) TriggerHealthCheck(c *gin.Context) {
	ctx := c.Request.Context()

	h.logger.WithFields(logrus.Fields{
		"action":    "trigger_health_check",
		"client_ip": c.ClientIP(),
	}).Info("手动触发健康检查")

	results, err := h.healthChecker.CheckAllServices(ctx)
	if err != nil {
		h.logger.WithError(err).Error("手动健康检查执行失败")
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "健康检查执行失败",
			"message": err.Error(),
		})
		return
	}

	// 统计检查结果
	summary := h.calculateCheckSummary(results)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "健康检查执行完成",
		"data": gin.H{
			"results": results,
			"summary": summary,
		},
		"timestamp": time.Now(),
	})
}

// CleanupOldRecords 清理旧的健康检查记录
// @Summary 清理旧的健康检查记录
// @Description 删除指定天数之前的健康检查记录
// @Tags 系统管理
// @Accept json
// @Produce json
// @Param days query int false "保留天数" default(30)
// @Success 200 {object} map[string]interface{} "清理成功"
// @Failure 400 {object} map[string]interface{} "参数错误"
// @Failure 500 {object} map[string]interface{} "清理失败"
// @Router /api/admin/service-status/cleanup [delete]
// @Security BearerAuth
func (h *HealthCheckHandler) CleanupOldRecords(c *gin.Context) {
	ctx := c.Request.Context()

	// 获取保留天数参数，默认30天
	keepDaysStr := c.DefaultQuery("days", "30")
	keepDays, err := strconv.Atoi(keepDaysStr)
	if err != nil || keepDays <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "无效的天数参数",
			"message": "天数必须是正整数",
		})
		return
	}

	h.logger.WithFields(logrus.Fields{
		"action":    "cleanup_old_records",
		"keep_days": keepDays,
		"client_ip": c.ClientIP(),
	}).Info("清理旧的健康检查记录")

	err = h.healthChecker.CleanupOldRecords(ctx, keepDays)
	if err != nil {
		h.logger.WithError(err).Error("清理旧记录失败")
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "清理记录失败",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "旧记录清理完成",
		"data": gin.H{
			"keep_days": keepDays,
		},
	})
}

// GetHealthSummary 获取健康状态摘要
// @Summary 获取系统健康状态摘要
// @Description 返回系统整体健康状态的简要摘要信息
// @Tags 系统管理
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{} "成功返回健康摘要"
// @Failure 500 {object} map[string]interface{} "内部服务器错误"
// @Router /api/admin/service-status/summary [get]
// @Security BearerAuth
func (h *HealthCheckHandler) GetHealthSummary(c *gin.Context) {
	ctx := c.Request.Context()

	statuses, err := h.healthChecker.GetAllServiceStatuses(ctx)
	if err != nil {
		h.logger.WithError(err).Error("获取服务状态摘要失败")
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "获取健康摘要失败",
			"message": err.Error(),
		})
		return
	}

	summary := h.calculateServiceSummary(statuses)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    summary,
	})
}

// calculateServiceSummary 计算服务状态摘要
func (h *HealthCheckHandler) calculateServiceSummary(statuses []models.ServiceHealthStatus) map[string]interface{} {
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

// calculateCheckSummary 计算检查结果摘要
func (h *HealthCheckHandler) calculateCheckSummary(results []service.HealthCheckResult) map[string]interface{} {
	var healthyCount, degradedCount, unhealthyCount int
	var totalResponseTime time.Duration
	var slowServices []string

	for _, result := range results {
		totalResponseTime += result.ResponseTime

		switch result.Status {
		case "healthy":
			healthyCount++
		case "degraded":
			degradedCount++
		case "unhealthy":
			unhealthyCount++
		}

		// 记录响应慢的服务
		if result.ResponseTime > 500*time.Millisecond {
			slowServices = append(slowServices, result.ServiceName)
		}
	}

	avgResponseTime := time.Duration(0)
	if len(results) > 0 {
		avgResponseTime = totalResponseTime / time.Duration(len(results))
	}

	return map[string]interface{}{
		"total_checked":        len(results),
		"healthy_count":        healthyCount,
		"degraded_count":       degradedCount,
		"unhealthy_count":      unhealthyCount,
		"slow_services":        slowServices,
		"avg_response_time_ms": avgResponseTime.Milliseconds(),
		"check_completed_at":   time.Now(),
	}
}
