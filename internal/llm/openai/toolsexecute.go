package openai

import (
	"context"
	"fmt"
	"sync"
	"time"

	"nu/internal/contracts"
	"nu/internal/telemetry"

	"github.com/openai/openai-go/v2"
)

func (c *Client) executeToolCall(
	ctx context.Context,
	toolCall openai.ChatCompletionMessageToolCallUnion,
	tools []contracts.Tool,
	memory contracts.Memory,
	history map[string]int,
	historyMu *sync.Mutex,
	resp *openai.ChatCompletion,
) openai.ChatCompletionMessageParamUnion {
	var selectedTool contracts.Tool
	for _, tool := range tools {
		if tool.Name() == toolCall.Function.Name {
			selectedTool = tool
			break
		}
	}
	if selectedTool == nil || selectedTool.Name() == "" {
		c.logger.Error(ctx, "Tool not found", map[string]interface{}{
			"toolName": toolCall.Function.Name, "toolcall": toolCall, "resp": resp,
		})
		errorMessage := fmt.Sprintf("Error: tool not found: %s", toolCall.Function.Name)
		if memory != nil {
			_ = memory.AddMessage(ctx, contracts.Message{
				Role: "assistant", Content: "",
				ToolCalls: []contracts.ToolCall{{ID: toolCall.ID, Name: toolCall.Function.Name, Arguments: toolCall.Function.Arguments}},
			})
			_ = memory.AddMessage(ctx, contracts.Message{
				Role: "tool", Content: errorMessage, ToolCallID: toolCall.ID,
				Metadata: map[string]interface{}{"tool_name": toolCall.Function.Name},
			})
		}
		now := time.Now()
		telemetry.AddToolCallToContext(ctx, telemetry.ToolCall{
			Name: toolCall.Function.Name, Arguments: toolCall.Function.Arguments, ID: toolCall.ID,
			Timestamp: now.Format(time.RFC3339), StartTime: now, Duration: 0, DurationMs: 0,
			Error: fmt.Sprintf("tool not found: %s", toolCall.Function.Name), Result: errorMessage,
		})
		return openai.ToolMessage(errorMessage, toolCall.ID)
	}

	c.logger.Info(ctx, "Executing tool", map[string]interface{}{"toolName": selectedTool.Name()})
	toolStartTime := time.Now()
	toolResult, err := selectedTool.Execute(ctx, toolCall.Function.Arguments)
	toolEndTime := time.Now()
	cacheKey := toolCall.Function.Name + ":" + toolCall.Function.Arguments
	historyMu.Lock()
	history[cacheKey]++
	callCount := history[cacheKey]
	historyMu.Unlock()
	if callCount > 1 {
		warning := fmt.Sprintf("\n\n[WARNING: This is call #%d to %s with identical parameters. You may be in a loop. Consider using the available information to provide a final answer.]", callCount, toolCall.Function.Name)
		if err == nil {
			toolResult += warning
		}
		c.logger.Warn(ctx, "Repetitive tool call detected", map[string]interface{}{"toolName": toolCall.Function.Name, "callCount": callCount})
	}

	executionDuration := toolEndTime.Sub(toolStartTime)
	toolCallTrace := telemetry.ToolCall{
		Name: toolCall.Function.Name, Arguments: toolCall.Function.Arguments, ID: toolCall.ID,
		Timestamp: toolStartTime.Format(time.RFC3339), StartTime: toolStartTime,
		Duration: executionDuration, DurationMs: executionDuration.Milliseconds(),
	}
	if memory != nil {
		content := toolResult
		if err != nil {
			content = fmt.Sprintf("Error: %v", err)
		}
		_ = memory.AddMessage(ctx, contracts.Message{
			Role: "assistant", Content: "",
			ToolCalls: []contracts.ToolCall{{ID: toolCall.ID, Name: toolCall.Function.Name, Arguments: toolCall.Function.Arguments}},
		})
		_ = memory.AddMessage(ctx, contracts.Message{
			Role: "tool", Content: content, ToolCallID: toolCall.ID,
			Metadata: map[string]interface{}{"tool_name": toolCall.Function.Name},
		})
	}

	if err != nil {
		c.logger.Error(ctx, "Error executing tool", map[string]interface{}{"toolName": selectedTool.Name(), "error": err.Error()})
		toolCallTrace.Error = err.Error()
		toolCallTrace.Result = fmt.Sprintf("Error: %v", err)
		telemetry.AddToolCallToContext(ctx, toolCallTrace)
		return openai.ToolMessage(fmt.Sprintf("Error: %v", err), toolCall.ID)
	}
	toolCallTrace.Result = toolResult
	telemetry.AddToolCallToContext(ctx, toolCallTrace)
	return openai.ToolMessage(toolResult, toolCall.ID)
}
