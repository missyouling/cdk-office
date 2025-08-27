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
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/linux-do/cdk-office/internal/models"
	"gorm.io/gorm"
)

// Router AI服务路由
type Router struct {
	serviceManager *ServiceManager
	db             *gorm.DB
}

// NewRouter 创建AI服务路由
func NewRouter(db *gorm.DB) *Router {
	serviceManager := NewServiceManager(db)
	serviceManager.StartHealthCheckRoutine()

	return &Router{
		serviceManager: serviceManager,
		db:             db,
	}
}

// RegisterRoutes 注册路由
func (r *Router) RegisterRoutes(rg *gin.RouterGroup) {
	ai := rg.Group("/ai")
	{
		// 智能问答
		ai.POST("/chat", r.Chat)
		ai.POST("/embedding", r.Embedding)
		ai.POST("/translate", r.Translate)

		// 服务配置管理（需要管理员权限）
		ai.GET("/services", r.GetServiceList)
		ai.POST("/services", r.CreateServiceConfig)
		ai.PUT("/services/:id", r.UpdateServiceConfig)
		ai.DELETE("/services/:id", r.DeleteServiceConfig)
		ai.POST("/services/:id/test", r.TestServiceConnection)

		// 健康检查
		ai.GET("/health", r.GetServiceHealth)
		ai.POST("/health/:id/check", r.CheckServiceNow)

		// 预设服务商
		ai.GET("/providers", r.GetPresetProviders)

		// 知识库管理
		ai.POST("/knowledge/sync", r.SyncKnowledge)
		ai.GET("/knowledge/qa", r.GetKnowledgeQA)
		ai.POST("/knowledge/feedback", r.SubmitFeedback)
	}
}

// Chat 智能问答
// @Summary 智能问答
// @Description 发送问题给AI系统获取回答
// @Tags AI
// @Accept json
// @Produce json
// @Param request body ChatRequest true "问答请求"
// @Success 200 {object} ChatResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/ai/chat [post]
func (r *Router) Chat(c *gin.Context) {
	var req ChatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 获取用户信息
	userID := c.GetHeader("X-User-ID")
	teamID := c.GetHeader("X-Team-ID")

	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// 获取默认AI对话服务
	service, err := r.serviceManager.GetDefaultService("ai_chat")
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "AI chat service not available"})
		return
	}

	// 创建客户端并发送请求
	client := r.serviceManager.createClient(service)

	chatReq := map[string]interface{}{
		"messages": []map[string]string{
			{"role": "user", "content": req.Question},
		},
		"max_tokens": 500,
	}

	response, err := client.Chat(c.Request.Context(), chatReq)
	if err != nil {
		// 如果主服务失败，尝试降级
		if fallbackErr := r.serviceManager.TriggerFallback(service.ID); fallbackErr == nil {
			// 重新获取服务并重试
			if fallbackService, getErr := r.serviceManager.GetDefaultService("ai_chat"); getErr == nil {
				fallbackClient := r.serviceManager.createClient(fallbackService)
				if response, err = fallbackClient.Chat(c.Request.Context(), chatReq); err == nil {
					// 降级成功，记录QA
					r.recordKnowledgeQA(userID, teamID, req.Question, response, fallbackService.ServiceName)
					c.JSON(http.StatusOK, r.formatChatResponse(response))
					return
				}
			}
		}

		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get AI response"})
		return
	}

	// 记录问答
	r.recordKnowledgeQA(userID, teamID, req.Question, response, service.ServiceName)

	c.JSON(http.StatusOK, r.formatChatResponse(response))
}

// Embedding 文本向量化
// @Summary 文本向量化
// @Description 将文本转换为向量表示
// @Tags AI
// @Accept json
// @Produce json
// @Param request body EmbeddingRequest true "向量化请求"
// @Success 200 {object} EmbeddingResponse
// @Router /api/ai/embedding [post]
func (r *Router) Embedding(c *gin.Context) {
	var req EmbeddingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	service, err := r.serviceManager.GetDefaultService("ai_embedding")
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "AI embedding service not available"})
		return
	}

	client := r.serviceManager.createClient(service)

	embeddingReq := map[string]interface{}{
		"input": req.Text,
	}

	response, err := client.Embedding(c.Request.Context(), embeddingReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get embedding"})
		return
	}

	c.JSON(http.StatusOK, response)
}

// Translate 文本翻译
// @Summary 文本翻译
// @Description 翻译文本到指定语言
// @Tags AI
// @Accept json
// @Produce json
// @Param request body TranslateRequest true "翻译请求"
// @Success 200 {object} TranslateResponse
// @Router /api/ai/translate [post]
func (r *Router) Translate(c *gin.Context) {
	var req TranslateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	service, err := r.serviceManager.GetDefaultService("ai_translation")
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "AI translation service not available"})
		return
	}

	client := r.serviceManager.createClient(service)

	translateReq := map[string]interface{}{
		"text": req.Text,
		"from": req.From,
		"to":   req.To,
	}

	response, err := client.Translate(c.Request.Context(), translateReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to translate"})
		return
	}

	c.JSON(http.StatusOK, response)
}

// GetServiceList 获取服务列表
// @Summary 获取AI服务列表
// @Description 获取所有配置的AI服务列表
// @Tags AI
// @Accept json
// @Produce json
// @Param type query string false "服务类型过滤"
// @Success 200 {object} ServiceListResponse
// @Router /api/ai/services [get]
func (r *Router) GetServiceList(c *gin.Context) {
	serviceType := c.Query("type")

	services, err := r.serviceManager.GetServiceList(serviceType)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"services": services,
		"total":    len(services),
	})
}

// CreateServiceConfig 创建服务配置
// @Summary 创建AI服务配置
// @Description 创建新的AI服务配置（需要管理员权限）
// @Tags AI
// @Accept json
// @Produce json
// @Param request body models.AIServiceConfig true "服务配置"
// @Success 201 {object} models.AIServiceConfig
// @Router /api/ai/services [post]
func (r *Router) CreateServiceConfig(c *gin.Context) {
	userID := c.GetHeader("X-User-ID")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	var config models.AIServiceConfig
	if err := c.ShouldBindJSON(&config); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := r.serviceManager.CreateServiceConfig(userID, &config); err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, config)
}

// UpdateServiceConfig 更新服务配置
// @Summary 更新AI服务配置
// @Description 更新AI服务配置（需要管理员权限）
// @Tags AI
// @Accept json
// @Produce json
// @Param id path string true "服务ID"
// @Param request body map[string]interface{} true "更新内容"
// @Success 200 {object} SuccessResponse
// @Router /api/ai/services/{id} [put]
func (r *Router) UpdateServiceConfig(c *gin.Context) {
	userID := c.GetHeader("X-User-ID")
	configID := c.Param("id")

	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	var updates map[string]interface{}
	if err := c.ShouldBindJSON(&updates); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := r.serviceManager.UpdateServiceConfig(userID, configID, updates); err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Service config updated successfully"})
}

// DeleteServiceConfig 删除服务配置
// @Summary 删除AI服务配置
// @Description 删除AI服务配置（需要管理员权限）
// @Tags AI
// @Accept json
// @Produce json
// @Param id path string true "服务ID"
// @Success 200 {object} SuccessResponse
// @Router /api/ai/services/{id} [delete]
func (r *Router) DeleteServiceConfig(c *gin.Context) {
	userID := c.GetHeader("X-User-ID")
	configID := c.Param("id")

	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	if err := r.serviceManager.DeleteServiceConfig(userID, configID); err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Service config deleted successfully"})
}

// TestServiceConnection 测试服务连接
// @Summary 测试AI服务连接
// @Description 测试AI服务连接状态
// @Tags AI
// @Accept json
// @Produce json
// @Param id path string true "服务ID"
// @Success 200 {object} SuccessResponse
// @Router /api/ai/services/{id}/test [post]
func (r *Router) TestServiceConnection(c *gin.Context) {
	userID := c.GetHeader("X-User-ID")
	configID := c.Param("id")

	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	if err := r.serviceManager.TestServiceConnection(userID, configID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Service connection test successful"})
}

// GetServiceHealth 获取服务健康状态
// @Summary 获取服务健康状态
// @Description 获取所有AI服务的健康状态
// @Tags AI
// @Accept json
// @Produce json
// @Success 200 {object} HealthStatusResponse
// @Router /api/ai/health [get]
func (r *Router) GetServiceHealth(c *gin.Context) {
	var statuses []models.ServiceStatus
	if err := r.db.Find(&statuses).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"statuses": statuses,
		"total":    len(statuses),
	})
}

// CheckServiceNow 立即检查服务
// @Summary 立即检查服务健康状态
// @Description 立即检查指定服务的健康状态
// @Tags AI
// @Accept json
// @Produce json
// @Param id path string true "服务ID"
// @Success 200 {object} HealthCheckResult
// @Router /api/ai/health/{id}/check [post]
func (r *Router) CheckServiceNow(c *gin.Context) {
	serviceID := c.Param("id")

	result, err := r.serviceManager.healthChecker.CheckServiceNow(serviceID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

// GetPresetProviders 获取预设服务商
// @Summary 获取预设服务商配置
// @Description 获取预定义的AI服务商配置模板
// @Tags AI
// @Accept json
// @Produce json
// @Success 200 {object} PresetServiceProviders
// @Router /api/ai/providers [get]
func (r *Router) GetPresetProviders(c *gin.Context) {
	providers := r.serviceManager.GetPresetProviders()
	c.JSON(http.StatusOK, providers)
}

// SyncKnowledge 同步知识库
func (r *Router) SyncKnowledge(c *gin.Context) {
	// TODO: 实现知识库同步功能
	c.JSON(http.StatusOK, gin.H{"message": "Knowledge sync feature coming soon"})
}

// GetKnowledgeQA 获取问答记录
func (r *Router) GetKnowledgeQA(c *gin.Context) {
	userID := c.GetHeader("X-User-ID")
	teamID := c.GetHeader("X-Team-ID")

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset := (page - 1) * limit

	var qas []models.KnowledgeQA
	query := r.db.Where("user_id = ?", userID)
	if teamID != "" {
		query = query.Where("team_id = ?", teamID)
	}

	if err := query.Order("created_at DESC").Offset(offset).Limit(limit).Find(&qas).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"qas":   qas,
		"page":  page,
		"limit": limit,
	})
}

// SubmitFeedback 提交反馈
func (r *Router) SubmitFeedback(c *gin.Context) {
	var req FeedbackRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 更新问答记录的反馈
	if err := r.db.Model(&models.KnowledgeQA{}).
		Where("id = ?", req.QAID).
		Update("feedback", req.Feedback).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Feedback submitted successfully"})
}

// 辅助方法

func (r *Router) recordKnowledgeQA(userID, teamID, question string, response map[string]interface{}, provider string) {
	// 提取答案
	answer := ""
	if choices, ok := response["choices"].([]interface{}); ok && len(choices) > 0 {
		if choice, ok := choices[0].(map[string]interface{}); ok {
			if message, ok := choice["message"].(map[string]interface{}); ok {
				if content, ok := message["content"].(string); ok {
					answer = content
				}
			}
		}
	}

	qa := &models.KnowledgeQA{
		UserID:     userID,
		TeamID:     teamID,
		Question:   question,
		Answer:     answer,
		AIProvider: provider,
	}

	r.db.Create(qa)
}

func (r *Router) formatChatResponse(response map[string]interface{}) gin.H {
	// 格式化响应为统一格式
	answer := ""
	if choices, ok := response["choices"].([]interface{}); ok && len(choices) > 0 {
		if choice, ok := choices[0].(map[string]interface{}); ok {
			if message, ok := choice["message"].(map[string]interface{}); ok {
				if content, ok := message["content"].(string); ok {
					answer = content
				}
			}
		}
	}

	return gin.H{
		"answer":     answer,
		"sources":    []string{},
		"confidence": 0.9,
	}
}

// 请求和响应结构

type ChatRequest struct {
	Question string `json:"question" binding:"required"`
}

type ChatResponse struct {
	Answer     string   `json:"answer"`
	Sources    []string `json:"sources"`
	Confidence float64  `json:"confidence"`
}

type EmbeddingRequest struct {
	Text string `json:"text" binding:"required"`
}

type EmbeddingResponse struct {
	Embedding []float64 `json:"embedding"`
}

type TranslateRequest struct {
	Text string `json:"text" binding:"required"`
	From string `json:"from" binding:"required"`
	To   string `json:"to" binding:"required"`
}

type TranslateResponse struct {
	TranslatedText string `json:"translated_text"`
}

type FeedbackRequest struct {
	QAID     string `json:"qa_id" binding:"required"`
	Feedback string `json:"feedback" binding:"required,oneof=positive negative"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

type SuccessResponse struct {
	Message string `json:"message"`
}
