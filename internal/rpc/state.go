package rpc

func (s *Server) stateData() map[string]any {
	s.mu.Lock()
	defer s.mu.Unlock()
	return map[string]any{
		"session_id":                s.sessionID,
		"sessionId":                 s.sessionID,
		"session_name":              s.sessionName,
		"cwd":                       s.cwd,
		"model":                     s.modelDataLocked(),
		"provider":                  s.provider,
		"api":                       s.api,
		"busy":                      s.running,
		"isStreaming":               s.running,
		"isCompacting":              false,
		"thinkingLevel":             s.thinkingLevel,
		"steeringMode":              s.steeringMode,
		"followUpMode":              s.followUpMode,
		"autoCompactionEnabled":     s.autoCompaction,
		"autoRetryEnabled":          s.autoRetry,
		"messageCount":              len(s.messages),
		"pendingMessageCount":       len(s.followUpQueue),
		"steering_queue":            append([]string(nil), s.steeringQueue...),
		"follow_up_queue":           append([]string(nil), s.followUpQueue...),
		"active_leaf":               s.leafIDLocked(),
		"settings":                  cloneMap(s.settings),
		"stdout_protocol_exclusive": true,
	}
}

func (s *Server) settingsData() map[string]any {
	s.mu.Lock()
	defer s.mu.Unlock()
	return cloneMap(s.settings)
}

func (s *Server) modelData() map[string]any {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.modelDataLocked()
}

func (s *Server) modelDataLocked() map[string]any {
	return map[string]any{
		"provider":     s.provider,
		"api":          s.api,
		"id":           s.model,
		"modelId":      s.model,
		"display_name": s.modelLabel,
		"displayName":  s.modelLabel,
	}
}

func (s *Server) thinking() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.thinkingLevel
}

func (s *Server) queueData() map[string]any {
	s.mu.Lock()
	defer s.mu.Unlock()
	return map[string]any{
		"steering":  append([]string(nil), s.steeringQueue...),
		"follow_up": append([]string(nil), s.followUpQueue...),
	}
}

func (s *Server) sessionStats() map[string]any {
	s.mu.Lock()
	defer s.mu.Unlock()
	var userCount, assistantCount int
	for _, message := range s.messages {
		switch message.Role {
		case "user":
			userCount++
		case "assistant":
			assistantCount++
		}
	}
	return map[string]any{
		"sessionId":      s.sessionID,
		"messageCount":   len(s.messages),
		"userCount":      userCount,
		"assistantCount": assistantCount,
	}
}

func (s *Server) userMessages() []map[string]any {
	s.mu.Lock()
	defer s.mu.Unlock()
	var messages []map[string]any
	for _, message := range s.messages {
		if message.Role == "user" {
			messages = append(messages, map[string]any{"entryId": message.ID, "text": message.Content})
		}
	}
	return messages
}

func (s *Server) entriesSince(since string) []rpcMessage {
	s.mu.Lock()
	defer s.mu.Unlock()
	messages := append([]rpcMessage(nil), s.messages...)
	if since == "" {
		return messages
	}
	for i, message := range messages {
		if message.ID == since {
			return messages[i+1:]
		}
	}
	return nil
}

func (s *Server) treeData() []map[string]any {
	s.mu.Lock()
	defer s.mu.Unlock()
	tree := make([]map[string]any, 0, len(s.messages))
	for i, message := range s.messages {
		parent := ""
		if i > 0 {
			parent = s.messages[i-1].ID
		}
		tree = append(tree, map[string]any{
			"id":       message.ID,
			"parentId": parent,
			"role":     message.Role,
			"text":     message.Content,
		})
	}
	return tree
}

func (s *Server) leafID() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.leafIDLocked()
}

func (s *Server) leafIDLocked() string {
	if len(s.messages) == 0 {
		return ""
	}
	return s.messages[len(s.messages)-1].ID
}

func (s *Server) lastAssistantText() any {
	s.mu.Lock()
	defer s.mu.Unlock()
	for i := len(s.messages) - 1; i >= 0; i-- {
		if s.messages[i].Role == "assistant" {
			return s.messages[i].Content
		}
	}
	return nil
}

func (s *Server) messagesData() []rpcMessage {
	s.mu.Lock()
	defer s.mu.Unlock()
	return append([]rpcMessage(nil), s.messages...)
}

func cloneMap(values map[string]any) map[string]any {
	cloned := make(map[string]any, len(values))
	for key, value := range values {
		cloned[key] = value
	}
	return cloned
}
