package logging

import (
	"encoding/json"
	"fmt"
	"math"
	"slices"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/dgraph-io/badger/v4"
)

// storeLogEntry stores a log entry in BadgerDB with optional indexing
func (p *LoggerPlugin) storeLogEntry(entry *LogEntry) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	// Serialize the log entry
	data, err := json.Marshal(entry)
	if err != nil {
		return fmt.Errorf("failed to marshal log entry: %w", err)
	}

	return p.db.Update(func(txn *badger.Txn) error {
		// Store the main log entry
		logKey := LogPrefix + entry.ID
		if err := txn.Set([]byte(logKey), data); err != nil {
			return err
		}

		// Create indexes
		if err := p.createIndexes(txn, entry); err != nil {
			return err
		}

		return nil
	})
}

// createIndexes creates various indexes for efficient searching
func (p *LoggerPlugin) createIndexes(txn *badger.Txn, entry *LogEntry) error {
	timestamp := entry.Timestamp.Unix()

	// Provider index
	if entry.Provider != "" {
		providerKey := fmt.Sprintf("%s%s%s:%d:%s", IndexPrefix, ProviderIndex, entry.Provider, timestamp, entry.ID)
		if err := txn.Set([]byte(providerKey), []byte(entry.ID)); err != nil {
			return err
		}
	}

	// Model index
	if entry.Model != "" {
		modelKey := fmt.Sprintf("%s%s%s:%d:%s", IndexPrefix, ModelIndex, entry.Model, timestamp, entry.ID)
		if err := txn.Set([]byte(modelKey), []byte(entry.ID)); err != nil {
			return err
		}
	}

	// Object index
	if entry.Object != "" {
		objectKey := fmt.Sprintf("%s%s%s:%d:%s", IndexPrefix, ObjectIndex, entry.Object, timestamp, entry.ID)
		if err := txn.Set([]byte(objectKey), []byte(entry.ID)); err != nil {
			return err
		}
	}

	// Timestamp index
	timestampKey := fmt.Sprintf("%s%s%d:%s", IndexPrefix, TimestampIndex, timestamp, entry.ID)
	if err := txn.Set([]byte(timestampKey), []byte(entry.ID)); err != nil {
		return err
	}

	// Status index
	statusKey := fmt.Sprintf("%s%s%s:%d:%s", IndexPrefix, StatusIndex, entry.Status, timestamp, entry.ID)
	if err := txn.Set([]byte(statusKey), []byte(entry.ID)); err != nil {
		return err
	}

	// Latency index (if available)
	if entry.Latency != nil {
		latencyBucket := getLatencyBucket(*entry.Latency)
		latencyKey := fmt.Sprintf("%s%s%d:%d:%s", IndexPrefix, LatencyIndex, latencyBucket, timestamp, entry.ID)
		if err := txn.Set([]byte(latencyKey), []byte(entry.ID)); err != nil {
			return err
		}
	}

	// Token count index (if available)
	if entry.TokenUsage != nil {
		tokenBucket := getTokenBucket(entry.TokenUsage.TotalTokens)
		tokenKey := fmt.Sprintf("%s%s%d:%d:%s", IndexPrefix, TokenIndex, tokenBucket, timestamp, entry.ID)
		if err := txn.Set([]byte(tokenKey), []byte(entry.ID)); err != nil {
			return err
		}
	}

	return nil
}

// SearchLogs searches for log entries based on filters and pagination
func (p *LoggerPlugin) SearchLogs(filters *SearchFilters, pagination *PaginationOptions) (*SearchResult, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if pagination == nil {
		pagination = &PaginationOptions{
			Limit:  50,
			Offset: 0,
			SortBy: "timestamp",
			Order:  "desc",
		}
	}

	var matchingIDs []string
	var allLogs []LogEntry
	seenIDs := make(map[string]bool)

	// Statistics variables
	var successfulRequests int64
	var totalLatency float64
	var totalTokens int64
	var logsWithLatency int64

	err := p.db.View(func(txn *badger.Txn) error {
		if filters != nil {
			// Use indexes for efficient filtering
			matchingIDs = p.searchWithIndexes(txn, filters)
		} else {
			// Fallback to full scan if indexing is disabled
			matchingIDs = p.searchFullScan(txn)
		}

		// Fetch all matching logs, deduplicating by ID
		for _, id := range matchingIDs {
			if !seenIDs[id] {
				if entry, err := p.getLogEntryByID(txn, id); err == nil && p.matchesFilters(entry, filters) {
					allLogs = append(allLogs, *entry)
					seenIDs[id] = true

					// Update statistics
					if entry.Status == "success" {
						successfulRequests++
					}
					if entry.Latency != nil {
						totalLatency += *entry.Latency
						logsWithLatency++
					}
					if entry.TokenUsage != nil {
						totalTokens += int64(entry.TokenUsage.TotalTokens)
					}
				}
			}
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	// Sort logs based on pagination options
	p.sortLogs(allLogs, pagination.SortBy, pagination.Order)

	// Apply pagination
	total := len(allLogs)
	start := pagination.Offset
	end := min(pagination.Offset+pagination.Limit, total)
	if start > total {
		start = total
	}

	// Calculate final statistics
	var successRate float64
	if total > 0 {
		successRate = float64(successfulRequests) / float64(total) * 100
	}

	var averageLatency float64
	if logsWithLatency > 0 {
		averageLatency = totalLatency / float64(logsWithLatency)
	}

	return &SearchResult{
		Logs:       allLogs[start:end],
		Pagination: *pagination,
		Stats: struct {
			TotalRequests  int64   `json:"total_requests"`
			SuccessRate    float64 `json:"success_rate"`
			AverageLatency float64 `json:"average_latency"`
			TotalTokens    int64   `json:"total_tokens"`
		}{
			TotalRequests:  int64(total),
			SuccessRate:    successRate,
			AverageLatency: averageLatency,
			TotalTokens:    totalTokens,
		},
	}, nil
}

// searchWithIndexes uses indexes to find matching log IDs efficiently
func (p *LoggerPlugin) searchWithIndexes(txn *badger.Txn, filters *SearchFilters) []string {
	var candidateIDs []string
	var hasFilters bool

	// Start with timestamp range if specified
	if filters.StartTime != nil || filters.EndTime != nil {
		candidateIDs = p.searchByTimeRange(txn, filters.StartTime, filters.EndTime)
		hasFilters = true
	}

	// Intersect with other filters
	if len(filters.Providers) > 0 {
		providerIDs := p.searchByProviders(txn, filters.Providers)
		if !hasFilters {
			candidateIDs = providerIDs
			hasFilters = true
		} else {
			candidateIDs = p.intersectIDLists(candidateIDs, providerIDs)
		}
	}

	if len(filters.Models) > 0 {
		modelIDs := p.searchByModels(txn, filters.Models)
		if !hasFilters {
			candidateIDs = modelIDs
			hasFilters = true
		} else {
			candidateIDs = p.intersectIDLists(candidateIDs, modelIDs)
		}
	}

	if len(filters.Status) > 0 {
		statusIDs := p.searchByStatus(txn, filters.Status)
		if !hasFilters {
			candidateIDs = statusIDs
			hasFilters = true
		} else {
			candidateIDs = p.intersectIDLists(candidateIDs, statusIDs)
		}
	}

	if len(filters.Objects) > 0 {
		objectIDs := p.searchByObjects(txn, filters.Objects)
		if !hasFilters {
			candidateIDs = objectIDs
			hasFilters = true
		} else {
			candidateIDs = p.intersectIDLists(candidateIDs, objectIDs)
		}
	}

	// Latency range filtering (using buckets for efficiency)
	if filters.MinLatency != nil || filters.MaxLatency != nil {
		latencyIDs := p.searchByLatencyRange(txn, filters.MinLatency, filters.MaxLatency)
		if !hasFilters {
			candidateIDs = latencyIDs
			hasFilters = true
		} else {
			candidateIDs = p.intersectIDLists(candidateIDs, latencyIDs)
		}
	}

	// Token range filtering (using buckets for efficiency)
	if filters.MinTokens != nil || filters.MaxTokens != nil {
		tokenIDs := p.searchByTokenRange(txn, filters.MinTokens, filters.MaxTokens)
		if !hasFilters {
			candidateIDs = tokenIDs
			hasFilters = true
		} else {
			candidateIDs = p.intersectIDLists(candidateIDs, tokenIDs)
		}
	}

	// If no filters were applied, return all logs
	if !hasFilters {
		return p.searchFullScan(txn)
	}

	return candidateIDs
}

// searchFullScan performs a full database scan (fallback when indexes are disabled)
func (p *LoggerPlugin) searchFullScan(txn *badger.Txn) []string {
	var matchingIDs []string

	opts := badger.DefaultIteratorOptions
	opts.PrefetchValues = false
	it := txn.NewIterator(opts)
	defer it.Close()

	prefix := []byte(LogPrefix)
	for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
		item := it.Item()
		key := string(item.Key())
		id := strings.TrimPrefix(key, LogPrefix)
		matchingIDs = append(matchingIDs, id)
	}

	return matchingIDs
}

// Helper methods for index-based searching
func (p *LoggerPlugin) searchByTimeRange(txn *badger.Txn, startTime, endTime *time.Time) []string {
	var ids []string

	opts := badger.DefaultIteratorOptions
	opts.PrefetchValues = false
	it := txn.NewIterator(opts)
	defer it.Close()

	prefix := []byte(IndexPrefix + TimestampIndex)
	for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
		item := it.Item()
		key := string(item.Key())

		// Extract timestamp from key
		parts := strings.Split(strings.TrimPrefix(key, IndexPrefix+TimestampIndex), ":")
		if len(parts) >= 2 {
			if timestamp, err := strconv.ParseInt(parts[0], 10, 64); err == nil {
				logTime := time.Unix(timestamp, 0)
				if (startTime == nil || logTime.After(*startTime)) &&
					(endTime == nil || logTime.Before(*endTime)) {
					if err := item.Value(func(val []byte) error {
						ids = append(ids, string(val))
						return nil
					}); err != nil {
						// Log error but continue processing
						p.logger.Debug(fmt.Sprintf("error getting log entry by ID: %v", err))
					}
				}
			}
		}
	}

	return ids
}

func (p *LoggerPlugin) searchByProviders(txn *badger.Txn, providers []string) []string {
	idMap := make(map[string]bool)

	for _, provider := range providers {
		prefix := []byte(IndexPrefix + ProviderIndex + provider + ":")
		opts := badger.DefaultIteratorOptions
		opts.PrefetchValues = false
		it := txn.NewIterator(opts)

		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
			item := it.Item()
			if err := item.Value(func(val []byte) error {
				idMap[string(val)] = true
				return nil
			}); err == nil {
				// Continue
			}
		}
		it.Close()
	}

	// Convert map to slice
	var ids []string
	for id := range idMap {
		ids = append(ids, id)
	}

	return ids
}

func (p *LoggerPlugin) searchByModels(txn *badger.Txn, models []string) []string {
	idMap := make(map[string]bool)

	for _, model := range models {
		prefix := []byte(IndexPrefix + ModelIndex + model + ":")
		opts := badger.DefaultIteratorOptions
		opts.PrefetchValues = false
		it := txn.NewIterator(opts)

		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
			item := it.Item()
			if err := item.Value(func(val []byte) error {
				idMap[string(val)] = true
				return nil
			}); err == nil {
				// Continue
			}
		}
		it.Close()
	}

	// Convert map to slice
	var ids []string
	for id := range idMap {
		ids = append(ids, id)
	}

	return ids
}

func (p *LoggerPlugin) searchByStatus(txn *badger.Txn, statuses []string) []string {
	idMap := make(map[string]bool)

	for _, status := range statuses {
		prefix := []byte(IndexPrefix + StatusIndex + status + ":")
		opts := badger.DefaultIteratorOptions
		opts.PrefetchValues = false
		it := txn.NewIterator(opts)

		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
			item := it.Item()
			if err := item.Value(func(val []byte) error {
				idMap[string(val)] = true
				return nil
			}); err == nil {
				// Continue
			}
		}
		it.Close()
	}

	// Convert map to slice
	var ids []string
	for id := range idMap {
		ids = append(ids, id)
	}

	return ids
}

func (p *LoggerPlugin) searchByObjects(txn *badger.Txn, objects []string) []string {
	idMap := make(map[string]bool)

	for _, object := range objects {
		prefix := []byte(IndexPrefix + ObjectIndex + object + ":")
		opts := badger.DefaultIteratorOptions
		opts.PrefetchValues = false
		it := txn.NewIterator(opts)

		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
			item := it.Item()
			if err := item.Value(func(val []byte) error {
				idMap[string(val)] = true
				return nil
			}); err == nil {
				// Continue
			}
		}
		it.Close()
	}

	// Convert map to slice
	var ids []string
	for id := range idMap {
		ids = append(ids, id)
	}

	return ids
}

func (p *LoggerPlugin) searchByLatencyRange(txn *badger.Txn, minLatency, maxLatency *float64) []string {
	idMap := make(map[string]bool)

	// Determine which latency buckets to search
	minBucket := 0
	maxBucket := int(math.Pow(10, 6)) // Very large bucket

	if minLatency != nil && *minLatency > 0 {
		minBucket = getLatencyBucket(*minLatency)
	}
	if maxLatency != nil && *maxLatency > 0 {
		maxBucket = getLatencyBucket(*maxLatency)
	}

	// Search through relevant latency buckets
	opts := badger.DefaultIteratorOptions
	opts.PrefetchValues = false
	it := txn.NewIterator(opts)
	defer it.Close()

	prefix := []byte(IndexPrefix + LatencyIndex)
	for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
		item := it.Item()
		key := string(item.Key())

		// Extract bucket from key
		parts := strings.Split(strings.TrimPrefix(key, IndexPrefix+LatencyIndex), ":")
		if len(parts) >= 3 {
			if bucket, err := strconv.Atoi(parts[0]); err == nil {
				if bucket >= minBucket && bucket <= maxBucket {
					if err := item.Value(func(val []byte) error {
						idMap[string(val)] = true
						return nil
					}); err != nil {
						// Log error but continue processing
						p.logger.Debug(fmt.Sprintf("error getting log entry by ID: %v", err))
					}
				}
			}
		}
	}

	// Convert map to slice
	var ids []string
	for id := range idMap {
		ids = append(ids, id)
	}

	return ids
}

func (p *LoggerPlugin) searchByTokenRange(txn *badger.Txn, minTokens, maxTokens *int) []string {
	idMap := make(map[string]bool)

	// Determine which token buckets to search
	minBucket := 0
	maxBucket := int(math.Pow(2, 20)) // Very large bucket

	if minTokens != nil && *minTokens > 0 {
		minBucket = getTokenBucket(*minTokens)
	}
	if maxTokens != nil && *maxTokens > 0 {
		maxBucket = getTokenBucket(*maxTokens)
	}

	// Search through relevant token buckets
	opts := badger.DefaultIteratorOptions
	opts.PrefetchValues = false
	it := txn.NewIterator(opts)
	defer it.Close()

	prefix := []byte(IndexPrefix + TokenIndex)
	for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
		item := it.Item()
		key := string(item.Key())

		// Extract bucket from key
		parts := strings.Split(strings.TrimPrefix(key, IndexPrefix+TokenIndex), ":")
		if len(parts) >= 3 {
			if bucket, err := strconv.Atoi(parts[0]); err == nil {
				if bucket >= minBucket && bucket <= maxBucket {
					if err := item.Value(func(val []byte) error {
						idMap[string(val)] = true
						return nil
					}); err != nil {
						// Log error but continue processing
						p.logger.Debug(fmt.Sprintf("error getting log entry by ID: %v", err))
					}
				}
			}
		}
	}

	// Convert map to slice
	var ids []string
	for id := range idMap {
		ids = append(ids, id)
	}

	return ids
}

// intersectIDLists returns the intersection of two ID lists
func (p *LoggerPlugin) intersectIDLists(list1, list2 []string) []string {
	if len(list1) == 0 || len(list2) == 0 {
		return []string{}
	}

	idMap := make(map[string]bool)
	for _, id := range list1 {
		idMap[id] = true
	}

	var result []string
	for _, id := range list2 {
		if idMap[id] {
			result = append(result, id)
		}
	}

	return result
}

// getLogEntryByID retrieves a log entry by ID
func (p *LoggerPlugin) getLogEntryByID(txn *badger.Txn, id string) (*LogEntry, error) {
	key := LogPrefix + id
	item, err := txn.Get([]byte(key))
	if err != nil {
		return nil, err
	}

	var entry LogEntry
	err = item.Value(func(val []byte) error {
		return json.Unmarshal(val, &entry)
	})

	return &entry, err
}

// matchesFilters checks if a log entry matches the given filters
func (p *LoggerPlugin) matchesFilters(entry *LogEntry, filters *SearchFilters) bool {
	if filters == nil {
		return true
	}

	// Provider filter
	if len(filters.Providers) > 0 {
		found := slices.Contains(filters.Providers, entry.Provider)
		if !found {
			return false
		}
	}

	// Model filter
	if len(filters.Models) > 0 {
		found := slices.Contains(filters.Models, entry.Model)
		if !found {
			return false
		}
	}

	// Status filter
	if len(filters.Status) > 0 {
		found := slices.Contains(filters.Status, entry.Status)
		if !found {
			return false
		}
	}

	// Object type filter
	if len(filters.Objects) > 0 {
		found := slices.Contains(filters.Objects, entry.Object)
		if !found {
			return false
		}
	}

	// Time range filter
	if filters.StartTime != nil && entry.Timestamp.Before(*filters.StartTime) {
		return false
	}
	if filters.EndTime != nil && entry.Timestamp.After(*filters.EndTime) {
		return false
	}

	// Latency filter
	if entry.Latency != nil {
		if filters.MinLatency != nil && *entry.Latency < *filters.MinLatency {
			return false
		}
		if filters.MaxLatency != nil && *entry.Latency > *filters.MaxLatency {
			return false
		}
	}

	// Token count filter
	if entry.TokenUsage != nil {
		if filters.MinTokens != nil && entry.TokenUsage.TotalTokens < *filters.MinTokens {
			return false
		}
		if filters.MaxTokens != nil && entry.TokenUsage.TotalTokens > *filters.MaxTokens {
			return false
		}
	}

	// Content search
	if filters.ContentSearch != "" {
		searchTerm := strings.ToLower(filters.ContentSearch)
		found := false

		// Search in input history
		for _, msg := range entry.InputHistory {
			if msg.Content.ContentStr != nil &&
				strings.Contains(strings.ToLower(*msg.Content.ContentStr), searchTerm) {
				found = true
				break
			}
		}

		// Search in output message
		if !found && entry.OutputMessage != nil && entry.OutputMessage.Content.ContentStr != nil &&
			strings.Contains(strings.ToLower(*entry.OutputMessage.Content.ContentStr), searchTerm) {
			found = true
		}

		if !found {
			return false
		}
	}

	return true
}

// sortLogs sorts log entries based on the specified criteria
func (p *LoggerPlugin) sortLogs(logs []LogEntry, sortBy, order string) {
	sort.Slice(logs, func(i, j int) bool {
		var less bool

		switch sortBy {
		case "latency":
			latencyI := float64(0)
			latencyJ := float64(0)
			if logs[i].Latency != nil {
				latencyI = *logs[i].Latency
			}
			if logs[j].Latency != nil {
				latencyJ = *logs[j].Latency
			}
			less = latencyI < latencyJ
		case "tokens":
			tokensI := 0
			tokensJ := 0
			if logs[i].TokenUsage != nil {
				tokensI = logs[i].TokenUsage.TotalTokens
			}
			if logs[j].TokenUsage != nil {
				tokensJ = logs[j].TokenUsage.TotalTokens
			}
			less = tokensI < tokensJ
		default: // timestamp
			less = logs[i].Timestamp.Before(logs[j].Timestamp)
		}

		if order == "desc" {
			return !less
		}
		return less
	})
}

// sortIDs sorts log IDs based on the specified criteria
func (p *LoggerPlugin) sortIDs(txn *badger.Txn, ids []string, sortBy, order string) {
	// Create a map to cache values for sorting
	cache := make(map[string]interface{})

	// Helper function to get cached value
	getValue := func(id string) interface{} {
		if val, ok := cache[id]; ok {
			return val
		}

		entry, err := p.getLogEntryByID(txn, id)
		if err != nil {
			return nil
		}

		var value interface{}
		switch sortBy {
		case "timestamp":
			value = entry.Timestamp.Unix()
		case "latency":
			if entry.Latency != nil {
				value = *entry.Latency
			}
		case "tokens":
			if entry.TokenUsage != nil {
				value = entry.TokenUsage.TotalTokens
			}
		}

		cache[id] = value
		return value
	}

	// Sort the IDs
	sort.Slice(ids, func(i, j int) bool {
		a := getValue(ids[i])
		b := getValue(ids[j])

		// Handle nil values
		if a == nil {
			return order == "desc"
		}
		if b == nil {
			return order == "asc"
		}

		// Compare based on type
		switch v := a.(type) {
		case int64:
			if order == "asc" {
				return v < b.(int64)
			}
			return v > b.(int64)
		case float64:
			if order == "asc" {
				return v < b.(float64)
			}
			return v > b.(float64)
		case int:
			if order == "asc" {
				return v < b.(int)
			}
			return v > b.(int)
		}

		return false
	})
}

// getLatencyBucket returns the logarithmic bucket (base 10) for a latency value
func getLatencyBucket(latency float64) int {
	if latency <= 0 {
		return 0
	}
	// Use floor(log10(latency)) to get the exponent, then 10^exponent for the bucket
	// This creates buckets like: 0-1ms, 1-10ms, 10-100ms, 100-1000ms, etc.
	exponent := math.Floor(math.Log10(latency))
	return int(math.Pow(10, exponent))
}

// getTokenBucket returns the power-of-2 bucket for a token count
func getTokenBucket(tokens int) int {
	if tokens <= 0 {
		return 0
	}
	// Use floor(log2(tokens)) to get the exponent, then 2^exponent for the bucket
	// This creates buckets like: 0-1, 1-2, 2-4, 4-8, 8-16, 16-32, etc.
	exponent := math.Floor(math.Log2(float64(tokens)))
	return int(math.Pow(2, exponent))
}

// LogManager defines the main interface that combines all logging functionality
type LogManager interface {
	// Search searches for log entries based on filters and pagination
	Search(filters *SearchFilters, pagination *PaginationOptions) (*SearchResult, error)

	// Get the number of dropped requests
	GetDroppedRequests() int64
}

type PluginLogManager struct {
	plugin *LoggerPlugin
}

func (p *PluginLogManager) Search(filters *SearchFilters, pagination *PaginationOptions) (*SearchResult, error) {
	return p.plugin.SearchLogs(filters, pagination)
}

func (p *PluginLogManager) GetDroppedRequests() int64 {
	return p.plugin.droppedRequests.Load()
}

func (p *LoggerPlugin) GetPluginLogManager() *PluginLogManager {
	return &PluginLogManager{
		plugin: p,
	}
}
