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

// KnowledgeQA 知识问答模型
type KnowledgeQA struct {
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
