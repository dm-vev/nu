package tui

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
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
	Version    string
	Home       string
	Branch     string
	Context    int
	Width      int
	Height     int
	Repaint    bool
}

// App wires line input, rendering, and the Nu agent.
type App struct {
	stdin  io.Reader
	width  int
	height int
	term   *Terminal

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
		opts.Width = envInt("COLUMNS", 80)
	}
	if opts.Height <= 0 {
		opts.Height = envInt("LINES", 24)
	}
	if strings.TrimSpace(opts.Branch) == "" {
		opts.Branch = currentGitBranch(opts.CWD)
	}
	return &App{
		stdin:  opts.Stdin,
		width:  opts.Width,
		height: opts.Height,
		term:   NewTerminal(opts.Stdout, opts.Repaint),
		editor: NewEditor(),
		state: State{
			Title:         "Nu",
			Version:       opts.Version,
			CWD:           opts.CWD,
			Home:          opts.Home,
			Branch:        opts.Branch,
			Provider:      opts.Provider,
			Model:         firstNonEmpty(opts.ModelLabel, opts.Model),
			ContextWindow: opts.Context,
			AutoCompact:   true,
			ContextFiles:  []string{"AGENTS.md"},
			Status:        "idle",
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

// Run starts the interactive loop.
func (a *App) Run(ctx context.Context) (runErr error) {
	restore, raw, err := enableRawInput(a.stdin)
	if err != nil {
		return err
	}
	if restore != nil {
		defer func() {
			if err := restore(); err != nil {
				a.rememberWriteErr(err)
				if runErr == nil {
					runErr = err
				}
			}
		}()
	}
	defer func() {
		if err := a.term.Close(); err != nil {
			a.rememberWriteErr(err)
			if runErr == nil {
				runErr = err
			}
		}
	}()
	a.render()
	if raw {
		stopResize := watchResize(a.render)
		defer stopResize()
		return a.runRaw(ctx)
	}
	return a.runLine(ctx)
}

func (a *App) runLine(ctx context.Context) error {
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

func (a *App) runRaw(ctx context.Context) error {
	reader := bufio.NewReader(a.stdin)
	for {
		if err := ctx.Err(); err != nil {
			return fmt.Errorf("run tui: %w", err)
		}
		ch, _, err := reader.ReadRune()
		if err != nil {
			if err == io.EOF {
				return a.writeErr
			}
			return fmt.Errorf("read tui input: %w", err)
		}
		switch ch {
		case 0x04:
			if a.editor.Snapshot().Text == "" {
				return a.writeErr
			}
		case '\r', '\n':
			text := a.editor.Submit()
			a.syncEditor()
			if strings.TrimSpace(text) == "" {
				a.render()
				continue
			}
			a.addUserMessage(text)
			a.render()
			if err := a.prompt(ctx, text); err != nil {
				return err
			}
		case 0x7f, 0x08:
			a.editor.Backspace()
			a.syncEditor()
			a.render()
		case 0x03:
			return a.writeErr
		default:
			if ch < 0x20 {
				continue
			}
			a.editor.Insert(string(ch))
			a.syncEditor()
			a.render()
		}
	}
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
	a.width, a.height = a.term.Size(a.width, a.height)
	frame := Render(a.state, a.width, a.height)
	a.mu.Unlock()
	if err := a.term.Draw(frame); err != nil {
		a.mu.Lock()
		if a.writeErr == nil {
			a.writeErr = err
		}
		a.mu.Unlock()
		return
	}
}

func (a *App) syncEditor() {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.state.Editor = a.editor.Snapshot()
}

func (a *App) rememberWriteErr(err error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.writeErr == nil {
		a.writeErr = err
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

func envInt(name string, fallback int) int {
	value := strings.TrimSpace(os.Getenv(name))
	if value == "" {
		return fallback
	}
	var parsed int
	if _, err := fmt.Sscanf(value, "%d", &parsed); err != nil || parsed <= 0 {
		return fallback
	}
	return parsed
}

func currentGitBranch(cwd string) string {
	dir := firstNonEmpty(cwd, ".")
	for {
		headPath := filepath.Join(dir, ".git", "HEAD")
		data, err := os.ReadFile(headPath)
		if err == nil {
			// Reading .git/HEAD is enough for the footer; shelling out to git would be slower and noisier.
			head := strings.TrimSpace(string(data))
			return strings.TrimPrefix(head, "ref: refs/heads/")
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return ""
		}
		dir = parent
	}
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
