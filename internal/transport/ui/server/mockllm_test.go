package server

import (
	"context"

	"nu/internal/contracts"
)

type MockLLM struct {
	response string
	err      error
}

func (m *MockLLM) Generate(ctx context.Context, prompt string, options ...contracts.GenerateOption) (string, error) {
	if m.err != nil {
		return "", m.err
	}
	return m.response, nil
}

func (m *MockLLM) GenerateWithTools(ctx context.Context, prompt string, tools []contracts.Tool, options ...contracts.GenerateOption) (string, error) {
	return m.Generate(ctx, prompt, options...)
}

func (m *MockLLM) GenerateDetailed(ctx context.Context, prompt string, options ...contracts.GenerateOption) (*contracts.LLMResponse, error) {
	content, err := m.Generate(ctx, prompt, options...)
	if err != nil {
		return nil, err
	}
	return &contracts.LLMResponse{Content: content, Model: m.Name()}, nil
}

func (m *MockLLM) GenerateWithToolsDetailed(ctx context.Context, prompt string, tools []contracts.Tool, options ...contracts.GenerateOption) (*contracts.LLMResponse, error) {
	content, err := m.GenerateWithTools(ctx, prompt, tools, options...)
	if err != nil {
		return nil, err
	}
	return &contracts.LLMResponse{Content: content, Model: m.Name()}, nil
}

func (m *MockLLM) Name() string {
	return "mock-llm"
}

func (m *MockLLM) SupportsStreaming() bool {
	return false
}
