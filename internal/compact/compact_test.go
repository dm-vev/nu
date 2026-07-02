package compact

import (
	"encoding/json"
	"testing"
	"time"

	"nu/internal/session"
)

func TestNUF090CompactionKeepsRecentBudget(t *testing.T) {
	plan := BuildPlan([]session.Entry{
		compactEntry("e1", 40, "", "", nil),
		compactEntry("e2", 40, "", "", nil),
		compactEntry("e3", 40, "", "", nil),
	}, 100, 20)

	if len(plan.Compact) != 1 || plan.Compact[0].ID != "e1" {
		t.Fatalf("Compact = %#v, want e1", plan.Compact)
	}
	if len(plan.Keep) != 2 || plan.Keep[0].ID != "e2" || plan.Keep[1].ID != "e3" {
		t.Fatalf("Keep = %#v, want e2,e3", plan.Keep)
	}
}

func TestNUF090CompactionDoesNotCutBeforeToolResult(t *testing.T) {
	plan := BuildPlan([]session.Entry{
		compactEntry("e1", 50, "user", "", nil),
		compactEntry("e2", 40, "assistant", "call-1", nil),
		compactEntry("e3", 20, "tool_result", "call-1", nil),
		compactEntry("e4", 10, "assistant", "", nil),
	}, 40, 10)

	if len(plan.Keep) != 3 || plan.Keep[0].ID != "e2" || plan.Keep[1].ID != "e3" || plan.Keep[2].ID != "e4" {
		t.Fatalf("Keep = %#v, want tool call/result suffix", plan.Keep)
	}
}

func TestNUF091BranchSummaryFindsCommonAncestor(t *testing.T) {
	summary := BuildBranchSummary(
		[]session.Entry{compactEntry("root", 1, "", "", nil), compactEntry("left", 1, "", "", nil)},
		[]session.Entry{compactEntry("root", 1, "", "", nil), compactEntry("right", 1, "", "", nil)},
	)

	if summary.CommonAncestorID != "root" {
		t.Fatalf("CommonAncestorID = %q, want root", summary.CommonAncestorID)
	}
	if len(summary.AbandonedEntryIDs) != 1 || summary.AbandonedEntryIDs[0] != "left" {
		t.Fatalf("AbandonedEntryIDs = %v, want left", summary.AbandonedEntryIDs)
	}
}

func TestNUF091BranchSummaryTracksFiles(t *testing.T) {
	summary := BuildBranchSummary(
		[]session.Entry{
			compactEntry("root", 1, "", "", nil),
			compactEntry("left-1", 1, "", "", []string{"a.go", "b.go"}),
			compactEntry("left-2", 1, "", "", []string{"a.go", "c.go"}),
		},
		[]session.Entry{compactEntry("root", 1, "", "", nil), compactEntry("right", 1, "", "", nil)},
	)

	want := []string{"a.go", "b.go", "c.go"}
	if len(summary.Files) != len(want) {
		t.Fatalf("Files = %v, want %v", summary.Files, want)
	}
	for i := range want {
		if summary.Files[i] != want[i] {
			t.Fatalf("Files = %v, want %v", summary.Files, want)
		}
	}
}

func compactEntry(id string, tokens int, role string, toolCallID string, files []string) session.Entry {
	payload := map[string]any{"tokens": tokens}
	if role != "" {
		payload["role"] = role
	}
	if toolCallID != "" {
		payload["tool_call_id"] = toolCallID
	}
	if len(files) > 0 {
		payload["files"] = files
	}
	data, _ := json.Marshal(payload)
	return session.Entry{
		Type:      "entry",
		Schema:    1,
		ID:        id,
		CreatedAt: time.Now().UTC(),
		Kind:      session.KindMessage,
		Payload:   data,
	}
}
