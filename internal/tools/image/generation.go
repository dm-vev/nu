package image

import (
	"sync"
	"time"

	"nu/internal/contracts"
	"nu/internal/data/storage"
)

// GenerationTool implements image generation as a tool for agents.
// When multi-turn editing is enabled, it automatically manages sessions
// to allow iterative image refinement through conversation.
type GenerationTool struct {
	generator     contracts.ImageGenerator
	storage       storage.Storage
	maxPromptLen  int
	defaultAspect string
	defaultFormat string

	// Multi-turn editing support
	multiTurnEditor   contracts.MultiTurnImageEditor
	multiTurnEnabled  bool
	multiTurnModel    string
	sessions          map[string]*generationSessionEntry
	sessionsMu        sync.RWMutex
	sessionTimeout    time.Duration
	maxSessionsPerOrg int
}

// Name returns the tool name
func (t *GenerationTool) Name() string {
	return "generate_image"
}

// DisplayName returns a human-friendly name
func (t *GenerationTool) DisplayName() string {
	return "Image Generator"
}

// Description returns what the tool does
func (t *GenerationTool) Description() string {
	if t.multiTurnEnabled {
		return `Generate and edit images through conversation. Supports iterative refinement.

Actions:
- generate: Create a new image (starts a session automatically for multi-turn editing)
- edit: Modify the current image in the active session
- end_session: Close the current editing session

The tool automatically maintains a session per conversation, allowing you to refine images iteratively.`
	}
	return "Generate images from text descriptions using AI. Provide a detailed prompt describing the image you want to create. Returns the URL of the generated image."
}

// Internal returns false as this is a user-visible tool
func (t *GenerationTool) Internal() bool {
	return false
}

// Parameters returns the tool's parameter specifications
func (t *GenerationTool) Parameters() map[string]contracts.ParameterSpec {
	params := map[string]contracts.ParameterSpec{
		"prompt": {
			Type:        "string",
			Description: "A detailed text description of the image to generate or the modification to apply.",
			Required:    true,
		},
		"aspect_ratio": {
			Type:        "string",
			Description: "The aspect ratio of the output image",
			Required:    false,
			Default:     t.defaultAspect,
			Enum:        []interface{}{"1:1", "16:9", "9:16", "4:3", "3:4", "2:3", "3:2", "21:9"},
		},
	}

	// Add multi-turn specific parameters
	if t.multiTurnEnabled {
		params["action"] = contracts.ParameterSpec{
			Type:        "string",
			Description: "The action to perform: 'generate' creates a new image (default), 'edit' modifies the current image, 'end_session' closes the editing session",
			Required:    false,
			Default:     "generate",
			Enum:        []interface{}{"generate", "edit", "end_session"},
		}
		params["image_size"] = contracts.ParameterSpec{
			Type:        "string",
			Description: "Output image resolution (for multi-turn editing)",
			Required:    false,
			Default:     "1K",
			Enum:        []interface{}{"1K", "2K", "4K"},
		}
	} else {
		params["output_format"] = contracts.ParameterSpec{
			Type:        "string",
			Description: "The output image format",
			Required:    false,
			Default:     t.defaultFormat,
			Enum:        []interface{}{"png", "jpeg"},
		}
	}

	return params
}
