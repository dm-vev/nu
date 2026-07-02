package app

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"

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
	SessionID  string
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

type jsonSessionHeader struct {
	Type       string    `json:"type"`
	Schema     int       `json:"schema"`
	ID         string    `json:"id"`
	CreatedAt  time.Time `json:"created_at"`
	CWD        string    `json:"cwd"`
	App        string    `json:"app"`
	AppVersion string    `json:"app_version"`
}

type printEventWriter struct {
	w   io.Writer
	err error
}

func (w *printEventWriter) emit(ev agent.Event) {
	if w.err != nil || ev.Type != "turn_end" {
		return
	}
	data, ok := ev.Data.(map[string]string)
	if !ok {
		return
	}
	if text := data["text"]; text != "" {
		// Print mode writes only final assistant text; live deltas stay internal.
		_, w.err = fmt.Fprintln(w.w, text)
	}
}

type jsonEventWriter struct {
	w   io.Writer
	err error
}

func (w *jsonEventWriter) emit(ev agent.Event) {
	if w.err != nil {
		return
	}
	w.err = writeJSONLine(w.w, ev)
}

func newAgent(opts Options, emit func(agent.Event)) *agent.Agent {
	if opts.Provider == nil {
		return nil
	}
	return agent.New(agent.Options{
		Provider:   opts.Provider,
		ProviderID: opts.ProviderID,
		API:        opts.API,
		Model:      opts.Model,
		Emit:       emit,
	})
}

func newJSONSessionHeader(opts Options) (jsonSessionHeader, error) {
	id := opts.SessionID
	if id == "" {
		generated, err := newSessionID()
		if err != nil {
			return jsonSessionHeader{}, err
		}
		id = generated
	}
	return jsonSessionHeader{
		Type:       "session",
		Schema:     1,
		ID:         id,
		CreatedAt:  time.Now().UTC(),
		CWD:        opts.CWD,
		App:        "nu",
		AppVersion: opts.Version.Version,
	}, nil
}

func newSessionID() (string, error) {
	var data [16]byte
	if _, err := rand.Read(data[:]); err != nil {
		return "", fmt.Errorf("generate session id: %w", err)
	}
	return hex.EncodeToString(data[:]), nil
}

func writeJSONLine(w io.Writer, value any) error {
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("marshal json line: %w", err)
	}
	if _, err := w.Write(append(data, '\n')); err != nil {
		return fmt.Errorf("write json line: %w", err)
	}
	return nil
}
