package anthropic

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"nu/internal/contracts"
)

type anthropicToolExecResult struct {
	index    int
	result   ToolResult
	toolName string
	toolJSON string
	err      error
}

func (c *Client) executeToolsParallel(ctx context.Context, toolCalls []ToolUse, tools []contracts.Tool, params *contracts.GenerateOptions, toolCallHistory map[string]int, iteration int) []ToolResult {
	if len(toolCalls) == 0 {
		return nil
	}
	c.logger.Info(ctx, "Executing tools in parallel", map[string]interface{}{"count": len(toolCalls), "iteration": iteration + 1})
	resultChan := make(chan anthropicToolExecResult, len(toolCalls))
	var wg sync.WaitGroup
	var historyMu sync.Mutex

	for i, toolCall := range toolCalls {
		toolName := toolCall.Name
		if toolName == "" {
			toolName = toolCall.RecipientName
		}
		if toolName == "" {
			c.logger.Error(ctx, "Tool call missing both Name and RecipientName", map[string]interface{}{"iteration": iteration + 1})
			resultChan <- anthropicToolExecResult{index: i, result: ToolResult{Type: "tool_result", Content: "Error: tool call missing name", ToolName: "unknown"}}
			continue
		}

		var selectedTool contracts.Tool
		for _, tool := range tools {
			if tool.Name() == toolName {
				selectedTool = tool
				break
			}
		}
		if selectedTool == nil {
			c.logger.Error(ctx, "Tool not found", map[string]interface{}{"toolName": toolName, "iteration": iteration + 1})
			errorMessage := fmt.Sprintf("Error: tool not found: %s", toolName)
			if params.Memory != nil {
				_ = params.Memory.AddMessage(ctx, contracts.Message{
					Role: "assistant", Content: "", ToolCalls: []contracts.ToolCall{{ID: toolCall.ID, Name: toolName, Arguments: "{}"}},
				})
				_ = params.Memory.AddMessage(ctx, contracts.Message{
					Role: "tool", Content: errorMessage, ToolCallID: toolCall.ID, Metadata: map[string]interface{}{"tool_name": toolName},
				})
			}
			resultChan <- anthropicToolExecResult{index: i, result: ToolResult{Type: "tool_result", Content: errorMessage, ToolName: toolName}, toolName: toolName}
			continue
		}

		var parameters map[string]interface{}
		if len(toolCall.Input) > 0 {
			parameters = toolCall.Input
		} else if len(toolCall.Parameters) > 0 {
			parameters = toolCall.Parameters
		}
		toolCallJSON, err := json.Marshal(parameters)
		if err != nil {
			c.logger.Error(ctx, "Error marshalling parameters", map[string]interface{}{"error": err.Error(), "iteration": iteration + 1})
			resultChan <- anthropicToolExecResult{
				index: i, result: ToolResult{Type: "tool_result", Content: fmt.Sprintf("Error: %v", err), ToolName: toolName},
				toolName: toolName, err: err,
			}
			continue
		}

		wg.Add(1)
		go func(idx int, tc ToolUse, tool contracts.Tool, tName string, tJSON []byte) {
			defer wg.Done()
			c.logger.Debug(ctx, "Tool parameters", map[string]interface{}{
				"toolName": tName, "parameters": string(tJSON), "iteration": iteration + 1,
			})
			c.logger.Info(ctx, "Executing tool (parallel)", map[string]interface{}{"toolName": tName, "iteration": iteration + 1})
			toolResult, execErr := tool.Execute(ctx, string(tJSON))

			historyMu.Lock()
			cacheKey := tName + ":" + string(tJSON)
			toolCallHistory[cacheKey]++
			callCount := toolCallHistory[cacheKey]
			historyMu.Unlock()
			if callCount > 2 {
				warning := fmt.Sprintf("\n\n[WARNING: This is call #%d to %s with identical parameters. You may be in a loop.]", callCount, tName)
				if execErr == nil {
					toolResult += warning
				}
				c.logger.Warn(ctx, "Repetitive tool call detected", map[string]interface{}{
					"toolName": tName, "callCount": callCount, "iteration": iteration + 1,
				})
			}

			if params.Memory != nil {
				_ = params.Memory.AddMessage(ctx, contracts.Message{
					Role: "assistant", Content: "", ToolCalls: []contracts.ToolCall{{ID: tc.ID, Name: tName, Arguments: string(tJSON)}},
				})
				if execErr != nil {
					_ = params.Memory.AddMessage(ctx, contracts.Message{
						Role: "tool", Content: fmt.Sprintf("Error: %v", execErr), ToolCallID: tc.ID, Metadata: map[string]interface{}{"tool_name": tName},
					})
				} else {
					_ = params.Memory.AddMessage(ctx, contracts.Message{
						Role: "tool", Content: toolResult, ToolCallID: tc.ID, Metadata: map[string]interface{}{"tool_name": tName},
					})
				}
			}

			if execErr != nil {
				c.logger.Error(ctx, "Error executing tool", map[string]interface{}{
					"toolName": tName, "error": execErr.Error(), "iteration": iteration + 1,
				})
				resultChan <- anthropicToolExecResult{
					index: idx, result: ToolResult{Type: "tool_result", Content: fmt.Sprintf("Error: %v", execErr), ToolName: tName},
					toolName: tName, toolJSON: string(tJSON), err: execErr,
				}
				return
			}
			resultChan <- anthropicToolExecResult{
				index: idx, result: ToolResult{Type: "tool_result", Content: toolResult, ToolName: tName},
				toolName: tName, toolJSON: string(tJSON),
			}
		}(i, toolCall, selectedTool, toolName, toolCallJSON)
	}

	go func() { wg.Wait(); close(resultChan) }()
	results := make([]anthropicToolExecResult, 0, len(toolCalls))
	for result := range resultChan {
		results = append(results, result)
	}
	toolResults := make([]ToolResult, len(results))
	for _, r := range results {
		if r.index < len(toolResults) {
			toolResults[r.index] = r.result
		}
	}
	filteredResults := make([]ToolResult, 0, len(toolResults))
	for _, r := range toolResults {
		if r.ToolName != "" {
			filteredResults = append(filteredResults, r)
		}
	}
	c.logger.Info(ctx, "Parallel tool execution completed", map[string]interface{}{
		"count": len(filteredResults), "iteration": iteration + 1,
	})
	return filteredResults
}

func (c *Client) buildMessagesWithMemory(ctx context.Context, prompt string, params *contracts.GenerateOptions) []Message {
	builder := anthropicNewMessageHistoryBuilder(c.logger)
	return builder.buildMessages(ctx, prompt, params)
}
