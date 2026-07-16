package image

import (
	"context"
	"fmt"
	"time"

	"nu/internal/contracts"
	"nu/internal/data/storage"
	"nu/internal/multitenancy"
)

func (t *EditTool) formatResponse(ctx context.Context, sessionID string, resp *contracts.ImageEditResponse, prompt string, isInitial bool) (string, error) {
	var result string

	if isInitial {
		result = fmt.Sprintf("Image editing session started with initial image.\n\nSession ID: %s\n\n", sessionID)
	} else {
		result = fmt.Sprintf("Image edited successfully.\n\nSession ID: %s\n\n", sessionID)
	}

	// Add text response if present
	if resp.Text != "" {
		result += fmt.Sprintf("Model response: %s\n\n", resp.Text)
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

				// Get org and thread IDs from context if available
				if orgID, err := multitenancy.GetOrgID(ctx); err == nil {
					metadata.OrgID = orgID
				}

				url, err := t.storage.Store(ctx, &image, metadata)
				if err != nil {
					// Log warning but continue with base64
					fmt.Printf("[imageedit] Storage failed, using base64: %v\n", err)
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

	result += fmt.Sprintf("\nUse action='edit' with session_id='%s' to continue refining, or action='end_session' to close.", sessionID)

	return result, nil
}

func (t *EditTool) formatImageURL(url string, image *contracts.GeneratedImage, index int) string {
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

func (t *EditTool) formatImageBase64(image *contracts.GeneratedImage, index int) string {
	result := ""
	if index > 0 {
		result = fmt.Sprintf("\n--- Image %d ---\n", index+1)
	}
	// Create data URI for direct embedding in markdown
	dataURI := fmt.Sprintf("data:%s;base64,%s", image.MimeType, image.Base64)
	result += fmt.Sprintf("![Generated image](%s)\n\n", dataURI)
	result += fmt.Sprintf("Format: %s\n", image.MimeType)
	result += fmt.Sprintf("Size: %d bytes\n", len(image.Data))
	return result
}
