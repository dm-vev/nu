package components

import (
	"github.com/dm-vev/nu/internal/tui/ansi"
	"github.com/dm-vev/nu/internal/tui/message"
	"strings"
	"testing"
)

func TestToolBlockToolBlockRendersBashCommandAndOutput(t *testing.T) {
	block := NewToolBlock(
		"bash",
		"call-1",
		`{"command":"pwd"}`,
		`{"output":"/tmp\n","exit_code":0}`,
		message.ToolSuccess,
		ToolBlockOptions{
			PaddingX:   1,
			PaddingY:   1,
			SuccessBg:  func(value string) string { return ansi.ToolSuccessBG + value + ansi.DefaultBG },
			TitleStyle: func(value string) string { return ansi.Bold + value + ansi.BoldOff },
		},
	)

	joined := strings.Join(block.Render(40), "\n")
	plain := ansi.Strip(joined)
	if !strings.Contains(plain, "$ pwd") || !strings.Contains(plain, "/tmp") {
		t.Fatalf("tool block = %q, want bash title and output", joined)
	}
	if !strings.Contains(joined, ansi.ToolSuccessBG) {
		t.Fatalf("tool block = %q, want success background", joined)
	}
}

func TestToolBlockToolBlockRendersFailedCommandWithErrorBackground(t *testing.T) {
	block := NewToolBlock(
		"bash",
		"call-1",
		`{"command":"false"}`,
		`{"output":"failed\n","exit_code":1}`,
		message.ToolSuccess,
		ToolBlockOptions{
			PaddingX: 1,
			ErrorBg:  func(value string) string { return ansi.ToolErrorBG + value + ansi.DefaultBG },
			ErrorStyle: func(value string) string {
				return ansi.Red + value + ansi.DefaultFG
			},
		},
	)

	joined := strings.Join(block.Render(40), "\n")
	if !strings.Contains(joined, ansi.ToolErrorBG) || !strings.Contains(joined, ansi.Red+"failed") {
		t.Fatalf("tool block = %q, want error background and text", joined)
	}
}

func TestToolBlockToolBlockRendersPatchDiffColors(t *testing.T) {
	block := NewToolBlock(
		"edit",
		"call-1",
		`{"path":"a.go"}`,
		`{"patch":"--- a.go\n+++ a.go\n@@\n-old\n+new\n"}`,
		message.ToolSuccess,
		ToolBlockOptions{
			AddedStyle:   func(value string) string { return ansi.Green + value + ansi.DefaultFG },
			RemovedStyle: func(value string) string { return ansi.Red + value + ansi.DefaultFG },
			ContextStyle: func(value string) string { return ansi.Muted + value + ansi.DefaultFG },
		},
	)

	joined := strings.Join(block.Render(60), "\n")
	if !strings.Contains(joined, ansi.Red+"-old") || !strings.Contains(joined, ansi.Green+"+new") {
		t.Fatalf("tool block = %q, want colored diff lines", joined)
	}
}
