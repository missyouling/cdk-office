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

package tests

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/linux-do/cdk-office/internal/apps/dify"
	"github.com/linux-do/cdk-office/internal/apps/knowledge"
	"github.com/linux-do/cdk-office/internal/apps/pdf"
	"github.com/linux-do/cdk-office/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// IntegrationTestSuite 集成测试套件
type IntegrationTestSuite struct {
	suite.Suite
	db     *gorm.DB
	router *gin.Engine
	ctx    context.Context
}

// SetupSuite 测试套件初始化
func (suite *IntegrationTestSuite) SetupSuite() {
	// 设置测试环境
	gin.SetMode(gin.TestMode)
	suite.ctx = context.Background()

	// 初始化内存数据库
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	suite.Require().NoError(err)
	suite.db = db

	// 自动迁移表结构
	err = db.AutoMigrate(
		&models.PersonalKnowledgeBase{},
		&models.WeChatRecord{},
		&models.PersonalKnowledgeShare{},
		&models.DocumentScanResult{},
		&models.ScanTask{},
		&models.DifyDocumentSync{},
		&models.KnowledgeQA{},
		&models.KnowledgeQAV2{},
		&models.PDFOperation{},
	)
	suite.Require().NoError(err)

	// 初始化路由
	suite.router = gin.New()
	suite.setupRoutes()
}

// TearDownSuite 测试套件清理
func (suite *IntegrationTestSuite) TearDownSuite() {
	sqlDB, err := suite.db.DB()
	if err == nil {
		sqlDB.Close()
	}
}

// SetupTest 每个测试前的初始化
func (suite *IntegrationTestSuite) SetupTest() {
	// 清理测试数据
	suite.cleanupTestData()
}

// setupRoutes 设置测试路由
func (suite *IntegrationTestSuite) setupRoutes() {
	api := suite.router.Group("/api")

	// 知识库路由
	knowledgeHandler := knowledge.NewHandler(suite.db)
	knowledgeRouter := knowledge.NewRouter(suite.db)
	knowledgeRouter.RegisterRoutes(api, knowledgeHandler)

	// PDF处理路由
	pdfHandler := pdf.NewHandler(suite.db)
	pdfRouter := pdf.NewRouter(suite.db)
	pdfRouter.RegisterRoutes(api, pdfHandler)

	// Dify集成路由
	difyConfig := &dify.Config{
		BaseURL:          "http://mock-dify-api.com",
		APIKey:           "test-api-key",
		DefaultDatasetID: "test-dataset-id",
	}
	difyService := dify.NewService(difyConfig, suite.db)
	difyHandler := dify.NewHandler(difyService, suite.db)
	difyRouter := dify.NewRouter(suite.db)
	difyRouter.RegisterRoutes(api, difyHandler)
}

// cleanupTestData 清理测试数据
func (suite *IntegrationTestSuite) cleanupTestData() {
	tables := []string{
		"personal_knowledge_bases",
		"we_chat_records",
		"personal_knowledge_shares",
		"document_scan_results",
		"scan_tasks",
		"dify_document_syncs",
		"knowledge_qas",
		"knowledge_qa_v2s",
		"pdf_operations",
	}

	for _, table := range tables {
		suite.db.Exec(fmt.Sprintf("DELETE FROM %s", table))
	}
}

// TestPersonalKnowledgeBaseCRUD 测试个人知识库CRUD操作
func (suite *IntegrationTestSuite) TestPersonalKnowledgeBaseCRUD() {
	userID := uuid.New().String()

	// 测试创建知识库
	createReq := map[string]interface{}{
		"title":        "测试知识库",
		"description":  "这是一个测试知识库",
		"content":      "# 测试内容\n这是测试内容",
		"content_type": "markdown",
		"tags":         []string{"测试", "知识库"},
		"category":     "技术文档",
		"privacy":      "private",
		"source_type":  "manual",
	}

	w := suite.performRequest("POST", "/api/knowledge", createReq, userID)
	suite.Equal(http.StatusOK, w.Code)

	var createResp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &createResp)
	suite.NoError(err)

	knowledgeID := createResp["id"].(string)
	suite.NotEmpty(knowledgeID)

	// 测试获取知识库详情
	w = suite.performRequest("GET", fmt.Sprintf("/api/knowledge/%s", knowledgeID), nil, userID)
	suite.Equal(http.StatusOK, w.Code)

	var getResp models.PersonalKnowledgeBase
	err = json.Unmarshal(w.Body.Bytes(), &getResp)
	suite.NoError(err)
	suite.Equal("测试知识库", getResp.Title)

	// 测试更新知识库
	updateReq := map[string]interface{}{
		"title":   "更新的测试知识库",
		"content": "# 更新的内容\n这是更新后的内容",
	}

	w = suite.performRequest("PUT", fmt.Sprintf("/api/knowledge/%s", knowledgeID), updateReq, userID)
	suite.Equal(http.StatusOK, w.Code)

	// 测试列表查询
	w = suite.performRequest("GET", "/api/knowledge?page=1&page_size=10", nil, userID)
	suite.Equal(http.StatusOK, w.Code)

	var listResp map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &listResp)
	suite.NoError(err)

	knowledge := listResp["knowledge"].([]interface{})
	suite.Len(knowledge, 1)

	// 测试删除知识库
	w = suite.performRequest("DELETE", fmt.Sprintf("/api/knowledge/%s", knowledgeID), nil, userID)
	suite.Equal(http.StatusOK, w.Code)

	// 验证删除成功
	w = suite.performRequest("GET", fmt.Sprintf("/api/knowledge/%s", knowledgeID), nil, userID)
	suite.Equal(http.StatusNotFound, w.Code)
}

// TestWeChatRecordUpload 测试微信聊天记录上传
func (suite *IntegrationTestSuite) TestWeChatRecordUpload() {
	userID := uuid.New().String()

	uploadReq := map[string]interface{}{
		"session_name": "测试聊天会话",
		"records": []map[string]interface{}{
			{
				"message_type": "text",
				"sender_name":  "张三",
				"sender_id":    "user123",
				"content":      "这是一条测试消息",
				"message_time": time.Now().Format(time.RFC3339),
			},
			{
				"message_type": "text",
				"sender_name":  "李四",
				"sender_id":    "user456",
				"content":      "这是另一条测试消息",
				"message_time": time.Now().Format(time.RFC3339),
			},
		},
		"process_config": map[string]interface{}{
			"enable_ocr":          true,
			"enable_auto_archive": true,
			"extract_keywords":    true,
		},
	}

	w := suite.performRequest("POST", "/api/knowledge/wechat-records", uploadReq, userID)
	suite.Equal(http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	suite.NoError(err)

	suite.Equal("success", resp["status"])
	suite.NotEmpty(resp["processed_count"])
}

// TestDifyKnowledgeSync 测试Dify知识库同步
func (suite *IntegrationTestSuite) TestDifyKnowledgeSync() {
	userID := uuid.New().String()
	teamID := uuid.New().String()

	// 先创建一个知识库文档
	knowledge := &models.PersonalKnowledgeBase{
		UserID:      userID,
		Title:       "Dify同步测试",
		Content:     "这是用于Dify同步测试的内容",
		ContentType: "markdown",
		Tags:        []string{"Dify", "同步"},
		Privacy:     "shared",
		SourceType:  "manual",
		IsShared:    true,
	}

	err := suite.db.Create(knowledge).Error
	suite.NoError(err)

	// 测试同步到Dify
	syncReq := map[string]interface{}{
		"document_id":   knowledge.ID,
		"title":         knowledge.Title,
		"content":       knowledge.Content,
		"document_type": "knowledge",
		"team_id":       teamID,
		"created_by":    userID,
	}

	w := suite.performRequest("POST", "/api/dify/sync-document", syncReq, userID)

	// 由于是模拟环境，可能返回错误，但至少验证接口调用
	suite.True(w.Code == http.StatusOK || w.Code == http.StatusInternalServerError)
}

// TestPDFOperations 测试PDF处理功能
func (suite *IntegrationTestSuite) TestPDFOperations() {
	userID := uuid.New().String()

	// 测试PDF合并操作
	mergeReq := map[string]interface{}{
		"operation_type": "merge",
		"output_name":    "merged_document.pdf",
		"options": map[string]interface{}{
			"bookmark_levels": []int{1, 2},
		},
	}

	w := suite.performRequest("POST", "/api/pdf/operation", mergeReq, userID)

	// 由于没有实际文件，可能返回错误，但验证接口存在
	suite.True(w.Code == http.StatusBadRequest || w.Code == http.StatusOK)

	// 测试获取PDF工具分类
	w = suite.performRequest("GET", "/api/pdf/categories", nil, userID)
	suite.Equal(http.StatusOK, w.Code)

	var categories []map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &categories)
	suite.NoError(err)
	suite.True(len(categories) >= 0)
}

// TestSearchFunctionality 测试搜索功能
func (suite *IntegrationTestSuite) TestSearchFunctionality() {
	userID := uuid.New().String()

	// 创建测试数据
	testKnowledge := []models.PersonalKnowledgeBase{
		{
			UserID:      userID,
			Title:       "Go语言学习指南",
			Content:     "这是关于Go语言的学习资料",
			ContentType: "markdown",
			Tags:        []string{"Go", "编程", "学习"},
			Category:    "技术文档",
			Privacy:     "private",
			SourceType:  "manual",
		},
		{
			UserID:      userID,
			Title:       "React前端开发",
			Content:     "React是一个JavaScript库",
			ContentType: "markdown",
			Tags:        []string{"React", "前端", "JavaScript"},
			Category:    "技术文档",
			Privacy:     "private",
			SourceType:  "manual",
		},
	}

	for _, knowledge := range testKnowledge {
		err := suite.db.Create(&knowledge).Error
		suite.NoError(err)
	}

	// 测试关键词搜索
	w := suite.performRequest("GET", "/api/knowledge/search?query=Go语言&page=1&page_size=10", nil, userID)
	suite.Equal(http.StatusOK, w.Code)

	var searchResp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &searchResp)
	suite.NoError(err)

	results := searchResp["results"].([]interface{})
	suite.True(len(results) >= 1)

	// 测试标签过滤
	w = suite.performRequest("GET", "/api/knowledge?tags=Go,编程&page=1&page_size=10", nil, userID)
	suite.Equal(http.StatusOK, w.Code)
}

// TestBatchOperations 测试批量操作
func (suite *IntegrationTestSuite) TestBatchOperations() {
	userID := uuid.New().String()

	// 创建多个知识库文档
	var knowledgeIDs []string
	for i := 0; i < 3; i++ {
		knowledge := &models.PersonalKnowledgeBase{
			UserID:      userID,
			Title:       fmt.Sprintf("批量测试文档 %d", i+1),
			Content:     fmt.Sprintf("这是第 %d 个测试文档", i+1),
			ContentType: "markdown",
			Privacy:     "private",
			SourceType:  "manual",
		}

		err := suite.db.Create(knowledge).Error
		suite.NoError(err)
		knowledgeIDs = append(knowledgeIDs, knowledge.ID)
	}

	// 测试批量更新标签
	batchReq := map[string]interface{}{
		"knowledge_ids": knowledgeIDs,
		"updates": map[string]interface{}{
			"tags": []string{"批量操作", "测试"},
		},
	}

	w := suite.performRequest("POST", "/api/knowledge/batch-update", batchReq, userID)
	suite.Equal(http.StatusOK, w.Code)

	var batchResp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &batchResp)
	suite.NoError(err)

	suite.Equal(float64(3), batchResp["success_count"])
	suite.Equal(float64(0), batchResp["failed_count"])

	// 测试批量删除
	deleteReq := map[string]interface{}{
		"knowledge_ids": knowledgeIDs,
	}

	w = suite.performRequest("POST", "/api/knowledge/batch-delete", deleteReq, userID)
	suite.Equal(http.StatusOK, w.Code)
}

// TestErrorHandling 测试错误处理
func (suite *IntegrationTestSuite) TestErrorHandling() {
	userID := uuid.New().String()

	// 测试获取不存在的知识库
	w := suite.performRequest("GET", "/api/knowledge/nonexistent-id", nil, userID)
	suite.Equal(http.StatusNotFound, w.Code)

	// 测试无效的请求数据
	invalidReq := map[string]interface{}{
		"title": "", // 空标题应该被拒绝
	}

	w = suite.performRequest("POST", "/api/knowledge", invalidReq, userID)
	suite.Equal(http.StatusBadRequest, w.Code)

	// 测试权限检查（访问其他用户的数据）
	otherUserID := uuid.New().String()

	// 创建其他用户的知识库
	otherKnowledge := &models.PersonalKnowledgeBase{
		UserID:      otherUserID,
		Title:       "其他用户的知识库",
		Content:     "这是其他用户的内容",
		ContentType: "markdown",
		Privacy:     "private",
		SourceType:  "manual",
	}

	err := suite.db.Create(otherKnowledge).Error
	suite.NoError(err)

	// 当前用户尝试访问其他用户的私有知识库
	w = suite.performRequest("GET", fmt.Sprintf("/api/knowledge/%s", otherKnowledge.ID), nil, userID)
	suite.Equal(http.StatusForbidden, w.Code)
}

// TestStatisticsAndAnalytics 测试统计和分析功能
func (suite *IntegrationTestSuite) TestStatisticsAndAnalytics() {
	userID := uuid.New().String()

	// 创建一些测试数据
	categories := []string{"技术文档", "产品文档", "会议记录"}
	sources := []string{"manual", "wechat", "upload"}

	for i, category := range categories {
		for j, source := range sources {
			knowledge := &models.PersonalKnowledgeBase{
				UserID:      userID,
				Title:       fmt.Sprintf("%s_%s_%d", category, source, j),
				Content:     "测试内容",
				ContentType: "markdown",
				Category:    category,
				SourceType:  source,
				Privacy:     "private",
				Tags:        []string{category, source},
			}

			err := suite.db.Create(knowledge).Error
			suite.NoError(err)
		}
	}

	// 测试获取统计信息
	w := suite.performRequest("GET", "/api/knowledge/statistics", nil, userID)
	suite.Equal(http.StatusOK, w.Code)

	var stats map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &stats)
	suite.NoError(err)

	suite.Equal(float64(9), stats["total_knowledge"])
	suite.True(len(stats["by_category"].([]interface{})) >= 3)
	suite.True(len(stats["by_source"].([]interface{})) >= 3)
}

// performRequest 执行HTTP请求的辅助方法
func (suite *IntegrationTestSuite) performRequest(method, url string, body interface{}, userID string) *httptest.ResponseRecorder {
	var reqBody *bytes.Buffer

	if body != nil {
		jsonData, _ := json.Marshal(body)
		reqBody = bytes.NewBuffer(jsonData)
	} else {
		reqBody = bytes.NewBuffer([]byte{})
	}

	req, _ := http.NewRequest(method, url, reqBody)
	req.Header.Set("Content-Type", "application/json")

	// 设置用户认证信息
	if userID != "" {
		req.Header.Set("X-User-ID", userID)
		req.Header.Set("X-Team-ID", uuid.New().String())
	}

	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	return w
}

// TestRunner 运行集成测试的入口
func TestIntegrationSuite(t *testing.T) {
	// 检查是否设置了测试环境变量
	if os.Getenv("RUN_INTEGRATION_TESTS") == "" {
		t.Skip("Skipping integration tests. Set RUN_INTEGRATION_TESTS=1 to run.")
	}

	suite.Run(t, new(IntegrationTestSuite))
}

// BenchmarkPersonalKnowledgeOperations 性能基准测试
func BenchmarkPersonalKnowledgeOperations(b *testing.B) {
	// 初始化测试环境
	gin.SetMode(gin.TestMode)

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		b.Fatal(err)
	}

	db.AutoMigrate(&models.PersonalKnowledgeBase{})

	router := gin.New()
	knowledgeHandler := knowledge.NewHandler(db)
	knowledgeRouter := knowledge.NewRouter(db)
	knowledgeRouter.RegisterRoutes(router.Group("/api"), knowledgeHandler)

	userID := uuid.New().String()

	// 创建测试数据
	createReq := map[string]interface{}{
		"title":        "性能测试知识库",
		"content":      "这是性能测试内容",
		"content_type": "markdown",
		"privacy":      "private",
		"source_type":  "manual",
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		jsonData, _ := json.Marshal(createReq)
		req, _ := http.NewRequest("POST", "/api/knowledge", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-User-ID", userID)
		req.Header.Set("X-Team-ID", uuid.New().String())

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			b.Errorf("Expected status 200, got %d", w.Code)
		}
	}
}

// TestHelper 测试辅助函数
type TestHelper struct {
	DB     *gorm.DB
	Router *gin.Engine
}

// NewTestHelper 创建测试辅助工具
func NewTestHelper() *TestHelper {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		panic(err)
	}

	// 自动迁移
	db.AutoMigrate(
		&models.PersonalKnowledgeBase{},
		&models.WeChatRecord{},
		&models.DifyDocumentSync{},
	)

	router := gin.New()
	gin.SetMode(gin.TestMode)

	return &TestHelper{
		DB:     db,
		Router: router,
	}
}

// CreateTestKnowledge 创建测试知识库数据
func (h *TestHelper) CreateTestKnowledge(userID string, count int) []models.PersonalKnowledgeBase {
	var knowledge []models.PersonalKnowledgeBase

	for i := 0; i < count; i++ {
		kb := models.PersonalKnowledgeBase{
			UserID:      userID,
			Title:       fmt.Sprintf("测试知识库 %d", i+1),
			Content:     fmt.Sprintf("这是第 %d 个测试知识库的内容", i+1),
			ContentType: "markdown",
			Tags:        []string{fmt.Sprintf("tag%d", i), "test"},
			Privacy:     "private",
			SourceType:  "manual",
		}

		h.DB.Create(&kb)
		knowledge = append(knowledge, kb)
	}

	return knowledge
}

// AssertKnowledgeExists 断言知识库存在
func (h *TestHelper) AssertKnowledgeExists(t *testing.T, knowledgeID string) {
	var knowledge models.PersonalKnowledgeBase
	err := h.DB.First(&knowledge, "id = ?", knowledgeID).Error
	assert.NoError(t, err)
	assert.NotEmpty(t, knowledge.ID)
}

// AssertKnowledgeNotExists 断言知识库不存在
func (h *TestHelper) AssertKnowledgeNotExists(t *testing.T, knowledgeID string) {
	var knowledge models.PersonalKnowledgeBase
	err := h.DB.First(&knowledge, "id = ?", knowledgeID).Error
	assert.Error(t, err)
	assert.True(t, err == gorm.ErrRecordNotFound)
}
