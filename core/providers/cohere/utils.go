package cohere

import "github.com/maximhq/bifrost/core/schemas"

var (
	// Maps provider-specific finish reasons to Bifrost format
	cohereFinishReasonToBifrost = map[CohereFinishReason]string{
		FinishReasonComplete:     "stop",
		FinishReasonStopSequence: "stop",
		FinishReasonMaxTokens:    "length",
		FinishReasonToolCall:     "tool_calls",
	}
)

// ConvertCohereFinishReasonToBifrost converts provider finish reasons to Bifrost format
func ConvertCohereFinishReasonToBifrost(providerReason CohereFinishReason) string {
	if bifrostReason, ok := cohereFinishReasonToBifrost[providerReason]; ok {
		return bifrostReason
	}
	return string(providerReason)
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
