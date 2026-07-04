package tui

import (
	"context"
	"fmt"
	"strings"

	"nu/internal/agent"
	"nu/internal/model"
	"nu/internal/slash"
	"nu/internal/tui/components/modelmenu"
	tuimessage "nu/internal/tui/message"
)

func (a *App) submit(value string) error {
	value = strings.TrimSpace(value)
	if value == "" {
		a.render()
		return nil
	}
	if name, args, ok := slash.Parse(value); ok {
		return a.runSlashCommand(name, args)
	}

	a.mu.Lock()
	agentRef := a.agent
	a.messages = append(a.messages, tuimessage.NewUser(value))
	a.rebuildChatLocked()
	a.mu.Unlock()
	a.render()

	if agentRef == nil {
		return fmt.Errorf("interactive mode requires agent handler")
	}
	a.startPrompt(agentRef, value)
	return nil
}

func (a *App) runSlashCommand(name string, args string) error {
	if _, ok := slash.Lookup(name); !ok {
		a.appendLocalMessage("Unknown command: /" + name)
		return nil
	}
	switch name {
	case "settings":
		a.appendLocalMessage(a.settingsCommandText())
	case "quit":
		a.requestQuit()
	case "new":
		a.mu.Lock()
		agentRef := a.agent
		a.messages = nil
		a.status.SetText("")
		a.rebuildChatLocked()
		a.mu.Unlock()
		if agentRef != nil {
			agentRef.Reset()
		}
		a.render()
	case "scoped-models":
		a.appendLocalMessage(a.scopedModelsCommandText())
	case "export":
		return a.handleExportSlash(args)
	case "import":
		return a.handleImportSlash(args, false)
	case "share":
		return a.handleShareSlash(args)
	case "copy":
		return a.handleCopySlash()
	case "name":
		a.handleNameSlash(args)
	case "session":
		a.appendLocalMessage(a.sessionCommandText())
	case "changelog":
		return a.handleChangelogSlash()
	case "hotkeys":
		a.appendLocalMessage("Hotkeys\n\n| Key | Action |\n| --- | --- |\n| Esc | interrupt / clear |\n| Ctrl+C | clear / exit |\n| Ctrl+D | delete / exit |\n| Ctrl+O | toggle startup help |\n| PageUp/PageDown | scroll history |\n| Tab | complete slash command |")
	case "fork":
		a.handleForkSlash(args)
	case "clone":
		a.handleCloneSlash()
	case "tree":
		a.appendLocalMessage(a.treeCommandText())
	case "trust":
		return a.handleTrustSlash(args)
	case "login":
		return a.handleLoginSlash(args)
	case "logout":
		return a.handleLogoutSlash(args)
	case "model":
		return a.handleModelSlash(args)
	case "compact":
		a.handleCompactSlash()
	case "resume":
		return a.handleImportSlash(args, true)
	case "reload":
		a.handleReloadSlash()
	default:
		a.appendLocalMessage("Unknown command: /" + name)
	}
	return nil
}

func (a *App) appendLocalMessage(text string) {
	a.mu.Lock()
	a.messages = append(a.messages, tuimessage.NewAssistantText(text))
	a.rebuildChatLocked()
	a.mu.Unlock()
	a.render()
}

func (a *App) sessionCommandText() string {
	a.mu.Lock()
	defer a.mu.Unlock()
	return fmt.Sprintf(
		"Session\n\n| Field | Value |\n| --- | --- |\n| name | %s |\n| id | %s |\n| cwd | %s |\n| provider | %s |\n| model | %s |\n| messages | %d |",
		firstNonEmpty(a.sessionName, "unnamed"),
		firstNonEmpty(a.sessionID, "memory"),
		a.cwd,
		a.provider,
		a.modelLabel,
		len(a.messages),
	)
}

func (a *App) handleModelSlash(args string) error {
	args = strings.TrimSpace(args)
	if args == "" {
		a.openModelMenu("")
		return nil
	}
	if selected, ok := a.findModel(args); ok {
		return a.selectModel(selected)
	}
	a.openModelMenu(args)
	return nil
}

func (a *App) openModelMenu(query string) {
	a.mu.Lock()
	a.models.SetModels(a.available)
	a.models.Open(query, a.provider, a.modelID)
	a.status.SetText("")
	a.mu.Unlock()
	a.render()
}

func (a *App) findModel(query string) (model.Model, bool) {
	a.mu.Lock()
	available := append([]model.Model(nil), a.available...)
	a.mu.Unlock()
	registry := model.NewRegistry(available)
	selected, err := registry.Resolve(query, authAll(available))
	if err != nil {
		return model.Model{}, false
	}
	return selected, true
}

func authAll(models []model.Model) map[string]bool {
	auth := map[string]bool{}
	for _, entry := range models {
		auth[entry.Provider] = true
	}
	return auth
}

func (a *App) selectModel(selected model.Model) error {
	if strings.TrimSpace(selected.Provider) == "" || strings.TrimSpace(selected.API) == "" || strings.TrimSpace(selected.ID) == "" {
		return fmt.Errorf("set model: provider, api, and model are required")
	}
	a.mu.Lock()
	agentRef := a.agent
	setModel := a.setModel
	ctx := a.submitCtx
	a.mu.Unlock()
	if ctx == nil {
		ctx = context.Background()
	}
	if setModel != nil {
		if err := setModel(ctx, selected); err != nil {
			return fmt.Errorf("set model %s/%s: %w", selected.Provider, selected.ID, err)
		}
	} else if agentRef != nil {
		if err := agentRef.SetModel(selected.Provider, selected.API, selected.ID); err != nil {
			return fmt.Errorf("set model %s/%s: %w", selected.Provider, selected.ID, err)
		}
	}

	label := modelDisplayName(selected)
	contextWindow := selected.ContextWindow
	if contextWindow <= 0 {
		contextWindow = defaultContext
	}
	a.mu.Lock()
	previous := firstNonEmpty(a.modelLabel, a.modelID, "unknown")
	a.provider = selected.Provider
	a.modelID = selected.ID
	a.modelLabel = label
	a.context = contextWindow
	a.models.Close()
	a.status.SetText("")
	opts := a.footer.Options()
	opts.Provider = a.provider
	opts.Model = label
	opts.Context = contextWindow
	a.footer.SetOptions(opts)
	a.messages = append(a.messages, tuimessage.NewAssistantText("Model switched from "+previous+" to "+label))
	a.rebuildChatLocked()
	a.mu.Unlock()
	a.render()
	return nil
}

func modelDisplayName(entry model.Model) string {
	if strings.TrimSpace(entry.DisplayName) != "" {
		return entry.DisplayName
	}
	return entry.ID
}

func (a *App) handleModelMenuInput(data string) bool {
	if !a.models.Visible() {
		return false
	}
	action := a.models.HandleInput(data)
	switch action {
	case modelmenu.ActionCancel:
		a.models.Close()
	case modelmenu.ActionSelect:
		selected, ok := a.models.Selected()
		if !ok {
			a.models.Close()
			break
		}
		if err := a.selectModel(selected); err != nil {
			a.appendError(err)
		}
	case modelmenu.ActionChanged:
	default:
	}
	return true
}

func (a *App) startPrompt(agentRef *agent.Agent, value string) {
	ctx := a.submitCtx
	if ctx == nil {
		ctx = context.Background()
	}
	a.promptWG.Add(1)
	go func() {
		defer a.promptWG.Done()
		if err := agentRef.Prompt(ctx, agent.Prompt{Text: value}); err != nil {
			a.appendError(fmt.Errorf("interactive prompt: %w", err))
		}
	}()
}

func (a *App) appendError(err error) {
	a.mu.Lock()
	a.status.SetText("")
	a.messages = append(a.messages, tuimessage.NewAssistantText("Error: "+err.Error()))
	a.rebuildChatLocked()
	a.mu.Unlock()
	a.render()
}

func (a *App) abortActiveTurn() bool {
	a.mu.Lock()
	agentRef := a.agent
	a.mu.Unlock()
	if agentRef == nil || !agentRef.Busy() {
		return false
	}
	// Abort is delegated to agent so provider streaming unwinds through the normal turn_end path.
	agentRef.Abort()
	a.mu.Lock()
	a.status.SetText("aborting")
	a.mu.Unlock()
	a.render()
	return true
}

func (a *App) rememberWriteErr(err error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.writeErr == nil {
		a.writeErr = err
	}
}

func (a *App) render() {
	if err := a.ui.RenderNow(); err != nil {
		a.rememberWriteErr(err)
	}
}
