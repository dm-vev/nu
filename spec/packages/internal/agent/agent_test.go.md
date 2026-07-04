# `internal/agent/agent_test.go`

## Status

Current: IMPLEMENTED_UNCOMMITTED
Implementation Commit: pending TUI commit; user requested no commits until TUI is complete.
Implementation Comments: Agent behavior tests, including prompt history and provider/model switching.

## Purpose

Verify agent prompt execution, tool loop behavior, abort/retry paths, and model/provider mutation without real provider calls.

## Acceptance Criteria

- Tests use fake providers only.
- Prompt history tests prove later prompts include prior user and assistant messages.
- Model switch tests prove later prompts use updated labels and, when supplied, the updated streamer.

## Tests

### `TestAgentPromptIncludesPreviousTurns`

Acceptance:
- The second prompt request includes the previous user prompt and assistant answer before the current user prompt.

### `TestAgentSetModelAffectsNextPrompt`

Acceptance:
- Later prompts use updated provider/api/model labels.

### `TestAgentSetProviderModelSwitchesStreamer`

Acceptance:
- Later prompts are sent to the replacement streamer, not the previous streamer.
