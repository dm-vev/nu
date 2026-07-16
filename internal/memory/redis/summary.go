package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/dm-vev/nu/contracts"
	"github.com/dm-vev/nu/internal/memory/conversation"
	"github.com/dm-vev/nu/internal/multitenancy"
)

// checkAndSummarize checks if summarization is needed and performs it
func (r *RedisMemory) checkAndSummarize(ctx context.Context) error {
	// Get conversation ID from context
	conversationID, err := conversation.ConversationKey(ctx)
	if err != nil {
		return fmt.Errorf("failed to get conversation ID: %w", err)
	}

	// Get organization ID from context
	orgID, err := multitenancy.GetOrgID(ctx)
	if err != nil {
		orgID = "default"
	}

	// Create Redis key
	key := fmt.Sprintf("%s%s:%s", r.keyPrefix, orgID, conversationID)

	// Get message count
	count, err := r.client.LLen(ctx, key).Result()
	if err != nil {
		return fmt.Errorf("failed to get message count: %w", err)
	}

	// Check if we need to summarize
	if count < int64(r.messageThreshold) {
		return nil
	}

	// Get messages to summarize (all but the most recent ones)
	keepRecent := r.messageThreshold / 3 // Keep 1/3 of threshold as recent messages
	summarizeCount := int(count) - keepRecent

	// Get messages to summarize
	results, err := r.client.LRange(ctx, key, 0, int64(summarizeCount-1)).Result()
	if err != nil {
		return fmt.Errorf("failed to get messages for summarization: %w", err)
	}

	// Parse messages
	var messages []contracts.Message
	for _, result := range results {
		var message contracts.Message
		if err := json.Unmarshal([]byte(result), &message); err != nil {
			return fmt.Errorf("failed to unmarshal message: %w", err)
		}
		messages = append(messages, message)
	}

	// Create summary
	summary, err := r.createSummary(ctx, messages)
	if err != nil {
		return fmt.Errorf("failed to create summary: %w", err)
	}

	// Store summary
	if err := r.storeSummary(ctx, summary); err != nil {
		return fmt.Errorf("failed to store summary: %w", err)
	}

	// Remove summarized messages from the main list
	for i := 0; i < summarizeCount; i++ {
		if err := r.client.LPop(ctx, key).Err(); err != nil {
			return fmt.Errorf("failed to remove summarized message: %w", err)
		}
	}

	// Rotate summaries if needed
	if err := r.rotateSummaries(ctx); err != nil {
		return fmt.Errorf("failed to rotate summaries: %w", err)
	}

	return nil
}

// createSummary generates a summary of the given messages using the LLM
func (r *RedisMemory) createSummary(ctx context.Context, messages []contracts.Message) (contracts.Message, error) {
	// Format messages for summarization
	var sb strings.Builder
	sb.WriteString("Summarize the following conversation concisely, preserving key information and context:\n\n")

	for _, msg := range messages {
		fmt.Fprintf(&sb, "%s: %s\n", msg.Role, msg.Content)
	}

	sb.WriteString("\nProvide a concise summary that captures the essential information from this conversation.")

	// Generate summary
	summary, err := r.llmClient.Generate(ctx, sb.String(), func(o *contracts.GenerateOptions) {
		o.LLMConfig = &contracts.LLMConfig{
			Temperature: 0.3, // Lower temperature for more consistent summaries
		}
	})
	if err != nil {
		return contracts.Message{}, fmt.Errorf("failed to generate summary: %w", err)
	}

	// Create summary message
	summaryMessage := contracts.Message{
		Role:    "system",
		Content: fmt.Sprintf("Previous conversation summary (%d messages): %s", len(messages), strings.TrimSpace(summary)),
		Metadata: map[string]interface{}{
			"is_summary":    true,
			"message_count": len(messages),
			"summarized_at": time.Now().Unix(),
		},
	}

	return summaryMessage, nil
}

// storeSummary stores a summary in Redis
func (r *RedisMemory) storeSummary(ctx context.Context, summary contracts.Message) error {
	// Get conversation ID from context
	conversationID, err := conversation.ConversationKey(ctx)
	if err != nil {
		return fmt.Errorf("failed to get conversation ID: %w", err)
	}

	// Get organization ID from context
	orgID, err := multitenancy.GetOrgID(ctx)
	if err != nil {
		orgID = "default"
	}

	// Create Redis key for summaries
	summaryKey := fmt.Sprintf("%s%s:%s", r.summaryKeyPrefix, orgID, conversationID)

	// Marshal summary
	summaryJSON, err := json.Marshal(summary)
	if err != nil {
		return fmt.Errorf("failed to marshal summary: %w", err)
	}

	// Add summary to Redis list
	if err := r.client.RPush(ctx, summaryKey, summaryJSON).Err(); err != nil {
		return fmt.Errorf("failed to store summary: %w", err)
	}

	// Set TTL on the summary key
	r.client.Expire(ctx, summaryKey, r.ttl)

	return nil
}

// getSummaries retrieves summaries from Redis
func (r *RedisMemory) getSummaries(ctx context.Context) ([]contracts.Message, error) {
	// Get conversation ID from context
	conversationID, err := conversation.ConversationKey(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get conversation ID: %w", err)
	}

	// Get organization ID from context
	orgID, err := multitenancy.GetOrgID(ctx)
	if err != nil {
		orgID = "default"
	}

	// Create Redis key for summaries
	summaryKey := fmt.Sprintf("%s%s:%s", r.summaryKeyPrefix, orgID, conversationID)

	// Get all summaries from Redis
	results, err := r.client.LRange(ctx, summaryKey, 0, -1).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get summaries from Redis: %w", err)
	}

	// Parse summaries
	var summaries []contracts.Message
	for _, result := range results {
		var summary contracts.Message
		if err := json.Unmarshal([]byte(result), &summary); err != nil {
			return nil, fmt.Errorf("failed to unmarshal summary: %w", err)
		}
		summaries = append(summaries, summary)
	}

	return summaries, nil
}

// rotateSummaries ensures we only keep the configured number of summaries
func (r *RedisMemory) rotateSummaries(ctx context.Context) error {
	// Get conversation ID from context
	conversationID, err := conversation.ConversationKey(ctx)
	if err != nil {
		return fmt.Errorf("failed to get conversation ID: %w", err)
	}

	// Get organization ID from context
	orgID, err := multitenancy.GetOrgID(ctx)
	if err != nil {
		orgID = "default"
	}

	// Create Redis key for summaries
	summaryKey := fmt.Sprintf("%s%s:%s", r.summaryKeyPrefix, orgID, conversationID)

	// Get summary count
	count, err := r.client.LLen(ctx, summaryKey).Result()
	if err != nil {
		return fmt.Errorf("failed to get summary count: %w", err)
	}

	// Remove old summaries if we exceed the limit
	if count > int64(r.summaryCount) {
		removeCount := int(count) - r.summaryCount
		for i := 0; i < removeCount; i++ {
			if err := r.client.LPop(ctx, summaryKey).Err(); err != nil {
				return fmt.Errorf("failed to remove old summary: %w", err)
			}
		}
	}

	return nil
}
