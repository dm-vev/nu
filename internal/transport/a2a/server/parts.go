package server

import (
	"fmt"

	"github.com/a2aproject/a2a-go/a2a"
)

func formatFilePart(part a2a.FilePart) string {
	switch file := part.File.(type) {
	case a2a.FileURI:
		name := file.Name
		if name == "" {
			name = file.URI
		}
		return fmt.Sprintf("[file: %s]", name)
	case a2a.FileBytes:
		name := file.Name
		if name == "" {
			name = "unnamed"
		}
		return fmt.Sprintf("[file: %s (base64: %d chars)]", name, len(file.Bytes))
	default:
		return "[file: unknown]"
	}
}
