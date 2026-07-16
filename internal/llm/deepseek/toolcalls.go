package deepseek

import (
	"context"
	"fmt"
	"sync"

	"github.com/dm-vev/nu/contracts"
)

// DeepSeekToolResult represents the result of a tool execution
type ToolResult struct {
	ToolCallID string
	ToolName   string
	Content    string
}

// executeToolsParallel executes multiple tools in parallel
func (c *Client) executeToolsParallel(ctx context.Context, toolCalls []ToolCall, tools []contracts.Tool) []ToolResult {
	type result struct {
		index      int
		toolCallID string
		toolName   string
		content    string
		err        error
	}

	resultCh := make(chan result, len(toolCalls))
	var wg sync.WaitGroup

	// Execute each tool call in a goroutine
	for i, toolCall := range toolCalls {
		wg.Add(1)
		go func(index int, tc ToolCall) {
			defer wg.Done()

			// Find the tool
			var tool contracts.Tool
			for _, t := range tools {
				if t.Name() == tc.Function.Name {
					tool = t
					break
				}
			}

			if tool == nil {
				c.logger.Error(ctx, "Tool not found", map[string]interface{}{
					"tool_name": tc.Function.Name,
				})
				resultCh <- result{
					index:      index,
					toolCallID: tc.ID,
					toolName:   tc.Function.Name,
					content:    fmt.Sprintf("Error: tool '%s' not found", tc.Function.Name),
					err:        fmt.Errorf("tool not found: %s", tc.Function.Name),
				}
				return
			}

			// Execute the tool
			c.logger.Info(ctx, "Executing tool", map[string]interface{}{
				"tool_name": tc.Function.Name,
				"arguments": tc.Function.Arguments,
			})

			content, err := tool.Execute(ctx, tc.Function.Arguments)
			if err != nil {
				c.logger.Error(ctx, "Tool execution failed", map[string]interface{}{
					"tool_name": tc.Function.Name,
					"error":     err.Error(),
				})
				content = fmt.Sprintf("Error executing tool: %v", err)
			}

			resultCh <- result{
				index:      index,
				toolCallID: tc.ID,
				toolName:   tc.Function.Name,
				content:    content,
				err:        nil,
			}
		}(i, toolCall)
	}

	// Wait for all goroutines to finish
	go func() {
		wg.Wait()
		close(resultCh)
	}()

	// Collect results
	results := make([]ToolResult, len(toolCalls))
	for r := range resultCh {
		results[r.index] = ToolResult{
			ToolCallID: r.toolCallID,
			ToolName:   r.toolName,
			Content:    r.content,
		}
	}

	return results
}

// convertToolsToDeepSeekFormat converts SDK tools to DeepSeek API format
func (c *Client) convertToolsToDeepSeekFormat(tools []contracts.Tool) []Tool {
	deepseekTools := make([]Tool, len(tools))

	for i, tool := range tools {
		// Convert parameters to JSON schema
		properties := make(map[string]interface{})
		required := []string{}

		for name, param := range tool.Parameters() {
			propDef := map[string]interface{}{
				"type":        param.Type,
				"description": param.Description,
			}

			if param.Default != nil {
				propDef["default"] = param.Default
			}

			if param.Items != nil {
				propDef["items"] = map[string]interface{}{
					"type": param.Items.Type,
				}
				if param.Items.Enum != nil {
					propDef["items"].(map[string]interface{})["enum"] = param.Items.Enum
				}
			}

			if param.Enum != nil {
				propDef["enum"] = param.Enum
			}

			properties[name] = propDef

			if param.Required {
				required = append(required, name)
			}
		}

		deepseekTools[i] = Tool{
			Type: "function",
			Function: FunctionDef{
				Name:        tool.Name(),
				Description: tool.Description(),
				Parameters: map[string]interface{}{
					"type":       "object",
					"properties": properties,
					"required":   required,
				},
			},
		}
	}

	return deepseekTools
}
