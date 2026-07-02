package bash

import (
	"bytes"
	"context"
	"errors"
	"os"
	"os/exec"
	"strings"
	"syscall"
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

	cmd := exec.CommandContext(runCtx, "sh", "-c", args.Command)
	cmd.Dir = cwd
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	cmd.Cancel = func() error {
		if cmd.Process == nil {
			return os.ErrProcessDone
		}
		return syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)
	}
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	timedOut := errors.Is(runCtx.Err(), context.DeadlineExceeded)
	exitCode := 0
	if err != nil {
		exitCode = 1
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			exitCode = exitErr.ExitCode()
		}
		if timedOut {
			exitCode = -1
		}
	}

	fullOutput := stdout.String() + stderr.String()
	output, truncated := toolkit.TruncateString(fullOutput, maxBytes)
	result := map[string]any{
		"stdout":    stdout.String(),
		"stderr":    stderr.String(),
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
