package azureopenai

import (
	"context"
	"time"

	"github.com/dm-vev/nu/contracts"
	"github.com/dm-vev/nu/internal/multitenancy"

	"github.com/openai/openai-go/v2"
	"github.com/openai/openai-go/v2/shared"
)

type azureOpenAIStreamToolRun struct {
	ctx           context.Context
	tools         []contracts.Tool
	params        *contracts.GenerateOptions
	maxIterations int
	events        chan contracts.StreamEvent
	openaiTools   []openai.ChatCompletionToolUnionParam
	messages      []openai.ChatCompletionMessageParamUnion
	filterContent bool
}

// GenerateWithToolsStream implements contracts.StreamingLLM.GenerateWithToolsStream with iterative tool calling
func (c *Client) GenerateWithToolsStream(
	ctx context.Context,
	prompt string,
	tools []contracts.Tool,
	options ...contracts.GenerateOption,
) (<-chan contracts.StreamEvent, error) {
	params := &contracts.GenerateOptions{LLMConfig: &contracts.LLMConfig{Temperature: 0.7}}
	for _, option := range options {
		option(params)
	}
	maxIterations := params.MaxIterations
	if maxIterations == 0 {
		maxIterations = 2
	}
	if id, err := multitenancy.GetOrgID(ctx); err == nil {
		ctx = multitenancy.WithOrgID(ctx, id)
	} else {
		ctx = multitenancy.WithOrgID(ctx, "default")
	}
	bufferSize := 100
	if params.StreamConfig != nil {
		bufferSize = params.StreamConfig.BufferSize
	}
	eventChan := make(chan contracts.StreamEvent, bufferSize)

	go func() {
		defer close(eventChan)
		openaiTools := make([]openai.ChatCompletionToolUnionParam, len(tools))
		for i, tool := range tools {
			openaiTools[i] = openai.ChatCompletionFunctionTool(shared.FunctionDefinitionParam{
				Name: tool.Name(), Description: openai.String(tool.Description()), Parameters: c.convertToOpenAISchema(tool.Parameters()),
			})
		}
		messages := []openai.ChatCompletionMessageParamUnion{}
		if params.SystemMessage != "" {
			messages = append(messages, openai.SystemMessage(params.SystemMessage))
			c.logger.Debug(ctx, "Using system message", map[string]interface{}{"system_message": params.SystemMessage})
		}
		builder := azureOpenAINewMessageHistoryBuilder(c.logger)
		messages = append(messages, builder.buildMessages(ctx, prompt, params.Memory)...)
		eventChan <- contracts.StreamEvent{
			Type: contracts.StreamEventMessageStart, Timestamp: time.Now(),
			Metadata: map[string]interface{}{"model": c.Model, "deployment": c.deployment, "tools": len(openaiTools)},
		}
		run := &azureOpenAIStreamToolRun{
			ctx: ctx, tools: tools, params: params, maxIterations: maxIterations,
			events: eventChan, openaiTools: openaiTools, messages: messages,
			filterContent: params.StreamConfig == nil || !params.StreamConfig.IncludeIntermediateMessages,
		}
		c.runToolStream(run)
	}()
	return eventChan, nil
}

func (c *Client) runToolStream(run *azureOpenAIStreamToolRun) {
	var capturedContentEvents []contracts.StreamEvent
	gotCompleteResponse := false
	for iteration := 0; iteration < run.maxIterations; iteration++ {
		result := c.runToolStreamIteration(run, iteration)
		if result.failed {
			return
		}
		if len(result.assistant.ToolCalls) == 0 {
			if result.hasContent {
				run.events <- contracts.StreamEvent{
					Type: contracts.StreamEventContentComplete, Timestamp: time.Now(),
					Metadata: map[string]interface{}{"iteration": iteration + 1},
				}
			}
			gotCompleteResponse = true
			break
		}

		c.logger.Info(run.ctx, "Processing tool calls", map[string]interface{}{"count": len(result.assistant.ToolCalls), "iteration": iteration + 1})
		for i, toolCall := range result.assistant.ToolCalls {
			c.logger.Debug(run.ctx, "Assistant tool call", map[string]interface{}{
				"index": i, "id": toolCall.ID, "id_length": len(toolCall.ID),
				"name": toolCall.Function.Name, "args_len": len(toolCall.Function.Arguments),
			})
		}
		result.assistant.Role = "assistant"
		run.messages = append(run.messages, result.assistant.ToParam())
		c.executeStreamToolCalls(run, result.assistant.ToolCalls, iteration)
		if run.filterContent && result.iterationHasContent {
			capturedContentEvents = append(capturedContentEvents, result.contentEvents...)
		}
	}

	if run.filterContent && len(capturedContentEvents) > 0 {
		c.logger.Debug(run.ctx, "Replaying captured content events from tool iterations", map[string]interface{}{"eventsCount": len(capturedContentEvents)})
		for _, contentEvent := range capturedContentEvents {
			run.events <- contentEvent
		}
	}
	if gotCompleteResponse {
		c.logger.Debug(run.ctx, "Skipping final synthesis call - already got complete response", map[string]interface{}{"maxIterations": run.maxIterations})
		run.events <- contracts.StreamEvent{Type: contracts.StreamEventMessageStop, Timestamp: time.Now()}
		return
	}
	if run.params.DisableFinalSummary {
		c.logger.Info(run.ctx, "DisableFinalSummary enabled, skipping final synthesis call", map[string]interface{}{"maxIterations": run.maxIterations})
		run.events <- contracts.StreamEvent{Type: contracts.StreamEventMessageStop, Timestamp: time.Now()}
		return
	}
	c.runFinalToolStream(run)
}
