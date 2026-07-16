package openai

import (
	"context"
	"fmt"
	"github.com/dm-vev/nu/internal/llm"

	"github.com/openai/openai-go/v2"
	"github.com/openai/openai-go/v2/shared"
)

// Chat uses the ChatCompletion API to have a conversation with a model.
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
		Model: openai.ChatModel(c.Model), Messages: chatMessages,
		Temperature: openai.Float(c.getTemperatureForModel(params.Temperature)),
	}
	if params.FrequencyPenalty != 0 {
		req.FrequencyPenalty = openai.Float(params.FrequencyPenalty)
	}
	if params.PresencePenalty != 0 {
		req.PresencePenalty = openai.Float(params.PresencePenalty)
	}
	if !openAIIsReasoningModel(c.Model) {
		req.TopP = openai.Float(params.TopP)
	}
	if len(params.StopSequences) > 0 {
		req.Stop = openai.ChatCompletionNewParamsStopUnion{OfStringArray: params.StopSequences}
	}
	if openAIIsReasoningModel(c.Model) && params.Reasoning != "" {
		req.ReasoningEffort = shared.ReasoningEffort(params.Reasoning)
		c.logger.Debug(ctx, "Setting reasoning effort", map[string]interface{}{"reasoning_effort": params.Reasoning})
	}

	var resp *openai.ChatCompletion
	var err error
	operation := func() error {
		c.logger.Debug(ctx, "Executing OpenAI Chat API request", map[string]interface{}{
			"model": c.Model, "temperature": req.Temperature, "top_p": req.TopP,
			"frequency_penalty": req.FrequencyPenalty, "presence_penalty": req.PresencePenalty,
			"stop_sequences": req.Stop, "messages": len(req.Messages), "reasoning_effort": params.Reasoning,
		})
		resp, err = c.ChatService.Completions.New(ctx, req)
		if err != nil {
			c.logger.Error(ctx, "Error from OpenAI Chat API", map[string]interface{}{"error": err.Error(), "model": c.Model})
			return fmt.Errorf("failed to create chat completion: %w", err)
		}
		return nil
	}
	if c.retryExecutor != nil {
		c.logger.Debug(ctx, "Using retry mechanism for OpenAI Chat request", map[string]interface{}{"model": c.Model})
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
	c.logger.Debug(ctx, "Successfully received chat response from OpenAI", map[string]interface{}{"model": c.Model})
	return resp.Choices[0].Message.Content, nil
}
