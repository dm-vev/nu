package generation

import (
	"time"

	"github.com/dm-vev/nu/contracts"
	"github.com/dm-vev/nu/internal/data/storage"
)

// Option configures a Tool.
type Option func(*Tool)

// WithMaxPromptLength sets the maximum prompt length.
func WithMaxPromptLength(maxLen int) Option {
	return func(t *Tool) {
		t.maxPromptLen = maxLen
	}
}

// WithDefaultAspectRatio sets the default aspect ratio.
func WithDefaultAspectRatio(ratio string) Option {
	return func(t *Tool) {
		t.defaultAspect = ratio
	}
}

// WithDefaultFormat sets the default output format.
func WithDefaultFormat(format string) Option {
	return func(t *Tool) {
		t.defaultFormat = format
	}
}

// WithMultiTurnEditor enables multi-turn image editing support.
// When enabled, the tool automatically manages sessions for iterative image refinement.
func WithMultiTurnEditor(editor contracts.MultiTurnImageEditor) Option {
	return func(t *Tool) {
		if editor != nil && editor.SupportsMultiTurnImageEditing() {
			t.multiTurnEditor = editor
			t.multiTurnEnabled = true
			t.sessions = make(map[string]*generationSessionEntry)
			// Start background cleanup
			go t.cleanupExpiredSessions()
		}
	}
}

// WithMultiTurnModel sets the model to use for multi-turn editing sessions
func WithMultiTurnModel(model string) Option {
	return func(t *Tool) {
		t.multiTurnModel = model
	}
}

// WithSessionTimeout sets how long sessions remain active without use
func WithSessionTimeout(timeout time.Duration) Option {
	return func(t *Tool) {
		t.sessionTimeout = timeout
	}
}

// WithMaxSessionsPerOrg limits concurrent sessions per organization
func WithMaxSessionsPerOrg(max int) Option {
	return func(t *Tool) {
		t.maxSessionsPerOrg = max
	}
}

// New creates an image generation tool.
func New(generator contracts.ImageGenerator, storage storage.Storage, options ...Option) *Tool {
	tool := &Tool{
		generator:         generator,
		storage:           storage,
		maxPromptLen:      2000,
		defaultAspect:     "1:1",
		defaultFormat:     "png",
		sessionTimeout:    30 * time.Minute,
		maxSessionsPerOrg: 10,
	}

	for _, opt := range options {
		opt(tool)
	}

	return tool
}
