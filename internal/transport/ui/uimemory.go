package ui

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"nu/internal/contracts"
)

// getConversationHistory returns conversation history with pagination
func (h *Server) getConversationHistory(limit, offset int) []MemoryEntry {
	// First, try to get from agent's memory system if available
	if memGetter, ok := interface{}(h.Agent).(interface{ GetMemory() contracts.Memory }); ok {
		if mem := memGetter.GetMemory(); mem != nil {
			return h.getMemoryFromAgent(mem, limit, offset)
		}
	}

	// Fallback to our in-memory storage
	total := len(h.conversationHistory)

	if offset >= total {
		return []MemoryEntry{}
	}

	end := offset + limit
	if end > total {
		end = total
	}

	// Return most recent entries first (reverse order)
	result := make([]MemoryEntry, 0, end-offset)
	for i := total - 1 - offset; i >= total-end; i-- {
		if i >= 0 {
			result = append(result, h.conversationHistory[i])
		}
	}

	return result
}

// getAllConversationsWithContext gets all conversations with request context (but ignores org isolation)
func (h *Server) getAllConversationsWithContext(ctx context.Context, limit, offset int) MemoryResponse {
	// For admin/debug view, we want to see all conversations from all orgs
	return h.getAllConversationsFromAllOrgs(limit, offset)
}

// getConversationMessagesWithContext gets messages with request context (but searches all orgs)
func (h *Server) getConversationMessagesWithContext(ctx context.Context, conversationID string, limit, offset int) MemoryResponse {
	// For admin/debug view, search across all orgs for the conversation
	return h.getConversationMessagesFromAllOrgs(conversationID, limit, offset)
}

// getAllConversationsFromAllOrgs gets conversations from all organizations
func (h *Server) getAllConversationsFromAllOrgs(limit, offset int) MemoryResponse {
	// Handle remote agents by making HTTP calls to their memory endpoint
	if h.Agent.IsRemote() {
		log.Println("Fetching conversations from remote agent memory")
		return h.getRemoteMemoryConversations(limit, offset)
	}

	// Check if memory supports cross-org operations
	if adminMem, ok := h.Agent.GetMemory().(contracts.AdminConversationMemory); ok {
		log.Println("Fetching conversations from admin conversation memory across all orgs")
		return h.buildConversationListFromAllOrgs(adminMem, limit, offset)
	}

	// Fallback: build conversation list from local history (all orgs)
	return h.buildConversationListFromLocalAllOrgs(limit, offset)
}

// getConversationMessagesFromAllOrgs searches for conversation across all orgs
func (h *Server) getConversationMessagesFromAllOrgs(conversationID string, limit, offset int) MemoryResponse {
	// Handle remote agents by making HTTP calls to their memory endpoint
	if h.Agent.IsRemote() {
		log.Printf("Fetching messages for conversation %s from remote agent memory", conversationID) // #nosec G706 - conversationID is a UUID from internal routing
		return h.getRemoteMemoryMessages(conversationID, limit, offset)
	}

	// Check if memory supports cross-org operations
	if adminMem, ok := h.Agent.GetMemory().(contracts.AdminConversationMemory); ok {
		log.Printf("Fetching messages for conversation %s from admin conversation memory across all orgs", conversationID) // #nosec G706 - conversationID is a UUID from internal routing
		return h.buildMessageListFromAllOrgs(adminMem, conversationID, limit, offset)
	}

	// Fallback: get messages from local history (search all orgs)
	return h.buildMessageListFromLocalAllOrgs(conversationID, limit, offset)
}

// getMemoryFromAgent retrieves memory from the agent's memory system (Redis, etc.)
func (h *Server) getMemoryFromAgent(mem contracts.Memory, limit, offset int) []MemoryEntry {
	ctx := context.Background()

	// Try to get messages from the agent's memory system
	messages, err := mem.GetMessages(ctx, contracts.WithLimit(limit+offset))
	if err != nil {
		// If we can't get from agent memory, fall back to our local storage
		return h.conversationHistory
	}

	// Convert agent memory messages to UI memory entries
	entries := make([]MemoryEntry, 0, len(messages))
	for i, msg := range messages {
		// Skip offset entries
		if i < offset {
			continue
		}

		entry := MemoryEntry{
			ID:             fmt.Sprintf("agent_mem_%d", i),
			Role:           string(msg.Role),
			Content:        msg.Content,
			Timestamp:      h.extractTimestamp(msg.Metadata),
			ConversationID: h.extractConversationID(msg.Metadata),
			Metadata:       msg.Metadata,
		}
		entries = append(entries, entry)
	}

	// If we got entries from agent memory, return them
	if len(entries) > 0 {
		return entries
	}

	// Otherwise fall back to local storage
	return h.conversationHistory
}

// extractTimestamp extracts timestamp from message metadata
func (h *Server) extractTimestamp(metadata map[string]interface{}) int64 {
	if metadata == nil {
		return time.Now().UnixMilli()
	}

	// Try different timestamp formats
	if ts, ok := metadata["timestamp"].(int64); ok {
		// Convert nanoseconds to milliseconds if needed
		if ts > 1e15 { // If it looks like nanoseconds
			return ts / 1e6
		}
		return ts
	}

	if ts, ok := metadata["timestamp"].(float64); ok {
		return int64(ts)
	}

	if timeStr, ok := metadata["timestamp"].(string); ok {
		if t, err := time.Parse(time.RFC3339, timeStr); err == nil {
			return t.UnixMilli()
		}
	}

	return time.Now().UnixMilli()
}

// extractConversationID extracts conversation ID from message metadata
func (h *Server) extractConversationID(metadata map[string]interface{}) string {
	if metadata == nil {
		return "default"
	}

	if convID, ok := metadata["conversation_id"].(string); ok {
		return convID
	}

	if convID, ok := metadata["conversationId"].(string); ok {
		return convID
	}

	return "default"
}

// searchConversationHistory searches through conversation history
func (h *Server) searchConversationHistory(query string) []MemoryEntry {
	if query == "" {
		return h.getConversationHistory(50, 0)
	}

	query = strings.ToLower(query)
	var results []MemoryEntry

	for i := len(h.conversationHistory) - 1; i >= 0; i-- {
		entry := h.conversationHistory[i]
		if strings.Contains(strings.ToLower(entry.Content), query) ||
			strings.Contains(strings.ToLower(entry.Role), query) {
			results = append(results, entry)
			if len(results) >= 50 { // Limit search results
				break
			}
		}
	}

	return results
}
