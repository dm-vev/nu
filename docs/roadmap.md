# Roadmap

Product scope is in `spec/functions.md`; this is implementation order.

## Completed Backend Connection

- imported the SDK source baseline at `v0.2.62` directly into `internal/`;
- rewrote package imports and preserved MIT attribution;
- replaced Nu provider/agent loop with SDK Agent and LLM packages;
- adapted Nu coding tools to SDK Tool;
- connected print, JSON, RPC, and TUI through `internal/agentui`;
- removed `internal/provider` and scripted-provider testkit.

## In Progress: Balanced Package Hierarchy

- preserve every feature/API behavior and owning test imported from pinned
  `v0.2.62`;
- migrate the temporary flat roots to the exact NUA-011 roots/subpackages;
- keep roots to shared types/orchestration, use normal filenames in subpackages,
  and reject one-helper packages;
- keep every TUI component in the single `tui/components` package;
- delete only superseded paths and add no compatibility wrappers;
- enforce SDK-to-Nu import direction and the production file-size rule;
- regenerate any imported protobuf affected by package/schema changes.

Exit: the full imported feature/test inventory remains covered, generated output
is reproducible under `transport/grpc/pb`, SDK code imports no Nu-owned package,
the exact package inventory matches, and full tests pass.

## Next: Session Memory

- adapt branchable Nu session active path to SDK Memory;
- preserve tool calls/results and compaction summaries;
- make `/new`, resume, fork, and clone operate on the same memory source.

Exit: SDK model context and Nu visible session tree cannot diverge.

## Next: SDK Configuration Exposure

- MCP connection config and trust UI;
- additional imported capability exposure only after its functional requirement
  is accepted and the ownership index is updated when package ownership changes.

Exit: each exposed SDK capability has trust, secret, stdout, and integration
tests. Package presence alone is not CLI exposure.

## Backend Updates

Upgrade only through the pinned procedure in `spec/backend.md`. Keep Nu patches
small and upstream-first; never recreate a removed provider or agent loop.
