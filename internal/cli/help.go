package cli

import (
	"runtime"
	"strings"
)

// ExtensionFlag describes a CLI flag contributed by an extension.
type ExtensionFlag struct {
	Name        string
	Description string
	Source      string
}

// VersionInfo contains build metadata for version output.
type VersionInfo struct {
	Name      string
	Version   string
	Commit    string
	BuildDate string
	GoVersion string
}

// Help renders stable CLI help text.
func Help(extensionFlags []ExtensionFlag) string {
	var b strings.Builder
	b.WriteString("Usage: nu [flags] [prompt]\n\n")
	b.WriteString("Modes:\n")
	b.WriteString("  --print              run once and print the result\n")
	b.WriteString("  --mode json          write JSONL events to stdout\n")
	b.WriteString("  --mode rpc           read JSONL commands from stdin\n\n")
	b.WriteString("Core flags:\n")
	b.WriteString("  --help               show help\n")
	b.WriteString("  --version            show version\n")
	b.WriteString("  --provider value     select provider\n")
	b.WriteString("  --model value        select model\n")
	b.WriteString("  --thinking value     off|minimal|low|medium|high|xhigh\n")
	b.WriteString("  --session value      resume session\n")
	b.WriteString("  --tools value        allow tools\n")
	b.WriteString("  --no-tools           disable tools\n")
	b.WriteString("  --no-context-files   skip context files\n")
	b.WriteString("  --approve            trust project for this run\n")
	b.WriteString("  --no-approve         block project resources for this run\n\n")
	b.WriteString("Commands:\n")
	b.WriteString("  package              manage resource packages\n")
	b.WriteString("  config               inspect or edit config\n")
	b.WriteString("  update               check or apply updates\n")
	b.WriteString("  export               export sessions\n")
	b.WriteString("  share                upload a private share artifact\n")
	b.WriteString("  list-models          list available models\n")

	if len(extensionFlags) > 0 {
		b.WriteString("\nExtension flags:\n")
		for _, flag := range extensionFlags {
			b.WriteString("  ")
			b.WriteString(flag.Name)
			if flag.Description != "" {
				b.WriteString("  ")
				b.WriteString(flag.Description)
			}
			if flag.Source != "" {
				b.WriteString(" (")
				b.WriteString(flag.Source)
				b.WriteString(")")
			}
			b.WriteString("\n")
		}
	}

	return b.String()
}

// Version renders stable one-line version text.
func Version(info VersionInfo) string {
	if info.Name == "" {
		info.Name = "nu"
	}
	if info.Version == "" {
		info.Version = "dev"
	}
	if info.GoVersion == "" {
		info.GoVersion = runtime.Version()
	}

	parts := []string{info.Name, info.Version}
	if info.Commit != "" {
		parts = append(parts, "commit="+info.Commit)
	}
	if info.BuildDate != "" {
		parts = append(parts, "built="+info.BuildDate)
	}
	parts = append(parts, "go="+info.GoVersion)
	return strings.Join(parts, " ")
}
