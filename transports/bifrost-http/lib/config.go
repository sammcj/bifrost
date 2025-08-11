// Package lib provides core functionality for the Bifrost HTTP service,
// including context propagation, header management, and integration with monitoring systems.
package lib

import (
	"github.com/maximhq/bifrost/core/schemas"
)

// HandlerStore provides access to runtime configuration values for handlers.
// This interface allows handlers to access only the configuration they need
// without depending on the entire ConfigStore, improving testability and decoupling.
type HandlerStore interface {
	// ShouldAllowDirectKeys returns whether direct API keys in headers are allowed
	ShouldAllowDirectKeys() bool
}

// ClientConfig represents the core configuration for Bifrost HTTP transport and the Bifrost Client.
// It includes settings for excess request handling, Prometheus metrics, and initial pool size.
type ClientConfig struct {
	DropExcessRequests      bool     `json:"drop_excess_requests"`      // Drop excess requests if the provider queue is full
	InitialPoolSize         int      `json:"initial_pool_size"`         // The initial pool size for the bifrost client
	PrometheusLabels        []string `json:"prometheus_labels"`         // The labels to be used for prometheus metrics
	EnableLogging           bool     `json:"enable_logging"`            // Enable logging of requests and responses
	EnableGovernance        bool     `json:"enable_governance"`         // Enable governance on all requests
	EnforceGovernanceHeader bool     `json:"enforce_governance_header"` // Enforce governance on all requests
	AllowDirectKeys         bool     `json:"allow_direct_keys"`         // Allow direct keys to be used for requests
	EnableCaching           bool     `json:"enable_caching"`            // Enable Redis caching plugin
	AllowedOrigins          []string `json:"allowed_origins,omitempty"` // Additional allowed origins for CORS and WebSocket (localhost is always allowed)
}

// ProviderConfig represents the configuration for a specific AI model provider.
// It includes API keys, network settings, and concurrency settings.
type ProviderConfig struct {
	Keys                     []schemas.Key                     `json:"keys"`                                  // API keys for the provider with UUIDs
	NetworkConfig            *schemas.NetworkConfig            `json:"network_config,omitempty"`              // Network-related settings
	ConcurrencyAndBufferSize *schemas.ConcurrencyAndBufferSize `json:"concurrency_and_buffer_size,omitempty"` // Concurrency settings
	ProxyConfig              *schemas.ProxyConfig              `json:"proxy_config,omitempty"`                // Proxy configuration
	SendBackRawResponse      bool                              `json:"send_back_raw_response"`                // Include raw response in BifrostResponse
}

// ConfigMap maps provider names to their configurations.
type ConfigMap map[schemas.ModelProvider]ProviderConfig

// BifrostHTTPConfig represents the complete configuration structure for Bifrost HTTP transport.
// It includes both provider configurations and MCP configuration.
type BifrostHTTPConfig struct {
	ClientConfig   *ClientConfig      `json:"client"`    // Client configuration
	ProviderConfig ConfigMap          `json:"providers"` // Provider configurations
	MCPConfig      *schemas.MCPConfig `json:"mcp"`       // MCP configuration (optional)
}
