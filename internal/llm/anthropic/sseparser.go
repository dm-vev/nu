package anthropic

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/dm-vev/nu/contracts"
)

// anthropicParseSSELine parses a single SSE line from Anthropic's API
func anthropicParseSSELine(line string) (*SSEEvent, error) {
	line = strings.TrimSpace(line)
	if line == "" || strings.HasPrefix(line, ":") {
		return nil, nil
	}

	if strings.HasPrefix(line, "data: ") {
		data := strings.TrimPrefix(line, "data: ")
		if data == "[DONE]" || data == "" || strings.TrimSpace(data) == "" {
			return &SSEEvent{Type: "done"}, nil
		}

		var event SSEEvent
		if err := json.Unmarshal([]byte(data), &event); err != nil {
			return nil, fmt.Errorf("failed to parse SSE event: %w (data: %q)", err, data)
		}
		return &event, nil
	}

	if strings.HasPrefix(line, "event: ") {
		return &SSEEvent{Type: strings.TrimPrefix(line, "event: ")}, nil
	}
	return nil, nil
}

// parseSSEStreamAndCapture parses SSE stream and captures content for memory storage
func (c *Client) parseSSEStreamAndCapture(ctx context.Context, scanner *bufio.Scanner, eventChan chan<- contracts.StreamEvent, req CompletionRequest, prompt string, params *contracts.GenerateOptions) string {
	var accumulatedContent strings.Builder
	var currentEvent *SSEEvent
	thinkingBlocks := make(map[int]bool)
	toolBlocks := make(map[int]struct {
		ID        string
		Name      string
		InputJSON strings.Builder
	})
	lineCount := 0

	for scanner.Scan() {
		lineCount++
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			if currentEvent != nil && len(currentEvent.Data) > 0 {
				if err := c.processCompleteSSEEventAndCapture(ctx, currentEvent, eventChan, thinkingBlocks, toolBlocks, &accumulatedContent); err != nil {
					c.logger.Error(ctx, "Failed to process SSE event", map[string]interface{}{
						"error": err.Error(), "event_type": currentEvent.Type, "event_data": string(currentEvent.Data),
					})
					eventChan <- contracts.StreamEvent{
						Type: contracts.StreamEventError, Error: fmt.Errorf("failed to process SSE event: %w", err), Timestamp: time.Now(),
					}
					break
				}
				currentEvent = nil
			}
			continue
		}

		if strings.HasPrefix(line, ":") {
			continue
		}
		if strings.HasPrefix(line, "event: ") {
			if currentEvent == nil {
				currentEvent = &SSEEvent{}
			}
			currentEvent.Type = strings.TrimPrefix(line, "event: ")
			continue
		}
		if strings.HasPrefix(line, "data: ") {
			dataContent := strings.TrimPrefix(line, "data: ")
			if dataContent == "[DONE]" {
				eventChan <- contracts.StreamEvent{Type: contracts.StreamEventMessageStop, Timestamp: time.Now()}
				break
			}
			if strings.TrimSpace(dataContent) == "" {
				continue
			}
			if currentEvent == nil {
				currentEvent = &SSEEvent{}
			}
			currentEvent.Data = json.RawMessage(dataContent)
		}
	}

	if currentEvent != nil && len(currentEvent.Data) > 0 {
		_ = c.processCompleteSSEEventAndCapture(ctx, currentEvent, eventChan, thinkingBlocks, toolBlocks, &accumulatedContent)
	}

	if err := scanner.Err(); err != nil {
		if ctx.Err() != nil {
			c.logger.Warn(ctx, "Scanner stopped due to context cancellation", map[string]interface{}{
				"error": err.Error(), "context_error": ctx.Err().Error(), "lines_processed": lineCount,
			})
		} else {
			c.logger.Error(ctx, "Scanner error during SSE parsing", map[string]interface{}{
				"error": err.Error(), "lines_processed": lineCount,
			})
			eventChan <- contracts.StreamEvent{
				Type: contracts.StreamEventError, Error: fmt.Errorf("scanner error after %d lines: %w", lineCount, err), Timestamp: time.Now(),
			}
		}
	}

	return accumulatedContent.String()
}

func (c *Client) processCompleteSSEEventAndCapture(ctx context.Context, event *SSEEvent, eventChan chan<- contracts.StreamEvent, thinkingBlocks map[int]bool, toolBlocks map[int]struct {
	ID        string
	Name      string
	InputJSON strings.Builder
}, accumulatedContent *strings.Builder) error {
	if event.Type == "done" || event.Type == "" {
		eventChan <- contracts.StreamEvent{Type: contracts.StreamEventMessageStop, Timestamp: time.Now()}
		return nil
	}

	streamEvent, err := c.convertAnthropicEventToStreamEvent(event, thinkingBlocks, toolBlocks)
	if err != nil {
		return fmt.Errorf("failed to convert event: %w", err)
	}
	if streamEvent != nil {
		if streamEvent.Type == contracts.StreamEventContentDelta && streamEvent.Content != "" {
			accumulatedContent.WriteString(streamEvent.Content)
		}
		eventChan <- *streamEvent
	}
	return nil
}
