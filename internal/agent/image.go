package agent

import (
	"context"
	"fmt"
	"os"
	"time"

	"google.golang.org/genai"

	agentconfig "nu/internal/agent/config"
	"nu/internal/contracts"
	"nu/internal/data/storage"
	"nu/internal/llm/gemini"
	"nu/internal/telemetry"
	"nu/internal/tools/image"
)

// createImageGenerationToolFromConfig creates an image generation tool from YAML configuration
func createImageGenerationToolFromConfig(config *agentconfig.ImageGenerationYAML, logger telemetry.Logger) (contracts.Tool, error) {
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

	// Build Gemini client options
	var geminiOptions []gemini.Option
	geminiOptions = append(geminiOptions, gemini.WithModel(model))

	// Check for Vertex AI credentials first
	googleCreds := ""
	projectID := ""
	location := ""

	if config.Config != nil {
		if creds, ok := config.Config["google_application_credentials"].(string); ok {
			googleCreds = creds
		}
		if proj, ok := config.Config["project_id"].(string); ok {
			projectID = proj
		}
		if loc, ok := config.Config["location"].(string); ok {
			location = loc
		}
	}

	// Fall back to environment variables
	if googleCreds == "" {
		googleCreds = os.Getenv("VERTEX_AI_GOOGLE_APPLICATION_CREDENTIALS_CONTENT")
	}
	if projectID == "" {
		projectID = os.Getenv("VERTEX_AI_PROJECT")
	}
	if location == "" {
		location = os.Getenv("VERTEX_AI_REGION")
	}
	if location == "" {
		location = "us-central1"
	}

	// Use Vertex AI if credentials and project are available
	if googleCreds != "" && projectID != "" {
		// Parse credentials (supports base64 encoded, file path, or raw JSON)
		credentialsJSON, err := parseGoogleCredentials(googleCreds)
		if err != nil {
			return nil, fmt.Errorf("failed to parse Google credentials for image generation: %w", err)
		}

		geminiOptions = append(geminiOptions,
			gemini.WithBackend(genai.BackendVertexAI),
			gemini.WithCredentialsJSON([]byte(credentialsJSON)),
			gemini.WithProjectID(projectID),
			gemini.WithLocation(location),
		)
	} else {
		// Fall back to API key authentication
		apiKey := ""
		if config.Config != nil {
			if key, ok := config.Config["api_key"].(string); ok {
				apiKey = key
			}
		}
		if apiKey == "" {
			apiKey = os.Getenv("GEMINI_API_KEY")
		}
		if apiKey == "" {
			apiKey = os.Getenv("GOOGLE_API_KEY")
		}
		if apiKey == "" {
			return nil, fmt.Errorf("credentials required for image generation: set GEMINI_API_KEY or Vertex AI credentials (VERTEX_AI_PROJECT + VERTEX_AI_GOOGLE_APPLICATION_CREDENTIALS_CONTENT)")
		}
		geminiOptions = append(geminiOptions, gemini.WithAPIKey(apiKey))
	}

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
	var imageStorage storage.Storage
	if config.Storage != nil {
		imageStorage, err = createImageStorageFromConfig(config.Storage)
		if err != nil {
			if logger != nil {
				logger.Warn(ctx, "Failed to create image storage, images will be returned as base64", map[string]interface{}{
					"error": err.Error(),
				})
			}
			// Continue without storage - tool will return base64 data
		}
	}

	// Build tool options
	var toolOptions []image.GenerationOption

	// Apply config options
	if config.Config != nil {
		if maxLen, ok := config.Config["max_prompt_length"].(int); ok {
			toolOptions = append(toolOptions, image.WithGenerationMaxPromptLength(maxLen))
		}
		if ratio, ok := config.Config["default_aspect_ratio"].(string); ok {
			toolOptions = append(toolOptions, image.WithGenerationDefaultAspectRatio(ratio))
		}
		if format, ok := config.Config["default_format"].(string); ok {
			toolOptions = append(toolOptions, image.WithGenerationDefaultFormat(format))
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

			if googleCreds != "" && projectID != "" {
				credentialsJSON, _ := parseGoogleCredentials(googleCreds)
				mtOptions = append(mtOptions,
					gemini.WithBackend(genai.BackendVertexAI),
					gemini.WithCredentialsJSON([]byte(credentialsJSON)),
					gemini.WithProjectID(projectID),
					gemini.WithLocation(location),
				)
			} else {
				apiKey := ""
				if config.Config != nil {
					if key, ok := config.Config["api_key"].(string); ok {
						apiKey = key
					}
				}
				if apiKey == "" {
					apiKey = os.Getenv("GEMINI_API_KEY")
				}
				if apiKey == "" {
					apiKey = os.Getenv("GOOGLE_API_KEY")
				}
				if apiKey != "" {
					mtOptions = append(mtOptions, gemini.WithAPIKey(apiKey))
				}
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
				toolOptions = append(toolOptions, image.WithGenerationMultiTurnEditor(mtClient))
				toolOptions = append(toolOptions, image.WithGenerationMultiTurnModel(multiTurnModel))

				// Apply multi-turn specific options
				if config.MultiTurnEditing.SessionTimeout != "" {
					if timeout, err := time.ParseDuration(config.MultiTurnEditing.SessionTimeout); err == nil {
						toolOptions = append(toolOptions, image.WithGenerationSessionTimeout(timeout))
					}
				}
				if config.MultiTurnEditing.MaxSessionsPerOrg != nil {
					toolOptions = append(toolOptions, image.WithGenerationMaxSessionsPerOrg(*config.MultiTurnEditing.MaxSessionsPerOrg))
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
	imgTool := image.NewGeneration(geminiClient, imageStorage, toolOptions...)

	return imgTool, nil
}
