package rpc

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"strings"
)

// WriteLine writes one strict JSONL record.
func WriteLine(w io.Writer, value any) error {
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("marshal jsonl record: %w", err)
	}
	data = append(data, '\n')
	if _, err := w.Write(data); err != nil {
		return fmt.Errorf("write jsonl record: %w", err)
	}
	return nil
}

// ReadLines reads LF-delimited JSONL records without Scanner's token ceiling.
func ReadLines(r io.Reader, onLine func(string) error) error {
	reader := bufio.NewReader(r)
	for {
		line, err := reader.ReadString('\n')
		if len(line) > 0 {
			// JSONL framing is LF-only; one CR before LF is accepted for clients using CRLF.
			line = strings.TrimSuffix(line, "\n")
			line = strings.TrimSuffix(line, "\r")
			if err := onLine(line); err != nil {
				return err
			}
		}
		if err == nil {
			continue
		}
		if err == io.EOF {
			return nil
		}
		return fmt.Errorf("read jsonl record: %w", err)
	}
}
