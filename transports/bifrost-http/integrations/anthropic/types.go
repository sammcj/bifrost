package anthropic

import (
	"encoding/json"
	"fmt"

	bifrost "github.com/maximhq/bifrost/core"
	"github.com/maximhq/bifrost/core/schemas"
)

var fnTypePtr = bifrost.Ptr(string(schemas.ToolChoiceTypeFunction))

// AnthropicContent represents content in Anthropic message format
type AnthropicContent struct {
	Type      string                `json:"type"`                  // "text", "image", "tool_use", "tool_result"
	Text      *string               `json:"text,omitempty"`        // For text content
	ToolUseID *string               `json:"tool_use_id,omitempty"` // For tool_result content
	ID        *string               `json:"id,omitempty"`          // For tool_use content
	Name      *string               `json:"name,omitempty"`        // For tool_use content
	Input     interface{}           `json:"input,omitempty"`       // For tool_use content
	Content   interface{}           `json:"content,omitempty"`     // For tool_result content
	Source    *AnthropicImageSource `json:"source,omitempty"`      // For image content
}

// AnthropicImageSource represents image source in Anthropic format
type AnthropicImageSource struct {
	Type      string `json:"type"`       // "base64" or "url"
	MediaType string `json:"media_type"` // "image/jpeg", "image/png", etc.
	Data      string `json:"data"`       // Base64-encoded image data
}

// AnthropicMessage represents a message in Anthropic format
type AnthropicMessage struct {
	Role    string             `json:"role"`    // "user", "assistant"
	Content []AnthropicContent `json:"content"` // Array of content blocks
}

// UnmarshalJSON implements custom JSON unmarshaling for AnthropicMessage
// to handle both string and array content formats
func (m *AnthropicMessage) UnmarshalJSON(data []byte) error {
	// First, try to unmarshal into a struct with Content as json.RawMessage
	var temp struct {
		Role    string          `json:"role"`
		Content json.RawMessage `json:"content"`
	}

	if err := json.Unmarshal(data, &temp); err != nil {
		return err
	}

	m.Role = temp.Role

	// Try to unmarshal content as string first
	var contentStr string
	if err := json.Unmarshal(temp.Content, &contentStr); err == nil {
		// It's a string, convert to AnthropicContent array
		m.Content = []AnthropicContent{
			{
				Type: "text",
				Text: &contentStr,
			},
		}
		return nil
	}

	// Try to unmarshal as array of AnthropicContent
	var contentArray []AnthropicContent
	if err := json.Unmarshal(temp.Content, &contentArray); err == nil {
		m.Content = contentArray
		return nil
	}

	return fmt.Errorf("content must be either a string or an array of content objects")
}

// AnthropicTool represents a tool in Anthropic format
type AnthropicTool struct {
	Name        string      `json:"name"`
	Description string      `json:"description"`
	InputSchema interface{} `json:"input_schema"`
}

// AnthropicToolChoice represents tool choice in Anthropic format
type AnthropicToolChoice struct {
	Type string `json:"type"`           // "auto", "any", "tool"
	Name string `json:"name,omitempty"` // For type "tool"
}

// AnthropicMessageRequest represents an Anthropic messages API request
type AnthropicMessageRequest struct {
	Model         string               `json:"model"`
	MaxTokens     int                  `json:"max_tokens"`
	Messages      []AnthropicMessage   `json:"messages"`
	System        *string              `json:"system,omitempty"`
	Temperature   *float64             `json:"temperature,omitempty"`
	TopP          *float64             `json:"top_p,omitempty"`
	TopK          *int                 `json:"top_k,omitempty"`
	StopSequences *[]string            `json:"stop_sequences,omitempty"`
	Stream        *bool                `json:"stream,omitempty"`
	Tools         *[]AnthropicTool     `json:"tools,omitempty"`
	ToolChoice    *AnthropicToolChoice `json:"tool_choice,omitempty"`
}

// ConvertToBifrostRequest converts an Anthropic messages request to Bifrost format
func (r *AnthropicMessageRequest) ConvertToBifrostRequest() *schemas.BifrostRequest {
	bifrostReq := &schemas.BifrostRequest{
		Provider: schemas.Anthropic,
		Model:    r.Model,
		Input: schemas.RequestInput{
			ChatCompletionInput: &[]schemas.BifrostMessage{},
		},
	}

	// Add system message if present
	if r.System != nil && *r.System != "" {
		systemMsg := schemas.BifrostMessage{
			Role:    schemas.ModelChatMessageRoleSystem,
			Content: r.System,
		}
		*bifrostReq.Input.ChatCompletionInput = append(*bifrostReq.Input.ChatCompletionInput, systemMsg)
	}

	// Convert messages
	for _, msg := range r.Messages {
		var bifrostMsg schemas.BifrostMessage
		bifrostMsg.Role = schemas.ModelChatMessageRole(msg.Role)

		// Handle different content types
		var toolCalls []schemas.ToolCall
		var textContents []string

		for _, content := range msg.Content {
			switch content.Type {
			case "text":
				if content.Text != nil {
					textContents = append(textContents, *content.Text)
				}
			case "image":
				if content.Source != nil {
					bifrostMsg.UserMessage = &schemas.UserMessage{
						ImageContent: &schemas.ImageContent{
							Type:      bifrost.Ptr(content.Source.Type),
							URL:       content.Source.Data,
							MediaType: &content.Source.MediaType,
						},
					}
				}
			case "tool_use":
				if content.ID != nil && content.Name != nil {
					tc := schemas.ToolCall{
						Type: fnTypePtr,
						ID:   content.ID,
						Function: schemas.FunctionCall{
							Name:      content.Name,
							Arguments: jsonifyInput(content.Input),
						},
					}
					toolCalls = append(toolCalls, tc)
				}
			case "tool_result":
				if content.ToolUseID != nil {
					bifrostMsg.ToolMessage = &schemas.ToolMessage{
						ToolCallID: content.ToolUseID,
					}
					if content.Content != nil {
						if str, ok := content.Content.(string); ok {
							bifrostMsg.Content = &str
						} else {
							jsonStr := jsonifyInput(content.Content)
							bifrostMsg.Content = &jsonStr
						}
					}
					if content.Source != nil {
						bifrostMsg.ToolMessage.ImageContent = &schemas.ImageContent{
							Type:      bifrost.Ptr(content.Source.Type),
							URL:       content.Source.Data,
							MediaType: &content.Source.MediaType,
						}
					}
					bifrostMsg.Role = schemas.ModelChatMessageRoleTool
				}
			}
		}

		// Concatenate all text contents
		if len(textContents) > 0 {
			concatenatedText := ""
			for i, text := range textContents {
				if i > 0 {
					concatenatedText += "\n" // Add newline separator between text blocks
				}
				concatenatedText += text
			}
			bifrostMsg.Content = &concatenatedText
		}

		if len(toolCalls) > 0 {
			bifrostMsg.AssistantMessage = &schemas.AssistantMessage{
				ToolCalls: &toolCalls,
			}
		}

		*bifrostReq.Input.ChatCompletionInput = append(*bifrostReq.Input.ChatCompletionInput, bifrostMsg)
	}

	// Convert parameters
	if r.MaxTokens > 0 || r.Temperature != nil || r.TopP != nil || r.TopK != nil || r.StopSequences != nil {
		params := &schemas.ModelParameters{}

		if r.MaxTokens > 0 {
			params.MaxTokens = &r.MaxTokens
		}
		if r.Temperature != nil {
			params.Temperature = r.Temperature
		}
		if r.TopP != nil {
			params.TopP = r.TopP
		}
		if r.TopK != nil {
			params.TopK = r.TopK
		}
		if r.StopSequences != nil {
			params.StopSequences = r.StopSequences
		}

		bifrostReq.Params = params
	}

	// Convert tools
	if r.Tools != nil {
		tools := []schemas.Tool{}
		for _, tool := range *r.Tools {
			// Convert input_schema to FunctionParameters
			params := schemas.FunctionParameters{
				Type: "object",
			}
			if tool.InputSchema != nil {
				if schemaMap, ok := tool.InputSchema.(map[string]interface{}); ok {
					if typeVal, ok := schemaMap["type"].(string); ok {
						params.Type = typeVal
					}
					if desc, ok := schemaMap["description"].(string); ok {
						params.Description = &desc
					}
					if required, ok := schemaMap["required"].([]interface{}); ok {
						reqStrings := make([]string, len(required))
						for i, req := range required {
							if reqStr, ok := req.(string); ok {
								reqStrings[i] = reqStr
							}
						}
						params.Required = reqStrings
					}
					if properties, ok := schemaMap["properties"].(map[string]interface{}); ok {
						params.Properties = properties
					}
					if enum, ok := schemaMap["enum"].([]interface{}); ok {
						enumStrings := make([]string, len(enum))
						for i, e := range enum {
							if eStr, ok := e.(string); ok {
								enumStrings[i] = eStr
							}
						}
						params.Enum = &enumStrings
					}
				}
			}

			t := schemas.Tool{
				Type: "function",
				Function: schemas.Function{
					Name:        tool.Name,
					Description: tool.Description,
					Parameters:  params,
				},
			}
			tools = append(tools, t)
		}
		if bifrostReq.Params == nil {
			bifrostReq.Params = &schemas.ModelParameters{}
		}
		bifrostReq.Params.Tools = &tools
	}

	// Convert tool choice
	if r.ToolChoice != nil {
		if bifrostReq.Params == nil {
			bifrostReq.Params = &schemas.ModelParameters{}
		}
		toolChoice := &schemas.ToolChoice{
			Type: func() schemas.ToolChoiceType {
				if r.ToolChoice.Type == "tool" {
					return schemas.ToolChoiceTypeFunction
				}
				return schemas.ToolChoiceType(r.ToolChoice.Type)
			}(),
		}
		if r.ToolChoice.Type == "tool" && r.ToolChoice.Name != "" {
			toolChoice.Function = schemas.ToolChoiceFunction{
				Name: r.ToolChoice.Name,
			}
		}
		bifrostReq.Params.ToolChoice = toolChoice
	}

	return bifrostReq
}

// Helper function to convert interface{} to JSON string
func jsonifyInput(input interface{}) string {
	if input == nil {
		return "{}"
	}
	jsonBytes, err := json.Marshal(input)
	if err != nil {
		return "{}"
	}
	return string(jsonBytes)
}

// AnthropicMessageResponse represents an Anthropic messages API response
type AnthropicMessageResponse struct {
	ID           string             `json:"id"`
	Type         string             `json:"type"`
	Role         string             `json:"role"`
	Content      []AnthropicContent `json:"content"`
	Model        string             `json:"model"`
	StopReason   *string            `json:"stop_reason,omitempty"`
	StopSequence *string            `json:"stop_sequence,omitempty"`
	Usage        *AnthropicUsage    `json:"usage,omitempty"`
}

// AnthropicUsage represents usage information in Anthropic format
type AnthropicUsage struct {
	InputTokens  int `json:"input_tokens"`
	OutputTokens int `json:"output_tokens"`
}

// DeriveAnthropicFromBifrostResponse converts a Bifrost response to Anthropic format
func DeriveAnthropicFromBifrostResponse(bifrostResp *schemas.BifrostResponse) *AnthropicMessageResponse {
	if bifrostResp == nil {
		return nil
	}

	anthropicResp := &AnthropicMessageResponse{
		ID:    bifrostResp.ID,
		Type:  "message",
		Role:  "assistant",
		Model: bifrostResp.Model,
	}

	// Convert usage information
	if bifrostResp.Usage != (schemas.LLMUsage{}) {
		anthropicResp.Usage = &AnthropicUsage{
			InputTokens:  bifrostResp.Usage.PromptTokens,
			OutputTokens: bifrostResp.Usage.CompletionTokens,
		}
	}

	// Convert choices to content
	var content []AnthropicContent
	if len(bifrostResp.Choices) > 0 {
		choice := bifrostResp.Choices[0] // Anthropic typically returns one choice

		if choice.FinishReason != nil {
			anthropicResp.StopReason = choice.FinishReason
		}
		if choice.StopString != nil {
			anthropicResp.StopSequence = choice.StopString
		}

		// Add thinking content if present
		if choice.Message.AssistantMessage != nil && choice.Message.AssistantMessage.Thought != nil && *choice.Message.AssistantMessage.Thought != "" {
			content = append(content, AnthropicContent{
				Type: "thinking",
				Text: choice.Message.AssistantMessage.Thought,
			})
		}

		// Add text content
		if choice.Message.Content != nil && *choice.Message.Content != "" {
			content = append(content, AnthropicContent{
				Type: "text",
				Text: choice.Message.Content,
			})
		}

		// Add tool calls as tool_use content
		if choice.Message.AssistantMessage != nil && choice.Message.AssistantMessage.ToolCalls != nil {
			for _, toolCall := range *choice.Message.AssistantMessage.ToolCalls {
				// Parse arguments JSON string back to map
				var input map[string]interface{}
				if toolCall.Function.Arguments != "" {
					if err := json.Unmarshal([]byte(toolCall.Function.Arguments), &input); err != nil {
						input = map[string]interface{}{}
					}
				} else {
					input = map[string]interface{}{}
				}

				tc := AnthropicContent{
					Type:  "tool_use",
					ID:    toolCall.ID,
					Name:  toolCall.Function.Name,
					Input: input,
				}
				content = append(content, tc)
			}
		}
	}

	if content == nil {
		content = []AnthropicContent{}
	}

	anthropicResp.Content = content
	return anthropicResp
}
