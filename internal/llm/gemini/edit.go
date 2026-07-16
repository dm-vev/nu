package gemini

import (
	"context"
	"fmt"

	"google.golang.org/genai"

	"github.com/dm-vev/nu/contracts"
)

// SupportsMultiTurnImageEditing returns true if the configured model supports
// multi-turn conversational image editing.
func (c *Client) SupportsMultiTurnImageEditing() bool {
	return SupportsMultiTurnImageEditing(c.model)
}

// CreateImageEditSession creates a new multi-turn image editing session.
// The session maintains conversation context for iterative image creation and modification.
func (c *Client) CreateImageEditSession(ctx context.Context, options *contracts.ImageEditSessionOptions) (contracts.ImageEditSession, error) {
	// Determine model to use
	model := c.model
	if options != nil && options.Model != "" {
		model = options.Model
	}

	// Validate model supports multi-turn image editing
	if !SupportsMultiTurnImageEditing(model) {
		// Try fallback to default image edit model
		if SupportsMultiTurnImageEditing(DefaultImageEditModel) {
			c.logger.Warn(ctx, "Model does not support multi-turn image editing, using default", map[string]interface{}{
				"requested_model": model,
				"fallback_model":  DefaultImageEditModel,
			})
			model = DefaultImageEditModel
		} else {
			return nil, fmt.Errorf("%w: model %s", contracts.ErrMultiTurnNotSupported, model)
		}
	}

	c.logger.Debug(ctx, "Creating image edit session", map[string]interface{}{
		"model":                  model,
		"has_system_instruction": options != nil && options.SystemInstruction != "",
	})

	// Build chat configuration with image response modalities
	config := &genai.GenerateContentConfig{
		ResponseModalities: []string{
			string(genai.ModalityText),
			string(genai.ModalityImage),
		},
	}

	// Add system instruction if provided
	if options != nil && options.SystemInstruction != "" {
		config.SystemInstruction = &genai.Content{
			Parts: []*genai.Part{
				{Text: options.SystemInstruction},
			},
		}
	}

	// Create chat session
	chat, err := c.genaiClient.Chats.Create(ctx, model, config, nil)
	if err != nil {
		c.logger.Error(ctx, "Failed to create image edit session", map[string]interface{}{
			"error": err.Error(),
			"model": model,
		})
		return nil, fmt.Errorf("failed to create image edit session: %w", err)
	}

	session := &ImageEditSession{
		client: c,
		chat:   chat,
		model:  model,
		logger: c.logger,
	}

	c.logger.Info(ctx, "Created image edit session", map[string]interface{}{
		"model": model,
	})

	return session, nil
}
