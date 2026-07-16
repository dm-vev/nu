package vllm

// vLLM API request/response structures
type GenerateRequest struct {
	Model         string   `json:"model"`
	Prompt        string   `json:"prompt"`
	Stream        bool     `json:"stream"`
	Temperature   float64  `json:"temperature,omitempty"`
	TopP          float64  `json:"top_p,omitempty"`
	TopK          int      `json:"top_k,omitempty"`
	MaxTokens     int      `json:"max_tokens,omitempty"`
	Stop          []string `json:"stop,omitempty"`
	UseBeamSearch bool     `json:"use_beam_search,omitempty"`
	BestOf        int      `json:"best_of,omitempty"`
	N             int      `json:"n,omitempty"`
}

type GenerateResponse struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	Model   string `json:"model"`
	Choices []struct {
		Index        int         `json:"index"`
		Text         string      `json:"text"`
		LogProbs     interface{} `json:"logprobs,omitempty"`
		FinishReason string      `json:"finish_reason"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
}

type ChatRequest struct {
	Model         string        `json:"model"`
	Messages      []ChatMessage `json:"messages"`
	Stream        bool          `json:"stream"`
	Temperature   float64       `json:"temperature,omitempty"`
	TopP          float64       `json:"top_p,omitempty"`
	TopK          int           `json:"top_k,omitempty"`
	MaxTokens     int           `json:"max_tokens,omitempty"`
	Stop          []string      `json:"stop,omitempty"`
	UseBeamSearch bool          `json:"use_beam_search,omitempty"`
	BestOf        int           `json:"best_of,omitempty"`
	N             int           `json:"n,omitempty"`
}

type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ChatResponse struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	Model   string `json:"model"`
	Choices []struct {
		Index        int         `json:"index"`
		Message      ChatMessage `json:"message"`
		FinishReason string      `json:"finish_reason"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
}

type ModelInfo struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	OwnedBy string `json:"owned_by"`
}

type ModelsResponse struct {
	Object string      `json:"object"`
	Data   []ModelInfo `json:"data"`
}
