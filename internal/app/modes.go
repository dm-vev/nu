package app

import (
	"context"
	"fmt"
	"strings"

	"nu/internal/agent"
	"nu/internal/cli"
)

func runMode(ctx context.Context, rt *Runtime, req cli.Request) error {
	select {
	case <-ctx.Done():
		return fmt.Errorf("run mode: %w", ctx.Err())
	default:
	}

	switch req.Command {
	case cli.CommandHelp:
		fmt.Fprint(rt.Options.Stdout, cli.Help(nil))
		return nil
	case cli.CommandVersion:
		fmt.Fprintln(rt.Options.Stdout, cli.Version(rt.Options.Version))
		return nil
	case cli.CommandChat:
		if req.Mode == cli.ModePrint {
			return runPrint(ctx, rt, req)
		}
		return fmt.Errorf("mode %q is not implemented yet", req.Mode)
	default:
		return fmt.Errorf("command %q is not implemented yet", req.Command)
	}
}

func runPrint(ctx context.Context, rt *Runtime, req cli.Request) error {
	if rt.Agent == nil {
		return fmt.Errorf("print mode requires agent handler")
	}
	return rt.Agent.Prompt(ctx, agent.Prompt{Text: strings.Join(req.Prompt, " ")})
}
