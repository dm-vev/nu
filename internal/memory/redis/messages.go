package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"time"

	"github.com/dm-vev/nu/contracts"
	"github.com/dm-vev/nu/internal/memory/conversation"
	"github.com/dm-vev/nu/internal/multitenancy"
)

// AddMessage adds a message to the memory with improved error handling and retry logic
func (r *RedisMemory) AddMessage(ctx context.Context, message contracts.Message) error {
	// Get conversation ID from context
	conversationID, err := conversation.ConversationKey(ctx)
	if err != nil {
		return err
	}

	// Get organization ID from context for multi-tenancy support
	orgID, err := multitenancy.GetOrgID(ctx)
	if err != nil {
		// If no organization ID is found, use a default
		orgID = "default"
	}

	// Create Redis key with org and conversation IDs for proper isolation
	key := fmt.Sprintf("%s%s:%s", r.keyPrefix, orgID, conversationID)

	// Validate message size if configured
	if r.maxMessageSize > 0 {
		messageBytes, err := json.Marshal(message)
		if err != nil {
			return fmt.Errorf("failed to marshal message: %w", err)
		}
		if len(messageBytes) > r.maxMessageSize {
			return fmt.Errorf("message size exceeds maximum allowed size of %d bytes", r.maxMessageSize)
		}
	}

	// Process message content (compression/encryption) if enabled
	processedMessage := message
	if r.compressionEnabled || r.encryptionKey != nil {
		processedMessage, err = r.processMessage(message)
		if err != nil {
			return fmt.Errorf("failed to process message: %w", err)
		}
	}

	// Implement retry logic for Redis operations
	var retryErr error
	for attempt := 0; attempt <= r.retryOptions.MaxRetries; attempt++ {
		if attempt > 0 {
			// Calculate backoff duration with exponential backoff
			backoffDuration := time.Duration(float64(r.retryOptions.RetryInterval) *
				math.Pow(r.retryOptions.BackoffFactor, float64(attempt-1)))
			time.Sleep(backoffDuration)
		}

		// Serialize message to JSON
		messageJSON, err := json.Marshal(processedMessage)
		if err != nil {
			return fmt.Errorf("failed to marshal message: %w", err)
		}

		// Add message to Redis list
		err = r.client.RPush(ctx, key, messageJSON).Err()
		if err == nil {
			// Set TTL on the key if not already set
			r.client.Expire(ctx, key, r.ttl)

			// Check if summarization is needed
			if r.summarizationEnabled {
				if err := r.checkAndSummarize(ctx); err != nil {
					// Log error but don't fail the message addition
					// TODO: Add proper logging
					_ = fmt.Sprintf("Failed to summarize messages: %v", err)
				}
			}

			return nil
		}

		retryErr = err
	}

	return fmt.Errorf("failed to add message to Redis after %d attempts: %w",
		r.retryOptions.MaxRetries, retryErr)
}

// GetMessages retrieves messages from the memory with improved filtering and pagination
func (r *RedisMemory) GetMessages(ctx context.Context, options ...contracts.GetMessagesOption) ([]contracts.Message, error) {
	// Get conversation ID from context
	conversationID, err := conversation.ConversationKey(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get conversation ID: %w", err)
	}

	// Get organization ID from context for multi-tenancy support
	orgID, err := multitenancy.GetOrgID(ctx)
	if err != nil {
		// If no organization ID is found, use a default
		orgID = "default"
	}

	// Create Redis key with org and conversation IDs
	key := fmt.Sprintf("%s%s:%s", r.keyPrefix, orgID, conversationID)

	// Apply options
	opts := &contracts.GetMessagesOptions{}
	for _, option := range options {
		option(opts)
	}

	var allMessages []contracts.Message

	// Get summaries first if summarization is enabled
	if r.summarizationEnabled {
		summaries, err := r.getSummaries(ctx)
		if err != nil {
			// Log error but continue without summaries
			_ = fmt.Sprintf("Failed to get summaries: %v", err)
		} else {
			allMessages = append(allMessages, summaries...)
		}
	}

	// Get all messages from Redis
	results, err := r.client.LRange(ctx, key, 0, -1).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get messages from Redis: %w", err)
	}

	// Parse messages
	for _, result := range results {
		var message contracts.Message
		if err := json.Unmarshal([]byte(result), &message); err != nil {
			return nil, fmt.Errorf("failed to unmarshal message: %w", err)
		}
		allMessages = append(allMessages, message)
	}

	// Filter by role if specified
	if len(opts.Roles) > 0 {
		var filtered []contracts.Message
		for _, msg := range allMessages {
			for _, role := range opts.Roles {
				if msg.Role == contracts.MessageRole(role) {
					filtered = append(filtered, msg)
					break
				}
			}
		}
		allMessages = filtered
	}

	// Apply limit if specified
	if opts.Limit > 0 && opts.Limit < len(allMessages) {
		allMessages = allMessages[len(allMessages)-opts.Limit:]
	}

	return allMessages, nil
}

// Clear clears the memory for a conversation
func (r *RedisMemory) Clear(ctx context.Context) error {
	// Get conversation ID from context
	conversationID, err := conversation.ConversationKey(ctx)
	if err != nil {
		return fmt.Errorf("failed to get conversation ID: %w", err)
	}

	// Get organization ID from context for multi-tenancy support
	orgID, err := multitenancy.GetOrgID(ctx)
	if err != nil {
		// If no organization ID is found, use a default
		orgID = "default"
	}

	// Create Redis key with org and conversation IDs
	key := fmt.Sprintf("%s%s:%s", r.keyPrefix, orgID, conversationID)

	// Delete the messages key from Redis
	err = r.client.Del(ctx, key).Err()
	if err != nil {
		return fmt.Errorf("failed to clear memory in Redis: %w", err)
	}

	// Clear summaries if summarization is enabled
	if r.summarizationEnabled {
		summaryKey := fmt.Sprintf("%s%s:%s", r.summaryKeyPrefix, orgID, conversationID)
		metaKey := fmt.Sprintf("%smeta:%s:%s", r.summaryKeyPrefix, orgID, conversationID)

		// Delete summary and metadata keys
		err = r.client.Del(ctx, summaryKey, metaKey).Err()
		if err != nil {
			return fmt.Errorf("failed to clear summaries in Redis: %w", err)
		}
	}

	return nil
}
