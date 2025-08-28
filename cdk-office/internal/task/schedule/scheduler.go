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

package schedule

import (
	"fmt"
	"log"
	"time"

	"github.com/linux-do/cdk-office/internal/config"
	"github.com/linux-do/cdk-office/internal/db"
	"github.com/linux-do/cdk-office/internal/logger"
	"github.com/linux-do/cdk-office/internal/models"
	"github.com/robfig/cron/v3"
)

// Scheduler 任务调度器
type Scheduler struct {
	cron *cron.Cron
}

// NewScheduler 创建新的调度器
func NewScheduler() *Scheduler {
	return &Scheduler{
		cron: cron.New(cron.WithSeconds()),
	}
}

// Start 启动调度器
func (s *Scheduler) Start() {
	// 初始化数据库连接
	db.Init()
	db.InitRedis()

	// 添加定时任务
	s.addJobs()

	// 启动调度器
	s.cron.Start()

	log.Println("[SCHEDULER] scheduler started")
}

// Stop 停止调度器
func (s *Scheduler) Stop() {
	s.cron.Stop()
	log.Println("[SCHEDULER] scheduler stopped")
}

// addJobs 添加定时任务
func (s *Scheduler) addJobs() {
	// 文档同步任务
	if config.Config.Schedule.DocumentSyncCron != "" {
		_, err := s.cron.AddFunc(config.Config.Schedule.DocumentSyncCron, s.syncDocuments)
		if err != nil {
			log.Printf("[SCHEDULER] failed to add document sync job: %v", err)
		} else {
			log.Println("[SCHEDULER] document sync job added")
		}
	}

	// 健康检查任务
	if config.Config.Schedule.HealthCheckCron != "" {
		_, err := s.cron.AddFunc(config.Config.Schedule.HealthCheckCron, s.healthCheck)
		if err != nil {
			log.Printf("[SCHEDULER] failed to add health check job: %v", err)
		} else {
			log.Println("[SCHEDULER] health check job added")
		}
	}

	// 归档任务
	if config.Config.Schedule.ArchiveCron != "" {
		_, err := s.cron.AddFunc(config.Config.Schedule.ArchiveCron, s.archiveDocuments)
		if err != nil {
			log.Printf("[SCHEDULER] failed to add archive job: %v", err)
		} else {
			log.Println("[SCHEDULER] archive job added")
		}
	}

	// 日程提醒任务
	if config.Config.Schedule.CalendarReminderCron != "" {
		_, err := s.cron.AddFunc(config.Config.Schedule.CalendarReminderCron, s.checkCalendarReminders)
		if err != nil {
			log.Printf("[SCHEDULER] failed to add calendar reminder job: %v", err)
		} else {
			log.Println("[SCHEDULER] calendar reminder job added")
		}
	}
}

// syncDocuments 文档同步任务
func (s *Scheduler) syncDocuments() {
	start := time.Now()
	logger.Info("starting document sync task")

	// TODO: 实现文档同步逻辑

	logger.Info("document sync task completed")
	logger.LogWithDuration(start, "document sync")
}

// healthCheck 健康检查任务
func (s *Scheduler) healthCheck() {
	start := time.Now()
	logger.Info("starting health check task")

	// TODO: 实现健康检查逻辑

	logger.Info("health check task completed")
	logger.LogWithDuration(start, "health check")
}

// archiveDocuments 归档任务
func (s *Scheduler) archiveDocuments() {
	start := time.Now()
	logger.Info("starting document archive task")

	// TODO: 实现文档归档逻辑

	logger.Info("document archive task completed")
	logger.LogWithDuration(start, "document archive")
}

// checkCalendarReminders 检查日程提醒任务
func (s *Scheduler) checkCalendarReminders() {
	start := time.Now()
	logger.Info("starting calendar reminder check task")

	// 获取数据库连接
	database := db.GetDB()
	if database == nil {
		log.Printf("[SCHEDULER] failed to get database connection")
		return
	}

	// 计算时间范围：从现在开始15分钟内
	now := time.Now()
	reminderWindow := now.Add(15 * time.Minute)

	// 查询即将开始的日程事件
	var events []models.CalendarEvent
	err := database.Where("start_time >= ? AND start_time <= ?", now, reminderWindow).
		Find(&events).Error

	if err != nil {
		log.Printf("[SCHEDULER] failed to query calendar events: %v", err)
		return
	}

	log.Printf("[SCHEDULER] found %d upcoming events in next 15 minutes", len(events))

	// 为每个事件创建提醒通知
	for _, event := range events {
		// 检查是否已经发送过提醒
		var existingNotification models.Notification
		exists := database.Where("related_id = ? AND related_type = 'calendar_event' AND type = 'calendar_reminder'", event.ID).
			First(&existingNotification).Error == nil

		if exists {
			continue // 已经发送过提醒，跳过
		}

		// 创建提醒
		notification := models.Notification{
			TeamID:      event.TeamID,
			UserID:      event.UserID,
			Title:       "日程提醒",
			Content:     格式化事件提醒内容(&event),
			Type:        "calendar_reminder",
			Category:    "important",
			Priority:    "normal",
			RelatedID:   event.ID,
			RelatedType: "calendar_event",
			CreatedBy:   "system",
		}

		if err := database.Create(&notification).Error; err != nil {
			log.Printf("[SCHEDULER] failed to create reminder notification for event %s: %v", event.ID, err)
		} else {
			log.Printf("[SCHEDULER] created reminder notification for event: %s", event.Title)
		}
	}

	logger.Info("calendar reminder check task completed")
	logger.LogWithDuration(start, "calendar reminder check")
}

// 格式化事件提醒内容
func 格式化事件提醒内容(event *models.CalendarEvent) string {
	startTime := event.StartTime.Format("15:04")
	if event.AllDay {
		return fmt.Sprintf("您的全天事件「%s」即将开始。", event.Title)
	}
	return fmt.Sprintf("您的日程「%s」将在 %s 开始。", event.Title, startTime)
}
