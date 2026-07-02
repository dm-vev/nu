package auth

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// Credential is one provider credential entry in auth.json.
type Credential struct {
	APIKey        string `json:"api_key"`
	APIKeyEnv     string `json:"api_key_env"`
	APIKeyCommand string `json:"api_key_command"`
}

// Store resolves provider credentials against a deterministic environment.
type Store struct {
	providers map[string]Credential
	env       map[string]string
}

type authFile struct {
	Providers map[string]Credential `json:"providers"`
}

var fallbackEnv = map[string][]string{
	"openai":    {"OPENAI_API_KEY"},
	"anthropic": {"ANTHROPIC_API_KEY"},
	"google":    {"GEMINI_API_KEY", "GOOGLE_API_KEY"},
	"bedrock":   {"AWS_ACCESS_KEY_ID"},
}

// Load reads auth.json and captures the supplied process environment.
func Load(path string, env []string) (Store, error) {
	store := Store{providers: map[string]Credential{}, env: parseEnv(env)}
	if strings.TrimSpace(path) == "" {
		return store, nil
	}
	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return store, nil
	}
	if err != nil {
		return Store{}, fmt.Errorf("read auth file: %w", err)
	}
	var file authFile
	if err := json.Unmarshal(data, &file); err != nil {
		return Store{}, fmt.Errorf("decode auth file: %w", err)
	}
	for providerID, credential := range file.Providers {
		store.providers[providerID] = credential
	}
	return store, nil
}

// ResolveAPIKey returns the API key for providerID when one is configured.
func (s Store) ResolveAPIKey(ctx context.Context, providerID string) (string, bool, error) {
	credential, ok := s.providers[providerID]
	if ok {
		return s.resolveCredential(ctx, providerID, credential)
	}
	for _, name := range fallbackEnv[providerID] {
		if value := strings.TrimSpace(s.env[name]); value != "" {
			return value, true, nil
		}
	}
	return "", false, nil
}

func (s Store) resolveCredential(ctx context.Context, providerID string, credential Credential) (string, bool, error) {
	if credential.APIKey != "" {
		return os.Expand(credential.APIKey, func(name string) string {
			return s.env[name]
		}), true, nil
	}
	if credential.APIKeyEnv != "" {
		value := strings.TrimSpace(s.env[credential.APIKeyEnv])
		return value, value != "", nil
	}
	if credential.APIKeyCommand != "" {
		return s.runKeyCommand(ctx, providerID, credential.APIKeyCommand)
	}
	return "", false, nil
}

func (s Store) runKeyCommand(ctx context.Context, providerID string, command string) (string, bool, error) {
	cmd := exec.CommandContext(ctx, "sh", "-c", command)
	var stdout bytes.Buffer
	cmd.Stdout = &stdout
	if err := cmd.Run(); err != nil {
		return "", false, fmt.Errorf("run api key command for %s: %w", providerID, err)
	}
	value := strings.TrimSpace(stdout.String())
	return value, value != "", nil
}

func parseEnv(env []string) map[string]string {
	values := make(map[string]string, len(env))
	for _, entry := range env {
		name, value, ok := strings.Cut(entry, "=")
		if ok {
			values[name] = value
		}
	}
	return values
}
