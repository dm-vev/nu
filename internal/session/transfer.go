package session

import (
	"context"
	"fmt"
	"io"
	"os"
)

// Export writes the validated JSONL session bytes to w.
func (s *Store) Export(ctx context.Context, ref Ref, w io.Writer) error {
	if _, err := s.Load(ctx, ref); err != nil {
		return err
	}
	data, err := os.ReadFile(s.path(ref))
	if err != nil {
		return fmt.Errorf("read session export %s: %w", s.path(ref), err)
	}
	if _, err := w.Write(data); err != nil {
		return fmt.Errorf("write session export: %w", err)
	}
	return nil
}

// Import validates JSONL session bytes and writes them as a new session file.
func (s *Store) Import(ctx context.Context, r io.Reader, target Ref) (Ref, error) {
	if err := ctx.Err(); err != nil {
		return Ref{}, fmt.Errorf("import session: %w", err)
	}
	// Import is a user-controlled input path, so bound memory before JSON validation.
	data, err := io.ReadAll(io.LimitReader(r, maxImportBytes+1))
	if err != nil {
		return Ref{}, fmt.Errorf("read session import: %w", err)
	}
	if len(data) > maxImportBytes {
		return Ref{}, fmt.Errorf("%w: limit %d bytes", ErrImportTooLarge, maxImportBytes)
	}
	header, entries, err := parseSession(data)
	if err != nil {
		return Ref{}, err
	}
	if target.ID == "" {
		target.ID = header.ID
	}
	header.ID = target.ID
	if err := s.writeSession(target, header, entries); err != nil {
		return Ref{}, err
	}
	return target, nil
}
