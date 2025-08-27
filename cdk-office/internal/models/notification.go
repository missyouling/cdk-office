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

// Notification 通知模型
type Notification struct {
	ID             string    `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	TeamID         string    `json:"team_id" gorm:"type:uuid;not null;index"`   // 团队ID
	UserID         string    `json:"user_id" gorm:"type:uuid;not null;index"`   // 接收用户ID
	Title          string    `json:"title" gorm:"size:255;not null"`            // 通知标题
	Content        string    `json:"content" gorm:"type:text"`                  // 通知内容
	Type           string    `json:"type" gorm:"size:50;not null;index"`        // 通知类型: system, approval, document, task, mention
	Category       string    `json:"category" gorm:"size:50;default:'general'"` // 通知分类: general, urgent, important
	Priority       string    `json:"priority" gorm:"size:20;default:'normal'"`  // 优先级: low, normal, high, urgent
	IsRead         bool      `json:"is_read" gorm:"default:false;index"`        // 是否已读
	ReadAt         time.Time `json:"read_at"`                                   // 阅读时间
	IsArchived     bool      `json:"is_archived" gorm:"default:false;index"`    // 是否已归档
	ArchivedAt     time.Time `json:"archived_at"`                               // 归档时间
	RelatedID      string    `json:"related_id" gorm:"size:100;index"`          // 关联ID（如审批ID、文档ID等）
	RelatedType    string    `json:"related_type" gorm:"size:50"`               // 关联类型
	ActionRequired bool      `json:"action_required" gorm:"default:false"`      // 是否需要操作
	ActionTaken    bool      `json:"action_taken" gorm:"default:false"`         // 是否已操作
	CreatedBy      string    `json:"created_by" gorm:"type:uuid"`               // 创建者
	CreatedAt      time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt      time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

// BeforeCreate 创建前钩子
func (n *Notification) BeforeCreate(tx *gorm.DB) error {
	if n.ID == "" {
		n.ID = uuid.New().String()
	}
	return nil
}

// NotificationTemplate 通知模板模型
type NotificationTemplate struct {
	ID          string    `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	TeamID      string    `json:"team_id" gorm:"type:uuid;not null;index"` // 团队ID
	Name        string    `json:"name" gorm:"size:255;not null"`           // 模板名称
	Description string    `json:"description" gorm:"type:text"`            // 描述
	Type        string    `json:"type" gorm:"size:50;not null"`            // 通知类型
	Subject     string    `json:"subject" gorm:"size:255"`                 // 通知主题
	Content     string    `json:"content" gorm:"type:text"`                // 通知内容模板
	IsDefault   bool      `json:"is_default" gorm:"default:false"`         // 是否默认模板
	CreatedBy   string    `json:"created_by" gorm:"type:uuid;not null"`
	UpdatedBy   string    `json:"updated_by" gorm:"type:uuid"`
	CreatedAt   time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt   time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

// BeforeCreate 创建前钩子
func (t *NotificationTemplate) BeforeCreate(tx *gorm.DB) error {
	if t.ID == "" {
		t.ID = uuid.New().String()
	}
	return nil
}

// NotificationPreference 用户通知偏好设置模型
type NotificationPreference struct {
	ID             string    `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	UserID         string    `json:"user_id" gorm:"type:uuid;not null;uniqueIndex"`        // 用户ID
	EmailEnabled   bool      `json:"email_enabled" gorm:"default:true"`                    // 邮件通知启用
	EmailFrequency string    `json:"email_frequency" gorm:"size:20;default:'immediately'"` // 邮件通知频率: immediately, daily, weekly
	PushEnabled    bool      `json:"push_enabled" gorm:"default:true"`                     // 推送通知启用
	InAppEnabled   bool      `json:"in_app_enabled" gorm:"default:true"`                   // 应用内通知启用
	SmsEnabled     bool      `json:"sms_enabled" gorm:"default:false"`                     // 短信通知启用
	DesktopEnabled bool      `json:"desktop_enabled" gorm:"default:true"`                  // 桌面通知启用
	SoundEnabled   bool      `json:"sound_enabled" gorm:"default:true"`                    // 声音提醒启用
	CreatedAt      time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt      time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

// BeforeCreate 创建前钩子
func (p *NotificationPreference) BeforeCreate(tx *gorm.DB) error {
	if p.ID == "" {
		p.ID = uuid.New().String()
	}
	return nil
}

// NotificationChannel 通知渠道模型
type NotificationChannel struct {
	ID        string    `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	TeamID    string    `json:"team_id" gorm:"type:uuid;not null;index"` // 团队ID
	Name      string    `json:"name" gorm:"size:100;not null"`           // 渠道名称
	Type      string    `json:"type" gorm:"size:50;not null"`            // 渠道类型: email, sms, webhook, slack, wechat
	Config    string    `json:"config" gorm:"type:text"`                 // 渠道配置(JSON格式)
	IsActive  bool      `json:"is_active" gorm:"default:true"`           // 是否激活
	CreatedBy string    `json:"created_by" gorm:"type:uuid;not null"`
	UpdatedBy string    `json:"updated_by" gorm:"type:uuid"`
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

// BeforeCreate 创建前钩子
func (c *NotificationChannel) BeforeCreate(tx *gorm.DB) error {
	if c.ID == "" {
		c.ID = uuid.New().String()
	}
	return nil
}
