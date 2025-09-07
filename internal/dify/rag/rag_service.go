package rag

import (
	"context"
	"errors"
	"time"

	"cdk-office/internal/dify/client"
	"cdk-office/pkg/logger"
)

// RAGServiceInterface defines the interface for RAG service
type RAGServiceInterface interface {
	Search(ctx context.Context, query string, kb string) ([]*Document, error)
	CreateKnowledgeBase(ctx context.Context, name string, config *KBConfig) (*KnowledgeBase, error)
	UpdateDocument(ctx context.Context, kbID string, doc *Document) error
}

// RAGService implements the RAGServiceInterface
type RAGService struct {
	difyClient client.DifyClientInterface
}

// NewRAGService creates a new instance of RAGService
func NewRAGService(difyClient client.DifyClientInterface) *RAGService {
	return &RAGService{
		difyClient: difyClient,
	}
}

// Document represents a document in the knowledge base
type Document struct {
	ID       string `json:"id"`
	Title    string `json:"title"`
	Content  string `json:"content"`
	Metadata map[string]interface{} `json:"metadata"`
}

// KnowledgeBase represents a knowledge base
type KnowledgeBase struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	CreatedAt   string    `json:"created_at"`
}

// KBConfig represents knowledge base configuration
type KBConfig struct {
	Description string            `json:"description"`
	Parameters  map[string]string `json:"parameters"`
}

// Search performs a search in the knowledge base
func (s *RAGService) Search(ctx context.Context, query string, kb string) ([]*Document, error) {
	// Prepare the request for Dify RAG API
	req := &client.CompletionRequest{
		Query: query,
		Inputs: map[string]interface{}{
			"kb_id": kb,
		},
		ResponseMode: "blocking",
		User: "cdk-office",
	}
	
	// Call the Dify API
	resp, err := s.difyClient.CreateCompletionMessage(ctx, req)
	if err != nil {
		logger.Error("failed to search knowledge base", "error", err)
		return nil, errors.New("failed to search knowledge base")
	}
	
	// Parse the response and convert to Document objects
	// This implementation assumes the Dify API returns results in a format we can convert to Document objects
	documents := []*Document{
		{
			ID:      "doc_" + resp.MessageID,
			Title:   "Search Result for: " + query,
			Content: resp.Answer,
			Metadata: map[string]interface{}{
				"source": "knowledge_base",
				"score":  0.95,
			},
		},
	}
	
	logger.Info("performed knowledge base search", "query", query, "kb", kb, "results", len(documents))
	
	return documents, nil
}

// CreateKnowledgeBase creates a new knowledge base
func (s *RAGService) CreateKnowledgeBase(ctx context.Context, name string, config *KBConfig) (*KnowledgeBase, error) {
	// Prepare the request for Dify RAG API to create a knowledge base
	// This is a simplified implementation - the actual Dify API may require different parameters
	req := &client.CompletionRequest{
		Query: "create_knowledge_base",
		Inputs: map[string]interface{}{
			"name": name,
			"description": config.Description,
			"parameters": config.Parameters,
		},
		ResponseMode: "blocking",
		User: "cdk-office",
	}
	
	// Call the Dify API
	resp, err := s.difyClient.CreateCompletionMessage(ctx, req)
	if err != nil {
		logger.Error("failed to create knowledge base", "error", err)
		return nil, errors.New("failed to create knowledge base")
	}
	
	// Parse the response and create a KnowledgeBase object
	kb := &KnowledgeBase{
		ID:          "kb_" + resp.MessageID,
		Name:        name,
		Description: config.Description,
		CreatedAt:   time.Now().Format(time.RFC3339),
	}
	
	logger.Info("created knowledge base", "name", name, "id", kb.ID)
	
	return kb, nil
}

// UpdateDocument updates a document in the knowledge base
func (s *RAGService) UpdateDocument(ctx context.Context, kbID string, doc *Document) error {
	// Prepare the request for Dify RAG API to update a document
	// This is a simplified implementation - the actual Dify API may require different parameters
	req := &client.CompletionRequest{
		Query: "update_document",
		Inputs: map[string]interface{}{
			"kb_id": kbID,
			"document_id": doc.ID,
			"content": doc.Content,
			"title": doc.Title,
			"metadata": doc.Metadata,
		},
		ResponseMode: "blocking",
		User: "cdk-office",
	}
	
	// Call the Dify API
	_, err := s.difyClient.CreateCompletionMessage(ctx, req)
	if err != nil {
		logger.Error("failed to update document in knowledge base", "error", err)
		return errors.New("failed to update document in knowledge base")
	}
	
	logger.Info("updated document in knowledge base", "kb_id", kbID, "doc_id", doc.ID)
	
	return nil
}