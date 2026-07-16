package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/dm-vev/nu/contracts"
)

// GetAllConversationsAcrossOrgs returns all conversation IDs from all organizations
func (r *RedisMemory) GetAllConversationsAcrossOrgs() (map[string][]string, error) {
	ctx := context.Background()

	// Search for all keys matching any org pattern
	pattern := fmt.Sprintf("%s*", r.keyPrefix)
	keys, err := r.client.Keys(ctx, pattern).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get all conversation keys: %w", err)
	}

	orgConversations := make(map[string][]string)

	for _, key := range keys {
		// Extract orgID and conversationID from key
		// Key format: keyPrefix + orgID + ":" + conversationID
		if strings.HasPrefix(key, r.keyPrefix) {
			remainder := strings.TrimPrefix(key, r.keyPrefix)
			parts := strings.SplitN(remainder, ":", 2)
			if len(parts) == 2 {
				orgID := parts[0]
				conversationID := parts[1]
				orgConversations[orgID] = append(orgConversations[orgID], conversationID)
			}
		}
	}

	return orgConversations, nil
}

// GetConversationMessagesAcrossOrgs finds conversation in any org and returns messages
func (r *RedisMemory) GetConversationMessagesAcrossOrgs(conversationID string) ([]contracts.Message, string, error) {
	ctx := context.Background()

	// Search for the conversation across all orgs
	pattern := fmt.Sprintf("%s*:%s", r.keyPrefix, conversationID)
	keys, err := r.client.Keys(ctx, pattern).Result()
	if err != nil {
		return nil, "", fmt.Errorf("failed to search for conversation: %w", err)
	}

	if len(keys) == 0 {
		return []contracts.Message{}, "", nil // Conversation not found
	}

	// Use the first match (there should typically be only one)
	key := keys[0]

	// Extract orgID from key
	if strings.HasPrefix(key, r.keyPrefix) {
		remainder := strings.TrimPrefix(key, r.keyPrefix)
		parts := strings.SplitN(remainder, ":", 2)
		if len(parts) == 2 {
			orgID := parts[0]

			// Get messages from Redis
			data, err := r.client.LRange(ctx, key, 0, -1).Result()
			if err != nil {
				return nil, "", fmt.Errorf("failed to get conversation messages: %w", err)
			}

			messages := make([]contracts.Message, 0, len(data))
			for _, item := range data {
				var msg contracts.Message
				if err := json.Unmarshal([]byte(item), &msg); err != nil {
					continue // Skip invalid messages
				}
				messages = append(messages, msg)
			}

			return messages, orgID, nil
		}
	}

	return []contracts.Message{}, "", nil
}

// GetMemoryStatisticsAcrossOrgs returns memory statistics across all organizations
func (r *RedisMemory) GetMemoryStatisticsAcrossOrgs() (totalConversations, totalMessages int, err error) {
	ctx := context.Background()

	// Search for all keys matching any org pattern
	pattern := fmt.Sprintf("%s*", r.keyPrefix)
	keys, err := r.client.Keys(ctx, pattern).Result()
	if err != nil {
		return 0, 0, fmt.Errorf("failed to get all conversation keys: %w", err)
	}

	totalConversations = len(keys)
	totalMessages = 0

	// Count messages in each conversation
	for _, key := range keys {
		count, err := r.client.LLen(ctx, key).Result()
		if err != nil {
			continue // Skip if we can't get count
		}
		totalMessages += int(count)
	}

	return totalConversations, totalMessages, nil
}
