package server

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/dm-vev/nu/contracts"
	"github.com/dm-vev/nu/internal/memory/conversation"
	"github.com/dm-vev/nu/internal/multitenancy"
)

// handleHealth provides a health check endpoint
func (h *Server) HandleHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "healthy",
		"agent":  h.Agent.GetName(),
		"time":   time.Now().Unix(),
	}); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}

// handleRun provides non-streaming agent execution
func (h *Server) HandleRun(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req StreamRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, fmt.Sprintf("Invalid JSON: %v", err), http.StatusBadRequest)
		return
	}

	if req.Input == "" {
		http.Error(w, "Input is required", http.StatusBadRequest)
		return
	}

	// Build context
	ctx := r.Context()
	if req.OrgID != "" {
		ctx = multitenancy.WithOrgID(ctx, req.OrgID)
	}
	if req.ConversationID != "" {
		ctx = conversation.WithConversationID(ctx, req.ConversationID)
	}

	// Execute agent with detailed tracking
	response, err := h.Agent.RunDetailed(ctx, req.Input)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"error": err.Error(),
		})
		return
	}

	// Log detailed execution information
	{
		executionDetails := map[string]interface{}{
			"endpoint":          "agent_run",
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
		log.Printf("[HTTP Server] Agent execution completed via HTTP API: %+v", executionDetails)
	}

	// Return result with execution details
	w.Header().Set("Content-Type", "application/json")
	responseData := map[string]interface{}{
		"output":            response.Content,
		"agent":             response.AgentName,
		"execution_summary": response.ExecutionSummary,
	}
	if response.Usage != nil {
		responseData["usage"] = response.Usage
	}
	if err := json.NewEncoder(w).Encode(responseData); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}

// handleMetadata provides agent metadata
func (h *Server) HandleMetadata(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	// Check if agent supports streaming
	_, supportsStreaming := interface{}(h.Agent).(contracts.StreamingAgent)

	if err := json.NewEncoder(w).Encode(map[string]interface{}{
		"name":               h.Agent.GetName(),
		"description":        h.Agent.GetDescription(),
		"supports_streaming": supportsStreaming,
		"capabilities": []string{
			"run",
			"stream",
			"metadata",
		},
		"endpoints": map[string]string{
			"run":      "/api/v1/agent/run",
			"stream":   "/api/v1/agent/stream",
			"metadata": "/api/v1/agent/metadata",
			"health":   "/health",
		},
	}); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}
