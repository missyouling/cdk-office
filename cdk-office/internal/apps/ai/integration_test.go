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
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/linux-do/cdk-office/internal/models"
)

// AIAPIIntegrationTestSuite AI API集成测试套件
type AIAPIIntegrationTestSuite struct {
	suite.Suite
	db          *gorm.DB
	router      *gin.Engine
	mockService *MockAIService
	handler     *Handler
}

// MockAIService 模拟AI服务
type MockAIService struct {
	mock.Mock
}

func (m *MockAIService) Chat(ctx context.Context, userID, teamID string, req *ChatRequest) (*ChatResponse, error) {
	args := m.Called(ctx, userID, teamID, req)
	return args.Get(0).(*ChatResponse), args.Error(1)
}

func (m *MockAIService) GetChatHistory(ctx context.Context, userID, teamID string, size, offset int) ([]*models.KnowledgeQA, int64, error) {
	args := m.Called(ctx, userID, teamID, size, offset)
	return args.Get(0).([]*models.KnowledgeQA), args.Get(1).(int64), args.Error(2)
}

func (m *MockAIService) UpdateFeedback(ctx context.Context, userID, messageID, feedback string) error {
	args := m.Called(ctx, userID, messageID, feedback)
	return args.Error(0)
}

func (m *MockAIService) GetStats(ctx context.Context, teamID string) (*ChatStats, error) {
	args := m.Called(ctx, teamID)
	return args.Get(0).(*ChatStats), args.Error(1)
}

// SetupSuite 套件初始化
func (suite *AIAPIIntegrationTestSuite) SetupSuite() {
	// 设置Gin为测试模式
	gin.SetMode(gin.TestMode)

	// 创建内存数据库
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	suite.Require().NoError(err)

	// 自动迁移模型
	err = db.AutoMigrate(&models.KnowledgeQA{})
	suite.Require().NoError(err)

	suite.db = db

	// 创建模拟服务
	suite.mockService = new(MockAIService)

	// 创建处理器
	suite.handler = &Handler{
		service: suite.mockService,
	}

	// 设置路由
	suite.setupRouter()
}

// SetupTest 每个测试前的设置
func (suite *AIAPIIntegrationTestSuite) SetupTest() {
	// 清理数据库
	suite.db.Exec("DELETE FROM knowledge_qas")

	// 重置mock
	suite.mockService.ExpectedCalls = nil
	suite.mockService.Calls = nil
}

// setupRouter 设置测试路由
func (suite *AIAPIIntegrationTestSuite) setupRouter() {
	suite.router = gin.New()

	// 添加认证中间件模拟
	suite.router.Use(func(c *gin.Context) {
		// 模拟用户认证信息
		c.Set("user_id", "test-user-id")
		c.Set("team_id", "test-team-id")
		c.Next()
	})

	// 注册AI路由
	api := suite.router.Group("/api")
	suite.handler.RegisterRoutes(api)
}

// TestChatAPI 测试AI聊天API
func (suite *AIAPIIntegrationTestSuite) TestChatAPI() {
	// 准备请求数据
	chatRequest := ChatRequest{
		Question: "什么是CDK-Office？",
		Context: map[string]interface{}{
			"source":     "web",
			"session_id": "test-session-123",
		},
	}

	// 设置mock期望
	expectedResponse := &ChatResponse{
		Answer: "CDK-Office是一个现代化的办公协作平台，提供文档管理、审批流程、合同管理等功能。",
		Sources: []DocumentInfo{
			{
				ID:      "doc-1",
				Name:    "CDK-Office用户手册.pdf",
				Snippet: "CDK-Office是一个基于云的办公自动化解决方案...",
				Score:   0.95,
			},
		},
		Confidence: 0.92,
		MessageID:  "msg-123",
		CreatedAt:  time.Now(),
	}

	suite.mockService.On("Chat", mock.Anything, "test-user-id", "test-team-id", mock.MatchedBy(func(req *ChatRequest) bool {
		return req.Question == chatRequest.Question
	})).Return(expectedResponse, nil)

	// 序列化请求数据
	jsonData, err := json.Marshal(chatRequest)
	suite.Require().NoError(err)

	// 创建HTTP请求
	req, err := http.NewRequest("POST", "/api/ai/chat", bytes.NewBuffer(jsonData))
	suite.Require().NoError(err)
	req.Header.Set("Content-Type", "application/json")

	// 发送请求
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	// 验证响应状态码
	assert.Equal(suite.T(), http.StatusOK, w.Code)

	// 验证响应头
	assert.Equal(suite.T(), "application/json; charset=utf-8", w.Header().Get("Content-Type"))

	// 验证响应JSON结构
	var response ChatResponse
	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)

	// 验证响应内容
	assert.Equal(suite.T(), expectedResponse.Answer, response.Answer)
	assert.Equal(suite.T(), expectedResponse.Confidence, response.Confidence)
	assert.Equal(suite.T(), expectedResponse.MessageID, response.MessageID)
	assert.Len(suite.T(), response.Sources, 1)
	assert.Equal(suite.T(), "doc-1", response.Sources[0].ID)
	assert.Equal(suite.T(), "CDK-Office用户手册.pdf", response.Sources[0].Name)

	// 验证mock被正确调用
	suite.mockService.AssertExpectations(suite.T())
}

// TestChatAPIInvalidRequest 测试无效请求
func (suite *AIAPIIntegrationTestSuite) TestChatAPIInvalidRequest() {
	// 测试空请求体
	req, err := http.NewRequest("POST", "/api/ai/chat", bytes.NewBuffer([]byte("")))
	suite.Require().NoError(err)
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusBadRequest, w.Code)

	var errorResponse models.ErrorResponse
	err = json.Unmarshal(w.Body.Bytes(), &errorResponse)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "INVALID_REQUEST", errorResponse.Code)
	assert.Contains(suite.T(), errorResponse.Message, "请求参数错误")

	// 测试缺少question字段
	invalidRequest := map[string]interface{}{
		"context": map[string]interface{}{
			"source": "web",
		},
	}

	jsonData, err := json.Marshal(invalidRequest)
	suite.Require().NoError(err)

	req, err = http.NewRequest("POST", "/api/ai/chat", bytes.NewBuffer(jsonData))
	suite.Require().NoError(err)
	req.Header.Set("Content-Type", "application/json")

	w = httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusBadRequest, w.Code)
}

// TestChatAPIServiceError 测试服务错误
func (suite *AIAPIIntegrationTestSuite) TestChatAPIServiceError() {
	// 准备请求数据
	chatRequest := ChatRequest{
		Question: "测试服务错误",
	}

	// 设置mock返回错误
	suite.mockService.On("Chat", mock.Anything, "test-user-id", "test-team-id", mock.Anything).Return(
		(*ChatResponse)(nil),
		assert.AnError,
	)

	jsonData, err := json.Marshal(chatRequest)
	suite.Require().NoError(err)

	req, err := http.NewRequest("POST", "/api/ai/chat", bytes.NewBuffer(jsonData))
	suite.Require().NoError(err)
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	// 验证错误响应
	assert.Equal(suite.T(), http.StatusInternalServerError, w.Code)

	var errorResponse models.ErrorResponse
	err = json.Unmarshal(w.Body.Bytes(), &errorResponse)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "AI_SERVICE_ERROR", errorResponse.Code)
	assert.Contains(suite.T(), errorResponse.Message, "AI服务出错")

	suite.mockService.AssertExpectations(suite.T())
}

// TestChatHistoryAPI 测试聊天历史API
func (suite *AIAPIIntegrationTestSuite) TestChatHistoryAPI() {
	// 准备历史数据
	expectedHistory := []*models.KnowledgeQA{
		{
			ID:         uuid.New().String(),
			UserID:     "test-user-id",
			TeamID:     "test-team-id",
			Question:   "什么是CDK-Office？",
			Answer:     "CDK-Office是一个办公协作平台",
			Confidence: 0.92,
			MessageID:  "msg-1",
			AIProvider: "dify",
			CreatedAt:  time.Now(),
		},
		{
			ID:         uuid.New().String(),
			UserID:     "test-user-id",
			TeamID:     "test-team-id",
			Question:   "如何使用员工管理功能？",
			Answer:     "员工管理功能位于主菜单中...",
			Confidence: 0.88,
			MessageID:  "msg-2",
			AIProvider: "dify",
			CreatedAt:  time.Now().Add(-1 * time.Hour),
		},
	}

	// 设置mock期望
	suite.mockService.On("GetChatHistory", mock.Anything, "test-user-id", "test-team-id", 20, 0).Return(
		expectedHistory,
		int64(2),
		nil,
	)

	// 创建请求
	req, err := http.NewRequest("GET", "/api/ai/chat/history?page=1&size=20", nil)
	suite.Require().NoError(err)

	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	// 验证响应
	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response ChatHistoryResponse
	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)

	// 验证响应结构
	assert.Len(suite.T(), response.Data, 2)
	assert.Equal(suite.T(), 1, response.Pagination.Page)
	assert.Equal(suite.T(), 20, response.Pagination.Size)
	assert.Equal(suite.T(), int64(2), response.Pagination.Total)

	// 验证历史数据
	assert.Equal(suite.T(), expectedHistory[0].Question, response.Data[0].Question)
	assert.Equal(suite.T(), expectedHistory[0].Answer, response.Data[0].Answer)

	suite.mockService.AssertExpectations(suite.T())
}

// TestUpdateFeedbackAPI 测试更新反馈API
func (suite *AIAPIIntegrationTestSuite) TestUpdateFeedbackAPI() {
	messageID := "msg-123"
	feedbackRequest := FeedbackRequest{
		Feedback: "这个回答很有帮助",
	}

	// 设置mock期望
	suite.mockService.On("UpdateFeedback", mock.Anything, "test-user-id", messageID, feedbackRequest.Feedback).Return(nil)

	// 序列化请求数据
	jsonData, err := json.Marshal(feedbackRequest)
	suite.Require().NoError(err)

	// 创建请求
	req, err := http.NewRequest("PATCH", "/api/ai/chat/"+messageID+"/feedback", bytes.NewBuffer(jsonData))
	suite.Require().NoError(err)
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	// 验证响应
	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response models.SuccessResponse
	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "SUCCESS", response.Code)
	assert.Equal(suite.T(), "反馈更新成功", response.Message)

	suite.mockService.AssertExpectations(suite.T())
}

// TestUpdateFeedbackAPINotFound 测试反馈API记录不存在
func (suite *AIAPIIntegrationTestSuite) TestUpdateFeedbackAPINotFound() {
	messageID := "nonexistent-msg"
	feedbackRequest := FeedbackRequest{
		Feedback: "测试反馈",
	}

	// 设置mock返回"记录不存在"错误
	suite.mockService.On("UpdateFeedback", mock.Anything, "test-user-id", messageID, feedbackRequest.Feedback).Return(
		fmt.Errorf("knowledge QA record not found"),
	)

	jsonData, err := json.Marshal(feedbackRequest)
	suite.Require().NoError(err)

	req, err := http.NewRequest("PATCH", "/api/ai/chat/"+messageID+"/feedback", bytes.NewBuffer(jsonData))
	suite.Require().NoError(err)
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	// 验证404响应
	assert.Equal(suite.T(), http.StatusNotFound, w.Code)

	var errorResponse models.ErrorResponse
	err = json.Unmarshal(w.Body.Bytes(), &errorResponse)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "RECORD_NOT_FOUND", errorResponse.Code)

	suite.mockService.AssertExpectations(suite.T())
}

// TestChatAPIWithoutAuth 测试未认证请求
func (suite *AIAPIIntegrationTestSuite) TestChatAPIWithoutAuth() {
	// 创建没有认证中间件的路由
	router := gin.New()
	handler := &Handler{service: suite.mockService}
	api := router.Group("/api")
	handler.RegisterRoutes(api)

	chatRequest := ChatRequest{
		Question: "测试问题",
	}

	jsonData, err := json.Marshal(chatRequest)
	suite.Require().NoError(err)

	req, err := http.NewRequest("POST", "/api/ai/chat", bytes.NewBuffer(jsonData))
	suite.Require().NoError(err)
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// 验证未授权响应
	assert.Equal(suite.T(), http.StatusUnauthorized, w.Code)

	var errorResponse models.ErrorResponse
	err = json.Unmarshal(w.Body.Bytes(), &errorResponse)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "UNAUTHORIZED", errorResponse.Code)
}

// TestConcurrentRequests 测试并发请求
func (suite *AIAPIIntegrationTestSuite) TestConcurrentRequests() {
	// 设置mock期望多次调用
	suite.mockService.On("Chat", mock.Anything, "test-user-id", "test-team-id", mock.Anything).Return(
		&ChatResponse{
			Answer:     "并发测试回答",
			Confidence: 0.9,
			MessageID:  "concurrent-msg",
			CreatedAt:  time.Now(),
		},
		nil,
	).Times(5)

	// 并发发送请求
	done := make(chan bool, 5)
	for i := 0; i < 5; i++ {
		go func(index int) {
			chatRequest := ChatRequest{
				Question: fmt.Sprintf("并发测试问题 %d", index),
			}

			jsonData, err := json.Marshal(chatRequest)
			assert.NoError(suite.T(), err)

			req, err := http.NewRequest("POST", "/api/ai/chat", bytes.NewBuffer(jsonData))
			assert.NoError(suite.T(), err)
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			suite.router.ServeHTTP(w, req)

			assert.Equal(suite.T(), http.StatusOK, w.Code)
			done <- true
		}(i)
	}

	// 等待所有请求完成
	for i := 0; i < 5; i++ {
		<-done
	}

	suite.mockService.AssertExpectations(suite.T())
}

// TestAPIResponseHeaders 测试API响应头
func (suite *AIAPIIntegrationTestSuite) TestAPIResponseHeaders() {
	// 设置mock
	suite.mockService.On("Chat", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(
		&ChatResponse{
			Answer:    "测试回答",
			MessageID: "test-msg",
			CreatedAt: time.Now(),
		},
		nil,
	)

	chatRequest := ChatRequest{Question: "测试问题"}
	jsonData, _ := json.Marshal(chatRequest)

	req, _ := http.NewRequest("POST", "/api/ai/chat", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	// 验证响应头
	assert.Equal(suite.T(), "application/json; charset=utf-8", w.Header().Get("Content-Type"))
	assert.NotEmpty(suite.T(), w.Header().Get("Date"))
}

// TestAPIErrorHandling 测试API错误处理
func (suite *AIAPIIntegrationTestSuite) TestAPIErrorHandling() {
	testCases := []struct {
		name           string
		requestBody    interface{}
		expectedStatus int
		expectedCode   string
	}{
		{
			name:           "空请求体",
			requestBody:    "",
			expectedStatus: http.StatusBadRequest,
			expectedCode:   "INVALID_REQUEST",
		},
		{
			name:           "无效JSON",
			requestBody:    "{invalid json}",
			expectedStatus: http.StatusBadRequest,
			expectedCode:   "INVALID_REQUEST",
		},
		{
			name: "缺少必需字段",
			requestBody: map[string]interface{}{
				"context": map[string]interface{}{},
			},
			expectedStatus: http.StatusBadRequest,
			expectedCode:   "INVALID_REQUEST",
		},
	}

	for _, tc := range testCases {
		suite.T().Run(tc.name, func(t *testing.T) {
			var jsonData []byte
			var err error

			if str, ok := tc.requestBody.(string); ok {
				jsonData = []byte(str)
			} else {
				jsonData, err = json.Marshal(tc.requestBody)
				assert.NoError(t, err)
			}

			req, err := http.NewRequest("POST", "/api/ai/chat", bytes.NewBuffer(jsonData))
			assert.NoError(t, err)
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			suite.router.ServeHTTP(w, req)

			assert.Equal(t, tc.expectedStatus, w.Code)

			if tc.expectedCode != "" {
				var errorResponse models.ErrorResponse
				err = json.Unmarshal(w.Body.Bytes(), &errorResponse)
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedCode, errorResponse.Code)
			}
		})
	}
}

// TestJSONResponseValidation 测试JSON响应验证
func (suite *AIAPIIntegrationTestSuite) TestJSONResponseValidation() {
	// 设置复杂的mock响应
	complexResponse := &ChatResponse{
		Answer: "这是一个详细的回答，包含了多个方面的信息。",
		Sources: []DocumentInfo{
			{
				ID:      "doc-1",
				Name:    "文档1.pdf",
				Snippet: "相关内容片段1",
				Score:   0.95,
			},
			{
				ID:      "doc-2",
				Name:    "文档2.docx",
				Snippet: "相关内容片段2",
				Score:   0.87,
			},
		},
		Confidence: 0.92,
		MessageID:  "complex-msg-123",
		CreatedAt:  time.Now(),
	}

	suite.mockService.On("Chat", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(
		complexResponse,
		nil,
	)

	chatRequest := ChatRequest{
		Question: "复杂测试问题",
		Context: map[string]interface{}{
			"source":     "web",
			"session_id": "session-123",
			"user_agent": "test-agent",
			"timestamp":  time.Now().Unix(),
		},
	}

	jsonData, err := json.Marshal(chatRequest)
	suite.Require().NoError(err)

	req, err := http.NewRequest("POST", "/api/ai/chat", bytes.NewBuffer(jsonData))
	suite.Require().NoError(err)
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	// 验证JSON格式正确
	var response ChatResponse
	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)

	// 验证响应结构完整性
	assert.Equal(suite.T(), complexResponse.Answer, response.Answer)
	assert.Equal(suite.T(), complexResponse.Confidence, response.Confidence)
	assert.Equal(suite.T(), complexResponse.MessageID, response.MessageID)
	assert.Len(suite.T(), response.Sources, 2)

	// 验证嵌套结构
	for i, source := range response.Sources {
		assert.Equal(suite.T(), complexResponse.Sources[i].ID, source.ID)
		assert.Equal(suite.T(), complexResponse.Sources[i].Name, source.Name)
		assert.Equal(suite.T(), complexResponse.Sources[i].Snippet, source.Snippet)
		assert.Equal(suite.T(), complexResponse.Sources[i].Score, source.Score)
	}

	suite.mockService.AssertExpectations(suite.T())
}

// 运行集成测试套件
func TestAIAPIIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(AIAPIIntegrationTestSuite))
}

// 基准测试API性能
func BenchmarkChatAPI(b *testing.B) {
	gin.SetMode(gin.TestMode)

	// 创建测试路由
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("user_id", "test-user-id")
		c.Set("team_id", "test-team-id")
		c.Next()
	})

	mockService := new(MockAIService)
	handler := &Handler{service: mockService}
	api := router.Group("/api")
	handler.RegisterRoutes(api)

	// 设置mock
	mockService.On("Chat", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(
		&ChatResponse{
			Answer:    "基准测试回答",
			MessageID: "benchmark-msg",
			CreatedAt: time.Now(),
		},
		nil,
	)

	chatRequest := ChatRequest{Question: "基准测试问题"}
	jsonData, _ := json.Marshal(chatRequest)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req, _ := http.NewRequest("POST", "/api/ai/chat", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			b.Errorf("Expected status 200, got %d", w.Code)
		}
	}
}
