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
