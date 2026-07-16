package app

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io"
	"time"
)

type jsonSessionHeader struct {
	Type       string    `json:"type"`
	Schema     int       `json:"schema"`
	ID         string    `json:"id"`
	CreatedAt  time.Time `json:"created_at"`
	CWD        string    `json:"cwd"`
	App        string    `json:"app"`
	AppVersion string    `json:"app_version"`
}

func newJSONSessionHeader(opts Options) (jsonSessionHeader, error) {
	id := opts.SessionID
	if id == "" {
		generated, err := newSessionID()
		if err != nil {
			return jsonSessionHeader{}, err
		}
		id = generated
	}
	version := opts.Version.Version
	if version == "" {
		version = "dev"
	}
	return jsonSessionHeader{
		Type:       "session",
		Schema:     1,
		ID:         id,
		CreatedAt:  time.Now().UTC(),
		CWD:        opts.CWD,
		App:        "nu",
		AppVersion: version,
	}, nil
}

func newSessionID() (string, error) {
	var data [16]byte
	if _, err := rand.Read(data[:]); err != nil {
		return "", fmt.Errorf("generate session id: %w", err)
	}
	data[6] = (data[6] & 0x0f) | 0x40
	data[8] = (data[8] & 0x3f) | 0x80
	return fmt.Sprintf(
		"%x-%x-%x-%x-%x",
		data[0:4],
		data[4:6],
		data[6:8],
		data[8:10],
		data[10:16],
	), nil
}

func writeJSONLine(w io.Writer, value any) error {
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("marshal json line: %w", err)
	}
	if _, err := w.Write(append(data, '\n')); err != nil {
		return fmt.Errorf("write json line: %w", err)
	}
	return nil
}
