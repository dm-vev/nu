package rpc

import (
	"fmt"
	"path/filepath"
	"strings"
)

func (s *Server) addMessageLocked(role string, text string) {
	s.nextMessageIndex++
	s.messages = append(s.messages, rpcMessage{
		ID:      fmt.Sprintf("m%d", s.nextMessageIndex),
		Role:    role,
		Content: text,
	})
}

func (s *Server) newSession(parent string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.sessionID = newID("session")
	s.messages = nil
	s.nextMessageIndex = 0
	if parent != "" {
		s.settings["parent_session"] = parent
	}
}

func (s *Server) currentSessionID() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.sessionID
}

func (s *Server) switchSession(path string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if strings.TrimSpace(path) == "" {
		s.sessionID = newID("session")
		return
	}
	s.sessionID = strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
}

func (s *Server) setSessionName(name string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.sessionName = strings.TrimSpace(name)
}

func (s *Server) setSettings(settings map[string]any, persist bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for key, value := range settings {
		s.settings[key] = value
	}
	s.settings["persist"] = persist
}

func (s *Server) setModel(providerID string, modelID string) error {
	s.mu.Lock()
	providerID = firstNonEmpty(providerID, s.provider)
	api := firstNonEmpty(s.api, "chat")
	a := s.agent
	s.mu.Unlock()

	if a != nil {
		if err := a.SetModel(providerID, api, modelID); err != nil {
			return err
		}
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	s.provider = providerID
	s.api = api
	s.model = modelID
	s.modelLabel = modelID
	return nil
}

func (s *Server) setThinking(level string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.thinkingLevel = firstNonEmpty(level, "off")
}

func (s *Server) cycleThinking() string {
	levels := []string{"off", "minimal", "low", "medium", "high", "xhigh"}
	s.mu.Lock()
	defer s.mu.Unlock()
	for i, level := range levels {
		if s.thinkingLevel == level {
			s.thinkingLevel = levels[(i+1)%len(levels)]
			return s.thinkingLevel
		}
	}
	s.thinkingLevel = "off"
	return s.thinkingLevel
}

func (s *Server) setAutoCompaction(enabled bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.autoCompaction = enabled
}

func (s *Server) setAutoRetry(enabled bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.autoRetry = enabled
}
