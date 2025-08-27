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

// DocumentScanTask 文档扫描任务模型
type DocumentScanTask struct {
	ID             string    `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	UserID         string    `json:"user_id" gorm:"type:uuid;not null;index"`
	TeamID         string    `json:"team_id" gorm:"type:uuid;not null;index"`
	TaskName       string    `json:"task_name" gorm:"size:255;not null"`
	Description    string    `json:"description" gorm:"type:text"`
	ScanType       string    `json:"scan_type" gorm:"size:50;not null"`       // mobile_scan, batch_upload, auto_import
	SourcePath     string    `json:"source_path" gorm:"size:500"`             // 源文件路径或目录
	TargetFolder   string    `json:"target_folder" gorm:"size:255"`           // 目标归档文件夹
	ProcessConfig  string    `json:"process_config" gorm:"type:text"`         // JSON格式的处理配置
	Status         string    `json:"status" gorm:"size:20;default:'pending'"` // pending, processing, completed, failed, cancelled
	Progress       int       `json:"progress" gorm:"default:0"`               // 进度百分比 0-100
	TotalFiles     int       `json:"total_files" gorm:"default:0"`
	ProcessedFiles int       `json:"processed_files" gorm:"default:0"`
	SuccessFiles   int       `json:"success_files" gorm:"default:0"`
	FailedFiles    int       `json:"failed_files" gorm:"default:0"`
	ErrorMessage   string    `json:"error_message" gorm:"type:text"`
	StartedAt      time.Time `json:"started_at"`
	CompletedAt    time.Time `json:"completed_at"`
	CreatedAt      time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt      time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

// BeforeCreate 创建前钩子
func (d *DocumentScanTask) BeforeCreate(tx *gorm.DB) error {
	if d.ID == "" {
		d.ID = uuid.New().String()
	}
	return nil
}

// DocumentScanResult 文档扫描结果模型
type DocumentScanResult struct {
	ID                   string    `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	TaskID               string    `json:"task_id" gorm:"type:uuid;not null;index"`
	UserID               string    `json:"user_id" gorm:"type:uuid;not null;index"`
	TeamID               string    `json:"team_id" gorm:"type:uuid;not null;index"`
	OriginalFileName     string    `json:"original_file_name" gorm:"size:255;not null"`
	OriginalFilePath     string    `json:"original_file_path" gorm:"size:500;not null"`
	ProcessedFilePath    string    `json:"processed_file_path" gorm:"size:500"`
	FileType             string    `json:"file_type" gorm:"size:50"`
	FileSize             int64     `json:"file_size"`
	ScanMethod           string    `json:"scan_method" gorm:"size:50"` // camera, upload, import
	ImageEnhanced        bool      `json:"image_enhanced" gorm:"default:false"`
	PerspectiveCorrected bool      `json:"perspective_corrected" gorm:"default:false"`
	OCRPerformed         bool      `json:"ocr_performed" gorm:"default:false"`
	OCRText              string    `json:"ocr_text" gorm:"type:text"`
	OCRConfidence        float32   `json:"ocr_confidence" gorm:"default:0"`
	DocumentID           string    `json:"document_id" gorm:"type:uuid;index"`              // 关联到Document表
	KnowledgeID          string    `json:"knowledge_id" gorm:"type:uuid;index"`             // 关联到PersonalKnowledgeBase表
	ProcessStatus        string    `json:"process_status" gorm:"size:20;default:'pending'"` // pending, processing, completed, failed
	ProcessSteps         string    `json:"process_steps" gorm:"type:text"`                  // JSON格式的处理步骤记录
	QualityScore         float32   `json:"quality_score" gorm:"default:0"`                  // 文档质量评分 0-100
	AutoCategories       []string  `json:"auto_categories" gorm:"type:text[]"`              // AI自动分类标签
	ExtractedMetadata    string    `json:"extracted_metadata" gorm:"type:text"`             // JSON格式的提取元数据
	ErrorMessage         string    `json:"error_message" gorm:"type:text"`
	ProcessedAt          time.Time `json:"processed_at"`
	CreatedAt            time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt            time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

// BeforeCreate 创建前钩子
func (d *DocumentScanResult) BeforeCreate(tx *gorm.DB) error {
	if d.ID == "" {
		d.ID = uuid.New().String()
	}
	return nil
}

// ScanProcessConfig 扫描处理配置结构
type ScanProcessConfig struct {
	ImageEnhancement      bool    `json:"image_enhancement"`      // 是否启用图像增强
	PerspectiveCorrection bool    `json:"perspective_correction"` // 是否启用透视矫正
	OCREnabled            bool    `json:"ocr_enabled"`            // 是否启用OCR
	OCRLanguage           string  `json:"ocr_language"`           // OCR识别语言
	AutoClassification    bool    `json:"auto_classification"`    // 是否启用AI自动分类
	QualityThreshold      float32 `json:"quality_threshold"`      // 质量阈值
	AutoArchive           bool    `json:"auto_archive"`           // 是否自动归档到知识库
	NotifyOnComplete      bool    `json:"notify_on_complete"`     // 完成时是否通知
}

// ScanProcessStep 扫描处理步骤
type ScanProcessStep struct {
	StepName     string      `json:"step_name"`
	Status       string      `json:"status"` // pending, processing, completed, failed, skipped
	StartTime    time.Time   `json:"start_time"`
	EndTime      time.Time   `json:"end_time"`
	Duration     int64       `json:"duration"` // 毫秒
	ErrorMessage string      `json:"error_message,omitempty"`
	OutputData   interface{} `json:"output_data,omitempty"`
}

// DocumentQualityMetrics 文档质量指标
type DocumentQualityMetrics struct {
	Clarity         float32  `json:"clarity"`                   // 清晰度 0-100
	Brightness      float32  `json:"brightness"`                // 亮度评分 0-100
	Contrast        float32  `json:"contrast"`                  // 对比度评分 0-100
	TextReadability float32  `json:"text_readability"`          // 文本可读性 0-100
	OverallScore    float32  `json:"overall_score"`             // 综合评分 0-100
	Recommendations []string `json:"recommendations,omitempty"` // 改进建议
}

// ExtractedMetadata 提取的元数据
type ExtractedMetadata struct {
	Title        string        `json:"title,omitempty"`
	Author       string        `json:"author,omitempty"`
	Date         time.Time     `json:"date,omitempty"`
	Subject      string        `json:"subject,omitempty"`
	Keywords     []string      `json:"keywords,omitempty"`
	Language     string        `json:"language,omitempty"`
	DocumentType string        `json:"document_type,omitempty"`
	PageCount    int           `json:"page_count,omitempty"`
	WordCount    int           `json:"word_count,omitempty"`
	CharCount    int           `json:"char_count,omitempty"`
	Entities     []NamedEntity `json:"entities,omitempty"` // 命名实体
}

// NamedEntity 命名实体
type NamedEntity struct {
	Text       string  `json:"text"`
	Type       string  `json:"type"` // PERSON, ORG, DATE, MONEY, etc.
	Confidence float32 `json:"confidence"`
	StartPos   int     `json:"start_pos"`
	EndPos     int     `json:"end_pos"`
}

// DocumentScanTemplate 文档扫描模板模型
type DocumentScanTemplate struct {
	ID            string    `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	UserID        string    `json:"user_id" gorm:"type:uuid;not null;index"`
	TeamID        string    `json:"team_id" gorm:"type:uuid;not null;index"`
	TemplateName  string    `json:"template_name" gorm:"size:255;not null"`
	Description   string    `json:"description" gorm:"type:text"`
	DocumentTypes []string  `json:"document_types" gorm:"type:text[]"` // 适用文档类型
	ProcessConfig string    `json:"process_config" gorm:"type:text"`   // JSON格式的处理配置
	IsDefault     bool      `json:"is_default" gorm:"default:false"`
	IsShared      bool      `json:"is_shared" gorm:"default:false"` // 是否共享给团队
	UsageCount    int       `json:"usage_count" gorm:"default:0"`   // 使用次数
	CreatedBy     string    `json:"created_by" gorm:"type:uuid;not null"`
	CreatedAt     time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt     time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

// BeforeCreate 创建前钩子
func (d *DocumentScanTemplate) BeforeCreate(tx *gorm.DB) error {
	if d.ID == "" {
		d.ID = uuid.New().String()
	}
	return nil
}
