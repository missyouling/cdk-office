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

package dify

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/linux-do/cdk-office/internal/db"
	"github.com/linux-do/cdk-office/internal/models"
	"gorm.io/gorm"
)

// Handler Dify处理器
type Handler struct {
	db      *gorm.DB
	service *Service
}

// NewHandler 创建Dify处理器
func NewHandler(config *Config) *Handler {
	database := db.GetDB()
	service := NewService(config, database)

	return &Handler{
		db:      database,
		service: service,
	}
}

// RegisterRoutes 注册路由
func (h *Handler) RegisterRoutes(router *gin.RouterGroup) {
	dify := router.Group("/dify")
	{
		// 智能问答
		dify.POST("/chat", h.Chat)
		dify.POST("/chat/streaming", h.StreamingChat)
		dify.GET("/chat/history", h.GetChatHistory)

		// 知识库管理
		dify.POST("/documents/sync", h.SyncDocument)
		dify.DELETE("/documents/:id/sync", h.DeleteDocumentSync)
		dify.GET("/documents/:id/sync-status", h.GetDocumentSyncStatus)
		dify.POST("/documents/batch-sync", h.BatchSyncDocuments)

		// 工作流
		dify.POST("/workflows/:id/run", h.RunWorkflow)
		dify.GET("/workflows/:id/status", h.GetWorkflowStatus)

		// 问卷分析
		dify.POST("/surveys/:id/analyze", h.AnalyzeSurvey)

		// 配置管理
		dify.GET("/config", h.GetConfig)
		dify.PUT("/config", h.UpdateConfig)
		dify.POST("/config/test", h.TestConnection)

		// 统计信息
		dify.GET("/statistics", h.GetStatistics)
		dify.GET("/health", h.HealthCheck)
	}
}

// Chat 智能问答
// @Summary 智能问答
// @Description 向Dify发送问题并获得AI回答
// @Tags dify
// @Accept json
// @Produce json
// @Param request body KnowledgeQARequest true "问答请求"
// @Success 200 {object} KnowledgeQAResponse
// @Failure 400 {object} map[string]interface{}
// @Router /api/dify/chat [post]
func (h *Handler) Chat(c *gin.Context) {
	var req KnowledgeQARequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// 设置用户和团队信息
	req.UserID = c.GetString("user_id")
	req.TeamID = c.GetString("team_id")

	// 调用服务
	response, err := h.service.Chat(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

// StreamingChat 流式智能问答
func (h *Handler) StreamingChat(c *gin.Context) {
	var req KnowledgeQARequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	req.UserID = c.GetString("user_id")
	req.TeamID = c.GetString("team_id")

	// 设置SSE头
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("Access-Control-Allow-Origin", "*")

	// 创建响应writer
	writer := c.Writer
	flusher, ok := writer.(http.Flusher)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Streaming not supported"})
		return
	}

	// 调用流式服务
	err := h.service.StreamingChat(c.Request.Context(), &req, func(response *StreamingChatResponse) error {
		data, _ := json.Marshal(response)
		fmt.Fprintf(writer, "data: %s\n\n", string(data))
		flusher.Flush()
		return nil
	})

	if err != nil {
		fmt.Fprintf(writer, "data: {\"error\": \"%s\"}\n\n", err.Error())
		flusher.Flush()
	}

	fmt.Fprintf(writer, "data: [DONE]\n\n")
	flusher.Flush()
}

// GetChatHistory 获取聊天历史
func (h *Handler) GetChatHistory(c *gin.Context) {
	userID := c.GetString("user_id")
	teamID := c.GetString("team_id")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	var history []models.KnowledgeQA
	var total int64

	query := h.db.Model(&models.KnowledgeQA{}).Where("user_id = ? AND team_id = ?", userID, teamID)

	// 获取总数
	query.Count(&total)

	// 分页查询
	offset := (page - 1) * limit
	if err := query.Offset(offset).Limit(limit).Order("created_at DESC").Find(&history).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch chat history"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"history": history,
		"total":   total,
		"page":    page,
		"limit":   limit,
	})
}

// SyncDocument 同步文档到Dify
func (h *Handler) SyncDocument(c *gin.Context) {
	var req DocumentSyncRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	req.TeamID = c.GetString("team_id")
	req.CreatedBy = c.GetString("user_id")

	response, err := h.service.SyncDocument(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

// DeleteDocumentSync 删除文档同步
func (h *Handler) DeleteDocumentSync(c *gin.Context) {
	documentID := c.Param("id")

	err := h.service.DeleteDocument(c.Request.Context(), documentID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Document sync deleted successfully"})
}

// GetDocumentSyncStatus 获取文档同步状态
func (h *Handler) GetDocumentSyncStatus(c *gin.Context) {
	documentID := c.Param("id")

	status, err := h.service.GetDocumentSyncStatus(c.Request.Context(), documentID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, status)
}

// BatchSyncDocuments 批量同步文档
func (h *Handler) BatchSyncDocuments(c *gin.Context) {
	var req struct {
		DocumentIDs []string `json:"document_ids" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	err := h.service.BatchSyncDocuments(c.Request.Context(), req.DocumentIDs)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Batch sync started successfully"})
}

// RunWorkflow 运行工作流
func (h *Handler) RunWorkflow(c *gin.Context) {
	workflowID := c.Param("id")

	var req struct {
		Inputs map[string]interface{} `json:"inputs"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	workflowReq := &WorkflowRunRequest{
		Inputs:       req.Inputs,
		ResponseMode: "blocking",
		User:         c.GetString("user_id"),
	}

	// 这里应该根据workflowID调用不同的工作流
	// 为了简化，直接调用客户端
	client := NewClient(h.service.config.BaseURL, h.service.config.APIKey)
	response, err := client.RunWorkflow(c.Request.Context(), workflowReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

// GetWorkflowStatus 获取工作流状态
func (h *Handler) GetWorkflowStatus(c *gin.Context) {
	workflowID := c.Param("id")

	// 这里应该实现获取工作流状态的逻辑
	// 目前返回模拟数据
	c.JSON(http.StatusOK, gin.H{
		"workflow_id": workflowID,
		"status":      "unknown",
		"message":     "Workflow status endpoint not implemented",
	})
}

// AnalyzeSurvey 分析问卷
func (h *Handler) AnalyzeSurvey(c *gin.Context) {
	surveyID := c.Param("id")

	var req struct {
		Responses []map[string]interface{} `json:"responses" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	response, err := h.service.RunSurveyAnalysis(c.Request.Context(), surveyID, req.Responses)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

// GetConfig 获取Dify配置
func (h *Handler) GetConfig(c *gin.Context) {
	// 隐藏敏感信息
	config := map[string]interface{}{
		"base_url":                     h.service.config.BaseURL,
		"default_dataset_id":           h.service.config.DefaultDatasetID,
		"survey_analysis_workflow_id":  h.service.config.SurveyAnalysisWorkflowID,
		"document_process_workflow_id": h.service.config.DocumentProcessWorkflowID,
		"knowledge_base_id":            h.service.config.KnowledgeBaseID,
		"enable_auto_sync":             h.service.config.EnableAutoSync,
		"sync_interval":                h.service.config.SyncInterval,
		"api_key_configured":           h.service.config.APIKey != "",
	}

	c.JSON(http.StatusOK, gin.H{"config": config})
}

// UpdateConfig 更新Dify配置
func (h *Handler) UpdateConfig(c *gin.Context) {
	var req Config
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// 更新配置
	h.service.config = &req

	// 重新创建客户端
	h.service.client = NewClient(req.BaseURL, req.APIKey)

	c.JSON(http.StatusOK, gin.H{"message": "Configuration updated successfully"})
}

// TestConnection 测试Dify连接
func (h *Handler) TestConnection(c *gin.Context) {
	// 发送一个简单的测试请求
	testReq := &ChatRequest{
		Query:        "Hello",
		ResponseMode: "blocking",
		User:         "test",
		Inputs:       make(map[string]interface{}),
	}

	_, err := h.service.client.Chat(c.Request.Context(), testReq)
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status": "failed",
			"error":  err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Connection to Dify is working",
	})
}

// GetStatistics 获取统计信息
func (h *Handler) GetStatistics(c *gin.Context) {
	teamID := c.GetString("team_id")

	// 问答统计
	var qaCount int64
	h.db.Model(&models.KnowledgeQA{}).Where("team_id = ?", teamID).Count(&qaCount)

	// 同步文档统计
	var syncCount int64
	h.db.Model(&models.DifyDocumentSync{}).Where("team_id = ?", teamID).Count(&syncCount)

	var syncedCount int64
	h.db.Model(&models.DifyDocumentSync{}).Where("team_id = ? AND sync_status = ?", teamID, "synced").Count(&syncedCount)

	// 今日问答统计
	var todayQACount int64
	h.db.Model(&models.KnowledgeQA{}).Where("team_id = ? AND DATE(created_at) = CURRENT_DATE", teamID).Count(&todayQACount)

	c.JSON(http.StatusOK, gin.H{
		"qa_total":         qaCount,
		"qa_today":         todayQACount,
		"documents_total":  syncCount,
		"documents_synced": syncedCount,
		"sync_rate":        float64(syncedCount) / float64(syncCount) * 100,
	})
}

// HealthCheck 健康检查
func (h *Handler) HealthCheck(c *gin.Context) {
	// 检查数据库连接
	sqlDB, err := h.db.DB()
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"status": "unhealthy", "error": "database connection failed"})
		return
	}

	if err := sqlDB.Ping(); err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"status": "unhealthy", "error": "database ping failed"})
		return
	}

	// 检查Dify连接（可选）
	testReq := &ChatRequest{
		Query:        "Health check",
		ResponseMode: "blocking",
		User:         "system",
		Inputs:       make(map[string]interface{}),
	}

	difyStatus := "unknown"
	if h.service.config.APIKey != "" {
		if _, err := h.service.client.Chat(c.Request.Context(), testReq); err == nil {
			difyStatus = "healthy"
		} else {
			difyStatus = "unhealthy"
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"status":      "healthy",
		"dify_status": difyStatus,
		"timestamp":   gin.H{},
	})
}
