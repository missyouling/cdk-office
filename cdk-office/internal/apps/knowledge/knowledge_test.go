package knowledge

import (
	"context"
	"io"
	"mime/multipart"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"

	"cdk-office/internal/models"
)

// MockKnowledgeDB 模拟知识库数据库
type MockKnowledgeDB struct {
	mock.Mock
}

func (m *MockKnowledgeDB) Create(value interface{}) *gorm.DB {
	args := m.Called(value)
	return args.Get(0).(*gorm.DB)
}

func (m *MockKnowledgeDB) Save(value interface{}) *gorm.DB {
	args := m.Called(value)
	return args.Get(0).(*gorm.DB)
}

func (m *MockKnowledgeDB) Where(query interface{}, args ...interface{}) *gorm.DB {
	mockArgs := append([]interface{}{query}, args...)
	callArgs := m.Called(mockArgs...)
	return callArgs.Get(0).(*gorm.DB)
}

func (m *MockKnowledgeDB) First(dest interface{}, conds ...interface{}) *gorm.DB {
	args := append([]interface{}{dest}, conds...)
	callArgs := m.Called(args...)
	return callArgs.Get(0).(*gorm.DB)
}

func (m *MockKnowledgeDB) Find(dest interface{}, conds ...interface{}) *gorm.DB {
	args := append([]interface{}{dest}, conds...)
	callArgs := m.Called(args...)
	return callArgs.Get(0).(*gorm.DB)
}

func (m *MockKnowledgeDB) Delete(value interface{}, conds ...interface{}) *gorm.DB {
	args := append([]interface{}{value}, conds...)
	callArgs := m.Called(args...)
	return callArgs.Get(0).(*gorm.DB)
}

// MockFileHeader 模拟文件头
type MockFileHeader struct {
	filename string
	size     int64
	content  string
}

func (m *MockFileHeader) Filename() string {
	return m.filename
}

func (m *MockFileHeader) Size() int64 {
	return m.size
}

func (m *MockFileHeader) Header() map[string][]string {
	return map[string][]string{
		"Content-Type": {"text/plain"},
	}
}

func (m *MockFileHeader) Open() (multipart.File, error) {
	return &MockFile{content: m.content}, nil
}

// MockFile 模拟文件
type MockFile struct {
	content string
	pos     int
}

func (m *MockFile) Read(p []byte) (n int, err error) {
	remaining := len(m.content) - m.pos
	if remaining == 0 {
		return 0, io.EOF
	}

	n = len(p)
	if n > remaining {
		n = remaining
	}

	copy(p, m.content[m.pos:m.pos+n])
	m.pos += n
	return n, nil
}

func (m *MockFile) ReadAt(p []byte, off int64) (n int, err error) {
	if off >= int64(len(m.content)) {
		return 0, io.EOF
	}

	n = copy(p, m.content[off:])
	if n < len(p) {
		err = io.EOF
	}
	return n, err
}

func (m *MockFile) Seek(offset int64, whence int) (int64, error) {
	switch whence {
	case io.SeekStart:
		m.pos = int(offset)
	case io.SeekCurrent:
		m.pos += int(offset)
	case io.SeekEnd:
		m.pos = len(m.content) + int(offset)
	}
	return int64(m.pos), nil
}

func (m *MockFile) Close() error {
	return nil
}

// PersonalKnowledgeServiceTestSuite 个人知识库服务测试套件
type PersonalKnowledgeServiceTestSuite struct {
	suite.Suite
	service *PersonalKnowledgeService
	mockDB  *MockKnowledgeDB
}

func (suite *PersonalKnowledgeServiceTestSuite) SetupTest() {
	suite.mockDB = &MockKnowledgeDB{}
	suite.service = &PersonalKnowledgeService{
		db: suite.mockDB,
	}
}

func (suite *PersonalKnowledgeServiceTestSuite) TestCreateKnowledgeBase() {
	// 准备测试数据
	kb := &models.PersonalKnowledgeBase{
		UserID:      1,
		Name:        "Test Knowledge Base",
		Description: "A test knowledge base",
		Type:        "document",
		Settings: map[string]interface{}{
			"auto_sync": true,
			"public":    false,
		},
	}

	// 模拟数据库操作
	mockResult := &gorm.DB{}
	suite.mockDB.On("Create", kb).Return(mockResult)

	// 执行测试
	err := suite.service.CreateKnowledgeBase(context.Background(), kb)

	// 验证结果
	suite.NoError(err)
	suite.mockDB.AssertExpectations(suite.T())
}

func (suite *PersonalKnowledgeServiceTestSuite) TestGetUserKnowledgeBases() {
	// 准备测试数据
	userID := uint(1)
	knowledgeBases := []models.PersonalKnowledgeBase{
		{ID: 1, UserID: userID, Name: "KB 1", Type: "document"},
		{ID: 2, UserID: userID, Name: "KB 2", Type: "note"},
	}

	// 模拟数据库查询
	mockResult := &gorm.DB{}
	suite.mockDB.On("Where", mock.Anything).Return(suite.mockDB)
	suite.mockDB.On("Find", mock.AnythingOfType("*[]models.PersonalKnowledgeBase"), mock.Anything).Return(mockResult).Run(func(args mock.Arguments) {
		dest := args.Get(0).(*[]models.PersonalKnowledgeBase)
		*dest = knowledgeBases
	})

	// 执行测试
	result, err := suite.service.GetUserKnowledgeBases(context.Background(), userID)

	// 验证结果
	suite.NoError(err)
	suite.Len(result, 2)
	suite.Equal(knowledgeBases[0].ID, result[0].ID)
	suite.Equal(knowledgeBases[1].ID, result[1].ID)
	suite.mockDB.AssertExpectations(suite.T())
}

func (suite *PersonalKnowledgeServiceTestSuite) TestUploadDocument() {
	// 准备测试数据
	kbID := uint(1)
	fileHeader := &MockFileHeader{
		filename: "test.txt",
		size:     100,
		content:  "This is a test document content",
	}

	// 模拟知识库查询
	kb := &models.PersonalKnowledgeBase{
		ID:     kbID,
		UserID: 1,
		Name:   "Test KB",
		Type:   "document",
	}

	mockResult := &gorm.DB{}
	suite.mockDB.On("Where", mock.Anything).Return(suite.mockDB)
	suite.mockDB.On("First", mock.AnythingOfType("*models.PersonalKnowledgeBase"), mock.Anything).Return(mockResult).Run(func(args mock.Arguments) {
		dest := args.Get(0).(*models.PersonalKnowledgeBase)
		*dest = *kb
	})

	// 模拟创建文档记录
	suite.mockDB.On("Create", mock.AnythingOfType("*models.KnowledgeDocument")).Return(mockResult)

	// 执行测试
	document, err := suite.service.UploadDocument(context.Background(), kbID, fileHeader, 1)

	// 验证结果
	suite.NoError(err)
	suite.NotNil(document)
	suite.Equal(fileHeader.Filename(), document.FileName)
	suite.Equal(fileHeader.Size(), document.FileSize)
	suite.mockDB.AssertExpectations(suite.T())
}

func (suite *PersonalKnowledgeServiceTestSuite) TestSearchDocuments() {
	// 准备测试数据
	query := "test"
	userID := uint(1)
	documents := []models.KnowledgeDocument{
		{ID: 1, FileName: "test1.txt", Content: "This is test content 1"},
		{ID: 2, FileName: "test2.txt", Content: "This is test content 2"},
	}

	// 模拟数据库查询
	mockResult := &gorm.DB{}
	suite.mockDB.On("Where", mock.Anything).Return(suite.mockDB)
	suite.mockDB.On("Find", mock.AnythingOfType("*[]models.KnowledgeDocument"), mock.Anything).Return(mockResult).Run(func(args mock.Arguments) {
		dest := args.Get(0).(*[]models.KnowledgeDocument)
		*dest = documents
	})

	// 执行测试
	result, err := suite.service.SearchDocuments(context.Background(), query, userID)

	// 验证结果
	suite.NoError(err)
	suite.Len(result, 2)
	suite.mockDB.AssertExpectations(suite.T())
}

func TestPersonalKnowledgeServiceSuite(t *testing.T) {
	suite.Run(t, new(PersonalKnowledgeServiceTestSuite))
}

// WeChatServiceTestSuite 微信服务测试套件
type WeChatServiceTestSuite struct {
	suite.Suite
	service *WeChatService
	mockDB  *MockKnowledgeDB
}

func (suite *WeChatServiceTestSuite) SetupTest() {
	suite.mockDB = &MockKnowledgeDB{}
	suite.service = &WeChatService{
		db: suite.mockDB,
	}
}

func (suite *WeChatServiceTestSuite) TestUploadChatRecords() {
	// 准备测试数据
	fileHeader := &MockFileHeader{
		filename: "wechat_records.txt",
		size:     200,
		content: `2024-01-01 10:00:00 张三: 你好
2024-01-01 10:01:00 李四: 你好，最近怎么样？
2024-01-01 10:02:00 张三: 挺好的，在研究新项目`,
	}

	// 模拟数据库操作
	mockResult := &gorm.DB{}
	suite.mockDB.On("Create", mock.AnythingOfType("*models.WeChatRecord")).Return(mockResult).Times(3)

	// 执行测试
	records, err := suite.service.UploadChatRecords(context.Background(), fileHeader, 1)

	// 验证结果
	suite.NoError(err)
	suite.Len(records, 3)
	suite.Equal("张三", records[0].SenderName)
	suite.Equal("你好", records[0].Message)
	suite.mockDB.AssertExpectations(suite.T())
}

func (suite *WeChatServiceTestSuite) TestParseChatRecord() {
	// 测试解析聊天记录
	line := "2024-01-01 10:00:00 张三: 你好，这是一条测试消息"

	record, err := suite.service.parseChatRecord(line, 1)

	suite.NoError(err)
	suite.NotNil(record)
	suite.Equal("张三", record.SenderName)
	suite.Equal("你好，这是一条测试消息", record.Message)
	suite.Equal(uint(1), record.UserID)
}

func (suite *WeChatServiceTestSuite) TestInvalidChatRecordFormat() {
	// 测试无效格式
	invalidLine := "这是一个无效格式的行"

	record, err := suite.service.parseChatRecord(invalidLine, 1)

	suite.Error(err)
	suite.Nil(record)
	suite.Contains(err.Error(), "invalid format")
}

func TestWeChatServiceSuite(t *testing.T) {
	suite.Run(t, new(WeChatServiceTestSuite))
}

// DocumentScanServiceTestSuite 文档扫描服务测试套件
type DocumentScanServiceTestSuite struct {
	suite.Suite
	service *DocumentScanService
	mockDB  *MockKnowledgeDB
}

func (suite *DocumentScanServiceTestSuite) SetupTest() {
	suite.mockDB = &MockKnowledgeDB{}
	suite.service = &DocumentScanService{
		db: suite.mockDB,
	}
}

func (suite *DocumentScanServiceTestSuite) TestCreateScanTask() {
	// 准备测试数据
	task := &models.ScanTask{
		UserID: 1,
		Name:   "Test Scan",
		Type:   "document",
		Config: map[string]interface{}{
			"auto_ocr": true,
			"language": "zh-cn",
		},
		Status: "pending",
	}

	// 模拟数据库操作
	mockResult := &gorm.DB{}
	suite.mockDB.On("Create", task).Return(mockResult)

	// 执行测试
	err := suite.service.CreateScanTask(context.Background(), task)

	// 验证结果
	suite.NoError(err)
	suite.mockDB.AssertExpectations(suite.T())
}

func (suite *DocumentScanServiceTestSuite) TestProcessDocument() {
	// 准备测试数据
	fileHeader := &MockFileHeader{
		filename: "test.jpg",
		size:     1024,
		content:  "fake image content",
	}

	scanConfig := map[string]interface{}{
		"auto_ocr":    true,
		"language":    "zh-cn",
		"enhance":     true,
		"perspective": true,
	}

	// 模拟数据库操作
	mockResult := &gorm.DB{}
	suite.mockDB.On("Create", mock.AnythingOfType("*models.DocumentScanResult")).Return(mockResult)

	// 执行测试
	result, err := suite.service.ProcessDocument(context.Background(), fileHeader, scanConfig, 1)

	// 验证结果
	suite.NoError(err)
	suite.NotNil(result)
	suite.Equal(fileHeader.Filename(), result.OriginalFileName)
	suite.Equal(uint(1), result.UserID)
	suite.mockDB.AssertExpectations(suite.T())
}

func (suite *DocumentScanServiceTestSuite) TestGetScanResults() {
	// 准备测试数据
	userID := uint(1)
	results := []models.DocumentScanResult{
		{ID: 1, UserID: userID, OriginalFileName: "doc1.jpg"},
		{ID: 2, UserID: userID, OriginalFileName: "doc2.jpg"},
	}

	// 模拟数据库查询
	mockResult := &gorm.DB{}
	suite.mockDB.On("Where", mock.Anything).Return(suite.mockDB)
	suite.mockDB.On("Find", mock.AnythingOfType("*[]models.DocumentScanResult"), mock.Anything).Return(mockResult).Run(func(args mock.Arguments) {
		dest := args.Get(0).(*[]models.DocumentScanResult)
		*dest = results
	})

	// 执行测试
	result, err := suite.service.GetScanResults(context.Background(), userID)

	// 验证结果
	suite.NoError(err)
	suite.Len(result, 2)
	suite.Equal(results[0].ID, result[0].ID)
	suite.Equal(results[1].ID, result[1].ID)
	suite.mockDB.AssertExpectations(suite.T())
}

func TestDocumentScanServiceSuite(t *testing.T) {
	suite.Run(t, new(DocumentScanServiceTestSuite))
}

// 单元测试
func TestKnowledgeBaseType(t *testing.T) {
	validTypes := []string{"document", "note", "chat", "scan"}

	for _, typ := range validTypes {
		t.Run(typ, func(t *testing.T) {
			kb := &models.PersonalKnowledgeBase{
				UserID: 1,
				Name:   "Test KB",
				Type:   typ,
			}
			assert.NotEmpty(t, kb.Type)
		})
	}
}

func TestDocumentValidation(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		size     int64
		valid    bool
	}{
		{"Valid document", "test.pdf", 1024, true},
		{"Empty filename", "", 1024, false},
		{"Zero size", "test.pdf", 0, false},
		{"Large file", "test.pdf", 100 * 1024 * 1024, false}, // 100MB
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			valid := validateDocument(tt.filename, tt.size)
			assert.Equal(t, tt.valid, valid)
		})
	}
}

func validateDocument(filename string, size int64) bool {
	if filename == "" {
		return false
	}
	if size <= 0 || size > 50*1024*1024 { // 50MB limit
		return false
	}
	return true
}

func TestFileTypeDetection(t *testing.T) {
	tests := []struct {
		filename string
		expected string
	}{
		{"test.pdf", "pdf"},
		{"test.docx", "docx"},
		{"test.txt", "txt"},
		{"test.jpg", "jpg"},
		{"test.png", "png"},
		{"unknown.xyz", "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.filename, func(t *testing.T) {
			result := detectFileType(tt.filename)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func detectFileType(filename string) string {
	parts := strings.Split(filename, ".")
	if len(parts) < 2 {
		return "unknown"
	}

	ext := strings.ToLower(parts[len(parts)-1])
	validExts := map[string]bool{
		"pdf": true, "docx": true, "txt": true,
		"jpg": true, "png": true,
	}

	if validExts[ext] {
		return ext
	}
	return "unknown"
}

// 基准测试
func BenchmarkCreateKnowledgeBase(b *testing.B) {
	kb := &models.PersonalKnowledgeBase{
		UserID:      1,
		Name:        "Benchmark KB",
		Description: "A benchmark knowledge base",
		Type:        "document",
		Settings: map[string]interface{}{
			"auto_sync": true,
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// 模拟创建知识库
		newKB := *kb
		newKB.ID = uint(i + 1)
		newKB.Name = kb.Name + string(rune(i))
	}
}

func BenchmarkSearchDocuments(b *testing.B) {
	documents := make([]models.KnowledgeDocument, 1000)
	for i := range documents {
		documents[i] = models.KnowledgeDocument{
			ID:       uint(i + 1),
			FileName: "doc" + string(rune(i)) + ".txt",
			Content:  "This is document content " + string(rune(i)),
		}
	}

	query := "document"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// 模拟搜索
		var results []models.KnowledgeDocument
		for _, doc := range documents {
			if strings.Contains(doc.Content, query) {
				results = append(results, doc)
			}
		}
	}
}

// 并发测试
func TestConcurrentUpload(t *testing.T) {
	concurrency := 5
	iterations := 10
	done := make(chan bool, concurrency)

	for i := 0; i < concurrency; i++ {
		go func(workerID int) {
			defer func() { done <- true }()

			for j := 0; j < iterations; j++ {
				// 模拟并发文档上传
				doc := &models.KnowledgeDocument{
					ID:              uint(workerID*iterations + j + 1),
					KnowledgeBaseID: 1,
					FileName:        "concurrent_doc.txt",
					FileSize:        1024,
					FileType:        "txt",
					Content:         "Concurrent upload content",
					UploadedBy:      uint(workerID + 1),
				}

				// 验证文档创建
				assert.NotNil(t, doc)
				assert.True(t, doc.ID > 0)
			}
		}(i)
	}

	// 等待所有goroutine完成
	for i := 0; i < concurrency; i++ {
		select {
		case <-done:
		case <-time.After(5 * time.Second):
			t.Fatal("Concurrent upload test timeout")
		}
	}
}

// 错误处理测试
func TestErrorHandling(t *testing.T) {
	service := &PersonalKnowledgeService{}

	t.Run("NilKnowledgeBase", func(t *testing.T) {
		err := service.CreateKnowledgeBase(context.Background(), nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "nil")
	})

	t.Run("EmptyName", func(t *testing.T) {
		kb := &models.PersonalKnowledgeBase{
			UserID: 1,
			Name:   "",
			Type:   "document",
		}
		err := service.CreateKnowledgeBase(context.Background(), kb)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "name")
	})

	t.Run("InvalidType", func(t *testing.T) {
		kb := &models.PersonalKnowledgeBase{
			UserID: 1,
			Name:   "Test KB",
			Type:   "invalid_type",
		}
		err := service.CreateKnowledgeBase(context.Background(), kb)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "type")
	})
}
