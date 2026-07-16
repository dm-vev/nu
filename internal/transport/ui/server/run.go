package server

import (
	"encoding/json"
	"fmt"
	"log"
	nethttp "net/http"
	"time"

	"github.com/dm-vev/nu/internal/memory/conversation"
	"github.com/dm-vev/nu/internal/multitenancy"
	httpserver "github.com/dm-vev/nu/internal/transport/http/server"
)

// addToConversationHistory adds an entry to local conversation history
func (h *Server) addToConversationHistory(role, content string, metadata map[string]interface{}) {
	entry := MemoryEntry{
		ID:        fmt.Sprintf("local_%d", time.Now().UnixNano()),
		Role:      role,
		Content:   content,
		Timestamp: time.Now().UnixMilli(),
		Metadata:  metadata,
	}

	h.conversationHistory = append(h.conversationHistory, entry)

	// Keep only last 1000 entries to avoid memory issues
	if len(h.conversationHistory) > 1000 {
		h.conversationHistory = h.conversationHistory[len(h.conversationHistory)-1000:]
	}
}

// handleRun handles non-streaming agent requests and captures conversations
func (h *Server) handleRun(w nethttp.ResponseWriter, r *nethttp.Request) {
	if r.Method != "POST" {
		nethttp.Error(w, "Method not allowed", nethttp.StatusMethodNotAllowed)
		return
	}

	var req httpserver.StreamRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		nethttp.Error(w, "Invalid JSON", nethttp.StatusBadRequest)
		return
	}

	if req.Input == "" {
		nethttp.Error(w, "Input is required", nethttp.StatusBadRequest)
		return
	}

	// Set up context with org ID if provided
	ctx := r.Context()
	if req.OrgID != "" {
		ctx = multitenancy.WithOrgID(ctx, req.OrgID)
	}

	// Add conversation ID if provided
	if req.ConversationID != "" {
		ctx = conversation.WithConversationID(ctx, req.ConversationID)
	}

	// Add user input to conversation history
	h.addToConversationHistory("user", req.Input, map[string]interface{}{
		"conversation_id": req.ConversationID,
		"org_id":          req.OrgID,
	})

	// Execute agent with detailed tracking
	response, err := h.Agent.RunDetailed(ctx, req.Input)

	// Add response to conversation history
	if err != nil {
		h.addToConversationHistory("error", err.Error(), map[string]interface{}{
			"conversation_id": req.ConversationID,
			"org_id":          req.OrgID,
		})

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"error":  err.Error(),
			"output": "",
		})
		return
	}

	// Log detailed execution information for UI chat
	{
		executionDetails := map[string]interface{}{
			"endpoint":          "ui_chat",
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
		log.Printf("[UI Server] Agent execution completed via UI chat: %+v", executionDetails)
	}

	h.addToConversationHistory("assistant", response.Content, map[string]interface{}{
		"conversation_id": req.ConversationID,
		"org_id":          req.OrgID,
	})

	w.Header().Set("Content-Type", "application/json")
	responseData := map[string]interface{}{
		"output":            response.Content,
		"error":             "",
		"execution_summary": response.ExecutionSummary,
	}
	if response.Usage != nil {
		responseData["usage"] = response.Usage
	}
	_ = json.NewEncoder(w).Encode(responseData)
}
