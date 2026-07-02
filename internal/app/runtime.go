package app

import (
	"io"
	"strings"

	"nu/internal/cli"
)

// Options carries process state into one app invocation.
type Options struct {
	Args    []string
	Env     []string
	CWD     string
	Home    string
	Stdin   io.Reader
	Stdout  io.Writer
	Stderr  io.Writer
	Version cli.VersionInfo

	Print func(*Runtime, cli.Request) error
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
