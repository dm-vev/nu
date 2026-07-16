package grpc

import (
	"context"
	"fmt"
	"net"
	"sync"
	"time"

	"nu/internal/agent"
	"nu/internal/contracts"
)

// Microservice represents a microservice wrapping an agent.
type Microservice struct {
	agent      *agent.Agent
	server     *AgentServer
	port       int
	running    bool
	serving    bool // New field to track if gRPC server is actually serving
	mu         sync.RWMutex
	cancelFunc context.CancelFunc
	servingCh  chan struct{} // Channel to signal when server starts serving

	// Event handlers
	thinkingHandlers   []func(string)
	contentHandlers    []func(string)
	toolCallHandlers   []func(*contracts.ToolCallEvent)
	toolResultHandlers []func(*contracts.ToolCallEvent)
	errorHandlers      []func(error)
	completeHandlers   []func()
	handlersMu         sync.RWMutex
}

// MicroserviceConfig configures an agent microservice.
type MicroserviceConfig struct {
	Port    int           // Port to run the service on (0 for auto-assign)
	Timeout time.Duration // Request timeout
}

// NewMicroservice creates an agent microservice.
func NewMicroservice(agent *agent.Agent, config MicroserviceConfig) (*Microservice, error) {
	if agent == nil {
		return nil, fmt.Errorf("agent cannot be nil")
	}

	if agent.IsRemote() {
		return nil, fmt.Errorf("cannot create microservice from remote agent")
	}

	if config.Port < 0 || config.Port > 65535 {
		return nil, fmt.Errorf("invalid port: %d", config.Port)
	}

	server := NewAgentServer(agent)

	return &Microservice{
		agent:     agent,
		server:    server,
		port:      config.Port,
		servingCh: make(chan struct{}),
	}, nil
}

// Start starts the microservice
func (m *Microservice) Start() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.running {
		return fmt.Errorf("microservice is already running")
	}

	// Create a listener first to get the actual port
	addr := fmt.Sprintf(":%d", m.port)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("failed to listen on port %d: %w", m.port, err)
	}

	// Update port if it was auto-assigned (port 0)
	if m.port == 0 {
		m.port = listener.Addr().(*net.TCPAddr).Port
	}

	// Create a context for the server
	_, cancel := context.WithCancel(context.Background())
	m.cancelFunc = cancel

	// Mark as running now that we have successfully bound to the port
	m.running = true

	// Start the server in a goroutine
	go func() {
		defer func() {
			m.mu.Lock()
			m.running = false
			m.serving = false
			m.mu.Unlock()
		}()

		// Signal that we're about to start serving
		m.mu.Lock()
		m.serving = true
		close(m.servingCh) // Signal that server is starting to serve
		m.mu.Unlock()

		err := m.server.StartWithListener(listener)
		if err != nil {
			fmt.Printf("Agent server error: %v\n", err)
		}
	}()

	fmt.Printf("Agent microservice '%s' started on port %d\n", m.agent.GetName(), m.port)
	return nil
}

// Stop stops the microservice
func (m *Microservice) Stop() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.running {
		return nil // Already stopped
	}

	// Stop the gRPC server
	m.server.Stop()

	// Cancel the context
	if m.cancelFunc != nil {
		m.cancelFunc()
	}

	m.running = false
	fmt.Printf("Agent microservice '%s' stopped\n", m.agent.GetName())
	return nil
}

// IsRunning returns true if the microservice is currently running
func (m *Microservice) IsRunning() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.running
}

// GetPort returns the port the microservice is running on
func (m *Microservice) GetPort() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.port
}

// GetURL returns the URL of the microservice
func (m *Microservice) GetURL() string {
	return fmt.Sprintf("localhost:%d", m.GetPort())
}

// GetAgent returns the underlying agent
func (m *Microservice) GetAgent() *agent.Agent {
	return m.agent
}

// WaitForReady waits for the microservice to be ready to serve requests
func (m *Microservice) WaitForReady(timeout time.Duration) error {
	deadline := time.Now().Add(timeout)

	// First, wait for the service to be marked as running
	for time.Now().Before(deadline) {
		if m.IsRunning() {
			break
		}
		time.Sleep(50 * time.Millisecond)
	}

	// If still not running after timeout, return error
	if !m.IsRunning() {
		return fmt.Errorf("microservice failed to start within %v", timeout)
	}

	// Wait for the server to start serving with timeout
	servingTimeout := time.Until(deadline)
	select {
	case <-m.servingCh:
		fmt.Printf("Debug: Server started serving on port %d\n", m.port)
		// Server has started serving, give it a moment to initialize
		time.Sleep(100 * time.Millisecond)
	case <-time.After(servingTimeout):
		return fmt.Errorf("microservice failed to start serving within %v", timeout)
	}

	// Now test gRPC health endpoint
	for time.Now().Before(deadline) {
		if err := m.testGRPCHealth(); err == nil {
			return nil // gRPC health check passed
		}
		time.Sleep(100 * time.Millisecond)
	}

	return fmt.Errorf("microservice not ready after %v", timeout)
}
