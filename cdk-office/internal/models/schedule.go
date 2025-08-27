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

package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ScheduledTask 调度任务模型
type ScheduledTask struct {
	ID           string     `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	TeamID       string     `json:"team_id" gorm:"type:uuid;not null;index"`
	Name         string     `json:"name" gorm:"size:255;not null"`          // 任务名称
	Description  string     `json:"description" gorm:"type:text"`           // 任务描述
	TaskType     string     `json:"task_type" gorm:"size:50;not null"`      // 任务类型: workflow, script, http_request, email
	TaskConfig   string     `json:"task_config" gorm:"type:text"`           // 任务配置(JSON格式)
	CronExpr     string     `json:"cron_expr" gorm:"size:100;not null"`     // Cron表达式
	TimeZone     string     `json:"time_zone" gorm:"size:50;default:'UTC'"` // 时区
	IsEnabled    bool       `json:"is_enabled" gorm:"default:true"`         // 是否启用
	MaxRetries   int        `json:"max_retries" gorm:"default:3"`           // 最大重试次数
	Timeout      int        `json:"timeout" gorm:"default:300"`             // 超时时间(秒)
	LastRunAt    *time.Time `json:"last_run_at"`                            // 上次运行时间
	NextRunAt    *time.Time `json:"next_run_at"`                            // 下次运行时间
	RunCount     int        `json:"run_count" gorm:"default:0"`             // 运行次数
	SuccessCount int        `json:"success_count" gorm:"default:0"`         // 成功次数
	FailureCount int        `json:"failure_count" gorm:"default:0"`         // 失败次数
	LastError    string     `json:"last_error" gorm:"type:text"`            // 最后一次错误信息
	CreatedBy    string     `json:"created_by" gorm:"type:uuid;not null"`
	UpdatedBy    string     `json:"updated_by" gorm:"type:uuid"`
	CreatedAt    time.Time  `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt    time.Time  `json:"updated_at" gorm:"autoUpdateTime"`
}

// BeforeCreate 创建前钩子
func (s *ScheduledTask) BeforeCreate(tx *gorm.DB) error {
	if s.ID == "" {
		s.ID = uuid.New().String()
	}
	return nil
}

// TaskExecution 任务执行记录模型
type TaskExecution struct {
	ID           string     `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	TaskID       string     `json:"task_id" gorm:"type:uuid;not null;index"`
	ExecutionID  string     `json:"execution_id" gorm:"size:100"`   // 执行ID(用于关联外部系统)
	Status       string     `json:"status" gorm:"size:20;not null"` // 状态: pending, running, completed, failed, cancelled
	StartTime    time.Time  `json:"start_time" gorm:"not null"`
	EndTime      *time.Time `json:"end_time"`
	Duration     int        `json:"duration"`                       // 执行时长(毫秒)
	RetryCount   int        `json:"retry_count" gorm:"default:0"`   // 重试次数
	ErrorMessage string     `json:"error_message" gorm:"type:text"` // 错误信息
	Output       string     `json:"output" gorm:"type:text"`        // 执行输出
	Metadata     string     `json:"metadata" gorm:"type:text"`      // 元数据(JSON格式)
	CreatedAt    time.Time  `json:"created_at" gorm:"autoCreateTime"`
}

// BeforeCreate 创建前钩子
func (t *TaskExecution) BeforeCreate(tx *gorm.DB) error {
	if t.ID == "" {
		t.ID = uuid.New().String()
	}
	return nil
}

// TaskTemplate 任务模板模型
type TaskTemplate struct {
	ID          string    `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	Name        string    `json:"name" gorm:"size:255;not null"`     // 模板名称
	Description string    `json:"description" gorm:"type:text"`      // 模板描述
	Category    string    `json:"category" gorm:"size:50"`           // 模板分类
	TaskType    string    `json:"task_type" gorm:"size:50;not null"` // 任务类型
	Template    string    `json:"template" gorm:"type:text"`         // 模板配置(JSON格式)
	Icon        string    `json:"icon" gorm:"size:100"`              // 图标
	Tags        string    `json:"tags" gorm:"type:text"`             // 标签(JSON数组)
	IsPublic    bool      `json:"is_public" gorm:"default:false"`    // 是否公开
	UseCount    int       `json:"use_count" gorm:"default:0"`        // 使用次数
	Rating      float32   `json:"rating" gorm:"default:0"`           // 评分
	CreatedBy   string    `json:"created_by" gorm:"type:uuid;not null"`
	TeamID      string    `json:"team_id" gorm:"type:uuid;index"` // 团队ID(为空表示全局模板)
	CreatedAt   time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt   time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

// BeforeCreate 创建前钩子
func (t *TaskTemplate) BeforeCreate(tx *gorm.DB) error {
	if t.ID == "" {
		t.ID = uuid.New().String()
	}
	return nil
}

// TaskNotification 任务通知模型
type TaskNotification struct {
	ID           string     `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	TaskID       string     `json:"task_id" gorm:"type:uuid;not null;index"`
	NotifyOn     string     `json:"notify_on" gorm:"size:20;not null"`   // 通知时机: success, failure, both
	NotifyType   string     `json:"notify_type" gorm:"size:20;not null"` // 通知类型: email, webhook, slack
	NotifyConfig string     `json:"notify_config" gorm:"type:text"`      // 通知配置(JSON格式)
	IsEnabled    bool       `json:"is_enabled" gorm:"default:true"`      // 是否启用
	LastNotifyAt *time.Time `json:"last_notify_at"`                      // 最后通知时间
	NotifyCount  int        `json:"notify_count" gorm:"default:0"`       // 通知次数
	CreatedBy    string     `json:"created_by" gorm:"type:uuid;not null"`
	CreatedAt    time.Time  `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt    time.Time  `json:"updated_at" gorm:"autoUpdateTime"`
}

// BeforeCreate 创建前钩子
func (t *TaskNotification) BeforeCreate(tx *gorm.DB) error {
	if t.ID == "" {
		t.ID = uuid.New().String()
	}
	return nil
}

// TaskLog 任务日志模型
type TaskLog struct {
	ID          string    `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	TaskID      string    `json:"task_id" gorm:"type:uuid;not null;index"`
	ExecutionID string    `json:"execution_id" gorm:"type:uuid;index"` // 执行记录ID
	Level       string    `json:"level" gorm:"size:20"`                // 日志级别: info, warn, error, debug
	Message     string    `json:"message" gorm:"type:text"`            // 日志消息
	Data        string    `json:"data" gorm:"type:text"`               // 相关数据(JSON格式)
	Source      string    `json:"source" gorm:"size:100"`              // 日志来源
	CreatedAt   time.Time `json:"created_at" gorm:"autoCreateTime"`
}

// BeforeCreate 创建前钩子
func (t *TaskLog) BeforeCreate(tx *gorm.DB) error {
	if t.ID == "" {
		t.ID = uuid.New().String()
	}
	return nil
}

// TaskDependency 任务依赖关系模型
type TaskDependency struct {
	ID          string    `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	TaskID      string    `json:"task_id" gorm:"type:uuid;not null;index"`      // 当前任务ID
	DependsOnID string    `json:"depends_on_id" gorm:"type:uuid;not null"`      // 依赖的任务ID
	DependType  string    `json:"depend_type" gorm:"size:20;default:'success'"` // 依赖类型: success, completion, failure
	IsEnabled   bool      `json:"is_enabled" gorm:"default:true"`               // 是否启用
	CreatedAt   time.Time `json:"created_at" gorm:"autoCreateTime"`
}

// BeforeCreate 创建前钩子
func (t *TaskDependency) BeforeCreate(tx *gorm.DB) error {
	if t.ID == "" {
		t.ID = uuid.New().String()
	}
	return nil
}

// ScheduleGroup 调度组模型
type ScheduleGroup struct {
	ID          string    `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	TeamID      string    `json:"team_id" gorm:"type:uuid;not null;index"`
	Name        string    `json:"name" gorm:"size:255;not null"`  // 组名称
	Description string    `json:"description" gorm:"type:text"`   // 组描述
	IsEnabled   bool      `json:"is_enabled" gorm:"default:true"` // 是否启用
	Priority    int       `json:"priority" gorm:"default:0"`      // 优先级
	CreatedBy   string    `json:"created_by" gorm:"type:uuid;not null"`
	CreatedAt   time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt   time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

// BeforeCreate 创建前钩子
func (s *ScheduleGroup) BeforeCreate(tx *gorm.DB) error {
	if s.ID == "" {
		s.ID = uuid.New().String()
	}
	return nil
}
