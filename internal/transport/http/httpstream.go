package http

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"nu/internal/contracts"
	"nu/internal/memory"
	"nu/internal/multitenancy"
)

// handleStream provides SSE streaming endpoint
func (h *Server) HandleStream(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse request
	var req StreamRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, fmt.Sprintf("Invalid JSON: %v", err), http.StatusBadRequest)
		return
	}

	if req.Input == "" {
		http.Error(w, "Input is required", http.StatusBadRequest)
		return
	}

	// Set SSE headers
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no") // Disable nginx buffering

	// Get flusher for immediate response sending
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "SSE not supported", http.StatusInternalServerError)
		return
	}

	// Build context
	ctx := r.Context()
	if req.OrgID != "" {
		ctx = multitenancy.WithOrgID(ctx, req.OrgID)
	}
	if req.ConversationID != "" {
		ctx = memory.WithConversationID(ctx, req.ConversationID)
	}

	// Check if agent supports streaming
	streamingAgent, ok := interface{}(h.Agent).(contracts.StreamingAgent)
	if !ok {
		// Fall back to non-streaming execution
		response, err := h.Agent.RunDetailed(ctx, req.Input)
		if err != nil {
			h.sendSSEEvent(w, flusher, "error", StreamEventData{
				Type:    "error",
				Error:   err.Error(),
				IsFinal: true,
			})
			return
		}

		// Log detailed execution information for streaming fallback
		{
			executionDetails := map[string]interface{}{
				"endpoint":          "agent_stream_fallback",
				"agent_name":        response.AgentName,
				"model_used":        response.Model,
				"response_length":   len(response.Content),
				"llm_calls":         response.ExecutionSummary.LLMCalls,
				"tool_calls":        response.ExecutionSummary.ToolCalls,
				"sub_agent_calls":   response.ExecutionSummary.SubAgentCalls,
				"execution_time_ms": response.ExecutionSummary.ExecutionTimeMs,
				"used_tools":        response.ExecutionSummary.UsedTools,
				"used_sub_agents":   response.ExecutionSummary.UsedSubAgents,
			}
			if response.Usage != nil {
				executionDetails["input_tokens"] = response.Usage.InputTokens
				executionDetails["output_tokens"] = response.Usage.OutputTokens
				executionDetails["total_tokens"] = response.Usage.TotalTokens
				executionDetails["reasoning_tokens"] = response.Usage.ReasoningTokens
			}
			log.Printf("[HTTP Server] Agent execution completed via streaming fallback: %+v", executionDetails)
		}

		h.sendSSEEvent(w, flusher, "content", StreamEventData{
			Type:    "content",
			Content: response.Content,
			IsFinal: true,
		})
		return
	}

	// Start streaming
	eventChan, err := streamingAgent.RunStream(ctx, req.Input)
	if err != nil {
		h.sendSSEEvent(w, flusher, "error", StreamEventData{
			Type:    "error",
			Error:   err.Error(),
			IsFinal: true,
		})
		return
	}

	// Send initial connection event
	h.sendSSEEvent(w, flusher, "connected", StreamEventData{
		Type: "connected",
		Metadata: map[string]interface{}{
			"agent": h.Agent.GetName(),
		},
	})

	// Stream events to client
	eventID := 0
	for event := range eventChan {
		eventID++

		// Convert agent event to HTTP event data
		eventData := h.convertAgentEventToHTTPEvent(event)

		// Determine event type for SSE
		var sseEventType string
		switch event.Type {
		case contracts.AgentEventContent:
			sseEventType = "content"
		case contracts.AgentEventThinking:
			sseEventType = "thinking"
		case contracts.AgentEventToolCall:
			sseEventType = "tool_call"
		case contracts.AgentEventToolResult:
			sseEventType = "tool_result"
		case contracts.AgentEventError:
			sseEventType = "error"
		case contracts.AgentEventComplete:
			sseEventType = "complete"
			eventData.IsFinal = true
		default:
			sseEventType = "content"
		}

		// Send SSE event
		h.sendSSEEventWithID(w, flusher, sseEventType, eventData, strconv.Itoa(eventID))

		// Check if client disconnected
		select {
		case <-ctx.Done():
			return
		default:
		}
	}

	// Send final completion event
	h.sendSSEEvent(w, flusher, "done", StreamEventData{
		Type:    "done",
		IsFinal: true,
	})
}
