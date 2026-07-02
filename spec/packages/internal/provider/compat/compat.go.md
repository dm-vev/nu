# `internal/provider/compat/compat.go`

## Status

Current: IMPLEMENTED
Implementation Commit: 4ddd508
Implementation Comments: Phase 3 compat adapter wraps OpenAI-compatible Chat Completions endpoints.

## TODO

- [x] Add or confirm the failing tests listed in this file.
- [x] Implement the file according to the function logic below.
- [x] Run the targeted package tests.
- [x] After implementation commit, replace `Implementation Commit` with the commit hash and summarize important comments.

## Purpose

Support custom OpenAI-compatible providers by reusing the OpenAI Chat
Completions adapter against a custom base URL.

## Functions

### `New(baseURL string, apiKey string) provider.Streamer`

Logic:

- Return an OpenAI Chat Completions adapter configured with the custom base URL.

Acceptance:

- no duplicate OpenAI-compatible request/parser code.

Tests:

- covered by OpenAI Chat adapter tests.
