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

package middleware

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/linux-do/cdk-office/internal/models"
	"github.com/linux-do/cdk-office/internal/services/isolation"
)

// DataIsolationMiddleware 数据隔离中间件
type DataIsolationMiddleware struct {
	isolationService *isolation.DataIsolationService
	db               *gorm.DB
}

// NewDataIsolationMiddleware 创建数据隔离中间件
func NewDataIsolationMiddleware(isolationService *isolation.DataIsolationService, db *gorm.DB) *DataIsolationMiddleware {
	return &DataIsolationMiddleware{
		isolationService: isolationService,
		db:               db,
	}
}

// TeamDataIsolation 团队数据隔离检查
func (m *DataIsolationMiddleware) TeamDataIsolation() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取用户信息
		userID := m.getUserID(c)
		teamID := m.getTeamID(c)
		userRole := m.getUserRole(c)

		if userID == "" || teamID == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
			c.Abort()
			return
		}

		// 获取资源信息
		resourceID := m.getResourceID(c)
		resourceType := m.getResourceType(c)
		actionType := m.getActionType(c)

		if resourceID == "" || resourceType == "" {
			// 如果无法获取资源信息，可能是列表查询等操作，跳过检查
			c.Next()
			return
		}

		// 获取资源所有者团队ID
		ownerTeamID, err := m.getResourceOwnerTeam(resourceID, resourceType)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get resource owner"})
			c.Abort()
			return
		}

		// 构建访问上下文
		accessCtx := &isolation.AccessContext{
			UserID:        userID,
			TeamID:        teamID,
			UserRole:      userRole,
			ResourceID:    resourceID,
			ResourceType:  resourceType,
			ActionType:    actionType,
			OwnerTeamID:   ownerTeamID,
			RequestSource: "web",
			Metadata: map[string]string{
				"ip_address":    c.ClientIP(),
				"user_agent":    c.GetHeader("User-Agent"),
				"session_id":    m.getSessionID(c),
				"resource_name": m.getResourceName(c, resourceID, resourceType),
			},
		}

		// 检查访问权限
		result, err := m.isolationService.CheckAccess(context.Background(), accessCtx)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Access check failed"})
			c.Abort()
			return
		}

		if !result.Allowed {
			// 记录访问拒绝
			c.JSON(http.StatusForbidden, gin.H{
				"error":             "Access denied",
				"reason":            result.Reason,
				"requires_approval": result.RequiresApproval,
			})
			c.Abort()
			return
		}

		// 设置访问限制到上下文
		if result.Restrictions != nil {
			c.Set("access_restrictions", result.Restrictions)
		}

		c.Next()
	}
}

// RequireTeamAccess 要求团队访问权限
func (m *DataIsolationMiddleware) RequireTeamAccess(allowedRoles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userRole := m.getUserRole(c)
		userTeamID := m.getTeamID(c)
		resourceTeamID := c.Param("team_id")

		// 检查是否为超级管理员
		if userRole == "super_admin" {
			c.Next()
			return
		}

		// 检查团队权限
		if resourceTeamID != "" && userTeamID != resourceTeamID {
			c.JSON(http.StatusForbidden, gin.H{"error": "Cross-team access denied"})
			c.Abort()
			return
		}

		// 检查角色权限
		if len(allowedRoles) > 0 {
			allowed := false
			for _, role := range allowedRoles {
				if userRole == role {
					allowed = true
					break
				}
			}
			if !allowed {
				c.JSON(http.StatusForbidden, gin.H{"error": "Insufficient role permissions"})
				c.Abort()
				return
			}
		}

		c.Next()
	}
}

// CrossTeamAccessOnly 仅超级管理员可跨团队访问
func (m *DataIsolationMiddleware) CrossTeamAccessOnly() gin.HandlerFunc {
	return func(c *gin.Context) {
		userRole := m.getUserRole(c)

		if userRole != "super_admin" {
			c.JSON(http.StatusForbidden, gin.H{
				"error": "Cross-team operations are restricted to super administrators only",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// DataVisibilityFilter 数据可见性过滤器
func (m *DataIsolationMiddleware) DataVisibilityFilter() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := m.getUserID(c)
		teamID := m.getTeamID(c)
		userRole := m.getUserRole(c)

		if userID == "" {
			c.Next()
			return
		}

		// 根据用户角色设置数据过滤条件
		filters := m.buildDataFilters(userRole, teamID)
		c.Set("data_filters", filters)

		c.Next()
	}
}

// RateLimitByRole 按角色限制访问频率
func (m *DataIsolationMiddleware) RateLimitByRole() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := m.getUserID(c)
		userRole := m.getUserRole(c)

		if userID == "" {
			c.Next()
			return
		}

		// 检查用户访问档案
		var profile models.UserDataAccessProfile
		if err := m.db.Where("user_id = ?", userID).First(&profile).Error; err != nil {
			c.Next()
			return
		}

		// 检查是否被封禁
		if profile.SecuritySettings.IsBlocked {
			if profile.SecuritySettings.BlockExpires != nil && time.Now().After(*profile.SecuritySettings.BlockExpires) {
				// 解除封禁
				profile.SecuritySettings.IsBlocked = false
				profile.SecuritySettings.BlockExpires = nil
				m.db.Save(&profile)
			} else {
				c.JSON(http.StatusForbidden, gin.H{
					"error": "User is temporarily blocked due to security violations",
				})
				c.Abort()
				return
			}
		}

		c.Next()
	}
}

// AuditLog 审计日志中间件
func (m *DataIsolationMiddleware) AuditLog() gin.HandlerFunc {
	return func(c *gin.Context) {
		startTime := time.Now()

		// 记录请求开始
		c.Next()

		// 记录请求结束
		duration := time.Since(startTime)

		// 异步记录审计日志
		go m.recordAuditLog(c, duration)
	}
}

// SystemVisibilityControl 系统可见性控制
func (m *DataIsolationMiddleware) SystemVisibilityControl() gin.HandlerFunc {
	return func(c *gin.Context) {
		userRole := m.getUserRole(c)

		// 获取系统可见性配置
		var config models.SystemVisibilityConfig
		m.db.First(&config)

		// 检查全局设置
		if !config.GlobalSettings.AllowGlobalSearch && strings.Contains(c.Request.URL.Path, "/search") {
			if userRole != "super_admin" {
				c.JSON(http.StatusForbidden, gin.H{"error": "Global search is disabled"})
				c.Abort()
				return
			}
		}

		if !config.GlobalSettings.AllowCrossTeamAccess && m.isCrossTeamRequest(c) {
			if userRole != "super_admin" {
				c.JSON(http.StatusForbidden, gin.H{"error": "Cross-team access is disabled"})
				c.Abort()
				return
			}
		}

		c.Next()
	}
}

// 辅助方法

func (m *DataIsolationMiddleware) getUserID(c *gin.Context) string {
	if userID := c.GetString("user_id"); userID != "" {
		return userID
	}
	if user, exists := c.Get("user"); exists {
		if u, ok := user.(map[string]interface{}); ok {
			if id, ok := u["id"].(string); ok {
				return id
			}
		}
	}
	return ""
}

func (m *DataIsolationMiddleware) getTeamID(c *gin.Context) string {
	if teamID := c.GetString("team_id"); teamID != "" {
		return teamID
	}
	if user, exists := c.Get("user"); exists {
		if u, ok := user.(map[string]interface{}); ok {
			if id, ok := u["team_id"].(string); ok {
				return id
			}
		}
	}
	return ""
}

func (m *DataIsolationMiddleware) getUserRole(c *gin.Context) string {
	if role := c.GetString("role"); role != "" {
		return role
	}
	if user, exists := c.Get("user"); exists {
		if u, ok := user.(map[string]interface{}); ok {
			if role, ok := u["role"].(string); ok {
				return role
			}
		}
	}
	return "normal_user"
}

func (m *DataIsolationMiddleware) getResourceID(c *gin.Context) string {
	// 尝试从URL参数获取
	if id := c.Param("id"); id != "" {
		return id
	}
	if id := c.Param("resource_id"); id != "" {
		return id
	}
	if id := c.Param("document_id"); id != "" {
		return id
	}
	if id := c.Param("knowledge_id"); id != "" {
		return id
	}

	// 尝试从查询参数获取
	if id := c.Query("id"); id != "" {
		return id
	}

	return ""
}

func (m *DataIsolationMiddleware) getResourceType(c *gin.Context) string {
	path := c.Request.URL.Path

	switch {
	case strings.Contains(path, "/knowledge"):
		return "knowledge"
	case strings.Contains(path, "/document"):
		return "document"
	case strings.Contains(path, "/survey"):
		return "survey"
	case strings.Contains(path, "/pdf"):
		return "pdf"
	case strings.Contains(path, "/scan"):
		return "scan"
	default:
		return "unknown"
	}
}

func (m *DataIsolationMiddleware) getActionType(c *gin.Context) string {
	method := c.Request.Method
	path := c.Request.URL.Path

	switch method {
	case "GET":
		if strings.Contains(path, "/download") {
			return "download"
		}
		return "view"
	case "POST":
		if strings.Contains(path, "/share") {
			return "share"
		}
		return "create"
	case "PUT", "PATCH":
		return "edit"
	case "DELETE":
		return "delete"
	default:
		return "unknown"
	}
}

func (m *DataIsolationMiddleware) getResourceOwnerTeam(resourceID, resourceType string) (string, error) {
	var teamID string
	var err error

	switch resourceType {
	case "knowledge":
		err = m.db.Model(&models.PersonalKnowledgeBase{}).
			Where("id = ?", resourceID).
			Select("team_id").
			Scan(&teamID).Error
	case "document":
		err = m.db.Model(&models.PersonalKnowledgeBase{}).
			Where("id = ?", resourceID).
			Select("team_id").
			Scan(&teamID).Error
	case "survey":
		err = m.db.Raw("SELECT team_id FROM surveys WHERE survey_id = ?", resourceID).
			Scan(&teamID).Error
	default:
		// 默认返回空团队ID，表示系统资源
		return "", nil
	}

	return teamID, err
}

func (m *DataIsolationMiddleware) getSessionID(c *gin.Context) string {
	if sessionID := c.GetString("session_id"); sessionID != "" {
		return sessionID
	}
	return ""
}

func (m *DataIsolationMiddleware) getResourceName(c *gin.Context, resourceID, resourceType string) string {
	var name string

	switch resourceType {
	case "knowledge", "document":
		m.db.Model(&models.PersonalKnowledgeBase{}).
			Where("id = ?", resourceID).
			Select("title").
			Scan(&name)
	case "survey":
		m.db.Raw("SELECT title FROM surveys WHERE survey_id = ?", resourceID).
			Scan(&name)
	}

	return name
}

func (m *DataIsolationMiddleware) buildDataFilters(userRole, teamID string) map[string]interface{} {
	filters := make(map[string]interface{})

	switch userRole {
	case "super_admin":
		// 超级管理员无过滤
		filters["team_id"] = nil
	case "team_manager":
		// 团队管理员只能看本团队和公开数据
		filters["team_visibility"] = []string{teamID, "public"}
	case "collaborator":
		// 协作用户可以看本团队和部分跨团队数据
		filters["team_visibility"] = []string{teamID, "team_public", "system_public"}
	case "normal_user":
		// 普通用户只能看公开数据
		filters["visibility"] = []string{"system_public", "team_public"}
	}

	return filters
}

func (m *DataIsolationMiddleware) isCrossTeamRequest(c *gin.Context) bool {
	userTeamID := m.getTeamID(c)
	resourceID := m.getResourceID(c)
	resourceType := m.getResourceType(c)

	if resourceID == "" || userTeamID == "" {
		return false
	}

	ownerTeamID, err := m.getResourceOwnerTeam(resourceID, resourceType)
	if err != nil || ownerTeamID == "" {
		return false
	}

	return userTeamID != ownerTeamID
}

func (m *DataIsolationMiddleware) recordAuditLog(c *gin.Context, duration time.Duration) {
	// 这里可以记录详细的审计日志
	userID := m.getUserID(c)
	if userID == "" {
		return
	}

	// 创建审计日志记录
	// 可以根据需要扩展记录的内容
}
