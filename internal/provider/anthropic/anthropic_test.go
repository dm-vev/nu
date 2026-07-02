package anthropic

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"nu/internal/provider"
)

func TestNUF030AnthropicMessagesRequestShape(t *testing.T) {
	var gotPath string
	var gotKey string
	var gotVersion string
	var gotBody map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotKey = r.Header.Get("x-api-key")
		gotVersion = r.Header.Get("anthropic-version")
		if err := json.NewDecoder(r.Body).Decode(&gotBody); err != nil {
			t.Fatalf("Decode body error = %v", err)
		}
		w.Header().Set("Content-Type", "text/event-stream")
		_, _ = w.Write([]byte("event: message_start\n"))
		_, _ = w.Write([]byte("data: {\"type\":\"message_start\",\"message\":{\"content\":[]}}\n\n"))
		_, _ = w.Write([]byte("event: message_delta\n"))
		_, _ = w.Write([]byte("data: {\"type\":\"message_delta\",\"delta\":{\"stop_reason\":\"end_turn\"}}\n\n"))
		_, _ = w.Write([]byte("event: message_stop\n"))
		_, _ = w.Write([]byte("data: {\"type\":\"message_stop\"}\n\n"))
	}))
	defer server.Close()

	client := New(Config{BaseURL: server.URL, APIKey: "anthropic-key", Client: server.Client()})
	ch, err := client.Stream(context.Background(), provider.Request{
		Provider: "anthropic",
		API:      "messages",
		Model:    "claude-opus-4-8",
		Messages: []provider.Message{
			{Role: "user", Content: "hello"},
			{Role: "tool", ToolCallID: "toolu_1", Content: "ok"},
		},
	})
	if err != nil {
		t.Fatalf("Stream error = %v", err)
	}
	if _, err := provider.Collect(ch); err != nil {
		t.Fatalf("Collect error = %v", err)
	}
	if gotPath != "/v1/messages" || gotKey != "anthropic-key" || gotVersion == "" {
		t.Fatalf("path/key/version=%q/%q/%q", gotPath, gotKey, gotVersion)
	}
	if gotBody["model"] != "claude-opus-4-8" || gotBody["stream"] != true {
		t.Fatalf("body=%#v, want model and stream", gotBody)
	}
	messages := gotBody["messages"].([]any)
	toolMessage := messages[1].(map[string]any)
	if toolMessage["role"] != "user" {
		t.Fatalf("tool message=%#v, want user role", toolMessage)
	}
	content := toolMessage["content"].([]any)[0].(map[string]any)
	if content["type"] != "tool_result" || content["tool_use_id"] != "toolu_1" {
		t.Fatalf("tool result content=%#v", content)
	}
}

func TestAnthropicPayloadIncludesAssistantToolUse(t *testing.T) {
	payload, err := BuildMessagesPayload(provider.Request{
		Provider: "anthropic",
		API:      "messages",
		Model:    "claude-opus-4-8",
		Messages: []provider.Message{
			{Role: "user", Content: "read"},
			{Role: "assistant", ToolCallID: "toolu_1", Name: "read", Content: `{"path":"a.txt"}`},
			{Role: "tool", ToolCallID: "toolu_1", Content: "ok"},
		},
	})
	if err != nil {
		t.Fatalf("BuildMessagesPayload error = %v", err)
	}
	messages := payload["messages"].([]map[string]any)
	assistant := messages[1]
	content := assistant["content"].([]map[string]any)[0]
	input := content["input"].(map[string]any)
	if assistant["role"] != "assistant" ||
		content["type"] != "tool_use" ||
		content["id"] != "toolu_1" ||
		content["name"] != "read" ||
		input["path"] != "a.txt" {
		t.Fatalf("assistant tool content = %#v", content)
	}
}
