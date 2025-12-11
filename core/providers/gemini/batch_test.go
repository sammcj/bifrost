package gemini

import (
	"testing"

	"github.com/maximhq/bifrost/core/schemas"
	"github.com/stretchr/testify/assert"
)

// TestToBifrostBatchStatus tests the conversion from Gemini batch states to Bifrost batch status
func TestToBifrostBatchStatus(t *testing.T) {
	tests := []struct {
		name        string
		geminiState string
		expected    schemas.BatchStatus
	}{
		{
			name:        "Pending state",
			geminiState: GeminiBatchStatePending,
			expected:    schemas.BatchStatusInProgress,
		},
		{
			name:        "Running state",
			geminiState: GeminiBatchStateRunning,
			expected:    schemas.BatchStatusInProgress,
		},
		{
			name:        "Succeeded state",
			geminiState: GeminiBatchStateSucceeded,
			expected:    schemas.BatchStatusCompleted,
		},
		{
			name:        "Failed state",
			geminiState: GeminiBatchStateFailed,
			expected:    schemas.BatchStatusFailed,
		},
		{
			name:        "Cancelling state",
			geminiState: GeminiBatchStateCancelling,
			expected:    schemas.BatchStatusCancelling,
		},
		{
			name:        "Cancelled state",
			geminiState: GeminiBatchStateCancelled,
			expected:    schemas.BatchStatusCancelled,
		},
		{
			name:        "Expired state",
			geminiState: GeminiBatchStateExpired,
			expected:    schemas.BatchStatusExpired,
		},
		{
			name:        "Unspecified state",
			geminiState: GeminiBatchStateUnspecified,
			expected:    schemas.BatchStatusInProgress,
		},
		{
			name:        "Unknown state defaults to InProgress",
			geminiState: "UNKNOWN_STATE",
			expected:    schemas.BatchStatusInProgress,
		},
		{
			name:        "Empty state defaults to InProgress",
			geminiState: "",
			expected:    schemas.BatchStatusInProgress,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ToBifrostBatchStatus(tt.geminiState)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestParseGeminiTimestamp tests the timestamp parsing function
func TestParseGeminiTimestamp(t *testing.T) {
	t.Run("Empty timestamp returns 0", func(t *testing.T) {
		result := parseGeminiTimestamp("")
		assert.Equal(t, int64(0), result)
	})

	t.Run("Invalid timestamp format returns 0", func(t *testing.T) {
		result := parseGeminiTimestamp("invalid-timestamp")
		assert.Equal(t, int64(0), result)
	})

	t.Run("Unix epoch returns 0", func(t *testing.T) {
		result := parseGeminiTimestamp("1970-01-01T00:00:00Z")
		assert.Equal(t, int64(0), result)
	})

	t.Run("Valid RFC3339 timestamp parses correctly", func(t *testing.T) {
		result := parseGeminiTimestamp("2024-01-15T10:30:00Z")
		// Should be a positive value for a date in 2024
		assert.Greater(t, result, int64(1704067200)) // Jan 1, 2024 00:00:00 UTC
		assert.Less(t, result, int64(1735689600))    // Jan 1, 2025 00:00:00 UTC
	})

	t.Run("Valid RFC3339 timestamp with timezone parses correctly", func(t *testing.T) {
		result := parseGeminiTimestamp("2024-01-15T10:30:00+00:00")
		// Should be a positive value for a date in 2024
		assert.Greater(t, result, int64(1704067200)) // Jan 1, 2024 00:00:00 UTC
		assert.Less(t, result, int64(1735689600))    // Jan 1, 2025 00:00:00 UTC
	})

	t.Run("Same timestamp with different formats produces same result", func(t *testing.T) {
		result1 := parseGeminiTimestamp("2024-01-15T10:30:00Z")
		result2 := parseGeminiTimestamp("2024-01-15T10:30:00+00:00")
		assert.Equal(t, result1, result2)
	})

	t.Run("Valid timestamp with nanoseconds parses correctly", func(t *testing.T) {
		result := parseGeminiTimestamp("2024-06-15T14:30:45.123456789Z")
		// Should be a positive value for a date in June 2024
		assert.Greater(t, result, int64(1717200000)) // June 1, 2024 approximately
		assert.Less(t, result, int64(1719792000))    // July 1, 2024 approximately
	})

	t.Run("Timestamps are monotonically increasing", func(t *testing.T) {
		result1 := parseGeminiTimestamp("2024-01-01T00:00:00Z")
		result2 := parseGeminiTimestamp("2024-06-01T00:00:00Z")
		result3 := parseGeminiTimestamp("2024-12-01T00:00:00Z")
		assert.Less(t, result1, result2)
		assert.Less(t, result2, result3)
	})

	t.Run("Invalid date returns 0", func(t *testing.T) {
		result := parseGeminiTimestamp("2024-13-45T25:61:61Z") // Invalid month, day, hour, minute, second
		assert.Equal(t, int64(0), result)
	})
}

// TestExtractBatchIDFromName tests the batch ID extraction function
func TestExtractBatchIDFromName(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Full resource name",
			input:    "batches/abc123",
			expected: "batches/abc123",
		},
		{
			name:     "Simple ID",
			input:    "abc123",
			expected: "abc123",
		},
		{
			name:     "Empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "Long resource name",
			input:    "batches/batch-xyz-12345-abcdef",
			expected: "batches/batch-xyz-12345-abcdef",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractBatchIDFromName(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestBuildBatchRequestItems tests the conversion from Bifrost batch requests to Gemini format
func TestBuildBatchRequestItems(t *testing.T) {
	tests := []struct {
		name     string
		requests []schemas.BatchRequestItem
		validate func(t *testing.T, result []GeminiBatchRequestItem)
	}{
		{
			name:     "Empty requests",
			requests: []schemas.BatchRequestItem{},
			validate: func(t *testing.T, result []GeminiBatchRequestItem) {
				assert.Len(t, result, 0)
			},
		},
		{
			name: "Single request with Body",
			requests: []schemas.BatchRequestItem{
				{
					CustomID: "req-1",
					Body: map[string]interface{}{
						"messages": []interface{}{
							map[string]interface{}{
								"role":    "user",
								"content": "Hello, world!",
							},
						},
					},
				},
			},
			validate: func(t *testing.T, result []GeminiBatchRequestItem) {
				assert.Len(t, result, 1)
				assert.NotNil(t, result[0].Metadata)
				assert.Equal(t, "req-1", result[0].Metadata.Key)
				assert.Len(t, result[0].Request.Contents, 1)
				assert.Equal(t, "user", result[0].Request.Contents[0].Role)
				assert.Len(t, result[0].Request.Contents[0].Parts, 1)
				assert.Equal(t, "Hello, world!", result[0].Request.Contents[0].Parts[0].Text)
			},
		},
		{
			name: "Single request with Params (fallback)",
			requests: []schemas.BatchRequestItem{
				{
					CustomID: "req-2",
					Params: map[string]interface{}{
						"messages": []interface{}{
							map[string]interface{}{
								"role":    "user",
								"content": "Test message",
							},
						},
					},
				},
			},
			validate: func(t *testing.T, result []GeminiBatchRequestItem) {
				assert.Len(t, result, 1)
				assert.NotNil(t, result[0].Metadata)
				assert.Equal(t, "req-2", result[0].Metadata.Key)
				assert.Len(t, result[0].Request.Contents, 1)
				assert.Equal(t, "Test message", result[0].Request.Contents[0].Parts[0].Text)
			},
		},
		{
			name: "Request with assistant role",
			requests: []schemas.BatchRequestItem{
				{
					CustomID: "req-3",
					Body: map[string]interface{}{
						"messages": []interface{}{
							map[string]interface{}{
								"role":    "assistant",
								"content": "I am an assistant",
							},
						},
					},
				},
			},
			validate: func(t *testing.T, result []GeminiBatchRequestItem) {
				assert.Len(t, result, 1)
				assert.Len(t, result[0].Request.Contents, 1)
				assert.Equal(t, "model", result[0].Request.Contents[0].Role)
			},
		},
		{
			name: "Request with system role is skipped",
			requests: []schemas.BatchRequestItem{
				{
					CustomID: "req-4",
					Body: map[string]interface{}{
						"messages": []interface{}{
							map[string]interface{}{
								"role":    "system",
								"content": "You are a helpful assistant",
							},
							map[string]interface{}{
								"role":    "user",
								"content": "Hello",
							},
						},
					},
				},
			},
			validate: func(t *testing.T, result []GeminiBatchRequestItem) {
				assert.Len(t, result, 1)
				// System message should be skipped
				assert.Len(t, result[0].Request.Contents, 1)
				assert.Equal(t, "user", result[0].Request.Contents[0].Role)
			},
		},
		{
			name: "Multiple requests",
			requests: []schemas.BatchRequestItem{
				{
					CustomID: "req-a",
					Body: map[string]interface{}{
						"messages": []interface{}{
							map[string]interface{}{
								"role":    "user",
								"content": "First message",
							},
						},
					},
				},
				{
					CustomID: "req-b",
					Body: map[string]interface{}{
						"messages": []interface{}{
							map[string]interface{}{
								"role":    "user",
								"content": "Second message",
							},
						},
					},
				},
			},
			validate: func(t *testing.T, result []GeminiBatchRequestItem) {
				assert.Len(t, result, 2)
				assert.Equal(t, "req-a", result[0].Metadata.Key)
				assert.Equal(t, "req-b", result[1].Metadata.Key)
			},
		},
		{
			name: "Request without custom ID",
			requests: []schemas.BatchRequestItem{
				{
					Body: map[string]interface{}{
						"messages": []interface{}{
							map[string]interface{}{
								"role":    "user",
								"content": "No custom ID",
							},
						},
					},
				},
			},
			validate: func(t *testing.T, result []GeminiBatchRequestItem) {
				assert.Len(t, result, 1)
				assert.Nil(t, result[0].Metadata)
			},
		},
		{
			name: "Request with nil Body and nil Params",
			requests: []schemas.BatchRequestItem{
				{
					CustomID: "req-empty",
				},
			},
			validate: func(t *testing.T, result []GeminiBatchRequestItem) {
				assert.Len(t, result, 1)
				assert.Len(t, result[0].Request.Contents, 0)
			},
		},
		{
			name: "Request with multi-turn conversation",
			requests: []schemas.BatchRequestItem{
				{
					CustomID: "req-multi",
					Body: map[string]interface{}{
						"messages": []interface{}{
							map[string]interface{}{
								"role":    "user",
								"content": "Hi there!",
							},
							map[string]interface{}{
								"role":    "assistant",
								"content": "Hello! How can I help?",
							},
							map[string]interface{}{
								"role":    "user",
								"content": "Tell me a joke",
							},
						},
					},
				},
			},
			validate: func(t *testing.T, result []GeminiBatchRequestItem) {
				assert.Len(t, result, 1)
				assert.Len(t, result[0].Request.Contents, 3)
				assert.Equal(t, "user", result[0].Request.Contents[0].Role)
				assert.Equal(t, "model", result[0].Request.Contents[1].Role)
				assert.Equal(t, "user", result[0].Request.Contents[2].Role)
			},
		},
		{
			name: "Request with messages but missing content",
			requests: []schemas.BatchRequestItem{
				{
					CustomID: "req-no-content",
					Body: map[string]interface{}{
						"messages": []interface{}{
							map[string]interface{}{
								"role": "user",
							},
						},
					},
				},
			},
			validate: func(t *testing.T, result []GeminiBatchRequestItem) {
				assert.Len(t, result, 1)
				assert.Len(t, result[0].Request.Contents, 1)
				// Content should have empty parts since content was missing
				assert.Len(t, result[0].Request.Contents[0].Parts, 0)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := buildBatchRequestItems(tt.requests)
			tt.validate(t, result)
		})
	}
}

// TestGeminiBatchStats_UnmarshalJSON tests the custom JSON unmarshaling for batch stats
func TestGeminiBatchStats_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		name     string
		json     string
		expected GeminiBatchStats
		wantErr  bool
	}{
		{
			name: "Valid stats with string numbers",
			json: `{"requestCount": "10", "pendingRequestCount": "5", "successfulRequestCount": "3"}`,
			expected: GeminiBatchStats{
				RequestCount:           10,
				PendingRequestCount:    5,
				SuccessfulRequestCount: 3,
			},
			wantErr: false,
		},
		{
			name: "Empty stats",
			json: `{}`,
			expected: GeminiBatchStats{
				RequestCount:           0,
				PendingRequestCount:    0,
				SuccessfulRequestCount: 0,
			},
			wantErr: false,
		},
		{
			name: "Partial stats",
			json: `{"requestCount": "100"}`,
			expected: GeminiBatchStats{
				RequestCount:           100,
				PendingRequestCount:    0,
				SuccessfulRequestCount: 0,
			},
			wantErr: false,
		},
		{
			name:     "Invalid number format",
			json:     `{"requestCount": "not-a-number"}`,
			expected: GeminiBatchStats{},
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var stats GeminiBatchStats
			err := stats.UnmarshalJSON([]byte(tt.json))

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, stats)
			}
		})
	}
}

// TestGeminiBatchStats_MarshalJSON tests the custom JSON marshaling for batch stats
func TestGeminiBatchStats_MarshalJSON(t *testing.T) {
	tests := []struct {
		name     string
		stats    GeminiBatchStats
		expected string
	}{
		{
			name: "Full stats",
			stats: GeminiBatchStats{
				RequestCount:           10,
				PendingRequestCount:    5,
				SuccessfulRequestCount: 3,
			},
			expected: `{"requestCount":"10","pendingRequestCount":"5","successfulRequestCount":"3"}`,
		},
		{
			name: "Zero stats",
			stats: GeminiBatchStats{
				RequestCount:           0,
				PendingRequestCount:    0,
				SuccessfulRequestCount: 0,
			},
			expected: `{"requestCount":"0","pendingRequestCount":"0","successfulRequestCount":"0"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := tt.stats.MarshalJSON()
			assert.NoError(t, err)
			assert.JSONEq(t, tt.expected, string(result))
		})
	}
}

// TestBatchStatusConstants verifies all batch state constants are defined correctly
func TestBatchStatusConstants(t *testing.T) {
	// Verify all constants exist and have expected values
	assert.Equal(t, "BATCH_STATE_UNSPECIFIED", GeminiBatchStateUnspecified)
	assert.Equal(t, "BATCH_STATE_PENDING", GeminiBatchStatePending)
	assert.Equal(t, "BATCH_STATE_RUNNING", GeminiBatchStateRunning)
	assert.Equal(t, "BATCH_STATE_SUCCEEDED", GeminiBatchStateSucceeded)
	assert.Equal(t, "BATCH_STATE_FAILED", GeminiBatchStateFailed)
	assert.Equal(t, "BATCH_STATE_CANCELLING", GeminiBatchStateCancelling)
	assert.Equal(t, "BATCH_STATE_CANCELLED", GeminiBatchStateCancelled)
	assert.Equal(t, "BATCH_STATE_EXPIRED", GeminiBatchStateExpired)
}

// TestBuildBatchRequestItems_RoleMapping tests that role mapping works correctly
func TestBuildBatchRequestItems_RoleMapping(t *testing.T) {
	tests := []struct {
		name         string
		inputRole    string
		expectedRole string
		shouldSkip   bool
	}{
		{
			name:         "User role stays user",
			inputRole:    "user",
			expectedRole: "user",
			shouldSkip:   false,
		},
		{
			name:         "Assistant role becomes model",
			inputRole:    "assistant",
			expectedRole: "model",
			shouldSkip:   false,
		},
		{
			name:       "System role is skipped",
			inputRole:  "system",
			shouldSkip: true,
		},
		{
			name:         "Other roles preserved",
			inputRole:    "function",
			expectedRole: "function",
			shouldSkip:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			requests := []schemas.BatchRequestItem{
				{
					CustomID: "test",
					Body: map[string]interface{}{
						"messages": []interface{}{
							map[string]interface{}{
								"role":    tt.inputRole,
								"content": "test content",
							},
						},
					},
				},
			}

			result := buildBatchRequestItems(requests)
			assert.Len(t, result, 1)

			if tt.shouldSkip {
				assert.Len(t, result[0].Request.Contents, 0)
			} else {
				assert.Len(t, result[0].Request.Contents, 1)
				assert.Equal(t, tt.expectedRole, result[0].Request.Contents[0].Role)
			}
		})
	}
}

// TestBuildBatchRequestItems_BodyVsParams tests priority of Body over Params
func TestBuildBatchRequestItems_BodyVsParams(t *testing.T) {
	// When both Body and Params are set, Body should take priority
	requests := []schemas.BatchRequestItem{
		{
			CustomID: "test",
			Body: map[string]interface{}{
				"messages": []interface{}{
					map[string]interface{}{
						"role":    "user",
						"content": "from body",
					},
				},
			},
			Params: map[string]interface{}{
				"messages": []interface{}{
					map[string]interface{}{
						"role":    "user",
						"content": "from params",
					},
				},
			},
		},
	}

	result := buildBatchRequestItems(requests)
	assert.Len(t, result, 1)
	assert.Len(t, result[0].Request.Contents, 1)
	assert.Equal(t, "from body", result[0].Request.Contents[0].Parts[0].Text)
}

// TestBuildBatchRequestItems_ParamsFallback tests that Params is used when Body is nil
func TestBuildBatchRequestItems_ParamsFallback(t *testing.T) {
	requests := []schemas.BatchRequestItem{
		{
			CustomID: "test",
			Body:     nil,
			Params: map[string]interface{}{
				"messages": []interface{}{
					map[string]interface{}{
						"role":    "user",
						"content": "from params",
					},
				},
			},
		},
	}

	result := buildBatchRequestItems(requests)
	assert.Len(t, result, 1)
	assert.Len(t, result[0].Request.Contents, 1)
	assert.Equal(t, "from params", result[0].Request.Contents[0].Parts[0].Text)
}
