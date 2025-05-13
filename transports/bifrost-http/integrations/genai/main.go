package genai

import (
	"encoding/json"

	"github.com/maximhq/bifrost/core/schemas"
	genai_sdk "google.golang.org/genai"
)

type GeminiChatRequest struct {
	Contents         []genai_sdk.Content        `json:"contents"`
	GenerationConfig genai_sdk.GenerationConfig `json:"generationConfig,omitempty"`
	SafetySettings   []genai_sdk.SafetySetting  `json:"safetySettings,omitempty"`
	Tools            []genai_sdk.Tool           `json:"tools,omitempty"`
	ToolConfig       genai_sdk.ToolConfig       `json:"toolConfig,omitempty"`
	Labels           map[string]string          `json:"labels,omitempty"`
}

func (r *GeminiChatRequest) ConvertToBifrostRequest(modelStr string) *schemas.BifrostRequest {
	bifrostReq := &schemas.BifrostRequest{
		Model: modelStr,
		Input: schemas.RequestInput{
			ChatCompletionInput: &[]schemas.Message{},
		},
	}

	// Convert messages (contents)
	for _, content := range r.Contents {
		// Assuming each content has one part for simplicity,
		// and that part is text. You might need to adjust this based on actual use case.
		if len(content.Parts) > 0 {
			bifrostMsg := schemas.Message{
				Role:    schemas.ModelChatMessageRole(content.Role),
				Content: &content.Parts[0].Text,
			}
			*bifrostReq.Input.ChatCompletionInput = append(*bifrostReq.Input.ChatCompletionInput, bifrostMsg)
		}
	}

	return bifrostReq
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
			candidate.TokenCount = int32(bifrostResp.Usage.CompletionTokens) // Assuming total completion tokens apply here
		}

		parts := []*genai_sdk.Part{}
		if choice.Message.Content != nil && *choice.Message.Content != "" {
			parts = append(parts, &genai_sdk.Part{Text: *choice.Message.Content})
		}

		if choice.Message.ToolCalls != nil {
			for _, toolCall := range *choice.Message.ToolCalls {
				argsMap := make(map[string]interface{})
				if toolCall.Function.Arguments != "" {
					// Attempt to unmarshal arguments, but don't fail if it's not valid JSON,
					// as BifrostResponse.FunctionCall.Arguments is a string.
					// genai.FunctionCall.Args expects map[string]any.
					json.Unmarshal([]byte(toolCall.Function.Arguments), &argsMap)
				}
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

		if len(parts) > 0 {
			candidate.Content = &genai_sdk.Content{
				Parts: parts,
				Role:  string(choice.Message.Role),
			}
		}

		genaiResp.Candidates[i] = candidate
	}

	return genaiResp
}
