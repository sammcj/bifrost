package anthropic

import (
	bifrost "github.com/maximhq/bifrost/core"
	"github.com/maximhq/bifrost/core/schemas"
	"github.com/maximhq/bifrost/transports/bifrost-http/integrations"
)

// AnthropicRouter holds route registrations for Anthropic endpoints.
// It supports standard chat completions and image-enabled vision capabilities.
type AnthropicRouter struct {
	*integrations.GenericRouter
}

// NewAnthropicRouter creates a new AnthropicRouter with the given bifrost client.
func NewAnthropicRouter(client *bifrost.Bifrost) *AnthropicRouter {
	routes := []integrations.RouteConfig{
		{
			Path:   "/anthropic/v1/messages",
			Method: "POST",
			GetRequestTypeInstance: func() interface{} {
				return &AnthropicMessageRequest{}
			},
			RequestConverter: func(req interface{}) *schemas.BifrostRequest {
				if anthropicReq, ok := req.(*AnthropicMessageRequest); ok {
					return anthropicReq.ConvertToBifrostRequest()
				}
				return nil
			},
			ResponseFunc: func(resp *schemas.BifrostResponse) interface{} {
				return DeriveAnthropicFromBifrostResponse(resp)
			},
		},
	}

	return &AnthropicRouter{
		GenericRouter: integrations.NewGenericRouter(client, routes),
	}
}
