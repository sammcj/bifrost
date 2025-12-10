package configstore

import (
	"crypto/sha256"
	"encoding/hex"
	"sort"

	"github.com/bytedance/sonic"
	"github.com/maximhq/bifrost/core/schemas"
	"github.com/maximhq/bifrost/framework/configstore/tables"
)

type EnvKeyType string

const (
	EnvKeyTypeAPIKey        EnvKeyType = "api_key"
	EnvKeyTypeAzureConfig   EnvKeyType = "azure_config"
	EnvKeyTypeVertexConfig  EnvKeyType = "vertex_config"
	EnvKeyTypeBedrockConfig EnvKeyType = "bedrock_config"
	EnvKeyTypeConnection    EnvKeyType = "connection_string"
	EnvKeyTypeMCPHeader     EnvKeyType = "mcp_header"
)

// EnvKeyInfo stores information about a key sourced from environment
type EnvKeyInfo struct {
	EnvVar     string                // The environment variable name (without env. prefix)
	Provider   schemas.ModelProvider // The provider this key belongs to (empty for core/mcp configs)
	KeyType    EnvKeyType            // Type of key (e.g., "api_key", "azure_config", "vertex_config", "bedrock_config", "connection_string", "mcp_header")
	ConfigPath string                // Path in config where this env var is used
	KeyID      string                // The key ID this env var belongs to (empty for non-key configs like bedrock_config, connection_string)
}

// ClientConfig represents the core configuration for Bifrost HTTP transport and the Bifrost Client.
// It includes settings for excess request handling, Prometheus metrics, and initial pool size.
type ClientConfig struct {
	DropExcessRequests      bool     `json:"drop_excess_requests"`                // Drop excess requests if the provider queue is full
	InitialPoolSize         int      `json:"initial_pool_size"`                   // The initial pool size for the bifrost client
	PrometheusLabels        []string `json:"prometheus_labels"`                   // The labels to be used for prometheus metrics
	EnableLogging           bool     `json:"enable_logging"`                      // Enable logging of requests and responses
	DisableContentLogging   bool     `json:"disable_content_logging"`             // Disable logging of content
	LogRetentionDays        int      `json:"log_retention_days" validate:"min=1"` // Number of days to retain logs (minimum 1 day)
	EnableGovernance        bool     `json:"enable_governance"`                   // Enable governance on all requests
	EnforceGovernanceHeader bool     `json:"enforce_governance_header"`           // Enforce governance on all requests
	AllowDirectKeys         bool     `json:"allow_direct_keys"`                   // Allow direct keys to be used for requests
	AllowedOrigins          []string `json:"allowed_origins,omitempty"`           // Additional allowed origins for CORS and WebSocket (localhost is always allowed)
	MaxRequestBodySizeMB    int      `json:"max_request_body_size_mb"`            // The maximum request body size in MB
	EnableLiteLLMFallbacks  bool     `json:"enable_litellm_fallbacks"`            // Enable litellm-specific fallbacks for text completion for Groq
}

// ProviderConfig represents the configuration for a specific AI model provider.
// It includes API keys, network settings, and concurrency settings.
type ProviderConfig struct {
	Keys                     []schemas.Key                     `json:"keys"`                                  // API keys for the provider with UUIDs
	NetworkConfig            *schemas.NetworkConfig            `json:"network_config,omitempty"`              // Network-related settings
	ConcurrencyAndBufferSize *schemas.ConcurrencyAndBufferSize `json:"concurrency_and_buffer_size,omitempty"` // Concurrency settings
	ProxyConfig              *schemas.ProxyConfig              `json:"proxy_config,omitempty"`                // Proxy configuration
	SendBackRawResponse      bool                              `json:"send_back_raw_response"`                // Include raw response in BifrostResponse
	CustomProviderConfig     *schemas.CustomProviderConfig     `json:"custom_provider_config,omitempty"`      // Custom provider configuration
	ConfigHash               string                            `json:"-"`
}

// GenerateConfigHash generates a SHA256 hash of the provider configuration.
// This is used to detect changes between config.json and database config.
// Keys are excluded as they are hashed separately.
func (p *ProviderConfig) GenerateConfigHash(providerName string) (string, error) {
	hash := sha256.New()

	// Hash provider name
	hash.Write([]byte(providerName))

	// Hash NetworkConfig
	if p.NetworkConfig != nil {
		data, err := sonic.Marshal(p.NetworkConfig)
		if err != nil {
			return "", err
		}
		hash.Write(data)
	}

	// Hash ConcurrencyAndBufferSize
	if p.ConcurrencyAndBufferSize != nil {
		data, err := sonic.Marshal(p.ConcurrencyAndBufferSize)
		if err != nil {
			return "", err
		}
		hash.Write(data)
	}

	// Hash ProxyConfig
	if p.ProxyConfig != nil {
		data, err := sonic.Marshal(p.ProxyConfig)
		if err != nil {
			return "", err
		}
		hash.Write(data)
	}

	// Hash CustomProviderConfig
	if p.CustomProviderConfig != nil {
		data, err := sonic.Marshal(p.CustomProviderConfig)
		if err != nil {
			return "", err
		}
		hash.Write(data)
	}

	// Hash SendBackRawResponse
	if p.SendBackRawResponse {
		hash.Write([]byte("sendBackRawResponse"))
	}

	return hex.EncodeToString(hash.Sum(nil)), nil
}

// GenerateKeyHash generates a SHA256 hash for an individual key.
// This is used to detect changes to keys between config.json and database.
// Skips: ID (dynamic UUID), timestamps
func GenerateKeyHash(key schemas.Key) (string, error) {
	hash := sha256.New()

	// Hash Name
	hash.Write([]byte(key.Name))

	// Hash Value
	hash.Write([]byte(key.Value))

	// Hash Models (key-level model restrictions)
	if len(key.Models) > 0 {
		sortedModels := make([]string, len(key.Models))
		copy(sortedModels, key.Models)
		sort.Strings(sortedModels)
		data, err := sonic.Marshal(sortedModels)
		if err != nil {
			return "", err
		}
		hash.Write(data)
	}

	// Hash Weight
	data, err := sonic.Marshal(key.Weight)
	if err != nil {
		return "", err
	}
	hash.Write(data)

	// Hash AzureKeyConfig
	if key.AzureKeyConfig != nil {
		data, err := sonic.Marshal(key.AzureKeyConfig)
		if err != nil {
			return "", err
		}
		hash.Write(data)
	}

	// Hash VertexKeyConfig
	if key.VertexKeyConfig != nil {
		data, err := sonic.Marshal(key.VertexKeyConfig)
		if err != nil {
			return "", err
		}
		hash.Write(data)
	}

	// Hash BedrockKeyConfig
	if key.BedrockKeyConfig != nil {
		data, err := sonic.Marshal(key.BedrockKeyConfig)
		if err != nil {
			return "", err
		}
		hash.Write(data)
	}

	return hex.EncodeToString(hash.Sum(nil)), nil
}

// VirtualKeyHashInput represents the fields used for virtual key hash generation.
// This struct is used to create a consistent hash from TableVirtualKey,
// excluding dynamic fields like ID, timestamps, and relationship objects.
type VirtualKeyHashInput struct {
	Name        string
	Description string
	Value       string
	IsActive    bool
	TeamID      *string
	CustomerID  *string
	BudgetID    *string
	RateLimitID *string
	// ProviderConfigs and MCPConfigs are hashed separately as they contain nested data
	ProviderConfigs []VirtualKeyProviderConfigHashInput
	MCPConfigs      []VirtualKeyMCPConfigHashInput
}

// VirtualKeyProviderConfigHashInput represents provider config fields for hashing
type VirtualKeyProviderConfigHashInput struct {
	Provider      string
	Weight        float64
	AllowedModels []string
	BudgetID      *string
	RateLimitID   *string
	KeyIDs        []string // Only key IDs, not full key objects
}

// VirtualKeyMCPConfigHashInput represents MCP config fields for hashing
type VirtualKeyMCPConfigHashInput struct {
	MCPClientID    uint
	ToolsToExecute []string
}

// GenerateVirtualKeyHash generates a SHA256 hash for a virtual key.
// This is used to detect changes to virtual keys between config.json and database.
// Skips: ID (primary key), CreatedAt, UpdatedAt, and relationship objects (Team, Customer, Budget, RateLimit)
func GenerateVirtualKeyHash(vk tables.TableVirtualKey) (string, error) {
	hash := sha256.New()

	// Hash Name
	hash.Write([]byte(vk.Name))

	// Hash Description
	hash.Write([]byte(vk.Description))

	// Hash Value
	hash.Write([]byte(vk.Value))

	// Hash IsActive
	if vk.IsActive {
		hash.Write([]byte("isActive:true"))
	} else {
		hash.Write([]byte("isActive:false"))
	}

	// Hash TeamID
	if vk.TeamID != nil {
		hash.Write([]byte("teamID:" + *vk.TeamID))
	}

	// Hash CustomerID
	if vk.CustomerID != nil {
		hash.Write([]byte("customerID:" + *vk.CustomerID))
	}

	// Hash BudgetID
	if vk.BudgetID != nil {
		hash.Write([]byte("budgetID:" + *vk.BudgetID))
	}

	// Hash RateLimitID
	if vk.RateLimitID != nil {
		hash.Write([]byte("rateLimitID:" + *vk.RateLimitID))
	}

	// Hash ProviderConfigs
	if len(vk.ProviderConfigs) > 0 {
		// Copy and sort provider configs for deterministic hashing
		sortedProviderConfigs := make([]tables.TableVirtualKeyProviderConfig, len(vk.ProviderConfigs))
		copy(sortedProviderConfigs, vk.ProviderConfigs)
		sort.Slice(sortedProviderConfigs, func(i, j int) bool {
			if sortedProviderConfigs[i].Provider != sortedProviderConfigs[j].Provider {
				return sortedProviderConfigs[i].Provider < sortedProviderConfigs[j].Provider
			}
			bi, bj := "", ""
			if sortedProviderConfigs[i].BudgetID != nil {
				bi = *sortedProviderConfigs[i].BudgetID
			}
			if sortedProviderConfigs[j].BudgetID != nil {
				bj = *sortedProviderConfigs[j].BudgetID
			}
			if bi != bj {
				return bi < bj
			}
			ri, rj := "", ""
			if sortedProviderConfigs[i].RateLimitID != nil {
				ri = *sortedProviderConfigs[i].RateLimitID
			}
			if sortedProviderConfigs[j].RateLimitID != nil {
				rj = *sortedProviderConfigs[j].RateLimitID
			}
			if ri != rj {
				return ri < rj
			}
			return sortedProviderConfigs[i].Weight < sortedProviderConfigs[j].Weight
		})

		providerConfigsForHash := make([]VirtualKeyProviderConfigHashInput, len(sortedProviderConfigs))
		for i, pc := range sortedProviderConfigs {
			// Sort key IDs for deterministic hashing
			keyIDs := make([]string, len(pc.Keys))
			for j, k := range pc.Keys {
				keyIDs[j] = k.KeyID
			}
			sort.Strings(keyIDs)

			// Sort allowed models for deterministic hashing
			sortedAllowedModels := make([]string, len(pc.AllowedModels))
			copy(sortedAllowedModels, pc.AllowedModels)
			sort.Strings(sortedAllowedModels)

			providerConfigsForHash[i] = VirtualKeyProviderConfigHashInput{
				Provider:      pc.Provider,
				Weight:        pc.Weight,
				AllowedModels: sortedAllowedModels,
				BudgetID:      pc.BudgetID,
				RateLimitID:   pc.RateLimitID,
				KeyIDs:        keyIDs,
			}
		}
		data, err := sonic.Marshal(providerConfigsForHash)
		if err != nil {
			return "", err
		}
		hash.Write(data)
	}

	// Hash MCPConfigs
	if len(vk.MCPConfigs) > 0 {
		// Copy and sort MCP configs for deterministic hashing
		sortedMCPConfigs := make([]tables.TableVirtualKeyMCPConfig, len(vk.MCPConfigs))
		copy(sortedMCPConfigs, vk.MCPConfigs)
		sort.Slice(sortedMCPConfigs, func(i, j int) bool {
			return sortedMCPConfigs[i].MCPClientID < sortedMCPConfigs[j].MCPClientID
		})

		mcpConfigsForHash := make([]VirtualKeyMCPConfigHashInput, len(sortedMCPConfigs))
		for i, mc := range sortedMCPConfigs {
			// Sort tools for deterministic hashing
			sortedTools := make([]string, len(mc.ToolsToExecute))
			copy(sortedTools, mc.ToolsToExecute)
			sort.Strings(sortedTools)

			mcpConfigsForHash[i] = VirtualKeyMCPConfigHashInput{
				MCPClientID:    mc.MCPClientID,
				ToolsToExecute: sortedTools,
			}
		}
		data, err := sonic.Marshal(mcpConfigsForHash)
		if err != nil {
			return "", err
		}
		hash.Write(data)
	}

	return hex.EncodeToString(hash.Sum(nil)), nil
}

// AuthConfig represents configured auth config for Bifrost dashboard
type AuthConfig struct {
	AdminUserName          string `json:"admin_username"`
	AdminPassword          string `json:"admin_password"`
	IsEnabled              bool   `json:"is_enabled"`
	DisableAuthOnInference bool   `json:"disable_auth_on_inference"`
}

// ConfigMap maps provider names to their configurations.
type ConfigMap map[schemas.ModelProvider]ProviderConfig

type GovernanceConfig struct {
	VirtualKeys []tables.TableVirtualKey `json:"virtual_keys"`
	Teams       []tables.TableTeam       `json:"teams"`
	Customers   []tables.TableCustomer   `json:"customers"`
	Budgets     []tables.TableBudget     `json:"budgets"`
	RateLimits  []tables.TableRateLimit  `json:"rate_limits"`
	AuthConfig  *AuthConfig              `json:"auth_config,omitempty"`
}
