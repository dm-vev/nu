package anthropic

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/dm-vev/nu/contracts"
)

func (c *Client) generateFinalToolResponse(ctx context.Context, messages []Message, params *contracts.GenerateOptions, maxTokens, maxIterations int) (string, error) {
	c.logger.Info(ctx, "Maximum iterations reached, making final call without tools", map[string]interface{}{"maxIterations": maxIterations})
	finalReq := CompletionRequest{
		Model: c.Model, Messages: messages, MaxTokens: maxTokens,
		Temperature: params.LLMConfig.Temperature, TopP: params.LLMConfig.TopP, Tools: nil,
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
	messages = append(messages, Message{Role: "user", Content: finalUserMessage})
	if params.ResponseFormat != nil {
		messages = append(messages, Message{Role: "assistant", Content: "{"})
		c.logger.Debug(ctx, "Added prefill for structured output in final call", nil)
	}
	finalReq.Messages = messages
	c.logger.Debug(ctx, "Making final request without tools", map[string]interface{}{"messages": len(finalReq.Messages)})

	finalHTTPReq, err := c.createHTTPRequest(ctx, &finalReq, "/v1/messages")
	if err != nil {
		return "", fmt.Errorf("failed to create final request: %w", err)
	}
	finalHTTPResp, err := c.HTTPClient.Do(finalHTTPReq)
	if err != nil {
		c.logger.Error(ctx, "Error in final call without tools", map[string]interface{}{"error": err.Error()})
		return "", fmt.Errorf("failed to send final request: %w", err)
	}
	defer func() {
		if closeErr := finalHTTPResp.Body.Close(); closeErr != nil {
			c.logger.Warn(ctx, "Failed to close final response body", map[string]interface{}{"error": closeErr.Error()})
		}
	}()
	finalRespBody, err := io.ReadAll(finalHTTPResp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read final response body: %w", err)
	}
	if finalHTTPResp.StatusCode != http.StatusOK {
		c.logger.Error(ctx, "Error from Anthropic API in final call", map[string]interface{}{
			"status_code": finalHTTPResp.StatusCode, "response": string(finalRespBody),
		})
		return "", fmt.Errorf("error from Anthropic API in final call: %s", string(finalRespBody))
	}
	c.logger.Debug(ctx, "Raw final response before unmarshaling", map[string]interface{}{
		"response_length": len(finalRespBody),
		"response_prefix": func() string {
			if len(finalRespBody) > 100 {
				return string(finalRespBody[:100])
			}
			return string(finalRespBody)
		}(),
		"first_char": func() string {
			if len(finalRespBody) > 0 {
				return fmt.Sprintf("'%c' (0x%02x)", finalRespBody[0], finalRespBody[0])
			}
			return "empty"
		}(),
	})

	var finalResp CompletionResponse
	if err = json.Unmarshal(finalRespBody, &finalResp); err != nil {
		c.logger.Error(ctx, "Failed to unmarshal final response", map[string]interface{}{
			"error": err.Error(), "response_length": len(finalRespBody),
			"response_sample": func() string {
				if len(finalRespBody) > 200 {
					return string(finalRespBody[:200])
				}
				return string(finalRespBody)
			}(),
		})
		return "", fmt.Errorf("failed to unmarshal final response: %w", err)
	}
	if finalResp.Content == nil {
		return "", fmt.Errorf("no content in final response")
	}
	var finalTextContent []string
	for _, contentBlock := range finalResp.Content {
		if contentBlock.Type == "text" {
			finalTextContent = append(finalTextContent, contentBlock.Text)
		}
	}
	if len(finalTextContent) == 0 {
		return "", fmt.Errorf("no text content in final response")
	}
	response := strings.Join(finalTextContent, "\n")
	if params.ResponseFormat != nil {
		response = "{" + response
		extractedJSON := anthropicExtractJSONFromResponse(response)
		if extractedJSON != response {
			c.logger.Debug(ctx, "Extracted JSON from final response", map[string]interface{}{
				"original_length": len(response), "extracted_length": len(extractedJSON),
			})
			response = extractedJSON
		}
	}
	c.logger.Info(ctx, "Successfully received final response without tools", map[string]interface{}{
		"response_length": len(response), "response_preview": response,
	})
	return response, nil
}
