package agentui

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"

	"github.com/dm-vev/nu/contracts"
	"github.com/dm-vev/nu/internal/memory/conversation"
	"github.com/dm-vev/nu/internal/multitenancy"
)

var ErrBusy = errors.New("agent busy")

type Config struct {
	ProviderID string
	API        string
	Model      string
}

type Builder func(context.Context, Config, contracts.Memory) (contracts.StreamingAgent, error)

type Options struct {
	Runner  contracts.StreamingAgent
	Builder Builder
	Memory  contracts.Memory
	Config  Config
	Emit    func(Event)
}

type Event struct {
	Type string `json:"type"`
	Data any    `json:"data,omitempty"`
}

type Prompt struct{ Text string }

type Controller struct {
	mu      sync.Mutex
	runner  contracts.StreamingAgent
	builder Builder
	memory  contracts.Memory
	emit    func(Event)
	config  Config
	busy    bool
	cancel  context.CancelFunc
}

// Agent keeps the existing TUI/RPC controller name while the backend lives in the public agent package.
type Agent = Controller

func New(opts Options) *Controller {
	return &Controller{runner: opts.Runner, builder: opts.Builder, memory: opts.Memory, emit: opts.Emit, config: opts.Config}
}

func (c *Controller) Prompt(ctx context.Context, input Prompt) error {
	runCtx, runner, err := c.start(ctx)
	if err != nil {
		return err
	}
	defer c.finish()
	return consumeStream(conversationContext(runCtx), runner, input.Text, c.emit)
}

func (c *Controller) Abort() {
	c.mu.Lock()
	cancel := c.cancel
	c.mu.Unlock()
	if cancel != nil {
		cancel()
	}
}

func (c *Controller) Reset() {
	if c.memory != nil {
		_ = c.memory.Clear(conversationContext(context.Background()))
	}
}

func (c *Controller) Busy() bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.busy
}

func (c *Controller) Config() Config {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.config
}

func (c *Controller) SetModel(providerID, api, model string) error {
	config := Config{ProviderID: strings.TrimSpace(providerID), API: strings.TrimSpace(api), Model: strings.TrimSpace(model)}
	if config.ProviderID == "" || config.API == "" || config.Model == "" {
		return fmt.Errorf("set model: provider, api, and model are required")
	}
	c.mu.Lock()
	if c.busy {
		c.mu.Unlock()
		return ErrBusy
	}
	builder, memoryStore := c.builder, c.memory
	c.mu.Unlock()
	if builder == nil {
		return fmt.Errorf("set model: SDK builder is unavailable")
	}
	runner, err := builder(context.Background(), config, memoryStore)
	if err != nil {
		return fmt.Errorf("set model: %w", err)
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.busy {
		return ErrBusy
	}
	c.runner, c.config = runner, config
	return nil
}

func (c *Controller) start(ctx context.Context) (context.Context, contracts.StreamingAgent, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.busy {
		return nil, nil, ErrBusy
	}
	if c.runner == nil {
		return nil, nil, fmt.Errorf("agent prompt: missing SDK runner")
	}
	runCtx, cancel := context.WithCancel(ctx)
	c.busy, c.cancel = true, cancel
	return runCtx, c.runner, nil
}

func (c *Controller) finish() {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.cancel != nil {
		c.cancel()
	}
	c.cancel, c.busy = nil, false
}

func conversationContext(ctx context.Context) context.Context {
	ctx = multitenancy.WithOrgID(ctx, "nu")
	return conversation.WithConversationID(ctx, "default")
}
