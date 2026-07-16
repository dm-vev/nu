package azureopenai

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"nu/internal/contracts"
	"nu/internal/multitenancy"

	"github.com/openai/openai-go/v2"
	"github.com/openai/openai-go/v2/shared"
)

type azureOpenAIToolExecutionState struct {
	tools       []contracts.Tool
	params      *contracts.GenerateOptions
	history     map[string]int
	historyLock sync.Mutex
}

// GenerateWithTools implements contracts.LLM.GenerateWithTools
func (c *Client) GenerateWithTools(ctx context.Context, prompt string, tools []contracts.Tool, options ...contracts.GenerateOption) (string, error) {
	params := &contracts.GenerateOptions{}
	for _, opt := range options {
		if opt != nil {
			opt(params)
		}
	}
	if params.LLMConfig == nil {
		params.LLMConfig = &contracts.LLMConfig{
			Temperature: 0.7, TopP: 1.0, FrequencyPenalty: 0.0, PresencePenalty: 0.0,
		}
	}
	maxIterations := params.MaxIterations
	if maxIterations == 0 {
		maxIterations = 2
	}

	orgID := "default"
	if id, err := multitenancy.GetOrgID(ctx); err == nil {
		orgID = id
	}
	ctx = context.WithValue(ctx, azureOpenAIOrganizationKey, orgID)

	openaiTools := make([]openai.ChatCompletionToolUnionParam, len(tools))
	for i, tool := range tools {
		openaiTools[i] = openai.ChatCompletionFunctionTool(shared.FunctionDefinitionParam{
			Name: tool.Name(), Description: openai.String(tool.Description()),
			Parameters: c.convertToOpenAISchema(tool.Parameters()),
		})
	}

	messages := []openai.ChatCompletionMessageParamUnion{}
	if params.SystemMessage != "" {
		messages = append(messages, openai.SystemMessage(params.SystemMessage))
		c.logger.Debug(ctx, "Using system message", map[string]interface{}{"system_message": params.SystemMessage})
	}
	builder := azureOpenAINewMessageHistoryBuilder(c.logger)
	messages = append(messages, builder.buildMessages(ctx, prompt, params.Memory)...)

	req := openai.ChatCompletionNewParams{
		Model: openai.ChatModel(c.deployment), Messages: messages, Tools: openaiTools,
		Temperature:      openai.Float(c.getTemperatureForModel(params.LLMConfig.Temperature)),
		FrequencyPenalty: openai.Float(params.LLMConfig.FrequencyPenalty),
		PresencePenalty:  openai.Float(params.LLMConfig.PresencePenalty),
	}
	if !azureOpenAIIsReasoningModel(c.Model) {
		req.TopP = openai.Float(params.LLMConfig.TopP)
		req.ParallelToolCalls = openai.Bool(true)
	}
	if len(params.LLMConfig.StopSequences) > 0 {
		req.Stop = openai.ChatCompletionNewParamsStopUnion{OfStringArray: params.LLMConfig.StopSequences}
	}
	if azureOpenAIIsReasoningModel(c.Model) && params.LLMConfig.Reasoning != "" {
		req.ReasoningEffort = shared.ReasoningEffort(params.LLMConfig.Reasoning)
		c.logger.Debug(ctx, "Setting reasoning effort", map[string]interface{}{"reasoning_effort": params.LLMConfig.Reasoning})
	}
	if params.ResponseFormat != nil {
		jsonSchema := shared.ResponseFormatJSONSchemaJSONSchemaParam{Name: params.ResponseFormat.Name, Schema: params.ResponseFormat.Schema}
		req.ResponseFormat = openai.ChatCompletionNewParamsResponseFormatUnion{OfJSONSchema: &shared.ResponseFormatJSONSchemaParam{Type: "json_schema", JSONSchema: jsonSchema}}
		c.logger.Debug(ctx, "Using response format", map[string]interface{}{"format": *params.ResponseFormat})
	}

	state := &azureOpenAIToolExecutionState{tools: tools, params: params, history: make(map[string]int)}
	var lastContent string
	for iteration := 0; iteration < maxIterations; iteration++ {
		req.Messages = messages
		reasoningEffort := "none"
		if params.LLMConfig != nil && params.LLMConfig.Reasoning != "" {
			reasoningEffort = params.LLMConfig.Reasoning
		}
		c.logger.Debug(ctx, "Sending request with tools to Azure OpenAI", map[string]interface{}{
			"model": c.Model, "deployment": c.deployment, "temperature": req.Temperature,
			"top_p": req.TopP, "frequency_penalty": req.FrequencyPenalty,
			"presence_penalty": req.PresencePenalty, "stop_sequences": req.Stop,
			"messages": len(req.Messages), "tools": len(req.Tools),
			"response_format": params.ResponseFormat != nil, "parallel_tools": req.ParallelToolCalls,
			"reasoning_effort": reasoningEffort, "iteration": iteration + 1, "maxIterations": maxIterations,
		})
		resp, err := c.ChatService.Completions.New(ctx, req)
		if err != nil {
			c.logger.Error(ctx, "Error from Azure OpenAI API", map[string]interface{}{"error": err.Error(), "deployment": c.deployment})
			return "", fmt.Errorf("failed to create chat completion: %w", err)
		}
		if len(resp.Choices) == 0 {
			return "", fmt.Errorf("no completions returned")
		}

		lastContent = strings.TrimSpace(resp.Choices[0].Message.Content)
		if len(resp.Choices[0].Message.ToolCalls) == 0 {
			return lastContent, nil
		}
		toolCalls := resp.Choices[0].Message.ToolCalls
		c.logger.Info(ctx, "Processing tool calls", map[string]interface{}{"count": len(toolCalls), "iteration": iteration + 1})
		messages = append(messages, resp.Choices[0].Message.ToParam())
		if err := c.processToolCalls(ctx, toolCalls, state, &messages, resp); err != nil {
			return "", err
		}
	}

	if params.DisableFinalSummary {
		c.logger.Info(ctx, "DisableFinalSummary enabled, skipping final summary call", map[string]interface{}{"maxIterations": maxIterations})
		return lastContent, nil
	}
	return c.generateFinalToolResponse(ctx, messages, params, maxIterations)
}

// GenerateWithToolsDetailed generates text with tools and returns detailed response information including token usage
func (c *Client) GenerateWithToolsDetailed(ctx context.Context, prompt string, tools []contracts.Tool, options ...contracts.GenerateOption) (*contracts.LLMResponse, error) {
	content, err := c.GenerateWithTools(ctx, prompt, tools, options...)
	if err != nil {
		return nil, err
	}
	return &contracts.LLMResponse{
		Content: content, Model: c.Model, StopReason: "", Usage: nil,
		Metadata: map[string]interface{}{"provider": "azure_openai", "deployment": c.deployment, "tools_used": true},
	}, nil
}
