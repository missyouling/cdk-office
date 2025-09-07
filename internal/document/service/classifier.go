package service

import (
	"cdk-office/internal/document/domain"
	"cdk-office/internal/dify/client"
	"cdk-office/pkg/logger"
	"context"
	"fmt"
)

// ClassifierInterface defines the interface for document classification service
type ClassifierInterface interface {
	ClassifyDocument(ctx context.Context, content string, document *domain.Document) (string, error)
}

// Classifier implements the ClassifierInterface
type Classifier struct {
	difyClient client.DifyClientInterface
}

// NewClassifier creates a new instance of Classifier
func NewClassifier(difyClient client.DifyClientInterface) *Classifier {
	return &Classifier{
		difyClient: difyClient,
	}
}

// ClassifyDocument classifies a document using Dify AI
func (c *Classifier) ClassifyDocument(ctx context.Context, content string, document *domain.Document) (string, error) {
	// Prepare the classification prompt
	prompt := fmt.Sprintf("Please classify the following document content into one of these categories: technical_document, business_document, legal_document, personal_document, other.\n\nDocument title: %s\n\nDocument content:\n%s", document.Title, content)

	// Create completion request
	req := &client.CompletionRequest{
		Query:        prompt,
		ResponseMode: "blocking",
		User:         document.OwnerID,
	}

	// Send request to Dify
	resp, err := c.difyClient.CreateCompletionMessage(ctx, req)
	if err != nil {
		logger.Error("failed to classify document with Dify", "error", err)
		return "", fmt.Errorf("failed to classify document: %v", err)
	}

	// Return the classification result
	return resp.Answer, nil
}