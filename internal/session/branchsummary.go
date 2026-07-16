package session

import "encoding/json"

// BranchSummary describes the entries abandoned when moving between branches.
type BranchSummary struct {
	CommonAncestorID  string
	AbandonedEntryIDs []string
	Files             []string
}

// BuildBranchSummary returns deterministic metadata for an abandoned branch.
func BuildBranchSummary(from []Entry, to []Entry) BranchSummary {
	ancestor := commonAncestor(from, to)
	summary := BranchSummary{CommonAncestorID: ancestor}
	start := 0
	for i, entry := range from {
		if entry.ID == ancestor {
			start = i + 1
			break
		}
	}
	seenFiles := map[string]bool{}
	for _, entry := range from[start:] {
		summary.AbandonedEntryIDs = append(summary.AbandonedEntryIDs, entry.ID)
		for _, file := range entryFiles(entry) {
			if seenFiles[file] {
				continue
			}
			seenFiles[file] = true
			summary.Files = append(summary.Files, file)
		}
	}
	return summary
}

func commonAncestor(from []Entry, to []Entry) string {
	limit := len(from)
	if len(to) < limit {
		limit = len(to)
	}
	ancestor := ""
	for i := 0; i < limit; i++ {
		if from[i].ID != to[i].ID {
			break
		}
		ancestor = from[i].ID
	}
	return ancestor
}

func entryFiles(entry Entry) []string {
	var meta struct {
		Files []string `json:"files"`
	}
	_ = json.Unmarshal(entry.Payload, &meta)
	return meta.Files
}
