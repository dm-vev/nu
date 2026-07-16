# `internal/tools/coding/read.go`

## Status

Current: IMPLEMENTED_UNCOMMITTED
Implementation Commit: -
Implementation Comments: Read behavior is owned by `internal/tools/coding`.

## TODO

- [x] Match the implemented file boundary.
- [x] Confirm package tests cover the owned behavior.
- [x] Run the targeted package tests.
- [ ] Record the implementation commit after commit.

## Purpose

Read one cwd-contained text or supported image file.

## Code Style

Check containment/cancellation before IO and return JSON `Result` values.

## Owned Logic

- `RunRead` decodes path/offset/limit, rejects directories and invalid offsets, and reads bytes.
- Supported images return MIME plus base64 data; text applies offset/limit and output truncation.

## Acceptance

- Text slices and truncation are accurate.
- PNG/JPEG/GIF/WebP files return image attachments.
- Lexical and symlink cwd escapes fail.

## Tests

- `TestNUF070ReadTextWithOffsetLimit`
- `TestNUF070ReadTruncatesLargeFile`
- `TestNUF070ReadImageAttachment`
- `TestReadRejectsSymlinkEscape`
