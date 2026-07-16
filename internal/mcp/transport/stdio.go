package transport

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/dm-vev/nu/contracts"
	"github.com/dm-vev/nu/internal/mcp/retry"
	"github.com/dm-vev/nu/telemetry"
)

// syncBuffer guards stderr because os/exec writes it from a background goroutine.
type syncBuffer struct {
	mu  sync.Mutex
	buf bytes.Buffer
}

func (sb *syncBuffer) Write(p []byte) (int, error) {
	sb.mu.Lock()
	defer sb.mu.Unlock()
	return sb.buf.Write(p)
}

func (sb *syncBuffer) String() string {
	sb.mu.Lock()
	defer sb.mu.Unlock()
	return sb.buf.String()
}

// StdioServerConfig holds configuration for a stdio MCP server
type StdioServerConfig struct {
	Command string
	Args    []string
	Env     []string
	Logger  telemetry.Logger
}

// NewStdioServer creates a new MCPServer that communicates over stdio using the official SDK
func NewStdioServer(ctx context.Context, config StdioServerConfig) (contracts.MCPServer, error) {
	return NewStdioServerWithRetry(ctx, config, nil)
}

// NewStdioServerWithRetry creates a new MCPServer with retry logic
func NewStdioServerWithRetry(ctx context.Context, config StdioServerConfig, retryConfig *retry.Config) (contracts.MCPServer, error) {
	if config.Command == "" {
		return nil, fmt.Errorf("command cannot be empty")
	}
	commandPath, err := exec.LookPath(config.Command)
	if err != nil {
		return nil, fmt.Errorf("invalid command %q: %v", config.Command, err)
	}
	if !filepath.IsAbs(commandPath) {
		return nil, fmt.Errorf("command path must be absolute for security: %q", commandPath)
	}
	if info, err := os.Stat(commandPath); err != nil {
		return nil, fmt.Errorf("command not accessible: %v", err)
	} else if info.IsDir() {
		return nil, fmt.Errorf("command path is a directory, not executable: %q", commandPath)
	}

	config.Logger.Debug(ctx, "Creating MCP server command", map[string]interface{}{
		"command": commandPath, "args": config.Args, "env_provided": len(config.Env),
	})
	if len(config.Env) > 0 {
		config.Logger.Debug(ctx, "MCP server environment variables (from config)", map[string]interface{}{
			"count": len(config.Env),
		})
		for i, envVar := range config.Env {
			parts := strings.SplitN(envVar, "=", 2)
			if len(parts) == 2 {
				key, value := parts[0], parts[1]
				if strings.Contains(strings.ToLower(key), "key") ||
					strings.Contains(strings.ToLower(key), "secret") ||
					strings.Contains(strings.ToLower(key), "password") ||
					strings.Contains(strings.ToLower(key), "token") {
					if len(value) > 8 {
						value = fmt.Sprintf("%s...%s (length: %d)", value[:4], value[len(value)-4:], len(value))
					} else {
						value = "***MASKED***"
					}
				}
				config.Logger.Debug(ctx, fmt.Sprintf("MCP env[%d]", i), map[string]interface{}{
					"key": key, "value": value,
				})
			} else {
				config.Logger.Debug(ctx, fmt.Sprintf("MCP env[%d]", i), map[string]interface{}{"raw": envVar})
			}
		}
	}

	// #nosec G204 -- commandPath is validated above with LookPath and security checks
	cmd := exec.CommandContext(ctx, commandPath, config.Args...)
	if len(config.Env) > 0 {
		cmd.Env = append(os.Environ(), config.Env...)
		config.Logger.Info(ctx, "[STDIO SERVER] Creating subprocess with command", map[string]interface{}{
			"command": commandPath, "args": config.Args, "env_count": len(config.Env),
		})
		for i, envVar := range config.Env {
			if len(envVar) > 60 {
				config.Logger.Debug(ctx, fmt.Sprintf("[STDIO SERVER ENV %d]", i), map[string]interface{}{
					"env": envVar[:60] + "...",
				})
			} else {
				config.Logger.Debug(ctx, fmt.Sprintf("[STDIO SERVER ENV %d]", i), map[string]interface{}{"env": envVar})
			}
		}
	}

	stderrBuf := &syncBuffer{}
	cmd.Stderr = stderrBuf
	transport := &mcp.CommandTransport{Command: cmd}
	server, mcpErr := newServerFromTransport(ctx, transport, "stdio-server", "stdio", nil, config.Logger)
	if mcpErr != nil {
		config.Logger.Error(ctx, "[STDIO SERVER ERROR] Failed to connect to MCP server", map[string]interface{}{
			"error": mcpErr.Error(), "error_type": mcpErr.ErrorType,
			"retryable": mcpErr.Retryable, "command": config.Command,
			"args": config.Args, "stderr": stderrBuf.String(),
		})
		return nil, mcpErr
	}
	return server, nil
}
