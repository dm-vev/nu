package builder

import (
	"context"
	"fmt"
	"net/url"
	"time"

	"github.com/dm-vev/nu/contracts"
	"github.com/dm-vev/nu/internal/mcp/config"
	"github.com/dm-vev/nu/internal/mcp/preset"
	"github.com/dm-vev/nu/telemetry"
)

type LazyMCPServerConfig = config.Config

// Builder provides a fluent interface for creating MCP server configurations
type Builder struct {
	servers      []contracts.MCPServer
	lazyConfigs  []LazyMCPServerConfig
	logger       telemetry.Logger
	retryOptions *RetryOptions
	healthCheck  bool
	timeout      time.Duration
	errors       []error
}

type RetryOptions struct {
	MaxAttempts       int
	InitialDelay      time.Duration
	MaxDelay          time.Duration
	BackoffMultiplier float64
}

func NewBuilder() *Builder {
	return &Builder{
		logger: telemetry.NewLogger(),
		retryOptions: &RetryOptions{
			MaxAttempts: 5, InitialDelay: time.Second,
			MaxDelay: 30 * time.Second, BackoffMultiplier: 2.0,
		},
		timeout: 30 * time.Second, healthCheck: true,
	}
}

func (b *Builder) WithLogger(logger telemetry.Logger) *Builder {
	b.logger = logger
	return b
}

func (b *Builder) WithRetry(maxAttempts int, initialDelay time.Duration) *Builder {
	b.retryOptions.MaxAttempts = maxAttempts
	b.retryOptions.InitialDelay = initialDelay
	return b
}

func (b *Builder) WithTimeout(timeout time.Duration) *Builder {
	b.timeout = timeout
	return b
}

func (b *Builder) WithHealthCheck(enabled bool) *Builder {
	b.healthCheck = enabled
	return b
}

func (b *Builder) AddServer(urlStr string) *Builder {
	server, config, err := b.parseServerURL(urlStr)
	if err != nil {
		b.errors = append(b.errors, fmt.Errorf("failed to parse server URL %q: %w", urlStr, err))
		return b
	}
	if server != nil {
		b.servers = append(b.servers, server)
	} else if config != nil {
		b.lazyConfigs = append(b.lazyConfigs, *config)
	}
	return b
}

func (b *Builder) AddStdioServer(name, command string, args ...string) *Builder {
	b.lazyConfigs = append(b.lazyConfigs, LazyMCPServerConfig{
		Name: name, Type: "stdio", Command: command, Args: args,
	})
	return b
}

func (b *Builder) AddHTTPServer(name, baseURL string) *Builder {
	b.lazyConfigs = append(b.lazyConfigs, LazyMCPServerConfig{Name: name, Type: "http", URL: baseURL})
	return b
}

func (b *Builder) AddHTTPServerWithAuth(name, baseURL, token string) *Builder {
	u, err := url.Parse(baseURL)
	if err != nil {
		b.errors = append(b.errors, fmt.Errorf("invalid URL %q: %w", baseURL, err))
		return b
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		b.errors = append(b.errors, fmt.Errorf("invalid URL scheme for HTTP server %q: expected http or https, got %q", baseURL, u.Scheme))
		return b
	}
	q := u.Query()
	q.Set("token", token)
	u.RawQuery = q.Encode()
	b.lazyConfigs = append(b.lazyConfigs, LazyMCPServerConfig{Name: name, Type: "http", URL: u.String()})
	return b
}

func (b *Builder) AddPreset(presetName string) *Builder {
	preset, err := preset.GetPreset(presetName)
	if err != nil {
		b.errors = append(b.errors, err)
		return b
	}
	switch preset.Type {
	case "stdio":
		b.AddStdioServer(preset.Name, preset.Command, preset.Args...)
	case "http":
		b.AddHTTPServer(preset.Name, preset.URL)
	}
	return b
}

func (b *Builder) Build(ctx context.Context) ([]contracts.MCPServer, []LazyMCPServerConfig, error) {
	if len(b.errors) > 0 {
		return nil, nil, fmt.Errorf("builder errors: %v", b.errors)
	}
	if b.healthCheck {
		for _, config := range b.lazyConfigs {
			if b.shouldInitializeEagerly(config) {
				server, err := b.Connect(ctx, config)
				if err != nil {
					b.logger.Warn(ctx, "Failed to initialize MCP server", map[string]interface{}{
						"server_name": config.Name, "error": err.Error(),
					})
					continue
				}
				b.servers = append(b.servers, server)
			}
		}
	}
	return b.servers, b.lazyConfigs, nil
}

func (b *Builder) BuildLazy() ([]LazyMCPServerConfig, error) {
	if len(b.errors) > 0 {
		return nil, fmt.Errorf("builder errors: %v", b.errors)
	}
	return b.lazyConfigs, nil
}
