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
	"log"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/linux-do/cdk-office/internal/apps/dify"
	"github.com/linux-do/cdk-office/internal/config"
	"github.com/linux-do/cdk-office/internal/middleware"
	"github.com/linux-do/cdk-office/internal/models"
)

// AIRouter AI模块路由管理器
type AIRouter struct {
	db           *gorm.DB
	service      *Service
	handler      *Handler
	documentSync *DocumentSyncService
	authzMdw     *middleware.AuthzMiddleware
}

// NewAIRouter 创建AI路由管理器
func NewAIRouter(db *gorm.DB) (*AIRouter, error) {
	// 创建Dify配置
	difyConfig := &dify.Config{
		APIKey:                   config.Config.Dify.APIKey,
		APIEndpoint:              config.Config.Dify.APIEndpoint,
		ChatEndpoint:             config.Config.Dify.ChatEndpoint,
		CompletionEndpoint:       config.Config.Dify.CompletionEndpoint,
		DatasetsEndpoint:         config.Config.Dify.DatasetsEndpoint,
		DocumentsEndpoint:        config.Config.Dify.DocumentsEndpoint,
		SurveyAnalysisWorkflowID: config.Config.Dify.SurveyAnalysisWorkflowID,
		DefaultDatasetID:         config.Config.Dify.KnowledgeBaseID,
		Timeout:                  config.Config.Dify.Timeout,
	}

	// 创建AI服务
	service := NewService(db, difyConfig)

	// 创建Handler
	handler := NewHandler(service)

	// 创建文档同步服务
	documentSync := NewDocumentSyncService(db, difyConfig)

	// 创建权限中间件
	authzMdw, err := middleware.NewAuthzMiddleware(db)
	if err != nil {
		log.Printf("Warning: Failed to initialize authorization middleware: %v", err)
		// 继续运行，但权限检查可能不可用
	}

	// 初始化默认权限策略
	if authzMdw != nil {
		if err := authzMdw.InitializeDefaultPolicies(); err != nil {
			log.Printf("Warning: Failed to initialize default policies: %v", err)
		}
	}

	return &AIRouter{
		db:           db,
		service:      service,
		handler:      handler,
		documentSync: documentSync,
		authzMdw:     authzMdw,
	}, nil
}

// RegisterRoutes 注册所有AI相关路由
func (r *AIRouter) RegisterRoutes(router *gin.RouterGroup) {
	// AI智能问答路由组
	aiGroup := router.Group("/ai")
	if r.authzMdw != nil {
		// 应用权限中间件
		aiGroup.Use(r.authzMdw.RequirePermission("ai:read"))
	}
	{
		// 智能问答接口（需要写权限）
		chatGroup := aiGroup.Group("/chat")
		if r.authzMdw != nil {
			chatGroup.Use(r.authzMdw.RequirePermission("ai:write"))
		}
		{
			chatGroup.POST("", r.handler.Chat)
			chatGroup.GET("/history", r.handler.GetChatHistory)
			chatGroup.PATCH("/:message_id/feedback", r.handler.UpdateFeedback)
		}

		// 统计信息（只需读权限）
		aiGroup.GET("/chat/stats", r.handler.GetStats)
	}

	// 文档同步路由组
	documentsGroup := router.Group("/documents")
	if r.authzMdw != nil {
		// 文档同步需要读和写权限
		documentsGroup.Use(r.authzMdw.RequireMultiplePermissions([]string{"document:read", "document:write"}))
	}
	{
		documentsGroup.POST("/:id/sync", r.handleSyncDocument)
		documentsGroup.GET("/:id/sync-status", r.handleGetSyncStatus)
		documentsGroup.POST("/:id/retry-sync", r.handleRetrySync)
	}

	// 权限管理路由组（仅管理员）
	if r.authzMdw != nil {
		permissionGroup := router.Group("/permissions")
		permissionGroup.Use(r.authzMdw.RequireAnyPermission([]string{"user:write", "team:write"}))
		{
			permissionGroup.POST("/users/:user_id/roles", r.handleAssignRole)
			permissionGroup.DELETE("/users/:user_id/roles/:role", r.handleRemoveRole)
			permissionGroup.GET("/users/:user_id", r.handleGetUserPermissions)
			permissionGroup.GET("/audit-logs", r.handleGetAuditLogs)
		}
	}
}

// 文档同步处理函数
func (r *AIRouter) handleSyncDocument(c *gin.Context) {
	documentID := c.Param("id")

	// 获取文档信息
	var doc struct {
		ID        string `json:"id"`
		Name      string `json:"name"`
		FileType  string `json:"file_type"`
		FileSize  int64  `json:"file_size"`
		TeamID    string `json:"team_id"`
		CreatedBy string `json:"created_by"`
	}

	// 这里应该从数据库获取文档信息，简化为从请求体获取
	if err := c.ShouldBindJSON(&doc); err != nil {
		c.JSON(400, gin.H{"error": "Invalid document data"})
		return
	}

	// 调用文档同步服务
	err := r.documentSync.SyncToDify(c.Request.Context(), &models.Document{
		ID:        documentID,
		Name:      doc.Name,
		FileType:  doc.FileType,
		FileSize:  doc.FileSize,
		TeamID:    doc.TeamID,
		CreatedBy: doc.CreatedBy,
	})

	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"message": "Document sync started"})
}

// 获取同步状态处理函数
func (r *AIRouter) handleGetSyncStatus(c *gin.Context) {
	documentID := c.Param("id")

	status, err := r.documentSync.GetSyncStatus(c.Request.Context(), documentID)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, status)
}

// 重试同步处理函数
func (r *AIRouter) handleRetrySync(c *gin.Context) {
	documentID := c.Param("id")

	err := r.documentSync.RetrySync(c.Request.Context(), documentID)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"message": "Retry sync initiated"})
}

// 分配角色处理函数
func (r *AIRouter) handleAssignRole(c *gin.Context) {
	if r.authzMdw == nil {
		c.JSON(500, gin.H{"error": "Authorization middleware not available"})
		return
	}

	userID := c.Param("user_id")
	teamID, exists := c.Get("team_id")
	if !exists {
		c.JSON(400, gin.H{"error": "Team ID not found"})
		return
	}

	var req struct {
		Role string `json:"role" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "Invalid request"})
		return
	}

	// 创建权限服务
	permissionService := middleware.NewPermissionService(r.authzMdw, r.db)

	err := permissionService.AssignUserRole(teamID.(string), userID, req.Role)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	// 记录审计日志
	r.logPermissionAudit(c, "role_assigned", userID, req.Role, "")

	c.JSON(200, gin.H{"message": "Role assigned successfully"})
}

// 移除角色处理函数
func (r *AIRouter) handleRemoveRole(c *gin.Context) {
	if r.authzMdw == nil {
		c.JSON(500, gin.H{"error": "Authorization middleware not available"})
		return
	}

	userID := c.Param("user_id")
	role := c.Param("role")
	teamID, exists := c.Get("team_id")
	if !exists {
		c.JSON(400, gin.H{"error": "Team ID not found"})
		return
	}

	permissionService := middleware.NewPermissionService(r.authzMdw, r.db)

	err := permissionService.RemoveUserRole(teamID.(string), userID, role)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	// 记录审计日志
	r.logPermissionAudit(c, "role_removed", userID, role, "")

	c.JSON(200, gin.H{"message": "Role removed successfully"})
}

// 获取用户权限处理函数
func (r *AIRouter) handleGetUserPermissions(c *gin.Context) {
	if r.authzMdw == nil {
		c.JSON(500, gin.H{"error": "Authorization middleware not available"})
		return
	}

	userID := c.Param("user_id")
	teamID, exists := c.Get("team_id")
	if !exists {
		c.JSON(400, gin.H{"error": "Team ID not found"})
		return
	}

	permissionService := middleware.NewPermissionService(r.authzMdw, r.db)

	permissions, err := permissionService.GetUserPermissions(userID, teamID.(string))
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"permissions": permissions})
}

// 获取审计日志处理函数
func (r *AIRouter) handleGetAuditLogs(c *gin.Context) {
	teamID, exists := c.Get("team_id")
	if !exists {
		c.JSON(400, gin.H{"error": "Team ID not found"})
		return
	}

	var logs []struct {
		ID        string `json:"id"`
		UserID    string `json:"user_id"`
		Action    string `json:"action"`
		Resource  string `json:"resource"`
		CreatedAt string `json:"created_at"`
	}

	// 查询审计日志
	err := r.db.Table("permission_audit_logs").
		Where("team_id = ?", teamID).
		Order("created_at DESC").
		Limit(100).
		Scan(&logs).Error

	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"logs": logs})
}

// 记录权限操作审计日志
func (r *AIRouter) logPermissionAudit(c *gin.Context, action, targetUserID, resource, oldValue string) {
	operatorID, _ := c.Get("user_id")
	teamID, _ := c.Get("team_id")

	auditLog := map[string]interface{}{
		"user_id":     targetUserID,
		"team_id":     teamID,
		"action":      action,
		"resource":    resource,
		"old_value":   oldValue,
		"operator_id": operatorID,
		"ip_address":  c.ClientIP(),
		"user_agent":  c.Request.UserAgent(),
	}

	r.db.Table("permission_audit_logs").Create(auditLog)
}

// GetDocumentSyncService 获取文档同步服务（供其他模块使用）
func (r *AIRouter) GetDocumentSyncService() *DocumentSyncService {
	return r.documentSync
}

// GetAuthzMiddleware 获取权限中间件（供其他模块使用）
func (r *AIRouter) GetAuthzMiddleware() *middleware.AuthzMiddleware {
	return r.authzMdw
}
