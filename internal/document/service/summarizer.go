package service

import (
	"cdk-office/internal/document/domain"
	"cdk-office/internal/dify/client"
	"cdk-office/pkg/logger"
	"context"
	"fmt"
)

// SummarizerInterface defines the interface for document summarization service
type SummarizerInterface interface {
	SummarizeDocument(ctx context.Context, content string, document *domain.Document) (string, error)
}

// Summarizer implements the SummarizerInterface
type Summarizer struct {
	difyClient client.DifyClientInterface
}

// NewSummarizer creates a new instance of Summarizer
func NewSummarizer(difyClient client.DifyClientInterface) *Summarizer {
	return &Summarizer{
		difyClient: difyClient,
	}
}

// SummarizeDocument generates a summary of the document content using Dify AI
func (s *Summarizer) SummarizeDocument(ctx context.Context, content string, document *domain.Document) (string, error) {
	// Prepare the summarization prompt
	prompt := fmt.Sprintf("Please generate a concise summary of the following document content. The summary should be no more than 200 words.\n\nDocument title: %s\n\nDocument content:\n%s", document.Title, content)

	// Create completion request
	req := &client.CompletionRequest{
		Query:        prompt,
		ResponseMode: "blocking",
		User:         document.OwnerID,
	}

	// Send request to Dify
	resp, err := s.difyClient.CreateCompletionMessage(ctx, req)
	if err != nil {
		logger.Error("failed to summarize document with Dify", "error", err)
		return "", fmt.Errorf("failed to summarize document: %v", err)
	}

	// Return the summary
	return resp.Answer, nil
}