package vectorstore

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/maximhq/bifrost/core/schemas"
	"github.com/redis/go-redis/v9"
)

// MockLogger implements schemas.Logger for testing
type MockLogger struct{}

func (m *MockLogger) Debug(msg string, args ...any)                     { fmt.Printf("DEBUG: "+msg+"\n", args...) }
func (m *MockLogger) Info(msg string, args ...any)                      { fmt.Printf("INFO: "+msg+"\n", args...) }
func (m *MockLogger) Warn(msg string, args ...any)                      { fmt.Printf("WARN: "+msg+"\n", args...) }
func (m *MockLogger) Error(msg string, args ...any)                     { fmt.Printf("ERROR: "+msg+"\n", args...) }
func (m *MockLogger) Fatal(msg string, args ...any)                     { fmt.Printf("FATAL: "+msg+"\n", args...) }
func (m *MockLogger) SetLevel(level schemas.LogLevel)                   { /* no-op for testing */ }
func (m *MockLogger) SetOutputType(outputType schemas.LoggerOutputType) { /* no-op for testing */ }

// Test configurations
func getTestRedisClusterConfig() RedisClusterConfig {
	// Use internal Docker network addresses from docker-compose.yml
	addrs := []string{
		"172.38.0.11:6379", // redis-1
		"172.38.0.12:6379", // redis-2
		"172.38.0.13:6379", // redis-3
		"172.38.0.14:6379", // redis-4
		"172.38.0.15:6379", // redis-5
		"172.38.0.16:6379", // redis-6
	}

	// Allow override via environment variable (fallback to localhost for external access)
	if envAddrs := os.Getenv("REDIS_CLUSTER_ADDRS"); envAddrs != "" {
		addrs = strings.Split(envAddrs, ",")
	} else if os.Getenv("USE_LOCALHOST_REDIS") == "true" {
		// Fallback to localhost addresses for external testing
		addrs = []string{
			"localhost:6371", // redis-1
			"localhost:6372", // redis-2
			"localhost:6373", // redis-3
			"localhost:6374", // redis-4
			"localhost:6375", // redis-5
			"localhost:6376", // redis-6
		}
	}

	return RedisClusterConfig{
		Addrs:           addrs,
		MaxRedirects:    3,
		ReadOnly:        false,
		RouteByLatency:  false,
		RouteRandomly:   false,
		PoolSize:        10,
		MinIdleConns:    1,
		MaxIdleConns:    5,
		ConnMaxLifetime: 30 * time.Minute,
		ConnMaxIdleTime: 5 * time.Minute,
		DialTimeout:     5 * time.Second,
		ReadTimeout:     3 * time.Second,
		WriteTimeout:    3 * time.Second,
		ContextTimeout:  10 * time.Second,
	}
}

// Helper function to check if Redis cluster is available and properly configured
func isRedisClusterAvailable(config RedisClusterConfig) bool {
	availableNodes := 0
	
	// First, check how many nodes are accessible
	for _, addr := range config.Addrs {
		client := redis.NewClient(&redis.Options{
			Addr:        addr,
			DialTimeout: 2 * time.Second,
		})

		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		err := client.Ping(ctx).Err()
		cancel()
		client.Close()

		if err != nil {
			fmt.Printf("Redis node %s not available: %v\n", addr, err)
		} else {
			availableNodes++
		}
	}

	if availableNodes == 0 {
		fmt.Println("No Redis nodes are available")
		return false
	}

	fmt.Printf("Found %d available Redis nodes out of %d\n", availableNodes, len(config.Addrs))

	// Try to create a cluster client and test basic functionality
	client := redis.NewClusterClient(&redis.ClusterOptions{
		Addrs:        config.Addrs,
		DialTimeout:  2 * time.Second,
		ReadTimeout:  2 * time.Second,
		WriteTimeout: 2 * time.Second,
		MaxRedirects: 3,
	})
	defer client.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Try a simple cluster operation to verify it's working
	testKey := "test:cluster:availability"
	err := client.Set(ctx, testKey, "test", time.Minute).Err()
	if err != nil {
		fmt.Printf("Cluster not properly configured: %v\n", err)
		return false
	}

	// Clean up test key
	client.Del(ctx, testKey)
	
	fmt.Println("Redis cluster is available and properly configured")
	return true
}

func TestRedisClusterStore_Connection(t *testing.T) {
	config := getTestRedisClusterConfig()
	logger := &MockLogger{}

	t.Run("successful connection", func(t *testing.T) {
		if !isRedisClusterAvailable(config) {
			t.Skip("Redis cluster not available, skipping test")
		}

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		store, err := newRedisClusterStore(ctx, config, logger)
		if err != nil {
			t.Fatalf("Failed to create Redis cluster store: %v", err)
		}
		if store == nil {
			t.Fatal("Store should not be nil")
		}

		// Test that we can actually use the connection
		err = store.Add(ctx, "test:connection", "test_value", time.Minute)
		if err != nil {
			t.Errorf("Should be able to add a key: %v", err)
		}

		value, err := store.GetChunk(ctx, "test:connection")
		if err != nil {
			t.Errorf("Should be able to get a key: %v", err)
		}
		if value != "test_value" {
			t.Errorf("Retrieved value should match: expected 'test_value', got '%s'", value)
		}

		// Cleanup
		err = store.Delete(ctx, []string{"test:connection"})
		if err != nil {
			t.Errorf("Should be able to delete keys: %v", err)
		}

		err = store.Close(ctx)
		if err != nil {
			t.Errorf("Should be able to close connection: %v", err)
		}
	})

	t.Run("connection with invalid addresses", func(t *testing.T) {
		invalidConfig := config
		invalidConfig.Addrs = []string{"localhost:9999", "localhost:9998"}

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		store, err := newRedisClusterStore(ctx, invalidConfig, logger)
		if err == nil {
			t.Error("Should fail with invalid addresses")
		}
		if store != nil {
			t.Error("Store should be nil on error")
		}
	})

	t.Run("connection with empty addresses", func(t *testing.T) {
		invalidConfig := config
		invalidConfig.Addrs = []string{}

		ctx := context.Background()

		store, err := newRedisClusterStore(ctx, invalidConfig, logger)
		if err == nil {
			t.Error("Should fail with empty addresses")
		}
		if store != nil {
			t.Error("Store should be nil on error")
		}
		if err != nil && !strings.Contains(err.Error(), "at least one Redis cluster address is required") {
			t.Errorf("Error should mention required addresses, got: %v", err)
		}
	})
}

func TestRedisClusterStore_BasicOperations(t *testing.T) {
	config := getTestRedisClusterConfig()
	logger := &MockLogger{}

	if !isRedisClusterAvailable(config) {
		t.Fatal("Redis cluster not available, skipping test")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	store, err := newRedisClusterStore(ctx, config, logger)
	if err != nil {
		t.Fatalf("Failed to create Redis cluster store: %v", err)
	}
	defer store.Close(ctx)

	t.Run("add and get single value", func(t *testing.T) {
		key := "test:single"
		value := "single_test_value"

		err := store.Add(ctx, key, value, time.Minute)
		if err != nil {
			t.Errorf("Should be able to add key: %v", err)
		}

		retrieved, err := store.GetChunk(ctx, key)
		if err != nil {
			t.Errorf("Should be able to get key: %v", err)
		}
		if retrieved != value {
			t.Errorf("Retrieved value should match: expected '%s', got '%s'", value, retrieved)
		}

		// Cleanup
		err = store.Delete(ctx, []string{key})
		if err != nil {
			t.Errorf("Should be able to delete key: %v", err)
		}
	})

	t.Run("add and get multiple values", func(t *testing.T) {
		keys := []string{"{test:multi}:1", "{test:multi}:2", "{test:multi}:3"}
		values := []string{"value1", "value2", "value3"}

		// Add multiple keys
		for i, key := range keys {
			err := store.Add(ctx, key, values[i], time.Minute)
			if err != nil {
				t.Errorf("Should be able to add key %s: %v", key, err)
			}
		}

		// Get multiple keys
		retrieved, err := store.GetChunks(ctx, keys)
		if err != nil {
			t.Errorf("Should be able to get multiple keys: %v", err)
		}
		if len(retrieved) != 3 {
			t.Errorf("Should retrieve 3 values, got %d", len(retrieved))
		}

		// Convert interface{} to strings and verify
		for i, val := range retrieved {
			if val != values[i] {
				t.Errorf("Retrieved value %d should match: expected '%s', got '%v'", i, values[i], val)
			}
		}

		// Cleanup
		err = store.Delete(ctx, keys)
		if err != nil {
			t.Errorf("Should be able to delete multiple keys: %v", err)
		}
	})

	t.Run("get non-existent key", func(t *testing.T) {
		_, err := store.GetChunk(ctx, "test:nonexistent")
		if err == nil {
			t.Error("Should return error for non-existent key")
		}
		if !errors.Is(err, ErrNotFound) {
			t.Errorf("Should return ErrNotFound error, got: %v", err)
		}
	})

	t.Run("delete non-existent keys", func(t *testing.T) {
		err := store.Delete(ctx, []string{"{test:nonexistent}:1", "{test:nonexistent}:2"})
		// Delete should not return error even if keys don't exist
		if err != nil {
			t.Errorf("Delete should not fail for non-existent keys: %v", err)
		}
	})
}

func TestRedisClusterStore_GetAllOperations(t *testing.T) {
	config := getTestRedisClusterConfig()
	logger := &MockLogger{}

	if !isRedisClusterAvailable(config) {
		t.Fatal("Redis cluster not available, skipping test")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	store, err := newRedisClusterStore(ctx, config, logger)
	if err != nil {
		t.Fatalf("Failed to create Redis cluster store: %v", err)
	}
	defer store.Close(ctx)

	// Setup test data
	testKeys := []string{
		"{test:getall}:item1",
		"{test:getall}:item2",
		"{test:getall}:item3",
		"{test:getall}:special:item4",
		"{other:key}:item5",
	}

	// Add test data
	for i, key := range testKeys {
		err := store.Add(ctx, key, fmt.Sprintf("value%d", i+1), time.Minute)
		if err != nil {
			t.Fatalf("Should be able to add test key %s: %v", key, err)
		}
	}

	t.Run("get all keys with pattern", func(t *testing.T) {
		keys, cursor, err := store.GetAll(ctx, "{test:getall}*", nil, 10)
		if err != nil {
			t.Errorf("Should be able to get keys with pattern: %v", err)
		}
		if cursor != nil {
			t.Error("Cursor should be nil when all results fit in one page")
		}

		// Should find the first 4 keys that match the pattern
		expectedKeys := []string{
			"{test:getall}:item1",
			"{test:getall}:item2",
			"{test:getall}:item3",
			"{test:getall}:special:item4",
		}

		// Since Redis cluster distributes keys across nodes, we need to check that
		// all expected keys are present (order might vary)
		if len(keys) != 4 {
			t.Errorf("Should find 4 matching keys, got %d", len(keys))
		}
		for _, expectedKey := range expectedKeys {
			found := false
			for _, key := range keys {
				if key == expectedKey {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("Should contain key %s", expectedKey)
			}
		}
	})

	t.Run("get all keys with pagination", func(t *testing.T) {
		// Use a smaller count to test pagination
		keys, cursor, err := store.GetAll(ctx, "{test:getall}*", nil, 2)
		if err != nil {
			t.Errorf("Should be able to get keys with pagination: %v", err)
		}

		// We should get some keys, and potentially a cursor for more
		if len(keys) == 0 && cursor == nil {
			t.Error("Should get some keys or have a cursor for more")
		}

		// If there's a cursor, try to get more
		if cursor != nil {
			moreKeys, nextCursor, err := store.GetAll(ctx, "{test:getall}*", cursor, 2)
			if err != nil {
				t.Errorf("Should be able to get more keys with cursor: %v", err)
			}
			// In a cluster, we might not get more keys if they're all on nodes we've already scanned
			// This is acceptable behavior

			// Continue until no more cursor
			allKeys := append(keys, moreKeys...)
			for nextCursor != nil {
				additionalKeys, newCursor, err := store.GetAll(ctx, "{test:getall}*", nextCursor, 2)
				if err != nil {
					t.Errorf("Should be able to continue pagination: %v", err)
				}
				allKeys = append(allKeys, additionalKeys...)
				nextCursor = newCursor
			}

			// Total should be 4 keys
			if len(allKeys) != 4 {
				t.Errorf("Should eventually find all 4 matching keys, got %d: %v", len(allKeys), allKeys)
			}
		}
	})

	// Cleanup
	err = store.Delete(ctx, testKeys)
	if err != nil {
		t.Errorf("Should be able to cleanup test keys: %v", err)
	}
}

func TestRedisClusterStore_TTLOperations(t *testing.T) {
	config := getTestRedisClusterConfig()
	logger := &MockLogger{}

	if !isRedisClusterAvailable(config) {
		t.Skip("Redis cluster not available, skipping test")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	store, err := newRedisClusterStore(ctx, config, logger)
	if err != nil {
		t.Fatalf("Failed to create Redis cluster store: %v", err)
	}
	defer store.Close(ctx)

	t.Run("key expires after TTL", func(t *testing.T) {
		key := "test:ttl:expire"
		value := "expires_soon"

		// Add key with short TTL
		err := store.Add(ctx, key, value, 2*time.Second)
		if err != nil {
			t.Errorf("Should be able to add key with TTL: %v", err)
		}

		// Key should exist immediately
		retrieved, err := store.GetChunk(ctx, key)
		if err != nil {
			t.Errorf("Key should exist immediately: %v", err)
		}
		if retrieved != value {
			t.Errorf("Value should match: expected '%s', got '%s'", value, retrieved)
		}

		// Wait for expiration
		time.Sleep(3 * time.Second)

		// Key should be expired
		_, err = store.GetChunk(ctx, key)
		if err == nil {
			t.Error("Key should be expired")
		}
		if !errors.Is(err, ErrNotFound) {
			t.Errorf("Should return ErrNotFound for expired key, got: %v", err)
		}
	})

	t.Run("key with zero TTL persists", func(t *testing.T) {
		key := "test:ttl:persist"
		value := "persists"

		// Add key with zero TTL (no expiration)
		err := store.Add(ctx, key, value, 0)
		if err != nil {
			t.Errorf("Should be able to add key with zero TTL: %v", err)
		}

		// Key should exist
		retrieved, err := store.GetChunk(ctx, key)
		if err != nil {
			t.Errorf("Key should exist: %v", err)
		}
		if retrieved != value {
			t.Errorf("Value should match: expected '%s', got '%s'", value, retrieved)
		}

		// Cleanup
		err = store.Delete(ctx, []string{key})
		if err != nil {
			t.Errorf("Should be able to delete persistent key: %v", err)
		}
	})
}

func TestRedisClusterStore_ConcurrentOperations(t *testing.T) {
	config := getTestRedisClusterConfig()
	logger := &MockLogger{}

	if !isRedisClusterAvailable(config) {
		t.Fatal("Redis cluster not available, skipping test")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	store, err := newRedisClusterStore(ctx, config, logger)
	if err != nil {
		t.Fatalf("Failed to create Redis cluster store: %v", err)
	}
	defer store.Close(ctx)

	t.Run("concurrent writes and reads", func(t *testing.T) {
		const numGoroutines = 10
		const numOperations = 50

		// Channel to collect errors
		errChan := make(chan error, numGoroutines*numOperations)

		// Start multiple goroutines doing concurrent operations
		for i := 0; i < numGoroutines; i++ {
			go func(routineID int) {
				for j := 0; j < numOperations; j++ {
					key := fmt.Sprintf("test:concurrent:%d:%d", routineID, j)
					value := fmt.Sprintf("value_%d_%d", routineID, j)

					// Add key
					if err := store.Add(ctx, key, value, time.Minute); err != nil {
						errChan <- fmt.Errorf("failed to add key %s: %w", key, err)
						continue
					}

					// Read key back
					retrieved, err := store.GetChunk(ctx, key)
					if err != nil {
						errChan <- fmt.Errorf("failed to get key %s: %w", key, err)
						continue
					}

					if retrieved != value {
						errChan <- fmt.Errorf("value mismatch for key %s: expected %s, got %s", key, value, retrieved)
						continue
					}

					// Delete key
					if err := store.Delete(ctx, []string{key}); err != nil {
						errChan <- fmt.Errorf("failed to delete key %s: %w", key, err)
						continue
					}
				}
			}(i)
		}

		// Wait a bit for operations to complete
		time.Sleep(10 * time.Second)

		// Check for errors
		close(errChan)
		var errors []error
		for err := range errChan {
			errors = append(errors, err)
		}

		if len(errors) > 0 {
			t.Errorf("Got %d errors during concurrent operations:", len(errors))
			for i, err := range errors {
				if i < 10 { // Limit output to first 10 errors
					t.Errorf("  Error %d: %v", i+1, err)
				}
			}
		}
	})
}

func TestRedisClusterStore_ClusterCursor(t *testing.T) {
	t.Run("cluster cursor serialization", func(t *testing.T) {
		cursor := ClusterCursor{
			NodeCursors: map[string]uint64{
				"node1:6379": 123,
				"node2:6379": 456,
				"node3:6379": 0,
			},
		}

		// Test JSON marshaling
		data, err := json.Marshal(cursor)
		if err != nil {
			t.Errorf("Should be able to marshal cursor: %v", err)
		}

		// Test JSON unmarshaling
		var unmarshaled ClusterCursor
		err = json.Unmarshal(data, &unmarshaled)
		if err != nil {
			t.Errorf("Should be able to unmarshal cursor: %v", err)
		}

		// Compare the maps
		if len(cursor.NodeCursors) != len(unmarshaled.NodeCursors) {
			t.Error("Cursors should have same length")
		}
		for k, v := range cursor.NodeCursors {
			if unmarshaled.NodeCursors[k] != v {
				t.Errorf("Cursor mismatch for key %s: expected %d, got %d", k, v, unmarshaled.NodeCursors[k])
			}
		}
	})

	t.Run("empty cursor handling", func(t *testing.T) {
		var cursor ClusterCursor
		data, err := json.Marshal(cursor)
		if err != nil {
			t.Errorf("Should be able to marshal empty cursor: %v", err)
		}

		var unmarshaled ClusterCursor
		err = json.Unmarshal(data, &unmarshaled)
		if err != nil {
			t.Errorf("Should be able to unmarshal empty cursor: %v", err)
		}
	})
}

func TestRedisClusterStore_ErrorHandling(t *testing.T) {
	config := getTestRedisClusterConfig()
	logger := &MockLogger{}

	if !isRedisClusterAvailable(config) {
		t.Fatal("Redis cluster not available, skipping test")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	store, err := newRedisClusterStore(ctx, config, logger)
	if err != nil {
		t.Fatalf("Failed to create Redis cluster store: %v", err)
	}
	defer store.Close(ctx)

	t.Run("context timeout", func(t *testing.T) {
		// Create a context that times out quickly
		shortCtx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
		defer cancel()

		// This should timeout
		err := store.Add(shortCtx, "test:timeout", "value", time.Minute)
		if err == nil {
			t.Error("Should timeout with short context")
		}
		if !strings.Contains(err.Error(), "context deadline exceeded") {
			t.Errorf("Should timeout with context deadline exceeded, got: %v", err)
		}
	})

	t.Run("invalid cursor in GetAll", func(t *testing.T) {
		// Add a test key first to ensure there are keys to find
		testKey := "{error:test}:key"
		err := store.Add(ctx, testKey, "test_value", time.Minute)
		if err != nil {
			t.Fatalf("Should be able to add test key: %v", err)
		}

		invalidCursor := "invalid_json"
		keys, cursor, err := store.GetAll(ctx, "{error:test}*", &invalidCursor, 10)

		// Should handle invalid cursor gracefully by starting fresh
		if err != nil {
			t.Errorf("Should handle invalid cursor gracefully: %v", err)
		}
		if keys == nil {
			t.Error("Should return keys")
		}
		if len(keys) == 0 {
			t.Error("Should find at least one key")
		}
		_ = cursor // cursor might be nil or valid

		// Cleanup
		store.Delete(ctx, []string{testKey})
	})
}

// Benchmark tests
func BenchmarkRedisClusterStore_Add(b *testing.B) {
	config := getTestRedisClusterConfig()
	logger := &MockLogger{}

	if !isRedisClusterAvailable(config) {
		b.Fatal("Redis cluster not available, skipping benchmark")
	}

	ctx := context.Background()
	store, err := newRedisClusterStore(ctx, config, logger)
	if err != nil {
		b.Fatalf("Failed to create store: %v", err)
	}
	defer store.Close(ctx)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			key := fmt.Sprintf("bench:add:%d", i)
			err := store.Add(ctx, key, "benchmark_value", time.Minute)
			if err != nil {
				b.Errorf("Failed to add key: %v", err)
			}
			i++
		}
	})
}

func BenchmarkRedisClusterStore_Get(b *testing.B) {
	config := getTestRedisClusterConfig()
	logger := &MockLogger{}

	if !isRedisClusterAvailable(config) {
		b.Fatal("Redis cluster not available, skipping benchmark")
	}

	ctx := context.Background()
	store, err := newRedisClusterStore(ctx, config, logger)
	if err != nil {
		b.Fatalf("Failed to create store: %v", err)
	}
	defer store.Close(ctx)

	// Pre-populate some keys
	for i := 0; i < 1000; i++ {
		key := fmt.Sprintf("bench:get:%d", i)
		err := store.Add(ctx, key, "benchmark_value", time.Minute)
		if err != nil {
			b.Fatalf("Failed to pre-populate key: %v", err)
		}
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			key := fmt.Sprintf("bench:get:%d", i%1000)
			_, err := store.GetChunk(ctx, key)
			if err != nil {
				b.Errorf("Failed to get key: %v", err)
			}
			i++
		}
	})
}
