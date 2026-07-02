package session

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestNUF080SessionAppendBuildsTree(t *testing.T) {
	store := OpenStore(t.TempDir())
	ref := Ref{ID: "s1"}
	root := testEntry("e1", "", KindMessage)
	child := testEntry("e2", "e1", KindMessage)

	if err := store.Append(context.Background(), ref, root); err != nil {
		t.Fatalf("Append root error = %v", err)
	}
	if err := store.Append(context.Background(), ref, child); err != nil {
		t.Fatalf("Append child error = %v", err)
	}

	sess, err := store.Load(context.Background(), ref)
	if err != nil {
		t.Fatalf("Load error = %v", err)
	}
	if sess.Tree.ActiveLeaf() != "e2" {
		t.Fatalf("ActiveLeaf = %q, want e2", sess.Tree.ActiveLeaf())
	}
	path, err := PathTo(sess.Tree, "e2")
	if err != nil {
		t.Fatalf("PathTo error = %v", err)
	}
	if len(path) != 2 || path[0].ID != "e1" || path[1].ID != "e2" {
		t.Fatalf("PathTo = %#v, want e1 -> e2", path)
	}
}

func TestNUF080SessionLoadRejectsBrokenParent(t *testing.T) {
	dir := t.TempDir()
	header := Header{
		Type:      "session",
		Schema:    schemaVersion,
		ID:        "s1",
		CreatedAt: time.Now().UTC(),
		App:       "nu",
	}
	headerData, err := json.Marshal(header)
	if err != nil {
		t.Fatalf("Marshal header error = %v", err)
	}
	entryData, err := MarshalEntry(testEntry("e1", "missing", KindMessage))
	if err != nil {
		t.Fatalf("MarshalEntry error = %v", err)
	}
	content := append(append(headerData, '\n'), append(entryData, '\n')...)
	if err := os.WriteFile(filepath.Join(dir, "s1.jsonl"), content, 0o600); err != nil {
		t.Fatalf("WriteFile error = %v", err)
	}

	_, err = OpenStore(dir).Load(context.Background(), Ref{ID: "s1"})
	if !errors.Is(err, ErrInvalidTree) {
		t.Fatalf("Load error = %v, want ErrInvalidTree", err)
	}
}

func TestNUF080SessionAppendRejectsBrokenParent(t *testing.T) {
	store := OpenStore(t.TempDir())
	err := store.Append(context.Background(), Ref{ID: "s1"}, testEntry("e1", "missing", KindMessage))
	if !errors.Is(err, ErrInvalidTree) {
		t.Fatalf("Append error = %v, want ErrInvalidTree", err)
	}
}

func TestSessionBuildTreeRejectsDuplicateID(t *testing.T) {
	_, err := BuildTree([]Entry{
		testEntry("e1", "", KindMessage),
		testEntry("e1", "", KindMessage),
	})
	if !errors.Is(err, ErrInvalidTree) {
		t.Fatalf("BuildTree error = %v, want ErrInvalidTree", err)
	}
}

func TestNUF082SelectingAssistantEntryMovesLeaf(t *testing.T) {
	tree, err := BuildTree([]Entry{
		testEntry("e1", "", KindMessage),
		testEntry("e2", "e1", KindMessage),
	})
	if err != nil {
		t.Fatalf("BuildTree error = %v", err)
	}
	if err := MoveLeaf(tree, "e1"); err != nil {
		t.Fatalf("MoveLeaf error = %v", err)
	}
	if tree.ActiveLeaf() != "e1" {
		t.Fatalf("ActiveLeaf = %q, want e1", tree.ActiveLeaf())
	}
}

func testEntry(id, parentID string, kind Kind) Entry {
	return Entry{
		Type:      "entry",
		Schema:    schemaVersion,
		ID:        id,
		ParentID:  parentID,
		CreatedAt: time.Now().UTC(),
		Kind:      kind,
		Payload:   json.RawMessage(`{}`),
	}
}
