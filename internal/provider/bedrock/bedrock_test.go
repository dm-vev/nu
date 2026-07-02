package bedrock

import (
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
