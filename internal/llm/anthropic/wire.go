package anthropic

// AnthropicMessage represents a message for Anthropic API
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// ToolUse represents a tool call for Anthropic API
type ToolUse struct {
	RecipientName string                 `json:"recipient_name"`
	Name          string                 `json:"name"`
	ID            string                 `json:"id"`
	Input         map[string]interface{} `json:"input"`
	Parameters    map[string]interface{} `json:"parameters"`
}

type ToolResult struct {
	Type     string `json:"type"`
	Content  string `json:"content"`
	ToolName string `json:"tool_name"`
}

type CompletionRequest struct {
	Model         string         `json:"model,omitempty"`
	Messages      []Message      `json:"messages"`
	MaxTokens     int            `json:"max_tokens,omitempty"`
	Temperature   float64        `json:"temperature,omitempty"`
	TopP          float64        `json:"top_p,omitempty"`
	TopK          int            `json:"top_k,omitempty"`
	StopSequences []string       `json:"stop_sequences,omitempty"`
	System        string         `json:"system,omitempty"`
	Tools         []Tool         `json:"tools,omitempty"`
	ToolChoice    interface{}    `json:"tool_choice,omitempty"`
	Stream        bool           `json:"stream,omitempty"`
	MetadataKey   string         `json:"metadata,omitempty"`
	Version       string         `json:"anthropicVersion,omitempty"`
	Thinking      *ReasoningSpec `json:"thinking,omitempty"`
}

type ReasoningSpec struct {
	Type         string `json:"type"`
	BudgetTokens int    `json:"budget_tokens,omitempty"`
}

type Tool struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	InputSchema map[string]interface{} `json:"input_schema"`
}

type ContentBlock struct {
	Type    string                 `json:"type"`
	Text    string                 `json:"text,omitempty"`
	ToolUse *ToolUse               `json:"tool_use,omitempty"`
	ID      string                 `json:"id,omitempty"`
	Name    string                 `json:"name,omitempty"`
	Input   map[string]interface{} `json:"input,omitempty"`
}

type CompletionResponse struct {
	ID         string         `json:"id"`
	Type       string         `json:"type"`
	Role       string         `json:"role"`
	Content    []ContentBlock `json:"content"`
	Model      string         `json:"model"`
	StopReason string         `json:"stop_reason"`
	Usage      Usage          `json:"usage"`
}

type Usage struct {
	InputTokens              int `json:"input_tokens"`
	OutputTokens             int `json:"output_tokens"`
	CacheCreationInputTokens int `json:"cache_creation_input_tokens,omitempty"`
	CacheReadInputTokens     int `json:"cache_read_input_tokens,omitempty"`
}
