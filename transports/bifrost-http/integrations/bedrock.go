package integrations

import (
	"context"
	"errors"
	"fmt"

	bifrost "github.com/maximhq/bifrost/core"
	"github.com/maximhq/bifrost/core/providers/bedrock"
	"github.com/maximhq/bifrost/core/schemas"
	"github.com/maximhq/bifrost/transports/bifrost-http/lib"
	"github.com/valyala/fasthttp"
)

// BedrockRouter handles AWS Bedrock-compatible API endpoints
type BedrockRouter struct {
	*GenericRouter
}

// createBedrockConverseRouteConfig creates a route configuration for the Bedrock Converse API endpoint
// Handles POST /bedrock/model/{modelId}/converse
func createBedrockConverseRouteConfig(pathPrefix string) RouteConfig {
	return RouteConfig{
		Type:   RouteConfigTypeBedrock,
		Path:   pathPrefix + "/model/{modelId}/converse",
		Method: "POST",
		GetRequestTypeInstance: func() interface{} {
			return &bedrock.BedrockConverseRequest{}
		},
		RequestConverter: func(req interface{}) (*schemas.BifrostRequest, error) {
			if bedrockReq, ok := req.(*bedrock.BedrockConverseRequest); ok {
				bifrostReq, err := bedrockReq.ToBifrostResponsesRequest()
				if err != nil {
					return nil, fmt.Errorf("failed to convert bedrock request: %w", err)
				}
				return &schemas.BifrostRequest{
					ResponsesRequest: bifrostReq,
				}, nil
			}
			return nil, errors.New("invalid request type")
		},
		ResponsesResponseConverter: func(resp *schemas.BifrostResponsesResponse) (interface{}, error) {
			return bedrock.ToBedrockConverseResponse(resp)
		},
		ErrorConverter: func(err *schemas.BifrostError) interface{} {
			return bedrock.ToBedrockError(err)
		},
		PreCallback: extractBedrockModelFromPath,
	}
}

// CreateBedrockRouteConfigs creates route configurations for Bedrock endpoints
func CreateBedrockRouteConfigs(pathPrefix string) []RouteConfig {
	return []RouteConfig{
		createBedrockConverseRouteConfig(pathPrefix),
	}
}

// NewBedrockRouter creates a new BedrockRouter with the given bifrost client
func NewBedrockRouter(client *bifrost.Bifrost, handlerStore lib.HandlerStore, logger schemas.Logger) *BedrockRouter {
	return &BedrockRouter{
		GenericRouter: NewGenericRouter(client, handlerStore, CreateBedrockRouteConfigs("/bedrock"), logger),
	}
}

// extractBedrockModelFromPath pre-callback extracts the modelId from the URL path
// and sets it in the Bedrock request
func extractBedrockModelFromPath(ctx *fasthttp.RequestCtx, bifrostCtx *context.Context, req interface{}) error {
	if bedrockReq, ok := req.(*bedrock.BedrockConverseRequest); ok {
		// Extract modelId from path parameter
		modelIDVal := ctx.UserValue("modelId")
		if modelIDVal == nil {
			return errors.New("modelId not found in path")
		}

		modelIDStr, ok := modelIDVal.(string)
		if !ok {
			return fmt.Errorf("modelId must be a string, got %T", modelIDVal)
		}
		if modelIDStr == "" {
			return errors.New("modelId cannot be empty")
		}

		// Set the model ID with bedrock/ prefix for provider routing
		bedrockReq.ModelID = "bedrock/" + modelIDStr

		return nil
	}
	return errors.New("invalid request type for bedrock model extraction")
}
