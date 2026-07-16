package agent

import (
	"context"
	"encoding/json"
	"fmt"
)

// Execute implements contracts.Tool.Execute
func (at *AgentTool) Execute(ctx context.Context, args string) (string, error) {
	agentName := at.agent.GetName()

	// Log the tool execution start
	at.logger.Debug(ctx, "Sub-agent tool execution started", map[string]interface{}{
		"sub_agent": agentName,
		"tool_name": at.name,
		"raw_args":  args,
	})

	// Parse the JSON arguments
	var params struct {
		Query   string                 `json:"query"`
		Context map[string]interface{} `json:"context,omitempty"`
	}

	if err := json.Unmarshal([]byte(args), &params); err != nil {
		at.logger.Error(ctx, "Failed to parse sub-agent tool arguments", map[string]interface{}{
			"sub_agent": agentName,
			"tool_name": at.name,
			"raw_args":  args,
			"error":     err.Error(),
		})
		return "", fmt.Errorf("failed to parse arguments: %w", err)
	}

	if params.Query == "" {
		at.logger.Error(ctx, "Sub-agent tool called with empty query", map[string]interface{}{
			"sub_agent": agentName,
			"tool_name": at.name,
			"args":      args,
		})
		return "", fmt.Errorf("query parameter is required")
	}

	// Log parsed parameters
	at.logger.Debug(ctx, "Sub-agent tool parameters parsed", map[string]interface{}{
		"sub_agent":      agentName,
		"tool_name":      at.name,
		"parsed_query":   params.Query,
		"parsed_context": params.Context,
	})

	// If context is provided, add it to the context
	if params.Context != nil {
		for key, value := range params.Context {
			ctx = context.WithValue(ctx, contextKey(key), value)
		}
	}

	return at.Run(ctx, params.Query)
}
