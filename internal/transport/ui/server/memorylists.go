package server

import (
	"fmt"
	"time"

	"github.com/dm-vev/nu/contracts"
)

// buildConversationListFromAllOrgs builds conversation list from all organizations
func (h *Server) buildConversationListFromAllOrgs(adminMem contracts.AdminConversationMemory, limit, offset int) MemoryResponse {
	orgConversations, err := adminMem.GetAllConversationsAcrossOrgs()
	if err != nil {
		// Return empty response on error
		return MemoryResponse{
			Mode:          "conversations",
			Conversations: []ConversationInfo{},
			Total:         0,
			Limit:         limit,
			Offset:        offset,
		}
	}

	var allConversationInfos []ConversationInfo

	// Iterate through all orgs and their conversations
	for orgID, conversations := range orgConversations {
		for _, convID := range conversations {
			// Get messages to determine last activity and message count
			messages, foundOrgID, err := adminMem.GetConversationMessagesAcrossOrgs(convID)
			if err != nil || foundOrgID != orgID {
				continue
			}

			if len(messages) > 0 {
				lastMessage := messages[len(messages)-1]

				// Truncate last message content for preview
				lastContent := lastMessage.Content
				if len(lastContent) > 100 {
					lastContent = lastContent[:100] + "..."
				}

				// Include orgID in conversation display
				displayID := fmt.Sprintf("[%s] %s", orgID, convID)

				allConversationInfos = append(allConversationInfos, ConversationInfo{
					ID:           convID, // Keep original ID for API calls
					MessageCount: len(messages),
					LastActivity: time.Now().Unix(), // TODO: get actual timestamp from last message
					LastMessage:  displayID + ": " + lastContent,
				})
			}
		}
	}

	// Apply pagination
	total := len(allConversationInfos)
	start := offset
	end := offset + limit
	if start >= total {
		allConversationInfos = []ConversationInfo{}
	} else {
		if end > total {
			end = total
		}
		allConversationInfos = allConversationInfos[start:end]
	}

	return MemoryResponse{
		Mode:          "conversations",
		Conversations: allConversationInfos,
		Total:         total,
		Limit:         limit,
		Offset:        offset,
	}
}

// buildConversationListFromLocalAllOrgs builds conversation list from local history across all orgs
func (h *Server) buildConversationListFromLocalAllOrgs(limit, offset int) MemoryResponse {
	// Group local conversation history by conversation ID (ignoring org isolation)
	conversationMap := make(map[string][]MemoryEntry)

	for _, entry := range h.conversationHistory {
		convID := entry.ConversationID
		if convID == "" {
			convID = "default"
		}
		conversationMap[convID] = append(conversationMap[convID], entry)
	}

	var conversationInfos []ConversationInfo
	for convID, entries := range conversationMap {
		if len(entries) > 0 {
			lastEntry := entries[len(entries)-1]
			lastContent := lastEntry.Content
			if len(lastContent) > 100 {
				lastContent = lastContent[:100] + "..."
			}

			conversationInfos = append(conversationInfos, ConversationInfo{
				ID:           convID,
				MessageCount: len(entries),
				LastActivity: lastEntry.Timestamp,
				LastMessage:  lastContent,
			})
		}
	}

	// Apply pagination
	total := len(conversationInfos)
	start := offset
	end := offset + limit
	if start >= total {
		conversationInfos = []ConversationInfo{}
	} else {
		if end > total {
			end = total
		}
		conversationInfos = conversationInfos[start:end]
	}

	return MemoryResponse{
		Mode:          "conversations",
		Conversations: conversationInfos,
		Total:         total,
		Limit:         limit,
		Offset:        offset,
	}
}

// buildMessageListFromAllOrgs builds message list from all organizations
func (h *Server) buildMessageListFromAllOrgs(adminMem contracts.AdminConversationMemory, conversationID string, limit, offset int) MemoryResponse {
	messages, orgID, err := adminMem.GetConversationMessagesAcrossOrgs(conversationID)
	if err != nil {
		// Return empty response on error
		return MemoryResponse{
			Mode:           "messages",
			Messages:       []MemoryEntry{},
			Total:          0,
			Limit:          limit,
			Offset:         offset,
			ConversationID: conversationID,
		}
	}

	var memoryEntries []MemoryEntry

	for i, msg := range messages {
		// Extract tool calls if present
		toolCallsInfo := ""
		if len(msg.ToolCalls) > 0 {
			toolCallsInfo = fmt.Sprintf(" [%d tool calls]", len(msg.ToolCalls))
		}

		// Include org info in the message content
		content := msg.Content + toolCallsInfo
		if orgID != "" {
			content = fmt.Sprintf("[%s] %s", orgID, content)
		}

		memoryEntries = append(memoryEntries, MemoryEntry{
			ID:             fmt.Sprintf("agent_msg_%d", i),
			Role:           string(msg.Role),
			Content:        content,
			Timestamp:      time.Now().Unix(), // TODO: get actual timestamp from message
			ConversationID: conversationID,
			Metadata:       msg.Metadata,
		})
	}

	// Apply pagination
	total := len(memoryEntries)
	start := offset
	end := offset + limit
	if start >= total {
		memoryEntries = []MemoryEntry{}
	} else {
		if end > total {
			end = total
		}
		memoryEntries = memoryEntries[start:end]
	}

	return MemoryResponse{
		Mode:           "messages",
		Messages:       memoryEntries,
		Total:          total,
		Limit:          limit,
		Offset:         offset,
		ConversationID: conversationID,
	}
}

// buildMessageListFromLocalAllOrgs builds message list from local history across all orgs
func (h *Server) buildMessageListFromLocalAllOrgs(conversationID string, limit, offset int) MemoryResponse {
	var filteredEntries []MemoryEntry

	// Filter entries by conversation ID (ignoring org isolation)
	for _, entry := range h.conversationHistory {
		entryConvID := entry.ConversationID
		if entryConvID == "" {
			entryConvID = "default"
		}
		if entryConvID == conversationID {
			filteredEntries = append(filteredEntries, entry)
		}
	}

	// Apply pagination
	total := len(filteredEntries)
	start := offset
	end := offset + limit
	if start >= total {
		filteredEntries = []MemoryEntry{}
	} else {
		if end > total {
			end = total
		}
		filteredEntries = filteredEntries[start:end]
	}

	return MemoryResponse{
		Mode:           "messages",
		Messages:       filteredEntries,
		Total:          total,
		Limit:          limit,
		Offset:         offset,
		ConversationID: conversationID,
	}
}
