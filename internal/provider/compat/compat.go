package compat

import (
	"nu/internal/provider"
	"nu/internal/provider/openai"
)

// New returns an OpenAI-compatible Chat Completions adapter.
func New(baseURL string, apiKey string) provider.Streamer {
	return openai.New(openai.Config{BaseURL: baseURL, APIKey: apiKey, API: "chat"})
}
