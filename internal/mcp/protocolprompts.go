package mcp

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"nu/internal/contracts"
)

// ListPrompts lists the prompts available on the MCP server
func (s *Server) ListPrompts(ctx context.Context) ([]contracts.MCPPrompt, error) {
	s.logger.Debug(ctx, "Listing MCP prompts", nil)
	resp, err := s.session.ListPrompts(ctx, &mcp.ListPromptsParams{})
	if err != nil {
		mcpErr := ClassifyError(err, "ListPrompts", "server", "unknown")
		// govulncheck:ignore GO-2025-4155 - err.Error() used for logging only, not exploitable
		s.logger.Error(ctx, "Failed to list MCP prompts", map[string]interface{}{
			"error":      err.Error(),
			"error_type": mcpErr.ErrorType,
			"retryable":  mcpErr.Retryable,
		})
		return nil, mcpErr
	}

	prompts := make([]contracts.MCPPrompt, 0, len(resp.Prompts))
	for _, p := range resp.Prompts {
		prompt := contracts.MCPPrompt{
			Name:        p.Name,
			Description: p.Description,
			Arguments:   make([]contracts.MCPPromptArgument, 0, len(p.Arguments)),
			Metadata:    make(map[string]string),
		}
		for _, arg := range p.Arguments {
			prompt.Arguments = append(prompt.Arguments, contracts.MCPPromptArgument{
				Name: arg.Name, Description: arg.Description, Required: arg.Required,
			})
		}
		prompts = append(prompts, prompt)
	}
	s.logger.Info(ctx, "Successfully listed MCP prompts", map[string]interface{}{
		"prompt_count": len(prompts),
	})
	return prompts, nil
}

// GetPrompt retrieves a specific prompt with variables
func (s *Server) GetPrompt(ctx context.Context, name string, variables map[string]interface{}) (*contracts.MCPPromptResult, error) {
	s.logger.Debug(ctx, "Getting MCP prompt", map[string]interface{}{
		"name": name, "variables": variables,
	})
	args := make(map[string]string)
	for k, v := range variables {
		if str, ok := v.(string); ok {
			args[k] = str
		} else {
			args[k] = fmt.Sprintf("%v", v)
		}
	}
	resp, err := s.session.GetPrompt(ctx, &mcp.GetPromptParams{Name: name, Arguments: args})
	if err != nil {
		mcpErr := ClassifyError(err, "GetPrompt", "server", "unknown")
		_ = mcpErr.WithMetadata("prompt_name", name)
		// govulncheck:ignore GO-2025-4155 - err.Error() used for logging only, not exploitable
		s.logger.Error(ctx, "Failed to get MCP prompt", map[string]interface{}{
			"name": name, "error": err.Error(),
			"error_type": mcpErr.ErrorType, "retryable": mcpErr.Retryable,
		})
		return nil, mcpErr
	}

	result := &contracts.MCPPromptResult{
		Variables: variables,
		Messages:  make([]contracts.MCPPromptMessage, 0, len(resp.Messages)),
		Metadata:  make(map[string]string),
	}
	for _, msg := range resp.Messages {
		message := contracts.MCPPromptMessage{Role: string(msg.Role)}
		if msg.Content != nil {
			if textContent, ok := msg.Content.(*mcp.TextContent); ok {
				message.Content = textContent.Text
			} else {
				message.Content = fmt.Sprintf("%v", msg.Content)
			}
		}
		result.Messages = append(result.Messages, message)
	}
	if len(result.Messages) == 1 {
		result.Prompt = result.Messages[0].Content
	}
	s.logger.Debug(ctx, "Successfully retrieved MCP prompt", map[string]interface{}{
		"name": name, "message_count": len(result.Messages),
	})
	return result, nil
}
