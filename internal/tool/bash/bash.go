package bash

import (
	"context"
	"errors"
	"strings"
	"time"

	"nu/internal/agent"
	"nu/internal/tool/toolkit"
)

// Run runs one shell command in cwd.
func Run(ctx context.Context, cwd string, raw string, maxBytes int) (agent.ToolResult, error) {
	var args struct {
		Command   string `json:"command"`
		TimeoutMS int    `json:"timeout_ms"`
	}
	if err := toolkit.DecodeArgs(raw, &args); err != nil {
		return agent.ToolResult{}, err
	}
	if strings.TrimSpace(args.Command) == "" {
		return agent.ToolResult{}, errors.New("bash: missing command")
	}
	runCtx := ctx
	cancel := func() {}
	if args.TimeoutMS > 0 {
		runCtx, cancel = context.WithTimeout(ctx, time.Duration(args.TimeoutMS)*time.Millisecond)
	}
	defer cancel()

	stdout, stderr, exitCode, timedOut := runCommand(runCtx, cwd, args.Command)
	fullOutput := stdout + stderr
	output, truncated := toolkit.TruncateString(fullOutput, maxBytes)
	result := map[string]any{
		"stdout":    stdout,
		"stderr":    stderr,
		"exit_code": exitCode,
		"timed_out": timedOut,
		"output":    output,
		"truncated": truncated,
	}
	if truncated {
		path, err := toolkit.PersistTempOutput(fullOutput)
		if err != nil {
			return agent.ToolResult{}, err
		}
		result["full_output_path"] = path
	}
	return toolkit.JSONResult(result)
}
