package rpc

import "strings"

func (s *Server) enqueueSteering(text string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if strings.TrimSpace(text) != "" {
		s.steeringQueue = append(s.steeringQueue, text)
	}
}

func (s *Server) enqueueFollowUp(text string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if strings.TrimSpace(text) != "" {
		s.followUpQueue = append(s.followUpQueue, text)
	}
}

func (s *Server) mergeSteeringLocked(text string) string {
	if len(s.steeringQueue) == 0 {
		return text
	}
	count := len(s.steeringQueue)
	if s.steeringMode == "one-at-a-time" {
		count = 1
	}
	steering := strings.Join(s.steeringQueue[:count], "\n")
	s.steeringQueue = append([]string(nil), s.steeringQueue[count:]...)
	return steering + "\n\n" + text
}

func (s *Server) popFollowUpLocked() string {
	if len(s.followUpQueue) == 0 {
		return ""
	}
	count := len(s.followUpQueue)
	if s.followUpMode == "one-at-a-time" {
		count = 1
	}
	next := strings.Join(s.followUpQueue[:count], "\n")
	s.followUpQueue = append([]string(nil), s.followUpQueue[count:]...)
	return next
}

func (s *Server) setQueueMode(queue string, mode string) {
	if mode != "one-at-a-time" {
		mode = "all"
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if queue == "steering" {
		s.steeringMode = mode
		return
	}
	s.followUpMode = mode
}
