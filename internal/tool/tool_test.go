package tool

import (
	"context"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"nu/internal/agent"
)

func TestBuiltinsExposesEveryPhaseTwoTool(t *testing.T) {
	tools := Builtins(t.TempDir())
	var names []string
	for name := range tools {
		names = append(names, name)
	}
	sort.Strings(names)
	want := []string{"bash", "edit", "find", "grep", "ls", "read", "write"}
	if strings.Join(names, ",") != strings.Join(want, ",") {
		t.Fatalf("Builtins keys = %v, want %v", names, want)
	}
}

func TestBuiltinsReadToolRuns(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "a.txt"), []byte("ok"), 0o644); err != nil {
		t.Fatalf("WriteFile error = %v", err)
	}
	result, err := Builtins(dir)["read"](context.Background(), agent.ToolCall{Arguments: `{"path":"a.txt"}`})
	if err != nil {
		t.Fatalf("read tool error = %v", err)
	}
	if !strings.Contains(result.Content, `"content":"ok"`) {
		t.Fatalf("read tool content = %q, want ok", result.Content)
	}
}
