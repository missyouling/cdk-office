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
	"net/http"
	"strings"

	"github.com/casbin/casbin/v2"
	"github.com/casbin/casbin/v2/model"
	gormadapter "github.com/casbin/gorm-adapter/v3"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/linux-do/cdk-office/internal/models"
)

// AuthzMiddleware Casbin权限控制中间件
type AuthzMiddleware struct {
	enforcer *casbin.Enforcer
}

// NewAuthzMiddleware 创建权限控制中间件
func NewAuthzMiddleware(db *gorm.DB) (*AuthzMiddleware, error) {
	// 创建Gorm适配器
	adapter, err := gormadapter.NewAdapterByDB(db)
	if err != nil {
		return nil, err
	}

	// 创建模型
	m, err := model.NewModelFromString(casbinModel)
	if err != nil {
		return nil, err
	}

	// 创建执行器
	enforcer, err := casbin.NewEnforcer(m, adapter)
	if err != nil {
		return nil, err
	}

	// 自动加载策略
	if err := enforcer.LoadPolicy(); err != nil {
		return nil, err
	}

	return &AuthzMiddleware{
		enforcer: enforcer,
	}, nil
}

// RequirePermission 权限验证中间件
func (m *AuthzMiddleware) RequirePermission(permission string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取用户信息
		userID, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, models.ErrorResponse{
				Code:    "UNAUTHORIZED",
				Message: "用户未登录",
			})
			c.Abort()
			return
		}

		teamID, exists := c.Get("team_id")
		if !exists {
			c.JSON(http.StatusBadRequest, models.ErrorResponse{
				Code:    "INVALID_TEAM",
				Message: "无效的团队信息",
			})
			c.Abort()
			return
		}

		// 检查权限
		resource := strings.Split(permission, ":")[0] // 例如 "team:read" -> "team"
		action := strings.Split(permission, ":")[1]   // 例如 "team:read" -> "read"

		// 构建权限检查参数
		// subject: user_id, object: team_id/resource, action: action
		subject := userID.(string)
		object := teamID.(string) + "/" + resource

		allowed, err := m.enforcer.Enforce(subject, object, action)
		if err != nil {
			c.JSON(http.StatusInternalServerError, models.ErrorResponse{
				Code:    "AUTHZ_ERROR",
				Message: "权限检查失败: " + err.Error(),
			})
			c.Abort()
			return
		}

		if !allowed {
			c.JSON(http.StatusForbidden, models.ErrorResponse{
				Code:    "PERMISSION_DENIED",
				Message: "权限不足，需要 " + permission + " 权限",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequireMultiplePermissions 多权限验证中间件
func (m *AuthzMiddleware) RequireMultiplePermissions(permissions []string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取用户信息
		userID, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, models.ErrorResponse{
				Code:    "UNAUTHORIZED",
				Message: "用户未登录",
			})
			c.Abort()
			return
		}

		teamID, exists := c.Get("team_id")
		if !exists {
			c.JSON(http.StatusBadRequest, models.ErrorResponse{
				Code:    "INVALID_TEAM",
				Message: "无效的团队信息",
			})
			c.Abort()
			return
		}

		subject := userID.(string)

		// 检查所有权限
		for _, permission := range permissions {
			resource := strings.Split(permission, ":")[0]
			action := strings.Split(permission, ":")[1]
			object := teamID.(string) + "/" + resource

			allowed, err := m.enforcer.Enforce(subject, object, action)
			if err != nil {
				c.JSON(http.StatusInternalServerError, models.ErrorResponse{
					Code:    "AUTHZ_ERROR",
					Message: "权限检查失败: " + err.Error(),
				})
				c.Abort()
				return
			}

			if !allowed {
				c.JSON(http.StatusForbidden, models.ErrorResponse{
					Code:    "PERMISSION_DENIED",
					Message: "权限不足，需要 " + permission + " 权限",
				})
				c.Abort()
				return
			}
		}

		c.Next()
	}
}

// RequireAnyPermission 任一权限验证中间件（只需满足其中一个权限即可）
func (m *AuthzMiddleware) RequireAnyPermission(permissions []string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取用户信息
		userID, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, models.ErrorResponse{
				Code:    "UNAUTHORIZED",
				Message: "用户未登录",
			})
			c.Abort()
			return
		}

		teamID, exists := c.Get("team_id")
		if !exists {
			c.JSON(http.StatusBadRequest, models.ErrorResponse{
				Code:    "INVALID_TEAM",
				Message: "无效的团队信息",
			})
			c.Abort()
			return
		}

		subject := userID.(string)
		hasPermission := false

		// 检查是否有任一权限
		for _, permission := range permissions {
			resource := strings.Split(permission, ":")[0]
			action := strings.Split(permission, ":")[1]
			object := teamID.(string) + "/" + resource

			allowed, err := m.enforcer.Enforce(subject, object, action)
			if err != nil {
				c.JSON(http.StatusInternalServerError, models.ErrorResponse{
					Code:    "AUTHZ_ERROR",
					Message: "权限检查失败: " + err.Error(),
				})
				c.Abort()
				return
			}

			if allowed {
				hasPermission = true
				break
			}
		}

		if !hasPermission {
			c.JSON(http.StatusForbidden, models.ErrorResponse{
				Code:    "PERMISSION_DENIED",
				Message: "权限不足，需要以下权限之一: " + strings.Join(permissions, ", "),
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// AddPolicy 添加权限策略
func (m *AuthzMiddleware) AddPolicy(userID, resource, action string) error {
	return m.enforcer.AddPolicy(userID, resource, action)
}

// RemovePolicy 移除权限策略
func (m *AuthzMiddleware) RemovePolicy(userID, resource, action string) error {
	return m.enforcer.RemovePolicy(userID, resource, action)
}

// AddRoleForUser 为用户添加角色
func (m *AuthzMiddleware) AddRoleForUser(userID, role string) error {
	return m.enforcer.AddRoleForUser(userID, role)
}

// DeleteRoleForUser 删除用户角色
func (m *AuthzMiddleware) DeleteRoleForUser(userID, role string) error {
	return m.enforcer.DeleteRoleForUser(userID, role)
}

// AddRolePermission 为角色添加权限
func (m *AuthzMiddleware) AddRolePermission(role, resource, action string) error {
	return m.enforcer.AddPolicy(role, resource, action)
}

// GetPermissionsForUser 获取用户权限
func (m *AuthzMiddleware) GetPermissionsForUser(userID string) [][]string {
	return m.enforcer.GetPermissionsForUser(userID)
}

// GetRolesForUser 获取用户角色
func (m *AuthzMiddleware) GetRolesForUser(userID string) ([]string, error) {
	return m.enforcer.GetRolesForUser(userID)
}

// SavePolicy 保存策略到数据库
func (m *AuthzMiddleware) SavePolicy() error {
	return m.enforcer.SavePolicy()
}

// LoadPolicy 从数据库加载策略
func (m *AuthzMiddleware) LoadPolicy() error {
	return m.enforcer.LoadPolicy()
}

// InitializeDefaultPolicies 初始化默认权限策略
func (m *AuthzMiddleware) InitializeDefaultPolicies() error {
	// 定义默认角色权限
	defaultPolicies := [][]string{
		// 超级管理员权限
		{"admin", "*/team", "read"},
		{"admin", "*/team", "write"},
		{"admin", "*/team", "delete"},
		{"admin", "*/document", "read"},
		{"admin", "*/document", "write"},
		{"admin", "*/document", "delete"},
		{"admin", "*/ai", "read"},
		{"admin", "*/ai", "write"},
		{"admin", "*/user", "read"},
		{"admin", "*/user", "write"},

		// 团队管理员权限
		{"manager", "*/team", "read"},
		{"manager", "*/team", "write"},
		{"manager", "*/document", "read"},
		{"manager", "*/document", "write"},
		{"manager", "*/ai", "read"},
		{"manager", "*/ai", "write"},
		{"manager", "*/user", "read"},

		// 普通用户权限
		{"user", "*/team", "read"},
		{"user", "*/document", "read"},
		{"user", "*/document", "write"},
		{"user", "*/ai", "read"},
		{"user", "*/ai", "write"},

		// 协作用户权限
		{"collaborator", "*/team", "read"},
		{"collaborator", "*/document", "read"},
		{"collaborator", "*/ai", "read"},
	}

	// 添加策略
	for _, policy := range defaultPolicies {
		if _, err := m.enforcer.AddPolicy(policy); err != nil {
			return err
		}
	}

	// 保存策略
	return m.enforcer.SavePolicy()
}

// Casbin模型定义
const casbinModel = `
[request_definition]
r = sub, obj, act

[policy_definition]
p = sub, obj, act

[role_definition]
g = _, _

[policy_effect]
e = some(where (p.eft == allow))

[matchers]
m = g(r.sub, p.sub) && keyMatch2(r.obj, p.obj) && r.act == p.act
`

// PermissionService 权限管理服务
type PermissionService struct {
	middleware *AuthzMiddleware
	db         *gorm.DB
}

// NewPermissionService 创建权限管理服务
func NewPermissionService(middleware *AuthzMiddleware, db *gorm.DB) *PermissionService {
	return &PermissionService{
		middleware: middleware,
		db:         db,
	}
}

// AssignUserRole 分配用户角色
func (s *PermissionService) AssignUserRole(teamID, userID, role string) error {
	// 添加用户角色，格式：userID -> teamID:role
	roleWithTeam := teamID + ":" + role
	return s.middleware.AddRoleForUser(userID, roleWithTeam)
}

// RemoveUserRole 移除用户角色
func (s *PermissionService) RemoveUserRole(teamID, userID, role string) error {
	roleWithTeam := teamID + ":" + role
	return s.middleware.DeleteRoleForUser(userID, roleWithTeam)
}

// CheckPermission 检查用户权限
func (s *PermissionService) CheckPermission(userID, teamID, resource, action string) (bool, error) {
	object := teamID + "/" + resource
	return s.middleware.enforcer.Enforce(userID, object, action)
}

// GetUserPermissions 获取用户在特定团队的权限
func (s *PermissionService) GetUserPermissions(userID, teamID string) ([]string, error) {
	permissions := []string{}

	// 获取用户在团队中的角色
	roles, err := s.middleware.GetRolesForUser(userID)
	if err != nil {
		return nil, err
	}

	// 过滤出当前团队的角色
	for _, role := range roles {
		if strings.HasPrefix(role, teamID+":") {
			// 获取该角色的权限
			rolePolicies := s.middleware.enforcer.GetPermissionsForUser(role)
			for _, policy := range rolePolicies {
				if len(policy) >= 3 {
					resource := strings.TrimPrefix(policy[1], "*/")
					action := policy[2]
					permissions = append(permissions, resource+":"+action)
				}
			}
		}
	}

	return permissions, nil
}
