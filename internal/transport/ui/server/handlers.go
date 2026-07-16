package server

import (
	"encoding/json"
	"fmt"
	"log"
	nethttp "net/http"
	"strconv"

	"github.com/dm-vev/nu/internal/memory/conversation"
)

// handleConfig provides detailed agent configuration
func (h *Server) handleConfig(w nethttp.ResponseWriter, r *nethttp.Request) {
	if r.Method != "GET" {
		nethttp.Error(w, "Method not allowed", nethttp.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	// Get agent tools - directly from agent interface
	tools := h.getToolNames()

	// Get system prompt - handle remote agents differently
	systemPrompt := h.getSystemPrompt()

	// Get model info - try to get from LLM
	model := h.getModelName()

	// Get memory info - directly from agent interface
	memInfo := h.getMemoryInfo()

	// Get datastore info
	datastoreInfo := h.getDataStoreInfo()

	response := AgentConfigResponse{
		Name:         h.Agent.GetName(),
		Description:  h.Agent.GetDescription(),
		Model:        model,
		SystemPrompt: systemPrompt,
		Tools:        tools,
		Memory:       memInfo,
		DataStore:    datastoreInfo,
		Features:     h.uiConfig.Features,
		UITheme:      h.uiConfig.Theme,
		SubAgents:    h.getSubAgentsList(),
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		nethttp.Error(w, "Failed to encode response", nethttp.StatusInternalServerError)
	}
}

// handleSubAgents provides list of sub-agents
func (h *Server) handleSubAgents(w nethttp.ResponseWriter, r *nethttp.Request) {
	if r.Method != "GET" {
		nethttp.Error(w, "Method not allowed", nethttp.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	subAgents := h.getSubAgentsList()

	if err := json.NewEncoder(w).Encode(map[string]interface{}{
		"sub_agents": subAgents,
		"count":      len(subAgents),
	}); err != nil {
		nethttp.Error(w, "Failed to encode response", nethttp.StatusInternalServerError)
	}
}

// handleDelegate handles delegation to sub-agents
func (h *Server) handleDelegate(w nethttp.ResponseWriter, r *nethttp.Request) {
	if r.Method != "POST" {
		nethttp.Error(w, "Method not allowed", nethttp.StatusMethodNotAllowed)
		return
	}

	var req DelegateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		nethttp.Error(w, fmt.Sprintf("Invalid JSON: %v", err), nethttp.StatusBadRequest)
		return
	}

	// Build context
	ctx := r.Context()
	if req.ConversationID != "" {
		ctx = conversation.WithConversationID(ctx, req.ConversationID)
	}
	_ = ctx // TODO: Use ctx when implementing actual delegation logic

	// Here you would implement the actual delegation logic
	// For now, we'll return a placeholder response
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(map[string]interface{}{
		"status":       "delegated",
		"sub_agent_id": req.SubAgentID,
		"task":         req.Task,
		"result":       "Sub-agent delegation not yet implemented",
	}); err != nil {
		nethttp.Error(w, "Failed to encode response", nethttp.StatusInternalServerError)
	}
}

// handleMemory provides memory browser functionality
func (h *Server) handleMemory(w nethttp.ResponseWriter, r *nethttp.Request) {
	if r.Method != "GET" {
		nethttp.Error(w, "Method not allowed", nethttp.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	// Parse query parameters
	limit := 100
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	offset := 0
	if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		}
	}

	conversationID := r.URL.Query().Get("conversation_id")

	var response MemoryResponse

	if conversationID != "" {
		// Get messages for specific conversation
		log.Printf("Getting messages for conversation: %s", conversationID) // #nosec G706 - conversationID is a UUID from internal routing
		response = h.getConversationMessagesWithContext(r.Context(), conversationID, limit, offset)
	} else {
		// Get all conversations
		log.Println("Getting all conversations")
		response = h.getAllConversationsWithContext(r.Context(), limit, offset)
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		nethttp.Error(w, "Failed to encode response", nethttp.StatusInternalServerError)
	}
}

// handleMemorySearch provides memory search functionality
func (h *Server) handleMemorySearch(w nethttp.ResponseWriter, r *nethttp.Request) {
	if r.Method != "GET" {
		nethttp.Error(w, "Method not allowed", nethttp.StatusMethodNotAllowed)
		return
	}

	query := r.URL.Query().Get("q")
	if query == "" {
		nethttp.Error(w, "Query parameter 'q' is required", nethttp.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	// Search conversation history
	results := h.searchConversationHistory(query)

	if err := json.NewEncoder(w).Encode(map[string]interface{}{
		"query":   query,
		"results": results,
		"count":   len(results),
	}); err != nil {
		nethttp.Error(w, "Failed to encode response", nethttp.StatusInternalServerError)
	}
}

// handleTools provides list of available tools
func (h *Server) handleTools(w nethttp.ResponseWriter, r *nethttp.Request) {
	if r.Method != "GET" {
		nethttp.Error(w, "Method not allowed", nethttp.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	tools := []map[string]interface{}{}

	// Check if agent is remote and handle accordingly
	if h.Agent.IsRemote() {
		// For remote agents, get tools from system prompt or use alternative method
		// Parse system prompt to extract tool information
		systemPrompt := h.getSystemPrompt()
		toolNames := h.parseToolsFromSystemPrompt(systemPrompt)
		for _, toolName := range toolNames {
			tools = append(tools, map[string]interface{}{
				"name":        toolName,
				"description": "Remote agent tool",
				"enabled":     true,
			})
		}
	} else {
		// Get tools from local agent
		agentTools := h.Agent.GetTools()
		for _, tool := range agentTools {
			tools = append(tools, map[string]interface{}{
				"name":        tool.Name(),
				"description": tool.Description(),
				"enabled":     true,
			})
		}
	}

	if err := json.NewEncoder(w).Encode(map[string]interface{}{
		"tools": tools,
		"count": len(tools),
	}); err != nil {
		nethttp.Error(w, "Failed to encode response", nethttp.StatusInternalServerError)
	}
}

// handleWebSocketChat handles WebSocket connections for real-time chat
func (h *Server) handleWebSocketChat(w nethttp.ResponseWriter, r *nethttp.Request) {
	// WebSocket implementation would go here
	// For now, return not implemented
	nethttp.Error(w, "WebSocket not yet implemented", nethttp.StatusNotImplemented)
}
