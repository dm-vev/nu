package anthropic

import (
	"context"
	"fmt"
	"time"

	"github.com/dm-vev/nu/contracts"
	"github.com/dm-vev/nu/internal/multitenancy"
)

// GenerateWithToolsStream implements contracts.StreamingLLM.GenerateWithToolsStream
func (c *Client) GenerateWithToolsStream(ctx context.Context, prompt string, tools []contracts.Tool, options ...contracts.GenerateOption) (<-chan contracts.StreamEvent, error) {
	c.logger.Debug(ctx, "[LLM RESPONSE DEBUG] GenerateWithToolsStream called (WITH TOOLS)", map[string]interface{}{
		"model": c.Model, "promptLength": len(prompt), "toolsCount": len(tools),
	})
	if c.Model == "" {
		return nil, fmt.Errorf("model not specified: use WithModel option when creating the client")
	}

	params := &contracts.GenerateOptions{LLMConfig: &contracts.LLMConfig{Temperature: 0.7}}
	for _, opt := range options {
		if opt != nil {
			opt(params)
		}
	}
	if id, err := multitenancy.GetOrgID(ctx); err == nil {
		ctx = multitenancy.WithOrgID(ctx, id)
	} else {
		ctx = multitenancy.WithOrgID(ctx, "default")
	}

	anthropicTools := make([]Tool, len(tools))
	for i, tool := range tools {
		properties := make(map[string]interface{})
		required := []string{}
		for name, param := range tool.Parameters() {
			properties[name] = map[string]interface{}{"type": param.Type, "description": param.Description}
			if param.Default != nil {
				properties[name].(map[string]interface{})["default"] = param.Default
			}
			if param.Required {
				required = append(required, name)
			}
			if param.Items != nil {
				properties[name].(map[string]interface{})["items"] = map[string]interface{}{"type": param.Items.Type}
				if param.Items.Enum != nil {
					properties[name].(map[string]interface{})["items"].(map[string]interface{})["enum"] = param.Items.Enum
				}
			}
			if param.Enum != nil {
				properties[name].(map[string]interface{})["enum"] = param.Enum
			}
		}
		anthropicTools[i] = Tool{
			Name: tool.Name(), Description: tool.Description(),
			InputSchema: map[string]interface{}{"type": "object", "properties": properties, "required": required},
		}
	}

	bufferSize := 100
	if params.StreamConfig != nil {
		bufferSize = params.StreamConfig.BufferSize
	}
	eventChan := make(chan contracts.StreamEvent, bufferSize)
	go func() {
		defer func() {
			defer func() { _ = recover() }()
			close(eventChan)
		}()
		if err := c.executeStreamingWithTools(ctx, prompt, anthropicTools, tools, params, eventChan); err != nil {
			select {
			case eventChan <- contracts.StreamEvent{Type: contracts.StreamEventError, Error: err, Timestamp: time.Now()}:
			case <-ctx.Done():
				return
			}
		}
	}()
	return eventChan, nil
}
