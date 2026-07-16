package gemini

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"google.golang.org/genai"

	"nu/internal/contracts"
	"nu/internal/multitenancy"
)

// GenerateStream generates text with streaming response using native Gemini streaming
func (c *Client) GenerateStream(ctx context.Context, prompt string, options ...contracts.GenerateOption) (<-chan contracts.StreamEvent, error) {
	// Convert options to params
	params := &contracts.GenerateOptions{}
	for _, opt := range options {
		if opt != nil {
			opt(params)
		}
	}

	// Get streaming config or use default
	streamConfig := contracts.DefaultStreamConfig()
	if params.StreamConfig != nil {
		streamConfig = *params.StreamConfig
	}

	// Check for organization ID in context
	orgID := "default"
	if id, err := multitenancy.GetOrgID(ctx); err == nil {
		orgID = id
	}
	_ = orgID

	// Build contents using unified builder
	builder := geminiNewMessageHistoryBuilder(c.logger)
	contents := builder.buildContents(ctx, prompt, params)

	// Add system instruction if provided or if reasoning is specified
	var systemInstruction *genai.Content
	systemMessage := params.SystemMessage

	// Log reasoning mode usage - only affects native thinking models (2.5 series)
	if params.LLMConfig != nil && params.LLMConfig.Reasoning != "" {
		if SupportsThinking(c.model) {
			c.logger.Debug(ctx, "Using reasoning mode with thinking-capable model", map[string]interface{}{
				"reasoning": params.LLMConfig.Reasoning,
				"model":     c.model,
			})
		} else {
			c.logger.Debug(ctx, "Reasoning mode specified for non-thinking model - native thinking tokens not available", map[string]interface{}{
				"reasoning":        params.LLMConfig.Reasoning,
				"model":            c.model,
				"supportsThinking": false,
			})
		}
	}

	if systemMessage != "" {
		systemInstruction = &genai.Content{
			Parts: []*genai.Part{
				{Text: systemMessage},
			},
		}
		c.logger.Debug(ctx, "Using system message", map[string]interface{}{"system_message": systemMessage})
	}

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

	// Set response format if provided
	if params.ResponseFormat != nil {
		if genConfig == nil {
			genConfig = &genai.GenerationConfig{}
		}
		genConfig.ResponseMIMEType = "application/json"

		// Convert schema for genai
		if schemaBytes, err := json.Marshal(params.ResponseFormat.Schema); err == nil {
			var schema *genai.Schema
			if err := json.Unmarshal(schemaBytes, &schema); err != nil {
				c.logger.Warn(ctx, "Failed to convert response schema", map[string]interface{}{"error": err.Error()})
			} else {
				genConfig.ResponseSchema = schema
			}
		}
		c.logger.Debug(ctx, "Using response format", map[string]interface{}{"format": *params.ResponseFormat})
	}

	// Create config
	config := &genai.GenerateContentConfig{
		SystemInstruction: systemInstruction,
	}

	// Apply generation config parameters directly to config
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

	// Add thinking configuration if supported and enabled
	if SupportsThinking(c.model) && c.thinkingConfig != nil {
		if c.thinkingConfig.IncludeThoughts || c.thinkingConfig.ThinkingBudget != nil {
			config.ThinkingConfig = &genai.ThinkingConfig{
				IncludeThoughts: c.thinkingConfig.IncludeThoughts,
				ThinkingBudget:  c.thinkingConfig.ThinkingBudget,
			}

			c.logger.Debug(ctx, "Enabled thinking configuration for streaming", map[string]interface{}{
				"includeThoughts": c.thinkingConfig.IncludeThoughts,
				"thinkingBudget":  c.thinkingConfig.ThinkingBudget,
			})
		}
	}

	// Create event channel
	eventCh := make(chan contracts.StreamEvent, streamConfig.BufferSize)

	// Start streaming goroutine
	go func() {
		defer close(eventCh)

		// Send message start event
		select {
		case eventCh <- contracts.StreamEvent{
			Type:      contracts.StreamEventMessageStart,
			Timestamp: time.Now(),
		}:
		case <-ctx.Done():
			return
		}

		c.logger.Debug(ctx, "Starting native Gemini streaming", map[string]interface{}{
			"model":           c.model,
			"thinkingEnabled": SupportsThinking(c.model) && c.thinkingConfig != nil && c.thinkingConfig.IncludeThoughts,
		})

		// Track accumulated content for memory storage
		var accumulatedContent strings.Builder

		// Start streaming
		streamIter := c.genaiClient.Models.GenerateContentStream(ctx, c.model, contents, config)

		for response, err := range streamIter {
			if err != nil {
				// Send error event
				select {
				case eventCh <- contracts.StreamEvent{
					Type:      contracts.StreamEventError,
					Error:     err,
					Timestamp: time.Now(),
				}:
				case <-ctx.Done():
				}
				return
			}

			// Process each candidate in the response
			for _, candidate := range response.Candidates {
				if candidate.Content == nil {
					continue
				}

				// Process each part in the content
				for _, part := range candidate.Content.Parts {
					if part.Text == "" {
						continue
					}

					// Check if this is thinking content
					if part.Thought {
						// Send thinking event
						select {
						case eventCh <- contracts.StreamEvent{
							Type:      contracts.StreamEventThinking,
							Content:   part.Text,
							Timestamp: time.Now(),
							Metadata: map[string]interface{}{
								"thought_signature": part.ThoughtSignature,
							},
						}:
						case <-ctx.Done():
							return
						}
					} else {
						// Send content delta event and accumulate for memory
						accumulatedContent.WriteString(part.Text)
						select {
						case eventCh <- contracts.StreamEvent{
							Type:      contracts.StreamEventContentDelta,
							Content:   part.Text,
							Timestamp: time.Now(),
						}:
						case <-ctx.Done():
							return
						}
					}
				}
			}
		}

		// Send content complete event
		select {
		case eventCh <- contracts.StreamEvent{
			Type:      contracts.StreamEventContentComplete,
			Timestamp: time.Now(),
		}:
		case <-ctx.Done():
			return
		}

		// Send message stop event
		select {
		case eventCh <- contracts.StreamEvent{
			Type:      contracts.StreamEventMessageStop,
			Timestamp: time.Now(),
		}:
		case <-ctx.Done():
			return
		}

		c.logger.Debug(ctx, "Successfully completed native Gemini streaming response", map[string]interface{}{
			"model": c.model,
		})
	}()

	return eventCh, nil
}
