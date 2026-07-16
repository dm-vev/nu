package gemini

// ReasoningMode defines the reasoning approach for the model
type ReasoningMode string

const (
	ReasoningModeNone          ReasoningMode = "none"
	ReasoningModeMinimal       ReasoningMode = "minimal"
	ReasoningModeComprehensive ReasoningMode = "comprehensive"
)

// SafetyThreshold represents the safety filtering threshold
type SafetyThreshold string

const (
	SafetyThresholdUnspecified         SafetyThreshold = "HARM_BLOCK_THRESHOLD_UNSPECIFIED"
	SafetyThresholdBlockLowAndAbove    SafetyThreshold = "BLOCK_LOW_AND_ABOVE"
	SafetyThresholdBlockMediumAndAbove SafetyThreshold = "BLOCK_MEDIUM_AND_ABOVE"
	SafetyThresholdBlockOnlyHigh       SafetyThreshold = "BLOCK_ONLY_HIGH"
	SafetyThresholdBlockNone           SafetyThreshold = "BLOCK_NONE"
)

// HarmCategory represents the harm category for safety filtering
type HarmCategory string

const (
	HarmCategoryUnspecified      HarmCategory = "HARM_CATEGORY_UNSPECIFIED"
	HarmCategoryDerogatory       HarmCategory = "HARM_CATEGORY_DEROGATORY"
	HarmCategoryToxicity         HarmCategory = "HARM_CATEGORY_TOXICITY"
	HarmCategoryViolence         HarmCategory = "HARM_CATEGORY_VIOLENCE"
	HarmCategorySexual           HarmCategory = "HARM_CATEGORY_SEXUAL"
	HarmCategoryMedical          HarmCategory = "HARM_CATEGORY_MEDICAL"
	HarmCategoryDangerous        HarmCategory = "HARM_CATEGORY_DANGEROUS"
	HarmCategoryHarassment       HarmCategory = "HARM_CATEGORY_HARASSMENT"
	HarmCategoryHateSpeech       HarmCategory = "HARM_CATEGORY_HATE_SPEECH"
	HarmCategorySexuallyExplicit HarmCategory = "HARM_CATEGORY_SEXUALLY_EXPLICIT"
	HarmCategoryDangerousContent HarmCategory = "HARM_CATEGORY_DANGEROUS_CONTENT"
)

// SafetySetting represents a safety setting for content filtering
type SafetySetting struct {
	Category  HarmCategory    `json:"category"`
	Threshold SafetyThreshold `json:"threshold"`
}

// DefaultSafetySettings returns default safety settings
func DefaultSafetySettings() []SafetySetting {
	return []SafetySetting{
		{
			Category:  HarmCategoryHarassment,
			Threshold: SafetyThresholdBlockMediumAndAbove,
		},
		{
			Category:  HarmCategoryHateSpeech,
			Threshold: SafetyThresholdBlockMediumAndAbove,
		},
		{
			Category:  HarmCategorySexuallyExplicit,
			Threshold: SafetyThresholdBlockMediumAndAbove,
		},
		{
			Category:  HarmCategoryDangerousContent,
			Threshold: SafetyThresholdBlockMediumAndAbove,
		},
	}
}

// ThinkingConfig represents thinking/reasoning configuration for Gemini models
type ThinkingConfig struct {
	// Whether to include thinking content in responses
	IncludeThoughts bool
	// Maximum tokens allocated for thinking (nil for dynamic thinking)
	ThinkingBudget *int32
	// Thought signatures for context preservation across multi-turn conversations
	ThoughtSignatures [][]byte
}

// DefaultThinkingConfig returns default thinking configuration
func DefaultThinkingConfig() ThinkingConfig {
	return ThinkingConfig{
		IncludeThoughts:   false,
		ThinkingBudget:    nil, // Dynamic thinking by default
		ThoughtSignatures: nil,
	}
}
