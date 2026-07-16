package gemini

import "fmt"

// IsVisionModel returns true if the model supports vision capabilities
func IsVisionModel(model string) bool {
	capabilities := GetModelCapabilities(model)
	return capabilities.SupportsVision
}

// IsAudioModel returns true if the model supports audio capabilities
func IsAudioModel(model string) bool {
	capabilities := GetModelCapabilities(model)
	return capabilities.SupportsAudio
}

// SupportsToolCalling returns true if the model supports function/tool calling
func SupportsToolCalling(model string) bool {
	capabilities := GetModelCapabilities(model)
	return capabilities.SupportsToolCalling
}

// GeminiSupportsThinking returns true if the model supports thinking capabilities
func SupportsThinking(model string) bool {
	capabilities := GetModelCapabilities(model)
	return capabilities.SupportsThinking
}

// SupportsImageGeneration returns true if the model supports image generation
func SupportsImageGeneration(model string) bool {
	capabilities := GetModelCapabilities(model)
	return capabilities.SupportsImageGeneration
}

// GetSupportedOutputFormats returns the supported output formats for image generation
func GetSupportedOutputFormats(model string) []string {
	capabilities := GetModelCapabilities(model)
	return capabilities.SupportedOutputFormats
}

// GetMaxThinkingTokens returns the maximum thinking tokens for a model
func GetMaxThinkingTokens(model string) *int32 {
	capabilities := GetModelCapabilities(model)
	return capabilities.MaxThinkingTokens
}

// ValidateThinkingBudget validates if a thinking budget is within model limits
func ValidateThinkingBudget(model string, budget int32) error {
	if !SupportsThinking(model) {
		return fmt.Errorf("model %s does not support thinking", model)
	}

	maxTokens := GetMaxThinkingTokens(model)
	if maxTokens != nil && budget > *maxTokens {
		return fmt.Errorf("thinking budget %d exceeds maximum %d for model %s", budget, *maxTokens, model)
	}

	return nil
}

// SupportsMultiTurnImageEditing returns true if the model supports multi-turn image editing
func SupportsMultiTurnImageEditing(model string) bool {
	capabilities := GetModelCapabilities(model)
	return capabilities.SupportsMultiTurnImageEditing
}

// GetSupportedImageSizes returns the supported image sizes for the model
func GetSupportedImageSizes(model string) []string {
	capabilities := GetModelCapabilities(model)
	return capabilities.SupportedImageSizes
}
