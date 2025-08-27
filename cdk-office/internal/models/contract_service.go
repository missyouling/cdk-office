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

// ContractServiceConfig 合同服务配置模型
type ContractServiceConfig struct {
	ID            string    `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	ServiceName   string    `json:"service_name" gorm:"size:100;not null"`   // 服务商名称
	ServiceType   string    `json:"service_type" gorm:"size:50;not null"`    // ca, sms, blockchain
	Provider      string    `json:"provider" gorm:"size:50;not null"`        // fadada, esign, aliyun, tencent等
	
	// 配置信息
	APIEndpoint   string    `json:"api_endpoint" gorm:"size:255"`
	APIKey        string    `json:"api_key" gorm:"size:255"`
	SecretKey     string    `json:"secret_key" gorm:"size:255"`
	AppID         string    `json:"app_id" gorm:"size:100"`
	AppSecret     string    `json:"app_secret" gorm:"size:255"`
	Region        string    `json:"region" gorm:"size:50"`
	
	// 高级配置
	MaxRetries    int       `json:"max_retries" gorm:"default:3"`
	Timeout       int       `json:"timeout" gorm:"default:30"`               // 秒
	RateLimit     int       `json:"rate_limit" gorm:"default:100"`           // 每分钟请求数
	
	// 自定义配置
	CustomHeaders string    `json:"custom_headers" gorm:"type:text"`         // JSON格式
	CustomParams  string    `json:"custom_params" gorm:"type:text"`          // JSON格式
	
	// 状态管理
	IsEnabled     bool      `json:"is_enabled" gorm:"default:true"`
	IsDefault     bool      `json:"is_default" gorm:"default:false"`
	Priority      int       `json:"priority" gorm:"default:0"`               // 优先级
	
	// 健康状态
	HealthStatus  string    `json:"health_status" gorm:"size:20;default:'unknown'"` // healthy, degraded, unavailable, unknown
	LastCheckAt   time.Time `json:"last_check_at"`
	ErrorCount    int       `json:"error_count" gorm:"default:0"`
	LastError     string    `json:"last_error" gorm:"type:text"`
	
	// 审计字段
	CreatedBy     string    `json:"created_by" gorm:"type:uuid"`
	UpdatedBy     string    `json:"updated_by" gorm:"type:uuid"`
	CreatedAt     time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt     time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

// BeforeCreate 创建前钩子
func (c *ContractServiceConfig) BeforeCreate(tx *gorm.DB) error {
	if c.ID == "" {
		c.ID = uuid.New().String()
	}
	return nil
}

// ContractWorkflow 合同工作流配置模型
type ContractWorkflow struct {
	ID                   string    `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	TeamID               string    `json:"team_id" gorm:"type:uuid;not null;index"`
	Name                 string    `json:"name" gorm:"size:255;not null"`
	Description          string    `json:"description" gorm:"type:text"`
	
	// 知识库集成配置
	AutoSubmitKnowledge  bool      `json:"auto_submit_knowledge" gorm:"default:true"`   // 自动提交到知识库
	RequireApproval      bool      `json:"require_approval" gorm:"default:false"`       // 是否需要审批
	ApprovalRoles        string    `json:"approval_roles" gorm:"type:text"`             // 审批角色(JSON数组)
	ApprovalTimeout      int       `json:"approval_timeout" gorm:"default:24"`          // 审批超时时间(小时)
	
	// 通知配置
	NotifyOnComplete     bool      `json:"notify_on_complete" gorm:"default:true"`      // 签署完成通知
	NotifyOnApproval     bool      `json:"notify_on_approval" gorm:"default:true"`      // 审批通知
	NotificationChannels string    `json:"notification_channels" gorm:"type:text"`     // 通知渠道(JSON数组)
	
	// Dify集成配置
	DifyDatasetID        string    `json:"dify_dataset_id" gorm:"size:100"`             // Dify数据集ID
	DifyProcessingMode   string    `json:"dify_processing_mode" gorm:"size:20;default:'auto'"` // auto, manual
	DifyTags             string    `json:"dify_tags" gorm:"type:text"`                  // 自动标签(JSON数组)
	
	// 状态管理
	IsActive             bool      `json:"is_active" gorm:"default:true"`
	IsDefault            bool      `json:"is_default" gorm:"default:false"`
	
	// 审计字段
	CreatedBy            string    `json:"created_by" gorm:"type:uuid;not null"`
	UpdatedBy            string    `json:"updated_by" gorm:"type:uuid"`
	CreatedAt            time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt            time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

// ContractStatistics 合同统计模型
type ContractStatistics struct {
	ID               string    `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	TeamID           string    `json:"team_id" gorm:"type:uuid;not null;index"`
	StatDate         time.Time `json:"stat_date" gorm:"type:date;not null;index"`  // 统计日期
	
	// 合同数量统计
	TotalContracts   int       `json:"total_contracts" gorm:"default:0"`           // 总合同数
	DraftContracts   int       `json:"draft_contracts" gorm:"default:0"`           // 草稿合同数
	PendingContracts int       `json:"pending_contracts" gorm:"default:0"`         // 待签署合同数
	SigningContracts int       `json:"signing_contracts" gorm:"default:0"`         // 签署中合同数
	CompletedContracts int     `json:"completed_contracts" gorm:"default:0"`       // 已完成合同数
	RejectedContracts int      `json:"rejected_contracts" gorm:"default:0"`        // 已拒绝合同数
	CancelledContracts int     `json:"cancelled_contracts" gorm:"default:0"`       // 已取消合同数
	ExpiredContracts int       `json:"expired_contracts" gorm:"default:0"`         // 已过期合同数
	
	// 效率统计
	AvgSigningTime   int       `json:"avg_signing_time" gorm:"default:0"`          // 平均签署时间(小时)
	SuccessRate      float64   `json:"success_rate" gorm:"type:decimal(5,2);default:0"` // 成功率
	
	// 知识库统计
	KnowledgeSubmissions int   `json:"knowledge_submissions" gorm:"default:0"`     // 知识库提交数
	KnowledgeApprovals   int   `json:"knowledge_approvals" gorm:"default:0"`       // 知识库审批通过数
	
	// 审计字段
	CreatedAt        time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt        time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

// ContractNotification 合同通知模型
type ContractNotification struct {
	ID           string    `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	ContractID   string    `json:"contract_id" gorm:"type:uuid;not null;index"`
	RecipientID  string    `json:"recipient_id" gorm:"type:uuid;not null"`        // 接收人ID
	
	// 通知内容
	Type         string    `json:"type" gorm:"size:50;not null"`                  // sign_request, sign_reminder, completed, rejected等
	Title        string    `json:"title" gorm:"size:255;not null"`
	Content      string    `json:"content" gorm:"type:text;not null"`
	ActionURL    string    `json:"action_url" gorm:"size:500"`                    // 操作链接
	
	// 发送配置
	Channels     string    `json:"channels" gorm:"type:text"`                     // 发送渠道(JSON数组): email, sms, system
	Priority     string    `json:"priority" gorm:"size:20;default:'normal'"`      // low, normal, high, urgent
	
	// 发送状态
	Status       string    `json:"status" gorm:"size:20;default:'pending'"`       // pending, sent, failed, read
	SentAt       *time.Time `json:"sent_at"`
	ReadAt       *time.Time `json:"read_at"`
	ErrorMessage string    `json:"error_message" gorm:"type:text"`
	
	// 重试机制
	RetryCount   int       `json:"retry_count" gorm:"default:0"`
	MaxRetries   int       `json:"max_retries" gorm:"default:3"`
	NextRetryAt  *time.Time `json:"next_retry_at"`
	
	// 审计字段
	CreatedAt    time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt    time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}