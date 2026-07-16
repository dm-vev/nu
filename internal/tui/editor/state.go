package editor

// State stores editor buffer and cursor.
type State struct {
	Lines      []string
	CursorLine int
	CursorCol  int
}

func initialState() State {
	return State{Lines: []string{""}}
}
