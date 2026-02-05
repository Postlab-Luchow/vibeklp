package llm

// Message represents a chat message
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// OpenRouterRequest represents the API request body
type OpenRouterRequest struct {
	Model          string                 `json:"model"`
	Messages       []Message              `json:"messages"`
	ResponseFormat map[string]interface{} `json:"response_format,omitempty"`
}

// OpenRouterResponse represents the API response
type OpenRouterResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
	} `json:"usage"`
	Error *struct {
		Message string `json:"message"`
		Code    int    `json:"code"`
	} `json:"error,omitempty"`
}

// ExtractionResult represents the result of an extraction
type ExtractionResult struct {
	Content      string
	PromptTokens int
	CompTokens   int
	FromCache    bool
}
