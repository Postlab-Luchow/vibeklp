package llm

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// OpenRouterClient handles communication with OpenRouter API
type OpenRouterClient struct {
	APIKey     string
	Model      string
	HTTPClient *http.Client
	BaseURL    string
}

// NewClient creates a new OpenRouter client
func NewClient(apiKey, model string) *OpenRouterClient {
	return &OpenRouterClient{
		APIKey:  apiKey,
		Model:   model,
		BaseURL: "https://openrouter.ai/api/v1",
		HTTPClient: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
}

// ExtractWithSchema sends HTML to LLM with JSON Schema for structured output
func (c *OpenRouterClient) ExtractWithSchema(
	html string,
	schema map[string]interface{},
	systemPrompt string,
) (*ExtractionResult, error) {
	reqBody := OpenRouterRequest{
		Model: c.Model,
		Messages: []Message{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: html},
		},
		ResponseFormat: map[string]interface{}{
			"type":        "json_schema",
			"json_schema": schema,
		},
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", c.BaseURL+"/chat/completions", bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.APIKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("HTTP-Referer", "https://github.com/musche/klp")
	req.Header.Set("X-Title", "KLP Crawler")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
	}

	var apiResp OpenRouterResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if apiResp.Error != nil {
		return nil, fmt.Errorf("API error: %s (code: %d)", apiResp.Error.Message, apiResp.Error.Code)
	}

	if len(apiResp.Choices) == 0 {
		return nil, fmt.Errorf("no choices in response")
	}

	return &ExtractionResult{
		Content:      apiResp.Choices[0].Message.Content,
		PromptTokens: apiResp.Usage.PromptTokens,
		CompTokens:   apiResp.Usage.CompletionTokens,
		FromCache:    false,
	}, nil
}
