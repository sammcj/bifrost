/**
 * Bifrost WASM Plugin for TypeScript/AssemblyScript
 *
 * This plugin demonstrates the proper structure for parsing inputs,
 * building responses, and handling context - similar to Go plugin patterns.
 *
 * Build with: npm run build
 */

import { JSON } from 'json-as'

// Memory management exports
import { free as _free, malloc as _malloc, readString, writeString } from './memory'

// Type definitions
import {
  HTTPInterceptInput,
  HTTPInterceptOutput,
  PostHookInput,
  PostHookOutput,
  PreHookInput,
  PreHookOutput
} from './types'

// =============================================================================
// Re-export memory functions for WASM host
// =============================================================================

export function malloc(size: u32): u32 {
  return _malloc(size)
}

export function free(ptr: u32): void {
  _free(ptr)
}

// =============================================================================
// Plugin Configuration
// =============================================================================

// Plugin configuration storage
let pluginConfig: string = ''

// =============================================================================
// Exported Plugin Functions
// =============================================================================

// Return the plugin name
export function get_name(): u64 {
  return writeString('hello-world-wasm-typescript')
}

// Initialize the plugin with config
// Returns 0 on success, non-zero on error
export function init(configPtr: u32, configLen: u32): i32 {
  // Parse and store configuration
  pluginConfig = readString(configPtr, configLen)

  // Validate configuration if needed
  // For this example, we just accept any config

  return 0 // Success
}

/**
 * HTTP transport intercept
 * Called at the HTTP layer before request enters Bifrost core.
 * Can modify headers, query params, or short-circuit with a response.
 */
export function http_intercept(inputPtr: u32, inputLen: u32): u64 {
  // Parse input using json-as
  const inputJson = readString(inputPtr, inputLen)
  const input = JSON.parse<HTTPInterceptInput>(inputJson)

  // Create output with context preserved
  const output = new HTTPInterceptOutput()
  output.context = input.context
  output.has_response = false

  // Example: Short-circuit health check endpoint
  // Uncomment to test:
  /*
  if (input.request.path === '/health') {
    output.has_response = true
    output.response = new HTTPResponse()
    output.response!.status_code = 200
    output.response!.body = 'eyJzdGF0dXMiOiJvayJ9' // base64 of {"status":"ok"}
    return writeString(JSON.stringify(output))
  }
  */

  return writeString(JSON.stringify(output))
}

/**
 * Pre-request hook
 * Called before request is sent to the provider.
 * Can modify the request or short-circuit with a response/error.
 */
export function pre_hook(inputPtr: u32, inputLen: u32): u64 {
  // Parse input using json-as
  const inputJson = readString(inputPtr, inputLen)
  const input = JSON.parse<PreHookInput>(inputJson)

  // Create output with context preserved
  const output = new PreHookOutput()
  output.context = input.context

  // Add custom values to context for tracking in post_hook
  output.context.plugin_processed = 'true'
  output.context.plugin_name = 'hello-world-wasm-typescript'

  // Get provider and model from request
  let provider = input.request.provider
  let model = input.request.model
  if (input.request.chat_request !== null) {
    provider = input.request.chat_request!.provider
    model = input.request.chat_request!.model
  }

  // Example: Short-circuit with mock response for specific model
  // Uncomment to test:
  /*
  if (model === 'mock-model') {
    output.has_short_circuit = true
    output.short_circuit = new PluginShortCircuit()

    const mockResponse = new BifrostResponse()
    mockResponse.chat_response = new BifrostChatResponse()
    mockResponse.chat_response!.id = 'mock-' + input.context.request_id
    mockResponse.chat_response!.model = 'mock-model'

    const choice = new ResponseChoice()
    choice.message.role = 'assistant'
    choice.message.content = 'This is a mock response from the WASM plugin!'
    choice.finish_reason = 'stop'
    mockResponse.chat_response!.choices.push(choice)

    mockResponse.chat_response!.usage.prompt_tokens = 10
    mockResponse.chat_response!.usage.completion_tokens = 15
    mockResponse.chat_response!.usage.total_tokens = 25

    output.short_circuit!.response = mockResponse
    return writeString(JSON.stringify(output))
  }
  */

  // Example: Short-circuit with rate limit error
  // Uncomment to test:
  /*
  if (shouldRateLimit(input.context.request_id)) {
    output.has_short_circuit = true
    output.short_circuit = new PluginShortCircuit()

    const error = new BifrostError()
    error.error.message = 'Rate limit exceeded'
    error.error.type = 'rate_limit'
    error.error.code = '429'
    error.status_code = 429

    output.short_circuit!.error = error
    return writeString(JSON.stringify(output))
  }
  */

  // Pass through - null request means use original
  return writeString(JSON.stringify(output))
}

/**
 * Post-response hook
 * Called after response is received from provider.
 * Can modify the response or error.
 */
export function post_hook(inputPtr: u32, inputLen: u32): u64 {
  // Parse input using json-as
  const inputJson = readString(inputPtr, inputLen)
  const input = JSON.parse<PostHookInput>(inputJson)

  // Create output with context preserved
  const output = new PostHookOutput()
  output.context = input.context

  // Check if our plugin processed this request
  const pluginProcessed = input.context.plugin_processed

  // Add completion marker if plugin was involved
  if (pluginProcessed === 'true') {
    output.context.post_hook_completed = 'true'
  }

  // Handle error case
  if (input.has_error && input.error !== null) {
    output.has_error = true
    output.error = input.error

    // Example: Modify error message
    // Uncomment to test:
    /*
    const modifiedError = new BifrostError()
    modifiedError.error.message = input.error!.error.message + ' (processed by WASM plugin)'
    modifiedError.error.type = input.error!.error.type
    modifiedError.error.code = input.error!.error.code
    modifiedError.status_code = input.error!.status_code
    output.error = modifiedError
    */

    return writeString(JSON.stringify(output))
  }

  // Handle success case - pass through response
  if (input.response !== null) {
    output.response = input.response

    // Example: Modify response model name
    // Uncomment to test:
    /*
    if (input.response!.chat_response !== null) {
      const modifiedResponse = new BifrostResponse()
      modifiedResponse.chat_response = input.response!.chat_response
      modifiedResponse.chat_response!.model += ' (via wasm-ts)'
      output.response = modifiedResponse
    }
    */
  }

  return writeString(JSON.stringify(output))
}

/**
 * Cleanup resources
 * Called when plugin is being unloaded.
 * Returns 0 on success, non-zero on error
 */
export function cleanup(): i32 {
  // Clear any stored configuration
  pluginConfig = ''

  // Perform any necessary cleanup
  return 0 // Success
}

// =============================================================================
// Helper Functions
// =============================================================================

// Example rate limit check function
function shouldRateLimit(_requestId: string): bool {
  // Implement your rate limiting logic here
  return false
}
