package openai

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"nu/internal/provider"
)

func TestNUF030OpenAIChatRequestShape(t *testing.T) {
	var gotPath string
	var gotAuth string
	var gotBody map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotAuth = r.Header.Get("Authorization")
		if err := json.NewDecoder(r.Body).Decode(&gotBody); err != nil {
			t.Fatalf("Decode body error = %v", err)
		}
		w.Header().Set("Content-Type", "text/event-stream")
		_, _ = w.Write([]byte("data: {\"choices\":[{\"delta\":{\"content\":\"hi\"}}]}\n\n"))
		_, _ = w.Write([]byte("data: {\"choices\":[{\"finish_reason\":\"stop\"}]}\n\n"))
	}))
	defer server.Close()

	client := New(Config{BaseURL: server.URL, APIKey: "test-key", API: "chat", Client: server.Client()})
	ch, err := client.Stream(context.Background(), provider.Request{
		Provider: "openai",
		API:      "chat",
		Model:    "gpt-4.1",
		Messages: []provider.Message{{Role: "user", Content: "hello"}},
	})
	if err != nil {
		t.Fatalf("Stream error = %v", err)
	}
	events, err := provider.Collect(ch)
	if err != nil {
		t.Fatalf("Collect error = %v", err)
	}
	if gotPath != "/chat/completions" || gotAuth != "Bearer test-key" {
		t.Fatalf("path/auth = %q/%q, want chat endpoint and bearer auth", gotPath, gotAuth)
	}
	if gotBody["model"] != "gpt-4.1" || gotBody["stream"] != true {
		t.Fatalf("body=%#v, want model and stream", gotBody)
	}
	messages := gotBody["messages"].([]any)
	first := messages[0].(map[string]any)
	if first["role"] != "user" || first["content"] != "hello" {
		t.Fatalf("message=%#v, want user hello", first)
	}
	if events[len(events)-1].StopReason != "stop" {
		t.Fatalf("last event=%#v, want stop", events[len(events)-1])
	}
}

func TestNUF030OpenAIResponsesToolCallStream(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		_, _ = w.Write([]byte("event: response.output_item.added\n"))
		_, _ = w.Write([]byte("data: {\"output_index\":0,\"item\":{\"type\":\"function_call\",\"id\":\"fc_1\",\"call_id\":\"call_1\",\"name\":\"read\"}}\n\n"))
		_, _ = w.Write([]byte("event: response.function_call_arguments.delta\n"))
		_, _ = w.Write([]byte("data: {\"output_index\":0,\"delta\":\"{\\\"path\\\":\\\"a.txt\\\"}\"}\n\n"))
		_, _ = w.Write([]byte("event: response.output_item.done\n"))
		_, _ = w.Write([]byte("data: {\"output_index\":0,\"item\":{\"type\":\"function_call\"}}\n\n"))
		_, _ = w.Write([]byte("event: response.completed\n"))
		_, _ = w.Write([]byte("data: {\"response\":{\"status\":\"completed\"}}\n\n"))
	}))
	defer server.Close()

	client := New(Config{BaseURL: server.URL, APIKey: "test-key", API: "responses", Client: server.Client()})
	ch, err := client.Stream(context.Background(), provider.Request{
		Provider: "openai",
		API:      "responses",
		Model:    "gpt-5.5",
		Messages: []provider.Message{{Role: "user", Content: "read"}},
	})
	if err != nil {
		t.Fatalf("Stream error = %v", err)
	}
	events, err := provider.Collect(ch)
	if err != nil {
		t.Fatalf("Collect error = %v", err)
	}
	want := []provider.EventType{
		provider.EventStart,
		provider.EventToolCallStart,
		provider.EventToolCallDelta,
		provider.EventToolCallEnd,
		provider.EventDone,
	}
	if len(events) != len(want) {
		t.Fatalf("events=%#v, want %d events", events, len(want))
	}
	for i, typ := range want {
		if events[i].Type != typ {
			t.Fatalf("event[%d]=%#v, want %s", i, events[i], typ)
		}
	}
	if events[1].ToolCallID != "call_1" || events[1].ToolName != "read" || events[2].Delta != `{"path":"a.txt"}` {
		t.Fatalf("tool events=%#v", events)
	}
	if events[4].StopReason != "tool_use" {
		t.Fatalf("done=%#v, want tool_use", events[4])
	}
}

func TestOpenAIChatPayloadIncludesAssistantToolCalls(t *testing.T) {
	payload, err := BuildChatPayload(provider.Request{
		Provider: "openai",
		API:      "chat",
		Model:    "gpt-4.1",
		Messages: []provider.Message{
			{Role: "user", Content: "read"},
			{Role: "assistant", ToolCallID: "call-1", Name: "read", Content: `{"path":"a.txt"}`},
			{Role: "tool", ToolCallID: "call-1", Content: "ok"},
		},
	})
	if err != nil {
		t.Fatalf("BuildChatPayload error = %v", err)
	}
	messages := payload["messages"].([]map[string]any)
	assistant := messages[1]
	if assistant["role"] != "assistant" {
		t.Fatalf("assistant message = %#v", assistant)
	}
	toolCalls := assistant["tool_calls"].([]map[string]any)
	function := toolCalls[0]["function"].(map[string]string)
	if toolCalls[0]["id"] != "call-1" || function["name"] != "read" || function["arguments"] != `{"path":"a.txt"}` {
		t.Fatalf("tool calls = %#v", toolCalls)
	}
}

func TestOpenAIResponsesPayloadIncludesFunctionCallHistory(t *testing.T) {
	payload, err := BuildResponsesPayload(provider.Request{
		Provider: "openai",
		API:      "responses",
		Model:    "gpt-5.5",
		Messages: []provider.Message{
			{Role: "user", Content: "read"},
			{Role: "assistant", ToolCallID: "call-1", Name: "read", Content: `{"path":"a.txt"}`},
			{Role: "tool", ToolCallID: "call-1", Content: "ok"},
		},
	})
	if err != nil {
		t.Fatalf("BuildResponsesPayload error = %v", err)
	}
	input := payload["input"].([]map[string]any)
	call := input[1]
	result := input[2]
	if call["type"] != "function_call" ||
		call["call_id"] != "call-1" ||
		call["name"] != "read" ||
		call["arguments"] != `{"path":"a.txt"}` ||
		result["type"] != "function_call_output" ||
		result["call_id"] != "call-1" ||
		result["output"] != "ok" {
		t.Fatalf("input = %#v", input)
	}
}
