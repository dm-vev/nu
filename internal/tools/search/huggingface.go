package search

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/dm-vev/nu/contracts"
)

type HuggingFaceModel struct {
	ID          string    `json:"_id"`
	ModelID     string    `json:"modelId"`
	Name        string    `json:"id"`
	Description string    `json:"description,omitempty"`
	URL         string    `json:"url,omitempty"`
	Downloads   int       `json:"downloads"`
	Likes       int       `json:"likes"`
	Private     bool      `json:"private"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"lastModified,omitempty"`
	Tags        []string  `json:"tags"`
	Task        string    `json:"pipeline_tag"`
	LibraryName string    `json:"library_name"`
}

// HuggingFaceTool implements a Hugging Face model search tool.
type HuggingFaceTool struct {
	baseURL    string
	httpClient *http.Client
}

// HuggingFaceOption configures a HuggingFaceTool.
type HuggingFaceOption func(*HuggingFaceTool)

// WithHuggingFaceBaseURL sets the base URL for the Hugging Face API.
func WithHuggingFaceBaseURL(baseURL string) HuggingFaceOption {
	return func(t *HuggingFaceTool) {
		t.baseURL = baseURL
	}
}

// WithHuggingFaceHTTPClient sets the HTTP client for the tool.
func WithHuggingFaceHTTPClient(client *http.Client) HuggingFaceOption {
	return func(t *HuggingFaceTool) {
		t.httpClient = client
	}
}

// NewHuggingFace creates a new Hugging Face model search tool.
func NewHuggingFace(options ...HuggingFaceOption) *HuggingFaceTool {
	tool := &HuggingFaceTool{
		baseURL:    "https://huggingface.co",
		httpClient: &http.Client{Timeout: 10 * time.Second},
	}

	for _, option := range options {
		option(tool)
	}

	return tool
}

// Name returns the name of the tool
func (t *HuggingFaceTool) Name() string {
	return "huggingface_search"
}

// DisplayName implements contracts.ToolWithDisplayName.DisplayName
func (t *HuggingFaceTool) DisplayName() string {
	return "Hugging Face Search"
}

// Description returns a description of what the tool does
func (t *HuggingFaceTool) Description() string {
	return "Search Hugging Face for AI models matching the given query"
}

// Internal implements contracts.InternalTool.Internal
func (t *HuggingFaceTool) Internal() bool {
	return false
}

// Parameters returns the parameters that the tool accepts
func (t *HuggingFaceTool) Parameters() map[string]contracts.ParameterSpec {
	return map[string]contracts.ParameterSpec{
		"query": {
			Type:        "string",
			Description: "The search query",
			Required:    true,
		},
		"limit": {
			Type:        "integer",
			Description: "Number of results to return",
			Required:    false,
			Default:     5,
		},
	}
}

// Run executes the tool with the given input
func (t *HuggingFaceTool) Run(ctx context.Context, input string) (string, error) {
	// Parse input as JSON
	var params map[string]interface{}
	if err := json.Unmarshal([]byte(input), &params); err != nil {
		// If not JSON, treat the input as the query
		params = map[string]interface{}{
			"query": input,
		}
	}

	// Get query parameter
	query, ok := params["query"].(string)
	if !ok || query == "" {
		return "", fmt.Errorf("query parameter is required")
	}

	// Get limit parameter
	limit := 5
	if l, ok := params["limit"].(float64); ok {
		limit = int(l)
	}

	// URL encode the search query
	encodedQuery := url.QueryEscape(query)

	// Build request URL
	url := fmt.Sprintf("%s/api/models?search=%s&sort=downloads&direction=-1&limit=%d", t.baseURL, encodedQuery, limit)

	// Create request
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Accept", "application/json")

	// Execute request
	resp, err := t.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to make request: %w", err)
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			err = fmt.Errorf("failed to close response body: %w", closeErr)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("hugging Face API returned non-200 status code: %d", resp.StatusCode)
	}

	var models []HuggingFaceModel
	if err := json.NewDecoder(resp.Body).Decode(&models); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	// Format results
	var sb strings.Builder
	fmt.Fprintf(&sb, "Found %d models matching '%s':\n\n", len(models), query)
	for i, model := range models {
		fmt.Fprintf(&sb, "%d. %s\n", i+1, model.Name)
		fmt.Fprintf(&sb, "   ID: %s\n", model.ModelID)
		if model.Description != "" {
			fmt.Fprintf(&sb, "   Description: %s\n", model.Description)
		}
		fmt.Fprintf(&sb, "   Downloads: %d\n", model.Downloads)
		fmt.Fprintf(&sb, "   Likes: %d\n", model.Likes)
		fmt.Fprintf(&sb, "   Task: %s\n", model.Task)
		fmt.Fprintf(&sb, "   Library: %s\n", model.LibraryName)
		fmt.Fprintf(&sb, "   Tags: %v\n\n", model.Tags)
	}

	return sb.String(), nil
}

// Execute implements the tool interface
func (t *HuggingFaceTool) Execute(ctx context.Context, args string) (string, error) {
	// Parse args as JSON
	var params struct {
		Query string `json:"query"`
		Limit int    `json:"limit,omitempty"`
	}

	if err := json.Unmarshal([]byte(args), &params); err != nil {
		return "", fmt.Errorf("failed to parse args: %w", err)
	}

	// Execute search
	return t.Run(ctx, fmt.Sprintf(`{"query": "%s", "limit": %d}`, params.Query, params.Limit))
}
