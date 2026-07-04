package ansi

const (
	Reset      = "\x1b[0m"
	Bold       = "\x1b[1m"
	BoldOff    = "\x1b[22m"
	Italic     = "\x1b[3m"
	ItalicOff  = "\x1b[23m"
	Inverse    = "\x1b[7m"
	InverseOff = "\x1b[27m"
	DefaultFG  = "\x1b[39m"
	DefaultBG  = "\x1b[49m"
	Text       = "\x1b[38;5;252m"
	Muted      = "\x1b[38;5;244m"
	Dim        = "\x1b[38;5;241m"
	Green      = "\x1b[38;5;29m"
	Red        = "\x1b[38;5;167m"
	Yellow     = "\x1b[38;5;222m"
	Context    = "\x1b[38;5;222m"

	UserMessageBG = "\x1b[48;5;238m"
	ToolPendingBG = "\x1b[48;5;236m"
	ToolSuccessBG = "\x1b[48;5;22m"
	ToolErrorBG   = "\x1b[48;5;52m"
)
