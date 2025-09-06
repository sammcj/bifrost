// Package logging provides utility functions and interfaces for the GORM-based logging plugin
package logging

import (
	"fmt"

	"github.com/maximhq/bifrost/core/schemas"
	"github.com/maximhq/bifrost/framework/logstore"
)

func (p *LoggerPlugin) calculateRequestTotalCost(result *schemas.BifrostResponse, provider schemas.ModelProvider, model string, requestType schemas.RequestType) float64 {
	if result == nil {
		return 0
	}

	cacheDebug := result.ExtraFields.CacheDebug

	if cacheDebug != nil {
		if cacheDebug.CacheHit {
			if cacheDebug.HitType != nil && *cacheDebug.HitType == "direct" {
				return 0
			} else if cacheDebug.ProviderUsed != nil && cacheDebug.ModelUsed != nil && cacheDebug.InputTokens != nil {
				return p.getSemanticCacheCost(*cacheDebug.ProviderUsed, *cacheDebug.ModelUsed, *cacheDebug.InputTokens)
			}

			// Don't over-bill cache hits if fields are missing.
			return 0
		} else {
			baseCost := p.pricingManager.CalculateCost(result, provider, model, requestType)
			var semanticCacheCost float64
			if cacheDebug.ProviderUsed != nil && cacheDebug.ModelUsed != nil && cacheDebug.InputTokens != nil {
				semanticCacheCost = p.getSemanticCacheCost(*cacheDebug.ProviderUsed, *cacheDebug.ModelUsed, *cacheDebug.InputTokens)
			}

			return baseCost + semanticCacheCost
		}
	}

	return p.pricingManager.CalculateCost(result, provider, model, requestType)
}

func (p *LoggerPlugin) getSemanticCacheCost(provider string, model string, inputTokens int) float64 {
	return p.pricingManager.CalculateCostFromUsage(provider, model, &schemas.LLMUsage{
		PromptTokens:     inputTokens,
		CompletionTokens: 0,
		TotalTokens:      inputTokens,
	}, schemas.EmbeddingRequest, false, false, nil, nil)
}

// LogManager defines the main interface that combines all logging functionality
type LogManager interface {
	// Search searches for log entries based on filters and pagination
	Search(filters *logstore.SearchFilters, pagination *logstore.PaginationOptions) (*logstore.SearchResult, error)

	// Get the number of dropped requests
	GetDroppedRequests() int64

	// GetAvailableModels returns all unique models from logs
	GetAvailableModels() []string
}

// PluginLogManager implements LogManager interface wrapping the plugin
type PluginLogManager struct {
	plugin *LoggerPlugin
}

func (p *PluginLogManager) Search(filters *logstore.SearchFilters, pagination *logstore.PaginationOptions) (*logstore.SearchResult, error) {
	if filters == nil || pagination == nil {
		return nil, fmt.Errorf("filters and pagination cannot be nil")
	}
	return p.plugin.SearchLogs(*filters, *pagination)
}

func (p *PluginLogManager) GetDroppedRequests() int64 {
	return p.plugin.droppedRequests.Load()
}

// GetAvailableModels returns all unique models from logs
func (p *PluginLogManager) GetAvailableModels() []string {
	return p.plugin.GetAvailableModels()
}

// GetPluginLogManager returns a LogManager interface for this plugin
func (p *LoggerPlugin) GetPluginLogManager() *PluginLogManager {
	return &PluginLogManager{
		plugin: p,
	}
}
