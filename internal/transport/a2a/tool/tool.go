package tool

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/a2aproject/a2a-go/a2a"

	"github.com/dm-vev/nu/contracts"
	"github.com/dm-vev/nu/internal/transport/a2a/client"
	"github.com/dm-vev/nu/telemetry"
)

// Tool wraps an A2A client as a contracts.Tool so that a remote
// A2A agent can be used as a tool by agent-sdk-go agents.
type Tool struct {
	client   *client.Client
	logger   telemetry.Logger
	nameOver string
}

// Option configures a Tool.
type Option func(*Tool)

// WithToolName overrides the auto-generated tool name.
// Use this to prevent name collisions when registering multiple remote agents.
func WithToolName(name string) Option {
	return func(t *Tool) {
		t.nameOver = name
	}
}

// New creates a tool from an A2A client.
func New(client *client.Client, opts ...Option) *Tool {
	t := &Tool{
		client: client,
		logger: client.Logger(),
	}
	for _, opt := range opts {
		opt(t)
	}
	return t
}

var _ contracts.Tool = (*Tool)(nil)

func (t *Tool) Name() string {
	if t.nameOver != "" {
		return t.nameOver
	}
	card := t.client.Card()
	return sanitizeToolName(card.Name)
}

func (t *Tool) Description() string {
	return t.client.Card().Description
}

func (t *Tool) Parameters() map[string]contracts.ParameterSpec {
	return map[string]contracts.ParameterSpec{
		"query": {
			Type:        "string",
			Description: fmt.Sprintf("The message to send to the remote %s A2A agent", t.client.Card().Name),
			Required:    true,
		},
	}
}

func (t *Tool) Run(ctx context.Context, input string) (string, error) {
	result, err := t.client.SendMessage(ctx, input)
	if err != nil {
		return "", fmt.Errorf("a2a tool %s: %w", t.Name(), err)
	}
	return extractResultText(ctx, t.logger, result), nil
}

func (t *Tool) Execute(ctx context.Context, args string) (string, error) {
	var params struct {
		Query string `json:"query"`
	}
	if err := json.Unmarshal([]byte(args), &params); err != nil {
		return "", fmt.Errorf("a2a tool: failed to parse arguments: %w", err)
	}
	if params.Query == "" {
		return "", fmt.Errorf("a2a tool: query parameter is required")
	}
	return t.Run(ctx, params.Query)
}

// ExtractResultText pulls text content from an A2A SendMessageResult.
// Non-text parts are converted with warnings logged via the provided logger.
func ExtractResultText(result a2a.SendMessageResult) string {
	return extractResultText(context.Background(), telemetry.NewLogger(), result)
}

// extractResultText is the internal version that accepts a logger for non-text part warnings.
func extractResultText(ctx context.Context, logger telemetry.Logger, result a2a.SendMessageResult) string {
	switch r := result.(type) {
	case *a2a.Task:
		if len(r.Artifacts) > 0 {
			var parts []string
			for _, artifact := range r.Artifacts {
				for _, p := range artifact.Parts {
					if text := partToText(ctx, logger, p); text != "" {
						parts = append(parts, text)
					}
				}
			}
			return strings.Join(parts, "\n")
		}
		if r.Status.Message != nil {
			return messagePartsToText(ctx, logger, r.Status.Message)
		}
		return ""
	case *a2a.Message:
		return messagePartsToText(ctx, logger, r)
	default:
		return fmt.Sprintf("%v", result)
	}
}

// messagePartsToText extracts text from all parts of a message.
func messagePartsToText(ctx context.Context, logger telemetry.Logger, msg *a2a.Message) string {
	if msg == nil {
		return ""
	}
	var parts []string
	for _, p := range msg.Parts {
		parts = append(parts, partToText(ctx, logger, p))
	}
	return strings.Join(parts, "\n")
}

// partToText converts any A2A Part to a text representation.
func partToText(ctx context.Context, logger telemetry.Logger, p a2a.Part) string {
	switch tp := p.(type) {
	case a2a.TextPart:
		return tp.Text
	case a2a.DataPart:
		logger.Warn(ctx, "A2A tool: non-text DataPart in result, converting to JSON", nil)
		data, err := json.Marshal(tp.Data)
		if err != nil {
			return fmt.Sprintf("[data: marshal error: %v]", err)
		}
		return string(data)
	case a2a.FilePart:
		logger.Warn(ctx, "A2A tool: non-text FilePart in result, using placeholder", nil)
		return formatFilePart(tp)
	default:
		logger.Warn(ctx, "A2A tool: unknown part type in result", map[string]any{
			"type": fmt.Sprintf("%T", p),
		})
		return fmt.Sprintf("%v", p)
	}
}

// formatFilePart produces a text representation of a FilePart.
func formatFilePart(fp a2a.FilePart) string {
	switch fc := fp.File.(type) {
	case a2a.FileURI:
		name := fc.Name
		if name == "" {
			name = fc.URI
		}
		return fmt.Sprintf("[file: %s]", name)
	case a2a.FileBytes:
		name := fc.Name
		if name == "" {
			name = "unnamed"
		}
		return fmt.Sprintf("[file: %s (base64: %d chars)]", name, len(fc.Bytes))
	default:
		return "[file: unknown]"
	}
}

func sanitizeToolName(name string) string {
	result := strings.Map(func(r rune) rune {
		if r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' || r >= '0' && r <= '9' || r == '_' {
			return r
		}
		return '_'
	}, name)
	result = strings.ToLower(result)
	// Collapse consecutive underscores and trim edges.
	for strings.Contains(result, "__") {
		result = strings.ReplaceAll(result, "__", "_")
	}
	result = strings.Trim(result, "_")
	if result == "" {
		return "remote_agent"
	}
	if result[0] >= '0' && result[0] <= '9' {
		result = "agent_" + result
	}
	return result
}
