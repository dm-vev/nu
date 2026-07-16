package grpc

import (
	"context"
	"sync"
	"time"

	"google.golang.org/grpc"

	"nu/internal/contracts"
	pb "nu/internal/transport/grpc/pb"
)

// RemoteAgentClient handles communication with remote agents via gRPC.
type RemoteAgentClient struct {
	url        string
	conn       *grpc.ClientConn
	client     pb.AgentServiceClient
	timeout    time.Duration
	retryCount int

	// Event handlers
	thinkingHandlers   []func(string)
	contentHandlers    []func(string)
	toolCallHandlers   []func(*contracts.ToolCallEvent)
	toolResultHandlers []func(*contracts.ToolCallEvent)
	errorHandlers      []func(error)
	completeHandlers   []func()
	handlersMu         sync.RWMutex
}

// RemoteAgentConfig configures the remote agent client.
type RemoteAgentConfig struct {
	URL        string
	Timeout    time.Duration
	RetryCount int
}

// NewRemoteAgentClient creates a remote agent client.
func NewRemoteAgentClient(config RemoteAgentConfig) *RemoteAgentClient {
	// Only set default timeout if no timeout was specified at all
	// If timeout is explicitly set to 0, keep it as 0 for infinite timeout
	timeout := config.Timeout
	if timeout == 0 && !isTimeoutExplicitlySet(config) {
		timeout = 30 * time.Minute // 30 minutes for long-running agents
	}

	if config.RetryCount == 0 {
		config.RetryCount = 3
	}

	return &RemoteAgentClient{
		url:        config.URL,
		timeout:    timeout,
		retryCount: config.RetryCount,
	}
}

// isTimeoutExplicitlySet checks if timeout was explicitly set in config
// For now, we'll assume any 0 value means infinite timeout
func isTimeoutExplicitlySet(config RemoteAgentConfig) bool {
	// We'll treat any 0 timeout as explicitly set for infinite timeout
	return true
}

// withTimeoutIfSet adds timeout to context if timeout > 0, otherwise returns context as-is (infinite timeout)
func (r *RemoteAgentClient) withTimeoutIfSet(ctx context.Context) (context.Context, context.CancelFunc) {
	if r.timeout > 0 {
		return context.WithTimeout(ctx, r.timeout) // #nosec G118 - cancel func is returned to caller
	}
	// For 0 timeout (infinite), return context as-is with a no-op cancel function
	return ctx, func() {}
}

// SetTimeout sets the timeout for requests
func (r *RemoteAgentClient) SetTimeout(timeout time.Duration) {
	r.timeout = timeout
}

// SetRetryCount sets the number of retries for failed requests
func (r *RemoteAgentClient) SetRetryCount(count int) {
	r.retryCount = count
}
