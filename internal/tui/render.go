package tui

import "strings"

const resetSuffix = "\x1b[0m\x1b]8;;\x07"

// Message is one UI-renderable chat message.
type Message struct {
	Role string
	Text string
}

// State is the complete render input for one frame.
type State struct {
	Title    string
	CWD      string
	Provider string
	Model    string
	Status   string
	Messages []Message
	Editor   EditorSnapshot
	Widgets  []string
	Overlays []string
}

// Frame is a deterministic terminal render result.
type Frame struct {
	Lines     []string
	CursorRow int
	CursorCol int
}

// Render builds one terminal frame without writing it.
func Render(state State, width int, height int) Frame {
	if width < 1 {
		width = 1
	}
	if height < 1 {
		height = 1
	}

	lines := []string{firstNonEmpty(state.Title, "Nu")}
	for _, message := range state.Messages {
		lines = append(lines, message.Role+": "+singleLine(message.Text))
	}
	if state.Status != "" {
		lines = append(lines, "status: "+singleLine(state.Status))
	}
	for _, widget := range state.Widgets {
		lines = append(lines, singleLine(widget))
	}
	if len(state.Overlays) > 0 {
		lines = append(lines, "overlay: "+state.Overlays[len(state.Overlays)-1])
	}
	lines = append(lines, "> "+singleLine(state.Editor.Text))
	lines = append(lines, footerLine(state))

	if len(lines) > height {
		// Keep the top of the frame stable; later raw-terminal diffing can optimize scrollback.
		lines = lines[:height]
	}
	rendered := make([]string, len(lines))
	for i, line := range lines {
		rendered[i] = withReset(truncateRunes(line, width))
	}
	cursorRow := len(rendered) - 2
	if cursorRow < 0 {
		cursorRow = len(rendered) - 1
	}
	if cursorRow < 0 {
		cursorRow = 0
	}
	cursorCol := 2 + state.Editor.Cursor
	if cursorCol >= width {
		cursorCol = width - 1
	}
	return Frame{Lines: rendered, CursorRow: cursorRow, CursorCol: cursorCol}
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

func footerLine(state State) string {
	left := firstNonEmpty(state.CWD, ".")
	right := strings.TrimSpace(strings.Join([]string{state.Provider, state.Model}, "/"))
	if right == "/" {
		return left
	}
	return left + "  " + right
}

func withReset(line string) string {
	return line + resetSuffix
}

func truncateRunes(text string, width int) string {
	runes := []rune(text)
	if len(runes) <= width {
		return text
	}
	if width <= 1 {
		return string(runes[:width])
	}
	return string(runes[:width-1]) + "…"
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
