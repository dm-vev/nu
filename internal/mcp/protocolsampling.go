package mcp

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"nu/internal/contracts"
)

// CreateMessage requests the client to generate a completion using its LLM
func (s *Server) CreateMessage(ctx context.Context, request *contracts.MCPSamplingRequest) (*contracts.MCPSamplingResponse, error) {
	s.logger.Debug(ctx, "Creating message via sampling", map[string]interface{}{
		"message_count": len(request.Messages),
		"system_prompt": request.SystemPrompt != "",
		"max_tokens":    request.MaxTokens,
	})
	samplingRequest := &mcp.CreateMessageParams{
		Messages: make([]*mcp.SamplingMessage, 0, len(request.Messages)),
	}
	for _, msg := range request.Messages {
		samplingMsg := &mcp.SamplingMessage{Role: mcp.Role(msg.Role)}
		switch msg.Content.Type {
		case "text":
			samplingMsg.Content = &mcp.TextContent{Text: msg.Content.Text}
		case "image":
			var imageData []byte
			if msg.Content.Data != "" {
				imageData = []byte(msg.Content.Data)
			}
			samplingMsg.Content = &mcp.ImageContent{Data: imageData, MIMEType: msg.Content.MimeType}
		default:
			samplingMsg.Content = &mcp.TextContent{Text: msg.Content.Text}
		}
		samplingRequest.Messages = append(samplingRequest.Messages, samplingMsg)
	}
	if request.SystemPrompt != "" {
		samplingRequest.SystemPrompt = request.SystemPrompt
	}
	if request.ModelPreferences != nil {
		samplingRequest.ModelPreferences = &mcp.ModelPreferences{
			CostPriority:         request.ModelPreferences.CostPriority,
			SpeedPriority:        request.ModelPreferences.SpeedPriority,
			IntelligencePriority: request.ModelPreferences.IntelligencePriority,
		}
		if len(request.ModelPreferences.Hints) > 0 {
			hints := make([]*mcp.ModelHint, 0, len(request.ModelPreferences.Hints))
			for _, hint := range request.ModelPreferences.Hints {
				hints = append(hints, &mcp.ModelHint{Name: hint.Name})
			}
			samplingRequest.ModelPreferences.Hints = hints
		}
	}
	if request.MaxTokens != nil {
		samplingRequest.MaxTokens = int64(*request.MaxTokens)
	}
	if request.Temperature != nil {
		samplingRequest.Temperature = *request.Temperature
	}
	if len(request.StopSequences) > 0 {
		samplingRequest.StopSequences = request.StopSequences
	}
	if request.IncludeContext != "" {
		samplingRequest.IncludeContext = request.IncludeContext
	}

	s.logger.Warn(ctx, "MCP Sampling feature not yet implemented in Go SDK", map[string]interface{}{
		"message_count": len(request.Messages),
		"system_prompt": request.SystemPrompt != "",
	})
	return nil, fmt.Errorf("MCP Sampling is not yet available in the Go SDK - this is a placeholder implementation for the 2025 specification")
}
