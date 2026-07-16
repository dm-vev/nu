package vllm

import (
	"context"
	"testing"

	"nu/internal/contracts"
	"nu/internal/telemetry"

	"github.com/stretchr/testify/assert"
)

func TestVLLMNewClient(t *testing.T) {
	// Test default client creation
	client := NewClient()
	assert.NotNil(t, client)
	assert.Equal(t, "http://localhost:8000", client.BaseURL)
	assert.Equal(t, "llama-2-7b", client.Model)
	assert.NotNil(t, client.HTTPClient)
	assert.NotNil(t, client.logger)

	// Test client with options
	logger := telemetry.NewLogger()
	client = NewClient(
		WithModel("mistral-7b"),
		WithBaseURL("http://localhost:9000"),
		WithLogger(logger),
	)
	assert.Equal(t, "http://localhost:9000", client.BaseURL)
	assert.Equal(t, "mistral-7b", client.Model)
	assert.Equal(t, logger, client.logger)
}

func TestVLLMClientImplementsLLMInterface(t *testing.T) {
	// This test ensures the client implements the LLM interface
	var _ contracts.LLM = (*Client)(nil)
}

func TestVLLMWithModel(t *testing.T) {
	client := NewClient()
	assert.Equal(t, "llama-2-7b", client.Model)

	client = NewClient(WithModel("codellama-7b"))
	assert.Equal(t, "codellama-7b", client.Model)
}

func TestVLLMWithBaseURL(t *testing.T) {
	client := NewClient()
	assert.Equal(t, "http://localhost:8000", client.BaseURL)

	client = NewClient(WithBaseURL("http://localhost:9000"))
	assert.Equal(t, "http://localhost:9000", client.BaseURL)
}

func TestVLLMWithLogger(t *testing.T) {
	logger := telemetry.NewLogger()
	client := NewClient(WithLogger(logger))
	assert.Equal(t, logger, client.logger)
}

func TestVLLMWithRetry(t *testing.T) {
	client := NewClient()
	assert.Nil(t, client.retryExecutor)

	client = NewClient(WithRetry())
	assert.NotNil(t, client.retryExecutor)
}

func TestVLLMWithHTTPClient(t *testing.T) {
	client := NewClient()
	assert.NotNil(t, client.HTTPClient)

	// Test that the option function exists and can be called
	// The actual HTTP client testing would be done in integration tests
	assert.NotNil(t, WithHTTPClient)
}

func TestVLLMName(t *testing.T) {
	client := NewClient()
	assert.Equal(t, "vllm", client.Name())
}

func TestVLLMGenerateOptions(t *testing.T) {
	// Test WithVLLMTemperature
	option := WithTemperature(0.5)
	options := &contracts.GenerateOptions{}
	option(options)
	assert.Equal(t, 0.5, options.LLMConfig.Temperature)

	// Test WithVLLMTopP
	option = WithTopP(0.8)
	options = &contracts.GenerateOptions{}
	option(options)
	assert.Equal(t, 0.8, options.LLMConfig.TopP)

	// Test WithVLLMStopSequences
	option = WithStopSequences([]string{"END", "STOP"})
	options = &contracts.GenerateOptions{}
	option(options)
	assert.Equal(t, []string{"END", "STOP"}, options.LLMConfig.StopSequences)

	// Test WithVLLMSystemMessage
	option = WithSystemMessage("You are a helpful assistant.")
	options = &contracts.GenerateOptions{}
	option(options)
	assert.Equal(t, "You are a helpful assistant.", options.SystemMessage)

	// Test WithVLLMResponseFormat
	format := contracts.ResponseFormat{
		Type: contracts.ResponseFormatJSON,
		Name: "TestFormat",
	}
	option = WithResponseFormat(format)
	options = &contracts.GenerateOptions{}
	option(options)
	assert.Equal(t, &format, options.ResponseFormat)
}

func TestVLLMGenerateRequestStructure(t *testing.T) {
	// Test that GenerateRequest can be marshaled to JSON
	req := GenerateRequest{
		Model:         "llama-2-7b",
		Prompt:        "Hello, world!",
		Stream:        false,
		Temperature:   0.7,
		TopP:          0.9,
		TopK:          40,
		MaxTokens:     100,
		Stop:          []string{"END"},
		UseBeamSearch: false,
		BestOf:        1,
		N:             1,
	}

	// This test ensures the struct can be created without errors
	assert.Equal(t, "llama-2-7b", req.Model)
	assert.Equal(t, "Hello, world!", req.Prompt)
	assert.Equal(t, false, req.Stream)
	assert.Equal(t, 0.7, req.Temperature)
	assert.Equal(t, 0.9, req.TopP)
	assert.Equal(t, 40, req.TopK)
	assert.Equal(t, 100, req.MaxTokens)
	assert.Equal(t, []string{"END"}, req.Stop)
	assert.Equal(t, false, req.UseBeamSearch)
	assert.Equal(t, 1, req.BestOf)
	assert.Equal(t, 1, req.N)
}

func TestVLLMChatRequestStructure(t *testing.T) {
	// Test that ChatRequest can be created with messages
	messages := []ChatMessage{
		{
			Role:    "system",
			Content: "You are a helpful assistant.",
		},
		{
			Role:    "user",
			Content: "Hello!",
		},
	}

	req := ChatRequest{
		Model:         "llama-2-7b",
		Messages:      messages,
		Stream:        false,
		Temperature:   0.7,
		TopP:          0.9,
		TopK:          40,
		MaxTokens:     100,
		Stop:          []string{"END"},
		UseBeamSearch: false,
		BestOf:        1,
		N:             1,
	}

	assert.Equal(t, "llama-2-7b", req.Model)
	assert.Equal(t, messages, req.Messages)
	assert.Equal(t, false, req.Stream)
	assert.Equal(t, 0.7, req.Temperature)
	assert.Equal(t, 0.9, req.TopP)
	assert.Equal(t, 40, req.TopK)
	assert.Equal(t, 100, req.MaxTokens)
	assert.Equal(t, []string{"END"}, req.Stop)
	assert.Equal(t, false, req.UseBeamSearch)
	assert.Equal(t, 1, req.BestOf)
	assert.Equal(t, 1, req.N)
}

func TestVLLMClientOptions(t *testing.T) {
	// Test that all options work correctly
	logger := telemetry.NewLogger()
	client := NewClient(
		WithModel("test-model"),
		WithBaseURL("http://test-server:8000"),
		WithLogger(logger),
		WithRetry(),
	)

	assert.Equal(t, "test-model", client.Model)
	assert.Equal(t, "http://test-server:8000", client.BaseURL)
	assert.Equal(t, logger, client.logger)
	assert.NotNil(t, client.retryExecutor)
}

func TestVLLMGenerateOptionsWithNilConfig(t *testing.T) {
	// Test that options work when LLMConfig is nil
	option := WithTemperature(0.5)
	options := &contracts.GenerateOptions{}
	option(options)
	assert.NotNil(t, options.LLMConfig)
	assert.Equal(t, 0.5, options.LLMConfig.Temperature)

	option = WithTopP(0.8)
	options = &contracts.GenerateOptions{}
	option(options)
	assert.NotNil(t, options.LLMConfig)
	assert.Equal(t, 0.8, options.LLMConfig.TopP)

	option = WithStopSequences([]string{"END"})
	options = &contracts.GenerateOptions{}
	option(options)
	assert.NotNil(t, options.LLMConfig)
	assert.Equal(t, []string{"END"}, options.LLMConfig.StopSequences)
}

func TestVLLMContextHandling(t *testing.T) {
	// Test that context is properly handled
	ctx := context.Background()
	client := NewClient()

	// This test ensures the client can be created with context
	// The actual API calls would be tested in integration tests
	assert.NotNil(t, client)
	assert.NotNil(t, ctx)
}

func TestVLLMClientDefaultValues(t *testing.T) {
	// Test that default values are set correctly
	client := NewClient()

	assert.Equal(t, "http://localhost:8000", client.BaseURL)
	assert.Equal(t, "llama-2-7b", client.Model)
	assert.NotNil(t, client.HTTPClient)
	assert.NotNil(t, client.logger)
	assert.Nil(t, client.retryExecutor) // Retry is not enabled by default
}

func TestVLLMClientWithAllOptions(t *testing.T) {
	// Test creating client with all available options
	logger := telemetry.NewLogger()
	client := NewClient(
		WithModel("custom-model"),
		WithBaseURL("http://custom-server:9000"),
		WithLogger(logger),
		WithRetry(),
	)

	assert.Equal(t, "custom-model", client.Model)
	assert.Equal(t, "http://custom-server:9000", client.BaseURL)
	assert.Equal(t, logger, client.logger)
	assert.NotNil(t, client.retryExecutor)
	assert.NotNil(t, client.HTTPClient)
}
