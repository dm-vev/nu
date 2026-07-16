package session

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"os"
)

func readSession(path string) (Header, []Entry, error) {
	file, err := os.Open(path)
	if err != nil {
		return Header{}, nil, fmt.Errorf("open session %s: %w", path, err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	scanner.Buffer(nil, 1024*1024)
	if !scanner.Scan() {
		if err := scanner.Err(); err != nil {
			return Header{}, nil, fmt.Errorf("read session header %s: %w", path, err)
		}
		return Header{}, nil, fmt.Errorf("read session header %s: empty file", path)
	}

	var header Header
	if err := json.Unmarshal(scanner.Bytes(), &header); err != nil {
		return Header{}, nil, fmt.Errorf("decode session header %s: %w", path, err)
	}
	if header.Type != "session" || header.Schema != schemaVersion || header.ID == "" {
		return Header{}, nil, fmt.Errorf("decode session header %s: invalid header", path)
	}

	var entries []Entry
	for scanner.Scan() {
		entry, err := UnmarshalEntry(scanner.Bytes())
		if err != nil {
			return Header{}, nil, fmt.Errorf("decode session entry %s: %w", path, err)
		}
		entries = append(entries, entry)
	}
	if err := scanner.Err(); err != nil {
		return Header{}, nil, fmt.Errorf("read session %s: %w", path, err)
	}
	return header, entries, nil
}

func readHeader(path string) (Header, error) {
	file, err := os.Open(path)
	if err != nil {
		return Header{}, fmt.Errorf("open session %s: %w", path, err)
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	scanner.Buffer(nil, 1024*1024)
	if !scanner.Scan() {
		if err := scanner.Err(); err != nil {
			return Header{}, fmt.Errorf("read session header %s: %w", path, err)
		}
		return Header{}, fmt.Errorf("read session header %s: empty file", path)
	}
	header, _, err := parseSession(scanner.Bytes())
	if err != nil {
		return Header{}, fmt.Errorf("decode session header %s: %w", path, err)
	}
	return header, nil
}

func parseSession(data []byte) (Header, []Entry, error) {
	scanner := bufio.NewScanner(bytes.NewReader(data))
	scanner.Buffer(nil, 1024*1024)
	if !scanner.Scan() {
		if err := scanner.Err(); err != nil {
			return Header{}, nil, fmt.Errorf("read session header: %w", err)
		}
		return Header{}, nil, fmt.Errorf("read session header: empty data")
	}
	var header Header
	if err := json.Unmarshal(scanner.Bytes(), &header); err != nil {
		return Header{}, nil, fmt.Errorf("decode session header: %w", err)
	}
	if header.Type != "session" || header.Schema != schemaVersion || header.ID == "" {
		return Header{}, nil, fmt.Errorf("decode session header: invalid header")
	}
	var entries []Entry
	for scanner.Scan() {
		entry, err := UnmarshalEntry(scanner.Bytes())
		if err != nil {
			return Header{}, nil, fmt.Errorf("decode session entry: %w", err)
		}
		entries = append(entries, entry)
	}
	if err := scanner.Err(); err != nil {
		return Header{}, nil, fmt.Errorf("read session: %w", err)
	}
	// Import validates tree links before any bytes are written to the store.
	if _, err := BuildTree(entries); err != nil {
		return Header{}, nil, err
	}
	return header, entries, nil
}

func writeJSONLine(file *os.File, value any) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	_, err = file.Write(append(data, '\n'))
	return err
}
