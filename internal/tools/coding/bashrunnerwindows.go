//go:build windows

package coding

import (
	"bytes"
	"context"
	"errors"
	"os/exec"
)

func runCommand(ctx context.Context, cwd string, command string) (string, string, int, bool) {
	cmd := exec.CommandContext(ctx, "cmd", "/C", command)
	cmd.Dir = cwd
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	timedOut := errors.Is(ctx.Err(), context.DeadlineExceeded)
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
	return stdout.String(), stderr.String(), exitCode, timedOut
}
