package toolblock

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

func (b *Block) formatContent() string {
	title, suppressArgs := b.title()
	parts := []string{b.opts.TitleStyle(title)}
	if !suppressArgs {
		if args := prettyJSON(b.arguments); args != "" {
			parts = append(parts, "", b.opts.TextStyle(args))
		}
	}
	if output := b.output(); output != "" {
		parts = append(parts, "", output)
	}
	return strings.Join(parts, "\n")
}

func (b *Block) title() (string, bool) {
	if b.toolName == "bash" {
		if command := stringField(b.arguments, "command"); command != "" {
			return "$ " + command, true
		}
	}
	if path := stringField(b.arguments, "path"); path != "" {
		return b.toolName + " " + path, false
	}
	if b.toolName != "" {
		return b.toolName, false
	}
	if b.toolID != "" {
		return b.toolID, false
	}
	return "tool", false
}

func (b *Block) output() string {
	values, ok := decodeObject(b.result)
	if !ok {
		return b.opts.TextStyle(strings.TrimSpace(b.result))
	}
	if patch, ok := values["patch"].(string); ok && strings.TrimSpace(patch) != "" {
		return renderDiff(patch, b.opts)
	}
	if output, ok := values["output"].(string); ok && output != "" {
		if b.resultLooksFailed() {
			return b.opts.ErrorStyle(strings.TrimRight(output, "\n"))
		}
		return b.opts.TextStyle(strings.TrimRight(output, "\n"))
	}
	stdout, _ := values["stdout"].(string)
	stderr, _ := values["stderr"].(string)
	combined := strings.TrimRight(stdout+stderr, "\n")
	if combined != "" {
		if b.resultLooksFailed() {
			return b.opts.ErrorStyle(combined)
		}
		return b.opts.TextStyle(combined)
	}
	return b.opts.TextStyle(prettyObject(values))
}

func (b *Block) resultLooksFailed() bool {
	values, ok := decodeObject(b.result)
	if !ok {
		return false
	}
	if timedOut, ok := values["timed_out"].(bool); ok && timedOut {
		return true
	}
	exitCode, ok := numericField(values, "exit_code")
	return ok && exitCode != 0
}

func stringField(raw string, key string) string {
	values, ok := decodeObject(raw)
	if !ok {
		return ""
	}
	value, _ := values[key].(string)
	return strings.TrimSpace(value)
}

func numericField(values map[string]any, key string) (int, bool) {
	switch value := values[key].(type) {
	case float64:
		return int(value), true
	case int:
		return value, true
	case json.Number:
		parsed, err := strconv.Atoi(value.String())
		return parsed, err == nil
	default:
		return 0, false
	}
}

func prettyJSON(raw string) string {
	values, ok := decodeObject(raw)
	if !ok {
		return strings.TrimSpace(raw)
	}
	return prettyObject(values)
}

func prettyObject(values map[string]any) string {
	if len(values) == 0 {
		return ""
	}
	data, err := json.MarshalIndent(values, "", "  ")
	if err != nil {
		return fmt.Sprint(values)
	}
	return string(data)
}

func decodeObject(raw string) (map[string]any, bool) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, false
	}
	var values map[string]any
	decoder := json.NewDecoder(strings.NewReader(raw))
	decoder.UseNumber()
	if err := decoder.Decode(&values); err != nil {
		return nil, false
	}
	return values, true
}
