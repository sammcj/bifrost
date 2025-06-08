// Package providers implements various LLM providers and their utility functions.
// This file contains common utility functions used across different provider implementations.
package providers

import (
	"context"
	"fmt"
	"net/url"
	"reflect"
	"regexp"
	"strings"
	"sync"

	"github.com/goccy/go-json"
	schemas "github.com/maximhq/bifrost/core/schemas"
	"github.com/valyala/fasthttp"
	"github.com/valyala/fasthttp/fasthttpproxy"

	"maps"
)

// bifrostResponsePool provides a pool for Bifrost response objects.
var bifrostResponsePool = sync.Pool{
	New: func() interface{} {
		return &schemas.BifrostResponse{}
	},
}

// dataURIRegex is a precompiled regex for matching data URI format patterns.
// It matches patterns like: data:image/png;base64,iVBORw0KGgo...
var dataURIRegex = regexp.MustCompile(`^data:([^;]+);base64,(.*)$`)

// acquireBifrostResponse gets a Bifrost response from the pool and resets it.
func acquireBifrostResponse() *schemas.BifrostResponse {
	resp := bifrostResponsePool.Get().(*schemas.BifrostResponse)
	*resp = schemas.BifrostResponse{} // Reset the struct
	return resp
}

// releaseBifrostResponse returns a Bifrost response to the pool.
func releaseBifrostResponse(resp *schemas.BifrostResponse) {
	if resp != nil {
		bifrostResponsePool.Put(resp)
	}
}

// mergeConfig merges a default configuration map with custom parameters.
// It creates a new map containing all default values, then overrides them with any custom values.
// Returns a new map containing the merged configuration.
func mergeConfig(defaultConfig map[string]interface{}, customParams map[string]interface{}) map[string]interface{} {
	merged := make(map[string]interface{})

	// Copy default config
	for k, v := range defaultConfig {
		merged[k] = v
	}

	// Override with custom parameters
	for k, v := range customParams {
		merged[k] = v
	}

	return merged
}

// prepareParams converts ModelParameters into a flat map of parameters.
// It handles both standard fields and extra parameters, using reflection to process
// the struct fields and their JSON tags.
// Returns a map containing all parameters ready for use in API requests.
func prepareParams(params *schemas.ModelParameters) map[string]interface{} {
	flatParams := make(map[string]interface{})

	// Return empty map if params is nil
	if params == nil {
		return flatParams
	}

	// Use reflection to get the type and value of params
	val := reflect.ValueOf(params).Elem()
	typ := val.Type()

	// Iterate through all fields
	for i := range val.NumField() {
		field := val.Field(i)
		fieldType := typ.Field(i)

		// Skip the ExtraParams field as it's handled separately
		if fieldType.Name == "ExtraParams" {
			continue
		}

		// Get the JSON tag name
		jsonTag := fieldType.Tag.Get("json")
		if jsonTag == "" || jsonTag == "-" {
			continue
		}

		// Strip out ,omitempty and others from the tag
		jsonTag = strings.Split(jsonTag, ",")[0]

		// Handle pointer fields
		if field.Kind() == reflect.Ptr && !field.IsNil() {
			flatParams[jsonTag] = field.Elem().Interface()
		}
	}

	// Handle ExtraParams
	maps.Copy(flatParams, params.ExtraParams)

	return flatParams
}

// IMPORTANT: This function does NOT truly cancel the underlying fasthttp network request if the
// context is done. The fasthttp client call will continue in its goroutine until it completes
// or times out based on its own settings. This function merely stops *waiting* for the
// fasthttp call and returns an error related to the context.
func makeRequestWithContext(ctx context.Context, client *fasthttp.Client, req *fasthttp.Request, resp *fasthttp.Response) *schemas.BifrostError {
	errChan := make(chan error, 1)

	go func() {
		// client.Do is a blocking call.
		// It will send an error (or nil for success) to errChan when it completes.
		errChan <- client.Do(req, resp)
	}()

	select {
	case <-ctx.Done():
		// Context was cancelled (e.g., deadline exceeded or manual cancellation).
		// Return a BifrostError indicating this.
		return &schemas.BifrostError{
			IsBifrostError: true,
			Error: schemas.ErrorField{
				Type:    StrPtr(schemas.RequestCancelled),
				Message: fmt.Sprintf("Request cancelled or timed out by context: %v", ctx.Err()),
				Error:   ctx.Err(),
			},
		}
	case err := <-errChan:
		// The fasthttp.Do call completed.
		if err != nil {
			// The HTTP request itself failed (e.g., connection error, fasthttp timeout).
			return &schemas.BifrostError{
				IsBifrostError: false,
				Error: schemas.ErrorField{
					Message: schemas.ErrProviderRequest,
					Error:   err,
				},
			}
		}
		// HTTP request was successful from fasthttp's perspective (err is nil).
		// The caller should check resp.StatusCode() for HTTP-level errors (4xx, 5xx).
		return nil
	}
}

// configureProxy sets up a proxy for the fasthttp client based on the provided configuration.
// It supports HTTP, SOCKS5, and environment-based proxy configurations.
// Returns the configured client or the original client if proxy configuration is invalid.
func configureProxy(client *fasthttp.Client, proxyConfig *schemas.ProxyConfig, logger schemas.Logger) *fasthttp.Client {
	if proxyConfig == nil {
		return client
	}

	var dialFunc fasthttp.DialFunc

	// Create the appropriate proxy based on type
	switch proxyConfig.Type {
	case schemas.NoProxy:
		return client
	case schemas.HttpProxy:
		if proxyConfig.URL == "" {
			logger.Warn("Warning: HTTP proxy URL is required for setting up proxy")
			return client
		}
		dialFunc = fasthttpproxy.FasthttpHTTPDialer(proxyConfig.URL)
	case schemas.Socks5Proxy:
		if proxyConfig.URL == "" {
			logger.Warn("Warning: SOCKS5 proxy URL is required for setting up proxy")
			return client
		}
		proxyUrl := proxyConfig.URL
		// Add authentication if provided
		if proxyConfig.Username != "" && proxyConfig.Password != "" {
			parsedURL, err := url.Parse(proxyConfig.URL)
			if err != nil {
				logger.Warn("Invalid proxy configuration: invalid SOCKS5 proxy URL")
				return client
			}
			// Set user and password in the parsed URL
			parsedURL.User = url.UserPassword(proxyConfig.Username, proxyConfig.Password)
			proxyUrl = parsedURL.String()
		}
		dialFunc = fasthttpproxy.FasthttpSocksDialer(proxyUrl)
	case schemas.EnvProxy:
		// Use environment variables for proxy configuration
		dialFunc = fasthttpproxy.FasthttpProxyHTTPDialer()
	default:
		logger.Warn(fmt.Sprintf("Invalid proxy configuration: unsupported proxy type: %s", proxyConfig.Type))
		return client
	}

	if dialFunc != nil {
		client.Dial = dialFunc
	}

	return client
}

// handleProviderAPIError processes error responses from provider APIs.
// It attempts to unmarshal the error response and returns a BifrostError
// with the appropriate status code and error information.
func handleProviderAPIError(resp *fasthttp.Response, errorResp any) *schemas.BifrostError {
	statusCode := resp.StatusCode()

	if err := json.Unmarshal(resp.Body(), &errorResp); err != nil {
		return &schemas.BifrostError{
			IsBifrostError: true,
			StatusCode:     &statusCode,
			Error: schemas.ErrorField{
				Message: schemas.ErrProviderResponseUnmarshal,
				Error:   err,
			},
		}
	}

	return &schemas.BifrostError{
		IsBifrostError: false,
		StatusCode:     &statusCode,
		Error:          schemas.ErrorField{},
	}
}

// handleProviderResponse handles common response parsing logic for provider responses.
// It attempts to parse the response body into the provided response type
// and returns either the parsed response or a BifrostError if parsing fails.
func handleProviderResponse[T any](responseBody []byte, response *T) (interface{}, *schemas.BifrostError) {
	var rawResponse interface{}

	var wg sync.WaitGroup
	var structuredErr, rawErr error

	wg.Add(2)
	go func() {
		defer wg.Done()
		structuredErr = json.Unmarshal(responseBody, response)
	}()
	go func() {
		defer wg.Done()
		rawErr = json.Unmarshal(responseBody, &rawResponse)
	}()
	wg.Wait()

	if structuredErr != nil {
		return nil, &schemas.BifrostError{
			IsBifrostError: true,
			Error: schemas.ErrorField{
				Message: schemas.ErrProviderDecodeStructured,
				Error:   structuredErr,
			},
		}
	}

	if rawErr != nil {
		return nil, &schemas.BifrostError{
			IsBifrostError: true,
			Error: schemas.ErrorField{
				Message: schemas.ErrProviderDecodeRaw,
				Error:   rawErr,
			},
		}
	}

	return rawResponse, nil
}

// getRoleFromMessage extracts and validates the role from a message map.
func getRoleFromMessage(msg map[string]interface{}) (schemas.ModelChatMessageRole, bool) {
	roleVal, exists := msg["role"]
	if !exists {
		return "", false // Role key doesn't exist
	}

	// Try direct assertion to ModelChatMessageRole
	roleAsModelType, ok := roleVal.(schemas.ModelChatMessageRole)
	if ok {
		return roleAsModelType, true
	}

	// Try assertion to string and then convert
	roleAsString, okStr := roleVal.(string)
	if okStr {
		return schemas.ModelChatMessageRole(roleAsString), true
	}

	return "", false // Role is of an unexpected or invalid type
}

// float64Ptr creates a pointer to a float64 value.
// This is a helper function for creating pointers to float64 values.
func float64Ptr(f float64) *float64 {
	return &f
}

// StrPtr creates a pointer to a string value.
// This is a helper function for creating pointers to string values.
func StrPtr(s string) *string {
	return &s
}

// coalesceString returns the string value of a pointer to a string, or an empty string if the pointer is nil.
// This is a helper function for safely handling pointer-to-string values.
func coalesceString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

// normalizeMediaType converts short media types to full media types
// e.g., "jpeg" -> "image/jpeg", "png" -> "image/png"
func normalizeMediaType(mediaType string) string {
	if mediaType == "" {
		return "image/jpeg" // default
	}

	// If it already has the image/ prefix, return as is
	if strings.HasPrefix(mediaType, "image/") {
		return mediaType
	}

	// Add image/ prefix for common formats
	switch strings.ToLower(mediaType) {
	case "jpeg", "jpg":
		return "image/jpeg"
	case "png":
		return "image/png"
	case "gif":
		return "image/gif"
	case "webp":
		return "image/webp"
	case "bmp":
		return "image/bmp"
	case "svg":
		return "image/svg+xml"
	default:
		return "image/" + mediaType
	}
}

// Normalize handles type inference and media type normalization for image content.
// It automatically detects content type from URL patterns and normalizes media types.
//
// NOTE: This function is called internally by the Bifrost system - you do not need to call it yourself.
// It is automatically invoked when processing image content in requests.
func normalizeImageContent(ic *schemas.ImageContent) {
	if ic == nil {
		return
	}

	// Handle unknown/empty type - try to infer from URL
	if ic.Type == "" && ic.URL != "" {
		if dataURIRegex.MatchString(ic.URL) {
			// Looks like base64 data URI
			ic.Type = schemas.ImageContentTypeBase64
		} else if strings.HasPrefix(ic.URL, "http://") || strings.HasPrefix(ic.URL, "https://") {
			// Looks like a regular URL
			ic.Type = schemas.ImageContentTypeURL
		} else {
			// Assume it's raw base64 data
			ic.Type = schemas.ImageContentTypeBase64
		}
	}

	// Normalize MediaType if provided
	if ic.MediaType != nil && *ic.MediaType != "" {
		normalizedMediaType := normalizeMediaType(*ic.MediaType)
		ic.MediaType = &normalizedMediaType
	}

}

// FormatDataURL modifies the image content struct in place to format data URL for base64 image content.
//
// NOTE: This function is called internally by the Bifrost system - you do not need to call it yourself.
// It is automatically invoked when processing image content for different providers.
//
// Parameters:
//   - includePrefix: Whether to include the "data:mediatype;base64," prefix
//   - true: URL will be in full data URI format (data:image/png;base64,iVBORw0KGgo...)
//   - false: URL will contain only the base64 data (iVBORw0KGgo...)
func FormatImageContent(imageContent *schemas.ImageContent, includePrefix bool) *schemas.ImageContent {
	if imageContent == nil {
		return nil
	}

	newImageContent := *imageContent

	normalizeImageContent(&newImageContent)

	if newImageContent.Type != schemas.ImageContentTypeBase64 {
		return &newImageContent
	}

	var finalMediaType string
	var base64Data string

	// Extract base64 data and media type from URL using precompiled regex
	if matches := dataURIRegex.FindStringSubmatch(newImageContent.URL); matches != nil {
		// URL already has data URI format
		existingMediaType := matches[1]
		base64Data = matches[2]

		// Determine final media type (prefer explicit MediaType field)
		if newImageContent.MediaType != nil && *newImageContent.MediaType != "" {
			finalMediaType = normalizeMediaType(*newImageContent.MediaType)
		} else {
			finalMediaType = normalizeMediaType(existingMediaType)
		}
	} else {
		// URL contains raw base64 data (no data URI prefix)
		base64Data = newImageContent.URL

		// Determine media type
		if newImageContent.MediaType != nil && *newImageContent.MediaType != "" {
			finalMediaType = normalizeMediaType(*newImageContent.MediaType)
		} else {
			finalMediaType = "image/jpeg" // default when no media type provided
		}
	}

	// Ensure MediaType field is always set with normalized value
	normalizedMediaType := finalMediaType
	newImageContent.MediaType = &normalizedMediaType

	// Set URL based on includePrefix preference
	if includePrefix {
		// Full data URI format
		newImageContent.URL = fmt.Sprintf("data:%s;base64,%s", finalMediaType, base64Data)
	} else {
		// Raw base64 data only
		newImageContent.URL = base64Data
	}

	return &newImageContent
}
