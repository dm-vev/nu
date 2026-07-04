package input

// Event is one decoded terminal input sequence.
type Event struct {
	Data string
}

const (
	bracketedPasteStart = "\x1b[200~"
	bracketedPasteEnd   = "\x1b[201~"
)
