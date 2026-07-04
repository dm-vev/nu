package tui

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"html"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"nu/internal/model"
	tuimessage "nu/internal/tui/message"
)

type slashExportRecord struct {
	Role string `json:"role"`
	Text string `json:"text"`
}

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

func (a *App) scopedModelsCommandText() string {
	a.mu.Lock()
	models := append([]modelSummary(nil), summarizeModels(a.available, a.provider, a.modelID)...)
	a.mu.Unlock()
	if len(models) == 0 {
		return "Scoped models\n\nNo models are visible for the current provider credentials."
	}
	var b strings.Builder
	b.WriteString("Scoped models\n\n| Current | Provider | Model | Display |\n| --- | --- | --- | --- |\n")
	for _, item := range models {
		current := ""
		if item.Current {
			current = "*"
		}
		fmt.Fprintf(&b, "| %s | %s | %s | %s |\n", current, item.Provider, item.ID, item.Display)
	}
	return b.String()
}

func (a *App) handleExportSlash(args string) error {
	path := strings.TrimSpace(args)
	if path == "" {
		path = "nu-session.html"
	}
	path = a.resolvePath(path)
	if err := a.exportMessages(path); err != nil {
		return err
	}
	a.appendLocalMessage("Exported session to `" + path + "`.")
	return nil
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

func (a *App) handleShareSlash(args string) error {
	path := strings.TrimSpace(args)
	if path == "" {
		path = filepath.Join(os.TempDir(), "nu-share-"+time.Now().UTC().Format("20060102-150405")+".html")
	} else {
		path = a.resolvePath(path)
	}
	if err := a.exportMessages(path); err != nil {
		return err
	}
	a.appendLocalMessage("Share export ready at `" + path + "`.")
	return nil
}

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
	a.messages = append([]tuimessage.Message(nil), a.messages[:kept]...)
	a.rebuildChatLocked()
	a.mu.Unlock()
	a.appendLocalMessage(fmt.Sprintf("Forked at message %d.", kept))
}

func (a *App) handleCloneSlash() {
	a.mu.Lock()
	cloned := make([]tuimessage.Message, 0, len(a.messages))
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
	messages := append([]tuimessage.Message(nil), a.messages...)
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

func (a *App) handleCompactSlash() {
	a.mu.Lock()
	if len(a.messages) <= 8 {
		a.mu.Unlock()
		a.appendLocalMessage("Nothing to compact.")
		return
	}
	removed := len(a.messages) - 6
	tail := append([]tuimessage.Message(nil), a.messages[len(a.messages)-6:]...)
	// ponytail: cheap local compaction; replace with model summary when persisted compaction lands.
	a.messages = append([]tuimessage.Message{tuimessage.NewAssistantText(fmt.Sprintf("Compacted %d earlier messages.", removed))}, tail...)
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

func (a *App) sessionSnapshot() []tuimessage.Message {
	a.mu.Lock()
	defer a.mu.Unlock()
	out := make([]tuimessage.Message, 0, len(a.messages))
	for _, msg := range a.messages {
		out = append(out, msg.Clone())
	}
	return out
}

func (a *App) exportMessages(path string) error {
	messages := a.sessionSnapshot()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("create export dir: %w", err)
	}
	var data []byte
	var err error
	switch strings.ToLower(filepath.Ext(path)) {
	case ".jsonl":
		data, err = renderJSONL(messages)
	case ".md", ".markdown":
		data = []byte(renderMarkdown(messages))
	default:
		data = []byte(renderHTML(messages))
	}
	if err != nil {
		return err
	}
	if err := os.WriteFile(path, data, 0o600); err != nil {
		return fmt.Errorf("write export %s: %w", path, err)
	}
	return nil
}

func (a *App) resolvePath(path string) string {
	if filepath.IsAbs(path) {
		return filepath.Clean(path)
	}
	return filepath.Join(firstNonEmpty(a.cwd, "."), path)
}

func (a *App) authPath() string {
	return filepath.Join(firstNonEmpty(a.home, a.cwd, "."), ".nu", "auth.json")
}

func (a *App) trustPath() string {
	return filepath.Join(firstNonEmpty(a.home, a.cwd, "."), ".nu", "agent", "trust.json")
}

func (a *App) lastAssistantText() string {
	a.mu.Lock()
	defer a.mu.Unlock()
	for i := len(a.messages) - 1; i >= 0; i-- {
		if a.messages[i].Role == tuimessage.RoleAssistant {
			return messageText(a.messages[i])
		}
	}
	return ""
}

func (a *App) forkIndexLocked(args string) int {
	if index, err := strconv.Atoi(strings.TrimSpace(args)); err == nil && index > 0 && index <= len(a.messages) {
		return index - 1
	}
	for i := len(a.messages) - 1; i >= 0; i-- {
		if a.messages[i].Role == tuimessage.RoleUser {
			return i
		}
	}
	return -1
}

type modelSummary struct {
	Provider string
	ID       string
	Display  string
	Current  bool
}

func summarizeModels(models []model.Model, providerID string, modelID string) []modelSummary {
	out := make([]modelSummary, 0, len(models))
	for _, entry := range models {
		out = append(out, modelSummary{
			Provider: entry.Provider,
			ID:       entry.ID,
			Display:  modelDisplayName(entry),
			Current:  entry.Provider == providerID && entry.ID == modelID,
		})
	}
	return out
}

func readExportedMessages(path string) ([]tuimessage.Message, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open import %s: %w", path, err)
	}
	defer file.Close()
	var messages []tuimessage.Message
	scanner := bufio.NewScanner(file)
	scanner.Buffer(nil, 1024*1024)
	for scanner.Scan() {
		var record slashExportRecord
		if err := json.Unmarshal(scanner.Bytes(), &record); err != nil {
			return nil, fmt.Errorf("decode import %s: %w", path, err)
		}
		switch record.Role {
		case tuimessage.RoleUser:
			messages = append(messages, tuimessage.NewUser(record.Text))
		case tuimessage.RoleAssistant:
			messages = append(messages, tuimessage.NewAssistantText(record.Text))
		default:
			return nil, fmt.Errorf("decode import %s: unknown role %q", path, record.Role)
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("read import %s: %w", path, err)
	}
	return messages, nil
}

func renderJSONL(messages []tuimessage.Message) ([]byte, error) {
	var b bytes.Buffer
	enc := json.NewEncoder(&b)
	for _, msg := range messages {
		if err := enc.Encode(slashExportRecord{Role: msg.Role, Text: messageText(msg)}); err != nil {
			return nil, fmt.Errorf("encode export: %w", err)
		}
	}
	return b.Bytes(), nil
}

func renderMarkdown(messages []tuimessage.Message) string {
	var b strings.Builder
	for _, msg := range messages {
		fmt.Fprintf(&b, "## %s\n\n%s\n\n", msg.Role, messageText(msg))
	}
	return b.String()
}

func renderHTML(messages []tuimessage.Message) string {
	var b strings.Builder
	b.WriteString("<!doctype html><meta charset=\"utf-8\"><title>Nu session</title>")
	for _, msg := range messages {
		fmt.Fprintf(&b, "<h2>%s</h2><pre>%s</pre>", html.EscapeString(msg.Role), html.EscapeString(messageText(msg)))
	}
	return b.String()
}

func messageText(msg tuimessage.Message) string {
	var b strings.Builder
	for _, part := range msg.Parts {
		if part.Text != "" {
			b.WriteString(part.Text)
		}
		if part.ToolArguments != "" {
			b.WriteString(part.ToolArguments)
		}
		if part.ToolResult != "" {
			b.WriteString(part.ToolResult)
		}
	}
	return b.String()
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

func writeBoolMap(path string, key string, value bool) error {
	values := map[string]bool{}
	data, err := os.ReadFile(path)
	if err == nil {
		_ = json.Unmarshal(data, &values)
	}
	values[key] = value
	return writeJSONFile(path, values)
}

func writeAuthProvider(path string, providerID string, credential map[string]string) error {
	file := map[string]map[string]map[string]string{"providers": {}}
	data, err := os.ReadFile(path)
	if err == nil {
		_ = json.Unmarshal(data, &file)
	}
	if file["providers"] == nil {
		file["providers"] = map[string]map[string]string{}
	}
	file["providers"][providerID] = credential
	return writeJSONFile(path, file)
}

func removeAuthProvider(path string, providerID string) error {
	file := map[string]map[string]map[string]string{"providers": {}}
	data, err := os.ReadFile(path)
	if err == nil {
		_ = json.Unmarshal(data, &file)
	}
	if file["providers"] == nil {
		file["providers"] = map[string]map[string]string{}
	}
	delete(file["providers"], providerID)
	return writeJSONFile(path, file)
}

func writeJSONFile(path string, value any) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("create dir %s: %w", filepath.Dir(path), err)
	}
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return fmt.Errorf("encode json %s: %w", path, err)
	}
	if err := os.WriteFile(path, append(data, '\n'), 0o600); err != nil {
		return fmt.Errorf("write json %s: %w", path, err)
	}
	return nil
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
