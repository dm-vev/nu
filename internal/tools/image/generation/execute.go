package generation

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/dm-vev/nu/contracts"
	"github.com/dm-vev/nu/internal/data/storage"
)

// Run executes the tool with the given input
func (t *Tool) Run(ctx context.Context, input string) (string, error) {
	return t.Execute(ctx, input)
}

// Execute implements the tool execution
func (t *Tool) Execute(ctx context.Context, args string) (string, error) {
	// Parse arguments
	var params struct {
		Prompt       string `json:"prompt"`
		Action       string `json:"action,omitempty"`
		AspectRatio  string `json:"aspect_ratio,omitempty"`
		OutputFormat string `json:"output_format,omitempty"`
		ImageSize    string `json:"image_size,omitempty"`
	}

	if err := json.Unmarshal([]byte(args), &params); err != nil {
		return "", fmt.Errorf("failed to parse arguments: %w", err)
	}

	// Set defaults
	if params.Action == "" {
		params.Action = "generate"
	}
	if params.AspectRatio == "" {
		params.AspectRatio = t.defaultAspect
	}

	// Use multi-turn editing if enabled
	if t.multiTurnEnabled {
		return t.executeMultiTurn(ctx, params.Action, params.Prompt, params.AspectRatio, params.ImageSize)
	}

	// Standard single-shot generation
	return t.executeSingleShot(ctx, params.Prompt, params.AspectRatio, params.OutputFormat)
}

// executeSingleShot performs standard one-shot image generation
func (t *Tool) executeSingleShot(ctx context.Context, prompt, aspectRatio, outputFormat string) (string, error) {
	// Validate prompt
	if prompt == "" {
		return "", fmt.Errorf("prompt is required")
	}

	if len(prompt) > t.maxPromptLen {
		return "", fmt.Errorf("prompt exceeds maximum length of %d characters", t.maxPromptLen)
	}

	// Set defaults
	if outputFormat == "" {
		outputFormat = t.defaultFormat
	}

	// Build request
	request := contracts.ImageGenerationRequest{
		Prompt: prompt,
		Options: &contracts.ImageGenerationOptions{
			NumberOfImages: 1,
			AspectRatio:    aspectRatio,
			OutputFormat:   outputFormat,
		},
	}

	// Generate image
	response, err := t.generator.GenerateImage(ctx, request)
	if err != nil {
		return "", fmt.Errorf("image generation failed: %w", err)
	}

	if len(response.Images) == 0 {
		return "", fmt.Errorf("no images were generated")
	}

	// Store image if storage is configured
	if t.storage != nil {
		metadata := storage.Metadata{
			Prompt:    prompt,
			CreatedAt: time.Now(),
		}

		url, err := t.storage.Store(ctx, &response.Images[0], metadata)
		if err != nil {
			// Log warning but don't fail - return base64 instead
			fmt.Printf("[imagegen] Storage failed, using base64: %v\n", err)
			return t.formatResultWithBase64(response, prompt), nil
		}
		response.Images[0].URL = url
		fmt.Printf("[imagegen] Image stored at: %s\n", url)
		// Format result with URL
		return t.formatResult(response, prompt, url), nil
	}

	// No storage configured - return base64 embedded image
	fmt.Printf("[imagegen] No storage configured, using base64\n")
	return t.formatResultWithBase64(response, prompt), nil
}
