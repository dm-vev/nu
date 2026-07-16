package gemini

import (
	"context"
	"testing"

	"google.golang.org/genai"

	"nu/internal/contracts"
)

// geminiFakeTool is a minimal contracts.Tool for exercising
// geminiConvertToolsToFunctionDeclarations without the rest of the SDK.
type geminiFakeTool struct {
	name   string
	params map[string]contracts.ParameterSpec
}

func (t *geminiFakeTool) Name() string                                    { return t.name }
func (t *geminiFakeTool) Description() string                             { return "" }
func (t *geminiFakeTool) Parameters() map[string]contracts.ParameterSpec  { return t.params }
func (t *geminiFakeTool) Execute(context.Context, string) (string, error) { return "", nil }
func (t *geminiFakeTool) Run(context.Context, string) (string, error)     { return "", nil }

// TestConvertTools_PropagatesArrayItems is a regression test for the bug
// where the streaming path dropped ParameterSpec.Items, causing Gemini to
// reject any MCP tool that exposed an "array" parameter with
// INVALID_ARGUMENT "properties[...].items: missing field".
func TestGeminiConvertTools_PropagatesArrayItems(t *testing.T) {
	tool := &geminiFakeTool{
		name: "publish",
		params: map[string]contracts.ParameterSpec{
			"labels": {
				Type: "array",
				Items: &contracts.ParameterSpec{
					Type: "string",
					Enum: []interface{}{"bug", "feature"},
				},
				Required: true,
			},
			"priority": {
				Type: "string",
				Enum: []interface{}{"low", "high"},
			},
		},
	}

	decls := geminiConvertToolsToFunctionDeclarations([]contracts.Tool{tool})
	if len(decls) != 1 {
		t.Fatalf("want 1 declaration, got %d", len(decls))
	}
	props := decls[0].Parameters.Properties
	labels := props["labels"]
	if labels == nil || labels.Type != genai.TypeArray {
		t.Fatalf("labels missing or wrong type: %+v", labels)
	}
	if labels.Items == nil || labels.Items.Type != genai.TypeString {
		t.Fatalf("labels.items dropped: %+v", labels.Items)
	}
	if len(labels.Items.Enum) != 2 {
		t.Fatalf("labels.items.enum dropped: %v", labels.Items.Enum)
	}
	if len(decls[0].Parameters.Required) != 1 || decls[0].Parameters.Required[0] != "labels" {
		t.Fatalf("required list wrong: %v", decls[0].Parameters.Required)
	}

	priority := props["priority"]
	if len(priority.Enum) != 2 {
		t.Fatalf("priority enum dropped: %v", priority.Enum)
	}
}

// TestConvertTools_ArrayWithoutItems_DefaultsToString covers MCP servers
// that advertise an `array` without `items` (common in the GitHub MCP
// server). Gemini requires items, so we default to string items instead
// of emitting an invalid schema.
func TestGeminiConvertTools_ArrayWithoutItems_DefaultsToString(t *testing.T) {
	tool := &geminiFakeTool{
		name: "noisy",
		params: map[string]contracts.ParameterSpec{
			"files": {Type: "array"},
		},
	}
	decls := geminiConvertToolsToFunctionDeclarations([]contracts.Tool{tool})
	files := decls[0].Parameters.Properties["files"]
	if files.Items == nil || files.Items.Type != genai.TypeString {
		t.Fatalf("expected default string items, got %+v", files.Items)
	}
}

// TestConvertTools_UnionTypes covers anyOf-style types represented as
// []string or []interface{} unions (e.g. ["array","null"]). The helper
// should pick the first non-null branch.
func TestGeminiConvertTools_UnionTypes(t *testing.T) {
	tool := &geminiFakeTool{
		name: "union",
		params: map[string]contracts.ParameterSpec{
			"maybeList": {
				Type: []string{"array", "null"},
				Items: &contracts.ParameterSpec{
					Type: "string",
				},
			},
			"maybeName": {
				Type: []interface{}{"null", "string"},
			},
		},
	}
	decls := geminiConvertToolsToFunctionDeclarations([]contracts.Tool{tool})
	props := decls[0].Parameters.Properties
	if props["maybeList"].Type != genai.TypeArray {
		t.Fatalf("maybeList type not resolved: %v", props["maybeList"].Type)
	}
	if props["maybeList"].Items == nil || props["maybeList"].Items.Type != genai.TypeString {
		t.Fatalf("maybeList.items dropped: %+v", props["maybeList"].Items)
	}
	if props["maybeName"].Type != genai.TypeString {
		t.Fatalf("maybeName type not resolved: %v", props["maybeName"].Type)
	}
}
