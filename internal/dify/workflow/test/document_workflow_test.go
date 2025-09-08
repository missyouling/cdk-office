package test

import (
	"context"
	"testing"

	"cdk-office/internal/dify/workflow"
	"cdk-office/internal/document/domain"
	"cdk-office/internal/document/service"
	"cdk-office/internal/shared/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockDifyClient is a mock implementation of the Dify client
type MockDifyClient struct {
	mock.Mock
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

// Custom error types for testing
type ExtractionError struct {
	Message string
}

func (e *ExtractionError) Error() string {
	return e.Message
}

type ProcessingError struct {
	Message string
}

func (e *ProcessingError) Error() string {
	return e.Message
}

func TestDocumentWorkflow_ProcessDocument(t *testing.T) {
	// Set up test database
	db := testutils.SetupTestDB()
	defer db.Migrator().DropTable(&domain.Document{}, &domain.DocumentVersion{})

	// Create mock services
	mockDocumentService := new(MockDocumentService)
	mockContentExtractor := new(MockContentExtractor)
	mockOCRExtractor := new(MockOCRExtractor)
	mockClassifier := new(MockClassifier)
	mockTagExtractor := new(MockTagExtractor)
	mockSummarizer := new(MockSummarizer)
	mockKnowledgeBase := new(MockKnowledgeBase)

	// Create a test document
	testDocument := &domain.Document{
		ID:          "doc-001",
		Title:       "测试文档",
		Description: "这是一个测试文档",
		FilePath:    "/path/to/test.pdf",
		FileSize:    1024,
		MimeType:    "application/pdf",
		OwnerID:     "user-001",
		TeamID:      "team-001",
		Status:      "active",
	}

	// Create workflow with actual database
	docWorkflow := workflow.NewDocumentWorkflow(
		nil, // difyClient not used in this test
		mockDocumentService,
		mockContentExtractor,
		mockOCRExtractor,
		mockClassifier,
		mockTagExtractor,
		mockSummarizer,
		mockKnowledgeBase,
	)

	// Set up expectations for successful processing
	mockContentExtractor.On("ExtractContent", testDocument).Return("这是文档的内容", nil)
	mockClassifier.On("ClassifyDocument", mock.Anything, "这是文档的内容", testDocument).Return("技术文档", nil)
	mockTagExtractor.On("ExtractTags", mock.Anything, "这是文档的内容", testDocument).Return([]string{"技术", "测试"}, nil)
	mockSummarizer.On("SummarizeDocument", mock.Anything, "这是文档的内容", testDocument).Return("这是文档的摘要", nil)
	mockKnowledgeBase.On("AddToKnowledgeBase", mock.Anything, testDocument, "这是文档的内容").Return(nil)

	// Test successful document processing
	err := docWorkflow.ProcessDocument(context.Background(), testDocument)
	assert.NoError(t, err)

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

func TestDocumentWorkflow_ProcessDocument_ContentExtractionFallback(t *testing.T) {
	// Set up test database
	db := testutils.SetupTestDB()
	defer db.Migrator().DropTable(&domain.Document{}, &domain.DocumentVersion{})

	// Create mock services
	mockDocumentService := new(MockDocumentService)
	mockContentExtractor := new(MockContentExtractor)
	mockOCRExtractor := new(MockOCRExtractor)
	mockClassifier := new(MockClassifier)
	mockTagExtractor := new(MockTagExtractor)
	mockSummarizer := new(MockSummarizer)
	mockKnowledgeBase := new(MockKnowledgeBase)

	// Create a test document
	testDocument := &domain.Document{
		ID:          "doc-002",
		Title:       "图片文档",
		Description: "这是一个图片文档",
		FilePath:    "/path/to/test.jpg",
		FileSize:    2048,
		MimeType:    "image/jpeg",
		OwnerID:     "user-002",
		TeamID:      "team-002",
		Status:      "active",
	}

	// Create workflow with actual database
	docWorkflow := workflow.NewDocumentWorkflow(
		nil, // difyClient not used in this test
		mockDocumentService,
		mockContentExtractor,
		mockOCRExtractor,
		mockClassifier,
		mockTagExtractor,
		mockSummarizer,
		mockKnowledgeBase,
	)

	// Set up expectations for content extraction fallback to OCR
	mockContentExtractor.On("ExtractContent", testDocument).Return("", &ExtractionError{Message: "无法提取内容"})
	mockOCRExtractor.On("ExtractOCRContent", testDocument).Return("这是OCR提取的内容", nil)
	mockClassifier.On("ClassifyDocument", mock.Anything, "这是OCR提取的内容", testDocument).Return("扫描文档", nil)
	mockTagExtractor.On("ExtractTags", mock.Anything, "这是OCR提取的内容", testDocument).Return([]string{"扫描", "OCR"}, nil)
	mockSummarizer.On("SummarizeDocument", mock.Anything, "这是OCR提取的内容", testDocument).Return("这是OCR文档的摘要", nil)
	mockKnowledgeBase.On("AddToKnowledgeBase", mock.Anything, testDocument, "这是OCR提取的内容").Return(nil)

	// Test document processing with OCR fallback
	err := docWorkflow.ProcessDocument(context.Background(), testDocument)
	assert.NoError(t, err)

	// Verify that the document was updated with AI results
	assert.Contains(t, testDocument.Description, "这是OCR文档的摘要")
	assert.Contains(t, testDocument.Description, "扫描文档")
	assert.Equal(t, `["扫描","OCR"]`, testDocument.Tags)

	// Verify that all expectations were met
	mockContentExtractor.AssertExpectations(t)
	mockOCRExtractor.AssertExpectations(t)
	mockClassifier.AssertExpectations(t)
	mockTagExtractor.AssertExpectations(t)
	mockSummarizer.AssertExpectations(t)
	mockKnowledgeBase.AssertExpectations(t)
}

func TestDocumentWorkflow_ProcessDocument_AIProcessingError(t *testing.T) {
	// Set up test database
	db := testutils.SetupTestDB()
	defer db.Migrator().DropTable(&domain.Document{}, &domain.DocumentVersion{})

	// Create mock services
	mockDocumentService := new(MockDocumentService)
	mockContentExtractor := new(MockContentExtractor)
	mockOCRExtractor := new(MockOCRExtractor)
	mockClassifier := new(MockClassifier)
	mockTagExtractor := new(MockTagExtractor)
	mockSummarizer := new(MockSummarizer)
	mockKnowledgeBase := new(MockKnowledgeBase)

	// Create a test document
	testDocument := &domain.Document{
		ID:          "doc-003",
		Title:       "错误测试文档",
		Description: "这是一个错误测试文档",
		FilePath:    "/path/to/error.pdf",
		FileSize:    1024,
		MimeType:    "application/pdf",
		OwnerID:     "user-003",
		TeamID:      "team-003",
		Status:      "active",
	}

	// Create workflow with actual database
	docWorkflow := workflow.NewDocumentWorkflow(
		nil, // difyClient not used in this test
		mockDocumentService,
		mockContentExtractor,
		mockOCRExtractor,
		mockClassifier,
		mockTagExtractor,
		mockSummarizer,
		mockKnowledgeBase,
	)

	// Set up expectations for AI processing error
	mockContentExtractor.On("ExtractContent", testDocument).Return("这是文档的内容", nil)
	mockClassifier.On("ClassifyDocument", mock.Anything, "这是文档的内容", testDocument).Return("", &ProcessingError{Message: "AI处理失败"})

	// Test document processing with AI error
	err := docWorkflow.ProcessDocument(context.Background(), testDocument)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to process document")

	// Verify that the document was not updated
	assert.Equal(t, "这是一个错误测试文档", testDocument.Description)

	// Verify that all expectations were met
	mockContentExtractor.AssertExpectations(t)
	mockClassifier.AssertExpectations(t)
}

func TestDocumentWorkflow_ProcessDocument_KnowledgeBaseError(t *testing.T) {
	// Set up test database
	db := testutils.SetupTestDB()
	defer db.Migrator().DropTable(&domain.Document{}, &domain.DocumentVersion{})

	// Create mock services
	mockDocumentService := new(MockDocumentService)
	mockContentExtractor := new(MockContentExtractor)
	mockOCRExtractor := new(MockOCRExtractor)
	mockClassifier := new(MockClassifier)
	mockTagExtractor := new(MockTagExtractor)
	mockSummarizer := new(MockSummarizer)
	mockKnowledgeBase := new(MockKnowledgeBase)

	// Create a test document
	testDocument := &domain.Document{
		ID:          "doc-004",
		Title:       "知识库错误测试文档",
		Description: "这是一个知识库错误测试文档",
		FilePath:    "/path/to/kb-error.pdf",
		FileSize:    1024,
		MimeType:    "application/pdf",
		OwnerID:     "user-004",
		TeamID:      "team-004",
		Status:      "active",
	}

	// Create workflow with actual database
	docWorkflow := workflow.NewDocumentWorkflow(
		nil, // difyClient not used in this test
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
	mockKnowledgeBase.On("AddToKnowledgeBase", mock.Anything, testDocument, "这是文档的内容").Return(&ProcessingError{Message: "知识库添加失败"})

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