package rpc

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"
)

type rpcMessage struct {
	ID      string `json:"id"`
	Role    string `json:"role"`
	Content string `json:"content"`
}

type commandEnvelope struct {
	ID                     string         `json:"id,omitempty"`
	Type                   string         `json:"type"`
	Message                string         `json:"message,omitempty"`
	StreamingBehavior      string         `json:"streaming_behavior,omitempty"`
	StreamingBehaviorCamel string         `json:"streamingBehavior,omitempty"`
	ParentSession          string         `json:"parent_session,omitempty"`
	ParentSessionCamel     string         `json:"parentSession,omitempty"`
	Provider               string         `json:"provider,omitempty"`
	ModelID                string         `json:"modelId,omitempty"`
	ModelIDSnake           string         `json:"model_id,omitempty"`
	Level                  string         `json:"level,omitempty"`
	Mode                   string         `json:"mode,omitempty"`
	CustomInstructions     string         `json:"customInstructions,omitempty"`
	CustomInstructionsAlt  string         `json:"custom_instructions,omitempty"`
	Enabled                bool           `json:"enabled,omitempty"`
	CommandText            string         `json:"command,omitempty"`
	ExcludeFromContext     bool           `json:"excludeFromContext,omitempty"`
	OutputPath             string         `json:"outputPath,omitempty"`
	SessionPath            string         `json:"sessionPath,omitempty"`
	EntryID                string         `json:"entryId,omitempty"`
	Since                  string         `json:"since,omitempty"`
	Name                   string         `json:"name,omitempty"`
	Settings               map[string]any `json:"settings,omitempty"`
	Persist                bool           `json:"persist,omitempty"`
}

type response map[string]any

// ReadLines reads LF-delimited JSONL records without Scanner's token ceiling.
func ReadLines(r io.Reader, onLine func(string) error) error {
	reader := bufio.NewReader(r)
	for {
		line, err := reader.ReadString('\n')
		if len(line) > 0 {
			// JSONL framing is LF-only; one CR before LF is accepted for clients using CRLF.
			line = strings.TrimSuffix(line, "\n")
			line = strings.TrimSuffix(line, "\r")
			if err := onLine(line); err != nil {
				return err
			}
		}
		if err == nil {
			continue
		}
		if err == io.EOF {
			return nil
		}
		return fmt.Errorf("read jsonl record: %w", err)
	}
}

func (s *Server) handleLine(ctx context.Context, line string) error {
	var command commandEnvelope
	if err := json.Unmarshal([]byte(line), &command); err != nil {
		return s.write(failure("", "parse", fmt.Sprintf("failed to parse command: %v", err)))
	}
	if command.Type == "extension_ui_response" {
		return nil
	}

	resp, ok := s.handleCommand(ctx, command)
	if !ok {
		return nil
	}
	return s.write(resp)
}

func success(id string, command string, data any) response {
	resp := response{"type": "response", "command": command, "success": true}
	if id != "" {
		resp["id"] = id
	}
	if data != nil {
		resp["data"] = data
	}
	return resp
}

func failure(id string, command string, message string) response {
	resp := response{"type": "response", "command": command, "success": false, "error": message}
	if id != "" {
		resp["id"] = id
	}
	return resp
}
