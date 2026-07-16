package prompts

import (
	"bytes"
	"context"
	"fmt"
	"text/template"
	"time"
)

// PromptTemplateFormat represents the format of a prompt template.
type PromptTemplateFormat string

const (
	// GoTemplate uses Go's text/template package
	PromptTemplateFormatGo PromptTemplateFormat = "go_template"

	// HandlebarsTemplate uses handlebars-style templates
	PromptTemplateFormatHandlebars PromptTemplateFormat = "handlebars"
)

// PromptTemplate represents a prompt template.
type PromptTemplate struct {
	ID          string
	Name        string
	Description string
	Content     string
	Version     string
	Format      PromptTemplateFormat
	CreatedAt   time.Time
	UpdatedAt   time.Time
	Tags        []string
	Metadata    map[string]interface{}

	// Parsed template (cached)
	parsed *template.Template
}

// PromptStore stores and retrieves prompt templates.
type PromptStore interface {
	// Get retrieves a template by ID and version
	Get(ctx context.Context, id string, version string) (*PromptTemplate, error)

	// List returns all templates matching the given filter
	List(ctx context.Context, filter map[string]interface{}) ([]*PromptTemplate, error)

	// Save stores a template
	Save(ctx context.Context, tmpl *PromptTemplate) error

	// Delete removes a template
	Delete(ctx context.Context, id string, version string) error
}

// PromptTemplateOption configures a prompt template.
type PromptTemplateOption func(*PromptTemplate)

// WithPromptTemplateVersion sets the template version.
func WithPromptTemplateVersion(version string) PromptTemplateOption {
	return func(t *PromptTemplate) {
		t.Version = version
	}
}

// WithPromptTemplateDescription sets the template description.
func WithPromptTemplateDescription(description string) PromptTemplateOption {
	return func(t *PromptTemplate) {
		t.Description = description
	}
}

// WithPromptTemplateTags sets the template tags.
func WithPromptTemplateTags(tags ...string) PromptTemplateOption {
	return func(t *PromptTemplate) {
		t.Tags = tags
	}
}

// WithPromptTemplateMetadata sets the template metadata.
func WithPromptTemplateMetadata(metadata map[string]interface{}) PromptTemplateOption {
	return func(t *PromptTemplate) {
		t.Metadata = metadata
	}
}

// WithPromptTemplateFormat sets the template format.
func WithPromptTemplateFormat(format PromptTemplateFormat) PromptTemplateOption {
	return func(t *PromptTemplate) {
		t.Format = format
	}
}

// NewPromptTemplate creates a prompt template.
func NewPromptTemplate(id string, name string, content string, options ...PromptTemplateOption) *PromptTemplate {
	now := time.Now()

	tmpl := &PromptTemplate{
		ID:        id,
		Name:      name,
		Content:   content,
		Version:   "1.0.0",
		Format:    PromptTemplateFormatGo,
		CreatedAt: now,
		UpdatedAt: now,
		Tags:      []string{},
		Metadata:  map[string]interface{}{},
	}

	for _, option := range options {
		option(tmpl)
	}

	return tmpl
}

// Render renders the template with the given data
func (t *PromptTemplate) Render(data map[string]interface{}) (string, error) {
	var err error

	// Parse template if not already parsed
	if t.parsed == nil {
		t.parsed, err = template.New(t.ID).Parse(t.Content)
		if err != nil {
			return "", fmt.Errorf("failed to parse template: %w", err)
		}
	}

	// Render template
	var buf bytes.Buffer
	err = t.parsed.Execute(&buf, data)
	if err != nil {
		return "", fmt.Errorf("failed to render template: %w", err)
	}

	return buf.String(), nil
}
