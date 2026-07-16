package anthropic

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/dm-vev/nu/contracts"
)

func (c *Client) executeFinalToolStream(ctx context.Context, messages []Message, params *contracts.GenerateOptions, eventChan chan<- contracts.StreamEvent, maxTokens, maxIterations, finalIterationCount int) error {
	c.logger.Info(ctx, "[LLM RESPONSE DEBUG] Making final synthesis call after tool iterations", map[string]interface{}{
		"maxIterations": maxIterations, "totalPreviousLLMCalls": finalIterationCount, "reason": "no_complete_response_received",
	})

	finalUserMessage := "Please provide your final response based on the information available. Do not request any additional tools."
	if params.ResponseFormat != nil {
		schemaJSON, err := json.MarshalIndent(params.ResponseFormat.Schema, "", "  ")
		if err == nil {
			exampleJSON := anthropicCreateExampleFromSchema(params.ResponseFormat.Schema)
			exampleStr, _ := json.MarshalIndent(exampleJSON, "", "  ")
			finalUserMessage = fmt.Sprintf(`%s

You must respond with a valid JSON object that exactly follows this schema:
%s

Here is an example of the expected JSON structure:
%s

CRITICAL INSTRUCTIONS:
- Output ONLY valid JSON, no additional text before or after
- Follow the EXACT structure shown in the schema and example
- Use the field names exactly as specified
- Ensure all required fields are present
- Pay special attention to array fields - they must be arrays of objects, not simple objects
- If a field is defined as an array in the schema, it MUST be an array in your response
- The JSON must be directly parsable and match the schema precisely`, finalUserMessage, string(schemaJSON), string(exampleStr))
		}
	}

	finalMessages := append(messages, Message{Role: "user", Content: finalUserMessage})
	finalReq := CompletionRequest{
		Model: c.Model, Messages: finalMessages, MaxTokens: maxTokens,
		Temperature: params.LLMConfig.Temperature, TopP: params.LLMConfig.TopP, Stream: true,
	}
	if params.SystemMessage != "" {
		if params.ResponseFormat != nil {
			finalReq.System = params.SystemMessage + "\n\nIMPORTANT: You must respond with valid JSON that matches the specified schema. Return ONLY the raw JSON object without any markdown formatting, code blocks, or wrapper text. Pay special attention to array fields - if a field is defined as an array in the schema, it MUST be an array in your response, not an object."
		} else {
			finalReq.System = params.SystemMessage
		}
	} else if params.ResponseFormat != nil {
		finalReq.System = "You must respond with valid JSON that matches the specified schema. Return ONLY the raw JSON object without any markdown formatting, code blocks, or wrapper text. Pay special attention to array fields - if a field is defined as an array in the schema, it MUST be an array in your response, not an object."
	}

	if params.LLMConfig != nil && params.LLMConfig.EnableReasoning && SupportsThinking(c.Model) {
		finalReq.Thinking = &ReasoningSpec{Type: "enabled"}
		if params.LLMConfig.ReasoningBudget > 0 {
			finalReq.Thinking.BudgetTokens = params.LLMConfig.ReasoningBudget
		}
		finalReq.Temperature = 1.0
		c.logger.Debug(ctx, "Getting final answer with reasoning after tools", map[string]interface{}{
			"model": c.Model, "budget_tokens": params.LLMConfig.ReasoningBudget,
			"max_tokens": maxTokens, "temperature": finalReq.Temperature,
		})
	}

	c.logger.Debug(ctx, "[LLM RESPONSE DEBUG] Executing final synthesis LLM call", map[string]interface{}{
		"finalCallNumber": finalIterationCount + 1, "messageCount": len(finalMessages),
	})
	err := c.executeStreamingRequestWithMemory(ctx, finalReq, eventChan, "", params)
	if err != nil {
		c.logger.Error(ctx, "[LLM RESPONSE DEBUG] Final synthesis call failed", map[string]interface{}{"error": err.Error()})
	} else {
		c.logger.Info(ctx, "[LLM RESPONSE DEBUG] Final synthesis call completed successfully", map[string]interface{}{"totalLLMCalls": finalIterationCount + 1})
	}
	return err
}
