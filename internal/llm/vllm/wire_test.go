package vllm

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestVLLMGenerateResponseStructure(t *testing.T) {
	// Test that GenerateResponse can handle typical response structure
	resp := GenerateResponse{
		ID:      "test-id",
		Object:  "text_completion",
		Created: 1234567890,
		Model:   "llama-2-7b",
		Choices: []struct {
			Index        int         `json:"index"`
			Text         string      `json:"text"`
			LogProbs     interface{} `json:"logprobs,omitempty"`
			FinishReason string      `json:"finish_reason"`
		}{
			{
				Index:        0,
				Text:         "Hello, world!",
				LogProbs:     nil,
				FinishReason: "stop",
			},
		},
		Usage: struct {
			PromptTokens     int `json:"prompt_tokens"`
			CompletionTokens int `json:"completion_tokens"`
			TotalTokens      int `json:"total_tokens"`
		}{
			PromptTokens:     10,
			CompletionTokens: 5,
			TotalTokens:      15,
		},
	}

	assert.Equal(t, "test-id", resp.ID)
	assert.Equal(t, "text_completion", resp.Object)
	assert.Equal(t, int64(1234567890), resp.Created)
	assert.Equal(t, "llama-2-7b", resp.Model)
	assert.Len(t, resp.Choices, 1)
	assert.Equal(t, 0, resp.Choices[0].Index)
	assert.Equal(t, "Hello, world!", resp.Choices[0].Text)
	assert.Equal(t, "stop", resp.Choices[0].FinishReason)
	assert.Equal(t, 10, resp.Usage.PromptTokens)
	assert.Equal(t, 5, resp.Usage.CompletionTokens)
	assert.Equal(t, 15, resp.Usage.TotalTokens)
}

func TestVLLMChatResponseStructure(t *testing.T) {
	// Test that ChatResponse can handle typical response structure
	resp := ChatResponse{
		ID:      "test-id",
		Object:  "chat.completion",
		Created: 1234567890,
		Model:   "llama-2-7b",
		Choices: []struct {
			Index        int         `json:"index"`
			Message      ChatMessage `json:"message"`
			FinishReason string      `json:"finish_reason"`
		}{
			{
				Index: 0,
				Message: ChatMessage{
					Role:    "assistant",
					Content: "Hello! How can I help you?",
				},
				FinishReason: "stop",
			},
		},
		Usage: struct {
			PromptTokens     int `json:"prompt_tokens"`
			CompletionTokens int `json:"completion_tokens"`
			TotalTokens      int `json:"total_tokens"`
		}{
			PromptTokens:     10,
			CompletionTokens: 5,
			TotalTokens:      15,
		},
	}

	assert.Equal(t, "test-id", resp.ID)
	assert.Equal(t, "chat.completion", resp.Object)
	assert.Equal(t, int64(1234567890), resp.Created)
	assert.Equal(t, "llama-2-7b", resp.Model)
	assert.Len(t, resp.Choices, 1)
	assert.Equal(t, 0, resp.Choices[0].Index)
	assert.Equal(t, "assistant", resp.Choices[0].Message.Role)
	assert.Equal(t, "Hello! How can I help you?", resp.Choices[0].Message.Content)
	assert.Equal(t, "stop", resp.Choices[0].FinishReason)
	assert.Equal(t, 10, resp.Usage.PromptTokens)
	assert.Equal(t, 5, resp.Usage.CompletionTokens)
	assert.Equal(t, 15, resp.Usage.TotalTokens)
}

func TestVLLMModelsResponseStructure(t *testing.T) {
	// Test that ModelsResponse can handle typical response structure
	resp := ModelsResponse{
		Object: "list",
		Data: []ModelInfo{
			{
				ID:      "llama-2-7b",
				Object:  "model",
				Created: 1234567890,
				OwnedBy: "vllm",
			},
			{
				ID:      "mistral-7b",
				Object:  "model",
				Created: 1234567891,
				OwnedBy: "vllm",
			},
		},
	}

	assert.Equal(t, "list", resp.Object)
	assert.Len(t, resp.Data, 2)
	assert.Equal(t, "llama-2-7b", resp.Data[0].ID)
	assert.Equal(t, "model", resp.Data[0].Object)
	assert.Equal(t, int64(1234567890), resp.Data[0].Created)
	assert.Equal(t, "vllm", resp.Data[0].OwnedBy)
	assert.Equal(t, "mistral-7b", resp.Data[1].ID)
	assert.Equal(t, "model", resp.Data[1].Object)
	assert.Equal(t, int64(1234567891), resp.Data[1].Created)
	assert.Equal(t, "vllm", resp.Data[1].OwnedBy)
}
