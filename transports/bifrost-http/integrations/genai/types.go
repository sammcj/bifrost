package genai

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"

	bifrost "github.com/maximhq/bifrost/core"
	"github.com/maximhq/bifrost/core/schemas"
	genai_sdk "google.golang.org/genai"
)

var fnTypePtr = bifrost.Ptr(string(schemas.ToolChoiceTypeFunction))

// ensureExtraParams ensures that bifrostReq.Params and bifrostReq.Params.ExtraParams are initialized
func ensureExtraParams(bifrostReq *schemas.BifrostRequest) {
	if bifrostReq.Params == nil {
		bifrostReq.Params = &schemas.ModelParameters{
			ExtraParams: make(map[string]interface{}),
		}
	}
	if bifrostReq.Params.ExtraParams == nil {
		bifrostReq.Params.ExtraParams = make(map[string]interface{})
	}
}

type GeminiChatRequest struct {
	Model              string                     `json:"model,omitempty"` // Model field for explicit model specification
	Contents           []genai_sdk.Content        `json:"contents"`
	SystemInstruction  *genai_sdk.Content         `json:"systemInstruction,omitempty"`
	GenerationConfig   genai_sdk.GenerationConfig `json:"generationConfig,omitempty"`
	SafetySettings     []genai_sdk.SafetySetting  `json:"safetySettings,omitempty"`
	Tools              []genai_sdk.Tool           `json:"tools,omitempty"`
	ToolConfig         genai_sdk.ToolConfig       `json:"toolConfig,omitempty"`
	Labels             map[string]string          `json:"labels,omitempty"`
	CachedContent      string                     `json:"cachedContent,omitempty"`
	ResponseModalities []string                   `json:"responseModalities,omitempty"`
}

func (r *GeminiChatRequest) ConvertToBifrostRequest() *schemas.BifrostRequest {
	bifrostReq := &schemas.BifrostRequest{
		Provider: schemas.Vertex,
		Model:    r.Model,
		Input: schemas.RequestInput{
			ChatCompletionInput: &[]schemas.BifrostMessage{},
		},
	}

	// Convert system instruction if present
	if r.SystemInstruction != nil {
		systemMsgs := r.convertContentToBifrostMessages(*r.SystemInstruction, schemas.ModelChatMessageRoleSystem)
		*bifrostReq.Input.ChatCompletionInput = append(*bifrostReq.Input.ChatCompletionInput, systemMsgs...)
	}

	// Convert messages (contents)
	for _, content := range r.Contents {
		messages := r.convertContentToBifrostMessages(content, schemas.ModelChatMessageRole(content.Role))
		*bifrostReq.Input.ChatCompletionInput = append(*bifrostReq.Input.ChatCompletionInput, messages...)
	}

	// Convert generation config to parameters
	if params := r.convertGenerationConfigToParams(); params != nil {
		bifrostReq.Params = params
	}

	// Convert safety settings
	if len(r.SafetySettings) > 0 {
		ensureExtraParams(bifrostReq)
		bifrostReq.Params.ExtraParams["safety_settings"] = r.SafetySettings
	}

	// Convert additional request fields
	if r.CachedContent != "" {
		ensureExtraParams(bifrostReq)
		bifrostReq.Params.ExtraParams["cached_content"] = r.CachedContent
	}

	// Convert response modalities
	if len(r.ResponseModalities) > 0 {
		ensureExtraParams(bifrostReq)
		bifrostReq.Params.ExtraParams["response_modalities"] = r.ResponseModalities
	}

	// Convert labels
	if len(r.Labels) > 0 {
		ensureExtraParams(bifrostReq)
		bifrostReq.Params.ExtraParams["labels"] = r.Labels
	}

	// Convert tools and tool config
	if len(r.Tools) > 0 {
		ensureExtraParams(bifrostReq)

		tools := make([]schemas.Tool, 0, len(r.Tools))
		for _, tool := range r.Tools {
			if len(tool.FunctionDeclarations) > 0 {
				for _, fn := range tool.FunctionDeclarations {
					bifrostTool := schemas.Tool{
						Type: "function",
						Function: schemas.Function{
							Name:        fn.Name,
							Description: fn.Description,
						},
					}
					// Convert parameters schema if present
					if fn.Parameters != nil {
						bifrostTool.Function.Parameters = r.convertSchemaToFunctionParameters(fn.Parameters)
					}
					tools = append(tools, bifrostTool)
				}
			}
			// Handle other tool types (Retrieval, GoogleSearch, etc.) as ExtraParams
			if tool.Retrieval != nil {
				bifrostReq.Params.ExtraParams["retrieval"] = tool.Retrieval
			}
			if tool.GoogleSearch != nil {
				bifrostReq.Params.ExtraParams["google_search"] = tool.GoogleSearch
			}
			if tool.CodeExecution != nil {
				bifrostReq.Params.ExtraParams["code_execution"] = tool.CodeExecution
			}
		}

		if len(tools) > 0 {
			bifrostReq.Params.Tools = &tools
		}
	}

	// Convert tool config
	if r.ToolConfig.FunctionCallingConfig != nil || r.ToolConfig.RetrievalConfig != nil {
		ensureExtraParams(bifrostReq)
		bifrostReq.Params.ExtraParams["tool_config"] = r.ToolConfig
	}

	return bifrostReq
}

// convertContentToBifrostMessage converts a Gemini Content to BifrostMessage(s)
// Returns multiple messages when there are multiple images to ensure each image gets its own message
func (r *GeminiChatRequest) convertContentToBifrostMessages(content genai_sdk.Content, role schemas.ModelChatMessageRole) []schemas.BifrostMessage {
	if len(content.Parts) == 0 {
		return nil
	}

	// Handle multiple parts - concatenate text parts and handle other types
	var textParts []string
	var toolCalls []schemas.ToolCall
	var imageContents []schemas.ImageContent

	for _, part := range content.Parts {
		switch {
		case part.Text != "":
			textParts = append(textParts, part.Text)

		case part.FunctionCall != nil:
			jsonArgs, err := json.Marshal(part.FunctionCall.Args)
			if err != nil {
				jsonArgs = []byte(fmt.Sprintf("%v", part.FunctionCall.Args))
			}
			toolCall := schemas.ToolCall{
				ID:   bifrost.Ptr(part.FunctionCall.ID),
				Type: fnTypePtr,
				Function: schemas.FunctionCall{
					Name:      &part.FunctionCall.Name,
					Arguments: string(jsonArgs),
				},
			}

			toolCalls = append(toolCalls, toolCall)

		case part.InlineData != nil:
			// Handle inline images/media
			imageContent := schemas.ImageContent{
				Type:      schemas.ImageContentTypeBase64,
				URL:       fmt.Sprintf("data:%s;base64,%s", part.InlineData.MIMEType, base64.StdEncoding.EncodeToString(part.InlineData.Data)),
				MediaType: &part.InlineData.MIMEType,
			}
			imageContents = append(imageContents, imageContent)

		case part.FileData != nil:
			// Handle file references
			imageContent := schemas.ImageContent{
				Type:      schemas.ImageContentTypeURL,
				URL:       part.FileData.FileURI,
				MediaType: &part.FileData.MIMEType,
			}
			imageContents = append(imageContents, imageContent)

		case part.FunctionResponse != nil:
			responseContent, err := json.Marshal(part.FunctionResponse.Response)
			if err != nil {
				responseContent = []byte(fmt.Sprintf("%v", part.FunctionResponse.Response))
			}

			toolResponseMsg := schemas.BifrostMessage{
				Role:    schemas.ModelChatMessageRoleTool,
				Content: bifrost.Ptr(string(responseContent)),
				ToolMessage: &schemas.ToolMessage{
					ToolCallID: &part.FunctionResponse.Name,
				},
			}

			return []schemas.BifrostMessage{toolResponseMsg}
		}
	}

	var messages []schemas.BifrostMessage

	// Create main message with text content and tool calls
	mainMsg := schemas.BifrostMessage{
		Role: role,
	}

	// Set text content if we have any
	if len(textParts) > 0 {
		combinedText := strings.Join(textParts, "\n\n")
		mainMsg.Content = &combinedText
	}

	// Set tool calls if we have any
	if len(toolCalls) > 0 && role == schemas.ModelChatMessageRoleAssistant {
		mainMsg.AssistantMessage = &schemas.AssistantMessage{
			ToolCalls: &toolCalls,
		}
	}

	// Add main message if it has content or tool calls
	if mainMsg.Content != nil || (mainMsg.AssistantMessage != nil && mainMsg.AssistantMessage.ToolCalls != nil) {
		messages = append(messages, mainMsg)
	}

	// Create separate messages for each image
	for _, imageContent := range imageContents {
		imageMsg := schemas.BifrostMessage{
			Role: role,
		}

		// Set image content based on role
		switch role {
		case schemas.ModelChatMessageRoleUser:
			imageMsg.UserMessage = &schemas.UserMessage{
				ImageContent: &imageContent,
			}
			messages = append(messages, imageMsg)

		case schemas.ModelChatMessageRoleTool:
			imageMsg.ToolMessage = &schemas.ToolMessage{
				ImageContent: &imageContent,
			}
			messages = append(messages, imageMsg)
		}
	}

	return messages
}

// convertGenerationConfigToParams converts Gemini GenerationConfig to ModelParameters
func (r *GeminiChatRequest) convertGenerationConfigToParams() *schemas.ModelParameters {
	params := &schemas.ModelParameters{
		ExtraParams: make(map[string]interface{}),
	}

	config := r.GenerationConfig

	// Map generation config fields to parameters
	if config.Temperature != nil {
		temp := float64(*config.Temperature)
		params.Temperature = &temp
	}
	if config.TopP != nil {
		params.TopP = bifrost.Ptr(float64(*config.TopP))
	}
	if config.TopK != nil {
		params.TopK = bifrost.Ptr(int(*config.TopK))
	}
	if config.MaxOutputTokens > 0 {
		maxTokens := int(config.MaxOutputTokens)
		params.MaxTokens = &maxTokens
	}
	if config.CandidateCount > 0 {
		params.ExtraParams["candidate_count"] = config.CandidateCount
	}
	if len(config.StopSequences) > 0 {
		params.StopSequences = &config.StopSequences
	}
	if config.PresencePenalty != nil {
		params.PresencePenalty = bifrost.Ptr(float64(*config.PresencePenalty))
	}
	if config.FrequencyPenalty != nil {
		params.FrequencyPenalty = bifrost.Ptr(float64(*config.FrequencyPenalty))
	}
	if config.Seed != nil {
		params.ExtraParams["seed"] = *config.Seed
	}
	if config.ResponseMIMEType != "" {
		params.ExtraParams["response_mime_type"] = config.ResponseMIMEType
	}
	if config.ResponseLogprobs {
		params.ExtraParams["response_logprobs"] = config.ResponseLogprobs
	}
	if config.Logprobs != nil {
		params.ExtraParams["logprobs"] = *config.Logprobs
	}

	return params
}

// convertSchemaToFunctionParameters converts genai.Schema to schemas.FunctionParameters
func (r *GeminiChatRequest) convertSchemaToFunctionParameters(schema *genai_sdk.Schema) schemas.FunctionParameters {
	params := schemas.FunctionParameters{
		Type: string(schema.Type),
	}

	if schema.Description != "" {
		params.Description = &schema.Description
	}

	if len(schema.Required) > 0 {
		params.Required = schema.Required
	}

	if len(schema.Properties) > 0 {
		params.Properties = make(map[string]interface{})
		for k, v := range schema.Properties {
			params.Properties[k] = v
		}
	}

	if len(schema.Enum) > 0 {
		params.Enum = &schema.Enum
	}

	return params
}

func DeriveGenAIFromBifrostResponse(bifrostResp *schemas.BifrostResponse) *genai_sdk.GenerateContentResponse {
	if bifrostResp == nil {
		return nil
	}

	genaiResp := &genai_sdk.GenerateContentResponse{
		Candidates: make([]*genai_sdk.Candidate, len(bifrostResp.Choices)),
	}

	if bifrostResp.Usage != (schemas.LLMUsage{}) {
		genaiResp.UsageMetadata = &genai_sdk.GenerateContentResponseUsageMetadata{
			PromptTokenCount:     int32(bifrostResp.Usage.PromptTokens),
			CandidatesTokenCount: int32(bifrostResp.Usage.CompletionTokens),
			TotalTokenCount:      int32(bifrostResp.Usage.TotalTokens),
		}
	}

	for i, choice := range bifrostResp.Choices {
		candidate := &genai_sdk.Candidate{
			Index: int32(choice.Index),
		}
		if choice.FinishReason != nil {
			candidate.FinishReason = genai_sdk.FinishReason(*choice.FinishReason)
		}

		if bifrostResp.Usage != (schemas.LLMUsage{}) {
			candidate.TokenCount = int32(bifrostResp.Usage.CompletionTokens)
		}

		parts := []*genai_sdk.Part{}
		if choice.Message.Content != nil && *choice.Message.Content != "" {
			parts = append(parts, &genai_sdk.Part{Text: *choice.Message.Content})
		}

		// Handle tool calls
		if choice.Message.AssistantMessage != nil && choice.Message.AssistantMessage.ToolCalls != nil {
			for _, toolCall := range *choice.Message.AssistantMessage.ToolCalls {
				argsMap := make(map[string]interface{})
				if toolCall.Function.Arguments != "" {
					// Attempt to unmarshal arguments, but don't fail if it's not valid JSON,
					// as BifrostResponse.FunctionCall.Arguments is a string.
					// genai.FunctionCall.Args expects map[string]any.
					json.Unmarshal([]byte(toolCall.Function.Arguments), &argsMap)
				}
				if toolCall.Function.Name != nil {
					fc := &genai_sdk.FunctionCall{
						Name: *toolCall.Function.Name,
						Args: argsMap,
					}
					if toolCall.ID != nil {
						fc.ID = *toolCall.ID
					}
					parts = append(parts, &genai_sdk.Part{FunctionCall: fc})
				}
			}
		}

		// Handle thinking content if present
		if choice.Message.AssistantMessage != nil && choice.Message.AssistantMessage.Thought != nil && *choice.Message.AssistantMessage.Thought != "" {
			parts = append(parts, &genai_sdk.Part{
				Text:    *choice.Message.AssistantMessage.Thought,
				Thought: true,
			})
		}

		if len(parts) > 0 {
			candidate.Content = &genai_sdk.Content{
				Parts: parts,
				Role:  string(choice.Message.Role),
			}
		}

		// Handle safety ratings if available (from ExtraFields)
		if bifrostResp.ExtraFields.RawResponse != nil {
			if rawMap, ok := bifrostResp.ExtraFields.RawResponse.(map[string]interface{}); ok {
				if candidates, ok := rawMap["candidates"].([]interface{}); ok && len(candidates) > i {
					if candidateMap, ok := candidates[i].(map[string]interface{}); ok {
						if safetyRatings, ok := candidateMap["safetyRatings"].([]interface{}); ok {
							var ratings []*genai_sdk.SafetyRating
							for _, rating := range safetyRatings {
								if ratingMap, ok := rating.(map[string]interface{}); ok {
									sr := &genai_sdk.SafetyRating{}
									if category, ok := ratingMap["category"].(string); ok {
										sr.Category = genai_sdk.HarmCategory(category)
									}
									if probability, ok := ratingMap["probability"].(string); ok {
										sr.Probability = genai_sdk.HarmProbability(probability)
									}
									if blocked, ok := ratingMap["blocked"].(bool); ok {
										sr.Blocked = blocked
									}
									ratings = append(ratings, sr)
								}
							}
							candidate.SafetyRatings = ratings
						}
					}
				}
			}
		}

		genaiResp.Candidates[i] = candidate
	}

	return genaiResp
}
