//go:build !windows

package coding

import (
	"bytes"
	"context"
	"errors"
	"os"
	"os/exec"
	"syscall"
)

func runCommand(ctx context.Context, cwd string, command string) (string, string, int, bool) {
	cmd := exec.CommandContext(ctx, "sh", "-c", command)
	cmd.Dir = cwd
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	cmd.Cancel = func() error {
		if cmd.Process == nil {
			return os.ErrProcessDone
		}
		// Kill the process group so timed-out shell children do not keep running.
		return syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)
	}
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
