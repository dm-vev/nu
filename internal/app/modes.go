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
	"nu/internal/rpc"
	"nu/internal/tui"
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
		if req.Mode == cli.ModeRPC {
			return runRPC(ctx, rt, req)
		}
		if req.Mode == cli.ModeInteractive {
			return runInteractive(ctx, rt, req)
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
	entries, registry, err := loadModelRegistry(modelsPath(rt.Options.Home, req.ModelsPath))
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
		line := fmt.Sprintf(
			"%s/%s\t%s\t%d\t%d",
			entry.Provider,
			entry.ID,
			entry.API,
			entry.ContextWindow,
			entry.MaxOutput,
		)
		if strings.TrimSpace(entry.DisplayName) != "" {
			line += "\t" + strings.TrimSpace(entry.DisplayName)
		}
		if _, err := fmt.Fprintln(rt.Options.Stdout, line); err != nil {
			return fmt.Errorf("write model list: %w", err)
		}
	}
	return nil
}

func runRPC(ctx context.Context, rt *Runtime, _ cli.Request) error {
	server := rpc.NewServer(rpc.Options{
		Stdin:      rt.Options.Stdin,
		Stdout:     rt.Options.Stdout,
		Stderr:     rt.Options.Stderr,
		CWD:        rt.Options.CWD,
		SessionID:  rt.Options.SessionID,
		Provider:   rt.Options.ProviderID,
		API:        rt.Options.API,
		Model:      rt.Options.Model,
		ModelLabel: rt.Options.ModelLabel,
	})
	a := newAgent(rt.Options, server.Emit)
	if a == nil {
		return fmt.Errorf("rpc mode requires agent handler")
	}
	server.SetAgent(a)
	return server.Run(ctx)
}

func runInteractive(ctx context.Context, rt *Runtime, req cli.Request) error {
	var a *agent.Agent
	ui := tui.NewApp(tui.AppOptions{
		Stdin:      rt.Options.Stdin,
		Stdout:     rt.Options.Stdout,
		Stderr:     rt.Options.Stderr,
		CWD:        rt.Options.CWD,
		Home:       rt.Options.Home,
		Provider:   rt.Options.ProviderID,
		Model:      rt.Options.Model,
		ModelLabel: rt.Options.ModelLabel,
		SessionID:  rt.Options.SessionID,
		Models:     rt.Options.Models,
		SetModel: func(ctx context.Context, selected model.Model) error {
			if a == nil {
				return fmt.Errorf("interactive agent is not ready")
			}
			store, err := auth.Load(authFilePath(rt.Options.Home), rt.Options.Env)
			if err != nil {
				return err
			}
			settings, err := loadProviderSettings(rt.Options.Home)
			if err != nil {
				return err
			}
			streamer, err := newProviderClient(ctx, rt.Options, store, req, settings, selected)
			if err != nil {
				return err
			}
			if err := a.SetProviderModel(streamer, selected.Provider, selected.API, selected.ID); err != nil {
				return err
			}
			return saveSelectedModel(rt.Options.Home, selected)
		},
		Version: rt.Options.Version.Version,
		Context: rt.Options.ModelContext,
		Repaint: true,
	})
	a = newAgent(rt.Options, ui.Emit)
	if a == nil {
		return fmt.Errorf("interactive mode requires agent handler")
	}
	ui.SetAgent(a)
	return ui.Run(ctx)
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
