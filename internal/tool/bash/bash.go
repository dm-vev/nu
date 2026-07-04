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
	if usesInteractiveSudo(args.Command) {
		result := map[string]any{
			"stdout":    "",
			"stderr":    "sudo is disabled in interactive mode; use sudo -n for cached credentials or sudo -S with explicit stdin.\n",
			"exit_code": 1,
			"timed_out": false,
			"output":    "sudo is disabled in interactive mode; use sudo -n for cached credentials or sudo -S with explicit stdin.\n",
			"truncated": false,
		}
		return toolkit.JSONResult(result)
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

func usesInteractiveSudo(command string) bool {
	fields := strings.Fields(command)
	for index, field := range fields {
		if strings.Trim(field, ";|&()") != "sudo" {
			continue
		}
		for _, next := range fields[index+1:] {
			next = strings.Trim(next, ";|&()")
			if next == "" {
				continue
			}
			if sudoNonInteractiveFlag(next) {
				return false
			}
			if !strings.HasPrefix(next, "-") {
				return true
			}
		}
		return true
	}
	return false
}

func sudoNonInteractiveFlag(value string) bool {
	if value == "--non-interactive" {
		return true
	}
	if !strings.HasPrefix(value, "-") || strings.HasPrefix(value, "--") {
		return false
	}
	return strings.Contains(value, "n") || strings.Contains(value, "S")
}
