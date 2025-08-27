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

package dify

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"gorm.io/gorm"

	"github.com/linux-do/cdk-office/internal/models"
)

// Service Dify集成服务
type Service struct {
	db     *gorm.DB
	client *Client
	config *Config
	cache  map[string]*CacheEntry
}

// Config Dify配置
type Config struct {
	BaseURL                   string `json:"base_url"`
	APIKey                    string `json:"api_key"`
	DefaultDatasetID          string `json:"default_dataset_id"`
	SurveyAnalysisWorkflowID  string `json:"survey_analysis_workflow_id"`
	DocumentProcessWorkflowID string `json:"document_process_workflow_id"`
	KnowledgeBaseID           string `json:"knowledge_base_id"`
	EnableAutoSync            bool   `json:"enable_auto_sync"`
	SyncInterval              int    `json:"sync_interval"` // 分钟
}

// CacheEntry 缓存条目
type CacheEntry struct {
	Data      interface{}
	ExpiresAt time.Time
}

// NewService 创建Dify集成服务
func NewService(config *Config, db *gorm.DB) *Service {
	client := NewClient(config.BaseURL, config.APIKey)

	return &Service{
		db:     db,
		client: client,
		config: config,
		cache:  make(map[string]*CacheEntry),
	}
}

// KnowledgeQARequest 知识问答请求
type KnowledgeQARequest struct {
	Question string                 `json:"question"`
	UserID   string                 `json:"user_id"`
	TeamID   string                 `json:"team_id"`
	Context  map[string]interface{} `json:"context,omitempty"`
}

// KnowledgeQAResponse 知识问答响应
type KnowledgeQAResponse struct {
	Answer     string                 `json:"answer"`
	Sources    []DocumentSource       `json:"sources"`
	Confidence float64                `json:"confidence"`
	MessageID  string                 `json:"message_id"`
	Usage      UsageInfo              `json:"usage"`
	Metadata   map[string]interface{} `json:"metadata"`
}

// DocumentSource 文档来源
type DocumentSource struct {
	DocumentID   string  `json:"document_id"`
	DocumentName string  `json:"document_name"`
	Snippet      string  `json:"snippet"`
	Score        float64 `json:"score"`
}

// DocumentSyncRequest 文档同步请求
type DocumentSyncRequest struct {
	DocumentID   string                 `json:"document_id"`
	Title        string                 `json:"title"`
	Content      string                 `json:"content"`
	DocumentType string                 `json:"document_type"`
	TeamID       string                 `json:"team_id"`
	CreatedBy    string                 `json:"created_by"`
	Metadata     map[string]interface{} `json:"metadata"`
}

// DocumentSyncResponse 文档同步响应
type DocumentSyncResponse struct {
	DifyDocumentID string `json:"dify_document_id"`
	BatchID        string `json:"batch_id"`
	Status         string `json:"status"`
	IndexingStatus string `json:"indexing_status"`
}

// Chat 智能问答
func (s *Service) Chat(ctx context.Context, req *KnowledgeQARequest) (*KnowledgeQAResponse, error) {
	// 记录问答历史
	qaRecord := &models.KnowledgeQA{
		UserID:     req.UserID,
		TeamID:     req.TeamID,
		Question:   req.Question,
		AIProvider: "dify",
	}

	// 构建Dify聊天请求
	chatReq := &ChatRequest{
		Query:        req.Question,
		ResponseMode: "blocking",
		User:         req.UserID,
		Inputs:       make(map[string]interface{}),
	}

	// 添加上下文信息
	if req.Context != nil {
		for key, value := range req.Context {
			chatReq.Inputs[key] = value
		}
	}

	// 添加团队信息作为上下文
	chatReq.Inputs["team_id"] = req.TeamID
	chatReq.Inputs["timestamp"] = time.Now().Unix()

	// 调用Dify API
	chatResp, err := s.client.Chat(ctx, chatReq)
	if err != nil {
		qaRecord.Answer = "抱歉，我暂时无法回答您的问题。"
		qaRecord.Feedback = "error: " + err.Error()
		s.db.Create(qaRecord)
		return nil, fmt.Errorf("failed to get answer from Dify: %w", err)
	}

	// 解析响应
	response := &KnowledgeQAResponse{
		Answer:    chatResp.Answer,
		MessageID: chatResp.MessageID,
		Sources:   []DocumentSource{},
		Metadata:  chatResp.Metadata,
	}

	// 提取文档来源信息
	if metadata := chatResp.Metadata; metadata != nil {
		if retrieval, ok := metadata["retriever_resources"].([]interface{}); ok {
			for _, item := range retrieval {
				if source, ok := item.(map[string]interface{}); ok {
					docSource := DocumentSource{
						DocumentID:   getStringFromMap(source, "document_id"),
						DocumentName: getStringFromMap(source, "document_name"),
						Snippet:      getStringFromMap(source, "content"),
						Score:        getFloatFromMap(source, "score"),
					}
					response.Sources = append(response.Sources, docSource)
				}
			}
		}
	}

	// 计算置信度
	response.Confidence = s.calculateConfidence(response)

	// 保存问答记录
	qaRecord.Answer = response.Answer
	qaRecord.Confidence = float32(response.Confidence)
	if len(response.Sources) > 0 {
		sourceIDs := make([]string, len(response.Sources))
		for i, source := range response.Sources {
			sourceIDs[i] = source.DocumentID
		}
		sourcesJSON, _ := json.Marshal(sourceIDs)
		qaRecord.Sources = string(sourcesJSON)
	}

	s.db.Create(qaRecord)

	log.Printf("Knowledge QA: User=%s, Question=%s, Confidence=%.2f", req.UserID, req.Question, response.Confidence)

	return response, nil
}

// StreamingChat 流式智能问答
func (s *Service) StreamingChat(ctx context.Context, req *KnowledgeQARequest, callback func(*StreamingChatResponse) error) error {
	chatReq := &ChatRequest{
		Query:        req.Question,
		ResponseMode: "streaming",
		User:         req.UserID,
		Inputs:       make(map[string]interface{}),
	}

	// 添加上下文信息
	if req.Context != nil {
		for key, value := range req.Context {
			chatReq.Inputs[key] = value
		}
	}

	chatReq.Inputs["team_id"] = req.TeamID
	chatReq.Inputs["timestamp"] = time.Now().Unix()

	return s.client.StreamingChat(ctx, chatReq, callback)
}

// SyncDocument 同步文档到Dify知识库
func (s *Service) SyncDocument(ctx context.Context, req *DocumentSyncRequest) (*DocumentSyncResponse, error) {
	if s.config.DefaultDatasetID == "" {
		return nil, fmt.Errorf("default dataset ID not configured")
	}

	// 检查文档是否已存在
	existingDoc, err := s.findExistingDocument(req.DocumentID)
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, fmt.Errorf("failed to check existing document: %w", err)
	}

	var response *DocumentSyncResponse

	if existingDoc != nil {
		// 更新现有文档
		err = s.client.UpdateDocument(ctx, s.config.DefaultDatasetID, existingDoc.DifyDocumentID, req.Title, req.Content, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to update document in Dify: %w", err)
		}

		response = &DocumentSyncResponse{
			DifyDocumentID: existingDoc.DifyDocumentID,
			Status:         "updated",
			IndexingStatus: "processing",
		}

		// 更新数据库记录
		existingDoc.Title = req.Title
		existingDoc.Content = req.Content
		existingDoc.UpdatedAt = time.Now()
		s.db.Save(existingDoc)

	} else {
		// 创建新文档
		uploadResp, err := s.client.UploadDocumentByText(ctx, s.config.DefaultDatasetID, req.Title, req.Content, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to upload document to Dify: %w", err)
		}

		response = &DocumentSyncResponse{
			DifyDocumentID: uploadResp.DocumentID,
			BatchID:        uploadResp.BatchID,
			Status:         "created",
			IndexingStatus: "processing",
		}

		// 保存同步记录
		syncRecord := &models.DifyDocumentSync{
			DocumentID:     req.DocumentID,
			DifyDocumentID: uploadResp.DocumentID,
			DatasetID:      s.config.DefaultDatasetID,
			Title:          req.Title,
			Content:        req.Content,
			DocumentType:   req.DocumentType,
			TeamID:         req.TeamID,
			SyncStatus:     "synced",
			CreatedBy:      req.CreatedBy,
		}

		s.db.Create(syncRecord)
	}

	log.Printf("Document synced: ID=%s, Dify ID=%s, Status=%s", req.DocumentID, response.DifyDocumentID, response.Status)

	return response, nil
}

// DeleteDocument 从Dify知识库删除文档
func (s *Service) DeleteDocument(ctx context.Context, documentID string) error {
	// 查找同步记录
	var syncRecord models.DifyDocumentSync
	if err := s.db.Where("document_id = ?", documentID).First(&syncRecord).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil // 文档未同步到Dify，无需删除
		}
		return fmt.Errorf("failed to find sync record: %w", err)
	}

	// 从Dify删除文档
	err := s.client.DeleteDocument(ctx, syncRecord.DatasetID, syncRecord.DifyDocumentID)
	if err != nil {
		log.Printf("Failed to delete document from Dify: %v", err)
		// 不返回错误，允许继续清理本地记录
	}

	// 删除同步记录
	s.db.Delete(&syncRecord)

	log.Printf("Document deleted from Dify: ID=%s, Dify ID=%s", documentID, syncRecord.DifyDocumentID)

	return nil
}

// GetDocumentSyncStatus 获取文档同步状态
func (s *Service) GetDocumentSyncStatus(ctx context.Context, documentID string) (*DocumentSyncStatus, error) {
	var syncRecord models.DifyDocumentSync
	if err := s.db.Where("document_id = ?", documentID).First(&syncRecord).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return &DocumentSyncStatus{
				DocumentID: documentID,
				IsSynced:   false,
				SyncStatus: "not_synced",
			}, nil
		}
		return nil, fmt.Errorf("failed to find sync record: %w", err)
	}

	// 检查Dify中的索引状态
	docInfo, err := s.client.GetDocumentIndexingStatus(ctx, syncRecord.DatasetID, syncRecord.DifyDocumentID)
	if err != nil {
		log.Printf("Failed to get indexing status from Dify: %v", err)
		return &DocumentSyncStatus{
			DocumentID:     documentID,
			DifyDocumentID: syncRecord.DifyDocumentID,
			IsSynced:       true,
			SyncStatus:     syncRecord.SyncStatus,
			IndexingStatus: "unknown",
			SyncedAt:       syncRecord.CreatedAt,
		}, nil
	}

	return &DocumentSyncStatus{
		DocumentID:     documentID,
		DifyDocumentID: syncRecord.DifyDocumentID,
		IsSynced:       true,
		SyncStatus:     syncRecord.SyncStatus,
		IndexingStatus: docInfo.IndexingStatus,
		SyncedAt:       syncRecord.CreatedAt,
		IndexedAt:      time.Unix(docInfo.CreatedAt, 0),
	}, nil
}

// DocumentSyncStatus 文档同步状态
type DocumentSyncStatus struct {
	DocumentID     string    `json:"document_id"`
	DifyDocumentID string    `json:"dify_document_id"`
	IsSynced       bool      `json:"is_synced"`
	SyncStatus     string    `json:"sync_status"`
	IndexingStatus string    `json:"indexing_status"`
	SyncedAt       time.Time `json:"synced_at"`
	IndexedAt      time.Time `json:"indexed_at"`
}

// RunSurveyAnalysis 运行问卷分析工作流
func (s *Service) RunSurveyAnalysis(ctx context.Context, surveyID string, responses []map[string]interface{}) (*WorkflowRunResponse, error) {
	if s.config.SurveyAnalysisWorkflowID == "" {
		return nil, fmt.Errorf("survey analysis workflow ID not configured")
	}

	req := &WorkflowRunRequest{
		Inputs: map[string]interface{}{
			"survey_id": surveyID,
			"responses": responses,
			"timestamp": time.Now().Unix(),
		},
		ResponseMode: "blocking",
		User:         "system",
	}

	return s.client.RunWorkflow(ctx, req)
}

// BatchSyncDocuments 批量同步文档
func (s *Service) BatchSyncDocuments(ctx context.Context, documentIDs []string) error {
	for _, docID := range documentIDs {
		// 这里应该从文档服务获取文档信息
		// 为了简化，这里只是一个占位符实现
		log.Printf("Syncing document: %s", docID)

		// 实际实现中应该：
		// 1. 获取文档详细信息
		// 2. 调用SyncDocument方法
		// 3. 处理错误和重试
	}

	return nil
}

// StartAutoSync 启动自动同步
func (s *Service) StartAutoSync(ctx context.Context) {
	if !s.config.EnableAutoSync {
		return
	}

	interval := time.Duration(s.config.SyncInterval) * time.Minute
	ticker := time.NewTicker(interval)

	go func() {
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				s.performAutoSync(ctx)
			}
		}
	}()

	log.Printf("Auto sync started with interval: %v", interval)
}

// performAutoSync 执行自动同步
func (s *Service) performAutoSync(ctx context.Context) {
	// 查找需要同步的文档
	var documents []models.Document
	if err := s.db.Where("updated_at > ?", time.Now().Add(-time.Duration(s.config.SyncInterval)*time.Minute)).Find(&documents).Error; err != nil {
		log.Printf("Failed to find documents for auto sync: %v", err)
		return
	}

	for _, doc := range documents {
		req := &DocumentSyncRequest{
			DocumentID:   doc.ID,
			Title:        doc.Name,
			Content:      doc.Content,
			DocumentType: doc.DocumentType,
			TeamID:       doc.TeamID,
			CreatedBy:    doc.CreatedBy,
		}

		if _, err := s.SyncDocument(ctx, req); err != nil {
			log.Printf("Failed to auto sync document %s: %v", doc.ID, err)
		}
	}

	log.Printf("Auto sync completed, processed %d documents", len(documents))
}

// 辅助函数

func (s *Service) findExistingDocument(documentID string) (*models.DifyDocumentSync, error) {
	var syncRecord models.DifyDocumentSync
	err := s.db.Where("document_id = ?", documentID).First(&syncRecord).Error
	if err != nil {
		return nil, err
	}
	return &syncRecord, nil
}

func (s *Service) calculateConfidence(response *KnowledgeQAResponse) float64 {
	// 简单的置信度计算逻辑
	// 实际项目中可以基于更复杂的算法
	baseConfidence := 0.5

	if len(response.Sources) > 0 {
		var totalScore float64
		for _, source := range response.Sources {
			totalScore += source.Score
		}
		avgScore := totalScore / float64(len(response.Sources))
		baseConfidence = 0.3 + (avgScore * 0.7)
	}

	// 基于答案长度调整置信度
	if len(response.Answer) > 50 {
		baseConfidence += 0.1
	}

	if baseConfidence > 1.0 {
		baseConfidence = 1.0
	}

	return baseConfidence
}

func getStringFromMap(m map[string]interface{}, key string) string {
	if value, ok := m[key]; ok {
		if str, ok := value.(string); ok {
			return str
		}
	}
	return ""
}

func getFloatFromMap(m map[string]interface{}, key string) float64 {
	if value, ok := m[key]; ok {
		if f, ok := value.(float64); ok {
			return f
		}
		if i, ok := value.(int); ok {
			return float64(i)
		}
	}
	return 0.0
}
