package rpc

import (
	"encoding/json"
	"fmt"
	"io"
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

func (s *Server) write(value any) error {
	s.writeMu.Lock()
	defer s.writeMu.Unlock()
	if s.writeErr != nil {
		return s.writeErr
	}
	if err := WriteLine(s.stdout, value); err != nil {
		s.writeErr = err
		return err
	}
	return nil
}

func (s *Server) setWriteErr(err error) {
	s.writeMu.Lock()
	defer s.writeMu.Unlock()
	if s.writeErr == nil {
		s.writeErr = err
	}
}

func (s *Server) currentWriteErr() error {
	s.writeMu.Lock()
	defer s.writeMu.Unlock()
	return s.writeErr
}
