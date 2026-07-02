package app

import (
	"fmt"
	"io"
	"strings"

	"nu/internal/agent"
	"nu/internal/cli"
	"nu/internal/provider"
)

// Options carries process state into one app invocation.
type Options struct {
	Args       []string
	Env        []string
	CWD        string
	Home       string
	Stdin      io.Reader
	Stdout     io.Writer
	Stderr     io.Writer
	Version    cli.VersionInfo
	Provider   provider.Streamer
	ProviderID string
	API        string
	Model      string
}

// Runtime holds dependencies shared by mode handlers.
type Runtime struct {
	Options Options
	Agent   *agent.Agent
}

func normalizeOptions(opts Options) Options {
	if opts.Stdin == nil {
		opts.Stdin = strings.NewReader("")
	}
	if opts.Stdout == nil {
		opts.Stdout = io.Discard
	}
	if opts.Stderr == nil {
		opts.Stderr = io.Discard
	}
	return opts
}

func newAgent(opts Options) *agent.Agent {
	if opts.Provider == nil {
		return nil
	}
	return agent.New(agent.Options{
		Provider:   opts.Provider,
		ProviderID: opts.ProviderID,
		API:        opts.API,
		Model:      opts.Model,
		Emit: func(ev agent.Event) {
			if ev.Type != "turn_end" {
				return
			}
			data, ok := ev.Data.(map[string]string)
			if !ok {
				return
			}
			if text := data["text"]; text != "" {
				// Print mode writes only final assistant text; live deltas stay internal.
				fmt.Fprintln(opts.Stdout, text)
			}
		},
	})
}
