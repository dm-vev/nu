package client

import (
	"time"

	"github.com/a2aproject/a2a-go/a2a"

	"github.com/dm-vev/nu/telemetry"
)

// Option configures an A2A client.
type Option func(*Client)

// WithLogger sets a logger for the A2A client.
func WithLogger(logger telemetry.Logger) Option {
	return func(c *Client) {
		c.logger = logger
	}
}

// WithTimeout sets the HTTP client timeout for the A2A client.
func WithTimeout(d time.Duration) Option {
	return func(c *Client) {
		c.timeout = d
	}
}

// WithBearerToken sets a static bearer token for authentication on the A2A client.
func WithBearerToken(token string) Option {
	return func(c *Client) {
		c.bearerToken = token
	}
}

// SendOption configures an individual send operation.
type SendOption func(*sendConfig)

type sendConfig struct {
	contextID string
	taskID    a2a.TaskID
}

// WithContextID sets the context ID for a multi-turn conversation.
func WithContextID(id string) SendOption {
	return func(c *sendConfig) {
		c.contextID = id
	}
}

// WithTaskID continues an existing task by referencing its ID.
func WithTaskID(id a2a.TaskID) SendOption {
	return func(c *sendConfig) {
		c.taskID = id
	}
}
