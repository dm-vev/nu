package bedrock

import (
	"encoding/binary"
	"hash/crc32"
	"net/http"
	"strings"
	"testing"
	"time"

	"nu/internal/provider"
)

func TestBedrockConverseRequestShape(t *testing.T) {
	payload, err := BuildConversePayload(provider.Request{
		Provider: "bedrock",
		API:      "converse-stream",
		Model:    "anthropic.claude-3-sonnet-20240229-v1:0",
		Messages: []provider.Message{{Role: "user", Content: "hello"}},
	})
	if err != nil {
		t.Fatalf("BuildConversePayload error = %v", err)
	}
	messages := payload["messages"].([]map[string]any)
	if messages[0]["role"] != "user" {
		t.Fatalf("messages=%#v, want user role", messages)
	}
	content := messages[0]["content"].([]map[string]string)
	if content[0]["text"] != "hello" {
		t.Fatalf("content=%#v, want hello text", content)
	}
}

func TestBedrockPayloadIncludesToolUseAndResult(t *testing.T) {
	payload, err := BuildConversePayload(provider.Request{
		Provider: "bedrock",
		API:      "converse-stream",
		Model:    "anthropic.claude-3-sonnet-20240229-v1:0",
		Messages: []provider.Message{
			{Role: "user", Content: "read"},
			{Role: "assistant", ToolCallID: "tooluse_1", Name: "read", Content: `{"path":"a.txt"}`},
			{Role: "tool", ToolCallID: "tooluse_1", Content: "ok"},
		},
	})
	if err != nil {
		t.Fatalf("BuildConversePayload error = %v", err)
	}
	messages := payload["messages"].([]map[string]any)
	toolUseContent := messages[1]["content"].([]map[string]any)[0]
	toolUse := toolUseContent["toolUse"].(map[string]any)
	input := toolUse["input"].(map[string]any)
	toolResultContent := messages[2]["content"].([]map[string]any)[0]
	toolResult := toolResultContent["toolResult"].(map[string]any)
	if messages[1]["role"] != "assistant" ||
		toolUse["toolUseId"] != "tooluse_1" ||
		toolUse["name"] != "read" ||
		input["path"] != "a.txt" ||
		messages[2]["role"] != "user" ||
		toolResult["toolUseId"] != "tooluse_1" {
		t.Fatalf("messages = %#v", messages)
	}
}

func TestBedrockSignAddsAuthorization(t *testing.T) {
	req, err := http.NewRequest(http.MethodPost, "https://bedrock-runtime.us-east-1.amazonaws.com/model/m/converse-stream", strings.NewReader("{}"))
	if err != nil {
		t.Fatalf("NewRequest error = %v", err)
	}
	creds := Credentials{AccessKeyID: "AKID", SecretAccessKey: "SECRET", Region: "us-east-1"}
	now := time.Date(2026, 7, 2, 12, 0, 0, 0, time.UTC)

	if err := Sign(req, []byte("{}"), creds, now); err != nil {
		t.Fatalf("Sign error = %v", err)
	}
	auth := req.Header.Get("Authorization")
	if !strings.Contains(auth, "AWS4-HMAC-SHA256") || !strings.Contains(auth, "Credential=AKID/20260702/us-east-1/bedrock/aws4_request") {
		t.Fatalf("Authorization=%q, want SigV4 credential scope", auth)
	}
	if req.Header.Get("X-Amz-Date") != "20260702T120000Z" {
		t.Fatalf("X-Amz-Date=%q", req.Header.Get("X-Amz-Date"))
	}
}

func TestBedrockRejectsOversizedEventFrame(t *testing.T) {
	frame := make([]byte, 12)
	binary.BigEndian.PutUint32(frame[0:4], maxEventStreamFrameBytes+1)
	binary.BigEndian.PutUint32(frame[4:8], 0)
	binary.BigEndian.PutUint32(frame[8:12], crc32.ChecksumIEEE(frame[0:8]))

	_, err := readEventStreamPayload(strings.NewReader(string(frame)))
	if err == nil || !strings.Contains(err.Error(), "too large") {
		t.Fatalf("readEventStreamPayload error = %v, want too large", err)
	}
}
