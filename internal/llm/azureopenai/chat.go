package azureopenai

import (
	"context"
	"fmt"
	"nu/internal/llm"

	"github.com/openai/openai-go/v2"
	"github.com/openai/openai-go/v2/shared"
)

// Chat uses the ChatCompletion API to have a conversation (messages) with a model
func (c *Client) Chat(ctx context.Context, messages []llm.Message, params *llm.GenerateParams) (string, error) {
	if params == nil {
		params = llm.DefaultGenerateParams()
	}

	chatMessages := make([]openai.ChatCompletionMessageParamUnion, len(messages))
	for i, msg := range messages {
		switch msg.Role {
		case "system":
			chatMessages[i] = openai.SystemMessage(msg.Content)
		case "user":
			chatMessages[i] = openai.UserMessage(msg.Content)
		case "assistant":
			chatMessages[i] = openai.AssistantMessage(msg.Content)
		case "tool":
			chatMessages[i] = openai.ToolMessage(msg.Content, msg.ToolCallID)
		default:
			chatMessages[i] = openai.UserMessage(msg.Content)
		}
	}

	req := openai.ChatCompletionNewParams{
		Model: openai.ChatModel(c.deployment), Messages: chatMessages,
		Temperature:      openai.Float(c.getTemperatureForModel(params.Temperature)),
		FrequencyPenalty: openai.Float(params.FrequencyPenalty), PresencePenalty: openai.Float(params.PresencePenalty),
	}
	if !azureOpenAIIsReasoningModel(c.Model) {
		req.TopP = openai.Float(params.TopP)
	}
	if len(params.StopSequences) > 0 {
		req.Stop = openai.ChatCompletionNewParamsStopUnion{OfStringArray: params.StopSequences}
	}
	if azureOpenAIIsReasoningModel(c.Model) && params.Reasoning != "" {
		req.ReasoningEffort = shared.ReasoningEffort(params.Reasoning)
		c.logger.Debug(ctx, "Setting reasoning effort", map[string]interface{}{"reasoning_effort": params.Reasoning})
	}

	var resp *openai.ChatCompletion
	var err error
	operation := func() error {
		c.logger.Debug(ctx, "Executing Azure OpenAI Chat API request", map[string]interface{}{
			"model": c.Model, "deployment": c.deployment, "temperature": req.Temperature,
			"top_p": req.TopP, "frequency_penalty": req.FrequencyPenalty,
			"presence_penalty": req.PresencePenalty, "stop_sequences": req.Stop,
			"messages": len(req.Messages), "reasoning_effort": params.Reasoning,
		})
		resp, err = c.ChatService.Completions.New(ctx, req)
		if err != nil {
			c.logger.Error(ctx, "Error from Azure OpenAI Chat API", map[string]interface{}{"error": err.Error(), "model": c.Model, "deployment": c.deployment})
			return fmt.Errorf("failed to create chat completion: %w", err)
		}
		return nil
	}
	if c.retryExecutor != nil {
		c.logger.Debug(ctx, "Using retry mechanism for Azure OpenAI Chat request", map[string]interface{}{"model": c.Model, "deployment": c.deployment})
		err = c.retryExecutor.Execute(ctx, operation)
	} else {
		err = operation()
	}
	if err != nil {
		return "", err
	}
	if len(resp.Choices) == 0 {
		return "", fmt.Errorf("no completions returned")
	}
	c.logger.Debug(ctx, "Successfully received chat response from Azure OpenAI", map[string]interface{}{"model": c.Model, "deployment": c.deployment})
	return resp.Choices[0].Message.Content, nil
}
