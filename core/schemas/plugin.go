// Package schemas defines the core schemas and types used by the Bifrost system.
package schemas

import "context"

// Plugin defines the interface for Bifrost plugins.
// Plugins can intercept and modify requests and responses at different stages
// of the processing pipeline.
// User can provide multiple plugins in the BifrostConfig.
// PreHooks are executed in the order they are registered.
// PostHooks are executed in the reverse order of PreHooks.

// PreHooks and PostHooks can be used to implement custom logic, such as:
// - Rate limiting
// - Caching
// - Logging
// - Monitoring

// No Plugin errors are returned to the caller, they are logged as warnings by the Bifrost instance.

type Plugin interface {
	// GetName returns the name of the plugin.
	GetName() string

	// PreHook is called before a request is processed by a provider.
	// It allows plugins to modify the request before it is sent to the provider.
	// The context parameter can be used to maintain state across plugin calls.
	// Returns the modified request, an optional response (if the plugin wants to short-circuit the provider call), and any error that occurred during processing.
	// If a response is returned, the provider call is skipped and only the PostHook methods of plugins that had their PreHook executed are called in reverse order.
	PreHook(ctx *context.Context, req *BifrostRequest) (*BifrostRequest, *BifrostResponse, error)

	// PostHook is called after a response is received from a provider.
	// It allows plugins to modify the response/error before it is returned to the caller.
	// Returns the modified response, bifrost error and any error that occurred during processing.
	PostHook(ctx *context.Context, result *BifrostResponse, err *BifrostError) (*BifrostResponse, *BifrostError, error)

	// Cleanup is called on bifrost shutdown.
	// It allows plugins to clean up any resources they have allocated.
	// Returns any error that occurred during cleanup, which will be logged as a warning by the Bifrost instance.
	Cleanup() error
}
