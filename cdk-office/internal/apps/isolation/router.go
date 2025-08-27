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

package isolation

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/linux-do/cdk-office/internal/middleware"
	"github.com/linux-do/cdk-office/internal/models"
	"github.com/linux-do/cdk-office/internal/services/isolation"
)

// IsolationRouter 数据隔离路由
type IsolationRouter struct {
	db               *gorm.DB
	handler          *IsolationHandler
	isolationService *isolation.DataIsolationService
}

// NewIsolationRouter 创建数据隔离路由
func NewIsolationRouter(db *gorm.DB, isolationService *isolation.DataIsolationService) *IsolationRouter {
	handler := NewIsolationHandler(db, isolationService)
	return &IsolationRouter{
		db:               db,
		handler:          handler,
		isolationService: isolationService,
	}
}

// RegisterRoutes 注册路由
func (r *IsolationRouter) RegisterRoutes(group *gin.RouterGroup, oauth middleware.OAuthMiddleware) {
	// 数据隔离中间件
	isolationMiddleware := middleware.NewDataIsolationMiddleware(r.isolationService, r.db)

	// 数据隔离管理路由 - 需要管理员权限
	isolation := group.Group("/isolation")
	isolation.Use(oauth.LoginRequired())
	{
		// 团队隔离策略管理（仅超级管理员）
		policies := isolation.Group("/policies")
		policies.Use(isolationMiddleware.CrossTeamAccessOnly())
		{
			policies.POST("", r.handler.CreateTeamIsolationPolicy)         // 创建团队隔离策略
			policies.GET("/:team_id", r.handler.GetTeamIsolationPolicy)    // 获取团队隔离策略
			policies.PUT("/:team_id", r.handler.UpdateTeamIsolationPolicy) // 更新团队隔离策略
		}

		// 跨团队访问申请
		requests := isolation.Group("/requests")
		{
			requests.POST("", r.handler.CreateCrossTeamAccessRequest)                     // 创建跨团队访问申请
			requests.GET("", r.handler.GetCrossTeamAccessRequests)                        // 获取跨团队访问申请列表
			requests.PUT("/:request_id/approve", r.handler.ApproveCrossTeamAccessRequest) // 审批跨团队访问申请
		}

		// 数据访问日志
		logs := isolation.Group("/logs")
		{
			logs.GET("/access", r.handler.GetDataAccessLogs) // 获取数据访问日志
			logs.GET("/violations", r.handler.GetViolations) // 获取违规记录
		}

		// 用户访问档案管理
		profiles := isolation.Group("/profiles")
		profiles.Use(isolationMiddleware.RequireTeamAccess("team_manager", "super_admin"))
		{
			profiles.GET("/:user_id", r.handler.GetUserAccessProfile)    // 获取用户访问档案
			profiles.PUT("/:user_id", r.handler.UpdateUserAccessProfile) // 更新用户访问档案
		}
	}

	// 应用数据隔离中间件到需要保护的路由
	r.applyIsolationMiddleware(group, isolationMiddleware, oauth)
}

// applyIsolationMiddleware 应用数据隔离中间件到各个模块
func (r *IsolationRouter) applyIsolationMiddleware(group *gin.RouterGroup, isolationMiddleware *middleware.DataIsolationMiddleware, oauth middleware.OAuthMiddleware) {
	// 知识库路由应用数据隔离
	knowledge := group.Group("/knowledge")
	knowledge.Use(oauth.LoginRequired())
	knowledge.Use(isolationMiddleware.TeamDataIsolation())
	knowledge.Use(isolationMiddleware.DataVisibilityFilter())
	{
		// 这里的路由会自动应用数据隔离检查
		// 具体的知识库路由在knowledge模块中定义
	}

	// 文档路由应用数据隔离
	documents := group.Group("/documents")
	documents.Use(oauth.LoginRequired())
	documents.Use(isolationMiddleware.TeamDataIsolation())
	documents.Use(isolationMiddleware.RateLimitByRole())
	{
		// 文档相关路由会自动应用数据隔离检查
	}

	// 问卷路由应用数据隔离
	surveys := group.Group("/surveys")
	surveys.Use(oauth.LoginRequired())
	surveys.Use(isolationMiddleware.TeamDataIsolation())
	{
		// 问卷相关路由会自动应用数据隔离检查
	}

	// PDF处理路由应用数据隔离
	pdf := group.Group("/pdf")
	pdf.Use(oauth.LoginRequired())
	pdf.Use(isolationMiddleware.TeamDataIsolation())
	{
		// PDF处理相关路由会自动应用数据隔离检查
	}

	// 文档扫描路由应用数据隔离
	scanner := group.Group("/scanner")
	scanner.Use(oauth.LoginRequired())
	scanner.Use(isolationMiddleware.TeamDataIsolation())
	scanner.Use(isolationMiddleware.RateLimitByRole())
	{
		// 文档扫描相关路由会自动应用数据隔离检查
	}

	// 系统级路由应用可见性控制
	system := group.Group("/system")
	system.Use(oauth.LoginRequired())
	system.Use(isolationMiddleware.SystemVisibilityControl())
	{
		// 系统级操作需要特殊的可见性控制
	}

	// 搜索路由应用数据过滤
	search := group.Group("/search")
	search.Use(oauth.LoginRequired())
	search.Use(isolationMiddleware.DataVisibilityFilter())
	search.Use(isolationMiddleware.SystemVisibilityControl())
	{
		// 搜索功能需要根据用户权限过滤结果
	}

	// 统计和分析路由应用团队隔离
	analytics := group.Group("/analytics")
	analytics.Use(oauth.LoginRequired())
	analytics.Use(isolationMiddleware.RequireTeamAccess("team_manager", "super_admin"))
	analytics.Use(isolationMiddleware.DataVisibilityFilter())
	{
		// 统计分析功能需要团队级别的访问控制
	}

	// 管理员专用路由
	admin := group.Group("/admin")
	admin.Use(oauth.LoginRequired())
	admin.Use(isolationMiddleware.CrossTeamAccessOnly())
	{
		// 管理员专用功能，只有超级管理员可以访问
		admin.GET("/teams", r.handler.ListTeamIsolationPolicies)            // 列出所有团队隔离策略
		admin.GET("/system/config", r.handler.GetSystemVisibilityConfig)    // 获取系统可见性配置
		admin.PUT("/system/config", r.handler.UpdateSystemVisibilityConfig) // 更新系统可见性配置
		admin.GET("/violations/summary", r.handler.GetViolationsSummary)    // 获取违规统计摘要
		admin.POST("/users/:user_id/block", r.handler.BlockUser)            // 封禁用户
		admin.POST("/users/:user_id/unblock", r.handler.UnblockUser)        // 解封用户
	}

	// 审计日志路由
	audit := group.Group("/audit")
	audit.Use(oauth.LoginRequired())
	audit.Use(isolationMiddleware.AuditLog())
	audit.Use(isolationMiddleware.RequireTeamAccess("team_manager", "super_admin"))
	{
		// 审计日志查看需要管理员权限
		audit.GET("/access-logs", r.handler.GetDetailedAccessLogs)            // 获取详细访问日志
		audit.GET("/operation-logs", r.handler.GetOperationLogs)              // 获取操作日志
		audit.GET("/cross-team-activities", r.handler.GetCrossTeamActivities) // 获取跨团队活动记录
		audit.GET("/user-activities/:user_id", r.handler.GetUserActivities)   // 获取特定用户的活动记录
	}

	// 数据导出路由（特殊权限）
	export := group.Group("/export")
	export.Use(oauth.LoginRequired())
	export.Use(isolationMiddleware.CrossTeamAccessOnly()) // 只有超级管理员可以导出
	{
		export.POST("/access-logs", r.handler.ExportAccessLogs)      // 导出访问日志
		export.POST("/violations", r.handler.ExportViolations)       // 导出违规记录
		export.POST("/team-data/:team_id", r.handler.ExportTeamData) // 导出团队数据
	}
}

// InitializeDefaultPolicies 初始化默认隔离策略
func (r *IsolationRouter) InitializeDefaultPolicies() error {
	// 检查是否已有系统可见性配置
	var config models.SystemVisibilityConfig
	if err := r.db.First(&config).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			// 创建默认系统配置
			defaultConfig := models.SystemVisibilityConfig{
				GlobalSettings: struct {
					AllowGlobalSearch    bool `json:"allow_global_search" gorm:"default:false"`
					AllowCrossTeamAccess bool `json:"allow_cross_team_access" gorm:"default:false"`
					RequireApproval      bool `json:"require_approval" gorm:"default:true"`
				}{
					AllowGlobalSearch:    false,
					AllowCrossTeamAccess: false,
					RequireApproval:      true,
				},
				AuditEnabled:       true,
				AlertEnabled:       true,
				MaxViolationCount:  5,
				ViolationBlockTime: 30,
				CreatedBy:          "system",
			}

			if err := r.db.Create(&defaultConfig).Error; err != nil {
				return err
			}
		} else {
			return err
		}
	}

	return nil
}
