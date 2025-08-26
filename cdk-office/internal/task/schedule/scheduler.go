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
	"log"
	"time"

	"github.com/linux-do/cdk-office/internal/config"
	"github.com/linux-do/cdk-office/internal/db"
	"github.com/linux-do/cdk-office/internal/logger"
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
