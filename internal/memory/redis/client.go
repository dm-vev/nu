package redis

import (
	"context"
	"fmt"
	"time"

	redisclient "github.com/go-redis/redis/v8"

	"github.com/dm-vev/nu/contracts"
)

// RedisMemory implements a Redis-backed memory store
type RedisMemory struct {
	client             *redisclient.Client
	ttl                time.Duration
	keyPrefix          string
	compressionEnabled bool
	encryptionKey      []byte
	maxMessageSize     int
	retryOptions       *RetryOptions

	// Summarization fields
	summarizationEnabled bool
	llmClient            contracts.LLM
	messageThreshold     int
	summaryCount         int
	summaryKeyPrefix     string
}

// RedisConfig contains configuration for Redis
type RedisConfig struct {
	// URL is the Redis URL (e.g., "localhost:6379")
	URL string

	// Password is the Redis password
	Password string

	// DB is the Redis database number
	DB int
}

// NewRedisMemory creates a new Redis-backed memory store
func NewRedisMemory(client *redisclient.Client, options ...RedisOption) *RedisMemory {
	memory := &RedisMemory{
		client:             client,
		ttl:                24 * time.Hour,  // Default TTL
		keyPrefix:          "agent:memory:", // Default prefix
		compressionEnabled: false,
		maxMessageSize:     1024 * 1024, // 1MB default max size
		retryOptions: &RetryOptions{
			MaxRetries:    3,
			RetryInterval: 100 * time.Millisecond,
			BackoffFactor: 2.0,
		},
		// Summarization defaults
		summarizationEnabled: false,
		messageThreshold:     50,
		summaryCount:         5,
		summaryKeyPrefix:     "agent:memory:summary:",
	}

	for _, option := range options {
		option(memory)
	}

	// Update summary key prefix if keyPrefix was changed by options
	if memory.summarizationEnabled && memory.summaryKeyPrefix == "agent:memory:summary:" {
		memory.summaryKeyPrefix = memory.keyPrefix + "summary:"
	}

	return memory
}

// NewRedisMemoryFromConfig creates a new Redis memory from configuration
func NewRedisMemoryFromConfig(config RedisConfig, options ...RedisOption) (*RedisMemory, error) {
	// Create Redis client
	client := redisclient.NewClient(&redisclient.Options{
		Addr:     config.URL,
		Password: config.Password,
		DB:       config.DB,
	})

	// Test connection
	ctx := context.Background()
	if _, err := client.Ping(ctx).Result(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	// Create Redis memory
	return NewRedisMemory(client, options...), nil
}

// Close closes the underlying Redis connection
func (r *RedisMemory) Close() error {
	if r.client != nil {
		return r.client.Close()
	}
	return nil
}
