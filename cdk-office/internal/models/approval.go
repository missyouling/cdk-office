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

// ApprovalProcess 审批流程模型
type ApprovalProcess struct {
	ID            string    `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	TeamID        string    `json:"team_id" gorm:"type:uuid;not null;index"`
	Name          string    `json:"name" gorm:"size:255;not null"`                // 审批流程名称
	Description   string    `json:"description" gorm:"type:text"`                 // 描述
	DocumentID    string    `json:"document_id" gorm:"type:uuid;index"`           // 关联文档ID
	DocumentName  string    `json:"document_name" gorm:"size:255"`                // 文档名称
	RequestorID   string    `json:"requestor_id" gorm:"type:uuid;not null;index"` // 申请人ID
	RequestorName string    `json:"requestor_name" gorm:"size:100"`               // 申请人姓名
	ApproverID    string    `json:"approver_id" gorm:"type:uuid;index"`           // 审批人ID
	ApproverName  string    `json:"approver_name" gorm:"size:100"`                // 审批人姓名
	Status        string    `json:"status" gorm:"size:20;default:'pending'"`      // 状态: pending, approved, rejected, cancelled
	ApprovalType  string    `json:"approval_type" gorm:"size:50;not null"`        // 审批类型: document_upload, document_update, document_delete
	Comments      string    `json:"comments" gorm:"type:text"`                    // 审批意见
	SubmittedAt   time.Time `json:"submitted_at" gorm:"not null"`                 // 提交时间
	ApprovedAt    time.Time `json:"approved_at"`                                  // 审批时间
	RejectedAt    time.Time `json:"rejected_at"`                                  // 拒绝时间
	CancelledAt   time.Time `json:"cancelled_at"`                                 // 取消时间
	Deadline      time.Time `json:"deadline"`                                     // 截止时间
	Priority      string    `json:"priority" gorm:"size:20;default:'normal'"`     // 优先级: low, normal, high, urgent
	CreatedBy     string    `json:"created_by" gorm:"type:uuid;not null"`
	UpdatedBy     string    `json:"updated_by" gorm:"type:uuid"`
	CreatedAt     time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt     time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

// BeforeCreate 创建前钩子
func (a *ApprovalProcess) BeforeCreate(tx *gorm.DB) error {
	if a.ID == "" {
		a.ID = uuid.New().String()
	}
	return nil
}

// ApprovalHistory 审批历史模型
type ApprovalHistory struct {
	ID         string    `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	ApprovalID string    `json:"approval_id" gorm:"type:uuid;not null;index"` // 关联审批ID
	ActorID    string    `json:"actor_id" gorm:"type:uuid;not null"`          // 操作人ID
	ActorName  string    `json:"actor_name" gorm:"size:100"`                  // 操作人姓名
	Action     string    `json:"action" gorm:"size:20;not null"`              // 操作: submit, approve, reject, cancel, comment
	Comments   string    `json:"comments" gorm:"type:text"`                   // 操作意见
	ActionTime time.Time `json:"action_time" gorm:"not null"`                 // 操作时间
	CreatedAt  time.Time `json:"created_at" gorm:"autoCreateTime"`
}

// BeforeCreate 创建前钩子
func (h *ApprovalHistory) BeforeCreate(tx *gorm.DB) error {
	if h.ID == "" {
		h.ID = uuid.New().String()
	}
	return nil
}

// ApprovalTemplate 审批模板模型
type ApprovalTemplate struct {
	ID            string    `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	TeamID        string    `json:"team_id" gorm:"type:uuid;not null;index"`
	Name          string    `json:"name" gorm:"size:255;not null"`         // 模板名称
	Description   string    `json:"description" gorm:"type:text"`          // 描述
	ApprovalType  string    `json:"approval_type" gorm:"size:50;not null"` // 审批类型
	ApproverRoles []string  `json:"approver_roles" gorm:"type:text[]"`     // 审批人角色
	Steps         []string  `json:"steps" gorm:"type:text[]"`              // 审批步骤
	AutoApprove   bool      `json:"auto_approve" gorm:"default:false"`     // 是否自动审批
	Conditions    string    `json:"conditions" gorm:"type:text"`           // 自动审批条件
	CreatedBy     string    `json:"created_by" gorm:"type:uuid;not null"`
	UpdatedBy     string    `json:"updated_by" gorm:"type:uuid"`
	CreatedAt     time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt     time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

// BeforeCreate 创建前钩子
func (t *ApprovalTemplate) BeforeCreate(tx *gorm.DB) error {
	if t.ID == "" {
		t.ID = uuid.New().String()
	}
	return nil
}

// ApprovalNotification 审批通知模型
type ApprovalNotification struct {
	ID               string    `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	ApprovalID       string    `json:"approval_id" gorm:"type:uuid;not null;index"` // 关联审批ID
	UserID           string    `json:"user_id" gorm:"type:uuid;not null;index"`     // 接收用户ID
	NotificationType string    `json:"notification_type" gorm:"size:20;not null"`   // 通知类型: submit, approve, reject, reminder
	Title            string    `json:"title" gorm:"size:255;not null"`              // 通知标题
	Content          string    `json:"content" gorm:"type:text"`                    // 通知内容
	IsRead           bool      `json:"is_read" gorm:"default:false"`                // 是否已读
	ReadAt           time.Time `json:"read_at"`                                     // 阅读时间
	CreatedAt        time.Time `json:"created_at" gorm:"autoCreateTime"`
}

// BeforeCreate 创建前钩子
func (n *ApprovalNotification) BeforeCreate(tx *gorm.DB) error {
	if n.ID == "" {
		n.ID = uuid.New().String()
	}
	return nil
}
