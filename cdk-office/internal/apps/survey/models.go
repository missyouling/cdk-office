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

package survey

import (
	"time"

	"gorm.io/datatypes"
)

// CreateSurveyRequest 创建问卷请求
type CreateSurveyRequest struct {
	Title           string         `json:"title" binding:"required,min=1,max=255"`
	Description     string         `json:"description"`
	JsonDefinition  datatypes.JSON `json:"json_definition" binding:"required"`
	IsPublic        bool           `json:"is_public"`
	MaxResponses    int            `json:"max_responses"`
	StartTime       *time.Time     `json:"start_time"`
	EndTime         *time.Time     `json:"end_time"`
	Tags            string         `json:"tags"`
}

// UpdateSurveyRequest 更新问卷请求
type UpdateSurveyRequest struct {
	Title           *string        `json:"title,omitempty"`
	Description     *string        `json:"description,omitempty"`
	JsonDefinition  datatypes.JSON `json:"json_definition,omitempty"`
	IsPublic        *bool          `json:"is_public,omitempty"`
	MaxResponses    *int           `json:"max_responses,omitempty"`
	StartTime       *time.Time     `json:"start_time,omitempty"`
	EndTime         *time.Time     `json:"end_time,omitempty"`
	Tags            *string        `json:"tags,omitempty"`
}

// SubmitResponseRequest 提交响应请求
type SubmitResponseRequest struct {
	ResponseData datatypes.JSON `json:"response_data" binding:"required"`
	TimeSpent    int            `json:"time_spent"`
	UserAgent    string         `json:"user_agent"`
}

// CreateTemplateRequest 创建模板请求
type CreateTemplateRequest struct {
	Name         string         `json:"name" binding:"required,min=1,max=255"`
	Description  string         `json:"description"`
	Category     string         `json:"category"`
	JsonTemplate datatypes.JSON `json:"json_template" binding:"required"`
	PreviewImage string         `json:"preview_image"`
	IsPublic     bool           `json:"is_public"`
	Tags         string         `json:"tags"`
}

// UpdateTemplateRequest 更新模板请求
type UpdateTemplateRequest struct {
	Name         *string        `json:"name,omitempty"`
	Description  *string        `json:"description,omitempty"`
	Category     *string        `json:"category,omitempty"`
	JsonTemplate datatypes.JSON `json:"json_template,omitempty"`
	PreviewImage *string        `json:"preview_image,omitempty"`
	IsPublic     *bool          `json:"is_public,omitempty"`
	Tags         *string        `json:"tags,omitempty"`
}

// SetPermissionRequest 设置权限请求
type SetPermissionRequest struct {
	UserID     string `json:"user_id" binding:"required"`
	CanView    bool   `json:"can_view"`
	CanEdit    bool   `json:"can_edit"`
	CanDelete  bool   `json:"can_delete"`
	CanManage  bool   `json:"can_manage"`
	CanAnalyze bool   `json:"can_analyze"`
	CanExport  bool   `json:"can_export"`
}

// TriggerAnalysisRequest 触发分析请求
type TriggerAnalysisRequest struct {
	AnalysisType   string `json:"analysis_type" binding:"required"` // basic, ai, custom
	DifyWorkflowID string `json:"dify_workflow_id,omitempty"`
}

// SurveyResponse 问卷响应
type SurveyResponse struct {
	ID           string         `json:"id"`
	SurveyID     string         `json:"survey_id"`
	Title        string         `json:"title"`
	Description  string         `json:"description"`
	Status       string         `json:"status"`
	IsPublic     bool           `json:"is_public"`
	ResponseCount int           `json:"response_count"`
	ViewCount    int            `json:"view_count"`
	Tags         string         `json:"tags"`
	CreatedBy    string         `json:"created_by"`
	TeamID       string         `json:"team_id"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	
	// 可选字段
	JsonDefinition *datatypes.JSON `json:"json_definition,omitempty"`
	MaxResponses   *int            `json:"max_responses,omitempty"`
	StartTime      *time.Time      `json:"start_time,omitempty"`
	EndTime        *time.Time      `json:"end_time,omitempty"`
	ShareURL       *string         `json:"share_url,omitempty"`
}

// ResponseResponse 响应数据响应
type ResponseResponse struct {
	ID           string         `json:"id"`
	SurveyID     string         `json:"survey_id"`
	UserID       string         `json:"user_id,omitempty"`
	ResponseData datatypes.JSON `json:"response_data"`
	TimeSpent    int            `json:"time_spent"`
	IsCompleted  bool           `json:"is_completed"`
	CompletedAt  time.Time      `json:"completed_at"`
	CreatedAt    time.Time      `json:"created_at"`
}

// AnalysisResponse 分析结果响应
type AnalysisResponse struct {
	ID             string         `json:"id"`
	SurveyID       string         `json:"survey_id"`
	AnalysisType   string         `json:"analysis_type"`
	ResultData     datatypes.JSON `json:"result_data"`
	DifyWorkflowID string         `json:"dify_workflow_id,omitempty"`
	RunID          string         `json:"run_id,omitempty"`
	Status         string         `json:"status"`
	ErrorMessage   string         `json:"error_message,omitempty"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
}

// TemplateResponse 模板响应
type TemplateResponse struct {
	ID           string         `json:"id"`
	Name         string         `json:"name"`
	Description  string         `json:"description"`
	Category     string         `json:"category"`
	PreviewImage string         `json:"preview_image"`
	IsPublic     bool           `json:"is_public"`
	UseCount     int            `json:"use_count"`
	Rating       float32        `json:"rating"`
	Tags         string         `json:"tags"`
	CreatedBy    string         `json:"created_by"`
	TeamID       string         `json:"team_id,omitempty"`
	CreatedAt    time.Time      `json:"created_at"`
	
	// 可选字段
	JsonTemplate *datatypes.JSON `json:"json_template,omitempty"`
}

// PermissionResponse 权限响应
type PermissionResponse struct {
	ID         string    `json:"id"`
	SurveyID   string    `json:"survey_id"`
	UserID     string    `json:"user_id"`
	UserName   string    `json:"user_name,omitempty"`   // 用户名（关联查询）
	UserEmail  string    `json:"user_email,omitempty"`  // 用户邮箱（关联查询）
	CanView    bool      `json:"can_view"`
	CanEdit    bool      `json:"can_edit"`
	CanDelete  bool      `json:"can_delete"`
	CanManage  bool      `json:"can_manage"`
	CanAnalyze bool      `json:"can_analyze"`
	CanExport  bool      `json:"can_export"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// FileResponse 文件响应
type FileResponse struct {
	ID          string    `json:"id"`
	SurveyID    string    `json:"survey_id"`
	FileName    string    `json:"file_name"`
	FileSize    int64     `json:"file_size"`
	MimeType    string    `json:"mime_type"`
	StorageType string    `json:"storage_type"`
	FileURL     string    `json:"file_url"`
	UploadedBy  string    `json:"uploaded_by"`
	CreatedAt   time.Time `json:"created_at"`
}

// ExportRequest 导出请求
type ExportRequest struct {
	Format      string   `json:"format" binding:"required"` // csv, xlsx, json
	Fields      []string `json:"fields,omitempty"`          // 要导出的字段
	FilterBy    string   `json:"filter_by,omitempty"`       // 过滤条件
	DateFrom    string   `json:"date_from,omitempty"`       // 开始日期
	DateTo      string   `json:"date_to,omitempty"`         // 结束日期
	IncludeUser bool     `json:"include_user"`              // 是否包含用户信息
}

// StatisticsResponse 统计信息响应
type StatisticsResponse struct {
	TotalSurveys     int                    `json:"total_surveys"`
	ActiveSurveys    int                    `json:"active_surveys"`
	TotalResponses   int                    `json:"total_responses"`
	ResponseRate     float64                `json:"response_rate"`
	AvgCompletionTime int                   `json:"avg_completion_time"`
	PopularTags      []string               `json:"popular_tags"`
	RecentActivity   []ActivityResponse     `json:"recent_activity"`
	SurveyStats      []SurveyStatsResponse  `json:"survey_stats"`
}

// ActivityResponse 活动响应
type ActivityResponse struct {
	Type        string    `json:"type"`        // created, published, responded, etc.
	SurveyID    string    `json:"survey_id"`
	SurveyTitle string    `json:"survey_title"`
	UserID      string    `json:"user_id,omitempty"`
	UserName    string    `json:"user_name,omitempty"`
	Timestamp   time.Time `json:"timestamp"`
}

// SurveyStatsResponse 问卷统计响应
type SurveyStatsResponse struct {
	SurveyID      string  `json:"survey_id"`
	Title         string  `json:"title"`
	ResponseCount int     `json:"response_count"`
	ViewCount     int     `json:"view_count"`
	ResponseRate  float64 `json:"response_rate"`
	AvgTimeSpent  int     `json:"avg_time_spent"`
	Status        string  `json:"status"`
}

// ListResponse 通用列表响应
type ListResponse struct {
	Data  interface{} `json:"data"`
	Total int64       `json:"total"`
	Page  int         `json:"page"`
	Pages int64       `json:"pages"`
}

// ErrorResponse 错误响应
type ErrorResponse struct {
	Error   string `json:"error"`
	Details string `json:"details,omitempty"`
	Code    string `json:"code,omitempty"`
}

// SuccessResponse 成功响应
type SuccessResponse struct {
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}