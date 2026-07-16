package image

import (
	"time"

	"nu/internal/contracts"
)

// GetImageReference returns an ImageReference for storing in memory
func GetImageReference(image *contracts.GeneratedImage, prompt string) contracts.ImageReference {
	return contracts.ImageReference{
		URL:       image.URL,
		MimeType:  image.MimeType,
		Prompt:    prompt,
		CreatedAt: time.Now(),
	}
}
