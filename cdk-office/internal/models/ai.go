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

// AIServiceConfig AI服务配置模型
type AIServiceConfig struct {
	ID            string    `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	ServiceName   string    `json:"service_name" gorm:"size:100"` // 服务商名称
	ServiceType   string    `json:"service_type"`                 // ai_chat, ai_embedding, ai_translation
	Provider      string    `json:"provider"`                     // openai, baidu, tencent, aliyun
	APIEndpoint   string    `json:"api_endpoint" gorm:"size:255"`
	APIKey        string    `json:"api_key" gorm:"size:255"`
	SecretKey     string    `json:"secret_key" gorm:"size:255"`
	AppID         string    `json:"app_id" gorm:"size:100"`
	Region        string    `json:"region" gorm:"size:50"`
	MaxRetries    int       `json:"max_retries" gorm:"default:3"`
	Timeout       int       `json:"timeout" gorm:"default:30"`       // 秒
	RateLimit     int       `json:"rate_limit" gorm:"default:100"`   // 每分钟请求数
	CustomHeaders string    `json:"custom_headers" gorm:"type:text"` // JSON格式
	CustomParams  string    `json:"custom_params" gorm:"type:text"`  // JSON格式
	IsEnabled     bool      `json:"is_enabled" gorm:"default:true"`
	IsDefault     bool      `json:"is_default" gorm:"default:false"`
	Priority      int       `json:"priority" gorm:"default:0"` // 优先级
	CreatedBy     string    `json:"created_by" gorm:"type:uuid"`
	UpdatedBy     string    `json:"updated_by" gorm:"type:uuid"`
	CreatedAt     time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt     time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

// BeforeCreate 创建前钩子
func (a *AIServiceConfig) BeforeCreate(tx *gorm.DB) error {
	if a.ID == "" {
		a.ID = uuid.New().String()
	}
	return nil
}

// KnowledgeQA 知识问答记录模型
type KnowledgeQA struct {
	ID         string    `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	UserID     string    `json:"user_id" gorm:"type:uuid;not null;index"`
	TeamID     string    `json:"team_id" gorm:"type:uuid;not null;index"`
	Question   string    `json:"question" gorm:"type:text;not null"`
	Answer     string    `json:"answer" gorm:"type:text"`
	Sources    string    `json:"sources" gorm:"type:text"` // JSON格式的文档源
	Confidence float32   `json:"confidence" gorm:"default:0"`
	Feedback   string    `json:"feedback" gorm:"type:text"`
	AIProvider string    `json:"ai_provider" gorm:"size:50"`
	MessageID  string    `json:"message_id" gorm:"size:100"`
	CreatedAt  time.Time `json:"created_at" gorm:"autoCreateTime"`
}

// BeforeCreate 创建前钩子
func (k *KnowledgeQA) BeforeCreate(tx *gorm.DB) error {
	if k.ID == "" {
		k.ID = uuid.New().String()
	}
	return nil
}

// DifyDocumentSync Dify文档同步记录模型
type DifyDocumentSync struct {
	ID             string    `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	DocumentID     string    `json:"document_id" gorm:"type:uuid;not null;index"` // 本地文档ID
	DifyDocumentID string    `json:"dify_document_id" gorm:"size:100;not null"`   // Dify文档ID
	DatasetID      string    `json:"dataset_id" gorm:"size:100;not null"`         // Dify数据集ID
	Title          string    `json:"title" gorm:"size:255;not null"`
	Content        string    `json:"content" gorm:"type:text"`
	DocumentType   string    `json:"document_type" gorm:"size:50"`
	TeamID         string    `json:"team_id" gorm:"type:uuid;not null;index"`
	SyncStatus     string    `json:"sync_status" gorm:"size:20;default:'pending'"` // pending, synced, failed
	IndexingStatus string    `json:"indexing_status" gorm:"size:20"`               // processing, completed, error
	ErrorMessage   string    `json:"error_message" gorm:"type:text"`
	CreatedBy      string    `json:"created_by" gorm:"type:uuid;not null"`
	CreatedAt      time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt      time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

// BeforeCreate 创建前钩子
func (d *DifyDocumentSync) BeforeCreate(tx *gorm.DB) error {
	if d.ID == "" {
		d.ID = uuid.New().String()
	}
	return nil
}

// KnowledgeQAV2 知识问答模型（重命名避免重复定义）
type KnowledgeQAV2 struct {
	ID         string    `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	UserID     string    `json:"user_id" gorm:"type:uuid;not null;index"`
	TeamID     string    `json:"team_id" gorm:"type:uuid;not null;index"`
	Question   string    `json:"question" gorm:"type:text;not null"`
	Answer     string    `json:"answer" gorm:"type:text"`
	Sources    []string  `json:"sources" gorm:"type:uuid[]"` // 引用文档ID
	Confidence float32   `json:"confidence"`
	Feedback   string    `json:"feedback" gorm:"size:20"` // positive, negative
	AIProvider string    `json:"ai_provider" gorm:"size:50"`
	CreatedAt  time.Time `json:"created_at" gorm:"autoCreateTime"`
}

// KnowledgeSyncRecord 知识同步记录模型
type KnowledgeSyncRecord struct {
	ID           string    `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	DocumentID   string    `json:"document_id" gorm:"type:uuid;not null;index"`
	Status       string    `json:"status" gorm:"size:20"` // pending, processing, success, failed
	ErrorMessage string    `json:"error_message" gorm:"type:text"`
	RetryCount   int       `json:"retry_count" gorm:"default:0"`
	CreatedAt    time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt    time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

// ServiceStatus 服务状态模型
type ServiceStatus struct {
	ID           string    `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	ServiceID    string    `json:"service_id" gorm:"type:uuid;not null;index"`
	ServiceType  string    `json:"service_type"`  // ai, ocr, sms, email
	Status       string    `json:"status"`        // healthy, degraded, unavailable
	ResponseTime int64     `json:"response_time"` // 毫秒
	SuccessRate  float64   `json:"success_rate"`  // 成功率 0-1
	ErrorCount   int       `json:"error_count"`   // 错误次数
	LastError    string    `json:"last_error" gorm:"type:text"`
	LastCheckAt  time.Time `json:"last_check_at"`
	UpdatedAt    time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

// PersonalKnowledgeBase 个人知识库模型
type PersonalKnowledgeBase struct {
	ID          string    `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	UserID      string    `json:"user_id" gorm:"type:uuid;not null;index"`
	Title       string    `json:"title" gorm:"size:255;not null"`
	Description string    `json:"description" gorm:"type:text"`
	Content     string    `json:"content" gorm:"type:text"`
	ContentType string    `json:"content_type" gorm:"size:50;default:'markdown'"` // markdown, text, html
	Tags        []string  `json:"tags" gorm:"type:text[]"`
	Category    string    `json:"category" gorm:"size:100"`                 // 分类：学习笔记、工作文档、生活记录等
	Privacy     string    `json:"privacy" gorm:"size:20;default:'private'"` // private, shared, public
	SourceType  string    `json:"source_type" gorm:"size:50"`               // manual, wechat, upload, scan
	SourceData  string    `json:"source_data" gorm:"type:text"`             // JSON格式的源数据
	IsShared    bool      `json:"is_shared" gorm:"default:false"`           // 是否已分享到团队知识库
	SharedAt    time.Time `json:"shared_at"`
	CreatedAt   time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt   time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

// BeforeCreate 创建前钩子
func (p *PersonalKnowledgeBase) BeforeCreate(tx *gorm.DB) error {
	if p.ID == "" {
		p.ID = uuid.New().String()
	}
	return nil
}

// WeChatRecord 微信聊天记录模型
type WeChatRecord struct {
	ID            string    `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	UserID        string    `json:"user_id" gorm:"type:uuid;not null;index"`
	SessionName   string    `json:"session_name" gorm:"size:255;not null"` // 聊天会话名称
	MessageType   string    `json:"message_type" gorm:"size:20;not null"`  // text, image, voice, video, file
	MessageID     string    `json:"message_id" gorm:"size:100;index"`      // 微信消息ID
	SenderName    string    `json:"sender_name" gorm:"size:100"`
	SenderID      string    `json:"sender_id" gorm:"size:100"`
	Content       string    `json:"content" gorm:"type:text"`                        // 文本内容或文件路径
	OriginalFile  string    `json:"original_file" gorm:"size:500"`                   // 原始文件路径
	ProcessedFile string    `json:"processed_file" gorm:"size:500"`                  // 处理后文件路径
	OCRText       string    `json:"ocr_text" gorm:"type:text"`                       // OCR识别的文本
	MessageTime   time.Time `json:"message_time"`                                    // 消息发送时间
	ProcessStatus string    `json:"process_status" gorm:"size:20;default:'pending'"` // pending, processing, completed, failed
	ExtractedInfo string    `json:"extracted_info" gorm:"type:text"`                 // JSON格式的提取信息
	IsArchived    bool      `json:"is_archived" gorm:"default:false"`                // 是否已归档到知识库
	ArchivedTo    string    `json:"archived_to" gorm:"type:uuid"`                    // 归档到的知识库文档ID
	CreatedAt     time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt     time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

// BeforeCreate 创建前钩子
func (w *WeChatRecord) BeforeCreate(tx *gorm.DB) error {
	if w.ID == "" {
		w.ID = uuid.New().String()
	}
	return nil
}

// PersonalKnowledgeShare 个人知识分享模型
type PersonalKnowledgeShare struct {
	ID           string    `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	KnowledgeID  string    `json:"knowledge_id" gorm:"type:uuid;not null;index"`
	UserID       string    `json:"user_id" gorm:"type:uuid;not null;index"`
	TeamID       string    `json:"team_id" gorm:"type:uuid;not null;index"`
	ApprovalID   string    `json:"approval_id" gorm:"type:uuid"`            // 关联审批流程ID
	ShareReason  string    `json:"share_reason" gorm:"type:text"`           // 分享理由
	Status       string    `json:"status" gorm:"size:20;default:'pending'"` // pending, approved, rejected
	ReviewerID   string    `json:"reviewer_id" gorm:"type:uuid"`
	ReviewReason string    `json:"review_reason" gorm:"type:text"`
	ReviewedAt   time.Time `json:"reviewed_at"`
	CreatedAt    time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt    time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

// BeforeCreate 创建前钩子
func (p *PersonalKnowledgeShare) BeforeCreate(tx *gorm.DB) error {
	if p.ID == "" {
		p.ID = uuid.New().String()
	}
	return nil
}
