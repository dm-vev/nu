package tui

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/dm-vev/nu/internal/tui/message"
)

func (a *App) handleCopySlash() error {
	text := a.lastAssistantText()
	if strings.TrimSpace(text) == "" {
		a.appendLocalMessage("No assistant message to copy.")
		return nil
	}
	if err := copyToClipboard(text); err != nil {
		a.appendLocalMessage("Copy failed: " + err.Error())
		return nil
	}
	a.appendLocalMessage("Copied last assistant message.")
	return nil
}

func (a *App) lastAssistantText() string {
	a.mu.Lock()
	defer a.mu.Unlock()
	for i := len(a.messages) - 1; i >= 0; i-- {
		if a.messages[i].Role == message.RoleAssistant {
			return messageText(a.messages[i])
		}
	}
	return ""
}

func copyToClipboard(text string) error {
	commands := [][]string{{"wl-copy"}, {"xclip", "-selection", "clipboard"}, {"xsel", "--clipboard", "--input"}, {"pbcopy"}}
	for _, candidate := range commands {
		if _, err := exec.LookPath(candidate[0]); err != nil {
			continue
		}
		cmd := exec.Command(candidate[0], candidate[1:]...)
		cmd.Stdin = strings.NewReader(text)
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("%s: %w", candidate[0], err)
		}
		return nil
	}
	return fmt.Errorf("no clipboard command found")
}
