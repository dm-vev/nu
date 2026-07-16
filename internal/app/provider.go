package app

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"

	"nu/internal/agentui"
	"nu/internal/app/auth"
	"nu/internal/app/cli"
	"nu/internal/contracts"
	"nu/internal/llm/anthropic"
	"nu/internal/llm/gemini"
	"nu/internal/llm/openai"
	"nu/internal/model"
)

var fireworksBaseURL = "https://api.fireworks.ai/inference/v1"

func configureProvider(ctx context.Context, opts Options, req cli.Request) (Options, error) {
	if opts.Runner != nil || opts.LLM != nil {
		return opts, nil
	}
	// Provider URLs are treated as OpenAI-compatible Chat Completions endpoints.
	if isProviderURL(req.Provider) {
		if strings.TrimSpace(req.Model) == "" {
			return Options{}, fmt.Errorf("provider URL requires --model")
		}
		opts.LLM = openai.NewClient(req.APIKey, openai.WithModel(strings.TrimSpace(req.Model)), openai.WithBaseURL(req.Provider), openai.WithLogger(discardSDKLogger{}))
		opts.ProviderID = "compat"
		opts.API = "chat"
		opts.Model = strings.TrimSpace(req.Model)
		opts.ModelLabel = opts.Model
		opts.ModelContext = 0
		opts.Models = []model.Model{{
			ID:          opts.Model,
			Provider:    opts.ProviderID,
			API:         opts.API,
			DisplayName: opts.ModelLabel,
			Enabled:     true,
		}}
		return opts, nil
	}

	// Runtime selection uses the same model metadata and auth rules as list-models.
	entries, registry, err := loadModelRegistry(modelsPath(opts.Home, req.ModelsPath))
	if err != nil {
		return Options{}, err
	}
	store, err := auth.Load(authFilePath(opts.Home), opts.Env)
	if err != nil {
		return Options{}, err
	}
	settings, err := loadProviderSettings(opts.Home)
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
	opts.Models = registry.Available(authState)

	selected, err := selectModel(registry, authState, req, settings)
	if err != nil {
		return Options{}, err
	}
	llm, err := newSDKLLM(ctx, opts, store, req, settings, selected)
	if err != nil {
		return Options{}, err
	}
	opts.LLM = llm
	opts.BuildLLM = func(ctx context.Context, config agentui.Config) (contracts.LLM, error) {
		return newSDKLLM(ctx, opts, store, req, settings, model.Model{Provider: config.ProviderID, API: config.API, ID: config.Model})
	}
	opts.ProviderID = selected.Provider
	opts.API = selected.API
	opts.Model = selected.ID
	opts.ModelLabel = firstNonEmpty(selected.DisplayName, selected.ID)
	opts.ModelContext = selected.ContextWindow
	return opts, nil
}

func newSDKLLM(
	ctx context.Context,
	opts Options,
	store auth.Store,
	req cli.Request,
	settings providerSettingsFile,
	selected model.Model,
) (contracts.LLM, error) {
	if setting, ok := settings.Providers[selected.Provider]; ok && strings.TrimSpace(setting.BaseURL) != "" {
		authProvider := firstNonEmpty(setting.AuthProvider, selected.Provider)
		key, err := providerAPIKey(ctx, store, req, authProvider)
		if err != nil {
			return nil, err
		}
		return openai.NewClient(key, openai.WithModel(selected.ID), openai.WithBaseURL(setting.BaseURL), openai.WithLogger(discardSDKLogger{})), nil
	}

	switch selected.Provider {
	case "openai":
		key, err := providerAPIKey(ctx, store, req, selected.Provider)
		if err != nil {
			return nil, err
		}
		return openai.NewClient(key, openai.WithModel(selected.ID), openai.WithLogger(discardSDKLogger{})), nil
	case "anthropic":
		key, err := providerAPIKey(ctx, store, req, selected.Provider)
		if err != nil {
			return nil, err
		}
		return anthropic.NewClient(key, anthropic.WithModel(selected.ID), anthropic.WithLogger(discardSDKLogger{})), nil
	case "google":
		key, err := providerAPIKey(ctx, store, req, selected.Provider)
		if err != nil {
			return nil, err
		}
		return gemini.NewClient(ctx, gemini.WithAPIKey(key), gemini.WithModel(selected.ID), gemini.WithLogger(discardSDKLogger{}))
	case "bedrock":
		creds, err := bedrockCredentials(opts.Env)
		if err != nil {
			return nil, err
		}
		awsConfig, err := awsconfig.LoadDefaultConfig(ctx,
			awsconfig.WithRegion(creds.Region),
			awsconfig.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(creds.AccessKeyID, creds.SecretAccessKey, creds.SessionToken)),
		)
		if err != nil {
			return nil, fmt.Errorf("bedrock config: %w", err)
		}
		return anthropic.NewClient("", anthropic.WithModel(selected.ID), anthropic.WithBedrockAWSConfig(awsConfig), anthropic.WithLogger(discardSDKLogger{})), nil
	case "fireworks":
		key, err := providerAPIKey(ctx, store, req, selected.Provider)
		if err != nil {
			return nil, err
		}
		return openai.NewClient(key, openai.WithModel(selected.ID), openai.WithBaseURL(fireworksBaseURL), openai.WithLogger(discardSDKLogger{})), nil
	case "compat":
		if !isProviderURL(req.Provider) {
			return nil, fmt.Errorf("provider %q requires a provider URL", selected.Provider)
		}
		return openai.NewClient(req.APIKey, openai.WithModel(selected.ID), openai.WithBaseURL(req.Provider), openai.WithLogger(discardSDKLogger{})), nil
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

type bedrockAuth struct {
	AccessKeyID     string
	SecretAccessKey string
	SessionToken    string
	Region          string
}

func bedrockCredentials(env []string) (bedrockAuth, error) {
	values := envMap(env)
	creds := bedrockAuth{
		AccessKeyID:     strings.TrimSpace(values["AWS_ACCESS_KEY_ID"]),
		SecretAccessKey: strings.TrimSpace(values["AWS_SECRET_ACCESS_KEY"]),
		SessionToken:    strings.TrimSpace(values["AWS_SESSION_TOKEN"]),
		Region:          firstNonEmpty(values["AWS_REGION"], values["AWS_DEFAULT_REGION"]),
	}
	if creds.AccessKeyID == "" || creds.SecretAccessKey == "" {
		return bedrockAuth{}, fmt.Errorf("bedrock: missing aws credentials")
	}
	if creds.Region == "" {
		creds.Region = "us-east-1"
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
