package rpc

import (
	"context"
	"errors"
	"strings"

	"github.com/dm-vev/nu/internal/agentui"
)

// Emit forwards agent events to RPC stdout and updates lightweight state.
func (s *Server) Emit(ev agentui.Event) {
	s.recordEvent(ev)
	if err := s.write(ev); err != nil {
		s.setWriteErr(err)
	}
}

func (s *Server) startPrompt(ctx context.Context, id string, text string) response {
	text = strings.TrimSpace(text)
	if text == "" {
		return failure(id, "prompt", "message cannot be empty")
	}

	s.mu.Lock()
	a := s.agent
	if a == nil {
		s.mu.Unlock()
		return failure(id, "prompt", "rpc prompt requires agent handler")
	}
	if s.running {
		s.mu.Unlock()
		return failure(id, "prompt", "agent busy")
	}
	// Mark busy before the goroutine starts so immediate RPC commands see the active stream.
	s.running = true
	promptText := s.mergeSteeringLocked(text)
	s.addMessageLocked("user", promptText)
	s.mu.Unlock()

	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		err := a.Prompt(ctx, agentui.Prompt{Text: promptText})
		s.finishPrompt(ctx, err)
	}()
	return success(id, "prompt", nil)
}

func (s *Server) finishPrompt(ctx context.Context, err error) {
	var next string
	s.mu.Lock()
	shuttingDown := s.shutdown
	if err == nil && !shuttingDown && len(s.followUpQueue) > 0 {
		// Follow-ups run only after the active provider turn has released the agentui.
		next = s.popFollowUpLocked()
	}
	s.running = false
	s.mu.Unlock()

	if err != nil && !shuttingDown && !errors.Is(err, context.Canceled) {
		_ = s.write(map[string]any{"type": "prompt_error", "error": err.Error()})
	}
	if next != "" {
		_ = s.write(map[string]any{"type": "queue_update", "data": s.queueData()})
		_ = s.startPrompt(ctx, "", next)
	}
}

func (s *Server) recordEvent(ev agentui.Event) {
	if ev.Type != "turn_end" {
		return
	}
	text := turnEndText(ev.Data)
	if text == "" {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.addMessageLocked("assistant", text)
}

func turnEndText(data any) string {
	values, ok := data.(map[string]string)
	if ok {
		return values["text"]
	}
	generic, ok := data.(map[string]any)
	if ok {
		text, _ := generic["text"].(string)
		return text
	}
	return ""
}
