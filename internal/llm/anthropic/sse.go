package anthropic

import (
	"encoding/json"
)

// AnthropicSSEEvent represents the structure of Anthropic's SSE events
type SSEEvent struct {
	Type string          `json:"type"`
	Data json.RawMessage `json:"data,omitempty"`
}

// MessageStartData represents the message_start event data.
type MessageStartData struct {
	Type    string `json:"type"`
	Message struct {
		ID      string `json:"id"`
		Role    string `json:"role"`
		Content []any  `json:"content"`
		Model   string `json:"model"`
		Usage   Usage  `json:"usage"`
	} `json:"message"`
}

// ContentBlockStart event data
type ContentBlockStartData struct {
	Type         string `json:"type"`
	Index        int    `json:"index"`
	ContentBlock struct {
		Type  string                 `json:"type"`
		Text  string                 `json:"text,omitempty"`
		ID    string                 `json:"id,omitempty"`
		Name  string                 `json:"name,omitempty"`
		Input map[string]interface{} `json:"input,omitempty"`
	} `json:"content_block"`
}

// ContentBlockDelta event data
type ContentBlockDeltaData struct {
	Type  string `json:"type"`
	Index int    `json:"index"`
	Delta struct {
		Type        string `json:"type"`
		Text        string `json:"text,omitempty"`
		Thinking    string `json:"thinking,omitempty"`
		PartialJSON string `json:"partial_json,omitempty"`
	} `json:"delta"`
}

// ContentBlockStop event data
type ContentBlockStopData struct {
	Type  string `json:"type"`
	Index int    `json:"index"`
}

// MessageDelta event data
type MessageDeltaData struct {
	Type  string `json:"type"`
	Delta struct {
		StopReason   string `json:"stop_reason,omitempty"`
		StopSequence string `json:"stop_sequence,omitempty"`
	} `json:"delta"`
	Usage Usage `json:"usage"`
}

// MessageStop event data
type MessageStopData struct {
	Type string `json:"type"`
}
