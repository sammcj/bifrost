package providers

import (
	"bifrost/interfaces"
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// OpenAIProvider implements the Provider interface for OpenAI
type OpenAIProvider struct {
	Config map[string]string
	client *http.Client
}

// NewOpenAIProvider creates a new OpenAI provider instance
func NewOpenAIProvider() *OpenAIProvider {
	return &OpenAIProvider{
		client: &http.Client{Timeout: time.Second * 30},
	}
}

func (provider *OpenAIProvider) GetProviderKey() interfaces.SupportedModelProvider {
	return interfaces.OpenAI
}

func (provider *OpenAIProvider) GetConfig() interface{} {
	return provider.Config
}

func (provider *OpenAIProvider) IsEnabled() bool {
	return true
}

// TextCompletion performs text completion
func (provider *OpenAIProvider) TextCompletion(model, key, text string, params *interfaces.ModelParameters) (*interfaces.CompletionResult, error) {
	return nil, fmt.Errorf("text completion is not supported by OpenAI")
}

// sanitizeParameters cleans up the parameters for OpenAI
func (provider *OpenAIProvider) sanitizeParameters(params *interfaces.ModelParameters) *interfaces.ModelParameters {
	sanitized := params
	if sanitized == nil {
		return nil
	}

	if params.ExtraParams != nil {
		// For logprobs, if it's disabled, we remove top_logprobs
		if _, exists := params.ExtraParams["logprobs"]; !exists {
			delete(sanitized.ExtraParams, "top_logprobs")
		}
	}

	return sanitized
}

// ChatCompletion implements chat completion using OpenAI's API
func (provider *OpenAIProvider) ChatCompletion(model, key string, messages []interface{}, params *interfaces.ModelParameters) (*interfaces.CompletionResult, error) {
	startTime := time.Now()

	// Format messages for OpenAI API
	var openAIMessages []map[string]interface{}
	for _, msg := range messages {
		if m, ok := msg.(map[string]interface{}); ok {
			role, _ := m["role"].(string)
			content := m["content"].(interface{})

			openAIMessages = append(openAIMessages, map[string]interface{}{
				"role":    role,
				"content": content,
			})
		}
	}

	// Prepare request body with default config
	requestBody := map[string]interface{}{
		"model":    model,
		"messages": openAIMessages,
	}

	// Sanitize parameters
	params = provider.sanitizeParameters(params)

	if params != nil {
		if params.ExtraParams != nil {
			requestBody = MergeConfig(requestBody, params.ExtraParams)
		}

		if params.TestRunEntryID != nil {
			requestBody["test_run_entry_id"] = *params.TestRunEntryID
		}

		if params.ToolChoice != nil {
			requestBody["tool_choice"] = *params.ToolChoice
		}

		if params.Tools != nil {
			requestBody["tools"] = params.Tools
		}

		if params.FunctionCall != nil {
			requestBody["function_call"] = *params.FunctionCall
		}

		if params.Functions != nil {
			requestBody["functions"] = params.Functions
		}

		if params.PromptTools != nil {
			requestBody["prompt_tools"] = params.PromptTools
		}
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("error marshaling request: %v", err)
	}

	// Create request
	req, err := http.NewRequest("POST", "https://api.openai.com/v1/chat/completions", bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("error creating request: %v", err)
	}

	// Add headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+key)

	// Make request
	resp, err := provider.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error making request: %v", err)
	}
	defer resp.Body.Close()

	latency := time.Since(startTime).Seconds()

	// Handle error response
	if resp.StatusCode != http.StatusOK {
		var errorResp struct {
			Error struct {
				Message string `json:"message"`
				Type    string `json:"type"`
				Param   any    `json:"param"`
				Code    string `json:"code"`
			} `json:"error"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&errorResp); err != nil {
			return nil, fmt.Errorf("error decoding error response: %v", err)
		}
		return nil, fmt.Errorf("OpenAI error: %s", errorResp.Error.Message)
	}

	// Decode response
	var rawResult struct {
		ID      string                              `json:"id"`
		Choices []interfaces.CompletionResultChoice `json:"choices"`
		Usage   interfaces.LLMUsage                 `json:"usage"`
		Model   string                              `json:"model"`
		Created interface{}                         `json:"created"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&rawResult); err != nil {
		return nil, fmt.Errorf("error decoding response: %v", err)
	}

	// Convert the raw result to CompletionResult
	result := &interfaces.CompletionResult{
		ID:      rawResult.ID,
		Choices: rawResult.Choices,
		Usage:   rawResult.Usage,
		Model:   rawResult.Model,
	}

	// Handle the created field conversion
	if rawResult.Created != nil {
		switch v := rawResult.Created.(type) {
		case float64:
			// Convert Unix timestamp to string
			result.Created = fmt.Sprintf("%d", int64(v))
		case string:
			result.Created = v
		}
	}

	// Add provider-specific information
	result.Provider = interfaces.OpenAI
	result.Usage.Latency = latency

	return result, nil
}
