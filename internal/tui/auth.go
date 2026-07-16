package tui

import (
	"fmt"
	"path/filepath"
	"strings"
)

func (a *App) handleTrustSlash(args string) error {
	trusted := strings.TrimSpace(args) != "no" && strings.TrimSpace(args) != "false"
	path := a.trustPath()
	if err := writeBoolMap(path, a.cwd, trusted); err != nil {
		return err
	}
	a.appendLocalMessage(fmt.Sprintf("Project trust saved: `%s` = `%t`.", a.cwd, trusted))
	return nil
}

func (a *App) handleLoginSlash(args string) error {
	fields := strings.Fields(args)
	if len(fields) < 2 {
		a.appendLocalMessage("Usage: `/login provider key`, `/login provider env ENV_NAME`, or `/login provider command ...`")
		return nil
	}
	providerID := fields[0]
	credential := map[string]string{}
	switch fields[1] {
	case "env":
		if len(fields) < 3 {
			a.appendLocalMessage("Usage: `/login provider env ENV_NAME`")
			return nil
		}
		credential["api_key_env"] = fields[2]
	case "command":
		if len(fields) < 3 {
			a.appendLocalMessage("Usage: `/login provider command COMMAND`")
			return nil
		}
		credential["api_key_command"] = strings.Join(fields[2:], " ")
	default:
		credential["api_key"] = strings.Join(fields[1:], " ")
	}
	if err := writeAuthProvider(a.authPath(), providerID, credential); err != nil {
		return err
	}
	a.appendLocalMessage("Saved credentials for `" + providerID + "`.")
	return nil
}

func (a *App) handleLogoutSlash(args string) error {
	providerID := strings.TrimSpace(args)
	if providerID == "" {
		a.appendLocalMessage("Usage: `/logout provider`")
		return nil
	}
	if err := removeAuthProvider(a.authPath(), providerID); err != nil {
		return err
	}
	a.appendLocalMessage("Removed credentials for `" + providerID + "`.")
	return nil
}

func (a *App) authPath() string {
	return filepath.Join(firstNonEmpty(a.home, a.cwd, "."), ".nu", "auth.json")
}

func (a *App) trustPath() string {
	return filepath.Join(firstNonEmpty(a.home, a.cwd, "."), ".nu", "agent", "trust.json")
}
