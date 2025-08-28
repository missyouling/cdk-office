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

package dashboard

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// Handler Dashboard HTTP处理器
type Handler struct {
	service *Service
}

// NewHandler 创建Dashboard HTTP处理器实例
func NewHandler(service *Service) *Handler {
	return &Handler{
		service: service,
	}
}

// CreateTodoItem 创建待办事项
func (h *Handler) CreateTodoItem(c *gin.Context) {
	var req struct {
		Title   string `json:"title" binding:"required"`
		DueDate string `json:"due_date,omitempty"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 从中间件获取用户信息（假设已有认证中间件）
	userID := c.GetString("user_id")
	teamID := c.GetString("team_id")

	if userID == "" || teamID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户未认证"})
		return
	}

	var dueDate *time.Time
	if req.DueDate != "" {
		parsedDate, err := time.Parse("2006-01-02T15:04:05Z07:00", req.DueDate)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "日期格式错误"})
			return
		}
		dueDate = &parsedDate
	}

	todo, err := h.service.CreateTodoItem(userID, teamID, req.Title, dueDate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, todo.ToResponse())
}

// GetTodoItems 获取待办事项列表
func (h *Handler) GetTodoItems(c *gin.Context) {
	userID := c.GetString("user_id")
	teamID := c.GetString("team_id")

	if userID == "" || teamID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户未认证"})
		return
	}

	var completed *bool
	if completedStr := c.Query("completed"); completedStr != "" {
		if parsedCompleted, err := strconv.ParseBool(completedStr); err == nil {
			completed = &parsedCompleted
		}
	}

	todos, err := h.service.GetTodoItems(userID, teamID, completed)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 转换为响应格式
	responses := make([]interface{}, len(todos))
	for i, todo := range todos {
		responses[i] = todo.ToResponse()
	}

	c.JSON(http.StatusOK, gin.H{"data": responses})
}

// UpdateTodoItem 更新待办事项
func (h *Handler) UpdateTodoItem(c *gin.Context) {
	id := c.Param("id")

	var req struct {
		Completed bool `json:"completed"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID := c.GetString("user_id")
	teamID := c.GetString("team_id")

	if userID == "" || teamID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户未认证"})
		return
	}

	if err := h.service.UpdateTodoItem(id, userID, teamID, req.Completed); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "待办事项更新成功"})
}

// DeleteTodoItem 删除待办事项
func (h *Handler) DeleteTodoItem(c *gin.Context) {
	id := c.Param("id")

	userID := c.GetString("user_id")
	teamID := c.GetString("team_id")

	if userID == "" || teamID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户未认证"})
		return
	}

	if err := h.service.DeleteTodoItem(id, userID, teamID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "待办事项删除成功"})
}

// CreateCalendarEvent 创建日程事件
func (h *Handler) CreateCalendarEvent(c *gin.Context) {
	var req struct {
		Title       string `json:"title" binding:"required"`
		Description string `json:"description"`
		StartTime   string `json:"start_time" binding:"required"`
		EndTime     string `json:"end_time" binding:"required"`
		AllDay      bool   `json:"all_day"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID := c.GetString("user_id")
	teamID := c.GetString("team_id")

	if userID == "" || teamID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户未认证"})
		return
	}

	startTime, err := time.Parse("2006-01-02T15:04:05Z07:00", req.StartTime)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "开始时间格式错误"})
		return
	}

	endTime, err := time.Parse("2006-01-02T15:04:05Z07:00", req.EndTime)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "结束时间格式错误"})
		return
	}

	if endTime.Before(startTime) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "结束时间不能早于开始时间"})
		return
	}

	event, err := h.service.CreateCalendarEvent(userID, teamID, req.Title, req.Description, startTime, endTime, req.AllDay)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, event.ToResponse())
}

// GetCalendarEvents 获取日程事件列表
func (h *Handler) GetCalendarEvents(c *gin.Context) {
	userID := c.GetString("user_id")
	teamID := c.GetString("team_id")

	if userID == "" || teamID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户未认证"})
		return
	}

	var startDate, endDate *time.Time

	if startDateStr := c.Query("start_date"); startDateStr != "" {
		if parsedDate, err := time.Parse("2006-01-02", startDateStr); err == nil {
			startDate = &parsedDate
		}
	}

	if endDateStr := c.Query("end_date"); endDateStr != "" {
		if parsedDate, err := time.Parse("2006-01-02", endDateStr); err == nil {
			endDate = &parsedDate
		}
	}

	events, err := h.service.GetCalendarEvents(userID, teamID, startDate, endDate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 转换为响应格式
	responses := make([]interface{}, len(events))
	for i, event := range events {
		responses[i] = event.ToResponse()
	}

	c.JSON(http.StatusOK, gin.H{"data": responses})
}

// GetUpcomingEvents 获取未来7天的日程事件
func (h *Handler) GetUpcomingEvents(c *gin.Context) {
	userID := c.GetString("user_id")
	teamID := c.GetString("team_id")

	if userID == "" || teamID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户未认证"})
		return
	}

	events, err := h.service.GetUpcomingEvents(userID, teamID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 转换为响应格式
	responses := make([]interface{}, len(events))
	for i, event := range events {
		responses[i] = event.ToResponse()
	}

	c.JSON(http.StatusOK, gin.H{"data": responses})
}

// UpdateCalendarEvent 更新日程事件
func (h *Handler) UpdateCalendarEvent(c *gin.Context) {
	id := c.Param("id")

	var req struct {
		Title       string `json:"title"`
		Description string `json:"description"`
		StartTime   string `json:"start_time"`
		EndTime     string `json:"end_time"`
		AllDay      *bool  `json:"all_day"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID := c.GetString("user_id")
	teamID := c.GetString("team_id")

	if userID == "" || teamID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户未认证"})
		return
	}

	updates := make(map[string]interface{})

	if req.Title != "" {
		updates["title"] = req.Title
	}
	if req.Description != "" {
		updates["description"] = req.Description
	}
	if req.StartTime != "" {
		if startTime, err := time.Parse("2006-01-02T15:04:05Z07:00", req.StartTime); err == nil {
			updates["start_time"] = startTime
		} else {
			c.JSON(http.StatusBadRequest, gin.H{"error": "开始时间格式错误"})
			return
		}
	}
	if req.EndTime != "" {
		if endTime, err := time.Parse("2006-01-02T15:04:05Z07:00", req.EndTime); err == nil {
			updates["end_time"] = endTime
		} else {
			c.JSON(http.StatusBadRequest, gin.H{"error": "结束时间格式错误"})
			return
		}
	}
	if req.AllDay != nil {
		updates["all_day"] = *req.AllDay
	}

	if len(updates) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "没有要更新的字段"})
		return
	}

	if err := h.service.UpdateCalendarEvent(id, userID, teamID, updates); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "日程事件更新成功"})
}

// DeleteCalendarEvent 删除日程事件
func (h *Handler) DeleteCalendarEvent(c *gin.Context) {
	id := c.Param("id")

	userID := c.GetString("user_id")
	teamID := c.GetString("team_id")

	if userID == "" || teamID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户未认证"})
		return
	}

	if err := h.service.DeleteCalendarEvent(id, userID, teamID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "日程事件删除成功"})
}

// GetDashboardStats 获取Dashboard统计信息
func (h *Handler) GetDashboardStats(c *gin.Context) {
	userID := c.GetString("user_id")
	teamID := c.GetString("team_id")

	if userID == "" || teamID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户未认证"})
		return
	}

	stats, err := h.service.GetDashboardStats(userID, teamID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, stats)
}

// GetNotifications 获取用户通知
func (h *Handler) GetNotifications(c *gin.Context) {
	userID := c.GetString("user_id")
	teamID := c.GetString("team_id")

	if userID == "" || teamID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "缺少用户信息"})
		return
	}

	// 获取查询参数
	limit := 10 // 默认获取10条通知
	if limitStr := c.Query("limit"); limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 && parsedLimit <= 50 {
			limit = parsedLimit
		}
	}

	unreadOnly := c.Query("unread_only") == "true"

	notifications, err := h.service.GetUserNotifications(userID, teamID, limit, unreadOnly)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": notifications})
}

// MarkNotificationAsRead 标记通知为已读
func (h *Handler) MarkNotificationAsRead(c *gin.Context) {
	userID := c.GetString("user_id")
	teamID := c.GetString("team_id")
	notificationID := c.Param("id")

	if userID == "" || teamID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "缺少用户信息"})
		return
	}

	if notificationID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "缺少通知ID"})
		return
	}

	err := h.service.MarkNotificationAsRead(notificationID, userID, teamID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "标记成功"})
}
