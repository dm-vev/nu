package session

import "encoding/json"

// Plan is the deterministic split between kept and compacted entries.
type Plan struct {
	Keep     []Entry
	Compact  []Entry
	CutIndex int
}

// BuildPlan selects the newest suffix that fits the available context budget.
func BuildPlan(entries []Entry, contextWindow int, reserveTokens int) Plan {
	budget := contextWindow - reserveTokens
	if budget <= 0 {
		budget = 0
	}
	total := 0
	for _, entry := range entries {
		total += estimateTokens(entry)
	}
	if total <= budget {
		return Plan{Keep: append([]Entry(nil), entries...), CutIndex: 0}
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
		Keep:     append([]Entry(nil), entries[cut:]...),
		Compact:  append([]Entry(nil), entries[:cut]...),
		CutIndex: cut,
	}
}

func includeToolCall(entries []Entry, cut int) int {
	if cut <= 0 || cut >= len(entries) {
		return cut
	}
	first := compactionMetadata(entries[cut])
	if first.Role != "tool_result" || first.ToolCallID == "" {
		return cut
	}
	for i := cut - 1; i >= 0; i-- {
		item := compactionMetadata(entries[i])
		if item.Role == "assistant" && item.ToolCallID == first.ToolCallID {
			return i
		}
	}
	return cut
}

type entryCompactionMetadata struct {
	Tokens     int    `json:"tokens"`
	Role       string `json:"role"`
	ToolCallID string `json:"tool_call_id"`
}

func compactionMetadata(entry Entry) entryCompactionMetadata {
	var meta entryCompactionMetadata
	_ = json.Unmarshal(entry.Payload, &meta)
	return meta
}

func estimateTokens(entry Entry) int {
	if tokens := compactionMetadata(entry).Tokens; tokens > 0 {
		return tokens
	}
	// ponytail: rough byte estimate; replace with tokenizer only when context cuts become inaccurate.
	tokens := len(entry.Payload) / 4
	if tokens < 1 {
		return 1
	}
	return tokens
}
