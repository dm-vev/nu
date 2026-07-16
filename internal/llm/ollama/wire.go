package ollama

// Ollama API request/response structures
type GenerateRequest struct {
	Model     string   `json:"model"`
	Prompt    string   `json:"prompt"`
	Stream    bool     `json:"stream"`
	Options   *Options `json:"options,omitempty"`
	System    string   `json:"system,omitempty"`
	Template  string   `json:"template,omitempty"`
	Context   []int    `json:"context,omitempty"`
	Format    string   `json:"format,omitempty"`
	Raw       bool     `json:"raw,omitempty"`
	KeepAlive string   `json:"keep_alive,omitempty"`
	Images    []string `json:"images,omitempty"`
}

type Options struct {
	Temperature   float64  `json:"temperature,omitempty"`
	TopP          float64  `json:"top_p,omitempty"`
	TopK          int      `json:"top_k,omitempty"`
	NumPredict    int      `json:"num_predict,omitempty"`
	Stop          []string `json:"stop,omitempty"`
	RepeatPenalty float64  `json:"repeat_penalty,omitempty"`
	Seed          int      `json:"seed,omitempty"`
}

type GenerateResponse struct {
	Model              string `json:"model"`
	CreatedAt          string `json:"created_at"`
	Response           string `json:"response"`
	Done               bool   `json:"done"`
	Context            []int  `json:"context,omitempty"`
	TotalDuration      int64  `json:"total_duration,omitempty"`
	LoadDuration       int64  `json:"load_duration,omitempty"`
	PromptEvalCount    int    `json:"prompt_eval_count,omitempty"`
	PromptEvalDuration int64  `json:"prompt_eval_duration,omitempty"`
	EvalCount          int    `json:"eval_count,omitempty"`
	EvalDuration       int64  `json:"eval_duration,omitempty"`
}

type ChatRequest struct {
	Model     string        `json:"model"`
	Messages  []ChatMessage `json:"messages"`
	Stream    bool          `json:"stream"`
	Tools     []Tool        `json:"tools,omitempty"`
	Options   *Options      `json:"options,omitempty"`
	Format    string        `json:"format,omitempty"`
	KeepAlive string        `json:"keep_alive,omitempty"`
}

type ChatMessage struct {
	Role      string     `json:"role"`
	Content   string     `json:"content"`
	ToolCalls []ToolCall `json:"tool_calls,omitempty"`
}

// OllamaTool is a function declaration sent to /api/chat for native tool use.
type Tool struct {
	Type     string   `json:"type"`
	Function Function `json:"function"`
}

type Function struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Parameters  map[string]interface{} `json:"parameters"`
}

// OllamaToolCall is a function invocation requested by the model in /api/chat.
type ToolCall struct {
	Function ToolCallFunction `json:"function"`
}

type ToolCallFunction struct {
	Name      string                 `json:"name"`
	Arguments map[string]interface{} `json:"arguments"`
}

type ChatResponse struct {
	Model              string      `json:"model"`
	CreatedAt          string      `json:"created_at"`
	Message            ChatMessage `json:"message"`
	Done               bool        `json:"done"`
	TotalDuration      int64       `json:"total_duration,omitempty"`
	LoadDuration       int64       `json:"load_duration,omitempty"`
	PromptEvalCount    int         `json:"prompt_eval_count,omitempty"`
	PromptEvalDuration int64       `json:"prompt_eval_duration,omitempty"`
	EvalCount          int         `json:"eval_count,omitempty"`
	EvalDuration       int64       `json:"eval_duration,omitempty"`
}
