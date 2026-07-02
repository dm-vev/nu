package app

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"nu/internal/agent"
	"nu/internal/auth"
	"nu/internal/cli"
	"nu/internal/model"
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
	case cli.CommandListModels:
		return runListModels(ctx, rt)
	case cli.CommandChat:
		if req.Mode == cli.ModePrint {
			return runPrint(ctx, rt, req)
		}
		if req.Mode == cli.ModeJSON {
			return runJSON(ctx, rt, req)
		}
		return fmt.Errorf("mode %q is not implemented yet", req.Mode)
	default:
		return fmt.Errorf("command %q is not implemented yet", req.Command)
	}
}

func runPrint(ctx context.Context, rt *Runtime, req cli.Request) error {
	writer := &printEventWriter{w: rt.Options.Stdout}
	a := newAgent(rt.Options, writer.emit)
	if a == nil {
		return fmt.Errorf("print mode requires agent handler")
	}
	if err := a.Prompt(ctx, agent.Prompt{Text: strings.Join(req.Prompt, " ")}); err != nil {
		return err
	}
	if writer.err != nil {
		return writer.err
	}
	return nil
}

func runListModels(ctx context.Context, rt *Runtime) error {
	authState, err := providerAuthState(ctx, rt.Options)
	if err != nil {
		return err
	}
	registry := model.NewRegistry(model.Builtins())
	for _, entry := range registry.Available(authState) {
		if _, err := fmt.Fprintf(
			rt.Options.Stdout,
			"%s/%s\t%s\t%d\t%d\n",
			entry.Provider,
			entry.ID,
			entry.API,
			entry.ContextWindow,
			entry.MaxOutput,
		); err != nil {
			return fmt.Errorf("write model list: %w", err)
		}
	}
	return nil
}

func providerAuthState(ctx context.Context, opts Options) (map[string]bool, error) {
	store, err := auth.Load(authFilePath(opts.Home), opts.Env)
	if err != nil {
		return nil, err
	}
	state := map[string]bool{}
	for _, providerID := range []string{"openai", "anthropic", "google", "bedrock"} {
		_, ok, err := store.ResolveAPIKey(ctx, providerID)
		if err != nil {
			return nil, err
		}
		if ok {
			state[providerID] = true
		}
	}
	return state, nil
}

func authFilePath(home string) string {
	if strings.TrimSpace(home) == "" {
		return ""
	}
	return filepath.Join(home, ".nu", "auth.json")
}

func runJSON(ctx context.Context, rt *Runtime, req cli.Request) error {
	writer := &jsonEventWriter{w: rt.Options.Stdout}
	a := newAgent(rt.Options, writer.emit)
	if a == nil {
		return fmt.Errorf("json mode requires agent handler")
	}
	header, err := newJSONSessionHeader(rt.Options)
	if err != nil {
		return err
	}
	// Header comes first so JSON clients can initialize session state before events.
	if err := writeJSONLine(rt.Options.Stdout, header); err != nil {
		return err
	}
	if err := a.Prompt(ctx, agent.Prompt{Text: strings.Join(req.Prompt, " ")}); err != nil {
		return err
	}
	if writer.err != nil {
		return writer.err
	}
	return nil
}
