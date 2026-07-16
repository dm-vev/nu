package image

import (
	"time"

	"nu/internal/contracts"
	"nu/internal/data/storage"
)

// GenerationOption configures a GenerationTool.
type GenerationOption func(*GenerationTool)

// WithGenerationMaxPromptLength sets the maximum prompt length.
func WithGenerationMaxPromptLength(maxLen int) GenerationOption {
	return func(t *GenerationTool) {
		t.maxPromptLen = maxLen
	}
}

// WithGenerationDefaultAspectRatio sets the default aspect ratio.
func WithGenerationDefaultAspectRatio(ratio string) GenerationOption {
	return func(t *GenerationTool) {
		t.defaultAspect = ratio
	}
}

// WithGenerationDefaultFormat sets the default output format.
func WithGenerationDefaultFormat(format string) GenerationOption {
	return func(t *GenerationTool) {
		t.defaultFormat = format
	}
}

// WithMultiTurnEditor enables multi-turn image editing support.
// When enabled, the tool automatically manages sessions for iterative image refinement.
func WithGenerationMultiTurnEditor(editor contracts.MultiTurnImageEditor) GenerationOption {
	return func(t *GenerationTool) {
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
func WithGenerationMultiTurnModel(model string) GenerationOption {
	return func(t *GenerationTool) {
		t.multiTurnModel = model
	}
}

// WithSessionTimeout sets how long sessions remain active without use
func WithGenerationSessionTimeout(timeout time.Duration) GenerationOption {
	return func(t *GenerationTool) {
		t.sessionTimeout = timeout
	}
}

// WithMaxSessionsPerOrg limits concurrent sessions per organization
func WithGenerationMaxSessionsPerOrg(max int) GenerationOption {
	return func(t *GenerationTool) {
		t.maxSessionsPerOrg = max
	}
}

// NewGeneration creates an image generation tool.
func NewGeneration(generator contracts.ImageGenerator, storage storage.Storage, options ...GenerationOption) *GenerationTool {
	tool := &GenerationTool{
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
