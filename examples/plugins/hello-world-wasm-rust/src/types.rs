//! Type definitions for Bifrost WASM plugins.
//! These structures mirror the Go SDK types for interoperability.

use serde::{Deserialize, Serialize};
use std::collections::HashMap;

// =============================================================================
// Context Structure
// =============================================================================

/// BifrostContext holds request-scoped values passed between hooks.
/// Common keys include:
/// - request_id: Unique identifier for the request
/// - Custom plugin values can be added and will be persisted across hooks
#[derive(Debug, Clone, Serialize, Deserialize, Default)]
pub struct BifrostContext {
    #[serde(default)]
    pub request_id: Option<String>,

    /// Custom values set by plugins
    #[serde(flatten)]
    pub values: HashMap<String, serde_json::Value>,
}

impl BifrostContext {
    pub fn new() -> Self {
        Self::default()
    }

    /// Set a custom value in the context
    pub fn set_value(&mut self, key: &str, value: impl Into<serde_json::Value>) {
        self.values.insert(key.to_string(), value.into());
    }

    /// Get a string value from the context
    pub fn get_string(&self, key: &str) -> Option<&str> {
        self.values.get(key).and_then(|v| v.as_str())
    }

    /// Get a boolean value from the context
    pub fn get_bool(&self, key: &str) -> Option<bool> {
        self.values.get(key).and_then(|v| v.as_bool())
    }
}

// =============================================================================
// HTTP Transport Structures
// =============================================================================

/// HTTPRequest represents an incoming HTTP request at the transport layer.
/// Body is base64-encoded.
#[derive(Debug, Clone, Serialize, Deserialize, Default)]
pub struct HTTPRequest {
    #[serde(default)]
    pub method: String,

    #[serde(default)]
    pub path: String,

    #[serde(default)]
    pub headers: HashMap<String, String>,

    #[serde(default)]
    pub query: HashMap<String, String>,

    /// Base64-encoded request body
    #[serde(default)]
    pub body: String,
}

/// HTTPResponse represents an HTTP response to return.
#[derive(Debug, Clone, Serialize, Deserialize, Default)]
pub struct HTTPResponse {
    #[serde(default)]
    pub status_code: i32,

    #[serde(default)]
    pub headers: HashMap<String, String>,

    /// Base64-encoded response body
    #[serde(default)]
    pub body: String,
}

/// HTTPInterceptInput is the input for http_intercept hook.
#[derive(Debug, Clone, Serialize, Deserialize, Default)]
pub struct HTTPInterceptInput {
    #[serde(default)]
    pub context: BifrostContext,

    #[serde(default)]
    pub request: HTTPRequest,
}

/// HTTPInterceptOutput is the output for http_intercept hook.
#[derive(Debug, Clone, Serialize, Deserialize, Default)]
pub struct HTTPInterceptOutput {
    pub context: BifrostContext,

    #[serde(default)]
    pub request: serde_json::Value,

    #[serde(skip_serializing_if = "Option::is_none")]
    pub response: Option<HTTPResponse>,

    #[serde(default)]
    pub has_response: bool,

    #[serde(default)]
    pub error: String,
}

// =============================================================================
// Chat Completion Structures (BifrostRequest)
// =============================================================================

/// ChatMessageRole represents the role of a message sender.
#[derive(Debug, Clone, Serialize, Deserialize, PartialEq)]
#[serde(rename_all = "lowercase")]
pub enum ChatMessageRole {
    User,
    Assistant,
    System,
    Tool,
    Developer,
}

impl Default for ChatMessageRole {
    fn default() -> Self {
        ChatMessageRole::User
    }
}

/// ChatMessageContent can be either a string or an array of content blocks.
#[derive(Debug, Clone, Serialize, Deserialize)]
#[serde(untagged)]
pub enum ChatMessageContent {
    Text(String),
    Blocks(Vec<ChatContentBlock>),
}

impl Default for ChatMessageContent {
    fn default() -> Self {
        ChatMessageContent::Text(String::new())
    }
}

/// ChatContentBlock represents a content block in a message.
#[derive(Debug, Clone, Serialize, Deserialize, Default)]
pub struct ChatContentBlock {
    #[serde(rename = "type")]
    pub block_type: String,

    #[serde(skip_serializing_if = "Option::is_none")]
    pub text: Option<String>,

    #[serde(skip_serializing_if = "Option::is_none")]
    pub image_url: Option<ImageUrl>,
}

/// ImageUrl represents an image URL in a content block.
#[derive(Debug, Clone, Serialize, Deserialize, Default)]
pub struct ImageUrl {
    pub url: String,

    #[serde(skip_serializing_if = "Option::is_none")]
    pub detail: Option<String>,
}

/// ChatMessage represents a message in the conversation.
#[derive(Debug, Clone, Serialize, Deserialize, Default)]
pub struct ChatMessage {
    #[serde(default)]
    pub role: ChatMessageRole,

    #[serde(skip_serializing_if = "Option::is_none")]
    pub content: Option<ChatMessageContent>,

    #[serde(skip_serializing_if = "Option::is_none")]
    pub name: Option<String>,

    #[serde(skip_serializing_if = "Option::is_none")]
    pub tool_call_id: Option<String>,

    #[serde(skip_serializing_if = "Option::is_none")]
    pub tool_calls: Option<Vec<ToolCall>>,
}

/// ToolCall represents a tool call made by the assistant.
#[derive(Debug, Clone, Serialize, Deserialize, Default)]
pub struct ToolCall {
    #[serde(default)]
    pub id: Option<String>,

    #[serde(rename = "type", default)]
    pub call_type: Option<String>,

    #[serde(default)]
    pub function: ToolCallFunction,
}

/// ToolCallFunction represents the function being called.
#[derive(Debug, Clone, Serialize, Deserialize, Default)]
pub struct ToolCallFunction {
    #[serde(default)]
    pub name: Option<String>,

    #[serde(default)]
    pub arguments: String,
}

/// ChatParameters contains optional parameters for chat completion.
#[derive(Debug, Clone, Serialize, Deserialize, Default)]
pub struct ChatParameters {
    #[serde(skip_serializing_if = "Option::is_none")]
    pub temperature: Option<f64>,

    #[serde(skip_serializing_if = "Option::is_none")]
    pub max_completion_tokens: Option<i32>,

    #[serde(skip_serializing_if = "Option::is_none")]
    pub top_p: Option<f64>,

    #[serde(skip_serializing_if = "Option::is_none")]
    pub frequency_penalty: Option<f64>,

    #[serde(skip_serializing_if = "Option::is_none")]
    pub presence_penalty: Option<f64>,

    #[serde(skip_serializing_if = "Option::is_none")]
    pub stop: Option<Vec<String>>,

    #[serde(skip_serializing_if = "Option::is_none")]
    pub tools: Option<Vec<ChatTool>>,

    /// Catch-all for additional parameters
    #[serde(flatten)]
    pub extra: HashMap<String, serde_json::Value>,
}

/// ChatTool represents a tool definition.
#[derive(Debug, Clone, Serialize, Deserialize, Default)]
pub struct ChatTool {
    #[serde(rename = "type")]
    pub tool_type: String,

    #[serde(skip_serializing_if = "Option::is_none")]
    pub function: Option<ChatToolFunction>,
}

/// ChatToolFunction represents a function definition.
#[derive(Debug, Clone, Serialize, Deserialize, Default)]
pub struct ChatToolFunction {
    pub name: String,

    #[serde(skip_serializing_if = "Option::is_none")]
    pub description: Option<String>,

    #[serde(skip_serializing_if = "Option::is_none")]
    pub parameters: Option<serde_json::Value>,
}

/// BifrostChatRequest represents a chat completion request.
#[derive(Debug, Clone, Serialize, Deserialize, Default)]
pub struct BifrostChatRequest {
    #[serde(default)]
    pub provider: String,

    #[serde(default)]
    pub model: String,

    #[serde(default)]
    pub input: Vec<ChatMessage>,

    #[serde(skip_serializing_if = "Option::is_none")]
    pub params: Option<ChatParameters>,

    #[serde(skip_serializing_if = "Option::is_none")]
    pub fallbacks: Option<Vec<Fallback>>,
}

/// Fallback represents a fallback provider/model.
#[derive(Debug, Clone, Serialize, Deserialize, Default)]
pub struct Fallback {
    pub provider: String,
    pub model: String,
}

/// BifrostRequest is the unified request structure.
/// Only one of the request types should be present.
#[derive(Debug, Clone, Serialize, Deserialize, Default)]
pub struct BifrostRequest {
    #[serde(skip_serializing_if = "Option::is_none")]
    pub chat_request: Option<BifrostChatRequest>,

    // Add other request types as needed
    #[serde(flatten)]
    pub extra: HashMap<String, serde_json::Value>,
}

impl BifrostRequest {
    /// Get provider and model from the request
    pub fn get_provider_model(&self) -> (String, String) {
        if let Some(ref chat) = self.chat_request {
            return (chat.provider.clone(), chat.model.clone());
        }
        (String::new(), String::new())
    }
}

// =============================================================================
// Response Structures (BifrostResponse)
// =============================================================================

/// LLMUsage contains token usage information.
#[derive(Debug, Clone, Serialize, Deserialize, Default)]
pub struct LLMUsage {
    #[serde(default)]
    pub prompt_tokens: i32,

    #[serde(default)]
    pub completion_tokens: i32,

    #[serde(default)]
    pub total_tokens: i32,

    #[serde(skip_serializing_if = "Option::is_none")]
    pub prompt_tokens_details: Option<serde_json::Value>,

    #[serde(skip_serializing_if = "Option::is_none")]
    pub completion_tokens_details: Option<serde_json::Value>,
}

/// ResponseChoice represents a single completion choice.
#[derive(Debug, Clone, Serialize, Deserialize, Default)]
pub struct ResponseChoice {
    #[serde(default)]
    pub index: i32,

    #[serde(skip_serializing_if = "Option::is_none")]
    pub message: Option<ChatMessage>,

    #[serde(skip_serializing_if = "Option::is_none")]
    pub delta: Option<ChatMessage>,

    #[serde(skip_serializing_if = "Option::is_none")]
    pub finish_reason: Option<String>,

    #[serde(skip_serializing_if = "Option::is_none")]
    pub logprobs: Option<serde_json::Value>,
}

/// BifrostChatResponse represents a chat completion response.
#[derive(Debug, Clone, Serialize, Deserialize, Default)]
pub struct BifrostChatResponse {
    #[serde(default)]
    pub id: String,

    #[serde(default)]
    pub model: String,

    #[serde(default)]
    pub choices: Vec<ResponseChoice>,

    #[serde(skip_serializing_if = "Option::is_none")]
    pub usage: Option<LLMUsage>,

    #[serde(skip_serializing_if = "Option::is_none")]
    pub created: Option<i64>,

    #[serde(skip_serializing_if = "Option::is_none")]
    pub object: Option<String>,

    #[serde(skip_serializing_if = "Option::is_none")]
    pub system_fingerprint: Option<String>,
}

/// BifrostResponse is the unified response structure.
#[derive(Debug, Clone, Serialize, Deserialize, Default)]
pub struct BifrostResponse {
    #[serde(skip_serializing_if = "Option::is_none")]
    pub chat_response: Option<BifrostChatResponse>,

    #[serde(flatten)]
    pub extra: HashMap<String, serde_json::Value>,
}

// =============================================================================
// Error Structure
// =============================================================================

/// ErrorField contains the error details.
#[derive(Debug, Clone, Serialize, Deserialize, Default)]
pub struct ErrorField {
    #[serde(default)]
    pub message: String,

    #[serde(skip_serializing_if = "Option::is_none", rename = "type")]
    pub error_type: Option<String>,

    #[serde(skip_serializing_if = "Option::is_none")]
    pub code: Option<String>,

    #[serde(skip_serializing_if = "Option::is_none")]
    pub param: Option<String>,
}

/// BifrostError represents an error response.
#[derive(Debug, Clone, Serialize, Deserialize, Default)]
pub struct BifrostError {
    #[serde(default)]
    pub error: ErrorField,

    #[serde(skip_serializing_if = "Option::is_none")]
    pub status_code: Option<i32>,

    #[serde(skip_serializing_if = "Option::is_none")]
    pub allow_fallbacks: Option<bool>,
}

impl BifrostError {
    /// Create a new error with a message
    pub fn new(message: &str) -> Self {
        Self {
            error: ErrorField {
                message: message.to_string(),
                ..Default::default()
            },
            ..Default::default()
        }
    }

    /// Set the error type
    pub fn with_type(mut self, error_type: &str) -> Self {
        self.error.error_type = Some(error_type.to_string());
        self
    }

    /// Set the error code
    pub fn with_code(mut self, code: &str) -> Self {
        self.error.code = Some(code.to_string());
        self
    }

    /// Set the status code
    pub fn with_status(mut self, status: i32) -> Self {
        self.status_code = Some(status);
        self
    }
}

// =============================================================================
// Short Circuit Structure
// =============================================================================

/// PluginShortCircuit allows plugins to short-circuit the request flow.
#[derive(Debug, Clone, Serialize, Deserialize, Default)]
pub struct PluginShortCircuit {
    #[serde(skip_serializing_if = "Option::is_none")]
    pub response: Option<BifrostResponse>,

    #[serde(skip_serializing_if = "Option::is_none")]
    pub error: Option<BifrostError>,
}

// =============================================================================
// Hook Input/Output Structures
// =============================================================================

/// PreHookInput is the input for pre_hook.
#[derive(Debug, Clone, Serialize, Deserialize, Default)]
pub struct PreHookInput {
    #[serde(default)]
    pub context: BifrostContext,

    #[serde(default)]
    pub request: serde_json::Value,
}

impl PreHookInput {
    /// Parse the request as a BifrostRequest
    pub fn parse_request(&self) -> Option<BifrostRequest> {
        serde_json::from_value(self.request.clone()).ok()
    }

    /// Get provider and model from the request
    pub fn get_provider_model(&self) -> (String, String) {
        if let Some(req) = self.parse_request() {
            return req.get_provider_model();
        }
        // Try direct access for simpler structures
        let provider = self.request.get("provider")
            .and_then(|v| v.as_str())
            .unwrap_or("")
            .to_string();
        let model = self.request.get("model")
            .and_then(|v| v.as_str())
            .unwrap_or("")
            .to_string();
        (provider, model)
    }
}

/// PreHookOutput is the output for pre_hook.
#[derive(Debug, Clone, Serialize, Deserialize, Default)]
pub struct PreHookOutput {
    pub context: BifrostContext,

    #[serde(default)]
    pub request: serde_json::Value,

    #[serde(skip_serializing_if = "Option::is_none")]
    pub short_circuit: Option<PluginShortCircuit>,

    #[serde(default)]
    pub has_short_circuit: bool,

    #[serde(default)]
    pub error: String,
}

/// PostHookInput is the input for post_hook.
#[derive(Debug, Clone, Serialize, Deserialize, Default)]
pub struct PostHookInput {
    #[serde(default)]
    pub context: BifrostContext,

    #[serde(default)]
    pub response: serde_json::Value,

    #[serde(default)]
    pub error: serde_json::Value,

    #[serde(default)]
    pub has_error: bool,
}

impl PostHookInput {
    /// Parse the response as a BifrostResponse
    pub fn parse_response(&self) -> Option<BifrostResponse> {
        serde_json::from_value(self.response.clone()).ok()
    }

    /// Parse the error as a BifrostError
    pub fn parse_error(&self) -> Option<BifrostError> {
        if self.has_error {
            serde_json::from_value(self.error.clone()).ok()
        } else {
            None
        }
    }
}

/// PostHookOutput is the output for post_hook.
#[derive(Debug, Clone, Serialize, Deserialize, Default)]
pub struct PostHookOutput {
    pub context: BifrostContext,

    #[serde(default)]
    pub response: serde_json::Value,

    #[serde(default)]
    pub error: serde_json::Value,

    #[serde(default)]
    pub has_error: bool,

    #[serde(default)]
    pub hook_error: String,
}

// =============================================================================
// Plugin Configuration
// =============================================================================

/// Plugin configuration (customize as needed)
#[derive(Debug, Clone, Serialize, Deserialize, Default)]
pub struct PluginConfig {
    #[serde(flatten)]
    pub values: HashMap<String, serde_json::Value>,
}

// =============================================================================
// Tests
// =============================================================================

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_context_serialization() {
        let mut ctx = BifrostContext::new();
        ctx.request_id = Some("test-123".to_string());
        ctx.set_value("custom_key", "custom_value");
        
        let json = serde_json::to_string(&ctx).unwrap();
        assert!(json.contains("request_id"));
        assert!(json.contains("custom_key"));
    }

    #[test]
    fn test_chat_message() {
        let msg = ChatMessage {
            role: ChatMessageRole::User,
            content: Some(ChatMessageContent::Text("Hello!".to_string())),
            ..Default::default()
        };
        
        let json = serde_json::to_string(&msg).unwrap();
        assert!(json.contains("user"));
        assert!(json.contains("Hello!"));
    }

    #[test]
    fn test_bifrost_error() {
        let error = BifrostError::new("Test error")
            .with_type("test_type")
            .with_code("500")
            .with_status(500);
        
        let json = serde_json::to_string(&error).unwrap();
        assert!(json.contains("Test error"));
        assert!(json.contains("test_type"));
    }

    #[test]
    fn test_pre_hook_input_parsing() {
        let json = r#"{
            "context": {"request_id": "test-123"},
            "request": {"provider": "openai", "model": "gpt-4"}
        }"#;
        
        let input: PreHookInput = serde_json::from_str(json).unwrap();
        assert_eq!(input.context.request_id, Some("test-123".to_string()));
        
        let (provider, model) = input.get_provider_model();
        assert_eq!(provider, "openai");
        assert_eq!(model, "gpt-4");
    }
}
