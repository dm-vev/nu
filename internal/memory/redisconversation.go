package memory

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"nu/internal/contracts"
	"nu/internal/multitenancy"
)

// GetAllConversations returns all conversation IDs for the current org
func (r *RedisMemory) GetAllConversations(ctx context.Context) ([]string, error) {
	// Get organization ID from context
	orgID, err := multitenancy.GetOrgID(ctx)
	if err != nil {
		return nil, fmt.Errorf("organization ID not found in context: %w", err)
	}

	// Search for all keys matching the pattern for this org
	pattern := fmt.Sprintf("%s%s:*", r.keyPrefix, orgID)
	keys, err := r.client.Keys(ctx, pattern).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get conversation keys: %w", err)
	}

	// Extract conversation IDs from keys
	conversations := make([]string, 0, len(keys))
	expectedPrefix := fmt.Sprintf("%s%s:", r.keyPrefix, orgID)

	for _, key := range keys {
		if strings.HasPrefix(key, expectedPrefix) {
			conversationID := strings.TrimPrefix(key, expectedPrefix)
			conversations = append(conversations, conversationID)
		}
	}

	return conversations, nil
}

// GetConversationMessages gets all messages for a specific conversation in current org
func (r *RedisMemory) GetConversationMessages(ctx context.Context, conversationID string) ([]contracts.Message, error) {
	// Get organization ID from context
	orgID, err := multitenancy.GetOrgID(ctx)
	if err != nil {
		return nil, fmt.Errorf("organization ID not found in context: %w", err)
	}

	// Create Redis key
	key := fmt.Sprintf("%s%s:%s", r.keyPrefix, orgID, conversationID)

	// Get all messages from Redis list
	data, err := r.client.LRange(ctx, key, 0, -1).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get conversation messages: %w", err)
	}

	messages := make([]contracts.Message, 0, len(data))
	for _, item := range data {
		var msg contracts.Message
		if err := json.Unmarshal([]byte(item), &msg); err != nil {
			continue // Skip invalid messages
		}
		messages = append(messages, msg)
	}

	return messages, nil
}

// GetMemoryStatistics returns basic memory statistics for current org
func (r *RedisMemory) GetMemoryStatistics(ctx context.Context) (totalConversations, totalMessages int, err error) {
	// Get organization ID from context
	orgID, err := multitenancy.GetOrgID(ctx)
	if err != nil {
		return 0, 0, fmt.Errorf("organization ID not found in context: %w", err)
	}

	// Search for all keys matching the pattern for this org
	pattern := fmt.Sprintf("%s%s:*", r.keyPrefix, orgID)
	keys, err := r.client.Keys(ctx, pattern).Result()
	if err != nil {
		return 0, 0, fmt.Errorf("failed to get conversation keys: %w", err)
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
