package prompts

import (
	"bytes"
	"fmt"
	"strings"
	"time"
)

// parseTemplateFile parses a template file
func parseTemplateFile(data string, id string, version string) (*PromptTemplate, error) {
	// Split into sections
	sections := strings.Split(data, "---\n")
	if len(sections) < 2 {
		return nil, fmt.Errorf("invalid template file format")
	}

	// Parse metadata
	metadata := sections[0]
	content := strings.Join(sections[1:], "---\n")

	// Parse metadata lines
	lines := strings.Split(metadata, "\n")
	tmpl := &PromptTemplate{
		ID:        id,
		Version:   version,
		Content:   content,
		Format:    PromptTemplateFormatGo,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Tags:      []string{},
		Metadata:  map[string]interface{}{},
	}

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		switch key {
		case "name":
			tmpl.Name = value
		case "description":
			tmpl.Description = value
		case "format":
			tmpl.Format = PromptTemplateFormat(value)
		case "tags":
			tmpl.Tags = strings.Split(value, ",")
			for i, tag := range tmpl.Tags {
				tmpl.Tags[i] = strings.TrimSpace(tag)
			}
		default:
			tmpl.Metadata[key] = value
		}
	}

	return tmpl, nil
}

// serializeTemplate serializes a template to a string
func serializeTemplate(tmpl *PromptTemplate) string {
	var buf bytes.Buffer

	// Write metadata
	fmt.Fprintf(&buf, "name: %s\n", tmpl.Name)
	fmt.Fprintf(&buf, "description: %s\n", tmpl.Description)
	fmt.Fprintf(&buf, "format: %s\n", tmpl.Format)

	if len(tmpl.Tags) > 0 {
		fmt.Fprintf(&buf, "tags: %s\n", strings.Join(tmpl.Tags, ", "))
	}

	for key, value := range tmpl.Metadata {
		fmt.Fprintf(&buf, "%s: %v\n", key, value)
	}

	// Write content
	buf.WriteString("---\n")
	buf.WriteString(tmpl.Content)

	return buf.String()
}
