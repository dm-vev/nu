package app

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"nu/internal/agent"
	"nu/internal/cli"
	"nu/internal/provider"
	"nu/internal/testkit"
)

func TestAppRunHelp(t *testing.T) {
	var stdout bytes.Buffer
	code := Run(context.Background(), Options{
		Args:   []string{"--help"},
		Stdout: &stdout,
	})
	if code != exitOK {
		t.Fatalf("Run exit code = %d, want %d", code, exitOK)
	}
	if !strings.Contains(stdout.String(), "Usage: nu") {
		t.Fatalf("Run help stdout = %q, want usage", stdout.String())
	}
}

func TestAppRunPrintModeUsesInjectedRuntime(t *testing.T) {
	var stdout bytes.Buffer
	fake := testkit.NewScriptedProvider(
		provider.Event{Type: provider.EventStart},
		provider.Event{Type: provider.EventText, Delta: "ok"},
		provider.Event{Type: provider.EventDone},
	)
	code := Run(context.Background(), Options{
		Args:     []string{"--print", "hello"},
		Stdout:   &stdout,
		Provider: fake,
	})
	if code != exitOK {
		t.Fatalf("Run exit code = %d, want %d", code, exitOK)
	}
	requests := fake.Requests()
	if len(requests) != 1 {
		t.Fatalf("Provider requests = %d, want 1", len(requests))
	}
	if requests[0].Messages[0].Content != "hello" {
		t.Fatalf("Provider prompt = %q, want hello", requests[0].Messages[0].Content)
	}
	if stdout.String() != "ok\n" {
		t.Fatalf("Run stdout = %q, want ok", stdout.String())
	}
}

func TestNUF170JSONModeStdoutIsOnlyJSONL(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	fake := testkit.NewScriptedProvider(
		provider.Event{Type: provider.EventStart},
		provider.Event{Type: provider.EventText, Delta: "ok"},
		provider.Event{Type: provider.EventDone},
	)
	code := Run(context.Background(), Options{
		Args:      []string{"--mode", "json", "hello"},
		CWD:       "/tmp/nu-test",
		Stdout:    &stdout,
		Stderr:    &stderr,
		Provider:  fake,
		SessionID: "s1",
	})
	if code != exitOK {
		t.Fatalf("Run exit code = %d, want %d; stderr=%q", code, exitOK, stderr.String())
	}
	if stderr.String() != "" {
		t.Fatalf("Run stderr = %q, want empty", stderr.String())
	}

	lines := strings.Split(strings.TrimSpace(stdout.String()), "\n")
	if len(lines) < 2 {
		t.Fatalf("JSON mode lines = %d, want header and events; stdout=%q", len(lines), stdout.String())
	}
	var header map[string]any
	if err := json.Unmarshal([]byte(lines[0]), &header); err != nil {
		t.Fatalf("Header JSON error = %v", err)
	}
	if header["type"] != "session" || header["id"] != "s1" {
		t.Fatalf("Header = %#v, want session s1", header)
	}
	for _, line := range lines {
		var obj map[string]any
		if err := json.Unmarshal([]byte(line), &obj); err != nil {
			t.Fatalf("JSONL line %q error = %v", line, err)
		}
	}
	var last map[string]any
	if err := json.Unmarshal([]byte(lines[len(lines)-1]), &last); err != nil {
		t.Fatalf("Last JSON error = %v", err)
	}
	if last["type"] != "turn_end" {
		t.Fatalf("Last event = %#v, want turn_end", last)
	}
}

func TestNUF170JSONModeFeedsToolResultBackToProvider(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	fake := testkit.NewScriptedProviderScripts(
		[]provider.Event{
			{Type: provider.EventStart},
			{Type: provider.EventToolCallStart, Index: 0, ToolCallID: "call-1", ToolName: "fake"},
			{Type: provider.EventToolCallEnd, Index: 0},
			{Type: provider.EventDone, StopReason: "tool_use"},
		},
		[]provider.Event{
			{Type: provider.EventStart},
			{Type: provider.EventText, Delta: "ok"},
			{Type: provider.EventDone},
		},
	)
	code := Run(context.Background(), Options{
		Args:     []string{"--mode", "json", "hello"},
		CWD:      "/tmp/nu-test",
		Stdout:   &stdout,
		Stderr:   &stderr,
		Provider: fake,
		Tools: map[string]agent.ToolFunc{
			"fake": func(context.Context, agent.ToolCall) (agent.ToolResult, error) {
				return agent.ToolResult{Content: "tool result"}, nil
			},
		},
	})
	if code != exitOK {
		t.Fatalf("Run exit code = %d, want %d; stderr=%q", code, exitOK, stderr.String())
	}
	requests := fake.Requests()
	if len(requests) != 2 {
		t.Fatalf("Provider requests = %d, want 2", len(requests))
	}
	lastMessage := requests[1].Messages[len(requests[1].Messages)-1]
	if lastMessage.Role != "tool" || lastMessage.Content != "tool result" {
		t.Fatalf("Second request last message = %#v, want tool result", lastMessage)
	}
	if !strings.Contains(stdout.String(), `"type":"tool_end"`) {
		t.Fatalf("JSON stdout missing tool_end event: %q", stdout.String())
	}
}

func TestNUF002DispatchRPCMode(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	fake := testkit.NewScriptedProvider()

	code := Run(context.Background(), Options{
		Args:     []string{"--mode", "rpc"},
		Stdin:    strings.NewReader(`{"id":"s","type":"get_state"}` + "\n" + `{"id":"q","type":"shutdown"}` + "\n"),
		Stdout:   &stdout,
		Stderr:   &stderr,
		Provider: fake,
	})
	if code != exitOK {
		t.Fatalf("Run exit code = %d, want %d; stderr=%q", code, exitOK, stderr.String())
	}
	if stderr.String() != "" {
		t.Fatalf("stderr = %q, want empty", stderr.String())
	}
	if !strings.Contains(stdout.String(), `"id":"s"`) || !strings.Contains(stdout.String(), `"command":"get_state"`) {
		t.Fatalf("RPC stdout = %q, want get_state response", stdout.String())
	}
	if !strings.Contains(stdout.String(), `"id":"q"`) || !strings.Contains(stdout.String(), `"command":"shutdown"`) {
		t.Fatalf("RPC stdout = %q, want shutdown response", stdout.String())
	}
}

func TestNUF002DispatchInteractiveMode(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	fake := testkit.NewScriptedProvider()

	code := Run(context.Background(), Options{
		Args:     []string{"--mode", "interactive"},
		Stdin:    strings.NewReader("/quit\n"),
		Stdout:   &stdout,
		Stderr:   &stderr,
		Provider: fake,
		CWD:      "/tmp/nu-test",
	})
	if code != exitOK {
		t.Fatalf("Run exit code = %d, want %d; stderr=%q", code, exitOK, stderr.String())
	}
	if !strings.Contains(stdout.String(), "Nu") || !strings.Contains(stdout.String(), "/tmp/nu-test") {
		t.Fatalf("interactive stdout = %q, want rendered frame", stdout.String())
	}
}

func TestJSONModeUsesBuiltinToolsByDefault(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "a.txt"), []byte("from built-in"), 0o644); err != nil {
		t.Fatalf("WriteFile error = %v", err)
	}
	var stdout bytes.Buffer
	fake := testkit.NewScriptedProviderScripts(
		[]provider.Event{
			{Type: provider.EventStart},
			{Type: provider.EventToolCallStart, Index: 0, ToolCallID: "call-1", ToolName: "read"},
			{Type: provider.EventToolCallDelta, Index: 0, Delta: `{"path":"a.txt"}`},
			{Type: provider.EventToolCallEnd, Index: 0},
			{Type: provider.EventDone, StopReason: "tool_use"},
		},
		[]provider.Event{
			{Type: provider.EventStart},
			{Type: provider.EventText, Delta: "ok"},
			{Type: provider.EventDone},
		},
	)

	code := Run(context.Background(), Options{
		Args:     []string{"--mode", "json", "read"},
		CWD:      dir,
		Stdout:   &stdout,
		Provider: fake,
	})
	if code != exitOK {
		t.Fatalf("Run exit code = %d, want %d", code, exitOK)
	}
	requests := fake.Requests()
	lastMessage := requests[1].Messages[len(requests[1].Messages)-1]
	if !strings.Contains(lastMessage.Content, "from built-in") {
		t.Fatalf("tool result = %q, want built-in read content", lastMessage.Content)
	}
}

func TestJSONSessionHeaderDefaults(t *testing.T) {
	header, err := newJSONSessionHeader(Options{CWD: "/tmp/nu-test"})
	if err != nil {
		t.Fatalf("newJSONSessionHeader error = %v", err)
	}
	if len(header.ID) != 36 || header.ID[14] != '4' {
		t.Fatalf("Session id = %q, want UUIDv4-like id", header.ID)
	}
	if header.AppVersion != "dev" {
		t.Fatalf("AppVersion = %q, want dev", header.AppVersion)
	}
}

func TestAppRunPrintModeWithoutAvailableModelFails(t *testing.T) {
	var stderr bytes.Buffer
	code := Run(context.Background(), Options{
		Args:   []string{"--print", "hello"},
		Stderr: &stderr,
	})
	if code != exitError {
		t.Fatalf("Run exit code = %d, want %d", code, exitError)
	}
	if !strings.Contains(stderr.String(), "resolve model: no available models") {
		t.Fatalf("Run stderr = %q, want no available models error", stderr.String())
	}
}

func TestListModelsUsesAuthState(t *testing.T) {
	var stdout bytes.Buffer
	code := Run(context.Background(), Options{
		Args:   []string{"--list-models"},
		Env:    []string{"OPENAI_API_KEY=test"},
		Stdout: &stdout,
	})
	if code != exitOK {
		t.Fatalf("Run exit code = %d, want %d", code, exitOK)
	}
	if !strings.Contains(stdout.String(), "openai/gpt-5.5") {
		t.Fatalf("stdout = %q, want OpenAI model", stdout.String())
	}
	if strings.Contains(stdout.String(), "anthropic/") {
		t.Fatalf("stdout = %q, should hide unauthenticated Anthropic models", stdout.String())
	}
}

func TestListModelsUsesCustomModelsPath(t *testing.T) {
	dir := t.TempDir()
	modelsPath := filepath.Join(dir, "models.json")
	raw := `{"models":[{"id":"local","provider":"compat","api":"chat","requires_auth":false,"context_window":7,"max_output":3}]}`
	if err := os.WriteFile(modelsPath, []byte(raw), 0o644); err != nil {
		t.Fatalf("WriteFile error = %v", err)
	}

	var stdout bytes.Buffer
	code := Run(context.Background(), Options{
		Args:   []string{"--list-models", "--models", modelsPath},
		Stdout: &stdout,
	})
	if code != exitOK {
		t.Fatalf("Run exit code = %d, want %d", code, exitOK)
	}
	if !strings.Contains(stdout.String(), "compat/local\tchat\t7\t3") {
		t.Fatalf("stdout = %q, want custom model", stdout.String())
	}
}

func TestListModelsUsesGlobalModelsPathByDefault(t *testing.T) {
	home := t.TempDir()
	agentDir := filepath.Join(home, ".nu", "agent")
	if err := os.MkdirAll(agentDir, 0o755); err != nil {
		t.Fatalf("MkdirAll error = %v", err)
	}
	raw := `{"models":[{"id":"global","provider":"fireworks","api":"chat","display_name":"Global Fireworks","requires_auth":false,"context_window":9,"max_output":4}]}`
	if err := os.WriteFile(filepath.Join(agentDir, "models.json"), []byte(raw), 0o644); err != nil {
		t.Fatalf("WriteFile error = %v", err)
	}

	var stdout bytes.Buffer
	code := Run(context.Background(), Options{
		Args:   []string{"--list-models"},
		Home:   home,
		Stdout: &stdout,
	})
	if code != exitOK {
		t.Fatalf("Run exit code = %d, want %d", code, exitOK)
	}
	if !strings.Contains(stdout.String(), "fireworks/global\tchat\t9\t4\tGlobal Fireworks") {
		t.Fatalf("stdout = %q, want global model", stdout.String())
	}
}

func TestListModelsIncludesDisplayName(t *testing.T) {
	dir := t.TempDir()
	modelsPath := filepath.Join(dir, "models.json")
	raw := `{"models":[{"id":"glm","provider":"fireworks","api":"chat","display_name":"GLM 5.2 Fast","requires_auth":false,"context_window":7,"max_output":3}]}`
	if err := os.WriteFile(modelsPath, []byte(raw), 0o644); err != nil {
		t.Fatalf("WriteFile error = %v", err)
	}

	var stdout bytes.Buffer
	code := Run(context.Background(), Options{
		Args:   []string{"--list-models", "--models", modelsPath},
		Stdout: &stdout,
	})
	if code != exitOK {
		t.Fatalf("Run exit code = %d, want %d", code, exitOK)
	}
	if !strings.Contains(stdout.String(), "fireworks/glm\tchat\t7\t3\tGLM 5.2 Fast") {
		t.Fatalf("stdout = %q, want display name", stdout.String())
	}
}

func TestPrintModeBuildsProviderFromCLI(t *testing.T) {
	var gotPath string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		w.Header().Set("Content-Type", "text/event-stream")
		_, _ = w.Write([]byte("data: {\"choices\":[{\"delta\":{\"content\":\"ok\"}}]}\n\n"))
		_, _ = w.Write([]byte("data: {\"choices\":[{\"finish_reason\":\"stop\"}]}\n\n"))
	}))
	defer server.Close()

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := Run(context.Background(), Options{
		Args:   []string{"--print", "--provider", server.URL, "--model", "local", "--api-key", "key", "hello"},
		Stdout: &stdout,
		Stderr: &stderr,
	})
	if code != exitOK {
		t.Fatalf("Run exit code = %d, want %d; stderr=%q", code, exitOK, stderr.String())
	}
	if gotPath != "/chat/completions" || stdout.String() != "ok\n" {
		t.Fatalf("path/stdout = %q/%q, want compat chat response", gotPath, stdout.String())
	}
}

func TestPrintModeBuildsFireworksProviderFromGlobalModels(t *testing.T) {
	home := t.TempDir()
	agentDir := filepath.Join(home, ".nu", "agent")
	if err := os.MkdirAll(agentDir, 0o755); err != nil {
		t.Fatalf("MkdirAll error = %v", err)
	}
	raw := `{"models":[{"id":"glm","provider":"fireworks","api":"chat","requires_auth":true}]}`
	if err := os.WriteFile(filepath.Join(agentDir, "models.json"), []byte(raw), 0o644); err != nil {
		t.Fatalf("WriteFile error = %v", err)
	}
	authRaw := `{"providers":{"fireworks":{"api_key":"key"}}}`
	if err := os.MkdirAll(filepath.Join(home, ".nu"), 0o755); err != nil {
		t.Fatalf("MkdirAll auth dir error = %v", err)
	}
	if err := os.WriteFile(filepath.Join(home, ".nu", "auth.json"), []byte(authRaw), 0o644); err != nil {
		t.Fatalf("WriteFile auth error = %v", err)
	}

	var gotPath string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		w.Header().Set("Content-Type", "text/event-stream")
		_, _ = w.Write([]byte("data: {\"choices\":[{\"delta\":{\"content\":\"ok\"}}]}\n\n"))
		_, _ = w.Write([]byte("data: {\"choices\":[{\"finish_reason\":\"stop\"}]}\n\n"))
	}))
	defer server.Close()

	oldBaseURL := fireworksBaseURL
	fireworksBaseURL = server.URL
	defer func() {
		fireworksBaseURL = oldBaseURL
	}()

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := Run(context.Background(), Options{
		Args:   []string{"--print", "hello"},
		Home:   home,
		Stdout: &stdout,
		Stderr: &stderr,
	})
	if code != exitOK {
		t.Fatalf("Run exit code = %d, want %d; stderr=%q", code, exitOK, stderr.String())
	}
	if gotPath != "/chat/completions" || stdout.String() != "ok\n" {
		t.Fatalf("path/stdout = %q/%q, want Fireworks compat response", gotPath, stdout.String())
	}
}

func TestSelectModelUsesOpenAIDefaultWhenAPIKeyMarksAllProviders(t *testing.T) {
	entries, registry, err := loadModelRegistry("")
	if err != nil {
		t.Fatalf("loadModelRegistry error = %v", err)
	}
	authState := map[string]bool{}
	markConfiguredProviders(authState, entries)

	selected, err := selectModel(registry, authState, cli.Request{})
	if err != nil {
		t.Fatalf("selectModel error = %v", err)
	}
	if selected.Provider != "openai" || selected.ID != "gpt-5.5" {
		t.Fatalf("selected = %s/%s, want openai/gpt-5.5", selected.Provider, selected.ID)
	}
}

func TestSelectModelUsesOpenAIDefaultForProviderOnly(t *testing.T) {
	_, registry, err := loadModelRegistry("")
	if err != nil {
		t.Fatalf("loadModelRegistry error = %v", err)
	}
	authState := map[string]bool{"openai": true}

	selected, err := selectModel(registry, authState, cli.Request{Provider: "openai"})
	if err != nil {
		t.Fatalf("selectModel error = %v", err)
	}
	if selected.Provider != "openai" || selected.ID != "gpt-5.5" {
		t.Fatalf("selected = %s/%s, want openai/gpt-5.5", selected.Provider, selected.ID)
	}
}
