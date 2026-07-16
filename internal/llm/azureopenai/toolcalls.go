package azureopenai

import (
	"context"
	"fmt"
	"time"

	"github.com/dm-vev/nu/contracts"
	"github.com/dm-vev/nu/telemetry"

	"github.com/openai/openai-go/v2"
)

func (c *Client) processToolCalls(
	ctx context.Context,
	toolCalls []openai.ChatCompletionMessageToolCallUnion,
	state *azureOpenAIToolExecutionState,
	messages *[]openai.ChatCompletionMessageParamUnion,
	resp *openai.ChatCompletion,
) error {
	for _, toolCall := range toolCalls {
		if toolCall.Function.Name == "multi_tool_use.parallel" {
			c.logger.Info(ctx, "Replacing multi_tool_use.parallel with parallel_tool_use", nil)
			toolCall.Function.Name = "parallel_tool_use"
		}
		if toolCall.Function.Name == "parallel_tool_use" {
			message, err := c.executeParallelToolCall(ctx, toolCall, state)
			if err != nil {
				return err
			}
			if message != nil {
				*messages = append(*messages, *message)
			}
			continue
		}

		message, toolCallTrace := c.executeSingleToolCall(ctx, toolCall, state, resp)
		*messages = append(*messages, message)
		if toolCallTrace != nil {
			telemetry.AddToolCallToContext(ctx, *toolCallTrace)
		}
	}
	return nil
}

func (c *Client) executeSingleToolCall(
	ctx context.Context,
	toolCall openai.ChatCompletionMessageToolCallUnion,
	state *azureOpenAIToolExecutionState,
	resp *openai.ChatCompletion,
) (openai.ChatCompletionMessageParamUnion, *telemetry.ToolCall) {
	var selectedTool contracts.Tool
	for _, tool := range state.tools {
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
		azureOpenAIStoreToolResult(ctx, state.params.Memory, toolCall.ID, toolCall.Function.Name, toolCall.Function.Name, toolCall.Function.Arguments, errorMessage, fmt.Errorf("tool not found: %s", toolCall.Function.Name))
		telemetry.AddToolCallToContext(ctx, telemetry.ToolCall{
			Name: toolCall.Function.Name, Arguments: toolCall.Function.Arguments, ID: toolCall.ID,
			Timestamp: time.Now().Format(time.RFC3339), StartTime: time.Now(), Duration: 0, DurationMs: 0,
			Error: fmt.Sprintf("tool not found: %s", toolCall.Function.Name), Result: errorMessage,
		})
		return openai.ToolMessage(errorMessage, toolCall.ID), nil
	}

	c.logger.Info(ctx, "Executing tool", map[string]interface{}{"toolName": selectedTool.Name()})
	toolStartTime := time.Now()
	toolResult, err := selectedTool.Execute(ctx, toolCall.Function.Arguments)
	toolEndTime := time.Now()

	cacheKey := toolCall.Function.Name + ":" + toolCall.Function.Arguments
	state.historyLock.Lock()
	state.history[cacheKey]++
	callCount := state.history[cacheKey]
	state.historyLock.Unlock()
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
	azureOpenAIStoreToolResult(ctx, state.params.Memory, toolCall.ID, toolCall.Function.Name, toolCall.Function.Name, toolCall.Function.Arguments, toolResult, err)
	if err != nil {
		c.logger.Error(ctx, "Error executing tool", map[string]interface{}{"toolName": selectedTool.Name(), "error": err.Error()})
		toolCallTrace.Error = err.Error()
		toolCallTrace.Result = fmt.Sprintf("Error: %v", err)
		return openai.ToolMessage(fmt.Sprintf("Error: %v", err), toolCall.ID), &toolCallTrace
	}
	toolCallTrace.Result = toolResult
	return openai.ToolMessage(toolResult, toolCall.ID), &toolCallTrace
}

func azureOpenAIStoreToolResult(ctx context.Context, memory contracts.Memory, id, name, metadataName, arguments, result string, err error) {
	if memory == nil {
		return
	}
	_ = memory.AddMessage(ctx, contracts.Message{
		Role: "assistant", Content: "", ToolCalls: []contracts.ToolCall{{ID: id, Name: name, Arguments: arguments}},
	})
	if err != nil {
		result = fmt.Sprintf("Error: %v", err)
	}
	_ = memory.AddMessage(ctx, contracts.Message{
		Role: "tool", Content: result, ToolCallID: id, Metadata: map[string]interface{}{"tool_name": metadataName},
	})
}
