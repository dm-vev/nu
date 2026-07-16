package generation

import (
	"context"
	"fmt"
	"time"

	"github.com/dm-vev/nu/contracts"
	"github.com/dm-vev/nu/internal/data/storage"
	"github.com/dm-vev/nu/internal/multitenancy"
)

// formatMultiTurnResponse formats the response from a multi-turn editing session
func (t *Tool) formatMultiTurnResponse(ctx context.Context, resp *contracts.ImageEditResponse, prompt string, isInitial bool) (string, error) {
	var result string

	if isInitial {
		result = "Image generated successfully.\n\n"
	} else {
		result = "Image edited successfully.\n\n"
	}

	// Add text response if present
	if resp.Text != "" {
		result += fmt.Sprintf("Model: %s\n\n", resp.Text)
	}

	// Process images
	if len(resp.Images) > 0 {
		for i, image := range resp.Images {
			// Try to store image if storage is configured
			if t.storage != nil {
				metadata := storage.Metadata{
					Prompt:    prompt,
					CreatedAt: time.Now(),
				}

				// Get org ID from context if available
				if orgID, err := multitenancy.GetOrgID(ctx); err == nil {
					metadata.OrgID = orgID
				}

				url, err := t.storage.Store(ctx, &image, metadata)
				if err != nil {
					// Log warning but continue with base64
					fmt.Printf("[imagegen] Storage failed, using base64: %v\n", err)
					result += t.formatImageBase64(&image, i)
				} else {
					result += t.formatImageURL(url, &image, i)
				}
			} else {
				// No storage configured - use base64
				result += t.formatImageBase64(&image, i)
			}
		}
	} else {
		result += "No image was generated in this turn.\n"
	}

	// Add usage info
	if resp.Usage != nil {
		result += fmt.Sprintf("\nTokens used: %d input, %d output\n", resp.Usage.InputTokens, resp.Usage.OutputTokens)
	}

	result += "\nYou can continue editing this image with action='edit', or use action='end_session' when done."

	return result, nil
}

// formatImageURL formats an image with its URL
func (t *Tool) formatImageURL(url string, image *contracts.GeneratedImage, index int) string {
	result := ""
	if index > 0 {
		result = fmt.Sprintf("\n--- Image %d ---\n", index+1)
	}
	// Use markdown image syntax for UI rendering
	result += fmt.Sprintf("![Generated image](%s)\n\n", url)
	result += fmt.Sprintf("Format: %s\n", image.MimeType)
	result += fmt.Sprintf("Size: %d bytes\n", len(image.Data))
	return result
}

// formatImageBase64 formats an image as base64 data URI
// For large images, it skips embedding to avoid token limits and UI performance issues
func (t *Tool) formatImageBase64(image *contracts.GeneratedImage, index int) string {
	result := ""
	if index > 0 {
		result = fmt.Sprintf("\n--- Image %d ---\n", index+1)
	}

	// Check image size - if too large, don't embed full base64 to avoid:
	// 1. Token limits in LLM conversation history
	// 2. UI performance issues (re-rendering large base64 on each keystroke)
	// A typical 1K image is ~500KB base64
	const maxBase64Size = 50000 // ~50KB limit - be conservative for UI performance
	if len(image.Base64) > maxBase64Size {
		// Image too large for embedding - provide info but not the data
		result += "[Image generated successfully]\n\n"
		result += fmt.Sprintf("Format: %s\n", image.MimeType)
		result += fmt.Sprintf("Size: %d bytes (%.1f KB)\n", len(image.Data), float64(len(image.Data))/1024)
		result += "\nNote: Image was generated but is too large to display inline.\n"
		result += "Configure GCS storage in your agents.yaml to get shareable image URLs.\n"
		return result
	}

	// Create data URI for direct embedding in markdown
	dataURI := fmt.Sprintf("data:%s;base64,%s", image.MimeType, image.Base64)
	result += fmt.Sprintf("![Generated image](%s)\n\n", dataURI)
	result += fmt.Sprintf("Format: %s\n", image.MimeType)
	result += fmt.Sprintf("Size: %d bytes\n", len(image.Data))
	return result
}

// formatResult creates a human-readable result string with URL
// The image is formatted using markdown syntax so UIs can render it
func (t *Tool) formatResult(response *contracts.ImageGenerationResponse, prompt, imageURL string) string {
	result := fmt.Sprintf("Successfully generated image for prompt: \"%s\"\n\n", truncateString(prompt, 100))

	if imageURL != "" {
		// Use markdown image syntax for UI rendering
		result += fmt.Sprintf("![Generated image](%s)\n\n", imageURL)
	}

	result += fmt.Sprintf("Format: %s\n", response.Images[0].MimeType)
	result += fmt.Sprintf("Size: %d bytes\n", len(response.Images[0].Data))

	if response.Usage != nil {
		result += fmt.Sprintf("\nTokens used: %d input, %d output\n",
			response.Usage.InputTokens, response.Usage.OutputTokens)
	}

	return result
}

// formatResultWithBase64 creates a result string with base64 data (fallback when storage fails)
// The image is embedded as a data URI using markdown syntax for UI rendering
func (t *Tool) formatResultWithBase64(response *contracts.ImageGenerationResponse, prompt string) string {
	result := fmt.Sprintf("Successfully generated image for prompt: \"%s\"\n\n", truncateString(prompt, 100))

	// Create data URI for direct embedding in markdown
	dataURI := fmt.Sprintf("data:%s;base64,%s", response.Images[0].MimeType, response.Images[0].Base64)
	result += fmt.Sprintf("![Generated image](%s)\n\n", dataURI)

	result += fmt.Sprintf("Format: %s\n", response.Images[0].MimeType)
	result += fmt.Sprintf("Size: %d bytes\n", len(response.Images[0].Data))

	return result
}

// truncateString truncates a string to maxLen characters
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
