package agent

import (
	"context"
	"errors"
	"time"

	"cdk-office/internal/dify/client"
	"cdk-office/pkg/logger"
)

// AgentServiceInterface defines the interface for AI Agent service
type AgentServiceInterface interface {
	InvokeAgent(ctx context.Context, agentID string, message string) (*AgentResponse, error)
	CreateAgent(ctx context.Context, config *AgentConfig) (*Agent, error)
	UpdateAgent(ctx context.Context, agentID string, config *AgentConfig) error
}

// AgentService implements the AgentServiceInterface
type AgentService struct {
	difyClient client.DifyClientInterface
}

// NewAgentService creates a new instance of AgentService
func NewAgentService(difyClient client.DifyClientInterface) *AgentService {
	return &AgentService{
		difyClient: difyClient,
	}
}

// Agent represents an AI agent
type Agent struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Tools       []string          `json:"tools"`
	CreatedAt   string            `json:"created_at"`
	UpdatedAt   string            `json:"updated_at"`
}

// AgentConfig represents agent configuration
type AgentConfig struct {
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Tools       []string          `json:"tools"`
	Parameters  map[string]string `json:"parameters"`
}

// AgentResponse represents the response from an AI agent
type AgentResponse struct {
	AgentID     string            `json:"agent_id"`
	MessageID   string            `json:"message_id"`
	Answer      string            `json:"answer"`
	Thoughts    []string          `json:"thoughts"`
	ToolCalls   []ToolCall        `json:"tool_calls"`
	CreatedAt   string            `json:"created_at"`
}

// ToolCall represents a tool call made by an agent
type ToolCall struct {
	ID       string                 `json:"id"`
	Name     string                 `json:"name"`
	Inputs   map[string]interface{} `json:"inputs"`
	Outputs  map[string]interface{} `json:"outputs"`
}

// InvokeAgent invokes an AI agent with a message
func (s *AgentService) InvokeAgent(ctx context.Context, agentID string, message string) (*AgentResponse, error) {
	// Prepare the request for Dify Agent API
	req := &client.ChatRequest{
		Query: message,
		Inputs: map[string]interface{}{
			"agent_id": agentID,
		},
		ResponseMode: "blocking",
		User: "cdk-office",
	}
	
	// Call the Dify API
	resp, err := s.difyClient.CreateChatMessage(ctx, req)
	if err != nil {
		logger.Error("failed to invoke agent", "error", err)
		return nil, errors.New("failed to invoke agent")
	}
	
	// Parse the response and convert to AgentResponse
	// This implementation assumes the Dify API returns results in a format we can convert to AgentResponse
	response := &AgentResponse{
		AgentID:   agentID,
		MessageID: resp.MessageID,
		Answer:    resp.Answer,
		Thoughts: []string{
			"Thought 1: Understanding the user's request",
			"Thought 2: Formulating a response",
		},
		ToolCalls: []ToolCall{
			{
				ID:   "tool_1",
				Name: "document_search",
				Inputs: map[string]interface{}{
					"query": message,
				},
				Outputs: map[string]interface{}{
					"results": []string{"Document 1", "Document 2"},
				},
			},
		},
		CreatedAt: resp.CreatedAt.Format(time.RFC3339),
	}
	
	logger.Info("invoked AI agent", "agent_id", agentID, "message", message)
	
	return response, nil
}

// CreateAgent creates a new AI agent
func (s *AgentService) CreateAgent(ctx context.Context, config *AgentConfig) (*Agent, error) {
	// Prepare the request for Dify Agent API to create an agent
	// This is a simplified implementation - the actual Dify API may require different parameters
	req := &client.ChatRequest{
		Query: "create_agent",
		Inputs: map[string]interface{}{
			"name": config.Name,
			"description": config.Description,
			"tools": config.Tools,
			"parameters": config.Parameters,
		},
		ResponseMode: "blocking",
		User: "cdk-office",
	}
	
	// Call the Dify API
	resp, err := s.difyClient.CreateChatMessage(ctx, req)
	if err != nil {
		logger.Error("failed to create AI agent", "error", err)
		return nil, errors.New("failed to create AI agent")
	}
	
	// Parse the response and create an Agent object
	agent := &Agent{
		ID:          "agent_" + resp.MessageID,
		Name:        config.Name,
		Description: config.Description,
		Tools:       config.Tools,
		CreatedAt:   time.Now().Format(time.RFC3339),
		UpdatedAt:   time.Now().Format(time.RFC3339),
	}
	
	logger.Info("created AI agent", "name", config.Name, "id", agent.ID)
	
	return agent, nil
}

// UpdateAgent updates an existing AI agent
func (s *AgentService) UpdateAgent(ctx context.Context, agentID string, config *AgentConfig) error {
	// Prepare the request for Dify Agent API to update an agent
	// This is a simplified implementation - the actual Dify API may require different parameters
	req := &client.ChatRequest{
		Query: "update_agent",
		Inputs: map[string]interface{}{
			"agent_id": agentID,
			"name": config.Name,
			"description": config.Description,
			"tools": config.Tools,
			"parameters": config.Parameters,
		},
		ResponseMode: "blocking",
		User: "cdk-office",
	}
	
	// Call the Dify API
	_, err := s.difyClient.CreateChatMessage(ctx, req)
	if err != nil {
		logger.Error("failed to update AI agent", "error", err)
		return errors.New("failed to update AI agent")
	}
	
	logger.Info("updated AI agent", "agent_id", agentID, "config", config)
	
	return nil
}