package app

import (
	"context"
	"io"
	"strings"

	"github.com/dm-vev/nu/contracts"
	"github.com/dm-vev/nu/internal/agentui"
	"github.com/dm-vev/nu/internal/app/cli"
	"github.com/dm-vev/nu/internal/model"
)

// Options carries process state into one app invocation.
type Options struct {
	Args         []string
	Env          []string
	CWD          string
	Home         string
	Stdin        io.Reader
	Stdout       io.Writer
	Stderr       io.Writer
	Version      cli.VersionInfo
	Runner       contracts.StreamingAgent
	LLM          contracts.LLM
	BuildLLM     func(context.Context, agentui.Config) (contracts.LLM, error)
	ProviderID   string
	API          string
	Model        string
	ModelLabel   string
	ModelContext int
	Models       []model.Model
	Tools        []contracts.Tool
	Memory       contracts.Memory
	SessionID    string
}

// Runtime holds dependencies shared by mode handlers.
type Runtime struct {
	Options Options
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
