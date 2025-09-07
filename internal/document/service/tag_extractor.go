package service

import (
	"cdk-office/internal/document/domain"
	"cdk-office/internal/dify/client"
	"cdk-office/pkg/logger"
	"context"
	"encoding/json"
	"fmt"
)

// TagExtractorInterface defines the interface for tag extraction service
type TagExtractorInterface interface {
	ExtractTags(ctx context.Context, content string, document *domain.Document) ([]string, error)
}

// TagExtractor implements the TagExtractorInterface
type TagExtractor struct {
	difyClient client.DifyClientInterface
}

// NewTagExtractor creates a new instance of TagExtractor
func NewTagExtractor(difyClient client.DifyClientInterface) *TagExtractor {
	return &TagExtractor{
		difyClient: difyClient,
	}
}

// ExtractTags extracts tags from document content using Dify AI
func (te *TagExtractor) ExtractTags(ctx context.Context, content string, document *domain.Document) ([]string, error) {
	// Prepare the tag extraction prompt
	prompt := fmt.Sprintf("Please extract relevant tags from the following document content. Return the tags as a JSON array of strings.\n\nDocument title: %s\n\nDocument content:\n%s", document.Title, content)

	// Create completion request
	req := &client.CompletionRequest{
		Query:        prompt,
		ResponseMode: "blocking",
		User:         document.OwnerID,
	}

	// Send request to Dify
	resp, err := te.difyClient.CreateCompletionMessage(ctx, req)
	if err != nil {
		logger.Error("failed to extract tags with Dify", "error", err)
		return nil, fmt.Errorf("failed to extract tags: %v", err)
	}

	// Parse the tags from the response
	var tags []string
	if err := json.Unmarshal([]byte(resp.Answer), &tags); err != nil {
		logger.Error("failed to parse tags from Dify response", "error", err, "response", resp.Answer)
		// Return a fallback set of tags
		return []string{"document", "ai_processed"}, nil
	}

	return tags, nil
}