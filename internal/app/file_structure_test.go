package app_test

import (
	"bytes"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func repositoryRoot(t *testing.T) string {
	t.Helper()
	_, testFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("locate repository root")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(testFile), "..", ".."))
}

func TestNUF212HierarchyHasNoOldPackagesOrFacade(t *testing.T) {
	root := repositoryRoot(t)
	for _, oldPath := range []string{
		"internal/auth",
		"internal/cli",
		"internal/agentconfig",
		"internal/executionplan",
		"internal/context",
		"internal/guardrails",
		"internal/prompts",
		"internal/agentsdk.go",
	} {
		if _, err := os.Stat(filepath.Join(root, oldPath)); !os.IsNotExist(err) {
			t.Errorf("superseded path still exists: %s", oldPath)
		}
	}

	files, err := filepath.Glob(filepath.Join(root, "internal/agent/*.go"))
	if err != nil {
		t.Fatal(err)
	}
	for _, file := range files {
		parsed, err := parser.ParseFile(token.NewFileSet(), file, nil, parser.ImportsOnly)
		if err != nil {
			t.Fatal(err)
		}
		for _, imported := range parsed.Imports {
			path := strings.Trim(imported.Path.Value, `"`)
			if strings.HasPrefix(path, "nu/internal/task") || strings.HasPrefix(path, "nu/internal/transport") {
				t.Errorf("agent has forbidden import %s: %s", path, file)
			}
		}
	}
}

func TestNUA009InternalPackagesMatchBalancedHierarchy(t *testing.T) {
	root := repositoryRoot(t)
	allowed := map[string]bool{
		"agent": true, "agentui": true, "app": true, "app/auth": true, "app/cli": true, "config": true,
		"agent/config": true, "agent/plans": true, "agent/guardrails": true, "agent/prompts": true,
		"contracts": true, "data": true, "llm": true,
		"mcp/builder": true, "mcp/client": true, "mcp/config": true, "mcp/fault": true,
		"mcp/lazy": true, "mcp/preset": true, "mcp/prompt": true, "mcp/registry": true,
		"mcp/resource": true, "mcp/retry": true, "mcp/sampling": true, "mcp/schema": true,
		"mcp/testkit": true, "mcp/tool": true, "mcp/transport": true,
		"data/embedding": true, "data/sql": true, "data/storage": true,
		"data/weaviate/graph": true, "data/weaviate/graph/entity": true,
		"data/weaviate/graph/extraction": true, "data/weaviate/graph/relationship": true,
		"data/weaviate/graph/search": true, "data/weaviate/vector": true,
		"data/weaviate/vector/metadata": true,
		"llm/anthropic":                 true, "llm/azureopenai": true, "llm/deepseek": true,
		"llm/gemini": true, "llm/ollama": true, "llm/openai": true, "llm/vllm": true,
		"memory/conversation": true, "memory/factory": true, "memory/history": true,
		"memory/redis": true, "memory/vector": true,
		"model": true, "multitenancy": true, "rpc": true,
		"session": true, "task": true, "telemetry": true, "telemetry/otel": true, "telemetry/langfuse": true, "testkit": true,
		"task/service": true, "task/workflow": true, "task/orchestration": true,
		"tools/agent": true, "tools/calculator": true, "tools/registry": true,
		"transport": true, "transport/remote": true,
		"transport/a2a/card": true, "transport/a2a/client": true, "transport/a2a/server": true,
		"transport/a2a/tool": true, "transport/grpc/client": true, "transport/grpc/server": true,
		"transport/grpc/microservice": true, "transport/grpc/pb": true,
		"transport/http/server": true, "transport/ui/server": true, "transport/ui/trace": true, "tui": true,
		"tui/ansi": true, "tui/core": true, "tui/editor": true, "tui/engine": true,
		"tui/input": true, "tui/message": true, "tui/terminal": true, "tui/components": true,
		"tools/coding": true, "tools/search": true, "tools/image": true, "tools/graphrag": true,
	}
	seen := map[string]bool{}
	err := filepath.WalkDir(filepath.Join(root, "internal"), func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}
		dir, err := filepath.Rel(filepath.Join(root, "internal"), filepath.Dir(path))
		if err != nil {
			return err
		}
		dir = filepath.ToSlash(dir)
		seen[dir] = true
		if !allowed[dir] {
			t.Errorf("unexpected internal package: %s", dir)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	for dir := range allowed {
		if dir == "data" || dir == "data/weaviate" { // Index-only roots intentionally have no production Go file.
			continue
		}
		if !seen[dir] {
			t.Errorf("missing internal package: %s", dir)
		}
	}
	for _, dir := range []string{
		"memory", "mcp", "tools", "transport/a2a", "transport/grpc", "transport/http", "transport/ui",
	} {
		if seen[dir] {
			t.Errorf("domain family root must not contain production Go files: %s", dir)
		}
	}
}

func TestNUF212AgentPackagesHaveFinalOwnership(t *testing.T) {
	root := repositoryRoot(t)
	agentRoot := filepath.Join(root, "internal", "agent")
	children := map[string]string{
		"config": "models.go", "plans": "plan.go",
		"guardrails": "guardrails.go", "prompts": "template.go",
	}

	for child, ordinaryFile := range children {
		childRoot := filepath.Join(agentRoot, child)
		if _, err := os.Stat(filepath.Join(childRoot, ordinaryFile)); err != nil {
			t.Errorf("missing ordinary agent filename %s/%s: %v", child, ordinaryFile, err)
		}
		files, err := filepath.Glob(filepath.Join(childRoot, "*.go"))
		if err != nil {
			t.Fatal(err)
		}
		for _, file := range files {
			name := filepath.Base(file)
			for _, prefix := range []string{"config_", "execution_plan_", "guardrail_", "prompt_"} {
				if strings.HasPrefix(name, prefix) {
					t.Errorf("agent child filename repeats package context: %s", file)
				}
			}
			parsed, err := parser.ParseFile(token.NewFileSet(), file, nil, parser.ImportsOnly)
			if err != nil {
				t.Fatal(err)
			}
			if parsed.Name.Name != child {
				t.Errorf("%s declares package %s, want %s", file, parsed.Name.Name, child)
			}
			for _, imported := range parsed.Imports {
				path := strings.Trim(imported.Path.Value, `"`)
				if path == "nu/internal/agent" || strings.HasPrefix(path, "nu/internal/transport") || strings.HasPrefix(path, "nu/internal/task") {
					t.Errorf("agent child has forbidden import %s: %s", path, file)
				}
			}
		}
	}
}

func TestNUF212TaskPackagesHaveFinalOwnership(t *testing.T) {
	root := repositoryRoot(t)
	taskRoot := filepath.Join(root, "internal", "task")
	allowedRootFiles := map[string]bool{
		"coremodels.go": true, "executor.go": true, "models.go": true,
		"planner.go": true, "simpleexecutor.go": true, "simpleplanner.go": true,
	}

	rootFiles, err := filepath.Glob(filepath.Join(taskRoot, "*.go"))
	if err != nil {
		t.Fatal(err)
	}
	for _, file := range rootFiles {
		if !strings.HasSuffix(file, "_test.go") && !allowedRootFiles[filepath.Base(file)] {
			t.Errorf("root task contains non-core implementation: %s", filepath.Base(file))
		}
		parsed, err := parser.ParseFile(token.NewFileSet(), file, nil, parser.ImportsOnly)
		if err != nil {
			t.Fatal(err)
		}
		for _, imported := range parsed.Imports {
			if strings.HasPrefix(strings.Trim(imported.Path.Value, `"`), "nu/internal/task/") {
				t.Errorf("root task must not import child packages: %s", file)
			}
		}
	}

	for child, ordinaryFile := range map[string]string{
		"service": "memory.go", "workflow": "models.go", "orchestration": "handoff.go",
	} {
		files, err := filepath.Glob(filepath.Join(taskRoot, child, "*.go"))
		if err != nil {
			t.Fatal(err)
		}
		if len(files) == 0 {
			t.Errorf("missing task child package: %s", child)
			continue
		}
		if _, err := os.Stat(filepath.Join(taskRoot, child, ordinaryFile)); err != nil {
			t.Errorf("missing ordinary %s filename %s: %v", child, ordinaryFile, err)
		}
		for _, file := range files {
			name := filepath.Base(file)
			if strings.HasPrefix(name, "service_") || strings.HasPrefix(name, "workflow_") || strings.HasPrefix(name, "orchestrator_") {
				t.Errorf("task child filename repeats package context: %s", file)
			}
			parsed, err := parser.ParseFile(token.NewFileSet(), file, nil, parser.PackageClauseOnly)
			if err != nil {
				t.Fatal(err)
			}
			if parsed.Name.Name != child {
				t.Errorf("%s declares package %s, want %s", file, parsed.Name.Name, child)
			}
		}
	}

	for _, file := range []string{"service/api.go", "orchestration/router.go"} {
		if _, err := os.Stat(filepath.Join(taskRoot, filepath.FromSlash(file))); err != nil {
			t.Errorf("missing ordinary task filename %s: %v", file, err)
		}
	}
}

func TestNUF204ToolRootDoesNotReExportChildPackages(t *testing.T) {
	root := repositoryRoot(t)
	files, err := filepath.Glob(filepath.Join(root, "internal/tools/*.go"))
	if err != nil {
		t.Fatal(err)
	}
	for _, file := range files {
		parsed, err := parser.ParseFile(token.NewFileSet(), file, nil, parser.ImportsOnly)
		if err != nil {
			t.Fatal(err)
		}
		for _, imported := range parsed.Imports {
			if strings.HasPrefix(strings.Trim(imported.Path.Value, `"`), "nu/internal/tools/") {
				t.Errorf("root tools must not import child packages: %s", file)
			}
		}
	}
}

func TestNUT009ProductionGoFilesAreAtMost300Lines(t *testing.T) {
	root := repositoryRoot(t)
	err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() && (entry.Name() == ".git" || entry.Name() == "vendor") {
			return filepath.SkipDir
		}
		if entry.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}

		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		if bytes.Contains(data, []byte("Code generated ")) && bytes.Contains(data, []byte("DO NOT EDIT.")) {
			return nil
		}
		lines := bytes.Count(data, []byte{'\n'})
		if len(data) > 0 && data[len(data)-1] != '\n' {
			lines++
		}
		if lines > 300 {
			t.Errorf("%s has %d lines; maximum is 300", path, lines)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestNUT010ProductionGoFilenamesHaveNoUnderscore(t *testing.T) {
	root := repositoryRoot(t)
	err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() && (entry.Name() == ".git" || entry.Name() == "vendor") {
			return filepath.SkipDir
		}
		if entry.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}
		if strings.Contains(filepath.Base(path), "_") {
			t.Errorf("production Go filename contains underscore: %s", path)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}
