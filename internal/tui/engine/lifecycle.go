package engine

import "github.com/dm-vev/nu/internal/tui/terminal"

// Start initializes terminal state.
func (t *TUI) Start() error {
	if t.started {
		return nil
	}
	t.started = true
	if err := t.terminal.Write(terminal.BracketedOn + terminal.HideCursor); err != nil {
		return err
	}
	return t.terminal.SetTitle(t.opts.Title)
}

// Stop restores terminal state and moves below the rendered frame.
func (t *TUI) Stop() error {
	if !t.started {
		return nil
	}
	if len(t.previousLines) > 0 {
		if down := len(t.previousLines) - 1 - t.cursorRow; down > 0 {
			if err := t.terminal.MoveBy(down); err != nil {
				return err
			}
		}
		if err := t.terminal.Write("\x1b[1G\r\n"); err != nil {
			return err
		}
	}
	return t.terminal.Write(terminal.SyncEnd + terminal.ShowCursor + terminal.MouseOff + terminal.BracketedOff)
}
