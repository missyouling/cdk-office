package workflow

import (
	"context"
	"errors"
	"time"

	"cdk-office/internal/dify/client"
	"cdk-office/internal/document/domain"
	"cdk-office/internal/document/service"
	"cdk-office/internal/shared/database"
	"cdk-office/pkg/logger"
	"gorm.io/gorm"
)

// DocumentWorkflowInterface defines the interface for document processing workflow
type DocumentWorkflowInterface interface {
	ProcessDocument(ctx context.Context, document *domain.Document) error
}

// DocumentWorkflow implements the DocumentWorkflowInterface
type DocumentWorkflow struct {
	db                 *gorm.DB
	difyClient         client.DifyClientInterface
	documentService    service.DocumentServiceInterface
	contentExtractor   service.ContentExtractorInterface
	ocrExtractor       service.OCRExtractorInterface
	classifier         service.ClassifierInterface
	tagExtractor       service.TagExtractorInterface
	summarizer         service.SummarizerInterface
	knowledgeBase      service.KnowledgeBaseInterface
}

// NewDocumentWorkflow creates a new instance of DocumentWorkflow
func NewDocumentWorkflow(
	difyClient client.DifyClientInterface,
	documentService service.DocumentServiceInterface,
	contentExtractor service.ContentExtractorInterface,
	ocrExtractor service.OCRExtractorInterface,
	classifier service.ClassifierInterface,
	tagExtractor service.TagExtractorInterface,
	summarizer service.SummarizerInterface,
	knowledgeBase service.KnowledgeBaseInterface,
) *DocumentWorkflow {
	return &DocumentWorkflow{
		db:                 database.GetDB(),
		difyClient:         difyClient,
		documentService:    documentService,
		contentExtractor:   contentExtractor,
		ocrExtractor:       ocrExtractor,
		classifier:         classifier,
		tagExtractor:       tagExtractor,
		summarizer:         summarizer,
		knowledgeBase:      knowledgeBase,
	}
}

// ProcessDocument processes a document through the AI workflow
func (w *DocumentWorkflow) ProcessDocument(ctx context.Context, document *domain.Document) error {
	// 1. Extract content from document
	content, err := w.extractContent(document)
	if err != nil {
		logger.Error("failed to extract content from document", "error", err)
		return errors.New("failed to process document")
	}

	// 2. Send content to Dify for AI processing
	aiResult, err := w.processWithAI(ctx, content, document)
	if err != nil {
		logger.Error("failed to process document with AI", "error", err)
		return errors.New("failed to process document")
	}

	// 3. Update document with AI results
	if err := w.updateDocumentWithAIResults(ctx, document, aiResult); err != nil {
		logger.Error("failed to update document with AI results", "error", err)
		return errors.New("failed to process document")
	}

	// 4. Add document to knowledge base
	if err := w.addToKnowledgeBase(ctx, document, content); err != nil {
		// Log error but don't fail the entire process
		logger.Error("failed to add document to knowledge base", "error", err)
	}

	// 5. Notify relevant users (simplified - in real implementation, you would send actual notifications)
	if err := w.notifyUsers(ctx, document); err != nil {
		logger.Error("failed to notify users", "error", err)
		// Don't return error here as the document processing was successful
	}

	return nil
}

// extractContent extracts content from a document
func (w *DocumentWorkflow) extractContent(document *domain.Document) (string, error) {
	// Try to extract content using the content extractor
	content, err := w.contentExtractor.ExtractContent(document)
	if err != nil {
		// If content extraction fails, try OCR extraction for image documents
		ocrContent, ocrErr := w.ocrExtractor.ExtractOCRContent(document)
		if ocrErr != nil {
			// If both extraction methods fail, return the original error
			return "", err
		}
		return ocrContent, nil
	}
	return content, nil
}

// processWithAI processes document content with AI through Dify
func (w *DocumentWorkflow) processWithAI(ctx context.Context, content string, document *domain.Document) (*AIResult, error) {
	// 1. Classify document
	classification, err := w.classifier.ClassifyDocument(ctx, content, document)
	if err != nil {
		return nil, err
	}

	// 2. Extract tags
	tags, err := w.tagExtractor.ExtractTags(ctx, content, document)
	if err != nil {
		return nil, err
	}

	// 3. Generate summary
	summary, err := w.summarizer.SummarizeDocument(ctx, content, document)
	if err != nil {
		return nil, err
	}

	return &AIResult{
		Classification: classification,
		Tags:          tags,
		Summary:       summary,
	}, nil
}

// addToKnowledgeBase adds the document to the Dify knowledge base
func (w *DocumentWorkflow) addToKnowledgeBase(ctx context.Context, document *domain.Document, content string) error {
	return w.knowledgeBase.AddToKnowledgeBase(ctx, document, content)
}

// updateDocumentWithAIResults updates the document with AI processing results
func (w *DocumentWorkflow) updateDocumentWithAIResults(ctx context.Context, document *domain.Document, aiResult *AIResult) error {
	// Update document with AI results
	document.Description = aiResult.Summary
	document.Category = aiResult.Classification

	// Convert tags to JSON string
	tagsJSON := "["
	for i, tag := range aiResult.Tags {
		if i > 0 {
			tagsJSON += ","
		}
		tagsJSON += "\"" + tag + "\""
	}
	tagsJSON += "]"
	document.Tags = tagsJSON

	document.UpdatedAt = time.Now()

	// Save updated document to database
	if err := w.db.Save(document).Error; err != nil {
		logger.Error("failed to update document with AI results", "error", err)
		return errors.New("failed to update document")
	}

	return nil
}

// notifyUsers notifies relevant users about document processing completion
func (w *DocumentWorkflow) notifyUsers(ctx context.Context, document *domain.Document) error {
	// In a real implementation, you would send notifications to relevant users
	// For now, we'll just log that this step would happen
	logger.Info("would notify users about document processing", "document_id", document.ID)
	return nil
}

// AIResult represents the results from AI processing
type AIResult struct {
	Classification string   `json:"classification"`
	Tags          []string `json:"tags"`
	Summary       string   `json:"summary"`
}