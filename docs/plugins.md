# Bifrost Plugin System

Bifrost provides a powerful plugin system that allows you to extend and customize the request/response pipeline. Plugins can be used to implement various functionalities like rate limiting, caching, logging, monitoring, and more.

## 1. How Plugins Work

Plugins in Bifrost follow a flexible interface that allows them to intercept and modify requests and responses at different stages of processing:

1. **PreHook**: Executed before a request is sent to a provider

   - Can modify the request
   - Can add custom headers or parameters
   - Can implement rate limiting or validation
   - Executed in the order they are registered
   - If a PreHook returns a response, the provider call is skipped and only the PostHook methods of plugins that had their PreHook executed are called in reverse order.

2. **PostHook**: Executed after receiving a response from a provider or a PreHook short-circuit
   - Can modify the response and/or error
   - Can recover from errors (set error to nil and provide a response)
   - Can invalidate a response (set response to nil and provide an error)
   - Both response and error may be nil; plugins must handle both cases
   - Executed in reverse order of PreHooks
   - Only truly empty errors (no message, no error, no status code, no type) are treated as recoveries by the pipeline

> **Note**: The plugin pipeline ensures symmetry: for every PreHook executed, the corresponding PostHook will be called in reverse order. Plugin authors should ensure their hooks are robust to both response and error being nil, and should not assume either is always present.

## 2. Plugin Interface

```go
// Plugin interface for Bifrost plugins
// See core/schemas/plugin.go for the authoritative definition

GetName() string
PreHook(ctx *context.Context, req *BifrostRequest) (*BifrostRequest, *BifrostResponse, error)
PostHook(ctx *context.Context, result *BifrostResponse, err *BifrostError) (*BifrostResponse, *BifrostError, error)
Cleanup() error
```

## 3. Building Custom Plugins

### Basic Plugin Structure

```go
// Example plugin skeleton

type CustomPlugin struct {
    // Your plugin fields
}

func (p *CustomPlugin) GetName() string {
    return "CustomPlugin"
}

func (p *CustomPlugin) PreHook(ctx *context.Context, req *BifrostRequest) (*BifrostRequest, *BifrostResponse, error) {
    // Modify request or add custom logic
    // Return nil for response to continue with provider call
    return req, nil, nil
}

func (p *CustomPlugin) PostHook(ctx *context.Context, result *BifrostResponse, err *BifrostError) (*BifrostResponse, *BifrostError, error) {
    // Modify response or error, or recover from error
    // Either result or err may be nil
    // To recover from an error, set err to nil and return a valid result
    // To invalidate a response, set the result to nil and return a non-nil err.
    return result, err, nil
}

func (p *CustomPlugin) Cleanup() error {
    // Clean up any resources
    return nil
}
```

### Example: Rate Limiting Plugin

```go
type RateLimitPlugin struct {
    limiter *rate.Limiter
}

func (p *RateLimitPlugin) GetName() string {
    return "RateLimitPlugin"
}

func NewRateLimitPlugin(rps float64) *RateLimitPlugin {
    return &RateLimitPlugin{
        limiter: rate.NewLimiter(rate.Limit(rps), 1),
    }
}

func (p *RateLimitPlugin) PreHook(ctx *context.Context, req *BifrostRequest) (*BifrostRequest, *BifrostResponse, error) {
    if err := p.limiter.Wait(*ctx); err != nil {
        return nil, nil, err
    }
    return req, nil, nil
}

func (p *RateLimitPlugin) PostHook(ctx *context.Context, result *BifrostResponse, err *BifrostError) (*BifrostResponse, *BifrostError, error) {
    // No-op for rate limiting
    return result, err, nil
}

func (p *RateLimitPlugin) Cleanup() error {
    // Rate limiter doesn't need cleanup
    return nil
}
```

### Example: Logging Plugin

```go
type LoggingPlugin struct {
    logger schemas.Logger
}

func (p *LoggingPlugin) GetName() string {
    return "LoggingPlugin"
}

func NewLoggingPlugin(logger schemas.Logger) *LoggingPlugin {
    return &LoggingPlugin{logger: logger}
}

func (p *LoggingPlugin) PreHook(ctx *context.Context, req *BifrostRequest) (*BifrostRequest, *BifrostResponse, error) {
    p.logger.Info(fmt.Sprintf("Request to %s with model %s", req.Provider, req.Model))
    return req, nil, nil
}

func (p *LoggingPlugin) PostHook(ctx *context.Context, result *BifrostResponse, err *BifrostError) (*BifrostResponse, *BifrostError, error) {
    if result != nil {
        p.logger.Info(fmt.Sprintf("Response from %s with %d tokens", result.Model, result.Usage.TotalTokens))
    }
    if err != nil {
        p.logger.Warn(fmt.Sprintf("Error: %v", err.Error.Message))
    }
    return result, err, nil
}

func (p *LoggingPlugin) Cleanup() error {
    // Logger doesn't need cleanup
    return nil
}
```

## 4. Using Plugins

### Initializing Bifrost with Plugins

```go
client, err := bifrost.Init(schemas.BifrostConfig{
    Account: &yourAccount,
    Plugins: []schemas.Plugin{
        NewRateLimitPlugin(10.0),    // 10 requests per second
        NewLoggingPlugin(logger),    // Custom logging
        // Add more plugins as needed
    },
})
```

## 5. Plugin Pipeline Symmetry and Order

- PreHooks are executed in the order they are registered.
- PostHooks are executed in the reverse order of PreHooks.
- If a PreHook returns a response, the provider call is skipped and only the PostHook methods of plugins that had their PreHook executed are called in reverse order.
- The plugin pipeline ensures that for every PreHook executed, the corresponding PostHook will be called.

## 6. Best Practices for Plugin Authors

- Always check for both response and error being nil in PostHook.
- Do not assume either is always present.
- To recover from an error, set err to nil and return a valid result. Only truly empty errors (no message, no error, no status code, no type) are treated as recoveries by the pipeline.
- To invalidate a response, set the result to nil and return a non-nil err.
- Keep plugin logic lightweight and avoid blocking operations in hooks.
- Use context for cancellation and state passing.
- Write unit tests for your plugins, including error recovery and response invalidation scenarios.

## 7. Available Plugins

Bifrost provides a **[Plugin Store](https://github.com/maximhq/bifrost/tree/main/plugins)** with one-line integrations. For each plugin, refer to its specific documentation for configuration and usage details:

`https://github.com/maximhq/bifrost/tree/main/plugins/{plugin_name}`

## 8. Plugin Development Guidelines

1. **Documentation**

   - Document your plugin's purpose and configuration
   - Provide usage examples
   - Include performance considerations

2. **Configuration**

   - Make plugins configurable
   - Use sensible defaults
   - Validate configuration

3. **Error Handling**

   - Use custom error types
   - Provide detailed error messages
   - Handle edge cases gracefully

4. **Contribution Process**

   - Open an issue first to discuss the plugin's use case and design
   - Create a pull request with a clear explanation of the plugin's purpose
   - Follow the plugin structure requirements below

5. **Plugin Structure**
   Each plugin should be organized as follows:

   ```
   plugins/
   └── your-plugin-name/
       ├── main.go           # Plugin implementation
       ├── plugin_test.go    # Plugin tests
       ├── README.md         # Documentation
       └── go.mod            # Module definition
   ```

   Example `main.go`:

   ```go
   package your_plugin_name

   import (
       "context"
       "github.com/maximhq/bifrost/core/schemas"
   )

   type YourPlugin struct {
       // Plugin fields
   }

   func (p *YourPlugin) GetName() string {
       return "YourPlugin"
   }

   func (p *YourPlugin) PreHook(ctx *context.Context, req *schemas.BifrostRequest) (*schemas.BifrostRequest, *schemas.BifrostResponse, error) {
       // Implementation
       return req, nil, nil
   }

   func (p *YourPlugin) PostHook(ctx *context.Context, result *schemas.BifrostResponse, err *schemas.BifrostError) (*schemas.BifrostResponse, *schemas.BifrostError, error) {
       // Implementation
       return result, err, nil
   }

   func (p *YourPlugin) Cleanup() error {
       // Clean up any resources
       return nil
   }
   ```

   Example `README.md`:

   ````markdown
   # Your Plugin Name

   Brief description of what the plugin does.

   ## Installation

   ```bash
   go get github.com/maximhq/bifrost/plugins/your-plugin-name
   ```

   ## Usage

   Explain plugin usage.

   ## Configuration

   Describe configuration options.

   ## Examples

   Show usage examples.
   ````
