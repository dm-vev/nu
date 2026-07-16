package gemini

import (
	"context"
	"encoding/base64"
	"fmt"

	"google.golang.org/genai"

	"nu/internal/contracts"
)

// SupportsImageGeneration returns true if the configured model supports image generation
func (c *Client) SupportsImageGeneration() bool {
	return SupportsImageGeneration(c.model)
}

// SupportedImageFormats returns the supported output formats for the configured model
func (c *Client) SupportedImageFormats() []string {
	formats := GetSupportedOutputFormats(c.model)
	if len(formats) == 0 {
		return []string{"png", "jpeg"}
	}
	return formats
}

// GenerateImage generates images from a text prompt using Gemini
func (c *Client) GenerateImage(ctx context.Context, request contracts.ImageGenerationRequest) (*contracts.ImageGenerationResponse, error) {
	// Validate model supports image generation
	if !c.SupportsImageGeneration() {
		return nil, fmt.Errorf("%w: model %s", contracts.ErrImageGenerationNotSupported, c.model)
	}

	// Validate prompt
	if request.Prompt == "" {
		return nil, contracts.ErrInvalidPrompt
	}

	// Apply defaults if options not provided
	opts := request.Options
	if opts == nil {
		opts = contracts.DefaultImageGenerationOptions()
	}
	if opts.NumberOfImages <= 0 {
		opts.NumberOfImages = 1
	}
	if opts.OutputFormat == "" {
		opts.OutputFormat = "png"
	}

	// Build the request parts
	parts := []*genai.Part{
		genai.NewPartFromText(request.Prompt),
	}

	// Add reference image if provided (for image-to-image generation)
	if request.ReferenceImage != nil {
		imageData := request.ReferenceImage.Data
		if imageData == nil && request.ReferenceImage.Base64 != "" {
			// Decode base64
			decoded, err := base64.StdEncoding.DecodeString(request.ReferenceImage.Base64)
			if err != nil {
				return nil, fmt.Errorf("failed to decode reference image: %w", err)
			}
			imageData = decoded
		}

		if len(imageData) > 0 {
			mimeType := request.ReferenceImage.MimeType
			if mimeType == "" {
				mimeType = "image/png"
			}
			parts = append(parts, &genai.Part{
				InlineData: &genai.Blob{
					Data:     imageData,
					MIMEType: mimeType,
				},
			})
		}
	}

	// Configure generation for image output
	config := &genai.GenerateContentConfig{
		ResponseModalities: []string{
			string(genai.ModalityImage),
		},
	}

	// Execute generation
	result, err := c.genaiClient.Models.GenerateContent(
		ctx,
		c.model,
		[]*genai.Content{{Role: "user", Parts: parts}},
		config,
	)
	if err != nil {
		// Check for specific error types
		if geminiIsContentBlockedError(err) {
			return nil, fmt.Errorf("%w: %v", contracts.ErrContentBlocked, err)
		}
		if geminiIsRateLimitError(err) {
			return nil, fmt.Errorf("%w: %v", contracts.ErrRateLimitExceeded, err)
		}
		return nil, fmt.Errorf("image generation failed: %w", err)
	}

	// Parse response
	return c.parseImageResponse(result, opts.OutputFormat)
}

// parseImageResponse extracts generated images from the API response
func (c *Client) parseImageResponse(result *genai.GenerateContentResponse, outputFormat string) (*contracts.ImageGenerationResponse, error) {
	response := &contracts.ImageGenerationResponse{
		Images:   make([]contracts.GeneratedImage, 0),
		Metadata: make(map[string]interface{}),
	}

	if result == nil || len(result.Candidates) == 0 {
		return nil, fmt.Errorf("no candidates in response")
	}

	for _, candidate := range result.Candidates {
		if candidate.Content == nil {
			continue
		}

		for _, part := range candidate.Content.Parts {
			// Check for text response (model might return text instead of image)
			if part.Text != "" {
				response.Metadata["text_response"] = part.Text
			}
			if part.InlineData != nil && part.InlineData.Data != nil {
				mimeType := part.InlineData.MIMEType
				if mimeType == "" {
					mimeType = "image/" + outputFormat
				}

				image := contracts.GeneratedImage{
					Data:     part.InlineData.Data,
					Base64:   base64.StdEncoding.EncodeToString(part.InlineData.Data),
					MimeType: mimeType,
				}

				// Extract finish reason if available
				if candidate.FinishReason != "" {
					image.FinishReason = string(candidate.FinishReason)
				}

				response.Images = append(response.Images, image)
			}
		}
	}

	// Extract usage if available
	if result.UsageMetadata != nil {
		response.Usage = &contracts.ImageUsage{
			InputTokens:     int(result.UsageMetadata.PromptTokenCount),
			OutputTokens:    int(result.UsageMetadata.CandidatesTokenCount),
			ImagesGenerated: len(response.Images),
		}
	}

	// Store model info in metadata
	response.Metadata["model"] = c.model

	if len(response.Images) == 0 {
		// Include any text response in the error for debugging
		if textResp, ok := response.Metadata["text_response"].(string); ok {
			return nil, fmt.Errorf("no images generated in response, model returned text: %s", textResp)
		}
		return nil, fmt.Errorf("no images generated in response")
	}

	return response, nil
}

// geminiIsContentBlockedError checks if the error is due to content being blocked
func geminiIsContentBlockedError(err error) bool {
	if err == nil {
		return false
	}
	errStr := err.Error()
	return geminiContains(errStr, "blocked") ||
		geminiContains(errStr, "safety") ||
		geminiContains(errStr, "SAFETY") ||
		geminiContains(errStr, "BLOCKED")
}

// geminiIsRateLimitError checks if the error is due to rate limiting
func geminiIsRateLimitError(err error) bool {
	if err == nil {
		return false
	}
	errStr := err.Error()
	return geminiContains(errStr, "rate limit") ||
		geminiContains(errStr, "quota") ||
		geminiContains(errStr, "RESOURCE_EXHAUSTED") ||
		geminiContains(errStr, "429")
}

// geminiContains checks if s geminiContains substr (case-insensitive would be better but keeping it simple)
func geminiContains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && geminiContainsSubstr(s, substr))
}

func geminiContainsSubstr(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
