package sampling

import (
	"context"
	"fmt"

	"nu/internal/contracts"
	"nu/internal/telemetry"
)

// SamplingManager provides high-level operations for MCP sampling
type SamplingManager struct {
	servers []contracts.MCPServer
	logger  telemetry.Logger
}

// NewSamplingManager creates a new sampling manager
func NewSamplingManager(servers []contracts.MCPServer) *SamplingManager {
	return &SamplingManager{
		servers: servers,
		logger:  telemetry.NewLogger(),
	}
}

// CreateTextMessage creates a simple text message using the first available server
func (sm *SamplingManager) CreateTextMessage(ctx context.Context, prompt string, opts ...SamplingOption) (*contracts.MCPSamplingResponse, error) {
	request := &contracts.MCPSamplingRequest{
		Messages: []contracts.MCPMessage{
			{
				Role: "user",
				Content: contracts.MCPContent{
					Type: "text",
					Text: prompt,
				},
			},
		},
		ModelPreferences: &contracts.MCPModelPreferences{
			IntelligencePriority: 0.7,
			SpeedPriority:        0.5,
			CostPriority:         0.3,
		},
	}

	// Apply options
	for _, opt := range opts {
		opt(request)
	}

	return sm.CreateMessage(ctx, request)
}

// CreateMessage creates a message using the first available server
func (sm *SamplingManager) CreateMessage(ctx context.Context, request *contracts.MCPSamplingRequest) (*contracts.MCPSamplingResponse, error) {
	if len(sm.servers) == 0 {
		return nil, fmt.Errorf("no MCP servers available for sampling")
	}

	// Try each server until one succeeds
	var lastErr error
	for i, server := range sm.servers {
		serverName := fmt.Sprintf("server-%d", i)

		sm.logger.Debug(ctx, "Attempting sampling with server", map[string]interface{}{
			"server": serverName,
		})

		response, err := server.CreateMessage(ctx, request)
		if err != nil {
			sm.logger.Warn(ctx, "Sampling failed on server", map[string]interface{}{
				"server": serverName,
				"error":  err.Error(),
			})
			lastErr = err
			continue
		}

		sm.logger.Debug(ctx, "Sampling succeeded on server", map[string]interface{}{
			"server": serverName,
			"model":  response.Model,
		})

		return response, nil
	}

	return nil, fmt.Errorf("sampling failed on all servers: %w", lastErr)
}

// CreateConversation creates a multi-turn conversation
func (sm *SamplingManager) CreateConversation(ctx context.Context, messages []contracts.MCPMessage, opts ...SamplingOption) (*contracts.MCPSamplingResponse, error) {
	request := &contracts.MCPSamplingRequest{
		Messages: messages,
		ModelPreferences: &contracts.MCPModelPreferences{
			IntelligencePriority: 0.8,
			SpeedPriority:        0.4,
			CostPriority:         0.2,
		},
	}

	// Apply options
	for _, opt := range opts {
		opt(request)
	}

	return sm.CreateMessage(ctx, request)
}

// CreateCodeGeneration creates a message optimized for code generation
func (sm *SamplingManager) CreateCodeGeneration(ctx context.Context, prompt string, language string, opts ...SamplingOption) (*contracts.MCPSamplingResponse, error) {
	systemPrompt := fmt.Sprintf("You are an expert programmer. Generate high-quality %s code based on the user's request.", language)
	if language == "" {
		systemPrompt = "You are an expert programmer. Generate high-quality code based on the user's request."
	}

	request := &contracts.MCPSamplingRequest{
		Messages: []contracts.MCPMessage{
			{
				Role: "user",
				Content: contracts.MCPContent{
					Type: "text",
					Text: prompt,
				},
			},
		},
		SystemPrompt: systemPrompt,
		ModelPreferences: &contracts.MCPModelPreferences{
			IntelligencePriority: 0.9, // Prioritize intelligence for code generation
			SpeedPriority:        0.3,
			CostPriority:         0.1,
		},
	}

	// Apply options
	for _, opt := range opts {
		opt(request)
	}

	return sm.CreateMessage(ctx, request)
}

// CreateSummary creates a message optimized for summarization tasks
func (sm *SamplingManager) CreateSummary(ctx context.Context, content string, maxLength int, opts ...SamplingOption) (*contracts.MCPSamplingResponse, error) {
	prompt := fmt.Sprintf("Please summarize the following content in no more than %d words:\n\n%s", maxLength, content)

	request := &contracts.MCPSamplingRequest{
		Messages: []contracts.MCPMessage{
			{
				Role: "user",
				Content: contracts.MCPContent{
					Type: "text",
					Text: prompt,
				},
			},
		},
		SystemPrompt: "You are a skilled summarizer. Create concise, accurate summaries that capture the key points.",
		ModelPreferences: &contracts.MCPModelPreferences{
			IntelligencePriority: 0.7,
			SpeedPriority:        0.6, // Prioritize speed for summaries
			CostPriority:         0.4,
		},
		MaxTokens: &maxLength,
	}

	// Apply options
	for _, opt := range opts {
		opt(request)
	}

	return sm.CreateMessage(ctx, request)
}

// CreateImageAnalysisMessage creates a message for image analysis
func (sm *SamplingManager) CreateImageAnalysisMessage(ctx context.Context, imageData, mimeType, prompt string, opts ...SamplingOption) (*contracts.MCPSamplingResponse, error) {
	request := &contracts.MCPSamplingRequest{
		Messages: []contracts.MCPMessage{
			{
				Role: "user",
				Content: contracts.MCPContent{
					Type:     "image",
					Data:     imageData, // base64 encoded
					MimeType: mimeType,
				},
			},
			{
				Role: "user",
				Content: contracts.MCPContent{
					Type: "text",
					Text: prompt,
				},
			},
		},
		ModelPreferences: &contracts.MCPModelPreferences{
			IntelligencePriority: 0.9, // High intelligence for image analysis
			SpeedPriority:        0.4,
			CostPriority:         0.2,
		},
	}

	// Apply options
	for _, opt := range opts {
		opt(request)
	}

	return sm.CreateMessage(ctx, request)
}
