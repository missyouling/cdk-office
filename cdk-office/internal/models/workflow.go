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

// WorkflowDefinition 工作流定义模型
type WorkflowDefinition struct {
	ID          string    `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	TeamID      string    `json:"team_id" gorm:"type:uuid;not null;index"`
	Name        string    `json:"name" gorm:"size:255;not null"`         // 工作流名称
	Description string    `json:"description" gorm:"type:text"`          // 工作流描述
	Definition  string    `json:"definition" gorm:"type:text"`           // 工作流定义(JSON格式)
	Version     int       `json:"version" gorm:"default:1"`              // 版本号
	Status      string    `json:"status" gorm:"size:20;default:'draft'"` // 状态: draft, active, archived
	Category    string    `json:"category" gorm:"size:50"`               // 工作流分类
	Tags        string    `json:"tags" gorm:"type:text"`                 // 标签(JSON数组)
	IsTemplate  bool      `json:"is_template" gorm:"default:false"`      // 是否为模板
	CreatedBy   string    `json:"created_by" gorm:"type:uuid;not null"`
	UpdatedBy   string    `json:"updated_by" gorm:"type:uuid"`
	CreatedAt   time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt   time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

// BeforeCreate 创建前钩子
func (w *WorkflowDefinition) BeforeCreate(tx *gorm.DB) error {
	if w.ID == "" {
		w.ID = uuid.New().String()
	}
	return nil
}

// WorkflowInstance 工作流实例模型
type WorkflowInstance struct {
	ID            string     `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	WorkflowDefID string     `json:"workflow_def_id" gorm:"type:uuid;not null;index"`
	TeamID        string     `json:"team_id" gorm:"type:uuid;not null;index"`
	Name          string     `json:"name" gorm:"size:255"`                    // 实例名称
	Status        string     `json:"status" gorm:"size:20;default:'pending'"` // 状态: pending, running, completed, failed, cancelled, paused
	CurrentStep   string     `json:"current_step" gorm:"size:255"`            // 当前步骤ID
	InputData     string     `json:"input_data" gorm:"type:text"`             // 输入数据(JSON格式)
	OutputData    string     `json:"output_data" gorm:"type:text"`            // 输出数据(JSON格式)
	Variables     string     `json:"variables" gorm:"type:text"`              // 变量数据(JSON格式)
	Priority      int        `json:"priority" gorm:"default:0"`               // 优先级
	StartedAt     time.Time  `json:"started_at"`                              // 开始时间
	CompletedAt   *time.Time `json:"completed_at"`                            // 完成时间
	ErrorMessage  string     `json:"error_message" gorm:"type:text"`          // 错误信息
	CreatedBy     string     `json:"created_by" gorm:"type:uuid;not null"`
	CreatedAt     time.Time  `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt     time.Time  `json:"updated_at" gorm:"autoUpdateTime"`
}

// BeforeCreate 创建前钩子
func (w *WorkflowInstance) BeforeCreate(tx *gorm.DB) error {
	if w.ID == "" {
		w.ID = uuid.New().String()
	}
	return nil
}

// WorkflowStep 工作流步骤实例模型
type WorkflowStep struct {
	ID           string     `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	WorkflowID   string     `json:"workflow_id" gorm:"type:uuid;not null;index"`
	StepDefID    string     `json:"step_def_id" gorm:"size:255"`             // 步骤定义ID
	Name         string     `json:"name" gorm:"size:255"`                    // 步骤名称
	Type         string     `json:"type" gorm:"size:50"`                     // 步骤类型: approval, condition, activity, parallel
	Status       string     `json:"status" gorm:"size:20;default:'pending'"` // 状态: pending, running, completed, failed, skipped
	InputData    string     `json:"input_data" gorm:"type:text"`             // 输入数据(JSON格式)
	OutputData   string     `json:"output_data" gorm:"type:text"`            // 输出数据(JSON格式)
	Config       string     `json:"config" gorm:"type:text"`                 // 步骤配置(JSON格式)
	AssignedTo   string     `json:"assigned_to" gorm:"type:uuid"`            // 分配给的用户ID
	StartedAt    time.Time  `json:"started_at"`                              // 开始时间
	CompletedAt  *time.Time `json:"completed_at"`                            // 完成时间
	ErrorMessage string     `json:"error_message" gorm:"type:text"`          // 错误信息
	RetryCount   int        `json:"retry_count" gorm:"default:0"`            // 重试次数
	CreatedAt    time.Time  `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt    time.Time  `json:"updated_at" gorm:"autoUpdateTime"`
}

// BeforeCreate 创建前钩子
func (w *WorkflowStep) BeforeCreate(tx *gorm.DB) error {
	if w.ID == "" {
		w.ID = uuid.New().String()
	}
	return nil
}

// WorkflowLog 工作流日志模型
type WorkflowLog struct {
	ID         string    `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	WorkflowID string    `json:"workflow_id" gorm:"type:uuid;not null;index"`
	StepID     string    `json:"step_id" gorm:"type:uuid;index"` // 步骤ID(可选)
	Level      string    `json:"level" gorm:"size:20"`           // 日志级别: info, warn, error, debug
	Message    string    `json:"message" gorm:"type:text"`       // 日志消息
	Data       string    `json:"data" gorm:"type:text"`          // 相关数据(JSON格式)
	UserID     string    `json:"user_id" gorm:"type:uuid"`       // 操作用户ID
	IPAddress  string    `json:"ip_address" gorm:"size:45"`      // IP地址
	UserAgent  string    `json:"user_agent" gorm:"size:500"`     // 用户代理
	CreatedAt  time.Time `json:"created_at" gorm:"autoCreateTime"`
}

// BeforeCreate 创建前钩子
func (w *WorkflowLog) BeforeCreate(tx *gorm.DB) error {
	if w.ID == "" {
		w.ID = uuid.New().String()
	}
	return nil
}

// WorkflowTemplate 工作流模板模型
type WorkflowTemplate struct {
	ID          string    `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	Name        string    `json:"name" gorm:"size:255;not null"`  // 模板名称
	Description string    `json:"description" gorm:"type:text"`   // 模板描述
	Category    string    `json:"category" gorm:"size:50"`        // 模板分类
	Definition  string    `json:"definition" gorm:"type:text"`    // 模板定义(JSON格式)
	Icon        string    `json:"icon" gorm:"size:100"`           // 图标
	Tags        string    `json:"tags" gorm:"type:text"`          // 标签(JSON数组)
	IsPublic    bool      `json:"is_public" gorm:"default:false"` // 是否公开
	UseCount    int       `json:"use_count" gorm:"default:0"`     // 使用次数
	Rating      float32   `json:"rating" gorm:"default:0"`        // 评分
	CreatedBy   string    `json:"created_by" gorm:"type:uuid;not null"`
	TeamID      string    `json:"team_id" gorm:"type:uuid;index"` // 团队ID(为空表示全局模板)
	CreatedAt   time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt   time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

// BeforeCreate 创建前钩子
func (w *WorkflowTemplate) BeforeCreate(tx *gorm.DB) error {
	if w.ID == "" {
		w.ID = uuid.New().String()
	}
	return nil
}

// WorkflowVariable 工作流变量模型
type WorkflowVariable struct {
	ID           string    `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	WorkflowID   string    `json:"workflow_id" gorm:"type:uuid;not null;index"`
	Name         string    `json:"name" gorm:"size:100;not null"`     // 变量名
	Type         string    `json:"type" gorm:"size:50"`               // 变量类型: string, number, boolean, object, array
	Value        string    `json:"value" gorm:"type:text"`            // 变量值(JSON格式)
	Description  string    `json:"description" gorm:"type:text"`      // 变量描述
	IsRequired   bool      `json:"is_required" gorm:"default:false"`  // 是否必需
	IsReadOnly   bool      `json:"is_read_only" gorm:"default:false"` // 是否只读
	DefaultValue string    `json:"default_value" gorm:"type:text"`    // 默认值
	CreatedAt    time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt    time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

// BeforeCreate 创建前钩子
func (w *WorkflowVariable) BeforeCreate(tx *gorm.DB) error {
	if w.ID == "" {
		w.ID = uuid.New().String()
	}
	return nil
}

// WorkflowPermission 工作流权限模型
type WorkflowPermission struct {
	ID         string    `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	WorkflowID string    `json:"workflow_id" gorm:"type:uuid;not null;index"`
	UserID     string    `json:"user_id" gorm:"type:uuid;not null;index"`
	Permission string    `json:"permission" gorm:"size:20;not null"`   // 权限类型: view, execute, edit, admin
	GrantedBy  string    `json:"granted_by" gorm:"type:uuid;not null"` // 授权人
	CreatedAt  time.Time `json:"created_at" gorm:"autoCreateTime"`
}

// BeforeCreate 创建前钩子
func (w *WorkflowPermission) BeforeCreate(tx *gorm.DB) error {
	if w.ID == "" {
		w.ID = uuid.New().String()
	}
	return nil
}

// WorkflowSchedule 工作流调度模型
type WorkflowSchedule struct {
	ID             string     `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	WorkflowDefID  string     `json:"workflow_def_id" gorm:"type:uuid;not null;index"`
	Name           string     `json:"name" gorm:"size:255;not null"`   // 调度名称
	CronExpression string     `json:"cron_expression" gorm:"size:100"` // Cron表达式
	IsEnabled      bool       `json:"is_enabled" gorm:"default:true"`  // 是否启用
	InputData      string     `json:"input_data" gorm:"type:text"`     // 输入数据(JSON格式)
	LastRunAt      *time.Time `json:"last_run_at"`                     // 上次运行时间
	NextRunAt      *time.Time `json:"next_run_at"`                     // 下次运行时间
	RunCount       int        `json:"run_count" gorm:"default:0"`      // 运行次数
	FailCount      int        `json:"fail_count" gorm:"default:0"`     // 失败次数
	CreatedBy      string     `json:"created_by" gorm:"type:uuid;not null"`
	CreatedAt      time.Time  `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt      time.Time  `json:"updated_at" gorm:"autoUpdateTime"`
}

// BeforeCreate 创建前钩子
func (w *WorkflowSchedule) BeforeCreate(tx *gorm.DB) error {
	if w.ID == "" {
		w.ID = uuid.New().String()
	}
	return nil
}

// PDFTask PDF处理任务模型
type PDFTask struct {
	ID           string     `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	UserID       string     `json:"user_id" gorm:"type:uuid;not null;index"`
	TeamID       string     `json:"team_id" gorm:"type:uuid;not null;index"`
	Operation    string     `json:"operation" gorm:"size:50;not null"`       // merge, split, compress, rotate, watermark, convert
	Status       string     `json:"status" gorm:"size:20;default:'pending'"` // pending, processing, completed, failed
	InputFiles   string     `json:"input_files" gorm:"type:text"`            // JSON格式的输入文件列表
	OutputFiles  string     `json:"output_files" gorm:"type:text"`           // JSON格式的输出文件列表
	Parameters   string     `json:"parameters" gorm:"type:text"`             // JSON格式的操作参数
	ErrorMessage string     `json:"error_message" gorm:"type:text"`          // 错误信息
	Progress     int        `json:"progress" gorm:"default:0"`               // 处理进度（0-100）
	StartedAt    *time.Time `json:"started_at"`                              // 开始处理时间
	CompletedAt  *time.Time `json:"completed_at"`                            // 完成时间
	CreatedAt    time.Time  `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt    time.Time  `json:"updated_at" gorm:"autoUpdateTime"`
}

// BeforeCreate 创建前钩子
func (p *PDFTask) BeforeCreate(tx *gorm.DB) error {
	if p.ID == "" {
		p.ID = uuid.New().String()
	}
	return nil
}

// PDFTemplate PDF模板模型
type PDFTemplate struct {
	ID          string    `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	Name        string    `json:"name" gorm:"size:255;not null"`  // 模板名称
	Description string    `json:"description" gorm:"type:text"`   // 模板描述
	Category    string    `json:"category" gorm:"size:50"`        // 模板分类
	Template    string    `json:"template" gorm:"type:text"`      // PDF模板内容
	IsPublic    bool      `json:"is_public" gorm:"default:false"` // 是否公开
	UseCount    int       `json:"use_count" gorm:"default:0"`     // 使用次数
	CreatedBy   string    `json:"created_by" gorm:"type:uuid;not null"`
	TeamID      string    `json:"team_id" gorm:"type:uuid;index"` // 团队ID(为空表示全局模板)
	CreatedAt   time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt   time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

// BeforeCreate 创建前钩子
func (p *PDFTemplate) BeforeCreate(tx *gorm.DB) error {
	if p.ID == "" {
		p.ID = uuid.New().String()
	}
	return nil
}

// FilePreview 文件预览记录模型
type FilePreview struct {
	ID           string     `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	DocumentID   string     `json:"document_id" gorm:"size:255;not null;index"`
	FileName     string     `json:"file_name" gorm:"size:255;not null"`
	FileType     string     `json:"file_type" gorm:"size:50"` // 文件扩展名
	UserID       string     `json:"user_id" gorm:"type:uuid;not null;index"`
	TeamID       string     `json:"team_id" gorm:"type:uuid;not null;index"`
	Provider     string     `json:"provider" gorm:"size:50;not null"`        // dify, kkfileview
	PreviewURL   string     `json:"preview_url" gorm:"size:500"`             // 预览链接
	ThumbnailURL string     `json:"thumbnail_url" gorm:"size:500"`           // 缩略图链接
	Status       string     `json:"status" gorm:"size:20;default:'pending'"` // pending, processing, completed, failed
	ErrorMessage string     `json:"error_message" gorm:"type:text"`          // 错误信息
	ViewCount    int        `json:"view_count" gorm:"default:0"`             // 查看次数
	LastViewedAt *time.Time `json:"last_viewed_at"`                          // 最后查看时间
	CreatedAt    time.Time  `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt    time.Time  `json:"updated_at" gorm:"autoUpdateTime"`
}

// BeforeCreate 创建前钩子
func (f *FilePreview) BeforeCreate(tx *gorm.DB) error {
	if f.ID == "" {
		f.ID = uuid.New().String()
	}
	return nil
}

// FilePreviewConfig 文件预览配置模型
type FilePreviewConfig struct {
	ID            string    `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	Provider      string    `json:"provider" gorm:"size:50;not null"` // dify, kkfileview
	IsEnabled     bool      `json:"is_enabled" gorm:"default:true"`   // 是否启用
	Configuration string    `json:"configuration" gorm:"type:text"`   // JSON格式的配置信息
	Description   string    `json:"description" gorm:"type:text"`     // 配置描述
	CreatedBy     string    `json:"created_by" gorm:"type:uuid;not null"`
	UpdatedBy     string    `json:"updated_by" gorm:"type:uuid"`
	CreatedAt     time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt     time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

// BeforeCreate 创建前钩子
func (f *FilePreviewConfig) BeforeCreate(tx *gorm.DB) error {
	if f.ID == "" {
		f.ID = uuid.New().String()
	}
	return nil
}
