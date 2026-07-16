# `internal/app/provider.go`

## Status

Current: IMPLEMENTED_UNCOMMITTED
Implementation Commit: -
Implementation Comments: Provider selection and SDK LLM construction now have one composition-boundary owner.

## TODO

- [x] Match the implemented file boundary.
- [x] Confirm package tests cover the owned behavior.
- [x] Run the targeted package tests.
- [ ] Record the implementation commit after commit.

## Purpose

Configure the selected provider and construct the matching imported SDK LLM client.

## Code Style

Keep credentials explicit, diagnostics off protocol stdout, and provider-specific SDK imports at the app boundary.

## Owned Logic

- `configureProvider` preserves injected runners/LLMs, handles OpenAI-compatible URLs, resolves registry/auth/settings, and fills runtime model metadata.
- `newSDKLLM` constructs OpenAI, Anthropic, Gemini, Bedrock, Fireworks, custom-base-URL, or compatibility clients.
- Credential helpers prefer a CLI key, resolve stored auth, parse injected AWS env, and default Bedrock region to `us-east-1`.
- URL/env/string helpers normalize provider inputs.

## Acceptance

- Provider URLs require a model and use the compatibility client.
- Selected providers receive the correct credentials, model, base URL, and stdout-safe logger.
- Unsupported or unauthenticated providers fail before prompting.

## Tests

- `TestPrintModeBuildsProviderFromCLI`
- `TestPrintModeBuildsFireworksProviderFromGlobalModels`
- `TestPrintModeBuildsProviderFromSettings`
