package rpc

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/dm-vev/nu/contracts"
	"github.com/dm-vev/nu/internal/agentui"
	"github.com/dm-vev/nu/internal/testkit"
)

func TestNUF171RPCPromptResponseCorrelation(t *testing.T) {
	fake := testkit.NewScriptedAgent(
		contracts.AgentStreamEvent{Type: contracts.AgentEventContent, Content: "ok"},
		contracts.AgentStreamEvent{Type: contracts.AgentEventComplete},
	)
	var stdout bytes.Buffer
	server := NewServer(Options{
		Stdin:     strings.NewReader(`{"id":"p1","type":"prompt","message":"hello"}` + "\n" + `{"id":"q","type":"shutdown"}` + "\n"),
		Stdout:    &stdout,
		CWD:       "/tmp/nu",
		SessionID: "s1",
		Provider:  "test",
		API:       "test",
		Model:     "test",
	})
	server.SetAgent(agentui.New(agentui.Options{Runner: fake, Emit: server.Emit}))

	if err := server.Run(context.Background()); err != nil {
		t.Fatalf("Run error = %v", err)
	}

	records := decodeJSONLLines(t, stdout.String())
	requireResponse(t, records, "p1", "prompt", true)
	requireResponse(t, records, "q", "shutdown", true)
	requireEvent(t, records, "turn_end")
}

func TestNUF171RPCRejectsPromptDuringStreamWithoutBehavior(t *testing.T) {
	stdinReader, stdinWriter := io.Pipe()
	stdout := &lockedBuffer{}
	fake := &blockingRPCProvider{started: make(chan struct{})}
	server := NewServer(Options{
		Stdin:     stdinReader,
		Stdout:    stdout,
		CWD:       "/tmp/nu",
		SessionID: "s1",
	})
	server.SetAgent(agentui.New(agentui.Options{Runner: fake, Emit: server.Emit}))

	errCh := make(chan error, 1)
	go func() {
		errCh <- server.Run(context.Background())
	}()

	writeRPCLine(t, stdinWriter, `{"id":"p1","type":"prompt","message":"one"}`)
	waitForResponse(t, stdout, "p1", "prompt", true)
	<-fake.started

	writeRPCLine(t, stdinWriter, `{"id":"p2","type":"prompt","message":"two"}`)
	second := waitForResponse(t, stdout, "p2", "prompt", false)
	if !strings.Contains(second["error"].(string), "busy") {
		t.Fatalf("second error = %q, want busy", second["error"])
	}

	writeRPCLine(t, stdinWriter, `{"id":"q","type":"shutdown"}`)
	waitForResponse(t, stdout, "q", "shutdown", true)
	_ = stdinWriter.Close()
	if err := <-errCh; err != nil {
		t.Fatalf("Run error = %v", err)
	}
}

func TestNUF171RPCShutdownWritesFinalResponse(t *testing.T) {
	var stdout bytes.Buffer
	server := NewServer(Options{
		Stdin:  strings.NewReader(`{"id":"q","type":"shutdown"}` + "\n"),
		Stdout: &stdout,
	})

	if err := server.Run(context.Background()); err != nil {
		t.Fatalf("Run error = %v", err)
	}

	records := decodeJSONLLines(t, stdout.String())
	requireResponse(t, records, "q", "shutdown", true)
}

func TestNUF052SteeringDeliveredBeforeNextProviderCall(t *testing.T) {
	fake := testkit.NewScriptedAgent(
		contracts.AgentStreamEvent{Type: contracts.AgentEventContent, Content: "ok"},
		contracts.AgentStreamEvent{Type: contracts.AgentEventComplete},
	)
	var stdout bytes.Buffer
	server := NewServer(Options{
		Stdin: strings.NewReader(
			`{"id":"s","type":"steer","message":"remember constraints"}` + "\n" +
				`{"id":"p","type":"prompt","message":"implement"}` + "\n" +
				`{"id":"q","type":"shutdown"}` + "\n",
		),
		Stdout: &stdout,
	})
	server.SetAgent(agentui.New(agentui.Options{Runner: fake, Emit: server.Emit}))

	if err := server.Run(context.Background()); err != nil {
		t.Fatalf("Run error = %v", err)
	}

	prompts := fake.Prompts()
	if len(prompts) != 1 {
		t.Fatalf("agent prompts = %d, want 1", len(prompts))
	}
	got := prompts[0]
	if !strings.Contains(got, "remember constraints") || !strings.Contains(got, "implement") {
		t.Fatalf("prompt = %q, want steering and user prompt", got)
	}
}

func TestNUF171RPCRecognizesPiCommandSet(t *testing.T) {
	lines := []string{
		`{"id":"steer","type":"steer","message":"s"}`,
		`{"id":"follow_up","type":"follow_up","message":"f"}`,
		`{"id":"abort","type":"abort"}`,
		`{"id":"new_session","type":"new_session","parentSession":"p"}`,
		`{"id":"get_state","type":"get_state"}`,
		`{"id":"state","type":"state"}`,
		`{"id":"set_settings","type":"set_settings","settings":{"theme":"dark"},"persist":true}`,
		`{"id":"set_model","type":"set_model","provider":"test","modelId":"next"}`,
		`{"id":"cycle_model","type":"cycle_model"}`,
		`{"id":"get_available_models","type":"get_available_models"}`,
		`{"id":"set_thinking_level","type":"set_thinking_level","level":"medium"}`,
		`{"id":"cycle_thinking_level","type":"cycle_thinking_level"}`,
		`{"id":"set_steering_mode","type":"set_steering_mode","mode":"one-at-a-time"}`,
		`{"id":"set_follow_up_mode","type":"set_follow_up_mode","mode":"all"}`,
		`{"id":"compact","type":"compact","customInstructions":"short"}`,
		`{"id":"set_auto_compaction","type":"set_auto_compaction","enabled":true}`,
		`{"id":"set_auto_retry","type":"set_auto_retry","enabled":false}`,
		`{"id":"abort_retry","type":"abort_retry"}`,
		`{"id":"bash","type":"bash","command":"printf ok"}`,
		`{"id":"abort_bash","type":"abort_bash"}`,
		`{"id":"get_session_stats","type":"get_session_stats"}`,
		`{"id":"export_html","type":"export_html","outputPath":"out.html"}`,
		`{"id":"switch_session","type":"switch_session","sessionPath":"abc.jsonl"}`,
		`{"id":"fork","type":"fork","entryId":"m1"}`,
		`{"id":"clone","type":"clone"}`,
		`{"id":"get_fork_messages","type":"get_fork_messages"}`,
		`{"id":"get_entries","type":"get_entries"}`,
		`{"id":"get_tree","type":"get_tree"}`,
		`{"id":"get_last_assistant_text","type":"get_last_assistant_text"}`,
		`{"id":"set_session_name","type":"set_session_name","name":"demo"}`,
		`{"id":"get_messages","type":"get_messages"}`,
		`{"id":"get_commands","type":"get_commands"}`,
		`{"id":"shutdown","type":"shutdown"}`,
	}
	var stdout bytes.Buffer
	server := NewServer(Options{
		Stdin:  strings.NewReader(strings.Join(lines, "\n") + "\n"),
		Stdout: &stdout,
		CWD:    t.TempDir(),
	})
	runner := testkit.NewScriptedAgent()
	server.SetAgent(agentui.New(agentui.Options{
		Runner: runner,
		Builder: func(context.Context, agentui.Config, contracts.Memory) (contracts.StreamingAgent, error) {
			return runner, nil
		},
		Emit: server.Emit,
	}))

	if err := server.Run(context.Background()); err != nil {
		t.Fatalf("Run error = %v", err)
	}

	records := decodeJSONLLines(t, stdout.String())
	for _, line := range lines {
		var command map[string]any
		if err := json.Unmarshal([]byte(line), &command); err != nil {
			t.Fatalf("decode command fixture: %v", err)
		}
		id := command["id"].(string)
		response := findResponse(records, id)
		if response == nil {
			t.Fatalf("missing response for %s in %q", id, stdout.String())
		}
		if response["success"] != true {
			t.Fatalf("response for %s = %#v, want success", id, response)
		}
	}
}

func decodeJSONLLines(t *testing.T, raw string) []map[string]any {
	t.Helper()
	var records []map[string]any
	for _, line := range strings.Split(strings.TrimSpace(raw), "\n") {
		if strings.TrimSpace(line) == "" {
			continue
		}
		var record map[string]any
		if err := json.Unmarshal([]byte(line), &record); err != nil {
			t.Fatalf("decode JSONL line %q: %v", line, err)
		}
		records = append(records, record)
	}
	return records
}

func findResponse(records []map[string]any, id string) map[string]any {
	for _, record := range records {
		if record["type"] == "response" && record["id"] == id {
			return record
		}
	}
	return nil
}

func requireResponse(t *testing.T, records []map[string]any, id string, command string, success bool) {
	t.Helper()
	for _, record := range records {
		if record["type"] == "response" && record["id"] == id && record["command"] == command {
			if record["success"] != success {
				t.Fatalf("response %s success = %#v, want %v", id, record["success"], success)
			}
			return
		}
	}
	t.Fatalf("missing response id=%s command=%s in %#v", id, command, records)
}

func requireEvent(t *testing.T, records []map[string]any, eventType string) {
	t.Helper()
	for _, record := range records {
		if record["type"] == eventType {
			return
		}
	}
	t.Fatalf("missing event %s in %#v", eventType, records)
}

func writeRPCLine(t *testing.T, writer io.Writer, line string) {
	t.Helper()
	if _, err := io.WriteString(writer, line+"\n"); err != nil {
		t.Fatalf("write RPC line: %v", err)
	}
}

func waitForResponse(t *testing.T, stdout *lockedBuffer, id string, command string, success bool) map[string]any {
	t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		for _, record := range decodeJSONLLines(t, stdout.String()) {
			if record["type"] == "response" && record["id"] == id && record["command"] == command {
				if record["success"] != success {
					t.Fatalf("response %#v, want success %v", record, success)
				}
				return record
			}
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatalf("missing response id=%s command=%s in %q", id, command, stdout.String())
	return nil
}

type blockingRPCProvider struct {
	started chan struct{}
}

func (p *blockingRPCProvider) RunStream(ctx context.Context, _ string) (<-chan contracts.AgentStreamEvent, error) {
	select {
	case <-p.started:
	default:
		close(p.started)
	}
	ch := make(chan contracts.AgentStreamEvent)
	go func() {
		defer close(ch)
		<-ctx.Done()
	}()
	return ch, nil
}

type lockedBuffer struct {
	mu  sync.Mutex
	buf bytes.Buffer
}

func (b *lockedBuffer) Write(data []byte) (int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.buf.Write(data)
}

func (b *lockedBuffer) String() string {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.buf.String()
}
