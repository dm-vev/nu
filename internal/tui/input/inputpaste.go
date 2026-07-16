package input

import "fmt"

func (d *Decoder) readPaste() (string, error) {
	buffer := make([]byte, 0)
	for {
		next, err := d.reader.ReadByte()
		if err != nil {
			return "", fmt.Errorf("read bracketed paste: %w", err)
		}
		buffer = append(buffer, next)
		if hasPasteEnd(buffer) {
			return string(buffer[:len(buffer)-len(bracketedPasteEnd)]), nil
		}
	}
}

func hasPasteEnd(buffer []byte) bool {
	if len(buffer) < len(bracketedPasteEnd) {
		return false
	}
	return string(buffer[len(buffer)-len(bracketedPasteEnd):]) == bracketedPasteEnd
}
