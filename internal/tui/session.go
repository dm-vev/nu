package tui

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/dm-vev/nu/internal/tui/message"
)

func (a *App) settingsCommandText() string {
	a.mu.Lock()
	defer a.mu.Unlock()
	return fmt.Sprintf(
		"Settings\n\n| Field | Value |\n| --- | --- |\n| cwd | %s |\n| home | %s |\n| provider | %s |\n| model | %s |\n| context | %d |\n| session | %s |",
		a.cwd,
		a.home,
		a.provider,
		a.modelLabel,
		a.context,
		firstNonEmpty(a.sessionName, a.sessionID, "unnamed"),
	)
}

func (a *App) handleImportSlash(args string, replace bool) error {
	path := strings.TrimSpace(args)
	if path == "" {
		if replace {
			a.appendLocalMessage("Usage: `/resume path/to/session.jsonl`")
		} else {
			a.appendLocalMessage("Usage: `/import path/to/session.jsonl`")
		}
		return nil
	}
	path = a.resolvePath(path)
	messages, err := readExportedMessages(path)
	if err != nil {
		return err
	}
	a.mu.Lock()
	if replace {
		a.messages = messages
	} else {
		a.messages = append(a.messages, messages...)
	}
	a.rebuildChatLocked()
	a.mu.Unlock()
	a.render()
	if replace {
		a.appendLocalMessage("Resumed `" + path + "`.")
		return nil
	}
	a.appendLocalMessage("Imported `" + path + "`.")
	return nil
}

func (a *App) handleNameSlash(args string) {
	name := strings.TrimSpace(args)
	if name == "" {
		a.mu.Lock()
		current := firstNonEmpty(a.sessionName, "unnamed")
		a.mu.Unlock()
		a.appendLocalMessage("Session name: `" + current + "`")
		return
	}
	a.mu.Lock()
	a.sessionName = name
	a.status.SetText("Session: " + name)
	a.mu.Unlock()
	a.render()
}

func (a *App) handleChangelogSlash() error {
	for _, path := range []string{"CHANGELOG.md", "docs/changelog.md"} {
		data, err := os.ReadFile(a.resolvePath(path))
		if err == nil {
			a.appendLocalMessage(string(data))
			return nil
		}
	}
	a.appendLocalMessage("No changelog file found.")
	return nil
}

func (a *App) handleForkSlash(args string) {
	a.mu.Lock()
	index := a.forkIndexLocked(args)
	if index < 0 {
		a.mu.Unlock()
		a.appendLocalMessage("No user message to fork from.")
		return
	}
	kept := index + 1
	a.messages = append([]message.Message(nil), a.messages[:kept]...)
	a.rebuildChatLocked()
	a.mu.Unlock()
	a.appendLocalMessage(fmt.Sprintf("Forked at message %d.", kept))
}

func (a *App) handleCloneSlash() {
	a.mu.Lock()
	cloned := make([]message.Message, 0, len(a.messages))
	for _, msg := range a.messages {
		cloned = append(cloned, msg.Clone())
	}
	a.messages = cloned
	a.status.SetText("Session cloned")
	a.mu.Unlock()
	a.render()
}

func (a *App) treeCommandText() string {
	a.mu.Lock()
	messages := append([]message.Message(nil), a.messages...)
	a.mu.Unlock()
	if len(messages) == 0 {
		return "Tree\n\nNo messages in the current session."
	}
	var b strings.Builder
	b.WriteString("Tree\n\n| # | Role | Preview |\n| --- | --- | --- |\n")
	for i, msg := range messages {
		fmt.Fprintf(&b, "| %d | %s | %s |\n", i+1, msg.Role, tableCell(preview(messageText(msg), 64)))
	}
	return b.String()
}

func (a *App) handleCompactSlash() {
	a.mu.Lock()
	if len(a.messages) <= 8 {
		a.mu.Unlock()
		a.appendLocalMessage("Nothing to compact.")
		return
	}
	removed := len(a.messages) - 6
	tail := append([]message.Message(nil), a.messages[len(a.messages)-6:]...)
	// ponytail: cheap local compaction; replace with model summary when persisted compaction lands.
	a.messages = append([]message.Message{message.NewAssistantText(fmt.Sprintf("Compacted %d earlier messages.", removed))}, tail...)
	a.rebuildChatLocked()
	a.mu.Unlock()
	a.render()
}

func (a *App) handleReloadSlash() {
	a.mu.Lock()
	a.branch = currentGitBranch(a.cwd)
	opts := a.footer.Options()
	opts.Branch = a.branch
	a.footer.SetOptions(opts)
	a.status.SetText("Reloaded")
	a.mu.Unlock()
	a.render()
}

func (a *App) forkIndexLocked(args string) int {
	if index, err := strconv.Atoi(strings.TrimSpace(args)); err == nil && index > 0 && index <= len(a.messages) {
		return index - 1
	}
	for i := len(a.messages) - 1; i >= 0; i-- {
		if a.messages[i].Role == message.RoleUser {
			return i
		}
	}
	return -1
}

func tableCell(value string) string {
	value = strings.ReplaceAll(value, "|", "\\|")
	value = strings.ReplaceAll(value, "\n", " ")
	return value
}

func preview(value string, limit int) string {
	value = strings.TrimSpace(strings.Join(strings.Fields(value), " "))
	if len([]rune(value)) <= limit {
		return value
	}
	runes := []rune(value)
	return string(runes[:limit]) + "..."
}
