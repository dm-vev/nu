package ui

import (
	"encoding/json"
	"fmt"
	nethttp "net/http"
)

// getRemoteMemoryConversations gets conversations from a remote agent via HTTP
func (h *Server) getRemoteMemoryConversations(limit, offset int) MemoryResponse {
	remoteURL := h.Agent.GetRemoteURL()
	if remoteURL == "" {
		return MemoryResponse{
			Mode:          "conversations",
			Conversations: []ConversationInfo{},
			Total:         0,
			Limit:         limit,
			Offset:        offset,
		}
	}

	// Make HTTP request to remote agent's memory endpoint
	url := fmt.Sprintf("%s/api/v1/memory?limit=%d&offset=%d", remoteURL, limit, offset)

	// #nosec G107,G704 - URL is constructed from validated parameters
	resp, err := nethttp.Get(url)
	if err != nil {
		return MemoryResponse{
			Mode:          "conversations",
			Conversations: []ConversationInfo{},
			Total:         0,
			Limit:         limit,
			Offset:        offset,
		}
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != nethttp.StatusOK {
		return MemoryResponse{
			Mode:          "conversations",
			Conversations: []ConversationInfo{},
			Total:         0,
			Limit:         limit,
			Offset:        offset,
		}
	}

	var remoteResponse MemoryResponse
	if err := json.NewDecoder(resp.Body).Decode(&remoteResponse); err != nil {
		return MemoryResponse{
			Mode:          "conversations",
			Conversations: []ConversationInfo{},
			Total:         0,
			Limit:         limit,
			Offset:        offset,
		}
	}

	return remoteResponse
}

// getRemoteMemoryMessages gets messages for a specific conversation from a remote agent via HTTP
func (h *Server) getRemoteMemoryMessages(conversationID string, limit, offset int) MemoryResponse {
	remoteURL := h.Agent.GetRemoteURL()
	if remoteURL == "" {
		return MemoryResponse{
			Mode:           "messages",
			Messages:       []MemoryEntry{},
			Total:          0,
			Limit:          limit,
			Offset:         offset,
			ConversationID: conversationID,
		}
	}

	// Make HTTP request to remote agent's memory endpoint for specific conversation
	url := fmt.Sprintf("%s/api/v1/memory?conversation_id=%s&limit=%d&offset=%d",
		remoteURL, conversationID, limit, offset)

	// #nosec G107,G704 - URL is constructed from validated parameters
	resp, err := nethttp.Get(url)
	if err != nil {
		return MemoryResponse{
			Mode:           "messages",
			Messages:       []MemoryEntry{},
			Total:          0,
			Limit:          limit,
			Offset:         offset,
			ConversationID: conversationID,
		}
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != nethttp.StatusOK {
		return MemoryResponse{
			Mode:           "messages",
			Messages:       []MemoryEntry{},
			Total:          0,
			Limit:          limit,
			Offset:         offset,
			ConversationID: conversationID,
		}
	}

	var remoteResponse MemoryResponse
	if err := json.NewDecoder(resp.Body).Decode(&remoteResponse); err != nil {
		return MemoryResponse{
			Mode:           "messages",
			Messages:       []MemoryEntry{},
			Total:          0,
			Limit:          limit,
			Offset:         offset,
			ConversationID: conversationID,
		}
	}

	return remoteResponse
}
