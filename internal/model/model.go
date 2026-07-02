package model

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// ThinkingLevel is the requested reasoning effort.
type ThinkingLevel string

const (
	ThinkingOff     ThinkingLevel = "off"
	ThinkingMinimal ThinkingLevel = "minimal"
	ThinkingLow     ThinkingLevel = "low"
	ThinkingMedium  ThinkingLevel = "medium"
	ThinkingHigh    ThinkingLevel = "high"
	ThinkingXHigh   ThinkingLevel = "xhigh"
)

// Model is one model registry entry.
type Model struct {
	ID              string          `json:"id"`
	Provider        string          `json:"provider"`
	API             string          `json:"api"`
	DisplayName     string          `json:"display_name,omitempty"`
	Aliases         []string        `json:"aliases,omitempty"`
	Patterns        []string        `json:"patterns,omitempty"`
	Enabled         bool            `json:"enabled"`
	RequiresAuth    bool            `json:"requires_auth"`
	Input           []string        `json:"input,omitempty"`
	ContextWindow   int             `json:"context_window,omitempty"`
	MaxOutput       int             `json:"max_output,omitempty"`
	CostInputMTok   float64         `json:"cost_input_mtok,omitempty"`
	CostOutputMTok  float64         `json:"cost_output_mtok,omitempty"`
	ThinkingLevels  []ThinkingLevel `json:"thinking_levels,omitempty"`
	SupportsTools   bool            `json:"supports_tools,omitempty"`
	SupportsImages  bool            `json:"supports_images,omitempty"`
	SupportsCaching bool            `json:"supports_caching,omitempty"`
}

// UnmarshalJSON defaults omitted enabled to true while preserving explicit false.
func (m *Model) UnmarshalJSON(data []byte) error {
	type alias Model
	raw := struct {
		Enabled *bool `json:"enabled"`
		*alias
	}{
		alias: (*alias)(m),
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	if raw.Enabled == nil {
		m.Enabled = true
	}
	return nil
}

// Registry resolves model names and lists visible models.
type Registry struct {
	models map[string]Model
	keys   []string
}

type customFile struct {
	Models []Model `json:"models"`
}

// Builtins returns Phase 3 built-in model metadata.
func Builtins() []Model {
	return []Model{
		{
			ID:             "gpt-5.5",
			Provider:       "openai",
			API:            "responses",
			Aliases:        []string{"gpt-5", "openai-default"},
			Patterns:       []string{"gpt-5*", "openai/gpt-5*"},
			Enabled:        true,
			RequiresAuth:   true,
			Input:          []string{"text", "image"},
			ContextWindow:  400000,
			MaxOutput:      128000,
			ThinkingLevels: []ThinkingLevel{ThinkingMinimal, ThinkingLow, ThinkingMedium, ThinkingHigh},
			SupportsTools:  true,
			SupportsImages: true,
		},
		{
			ID:             "gpt-4.1",
			Provider:       "openai",
			API:            "chat",
			Aliases:        []string{"gpt-4.1-chat"},
			Patterns:       []string{"gpt-4.1*", "openai/gpt-4.1*"},
			Enabled:        true,
			RequiresAuth:   true,
			Input:          []string{"text", "image"},
			ContextWindow:  1000000,
			MaxOutput:      32768,
			SupportsTools:  true,
			SupportsImages: true,
		},
		{
			ID:             "claude-opus-4-8",
			Provider:       "anthropic",
			API:            "messages",
			Aliases:        []string{"opus", "claude-opus"},
			Patterns:       []string{"claude-opus*", "anthropic/claude-opus*"},
			Enabled:        true,
			RequiresAuth:   true,
			Input:          []string{"text", "image"},
			ContextWindow:  200000,
			MaxOutput:      128000,
			ThinkingLevels: []ThinkingLevel{ThinkingLow, ThinkingMedium, ThinkingHigh, ThinkingXHigh},
			SupportsTools:  true,
			SupportsImages: true,
		},
		{
			ID:             "gemini-3.5-flash",
			Provider:       "google",
			API:            "generateContent",
			Aliases:        []string{"gemini-flash"},
			Patterns:       []string{"gemini-*", "google/gemini-*"},
			Enabled:        true,
			RequiresAuth:   true,
			Input:          []string{"text", "image", "audio", "video"},
			ContextWindow:  1000000,
			MaxOutput:      65536,
			ThinkingLevels: []ThinkingLevel{ThinkingMinimal, ThinkingLow, ThinkingMedium, ThinkingHigh},
			SupportsTools:  true,
			SupportsImages: true,
		},
		{
			ID:             "anthropic.claude-3-sonnet-20240229-v1:0",
			Provider:       "bedrock",
			API:            "converse-stream",
			Aliases:        []string{"bedrock-sonnet"},
			Patterns:       []string{"bedrock/*", "anthropic.claude-*"},
			Enabled:        true,
			RequiresAuth:   true,
			Input:          []string{"text", "image"},
			ContextWindow:  200000,
			MaxOutput:      4096,
			SupportsTools:  true,
			SupportsImages: true,
		},
	}
}

// NewRegistry builds a deterministic model registry.
func NewRegistry(models []Model) Registry {
	registry := Registry{models: make(map[string]Model, len(models))}
	for _, model := range models {
		if model.Provider == "" || model.ID == "" {
			continue
		}
		key := modelKey(model.Provider, model.ID)
		registry.models[key] = model
	}
	registry.keys = make([]string, 0, len(registry.models))
	for key := range registry.models {
		registry.keys = append(registry.keys, key)
	}
	sort.Strings(registry.keys)
	return registry
}

// LoadCustom reads custom models from models.json.
func LoadCustom(path string) ([]Model, error) {
	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("read models file: %w", err)
	}
	var file customFile
	if err := json.Unmarshal(data, &file); err != nil {
		return nil, fmt.Errorf("decode models file: %w", err)
	}
	for _, model := range file.Models {
		if model.Provider == "" || model.API == "" || model.ID == "" {
			return nil, fmt.Errorf("decode models file: missing provider, api, or id")
		}
	}
	return file.Models, nil
}

// Resolve finds one visible model by exact name, alias, or glob pattern.
func (r Registry) Resolve(pattern string, auth map[string]bool) (Model, error) {
	pattern = strings.TrimSpace(pattern)
	if pattern == "" {
		return Model{}, fmt.Errorf("resolve model: missing pattern")
	}
	for _, model := range r.Available(auth) {
		if matchesModel(pattern, model) {
			return model, nil
		}
	}
	return Model{}, fmt.Errorf("resolve model %q: not found", pattern)
}

// Available returns visible models under the current auth state.
func (r Registry) Available(auth map[string]bool) []Model {
	out := make([]Model, 0, len(r.keys))
	for _, key := range r.keys {
		model := r.models[key]
		if !model.Enabled {
			continue
		}
		if model.RequiresAuth && !auth[model.Provider] {
			continue
		}
		out = append(out, model)
	}
	return out
}

// ThinkingFor returns the provider-specific request fragment for level.
func ThinkingFor(providerID string, api string, level ThinkingLevel) (map[string]any, error) {
	if level == "" || level == ThinkingOff {
		return nil, nil
	}
	if !validThinking(level) {
		return nil, fmt.Errorf("thinking level %q is not supported", level)
	}
	switch providerID + "/" + api {
	case "openai/responses":
		if level == ThinkingXHigh {
			return nil, fmt.Errorf("thinking level %q is not supported by openai responses", level)
		}
		return map[string]any{"reasoning": map[string]any{"effort": string(level)}}, nil
	case "anthropic/messages":
		return map[string]any{"thinking": map[string]any{"type": "enabled", "budget_tokens": thinkingBudget(level)}}, nil
	case "google/generateContent":
		return map[string]any{
			"generationConfig": map[string]any{
				"thinkingConfig": map[string]any{"thinkingBudget": thinkingBudget(level)},
			},
		}, nil
	default:
		return nil, fmt.Errorf("thinking is not supported by %s/%s", providerID, api)
	}
}

func matchesModel(pattern string, model Model) bool {
	candidates := []string{model.ID, modelKey(model.Provider, model.ID)}
	candidates = append(candidates, model.Aliases...)
	for _, candidate := range candidates {
		if candidate == pattern || glob(pattern, candidate) {
			return true
		}
	}
	for _, configured := range model.Patterns {
		if configured == pattern || glob(configured, pattern) {
			return true
		}
	}
	return false
}

func modelKey(providerID, id string) string {
	return providerID + "/" + id
}

func glob(pattern, value string) bool {
	ok, err := filepath.Match(pattern, value)
	return err == nil && ok
}

func validThinking(level ThinkingLevel) bool {
	switch level {
	case ThinkingMinimal, ThinkingLow, ThinkingMedium, ThinkingHigh, ThinkingXHigh:
		return true
	default:
		return false
	}
}

func thinkingBudget(level ThinkingLevel) int {
	switch level {
	case ThinkingMinimal:
		return 1024
	case ThinkingLow:
		return 4096
	case ThinkingMedium:
		return 8192
	case ThinkingHigh:
		return 16384
	case ThinkingXHigh:
		return 32768
	default:
		return 0
	}
}
