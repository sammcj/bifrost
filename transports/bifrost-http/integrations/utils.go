package integrations

import (
	"encoding/json"

	"github.com/fasthttp/router"
	bifrost "github.com/maximhq/bifrost/core"
	"github.com/maximhq/bifrost/core/schemas"
	"github.com/maximhq/bifrost/transports/bifrost-http/lib"
	"github.com/valyala/fasthttp"
)

// ExtensionRouter defines the interface that all integration routers must implement
// to register their routes with the main HTTP router.
type ExtensionRouter interface {
	RegisterRoutes(r *router.Router)
}

// RequestConverter is a function that converts integration-specific requests to Bifrost format.
// It takes the parsed request object and returns a BifrostRequest ready for processing.
type RequestConverter func(req interface{}) *schemas.BifrostRequest

// ResponseConverter is a function that converts Bifrost responses to integration-specific format.
// It takes a BifrostResponse and returns the format expected by the specific integration.
type ResponseConverter func(*schemas.BifrostResponse) interface{}

// PreRequestCallback is called before processing the request.
// It can be used to modify the request object (e.g., extract model from URL parameters)
// or perform validation. If it returns an error, the request processing stops.
type PreRequestCallback func(ctx *fasthttp.RequestCtx, req interface{}) error

// PostRequestCallback is called after processing the request but before sending the response.
// It can be used to modify the response or perform additional logging/metrics.
// If it returns an error, an error response is sent instead of the success response.
type PostRequestCallback func(ctx *fasthttp.RequestCtx, req interface{}, resp *schemas.BifrostResponse) error

// RouteConfig defines configuration for a single HTTP route in an integration.
// Each route specifies how to handle requests for a specific endpoint.
type RouteConfig struct {
	Path             string              // HTTP path pattern (e.g., "/openai/v1/chat/completions")
	Method           string              // HTTP method (POST, GET, PUT, DELETE)
	RequestType      interface{}         // Factory function to create request instance
	RequestConverter RequestConverter    // Function to convert request to BifrostRequest
	ResponseFunc     ResponseConverter   // Function to convert BifrostResponse to integration format
	PreCallback      PreRequestCallback  // Optional: called before request processing
	PostCallback     PostRequestCallback // Optional: called after request processing
}

// GenericRouter provides a reusable router implementation for all integrations.
// It handles the common flow of: parse request → convert to Bifrost → execute → convert response.
// Integration-specific logic is handled through the RouteConfig callbacks and converters.
type GenericRouter struct {
	client *bifrost.Bifrost // Bifrost client for executing requests
	routes []RouteConfig    // List of route configurations
}

// NewGenericRouter creates a new generic router with the given bifrost client and route configurations.
// Each integration should create their own routes and pass them to this constructor.
func NewGenericRouter(client *bifrost.Bifrost, routes []RouteConfig) *GenericRouter {
	return &GenericRouter{
		client: client,
		routes: routes,
	}
}

// RegisterRoutes registers all configured routes on the given fasthttp router.
// This method implements the ExtensionRouter interface.
func (g *GenericRouter) RegisterRoutes(r *router.Router) {
	for _, route := range g.routes {
		handler := g.createHandler(route)
		switch route.Method {
		case "POST":
			r.POST(route.Path, handler)
		case "GET":
			r.GET(route.Path, handler)
		case "PUT":
			r.PUT(route.Path, handler)
		case "DELETE":
			r.DELETE(route.Path, handler)
		default:
			r.POST(route.Path, handler) // Default to POST
		}
	}
}

// createHandler creates a fasthttp handler for the given route configuration.
// The handler follows this flow:
// 1. Parse JSON request body into the configured request type (for methods that expect bodies)
// 2. Execute pre-callback (if configured) for request modification/validation
// 3. Convert request to BifrostRequest using the configured converter
// 4. Execute the request through Bifrost
// 5. Execute post-callback (if configured) for response modification
// 6. Convert and send the response using the configured response converter
func (g *GenericRouter) createHandler(config RouteConfig) fasthttp.RequestHandler {
	return func(ctx *fasthttp.RequestCtx) {
		// Parse request body into the integration-specific request type
		// Skip JSON unmarshalling for methods that typically don't have request bodies
		req := config.RequestType
		method := string(ctx.Method())

		if method != "GET" && method != "DELETE" {
			// Use ctx.Request.Body() instead of ctx.PostBody() to support all HTTP methods
			body := ctx.Request.Body()
			if len(body) > 0 {
				if err := json.Unmarshal(body, req); err != nil {
					g.sendError(ctx, fasthttp.StatusBadRequest, "Invalid JSON: "+err.Error())
					return
				}
			}
		}

		// Execute pre-request callback if configured
		// This is typically used for extracting data from URL parameters
		// or performing request-specific validation
		if config.PreCallback != nil {
			if err := config.PreCallback(ctx, req); err != nil {
				g.sendError(ctx, fasthttp.StatusBadRequest, err.Error())
				return
			}
		}

		// Convert the integration-specific request to Bifrost format
		bifrostReq := config.RequestConverter(req)
		if bifrostReq == nil {
			g.sendError(ctx, fasthttp.StatusBadRequest, "Invalid request")
			return
		}
		if bifrostReq.Model == "" {
			g.sendError(ctx, fasthttp.StatusBadRequest, "Model parameter is required")
			return
		}

		// Execute the request through Bifrost
		bifrostCtx := lib.ConvertToBifrostContext(ctx)
		result, err := g.client.ChatCompletionRequest(*bifrostCtx, bifrostReq)
		if err != nil {
			g.sendError(ctx, func() int {
				if err.StatusCode != nil {
					return *err.StatusCode
				}
				return fasthttp.StatusInternalServerError
			}(), err)
			return
		}
		// Execute post-request callback if configured
		// This is typically used for response modification or additional processing
		if config.PostCallback != nil {
			if err := config.PostCallback(ctx, req, result); err != nil {
				g.sendError(ctx, fasthttp.StatusInternalServerError, err.Error())
				return
			}
		}

		// Convert Bifrost response to integration-specific format and send
		response := config.ResponseFunc(result)
		g.sendSuccess(ctx, response)
	}
}

// sendError sends an error response with the appropriate status code and JSON body.
// It handles different error types (string, error interface, or arbitrary objects).
func (g *GenericRouter) sendError(ctx *fasthttp.RequestCtx, statusCode int, err interface{}) {
	ctx.SetStatusCode(statusCode)
	ctx.SetContentType("application/json")

	var errorBody []byte
	switch e := err.(type) {
	case string:
		errorBody, _ = json.Marshal(map[string]string{"error": e})
	case error:
		errorBody, _ = json.Marshal(map[string]string{"error": e.Error()})
	default:
		errorBody, _ = json.Marshal(err)
	}
	ctx.SetBody(errorBody)
}

// sendSuccess sends a successful response with HTTP 200 status and JSON body.
func (g *GenericRouter) sendSuccess(ctx *fasthttp.RequestCtx, response interface{}) {
	ctx.SetStatusCode(fasthttp.StatusOK)
	ctx.SetContentType("application/json")

	responseBody, err := json.Marshal(response)
	if err != nil {
		g.sendError(ctx, fasthttp.StatusInternalServerError, "failed to encode response: "+err.Error())
		return
	}

	ctx.SetBody(responseBody)
}
