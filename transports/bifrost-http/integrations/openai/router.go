package openai

import (
	bifrost "github.com/maximhq/bifrost/core"
	"github.com/maximhq/bifrost/core/schemas"
	"github.com/maximhq/bifrost/transports/bifrost-http/integrations"
)

// OpenAIRouter holds route registrations for OpenAI endpoints.
// It supports standard chat completions and image-enabled vision capabilities.
type OpenAIRouter struct {
	*integrations.GenericRouter
}

// NewOpenAIRouter creates a new OpenAIRouter with the given bifrost client.
func NewOpenAIRouter(client *bifrost.Bifrost) *OpenAIRouter {
	routes := []integrations.RouteConfig{
		{
			Path:        "/openai/v1/chat/completions",
			Method:      "POST",
			RequestType: &OpenAIChatRequest{},
			RequestConverter: func(req interface{}) *schemas.BifrostRequest {
				if openaiReq, ok := req.(*OpenAIChatRequest); ok {
					return openaiReq.ConvertToBifrostRequest()
				}
				return nil
			},
			ResponseFunc: func(resp *schemas.BifrostResponse) interface{} {
				return DeriveOpenAIFromBifrostResponse(resp)
			},
		},
	}

	return &OpenAIRouter{
		GenericRouter: integrations.NewGenericRouter(client, routes),
	}
}
