package compact

import (
	"encoding/json"

	"nu/internal/session"
)

// Plan is the deterministic split between kept and compacted entries.
type Plan struct {
	Keep     []session.Entry
	Compact  []session.Entry
	CutIndex int
}

// BranchSummary describes the entries abandoned when moving between branches.
type BranchSummary struct {
	CommonAncestorID  string
	AbandonedEntryIDs []string
	Files             []string
}

// BuildPlan selects the newest suffix that fits the available context budget.
func BuildPlan(entries []session.Entry, contextWindow int, reserveTokens int) Plan {
	budget := contextWindow - reserveTokens
	if budget <= 0 {
		budget = 0
	}
	total := 0
	for _, entry := range entries {
		total += estimateTokens(entry)
	}
	if total <= budget {
		return Plan{Keep: append([]session.Entry(nil), entries...), CutIndex: 0}
	}

	// Walk backward so recent context survives before older entries are compacted.
	used := 0
	cut := len(entries)
	for i := len(entries) - 1; i >= 0; i-- {
		cost := estimateTokens(entries[i])
		if used+cost > budget {
			break
		}
		used += cost
		cut = i
	}
	// Tool results need their matching assistant tool-call context.
	cut = includeToolCall(entries, cut)
	return Plan{
		Keep:     append([]session.Entry(nil), entries[cut:]...),
		Compact:  append([]session.Entry(nil), entries[:cut]...),
		CutIndex: cut,
	}
}

func includeToolCall(entries []session.Entry, cut int) int {
	if cut <= 0 || cut >= len(entries) {
		return cut
	}
	first := metadata(entries[cut])
	if first.Role != "tool_result" || first.ToolCallID == "" {
		return cut
	}
	for i := cut - 1; i >= 0; i-- {
		item := metadata(entries[i])
		if item.Role == "assistant" && item.ToolCallID == first.ToolCallID {
			return i
		}
	}
	return cut
}

// BuildBranchSummary returns deterministic metadata for an abandoned branch.
func BuildBranchSummary(from []session.Entry, to []session.Entry) BranchSummary {
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
		for _, file := range metadata(entry).Files {
			if seenFiles[file] {
				continue
			}
			seenFiles[file] = true
			summary.Files = append(summary.Files, file)
		}
	}
	return summary
}

func commonAncestor(from []session.Entry, to []session.Entry) string {
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

type entryMetadata struct {
	Tokens     int      `json:"tokens"`
	Role       string   `json:"role"`
	ToolCallID string   `json:"tool_call_id"`
	Files      []string `json:"files"`
}

func metadata(entry session.Entry) entryMetadata {
	var meta entryMetadata
	_ = json.Unmarshal(entry.Payload, &meta)
	return meta
}

func estimateTokens(entry session.Entry) int {
	if tokens := metadata(entry).Tokens; tokens > 0 {
		return tokens
	}
	// ponytail: rough byte estimate; replace with tokenizer only when context cuts become inaccurate.
	tokens := len(entry.Payload) / 4
	if tokens < 1 {
		return 1
	}
	return tokens
}
