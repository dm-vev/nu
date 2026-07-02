package tui

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"strings"
	"sync"

	"nu/internal/agent"
)

// AppOptions configures interactive mode.
type AppOptions struct {
	Stdin      io.Reader
	Stdout     io.Writer
	Stderr     io.Writer
	CWD        string
	Provider   string
	Model      string
	ModelLabel string
	Width      int
	Height     int
}

// App wires line input, rendering, and the Nu agent.
type App struct {
	stdin  io.Reader
	stdout io.Writer
	stderr io.Writer
	width  int
	height int

	mu       sync.Mutex
	agent    *agent.Agent
	editor   *Editor
	state    State
	writeErr error
}

// NewApp creates an idle interactive app.
func NewApp(opts AppOptions) *App {
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
		opts.Width = 80
	}
	if opts.Height <= 0 {
		opts.Height = 24
	}
	return &App{
		stdin:  opts.Stdin,
		stdout: opts.Stdout,
		stderr: opts.Stderr,
		width:  opts.Width,
		height: opts.Height,
		editor: NewEditor(),
		state: State{
			Title:    "Nu",
			CWD:      opts.CWD,
			Provider: opts.Provider,
			Model:    firstNonEmpty(opts.ModelLabel, opts.Model),
			Status:   "idle",
		},
	}
}

// SetAgent injects the provider-backed agent.
func (a *App) SetAgent(agentRef *agent.Agent) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.agent = agentRef
}

// Emit updates UI state from one agent event.
func (a *App) Emit(ev agent.Event) {
	a.mu.Lock()
	switch ev.Type {
	case "turn_start":
		a.state.Status = "working"
	case "message_update":
		a.appendAssistantDeltaLocked(eventText(ev.Data, "delta"))
	case "turn_end":
		a.state.Status = "idle"
		if text := eventText(ev.Data, "text"); text != "" {
			a.replaceLastAssistantLocked(text)
		}
	case "tool_start":
		a.state.Status = "tool"
	case "tool_end":
		a.state.Status = "working"
	}
	a.mu.Unlock()
	a.render()
}

// Run starts the line-oriented interactive loop.
func (a *App) Run(ctx context.Context) error {
	a.render()
	scanner := bufio.NewScanner(a.stdin)
	for scanner.Scan() {
		if err := ctx.Err(); err != nil {
			return fmt.Errorf("run tui: %w", err)
		}
		line := scanner.Text()
		if line == "/quit" || line == "/exit" {
			return a.writeErr
		}
		if strings.TrimSpace(line) == "" {
			continue
		}
		a.editor.Insert(line)
		text := a.editor.Submit()
		a.addUserMessage(text)
		a.render()
		if err := a.prompt(ctx, text); err != nil {
			return err
		}
	}
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("read tui input: %w", err)
	}
	return a.writeErr
}

func (a *App) prompt(ctx context.Context, text string) error {
	a.mu.Lock()
	agentRef := a.agent
	a.mu.Unlock()
	if agentRef == nil {
		return fmt.Errorf("interactive mode requires agent handler")
	}
	// ponytail: line-oriented input is enough for current tests; raw mode belongs in the terminal driver slice.
	if err := agentRef.Prompt(ctx, agent.Prompt{Text: text}); err != nil {
		return fmt.Errorf("interactive prompt: %w", err)
	}
	return nil
}

func (a *App) addUserMessage(text string) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.state.Messages = append(a.state.Messages, Message{Role: "user", Text: text})
	a.state.Editor = a.editor.Snapshot()
}

func (a *App) render() {
	a.mu.Lock()
	frame := Render(a.state, a.width, a.height)
	a.mu.Unlock()
	for _, line := range frame.Lines {
		if _, err := fmt.Fprintln(a.stdout, line); err != nil {
			a.mu.Lock()
			if a.writeErr == nil {
				a.writeErr = fmt.Errorf("write tui frame: %w", err)
			}
			a.mu.Unlock()
			return
		}
	}
}

func (a *App) appendAssistantDeltaLocked(delta string) {
	if delta == "" {
		return
	}
	last := len(a.state.Messages) - 1
	if last >= 0 && a.state.Messages[last].Role == "assistant" {
		a.state.Messages[last].Text += delta
		return
	}
	a.state.Messages = append(a.state.Messages, Message{Role: "assistant", Text: delta})
}

func (a *App) replaceLastAssistantLocked(text string) {
	last := len(a.state.Messages) - 1
	if last >= 0 && a.state.Messages[last].Role == "assistant" {
		a.state.Messages[last].Text = text
		return
	}
	a.state.Messages = append(a.state.Messages, Message{Role: "assistant", Text: text})
}

func eventText(data any, key string) string {
	values, ok := data.(map[string]string)
	if ok {
		return values[key]
	}
	generic, ok := data.(map[string]any)
	if ok {
		text, _ := generic[key].(string)
		return text
	}
	return ""
}
