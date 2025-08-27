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

package notification

import (
	"net/http"
	"strconv"

	"cdk-office/internal/models"

	"github.com/gin-gonic/gin"
)

// Handler 通知中心HTTP处理函数
type Handler struct {
	service *Service
}

// NewHandler 创建通知中心HTTP处理函数实例
func NewHandler(service *Service) *Handler {
	return &Handler{
		service: service,
	}
}

// CreateNotification 创建通知
func (h *Handler) CreateNotification(c *gin.Context) {
	var req struct {
		TeamID         string `json:"team_id" binding:"required"`
		UserID         string `json:"user_id" binding:"required"`
		Title          string `json:"title" binding:"required"`
		Content        string `json:"content" binding:"required"`
		Type           string `json:"type" binding:"required"`
		Category       string `json:"category"`
		Priority       string `json:"priority"`
		RelatedID      string `json:"related_id"`
		RelatedType    string `json:"related_type"`
		ActionRequired bool   `json:"action_required"`
		CreatedBy      string `json:"created_by"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	notification := &models.Notification{
		TeamID:         req.TeamID,
		UserID:         req.UserID,
		Title:          req.Title,
		Content:        req.Content,
		Type:           req.Type,
		Category:       req.Category,
		Priority:       req.Priority,
		RelatedID:      req.RelatedID,
		RelatedType:    req.RelatedType,
		ActionRequired: req.ActionRequired,
		CreatedBy:      req.CreatedBy,
	}

	if err := h.service.CreateNotification(notification); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, notification)
}

// GetNotificationByID 根据ID获取通知
func (h *Handler) GetNotificationByID(c *gin.Context) {
	id := c.Param("id")

	notification, err := h.service.GetNotificationByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "通知不存在"})
		return
	}

	c.JSON(http.StatusOK, notification)
}

// ListNotifications 获取通知列表
func (h *Handler) ListNotifications(c *gin.Context) {
	userID := c.Query("user_id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user_id参数不能为空"})
		return
	}

	// 构建筛选条件
	filters := make(map[string]interface{})

	if isRead := c.Query("is_read"); isRead != "" {
		if isRead == "true" {
			filters["is_read"] = true
		} else if isRead == "false" {
			filters["is_read"] = false
		}
	}

	if isArchived := c.Query("is_archived"); isArchived != "" {
		if isArchived == "true" {
			filters["is_archived"] = true
		} else if isArchived == "false" {
			filters["is_archived"] = false
		}
	}

	if notificationType := c.Query("type"); notificationType != "" {
		filters["type"] = notificationType
	}

	if priority := c.Query("priority"); priority != "" {
		filters["priority"] = priority
	}

	page, _ := strconv.Atoi(c.Query("page"))
	if page <= 0 {
		page = 1
	}

	pageSize, _ := strconv.Atoi(c.Query("page_size"))
	if pageSize <= 0 {
		pageSize = 10
	}

	notifications, total, err := h.service.ListNotifications(userID, filters, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":      notifications,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// MarkAsRead 标记通知为已读
func (h *Handler) MarkAsRead(c *gin.Context) {
	id := c.Param("id")

	if err := h.service.MarkAsRead(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "通知标记为已读"})
}

// MarkMultipleAsRead 批量标记通知为已读
func (h *Handler) MarkMultipleAsRead(c *gin.Context) {
	var req struct {
		IDs []string `json:"ids" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.service.MarkMultipleAsRead(req.IDs); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "通知批量标记为已读"})
}

// MarkAllAsRead 标记所有通知为已读
func (h *Handler) MarkAllAsRead(c *gin.Context) {
	userID := c.Query("user_id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user_id参数不能为空"})
		return
	}

	if err := h.service.MarkAllAsRead(userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "所有通知标记为已读"})
}

// ArchiveNotification 归档通知
func (h *Handler) ArchiveNotification(c *gin.Context) {
	id := c.Param("id")

	if err := h.service.ArchiveNotification(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "通知已归档"})
}

// DeleteNotification 删除通知
func (h *Handler) DeleteNotification(c *gin.Context) {
	id := c.Param("id")

	if err := h.service.DeleteNotification(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "通知已删除"})
}

// CreateTemplate 创建通知模板
func (h *Handler) CreateTemplate(c *gin.Context) {
	var req struct {
		TeamID      string `json:"team_id" binding:"required"`
		Name        string `json:"name" binding:"required"`
		Description string `json:"description"`
		Type        string `json:"type" binding:"required"`
		Subject     string `json:"subject"`
		Content     string `json:"content" binding:"required"`
		IsDefault   bool   `json:"is_default"`
		CreatedBy   string `json:"created_by" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	template := &models.NotificationTemplate{
		TeamID:      req.TeamID,
		Name:        req.Name,
		Description: req.Description,
		Type:        req.Type,
		Subject:     req.Subject,
		Content:     req.Content,
		IsDefault:   req.IsDefault,
		CreatedBy:   req.CreatedBy,
		UpdatedBy:   req.CreatedBy,
	}

	if err := h.service.CreateTemplate(template); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, template)
}

// ListTemplates 获取通知模板列表
func (h *Handler) ListTemplates(c *gin.Context) {
	teamID := c.Query("team_id")
	if teamID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "team_id参数不能为空"})
		return
	}

	templates, err := h.service.ListTemplates(teamID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, templates)
}

// GetTemplateByID 根据ID获取通知模板
func (h *Handler) GetTemplateByID(c *gin.Context) {
	id := c.Param("id")

	template, err := h.service.GetTemplateByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "通知模板不存在"})
		return
	}

	c.JSON(http.StatusOK, template)
}

// UpdateTemplate 更新通知模板
func (h *Handler) UpdateTemplate(c *gin.Context) {
	id := c.Param("id")

	var req struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		Type        string `json:"type"`
		Subject     string `json:"subject"`
		Content     string `json:"content"`
		IsDefault   bool   `json:"is_default"`
		UpdatedBy   string `json:"updated_by" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	template, err := h.service.GetTemplateByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "通知模板不存在"})
		return
	}

	// 更新字段
	if req.Name != "" {
		template.Name = req.Name
	}
	if req.Description != "" {
		template.Description = req.Description
	}
	if req.Type != "" {
		template.Type = req.Type
	}
	if req.Subject != "" {
		template.Subject = req.Subject
	}
	if req.Content != "" {
		template.Content = req.Content
	}
	template.IsDefault = req.IsDefault
	template.UpdatedBy = req.UpdatedBy

	if err := h.service.UpdateTemplate(template); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, template)
}

// DeleteTemplate 删除通知模板
func (h *Handler) DeleteTemplate(c *gin.Context) {
	id := c.Param("id")

	if err := h.service.DeleteTemplate(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "通知模板已删除"})
}

// GetUserPreference 获取用户通知偏好设置
func (h *Handler) GetUserPreference(c *gin.Context) {
	userID := c.Query("user_id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user_id参数不能为空"})
		return
	}

	preference, err := h.service.GetUserPreference(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, preference)
}

// UpdateUserPreference 更新用户通知偏好设置
func (h *Handler) UpdateUserPreference(c *gin.Context) {
	var req struct {
		UserID         string `json:"user_id" binding:"required"`
		EmailEnabled   *bool  `json:"email_enabled"`
		EmailFrequency string `json:"email_frequency"`
		PushEnabled    *bool  `json:"push_enabled"`
		InAppEnabled   *bool  `json:"in_app_enabled"`
		SmsEnabled     *bool  `json:"sms_enabled"`
		DesktopEnabled *bool  `json:"desktop_enabled"`
		SoundEnabled   *bool  `json:"sound_enabled"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	preference, err := h.service.GetUserPreference(req.UserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 更新字段
	if req.EmailEnabled != nil {
		preference.EmailEnabled = *req.EmailEnabled
	}
	if req.EmailFrequency != "" {
		preference.EmailFrequency = req.EmailFrequency
	}
	if req.PushEnabled != nil {
		preference.PushEnabled = *req.PushEnabled
	}
	if req.InAppEnabled != nil {
		preference.InAppEnabled = *req.InAppEnabled
	}
	if req.SmsEnabled != nil {
		preference.SmsEnabled = *req.SmsEnabled
	}
	if req.DesktopEnabled != nil {
		preference.DesktopEnabled = *req.DesktopEnabled
	}
	if req.SoundEnabled != nil {
		preference.SoundEnabled = *req.SoundEnabled
	}

	if err := h.service.UpdateUserPreference(preference); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, preference)
}

// CreateChannel 创建通知渠道
func (h *Handler) CreateChannel(c *gin.Context) {
	var req struct {
		TeamID    string `json:"team_id" binding:"required"`
		Name      string `json:"name" binding:"required"`
		Type      string `json:"type" binding:"required"`
		Config    string `json:"config"`
		IsActive  bool   `json:"is_active"`
		CreatedBy string `json:"created_by" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	channel := &models.NotificationChannel{
		TeamID:    req.TeamID,
		Name:      req.Name,
		Type:      req.Type,
		Config:    req.Config,
		IsActive:  req.IsActive,
		CreatedBy: req.CreatedBy,
		UpdatedBy: req.CreatedBy,
	}

	if err := h.service.CreateChannel(channel); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, channel)
}

// ListChannels 获取通知渠道列表
func (h *Handler) ListChannels(c *gin.Context) {
	teamID := c.Query("team_id")
	if teamID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "team_id参数不能为空"})
		return
	}

	channels, err := h.service.ListChannels(teamID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, channels)
}

// GetChannelByID 根据ID获取通知渠道
func (h *Handler) GetChannelByID(c *gin.Context) {
	id := c.Param("id")

	channel, err := h.service.GetChannelByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "通知渠道不存在"})
		return
	}

	c.JSON(http.StatusOK, channel)
}

// UpdateChannel 更新通知渠道
func (h *Handler) UpdateChannel(c *gin.Context) {
	id := c.Param("id")

	var req struct {
		Name      string `json:"name"`
		Type      string `json:"type"`
		Config    string `json:"config"`
		IsActive  *bool  `json:"is_active"`
		UpdatedBy string `json:"updated_by" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	channel, err := h.service.GetChannelByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "通知渠道不存在"})
		return
	}

	// 更新字段
	if req.Name != "" {
		channel.Name = req.Name
	}
	if req.Type != "" {
		channel.Type = req.Type
	}
	if req.Config != "" {
		channel.Config = req.Config
	}
	if req.IsActive != nil {
		channel.IsActive = *req.IsActive
	}
	channel.UpdatedBy = req.UpdatedBy

	if err := h.service.UpdateChannel(channel); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, channel)
}

// DeleteChannel 删除通知渠道
func (h *Handler) DeleteChannel(c *gin.Context) {
	id := c.Param("id")

	if err := h.service.DeleteChannel(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "通知渠道已删除"})
}

// GetUnreadCount 获取未读通知数量
func (h *Handler) GetUnreadCount(c *gin.Context) {
	userID := c.Query("user_id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user_id参数不能为空"})
		return
	}

	count, err := h.service.GetUnreadCount(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"unread_count": count})
}
