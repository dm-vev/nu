package session

import (
	"errors"
	"fmt"
)

// ErrInvalidTree is returned when entries cannot form a valid append-only tree.
var ErrInvalidTree = errors.New("invalid session tree")

// Tree is an in-memory view over session entries.
type Tree struct {
	entries    map[string]Entry
	parent     map[string]string
	children   map[string][]string
	roots      []string
	leaves     map[string]bool
	activeLeaf string
}

// BuildTree validates entries and builds tree indexes.
func BuildTree(entries []Entry) (*Tree, error) {
	tree := &Tree{
		entries:  make(map[string]Entry, len(entries)),
		parent:   make(map[string]string, len(entries)),
		children: make(map[string][]string, len(entries)),
		leaves:   make(map[string]bool, len(entries)),
	}

	for _, entry := range entries {
		if _, exists := tree.entries[entry.ID]; exists {
			return nil, fmt.Errorf("%w: duplicate id %s", ErrInvalidTree, entry.ID)
		}
		if entry.ParentID != "" {
			if _, exists := tree.entries[entry.ParentID]; !exists {
				return nil, fmt.Errorf("%w: missing parent %s", ErrInvalidTree, entry.ParentID)
			}
			tree.children[entry.ParentID] = append(tree.children[entry.ParentID], entry.ID)
			delete(tree.leaves, entry.ParentID)
		} else {
			tree.roots = append(tree.roots, entry.ID)
		}
		tree.entries[entry.ID] = entry
		tree.parent[entry.ID] = entry.ParentID
		tree.leaves[entry.ID] = true
		tree.activeLeaf = entry.ID
	}

	return tree, nil
}

// PathTo returns the ordered root-to-leaf path.
func PathTo(tree *Tree, leaf string) ([]Entry, error) {
	if tree == nil {
		return nil, fmt.Errorf("%w: nil tree", ErrInvalidTree)
	}
	if _, exists := tree.entries[leaf]; !exists {
		return nil, fmt.Errorf("%w: missing leaf %s", ErrInvalidTree, leaf)
	}

	var reversed []Entry
	seen := make(map[string]bool)
	for id := leaf; id != ""; id = tree.parent[id] {
		if seen[id] {
			return nil, fmt.Errorf("%w: cycle at %s", ErrInvalidTree, id)
		}
		seen[id] = true
		reversed = append(reversed, tree.entries[id])
	}

	path := make([]Entry, 0, len(reversed))
	for i := len(reversed) - 1; i >= 0; i-- {
		path = append(path, reversed[i])
	}
	return path, nil
}

// MoveLeaf changes the active leaf in memory only.
func MoveLeaf(tree *Tree, entryID string) error {
	if tree == nil {
		return fmt.Errorf("%w: nil tree", ErrInvalidTree)
	}
	if _, exists := tree.entries[entryID]; !exists {
		return fmt.Errorf("%w: missing entry %s", ErrInvalidTree, entryID)
	}
	tree.activeLeaf = entryID
	return nil
}

// ActiveLeaf returns the currently selected leaf id.
func (tree *Tree) ActiveLeaf() string {
	if tree == nil {
		return ""
	}
	return tree.activeLeaf
}
