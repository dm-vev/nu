package azureopenai

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"

	"github.com/dm-vev/nu/contracts"

	"github.com/openai/openai-go/v2"
)

type azureOpenAIParallelToolResult struct {
	index  int
	result string
	err    error
}

func (c *Client) executeParallelToolCall(
	ctx context.Context,
	toolCall openai.ChatCompletionMessageToolCallUnion,
	state *azureOpenAIToolExecutionState,
) (*openai.ChatCompletionMessageParamUnion, error) {
	c.logger.Info(ctx, "Parallel tool call", map[string]interface{}{"toolName": toolCall.Function.Name})
	var toolUsesWrapper struct {
		ToolUses []map[string]interface{} `json:"tool_uses"`
	}
	if err := json.Unmarshal([]byte(toolCall.Function.Arguments), &toolUsesWrapper); err != nil {
		c.logger.Error(ctx, "Error unmarshalling tool uses", map[string]interface{}{"error": err.Error()})
		return nil, nil
	}

	resultCh := make(chan azureOpenAIParallelToolResult, len(toolUsesWrapper.ToolUses))
	var wg sync.WaitGroup
	for i, toolUse := range toolUsesWrapper.ToolUses {
		wg.Add(1)
		go func(index int, toolUse map[string]interface{}) {
			defer wg.Done()
			toolName := toolUse["recipient_name"].(string)
			parameters := toolUse["parameters"].(map[string]interface{})
			c.logger.Info(ctx, "Parallel tool use", map[string]interface{}{"toolName": toolName, "parameters": parameters})
			paramsBytes, err := json.Marshal(parameters)
			if err != nil {
				c.logger.Error(ctx, "Error marshalling parameters", map[string]interface{}{"error": err.Error()})
				resultCh <- azureOpenAIParallelToolResult{index: index, err: err}
				return
			}

			var tool contracts.Tool
			for _, candidate := range state.tools {
				if candidate.Name() == toolName {
					tool = candidate
					break
				}
			}
			if tool == nil {
				err := fmt.Errorf("tool not found: %s", toolName)
				c.logger.Error(ctx, "Tool not found in parallel execution", map[string]interface{}{"toolName": toolName})
				resultCh <- azureOpenAIParallelToolResult{index: index, err: err}
				return
			}

			c.logger.Info(ctx, "Executing tool", map[string]interface{}{"toolName": toolName, "parameters": string(paramsBytes)})
			result, err := tool.Execute(ctx, string(paramsBytes))
			cacheKey := toolName + ":" + string(paramsBytes)
			state.historyLock.Lock()
			state.history[cacheKey]++
			callCount := state.history[cacheKey]
			state.historyLock.Unlock()
			if callCount > 2 {
				warning := fmt.Sprintf("\n\n[WARNING: This is call #%d to %s with identical parameters. You may be in a loop. Consider using the available information to provide a final answer.]", callCount, toolName)
				if err == nil {
					result += warning
				}
				c.logger.Warn(ctx, "Repetitive tool call detected in parallel execution", map[string]interface{}{"toolName": toolName, "callCount": callCount})
			}
			azureOpenAIStoreToolResult(ctx, state.params.Memory, toolCall.ID, toolName, toolCall.Function.Name, string(paramsBytes), result, err)
			resultCh <- azureOpenAIParallelToolResult{index: index, result: result, err: err}
		}(i, toolUse)
	}
	go func() {
		wg.Wait()
		close(resultCh)
	}()

	toolResults := make([]string, len(toolUsesWrapper.ToolUses))
	for result := range resultCh {
		if result.err != nil {
			c.logger.Error(ctx, "Error executing tool", map[string]interface{}{"error": result.err.Error()})
			return nil, fmt.Errorf("error executing tool: %s", result.err.Error())
		}
		toolResults[result.index] = result.result
	}
	structuredResults := make([]string, 0, len(toolUsesWrapper.ToolUses))
	for i, toolUse := range toolUsesWrapper.ToolUses {
		toolName := toolUse["recipient_name"].(string)
		structuredResults = append(structuredResults, fmt.Sprintf("Tool: %s\nResult: %s", toolName, toolResults[i]))
	}
	message := openai.ToolMessage(strings.Join(structuredResults, "\n\n"), toolCall.ID)
	return &message, nil
}
