package image

import "time"

// EditOption configures an EditTool.
type EditOption func(*EditTool)

// WithEditMaxPromptLength sets the maximum prompt length.
func WithEditMaxPromptLength(maxLen int) EditOption {
	return func(t *EditTool) {
		t.maxPromptLen = maxLen
	}
}

// WithEditSessionTimeout sets the session timeout duration.
func WithEditSessionTimeout(timeout time.Duration) EditOption {
	return func(t *EditTool) {
		t.sessionTimeout = timeout
	}
}

// WithEditMaxSessions sets the maximum number of concurrent sessions.
func WithEditMaxSessions(max int) EditOption {
	return func(t *EditTool) {
		t.maxSessions = max
	}
}

// WithEditDefaultModel sets the default model for new sessions.
func WithEditDefaultModel(model string) EditOption {
	return func(t *EditTool) {
		t.defaultModel = model
	}
}
