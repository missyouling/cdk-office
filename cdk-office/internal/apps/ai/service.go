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

package ai

import (
	"context"
	"fmt"
	"log"
	"time"

	"gorm.io/gorm"

	"github.com/linux-do/cdk-office/internal/apps/dify"
	"github.com/linux-do/cdk-office/internal/models"
)

// Service AI智能问答服务
type Service struct {
	db          *gorm.DB
	difyService *dify.Service
}

// NewService 创建AI服务实例
func NewService(db *gorm.DB, difyConfig *dify.Config) *Service {
	difyService := dify.NewService(difyConfig, db)

	return &Service{
		db:          db,
		difyService: difyService,
	}
}

// ChatRequest 智能问答请求
type ChatRequest struct {
	Question string                 `json:"question" binding:"required"`
	Context  map[string]interface{} `json:"context,omitempty"`
}

// ChatResponse 智能问答响应
type ChatResponse struct {
	Answer     string         `json:"answer"`
	Sources    []DocumentInfo `json:"sources"`
	Confidence float32        `json:"confidence"`
	MessageID  string         `json:"message_id"`
	CreatedAt  time.Time      `json:"created_at"`
}

// DocumentInfo 文档信息
type DocumentInfo struct {
	ID      string  `json:"id"`
	Name    string  `json:"name"`
	Snippet string  `json:"snippet"`
	Score   float64 `json:"score"`
}

// Chat 智能问答核心方法
func (s *Service) Chat(ctx context.Context, userID, teamID string, req *ChatRequest) (*ChatResponse, error) {
	// 参数验证
	if req.Question == "" {
		return nil, fmt.Errorf("question cannot be empty")
	}

	// 创建Dify请求
	difyReq := &dify.KnowledgeQARequest{
		Question: req.Question,
		UserID:   userID,
		TeamID:   teamID,
		Context:  req.Context,
	}

	// 调用Dify服务
	difyResp, err := s.difyService.Chat(ctx, difyReq)
	if err != nil {
		// 记录失败的问答
		qaRecord := &models.KnowledgeQA{
			UserID:     userID,
			TeamID:     teamID,
			Question:   req.Question,
			Answer:     "抱歉，我暂时无法回答您的问题。",
			Confidence: 0.0,
			Feedback:   fmt.Sprintf("error: %v", err),
			AIProvider: "dify",
		}
		s.db.Create(qaRecord)

		return nil, fmt.Errorf("failed to get AI response: %w", err)
	}

	// 转换响应格式
	sources := make([]DocumentInfo, len(difyResp.Sources))
	for i, source := range difyResp.Sources {
		sources[i] = DocumentInfo{
			ID:      source.DocumentID,
			Name:    source.DocumentName,
			Snippet: source.Snippet,
			Score:   source.Score,
		}
	}

	response := &ChatResponse{
		Answer:     difyResp.Answer,
		Sources:    sources,
		Confidence: float32(difyResp.Confidence),
		MessageID:  difyResp.MessageID,
		CreatedAt:  time.Now(),
	}

	// 记录成功的问答（Dify服务中已经记录，这里是业务层的额外处理）
	log.Printf("AI Chat - User: %s, Team: %s, Question: %s, Confidence: %.2f",
		userID, teamID, req.Question, response.Confidence)

	return response, nil
}

// GetChatHistory 获取用户问答历史
func (s *Service) GetChatHistory(ctx context.Context, userID, teamID string, limit, offset int) ([]*models.KnowledgeQA, int64, error) {
	var history []*models.KnowledgeQA
	var total int64

	// 构建查询
	query := s.db.Model(&models.KnowledgeQA{}).Where("user_id = ? AND team_id = ?", userID, teamID)

	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count chat history: %w", err)
	}

	// 分页查询
	if err := query.Order("created_at DESC").Limit(limit).Offset(offset).Find(&history).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to fetch chat history: %w", err)
	}

	return history, total, nil
}

// UpdateFeedback 更新问答反馈
func (s *Service) UpdateFeedback(ctx context.Context, userID, messageID, feedback string) error {
	result := s.db.Model(&models.KnowledgeQA{}).
		Where("user_id = ? AND message_id = ?", userID, messageID).
		Update("feedback", feedback)

	if result.Error != nil {
		return fmt.Errorf("failed to update feedback: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("knowledge QA record not found")
	}

	return nil
}

// GetStats 获取问答统计信息
func (s *Service) GetStats(ctx context.Context, teamID string) (*ChatStats, error) {
	var stats ChatStats

	// 总问答数
	s.db.Model(&models.KnowledgeQA{}).Where("team_id = ?", teamID).Count(&stats.TotalChats)

	// 今日问答数
	today := time.Now().Truncate(24 * time.Hour)
	s.db.Model(&models.KnowledgeQA{}).
		Where("team_id = ? AND created_at >= ?", teamID, today).
		Count(&stats.TodayChats)

	// 平均置信度
	var avgConfidence float64
	s.db.Model(&models.KnowledgeQA{}).
		Where("team_id = ? AND confidence > 0", teamID).
		Select("AVG(confidence)").Scan(&avgConfidence)
	stats.AvgConfidence = float32(avgConfidence)

	// 最活跃用户
	type UserActivity struct {
		UserID string `json:"user_id"`
		Count  int64  `json:"count"`
	}

	var topUsers []UserActivity
	s.db.Model(&models.KnowledgeQA{}).
		Where("team_id = ?", teamID).
		Select("user_id, COUNT(*) as count").
		Group("user_id").
		Order("count DESC").
		Limit(5).
		Scan(&topUsers)

	stats.TopUsers = topUsers

	return &stats, nil
}

// ChatStats 问答统计信息
type ChatStats struct {
	TotalChats    int64         `json:"total_chats"`
	TodayChats    int64         `json:"today_chats"`
	AvgConfidence float32       `json:"avg_confidence"`
	TopUsers      []interface{} `json:"top_users"`
}

// DocumentSyncService 文档同步服务
type DocumentSyncService struct {
	db          *gorm.DB
	difyService *dify.Service
}

// NewDocumentSyncService 创建文档同步服务
func NewDocumentSyncService(db *gorm.DB, difyConfig *dify.Config) *DocumentSyncService {
	difyService := dify.NewService(difyConfig, db)

	return &DocumentSyncService{
		db:          db,
		difyService: difyService,
	}
}

// SyncToDify 异步同步文档到Dify知识库
func (s *DocumentSyncService) SyncToDify(ctx context.Context, doc *models.Document) error {
	// 创建同步记录
	syncRecord := &models.DifyDocumentSync{
		DocumentID:   doc.ID,
		TeamID:       doc.TeamID,
		Title:        doc.Name,
		DocumentType: doc.FileType,
		SyncStatus:   "pending",
		CreatedBy:    doc.CreatedBy,
	}

	if err := s.db.Create(syncRecord).Error; err != nil {
		return fmt.Errorf("failed to create sync record: %w", err)
	}

	// 异步处理同步
	go s.asyncSyncDocument(context.Background(), doc, syncRecord)

	return nil
}

// asyncSyncDocument 异步同步文档处理
func (s *DocumentSyncService) asyncSyncDocument(ctx context.Context, doc *models.Document, syncRecord *models.DifyDocumentSync) {
	// 更新状态为处理中
	s.db.Model(syncRecord).Update("sync_status", "processing")

	// 1. 提取文档内容
	content, err := s.extractDocumentContent(doc)
	if err != nil {
		log.Printf("Failed to extract content from document %s: %v", doc.ID, err)
		s.updateSyncError(syncRecord, fmt.Sprintf("content extraction failed: %v", err))
		return
	}

	// 2. 准备同步请求
	syncReq := &dify.DocumentSyncRequest{
		DocumentID:   doc.ID,
		Title:        doc.Name,
		Content:      content,
		DocumentType: doc.FileType,
		TeamID:       doc.TeamID,
		CreatedBy:    doc.CreatedBy,
		Metadata: map[string]interface{}{
			"file_size":  doc.FileSize,
			"mime_type":  doc.MimeType,
			"created_at": doc.CreatedAt,
			"updated_at": doc.UpdatedAt,
			"version":    doc.Version,
		},
	}

	// 3. 调用Dify同步接口（设置超时）
	syncCtx, cancel := context.WithTimeout(ctx, 5*time.Minute)
	defer cancel()

	resp, err := s.difyService.SyncDocument(syncCtx, syncReq)
	if err != nil {
		log.Printf("Failed to sync document %s to Dify: %v", doc.ID, err)
		s.updateSyncError(syncRecord, fmt.Sprintf("dify sync failed: %v", err))
		return
	}

	// 4. 更新同步状态
	updates := map[string]interface{}{
		"dify_document_id": resp.DifyDocumentID,
		"sync_status":      "synced",
		"indexing_status":  resp.IndexingStatus,
		"updated_at":       time.Now(),
	}

	if err := s.db.Model(syncRecord).Updates(updates).Error; err != nil {
		log.Printf("Failed to update sync record for document %s: %v", doc.ID, err)
		return
	}

	// 5. 更新原文档的Dify ID
	s.db.Model(doc).Updates(map[string]interface{}{
		"dify_document_id": resp.DifyDocumentID,
		"last_sync_at":     time.Now(),
	})

	log.Printf("Document %s successfully synced to Dify (ID: %s)", doc.ID, resp.DifyDocumentID)
}

// extractDocumentContent 提取文档内容
func (s *DocumentSyncService) extractDocumentContent(doc *models.Document) (string, error) {
	// 这里应该根据文件类型提取内容
	// 为了简化，假设文档内容已经在某个地方可以获取

	// 对于大文件，需要分块处理
	if doc.FileSize > 10*1024*1024 { // 10MB
		return s.extractLargeFileContent(doc)
	}

	// TODO: 实现具体的内容提取逻辑
	// 可能需要调用OCR服务、PDF解析器等
	return fmt.Sprintf("Content of document: %s\nFile: %s\nType: %s",
		doc.Name, doc.FileName, doc.FileType), nil
}

// extractLargeFileContent 处理大文件内容提取
func (s *DocumentSyncService) extractLargeFileContent(doc *models.Document) (string, error) {
	// 对于大文件，可能需要分段处理
	// 这里是一个简化的实现
	log.Printf("Processing large file: %s (Size: %d bytes)", doc.Name, doc.FileSize)

	// TODO: 实现大文件分块处理逻辑
	return fmt.Sprintf("Large file content summary: %s\nFile: %s\nSize: %d bytes",
		doc.Name, doc.FileName, doc.FileSize), nil
}

// updateSyncError 更新同步错误信息
func (s *DocumentSyncService) updateSyncError(syncRecord *models.DifyDocumentSync, errorMsg string) {
	updates := map[string]interface{}{
		"sync_status":   "failed",
		"error_message": errorMsg,
		"updated_at":    time.Now(),
	}
	s.db.Model(syncRecord).Updates(updates)
}

// GetSyncStatus 获取文档同步状态
func (s *DocumentSyncService) GetSyncStatus(ctx context.Context, documentID string) (*models.DifyDocumentSync, error) {
	var syncRecord models.DifyDocumentSync

	if err := s.db.Where("document_id = ?", documentID).First(&syncRecord).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("document not synced")
		}
		return nil, fmt.Errorf("failed to get sync status: %w", err)
	}

	return &syncRecord, nil
}

// RetrySync 重试失败的同步
func (s *DocumentSyncService) RetrySync(ctx context.Context, documentID string) error {
	// 获取文档信息
	var doc models.Document
	if err := s.db.Where("id = ?", documentID).First(&doc).Error; err != nil {
		return fmt.Errorf("document not found: %w", err)
	}

	// 获取同步记录
	var syncRecord models.DifyDocumentSync
	if err := s.db.Where("document_id = ?", documentID).First(&syncRecord).Error; err != nil {
		return fmt.Errorf("sync record not found: %w", err)
	}

	// 重置状态并重新同步
	s.db.Model(&syncRecord).Updates(map[string]interface{}{
		"sync_status":   "pending",
		"error_message": "",
		"updated_at":    time.Now(),
	})

	// 启动异步同步
	go s.asyncSyncDocument(context.Background(), &doc, &syncRecord)

	return nil
}
