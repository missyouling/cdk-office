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

package filepreview

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/linux-do/cdk-office/internal/db"
	"github.com/linux-do/cdk-office/internal/models"
	"gorm.io/gorm"
)

// Handler 文件预览请求处理器
type Handler struct {
	db      *gorm.DB
	service *Service
}

// NewHandler 创建文件预览请求处理器
func NewHandler() *Handler {
	database := db.GetDB()

	// 从配置中读取文件预览服务配置
	config := &Config{
		Provider: "dify", // 默认使用Dify原生预览
		DifyURL:  "http://dify-api:5001",
		KKFileView: KKFileViewConfig{
			Enabled: false, // 默认不启用KKFileView
			URL:     "http://kkfileview:8012",
			Timeout: 30,
		},
	}

	// 可以从环境变量或配置文件中读取配置
	// config = loadConfigFromEnv(config)

	service := NewService(config, database)

	return &Handler{
		db:      database,
		service: service,
	}
}

// RegisterRoutes 注册路由
func (h *Handler) RegisterRoutes(router *gin.RouterGroup) {
	preview := router.Group("/preview")
	{
		// 文件预览接口
		preview.POST("/generate", h.GeneratePreview)
		preview.GET("/url/:documentId", h.GetPreviewURL)
		preview.GET("/supported-types", h.GetSupportedTypes)
		preview.GET("/check/:filename", h.CheckSupport)

		// 预览历史
		preview.GET("/history", h.GetPreviewHistory)
		preview.DELETE("/history/:id", h.DeletePreviewRecord)

		// 服务状态
		preview.GET("/health", h.HealthCheck)
		preview.GET("/config", h.GetConfig)

		// 管理接口
		preview.POST("/cleanup", h.CleanupHistory)
	}
}

// GeneratePreview 生成文件预览
// @Summary 生成文件预览
// @Description 为指定文件生成预览链接
// @Tags file-preview
// @Accept json
// @Produce json
// @Param request body PreviewRequest true "预览请求"
// @Success 200 {object} PreviewResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/preview/generate [post]
func (h *Handler) GeneratePreview(c *gin.Context) {
	var req PreviewRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// 设置用户信息
	req.UserID = c.GetString("user_id")
	req.TeamID = c.GetString("team_id")

	// 验证请求参数
	if req.DocumentID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Document ID is required"})
		return
	}

	if req.FileName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "File name is required"})
		return
	}

	// 从文件名获取文件类型
	if req.FileType == "" {
		req.FileType = GetFileExtension(req.FileName)
	}

	// 生成预览
	response, err := h.service.Preview(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

// GetPreviewURL 获取预览URL
// @Summary 获取文件预览URL
// @Description 根据文档ID获取预览URL
// @Tags file-preview
// @Produce json
// @Param documentId path string true "文档ID"
// @Param fileType query string false "文件类型"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Router /api/preview/url/{documentId} [get]
func (h *Handler) GetPreviewURL(c *gin.Context) {
	documentID := c.Param("documentId")
	fileType := c.Query("fileType")

	if documentID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Document ID is required"})
		return
	}

	// 构建简单的预览请求
	req := &PreviewRequest{
		DocumentID: documentID,
		FileType:   fileType,
		UserID:     c.GetString("user_id"),
		TeamID:     c.GetString("team_id"),
	}

	response, err := h.service.Preview(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"preview_url": response.PreviewURL,
		"provider":    response.Provider,
		"supported":   response.Supported,
	})
}

// GetSupportedTypes 获取支持的文件类型
// @Summary 获取支持的文件类型
// @Description 获取当前预览服务支持的所有文件类型
// @Tags file-preview
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /api/preview/supported-types [get]
func (h *Handler) GetSupportedTypes(c *gin.Context) {
	supportedTypes := h.service.GetSupportedTypes()

	c.JSON(http.StatusOK, gin.H{
		"supported_types": supportedTypes,
		"provider":        h.service.config.Provider,
		"total_count":     len(supportedTypes),
	})
}

// CheckSupport 检查文件是否支持预览
// @Summary 检查文件预览支持
// @Description 检查指定文件名是否支持预览
// @Tags file-preview
// @Produce json
// @Param filename path string true "文件名"
// @Success 200 {object} map[string]interface{}
// @Router /api/preview/check/{filename} [get]
func (h *Handler) CheckSupport(c *gin.Context) {
	filename := c.Param("filename")

	if filename == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Filename is required"})
		return
	}

	supported := h.service.IsPreviewSupported(filename)
	fileType := GetFileExtension(filename)

	c.JSON(http.StatusOK, gin.H{
		"filename":  filename,
		"file_type": fileType,
		"supported": supported,
		"provider":  h.service.config.Provider,
	})
}

// GetPreviewHistory 获取预览历史
// @Summary 获取预览历史
// @Description 获取当前用户的文件预览历史记录
// @Tags file-preview
// @Produce json
// @Param page query int false "页码" default(1)
// @Param limit query int false "每页数量" default(10)
// @Success 200 {object} map[string]interface{}
// @Router /api/preview/history [get]
func (h *Handler) GetPreviewHistory(c *gin.Context) {
	userID := c.GetString("user_id")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	if limit > 100 {
		limit = 100 // 限制最大页面大小
	}

	history, err := h.service.GetPreviewHistory(userID, page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch preview history"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"history": history,
		"page":    page,
		"limit":   limit,
	})
}

// DeletePreviewRecord 删除预览记录
func (h *Handler) DeletePreviewRecord(c *gin.Context) {
	recordID := c.Param("id")
	userID := c.GetString("user_id")

	// 验证记录所有权
	var preview models.FilePreview
	if err := h.db.Where("id = ? AND user_id = ?", recordID, userID).First(&preview).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Preview record not found"})
		return
	}

	// 删除记录
	if err := h.db.Delete(&preview).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete preview record"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Preview record deleted successfully"})
}

// HealthCheck 健康检查
// @Summary 文件预览服务健康检查
// @Description 检查文件预览服务的健康状态
// @Tags file-preview
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 503 {object} map[string]interface{}
// @Router /api/preview/health [get]
func (h *Handler) HealthCheck(c *gin.Context) {
	err := h.service.HealthCheck()
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status":   "unhealthy",
			"provider": h.service.config.Provider,
			"error":    err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":    "healthy",
		"provider":  h.service.config.Provider,
		"timestamp": gin.H{"checked_at": gin.H{}},
	})
}

// GetConfig 获取预览服务配置
func (h *Handler) GetConfig(c *gin.Context) {
	config := map[string]interface{}{
		"provider":        h.service.config.Provider,
		"supported_types": h.service.GetSupportedTypes(),
	}

	// 根据提供者返回相应配置
	switch h.service.config.Provider {
	case "kkfileview":
		config["kkfileview"] = map[string]interface{}{
			"enabled": h.service.config.KKFileView.Enabled,
			"url":     h.service.config.KKFileView.URL,
		}
	case "dify":
		config["dify"] = map[string]interface{}{
			"url": h.service.config.DifyURL,
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"config": config,
	})
}

// CleanupHistory 清理预览历史
func (h *Handler) CleanupHistory(c *gin.Context) {
	// 只允许管理员执行清理操作
	userRole := c.GetString("user_role")
	if userRole != "admin" && userRole != "super_admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Insufficient permissions"})
		return
	}

	var req struct {
		OlderThanDays int `json:"older_than_days" binding:"required,min=1"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	err := h.service.CleanupPreviewHistory(req.OlderThanDays)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":         "Preview history cleaned up successfully",
		"older_than_days": req.OlderThanDays,
	})
}

// PreviewFileRequest 预览文件请求结构
type PreviewFileRequest struct {
	DocumentID string `json:"document_id" binding:"required"`
	FileName   string `json:"file_name" binding:"required"`
	FileURL    string `json:"file_url" binding:"required"`
	FileType   string `json:"file_type"`
}

// PreviewFile 预览文件（通用接口）
// @Summary 预览文件
// @Description 通用文件预览接口，支持多种文件格式
// @Tags file-preview
// @Accept json
// @Produce json
// @Param request body PreviewFileRequest true "预览文件请求"
// @Success 200 {object} PreviewResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/preview/file [post]
func (h *Handler) PreviewFile(c *gin.Context) {
	var req PreviewFileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// 构建预览请求
	previewReq := &PreviewRequest{
		DocumentID: req.DocumentID,
		FileName:   req.FileName,
		FileType:   req.FileType,
		FileURL:    req.FileURL,
		UserID:     c.GetString("user_id"),
		TeamID:     c.GetString("team_id"),
	}

	// 从文件名获取文件类型（如果未提供）
	if previewReq.FileType == "" {
		previewReq.FileType = GetFileExtension(req.FileName)
	}

	// 生成预览
	response, err := h.service.Preview(c.Request.Context(), previewReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}
