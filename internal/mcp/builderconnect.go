package mcp

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"time"

	"nu/internal/contracts"
)

func (b *Builder) parseServerURL(urlStr string) (contracts.MCPServer, *LazyMCPServerConfig, error) {
	u, err := url.Parse(urlStr)
	if err != nil {
		return nil, nil, err
	}
	switch u.Scheme {
	case "stdio":
		parts := strings.Split(strings.TrimPrefix(u.Path, "/"), "/")
		if len(parts) < 2 {
			return nil, nil, fmt.Errorf("invalid stdio URL format")
		}
		name := u.Host
		if name == "" {
			name = parts[0]
			parts = parts[1:]
		}
		command := "/" + strings.Join(parts, "/")
		args := []string{}
		for key, values := range u.Query() {
			for _, value := range values {
				if value != "" {
					args = append(args, fmt.Sprintf("--%s=%s", key, value))
				} else {
					args = append(args, fmt.Sprintf("--%s", key))
				}
			}
		}
		return nil, &LazyMCPServerConfig{
			Name: name, Type: "stdio", Command: command, Args: args,
		}, nil
	case "http", "https":
		name := u.Host
		q := u.Query()
		token := q.Get("token")
		transport := strings.ToLower(strings.TrimSpace(q.Get("transport")))
		q.Del("transport")
		q.Del("token")
		u.RawQuery = q.Encode()
		config := &LazyMCPServerConfig{
			Name: name, Type: "http", URL: u.String(), Token: token,
			HttpTransportMode: transport,
		}
		if transport != "sse" && transport != "streamable" {
			config.HttpTransportMode = ""
		}
		return nil, config, nil
	case "mcp":
		preset, err := GetPreset(u.Host + u.Path)
		if err != nil {
			return nil, nil, err
		}
		return nil, &preset, nil
	default:
		return nil, nil, fmt.Errorf("unsupported URL scheme: %s", u.Scheme)
	}
}

func (b *Builder) initializeServer(ctx context.Context, config LazyMCPServerConfig) (contracts.MCPServer, error) {
	ctx, cancel := context.WithTimeout(ctx, b.timeout)
	defer cancel()
	var server contracts.MCPServer
	var err error
	retryConfig := &RetryConfig{
		MaxAttempts: b.retryOptions.MaxAttempts, InitialDelay: b.retryOptions.InitialDelay,
		MaxDelay: b.retryOptions.MaxDelay, BackoffMultiplier: b.retryOptions.BackoffMultiplier,
	}
	switch config.Type {
	case "stdio":
		server, err = NewStdioServerWithRetry(ctx, StdioServerConfig{
			Command: config.Command, Args: config.Args, Env: config.Env, Logger: b.logger,
		}, retryConfig)
	case "http":
		server, err = NewHTTPWithRetry(ctx, HTTPConfig{
			BaseURL: config.URL, Token: config.Token, Logger: b.logger,
			ProtocolType: ServerProtocolType(config.HttpTransportMode),
		}, retryConfig)
	default:
		return nil, fmt.Errorf("unsupported server type: %s", config.Type)
	}
	if err != nil && b.retryOptions.MaxAttempts > 1 {
		server, err = b.retryConnection(ctx, config)
	}
	return server, err
}

func (b *Builder) retryConnection(ctx context.Context, config LazyMCPServerConfig) (contracts.MCPServer, error) {
	delay := b.retryOptions.InitialDelay
	for attempt := 1; attempt <= b.retryOptions.MaxAttempts; attempt++ {
		b.logger.Debug(ctx, "Retrying MCP connection", map[string]interface{}{
			"server_name": config.Name, "attempt": attempt, "max_attempts": b.retryOptions.MaxAttempts,
		})
		server, err := b.initializeServer(ctx, config)
		if err == nil {
			return server, nil
		}
		if attempt < b.retryOptions.MaxAttempts {
			time.Sleep(delay)
			delay = time.Duration(float64(delay) * b.retryOptions.BackoffMultiplier)
			if delay > b.retryOptions.MaxDelay {
				delay = b.retryOptions.MaxDelay
			}
		}
	}
	return nil, fmt.Errorf("failed to connect after %d attempts", b.retryOptions.MaxAttempts)
}

func (b *Builder) shouldInitializeEagerly(config LazyMCPServerConfig) bool {
	return false
}
