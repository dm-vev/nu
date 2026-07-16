package gemini

import "strings"

// ModelCapabilities represents the capabilities of different Gemini models
type ModelCapabilities struct {
	SupportsStreaming             bool
	SupportsToolCalling           bool
	SupportsVision                bool
	SupportsAudio                 bool
	SupportsThinking              bool
	SupportsImageGeneration       bool // Whether the model can generate images
	SupportsMultiTurnImageEditing bool // Whether the model supports conversational image editing
	MaxInputTokens                int
	MaxOutputTokens               int
	MaxThinkingTokens             *int32 // nil if thinking not supported
	MaxReferenceImages            int    // Max reference images for multi-turn editing (Gemini 3 Pro: 14)
	MaxObjectImages               int    // Max high-fidelity object images (Gemini 3 Pro: 6)
	MaxHumanImages                int    // Max human images for character consistency (Gemini 3 Pro: 5)
	SupportedMimeTypes            []string
	SupportedOutputFormats        []string // Output formats for image generation (e.g., "png", "jpeg")
	SupportedImageSizes           []string // Supported image sizes (e.g., "1K", "2K", "4K")
}

// GetModelCapabilities returns the capabilities for a given model
func GetModelCapabilities(model string) ModelCapabilities {
	switch model {
	case Model25Pro:
		maxThinking := int32(32768) // 32K tokens for Pro
		return ModelCapabilities{
			SupportsStreaming:   true,
			SupportsToolCalling: true,
			SupportsVision:      true,
			SupportsAudio:       true,
			SupportsThinking:    true,
			MaxInputTokens:      2097152, // 2M tokens
			MaxOutputTokens:     8192,
			MaxThinkingTokens:   &maxThinking,
			SupportedMimeTypes: []string{
				"image/png", "image/jpeg", "image/webp", "image/heic", "image/heif",
				"audio/wav", "audio/mp3", "audio/aiff", "audio/aac", "audio/ogg", "audio/flac",
				"video/mp4", "video/mpeg", "video/mov", "video/avi", "video/flv", "video/mpv", "video/webm", "video/wmv", "video/3gpp",
				"text/plain", "text/html", "text/css", "text/javascript", "application/x-javascript", "text/x-typescript",
				"application/pdf",
			},
		}
	case Model25Flash:
		maxThinking := int32(24576) // 24K tokens for Flash
		return ModelCapabilities{
			SupportsStreaming:   true,
			SupportsToolCalling: true,
			SupportsVision:      true,
			SupportsAudio:       true,
			SupportsThinking:    true,
			MaxInputTokens:      1048576, // 1M tokens
			MaxOutputTokens:     8192,
			MaxThinkingTokens:   &maxThinking,
			SupportedMimeTypes: []string{
				"image/png", "image/jpeg", "image/webp", "image/heic", "image/heif",
				"audio/wav", "audio/mp3", "audio/aiff", "audio/aac", "audio/ogg", "audio/flac",
				"video/mp4", "video/mpeg", "video/mov", "video/avi", "video/flv", "video/mpv", "video/webm", "video/wmv", "video/3gpp",
				"text/plain", "text/html", "text/css", "text/javascript", "application/x-javascript", "text/x-typescript",
				"application/pdf",
			},
		}
	case Model25FlashLite:
		return ModelCapabilities{
			SupportsStreaming:   true,
			SupportsToolCalling: true,
			SupportsVision:      false,
			SupportsAudio:       false,
			SupportsThinking:    false, // Lite model doesn't support thinking
			MaxInputTokens:      32768,
			MaxOutputTokens:     8192,
			MaxThinkingTokens:   nil,
			SupportedMimeTypes: []string{
				"text/plain",
			},
		}
	case Model20Flash:
		return ModelCapabilities{
			SupportsStreaming:   true,
			SupportsToolCalling: true,
			SupportsVision:      true,
			SupportsAudio:       false,
			SupportsThinking:    false,   // 2.0 and 1.5 models don't support thinking
			MaxInputTokens:      1048576, // 1M tokens
			MaxOutputTokens:     8192,
			MaxThinkingTokens:   nil,
			SupportedMimeTypes: []string{
				"image/png", "image/jpeg", "image/webp", "image/heic", "image/heif",
				"video/mp4", "video/mpeg", "video/mov", "video/avi", "video/flv", "video/mpv", "video/webm", "video/wmv", "video/3gpp",
				"text/plain",
			},
		}
	case Model20FlashLite:
		return ModelCapabilities{
			SupportsStreaming:   true,
			SupportsToolCalling: true,
			SupportsVision:      false,
			SupportsAudio:       false,
			MaxInputTokens:      32768,
			MaxOutputTokens:     8192,
			MaxThinkingTokens:   nil,
			SupportedMimeTypes: []string{
				"text/plain",
			},
		}
	case Model15Pro:
		return ModelCapabilities{
			SupportsStreaming:   true,
			SupportsToolCalling: true,
			SupportsVision:      true,
			SupportsAudio:       false,
			SupportsThinking:    false,   // 2.0 and 1.5 models don't support thinking
			MaxInputTokens:      2097152, // 2M tokens
			MaxOutputTokens:     8192,
			MaxThinkingTokens:   nil,
			SupportedMimeTypes: []string{
				"image/png", "image/jpeg", "image/webp", "image/heic", "image/heif",
				"video/mp4", "video/mpeg", "video/mov", "video/avi", "video/flv", "video/mpv", "video/webm", "video/wmv", "video/3gpp",
				"text/plain", "text/html", "text/css", "text/javascript", "application/x-javascript", "text/x-typescript",
				"application/pdf",
			},
		}
	case Model15Flash:
		return ModelCapabilities{
			SupportsStreaming:   true,
			SupportsToolCalling: true,
			SupportsVision:      true,
			SupportsAudio:       false,
			SupportsThinking:    false,   // 2.0 and 1.5 models don't support thinking
			MaxInputTokens:      1048576, // 1M tokens
			MaxOutputTokens:     8192,
			MaxThinkingTokens:   nil,
			SupportedMimeTypes: []string{
				"image/png", "image/jpeg", "image/webp", "image/heic", "image/heif",
				"video/mp4", "video/mpeg", "video/mov", "video/avi", "video/flv", "video/mpv", "video/webm", "video/wmv", "video/3gpp",
				"text/plain", "text/html", "text/css", "text/javascript", "application/x-javascript", "text/x-typescript",
				"application/pdf",
			},
		}
	case Model15Flash8B:
		return ModelCapabilities{
			SupportsStreaming:   true,
			SupportsToolCalling: true,
			SupportsVision:      true,
			SupportsAudio:       false,
			SupportsThinking:    false,   // 2.0 and 1.5 models don't support thinking
			MaxInputTokens:      1048576, // 1M tokens
			MaxOutputTokens:     8192,
			MaxThinkingTokens:   nil,
			SupportedMimeTypes: []string{
				"image/png", "image/jpeg", "image/webp", "image/heic", "image/heif",
				"video/mp4", "video/mpeg", "video/mov", "video/avi", "video/flv", "video/mpv", "video/webm", "video/wmv", "video/3gpp",
				"text/plain",
			},
		}
	// Preview/Experimental models
	case ModelLive25FlashPreview, Model25FlashPreviewNativeAudio, Model25FlashExpNativeAudioThinking,
		Model25FlashPreviewTTS, Model25ProPreviewTTS, Model20FlashLive001:
		return ModelCapabilities{
			SupportsStreaming:   true,
			SupportsToolCalling: true,
			SupportsVision:      true,
			SupportsAudio:       true,
			MaxInputTokens:      1048576, // 1M tokens
			MaxOutputTokens:     8192,
			MaxThinkingTokens:   nil,
			SupportedMimeTypes: []string{
				"image/png", "image/jpeg", "image/webp", "image/heic", "image/heif",
				"audio/wav", "audio/mp3", "audio/aiff", "audio/aac", "audio/ogg", "audio/flac",
				"video/mp4", "video/mpeg", "video/mov", "video/avi", "video/flv", "video/mpv", "video/webm", "video/wmv", "video/3gpp",
				"text/plain",
			},
		}
	case Model20FlashPreviewImageGen:
		return ModelCapabilities{
			SupportsStreaming:       true,
			SupportsToolCalling:     true,
			SupportsVision:          true,
			SupportsAudio:           false,
			SupportsThinking:        false,   // 2.0 and 1.5 models don't support thinking
			SupportsImageGeneration: true,    // Can generate images
			MaxInputTokens:          1048576, // 1M tokens
			MaxOutputTokens:         8192,
			MaxThinkingTokens:       nil,
			SupportedMimeTypes: []string{
				"image/png", "image/jpeg", "image/webp", "image/heic", "image/heif",
				"text/plain",
			},
			SupportedOutputFormats: []string{"png", "jpeg"},
		}
	case Model25FlashImage:
		return ModelCapabilities{
			SupportsStreaming:             true,
			SupportsToolCalling:           false, // Image gen models typically don't support tools
			SupportsVision:                true,  // Can accept images as input for image-to-image
			SupportsAudio:                 false,
			SupportsThinking:              false,
			SupportsImageGeneration:       true, // Primary purpose: generate images
			SupportsMultiTurnImageEditing: true, // Supports chat-based image editing
			MaxInputTokens:                32768,
			MaxOutputTokens:               8192,
			MaxThinkingTokens:             nil,
			SupportedMimeTypes: []string{
				"image/png", "image/jpeg", "image/webp",
				"text/plain",
			},
			SupportedOutputFormats: []string{"png", "jpeg"},
			SupportedImageSizes:    []string{"1K", "2K"},
		}
	case Model35Flash, Model31ProPreview, Model3FlashPreview:
		maxThinking := int32(24576)
		return ModelCapabilities{
			SupportsStreaming:   true,
			SupportsToolCalling: true,
			SupportsVision:      true,
			SupportsAudio:       true,
			SupportsThinking:    true,
			MaxInputTokens:      1048576,
			MaxOutputTokens:     65536,
			MaxThinkingTokens:   &maxThinking,
			SupportedMimeTypes: []string{
				"image/png", "image/jpeg", "image/webp", "image/heic", "image/heif",
				"audio/wav", "audio/mp3", "audio/aiff", "audio/aac", "audio/ogg", "audio/flac",
				"video/mp4", "video/mpeg", "video/quicktime", "video/avi", "video/x-flv", "video/mpg", "video/webm", "video/wmv", "video/3gpp",
				"text/plain", "text/html", "text/css", "text/javascript", "application/x-javascript", "text/x-typescript",
				"application/pdf",
			},
		}
	case Model3ProImagePreview:
		// Nano Banana Pro - Google's most advanced image generation and editing model
		return ModelCapabilities{
			SupportsStreaming:             true,
			SupportsToolCalling:           false, // Image gen models typically don't support tools
			SupportsVision:                true,  // Can accept images as input
			SupportsAudio:                 false,
			SupportsThinking:              true, // Uses "Thinking" for complex instructions
			SupportsImageGeneration:       true,
			SupportsMultiTurnImageEditing: true, // Primary feature: multi-turn image editing
			MaxInputTokens:                32768,
			MaxOutputTokens:               8192,
			MaxThinkingTokens:             nil,
			MaxReferenceImages:            14, // Up to 14 reference images
			MaxObjectImages:               6,  // Up to 6 high-fidelity object images
			MaxHumanImages:                5,  // Up to 5 human images for character consistency
			SupportedMimeTypes: []string{
				"image/png", "image/jpeg", "image/webp",
				"text/plain",
			},
			SupportedOutputFormats: []string{"png", "jpeg"},
			SupportedImageSizes:    []string{"1K", "2K", "4K"},
		}
	default:
		// Gemini 3.x models not explicitly listed above default to thinking-enabled
		// so thought_signature is handled correctly in tool loops.
		if strings.HasPrefix(model, "gemini-3") {
			maxThinking := int32(24576)
			return ModelCapabilities{
				SupportsStreaming:   true,
				SupportsToolCalling: true,
				SupportsVision:      true,
				SupportsAudio:       true,
				SupportsThinking:    true,
				MaxInputTokens:      1048576,
				MaxOutputTokens:     65536,
				MaxThinkingTokens:   &maxThinking,
				SupportedMimeTypes: []string{
					"image/png", "image/jpeg", "image/webp", "image/heic", "image/heif",
					"audio/wav", "audio/mp3", "audio/aiff", "audio/aac", "audio/ogg", "audio/flac",
					"video/mp4", "video/mpeg", "video/quicktime", "video/avi", "video/x-flv", "video/mpg", "video/webm", "video/wmv", "video/3gpp",
					"text/plain", "text/html", "text/css", "text/javascript", "application/x-javascript", "text/x-typescript",
					"application/pdf",
				},
			}
		}
		// Return default capabilities for unknown models
		return ModelCapabilities{
			SupportsStreaming:   true,
			SupportsToolCalling: true,
			SupportsVision:      false,
			SupportsAudio:       false,
			SupportsThinking:    false,
			MaxInputTokens:      32768,
			MaxOutputTokens:     2048,
			MaxThinkingTokens:   nil,
			SupportedMimeTypes: []string{
				"text/plain",
			},
		}
	}
}
