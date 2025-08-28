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

package router

import (
	"context"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"

	"github.com/linux-do/cdk-office/internal/handler"
	"github.com/linux-do/cdk-office/internal/middleware"
	"github.com/linux-do/cdk-office/internal/service"
)

// SetupHealthCheckRoutes 设置健康检查相关路由
func SetupHealthCheckRoutes(router *gin.Engine, db *gorm.DB, logger *logrus.Logger) {
	// 创建健康检查服务和处理器
	healthChecker := service.NewServiceHealthChecker(db, logger)
	healthHandler := handler.NewHealthCheckHandler(healthChecker, logger)

	// 管理员API路由组
	adminAPI := router.Group("/api/admin")
	{
		// 应用权限中间件（需要超级管理员权限）
		adminAPI.Use(middleware.RequireSuperAdmin())

		// 服务状态相关路由
		serviceStatus := adminAPI.Group("/service-status")
		{
			// 获取所有服务状态
			serviceStatus.GET("", healthHandler.GetServiceStatus)

			// 获取指定服务状态
			serviceStatus.GET("/:service_name", healthHandler.GetServiceStatusByName)

			// 获取健康状态摘要
			serviceStatus.GET("/summary", healthHandler.GetHealthSummary)

			// 手动触发健康检查
			serviceStatus.POST("/check", healthHandler.TriggerHealthCheck)

			// 清理旧记录
			serviceStatus.DELETE("/cleanup", healthHandler.CleanupOldRecords)
		}
	}

	// 公共健康检查端点（不需要认证）
	publicHealth := router.Group("/health")
	{
		// 基础健康检查
		publicHealth.GET("", func(c *gin.Context) {
			c.JSON(200, gin.H{
				"status":    "ok",
				"timestamp": "2025-01-20T12:00:00Z",
				"service":   "cdk-office",
				"version":   "1.0.0",
			})
		})

		// 就绪检查
		publicHealth.GET("/ready", func(c *gin.Context) {
			// 检查数据库连接
			sqlDB, err := db.DB()
			if err != nil {
				c.JSON(503, gin.H{
					"status": "not ready",
					"reason": "database connection failed",
				})
				return
			}

			if err := sqlDB.Ping(); err != nil {
				c.JSON(503, gin.H{
					"status": "not ready",
					"reason": "database ping failed",
				})
				return
			}

			c.JSON(200, gin.H{
				"status": "ready",
			})
		})

		// 存活检查
		publicHealth.GET("/live", func(c *gin.Context) {
			c.JSON(200, gin.H{
				"status": "alive",
			})
		})
	}

	logger.Info("健康检查路由设置完成")
}

// SetupPeriodicHealthCheck 设置定期健康检查
func SetupPeriodicHealthCheck(db *gorm.DB, logger *logrus.Logger) *service.ServiceHealthChecker {
	healthChecker := service.NewServiceHealthChecker(db, logger)

	// 在后台启动定期健康检查（每5分钟执行一次）
	go func() {
		ctx := context.Background()
		healthChecker.StartPeriodicHealthCheck(ctx, 5*time.Minute)
	}()

	logger.Info("定期健康检查已启动，间隔：5分钟")
	return healthChecker
}
