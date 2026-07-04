package terminal

const (
	SyncStart    = "\x1b[?2026h"
	SyncEnd      = "\x1b[?2026l"
	HideCursor   = "\x1b[?25l"
	ShowCursor   = "\x1b[?25h"
	BracketedOn  = "\x1b[?2004h"
	BracketedOff = "\x1b[?2004l"
	MouseOff     = "\x1b[?1006l\x1b[?1000l"
)
