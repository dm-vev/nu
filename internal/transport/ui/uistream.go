package ui

import (
	"encoding/json"
	"fmt"
	"log"
	nethttp "net/http"
	"strings"
	"time"

	"nu/internal/contracts"
	"nu/internal/memory"
	"nu/internal/multitenancy"
	"nu/internal/transport/http"
)

// handleStream handles streaming agent requests and captures conversations
func (h *Server) handleStream(w nethttp.ResponseWriter, r *nethttp.Request) {
	if r.Method != "POST" {
		nethttp.Error(w, "Method not allowed", nethttp.StatusMethodNotAllowed)
		return
	}

	var req http.StreamRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		nethttp.Error(w, "Invalid JSON", nethttp.StatusBadRequest)
		return
	}

	if req.Input == "" {
		nethttp.Error(w, "Input is required", nethttp.StatusBadRequest)
		return
	}

	// Set up SSE headers
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	// Set up context with org ID if provided
	ctx := r.Context()
	if req.OrgID != "" {
		ctx = multitenancy.WithOrgID(ctx, req.OrgID)
	}

	// Add conversation ID if provided
	if req.ConversationID != "" {
		ctx = memory.WithConversationID(ctx, req.ConversationID)
	}

	// Add user input to conversation history
	h.addToConversationHistory("user", req.Input, map[string]interface{}{
		"conversation_id": req.ConversationID,
		"org_id":          req.OrgID,
	})

	// Check if agent supports streaming
	streamingAgent, ok := interface{}(h.Agent).(contracts.StreamingAgent)
	if !ok {
		// Fall back to non-streaming with detailed tracking
		response, err := h.Agent.RunDetailed(ctx, req.Input)

		if err != nil {
			h.addToConversationHistory("error", err.Error(), map[string]interface{}{
				"conversation_id": req.ConversationID,
				"org_id":          req.OrgID,
			})

			event := http.SSEEvent{
				Event:     "error",
				Data:      http.StreamEventData{Type: "error", Content: err.Error(), IsFinal: true},
				Timestamp: time.Now().UnixMilli(),
			}
			h.sendSSEEvent(w, event)
			return
		}

		// Log detailed execution information for UI streaming fallback
		{
			executionDetails := map[string]interface{}{
				"endpoint":          "ui_stream_fallback",
				"conversation_id":   req.ConversationID,
				"org_id":            req.OrgID,
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
			log.Printf("[UI Server] Agent execution completed via UI streaming fallback: %+v", executionDetails)
		}

		h.addToConversationHistory("assistant", response.Content, map[string]interface{}{
			"conversation_id": req.ConversationID,
			"org_id":          req.OrgID,
		})

		event := http.SSEEvent{
			Event:     "content",
			Data:      http.StreamEventData{Type: "content", Content: response.Content, IsFinal: true},
			Timestamp: time.Now().UnixMilli(),
		}
		h.sendSSEEvent(w, event)
		return
	}

	// Stream events from agent
	eventChan, err := streamingAgent.RunStream(ctx, req.Input)
	if err != nil {
		h.addToConversationHistory("error", err.Error(), map[string]interface{}{
			"conversation_id": req.ConversationID,
			"org_id":          req.OrgID,
		})

		event := http.SSEEvent{
			Event:     "error",
			Data:      http.StreamEventData{Type: "error", Content: err.Error(), IsFinal: true},
			Timestamp: time.Now().UnixMilli(),
		}
		h.sendSSEEvent(w, event)
		return
	}

	var fullResponse strings.Builder
	for agentEvent := range eventChan {
		// Collect content for conversation history
		if agentEvent.Content != "" && agentEvent.Type == contracts.AgentEventContent {
			fullResponse.WriteString(agentEvent.Content)
		}

		// Convert agent event to stream event data
		eventData := http.StreamEventData{
			Type:         string(agentEvent.Type),
			Content:      agentEvent.Content,
			ThinkingStep: agentEvent.ThinkingStep,
			IsFinal:      agentEvent.Type == contracts.AgentEventComplete,
		}

		if agentEvent.ToolCall != nil {
			eventData.ToolCall = &http.ToolCallData{
				ID:        agentEvent.ToolCall.ID,
				Name:      agentEvent.ToolCall.Name,
				Arguments: agentEvent.ToolCall.Arguments,
				Result:    agentEvent.ToolCall.Result,
				Status:    agentEvent.ToolCall.Status,
			}
		}

		if agentEvent.Error != nil {
			eventData.Error = agentEvent.Error.Error()
		}

		if agentEvent.Metadata != nil {
			eventData.Metadata = agentEvent.Metadata
		}

		event := http.SSEEvent{
			Event:     string(agentEvent.Type),
			Data:      eventData,
			Timestamp: agentEvent.Timestamp.UnixMilli(),
		}

		h.sendSSEEvent(w, event)

		// Flush for real-time streaming
		if flusher, ok := w.(nethttp.Flusher); ok {
			flusher.Flush()
		}
	}

	// Add final response to conversation history
	if fullResponse.Len() > 0 {
		h.addToConversationHistory("assistant", fullResponse.String(), map[string]interface{}{
			"conversation_id": req.ConversationID,
			"org_id":          req.OrgID,
		})
	}
}

// sendSSEEvent sends a server-sent event
func (h *Server) sendSSEEvent(w nethttp.ResponseWriter, event http.SSEEvent) {
	data, err := json.Marshal(event.Data)
	if err != nil {
		return
	}

	_, _ = fmt.Fprintf(w, "event: %s\n", event.Event)
	_, _ = fmt.Fprintf(w, "data: %s\n", string(data))
	if event.ID != "" {
		_, _ = fmt.Fprintf(w, "id: %s\n", event.ID)
	}
	_, _ = fmt.Fprintf(w, "\n")
}
