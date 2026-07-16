package gemini

import (
	"context"
	"encoding/json"
	"fmt"

	"google.golang.org/genai"

	"nu/internal/contracts"
)

// generateWithToolsAndStream executes tool calling with real-time streaming events
func (c *Client) generateWithToolsAndStream(ctx context.Context, prompt string, tools []contracts.Tool, params *contracts.GenerateOptions, maxIterations int, eventCh chan contracts.StreamEvent) (string, error) {
	// Determine if we should filter intermediate content (for backward compatibility)
	filterIntermediateContent := params.StreamConfig == nil || !params.StreamConfig.IncludeIntermediateMessages

	// Track captured content for final iteration replay if filtering is enabled
	var capturedContentEvents []contracts.StreamEvent
	// Track whether any tool was actually executed across iterations. Used to
	// decide whether an iteration that returns no tool calls and no content
	// should fall through to the final-synthesis call (forcing the model to
	// answer from tool results) instead of returning an empty response.
	var executedAnyTool bool
	// Build tool map for quick lookup
	toolMap := make(map[string]contracts.Tool)
	for _, tool := range tools {
		toolMap[tool.Name()] = tool
	}

	// Track tool calls for clean loop continuation

	// Build contents using unified builder
	builder := geminiNewMessageHistoryBuilder(c.logger)
	contents := builder.buildContents(ctx, prompt, params)

	// Add system instruction if provided
	var systemInstruction *genai.Content
	if params.SystemMessage != "" {
		systemInstruction = &genai.Content{
			Parts: []*genai.Part{
				{Text: params.SystemMessage},
			},
		}
		c.logger.Debug(ctx, "Using system message for tool streaming", map[string]interface{}{"system_message": params.SystemMessage})
	}

	// Convert tools to Gemini format - all function declarations in a single tool.
	// Shared with the non-streaming path so array `items` and other schema
	// details stay consistent between Ask and Stream.
	geminiTools := []*genai.Tool{
		{
			FunctionDeclarations: geminiConvertToolsToFunctionDeclarations(tools),
		},
	}

	// Main conversation loop with streaming events
	for iteration := 0; iteration < maxIterations; iteration++ {
		// Set generation config
		var genConfig *genai.GenerationConfig
		if params.LLMConfig != nil {
			genConfig = &genai.GenerationConfig{}

			if params.LLMConfig.Temperature > 0 {
				temp := float32(params.LLMConfig.Temperature)
				genConfig.Temperature = &temp
			}
			if params.LLMConfig.TopP > 0 {
				topP := float32(params.LLMConfig.TopP)
				genConfig.TopP = &topP
			}
			if len(params.LLMConfig.StopSequences) > 0 {
				genConfig.StopSequences = params.LLMConfig.StopSequences
			}
		}

		// Apply max output tokens if configured at client level
		c.applyMaxOutputTokens(&genConfig)

		// Create config
		config := &genai.GenerateContentConfig{
			SystemInstruction: systemInstruction,
			Tools:             geminiTools,
		}

		// Apply generation config parameters
		if genConfig != nil {
			if genConfig.Temperature != nil {
				config.Temperature = genConfig.Temperature
			}
			if genConfig.TopP != nil {
				config.TopP = genConfig.TopP
			}
			if len(genConfig.StopSequences) > 0 {
				config.StopSequences = genConfig.StopSequences
			}
		}

		c.logger.Debug(ctx, "Sending request with tools for streaming", map[string]interface{}{
			"contents":      len(contents),
			"iteration":     iteration + 1,
			"maxIterations": maxIterations,
			"model":         c.model,
			"tools":         len(tools),
		})

		// Execute streaming request and collect tool calls
		shouldFilter := filterIntermediateContent && len(tools) > 0 && iteration < maxIterations-1
		var iterationContentEvents []contracts.StreamEvent
		toolCalls, hasContent, err := c.executeStreamingRequestWithToolCapture(ctx, contents, config, eventCh, shouldFilter, &iterationContentEvents)
		if err != nil {
			return "", err
		}

		// If we had content during this iteration and tools were called, capture it for final replay
		if shouldFilter && hasContent && len(toolCalls) > 0 {
			capturedContentEvents = append(capturedContentEvents, iterationContentEvents...)
		}

		// If no tool calls, we're done with the iteration loop
		if len(toolCalls) == 0 {
			// The model produced no content AND no tool call, but tools ran in
			// an earlier iteration: the model ended the turn without
			// synthesizing the tool results into an answer (observed with
			// Gemini on large tool outputs, where it emits a function call and
			// then an empty candidate). Returning here yields an empty
			// response even though the conversation holds tool results. Break
			// instead so the final-call-without-tools below forces a textual
			// answer. The maxIterations-exhaustion path already relies on that
			// same synthesis call.
			if !hasContent && executedAnyTool {
				break
			}
			// No tool calls means we have received the final response content
			// If content was filtered (captured), we need to replay it now
			if shouldFilter && len(iterationContentEvents) > 0 {
				for _, event := range iterationContentEvents {
					select {
					case eventCh <- event:
					case <-ctx.Done():
						return "", ctx.Err()
					}
				}
			}
			return "", nil
		}

		executedAnyTool = true
		contents, err = c.executeStreamToolCalls(ctx, contents, toolCalls, tools, iteration, eventCh)
		if err != nil {
			return "", err
		}
	}

	// Replay captured content events if we filtered them during iterations
	if filterIntermediateContent && len(capturedContentEvents) > 0 {
		c.logger.Debug(ctx, "Replaying captured content events from tool iterations", map[string]interface{}{
			"eventsCount": len(capturedContentEvents),
		})
		for _, contentEvent := range capturedContentEvents {
			select {
			case eventCh <- contentEvent:
			case <-ctx.Done():
				return "", ctx.Err()
			}
		}
	}

	// After all tool iterations, make a final call without tools to get the synthesized answer
	// This ensures the LLM provides a final response after processing all tool results

	// If DisableFinalSummary is enabled, skip the final synthesis call
	if params.DisableFinalSummary {
		c.logger.Info(ctx, "DisableFinalSummary enabled, skipping final synthesis call", map[string]interface{}{
			"maxIterations": maxIterations,
		})
		return "", nil
	}

	c.logger.Info(ctx, "Tool loop ended, making final call without tools to synthesize answer", map[string]interface{}{
		"maxIterations": maxIterations,
	})

	// Add a message to inform the LLM this is the final call
	contents = append(contents, &genai.Content{
		Role: "user",
		Parts: []*genai.Part{
			{Text: "Please provide your final response based on the information available. Do not request any additional functions."},
		},
	})

	// Set generation config without tools
	var genConfig *genai.GenerationConfig
	if params.LLMConfig != nil {
		genConfig = &genai.GenerationConfig{}

		if params.LLMConfig.Temperature > 0 {
			temp := float32(params.LLMConfig.Temperature)
			genConfig.Temperature = &temp
		}
		if params.LLMConfig.TopP > 0 {
			topP := float32(params.LLMConfig.TopP)
			genConfig.TopP = &topP
		}
		if len(params.LLMConfig.StopSequences) > 0 {
			genConfig.StopSequences = params.LLMConfig.StopSequences
		}
	}

	// Apply max output tokens if configured at client level
	c.applyMaxOutputTokens(&genConfig)

	// Add ResponseFormat if specified
	if params.ResponseFormat != nil {
		if genConfig == nil {
			genConfig = &genai.GenerationConfig{}
		}
		genConfig.ResponseMIMEType = "application/json"

		// Convert schema for genai
		if schemaBytes, err := json.Marshal(params.ResponseFormat.Schema); err == nil {
			var schema *genai.Schema
			if err := json.Unmarshal(schemaBytes, &schema); err != nil {
				c.logger.Warn(ctx, "Failed to convert response schema for final call", map[string]interface{}{"error": err.Error()})
			} else {
				genConfig.ResponseSchema = schema
			}
		}
	}

	config := &genai.GenerateContentConfig{
		SystemInstruction: systemInstruction,
		// No tools in final request - we want a final answer
	}

	// Apply generation config parameters
	if genConfig != nil {
		if genConfig.Temperature != nil {
			config.Temperature = genConfig.Temperature
		}
		if genConfig.TopP != nil {
			config.TopP = genConfig.TopP
		}
		if len(genConfig.StopSequences) > 0 {
			config.StopSequences = genConfig.StopSequences
		}
		if genConfig.ResponseMIMEType != "" {
			config.ResponseMIMEType = genConfig.ResponseMIMEType
		}
		if genConfig.ResponseSchema != nil {
			config.ResponseSchema = genConfig.ResponseSchema
		}
	}

	// Execute final request to get synthesized answer using streaming (no filtering for final call)
	_, _, err := c.executeStreamingRequestWithToolCapture(ctx, contents, config, eventCh, false, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create final content: %w", err)
	}

	return "", nil
}
