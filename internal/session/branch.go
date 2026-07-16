package session

import (
	"context"
	"fmt"
	"time"
)

// Fork writes a new session containing the path from root to entryID.
func (s *Store) Fork(ctx context.Context, source Ref, target Ref, entryID string) error {
	sourceSession, err := s.Load(ctx, source)
	if err != nil {
		return err
	}
	entries, err := PathTo(sourceSession.Tree, entryID)
	if err != nil {
		return err
	}
	return s.createFromEntries(ctx, target, sourceSession.Header, entries)
}

// Clone writes a new session containing the active branch.
func (s *Store) Clone(ctx context.Context, source Ref, target Ref) error {
	sourceSession, err := s.Load(ctx, source)
	if err != nil {
		return err
	}
	entries, err := PathTo(sourceSession.Tree, sourceSession.Tree.ActiveLeaf())
	if err != nil {
		return err
	}
	return s.createFromEntries(ctx, target, sourceSession.Header, entries)
}

func (s *Store) createFromEntries(ctx context.Context, target Ref, source Header, entries []Entry) error {
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("create session: %w", err)
	}
	if target.ID == "" {
		return fmt.Errorf("create session: missing target id")
	}
	header := source
	header.ID = target.ID
	header.CreatedAt = time.Now().UTC()
	return s.writeSession(target, header, entries)
}
