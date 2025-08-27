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
	"fmt"
	"time"

	"cdk-office/internal/db"
	"cdk-office/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Service 通知中心服务
type Service struct {
	db *gorm.DB
}

// NewService 创建通知中心服务实例
func NewService() *Service {
	return &Service{
		db: db.GetDB(),
	}
}

// CreateNotification 创建通知
func (s *Service) CreateNotification(notification *models.Notification) error {
	// 设置默认值
	if notification.ID == "" {
		notification.ID = uuid.New().String()
	}

	// 保存到数据库
	return s.db.Create(notification).Error
}

// GetNotificationByID 根据ID获取通知
func (s *Service) GetNotificationByID(id string) (*models.Notification, error) {
	var notification models.Notification
	err := s.db.Where("id = ?", id).First(&notification).Error
	if err != nil {
		return nil, err
	}
	return &notification, nil
}

// ListNotifications 获取通知列表
func (s *Service) ListNotifications(userID string, filters map[string]interface{}, page, pageSize int) ([]models.Notification, int64, error) {
	var notifications []models.Notification
	var total int64

	query := s.db.Model(&models.Notification{}).Where("user_id = ?", userID)

	// 应用筛选条件
	if isRead, ok := filters["is_read"]; ok {
		query = query.Where("is_read = ?", isRead)
	}

	if isArchived, ok := filters["is_archived"]; ok {
		query = query.Where("is_archived = ?", isArchived)
	}

	if notificationType, ok := filters["type"]; ok {
		query = query.Where("type = ?", notificationType)
	}

	if priority, ok := filters["priority"]; ok {
		query = query.Where("priority = ?", priority)
	}

	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (page - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).Order("created_at DESC").Find(&notifications).Error; err != nil {
		return nil, 0, err
	}

	return notifications, total, nil
}

// MarkAsRead 标记通知为已读
func (s *Service) MarkAsRead(id string) error {
	now := time.Now()
	return s.db.Model(&models.Notification{}).Where("id = ?", id).Updates(map[string]interface{}{
		"is_read": true,
		"read_at": &now,
	}).Error
}

// MarkMultipleAsRead 批量标记通知为已读
func (s *Service) MarkMultipleAsRead(ids []string) error {
	now := time.Now()
	return s.db.Model(&models.Notification{}).Where("id IN ?", ids).Updates(map[string]interface{}{
		"is_read": true,
		"read_at": &now,
	}).Error
}

// MarkAllAsRead 标记所有通知为已读
func (s *Service) MarkAllAsRead(userID string) error {
	now := time.Now()
	return s.db.Model(&models.Notification{}).Where("user_id = ? AND is_read = ?", userID, false).Updates(map[string]interface{}{
		"is_read": true,
		"read_at": &now,
	}).Error
}

// ArchiveNotification 归档通知
func (s *Service) ArchiveNotification(id string) error {
	now := time.Now()
	return s.db.Model(&models.Notification{}).Where("id = ?", id).Updates(map[string]interface{}{
		"is_archived": true,
		"archived_at": &now,
	}).Error
}

// DeleteNotification 删除通知
func (s *Service) DeleteNotification(id string) error {
	return s.db.Delete(&models.Notification{}, "id = ?", id).Error
}

// CreateTemplate 创建通知模板
func (s *Service) CreateTemplate(template *models.NotificationTemplate) error {
	// 设置默认值
	if template.ID == "" {
		template.ID = uuid.New().String()
	}

	// 如果是默认模板，取消其他默认模板
	if template.IsDefault {
		s.db.Model(&models.NotificationTemplate{}).Where("team_id = ? AND is_default = ?", template.TeamID, true).Update("is_default", false)
	}

	// 保存到数据库
	return s.db.Create(template).Error
}

// ListTemplates 获取通知模板列表
func (s *Service) ListTemplates(teamID string) ([]models.NotificationTemplate, error) {
	var templates []models.NotificationTemplate
	err := s.db.Where("team_id = ?", teamID).Order("created_at DESC").Find(&templates).Error
	return templates, err
}

// GetTemplateByID 根据ID获取通知模板
func (s *Service) GetTemplateByID(id string) (*models.NotificationTemplate, error) {
	var template models.NotificationTemplate
	err := s.db.Where("id = ?", id).First(&template).Error
	if err != nil {
		return nil, err
	}
	return &template, nil
}

// UpdateTemplate 更新通知模板
func (s *Service) UpdateTemplate(template *models.NotificationTemplate) error {
	// 如果是默认模板，取消其他默认模板
	if template.IsDefault {
		s.db.Model(&models.NotificationTemplate{}).Where("team_id = ? AND id != ? AND is_default = ?", template.TeamID, template.ID, true).Update("is_default", false)
	}

	// 保存到数据库
	return s.db.Save(template).Error
}

// DeleteTemplate 删除通知模板
func (s *Service) DeleteTemplate(id string) error {
	return s.db.Delete(&models.NotificationTemplate{}, "id = ?", id).Error
}

// GetUserPreference 获取用户通知偏好设置
func (s *Service) GetUserPreference(userID string) (*models.NotificationPreference, error) {
	var preference models.NotificationPreference
	err := s.db.Where("user_id = ?", userID).First(&preference).Error
	if err != nil {
		// 如果没有找到，创建默认偏好设置
		if err == gorm.ErrRecordNotFound {
			preference = models.NotificationPreference{
				UserID:         userID,
				EmailEnabled:   true,
				EmailFrequency: "immediately",
				PushEnabled:    true,
				InAppEnabled:   true,
				SmsEnabled:     false,
				DesktopEnabled: true,
				SoundEnabled:   true,
			}
			return &preference, s.db.Create(&preference).Error
		}
		return nil, err
	}
	return &preference, nil
}

// UpdateUserPreference 更新用户通知偏好设置
func (s *Service) UpdateUserPreference(preference *models.NotificationPreference) error {
	// 设置默认值
	if preference.ID == "" {
		preference.ID = uuid.New().String()
	}

	// 保存到数据库
	return s.db.Save(preference).Error
}

// CreateChannel 创建通知渠道
func (s *Service) CreateChannel(channel *models.NotificationChannel) error {
	// 设置默认值
	if channel.ID == "" {
		channel.ID = uuid.New().String()
	}

	// 保存到数据库
	return s.db.Create(channel).Error
}

// ListChannels 获取通知渠道列表
func (s *Service) ListChannels(teamID string) ([]models.NotificationChannel, error) {
	var channels []models.NotificationChannel
	err := s.db.Where("team_id = ?", teamID).Order("created_at DESC").Find(&channels).Error
	return channels, err
}

// GetChannelByID 根据ID获取通知渠道
func (s *Service) GetChannelByID(id string) (*models.NotificationChannel, error) {
	var channel models.NotificationChannel
	err := s.db.Where("id = ?", id).First(&channel).Error
	if err != nil {
		return nil, err
	}
	return &channel, nil
}

// UpdateChannel 更新通知渠道
func (s *Service) UpdateChannel(channel *models.NotificationChannel) error {
	// 保存到数据库
	return s.db.Save(channel).Error
}

// DeleteChannel 删除通知渠道
func (s *Service) DeleteChannel(id string) error {
	return s.db.Delete(&models.NotificationChannel{}, "id = ?", id).Error
}

// SendNotification 发送通知
func (s *Service) SendNotification(notification *models.Notification) error {
	// 创建通知
	if err := s.CreateNotification(notification); err != nil {
		return fmt.Errorf("创建通知失败: %v", err)
	}

	// 获取用户偏好设置
	preference, err := s.GetUserPreference(notification.UserID)
	if err != nil {
		return fmt.Errorf("获取用户偏好设置失败: %v", err)
	}

	// 根据用户偏好设置发送通知
	if preference.InAppEnabled {
		// TODO: 实现应用内通知发送逻辑
		// 这里可以集成WebSocket或其他实时通信机制
	}

	if preference.EmailEnabled && preference.EmailFrequency == "immediately" {
		// TODO: 实现邮件通知发送逻辑
	}

	if preference.PushEnabled {
		// TODO: 实现推送通知发送逻辑
	}

	if preference.SmsEnabled {
		// TODO: 实现短信通知发送逻辑
	}

	if preference.DesktopEnabled {
		// TODO: 实现桌面通知发送逻辑
	}

	return nil
}

// GetUnreadCount 获取未读通知数量
func (s *Service) GetUnreadCount(userID string) (int64, error) {
	var count int64
	err := s.db.Model(&models.Notification{}).Where("user_id = ? AND is_read = ?", userID, false).Count(&count).Error
	return count, err
}
