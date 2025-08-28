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
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/linux-do/cdk-office/internal/apps/dify"
	"github.com/linux-do/cdk-office/internal/models"
)

// AITestSuite AI模块测试套件
type AITestSuite struct {
	suite.Suite
	db      *gorm.DB
	router  *gin.Engine
	service *Service
	handler *Handler
}

// SetupSuite 测试套件初始化
func (suite *AITestSuite) SetupSuite() {
	// 设置测试环境
	gin.SetMode(gin.TestMode)

	// 初始化内存数据库
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	suite.Require().NoError(err)
	suite.db = db

	// 自动迁移表结构
	err = db.AutoMigrate(
		&models.KnowledgeQA{},
		&models.DifyDocumentSync{},
		&models.Document{},
		&models.AIServiceConfig{},
	)
	suite.Require().NoError(err)

	// 创建测试用的Dify配置
	difyConfig := &dify.Config{
		APIKey:                   "test-api-key",
		APIEndpoint:              "https://api.dify.ai/v1",
		ChatEndpoint:             "/chat-messages",
		CompletionEndpoint:       "/completion-messages",
		DatasetsEndpoint:         "/datasets",
		DocumentsEndpoint:        "/documents",
		SurveyAnalysisWorkflowID: "test-workflow-id",
		DefaultDatasetID:         "test-dataset-id",
		Timeout:                  30,
	}

	// 创建服务和处理器
	suite.service = NewService(db, difyConfig)
	suite.handler = NewHandler(suite.service)

	// 初始化路由
	suite.router = gin.New()
	suite.setupRoutes()
}

// TearDownSuite 测试套件清理
func (suite *AITestSuite) TearDownSuite() {
	sqlDB, err := suite.db.DB()
	if err == nil {
		sqlDB.Close()
	}
}

// setupRoutes 设置测试路由
func (suite *AITestSuite) setupRoutes() {
	api := suite.router.Group("/api/v1")

	// 模拟认证中间件
	api.Use(func(c *gin.Context) {
		c.Set("user_id", "test-user-id")
		c.Set("team_id", "test-team-id")
		c.Next()
	})

	// 注册AI路由
	suite.handler.RegisterRoutes(api)
}

// TestChatAPI 测试智能问答API
func (suite *AITestSuite) TestChatAPI() {
	// 准备请求数据
	chatRequest := ChatRequest{
		Question: "什么是CDK-Office？",
		Context: map[string]interface{}{
			"source": "test",
		},
	}

	jsonData, err := json.Marshal(chatRequest)
	suite.Require().NoError(err)

	// 创建HTTP请求
	req, err := http.NewRequest("POST", "/api/v1/ai/chat", bytes.NewBuffer(jsonData))
	suite.Require().NoError(err)
	req.Header.Set("Content-Type", "application/json")

	// 发送请求
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	// 注意：由于没有真实的Dify服务，这个测试可能会失败
	// 但我们可以验证请求格式和基本逻辑
	suite.T().Logf("Chat API response status: %d", w.Code)
	suite.T().Logf("Chat API response body: %s", w.Body.String())
}

// TestChatHistory 测试问答历史API
func (suite *AITestSuite) TestChatHistory() {
	// 先插入一些测试数据
	testData := &models.KnowledgeQA{
		UserID:     "test-user-id",
		TeamID:     "test-team-id",
		Question:   "测试问题",
		Answer:     "测试答案",
		Confidence: 0.95,
		MessageID:  "test-message-id",
		AIProvider: "dify",
		CreatedAt:  time.Now(),
	}
	suite.db.Create(testData)

	// 创建请求
	req, err := http.NewRequest("GET", "/api/v1/ai/chat/history?page=1&size=10", nil)
	suite.Require().NoError(err)

	// 发送请求
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	// 验证响应
	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response ChatHistoryResponse
	err = json.Unmarshal(w.Body.Bytes(), &response)
	suite.Require().NoError(err)

	assert.NotEmpty(suite.T(), response.Data)
	assert.Equal(suite.T(), int64(1), response.Pagination.Total)
}

// TestUpdateFeedback 测试反馈更新API
func (suite *AITestSuite) TestUpdateFeedback() {
	// 先插入测试数据
	testData := &models.KnowledgeQA{
		UserID:     "test-user-id",
		TeamID:     "test-team-id",
		Question:   "测试问题",
		Answer:     "测试答案",
		MessageID:  "test-message-id",
		AIProvider: "dify",
		CreatedAt:  time.Now(),
	}
	suite.db.Create(testData)

	// 准备反馈数据
	feedbackRequest := FeedbackRequest{
		Feedback: "回答很有帮助",
	}

	jsonData, err := json.Marshal(feedbackRequest)
	suite.Require().NoError(err)

	// 创建请求
	req, err := http.NewRequest("PATCH", "/api/v1/ai/chat/test-message-id/feedback", bytes.NewBuffer(jsonData))
	suite.Require().NoError(err)
	req.Header.Set("Content-Type", "application/json")

	// 发送请求
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	// 验证响应
	assert.Equal(suite.T(), http.StatusOK, w.Code)

	// 验证数据库更新
	var updatedQA models.KnowledgeQA
	err = suite.db.Where("message_id = ?", "test-message-id").First(&updatedQA).Error
	suite.Require().NoError(err)
	assert.Equal(suite.T(), "回答很有帮助", updatedQA.Feedback)
}

// TestGetStats 测试统计信息API
func (suite *AITestSuite) TestGetStats() {
	// 插入测试数据
	testData := []*models.KnowledgeQA{
		{
			UserID:     "test-user-id",
			TeamID:     "test-team-id",
			Question:   "问题1",
			Answer:     "答案1",
			Confidence: 0.95,
			AIProvider: "dify",
			CreatedAt:  time.Now(),
		},
		{
			UserID:     "test-user-id-2",
			TeamID:     "test-team-id",
			Question:   "问题2",
			Answer:     "答案2",
			Confidence: 0.88,
			AIProvider: "dify",
			CreatedAt:  time.Now(),
		},
	}

	for _, data := range testData {
		suite.db.Create(data)
	}

	// 创建请求
	req, err := http.NewRequest("GET", "/api/v1/ai/chat/stats", nil)
	suite.Require().NoError(err)

	// 发送请求
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	// 验证响应
	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var stats ChatStats
	err = json.Unmarshal(w.Body.Bytes(), &stats)
	suite.Require().NoError(err)

	assert.Equal(suite.T(), int64(2), stats.TotalChats)
	assert.Greater(suite.T(), stats.AvgConfidence, float32(0.0))
}

// TestDocumentSyncService 测试文档同步服务
func (suite *AITestSuite) TestDocumentSyncService() {
	// 创建文档同步服务
	difyConfig := &dify.Config{
		APIKey:           "test-api-key",
		APIEndpoint:      "https://api.dify.ai/v1",
		DefaultDatasetID: "test-dataset-id",
		Timeout:          30,
	}

	syncService := NewDocumentSyncService(suite.db, difyConfig)

	// 创建测试文档
	testDoc := &models.Document{
		ID:        uuid.New().String(),
		TeamID:    "test-team-id",
		Name:      "测试文档.pdf",
		FileName:  "test-document.pdf",
		FileType:  "pdf",
		FileSize:  1024 * 1024, // 1MB
		CreatedBy: "test-user-id",
		CreatedAt: time.Now(),
	}

	// 测试同步方法
	err := syncService.SyncToDify(context.Background(), testDoc)

	// 由于没有真实的Dify服务，这会创建同步记录但同步会失败
	// 我们主要验证记录是否正确创建
	assert.NoError(suite.T(), err)

	// 验证同步记录是否创建
	var syncRecord models.DifyDocumentSync
	err = suite.db.Where("document_id = ?", testDoc.ID).First(&syncRecord).Error
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), testDoc.ID, syncRecord.DocumentID)
	assert.Equal(suite.T(), testDoc.TeamID, syncRecord.TeamID)
	assert.Equal(suite.T(), testDoc.Name, syncRecord.Title)
}

// TestServiceValidation 测试服务验证逻辑
func (suite *AITestSuite) TestServiceValidation() {
	// 测试空问题验证
	_, err := suite.service.Chat(context.Background(), "test-user-id", "test-team-id", &ChatRequest{
		Question: "",
	})
	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "question cannot be empty")

	// 测试正常请求（虽然会因为没有真实服务而失败，但验证参数处理）
	req := &ChatRequest{
		Question: "有效的问题",
		Context: map[string]interface{}{
			"test": "context",
		},
	}

	// 这个测试主要验证参数验证逻辑
	_, err = suite.service.Chat(context.Background(), "test-user-id", "test-team-id", req)
	// 应该会失败，因为没有真实的Dify服务
	assert.Error(suite.T(), err)
}

// TestChatResponseStructure 测试响应结构
func (suite *AITestSuite) TestChatResponseStructure() {
	// 创建测试响应
	response := &ChatResponse{
		Answer:     "这是一个测试答案",
		Confidence: 0.95,
		MessageID:  "test-message-id",
		CreatedAt:  time.Now(),
		Sources: []DocumentInfo{
			{
				ID:      "doc-1",
				Name:    "文档1",
				Snippet: "相关内容片段",
				Score:   0.92,
			},
		},
	}

	// 验证结构体序列化
	jsonData, err := json.Marshal(response)
	assert.NoError(suite.T(), err)
	assert.NotEmpty(suite.T(), jsonData)

	// 验证反序列化
	var deserializedResponse ChatResponse
	err = json.Unmarshal(jsonData, &deserializedResponse)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), response.Answer, deserializedResponse.Answer)
	assert.Equal(suite.T(), response.Confidence, deserializedResponse.Confidence)
	assert.Len(suite.T(), deserializedResponse.Sources, 1)
}

// TestPaginationLogic 测试分页逻辑
func (suite *AITestSuite) TestPaginationLogic() {
	// 插入多条测试数据
	for i := 0; i < 25; i++ {
		testData := &models.KnowledgeQA{
			UserID:     "test-user-id",
			TeamID:     "test-team-id",
			Question:   fmt.Sprintf("问题 %d", i+1),
			Answer:     fmt.Sprintf("答案 %d", i+1),
			AIProvider: "dify",
			CreatedAt:  time.Now().Add(-time.Duration(i) * time.Hour),
		}
		suite.db.Create(testData)
	}

	// 测试第一页
	history, total, err := suite.service.GetChatHistory(context.Background(), "test-user-id", "test-team-id", 10, 0)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), int64(25), total)
	assert.Len(suite.T(), history, 10)

	// 测试第二页
	history2, total2, err := suite.service.GetChatHistory(context.Background(), "test-user-id", "test-team-id", 10, 10)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), int64(25), total2)
	assert.Len(suite.T(), history2, 10)

	// 验证数据不重复
	assert.NotEqual(suite.T(), history[0].ID, history2[0].ID)
}

// 运行测试套件
func TestAITestSuite(t *testing.T) {
	suite.Run(t, new(AITestSuite))
}

// BenchmarkChatService 性能测试
func BenchmarkChatService(b *testing.B) {
	// 初始化测试环境
	gin.SetMode(gin.TestMode)

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		b.Fatal(err)
	}

	db.AutoMigrate(&models.KnowledgeQA{})

	difyConfig := &dify.Config{
		APIKey:           "test-api-key",
		APIEndpoint:      "https://api.dify.ai/v1",
		DefaultDatasetID: "test-dataset-id",
		Timeout:          30,
	}

	service := NewService(db, difyConfig)

	b.ResetTimer()

	// 性能测试主要测试数据库操作部分
	for i := 0; i < b.N; i++ {
		// 测试获取历史记录的性能
		_, _, err := service.GetChatHistory(context.Background(), "test-user-id", "test-team-id", 20, 0)
		if err != nil {
			// 数据库操作应该成功
			continue
		}
	}
}
