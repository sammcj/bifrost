//go:build !tinygo && !wasm

package schemas

import (
	"github.com/valyala/fasthttp"
)

// BifrostHTTPMiddleware is a middleware function for the Bifrost HTTP transport.
// It follows the standard pattern: receives the next handler and returns a new handler.
// Used internally for CORS, Auth, Tracing middleware. Plugins use HTTPTransportIntercept instead.
type BifrostHTTPMiddleware func(next fasthttp.RequestHandler) fasthttp.RequestHandler
