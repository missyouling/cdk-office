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

package approval

import (
	"net/http"
	"strconv"

	"cdk-office/internal/models"

	"github.com/gin-gonic/gin"
)

// Handler 审批流程HTTP处理函数
type Handler struct {
	service *Service
}

// NewHandler 创建审批流程HTTP处理函数实例
func NewHandler(service *Service) *Handler {
	return &Handler{
		service: service,
	}
}

// CreateApproval 创建审批流程
func (h *Handler) CreateApproval(c *gin.Context) {
	var req struct {
		TeamID        string `json:"team_id" binding:"required"`
		Name          string `json:"name" binding:"required"`
		Description   string `json:"description"`
		DocumentID    string `json:"document_id"`
		DocumentName  string `json:"document_name"`
		RequestorID   string `json:"requestor_id" binding:"required"`
		RequestorName string `json:"requestor_name" binding:"required"`
		ApproverID    string `json:"approver_id"`
		ApproverName  string `json:"approver_name"`
		ApprovalType  string `json:"approval_type" binding:"required"`
		Comments      string `json:"comments"`
		Deadline      string `json:"deadline"`
		Priority      string `json:"priority"`
		CreatedBy     string `json:"created_by" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	approval := &models.ApprovalProcess{
		TeamID:        req.TeamID,
		Name:          req.Name,
		Description:   req.Description,
		DocumentID:    req.DocumentID,
		DocumentName:  req.DocumentName,
		RequestorID:   req.RequestorID,
		RequestorName: req.RequestorName,
		ApproverID:    req.ApproverID,
		ApproverName:  req.ApproverName,
		ApprovalType:  req.ApprovalType,
		Comments:      req.Comments,
		Priority:      req.Priority,
		CreatedBy:     req.CreatedBy,
		UpdatedBy:     req.CreatedBy,
	}

	if err := h.service.CreateApproval(approval); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, approval)
}

// GetApprovalByID 根据ID获取审批流程
func (h *Handler) GetApprovalByID(c *gin.Context) {
	id := c.Param("id")

	approval, err := h.service.GetApprovalByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "审批流程不存在"})
		return
	}

	c.JSON(http.StatusOK, approval)
}

// ListApprovals 获取审批流程列表
func (h *Handler) ListApprovals(c *gin.Context) {
	teamID := c.Query("team_id")
	if teamID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "team_id参数不能为空"})
		return
	}

	status := c.Query("status")

	page, _ := strconv.Atoi(c.Query("page"))
	if page <= 0 {
		page = 1
	}

	pageSize, _ := strconv.Atoi(c.Query("page_size"))
	if pageSize <= 0 {
		pageSize = 10
	}

	approvals, total, err := h.service.ListApprovals(teamID, status, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":      approvals,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// UpdateApprovalStatus 更新审批状态
func (h *Handler) UpdateApprovalStatus(c *gin.Context) {
	id := c.Param("id")

	var req struct {
		Status    string `json:"status" binding:"required,oneof=pending approved rejected cancelled"`
		Comments  string `json:"comments"`
		ActorID   string `json:"actor_id" binding:"required"`
		ActorName string `json:"actor_name" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.service.UpdateApprovalStatus(id, req.Status, req.Comments, req.ActorID, req.ActorName); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "审批状态更新成功"})
}

// GetApprovalHistory 获取审批历史
func (h *Handler) GetApprovalHistory(c *gin.Context) {
	approvalID := c.Param("id")

	history, err := h.service.GetApprovalHistory(approvalID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, history)
}

// CreateApprovalTemplate 创建审批模板
func (h *Handler) CreateApprovalTemplate(c *gin.Context) {
	var req struct {
		TeamID        string   `json:"team_id" binding:"required"`
		Name          string   `json:"name" binding:"required"`
		Description   string   `json:"description"`
		ApprovalType  string   `json:"approval_type" binding:"required"`
		ApproverRoles []string `json:"approver_roles"`
		Steps         []string `json:"steps"`
		AutoApprove   bool     `json:"auto_approve"`
		Conditions    string   `json:"conditions"`
		CreatedBy     string   `json:"created_by" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	template := &models.ApprovalTemplate{
		TeamID:        req.TeamID,
		Name:          req.Name,
		Description:   req.Description,
		ApprovalType:  req.ApprovalType,
		ApproverRoles: req.ApproverRoles,
		Steps:         req.Steps,
		AutoApprove:   req.AutoApprove,
		Conditions:    req.Conditions,
		CreatedBy:     req.CreatedBy,
		UpdatedBy:     req.CreatedBy,
	}

	if err := h.service.CreateApprovalTemplate(template); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, template)
}

// ListApprovalTemplates 获取审批模板列表
func (h *Handler) ListApprovalTemplates(c *gin.Context) {
	teamID := c.Query("team_id")
	if teamID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "team_id参数不能为空"})
		return
	}

	templates, err := h.service.ListApprovalTemplates(teamID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, templates)
}

// CreateNotification 创建审批通知
func (h *Handler) CreateNotification(c *gin.Context) {
	var req struct {
		ApprovalID       string `json:"approval_id" binding:"required"`
		UserID           string `json:"user_id" binding:"required"`
		NotificationType string `json:"notification_type" binding:"required"`
		Title            string `json:"title" binding:"required"`
		Content          string `json:"content" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	notification := &models.ApprovalNotification{
		ApprovalID:       req.ApprovalID,
		UserID:           req.UserID,
		NotificationType: req.NotificationType,
		Title:            req.Title,
		Content:          req.Content,
	}

	if err := h.service.CreateNotification(notification); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, notification)
}

// ListNotifications 获取用户通知列表
func (h *Handler) ListNotifications(c *gin.Context) {
	userID := c.Query("user_id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user_id参数不能为空"})
		return
	}

	var isRead *bool
	if read := c.Query("is_read"); read != "" {
		if read == "true" {
			b := true
			isRead = &b
		} else if read == "false" {
			b := false
			isRead = &b
		}
	}

	notifications, err := h.service.ListNotifications(userID, isRead)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, notifications)
}

// MarkNotificationAsRead 标记通知为已读
func (h *Handler) MarkNotificationAsRead(c *gin.Context) {
	id := c.Param("id")

	if err := h.service.MarkNotificationAsRead(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "通知标记为已读"})
}
