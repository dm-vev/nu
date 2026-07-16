package openai

import (
	"context"
	"time"

	"nu/internal/contracts"
	"nu/internal/multitenancy"

	"github.com/openai/openai-go/v2"
	"github.com/openai/openai-go/v2/shared"
)

// GenerateWithToolsStream streams iterative tool calling.
func (c *Client) GenerateWithToolsStream(ctx context.Context, prompt string, tools []contracts.Tool, options ...contracts.GenerateOption) (<-chan contracts.StreamEvent, error) {
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
		c.runToolsStream(ctx, prompt, tools, params, maxIterations, eventChan)
	}()
	return eventChan, nil
}

func (c *Client) runToolsStream(ctx context.Context, prompt string, tools []contracts.Tool, params *contracts.GenerateOptions, maxIterations int, eventChan chan<- contracts.StreamEvent) {
	openaiTools := make([]openai.ChatCompletionToolUnionParam, len(tools))
	for i, tool := range tools {
		openaiTools[i] = openai.ChatCompletionFunctionTool(shared.FunctionDefinitionParam{
			Name: tool.Name(), Description: openai.String(tool.Description()), Parameters: c.convertToOpenAISchema(tool.Parameters()),
		})
	}
	messages := []openai.ChatCompletionMessageParamUnion{}
	if params.SystemMessage != "" {
		messages = append(messages, openai.SystemMessage(params.SystemMessage))
		c.logger.Debug(ctx, "Using system message for tools", map[string]interface{}{"system_message": params.SystemMessage})
	}
	builder := openAINewMessageHistoryBuilder(c.logger)
	messages = append(messages, builder.buildMessages(ctx, prompt, params.Memory)...)
	eventChan <- contracts.StreamEvent{
		Type: contracts.StreamEventMessageStart, Timestamp: time.Now(),
		Metadata: map[string]interface{}{"model": c.Model, "tools": len(openaiTools)},
	}

	filterIntermediateContent := params.StreamConfig == nil || !params.StreamConfig.IncludeIntermediateMessages
	var capturedContentEvents []contracts.StreamEvent
	gotCompleteResponse := false
	for iteration := 0; iteration < maxIterations; iteration++ {
		result := c.runToolStreamIteration(ctx, messages, openaiTools, tools, params, iteration, maxIterations, filterIntermediateContent, eventChan)
		if result.abort {
			return
		}
		messages = result.messages
		if result.complete {
			gotCompleteResponse = true
			break
		}
		if filterIntermediateContent && result.hadContent && result.calledTools {
			capturedContentEvents = append(capturedContentEvents, result.contentEvents...)
		}
	}

	if filterIntermediateContent && len(capturedContentEvents) > 0 {
		c.logger.Debug(ctx, "Replaying captured content events from tool iterations", map[string]interface{}{"eventsCount": len(capturedContentEvents)})
		for _, contentEvent := range capturedContentEvents {
			eventChan <- contentEvent
		}
	}
	if gotCompleteResponse {
		c.logger.Debug(ctx, "Skipping final synthesis call - already got complete response", map[string]interface{}{"maxIterations": maxIterations})
		eventChan <- contracts.StreamEvent{Type: contracts.StreamEventMessageStop, Timestamp: time.Now()}
		return
	}
	if params.DisableFinalSummary {
		c.logger.Info(ctx, "DisableFinalSummary enabled, skipping final synthesis call", map[string]interface{}{"maxIterations": maxIterations})
		eventChan <- contracts.StreamEvent{Type: contracts.StreamEventMessageStop, Timestamp: time.Now()}
		return
	}
	c.streamFinalToolResponse(ctx, messages, params, maxIterations, eventChan)
}
