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

// Contract 合同模型
type Contract struct {
	ID                string              `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	TeamID            string              `json:"team_id" gorm:"type:uuid;not null;index"`
	Title             string              `json:"title" gorm:"size:255;not null"`
	Description       string              `json:"description" gorm:"type:text"`
	TemplateID        string              `json:"template_id" gorm:"type:uuid;index"`
	
	// 合同内容
	Content           string              `json:"content" gorm:"type:text"`
	OriginalFileURL   string              `json:"original_file_url" gorm:"size:500"`  // 原始文件URL
	FinalFileURL      string              `json:"final_file_url" gorm:"size:500"`     // 签署完成后的文件URL
	FileHash          string              `json:"file_hash" gorm:"size:128"`          // 文件哈希值
	
	// 状态管理
	Status            string              `json:"status" gorm:"size:20;default:'draft'"` // draft, pending, signing, completed, rejected, cancelled, expired
	Progress          int                 `json:"progress" gorm:"default:0"`              // 签署进度百分比
	
	// 签署配置
	SignMode          string              `json:"sign_mode" gorm:"size:20;default:'sequential'"` // sequential(顺序), parallel(并行)
	RequireCA         bool                `json:"require_ca" gorm:"default:true"`                 // 是否需要CA证书
	RequireBlockchain bool                `json:"require_blockchain" gorm:"default:false"`       // 是否需要区块链存证
	
	// 时间管理
	StartTime         time.Time           `json:"start_time"`                        // 签署开始时间
	ExpireTime        time.Time           `json:"expire_time"`                       // 合同过期时间
	CompletedAt       *time.Time          `json:"completed_at"`                      // 签署完成时间
	
	// 区块链存证信息
	BlockchainTxHash  string              `json:"blockchain_tx_hash" gorm:"size:128"` // 区块链交易哈希
	EvidenceURL       string              `json:"evidence_url" gorm:"size:500"`       // 存证报告URL
	
	// 关联关系
	Signers           []ContractSigner    `json:"signers" gorm:"foreignKey:ContractID"`
	Logs              []ContractLog       `json:"logs" gorm:"foreignKey:ContractID"`
	
	// 审计字段
	CreatedBy         string              `json:"created_by" gorm:"type:uuid;not null"`
	UpdatedBy         string              `json:"updated_by" gorm:"type:uuid"`
	CreatedAt         time.Time           `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt         time.Time           `json:"updated_at" gorm:"autoUpdateTime"`
}

// BeforeCreate 创建前钩子
func (c *Contract) BeforeCreate(tx *gorm.DB) error {
	if c.ID == "" {
		c.ID = uuid.New().String()
	}
	return nil
}

// ContractSigner 合同签署人模型
type ContractSigner struct {
	ID             string     `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	ContractID     string     `json:"contract_id" gorm:"type:uuid;not null;index"`
	
	// 签署人信息
	SignerType     string     `json:"signer_type" gorm:"size:20;not null"` // person(个人), company(企业)
	Name           string     `json:"name" gorm:"size:100;not null"`
	Email          string     `json:"email" gorm:"size:100"`
	Phone          string     `json:"phone" gorm:"size:20"`
	IdCard         string     `json:"id_card" gorm:"size:30"`              // 身份证号(个人)
	CompanyName    string     `json:"company_name" gorm:"size:200"`        // 企业名称
	UnifiedCode    string     `json:"unified_code" gorm:"size:50"`         // 统一社会信用代码(企业)
	
	// 签署配置
	SignOrder      int        `json:"sign_order" gorm:"not null"`          // 签署顺序
	SignPosition   string     `json:"sign_position" gorm:"type:text"`      // 签署位置信息(JSON格式)
	SignType       string     `json:"sign_type" gorm:"size:20;default:'signature'"` // signature(签名), seal(印章)
	
	// 签署状态
	Status         string     `json:"status" gorm:"size:20;default:'pending'"` // pending, signed, rejected
	SignTime       *time.Time `json:"sign_time"`                               // 签署时间
	SignIP         string     `json:"sign_ip" gorm:"size:45"`                  // 签署IP
	SignLocation   string     `json:"sign_location" gorm:"size:200"`           // 签署地点
	
	// 认证信息
	CertificateID  string     `json:"certificate_id" gorm:"size:100"`  // CA证书ID
	SignatureImage string     `json:"signature_image" gorm:"size:500"` // 签名图片URL
	
	// 审计字段
	CreatedAt      time.Time  `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt      time.Time  `json:"updated_at" gorm:"autoUpdateTime"`
}

// ContractTemplate 合同模板模型
type ContractTemplate struct {
	ID             string                   `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	TeamID         string                   `json:"team_id" gorm:"type:uuid;index"`        // 空值表示公共模板
	Name           string                   `json:"name" gorm:"size:255;not null"`
	Description    string                   `json:"description" gorm:"type:text"`
	Category       string                   `json:"category" gorm:"size:50"`               // 模板分类
	
	// 模板内容
	Content        string                   `json:"content" gorm:"type:text;not null"`     // 模板内容(HTML)
	Fields         string                   `json:"fields" gorm:"type:text"`               // 可变字段定义(JSON)
	SignPositions  string                   `json:"sign_positions" gorm:"type:text"`       // 签署位置定义(JSON)
	
	// 模板配置
	IsPublic       bool                     `json:"is_public" gorm:"default:false"`        // 是否为公共模板
	IsActive       bool                     `json:"is_active" gorm:"default:true"`
	Version        int                      `json:"version" gorm:"default:1"`
	DownloadCount  int                      `json:"download_count" gorm:"default:0"`
	
	// 关联关系
	Contracts      []Contract               `json:"contracts" gorm:"foreignKey:TemplateID"`
	
	// 审计字段
	CreatedBy      string                   `json:"created_by" gorm:"type:uuid"`
	UpdatedBy      string                   `json:"updated_by" gorm:"type:uuid"`
	CreatedAt      time.Time                `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt      time.Time                `json:"updated_at" gorm:"autoUpdateTime"`
}

// ContractLog 合同操作日志模型
type ContractLog struct {
	ID           string    `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	ContractID   string    `json:"contract_id" gorm:"type:uuid;not null;index"`
	
	// 操作信息
	Action       string    `json:"action" gorm:"size:50;not null"`       // create, send, sign, reject, cancel, complete等
	Description  string    `json:"description" gorm:"size:500"`
	OperatorID   string    `json:"operator_id" gorm:"type:uuid"`
	OperatorName string    `json:"operator_name" gorm:"size:100"`
	
	// 操作详情
	Details      string    `json:"details" gorm:"type:text"`             // 详细信息(JSON格式)
	IPAddress    string    `json:"ip_address" gorm:"size:45"`
	UserAgent    string    `json:"user_agent" gorm:"size:500"`
	
	// 审计字段
	CreatedAt    time.Time `json:"created_at" gorm:"autoCreateTime"`
}

// ContractFile 合同文件模型
type ContractFile struct {
	ID         string    `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	ContractID string    `json:"contract_id" gorm:"type:uuid;not null;index"`
	
	// 文件信息
	FileName   string    `json:"file_name" gorm:"size:255;not null"`
	FileURL    string    `json:"file_url" gorm:"size:500;not null"`
	FileSize   int64     `json:"file_size"`
	FileType   string    `json:"file_type" gorm:"size:50"`        // original, signed, evidence
	MimeType   string    `json:"mime_type" gorm:"size:100"`
	FileHash   string    `json:"file_hash" gorm:"size:128"`
	
	// 存储信息
	StorageType string   `json:"storage_type" gorm:"size:20"`     // local, oss, cos等
	StoragePath string   `json:"storage_path" gorm:"size:500"`
	
	// 审计字段
	CreatedAt   time.Time `json:"created_at" gorm:"autoCreateTime"`
}

// KnowledgeSubmission 知识库提交记录模型
type KnowledgeSubmission struct {
	ID                 string     `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	ContractID         string     `json:"contract_id" gorm:"type:uuid;not null;index"`
	DocumentID         string     `json:"document_id" gorm:"type:uuid"`                    // 关联的文档ID
	
	// 提交信息
	SubmissionType     string     `json:"submission_type" gorm:"size:20"`                  // auto, manual
	Status             string     `json:"status" gorm:"size:20;default:'pending'"`        // pending, approved, rejected, completed
	AutoProcessing     bool       `json:"auto_processing" gorm:"default:true"`             // 是否自动处理
	
	// 审批信息
	ApprovalRequired   bool       `json:"approval_required" gorm:"default:false"`
	ApprovalStatus     string     `json:"approval_status" gorm:"size:20"`                  // pending, approved, rejected
	ApprovalComments   string     `json:"approval_comments" gorm:"type:text"`
	ApprovalBy         string     `json:"approval_by" gorm:"type:uuid"`
	ApprovalAt         *time.Time `json:"approval_at"`
	
	// Dify处理状态
	DifyWorkflowStatus string     `json:"dify_workflow_status" gorm:"size:20"`             // pending, processing, completed, failed
	DifyJobID          string     `json:"dify_job_id" gorm:"size:100"`
	DifyResult         string     `json:"dify_result" gorm:"type:text"`                    // 处理结果(JSON)
	
	// 审计字段
	CreatedBy          string     `json:"created_by" gorm:"type:uuid;not null"`
	CreatedAt          time.Time  `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt          time.Time  `json:"updated_at" gorm:"autoUpdateTime"`
}