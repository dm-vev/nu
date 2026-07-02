package app

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"strings"
	"time"

	"nu/internal/agent"
	"nu/internal/auth"
	"nu/internal/cli"
	"nu/internal/model"
	"nu/internal/provider"
	"nu/internal/provider/anthropic"
	"nu/internal/provider/bedrock"
	"nu/internal/provider/compat"
	"nu/internal/provider/google"
	"nu/internal/provider/openai"
	"nu/internal/tool"
)

// Options carries process state into one app invocation.
type Options struct {
	Args       []string
	Env        []string
	CWD        string
	Home       string
	Stdin      io.Reader
	Stdout     io.Writer
	Stderr     io.Writer
	Version    cli.VersionInfo
	Provider   provider.Streamer
	ProviderID string
	API        string
	Model      string
	Tools      map[string]agent.ToolFunc
	SessionID  string
}

// Runtime holds dependencies shared by mode handlers.
type Runtime struct {
	Options Options
}

func normalizeOptions(opts Options) Options {
	if opts.Stdin == nil {
		opts.Stdin = strings.NewReader("")
	}
	if opts.Stdout == nil {
		opts.Stdout = io.Discard
	}
	if opts.Stderr == nil {
		opts.Stderr = io.Discard
	}
	return opts
}

type jsonSessionHeader struct {
	Type       string    `json:"type"`
	Schema     int       `json:"schema"`
	ID         string    `json:"id"`
	CreatedAt  time.Time `json:"created_at"`
	CWD        string    `json:"cwd"`
	App        string    `json:"app"`
	AppVersion string    `json:"app_version"`
}

type printEventWriter struct {
	w   io.Writer
	err error
}

func (w *printEventWriter) emit(ev agent.Event) {
	if w.err != nil || ev.Type != "turn_end" {
		return
	}
	data, ok := ev.Data.(map[string]string)
	if !ok {
		return
	}
	if text := data["text"]; text != "" {
		// Print mode writes only final assistant text; live deltas stay internal.
		_, w.err = fmt.Fprintln(w.w, text)
	}
}

type jsonEventWriter struct {
	w   io.Writer
	err error
}

func (w *jsonEventWriter) emit(ev agent.Event) {
	if w.err != nil {
		return
	}
	w.err = writeJSONLine(w.w, ev)
}

func newAgent(opts Options, emit func(agent.Event)) *agent.Agent {
	if opts.Provider == nil {
		return nil
	}
	tools := opts.Tools
	if tools == nil {
		tools = tool.Builtins(opts.CWD)
	}
	return agent.New(agent.Options{
		Provider:   opts.Provider,
		ProviderID: opts.ProviderID,
		API:        opts.API,
		Model:      opts.Model,
		Tools:      tools,
		Emit:       emit,
	})
}

func configureProvider(ctx context.Context, opts Options, req cli.Request) (Options, error) {
	if opts.Provider != nil {
		return opts, nil
	}
	// Provider URLs are treated as OpenAI-compatible Chat Completions endpoints.
	if isProviderURL(req.Provider) {
		if strings.TrimSpace(req.Model) == "" {
			return Options{}, fmt.Errorf("provider URL requires --model")
		}
		opts.Provider = compat.New(req.Provider, req.APIKey)
		opts.ProviderID = "compat"
		opts.API = "chat"
		opts.Model = strings.TrimSpace(req.Model)
		return opts, nil
	}

	// Runtime selection uses the same model metadata and auth rules as list-models.
	entries, registry, err := loadModelRegistry(req.ModelsPath)
	if err != nil {
		return Options{}, err
	}
	store, err := auth.Load(authFilePath(opts.Home), opts.Env)
	if err != nil {
		return Options{}, err
	}
	authState, err := providerAuthState(ctx, store, entries)
	if err != nil {
		return Options{}, err
	}
	if strings.TrimSpace(req.APIKey) != "" {
		// A CLI key proves auth for selection, then the selected provider consumes it.
		markConfiguredProviders(authState, entries)
	}

	selected, err := selectModel(registry, authState, req)
	if err != nil {
		return Options{}, err
	}
	streamer, err := newProviderClient(ctx, opts, store, req, selected)
	if err != nil {
		return Options{}, err
	}
	opts.Provider = streamer
	opts.ProviderID = selected.Provider
	opts.API = selected.API
	opts.Model = selected.ID
	return opts, nil
}

func loadModelRegistry(path string) ([]model.Model, model.Registry, error) {
	entries := model.Builtins()
	if strings.TrimSpace(path) != "" {
		custom, err := model.LoadCustom(path)
		if err != nil {
			return nil, model.Registry{}, err
		}
		entries = append(entries, custom...)
	}
	return entries, model.NewRegistry(entries), nil
}

func providerAuthState(ctx context.Context, store auth.Store, entries []model.Model) (map[string]bool, error) {
	state := map[string]bool{}
	seen := map[string]bool{}
	for _, entry := range entries {
		if seen[entry.Provider] {
			continue
		}
		seen[entry.Provider] = true
		// Auth resolution may run configured commands, so do it once per provider.
		_, ok, err := store.ResolveAPIKey(ctx, entry.Provider)
		if err != nil {
			return nil, err
		}
		if ok {
			state[entry.Provider] = true
		}
	}
	return state, nil
}

func markConfiguredProviders(state map[string]bool, entries []model.Model) {
	for _, entry := range entries {
		state[entry.Provider] = true
	}
}

func selectModel(registry model.Registry, authState map[string]bool, req cli.Request) (model.Model, error) {
	providerID := strings.TrimSpace(req.Provider)
	modelID := strings.TrimSpace(req.Model)
	if modelID != "" {
		if providerID != "" {
			return selectProviderModel(registry, authState, providerID, modelID)
		}
		return registry.Resolve(modelID, authState)
	}

	available := registry.Available(authState)
	if providerID != "" {
		if selected, ok := defaultModelForProvider(registry, authState, providerID); ok {
			return selected, nil
		}
		for _, entry := range available {
			if entry.Provider == providerID {
				return entry, nil
			}
		}
		return model.Model{}, fmt.Errorf("resolve provider %q: no available models", providerID)
	}
	if len(available) == 0 {
		return model.Model{}, fmt.Errorf("resolve model: no available models")
	}
	// The global default should be stable instead of depending on registry sort order.
	if selected, err := registry.Resolve("openai-default", authState); err == nil {
		return selected, nil
	}
	return available[0], nil
}

func defaultModelForProvider(registry model.Registry, authState map[string]bool, providerID string) (model.Model, bool) {
	var selector string
	switch providerID {
	case "openai":
		selector = "openai-default"
	default:
		return model.Model{}, false
	}
	selected, err := registry.Resolve(selector, authState)
	if err != nil || selected.Provider != providerID {
		return model.Model{}, false
	}
	return selected, true
}

func selectProviderModel(registry model.Registry, authState map[string]bool, providerID string, modelID string) (model.Model, error) {
	if strings.Contains(modelID, "/") {
		selected, err := registry.Resolve(modelID, authState)
		if err != nil {
			return model.Model{}, err
		}
		if selected.Provider != providerID {
			return model.Model{}, fmt.Errorf("resolve model %q: provider is %s, want %s", modelID, selected.Provider, providerID)
		}
		return selected, nil
	}
	return registry.Resolve(providerID+"/"+modelID, authState)
}

func newProviderClient(
	ctx context.Context,
	opts Options,
	store auth.Store,
	req cli.Request,
	selected model.Model,
) (provider.Streamer, error) {
	switch selected.Provider {
	case "openai":
		key, err := providerAPIKey(ctx, store, req, selected.Provider)
		if err != nil {
			return nil, err
		}
		return openai.New(openai.Config{APIKey: key, API: selected.API}), nil
	case "anthropic":
		key, err := providerAPIKey(ctx, store, req, selected.Provider)
		if err != nil {
			return nil, err
		}
		return anthropic.New(anthropic.Config{APIKey: key}), nil
	case "google":
		key, err := providerAPIKey(ctx, store, req, selected.Provider)
		if err != nil {
			return nil, err
		}
		return google.New(google.Config{APIKey: key}), nil
	case "bedrock":
		creds, err := bedrockCredentials(opts.Env)
		if err != nil {
			return nil, err
		}
		return bedrock.New(bedrock.Config{Credentials: creds}), nil
	default:
		return nil, fmt.Errorf("provider %q is not implemented", selected.Provider)
	}
}

func providerAPIKey(ctx context.Context, store auth.Store, req cli.Request, providerID string) (string, error) {
	if key := strings.TrimSpace(req.APIKey); key != "" {
		return key, nil
	}
	key, ok, err := store.ResolveAPIKey(ctx, providerID)
	if err != nil {
		return "", err
	}
	if !ok {
		return "", fmt.Errorf("provider %q requires an api key", providerID)
	}
	return key, nil
}

func bedrockCredentials(env []string) (bedrock.Credentials, error) {
	values := envMap(env)
	creds := bedrock.Credentials{
		AccessKeyID:     strings.TrimSpace(values["AWS_ACCESS_KEY_ID"]),
		SecretAccessKey: strings.TrimSpace(values["AWS_SECRET_ACCESS_KEY"]),
		SessionToken:    strings.TrimSpace(values["AWS_SESSION_TOKEN"]),
		Region:          firstNonEmpty(values["AWS_REGION"], values["AWS_DEFAULT_REGION"]),
	}
	if creds.AccessKeyID == "" || creds.SecretAccessKey == "" {
		return bedrock.Credentials{}, fmt.Errorf("bedrock: missing aws credentials")
	}
	return creds, nil
}

func envMap(env []string) map[string]string {
	values := make(map[string]string, len(env))
	for _, entry := range env {
		name, value, ok := strings.Cut(entry, "=")
		if ok {
			values[name] = value
		}
	}
	return values
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}

func isProviderURL(value string) bool {
	parsed, err := url.Parse(strings.TrimSpace(value))
	if err != nil {
		return false
	}
	return parsed.Host != "" && (parsed.Scheme == "http" || parsed.Scheme == "https")
}

func newJSONSessionHeader(opts Options) (jsonSessionHeader, error) {
	id := opts.SessionID
	if id == "" {
		generated, err := newSessionID()
		if err != nil {
			return jsonSessionHeader{}, err
		}
		id = generated
	}
	version := opts.Version.Version
	if version == "" {
		version = "dev"
	}
	return jsonSessionHeader{
		Type:       "session",
		Schema:     1,
		ID:         id,
		CreatedAt:  time.Now().UTC(),
		CWD:        opts.CWD,
		App:        "nu",
		AppVersion: version,
	}, nil
}

func newSessionID() (string, error) {
	var data [16]byte
	if _, err := rand.Read(data[:]); err != nil {
		return "", fmt.Errorf("generate session id: %w", err)
	}
	data[6] = (data[6] & 0x0f) | 0x40
	data[8] = (data[8] & 0x3f) | 0x80
	return fmt.Sprintf(
		"%x-%x-%x-%x-%x",
		data[0:4],
		data[4:6],
		data[6:8],
		data[8:10],
		data[10:16],
	), nil
}

func writeJSONLine(w io.Writer, value any) error {
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("marshal json line: %w", err)
	}
	if _, err := w.Write(append(data, '\n')); err != nil {
		return fmt.Errorf("write json line: %w", err)
	}
	return nil
}
