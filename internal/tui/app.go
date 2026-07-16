package tui

import (
	"context"
	"sync"

	"nu/internal/agentui"
	"nu/internal/model"
	"nu/internal/tui/components"
	"nu/internal/tui/core"
	"nu/internal/tui/editor"
	"nu/internal/tui/engine"
	"nu/internal/tui/message"
	"nu/internal/tui/terminal"
)

// App wires Nu agent events into the component TUI.
type App struct {
	mu     sync.Mutex
	agent  *agentui.Agent
	term   *terminal.Terminal
	ui     *engine.TUI
	editor *editor.Editor

	header *components.Header
	chat   *core.Container
	menu   *components.CommandMenu
	models *components.ModelMenu
	status *components.Status
	footer *components.Footer

	cwd         string
	home        string
	branch      string
	provider    string
	modelID     string
	modelLabel  string
	sessionID   string
	sessionName string
	version     string
	context     int
	available   []model.Model
	setModel    func(context.Context, model.Model) error
	messages    []message.Message
	writeErr    error
	quit        bool
	promptWG    sync.WaitGroup
	submitCtx   context.Context
}

// NewApp creates an idle interactive app.
func NewApp(opts AppOptions) *App {
	opts = normalizeOptions(opts)
	choices := modelChoices(opts)
	term := terminal.New(opts.Stdin, opts.Stdout, opts.Width, opts.Height)
	ui := engine.New(term, engine.Options{Title: windowTitle(opts.CWD), MinRenderRows: opts.Height})
	app := &App{
		term:        term,
		ui:          ui,
		editor:      editor.New(),
		header:      components.NewHeader(headerOptions(opts)),
		chat:        &core.Container{},
		menu:        components.NewCommandMenu(components.SlashBuiltins(), commandMenuOptions()),
		models:      components.NewModelMenu(choices, modelMenuOptions()),
		status:      components.NewStatus(muted, statusFrames(opts)...),
		footer:      components.NewFooter(footerOptions(opts)),
		cwd:         opts.CWD,
		home:        opts.Home,
		branch:      firstNonEmpty(opts.Branch, currentGitBranch(opts.CWD)),
		provider:    opts.Provider,
		modelID:     opts.Model,
		modelLabel:  firstNonEmpty(opts.ModelLabel, opts.Model),
		sessionID:   opts.SessionID,
		sessionName: opts.SessionName,
		version:     opts.Version,
		context:     opts.Context,
		available:   choices,
		setModel:    opts.SetModel,
	}
	if limitedCharset(opts) {
		app.editor.SetBorderRune('-')
		app.editor.SetStyles(muted, ansiText)
	} else {
		app.editor.SetStyles(green, ansiText)
	}
	app.editor.SetChangeHandler(func(text string) {
		app.menu.SetText(text)
	})
	app.buildLayout()
	return app
}

func modelChoices(opts AppOptions) []model.Model {
	if len(opts.Models) > 0 {
		return append([]model.Model(nil), opts.Models...)
	}
	if opts.Provider == "" || opts.Model == "" {
		return nil
	}
	return []model.Model{{
		ID:          opts.Model,
		Provider:    opts.Provider,
		API:         "chat",
		DisplayName: opts.ModelLabel,
		Enabled:     true,
	}}
}

// SetAgent injects the provider-backed agentui.
func (a *App) SetAgent(agentRef *agentui.Agent) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.agent = agentRef
}

func (a *App) requestQuit() {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.quit = true
}

func (a *App) shouldQuit() bool {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.quit
}
