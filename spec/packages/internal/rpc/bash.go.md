# `internal/rpc/bash.go`

## Status

Current: IMPLEMENTED_UNCOMMITTED
Implementation Commit: -
Implementation Comments: RPC bash adaptation was split from generic dispatch and delegates to the root Nu tool package.

## TODO

- [x] Match the implemented file boundary.
- [x] Confirm package tests cover the owned behavior.
- [x] Run the targeted package tests.
- [ ] Record the implementation commit after commit.

## Purpose

Adapt an RPC bash command to `tools.RunBash` and return decoded result data.

## Code Style

Keep execution policy in `internal/tools/coding`; this file owns only RPC argument/result conversion.

## Owned Logic

- Encode command text as tool arguments and run under server cwd with a 16 KiB display limit.
- Decode JSON tool content to protocol data, falling back to the raw content string.

## Acceptance

- RPC bash uses the same safety and timeout behavior as the built-in tool.
- Valid tool JSON is returned as structured response data.

## Tests

- `TestNUF171RPCRecognizesPiCommandSet`
