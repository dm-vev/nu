package image

import (
	"context"
	"encoding/json"
	"fmt"
)

// Run executes the tool with the given input
func (t *EditTool) Run(ctx context.Context, input string) (string, error) {
	return t.Execute(ctx, input)
}

// Execute implements the tool execution
func (t *EditTool) Execute(ctx context.Context, args string) (string, error) {
	var params struct {
		Action      string `json:"action"`
		SessionID   string `json:"session_id,omitempty"`
		Prompt      string `json:"prompt,omitempty"`
		AspectRatio string `json:"aspect_ratio,omitempty"`
		ImageSize   string `json:"image_size,omitempty"`
	}

	if err := json.Unmarshal([]byte(args), &params); err != nil {
		return "", fmt.Errorf("failed to parse arguments: %w", err)
	}

	switch params.Action {
	case "start_session":
		return t.startSession(ctx, params.Prompt, params.AspectRatio, params.ImageSize)
	case "edit":
		if params.SessionID == "" {
			return "", fmt.Errorf("session_id is required for 'edit' action")
		}
		if params.Prompt == "" {
			return "", fmt.Errorf("prompt is required for 'edit' action")
		}
		return t.editImage(ctx, params.SessionID, params.Prompt, params.AspectRatio, params.ImageSize)
	case "end_session":
		if params.SessionID == "" {
			return "", fmt.Errorf("session_id is required for 'end_session' action")
		}
		return t.endSession(ctx, params.SessionID)
	default:
		return "", fmt.Errorf("unknown action: %s. Valid actions: start_session, edit, end_session", params.Action)
	}
}
