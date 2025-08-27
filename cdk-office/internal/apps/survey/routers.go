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
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"cdk-office/internal/models"
	"cdk-office/internal/storage"
)

// SurveyHandler 问卷处理器
type SurveyHandler struct {
	db             *gorm.DB
	storageService *storage.StorageService
	difyService    *DifyService // Dify AI分析服务
}

// NewSurveyHandler 创建问卷处理器
func NewSurveyHandler(db *gorm.DB, storageService *storage.StorageService) *SurveyHandler {
	return &SurveyHandler{
		db:             db,
		storageService: storageService,
	}
}

// RegisterRoutes 注册路由
func (h *SurveyHandler) RegisterRoutes(router *gin.RouterGroup) {
	// 问卷管理路由
	surveys := router.Group("")
	{
		// 问卷CRUD操作
		surveys.POST("", h.CreateSurvey)                    // 创建问卷
		surveys.GET("", h.ListSurveys)                      // 获取问卷列表
		surveys.GET("/:id", h.GetSurvey)                    // 获取问卷详情
		surveys.PUT("/:id", PermissionMiddleware.RequireEditPermission(), h.UpdateSurvey)      // 更新问卷
		surveys.DELETE("/:id", PermissionMiddleware.RequireDeletePermission(), h.DeleteSurvey) // 删除问卷
		
		// 问卷状态管理
		surveys.POST("/:id/publish", PermissionMiddleware.RequireManagePermission(), h.PublishSurvey) // 发布问卷
		surveys.POST("/:id/close", PermissionMiddleware.RequireManagePermission(), h.CloseSurvey)     // 关闭问卷

		// 问卷响应管理
		surveys.POST("/:id/responses", h.SubmitResponse)                                              // 提交响应
		surveys.GET("/:id/responses", PermissionMiddleware.RequireAnalyzePermission(), h.GetResponses) // 获取响应列表
		surveys.GET("/responses/:responseId", h.GetResponse)                                          // 获取响应详情
		surveys.DELETE("/responses/:responseId", PermissionMiddleware.RequireManagePermission(), h.DeleteResponse) // 删除响应

		// 问卷分析功能
		surveys.POST("/:id/analyze", PermissionMiddleware.RequireAnalyzePermission(), h.TriggerAnalysis) // 触发分析
		surveys.GET("/:id/analysis", PermissionMiddleware.RequireAnalyzePermission(), h.GetAnalysis)    // 获取分析结果
		surveys.GET("/:id/export", PermissionMiddleware.RequireExportPermission(), h.ExportData)        // 导出数据

		// 问卷权限管理
		surveys.POST("/:id/permissions", PermissionMiddleware.RequireManagePermission(), h.SetPermissions)             // 设置权限
		surveys.GET("/:id/permissions", PermissionMiddleware.RequireManagePermission(), h.GetPermissions)              // 获取权限
		surveys.DELETE("/:id/permissions/:userId", PermissionMiddleware.RequireManagePermission(), h.RemovePermission) // 移除权限

		// 问卷文件管理
		surveys.POST("/:id/files", PermissionMiddleware.RequireEditPermission(), h.UploadFile)    // 上传文件
		surveys.GET("/:id/files", h.ListFiles)                                                    // 获取文件列表
		surveys.DELETE("/files/:fileId", PermissionMiddleware.RequireManagePermission(), h.DeleteFile) // 删除文件
	}

	// 问卷模板管理
	templates := router.Group("/templates")
	{
		templates.GET("", h.ListTemplates)                                             // 获取模板列表
		templates.POST("", h.CreateTemplate)                                           // 创建模板
		templates.GET("/:id", h.GetTemplate)                                           // 获取模板详情
		templates.PUT("/:id", PermissionMiddleware.SurveyOwnerPermission(), h.UpdateTemplate)  // 更新模板
		templates.DELETE("/:id", PermissionMiddleware.SurveyOwnerPermission(), h.DeleteTemplate) // 删除模板
	}

// RegisterPublicRoutes 注册公开访问路由（无需认证）
func (h *SurveyHandler) RegisterPublicRoutes(router *gin.RouterGroup) {
	// 公开问卷访问（无需认证）
	router.GET("/:id", PermissionMiddleware.PublicSurveyAccessible(), h.GetPublicSurvey)
	
	// 公开问卷响应提交（无需认证）
	router.POST("/:id/responses", PermissionMiddleware.PublicSurveyAccessible(), h.SubmitPublicResponse)
	
	// 公开问卷统计信息（无需认证）
	router.GET("/:id/stats", PermissionMiddleware.PublicSurveyAccessible(), h.GetPublicSurveyStats)
}
}
}

// CreateSurvey 创建问卷
func (h *SurveyHandler) CreateSurvey(c *gin.Context) {
	// 从OAuth中间件获取用户信息
	userID := c.GetString("user_id")
	teamID := c.GetString("team_id")
	
	// 如果没有获取到用户信息，尝试从上下文获取
	if userID == "" {
		if user, exists := c.Get("user"); exists {
			if u, ok := user.(map[string]interface{}); ok {
				if id, ok := u["id"].(string); ok {
					userID = id
				}
				if tid, ok := u["team_id"].(string); ok {
					teamID = tid
				}
			}
		}
	}
	
	if userID == "" || teamID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User authentication required"})
		return
	}

	var req CreateSurveyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data", "details": err.Error()})
		return
	}

	survey := &models.Survey{
		Title:          req.Title,
		Description:    req.Description,
		JsonDefinition: req.JsonDefinition,
		CreatedBy:      userID,
		TeamID:         teamID,
		Status:         "draft",
		IsPublic:       req.IsPublic,
		MaxResponses:   req.MaxResponses,
		StartTime:      req.StartTime,
		EndTime:        req.EndTime,
		Tags:           req.Tags,
	}

	if err := h.db.Create(survey).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create survey"})
		return
	}

	// 为创建者设置默认权限
	permission := &models.SurveyPermission{
		SurveyID:   survey.SurveyID,
		UserID:     userID,
		TeamID:     teamID,
		CanView:    true,
		CanEdit:    true,
		CanDelete:  true,
		CanManage:  true,
		CanAnalyze: true,
		CanExport:  true,
	}
	h.db.Create(permission)

	c.JSON(http.StatusCreated, survey)
}

// ListSurveys 获取问卷列表
func (h *SurveyHandler) ListSurveys(c *gin.Context) {
	// 从OAuth中间件获取用户信息
	userID := c.GetString("user_id")
	teamID := c.GetString("team_id")
	
	// 如果没有获取到用户信息，尝试从上下文获取
	if userID == "" {
		if user, exists := c.Get("user"); exists {
			if u, ok := user.(map[string]interface{}); ok {
				if id, ok := u["id"].(string); ok {
					userID = id
				}
				if tid, ok := u["team_id"].(string); ok {
					teamID = tid
				}
			}
		}
	}
	
	if userID == "" || teamID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User authentication required"})
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	status := c.Query("status")
	search := c.Query("search")

	offset := (page - 1) * pageSize

	query := h.db.Where("team_id = ? OR is_public = ?", teamID, true)

	if status != "" {
		query = query.Where("status = ?", status)
	}

	if search != "" {
		query = query.Where("title ILIKE ? OR description ILIKE ?", "%"+search+"%", "%"+search+"%")
	}

	var surveys []models.Survey
	var total int64

	query.Model(&models.Survey{}).Count(&total)
	if err := query.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&surveys).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get surveys"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"surveys": surveys,
		"total":   total,
		"page":    page,
		"pages":   (total + int64(pageSize) - 1) / int64(pageSize),
	})
}

// GetSurvey 获取问卷详情
func (h *SurveyHandler) GetSurvey(c *gin.Context) {
	surveyID := c.Param("id")
	// 从OAuth中间件获取用户信息
	userID := c.GetString("user_id")
	teamID := c.GetString("team_id")
	
	// 如果没有获取到用户信息，尝试从上下文获取
	if userID == "" {
		if user, exists := c.Get("user"); exists {
			if u, ok := user.(map[string]interface{}); ok {
				if id, ok := u["id"].(string); ok {
					userID = id
				}
				if tid, ok := u["team_id"].(string); ok {
					teamID = tid
				}
			}
		}
	}
	
	if userID == "" || teamID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User authentication required"})
		return
	}

	var survey models.Survey
	if err := h.db.Where("survey_id = ?", surveyID).First(&survey).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Survey not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get survey"})
		}
		return
	}

	// 检查权限
	if !survey.IsPublic && survey.TeamID != teamID {
		// 检查是否有特定权限
		var permission models.SurveyPermission
		if err := h.db.Where("survey_id = ? AND user_id = ? AND can_view = ?", surveyID, userID, true).First(&permission).Error; err != nil {
			c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
			return
		}
	}

	// 增加浏览次数
	h.db.Model(&survey).UpdateColumn("view_count", gorm.Expr("view_count + ?", 1))

	c.JSON(http.StatusOK, survey)
}

// UpdateSurvey 更新问卷
func (h *SurveyHandler) UpdateSurvey(c *gin.Context) {
	surveyID := c.Param("id")
	userID := c.GetString("user_id")

	var req UpdateSurveyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data"})
		return
	}

	var survey models.Survey
	if err := h.db.Where("survey_id = ?", surveyID).First(&survey).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Survey not found"})
		return
	}

	// 检查编辑权限
	if !h.hasEditPermission(surveyID, userID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "No edit permission"})
		return
	}

	// 更新字段
	updates := map[string]interface{}{}
	if req.Title != nil {
		updates["title"] = *req.Title
	}
	if req.Description != nil {
		updates["description"] = *req.Description
	}
	if req.JsonDefinition != nil {
		updates["json_definition"] = req.JsonDefinition
	}
	if req.Tags != nil {
		updates["tags"] = *req.Tags
	}
	if req.IsPublic != nil {
		updates["is_public"] = *req.IsPublic
	}
	if req.MaxResponses != nil {
		updates["max_responses"] = *req.MaxResponses
	}
	if req.StartTime != nil {
		updates["start_time"] = req.StartTime
	}
	if req.EndTime != nil {
		updates["end_time"] = req.EndTime
	}

	if err := h.db.Model(&survey).Updates(updates).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update survey"})
		return
	}

	c.JSON(http.StatusOK, survey)
}

// DeleteSurvey 删除问卷
func (h *SurveyHandler) DeleteSurvey(c *gin.Context) {
	surveyID := c.Param("id")
	userID := c.GetString("user_id")

	var survey models.Survey
	if err := h.db.Where("survey_id = ?", surveyID).First(&survey).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Survey not found"})
		return
	}

	// 检查删除权限
	if !h.hasDeletePermission(surveyID, userID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "No delete permission"})
		return
	}

	// 软删除问卷
	if err := h.db.Delete(&survey).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete survey"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Survey deleted successfully"})
}

// PublishSurvey 发布问卷
func (h *SurveyHandler) PublishSurvey(c *gin.Context) {
	surveyID := c.Param("id")
	userID := c.GetString("user_id")

	if !h.hasManagePermission(surveyID, userID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "No manage permission"})
		return
	}

	var survey models.Survey
	if err := h.db.Where("survey_id = ?", surveyID).First(&survey).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Survey not found"})
		return
	}

	if err := h.db.Model(&survey).Update("status", "active").Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to publish survey"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Survey published successfully"})
}

// CloseSurvey 关闭问卷
func (h *SurveyHandler) CloseSurvey(c *gin.Context) {
	surveyID := c.Param("id")
	userID := c.GetString("user_id")

	if !h.hasManagePermission(surveyID, userID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "No manage permission"})
		return
	}

	var survey models.Survey
	if err := h.db.Where("survey_id = ?", surveyID).First(&survey).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Survey not found"})
		return
	}

	if err := h.db.Model(&survey).Update("status", "closed").Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to close survey"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Survey closed successfully"})
}

// 权限检查辅助方法
func (h *SurveyHandler) hasEditPermission(surveyID, userID string) bool {
	var permission models.SurveyPermission
	err := h.db.Where("survey_id = ? AND user_id = ? AND can_edit = ?", surveyID, userID, true).First(&permission).Error
	return err == nil
}

func (h *SurveyHandler) hasDeletePermission(surveyID, userID string) bool {
	var permission models.SurveyPermission
	err := h.db.Where("survey_id = ? AND user_id = ? AND can_delete = ?", surveyID, userID, true).First(&permission).Error
	return err == nil
}

func (h *SurveyHandler) hasManagePermission(surveyID, userID string) bool {
	var permission models.SurveyPermission
	err := h.db.Where("survey_id = ? AND user_id = ? AND can_manage = ?", surveyID, userID, true).First(&permission).Error
	return err == nil
}

// ==================== 公开访问处理器方法 ====================

// GetPublicSurvey 获取公开问卷（无需认证）
func (h *SurveyHandler) GetPublicSurvey(c *gin.Context) {
	// 从中间件获取survey信息（已通过PublicSurveyAccessible检查）
	survey, exists := c.Get("survey")
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "Survey not found"})
		return
	}

	surveyObj := survey.(models.Survey)

	// 增加浏览次数
	h.db.Model(&surveyObj).UpdateColumn("view_count", gorm.Expr("view_count + ?", 1))

	// 只返回公开访问需要的字段
	response := gin.H{
		"survey_id":       surveyObj.SurveyID,
		"title":           surveyObj.Title,
		"description":     surveyObj.Description,
		"json_definition": surveyObj.JsonDefinition,
		"tags":            surveyObj.Tags,
		"view_count":      surveyObj.ViewCount + 1,
		"response_count":  surveyObj.ResponseCount,
		"max_responses":   surveyObj.MaxResponses,
		"start_time":      surveyObj.StartTime,
		"end_time":        surveyObj.EndTime,
		"status":          surveyObj.Status,
	}

	c.JSON(http.StatusOK, response)
}

// SubmitPublicResponse 提交公开问卷响应（无需认证）
func (h *SurveyHandler) SubmitPublicResponse(c *gin.Context) {
	// 从中间件获取survey信息
	survey, exists := c.Get("survey")
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "Survey not found"})
		return
	}

	surveyObj := survey.(models.Survey)

	// 检查问卷是否还在活跃状态
	if surveyObj.Status != "active" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Survey is not active"})
		return
	}

	// 检查是否超过最大响应数
	if surveyObj.MaxResponses > 0 && surveyObj.ResponseCount >= surveyObj.MaxResponses {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Survey has reached maximum responses"})
		return
	}

	var req SubmitResponseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data"})
		return
	}

	// 创建匿名响应记录
	response := &models.SurveyResponse{
		SurveyID:     surveyObj.SurveyID,
		ResponseData: req.ResponseData,
		IPAddress:    c.ClientIP(),
		UserAgent:    c.Request.UserAgent(),
		TimeSpent:    req.TimeSpent,
		IsAnonymous:  true, // 标记为匿名响应
	}

	// 开始事务
	tx := h.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 保存响应
	if err := tx.Create(response).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save response"})
		return
	}

	// 更新问卷响应计数
	if err := tx.Model(&surveyObj).UpdateColumn("response_count", gorm.Expr("response_count + ?", 1)).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update survey"})
		return
	}

	tx.Commit()

	// 异步触发AI分析（如果启用）
	if h.difyService != nil {
		go func() {
			if err := h.difyService.AnalyzeSurveyResponse(response); err != nil {
				// 记录分析失败日志
				// log.Printf("Failed to analyze survey response: %v", err)
			}
		}()
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":     "Response submitted successfully",
		"response_id": response.ResponseID,
	})
}

// GetPublicSurveyStats 获取公开问卷统计信息（无需认证）
func (h *SurveyHandler) GetPublicSurveyStats(c *gin.Context) {
	// 从中间件获取survey信息
	survey, exists := c.Get("survey")
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "Survey not found"})
		return
	}

	surveyObj := survey.(models.Survey)

	// 只返回基本的统计信息
	stats := gin.H{
		"survey_id":      surveyObj.SurveyID,
		"title":          surveyObj.Title,
		"view_count":     surveyObj.ViewCount,
		"response_count": surveyObj.ResponseCount,
		"max_responses":  surveyObj.MaxResponses,
		"status":         surveyObj.Status,
		"created_at":     surveyObj.CreatedAt,
	}

	// 如果问卷已结束，可以显示更多统计信息
	if surveyObj.Status == "closed" || surveyObj.Status == "completed" {
		// 计算完成率
		if surveyObj.ViewCount > 0 {
			completionRate := float64(surveyObj.ResponseCount) / float64(surveyObj.ViewCount) * 100
			stats["completion_rate"] = completionRate
		}
	}

	c.JSON(http.StatusOK, stats)
}