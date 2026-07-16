# `internal/app/modelselection.go`

## Status

Current: IMPLEMENTED_UNCOMMITTED
Implementation Commit: -
Implementation Comments: Registry loading, auth availability, and deterministic default selection were split from runtime/provider construction.

## TODO

- [x] Match the implemented file boundary.
- [x] Confirm package tests cover the owned behavior.
- [x] Run the targeted package tests.
- [ ] Record the implementation commit after commit.

## Purpose

Resolve the model registry and select an available provider/model from CLI, auth, and saved settings.

## Code Style

Keep selection deterministic and side-effect free except for explicit auth resolution.

## Owned Logic

- Load built-in plus optional custom models and derive the default models path.
- Resolve provider auth once per provider and mark CLI-key providers available.
- Prefer explicit model/provider, saved defaults, the stable OpenAI default, then the first available model.
- Validate provider-qualified model selections and reject unavailable combinations.

## Acceptance

- List and runtime selection use the same registry/auth state.
- Saved and provider-scoped defaults are restored only when available.
- Missing or mismatched selections return contextual errors.

## Tests

- `TestListModelsUsesAuthState`
- `TestListModelsUsesCustomModelsPath`
- `TestSavedModelSelectionRestoresDefault`
- `TestSelectModelUsesOpenAIDefaultWhenAPIKeyMarksAllProviders`
- `TestSelectModelUsesOpenAIDefaultForProviderOnly`
