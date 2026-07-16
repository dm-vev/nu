package image

import (
	"context"
	"fmt"
	"time"

	"github.com/dm-vev/nu/agent/config"
	"github.com/dm-vev/nu/contracts"
	"github.com/dm-vev/nu/internal/llm/gemini"
	"github.com/dm-vev/nu/internal/tools/image/generation"
	"github.com/dm-vev/nu/telemetry"
)

// CreateGenerationTool creates an image generation tool from YAML configuration.
func CreateGenerationTool(config *config.ImageGenerationYAML, logger telemetry.Logger) (contracts.Tool, error) {
	if config == nil {
		return nil, nil
	}

	ctx := context.Background()

	// Determine provider (default to gemini)
	provider := config.Provider
	if provider == "" {
		provider = "gemini"
	}

	// Currently only Gemini is supported
	if provider != "gemini" {
		return nil, fmt.Errorf("unsupported image generation provider: %s (only 'gemini' is supported)", provider)
	}

	// Determine model (default to gemini-2.5-flash-image)
	model := config.Model
	if model == "" {
		model = gemini.Model25FlashImage
	}

	credentials := resolveCredentials(config)
	geminiOptions := []gemini.Option{gemini.WithModel(model)}
	authOptions, err := credentials.geminiOptions()
	if err != nil {
		return nil, err
	}
	geminiOptions = append(geminiOptions, authOptions...)

	// Create Gemini client for image generation
	geminiClient, err := gemini.NewClient(ctx, geminiOptions...)
	if err != nil {
		return nil, fmt.Errorf("failed to create Gemini client for image generation: %w", err)
	}

	// Verify the model supports image generation
	if !geminiClient.SupportsImageGeneration() {
		return nil, fmt.Errorf("model %s does not support image generation", model)
	}

	// Create storage backend
	imageStorage, err := createStorageFromConfig(config.Storage)
	if err != nil {
		if logger != nil {
			logger.Warn(ctx, "Failed to create image storage, images will be returned as base64", map[string]interface{}{
				"error": err.Error(),
			})
		}
		// Continue without storage - tool will return base64 data
	}

	// Build tool options
	var toolOptions []generation.Option

	// Apply config options
	if config.Config != nil {
		if maxLen, ok := config.Config["max_prompt_length"].(int); ok {
			toolOptions = append(toolOptions, generation.WithMaxPromptLength(maxLen))
		}
		if ratio, ok := config.Config["default_aspect_ratio"].(string); ok {
			toolOptions = append(toolOptions, generation.WithDefaultAspectRatio(ratio))
		}
		if format, ok := config.Config["default_format"].(string); ok {
			toolOptions = append(toolOptions, generation.WithDefaultFormat(format))
		}
	}

	// Check if multi-turn editing is enabled
	if config.MultiTurnEditing != nil {
		mtEnabled := config.MultiTurnEditing.Enabled == nil || *config.MultiTurnEditing.Enabled
		if mtEnabled {
			// Create a separate client for multi-turn editing (may use different model)
			multiTurnModel := config.MultiTurnEditing.Model
			if multiTurnModel == "" {
				multiTurnModel = model // Fall back to the same model
			}

			// Create multi-turn client options (same auth, different model)
			var mtOptions []gemini.Option
			mtOptions = append(mtOptions, gemini.WithModel(multiTurnModel))

			if authOptions, authErr := credentials.geminiOptions(); authErr == nil {
				mtOptions = append(mtOptions, authOptions...)
			}

			mtClient, err := gemini.NewClient(ctx, mtOptions...)
			if err != nil {
				if logger != nil {
					logger.Warn(ctx, "Failed to create multi-turn client, multi-turn editing disabled", map[string]interface{}{
						"error": err.Error(),
					})
				}
			} else if mtClient.SupportsMultiTurnImageEditing() {
				// Add multi-turn support to the tool
				toolOptions = append(toolOptions, generation.WithMultiTurnEditor(mtClient))
				toolOptions = append(toolOptions, generation.WithMultiTurnModel(multiTurnModel))

				// Apply multi-turn specific options
				if config.MultiTurnEditing.SessionTimeout != "" {
					if timeout, err := time.ParseDuration(config.MultiTurnEditing.SessionTimeout); err == nil {
						toolOptions = append(toolOptions, generation.WithSessionTimeout(timeout))
					}
				}
				if config.MultiTurnEditing.MaxSessionsPerOrg != nil {
					toolOptions = append(toolOptions, generation.WithMaxSessionsPerOrg(*config.MultiTurnEditing.MaxSessionsPerOrg))
				}

				if logger != nil {
					logger.Info(ctx, "Multi-turn image editing enabled", map[string]interface{}{
						"model": multiTurnModel,
					})
				}
			} else {
				if logger != nil {
					logger.Warn(ctx, "Model does not support multi-turn image editing", map[string]interface{}{
						"model": multiTurnModel,
					})
				}
			}
		}
	}

	// Create the image generation tool
	imgTool := generation.New(geminiClient, imageStorage, toolOptions...)

	return imgTool, nil
}
