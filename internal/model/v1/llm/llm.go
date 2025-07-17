package llm

import (
	"time"

	v1 "github.com/KubeOperator/kubepi/internal/model/v1"
)

// LLMModel represents an LLM model configuration
type LLMModel struct {
	ID           string      `json:"id" storm:"id"`
	Name         string      `json:"name" storm:"unique"`
	BaseURI      string      `json:"baseUri"`
	ModelName    string      `json:"modelName"`
	APIKey       string      `json:"apiKey"`
	Temperature  float64     `json:"temperature"`
	Metadata     v1.Metadata `json:"metadata"`
	CreatedAt    time.Time   `json:"createdAt" storm:"index"`
	LastTestTime time.Time   `json:"lastTestTime"`
	Status       string      `json:"status"` // "available", "unavailable", "untested"
}

// ChatRequest represents a request to the LLM API
type ChatRequest struct {
	Model       string         `json:"model"`
	Messages    []ChatMessage  `json:"messages"`
	Temperature float64        `json:"temperature"`
	Stream      bool           `json:"stream,omitempty"`
	Options     map[string]any `json:"options,omitempty"`
}

// ChatMessage represents a message in a chat
type ChatMessage struct {
	Role             string `json:"role"`
	Content          string `json:"content"`
	ReasoningContent string `json:"reasoning_content,omitempty"`
}

// ChatResponse represents a response from the LLM API
type ChatResponse struct {
	ID      string        `json:"id"`
	Object  string        `json:"object"`
	Created int64         `json:"created"`
	Model   string        `json:"model"`
	Choices []ChatChoice  `json:"choices"`
	Usage   ChatUsage     `json:"usage"`
	Error   *ChatError    `json:"error,omitempty"`
}

// ChatChoice represents a choice in a response
type ChatChoice struct {
	Index        int         `json:"index"`
	Message      ChatMessage `json:"message"`
	FinishReason string      `json:"finish_reason"`
}

// ChatUsage represents usage information
type ChatUsage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// ChatError represents an error in a response
type ChatError struct {
	Message string `json:"message"`
	Type    string `json:"type"`
	Code    string `json:"code"`
}

// TestRequest represents a request to test an LLM model
type TestRequest struct {
	Content string `json:"content"`
} 