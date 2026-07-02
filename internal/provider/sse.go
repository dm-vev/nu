package provider

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"strings"
)

// SSEvent is one parsed server-sent event frame.
type SSEvent struct {
	Event string
	Data  string
}

// ReadSSE parses server-sent events from r.
func ReadSSE(ctx context.Context, r io.Reader, emit func(SSEvent) error) error {
	scanner := bufio.NewScanner(r)
	scanner.Buffer(make([]byte, 0, 64*1024), 4*1024*1024)
	var eventName string
	var data []string
	flush := func() error {
		if len(data) == 0 {
			eventName = ""
			return nil
		}
		ev := SSEvent{Event: eventName, Data: strings.Join(data, "\n")}
		eventName = ""
		data = nil
		if err := emit(ev); err != nil {
			return fmt.Errorf("emit sse event: %w", err)
		}
		return nil
	}
	for scanner.Scan() {
		if err := ctx.Err(); err != nil {
			return fmt.Errorf("read sse: %w", err)
		}
		line := strings.TrimSuffix(scanner.Text(), "\r")
		if line == "" {
			if err := flush(); err != nil {
				return err
			}
			continue
		}
		if strings.HasPrefix(line, ":") {
			continue
		}
		field, value, _ := strings.Cut(line, ":")
		value = strings.TrimPrefix(value, " ")
		switch field {
		case "event":
			eventName = value
		case "data":
			data = append(data, value)
		}
	}
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("scan sse: %w", err)
	}
	return flush()
}
