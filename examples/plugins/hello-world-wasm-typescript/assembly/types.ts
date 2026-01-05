/**
 * Type definitions for Bifrost WASM plugins.
 * These structures mirror the Go SDK types for interoperability.
 */

import { JSON } from 'json-as'

// =============================================================================
// Context Structure
// =============================================================================

/**
 * BifrostContext holds request-scoped values passed between hooks.
 * Common keys include:
 * - request_id: Unique identifier for the request
 * - Custom plugin values can be added and will be persisted across hooks
 */
@json
export class BifrostContext {
  request_id: string = ''

  // Custom values for plugin use (add more as needed)
  plugin_processed: string = ''
  plugin_name: string = ''
  post_hook_completed: string = ''
}

// =============================================================================
// HTTP Transport Structures
// =============================================================================

/**
 * HTTPRequest represents an incoming HTTP request at the transport layer.
 * Body is base64-encoded.
 */
@json
export class HTTPRequest {
  method: string = ''
  path: string = ''
  body: string = '' // base64 encoded
  headers: Map<string, string> = new Map<string, string>()
  query: Map<string, string> = new Map<string, string>()
}

/**
 * HTTPResponse represents an HTTP response to return.
 */
@json
export class HTTPResponse {
  status_code: i32 = 200
  body: string = '' // base64 encoded
  headers: Map<string, string> = new Map<string, string>()
}

/**
 * HTTPInterceptInput is the input for http_intercept hook.
 */
@json
export class HTTPInterceptInput {
  context: BifrostContext = new BifrostContext()
  request: HTTPRequest = new HTTPRequest()
}

/**
 * HTTPInterceptOutput is the output for http_intercept hook.
 */
@json
export class HTTPInterceptOutput {
  context: BifrostContext = new BifrostContext()
  request: HTTPRequest | null = null
  response: HTTPResponse | null = null
  has_response: bool = false
  error: string = ''
}

// =============================================================================
// Chat Completion Structures (BifrostRequest)
// =============================================================================

/**
 * ChatMessage represents a message in the conversation.
 */
@json
export class ChatMessage {
  role: string = ''  // "user", "assistant", "system", "tool"
  content: string = ''
  name: string = ''
  tool_call_id: string = ''
}

/**
 * ChatParameters contains optional parameters for chat completion.
 */
@json
export class ChatParameters {
  temperature: f64 = 0
  max_completion_tokens: i32 = 0
  top_p: f64 = 0
}

/**
 * BifrostChatRequest represents a chat completion request.
 */
@json
export class BifrostChatRequest {
  provider: string = ''
  model: string = ''
  input: ChatMessage[] = []
  params: ChatParameters = new ChatParameters()
}

/**
 * BifrostRequest is the unified request structure.
 */
@json
export class BifrostRequest {
  chat_request: BifrostChatRequest | null = null

  // Direct fields for simpler request structures
  provider: string = ''
  model: string = ''
  input: ChatMessage[] = []
  params: ChatParameters | null = null
}

// =============================================================================
// Response Structures (BifrostResponse)
// =============================================================================

/**
 * LLMUsage contains token usage information.
 */
@json
export class LLMUsage {
  prompt_tokens: i32 = 0
  completion_tokens: i32 = 0
  total_tokens: i32 = 0
}

/**
 * ResponseChoice represents a single completion choice.
 */
@json
export class ResponseChoice {
  index: i32 = 0
  message: ChatMessage = new ChatMessage()
  finish_reason: string = 'stop'
}

/**
 * BifrostChatResponse represents a chat completion response.
 */
@json
export class BifrostChatResponse {
  id: string = ''
  model: string = ''
  choices: ResponseChoice[] = []
  usage: LLMUsage = new LLMUsage()
}

/**
 * BifrostResponse is the unified response structure.
 */
@json
export class BifrostResponse {
  chat_response: BifrostChatResponse | null = null
}

// =============================================================================
// Error Structure
// =============================================================================

/**
 * ErrorField contains the error details.
 */
@json
export class ErrorField {
  message: string = ''
  type: string = ''
  code: string = ''
}

/**
 * BifrostError represents an error response.
 */
@json
export class BifrostError {
  error: ErrorField = new ErrorField()
  status_code: i32 = 0
}

// =============================================================================
// Short Circuit Structure
// =============================================================================

/**
 * PluginShortCircuit allows plugins to short-circuit the request flow.
 */
@json
export class PluginShortCircuit {
  response: BifrostResponse | null = null
  error: BifrostError | null = null
}

// =============================================================================
// Hook Input/Output Structures
// =============================================================================

/**
 * PreHookInput is the input for pre_hook.
 */
@json
export class PreHookInput {
  context: BifrostContext = new BifrostContext()
  request: BifrostRequest = new BifrostRequest()
}

/**
 * PreHookOutput is the output for pre_hook.
 */
@json
export class PreHookOutput {
  context: BifrostContext = new BifrostContext()
  request: BifrostRequest | null = null
  short_circuit: PluginShortCircuit | null = null
  has_short_circuit: bool = false
  error: string = ''
}

/**
 * PostHookInput is the input for post_hook.
 */
@json
export class PostHookInput {
  context: BifrostContext = new BifrostContext()
  response: BifrostResponse | null = null
  error: BifrostError | null = null
  has_error: bool = false
}

/**
 * PostHookOutput is the output for post_hook.
 */
@json
export class PostHookOutput {
  context: BifrostContext = new BifrostContext()
  response: BifrostResponse | null = null
  error: BifrostError | null = null
  has_error: bool = false
  hook_error: string = ''
}
