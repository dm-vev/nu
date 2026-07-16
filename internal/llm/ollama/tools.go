package ollama

import (
	"context"
	"encoding/json"
	"fmt"

	"nu/internal/contracts"
)

// GenerateWithTools generates text using Ollama's native /api/chat tool support.
// The model is given the full tool list and may invoke any subset; we execute each
// returned tool_call and feed the result back as a tool message until the model
// returns a final answer with no further tool calls (#202).
func (c *Client) GenerateWithTools(ctx context.Context, prompt string, tools []contracts.Tool, options ...contracts.GenerateOption) (string, error) {
	if len(tools) == 0 {
		return c.Generate(ctx, prompt, options...)
	}

	params := &contracts.GenerateOptions{
		LLMConfig: &contracts.LLMConfig{Temperature: 0.7},
	}
	for _, option := range options {
		option(params)
	}

	// Build initial conversation. Memory history is inlined into the user
	// prompt (same convention as Generate) so we don't need a separate
	// per-turn history schema for the chat endpoint.
	messages := make([]ChatMessage, 0, 4)
	if params.SystemMessage != "" {
		messages = append(messages, ChatMessage{Role: "system", Content: params.SystemMessage})
	}
	messages = append(messages, ChatMessage{
		Role:    "user",
		Content: c.buildPromptWithMemory(ctx, prompt, params),
	})

	// Convert agent tools to Ollama function declarations
	ollamaTools := make([]Tool, 0, len(tools))
	for _, t := range tools {
		ollamaTools = append(ollamaTools, Tool{
			Type: "function",
			Function: Function{
				Name:        t.Name(),
				Description: t.Description(),
				Parameters:  ollamaToolParametersToJSONSchema(t.Parameters()),
			},
		})
	}

	maxIterations := params.MaxIterations
	if maxIterations <= 0 {
		maxIterations = 10
	}

	for iter := 0; iter < maxIterations; iter++ {
		req := ChatRequest{
			Model:    c.Model,
			Messages: messages,
			Stream:   false,
			Tools:    ollamaTools,
			Options: &Options{
				Temperature: params.LLMConfig.Temperature,
				TopP:        params.LLMConfig.TopP,
				Stop:        params.LLMConfig.StopSequences,
			},
		}

		resp, err := c.makeRequest(ctx, "/api/chat", req)
		if err != nil {
			return "", fmt.Errorf("failed to chat with tools: %w", err)
		}

		var chatResp ChatResponse
		if err := json.Unmarshal(resp, &chatResp); err != nil {
			return "", fmt.Errorf("failed to unmarshal tool-chat response: %w", err)
		}

		// No tool calls means the model produced its final answer.
		if len(chatResp.Message.ToolCalls) == 0 {
			return chatResp.Message.Content, nil
		}

		// Persist the assistant message that requested the tool calls.
		messages = append(messages, chatResp.Message)

		// Synthesize one ID per tool_call so the same tool invoked twice
		// in a single turn doesn't collide on a shared "ollama:<name>" key.
		// IDs are stable for the duration of the loop iteration; they're
		// persisted on both the assistant ToolCall and the corresponding
		// tool-result message so consumers can pair them later.
		callIDs := make([]string, len(chatResp.Message.ToolCalls))
		for idx, call := range chatResp.Message.ToolCalls {
			callIDs[idx] = fmt.Sprintf("ollama:%s:%d:%d", call.Function.Name, iter, idx)
		}

		// Mirror the assistant tool-call message into Memory so subsequent
		// agent turns can see the tool exchanges (matches OpenAI client
		// convention; addresses the #325 review BLOCKER on memory loss).
		if params.Memory != nil {
			toolCallSummaries := make([]contracts.ToolCall, 0, len(chatResp.Message.ToolCalls))
			for idx, call := range chatResp.Message.ToolCalls {
				argsBytes, _ := json.Marshal(call.Function.Arguments)
				toolCallSummaries = append(toolCallSummaries, contracts.ToolCall{
					ID:        callIDs[idx],
					Name:      call.Function.Name,
					Arguments: string(argsBytes),
				})
			}
			_ = params.Memory.AddMessage(ctx, contracts.Message{
				Role:      contracts.RoleAssistant,
				Content:   chatResp.Message.Content,
				ToolCalls: toolCallSummaries,
			})
		}

		// Execute each tool call and append its result as a tool message.
		for idx, call := range chatResp.Message.ToolCalls {
			callID := callIDs[idx]
			tool := ollamaFindToolByName(tools, call.Function.Name)
			if tool == nil {
				errMsg := fmt.Sprintf("error: tool %q not found", call.Function.Name)
				messages = append(messages, ChatMessage{Role: "tool", Content: errMsg})
				ollamaPersistToolResultMessage(ctx, params.Memory, callID, call.Function.Name, errMsg)
				continue
			}

			argsJSON, err := json.Marshal(call.Function.Arguments)
			if err != nil {
				errMsg := fmt.Sprintf("error: failed to encode arguments: %v", err)
				messages = append(messages, ChatMessage{Role: "tool", Content: errMsg})
				ollamaPersistToolResultMessage(ctx, params.Memory, callID, call.Function.Name, errMsg)
				continue
			}

			result, err := tool.Execute(ctx, string(argsJSON))
			if err != nil {
				errMsg := fmt.Sprintf("error: %v", err)
				messages = append(messages, ChatMessage{Role: "tool", Content: errMsg})
				ollamaPersistToolResultMessage(ctx, params.Memory, callID, call.Function.Name, errMsg)
				continue
			}

			messages = append(messages, ChatMessage{Role: "tool", Content: result})
			ollamaPersistToolResultMessage(ctx, params.Memory, callID, call.Function.Name, result)
		}
	}

	return "", fmt.Errorf("ollama tool loop exceeded max iterations (%d)", maxIterations)
}

// ollamaToolParametersToJSONSchema converts the SDK's ParameterSpec map into the JSON
// Schema object Ollama expects under function.parameters.
func ollamaToolParametersToJSONSchema(params map[string]contracts.ParameterSpec) map[string]interface{} {
	properties := make(map[string]interface{}, len(params))
	required := make([]string, 0)
	for name, spec := range params {
		field := map[string]interface{}{}
		if spec.Type != nil {
			field["type"] = spec.Type
		}
		if spec.Description != "" {
			field["description"] = spec.Description
		}
		if spec.Default != nil {
			field["default"] = spec.Default
		}
		if spec.Enum != nil {
			field["enum"] = spec.Enum
		}
		if spec.Items != nil {
			itemSchema := map[string]interface{}{}
			if spec.Items.Type != nil {
				itemSchema["type"] = spec.Items.Type
			}
			if spec.Items.Description != "" {
				itemSchema["description"] = spec.Items.Description
			}
			if spec.Items.Enum != nil {
				itemSchema["enum"] = spec.Items.Enum
			}
			field["items"] = itemSchema
		}
		properties[name] = field
		if spec.Required {
			required = append(required, name)
		}
	}
	schema := map[string]interface{}{
		"type":       "object",
		"properties": properties,
	}
	if len(required) > 0 {
		schema["required"] = required
	}
	return schema
}

func ollamaFindToolByName(tools []contracts.Tool, name string) contracts.Tool {
	for _, t := range tools {
		if t.Name() == name {
			return t
		}
	}
	return nil
}
