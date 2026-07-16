package image

import (
	"sync"
	"time"

	"nu/internal/contracts"
	"nu/internal/data/storage"
)

// EditTool implements multi-turn image editing as an agent tool.
// It manages sessions that allow iterative image creation and refinement
// through conversation, maintaining context between edits.
type EditTool struct {
	editor         contracts.MultiTurnImageEditor
	storage        storage.Storage
	sessions       map[string]*editSessionEntry
	sessionsMu     sync.RWMutex
	maxPromptLen   int
	sessionTimeout time.Duration
	maxSessions    int
	defaultModel   string
}

// NewEdit creates a multi-turn image editing tool.
func NewEdit(editor contracts.MultiTurnImageEditor, storage storage.Storage, options ...EditOption) *EditTool {
	tool := &EditTool{
		editor:         editor,
		storage:        storage,
		sessions:       make(map[string]*editSessionEntry),
		maxPromptLen:   2000,
		sessionTimeout: 30 * time.Minute,
		maxSessions:    10,
		defaultModel:   "", // Will use editor's default
	}

	for _, opt := range options {
		opt(tool)
	}

	// Start background cleanup goroutine
	go tool.cleanupExpiredSessions()

	return tool
}

// Name returns the tool name
func (t *EditTool) Name() string {
	return "edit_image"
}

// DisplayName returns a human-friendly name
func (t *EditTool) DisplayName() string {
	return "Image Editor"
}

// Description returns what the tool does
func (t *EditTool) Description() string {
	return `Multi-turn image editing tool for iterative image creation and refinement through conversation.

Actions:
- start_session: Begin a new editing session, optionally with an initial prompt to generate the first image
- edit: Send a modification request to an existing session (requires session_id)
- end_session: Close a session when done (requires session_id)

The session maintains conversation context, allowing you to progressively refine images based on previous results.`
}

// Internal returns false as this is a user-visible tool
func (t *EditTool) Internal() bool {
	return false
}

// Parameters returns the tool's parameter specifications
func (t *EditTool) Parameters() map[string]contracts.ParameterSpec {
	return map[string]contracts.ParameterSpec{
		"action": {
			Type:        "string",
			Description: "The action to perform: start_session (create new session), edit (modify existing image), or end_session (close session)",
			Required:    true,
			Enum:        []interface{}{"start_session", "edit", "end_session"},
		},
		"session_id": {
			Type:        "string",
			Description: "Session ID (required for 'edit' and 'end_session' actions). Returned when starting a new session.",
			Required:    false,
		},
		"prompt": {
			Type:        "string",
			Description: "Text description for image generation or modification request. Required for 'start_session' (initial image) and 'edit' actions.",
			Required:    false,
		},
		"aspect_ratio": {
			Type:        "string",
			Description: "Output image aspect ratio",
			Required:    false,
			Default:     "1:1",
			Enum:        []interface{}{"1:1", "2:3", "3:2", "16:9", "21:9"},
		},
		"image_size": {
			Type:        "string",
			Description: "Output image resolution",
			Required:    false,
			Default:     "1K",
			Enum:        []interface{}{"1K", "2K", "4K"},
		},
	}
}
