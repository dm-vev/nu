package cli

import "testing"

func TestNUF001ParseKnownFlags(t *testing.T) {
	req, diagnostics := Parse([]string{"--print", "--mode", "json", "--thinking", "high", "@README.md", "hello"})
	if len(diagnostics) != 0 {
		t.Fatalf("Parse diagnostics = %v, want none", diagnostics)
	}
	if req.Mode != ModeJSON {
		t.Fatalf("Request mode = %q, want %q", req.Mode, ModeJSON)
	}
	if len(req.FileArgs) != 1 || req.FileArgs[0] != "@README.md" {
		t.Fatalf("Request file args = %v, want @README.md", req.FileArgs)
	}
	if len(req.Prompt) != 1 || req.Prompt[0] != "hello" {
		t.Fatalf("Request prompt = %v, want hello", req.Prompt)
	}
}

func TestNUF001UnknownFlagsArePreservedForExtensions(t *testing.T) {
	req, diagnostics := Parse([]string{"--extension-flag", "value", "--x=1", "prompt"})
	if len(diagnostics) != 0 {
		t.Fatalf("Parse diagnostics = %v, want none", diagnostics)
	}
	if len(req.ExtensionFlags) != 2 || req.ExtensionFlags[0] != "--extension-flag=value" || req.ExtensionFlags[1] != "--x=1" {
		t.Fatalf("Request extension flags = %v, want extension flags with values", req.ExtensionFlags)
	}
	if len(req.Prompt) != 1 || req.Prompt[0] != "prompt" {
		t.Fatalf("Request prompt = %v, want prompt", req.Prompt)
	}
}

func TestNUF001InvalidThinkingLevelReportsDiagnostic(t *testing.T) {
	_, diagnostics := Parse([]string{"--thinking", "huge"})
	if len(diagnostics) != 1 {
		t.Fatalf("Parse diagnostics count = %d, want 1", len(diagnostics))
	}
}
