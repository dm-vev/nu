package openai

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"

	"nu/internal/contracts"

	"github.com/openai/openai-go/v2"
)

type openAIParallelToolResult struct {
	index  int
	result string
	err    error
}

func (c *Client) executeParallelToolCall(
	ctx context.Context,
	toolCall openai.ChatCompletionMessageToolCallUnion,
	tools []contracts.Tool,
	memory contracts.Memory,
	history map[string]int,
	historyMu *sync.Mutex,
) (openai.ChatCompletionMessageParamUnion, bool, error) {
	c.logger.Info(ctx, "Parallel tool call", map[string]interface{}{"toolName": toolCall.Function.Name})
	var wrapper struct {
		ToolUses []map[string]interface{} `json:"tool_uses"`
	}
	if err := json.Unmarshal([]byte(toolCall.Function.Arguments), &wrapper); err != nil {
		c.logger.Error(ctx, "Error unmarshalling tool uses", map[string]interface{}{"error": err.Error()})
		return openai.ChatCompletionMessageParamUnion{}, false, nil
	}

	resultCh := make(chan openAIParallelToolResult, len(wrapper.ToolUses))
	var wg sync.WaitGroup
	for i, toolUse := range wrapper.ToolUses {
		wg.Add(1)
		go func(index int, toolUse map[string]interface{}) {
			defer wg.Done()
			toolName := toolUse["recipient_name"].(string)
			parameters := toolUse["parameters"].(map[string]interface{})
			c.logger.Info(ctx, "Parallel tool use", map[string]interface{}{"toolName": toolName, "parameters": parameters})
			paramsBytes, err := json.Marshal(parameters)
			if err != nil {
				c.logger.Error(ctx, "Error marshalling parameters", map[string]interface{}{"error": err.Error()})
				resultCh <- openAIParallelToolResult{index: index, err: err}
				return
			}

			var selected contracts.Tool
			for _, candidate := range tools {
				if candidate.Name() == toolName {
					selected = candidate
					break
				}
			}
			if selected == nil {
				err := fmt.Errorf("tool not found: %s", toolName)
				c.logger.Error(ctx, "Tool not found in parallel execution", map[string]interface{}{"toolName": toolName})
				resultCh <- openAIParallelToolResult{index: index, err: err}
				return
			}

			c.logger.Info(ctx, "Executing tool", map[string]interface{}{"toolName": toolName, "parameters": string(paramsBytes)})
			result, err := selected.Execute(ctx, string(paramsBytes))
			cacheKey := toolName + ":" + string(paramsBytes)
			historyMu.Lock()
			history[cacheKey]++
			callCount := history[cacheKey]
			historyMu.Unlock()
			if callCount > 2 {
				warning := fmt.Sprintf("\n\n[WARNING: This is call #%d to %s with identical parameters. You may be in a loop. Consider using the available information to provide a final answer.]", callCount, toolName)
				if err == nil {
					result += warning
				}
				c.logger.Warn(ctx, "Repetitive tool call detected in parallel execution", map[string]interface{}{"toolName": toolName, "callCount": callCount})
			}

			if memory != nil {
				content := result
				if err != nil {
					content = fmt.Sprintf("Error: %v", err)
				}
				_ = memory.AddMessage(ctx, contracts.Message{
					Role: "assistant", Content: "",
					ToolCalls: []contracts.ToolCall{{ID: toolCall.ID, Name: toolName, Arguments: string(paramsBytes)}},
				})
				_ = memory.AddMessage(ctx, contracts.Message{
					Role: "tool", Content: content, ToolCallID: toolCall.ID,
					Metadata: map[string]interface{}{"tool_name": toolCall.Function.Name},
				})
			}
			resultCh <- openAIParallelToolResult{index: index, result: result, err: err}
		}(i, toolUse)
	}
	go func() {
		wg.Wait()
		close(resultCh)
	}()

	results := make([]string, len(wrapper.ToolUses))
	for result := range resultCh {
		if result.err != nil {
			c.logger.Error(ctx, "Error executing tool", map[string]interface{}{"error": result.err.Error()})
			return openai.ChatCompletionMessageParamUnion{}, false, fmt.Errorf("error executing tool: %s", result.err.Error())
		}
		results[result.index] = result.result
	}
	structuredResults := make([]string, 0, len(wrapper.ToolUses))
	for i, toolUse := range wrapper.ToolUses {
		toolName := toolUse["recipient_name"].(string)
		structuredResults = append(structuredResults, fmt.Sprintf("Tool: %s\nResult: %s", toolName, results[i]))
	}
	return openai.ToolMessage(strings.Join(structuredResults, "\n\n"), toolCall.ID), true, nil
}
