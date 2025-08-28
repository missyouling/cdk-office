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
)

// Handler AI模块HTTP处理器
type Handler struct {
	service *Service
}

// NewHandler 创建Handler实例
func NewHandler(service *Service) *Handler {
	return &Handler{
		service: service,
	}
}

// Chat 智能问答接口
// @Summary 智能问答
// @Description 通过AI进行智能问答，支持上下文对话
// @Tags AI
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param request body ChatRequest true "问答请求"
// @Success 200 {object} ChatResponse "问答响应"
// @Failure 400 {object} models.ErrorResponse "请求参数错误"
// @Failure 401 {object} models.ErrorResponse "未授权"
// @Failure 403 {object} models.ErrorResponse "权限不足"
// @Failure 500 {object} models.ErrorResponse "服务器内部错误"
// @Router /api/ai/chat [post]
func (h *Handler) Chat(c *gin.Context) {
	// 获取用户信息
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Code:    "UNAUTHORIZED",
			Message: "用户未登录",
		})
		return
	}

	teamID, exists := c.Get("team_id")
	if !exists {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Code:    "INVALID_TEAM",
			Message: "无效的团队信息",
		})
		return
	}

	// 解析请求
	var req ChatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Code:    "INVALID_REQUEST",
			Message: "请求参数错误: " + err.Error(),
		})
		return
	}

	// 调用服务
	response, err := h.service.Chat(c.Request.Context(), userID.(string), teamID.(string), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Code:    "AI_SERVICE_ERROR",
			Message: "AI服务出错: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, response)
}

// GetChatHistory 获取问答历史
// @Summary 获取问答历史
// @Description 获取用户的问答历史记录
// @Tags AI
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param page query int false "页码" default(1)
// @Param size query int false "每页数量" default(20)
// @Success 200 {object} ChatHistoryResponse "问答历史"
// @Failure 400 {object} models.ErrorResponse "请求参数错误"
// @Failure 401 {object} models.ErrorResponse "未授权"
// @Failure 500 {object} models.ErrorResponse "服务器内部错误"
// @Router /api/ai/chat/history [get]
func (h *Handler) GetChatHistory(c *gin.Context) {
	// 获取用户信息
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Code:    "UNAUTHORIZED",
			Message: "用户未登录",
		})
		return
	}

	teamID, exists := c.Get("team_id")
	if !exists {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Code:    "INVALID_TEAM",
			Message: "无效的团队信息",
		})
		return
	}

	// 解析分页参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "20"))

	if page < 1 {
		page = 1
	}
	if size < 1 || size > 100 {
		size = 20
	}

	offset := (page - 1) * size

	// 调用服务
	history, total, err := h.service.GetChatHistory(c.Request.Context(), userID.(string), teamID.(string), size, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Code:    "SERVICE_ERROR",
			Message: "获取历史记录失败: " + err.Error(),
		})
		return
	}

	response := ChatHistoryResponse{
		Data: history,
		Pagination: PaginationInfo{
			Page:  page,
			Size:  size,
			Total: total,
		},
	}

	c.JSON(http.StatusOK, response)
}

// UpdateFeedback 更新问答反馈
// @Summary 更新问答反馈
// @Description 用户对AI回答进行反馈评价
// @Tags AI
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param message_id path string true "消息ID"
// @Param request body FeedbackRequest true "反馈请求"
// @Success 200 {object} models.SuccessResponse "成功响应"
// @Failure 400 {object} models.ErrorResponse "请求参数错误"
// @Failure 401 {object} models.ErrorResponse "未授权"
// @Failure 404 {object} models.ErrorResponse "记录不存在"
// @Failure 500 {object} models.ErrorResponse "服务器内部错误"
// @Router /api/ai/chat/{message_id}/feedback [patch]
func (h *Handler) UpdateFeedback(c *gin.Context) {
	// 获取用户信息
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Code:    "UNAUTHORIZED",
			Message: "用户未登录",
		})
		return
	}

	// 获取消息ID
	messageID := c.Param("message_id")
	if messageID == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Code:    "INVALID_MESSAGE_ID",
			Message: "消息ID不能为空",
		})
		return
	}

	// 解析请求
	var req FeedbackRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Code:    "INVALID_REQUEST",
			Message: "请求参数错误: " + err.Error(),
		})
		return
	}

	// 调用服务
	err := h.service.UpdateFeedback(c.Request.Context(), userID.(string), messageID, req.Feedback)
	if err != nil {
		if err.Error() == "knowledge QA record not found" {
			c.JSON(http.StatusNotFound, models.ErrorResponse{
				Code:    "RECORD_NOT_FOUND",
				Message: "问答记录不存在",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Code:    "SERVICE_ERROR",
			Message: "更新反馈失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse{
		Code:    "SUCCESS",
		Message: "反馈更新成功",
	})
}

// GetStats 获取问答统计
// @Summary 获取问答统计
// @Description 获取团队的问答统计信息
// @Tags AI
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {object} ChatStats "统计信息"
// @Failure 401 {object} models.ErrorResponse "未授权"
// @Failure 500 {object} models.ErrorResponse "服务器内部错误"
// @Router /api/ai/chat/stats [get]
func (h *Handler) GetStats(c *gin.Context) {
	// 获取团队信息
	teamID, exists := c.Get("team_id")
	if !exists {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Code:    "INVALID_TEAM",
			Message: "无效的团队信息",
		})
		return
	}

	// 调用服务
	stats, err := h.service.GetStats(c.Request.Context(), teamID.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Code:    "SERVICE_ERROR",
			Message: "获取统计信息失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, stats)
}

// RegisterRoutes 注册路由
func (h *Handler) RegisterRoutes(router *gin.RouterGroup) {
	ai := router.Group("/ai")
	{
		ai.POST("/chat", h.Chat)
		ai.GET("/chat/history", h.GetChatHistory)
		ai.PATCH("/chat/:message_id/feedback", h.UpdateFeedback)
		ai.GET("/chat/stats", h.GetStats)
	}
}

// 响应结构体定义

// ChatHistoryResponse 问答历史响应
type ChatHistoryResponse struct {
	Data       []*models.KnowledgeQA `json:"data"`
	Pagination PaginationInfo        `json:"pagination"`
}

// PaginationInfo 分页信息
type PaginationInfo struct {
	Page  int   `json:"page"`
	Size  int   `json:"size"`
	Total int64 `json:"total"`
}

// FeedbackRequest 反馈请求
type FeedbackRequest struct {
	Feedback string `json:"feedback" binding:"required"`
}

// ChatStats 问答统计
type ChatStats struct {
	TotalChats     int64            `json:"total_chats"`
	TodayChats     int64            `json:"today_chats"`
	AvgConfidence  float32          `json:"avg_confidence"`
	TopUsers       []UserActivity   `json:"top_users"`
	RecentActivity []ActivityRecord `json:"recent_activity"`
}

// UserActivity 用户活动统计
type UserActivity struct {
	UserID   string `json:"user_id"`
	UserName string `json:"user_name"`
	Count    int64  `json:"count"`
}

// ActivityRecord 活动记录
type ActivityRecord struct {
	Date  string `json:"date"`
	Count int64  `json:"count"`
}
