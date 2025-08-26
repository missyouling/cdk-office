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
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"cdk-office/internal/models"
	"cdk-office/internal/db"
)

// Service 审批流程服务
type Service struct {
	db *gorm.DB
}

// NewService 创建审批流程服务实例
func NewService() *Service {
	return &Service{
		db: db.GetDB(),
	}
}

// CreateApproval 创建审批流程
func (s *Service) CreateApproval(approval *models.ApprovalProcess) error {
	// 设置默认值
	if approval.ID == "" {
		approval.ID = uuid.New().String()
	}
	approval.SubmittedAt = time.Now()
	
	// 如果没有设置状态，默认为待审批
	if approval.Status == "" {
		approval.Status = "pending"
	}
	
	// 保存到数据库
	return s.db.Create(approval).Error
}

// GetApprovalByID 根据ID获取审批流程
func (s *Service) GetApprovalByID(id string) (*models.ApprovalProcess, error) {
	var approval models.ApprovalProcess
	err := s.db.Where("id = ?", id).First(&approval).Error
	if err != nil {
		return nil, err
	}
	return &approval, nil
}

// ListApprovals 获取审批流程列表
func (s *Service) ListApprovals(teamID string, status string, page, pageSize int) ([]models.ApprovalProcess, int64, error) {
	var approvals []models.ApprovalProcess
	var total int64
	
	query := s.db.Model(&models.ApprovalProcess{}).Where("team_id = ?", teamID)
	
	// 根据状态筛选
	if status != "" {
		query = query.Where("status = ?", status)
	}
	
	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	
	// 分页查询
	offset := (page - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).Order("created_at DESC").Find(&approvals).Error; err != nil {
		return nil, 0, err
	}
	
	return approvals, total, nil
}

// UpdateApprovalStatus 更新审批状态
func (s *Service) UpdateApprovalStatus(id, status, comments, actorID, actorName string) error {
	// 获取当前审批流程
	approval, err := s.GetApprovalByID(id)
	if err != nil {
		return err
	}
	
	// 更新状态和时间
	approval.Status = status
	approval.UpdatedBy = actorID
	
	// 根据状态更新对应的时间字段
	now := time.Now()
	switch status {
	case "approved":
		approval.ApprovedAt = &now
	case "rejected":
		approval.RejectedAt = &now
	case "cancelled":
		approval.CancelledAt = &now
	}
	
	// 更新审批意见
	if comments != "" {
		approval.Comments = comments
	}
	
	// 保存更新
	if err := s.db.Save(approval).Error; err != nil {
		return err
	}
	
	// 记录审批历史
	history := &models.ApprovalHistory{
		ApprovalID: id,
		ActorID:    actorID,
		ActorName:  actorName,
		Action:     status,
		Comments:   comments,
		ActionTime: now,
	}
	
	return s.db.Create(history).Error
}

// GetApprovalHistory 获取审批历史
func (s *Service) GetApprovalHistory(approvalID string) ([]models.ApprovalHistory, error) {
	var history []models.ApprovalHistory
	err := s.db.Where("approval_id = ?", approvalID).Order("action_time ASC").Find(&history).Error
	return history, err
}

// CreateApprovalTemplate 创建审批模板
func (s *Service) CreateApprovalTemplate(template *models.ApprovalTemplate) error {
	// 设置默认值
	if template.ID == "" {
		template.ID = uuid.New().String()
	}
	
	// 保存到数据库
	return s.db.Create(template).Error
}

// ListApprovalTemplates 获取审批模板列表
func (s *Service) ListApprovalTemplates(teamID string) ([]models.ApprovalTemplate, error) {
	var templates []models.ApprovalTemplate
	err := s.db.Where("team_id = ?", teamID).Order("created_at DESC").Find(&templates).Error
	return templates, err
}

// CreateNotification 创建审批通知
func (s *Service) CreateNotification(notification *models.ApprovalNotification) error {
	// 设置默认值
	if notification.ID == "" {
		notification.ID = uuid.New().String()
	}
	
	// 保存到数据库
	return s.db.Create(notification).Error
}

// ListNotifications 获取用户通知列表
func (s *Service) ListNotifications(userID string, isRead *bool) ([]models.ApprovalNotification, error) {
	var notifications []models.ApprovalNotification
	query := s.db.Where("user_id = ?", userID)
	
	if isRead != nil {
		query = query.Where("is_read = ?", *isRead)
	}
	
	err := query.Order("created_at DESC").Find(&notifications).Error
	return notifications, err
}

// MarkNotificationAsRead 标记通知为已读
func (s *Service) MarkNotificationAsRead(id string) error {
	return s.db.Model(&models.ApprovalNotification{}).Where("id = ?", id).Update("is_read", true).Update("read_at", time.Now()).Error
}