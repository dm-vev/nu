package slash

import "strings"

// Command is one slash command shown in the interactive command menu.
type Command struct {
	Name        string
	Description string
	Source      string
}

var builtins = []Command{
	{Name: "settings", Description: "Open settings menu", Source: "builtin"},
	{Name: "model", Description: "Select model (opens selector UI)", Source: "builtin"},
	{Name: "scoped-models", Description: "Enable/disable models for Ctrl+P cycling", Source: "builtin"},
	{Name: "export", Description: "Export session (HTML default, or specify path: .html/.jsonl)", Source: "builtin"},
	{Name: "import", Description: "Import and resume a session from a JSONL file", Source: "builtin"},
	{Name: "share", Description: "Share session as a secret GitHub gist", Source: "builtin"},
	{Name: "copy", Description: "Copy last agent message to clipboard", Source: "builtin"},
	{Name: "name", Description: "Set session display name", Source: "builtin"},
	{Name: "session", Description: "Show session info and stats", Source: "builtin"},
	{Name: "changelog", Description: "Show changelog entries", Source: "builtin"},
	{Name: "hotkeys", Description: "Show all keyboard shortcuts", Source: "builtin"},
	{Name: "fork", Description: "Create a new fork from a previous user message", Source: "builtin"},
	{Name: "clone", Description: "Duplicate the current session at the current position", Source: "builtin"},
	{Name: "tree", Description: "Navigate session tree (switch branches)", Source: "builtin"},
	{Name: "trust", Description: "Save project trust decision for future sessions", Source: "builtin"},
	{Name: "login", Description: "Configure provider authentication", Source: "builtin"},
	{Name: "logout", Description: "Remove provider authentication", Source: "builtin"},
	{Name: "new", Description: "Start a new session", Source: "builtin"},
	{Name: "compact", Description: "Manually compact the session context", Source: "builtin"},
	{Name: "resume", Description: "Resume a different session", Source: "builtin"},
	{Name: "reload", Description: "Reload keybindings, extensions, skills, prompts, and themes", Source: "builtin"},
	{Name: "quit", Description: "Quit Nu", Source: "builtin"},
}

// Builtins returns Pi-compatible built-in slash commands.
func Builtins() []Command {
	return append([]Command(nil), builtins...)
}

// Lookup returns a built-in command by name.
func Lookup(name string) (Command, bool) {
	name = strings.TrimPrefix(strings.TrimSpace(name), "/")
	for _, command := range builtins {
		if command.Name == name {
			return command, true
		}
	}
	return Command{}, false
}

// Parse splits a slash input into command name and arguments.
func Parse(input string) (string, string, bool) {
	input = strings.TrimSpace(input)
	if !strings.HasPrefix(input, "/") || input == "/" {
		return "", "", false
	}
	input = strings.TrimPrefix(input, "/")
	name, args, _ := strings.Cut(input, " ")
	return strings.TrimSpace(name), strings.TrimSpace(args), true
}

// Filter returns commands matching a slash prefix.
func Filter(prefix string, limit int) []Command {
	prefix = strings.TrimPrefix(strings.TrimSpace(prefix), "/")
	matches := make([]Command, 0, len(builtins))
	for _, command := range builtins {
		if prefix == "" || strings.HasPrefix(command.Name, prefix) || strings.Contains(command.Name, prefix) {
			matches = append(matches, command)
			if limit > 0 && len(matches) >= limit {
				break
			}
		}
	}
	return matches
}
