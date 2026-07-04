package agent

import (
	"context"
	"errors"
	"fmt"
	"strings"
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
	ToolDefs   []provider.ToolDefinition
	Emit       func(Event)
}

// Config is the provider identity used for future turns.
type Config struct {
	ProviderID string
	API        string
	Model      string
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

	mu      sync.Mutex
	busy    bool
	cancel  context.CancelFunc
	history []provider.Message
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
	runCtx, opts, history, err := a.start(ctx)
	if err != nil {
		return err
	}
	defer a.finish()

	state := &State{
		Provider:   opts.Provider,
		ProviderID: opts.ProviderID,
		API:        opts.API,
		Model:      opts.Model,
		Tools:      opts.Tools,
		ToolDefs:   append([]provider.ToolDefinition(nil), opts.ToolDefs...),
		Emit:       opts.Emit,
	}
	if err := runTurn(runCtx, state, TurnInput{Prompt: input.Text, History: history}); err != nil {
		return err
	}
	a.replaceHistory(state.messages)
	return nil
}

// Abort cancels the active provider stream.
func (a *Agent) Abort() {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.cancel != nil {
		a.cancel()
	}
}

// Reset clears remembered prompt history.
func (a *Agent) Reset() {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.history = nil
}

// Busy reports whether a prompt currently owns the agent.
func (a *Agent) Busy() bool {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.busy
}

// Config returns the provider labels used for future prompts.
func (a *Agent) Config() Config {
	a.mu.Lock()
	defer a.mu.Unlock()
	return Config{ProviderID: a.opts.ProviderID, API: a.opts.API, Model: a.opts.Model}
}

// SetModel updates provider labels for later prompts.
func (a *Agent) SetModel(providerID string, api string, model string) error {
	return a.SetProviderModel(nil, providerID, api, model)
}

// SetProviderModel updates the provider stream and labels for later prompts.
func (a *Agent) SetProviderModel(streamer provider.Streamer, providerID string, api string, model string) error {
	providerID = strings.TrimSpace(providerID)
	api = strings.TrimSpace(api)
	model = strings.TrimSpace(model)
	if providerID == "" || api == "" || model == "" {
		return fmt.Errorf("set model: provider, api, and model are required")
	}

	a.mu.Lock()
	defer a.mu.Unlock()
	if a.busy {
		return ErrBusy
	}
	if streamer != nil {
		a.opts.Provider = streamer
	}
	a.opts.ProviderID = providerID
	a.opts.API = api
	a.opts.Model = model
	return nil
}

func (a *Agent) start(ctx context.Context) (context.Context, Options, []provider.Message, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.busy {
		return nil, Options{}, nil, ErrBusy
	}
	if a.opts.Provider == nil {
		return nil, Options{}, nil, fmt.Errorf("agent prompt: missing provider")
	}
	ctx, cancel := context.WithCancel(ctx)
	a.busy = true
	a.cancel = cancel
	return ctx, a.opts, append([]provider.Message(nil), a.history...), nil
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

func (a *Agent) replaceHistory(messages []provider.Message) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.history = append(a.history[:0], messages...)
}
