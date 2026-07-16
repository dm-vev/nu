package components

import (
	"nu/internal/tui/ansi"
	"strings"
	"testing"
)

func TestMarkdownMarkdownRendersInlineStyles(t *testing.T) {
	md := NewMarkdown("**bold** *em* `code`", MarkdownOptions{
		StrongStyle: func(value string) string { return ansi.Bold + value + ansi.BoldOff },
		EmphasisStyle: func(value string) string {
			return ansi.Italic + value + ansi.ItalicOff
		},
		CodeStyle: func(value string) string { return ansi.Green + value + ansi.DefaultFG },
	})

	joined := strings.Join(md.Render(80), "\n")
	for _, want := range []string{ansi.Bold + "bold", ansi.Italic + "em", ansi.Green + "code"} {
		if !strings.Contains(joined, want) {
			t.Fatalf("rendered markdown = %q, want %q", joined, want)
		}
	}
}

func TestMarkdownMarkdownRendersBlocksAndWraps(t *testing.T) {
	md := NewMarkdown("# Title\n\n- alpha beta gamma\n> quote\n```go\nfmt.Println()\n```", MarkdownOptions{})
	lines := md.Render(14)
	plain := ansi.Strip(strings.Join(lines, "\n"))

	for _, want := range []string{"Title", "- alpha", "│ quote", "fmt.Println"} {
		if !strings.Contains(plain, want) {
			t.Fatalf("lines = %#v, want %q", lines, want)
		}
	}
	for _, line := range lines {
		if ansi.VisibleWidth(line) > 14 {
			t.Fatalf("line width = %d for %q, want <= 14", ansi.VisibleWidth(line), line)
		}
	}
}

func TestMarkdownMarkdownRendersPipeTables(t *testing.T) {
	md := NewMarkdown("| Name | Age |\n| --- | ---: |\n| **Ann** | 7 |\n| Bob | 12 |", MarkdownOptions{
		StrongStyle: func(value string) string { return ansi.Bold + value + ansi.BoldOff },
	})

	lines := md.Render(40)
	plain := ansi.Strip(strings.Join(lines, "\n"))
	for _, want := range []string{"Name  Age", "Ann   7", "Bob   12"} {
		if !strings.Contains(plain, want) {
			t.Fatalf("table = %#v, want %q", plain, want)
		}
	}
	if strings.Contains(plain, "---") {
		t.Fatalf("table = %#v, separator row should not render", plain)
	}
	if !strings.Contains(strings.Join(lines, "\n"), ansi.Bold+"Ann") {
		t.Fatalf("table = %#v, want inline styling in cells", lines)
	}
}
