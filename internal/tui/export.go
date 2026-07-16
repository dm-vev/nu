package tui

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"html"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/dm-vev/nu/internal/tui/message"
)

type slashExportRecord struct {
	Role string `json:"role"`
	Text string `json:"text"`
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

func (a *App) sessionSnapshot() []message.Message {
	a.mu.Lock()
	defer a.mu.Unlock()
	out := make([]message.Message, 0, len(a.messages))
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

func readExportedMessages(path string) ([]message.Message, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open import %s: %w", path, err)
	}
	defer file.Close()
	var messages []message.Message
	scanner := bufio.NewScanner(file)
	scanner.Buffer(nil, 1024*1024)
	for scanner.Scan() {
		var record slashExportRecord
		if err := json.Unmarshal(scanner.Bytes(), &record); err != nil {
			return nil, fmt.Errorf("decode import %s: %w", path, err)
		}
		switch record.Role {
		case message.RoleUser:
			messages = append(messages, message.NewUser(record.Text))
		case message.RoleAssistant:
			messages = append(messages, message.NewAssistantText(record.Text))
		default:
			return nil, fmt.Errorf("decode import %s: unknown role %q", path, record.Role)
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("read import %s: %w", path, err)
	}
	return messages, nil
}

func renderJSONL(messages []message.Message) ([]byte, error) {
	var b bytes.Buffer
	enc := json.NewEncoder(&b)
	for _, msg := range messages {
		if err := enc.Encode(slashExportRecord{Role: msg.Role, Text: messageText(msg)}); err != nil {
			return nil, fmt.Errorf("encode export: %w", err)
		}
	}
	return b.Bytes(), nil
}

func renderMarkdown(messages []message.Message) string {
	var b strings.Builder
	for _, msg := range messages {
		fmt.Fprintf(&b, "## %s\n\n%s\n\n", msg.Role, messageText(msg))
	}
	return b.String()
}

func renderHTML(messages []message.Message) string {
	var b strings.Builder
	b.WriteString("<!doctype html><meta charset=\"utf-8\"><title>Nu session</title>")
	for _, msg := range messages {
		fmt.Fprintf(&b, "<h2>%s</h2><pre>%s</pre>", html.EscapeString(msg.Role), html.EscapeString(messageText(msg)))
	}
	return b.String()
}

func messageText(msg message.Message) string {
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
