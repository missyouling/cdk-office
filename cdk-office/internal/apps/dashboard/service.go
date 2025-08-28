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
	"fmt"
	"time"

	"cdk-office/internal/models"

	"gorm.io/gorm"
)

// Service Dashboard服务
type Service struct {
	db *gorm.DB
}

// NewService 创建Dashboard服务实例
func NewService(db *gorm.DB) *Service {
	return &Service{
		db: db,
	}
}

// CreateTodoItem 创建待办事项
func (s *Service) CreateTodoItem(userID, teamID, title string, dueDate *time.Time) (*models.TodoItem, error) {
	todo := &models.TodoItem{
		UserID:  userID,
		TeamID:  teamID,
		Title:   title,
		DueDate: dueDate,
	}

	if err := s.db.Create(todo).Error; err != nil {
		return nil, fmt.Errorf("创建待办事项失败: %w", err)
	}

	return todo, nil
}

// GetTodoItems 获取待办事项列表
func (s *Service) GetTodoItems(userID, teamID string, completed *bool) ([]models.TodoItem, error) {
	var todos []models.TodoItem

	query := s.db.Where("user_id = ? AND team_id = ?", userID, teamID)

	if completed != nil {
		query = query.Where("completed = ?", *completed)
	}

	if err := query.Order("created_at DESC").Find(&todos).Error; err != nil {
		return nil, fmt.Errorf("获取待办事项列表失败: %w", err)
	}

	return todos, nil
}

// UpdateTodoItem 更新待办事项
func (s *Service) UpdateTodoItem(id, userID, teamID string, completed bool) error {
	result := s.db.Model(&models.TodoItem{}).
		Where("id = ? AND user_id = ? AND team_id = ?", id, userID, teamID).
		Update("completed", completed)

	if result.Error != nil {
		return fmt.Errorf("更新待办事项失败: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("待办事项不存在或无权限")
	}

	return nil
}

// DeleteTodoItem 删除待办事项
func (s *Service) DeleteTodoItem(id, userID, teamID string) error {
	result := s.db.Where("id = ? AND user_id = ? AND team_id = ?", id, userID, teamID).
		Delete(&models.TodoItem{})

	if result.Error != nil {
		return fmt.Errorf("删除待办事项失败: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("待办事项不存在或无权限")
	}

	return nil
}

// CreateCalendarEvent 创建日程事件
func (s *Service) CreateCalendarEvent(userID, teamID, title, description string, startTime, endTime time.Time, allDay bool) (*models.CalendarEvent, error) {
	event := &models.CalendarEvent{
		UserID:      userID,
		TeamID:      teamID,
		Title:       title,
		Description: description,
		StartTime:   startTime,
		EndTime:     endTime,
		AllDay:      allDay,
	}

	if err := s.db.Create(event).Error; err != nil {
		return nil, fmt.Errorf("创建日程事件失败: %w", err)
	}

	return event, nil
}

// GetCalendarEvents 获取日程事件列表
func (s *Service) GetCalendarEvents(userID, teamID string, startDate, endDate *time.Time) ([]models.CalendarEvent, error) {
	var events []models.CalendarEvent

	query := s.db.Where("user_id = ? AND team_id = ?", userID, teamID)

	if startDate != nil && endDate != nil {
		query = query.Where("start_time >= ? AND start_time <= ?", *startDate, *endDate)
	}

	if err := query.Order("start_time ASC").Find(&events).Error; err != nil {
		return nil, fmt.Errorf("获取日程事件列表失败: %w", err)
	}

	return events, nil
}

// GetUpcomingEvents 获取未来7天的日程事件
func (s *Service) GetUpcomingEvents(userID, teamID string) ([]models.CalendarEvent, error) {
	now := time.Now()
	sevenDaysLater := now.AddDate(0, 0, 7)

	return s.GetCalendarEvents(userID, teamID, &now, &sevenDaysLater)
}

// UpdateCalendarEvent 更新日程事件
func (s *Service) UpdateCalendarEvent(id, userID, teamID string, updates map[string]interface{}) error {
	result := s.db.Model(&models.CalendarEvent{}).
		Where("id = ? AND user_id = ? AND team_id = ?", id, userID, teamID).
		Updates(updates)

	if result.Error != nil {
		return fmt.Errorf("更新日程事件失败: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("日程事件不存在或无权限")
	}

	return nil
}

// DeleteCalendarEvent 删除日程事件
func (s *Service) DeleteCalendarEvent(id, userID, teamID string) error {
	result := s.db.Where("id = ? AND user_id = ? AND team_id = ?", id, userID, teamID).
		Delete(&models.CalendarEvent{})

	if result.Error != nil {
		return fmt.Errorf("删除日程事件失败: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("日程事件不存在或无权限")
	}

	return nil
}

// GetDashboardStats 获取Dashboard统计信息
func (s *Service) GetDashboardStats(userID, teamID string) (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// 统计待办事项
	var totalTodos, completedTodos, overdueTodos int64

	s.db.Model(&models.TodoItem{}).
		Where("user_id = ? AND team_id = ?", userID, teamID).
		Count(&totalTodos)

	s.db.Model(&models.TodoItem{}).
		Where("user_id = ? AND team_id = ? AND completed = ?", userID, teamID, true).
		Count(&completedTodos)

	s.db.Model(&models.TodoItem{}).
		Where("user_id = ? AND team_id = ? AND completed = ? AND due_date < ?", userID, teamID, false, time.Now()).
		Count(&overdueTodos)

	// 统计今日日程
	today := time.Now().Truncate(24 * time.Hour)
	tomorrow := today.AddDate(0, 0, 1)

	var todayEvents int64
	s.db.Model(&models.CalendarEvent{}).
		Where("user_id = ? AND team_id = ? AND start_time >= ? AND start_time < ?", userID, teamID, today, tomorrow).
		Count(&todayEvents)

	stats["total_todos"] = totalTodos
	stats["completed_todos"] = completedTodos
	stats["overdue_todos"] = overdueTodos
	stats["pending_todos"] = totalTodos - completedTodos
	stats["today_events"] = todayEvents

	return stats, nil
}

// GetUserNotifications 获取用户通知列表
func (s *Service) GetUserNotifications(userID, teamID string, limit int, unreadOnly bool) ([]models.Notification, error) {
	var notifications []models.Notification

	query := s.db.Where("user_id = ? AND team_id = ?", userID, teamID)

	if unreadOnly {
		query = query.Where("is_read = ?", false)
	}

	if err := query.Order("created_at DESC").Limit(limit).Find(&notifications).Error; err != nil {
		return nil, fmt.Errorf("获取通知列表失败: %w", err)
	}

	return notifications, nil
}

// MarkNotificationAsRead 标记通知为已读
func (s *Service) MarkNotificationAsRead(notificationID, userID, teamID string) error {
	now := time.Now()
	result := s.db.Model(&models.Notification{}).
		Where("id = ? AND user_id = ? AND team_id = ?", notificationID, userID, teamID).
		Updates(map[string]interface{}{
			"is_read": true,
			"read_at": now,
		})

	if result.Error != nil {
		return fmt.Errorf("标记通知已读失败: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("通知不存在或无权限")
	}

	return nil
}
