// Package governance provides utility functions for the governance plugin
package governance

import (
	"context"

	"github.com/maximhq/bifrost/core/schemas"
)

// extractHeadersFromContext extracts governance headers from context (standalone version)
func extractHeadersFromContext(ctx context.Context) map[string]string {
	headers := make(map[string]string)

	// Extract governance headers using schemas.BifrostContextKey
	if teamID := getStringFromContext(ctx, schemas.BifrostContextKey("x-bf-team")); teamID != "" {
		headers["x-bf-team"] = teamID
	}
	if userID := getStringFromContext(ctx, schemas.BifrostContextKey("x-bf-user")); userID != "" {
		headers["x-bf-user"] = userID
	}
	if customerID := getStringFromContext(ctx, schemas.BifrostContextKey("x-bf-customer")); customerID != "" {
		headers["x-bf-customer"] = customerID
	}

	return headers
}

// getStringFromContext safely extracts a string value from context
func getStringFromContext(ctx context.Context, key any) string {
	if value := ctx.Value(key); value != nil {
		if str, ok := value.(string); ok {
			return str
		}
	}
	return ""
}
