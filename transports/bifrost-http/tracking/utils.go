package tracking

import (
	"context"
	"strings"

	"github.com/valyala/fasthttp"
)

// ConvertToBifrostContext converts a FastHTTP RequestCtx to a Bifrost context,
// copying all prometheus header values (x-bf-prom-*) to the new context.
func ConvertToBifrostContext(ctx *fasthttp.RequestCtx) *context.Context {
	bifrostCtx := context.Background()

	// Copy all prometheus header values to the new context
	ctx.Request.Header.VisitAll(func(key, value []byte) {
		keyStr := strings.ToLower(string(key))
		if strings.HasPrefix(keyStr, "x-bf-prom-") {
			labelName := strings.TrimPrefix(keyStr, "x-bf-prom-")
			bifrostCtx = context.WithValue(bifrostCtx, prometheusContextKey(labelName), string(value))
		}
	})

	return &bifrostCtx
}
