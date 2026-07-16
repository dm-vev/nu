package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/dm-vev/nu/contracts"
)

// convertAgentEventToHTTPEvent converts agent stream events to HTTP event format
func (h *Server) convertAgentEventToHTTPEvent(event contracts.AgentStreamEvent) StreamEventData {
	eventData := StreamEventData{
		Type:     string(event.Type),
		Content:  event.Content,
		Metadata: event.Metadata,
		IsFinal:  false,
	}

	if event.ThinkingStep != "" {
		eventData.ThinkingStep = event.ThinkingStep
	}

	if event.ToolCall != nil {
		eventData.ToolCall = &ToolCallData{
			ID:        event.ToolCall.ID,
			Name:      event.ToolCall.Name,
			Arguments: event.ToolCall.Arguments,
			Result:    event.ToolCall.Result,
			Status:    event.ToolCall.Status,
		}
	}

	if event.Error != nil {
		eventData.Error = event.Error.Error()
	}

	return eventData
}

// sendSSEEvent sends a Server-Sent Event
func (h *Server) sendSSEEvent(w http.ResponseWriter, flusher http.Flusher, eventType string, data StreamEventData) {
	h.sendSSEEventWithID(w, flusher, eventType, data, "")
}

// sendSSEEventWithID sends a Server-Sent Event with ID
func (h *Server) sendSSEEventWithID(w http.ResponseWriter, flusher http.Flusher, eventType string, data StreamEventData, id string) {
	// Add timestamp
	data.Timestamp = time.Now().UnixMilli()

	// Convert data to JSON
	jsonData, err := json.Marshal(data)
	if err != nil {
		// Fallback to error event
		_, _ = fmt.Fprintf(w, "event: error\ndata: {\"error\": \"Failed to marshal event data\"}\n\n")
		flusher.Flush()
		return
	}

	// Write SSE event
	if id != "" {
		_, _ = fmt.Fprintf(w, "id: %s\n", id)
	}
	_, _ = fmt.Fprintf(w, "event: %s\n", eventType)
	_, _ = fmt.Fprintf(w, "data: %s\n\n", string(jsonData))

	flusher.Flush()
}
