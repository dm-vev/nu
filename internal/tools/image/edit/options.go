package edit

import "time"

// Option configures an Tool.
type Option func(*Tool)

// WithMaxPromptLength sets the maximum prompt length.
func WithMaxPromptLength(maxLen int) Option {
	return func(t *Tool) {
		t.maxPromptLen = maxLen
	}
}

// WithSessionTimeout sets the session timeout duration.
func WithSessionTimeout(timeout time.Duration) Option {
	return func(t *Tool) {
		t.sessionTimeout = timeout
	}
}

// WithMaxSessions sets the maximum number of concurrent sessions.
func WithMaxSessions(max int) Option {
	return func(t *Tool) {
		t.maxSessions = max
	}
}

// WithDefaultModel sets the default model for new sessions.
func WithDefaultModel(model string) Option {
	return func(t *Tool) {
		t.defaultModel = model
	}
}
