package tui

import (
	"path/filepath"
	"strconv"
	"strings"
	"unicode/utf8"
)

const (
	ansiReset       = "\x1b[0m"
	ansiBold        = "\x1b[1m"
	ansiBoldOff     = "\x1b[22m"
	ansiDefaultFG   = "\x1b[39m"
	ansiText        = "\x1b[38;5;252m"
	ansiMuted       = "\x1b[38;5;244m"
	ansiDim         = "\x1b[38;5;241m"
	ansiBorder      = "\x1b[38;5;29m"
	ansiDarkGreen   = "\x1b[38;5;29m"
	ansiContext     = "\x1b[38;5;222m"
	defaultContext  = 128000
	resetSuffix     = ansiReset + "\x1b]8;;\x07"
	defaultHelpLine = "escape interrupt · ctrl+c/ctrl+d clear/exit · / commands · ! bash · ctrl+o more"
	defaultHintLine = "Press ctrl+o to show full startup help and loaded resources."
	defaultOnboard  = "Nu can explain its own features and look up its docs. Ask it how to use or extend Nu."
	ellipsis        = "…"
)

// Message is one UI-renderable chat message.
type Message struct {
	Role string
	Text string
}

// State is the complete render input for one frame.
type State struct {
	Title         string
	Version       string
	CWD           string
	Home          string
	Branch        string
	Provider      string
	Model         string
	ContextWindow int
	AutoCompact   bool
	ContextFiles  []string
	Status        string
	Messages      []Message
	Editor        EditorSnapshot
	Widgets       []string
	Overlays      []string
}

// Frame is a deterministic terminal render result.
type Frame struct {
	Lines     []string
	CursorRow int
	CursorCol int
	Width     int
	Title     string
}

// Render builds one terminal frame without writing it.
func Render(state State, width int, height int) Frame {
	if width < 1 {
		width = 1
	}
	if height < 1 {
		height = 1
	}

	lines := startupLines(state)
	for _, message := range state.Messages {
		lines = append(lines, renderMessage(message)...)
	}
	if state.Status != "" && state.Status != "idle" {
		lines = append(lines, ansiDim+"status: "+ansiText+singleLine(state.Status))
	}
	for _, widget := range state.Widgets {
		lines = append(lines, ansiDim+singleLine(widget))
	}
	if len(state.Overlays) > 0 {
		lines = append(lines, ansiDarkGreen+"overlay: "+ansiText+state.Overlays[len(state.Overlays)-1])
	}

	editorRow := len(lines) + 1
	lines = append(lines, ansiBorder+strings.Repeat("─", width))
	lines = append(lines, ansiText+singleLine(state.Editor.Text))
	lines = append(lines, ansiBorder+strings.Repeat("─", width))
	lines = append(lines, footerPathLine(state))
	lines = append(lines, footerStatsLine(state, width))

	if len(lines) > height {
		// Keep the editor and footer visible; old chat history can scroll off first.
		cut := len(lines) - height
		if cut > editorRow {
			cut = editorRow
		}
		lines = append(lines[:0], lines[cut:]...)
		editorRow -= cut
	}
	rendered := make([]string, len(lines))
	for i, line := range lines {
		rendered[i] = withReset(padANSI(truncateRunes(line, width), width))
	}
	cursorRow := editorRow
	return Frame{
		Lines:     rendered,
		CursorRow: clamp(cursorRow, 0, len(rendered)-1),
		CursorCol: clamp(state.Editor.Cursor+1, 1, width),
		Width:     width,
		Title:     windowTitle(state),
	}
}

// StripANSI removes escape sequences emitted by Render.
func StripANSI(text string) string {
	var out strings.Builder
	for i := 0; i < len(text); {
		if text[i] != 0x1b {
			out.WriteByte(text[i])
			i++
			continue
		}
		if i+1 < len(text) && text[i+1] == '[' {
			i += 2
			for i < len(text) {
				ch := text[i]
				i++
				if ch >= '@' && ch <= '~' {
					break
				}
			}
			continue
		}
		if i+1 < len(text) && text[i+1] == ']' {
			i += 2
			for i < len(text) && text[i] != 0x07 {
				i++
			}
			if i < len(text) {
				i++
			}
			continue
		}
		i++
	}
	return out.String()
}

func renderMessage(message Message) []string {
	role := strings.ToLower(strings.TrimSpace(message.Role))
	switch role {
	case "user":
		return []string{"", " " + ansiDarkGreen + singleLine(message.Text)}
	case "assistant":
		return []string{"", " " + ansiText + singleLine(message.Text)}
	default:
		return []string{"", ansiDim + role + ": " + ansiText + singleLine(message.Text)}
	}
}

func withReset(line string) string {
	return line + resetSuffix
}

func startupLines(state State) []string {
	title := firstNonEmpty(state.Title, "Nu")
	version := strings.TrimSpace(state.Version)
	logo := " " + ansiBold + ansiDarkGreen + title + ansiDefaultFG + ansiBoldOff
	if version != "" {
		logo += ansiDim + " v" + version + ansiDefaultFG
	}
	lines := []string{
		"",
		logo,
		" " + helpLine(),
		" " + ansiDim + defaultHintLine,
		"",
		" " + ansiDim + defaultOnboard,
		"",
		"",
	}
	files := state.ContextFiles
	if len(files) == 0 {
		files = []string{"AGENTS.md"}
	}
	lines = append(lines, ansiContext+"[Context]"+ansiDefaultFG)
	for _, file := range files {
		lines = append(lines, ansiDim+"  "+singleLine(file)+ansiDefaultFG)
	}
	lines = append(lines, "", "")
	return lines
}

func helpLine() string {
	parts := strings.Split(defaultHelpLine, " ")
	for i, part := range parts {
		if part == "escape" || part == "ctrl+c/ctrl+d" || part == "/" || part == "!" || part == "ctrl+o" {
			parts[i] = ansiDim + part + ansiDefaultFG
			continue
		}
		if part == "·" {
			parts[i] = ansiMuted + part + ansiDefaultFG
			continue
		}
		parts[i] = ansiMuted + part + ansiDefaultFG
	}
	return strings.Join(parts, " ")
}

func footerPathLine(state State) string {
	return ansiDim + footerPath(state) + ansiDefaultFG
}

func footerPath(state State) string {
	path := firstNonEmpty(state.CWD, ".")
	if state.Home != "" {
		home := filepath.Clean(state.Home)
		cwd := filepath.Clean(path)
		if cwd == home {
			path = "~"
		} else if rel, err := filepath.Rel(home, cwd); err == nil && rel != "." && !strings.HasPrefix(rel, "..") {
			path = "~" + string(filepath.Separator) + rel
		}
	}
	if strings.TrimSpace(state.Branch) != "" {
		path += " (" + strings.TrimSpace(state.Branch) + ")"
	}
	return path
}

func footerStatsLine(state State, width int) string {
	contextWindow := state.ContextWindow
	if contextWindow <= 0 {
		contextWindow = defaultContext
	}
	auto := ""
	if state.AutoCompact {
		auto = " (auto)"
	}
	left := "0.0%/" + formatTokens(contextWindow) + auto
	right := strings.Trim(strings.Join([]string{state.Provider, state.Model}, "/"), "/")
	if right == "" {
		return ansiDim + left + ansiDefaultFG
	}
	leftWidth := visibleLen(left)
	rightWidth := visibleLen(right)
	padding := 2
	if width > leftWidth+rightWidth {
		padding = width - leftWidth - rightWidth
	}
	return ansiDim + left + strings.Repeat(" ", padding) + right + ansiDefaultFG
}

func formatTokens(count int) string {
	if count < 1000 {
		return strconv.Itoa(count)
	}
	if count%1024 == 0 && count < 1024*1024 {
		return strconv.Itoa(count/1024) + "k"
	}
	if count < 1000000 {
		return strconv.Itoa(count/1000) + "k"
	}
	return strconv.Itoa(count/1000000) + "M"
}

func truncateRunes(text string, width int) string {
	if width <= 0 {
		return ""
	}
	if visibleLen(text) <= width {
		return text
	}
	limit := width
	if width > 1 {
		limit = width - 1
	}
	var out strings.Builder
	visible := 0
	for i := 0; i < len(text); {
		if text[i] == 0x1b {
			next := ansiEnd(text, i)
			out.WriteString(text[i:next])
			i = next
			continue
		}
		if visible >= limit {
			break
		}
		r, size := utf8.DecodeRuneInString(text[i:])
		out.WriteRune(r)
		visible++
		i += size
	}
	if width > 1 {
		out.WriteString(ellipsis)
	}
	return out.String()
}

func ansiEnd(text string, start int) int {
	if start+1 >= len(text) {
		return start + 1
	}
	if text[start+1] == '[' {
		i := start + 2
		for i < len(text) {
			ch := text[i]
			i++
			if ch >= '@' && ch <= '~' {
				return i
			}
		}
		return len(text)
	}
	if text[start+1] == ']' {
		i := start + 2
		for i < len(text) && text[i] != 0x07 {
			i++
		}
		if i < len(text) {
			return i + 1
		}
		return len(text)
	}
	return start + 1
}

func singleLine(text string) string {
	return strings.Join(strings.Fields(strings.ReplaceAll(text, "\n", " ")), " ")
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}

func padANSI(text string, width int) string {
	if width <= 0 {
		return ""
	}
	visible := visibleLen(text)
	if visible >= width {
		return text
	}
	return text + strings.Repeat(" ", width-visible)
}

func visibleLen(text string) int {
	visible := 0
	for i := 0; i < len(text); {
		if text[i] == 0x1b {
			i = ansiEnd(text, i)
			continue
		}
		_, size := utf8.DecodeRuneInString(text[i:])
		visible++
		i += size
	}
	return visible
}

func clamp(value int, low int, high int) int {
	if high < low {
		return low
	}
	if value < low {
		return low
	}
	if value > high {
		return high
	}
	return value
}

func windowTitle(state State) string {
	title := firstNonEmpty(state.Title, "Nu")
	base := filepath.Base(firstNonEmpty(state.CWD, "."))
	if base == "." || base == string(filepath.Separator) {
		return title
	}
	return title + " - " + base
}
