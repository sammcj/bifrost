package tests

import (
	"bifrost/interfaces"
	"bifrost/providers"
	"fmt"
)

// BaseAccount provides a basic implementation of the Account interface
type BaseAccount struct {
	providers map[string]interfaces.Provider
	keys      map[string][]interfaces.Key
	config    map[string]interfaces.ConcurrencyAndBufferSize
}

type ProviderConfig map[string]struct {
	Keys              []interfaces.Key                    `json:"keys"`
	ConcurrencyConfig interfaces.ConcurrencyAndBufferSize `json:"concurrency_config"`
}

func (baseAccount *BaseAccount) Init(config ProviderConfig) error {
	baseAccount.providers = make(map[string]interfaces.Provider)
	baseAccount.keys = make(map[string][]interfaces.Key)
	baseAccount.config = make(map[string]interfaces.ConcurrencyAndBufferSize)

	for providerKey, providerData := range config {
		// Create provider instance based on the key
		provider, err := baseAccount.createProvider(providerKey, providerData.Keys)
		if err != nil {
			return fmt.Errorf("failed to create provider %s: %v", providerKey, err)
		}

		fmt.Println("✅ provider created")

		// Add provider to the account
		baseAccount.AddProvider(provider)

		// Add keys for the provider
		for _, keyData := range providerData.Keys {
			key := interfaces.Key{
				Value:  keyData.Value,
				Models: keyData.Models,
				Weight: keyData.Weight,
			}

			baseAccount.AddKey(providerKey, key)
		}

		// Set provider configuration
		baseAccount.SetProviderConcurrencyConfig(providerKey, providerData.ConcurrencyConfig)
	}

	return nil
}

// createProvider creates a new provider instance based on the provider key
func (ba *BaseAccount) createProvider(providerKey string, keys []interfaces.Key) (interfaces.Provider, error) {
	if len(keys) == 0 {
		return nil, fmt.Errorf("no keys found for provider: %s", providerKey)
	}

	switch interfaces.SupportedModelProvider(providerKey) {
	case interfaces.OpenAI:
		return providers.NewOpenAIProvider(), nil
	case interfaces.Anthropic:
		return providers.NewAnthropicProvider(), nil
	default:
		return nil, fmt.Errorf("unsupported provider: %s", providerKey)
	}
}

// GetInitiallyConfiguredProviders returns all configured providers
func (ba *BaseAccount) GetInitiallyConfiguredProviders() ([]interfaces.Provider, error) {
	providers := make([]interfaces.Provider, 0, len(ba.providers))
	for _, provider := range ba.providers {
		providers = append(providers, provider)
	}

	return providers, nil
}

// GetKeysForProvider returns all keys associated with a provider
func (ba *BaseAccount) GetKeysForProvider(provider interfaces.Provider) ([]interfaces.Key, error) {
	providerKey := string(provider.GetProviderKey())
	if keys, exists := ba.keys[providerKey]; exists {
		return keys, nil
	}

	return nil, fmt.Errorf("no keys found for provider: %s", providerKey)
}

// GetConcurrencyAndBufferSizeForProvider returns the concurrency and buffer size settings for a provider
func (ba *BaseAccount) GetConcurrencyAndBufferSizeForProvider(provider interfaces.Provider) (interfaces.ConcurrencyAndBufferSize, error) {
	providerKey := string(provider.GetProviderKey())
	if config, exists := ba.config[providerKey]; exists {
		return config, nil
	}

	// Default values if not configured
	return interfaces.ConcurrencyAndBufferSize{
		Concurrency: 5,
		BufferSize:  100,
	}, nil
}

// AddProvider adds a new provider to the account
func (ba *BaseAccount) AddProvider(provider interfaces.Provider) {
	ba.providers[string(provider.GetProviderKey())] = provider
}

// AddKey adds a new key for a provider
func (ba *BaseAccount) AddKey(providerKey string, key interfaces.Key) {
	ba.keys[providerKey] = append(ba.keys[providerKey], key)
}

// SetProviderConfig sets the concurrency and buffer size for a provider
func (ba *BaseAccount) SetProviderConcurrencyConfig(providerKey string, config interfaces.ConcurrencyAndBufferSize) {
	ba.config[providerKey] = config
}
