package workflow

import (
	"context"
	"testing"

	"cdk-office/internal/dify/client"
	"cdk-office/internal/document/domain"
	"cdk-office/internal/document/service"
	"cdk-office/internal/shared/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockDifyClient is a mock implementation of the DifyClientInterface
type MockDifyClient struct {
	mock.Mock
}

func (m *MockDifyClient) CreateCompletionMessage(ctx context.Context, req *client.CompletionRequest) (*client.CompletionResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*client.CompletionResponse), args.Error(1)
}

func (m *MockDifyClient) CreateChatMessage(ctx context.Context, req *client.ChatRequest) (*client.ChatResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*client.ChatResponse), args.Error(1)
}

func (m *MockDifyClient) UploadFile(ctx context.Context, req *client.FileUploadRequest) (*client.FileUploadResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*client.FileUploadResponse), args.Error(1)
}

// MockDocumentService is a mock implementation of the DocumentServiceInterface
type MockDocumentService struct {
	mock.Mock
}

func (m *MockDocumentService) Upload(ctx context.Context, req *service.UploadRequest) (*domain.Document, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Document), args.Error(1)
}

func (m *MockDocumentService) GetDocument(ctx context.Context, docID string) (*domain.Document, error) {
	args := m.Called(ctx, docID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Document), args.Error(1)
}

func (m *MockDocumentService) UpdateDocument(ctx context.Context, docID string, req *service.UpdateRequest) error {
	args := m.Called(ctx, docID, req)
	return args.Error(0)
}

func (m *MockDocumentService) DeleteDocument(ctx context.Context, docID string) error {
	args := m.Called(ctx, docID)
	return args.Error(0)
}

func (m *MockDocumentService) GetDocumentVersions(ctx context.Context, docID string) ([]*domain.DocumentVersion, error) {
	args := m.Called(ctx, docID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.DocumentVersion), args.Error(1)
}

// MockContentExtractor is a mock implementation of the ContentExtractorInterface
type MockContentExtractor struct {
	mock.Mock
}

func (m *MockContentExtractor) ExtractContent(document *domain.Document) (string, error) {
	args := m.Called(document)
	return args.String(0), args.Error(1)
}

// MockOCRExtractor is a mock implementation of the OCRExtractorInterface
type MockOCRExtractor struct {
	mock.Mock
}

func (m *MockOCRExtractor) ExtractOCRContent(document *domain.Document) (string, error) {
	args := m.Called(document)
	return args.String(0), args.Error(1)
}

// MockClassifier is a mock implementation of the ClassifierInterface
type MockClassifier struct {
	mock.Mock
}

func (m *MockClassifier) ClassifyDocument(ctx context.Context, content string, document *domain.Document) (string, error) {
	args := m.Called(ctx, content, document)
	return args.String(0), args.Error(1)
}

// MockTagExtractor is a mock implementation of the TagExtractorInterface
type MockTagExtractor struct {
	mock.Mock
}

func (m *MockTagExtractor) ExtractTags(ctx context.Context, content string, document *domain.Document) ([]string, error) {
	args := m.Called(ctx, content, document)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]string), args.Error(1)
}

// MockSummarizer is a mock implementation of the SummarizerInterface
type MockSummarizer struct {
	mock.Mock
}

func (m *MockSummarizer) SummarizeDocument(ctx context.Context, content string, document *domain.Document) (string, error) {
	args := m.Called(ctx, content, document)
	return args.String(0), args.Error(1)
}

// MockKnowledgeBase is a mock implementation of the KnowledgeBaseInterface
type MockKnowledgeBase struct {
	mock.Mock
}

func (m *MockKnowledgeBase) AddToKnowledgeBase(ctx context.Context, document *domain.Document, content string) error {
	args := m.Called(ctx, document, content)
	return args.Error(0)
}

// Additional test scenarios for DocumentWorkflow

// TestDocumentWorkflow_ProcessDocument_WithDifyClient tests document processing with Dify client integration
func TestDocumentWorkflow_ProcessDocument_WithDifyClient(t *testing.T) {
	// Set up test database
	db := testutils.SetupTestDB()
	defer db.Migrator().DropTable(&domain.Document{}, &domain.DocumentVersion{})

	// Create mock services
	mockDifyClient := new(MockDifyClient)
	mockDocumentService := new(MockDocumentService)
	mockContentExtractor := new(MockContentExtractor)
	mockOCRExtractor := new(MockOCRExtractor)
	mockClassifier := new(MockClassifier)
	mockTagExtractor := new(MockTagExtractor)
	mockSummarizer := new(MockSummarizer)
	mockKnowledgeBase := new(MockKnowledgeBase)

	// Create a test document
	testDocument := &domain.Document{
		ID:          "doc-005",
		Title:       "Dify集成测试文档",
		Description: "这是一个Dify集成测试文档",
		FilePath:    "/path/to/dify-test.pdf",
		FileSize:    1024,
		MimeType:    "application/pdf",
		OwnerID:     "user-005",
		TeamID:      "team-005",
		Status:      "active",
	}

	// Create workflow with actual database
	docWorkflow := NewDocumentWorkflow(
		mockDifyClient,
		mockDocumentService,
		mockContentExtractor,
		mockOCRExtractor,
		mockClassifier,
		mockTagExtractor,
		mockSummarizer,
		mockKnowledgeBase,
	)

	// Set up expectations for successful processing with Dify client
	mockContentExtractor.On("ExtractContent", testDocument).Return("这是文档的内容", nil)
	mockClassifier.On("ClassifyDocument", mock.Anything, "这是文档的内容", testDocument).Return("技术文档", nil)
	mockTagExtractor.On("ExtractTags", mock.Anything, "这是文档的内容", testDocument).Return([]string{"技术", "Dify"}, nil)
	mockSummarizer.On("SummarizeDocument", mock.Anything, "这是文档的内容", testDocument).Return("这是文档的摘要", nil)
	mockKnowledgeBase.On("AddToKnowledgeBase", mock.Anything, testDocument, "这是文档的内容").Return(nil)

	// Test successful document processing with Dify client
	err := docWorkflow.ProcessDocument(context.Background(), testDocument)
	assert.NoError(t, err)

	// Verify that the document was updated with AI results
	assert.Contains(t, testDocument.Description, "这是文档的摘要")
	assert.Contains(t, testDocument.Description, "技术文档")
	assert.Equal(t, `["技术","Dify"]`, testDocument.Tags)

	// Verify that all expectations were met
	mockContentExtractor.AssertExpectations(t)
	mockClassifier.AssertExpectations(t)
	mockTagExtractor.AssertExpectations(t)
	mockSummarizer.AssertExpectations(t)
	mockKnowledgeBase.AssertExpectations(t)
}

// TestDocumentWorkflow_ProcessDocument_EmptyContent tests document processing with empty content
func TestDocumentWorkflow_ProcessDocument_EmptyContent(t *testing.T) {
	// Set up test database
	db := testutils.SetupTestDB()
	defer db.Migrator().DropTable(&domain.Document{}, &domain.DocumentVersion{})

	// Create mock services
	mockDifyClient := new(MockDifyClient)
	mockDocumentService := new(MockDocumentService)
	mockContentExtractor := new(MockContentExtractor)
	mockOCRExtractor := new(MockOCRExtractor)
	mockClassifier := new(MockClassifier)
	mockTagExtractor := new(MockTagExtractor)
	mockSummarizer := new(MockSummarizer)
	mockKnowledgeBase := new(MockKnowledgeBase)

	// Create a test document
	testDocument := &domain.Document{
		ID:          "doc-006",
		Title:       "空内容测试文档",
		Description: "这是一个空内容测试文档",
		FilePath:    "/path/to/empty-test.pdf",
		FileSize:    1024,
		MimeType:    "application/pdf",
		OwnerID:     "user-006",
		TeamID:      "team-006",
		Status:      "active",
	}

	// Create workflow with actual database
	docWorkflow := NewDocumentWorkflow(
		mockDifyClient,
		mockDocumentService,
		mockContentExtractor,
		mockOCRExtractor,
		mockClassifier,
		mockTagExtractor,
		mockSummarizer,
		mockKnowledgeBase,
	)

	// Set up expectations for empty content
	mockContentExtractor.On("ExtractContent", testDocument).Return("", nil)
	mockClassifier.On("ClassifyDocument", mock.Anything, "", testDocument).Return("未知文档", nil)
	mockTagExtractor.On("ExtractTags", mock.Anything, "", testDocument).Return([]string{}, nil)
	mockSummarizer.On("SummarizeDocument", mock.Anything, "", testDocument).Return("无法生成摘要", nil)
	mockKnowledgeBase.On("AddToKnowledgeBase", mock.Anything, testDocument, "").Return(nil)

	// Test document processing with empty content
	err := docWorkflow.ProcessDocument(context.Background(), testDocument)
	assert.NoError(t, err)

	// Verify that the document was updated with AI results
	assert.Contains(t, testDocument.Description, "无法生成摘要")
	assert.Contains(t, testDocument.Description, "未知文档")
	assert.Equal(t, `[]`, testDocument.Tags)

	// Verify that all expectations were met
	mockContentExtractor.AssertExpectations(t)
	mockClassifier.AssertExpectations(t)
	mockTagExtractor.AssertExpectations(t)
	mockSummarizer.AssertExpectations(t)
	mockKnowledgeBase.AssertExpectations(t)
}

// TestDocumentWorkflow_ProcessDocument_LargeDocument tests document processing with a large document
func TestDocumentWorkflow_ProcessDocument_LargeDocument(t *testing.T) {
	// Set up test database
	db := testutils.SetupTestDB()
	defer db.Migrator().DropTable(&domain.Document{}, &domain.DocumentVersion{})

	// Create mock services
	mockDifyClient := new(MockDifyClient)
	mockDocumentService := new(MockDocumentService)
	mockContentExtractor := new(MockContentExtractor)
	mockOCRExtractor := new(MockOCRExtractor)
	mockClassifier := new(MockClassifier)
	mockTagExtractor := new(MockTagExtractor)
	mockSummarizer := new(MockSummarizer)
	mockKnowledgeBase := new(MockKnowledgeBase)

	// Create a large test document
	testDocument := &domain.Document{
		ID:          "doc-007",
		Title:       "大文档测试",
		Description: "这是一个大文档测试",
		FilePath:    "/path/to/large-test.pdf",
		FileSize:    1024 * 1024 * 10, // 10MB
		MimeType:    "application/pdf",
		OwnerID:     "user-007",
		TeamID:      "team-007",
		Status:      "active",
	}

	// Create workflow with actual database
	docWorkflow := NewDocumentWorkflow(
		mockDifyClient,
		mockDocumentService,
		mockContentExtractor,
		mockOCRExtractor,
		mockClassifier,
		mockTagExtractor,
		mockSummarizer,
		mockKnowledgeBase,
	)

	// Create large content
	largeContent := ""
	for i := 0; i < 10000; i++ {
		largeContent += "这是很长的文档内容，用于测试大文档处理。"
	}

	// Set up expectations for large content
	mockContentExtractor.On("ExtractContent", testDocument).Return(largeContent, nil)
	mockClassifier.On("ClassifyDocument", mock.Anything, largeContent, testDocument).Return("大型文档", nil)
	mockTagExtractor.On("ExtractTags", mock.Anything, largeContent, testDocument).Return([]string{"大文档", "测试"}, nil)
	mockSummarizer.On("SummarizeDocument", mock.Anything, largeContent, testDocument).Return("这是大型文档的摘要", nil)
	mockKnowledgeBase.On("AddToKnowledgeBase", mock.Anything, testDocument, largeContent).Return(nil)

	// Test document processing with large content
	err := docWorkflow.ProcessDocument(context.Background(), testDocument)
	assert.NoError(t, err)

	// Verify that the document was updated with AI results
	assert.Contains(t, testDocument.Description, "这是大型文档的摘要")
	assert.Contains(t, testDocument.Description, "大型文档")
	assert.Equal(t, `["大文档","测试"]`, testDocument.Tags)

	// Verify that all expectations were met
	mockContentExtractor.AssertExpectations(t)
	mockClassifier.AssertExpectations(t)
	mockTagExtractor.AssertExpectations(t)
	mockSummarizer.AssertExpectations(t)
	mockKnowledgeBase.AssertExpectations(t)
}

// TestDocumentWorkflow_ProcessDocument_MultipleVersions tests document processing with multiple versions
func TestDocumentWorkflow_ProcessDocument_MultipleVersions(t *testing.T) {
	// Set up test database
	db := testutils.SetupTestDB()
	defer db.Migrator().DropTable(&domain.Document{}, &domain.DocumentVersion{})

	// Create mock services
	mockDifyClient := new(MockDifyClient)
	mockDocumentService := new(MockDocumentService)
	mockContentExtractor := new(MockContentExtractor)
	mockOCRExtractor := new(MockOCRExtractor)
	mockClassifier := new(MockClassifier)
	mockTagExtractor := new(MockTagExtractor)
	mockSummarizer := new(MockSummarizer)
	mockKnowledgeBase := new(MockKnowledgeBase)

	// Create a test document
	testDocument := &domain.Document{
		ID:          "doc-008",
		Title:       "多版本测试文档",
		Description: "这是一个多版本测试文档",
		FilePath:    "/path/to/version-test.pdf",
		FileSize:    1024,
		MimeType:    "application/pdf",
		OwnerID:     "user-008",
		TeamID:      "team-008",
		Status:      "active",
	}

	// Create workflow with actual database
	docWorkflow := NewDocumentWorkflow(
		mockDifyClient,
		mockDocumentService,
		mockContentExtractor,
		mockOCRExtractor,
		mockClassifier,
		mockTagExtractor,
		mockSummarizer,
		mockKnowledgeBase,
	)

	// Process the document for the first time
	mockContentExtractor.On("ExtractContent", testDocument).Return("这是文档的第一个版本内容", nil).Once()
	mockClassifier.On("ClassifyDocument", mock.Anything, "这是文档的第一个版本内容", testDocument).Return("版本1文档", nil).Once()
	mockTagExtractor.On("ExtractTags", mock.Anything, "这是文档的第一个版本内容", testDocument).Return([]string{"版本1", "测试"}, nil).Once()
	mockSummarizer.On("SummarizeDocument", mock.Anything, "这是文档的第一个版本内容", testDocument).Return("这是第一个版本的摘要", nil).Once()
	mockKnowledgeBase.On("AddToKnowledgeBase", mock.Anything, testDocument, "这是文档的第一个版本内容").Return(nil).Once()

	err := docWorkflow.ProcessDocument(context.Background(), testDocument)
	assert.NoError(t, err)

	// Verify first processing results
	assert.Contains(t, testDocument.Description, "这是第一个版本的摘要")
	assert.Contains(t, testDocument.Description, "版本1文档")
	assert.Equal(t, `["版本1","测试"]`, testDocument.Tags)

	// Process the document for the second time (simulating a new version)
	mockContentExtractor.On("ExtractContent", testDocument).Return("这是文档的第二个版本内容", nil).Once()
	mockClassifier.On("ClassifyDocument", mock.Anything, "这是文档的第二个版本内容", testDocument).Return("版本2文档", nil).Once()
	mockTagExtractor.On("ExtractTags", mock.Anything, "这是文档的第二个版本内容", testDocument).Return([]string{"版本2", "更新"}, nil).Once()
	mockSummarizer.On("SummarizeDocument", mock.Anything, "这是文档的第二个版本内容", testDocument).Return("这是第二个版本的摘要", nil).Once()
	mockKnowledgeBase.On("AddToKnowledgeBase", mock.Anything, testDocument, "这是文档的第二个版本内容").Return(nil).Once()

	err = docWorkflow.ProcessDocument(context.Background(), testDocument)
	assert.NoError(t, err)

	// Verify second processing results
	assert.Contains(t, testDocument.Description, "这是第二个版本的摘要")
	assert.Contains(t, testDocument.Description, "版本2文档")
	assert.Equal(t, `["版本2","更新"]`, testDocument.Tags)

	// Verify that all expectations were met
	mockContentExtractor.AssertExpectations(t)
	mockClassifier.AssertExpectations(t)
	mockTagExtractor.AssertExpectations(t)
	mockSummarizer.AssertExpectations(t)
	mockKnowledgeBase.AssertExpectations(t)
}

// TestDocumentWorkflow_ExtractContent_Failure tests extractContent with both extraction methods failing
func TestDocumentWorkflow_ExtractContent_Failure(t *testing.T) {
	// Set up test database
	db := testutils.SetupTestDB()
	defer db.Migrator().DropTable(&domain.Document{}, &domain.DocumentVersion{})

	// Create mock services
	mockDifyClient := new(MockDifyClient)
	mockDocumentService := new(MockDocumentService)
	mockContentExtractor := new(MockContentExtractor)
	mockOCRExtractor := new(MockOCRExtractor)
	mockClassifier := new(MockClassifier)
	mockTagExtractor := new(MockTagExtractor)
	mockSummarizer := new(MockSummarizer)
	mockKnowledgeBase := new(MockKnowledgeBase)

	// Create a test document
	testDocument := &domain.Document{
		ID:          "doc-009",
		Title:       "提取失败测试文档",
		Description: "这是一个提取失败测试文档",
		FilePath:    "/path/to/failure-test.pdf",
		FileSize:    1024,
		MimeType:    "application/pdf",
		OwnerID:     "user-009",
		TeamID:      "team-009",
		Status:      "active",
	}

	// Create workflow with actual database
	docWorkflow := NewDocumentWorkflow(
		mockDifyClient,
		mockDocumentService,
		mockContentExtractor,
		mockOCRExtractor,
		mockClassifier,
		mockTagExtractor,
		mockSummarizer,
		mockKnowledgeBase,
	)

	// Set up expectations for both extraction methods failing
	mockContentExtractor.On("ExtractContent", testDocument).Return("", assert.AnError)
	mockOCRExtractor.On("ExtractOCRContent", testDocument).Return("", assert.AnError)

	// Test document processing with both extraction methods failing
	err := docWorkflow.ProcessDocument(context.Background(), testDocument)
	assert.Error(t, err)
	assert.Equal(t, "failed to process document", err.Error())

	// Verify that all expectations were met
	mockContentExtractor.AssertExpectations(t)
	mockOCRExtractor.AssertExpectations(t)
}

// TestDocumentWorkflow_ExtractContent_OCROnly tests extractContent with only OCR extraction succeeding
func TestDocumentWorkflow_ExtractContent_OCROnly(t *testing.T) {
	// Set up test database
	db := testutils.SetupTestDB()
	defer db.Migrator().DropTable(&domain.Document{}, &domain.DocumentVersion{})

	// Create mock services
	mockDifyClient := new(MockDifyClient)
	mockDocumentService := new(MockDocumentService)
	mockContentExtractor := new(MockContentExtractor)
	mockOCRExtractor := new(MockOCRExtractor)
	mockClassifier := new(MockClassifier)
	mockTagExtractor := new(MockTagExtractor)
	mockSummarizer := new(MockSummarizer)
	mockKnowledgeBase := new(MockKnowledgeBase)

	// Create a test document
	testDocument := &domain.Document{
		ID:          "doc-010",
		Title:       "OCR提取测试文档",
		Description: "这是一个OCR提取测试文档",
		FilePath:    "/path/to/ocr-test.png",
		FileSize:    2048,
		MimeType:    "image/png",
		OwnerID:     "user-010",
		TeamID:      "team-010",
		Status:      "active",
	}

	// Create workflow with actual database
	docWorkflow := NewDocumentWorkflow(
		mockDifyClient,
		mockDocumentService,
		mockContentExtractor,
		mockOCRExtractor,
		mockClassifier,
		mockTagExtractor,
		mockSummarizer,
		mockKnowledgeBase,
	)

	// Set up expectations for content extraction failing but OCR extraction succeeding
	mockContentExtractor.On("ExtractContent", testDocument).Return("", assert.AnError)
	mockOCRExtractor.On("ExtractOCRContent", testDocument).Return("这是OCR提取的内容", nil)
	mockClassifier.On("ClassifyDocument", mock.Anything, "这是OCR提取的内容", testDocument).Return("OCR文档", nil)
	mockTagExtractor.On("ExtractTags", mock.Anything, "这是OCR提取的内容", testDocument).Return([]string{"OCR", "图像"}, nil)
	mockSummarizer.On("SummarizeDocument", mock.Anything, "这是OCR提取的内容", testDocument).Return("这是OCR文档的摘要", nil)
	mockKnowledgeBase.On("AddToKnowledgeBase", mock.Anything, testDocument, "这是OCR提取的内容").Return(nil)

	// Test document processing with OCR extraction only
	err := docWorkflow.ProcessDocument(context.Background(), testDocument)
	assert.NoError(t, err)

	// Verify that the document was updated with AI results
	assert.Contains(t, testDocument.Description, "这是OCR文档的摘要")
	assert.Contains(t, testDocument.Description, "OCR文档")
	assert.Equal(t, `["OCR","图像"]`, testDocument.Tags)

	// Verify that all expectations were met
	mockContentExtractor.AssertExpectations(t)
	mockOCRExtractor.AssertExpectations(t)
	mockClassifier.AssertExpectations(t)
	mockTagExtractor.AssertExpectations(t)
	mockSummarizer.AssertExpectations(t)
	mockKnowledgeBase.AssertExpectations(t)
}

// TestDocumentWorkflow_ProcessWithAI_Failure tests processWithAI with classification failure
func TestDocumentWorkflow_ProcessWithAI_Failure(t *testing.T) {
	// Set up test database
	db := testutils.SetupTestDB()
	defer db.Migrator().DropTable(&domain.Document{}, &domain.DocumentVersion{})

	// Create mock services
	mockDifyClient := new(MockDifyClient)
	mockDocumentService := new(MockDocumentService)
	mockContentExtractor := new(MockContentExtractor)
	mockOCRExtractor := new(MockOCRExtractor)
	mockClassifier := new(MockClassifier)
	mockTagExtractor := new(MockTagExtractor)
	mockSummarizer := new(MockSummarizer)
	mockKnowledgeBase := new(MockKnowledgeBase)

	// Create a test document
	testDocument := &domain.Document{
		ID:          "doc-011",
		Title:       "AI处理失败测试文档",
		Description: "这是一个AI处理失败测试文档",
		FilePath:    "/path/to/ai-failure-test.pdf",
		FileSize:    1024,
		MimeType:    "application/pdf",
		OwnerID:     "user-011",
		TeamID:      "team-011",
		Status:      "active",
	}

	// Create workflow with actual database
	docWorkflow := NewDocumentWorkflow(
		mockDifyClient,
		mockDocumentService,
		mockContentExtractor,
		mockOCRExtractor,
		mockClassifier,
		mockTagExtractor,
		mockSummarizer,
		mockKnowledgeBase,
	)

	// Set up expectations for AI processing failure
	mockContentExtractor.On("ExtractContent", testDocument).Return("这是文档的内容", nil)
	mockClassifier.On("ClassifyDocument", mock.Anything, "这是文档的内容", testDocument).Return("", assert.AnError)

	// Test document processing with AI classification failure
	err := docWorkflow.ProcessDocument(context.Background(), testDocument)
	assert.Error(t, err)
	assert.Equal(t, "failed to process document", err.Error())

	// Verify that all expectations were met
	mockContentExtractor.AssertExpectations(t)
	mockClassifier.AssertExpectations(t)
}

// TestDocumentWorkflow_AddToKnowledgeBase_Error tests addToKnowledgeBase with error
func TestDocumentWorkflow_AddToKnowledgeBase_Error(t *testing.T) {
	// Set up test database
	db := testutils.SetupTestDB()
	defer db.Migrator().DropTable(&domain.Document{}, &domain.DocumentVersion{})

	// Create mock services
	mockDifyClient := new(MockDifyClient)
	mockDocumentService := new(MockDocumentService)
	mockContentExtractor := new(MockContentExtractor)
	mockOCRExtractor := new(MockOCRExtractor)
	mockClassifier := new(MockClassifier)
	mockTagExtractor := new(MockTagExtractor)
	mockSummarizer := new(MockSummarizer)
	mockKnowledgeBase := new(MockKnowledgeBase)

	// Create a test document
	testDocument := &domain.Document{
		ID:          "doc-012",
		Title:       "知识库添加失败测试文档",
		Description: "这是一个知识库添加失败测试文档",
		FilePath:    "/path/to/kb-error-test.pdf",
		FileSize:    1024,
		MimeType:    "application/pdf",
		OwnerID:     "user-012",
		TeamID:      "team-012",
		Status:      "active",
	}

	// Create workflow with actual database
	docWorkflow := NewDocumentWorkflow(
		mockDifyClient,
		mockDocumentService,
		mockContentExtractor,
		mockOCRExtractor,
		mockClassifier,
		mockTagExtractor,
		mockSummarizer,
		mockKnowledgeBase,
	)

	// Set up expectations - everything succeeds except knowledge base
	mockContentExtractor.On("ExtractContent", testDocument).Return("这是文档的内容", nil)
	mockClassifier.On("ClassifyDocument", mock.Anything, "这是文档的内容", testDocument).Return("技术文档", nil)
	mockTagExtractor.On("ExtractTags", mock.Anything, "这是文档的内容", testDocument).Return([]string{"技术", "测试"}, nil)
	mockSummarizer.On("SummarizeDocument", mock.Anything, "这是文档的内容", testDocument).Return("这是文档的摘要", nil)
	mockKnowledgeBase.On("AddToKnowledgeBase", mock.Anything, testDocument, "这是文档的内容").Return(assert.AnError)

	// Test document processing with knowledge base error (should not fail the entire process)
	err := docWorkflow.ProcessDocument(context.Background(), testDocument)
	assert.NoError(t, err) // Process should succeed even if knowledge base fails

	// Verify that the document was updated with AI results
	assert.Contains(t, testDocument.Description, "这是文档的摘要")
	assert.Contains(t, testDocument.Description, "技术文档")
	assert.Equal(t, `["技术","测试"]`, testDocument.Tags)

	// Verify that all expectations were met
	mockContentExtractor.AssertExpectations(t)
	mockClassifier.AssertExpectations(t)
	mockTagExtractor.AssertExpectations(t)
	mockSummarizer.AssertExpectations(t)
	mockKnowledgeBase.AssertExpectations(t)
}