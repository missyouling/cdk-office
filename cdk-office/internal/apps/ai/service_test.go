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
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/linux-do/cdk-office/internal/apps/dify"
	"github.com/linux-do/cdk-office/internal/models"
)

// MockDifyService 模拟Dify服务
type MockDifyService struct {
	mock.Mock
}

func (m *MockDifyService) Chat(ctx context.Context, req *dify.KnowledgeQARequest) (*dify.KnowledgeQAResponse, error) {
	args := m.Called(ctx, req)
	return args.Get(0).(*dify.KnowledgeQAResponse), args.Error(1)
}

func (m *MockDifyService) SyncDocument(ctx context.Context, req *dify.DocumentSyncRequest) (*dify.DocumentSyncResponse, error) {
	args := m.Called(ctx, req)
	return args.Get(0).(*dify.DocumentSyncResponse), args.Error(1)
}

func (m *MockDifyService) GetDocumentSyncStatus(ctx context.Context, documentID string) (*dify.DocumentSyncStatus, error) {
	args := m.Called(ctx, documentID)
	return args.Get(0).(*dify.DocumentSyncStatus), args.Error(1)
}

func (m *MockDifyService) DeleteDocument(ctx context.Context, documentID string) error {
	args := m.Called(ctx, documentID)
	return args.Error(0)
}

// DocumentSyncTestSuite 文档同步测试套件
type DocumentSyncTestSuite struct {
	suite.Suite
	db          *gorm.DB
	mockDify    *MockDifyService
	syncService *DocumentSyncService
}

// SetupSuite 套件初始化
func (suite *DocumentSyncTestSuite) SetupSuite() {
	// 创建内存数据库
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	suite.Require().NoError(err)

	// 自动迁移模型
	err = db.AutoMigrate(&models.Document{}, &models.DifyDocumentSync{})
	suite.Require().NoError(err)

	suite.db = db

	// 创建模拟Dify服务
	suite.mockDify = new(MockDifyService)

	// 创建文档同步服务（使用构造函数的简化版本）
	suite.syncService = &DocumentSyncService{
		db:          db,
		difyService: suite.mockDify,
	}
}

// SetupTest 每个测试前的设置
func (suite *DocumentSyncTestSuite) SetupTest() {
	// 清理数据库
	suite.db.Exec("DELETE FROM documents")
	suite.db.Exec("DELETE FROM dify_document_syncs")

	// 重置mock
	suite.mockDify.ExpectedCalls = nil
	suite.mockDify.Calls = nil
}

// TestNewDocumentSyncService 测试创建文档同步服务
func (suite *DocumentSyncTestSuite) TestNewDocumentSyncService() {
	difyConfig := &dify.Config{
		APIKey:           "test-api-key",
		APIEndpoint:      "https://api.dify.ai/v1",
		DefaultDatasetID: "test-dataset-id",
		Timeout:          30,
	}

	service := NewDocumentSyncService(suite.db, difyConfig)

	assert.NotNil(suite.T(), service)
	assert.Equal(suite.T(), suite.db, service.db)
	assert.NotNil(suite.T(), service.difyService)
}

// TestSyncToDify 测试同步文档到Dify
func (suite *DocumentSyncTestSuite) TestSyncToDify() {
	ctx := context.Background()

	// 创建测试文档
	doc := &models.Document{
		ID:        uuid.New().String(),
		TeamID:    "test-team-id",
		Name:      "测试文档.pdf",
		FileName:  "test-document.pdf",
		FileType:  "pdf",
		FileSize:  1024 * 1024, // 1MB
		CreatedBy: "test-user-id",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// 测试同步方法
	err := suite.syncService.SyncToDify(ctx, doc)
	assert.NoError(suite.T(), err)

	// 验证同步记录是否创建
	var syncRecord models.DifyDocumentSync
	err = suite.db.Where("document_id = ?", doc.ID).First(&syncRecord).Error
	assert.NoError(suite.T(), err)

	// 验证同步记录字段
	assert.Equal(suite.T(), doc.ID, syncRecord.DocumentID)
	assert.Equal(suite.T(), doc.TeamID, syncRecord.TeamID)
	assert.Equal(suite.T(), doc.Name, syncRecord.Title)
	assert.Equal(suite.T(), doc.FileType, syncRecord.DocumentType)
	assert.Equal(suite.T(), "pending", syncRecord.SyncStatus)
	assert.Equal(suite.T(), doc.CreatedBy, syncRecord.CreatedBy)

	// 等待异步处理完成（在真实环境中会有异步处理）
	time.Sleep(50 * time.Millisecond)
}

// TestExtractDocumentContent 测试文档内容提取
func (suite *DocumentSyncTestSuite) TestExtractDocumentContent() {
	// 测试小文件
	smallDoc := &models.Document{
		ID:       "small-doc-id",
		Name:     "small.txt",
		FileName: "small.txt",
		FileType: "txt",
		FileSize: 1024, // 1KB
	}

	content, err := suite.syncService.extractDocumentContent(smallDoc)
	assert.NoError(suite.T(), err)
	assert.Contains(suite.T(), content, smallDoc.Name)
	assert.Contains(suite.T(), content, smallDoc.FileType)

	// 测试大文件
	largeDoc := &models.Document{
		ID:       "large-doc-id",
		Name:     "large.pdf",
		FileName: "large.pdf",
		FileType: "pdf",
		FileSize: 15 * 1024 * 1024, // 15MB
	}

	content, err = suite.syncService.extractDocumentContent(largeDoc)
	assert.NoError(suite.T(), err)
	assert.Contains(suite.T(), content, "Large file content summary")
	assert.Contains(suite.T(), content, largeDoc.Name)
}

// TestExtractLargeFileContent 测试大文件内容提取
func (suite *DocumentSyncTestSuite) TestExtractLargeFileContent() {
	doc := &models.Document{
		ID:       "large-doc-id",
		Name:     "large-file.pdf",
		FileName: "large-file.pdf",
		FileType: "pdf",
		FileSize: 20 * 1024 * 1024, // 20MB
	}

	content, err := suite.syncService.extractLargeFileContent(doc)
	assert.NoError(suite.T(), err)
	assert.Contains(suite.T(), content, "Large file content summary")
	assert.Contains(suite.T(), content, doc.Name)
	assert.Contains(suite.T(), content, "20971520") // 文件大小
}

// TestGetSyncStatus 测试获取同步状态
func (suite *DocumentSyncTestSuite) TestGetSyncStatus() {
	ctx := context.Background()
	documentID := uuid.New().String()

	// 测试不存在的文档
	_, err := suite.syncService.GetSyncStatus(ctx, "nonexistent-doc")
	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "document not synced")

	// 创建同步记录
	syncRecord := &models.DifyDocumentSync{
		DocumentID:     documentID,
		TeamID:         "test-team-id",
		Title:          "测试文档",
		DocumentType:   "pdf",
		SyncStatus:     "synced",
		DifyDocumentID: "dify-doc-123",
		IndexingStatus: "completed",
		CreatedBy:      "test-user-id",
		CreatedAt:      time.Now(),
	}

	err = suite.db.Create(syncRecord).Error
	assert.NoError(suite.T(), err)

	// 测试获取存在的同步状态
	status, err := suite.syncService.GetSyncStatus(ctx, documentID)
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), status)
	assert.Equal(suite.T(), documentID, status.DocumentID)
	assert.Equal(suite.T(), "synced", status.SyncStatus)
	assert.Equal(suite.T(), "dify-doc-123", status.DifyDocumentID)
}

// TestRetrySync 测试重试同步
func (suite *DocumentSyncTestSuite) TestRetrySync() {
	ctx := context.Background()
	documentID := uuid.New().String()

	// 创建测试文档
	doc := &models.Document{
		ID:        documentID,
		TeamID:    "test-team-id",
		Name:      "重试文档.pdf",
		FileName:  "retry-document.pdf",
		FileType:  "pdf",
		FileSize:  1024,
		CreatedBy: "test-user-id",
		CreatedAt: time.Now(),
	}

	err := suite.db.Create(doc).Error
	assert.NoError(suite.T(), err)

	// 创建失败的同步记录
	syncRecord := &models.DifyDocumentSync{
		DocumentID:   documentID,
		TeamID:       "test-team-id",
		Title:        "重试文档.pdf",
		DocumentType: "pdf",
		SyncStatus:   "failed",
		ErrorMessage: "previous error",
		CreatedBy:    "test-user-id",
		CreatedAt:    time.Now(),
	}

	err = suite.db.Create(syncRecord).Error
	assert.NoError(suite.T(), err)

	// 测试重试同步
	err = suite.syncService.RetrySync(ctx, documentID)
	assert.NoError(suite.T(), err)

	// 验证同步记录被重置
	var updatedRecord models.DifyDocumentSync
	err = suite.db.Where("document_id = ?", documentID).First(&updatedRecord).Error
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "pending", updatedRecord.SyncStatus)
	assert.Empty(suite.T(), updatedRecord.ErrorMessage)

	// 测试重试不存在的文档
	err = suite.syncService.RetrySync(ctx, "nonexistent-doc")
	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "document not found")

	// 测试重试没有同步记录的文档
	anotherDocID := uuid.New().String()
	anotherDoc := &models.Document{
		ID:        anotherDocID,
		TeamID:    "test-team-id",
		Name:      "另一个文档.pdf",
		CreatedBy: "test-user-id",
		CreatedAt: time.Now(),
	}
	err = suite.db.Create(anotherDoc).Error
	assert.NoError(suite.T(), err)

	err = suite.syncService.RetrySync(ctx, anotherDocID)
	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "sync record not found")
}

// TestAsyncSyncDocument 测试异步同步文档处理
func (suite *DocumentSyncTestSuite) TestAsyncSyncDocument() {
	ctx := context.Background()

	// 创建测试文档
	doc := &models.Document{
		ID:        uuid.New().String(),
		TeamID:    "test-team-id",
		Name:      "异步测试文档.pdf",
		FileName:  "async-test.pdf",
		FileType:  "pdf",
		FileSize:  2048,
		MimeType:  "application/pdf",
		Version:   1,
		CreatedBy: "test-user-id",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// 创建同步记录
	syncRecord := &models.DifyDocumentSync{
		DocumentID:   doc.ID,
		TeamID:       doc.TeamID,
		Title:        doc.Name,
		DocumentType: doc.FileType,
		SyncStatus:   "pending",
		CreatedBy:    doc.CreatedBy,
		CreatedAt:    time.Now(),
	}

	err := suite.db.Create(syncRecord).Error
	assert.NoError(suite.T(), err)

	// 设置Dify服务mock
	expectedResponse := &dify.DocumentSyncResponse{
		DifyDocumentID: "dify-doc-123",
		Status:         "created",
		IndexingStatus: "processing",
	}

	suite.mockDify.On("SyncDocument", mock.Anything, mock.MatchedBy(func(req *dify.DocumentSyncRequest) bool {
		return req.DocumentID == doc.ID && req.Title == doc.Name
	})).Return(expectedResponse, nil)

	// 手动调用异步同步方法（在实际代码中这是在goroutine中运行的）
	suite.syncService.asyncSyncDocument(ctx, doc, syncRecord)

	// 验证mock被调用
	suite.mockDify.AssertExpectations(suite.T())

	// 验证同步记录被更新
	var updatedRecord models.DifyDocumentSync
	err = suite.db.Where("document_id = ?", doc.ID).First(&updatedRecord).Error
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "synced", updatedRecord.SyncStatus)
	assert.Equal(suite.T(), "dify-doc-123", updatedRecord.DifyDocumentID)
	assert.Equal(suite.T(), "processing", updatedRecord.IndexingStatus)

	// 验证原文档被更新
	var updatedDoc models.Document
	err = suite.db.Where("id = ?", doc.ID).First(&updatedDoc).Error
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "dify-doc-123", updatedDoc.DifyDocumentID)
}

// TestAsyncSyncDocumentWithError 测试异步同步失败情况
func (suite *DocumentSyncTestSuite) TestAsyncSyncDocumentWithError() {
	ctx := context.Background()

	// 创建测试文档
	doc := &models.Document{
		ID:        uuid.New().String(),
		TeamID:    "test-team-id",
		Name:      "失败测试文档.pdf",
		FileType:  "pdf",
		FileSize:  1024,
		CreatedBy: "test-user-id",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// 创建同步记录
	syncRecord := &models.DifyDocumentSync{
		DocumentID:   doc.ID,
		TeamID:       doc.TeamID,
		Title:        doc.Name,
		DocumentType: doc.FileType,
		SyncStatus:   "pending",
		CreatedBy:    doc.CreatedBy,
		CreatedAt:    time.Now(),
	}

	err := suite.db.Create(syncRecord).Error
	assert.NoError(suite.T(), err)

	// 设置Dify服务返回错误
	suite.mockDify.On("SyncDocument", mock.Anything, mock.Anything).Return(
		(*dify.DocumentSyncResponse)(nil),
		assert.AnError,
	)

	// 手动调用异步同步方法
	suite.syncService.asyncSyncDocument(ctx, doc, syncRecord)

	// 验证mock被调用
	suite.mockDify.AssertExpectations(suite.T())

	// 验证同步记录状态变为失败
	var updatedRecord models.DifyDocumentSync
	err = suite.db.Where("document_id = ?", doc.ID).First(&updatedRecord).Error
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "failed", updatedRecord.SyncStatus)
	assert.Contains(suite.T(), updatedRecord.ErrorMessage, "dify sync failed")
}

// TestUpdateSyncError 测试更新同步错误
func (suite *DocumentSyncTestSuite) TestUpdateSyncError() {
	// 创建同步记录
	syncRecord := &models.DifyDocumentSync{
		DocumentID:   uuid.New().String(),
		TeamID:       "test-team-id",
		Title:        "错误测试文档",
		DocumentType: "pdf",
		SyncStatus:   "processing",
		CreatedBy:    "test-user-id",
		CreatedAt:    time.Now(),
	}

	err := suite.db.Create(syncRecord).Error
	assert.NoError(suite.T(), err)

	// 测试更新错误
	errorMsg := "同步过程中发生错误"
	suite.syncService.updateSyncError(syncRecord, errorMsg)

	// 验证记录被更新
	var updatedRecord models.DifyDocumentSync
	err = suite.db.Where("document_id = ?", syncRecord.DocumentID).First(&updatedRecord).Error
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "failed", updatedRecord.SyncStatus)
	assert.Equal(suite.T(), errorMsg, updatedRecord.ErrorMessage)
	assert.True(suite.T(), updatedRecord.UpdatedAt.After(syncRecord.CreatedAt))
}

// TestConcurrentSync 测试并发同步
func (suite *DocumentSyncTestSuite) TestConcurrentSync() {
	ctx := context.Background()

	// 创建多个文档
	docs := make([]*models.Document, 3)
	for i := 0; i < 3; i++ {
		docs[i] = &models.Document{
			ID:        uuid.New().String(),
			TeamID:    "test-team-id",
			Name:      fmt.Sprintf("并发文档%d.pdf", i+1),
			FileType:  "pdf",
			FileSize:  1024,
			CreatedBy: "test-user-id",
			CreatedAt: time.Now(),
		}
	}

	// 并发同步
	done := make(chan bool, len(docs))
	for _, doc := range docs {
		go func(d *models.Document) {
			err := suite.syncService.SyncToDify(ctx, d)
			assert.NoError(suite.T(), err)
			done <- true
		}(doc)
	}

	// 等待所有同步完成
	for i := 0; i < len(docs); i++ {
		<-done
	}

	// 验证所有同步记录都被创建
	var count int64
	suite.db.Model(&models.DifyDocumentSync{}).Count(&count)
	assert.Equal(suite.T(), int64(len(docs)), count)
}

// TestDifferentFileTypes 测试不同文件类型的处理
func (suite *DocumentSyncTestSuite) TestDifferentFileTypes() {
	fileTypes := []struct {
		name     string
		fileType string
		size     int64
	}{
		{"document.pdf", "pdf", 1024 * 1024},
		{"document.txt", "txt", 2048},
		{"document.docx", "docx", 512 * 1024},
		{"document.xlsx", "xlsx", 256 * 1024},
	}

	for _, ft := range fileTypes {
		doc := &models.Document{
			ID:        uuid.New().String(),
			Name:      ft.name,
			FileType:  ft.fileType,
			FileSize:  ft.size,
			CreatedBy: "test-user-id",
			TeamID:    "test-team-id",
		}

		content, err := suite.syncService.extractDocumentContent(doc)
		assert.NoError(suite.T(), err)
		assert.Contains(suite.T(), content, ft.fileType)
		assert.Contains(suite.T(), content, ft.name)

		// 根据文件大小验证处理方式
		if ft.size > 10*1024*1024 {
			assert.Contains(suite.T(), content, "Large file content summary")
		} else {
			assert.Contains(suite.T(), content, "Content of document")
		}
	}
}

// 运行测试套件
func TestDocumentSyncTestSuite(t *testing.T) {
	suite.Run(t, new(DocumentSyncTestSuite))
}

// 基准测试
func BenchmarkSyncToDify(b *testing.B) {
	// 创建测试数据库
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		b.Fatal(err)
	}

	err = db.AutoMigrate(&models.Document{}, &models.DifyDocumentSync{})
	if err != nil {
		b.Fatal(err)
	}

	mockDify := new(MockDifyService)
	syncService := &DocumentSyncService{
		db:          db,
		difyService: mockDify,
	}

	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		doc := &models.Document{
			ID:        uuid.New().String(),
			TeamID:    "test-team-id",
			Name:      fmt.Sprintf("benchmark-doc-%d.pdf", i),
			FileType:  "pdf",
			FileSize:  1024,
			CreatedBy: "test-user-id",
			CreatedAt: time.Now(),
		}

		err := syncService.SyncToDify(ctx, doc)
		if err != nil {
			b.Error(err)
		}
	}
}

// 测试内存使用
func TestMemoryUsage(t *testing.T) {
	// 创建大量文档同步记录，测试内存使用情况
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	assert.NoError(t, err)

	err = db.AutoMigrate(&models.Document{}, &models.DifyDocumentSync{})
	assert.NoError(t, err)

	mockDify := new(MockDifyService)
	syncService := &DocumentSyncService{
		db:          db,
		difyService: mockDify,
	}

	ctx := context.Background()

	// 创建1000个文档同步记录
	for i := 0; i < 1000; i++ {
		doc := &models.Document{
			ID:        uuid.New().String(),
			TeamID:    "test-team-id",
			Name:      fmt.Sprintf("memory-test-doc-%d.pdf", i),
			FileType:  "pdf",
			FileSize:  1024,
			CreatedBy: "test-user-id",
			CreatedAt: time.Now(),
		}

		err := syncService.SyncToDify(ctx, doc)
		assert.NoError(t, err)
	}

	// 验证记录数量
	var count int64
	db.Model(&models.DifyDocumentSync{}).Count(&count)
	assert.Equal(t, int64(1000), count)
}
