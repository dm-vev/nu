package tui

import "github.com/dm-vev/nu/internal/tui/ansi"

func green(value string) string {
	return ansi.Green + value + ansi.DefaultFG
}

func greenBold(value string) string {
	return ansi.Green + ansi.Bold + value + ansi.BoldOff + ansi.DefaultFG
}

func red(value string) string {
	return ansi.Red + value + ansi.DefaultFG
}

func dim(value string) string {
	return ansi.Dim + value + ansi.BoldOff + ansi.DefaultFG
}

func muted(value string) string {
	return ansi.Muted + value + ansi.BoldOff + ansi.DefaultFG
}

func ansiText(value string) string {
	return ansi.Text + value + ansi.BoldOff + ansi.DefaultFG
}

func boldText(value string) string {
	return ansi.Text + ansi.Bold + value + ansi.BoldOff + ansi.DefaultFG
}

func italicText(value string) string {
	return ansi.Text + ansi.Italic + value + ansi.ItalicOff + ansi.DefaultFG
}

func inlineCode(value string) string {
	return ansi.Yellow + value + ansi.DefaultFG
}

func thinkingText(value string) string {
	return ansi.Dim + ansi.Italic + value + ansi.ItalicOff + ansi.BoldOff + ansi.DefaultFG
}

func thinkingStrong(value string) string {
	return ansi.Dim + ansi.Italic + ansi.Bold + value + ansi.BoldOff + ansi.ItalicOff + ansi.DefaultFG
}

func userBackground(value string) string {
	return ansi.UserMessageBG + value + ansi.DefaultBG
}

func toolPendingBackground(value string) string {
	return ansi.ToolPendingBG + value + ansi.DefaultBG
}

func toolSuccessBackground(value string) string {
	return ansi.ToolSuccessBG + value + ansi.DefaultBG
}

func toolErrorBackground(value string) string {
	return ansi.ToolErrorBG + value + ansi.DefaultBG
}
