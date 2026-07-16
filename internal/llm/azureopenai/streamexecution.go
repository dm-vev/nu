package azureopenai

import (
	"context"
	"fmt"
	"time"

	"github.com/dm-vev/nu/contracts"
	"github.com/dm-vev/nu/telemetry"

	"github.com/openai/openai-go/v2"
)

func (c *Client) executeStreamToolCalls(run *azureOpenAIStreamToolRun, toolCalls []openai.ChatCompletionMessageToolCallUnion, iteration int) {
	for _, toolCall := range toolCalls {
		var foundTool contracts.Tool
		for _, tool := range run.tools {
			if tool.Name() == toolCall.Function.Name {
				foundTool = tool
				break
			}
		}
		if foundTool == nil {
			c.recordMissingStreamTool(run, toolCall)
			continue
		}

		c.logger.Info(run.ctx, "Executing tool", map[string]interface{}{"toolName": foundTool.Name()})
		toolStartTime := time.Now()
		result, err := foundTool.Execute(run.ctx, toolCall.Function.Arguments)
		toolEndTime := time.Now()
		executionDuration := toolEndTime.Sub(toolStartTime)
		toolCallTrace := telemetry.ToolCall{
			Name: toolCall.Function.Name, Arguments: toolCall.Function.Arguments, ID: toolCall.ID,
			Timestamp: toolStartTime.Format(time.RFC3339), StartTime: toolStartTime,
			Duration: executionDuration, DurationMs: executionDuration.Milliseconds(),
		}
		if err != nil {
			c.logger.Error(run.ctx, "Tool execution error", map[string]interface{}{"tool_name": toolCall.Function.Name, "error": err.Error()})
			result = fmt.Sprintf("Error executing tool: %v", err)
			toolCallTrace.Error = err.Error()
		}
		toolCallTrace.Result = result
		run.events <- contracts.StreamEvent{
			Type: contracts.StreamEventToolResult, Timestamp: time.Now(),
			ToolCall: &contracts.ToolCall{ID: toolCall.ID, Name: toolCall.Function.Name, Arguments: toolCall.Function.Arguments},
			Metadata: map[string]interface{}{"iteration": iteration + 1, "result": result},
		}
		c.logger.Debug(run.ctx, "Adding tool result to conversation", map[string]interface{}{
			"tool_call_id": toolCall.ID, "id_length": len(toolCall.ID),
			"tool_name": toolCall.Function.Name, "result_length": len(result),
		})
		fmt.Printf("DEBUG AzureOpenAI: Adding tool call %s to tracing context\n", toolCallTrace.Name)
		telemetry.AddToolCallToContext(run.ctx, toolCallTrace)
		if currentToolCalls := telemetry.GetToolCallsFromContext(run.ctx); currentToolCalls != nil {
			fmt.Printf("DEBUG AzureOpenAI: Context now has %d tool calls\n", len(currentToolCalls))
		} else {
			fmt.Printf("DEBUG AzureOpenAI: WARNING: Context has nil tool calls after adding\n")
		}
		azureOpenAIStoreStreamToolResult(run.ctx, run.params.Memory, toolCall.ID, toolCall.Function.Name, toolCall.Function.Arguments, result)
		toolMessage := openai.ToolMessage(result, toolCall.ID)
		c.logger.Debug(run.ctx, "Created tool message", map[string]interface{}{"message_type": "tool"})
		run.messages = append(run.messages, toolMessage)
	}
}

func (c *Client) recordMissingStreamTool(run *azureOpenAIStreamToolRun, toolCall openai.ChatCompletionMessageToolCallUnion) {
	c.logger.Error(run.ctx, "Tool not found", map[string]interface{}{"tool_name": toolCall.Function.Name})
	now := time.Now()
	errorMessage := fmt.Sprintf("Error: tool not found: %s", toolCall.Function.Name)
	telemetry.AddToolCallToContext(run.ctx, telemetry.ToolCall{
		Name: toolCall.Function.Name, Arguments: toolCall.Function.Arguments, ID: toolCall.ID,
		Timestamp: now.Format(time.RFC3339), StartTime: now, Duration: 0, DurationMs: 0,
		Error: fmt.Sprintf("tool not found: %s", toolCall.Function.Name), Result: errorMessage,
	})
	azureOpenAIStoreStreamToolResult(run.ctx, run.params.Memory, toolCall.ID, toolCall.Function.Name, toolCall.Function.Arguments, errorMessage)
}

func azureOpenAIStoreStreamToolResult(ctx context.Context, memory contracts.Memory, id, name, arguments, result string) {
	if memory == nil {
		return
	}
	_ = memory.AddMessage(ctx, contracts.Message{
		Role: "assistant", Content: "", ToolCalls: []contracts.ToolCall{{ID: id, Name: name, Arguments: arguments}},
	})
	_ = memory.AddMessage(ctx, contracts.Message{
		Role: "tool", Content: result, ToolCallID: id, Metadata: map[string]interface{}{"tool_name": name},
	})
}
