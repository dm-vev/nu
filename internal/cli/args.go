package cli

import "strings"

type thinkingLevel string

const (
	thinkingOff     thinkingLevel = "off"
	thinkingMinimal thinkingLevel = "minimal"
	thinkingLow     thinkingLevel = "low"
	thinkingMedium  thinkingLevel = "medium"
	thinkingHigh    thinkingLevel = "high"
	thinkingXHigh   thinkingLevel = "xhigh"
)

var knownValueFlags = map[string]bool{
	"--mode":            true,
	"--provider":        true,
	"--model":           true,
	"--api-key":         true,
	"--thinking":        true,
	"--session":         true,
	"--session-id":      true,
	"--session-dir":     true,
	"--name":            true,
	"--models":          true,
	"--tools":           true,
	"--exclude-tools":   true,
	"--extension":       true,
	"--skill":           true,
	"--prompt-template": true,
	"--theme":           true,
}

var knownBoolFlags = map[string]bool{
	"--help":                true,
	"--version":             true,
	"--print":               true,
	"--continue":            true,
	"--resume":              true,
	"--fork":                true,
	"--no-session":          true,
	"--no-tools":            true,
	"--no-builtin-tools":    true,
	"--no-extensions":       true,
	"--no-skills":           true,
	"--no-prompt-templates": true,
	"--no-themes":           true,
	"--no-context-files":    true,
	"--list-models":         true,
	"--offline":             true,
	"--verbose":             true,
	"--approve":             true,
	"--no-approve":          true,
}

// Parse converts argv into a Request and non-fatal diagnostics.
func Parse(args []string) (Request, []Diagnostic) {
	req := Request{Command: CommandChat, Mode: ModeInteractive}
	var diagnostics []Diagnostic

	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch {
		case arg == "--help":
			req.Command = CommandHelp
		case arg == "--version":
			req.Command = CommandVersion
		case arg == "--print":
			req.Mode = ModePrint
		case arg == "--list-models":
			req.Command = CommandListModels
		case strings.HasPrefix(arg, "@") && len(arg) > 1:
			req.FileArgs = append(req.FileArgs, arg)
		case arg == "--mode":
			value, ok := nextValue(args, &i)
			if !ok {
				diagnostics = append(diagnostics, Diagnostic{Message: "--mode requires a value"})
				continue
			}
			mode, ok := parseMode(value)
			if !ok {
				diagnostics = append(diagnostics, Diagnostic{Message: "invalid --mode: " + value})
				continue
			}
			req.Mode = mode
		case arg == "--provider":
			value, ok := nextValue(args, &i)
			if !ok {
				diagnostics = append(diagnostics, Diagnostic{Message: "--provider requires a value"})
				continue
			}
			req.Provider = value
		case arg == "--model":
			value, ok := nextValue(args, &i)
			if !ok {
				diagnostics = append(diagnostics, Diagnostic{Message: "--model requires a value"})
				continue
			}
			req.Model = value
		case arg == "--api-key":
			value, ok := nextValue(args, &i)
			if !ok {
				diagnostics = append(diagnostics, Diagnostic{Message: "--api-key requires a value"})
				continue
			}
			req.APIKey = value
		case arg == "--models":
			value, ok := nextValue(args, &i)
			if !ok {
				diagnostics = append(diagnostics, Diagnostic{Message: "--models requires a value"})
				continue
			}
			req.ModelsPath = value
		case arg == "--thinking":
			value, ok := nextValue(args, &i)
			if !ok {
				diagnostics = append(diagnostics, Diagnostic{Message: "--thinking requires a value"})
				continue
			}
			if _, ok := parseThinkingLevel(value); !ok {
				diagnostics = append(diagnostics, Diagnostic{Message: "invalid --thinking: " + value})
			}
		case knownValueFlags[arg]:
			if _, ok := nextValue(args, &i); !ok {
				diagnostics = append(diagnostics, Diagnostic{Message: arg + " requires a value"})
			}
		case knownBoolFlags[arg]:
		case strings.HasPrefix(arg, "--"):
			// Preserve extension-owned flag values without interpreting their schema.
			flag := arg
			if !strings.Contains(arg, "=") && i+1 < len(args) && !strings.HasPrefix(args[i+1], "-") {
				i++
				flag += "=" + args[i]
			}
			req.ExtensionFlags = append(req.ExtensionFlags, flag)
		default:
			req.Prompt = append(req.Prompt, arg)
		}
	}

	return req, diagnostics
}

func nextValue(args []string, index *int) (string, bool) {
	if *index+1 >= len(args) {
		return "", false
	}
	*index++
	return args[*index], true
}

func parseMode(value string) (Mode, bool) {
	switch Mode(strings.ToLower(value)) {
	case ModeInteractive:
		return ModeInteractive, true
	case ModePrint:
		return ModePrint, true
	case ModeJSON:
		return ModeJSON, true
	case ModeRPC:
		return ModeRPC, true
	default:
		return "", false
	}
}

func parseThinkingLevel(value string) (thinkingLevel, bool) {
	switch thinkingLevel(strings.ToLower(strings.TrimSpace(value))) {
	case thinkingOff:
		return thinkingOff, true
	case thinkingMinimal:
		return thinkingMinimal, true
	case thinkingLow:
		return thinkingLow, true
	case thinkingMedium:
		return thinkingMedium, true
	case thinkingHigh:
		return thinkingHigh, true
	case thinkingXHigh:
		return thinkingXHigh, true
	default:
		return "", false
	}
}
