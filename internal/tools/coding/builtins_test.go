package coding

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

func TestBuiltinsExposesEveryPhaseTwoTool(t *testing.T) {
	tools := Builtins(t.TempDir())
	var names []string
	for _, tool := range tools {
		names = append(names, tool.Name())
	}
	sort.Strings(names)
	want := []string{"bash", "edit", "find", "grep", "ls", "read", "write"}
	if strings.Join(names, ",") != strings.Join(want, ",") {
		t.Fatalf("Builtins keys = %v, want %v", names, want)
	}
}

func TestDefinitionsExposeBashSchema(t *testing.T) {
	tools := Builtins(t.TempDir())
	for _, tool := range tools {
		if tool.Name() != "bash" {
			continue
		}
		if tool.Description() == "" || !tool.Parameters()["command"].Required {
			t.Fatalf("bash definition = %#v, want required command", tool)
		}
		return
	}
	t.Fatal("Builtins() missing bash")
}

func TestBuiltinsReadToolRuns(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "a.txt"), []byte("ok"), 0o644); err != nil {
		t.Fatalf("WriteFile error = %v", err)
	}
	var readTool interface {
		Execute(context.Context, string) (string, error)
	}
	for _, candidate := range Builtins(dir) {
		if candidate.Name() == "read" {
			readTool = candidate
		}
	}
	if readTool == nil {
		t.Fatal("read tool missing")
	}
	result, err := readTool.Execute(context.Background(), `{"path":"a.txt"}`)
	if err != nil {
		t.Fatalf("read tool error = %v", err)
	}
	if !strings.Contains(result, `"content":"ok"`) {
		t.Fatalf("read tool content = %q, want ok", result)
	}
}

func decodeResult(t *testing.T, raw string) map[string]any {
	t.Helper()
	var got map[string]any
	if err := json.Unmarshal([]byte(raw), &got); err != nil {
		t.Fatalf("result JSON error = %v; raw=%q", err, raw)
	}
	return got
}

func mustRead(t *testing.T, path string) []byte {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile error = %v", err)
	}
	return data
}

func toStrings(t *testing.T, raw any) []string {
	t.Helper()
	values, ok := raw.([]any)
	if !ok {
		t.Fatalf("value = %#v, want []any", raw)
	}
	out := make([]string, 0, len(values))
	for _, value := range values {
		text, ok := value.(string)
		if !ok {
			t.Fatalf("item = %#v, want string", value)
		}
		out = append(out, text)
	}
	return out
}
