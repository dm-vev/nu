package app

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"nu/internal/agent"
	"nu/internal/auth"
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
	case cli.CommandListModels:
		return runListModels(ctx, rt, req)
	case cli.CommandChat:
		opts, err := configureProvider(ctx, rt.Options, req)
		if err != nil {
			return err
		}
		rt = &Runtime{Options: opts}
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

func runListModels(ctx context.Context, rt *Runtime, req cli.Request) error {
	// List output must reflect the same custom registry used for runtime selection.
	entries, registry, err := loadModelRegistry(req.ModelsPath)
	if err != nil {
		return err
	}
	store, err := auth.Load(authFilePath(rt.Options.Home), rt.Options.Env)
	if err != nil {
		return err
	}
	authState, err := providerAuthState(ctx, store, entries)
	if err != nil {
		return err
	}
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
