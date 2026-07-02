package agent

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"nu/internal/provider"
)

// ErrBusy is returned when a prompt is already running.
var ErrBusy = errors.New("agent busy")

// Options configures an Agent.
type Options struct {
	Provider   provider.Streamer
	ProviderID string
	API        string
	Model      string
	Tools      map[string]ToolFunc
	Emit       func(Event)
}

// Event is one agent event emitted to app/RPC boundaries.
type Event struct {
	Type string `json:"type"`
	Data any    `json:"data,omitempty"`
}

// Prompt is one user prompt.
type Prompt struct {
	Text string
}

// ToolCall is one finalized provider tool request.
type ToolCall struct {
	ID        string
	Name      string
	Arguments string
}

// ToolResult is one tool result fed back to the provider.
type ToolResult struct {
	Content string
}

// ToolFunc executes one tool call.
type ToolFunc func(context.Context, ToolCall) (ToolResult, error)

// Agent owns prompt execution state.
type Agent struct {
	opts Options

	mu     sync.Mutex
	busy   bool
	cancel context.CancelFunc
}

// New constructs an idle agent.
func New(opts Options) *Agent {
	if opts.ProviderID == "" {
		opts.ProviderID = "test"
	}
	if opts.API == "" {
		opts.API = "test"
	}
	if opts.Model == "" {
		opts.Model = "test"
	}
	if len(opts.Tools) > 0 {
		tools := make(map[string]ToolFunc, len(opts.Tools))
		for name, tool := range opts.Tools {
			tools[name] = tool
		}
		opts.Tools = tools
	}
	return &Agent{opts: opts}
}

// Prompt sends one prompt to the provider.
func (a *Agent) Prompt(ctx context.Context, input Prompt) error {
	runCtx, err := a.start(ctx)
	if err != nil {
		return err
	}
	defer a.finish()

	state := &State{
		Provider:   a.opts.Provider,
		ProviderID: a.opts.ProviderID,
		API:        a.opts.API,
		Model:      a.opts.Model,
		Tools:      a.opts.Tools,
		Emit:       a.opts.Emit,
	}
	return runTurn(runCtx, state, TurnInput{Prompt: input.Text})
}

// Abort cancels the active provider stream.
func (a *Agent) Abort() {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.cancel != nil {
		a.cancel()
	}
}

func (a *Agent) start(ctx context.Context) (context.Context, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.busy {
		return nil, ErrBusy
	}
	if a.opts.Provider == nil {
		return nil, fmt.Errorf("agent prompt: missing provider")
	}
	ctx, cancel := context.WithCancel(ctx)
	a.busy = true
	a.cancel = cancel
	return ctx, nil
}

func (a *Agent) finish() {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.cancel != nil {
		a.cancel()
	}
	a.cancel = nil
	a.busy = false
}
