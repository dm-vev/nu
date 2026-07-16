package app

import (
	"fmt"
	"io"

	"nu/internal/agentui"
)

type printEventWriter struct {
	w   io.Writer
	err error
}

func (w *printEventWriter) emit(ev agentui.Event) {
	if w.err != nil || ev.Type != "turn_end" {
		return
	}
	data, ok := ev.Data.(map[string]string)
	if !ok {
		return
	}
	if text := data["text"]; text != "" {
		// Print mode writes only final assistant text; live deltas stay internal.
		_, w.err = fmt.Fprintln(w.w, text)
	}
}

type jsonEventWriter struct {
	w   io.Writer
	err error
}

func (w *jsonEventWriter) emit(ev agentui.Event) {
	if w.err != nil {
		return
	}
	w.err = writeJSONLine(w.w, ev)
}
