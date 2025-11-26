package integrations

import (
	"context"
	"errors"

	bifrost "github.com/maximhq/bifrost/core"
	"github.com/maximhq/bifrost/core/providers/cohere"
	"github.com/maximhq/bifrost/core/schemas"
	"github.com/maximhq/bifrost/transports/bifrost-http/lib"
)

// CohereRouter holds route registrations for Cohere endpoints.
// It supports Cohere's v2 chat API and embeddings API.
type CohereRouter struct {
	*GenericRouter
}

// NewCohereRouter creates a new CohereRouter with the given bifrost client.
func NewCohereRouter(client *bifrost.Bifrost, handlerStore lib.HandlerStore, logger schemas.Logger) *CohereRouter {
	return &CohereRouter{
		GenericRouter: NewGenericRouter(client, handlerStore, CreateCohereRouteConfigs("/cohere"), logger),
	}
}

// CreateCohereRouteConfigs creates route configurations for Cohere v2 API endpoints.
func CreateCohereRouteConfigs(pathPrefix string) []RouteConfig {
	var routes []RouteConfig

	// Chat completions endpoint (v2/chat)
	routes = append(routes, RouteConfig{
		Path:   pathPrefix + "/v2/chat",
		Method: "POST",
		GetRequestTypeInstance: func() interface{} {
			return &cohere.CohereChatRequest{}
		},
		RequestConverter: func(ctx *context.Context, req interface{}) (*schemas.BifrostRequest, error) {
			if cohereReq, ok := req.(*cohere.CohereChatRequest); ok {
				return &schemas.BifrostRequest{
					ChatRequest: cohereChatRequestToBifrost(cohereReq),
				}, nil
			}
			return nil, errors.New("invalid request type")
		},
		ChatResponseConverter: func(ctx *context.Context, resp *schemas.BifrostChatResponse) (interface{}, error) {
			if resp.ExtraFields.Provider == schemas.Cohere {
				if resp.ExtraFields.RawResponse != nil {
					return resp.ExtraFields.RawResponse, nil
				}
			}
			return resp, nil
		},
		ErrorConverter: func(ctx *context.Context, err *schemas.BifrostError) interface{} {
			return err
		},
		StreamConfig: &StreamConfig{
			ChatStreamResponseConverter: func(ctx *context.Context, resp *schemas.BifrostChatResponse) (string, interface{}, error) {
				if resp.ExtraFields.Provider == schemas.Cohere {
					if resp.ExtraFields.RawResponse != nil {
						return "", resp.ExtraFields.RawResponse, nil
					}
				}
				return "", resp, nil
			},
			ErrorConverter: func(ctx *context.Context, err *schemas.BifrostError) interface{} {
				return err
			},
		},
	})

	// Embeddings endpoint (v2/embed)
	routes = append(routes, RouteConfig{
		Path:   pathPrefix + "/v2/embed",
		Method: "POST",
		GetRequestTypeInstance: func() interface{} {
			return &cohere.CohereEmbeddingRequest{}
		},
		RequestConverter: func(ctx *context.Context, req interface{}) (*schemas.BifrostRequest, error) {
			if cohereReq, ok := req.(*cohere.CohereEmbeddingRequest); ok {
				return &schemas.BifrostRequest{
					EmbeddingRequest: cohereEmbeddingRequestToBifrost(cohereReq),
				}, nil
			}
			return nil, errors.New("invalid embedding request type")
		},
		EmbeddingResponseConverter: func(ctx *context.Context, resp *schemas.BifrostEmbeddingResponse) (interface{}, error) {
			if resp.ExtraFields.Provider == schemas.Cohere {
				if resp.ExtraFields.RawResponse != nil {
					return resp.ExtraFields.RawResponse, nil
				}
			}
			return resp, nil
		},
		ErrorConverter: func(ctx *context.Context, err *schemas.BifrostError) interface{} {
			return err
		},
	})

	return routes
}

// cohereChatRequestToBifrost converts a Cohere v2 chat request to Bifrost format
func cohereChatRequestToBifrost(req *cohere.CohereChatRequest) *schemas.BifrostChatRequest {
	if req == nil {
		return nil
	}

	provider, model := schemas.ParseModelString(req.Model, schemas.Cohere)

	bifrostReq := &schemas.BifrostChatRequest{
		Provider: provider,
		Model:    model,
		Input:    convertCohereMessagesToBifrost(req.Messages),
		Params:   &schemas.ChatParameters{},
	}

	// Convert parameters
	if req.MaxTokens != nil {
		bifrostReq.Params.MaxCompletionTokens = req.MaxTokens
	}
	if req.Temperature != nil {
		bifrostReq.Params.Temperature = req.Temperature
	}
	if req.P != nil {
		bifrostReq.Params.TopP = req.P
	}
	if req.StopSequences != nil {
		bifrostReq.Params.Stop = req.StopSequences
	}
	if req.FrequencyPenalty != nil {
		bifrostReq.Params.FrequencyPenalty = req.FrequencyPenalty
	}
	if req.PresencePenalty != nil {
		bifrostReq.Params.PresencePenalty = req.PresencePenalty
	}

	// Convert tools
	if req.Tools != nil {
		bifrostTools := make([]schemas.ChatTool, len(req.Tools))
		for i, tool := range req.Tools {
			bifrostTools[i] = schemas.ChatTool{
				Type: schemas.ChatToolTypeFunction,
				Function: &schemas.ChatToolFunction{
					Name:        tool.Function.Name,
					Description: tool.Function.Description,
					Parameters:  convertInterfaceToToolFunctionParameters(tool.Function.Parameters),
				},
			}
		}
		bifrostReq.Params.Tools = bifrostTools
	}

	// Convert tool choice
	if req.ToolChoice != nil {
		switch *req.ToolChoice {
		case cohere.ToolChoiceNone:
			bifrostReq.Params.ToolChoice = &schemas.ChatToolChoice{
				ChatToolChoiceStr: schemas.Ptr(string(schemas.ChatToolChoiceTypeNone)),
			}
		case cohere.ToolChoiceRequired:
			bifrostReq.Params.ToolChoice = &schemas.ChatToolChoice{
				ChatToolChoiceStr: schemas.Ptr(string(schemas.ChatToolChoiceTypeRequired)),
			}
		case cohere.ToolChoiceAuto:
			bifrostReq.Params.ToolChoice = &schemas.ChatToolChoice{
				ChatToolChoiceStr: schemas.Ptr(string(schemas.ChatToolChoiceTypeAny)),
			}
		}
	}

	// Convert extra params
	extraParams := make(map[string]interface{})
	if req.SafetyMode != nil {
		extraParams["safety_mode"] = *req.SafetyMode
	}
	if req.LogProbs != nil {
		extraParams["log_probs"] = *req.LogProbs
	}
	if req.StrictToolChoice != nil {
		extraParams["strict_tool_choice"] = *req.StrictToolChoice
	}
	if req.Thinking != nil {
		thinkingMap := map[string]interface{}{
			"type": string(req.Thinking.Type),
		}
		if req.Thinking.TokenBudget != nil {
			thinkingMap["token_budget"] = *req.Thinking.TokenBudget
		}
		extraParams["thinking"] = thinkingMap
	}
	if len(extraParams) > 0 {
		bifrostReq.Params.ExtraParams = extraParams
	}

	return bifrostReq
}

// convertCohereMessagesToBifrost converts Cohere messages to Bifrost format
func convertCohereMessagesToBifrost(messages []cohere.CohereMessage) []schemas.ChatMessage {
	if messages == nil {
		return nil
	}

	bifrostMessages := make([]schemas.ChatMessage, len(messages))
	for i, msg := range messages {
		bifrostMsg := schemas.ChatMessage{
			Role: schemas.ChatMessageRole(msg.Role),
		}

		// Convert content
		if msg.Content != nil {
			if msg.Content.IsString() {
				bifrostMsg.Content = &schemas.ChatMessageContent{
					ContentStr: msg.Content.GetString(),
				}
			} else if msg.Content.IsBlocks() {
				var contentBlocks []schemas.ChatContentBlock
				for _, block := range msg.Content.GetBlocks() {
					switch block.Type {
					case cohere.CohereContentBlockTypeText:
						contentBlocks = append(contentBlocks, schemas.ChatContentBlock{
							Type: schemas.ChatContentBlockTypeText,
							Text: block.Text,
						})
					case cohere.CohereContentBlockTypeImage:
						if block.ImageURL != nil {
							contentBlocks = append(contentBlocks, schemas.ChatContentBlock{
								Type: schemas.ChatContentBlockTypeImage,
								ImageURLStruct: &schemas.ChatInputImage{
									URL: block.ImageURL.URL,
								},
							})
						}
					}
				}
				if len(contentBlocks) > 0 {
					bifrostMsg.Content = &schemas.ChatMessageContent{
						ContentBlocks: contentBlocks,
					}
				}
			}
		}

		// Convert tool calls (for assistant messages)
		if msg.ToolCalls != nil {
			var toolCalls []schemas.ChatAssistantMessageToolCall
			for j, tc := range msg.ToolCalls {
				toolCall := schemas.ChatAssistantMessageToolCall{
					Index: uint16(j),
					ID:    tc.ID,
				}
				if tc.Function != nil {
					toolCall.Function = schemas.ChatAssistantMessageToolCallFunction{
						Name:      tc.Function.Name,
						Arguments: tc.Function.Arguments,
					}
				}
				toolCalls = append(toolCalls, toolCall)
			}
			if len(toolCalls) > 0 {
				bifrostMsg.ChatAssistantMessage = &schemas.ChatAssistantMessage{
					ToolCalls: toolCalls,
				}
			}
		}

		// Convert tool call ID (for tool messages)
		if msg.ToolCallID != nil {
			bifrostMsg.ChatToolMessage = &schemas.ChatToolMessage{
				ToolCallID: msg.ToolCallID,
			}
		}

		bifrostMessages[i] = bifrostMsg
	}

	return bifrostMessages
}

// cohereEmbeddingRequestToBifrost converts a Cohere embedding request to Bifrost format
func cohereEmbeddingRequestToBifrost(req *cohere.CohereEmbeddingRequest) *schemas.BifrostEmbeddingRequest {
	if req == nil {
		return nil
	}

	provider, model := schemas.ParseModelString(req.Model, schemas.Cohere)

	bifrostReq := &schemas.BifrostEmbeddingRequest{
		Provider: provider,
		Model:    model,
		Input:    &schemas.EmbeddingInput{},
		Params:   &schemas.EmbeddingParameters{},
	}

	// Convert texts
	if len(req.Texts) > 0 {
		if len(req.Texts) == 1 {
			bifrostReq.Input.Text = &req.Texts[0]
		} else {
			bifrostReq.Input.Texts = req.Texts
		}
	}

	// Convert parameters
	if req.OutputDimension != nil {
		bifrostReq.Params.Dimensions = req.OutputDimension
	}

	// Convert extra params
	extraParams := make(map[string]interface{})
	if req.InputType != "" {
		extraParams["input_type"] = req.InputType
	}
	if req.EmbeddingTypes != nil {
		extraParams["embedding_types"] = req.EmbeddingTypes
	}
	if req.Truncate != nil {
		extraParams["truncate"] = *req.Truncate
	}
	if req.MaxTokens != nil {
		extraParams["max_tokens"] = *req.MaxTokens
	}
	if len(extraParams) > 0 {
		bifrostReq.Params.ExtraParams = extraParams
	}

	return bifrostReq
}

// convertInterfaceToToolFunctionParameters converts an interface{} to ToolFunctionParameters
// This handles the conversion from Cohere's flexible parameter format to Bifrost's structured format
func convertInterfaceToToolFunctionParameters(params interface{}) *schemas.ToolFunctionParameters {
	if params == nil {
		return nil
	}

	// Try to convert from map[string]interface{}
	paramsMap, ok := params.(map[string]interface{})
	if !ok {
		return nil
	}

	result := &schemas.ToolFunctionParameters{}

	// Extract type
	if typeVal, ok := paramsMap["type"].(string); ok {
		result.Type = typeVal
	}

	// Extract description
	if descVal, ok := paramsMap["description"].(string); ok {
		result.Description = &descVal
	}

	// Extract required
	if requiredVal, ok := paramsMap["required"].([]interface{}); ok {
		required := make([]string, 0, len(requiredVal))
		for _, v := range requiredVal {
			if s, ok := v.(string); ok {
				required = append(required, s)
			}
		}
		result.Required = required
	}

	// Extract properties
	if propsVal, ok := paramsMap["properties"].(map[string]interface{}); ok {
		result.Properties = &propsVal
	}

	// Extract enum
	if enumVal, ok := paramsMap["enum"].([]interface{}); ok {
		enum := make([]string, 0, len(enumVal))
		for _, v := range enumVal {
			if s, ok := v.(string); ok {
				enum = append(enum, s)
			}
		}
		result.Enum = enum
	}

	// Extract additionalProperties
	if addPropsVal, ok := paramsMap["additionalProperties"].(bool); ok {
		result.AdditionalProperties = &addPropsVal
	}

	return result
}
