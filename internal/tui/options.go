package tui

import (
	"context"
	"io"
	"os"
	"strings"

	"nu/internal/model"
)

const defaultContext = 128000

// AppOptions configures interactive mode.
type AppOptions struct {
	Stdin       io.Reader
	Stdout      io.Writer
	Stderr      io.Writer
	CWD         string
	Provider    string
	Model       string
	ModelLabel  string
	SessionID   string
	SessionName string
	Models      []model.Model
	SetModel    func(context.Context, model.Model) error
	Version     string
	Home        string
	Branch      string
	Context     int
	Width       int
	Height      int
	Repaint     bool
	ASCII       bool
}

func normalizeOptions(opts AppOptions) AppOptions {
	if opts.Stdin == nil {
		opts.Stdin = strings.NewReader("")
	}
	if opts.Stdout == nil {
		opts.Stdout = io.Discard
	}
	if opts.Stderr == nil {
		opts.Stderr = io.Discard
	}
	if opts.Width <= 0 {
		opts.Width = envInt("COLUMNS", 80)
	}
	if opts.Height <= 0 {
		opts.Height = envInt("LINES", 24)
	}
	if opts.Context <= 0 {
		opts.Context = defaultContext
	}
	return opts
}

func limitedCharset(opts AppOptions) bool {
	if opts.ASCII {
		return true
	}
	switch strings.ToLower(strings.TrimSpace(os.Getenv("NU_TUI_ASCII"))) {
	case "1", "true", "yes", "on":
		return true
	}
	switch strings.ToLower(strings.TrimSpace(os.Getenv("TERM"))) {
	case "linux", "dumb", "vt100", "vt102", "ansi":
		return true
	default:
		return false
	}
}

func statusFrames(opts AppOptions) []string {
	if limitedCharset(opts) {
		return []string{"-", "\\", "|", "/"}
	}
	return nil
}
