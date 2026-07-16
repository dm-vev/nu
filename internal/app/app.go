package app

import (
	"context"
	"fmt"

	"github.com/dm-vev/nu/internal/app/cli"
)

const (
	exitOK    = 0
	exitError = 1
	exitUsage = 2
)

// Run executes one Nu invocation and returns a process exit code.
func Run(ctx context.Context, opts Options) int {
	req, diagnostics := cli.Parse(opts.Args)
	opts = normalizeOptions(opts)

	for _, diagnostic := range diagnostics {
		fmt.Fprintln(opts.Stderr, diagnostic.Message)
	}
	if len(diagnostics) > 0 {
		return exitUsage
	}

	rt, err := NewRuntime(ctx, opts)
	if err != nil {
		fmt.Fprintln(opts.Stderr, err)
		return exitError
	}

	if err := runMode(ctx, rt, req); err != nil {
		fmt.Fprintln(opts.Stderr, err)
		return exitError
	}
	return exitOK
}

// NewRuntime constructs runtime dependencies without starting long-lived work.
func NewRuntime(ctx context.Context, opts Options) (*Runtime, error) {
	select {
	case <-ctx.Done():
		return nil, fmt.Errorf("construct runtime: %w", ctx.Err())
	default:
	}

	opts = normalizeOptions(opts)
	return &Runtime{Options: opts}, nil
}
