# `internal/share/share.go`

## Status

Current: TODO
Implementation Commit: -
Implementation Comments: Not implemented yet.

## TODO

- [ ] Add or confirm the failing tests listed in this file.
- [ ] Implement the file according to the function logic below.
- [ ] Run the targeted package tests.
- [ ] After implementation commit, replace `Implementation Commit` with the commit hash and summarize important comments.

## Purpose

Create and upload private share artifacts after explicit user action.

## Code Style

No ambient uploads. Network client, target endpoint, session source, and user
confirmation are injected. Do not read credentials from process globals here.

## Functions

### `BuildArtifact(ctx context.Context, sess *session.Session, opts ArtifactOptions) (Artifact, error)`

Logic:

- Check `ctx` before walking large sessions.
- Select the active branch or explicit entry range requested by options.
- Redact secrets from tool details, environment snapshots, headers, and auth-like
  fields before serialization.
- Encode artifact metadata, messages, tool outputs, and attachments into a
  deterministic private artifact payload.

Acceptance:

- builds a deterministic artifact from a selected session branch;
- redacts secret-like fields before upload.

### `Upload(ctx context.Context, artifact Artifact, opts UploadOptions) (Result, error)`

Logic:

- Require an explicit target and explicit user confirmation token in options.
- Refuse upload when target, confirmation, or network client is missing.
- Send the artifact with the injected HTTP client and configured auth.
- Return share id/url and visibility metadata; never print or persist it here.

Acceptance:

- refuses upload without explicit target;
- refuses upload without explicit action confirmation;
- uses only the injected network client.

Tests:

- `TestNUF181ShareRequiresExplicitTarget`
- `TestNUF181ShareRequiresExplicitAction`
