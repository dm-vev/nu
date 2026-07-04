package tui

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"nu/internal/agent"
	"nu/internal/model"
	"nu/internal/slash"
	"nu/internal/tui/ansi"
	tuimessage "nu/internal/tui/message"
)

func TestTUIAppRendersPiStyleComponentTree(t *testing.T) {
	var out bytes.Buffer
	app := NewApp(AppOptions{
		Stdout:     &out,
		CWD:        "/tmp/nu",
		Provider:   "fireworks",
		ModelLabel: "GLM 5.2 Fast",
		Width:      80,
		Height:     12,
		Repaint:    true,
	})

	app.Emit(agent.Event{Type: "turn_start"})
	app.Emit(agent.Event{Type: "message_update", Data: map[string]string{"delta": "hello"}})

	got := out.String()
	if strings.Count(got, "\x1b[?2026h") < 2 {
		t.Fatalf("output = %q, want synchronized renders", got)
	}
	if strings.Count(got, "\x1b[2J") != 1 {
		t.Fatalf("output = %q, want exactly one initial clear", got)
	}
	for _, want := range []string{"Nu", "bubbling", "hello", "fireworks/GLM 5.2 Fast"} {
		if !strings.Contains(got, want) {
			t.Fatalf("output = %q, want %q", got, want)
		}
	}
}

func TestTUIAppRendersStructuredMessageParts(t *testing.T) {
	var out bytes.Buffer
	app := NewApp(AppOptions{
		Stdout:     &out,
		CWD:        "/tmp/nu",
		Provider:   "fireworks",
		ModelLabel: "GLM 5.2 Fast",
		Width:      96,
		Height:     18,
		Repaint:    true,
	})

	app.Emit(agent.Event{Type: "turn_start"})
	app.Emit(agent.Event{
		Type: "message_update",
		Data: map[string]string{"kind": "thinking", "thinking_delta": "I should inspect **state**."},
	})
	app.Emit(agent.Event{
		Type: "tool_start",
		Data: map[string]string{"id": "call-1", "name": "bash", "arguments": `{"command":"pwd"}`},
	})
	app.Emit(agent.Event{
		Type: "tool_end",
		Data: map[string]string{
			"id":     "call-1",
			"name":   "bash",
			"result": `{"output":"/tmp/nu\n","exit_code":0}`,
			"error":  "false",
		},
	})
	app.Emit(agent.Event{Type: "message_update", Data: map[string]string{"delta": "Final `answer`."}})

	got := out.String()
	plain := ansi.Strip(got)
	for _, want := range []string{"I should inspect state.", "$ pwd", "/tmp/nu", "Final answer."} {
		if !strings.Contains(plain, want) {
			t.Fatalf("output = %q, want %q", plain, want)
		}
	}
	for _, want := range []string{ansi.Italic, ansi.ToolSuccessBG, ansi.Yellow + "answer"} {
		if !strings.Contains(got, want) {
			t.Fatalf("output = %q, want ANSI sequence %q", got, want)
		}
	}
}

func TestTUIAppRendersUserMessageTextWhite(t *testing.T) {
	var out bytes.Buffer
	app := NewApp(AppOptions{Stdout: &out, Width: 80, Height: 12})

	app.messages = append(app.messages, tuimessage.NewUser("привет"))
	app.rebuildChatLocked()
	app.render()

	got := out.String()
	if !strings.Contains(got, ansi.Text+"привет") {
		t.Fatalf("output = %q, want white user text", got)
	}
	if strings.Contains(got, ansi.Green+"привет") {
		t.Fatalf("output = %q, user text should not be green", got)
	}
}

func TestTUIAppAnchorsEditorAndFooterToBottom(t *testing.T) {
	var out bytes.Buffer
	app := NewApp(AppOptions{
		Stdout:     &out,
		CWD:        "/tmp/nu",
		Provider:   "fireworks",
		ModelLabel: "GLM 5.2 Fast",
		Width:      80,
		Height:     24,
	})

	app.render()

	lines := strings.Split(ansi.Strip(out.String()), "\r\n")
	if len(lines) != 24 {
		t.Fatalf("rendered lines = %d, want 24: %#v", len(lines), lines)
	}
	if !strings.Contains(lines[22], "/tmp/nu") || !strings.Contains(lines[23], "fireworks/GLM 5.2 Fast") {
		t.Fatalf("footer not anchored to bottom: %#v", lines[20:])
	}
	if !editorBorderLine(lines[19]) || !editorBorderLine(lines[21]) {
		t.Fatalf("editor not directly above footer: %#v", lines[18:])
	}
}

func TestTUIAppKeepsStatusLineAboveEditor(t *testing.T) {
	var out bytes.Buffer
	app := NewApp(AppOptions{
		Stdout:  &out,
		Width:   80,
		Height:  12,
		Repaint: true,
	})

	app.Emit(agent.Event{Type: "turn_start"})

	lines := strings.Split(ansi.Strip(out.String()), "\r\n")
	if len(lines) != 12 {
		t.Fatalf("rendered lines = %d, want 12: %#v", len(lines), lines)
	}
	if !strings.Contains(lines[6], "running") {
		t.Fatalf("status line = %q, want running above editor", lines[6])
	}
	if !editorBorderLine(lines[7]) {
		t.Fatalf("editor top line = %q, want editor directly below status", lines[7])
	}
}

func editorBorderLine(line string) bool {
	return strings.Contains(line, "─") || strings.Contains(line, "-")
}

func TestTUIAppUsesLimitedCharsetWhenRequested(t *testing.T) {
	var out bytes.Buffer
	app := NewApp(AppOptions{
		Stdout:  &out,
		Width:   80,
		Height:  12,
		Repaint: true,
		ASCII:   true,
	})

	app.Emit(agent.Event{Type: "turn_start"})

	got := out.String()
	plain := ansi.Strip(got)
	if strings.ContainsAny(plain, "·✢✳✶✻✽─") {
		t.Fatalf("output = %q, want no unsupported spinner or border glyphs", plain)
	}
	if !strings.Contains(plain, "- running") {
		t.Fatalf("output = %q, want ASCII spinner", plain)
	}
	if !strings.Contains(got, ansi.Muted+"--------------------------------------------------------------------------------") {
		t.Fatalf("output = %q, want muted ASCII prompt line", got)
	}
}

func TestTUIHandleRawInputTogglesHeader(t *testing.T) {
	app := NewApp(AppOptions{Width: 80, Height: 12})

	if app.header.Expanded() {
		t.Fatalf("header should start compact")
	}
	if exit := app.handleRawInput("\x0f"); exit {
		t.Fatalf("ctrl+o should not exit")
	}
	if !app.header.Expanded() {
		t.Fatalf("header should expand after ctrl+o")
	}
}

func TestTUIHandleRawInputScrollsViewport(t *testing.T) {
	var out bytes.Buffer
	app := NewApp(AppOptions{Stdout: &out, Width: 80, Height: 8, Repaint: true})
	for _, text := range []string{"one", "two", "three", "four", "five"} {
		app.messages = append(app.messages, tuimessage.NewAssistantText(text))
	}
	app.rebuildChatLocked()
	app.render()
	bottom := ansi.Strip(out.String())

	out.Reset()
	app.handleRawInput("\x1b[5~")
	app.render()
	scrolled := ansi.Strip(out.String())

	if bottom == scrolled {
		t.Fatalf("page up did not change viewport")
	}
	if !strings.Contains(scrolled, "three") {
		t.Fatalf("scrolled viewport = %q, want older message", scrolled)
	}
}

func TestTUIRateLimitShowsFooterNotice(t *testing.T) {
	var out bytes.Buffer
	app := NewApp(AppOptions{Stdout: &out, Width: 80, Height: 12, Repaint: true})

	app.Emit(agent.Event{Type: "rate_limit", Data: map[string]string{"attempt": "1", "max": "5"}})

	got := out.String()
	if !strings.Contains(ansi.Strip(got), "Rate limit") {
		t.Fatalf("output = %q, want footer rate limit notice", got)
	}
	if !strings.Contains(got, ansi.Red+"Rate limit") {
		t.Fatalf("output = %q, want red rate limit notice", got)
	}
}

func TestTUICommandMenuRendersAndCompletes(t *testing.T) {
	var out bytes.Buffer
	app := NewApp(AppOptions{Stdout: &out, Width: 96, Height: 16, Repaint: true})

	app.handleRawInput("/")
	app.handleRawInput("m")
	app.handleRawInput("o")
	app.render()

	plain := ansi.Strip(out.String())
	if !strings.Contains(plain, "/model") || !strings.Contains(plain, "/scoped-models") {
		t.Fatalf("output = %q, want command menu", plain)
	}

	app.handleRawInput("\t")
	if got := app.editor.Text(); got != "/model " {
		t.Fatalf("editor = %q, want completed model command", got)
	}
}

func TestTUICommandMenuEnterRunsSelectedCommand(t *testing.T) {
	var out bytes.Buffer
	app := NewApp(AppOptions{
		Stdout:   &out,
		Width:    96,
		Height:   18,
		Repaint:  true,
		Provider: "fireworks",
		Model:    "glm-fast",
		Models: []model.Model{
			{ID: "glm-fast", Provider: "fireworks", API: "chat", DisplayName: "GLM Fast", Enabled: true},
		},
	})

	app.handleRawInput("/")
	app.handleRawInput("\x1b[B")
	app.handleRawInput("\r")

	plain := ansi.Strip(out.String())
	if !strings.Contains(plain, "Search:") {
		t.Fatalf("output = %q, want /model selector after enter on selected command", plain)
	}
	if strings.Contains(plain, "Unknown command") {
		t.Fatalf("output = %q, selector enter should not submit raw slash text", plain)
	}
}

func TestTUISlashSessionDoesNotCallAgent(t *testing.T) {
	var out bytes.Buffer
	app := NewApp(AppOptions{Stdout: &out, CWD: "/tmp/nu", Width: 96, Height: 16, Repaint: true})

	if err := app.submit("/session"); err != nil {
		t.Fatalf("submit error = %v", err)
	}

	plain := ansi.Strip(out.String())
	if !strings.Contains(plain, "Session") || !strings.Contains(plain, "/tmp/nu") {
		t.Fatalf("output = %q, want local session command output", plain)
	}
	if strings.Contains(plain, "interactive mode requires agent handler") {
		t.Fatalf("output = %q, slash command should not call agent", plain)
	}
}

func TestTUISlashQuitRequestsExit(t *testing.T) {
	app := NewApp(AppOptions{Width: 80, Height: 12})

	if err := app.submit("/quit"); err != nil {
		t.Fatalf("submit error = %v", err)
	}
	if !app.shouldQuit() {
		t.Fatalf("quit flag = false, want true")
	}
}

func TestTUISlashModelOpensSelectorAndSelectsModel(t *testing.T) {
	var out bytes.Buffer
	var selected string
	app := NewApp(AppOptions{
		Stdout:     &out,
		Provider:   "fireworks",
		Model:      "glm-fast",
		ModelLabel: "GLM Fast",
		Models: []model.Model{
			{ID: "gpt-test", Provider: "openai", API: "chat", DisplayName: "GPT Test", Enabled: true},
			{ID: "glm-fast", Provider: "fireworks", API: "chat", DisplayName: "GLM Fast", Enabled: true},
		},
		SetModel: func(_ context.Context, entry model.Model) error {
			selected = entry.Provider + "/" + entry.ID
			return nil
		},
		Width:   96,
		Height:  20,
		Repaint: true,
	})

	if err := app.submit("/model"); err != nil {
		t.Fatalf("submit error = %v", err)
	}

	plain := ansi.Strip(out.String())
	if !strings.Contains(plain, "Search:") || !strings.Contains(plain, "glm-fast [fireworks] *") {
		t.Fatalf("output = %q, want model selector", plain)
	}

	app.handleRawInput("\x1b[B")
	app.handleRawInput("\r")

	if selected != "openai/gpt-test" {
		t.Fatalf("selected = %q, want openai/gpt-test", selected)
	}
	plain = ansi.Strip(out.String())
	if strings.Contains(plain, "Model: GPT Test") {
		t.Fatalf("output = %q, should not render selected model in animated status", plain)
	}
	if !strings.Contains(plain, "Model switched from GLM Fast to GPT Test") ||
		!strings.Contains(plain, "openai/GPT Test") {
		t.Fatalf("output = %q, want selected model chat notice and footer", plain)
	}
}

func TestTUISlashModelExactMatchSelectsWithoutMenu(t *testing.T) {
	var selected string
	app := NewApp(AppOptions{
		Provider:   "fireworks",
		Model:      "glm-fast",
		ModelLabel: "GLM Fast",
		Models: []model.Model{
			{ID: "gpt-test", Provider: "openai", API: "chat", DisplayName: "GPT Test", Enabled: true},
			{ID: "glm-fast", Provider: "fireworks", API: "chat", DisplayName: "GLM Fast", Enabled: true},
		},
		SetModel: func(_ context.Context, entry model.Model) error {
			selected = entry.Provider + "/" + entry.ID
			return nil
		},
		Width:  80,
		Height: 12,
	})

	if err := app.submit("/model openai/gpt-test"); err != nil {
		t.Fatalf("submit error = %v", err)
	}
	if app.models.Visible() {
		t.Fatalf("model selector should stay closed after exact match")
	}
	if selected != "openai/gpt-test" {
		t.Fatalf("selected = %q, want openai/gpt-test", selected)
	}
}

func TestTUIAllBuiltinSlashCommandsHaveHandlers(t *testing.T) {
	dir := t.TempDir()
	importPath := filepath.Join(dir, "import.jsonl")
	if err := os.WriteFile(importPath, []byte(
		"{\"role\":\"user\",\"text\":\"hello\"}\n{\"role\":\"assistant\",\"text\":\"hi\"}\n",
	), 0o600); err != nil {
		t.Fatalf("WriteFile import error = %v", err)
	}

	var out bytes.Buffer
	app := NewApp(AppOptions{
		Stdout:   &out,
		CWD:      dir,
		Home:     filepath.Join(dir, "home"),
		Provider: "fireworks",
		Model:    "glm-fast",
		Models: []model.Model{
			{ID: "glm-fast", Provider: "fireworks", API: "chat", DisplayName: "GLM Fast", Enabled: true},
		},
		Width:   100,
		Height:  24,
		Repaint: true,
	})
	app.messages = []tuimessage.Message{tuimessage.NewUser("hello"), tuimessage.NewAssistantText("hi")}
	app.rebuildChatLocked()

	args := map[string]string{
		"export": filepath.Join(dir, "out.jsonl"),
		"import": importPath,
		"name":   "demo",
		"login":  "fireworks env FIREWORKS_API_KEY",
		"logout": "fireworks",
		"resume": importPath,
	}
	for _, command := range slash.Builtins() {
		input := "/" + command.Name
		if value := args[command.Name]; value != "" {
			input += " " + value
		}
		if err := app.submit(input); err != nil {
			t.Fatalf("submit %s error = %v", input, err)
		}
	}

	plain := ansi.Strip(out.String())
	if strings.Contains(plain, "backend is not implemented") {
		t.Fatalf("output = %q, builtin command hit placeholder", plain)
	}
}
