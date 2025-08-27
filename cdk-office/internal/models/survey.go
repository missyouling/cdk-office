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
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// Survey 问卷模型
type Survey struct {
	ID              string         `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	SurveyID        string         `json:"survey_id" gorm:"size:100;uniqueIndex;not null"` // 问卷唯一标识
	Title           string         `json:"title" gorm:"size:255;not null"`                  // 问卷标题
	Description     string         `json:"description" gorm:"type:text"`                    // 问卷描述
	JsonDefinition  datatypes.JSON `json:"json_definition" gorm:"type:jsonb"`               // SurveyJS JSON定义
	CreatedBy       string         `json:"created_by" gorm:"type:uuid;not null;index"`     // 创建者ID
	TeamID          string         `json:"team_id" gorm:"type:uuid;not null;index"`        // 团队ID
	Status          string         `json:"status" gorm:"size:50;default:'draft'"`          // 状态: draft, active, closed, archived
	IsPublic        bool           `json:"is_public" gorm:"default:false"`                 // 是否公开
	MaxResponses    int            `json:"max_responses" gorm:"default:0"`                 // 最大响应数量，0为无限制
	StartTime       *time.Time     `json:"start_time"`                                     // 开始时间
	EndTime         *time.Time     `json:"end_time"`                                       // 结束时间
	Tags            string         `json:"tags" gorm:"type:text"`                          // 标签，逗号分隔
	ResponseCount   int            `json:"response_count" gorm:"default:0"`                // 响应数量
	ViewCount       int            `json:"view_count" gorm:"default:0"`                    // 浏览次数
	ShareURL        string         `json:"share_url" gorm:"size:500"`                      // 分享链接
	CreatedAt       time.Time      `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt       time.Time      `json:"updated_at" gorm:"autoUpdateTime"`
}

// BeforeCreate 创建前钩子
func (s *Survey) BeforeCreate(tx *gorm.DB) error {
	if s.ID == "" {
		s.ID = uuid.New().String()
	}
	if s.SurveyID == "" {
		s.SurveyID = "survey_" + uuid.New().String()[:8]
	}
	return nil
}

// SurveyResponse 问卷响应模型
type SurveyResponse struct {
	ID           string         `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	SurveyID     string         `json:"survey_id" gorm:"size:100;not null;index"`           // 问卷ID
	UserID       string         `json:"user_id" gorm:"type:uuid;index"`                     // 用户ID，可为空（匿名用户）
	TeamID       string         `json:"team_id" gorm:"type:uuid;not null;index"`            // 团队ID
	ResponseData datatypes.JSON `json:"response_data" gorm:"type:jsonb"`                    // 响应数据
	TimeSpent    int            `json:"time_spent" gorm:"default:0"`                        // 用时（秒）
	IPAddress    string         `json:"ip_address" gorm:"size:45"`                          // IP地址
	UserAgent    string         `json:"user_agent" gorm:"size:500"`                         // 用户代理
	IsCompleted  bool           `json:"is_completed" gorm:"default:true"`                   // 是否完成
	CompletedAt  time.Time      `json:"completed_at"`                                       // 完成时间
	CreatedAt    time.Time      `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt    time.Time      `json:"updated_at" gorm:"autoUpdateTime"`
}

// BeforeCreate 创建前钩子
func (sr *SurveyResponse) BeforeCreate(tx *gorm.DB) error {
	if sr.ID == "" {
		sr.ID = uuid.New().String()
	}
	if sr.IsCompleted && sr.CompletedAt.IsZero() {
		sr.CompletedAt = time.Now()
	}
	return nil
}

// SurveyAnalysis 问卷分析模型
type SurveyAnalysis struct {
	ID             string         `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	SurveyID       string         `json:"survey_id" gorm:"size:100;not null;index"`     // 问卷ID
	AnalysisType   string         `json:"analysis_type" gorm:"size:50;not null"`        // 分析类型: basic, ai, custom
	ResultData     datatypes.JSON `json:"result_data" gorm:"type:jsonb"`                // 分析结果数据
	DifyWorkflowID string         `json:"dify_workflow_id" gorm:"size:100"`             // Dify工作流ID
	RunID          string         `json:"run_id" gorm:"size:100"`                       // Dify运行ID
	Status         string         `json:"status" gorm:"size:50;default:'pending'"`     // 状态: pending, running, completed, failed
	ErrorMessage   string         `json:"error_message" gorm:"type:text"`               // 错误信息
	CreatedAt      time.Time      `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt      time.Time      `json:"updated_at" gorm:"autoUpdateTime"`
}

// BeforeCreate 创建前钩子
func (sa *SurveyAnalysis) BeforeCreate(tx *gorm.DB) error {
	if sa.ID == "" {
		sa.ID = uuid.New().String()
	}
	return nil
}

// SurveyPermission 问卷权限模型
type SurveyPermission struct {
	ID        string    `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	SurveyID  string    `json:"survey_id" gorm:"size:100;not null;index"`        // 问卷ID
	UserID    string    `json:"user_id" gorm:"type:uuid;not null;index"`         // 用户ID
	TeamID    string    `json:"team_id" gorm:"type:uuid;not null;index"`         // 团队ID
	CanView   bool      `json:"can_view" gorm:"default:true"`                    // 可查看
	CanEdit   bool      `json:"can_edit" gorm:"default:false"`                   // 可编辑
	CanDelete bool      `json:"can_delete" gorm:"default:false"`                 // 可删除
	CanManage bool      `json:"can_manage" gorm:"default:false"`                 // 可管理（包括分享、权限设置等）
	CanAnalyze bool     `json:"can_analyze" gorm:"default:false"`                // 可分析
	CanExport bool      `json:"can_export" gorm:"default:false"`                 // 可导出
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

// BeforeCreate 创建前钩子
func (sp *SurveyPermission) BeforeCreate(tx *gorm.DB) error {
	if sp.ID == "" {
		sp.ID = uuid.New().String()
	}
	return nil
}

// SurveyTemplate 问卷模板模型
type SurveyTemplate struct {
	ID           string         `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	Name         string         `json:"name" gorm:"size:255;not null"`                      // 模板名称
	Description  string         `json:"description" gorm:"type:text"`                       // 模板描述
	Category     string         `json:"category" gorm:"size:100"`                           // 模板分类
	JsonTemplate datatypes.JSON `json:"json_template" gorm:"type:jsonb"`                    // 模板JSON
	PreviewImage string         `json:"preview_image" gorm:"size:500"`                      // 预览图片
	IsPublic     bool           `json:"is_public" gorm:"default:false"`                     // 是否公开
	UseCount     int            `json:"use_count" gorm:"default:0"`                         // 使用次数
	Rating       float32        `json:"rating" gorm:"default:0"`                            // 评分
	CreatedBy    string         `json:"created_by" gorm:"type:uuid;not null;index"`        // 创建者ID
	TeamID       string         `json:"team_id" gorm:"type:uuid;index"`                     // 团队ID，为空表示全局模板
	Tags         string         `json:"tags" gorm:"type:text"`                              // 标签
	CreatedAt    time.Time      `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt    time.Time      `json:"updated_at" gorm:"autoUpdateTime"`
}

// BeforeCreate 创建前钩子
func (st *SurveyTemplate) BeforeCreate(tx *gorm.DB) error {
	if st.ID == "" {
		st.ID = uuid.New().String()
	}
	return nil
}

// SurveyFile 问卷文件模型（用于上传的文件）
type SurveyFile struct {
	ID          string    `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	SurveyID    string    `json:"survey_id" gorm:"size:100;not null;index"`       // 问卷ID
	FileName    string    `json:"file_name" gorm:"size:255;not null"`             // 文件名
	FilePath    string    `json:"file_path" gorm:"size:500;not null"`             // 文件路径
	FileSize    int64     `json:"file_size"`                                      // 文件大小
	MimeType    string    `json:"mime_type" gorm:"size:100"`                      // MIME类型
	StorageType string    `json:"storage_type" gorm:"size:50"`                    // 存储类型: local, cloud_db, s3, webdav
	FileURL     string    `json:"file_url" gorm:"size:500"`                       // 文件访问URL
	UploadedBy  string    `json:"uploaded_by" gorm:"type:uuid;not null;index"`   // 上传者ID
	TeamID      string    `json:"team_id" gorm:"type:uuid;not null;index"`        // 团队ID
	CreatedAt   time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt   time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

// BeforeCreate 创建前钩子
func (sf *SurveyFile) BeforeCreate(tx *gorm.DB) error {
	if sf.ID == "" {
		sf.ID = uuid.New().String()
	}
	return nil
}