package gemini

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"google.golang.org/genai"

	"github.com/dm-vev/nu/contracts"
)

// executeStreamingRequestWithToolCapture executes a streaming request and captures tool calls
func (c *Client) executeStreamingRequestWithToolCapture(
	ctx context.Context,
	contents []*genai.Content,
	config *genai.GenerateContentConfig,
	eventCh chan<- contracts.StreamEvent,
	filterContent bool,
	capturedEvents *[]contracts.StreamEvent,
) ([]contracts.ToolCall, bool, error) {
	var toolCalls []contracts.ToolCall
	var hasContent bool

	c.logger.Debug(ctx, "Executing Gemini streaming request with tool capture", map[string]interface{}{
		"model":         c.model,
		"filterContent": filterContent,
	})

	// Add thinking configuration if supported and enabled
	if SupportsThinking(c.model) && c.thinkingConfig != nil {
		if c.thinkingConfig.IncludeThoughts || c.thinkingConfig.ThinkingBudget != nil {
			config.ThinkingConfig = &genai.ThinkingConfig{
				IncludeThoughts: c.thinkingConfig.IncludeThoughts,
				ThinkingBudget:  c.thinkingConfig.ThinkingBudget,
			}

			c.logger.Debug(ctx, "Enabled thinking configuration for tool streaming", map[string]interface{}{
				"includeThoughts": c.thinkingConfig.IncludeThoughts,
				"thinkingBudget":  c.thinkingConfig.ThinkingBudget,
			})
		}
	}

	// Generate content with tools using streaming
	streamIter := c.genaiClient.Models.GenerateContentStream(ctx, c.model, contents, config)

	for response, err := range streamIter {
		if err != nil {
			return nil, false, fmt.Errorf("failed to generate content stream: %w", err)
		}

		// Process each candidate in the response
		for _, candidate := range response.Candidates {
			if candidate.Content == nil {
				continue
			}

			// Process each part in the content
			for _, part := range candidate.Content.Parts {
				if part.FunctionCall != nil {
					// This is a tool call - capture it
					argsBytes, _ := json.Marshal(part.FunctionCall.Args)
					toolCall := contracts.ToolCall{
						ID:               fmt.Sprintf("geminiTool_%s", part.FunctionCall.Name),
						Name:             part.FunctionCall.Name,
						Arguments:        string(argsBytes),
						ThoughtSignature: part.ThoughtSignature,
					}
					toolCalls = append(toolCalls, toolCall)

					// Send tool use event to stream
					select {
					case eventCh <- contracts.StreamEvent{
						Type:      contracts.StreamEventToolUse,
						Timestamp: time.Now(),
						ToolCall:  &toolCall,
					}:
					case <-ctx.Done():
						return nil, false, ctx.Err()
					}
				} else if part.Text != "" {
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
							return nil, false, ctx.Err()
						}
					} else {
						// This is content
						hasContent = true
						contentEvent := contracts.StreamEvent{
							Type:      contracts.StreamEventContentDelta,
							Content:   part.Text,
							Timestamp: time.Now(),
						}

						if filterContent && capturedEvents != nil {
							// Capture content for potential replay later
							*capturedEvents = append(*capturedEvents, contentEvent)
						} else {
							// Stream content immediately
							select {
							case eventCh <- contentEvent:
							case <-ctx.Done():
								return nil, false, ctx.Err()
							}
						}
					}
				}
			}
		}
	}

	return toolCalls, hasContent, nil
}
