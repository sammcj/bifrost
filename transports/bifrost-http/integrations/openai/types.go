package openai

import (
	"strings"

	bifrost "github.com/maximhq/bifrost/core"
	"github.com/maximhq/bifrost/core/schemas"
)

var fnTypePtr = bifrost.Ptr(string(schemas.ToolChoiceTypeFunction))

// OpenAIContentPart represents a part of the content (text or image) in OpenAI format
type OpenAIContentPart struct {
	Type     string          `json:"type"`
	Text     *string         `json:"text,omitempty"`
	ImageURL *OpenAIImageURL `json:"image_url,omitempty"`
}

// OpenAIImageURL represents an image URL with optional detail level in OpenAI format
type OpenAIImageURL struct {
	URL    string  `json:"url"`
	Detail *string `json:"detail,omitempty"`
}

// OpenAIMessage represents a message in the OpenAI chat format
type OpenAIMessage struct {
	Role         string                `json:"role"`
	Content      interface{}           `json:"content,omitempty"` // Can be string or []OpenAIContentPart
	Name         *string               `json:"name,omitempty"`
	ToolCalls    *[]schemas.ToolCall   `json:"tool_calls,omitempty"` // Reuse schema type
	ToolCallID   *string               `json:"tool_call_id,omitempty"`
	FunctionCall *schemas.FunctionCall `json:"function_call,omitempty"` // Reuse schema type
}

// OpenAIChatRequest represents an OpenAI chat completion request
type OpenAIChatRequest struct {
	Model            string              `json:"model"`
	Messages         []OpenAIMessage     `json:"messages"`
	MaxTokens        *int                `json:"max_tokens,omitempty"`
	Temperature      *float64            `json:"temperature,omitempty"`
	TopP             *float64            `json:"top_p,omitempty"`
	N                *int                `json:"n,omitempty"`
	Stop             interface{}         `json:"stop,omitempty"`
	PresencePenalty  *float64            `json:"presence_penalty,omitempty"`
	FrequencyPenalty *float64            `json:"frequency_penalty,omitempty"`
	LogitBias        map[string]float64  `json:"logit_bias,omitempty"`
	User             *string             `json:"user,omitempty"`
	Functions        *[]schemas.Function `json:"functions,omitempty"` // Reuse schema type
	FunctionCall     interface{}         `json:"function_call,omitempty"`
	Tools            *[]schemas.Tool     `json:"tools,omitempty"` // Reuse schema type
	ToolChoice       interface{}         `json:"tool_choice,omitempty"`
	Stream           *bool               `json:"stream,omitempty"`
	LogProbs         *bool               `json:"logprobs,omitempty"`
	TopLogProbs      *int                `json:"top_logprobs,omitempty"`
	ResponseFormat   interface{}         `json:"response_format,omitempty"`
	Seed             *int                `json:"seed,omitempty"`
}

// OpenAIChatResponse represents an OpenAI chat completion response
type OpenAIChatResponse struct {
	ID                string            `json:"id"`
	Object            string            `json:"object"`
	Created           int               `json:"created"`
	Model             string            `json:"model"`
	Choices           []OpenAIChoice    `json:"choices"`
	Usage             *schemas.LLMUsage `json:"usage,omitempty"` // Reuse schema type
	SystemFingerprint *string           `json:"system_fingerprint,omitempty"`
}

// OpenAIChoice represents a choice in the OpenAI response
type OpenAIChoice struct {
	Index        int           `json:"index"`
	Message      OpenAIMessage `json:"message"`
	FinishReason *string       `json:"finish_reason,omitempty"`
	LogProbs     interface{}   `json:"logprobs,omitempty"`
}

// convertOpenAIContent handles both string and structured content formats
// It returns the text content and any image content found
func convertOpenAIContent(content interface{}) (*string, *schemas.ImageContent) {
	if content == nil {
		return nil, nil
	}

	switch c := content.(type) {
	case string:
		return &c, nil
	case []interface{}:
		// Handle array of content parts (for vision API)
		var textParts []string
		var imageContent *schemas.ImageContent

		for _, part := range c {
			if partMap, ok := part.(map[string]interface{}); ok {
				if partType, exists := partMap["type"].(string); exists {
					switch partType {
					case "text":
						if text, textExists := partMap["text"].(string); textExists {
							textParts = append(textParts, text)
						}
					case "image_url":
						if imageURL, ok := partMap["image_url"].(map[string]interface{}); ok {
							if url, urlExists := imageURL["url"].(string); urlExists {
								// Initialize imageContent if we have a URL
								imageContent = &schemas.ImageContent{
									URL: url,
								}

								// Get detail level if specified
								if detail, detailExists := imageURL["detail"].(string); detailExists {
									imageContent.Detail = &detail
								}
							}
						}
					}
				}
			}
		}

		var textContent *string
		if len(textParts) > 0 {
			combined := strings.Join(textParts, " ")
			textContent = &combined
		}

		return textContent, imageContent
	}

	return nil, nil
}

// ConvertToBifrostRequest converts an OpenAI chat request to Bifrost format
func (r *OpenAIChatRequest) ConvertToBifrostRequest() *schemas.BifrostRequest {
	bifrostReq := &schemas.BifrostRequest{
		Provider: schemas.OpenAI,
		Model:    r.Model,
		Input: schemas.RequestInput{
			ChatCompletionInput: &[]schemas.BifrostMessage{},
		},
	}

	// Convert messages
	r.convertMessages(bifrostReq)

	// Convert parameters using dynamic reflection-based mapping
	bifrostReq.Params = r.convertParameters()

	// Convert tools and tool choices
	r.convertToolsAndChoices(bifrostReq)

	return bifrostReq
}

// convertParameters converts OpenAI request parameters to Bifrost ModelParameters
// using direct field access for better performance and type safety.
func (r *OpenAIChatRequest) convertParameters() *schemas.ModelParameters {
	params := &schemas.ModelParameters{
		ExtraParams: make(map[string]interface{}),
	}

	// Direct field mapping with type safety - much faster than reflection
	if r.MaxTokens != nil {
		params.ExtraParams["max_tokens"] = *r.MaxTokens
	}
	if r.Temperature != nil {
		params.ExtraParams["temperature"] = *r.Temperature
	}
	if r.TopP != nil {
		params.ExtraParams["top_p"] = *r.TopP
	}
	if r.PresencePenalty != nil {
		params.ExtraParams["presence_penalty"] = *r.PresencePenalty
	}
	if r.FrequencyPenalty != nil {
		params.ExtraParams["frequency_penalty"] = *r.FrequencyPenalty
	}
	if r.N != nil {
		params.ExtraParams["n"] = *r.N
	}
	if r.LogProbs != nil {
		params.ExtraParams["logprobs"] = *r.LogProbs
	}
	if r.TopLogProbs != nil {
		params.ExtraParams["top_logprobs"] = *r.TopLogProbs
	}
	if r.Stop != nil {
		params.ExtraParams["stop"] = r.Stop
	}
	if r.LogitBias != nil {
		params.ExtraParams["logit_bias"] = r.LogitBias
	}
	if r.User != nil {
		params.ExtraParams["user"] = *r.User
	}
	if r.Stream != nil {
		params.ExtraParams["stream"] = *r.Stream
	}
	if r.ResponseFormat != nil {
		params.ExtraParams["response_format"] = r.ResponseFormat
	}
	if r.Seed != nil {
		params.ExtraParams["seed"] = *r.Seed
	}

	// Return nil if no parameters were set
	if len(params.ExtraParams) == 0 {
		return nil
	}

	return params
}

// convertMessages handles message conversion from OpenAI to Bifrost format
func (r *OpenAIChatRequest) convertMessages(bifrostReq *schemas.BifrostRequest) {
	for _, msg := range r.Messages {
		bifrostMsg := r.convertSingleMessage(msg)
		*bifrostReq.Input.ChatCompletionInput = append(*bifrostReq.Input.ChatCompletionInput, bifrostMsg)
	}
}

// convertSingleMessage converts a single OpenAI message to Bifrost format
func (r *OpenAIChatRequest) convertSingleMessage(msg OpenAIMessage) schemas.BifrostMessage {
	// Handle different content formats (string vs array of content parts)
	textContent, imageContent := convertOpenAIContent(msg.Content)

	// Create BifrostMessage with proper embedded struct setup
	bifrostMsg := schemas.BifrostMessage{
		Role:    schemas.ModelChatMessageRole(msg.Role),
		Content: textContent,
	}

	// Set embedded struct based on message role and image content
	if imageContent != nil {
		r.setImageContent(&bifrostMsg, msg.Role, imageContent)
	}

	// Handle tool calls and function calls for assistant messages
	r.setToolCalls(&bifrostMsg, msg)

	// Handle tool messages
	r.setToolMessage(&bifrostMsg, msg)

	return bifrostMsg
}

// setImageContent sets image content based on message role
func (r *OpenAIChatRequest) setImageContent(bifrostMsg *schemas.BifrostMessage, role string, imageContent *schemas.ImageContent) {
	switch role {
	case "user":
		bifrostMsg.Role = schemas.ModelChatMessageRoleUser
		bifrostMsg.UserMessage = &schemas.UserMessage{
			ImageContent: imageContent,
		}
	case "tool":
		bifrostMsg.Role = schemas.ModelChatMessageRoleTool
		bifrostMsg.ToolMessage = &schemas.ToolMessage{
			ImageContent: imageContent,
		}
	}
}

// setToolCalls handles tool calls and function calls for assistant messages
func (r *OpenAIChatRequest) setToolCalls(bifrostMsg *schemas.BifrostMessage, msg OpenAIMessage) {
	var toolCalls []schemas.ToolCall

	// Prioritize modern tool_calls format, only use legacy function_call if tool_calls is not present
	if msg.ToolCalls != nil {
		// Tool calls are already in the right format (schemas.ToolCall)
		toolCalls = *msg.ToolCalls
	} else if msg.FunctionCall != nil {
		// Add legacy function call only if no modern tool calls exist
		tc := schemas.ToolCall{
			Type:     fnTypePtr,
			Function: *msg.FunctionCall, // Already schemas.FunctionCall type
		}
		toolCalls = append(toolCalls, tc)
	}

	// Assign AssistantMessage only if we have tool calls
	if len(toolCalls) > 0 {
		bifrostMsg.AssistantMessage = &schemas.AssistantMessage{
			ToolCalls: &toolCalls,
		}
	}
}

// setToolMessage handles tool message-specific fields
func (r *OpenAIChatRequest) setToolMessage(bifrostMsg *schemas.BifrostMessage, msg OpenAIMessage) {
	if msg.ToolCallID != nil {
		if bifrostMsg.ToolMessage == nil {
			bifrostMsg.ToolMessage = &schemas.ToolMessage{}
		}
		bifrostMsg.ToolMessage.ToolCallID = msg.ToolCallID
	}
}

// convertToolsAndChoices handles tools and tool choice conversion
func (r *OpenAIChatRequest) convertToolsAndChoices(bifrostReq *schemas.BifrostRequest) {
	// Convert tools
	allTools := r.convertTools()
	if len(allTools) > 0 {
		r.ensureParams(bifrostReq)
		bifrostReq.Params.Tools = &allTools
	}

	// Convert tool choice
	toolChoice := r.convertToolChoice()
	if toolChoice != nil {
		r.ensureParams(bifrostReq)
		bifrostReq.Params.ToolChoice = toolChoice
	}
}

// convertTools combines modern Tools and legacy Functions into a unified tool list
func (r *OpenAIChatRequest) convertTools() []schemas.Tool {
	var allTools []schemas.Tool

	// Handle modern Tools field (already schemas.Tool type)
	if r.Tools != nil {
		allTools = append(allTools, *r.Tools...)
	}

	// Handle legacy Functions field
	if r.Functions != nil {
		for _, function := range *r.Functions {
			t := schemas.Tool{
				Type:     string(schemas.ToolChoiceTypeFunction),
				Function: function, // Already schemas.Function type
			}
			allTools = append(allTools, t)
		}
	}

	return allTools
}

// convertToolChoice handles both modern tool_choice and legacy function_call
func (r *OpenAIChatRequest) convertToolChoice() *schemas.ToolChoice {
	if r.ToolChoice == nil && r.FunctionCall == nil {
		return nil
	}

	// Handle ToolChoice (modern format) first
	if r.ToolChoice != nil {
		return r.parseToolChoice(r.ToolChoice)
	}

	// Handle legacy FunctionCall
	if r.FunctionCall != nil {
		return r.parseFunctionCall(r.FunctionCall)
	}

	return nil
}

// parseToolChoice parses modern tool_choice format
func (r *OpenAIChatRequest) parseToolChoice(toolChoice interface{}) *schemas.ToolChoice {
	tc := &schemas.ToolChoice{}

	switch v := toolChoice.(type) {
	case string:
		tc.Type = r.parseToolChoiceString(v)
	case map[string]interface{}:
		r.parseToolChoiceObject(tc, v)
	}

	return tc
}

// parseToolChoiceString handles string tool choice values
func (r *OpenAIChatRequest) parseToolChoiceString(value string) schemas.ToolChoiceType {
	switch value {
	case "none":
		return schemas.ToolChoiceTypeNone
	case "auto":
		return schemas.ToolChoiceTypeAuto
	case "required":
		return schemas.ToolChoiceTypeRequired
	default:
		return schemas.ToolChoiceTypeAuto // fallback
	}
}

// parseToolChoiceObject handles object tool choice values
func (r *OpenAIChatRequest) parseToolChoiceObject(tc *schemas.ToolChoice, obj map[string]interface{}) {
	typeVal, ok := obj["type"].(string)
	if !ok {
		tc.Type = schemas.ToolChoiceTypeAuto
		return
	}

	switch typeVal {
	case "function":
		tc.Type = schemas.ToolChoiceTypeFunction
		if functionVal, ok := obj["function"].(map[string]interface{}); ok {
			if name, ok := functionVal["name"].(string); ok {
				tc.Function = schemas.ToolChoiceFunction{Name: name}
			}
		}
	case "none":
		tc.Type = schemas.ToolChoiceTypeNone
	case "auto":
		tc.Type = schemas.ToolChoiceTypeAuto
	case "required":
		tc.Type = schemas.ToolChoiceTypeRequired
	default:
		tc.Type = schemas.ToolChoiceTypeAuto // fallback
	}
}

// parseFunctionCall handles legacy function_call format
func (r *OpenAIChatRequest) parseFunctionCall(functionCall interface{}) *schemas.ToolChoice {
	tc := &schemas.ToolChoice{}

	switch v := functionCall.(type) {
	case string:
		switch v {
		case "none":
			tc.Type = schemas.ToolChoiceTypeNone
		case "auto":
			tc.Type = schemas.ToolChoiceTypeAuto
		default:
			tc.Type = schemas.ToolChoiceTypeAuto // fallback
		}
	case map[string]interface{}:
		if name, ok := v["name"].(string); ok {
			tc.Type = schemas.ToolChoiceTypeFunction
			tc.Function = schemas.ToolChoiceFunction{Name: name}
		}
	}

	return tc
}

// ensureParams ensures bifrostReq.Params is initialized
func (r *OpenAIChatRequest) ensureParams(bifrostReq *schemas.BifrostRequest) {
	if bifrostReq.Params == nil {
		bifrostReq.Params = &schemas.ModelParameters{}
	}
}

// extractLegacyFunctionCall returns the FunctionCall for legacy compatibility
// when exactly one function tool-call is present, otherwise returns nil
func extractLegacyFunctionCall(toolCalls []schemas.ToolCall) *schemas.FunctionCall {
	if len(toolCalls) == 1 && toolCalls[0].Type != nil && *toolCalls[0].Type == string(schemas.ToolChoiceTypeFunction) {
		return &toolCalls[0].Function
	}
	return nil
}

// DeriveOpenAIFromBifrostResponse converts a Bifrost response to OpenAI format
func DeriveOpenAIFromBifrostResponse(bifrostResp *schemas.BifrostResponse) *OpenAIChatResponse {
	if bifrostResp == nil {
		return nil
	}

	openaiResp := &OpenAIChatResponse{
		ID:      bifrostResp.ID,
		Object:  "chat.completion",
		Created: bifrostResp.Created,
		Model:   bifrostResp.Model,
		Choices: make([]OpenAIChoice, len(bifrostResp.Choices)),
	}

	if bifrostResp.SystemFingerprint != nil {
		openaiResp.SystemFingerprint = bifrostResp.SystemFingerprint
	}

	// Convert usage information (using schemas.LLMUsage directly)
	if bifrostResp.Usage != (schemas.LLMUsage{}) {
		openaiResp.Usage = &bifrostResp.Usage
	}

	// Convert choices
	for i, choice := range bifrostResp.Choices {
		openaiChoice := OpenAIChoice{
			Index:        choice.Index,
			FinishReason: choice.FinishReason,
		}

		// Convert message
		msg := OpenAIMessage{
			Role: string(choice.Message.Role),
		}

		// Convert content back to proper format
		if choice.Message.Content != nil {
			msg.Content = *choice.Message.Content
		}

		// Convert tool calls for assistant messages (already in schemas.ToolCall format)
		if choice.Message.AssistantMessage != nil && choice.Message.AssistantMessage.ToolCalls != nil {
			msg.ToolCalls = choice.Message.AssistantMessage.ToolCalls
			msg.FunctionCall = extractLegacyFunctionCall(*choice.Message.AssistantMessage.ToolCalls)
		}

		// Handle tool messages - propagate tool_call_id
		if choice.Message.ToolMessage != nil && choice.Message.ToolMessage.ToolCallID != nil {
			msg.ToolCallID = choice.Message.ToolMessage.ToolCallID
		}

		openaiChoice.Message = msg
		openaiResp.Choices[i] = openaiChoice
	}

	return openaiResp
}
