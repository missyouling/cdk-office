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

package knowledge

import "github.com/linux-do/cdk-office/internal/models"

// CreateKnowledgeRequest 创建知识请求
type CreateKnowledgeRequest struct {
	UserID      string                 `json:"user_id" binding:"required"`
	Title       string                 `json:"title" binding:"required,max=255"`
	Description string                 `json:"description"`
	Content     string                 `json:"content" binding:"required"`
	ContentType string                 `json:"content_type"` // markdown, text, html
	Tags        []string               `json:"tags"`
	Category    string                 `json:"category"`
	Privacy     string                 `json:"privacy"`     // private, shared, public
	SourceType  string                 `json:"source_type"` // manual, wechat, upload, scan
	SourceData  map[string]interface{} `json:"source_data,omitempty"`
}

// UpdateKnowledgeRequest 更新知识请求
type UpdateKnowledgeRequest struct {
	Title       string   `json:"title,omitempty" binding:"max=255"`
	Description string   `json:"description,omitempty"`
	Content     string   `json:"content,omitempty"`
	ContentType string   `json:"content_type,omitempty"`
	Tags        []string `json:"tags,omitempty"`
	Category    string   `json:"category,omitempty"`
	Privacy     string   `json:"privacy,omitempty"`
}

// ListKnowledgeRequest 列出知识请求
type ListKnowledgeRequest struct {
	UserID     string   `json:"user_id" binding:"required"`
	Page       int      `json:"page" binding:"min=1"`
	PageSize   int      `json:"page_size" binding:"min=1,max=100"`
	Category   string   `json:"category,omitempty"`
	Privacy    string   `json:"privacy,omitempty"`
	SourceType string   `json:"source_type,omitempty"`
	Keyword    string   `json:"keyword,omitempty"`
	Tags       []string `json:"tags,omitempty"`
	SortBy     string   `json:"sort_by,omitempty"` // created_at, updated_at, title
}

// ListKnowledgeResponse 列出知识响应
type ListKnowledgeResponse struct {
	Knowledge []models.PersonalKnowledgeBase `json:"knowledge"`
	Total     int64                          `json:"total"`
	Page      int                            `json:"page"`
	PageSize  int                            `json:"page_size"`
}

// ShareToTeamRequest 分享到团队请求
type ShareToTeamRequest struct {
	KnowledgeID string `json:"knowledge_id" binding:"required"`
	UserID      string `json:"user_id" binding:"required"`
	TeamID      string `json:"team_id" binding:"required"`
	ShareReason string `json:"share_reason" binding:"required"`
}

// SearchKnowledgeRequest 搜索知识请求
type SearchKnowledgeRequest struct {
	UserID     string   `json:"user_id" binding:"required"`
	Query      string   `json:"query" binding:"required"`
	Tags       []string `json:"tags,omitempty"`
	Category   string   `json:"category,omitempty"`
	SourceType string   `json:"source_type,omitempty"`
	Page       int      `json:"page" binding:"min=1"`
	PageSize   int      `json:"page_size" binding:"min=1,max=100"`
}

// SearchKnowledgeResponse 搜索知识响应
type SearchKnowledgeResponse struct {
	Results  []models.PersonalKnowledgeBase `json:"results"`
	Total    int64                          `json:"total"`
	Page     int                            `json:"page"`
	PageSize int                            `json:"page_size"`
	Query    string                         `json:"query"`
}

// KnowledgeStatistics 知识库统计信息
type KnowledgeStatistics struct {
	TotalKnowledge  int64          `json:"total_knowledge"`
	SharedKnowledge int64          `json:"shared_knowledge"`
	WeeklyAdded     int64          `json:"weekly_added"`
	ByCategory      []CategoryStat `json:"by_category"`
	BySource        []SourceStat   `json:"by_source"`
}

// CategoryStat 分类统计
type CategoryStat struct {
	Category string `json:"category"`
	Count    int64  `json:"count"`
}

// SourceStat 来源统计
type SourceStat struct {
	SourceType string `json:"source_type"`
	Count      int64  `json:"count"`
}

// TagStat 标签统计
type TagStat struct {
	Tag   string `json:"tag"`
	Count int64  `json:"count"`
}

// WeChatUploadRequest 微信聊天记录上传请求
type WeChatUploadRequest struct {
	UserID        string               `json:"user_id" binding:"required"`
	SessionName   string               `json:"session_name" binding:"required"`
	Records       []WeChatRecordData   `json:"records" binding:"required"`
	ProcessConfig *WeChatProcessConfig `json:"process_config,omitempty"`
}

// WeChatRecordData 微信聊天记录数据
type WeChatRecordData struct {
	MessageID   string                 `json:"message_id"`
	MessageType string                 `json:"message_type" binding:"required"` // text, image, voice, video, file
	SenderName  string                 `json:"sender_name"`
	SenderID    string                 `json:"sender_id"`
	Content     string                 `json:"content"`
	MessageTime string                 `json:"message_time"`
	FileData    string                 `json:"file_data,omitempty"` // base64编码的文件数据
	FileName    string                 `json:"file_name,omitempty"`
	ExtraData   map[string]interface{} `json:"extra_data,omitempty"`
}

// WeChatProcessConfig 微信聊天记录处理配置
type WeChatProcessConfig struct {
	EnableOCR          bool     `json:"enable_ocr"`           // 是否对图片进行OCR
	EnableAutoArchive  bool     `json:"enable_auto_archive"`  // 是否自动归档到知识库
	FilterMessageTypes []string `json:"filter_message_types"` // 过滤的消息类型
	ExtractKeywords    bool     `json:"extract_keywords"`     // 是否提取关键词
	GroupBySession     bool     `json:"group_by_session"`     // 是否按会话分组
}

// WeChatUploadResponse 微信聊天记录上传响应
type WeChatUploadResponse struct {
	ProcessedCount int                   `json:"processed_count"`
	FailedCount    int                   `json:"failed_count"`
	Records        []models.WeChatRecord `json:"records"`
	Errors         []WeChatProcessError  `json:"errors,omitempty"`
}

// WeChatProcessError 微信处理错误
type WeChatProcessError struct {
	MessageID string `json:"message_id"`
	Error     string `json:"error"`
}

// BatchDeleteRequest 批量删除请求
type BatchDeleteRequest struct {
	UserID       string   `json:"user_id" binding:"required"`
	KnowledgeIDs []string `json:"knowledge_ids" binding:"required"`
}

// BatchDeleteResponse 批量删除响应
type BatchDeleteResponse struct {
	SuccessCount int      `json:"success_count"`
	FailedCount  int      `json:"failed_count"`
	FailedIDs    []string `json:"failed_ids,omitempty"`
}

// BatchUpdateRequest 批量更新请求
type BatchUpdateRequest struct {
	UserID       string                 `json:"user_id" binding:"required"`
	KnowledgeIDs []string               `json:"knowledge_ids" binding:"required"`
	Updates      map[string]interface{} `json:"updates" binding:"required"`
}

// BatchUpdateResponse 批量更新响应
type BatchUpdateResponse struct {
	SuccessCount int      `json:"success_count"`
	FailedCount  int      `json:"failed_count"`
	FailedIDs    []string `json:"failed_ids,omitempty"`
}

// ImportKnowledgeRequest 导入知识请求
type ImportKnowledgeRequest struct {
	UserID  string        `json:"user_id" binding:"required"`
	Format  string        `json:"format" binding:"required"` // json, csv, markdown
	Data    string        `json:"data" binding:"required"`   // 导入数据
	Options ImportOptions `json:"options,omitempty"`
}

// ImportOptions 导入选项
type ImportOptions struct {
	SkipDuplicates  bool   `json:"skip_duplicates"`
	DefaultCategory string `json:"default_category"`
	DefaultPrivacy  string `json:"default_privacy"`
	TagPrefix       string `json:"tag_prefix"`
}

// ImportKnowledgeResponse 导入知识响应
type ImportKnowledgeResponse struct {
	ImportedCount int      `json:"imported_count"`
	SkippedCount  int      `json:"skipped_count"`
	FailedCount   int      `json:"failed_count"`
	Errors        []string `json:"errors,omitempty"`
}

// ExportKnowledgeRequest 导出知识请求
type ExportKnowledgeRequest struct {
	UserID         string        `json:"user_id" binding:"required"`
	Format         string        `json:"format" binding:"required"` // json, csv, markdown, pdf
	KnowledgeIDs   []string      `json:"knowledge_ids,omitempty"`   // 为空则导出全部
	IncludePrivate bool          `json:"include_private"`
	Options        ExportOptions `json:"options,omitempty"`
}

// ExportOptions 导出选项
type ExportOptions struct {
	IncludeMetadata bool   `json:"include_metadata"`
	IncludeTags     bool   `json:"include_tags"`
	DateFormat      string `json:"date_format"`
	Encoding        string `json:"encoding"`
}

// ExportKnowledgeResponse 导出知识响应
type ExportKnowledgeResponse struct {
	ExportedCount int    `json:"exported_count"`
	Format        string `json:"format"`
	Data          string `json:"data"` // 导出数据内容或文件路径
	FileName      string `json:"file_name"`
	FileSize      int64  `json:"file_size"`
}

// ListWeChatRecordsRequest 列出微信聊天记录请求
type ListWeChatRecordsRequest struct {
	UserID      string `json:"user_id" binding:"required"`
	SessionName string `json:"session_name,omitempty"`
	MessageType string `json:"message_type,omitempty"`
	StartDate   string `json:"start_date,omitempty"` // YYYY-MM-DD
	EndDate     string `json:"end_date,omitempty"`   // YYYY-MM-DD
	Keyword     string `json:"keyword,omitempty"`
	Page        int    `json:"page" binding:"min=1"`
	PageSize    int    `json:"page_size" binding:"min=1,max=100"`
}

// ListWeChatRecordsResponse 列出微信聊天记录响应
type ListWeChatRecordsResponse struct {
	Records  []models.WeChatRecord `json:"records"`
	Total    int64                 `json:"total"`
	Page     int                   `json:"page"`
	PageSize int                   `json:"page_size"`
}

// ArchiveWeChatRecordRequest 归档微信聊天记录请求
type ArchiveWeChatRecordRequest struct {
	RecordID    string   `json:"record_id" binding:"required"`
	UserID      string   `json:"user_id" binding:"required"`
	Title       string   `json:"title" binding:"required"`
	Description string   `json:"description,omitempty"`
	Tags        []string `json:"tags,omitempty"`
	Category    string   `json:"category,omitempty"`
}

// WeChatStatistics 微信聊天记录统计
type WeChatStatistics struct {
	TotalRecords    int64            `json:"total_records"`
	ArchivedRecords int64            `json:"archived_records"`
	WeeklyAdded     int64            `json:"weekly_added"`
	ByType          []WeChatTypeStat `json:"by_type"`
}

// WeChatTypeStat 微信消息类型统计
type WeChatTypeStat struct {
	MessageType string `json:"message_type"`
	Count       int64  `json:"count"`
}

// SubmitShareApplicationRequest 提交分享申请请求
type SubmitShareApplicationRequest struct {
	KnowledgeID    string `json:"knowledge_id" binding:"required"`
	UserID         string `json:"user_id" binding:"required"`
	TeamID         string `json:"team_id" binding:"required"`
	ShareReason    string `json:"share_reason" binding:"required"`
	CreateWorkflow bool   `json:"create_workflow"` // 是否创建审批工作流
}

// ListShareApplicationsRequest 列出分享申请请求
type ListShareApplicationsRequest struct {
	UserID   string `json:"user_id" binding:"required"`
	TeamID   string `json:"team_id,omitempty"`
	Role     string `json:"role,omitempty"` // admin, user
	Status   string `json:"status,omitempty"`
	Page     int    `json:"page" binding:"min=1"`
	PageSize int    `json:"page_size" binding:"min=1,max=100"`
}

// ListShareApplicationsResponse 列出分享申请响应
type ListShareApplicationsResponse struct {
	Applications []ShareApplicationWithKnowledge `json:"applications"`
	Total        int64                           `json:"total"`
	Page         int                             `json:"page"`
	PageSize     int                             `json:"page_size"`
}

// ShareApplicationWithKnowledge 包含知识信息的分享申请
type ShareApplicationWithKnowledge struct {
	models.PersonalKnowledgeShare
	Knowledge *models.PersonalKnowledgeBase `json:"knowledge"`
}

// ReviewShareApplicationRequest 审批分享申请请求
type ReviewShareApplicationRequest struct {
	ShareID      string `json:"share_id" binding:"required"`
	ReviewerID   string `json:"reviewer_id" binding:"required"`
	Decision     string `json:"decision" binding:"required,oneof=approved rejected"`
	ReviewReason string `json:"review_reason"`
}

// ShareApplicationDetail 分享申请详情
type ShareApplicationDetail struct {
	models.PersonalKnowledgeShare
	Knowledge     *models.PersonalKnowledgeBase `json:"knowledge"`
	WorkflowTasks []models.WorkflowTask         `json:"workflow_tasks,omitempty"`
}

// ShareStatistics 分享统计信息
type ShareStatistics struct {
	TotalApplications    int64 `json:"total_applications"`
	PendingApplications  int64 `json:"pending_applications"`
	ApprovedApplications int64 `json:"approved_applications"`
	RejectedApplications int64 `json:"rejected_applications"`
	WeeklyApplications   int64 `json:"weekly_applications"`
}
