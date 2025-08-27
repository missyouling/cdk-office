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

package survey

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"cdk-office/internal/db"
	"cdk-office/internal/models"
)

// SurveyPermissionMiddleware Survey权限控制中间件
type SurveyPermissionMiddleware struct {
	db *gorm.DB
}

// NewSurveyPermissionMiddleware 创建Survey权限中间件
func NewSurveyPermissionMiddleware() *SurveyPermissionMiddleware {
	return &SurveyPermissionMiddleware{
		db: db.GetDB(),
	}
}

// RequireEditPermission 需要编辑权限
func (m *SurveyPermissionMiddleware) RequireEditPermission() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !m.hasPermission(c, "can_edit") {
			c.JSON(http.StatusForbidden, gin.H{"error": "No edit permission"})
			c.Abort()
			return
		}
		c.Next()
	}
}

// RequireDeletePermission 需要删除权限
func (m *SurveyPermissionMiddleware) RequireDeletePermission() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !m.hasPermission(c, "can_delete") {
			c.JSON(http.StatusForbidden, gin.H{"error": "No delete permission"})
			c.Abort()
			return
		}
		c.Next()
	}
}

// RequireManagePermission 需要管理权限
func (m *SurveyPermissionMiddleware) RequireManagePermission() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !m.hasPermission(c, "can_manage") {
			c.JSON(http.StatusForbidden, gin.H{"error": "No manage permission"})
			c.Abort()
			return
		}
		c.Next()
	}
}

// RequireAnalyzePermission 需要分析权限
func (m *SurveyPermissionMiddleware) RequireAnalyzePermission() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !m.hasPermission(c, "can_analyze") {
			c.JSON(http.StatusForbidden, gin.H{"error": "No analyze permission"})
			c.Abort()
			return
		}
		c.Next()
	}
}

// RequireExportPermission 需要导出权限
func (m *SurveyPermissionMiddleware) RequireExportPermission() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !m.hasPermission(c, "can_export") {
			c.JSON(http.StatusForbidden, gin.H{"error": "No export permission"})
			c.Abort()
			return
		}
		c.Next()
	}
}

// SurveyOwnerPermission 问卷创建者权限检查
func (m *SurveyPermissionMiddleware) SurveyOwnerPermission() gin.HandlerFunc {
	return func(c *gin.Context) {
		surveyID := c.Param("id")
		userID := m.getUserID(c)
		
		if userID == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
			c.Abort()
			return
		}

		var survey models.Survey
		if err := m.db.Where("survey_id = ?", surveyID).First(&survey).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				c.JSON(http.StatusNotFound, gin.H{"error": "Survey not found"})
			} else {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
			}
			c.Abort()
			return
		}

		// 检查是否为创建者
		if survey.CreatedBy != userID {
			c.JSON(http.StatusForbidden, gin.H{"error": "Only survey creator can perform this action"})
			c.Abort()
			return
		}

		c.Next()
	}
}

// PublicSurveyAccessible 公开问卷可访问检查
func (m *SurveyPermissionMiddleware) PublicSurveyAccessible() gin.HandlerFunc {
	return func(c *gin.Context) {
		surveyID := c.Param("id")

		var survey models.Survey
		if err := m.db.Where("survey_id = ? AND is_public = ? AND status = ?", 
			surveyID, true, "active").First(&survey).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				c.JSON(http.StatusNotFound, gin.H{"error": "Public survey not found or not active"})
			} else {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
			}
			c.Abort()
			return
		}

		// 将survey信息存入context，避免重复查询
		c.Set("survey", survey)
		c.Next()
	}
}

// hasPermission 检查用户是否有指定权限
func (m *SurveyPermissionMiddleware) hasPermission(c *gin.Context, permission string) bool {
	surveyID := c.Param("id")
	userID := m.getUserID(c)
	
	if userID == "" {
		return false
	}

	// 检查是否为问卷创建者（创建者拥有所有权限）
	var survey models.Survey
	if err := m.db.Where("survey_id = ?", surveyID).First(&survey).Error; err == nil {
		if survey.CreatedBy == userID {
			return true
		}
	}

	// 检查具体权限
	var perm models.SurveyPermission
	query := m.db.Where("survey_id = ? AND user_id = ?", surveyID, userID)
	
	switch permission {
	case "can_view":
		query = query.Where("can_view = ?", true)
	case "can_edit":
		query = query.Where("can_edit = ?", true)
	case "can_delete":
		query = query.Where("can_delete = ?", true)
	case "can_manage":
		query = query.Where("can_manage = ?", true)
	case "can_analyze":
		query = query.Where("can_analyze = ?", true)
	case "can_export":
		query = query.Where("can_export = ?", true)
	default:
		return false
	}

	err := query.First(&perm).Error
	return err == nil
}

// getUserID 从上下文获取用户ID
func (m *SurveyPermissionMiddleware) getUserID(c *gin.Context) string {
	// 首先尝试从OAuth中间件设置的字段获取
	if userID := c.GetString("user_id"); userID != "" {
		return userID
	}

	// 尝试从用户对象获取
	if user, exists := c.Get("user"); exists {
		if u, ok := user.(map[string]interface{}); ok {
			if id, ok := u["id"].(string); ok {
				return id
			}
		}
	}

	return ""
}

// getTeamID 从上下文获取团队ID
func (m *SurveyPermissionMiddleware) getTeamID(c *gin.Context) string {
	// 首先尝试从OAuth中间件设置的字段获取
	if teamID := c.GetString("team_id"); teamID != "" {
		return teamID
	}

	// 尝试从用户对象获取
	if user, exists := c.Get("user"); exists {
		if u, ok := user.(map[string]interface{}); ok {
			if id, ok := u["team_id"].(string); ok {
				return id
			}
		}
	}

	return ""
}

// AdminRequired 需要管理员权限
func AdminRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 检查用户是否为管理员
		if user, exists := c.Get("user"); exists {
			if u, ok := user.(map[string]interface{}); ok {
				if isAdmin, ok := u["is_admin"].(bool); ok && isAdmin {
					c.Next()
					return
				}
			}
		}

		c.JSON(http.StatusForbidden, gin.H{"error": "Admin permission required"})
		c.Abort()
	}
}

// TeamManagerRequired 需要团队管理员权限
func TeamManagerRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 检查用户角色
		if user, exists := c.Get("user"); exists {
			if u, ok := user.(map[string]interface{}); ok {
				// 检查是否为超级管理员
				if isAdmin, ok := u["is_admin"].(bool); ok && isAdmin {
					c.Next()
					return
				}
				
				// 检查是否为团队管理员
				if role, ok := u["role"].(string); ok && (role == "team_manager" || role == "manager") {
					c.Next()
					return
				}
			}
		}

		c.JSON(http.StatusForbidden, gin.H{"error": "Team manager permission required"})
		c.Abort()
	}
}

// 全局中间件实例
var PermissionMiddleware = NewSurveyPermissionMiddleware()