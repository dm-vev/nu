package session

import (
	"bytes"
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

func TestSessionExportImportRoundTrip(t *testing.T) {
	source := OpenStore(t.TempDir())
	if err := source.Append(context.Background(), Ref{ID: "s1"}, testEntry("e1", "", KindMessage)); err != nil {
		t.Fatalf("Append error = %v", err)
	}

	var out bytes.Buffer
	if err := source.Export(context.Background(), Ref{ID: "s1"}, &out); err != nil {
		t.Fatalf("Export error = %v", err)
	}

	target := OpenStore(t.TempDir())
	imported, err := target.Import(context.Background(), bytes.NewReader(out.Bytes()), Ref{ID: "copy"})
	if err != nil {
		t.Fatalf("Import error = %v", err)
	}
	loaded, err := target.Load(context.Background(), imported)
	if err != nil {
		t.Fatalf("Load imported error = %v", err)
	}
	if loaded.Header.ID != "copy" || len(loaded.Entries) != 1 || loaded.Entries[0].ID != "e1" {
		t.Fatalf("Imported session = %#v", loaded)
	}
}

func TestSessionAppendUsesRefCWD(t *testing.T) {
	store := OpenStore(t.TempDir())
	cwd := filepath.Join(t.TempDir(), "work")
	ref := Ref{ID: "s1", CWD: cwd}
	if err := store.Append(context.Background(), ref, testEntry("e1", "", KindMessage)); err != nil {
		t.Fatalf("Append error = %v", err)
	}

	latest, err := store.LatestByCWD(context.Background(), cwd)
	if err != nil {
		t.Fatalf("LatestByCWD error = %v", err)
	}
	if latest.ID != "s1" {
		t.Fatalf("LatestByCWD = %#v, want s1", latest)
	}
}

func TestSessionImportRejectsOversizedInput(t *testing.T) {
	store := OpenStore(t.TempDir())
	oversized := bytes.Repeat([]byte("x"), maxImportBytes+1)

	_, err := store.Import(context.Background(), bytes.NewReader(oversized), Ref{ID: "big"})
	if err == nil || !errors.Is(err, ErrImportTooLarge) {
		t.Fatalf("Import error = %v, want ErrImportTooLarge", err)
	}
	if _, statErr := os.Stat(filepath.Join(store.root, "big.jsonl")); !errors.Is(statErr, os.ErrNotExist) {
		t.Fatalf("Import created file for oversized input: %v", statErr)
	}
}

func TestNUF081ContinueLatestByCWD(t *testing.T) {
	dir := t.TempDir()
	cwd := filepath.Join(dir, "work")
	writeSession(t, dir, Header{
		Type:      "session",
		Schema:    schemaVersion,
		ID:        "old",
		CreatedAt: time.Date(2026, 7, 2, 10, 0, 0, 0, time.UTC),
		CWD:       cwd,
		App:       "nu",
	}, testEntry("e1", "", KindMessage))
	writeSession(t, dir, Header{
		Type:      "session",
		Schema:    schemaVersion,
		ID:        "new",
		CreatedAt: time.Date(2026, 7, 2, 11, 0, 0, 0, time.UTC),
		CWD:       cwd,
		App:       "nu",
	}, testEntry("e2", "", KindMessage))

	ref, err := OpenStore(dir).LatestByCWD(context.Background(), cwd)
	if err != nil {
		t.Fatalf("LatestByCWD error = %v", err)
	}
	if ref.ID != "new" {
		t.Fatalf("LatestByCWD = %#v, want new", ref)
	}
}

func TestNUF081ResumeByPathOrPartialID(t *testing.T) {
	dir := t.TempDir()
	header := Header{
		Type:      "session",
		Schema:    schemaVersion,
		ID:        "abcdef",
		CreatedAt: time.Now().UTC(),
		App:       "nu",
	}
	path := writeSession(t, dir, header, testEntry("e1", "", KindMessage))
	store := OpenStore(dir)

	byPath, err := store.Resolve(context.Background(), path)
	if err != nil {
		t.Fatalf("Resolve path error = %v", err)
	}
	if byPath.ID != "abcdef" || byPath.Path != path {
		t.Fatalf("Resolve path = %#v", byPath)
	}
	byPartial, err := store.Resolve(context.Background(), "abc")
	if err != nil {
		t.Fatalf("Resolve partial error = %v", err)
	}
	if byPartial.ID != "abcdef" || byPartial.Path != "" {
		t.Fatalf("Resolve partial = %#v", byPartial)
	}
}

func TestNUF081ForkStartsNewFileFromUserEntry(t *testing.T) {
	store := OpenStore(t.TempDir())
	ref := Ref{ID: "s1"}
	for _, entry := range []Entry{
		testEntry("e1", "", KindMessage),
		testEntry("e2", "e1", KindMessage),
		testEntry("e3", "e1", KindMessage),
	} {
		if err := store.Append(context.Background(), ref, entry); err != nil {
			t.Fatalf("Append %s error = %v", entry.ID, err)
		}
	}

	if err := store.Fork(context.Background(), ref, Ref{ID: "fork"}, "e2"); err != nil {
		t.Fatalf("Fork error = %v", err)
	}
	loaded, err := store.Load(context.Background(), Ref{ID: "fork"})
	if err != nil {
		t.Fatalf("Load fork error = %v", err)
	}
	if len(loaded.Entries) != 2 || loaded.Tree.ActiveLeaf() != "e2" {
		t.Fatalf("Fork entries/leaf = %d/%q", len(loaded.Entries), loaded.Tree.ActiveLeaf())
	}
}

func TestNUF081CloneCopiesActiveBranch(t *testing.T) {
	store := OpenStore(t.TempDir())
	ref := Ref{ID: "s1"}
	for _, entry := range []Entry{
		testEntry("e1", "", KindMessage),
		testEntry("e2", "e1", KindMessage),
		testEntry("e3", "e1", KindMessage),
	} {
		if err := store.Append(context.Background(), ref, entry); err != nil {
			t.Fatalf("Append %s error = %v", entry.ID, err)
		}
	}

	if err := store.Clone(context.Background(), ref, Ref{ID: "clone"}); err != nil {
		t.Fatalf("Clone error = %v", err)
	}
	loaded, err := store.Load(context.Background(), Ref{ID: "clone"})
	if err != nil {
		t.Fatalf("Load clone error = %v", err)
	}
	if len(loaded.Entries) != 2 || loaded.Entries[1].ID != "e3" {
		t.Fatalf("Clone entries = %#v", loaded.Entries)
	}
}

func TestSessionStateEntrySetsActiveLeaf(t *testing.T) {
	state := testEntry("state", "", KindExtension)
	state.Payload = json.RawMessage(`{"name":"session_state","active_leaf":"e2"}`)
	tree, err := BuildTree([]Entry{
		testEntry("e1", "", KindMessage),
		testEntry("e2", "e1", KindMessage),
		state,
	})
	if err != nil {
		t.Fatalf("BuildTree error = %v", err)
	}
	if tree.ActiveLeaf() != "e2" {
		t.Fatalf("ActiveLeaf = %q, want e2", tree.ActiveLeaf())
	}
}

func TestNUF080SessionAppendRejectsBrokenParent(t *testing.T) {
	store := OpenStore(t.TempDir())
	err := store.Append(context.Background(), Ref{ID: "s1"}, testEntry("e1", "missing", KindMessage))
	if !errors.Is(err, ErrInvalidTree) {
		t.Fatalf("Append error = %v, want ErrInvalidTree", err)
	}
}

func TestSessionAppendRejectsDuplicateID(t *testing.T) {
	store := OpenStore(t.TempDir())
	ref := Ref{ID: "s1"}
	entry := testEntry("e1", "", KindMessage)
	if err := store.Append(context.Background(), ref, entry); err != nil {
		t.Fatalf("Append first error = %v", err)
	}
	err := store.Append(context.Background(), ref, entry)
	if !errors.Is(err, ErrInvalidTree) {
		t.Fatalf("Append duplicate error = %v, want ErrInvalidTree", err)
	}
}

func TestSessionEntryPreservesExtraFields(t *testing.T) {
	entry := testEntry("e1", "", KindMessage)
	entry.Extra = map[string]json.RawMessage{"future": json.RawMessage(`{"ok":true}`)}
	data, err := MarshalEntry(entry)
	if err != nil {
		t.Fatalf("MarshalEntry error = %v", err)
	}
	roundTrip, err := UnmarshalEntry(data)
	if err != nil {
		t.Fatalf("UnmarshalEntry error = %v", err)
	}
	if string(roundTrip.Extra["future"]) != `{"ok":true}` {
		t.Fatalf("Extra = %s, want future field", roundTrip.Extra["future"])
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

func writeSession(t *testing.T, dir string, header Header, entries ...Entry) string {
	t.Helper()
	data, err := json.Marshal(header)
	if err != nil {
		t.Fatalf("Marshal header error = %v", err)
	}
	data = append(data, '\n')
	for _, entry := range entries {
		entryData, err := MarshalEntry(entry)
		if err != nil {
			t.Fatalf("MarshalEntry error = %v", err)
		}
		data = append(data, append(entryData, '\n')...)
	}
	path := filepath.Join(dir, header.ID+".jsonl")
	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatalf("WriteFile error = %v", err)
	}
	return path
}
