package redis

import (
	"time"

	"nu/internal/contracts"
)

// RetryOptions configures retry behavior for Redis operations
type RetryOptions struct {
	MaxRetries    int
	RetryInterval time.Duration
	BackoffFactor float64
}

// RedisOption represents an option for configuring the Redis memory
type RedisOption func(*RedisMemory)

// WithTTL sets the TTL for Redis keys
func WithTTL(ttl time.Duration) RedisOption {
	return func(r *RedisMemory) {
		r.ttl = ttl
	}
}

// WithKeyPrefix sets a custom prefix for Redis keys
func WithKeyPrefix(prefix string) RedisOption {
	return func(r *RedisMemory) {
		r.keyPrefix = prefix
	}
}

// WithCompression enables compression for stored messages
func WithCompression(enabled bool) RedisOption {
	return func(r *RedisMemory) {
		r.compressionEnabled = enabled
	}
}

// WithEncryption enables encryption for stored messages
func WithEncryption(key []byte) RedisOption {
	return func(r *RedisMemory) {
		r.encryptionKey = key
	}
}

// WithMaxMessageSize sets the maximum size for stored messages
func WithMaxMessageSize(size int) RedisOption {
	return func(r *RedisMemory) {
		r.maxMessageSize = size
	}
}

// WithRetryOptions configures retry behavior for Redis operations
func WithRetryOptions(options *RetryOptions) RedisOption {
	return func(r *RedisMemory) {
		r.retryOptions = options
	}
}

// WithSummarization enables automatic summarization of old messages
func WithSummarization(llm contracts.LLM, messageThreshold int, summaryCount int) RedisOption {
	return func(r *RedisMemory) {
		r.summarizationEnabled = true
		r.llmClient = llm
		r.messageThreshold = messageThreshold
		r.summaryCount = summaryCount
		r.summaryKeyPrefix = r.keyPrefix + "summary:"
	}
}
