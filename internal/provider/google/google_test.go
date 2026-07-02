package google

import (
	"testing"

	"nu/internal/provider"
)

func TestGoogleGenerateContentRequestShape(t *testing.T) {
	payload, err := BuildGenerateContentPayload(provider.Request{
		Provider: "google",
		API:      "generateContent",
		Model:    "gemini-3.5-flash",
		Messages: []provider.Message{
			{Role: "user", Content: "hello"},
			{Role: "assistant", Content: "hi"},
		},
	})
	if err != nil {
		t.Fatalf("BuildGenerateContentPayload error = %v", err)
	}
	contents := payload["contents"].([]map[string]any)
	if contents[0]["role"] != "user" || contents[1]["role"] != "model" {
		t.Fatalf("contents=%#v, want user/model roles", contents)
	}
	firstParts := contents[0]["parts"].([]map[string]string)
	if firstParts[0]["text"] != "hello" {
		t.Fatalf("parts=%#v, want hello text", firstParts)
	}
}

func TestGooglePayloadIncludesFunctionCallAndResponse(t *testing.T) {
	payload, err := BuildGenerateContentPayload(provider.Request{
		Provider: "google",
		API:      "generateContent",
		Model:    "gemini-3.5-flash",
		Messages: []provider.Message{
			{Role: "user", Content: "read"},
			{Role: "assistant", ToolCallID: "google-call-0", Name: "read", Content: `{"path":"a.txt"}`},
			{Role: "tool", ToolCallID: "google-call-0", Name: "read", Content: `{"content":"ok"}`},
		},
	})
	if err != nil {
		t.Fatalf("BuildGenerateContentPayload error = %v", err)
	}
	contents := payload["contents"].([]map[string]any)
	callPart := contents[1]["parts"].([]map[string]any)[0]
	functionCall := callPart["functionCall"].(map[string]any)
	callArgs := functionCall["args"].(map[string]any)
	responsePart := contents[2]["parts"].([]map[string]any)[0]
	functionResponse := responsePart["functionResponse"].(map[string]any)
	response := functionResponse["response"].(map[string]any)
	if contents[1]["role"] != "model" ||
		functionCall["name"] != "read" ||
		callArgs["path"] != "a.txt" ||
		contents[2]["role"] != "user" ||
		functionResponse["name"] != "read" ||
		response["content"] != "ok" {
		t.Fatalf("contents = %#v", contents)
	}
}
