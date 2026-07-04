# `internal/app/runtime.go`

## Status

Current: IMPLEMENTED
Implementation Commit: a44b95f
Implementation Comments: Runtime carries process IO, provider settings, visible model candidates, tool functions and definitions, optional session id, mode-specific emitters, default built-ins, URL-compatible provider support, OpenAI default selection, Phase 3 provider construction helpers, selected model display labels, global models/settings file defaults, and configured OpenAI-compatible providers.

## TODO

- [x] Add or confirm the failing tests listed in this file.
- [x] Implement the file according to the function logic below.
- [x] Run the targeted package tests.
- [x] After implementation commit, replace `Implementation Commit` with the commit hash and summarize important comments.

## Purpose

Shared runtime object passed to mode handlers.

## Code Style

Plain structs with concrete dependencies. Construct mode-specific agents with
the emitter that mode needs. Provider construction stays in runtime helpers so
tests can still inject fake providers.

## Types

### `type Runtime struct`

Logic:

- Store normalized process IO, provider settings, tool functions, tool
  definitions, and session id.
- Keep ownership explicit: runtime closes only components it created.
- Do not start goroutines, TUI loops, RPC loops, or provider streams during construction.

Acceptance:

- contains normalized process IO, provider settings, tool functions, tool
  definitions, and session id;
- has no goroutines after construction.

### `type Options struct`

Logic:

- Carry argv, environment, cwd, home, stdin, stdout, stderr, version, optional
  provider settings, selected model display name, visible model candidates, tool functions, tool
  definitions, and optional session id from `cmd/nu`.
- Keep options serializable enough for integration tests to construct without process globals.
- Do not store parsed CLI state here; parsing output belongs to `cli.Request`.

Acceptance:

- carries argv, env, cwd, home, stdin, stdout, stderr, version metadata,
  optional provider, selected model display name, visible model candidates, tool functions, tool
  definitions, and optional session id.

## Functions

### `newAgent(opts Options, emit func(agent.Event)) *agent.Agent`

Logic:

- Return nil when no provider is configured.
- Create `agent.Agent` with provider id, API, model, and provider stream.
- Pass configured tool functions through to the agent.
- Use `tool.Builtins(opts.CWD)` when no tool map is supplied.
- Pass configured tool definitions through to the agent.
- Use `tool.Definitions()` when no tool definition list is supplied.
- Install the mode-specific event callback.

Acceptance:

- no provider means no agent;
- mode-specific emitter receives agent events.
- default built-ins are available when `Options.Tools` is nil.
- default tool schemas are available when `Options.ToolDefs` is nil.

### `configureProvider(ctx context.Context, opts Options, req cli.Request) (Options, error)`

Logic:

- Return injected `Options.Provider` unchanged for tests.
- Load auth from `~/.nu/auth.json` and `Options.Env`.
- Load provider settings from `~/.nu/agent/settings.json` when present.
- Build a model registry from built-ins plus explicit `req.ModelsPath` or,
  when omitted, global `~/.nu/agent/models.json`.
- Treat URL `req.Provider` values as OpenAI-compatible Chat Completions
  endpoints and require `req.Model`.
- Resolve `req.Model` when supplied.
- Preserve custom `display_name` from selected model metadata for UI/RPC/list
  output while keeping provider requests on the real model id.
- Preserve the visible model candidate list for interactive `/model` selection.
- Prefer provider-specific defaults such as `openai-default` when only a
  provider is supplied.
- Prefer saved `default_provider/default_model` from settings when no CLI model is supplied.
- Prefer `openai-default` when no provider/model is supplied and OpenAI is
  available, otherwise choose the first visible model.
- Use `req.APIKey` before auth store values.
- Construct OpenAI, Anthropic, Google, Bedrock, configured, or OpenAI-compatible
  custom provider adapters.

Acceptance:

- `--print --provider openai --model ...` can create a real provider adapter;
- `--print --provider http://... --model ...` reaches a compat chat endpoint;
- `--models` custom entries can affect runtime selection;
- custom `display_name` entries affect display labels without changing provider
  model ids;
- global `~/.nu/agent/models.json` is used when `--models` is omitted;
- default selection is stable and does not depend on registry sort order.
- interactive mode receives the same visible model list as runtime selection.
- configured providers with `base_url` use the OpenAI-compatible chat adapter.
- saved `/model` choices are restored on later app starts.

### `loadModelRegistry(path string) ([]model.Model, model.Registry, error)`

Logic:

- Start with built-in model entries.
- Load and append custom entries when `path` is non-empty.
- Let later entries override earlier entries through `model.NewRegistry`.

Acceptance:

- custom `models.json` entries appear in list-models and runtime selection.

### `modelsPath(home string, explicit string) string`

Logic:

- Return `explicit` when non-empty.
- Return empty when home is empty.
- Otherwise return `~/.nu/agent/models.json` under the supplied home.

Acceptance:

- runtime uses global models by default without reading process globals.

### `loadProviderSettings(home string) (providerSettingsFile, error)`

Logic:

- Return empty settings when home is empty or `~/.nu/agent/settings.json` is missing.
- Decode configured provider settings keyed by provider id.
- Preserve root default provider/model plus provider `base_url`, optional `auth_provider`, and provider `default_model`.

Acceptance:

- `TestPrintModeBuildsProviderFromSettings` fails if a configured provider cannot build a compat client.

### `saveSelectedModel(home string, selected model.Model) error`

Logic:

- No-op when home is empty.
- Load existing settings, preserve provider connection fields, and write the selected model as root `default_provider/default_model`.
- Also update the selected provider's `default_model` for provider-only selection.

Acceptance:

- `TestSavedModelSelectionRestoresDefault` fails if `/model` selections cannot survive restart.

### `writeProviderSettings(home string, settings providerSettingsFile) error`

Logic:

- Create `~/.nu/agent` when needed.
- Write formatted JSON through a temporary file and rename it into place.

Acceptance:

- A partial settings write does not leave a truncated settings file on normal filesystem errors.

### `providerAuthState(ctx context.Context, store auth.Store, entries []model.Model) (map[string]bool, error)`

Logic:

- Resolve auth once for each provider present in model entries.
- Use the supplied auth store and context; do not read process globals.
- Return only providers with usable credentials.

Acceptance:

- custom provider ids can be checked through auth.json;
- Bedrock is available only when auth reports complete AWS credentials.

### `markConfiguredProviders(state map[string]bool, entries []model.Model)`

Logic:

- Mark all model-entry providers as available when an explicit CLI API key is
  supplied.
- Leave actual key consumption to `newProviderClient`.

Acceptance:

- `--api-key` can unlock model resolution before provider construction.

### `selectModel(registry model.Registry, authState map[string]bool, req cli.Request, settings providerSettingsFile) (model.Model, error)`

Logic:

- Honor explicit provider/model selectors first.
- Enforce provider match when both provider and provider-qualified model are
  supplied.
- Use provider `default_model` when a provider is supplied without a model.
- Use root saved `default_provider/default_model` when neither provider nor model is supplied.
- Prefer provider default aliases for provider-only selection.
- Prefer global `openai-default` before falling back to registry ordering.

Acceptance:

- provider/model selectors resolve the intended provider entry;
- provider mismatches fail clearly;
- saved defaults resolve before generic fallback ordering;
- default OpenAI selection is stable.

### `configuredDefault(registry model.Registry, authState map[string]bool, settings providerSettingsFile) (model.Model, bool)`

Logic:

- Resolve the saved root default provider/model when present.
- Fall back to saved provider-specific default when only provider is present.

Acceptance:

- Missing, unauthenticated, or deleted saved models are ignored and normal fallback continues.

### `configuredDefaultForProvider(registry model.Registry, authState map[string]bool, providerID string, settings providerSettingsFile) (model.Model, bool)`

Logic:

- Resolve one provider's saved `default_model`.

Acceptance:

- Provider-only startup can restore the last selected model for that provider.

### `defaultModelForProvider(registry model.Registry, authState map[string]bool, providerID string) (model.Model, bool)`

Logic:

- Return provider-specific default aliases where defined.
- Verify the resolved default still belongs to the requested provider.

Acceptance:

- `openai` provider-only selection resolves `openai-default`.

### `selectProviderModel(registry model.Registry, authState map[string]bool, providerID string, modelID string) (model.Model, error)`

Logic:

- Resolve `provider/model` for unqualified model ids.
- Resolve already qualified model ids and reject provider mismatches.

Acceptance:

- `--provider openai --model gpt-5.5` resolves OpenAI;
- `--provider openai --model anthropic/...` fails.

### `newProviderClient(ctx context.Context, opts Options, store auth.Store, req cli.Request, settings providerSettingsFile, selected model.Model) (provider.Streamer, error)`

Logic:

- Construct the concrete provider adapter for the selected model provider.
- Prefer a configured `base_url` from provider settings and build an OpenAI-compatible chat streamer.
- Use `auth_provider` from settings when supplied; otherwise use selected provider id.
- Resolve API-key providers through `providerAPIKey`.
- Resolve Bedrock credentials from the injected environment.
- Return a clear error for unsupported provider ids.

Acceptance:

- OpenAI, Anthropic, Google, and Bedrock selections create streamers;
- configured providers such as Fireworks and Geartech create an OpenAI-compatible chat streamer;
- URL-compatible `compat` selections can be recreated for interactive model switching;
- unsupported custom providers fail before network work.

### `providerAPIKey(ctx context.Context, store auth.Store, req cli.Request, providerID string) (string, error)`

Logic:

- Prefer explicit `req.APIKey`.
- Otherwise resolve the provider key from auth store.
- Error when the selected provider has no key.

Acceptance:

- CLI API keys override auth.json and environment.

### `bedrockCredentials(env []string) (bedrock.Credentials, error)`

Logic:

- Read AWS access key, secret key, session token, and region from injected env.
- Prefer `AWS_REGION` over `AWS_DEFAULT_REGION`.
- Require both access key and secret key.

Acceptance:

- Bedrock construction fails before signing when credentials are incomplete.

### `envMap(env []string) map[string]string`

Logic:

- Convert `KEY=value` entries into a lookup map.
- Ignore malformed entries without `=`.

Acceptance:

- runtime helpers use injected env instead of process globals.

### `firstNonEmpty(values ...string) string`

Logic:

- Return the first non-empty trimmed string.
- Return empty when all values are empty.

Acceptance:

- region fallback stays deterministic.

### `isProviderURL(value string) bool`

Logic:

- Parse the provider selector as a URL.
- Accept only `http` and `https` URLs with a host.

Acceptance:

- compat providers are selected only for explicit HTTP(S) endpoints.

### `newJSONSessionHeader(opts Options) (jsonSessionHeader, error)`

Logic:

- Use `opts.SessionID` when supplied.
- Generate a UUIDv4-like id when no session id is supplied.
- Default empty app version to `dev`.
- Fill the session header fields used by JSON mode.

Acceptance:

- produces a session header with id, schema, cwd, app, and app version.

### `writeJSONLine(w io.Writer, value any) error`

Logic:

- Marshal `value` with `encoding/json`.
- Write bytes plus one LF.
- Wrap marshal/write errors with context.

Acceptance:

- writes exactly one valid JSON object line per call.

Tests:

- `TestAppRunPrintModeUsesInjectedRuntime`
- `TestNUF170JSONModeStdoutIsOnlyJSONL`
- `TestJSONSessionHeaderDefaults`
- `TestJSONModeUsesBuiltinToolsByDefault`
- `TestPrintModeBuildsProviderFromCLI`
- `TestSelectModelUsesOpenAIDefaultWhenAPIKeyMarksAllProviders`
- `TestSelectModelUsesOpenAIDefaultForProviderOnly`
