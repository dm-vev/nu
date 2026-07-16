package gemini

// Model constants for Gemini API
const (
	// Stable models
	Model25Pro       = "gemini-2.5-pro"
	Model25Flash     = "gemini-2.5-flash"
	Model25FlashLite = "gemini-2.5-flash-lite"
	Model20Flash     = "gemini-2.0-flash"
	Model20FlashLite = "gemini-2.0-flash-lite"
	Model15Pro       = "gemini-1.5-pro"
	Model15Flash     = "gemini-1.5-flash"
	Model15Flash8B   = "gemini-1.5-flash-8b"

	// Preview/Experimental models
	ModelLive25FlashPreview            = "gemini-live-2.5-flash-preview"
	Model25FlashPreviewNativeAudio     = "gemini-2.5-flash-preview-native-audio-dialog"
	Model25FlashExpNativeAudioThinking = "gemini-2.5-flash-exp-native-audio-thinking-dialog"
	Model25FlashPreviewTTS             = "gemini-2.5-flash-preview-tts"
	Model25ProPreviewTTS               = "gemini-2.5-pro-preview-tts"
	Model20FlashPreviewImageGen        = "gemini-2.0-flash-preview-image-generation"
	Model20FlashLive001                = "gemini-2.0-flash-live-001"

	// Image generation models
	Model25FlashImage = "gemini-2.5-flash-image"

	// Gemini 3.x text models
	Model35Flash       = "gemini-3.5-flash"
	Model31ProPreview  = "gemini-3.1-pro-preview"
	Model3FlashPreview = "gemini-3-flash-preview"

	// Multi-turn image editing models (Nano Banana Pro)
	Model3ProImagePreview = "gemini-3-pro-image-preview"

	// Default model
	DefaultModel = Model15Flash

	// Default model for multi-turn image editing
	DefaultImageEditModel = Model3ProImagePreview
)
