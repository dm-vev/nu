package coding

import (
	"context"

	"github.com/dm-vev/nu/contracts"
)

const defaultMaxOutputBytes = 16 * 1024

type codingTool struct {
	name        string
	description string
	parameters  map[string]contracts.ParameterSpec
	execute     func(context.Context, string) (string, error)
}

func (t codingTool) Name() string                                   { return t.name }
func (t codingTool) Description() string                            { return t.description }
func (t codingTool) Parameters() map[string]contracts.ParameterSpec { return t.parameters }
func (t codingTool) Run(ctx context.Context, input string) (string, error) {
	return t.execute(ctx, input)
}
func (t codingTool) Execute(ctx context.Context, input string) (string, error) {
	return t.execute(ctx, input)
}

func Builtins(cwd string) []contracts.Tool {
	return []contracts.Tool{
		codingTool{name: "bash", description: "Run a shell command in the current project directory.", parameters: map[string]contracts.ParameterSpec{
			"command": {Type: "string", Required: true}, "timeout_ms": {Type: "integer"},
		}, execute: func(ctx context.Context, input string) (string, error) {
			result, err := RunBash(ctx, cwd, input, defaultMaxOutputBytes)
			return result.Content, err
		}},
		codingTool{name: "read", description: "Read a file under the current project directory.", parameters: map[string]contracts.ParameterSpec{
			"path": {Type: "string", Required: true}, "offset": {Type: "integer"}, "limit": {Type: "integer"},
		}, execute: func(ctx context.Context, input string) (string, error) {
			result, err := RunRead(ctx, cwd, input, defaultMaxOutputBytes)
			return result.Content, err
		}},
		codingTool{name: "write", description: "Create or overwrite a file under the current project directory.", parameters: map[string]contracts.ParameterSpec{
			"path": {Type: "string", Required: true}, "content": {Type: "string", Required: true},
		}, execute: func(ctx context.Context, input string) (string, error) {
			result, err := RunWrite(ctx, cwd, input)
			return result.Content, err
		}},
		codingTool{name: "edit", description: "Apply exact text replacements to a file. Each replacement contains old and new strings.", parameters: map[string]contracts.ParameterSpec{
			"path":         {Type: "string", Required: true},
			"replacements": {Type: "array", Required: true, Items: &contracts.ParameterSpec{Type: "object"}},
		}, execute: func(ctx context.Context, input string) (string, error) {
			result, err := RunEdit(ctx, cwd, input)
			return result.Content, err
		}},
		codingTool{name: "grep", description: "Search files under the current project directory.", parameters: map[string]contracts.ParameterSpec{
			"pattern": {Type: "string", Required: true}, "literal": {Type: "boolean"}, "ignore_case": {Type: "boolean"},
			"glob": {Type: "string"}, "root": {Type: "string"}, "limit": {Type: "integer"},
		}, execute: func(ctx context.Context, input string) (string, error) {
			result, err := RunGrep(ctx, cwd, input, defaultMaxOutputBytes)
			return result.Content, err
		}},
		codingTool{name: "find", description: "Find files under the current project directory.", parameters: map[string]contracts.ParameterSpec{
			"root": {Type: "string"}, "glob": {Type: "string"}, "limit": {Type: "integer"},
		}, execute: func(ctx context.Context, input string) (string, error) {
			result, err := RunFind(ctx, cwd, input, defaultMaxOutputBytes)
			return result.Content, err
		}},
		codingTool{name: "ls", description: "List a directory under the current project directory.", parameters: map[string]contracts.ParameterSpec{
			"path": {Type: "string"},
		}, execute: func(ctx context.Context, input string) (string, error) {
			result, err := RunLS(ctx, cwd, input, defaultMaxOutputBytes)
			return result.Content, err
		}},
	}
}
