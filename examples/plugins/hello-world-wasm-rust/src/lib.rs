//! Bifrost WASM Plugin for Rust
//!
//! This plugin demonstrates the proper structure for parsing inputs,
//! building responses, and handling context - similar to Go plugin patterns.
//!
//! Build with: cargo build --release --target wasm32-unknown-unknown

mod memory;
mod types;

use memory::{read_string, write_string};
use types::*;

// Global configuration storage
static mut PLUGIN_CONFIG: Option<PluginConfig> = None;

// =============================================================================
// Exported Plugin Functions
// =============================================================================

/// Return the plugin name
#[no_mangle]
pub extern "C" fn get_name() -> u64 {
    write_string("hello-world-wasm-rust")
}

/// Initialize the plugin with config
/// Returns 0 on success, non-zero on error
#[no_mangle]
pub extern "C" fn init(config_ptr: u32, config_len: u32) -> i32 {
    let config_str = read_string(config_ptr, config_len);
    
    // Parse configuration
    let config: PluginConfig = if config_str.is_empty() {
        PluginConfig::default()
    } else {
        match serde_json::from_str(&config_str) {
            Ok(c) => c,
            Err(_) => return 1, // Config parse error
        }
    };
    
    // Store configuration
    unsafe {
        PLUGIN_CONFIG = Some(config);
    }
    
    0 // Success
}

/// HTTP transport intercept
/// Called at the HTTP layer before request enters Bifrost core.
/// Can modify headers, query params, or short-circuit with a response.
#[no_mangle]
pub extern "C" fn http_intercept(input_ptr: u32, input_len: u32) -> u64 {
    let input_str = read_string(input_ptr, input_len);
    
    // Parse input
    let input: HTTPInterceptInput = match serde_json::from_str(&input_str) {
        Ok(i) => i,
        Err(e) => {
            let output = HTTPInterceptOutput {
                error: format!("Failed to parse input: {}", e),
                ..Default::default()
            };
            return write_string(&serde_json::to_string(&output).unwrap_or_default());
        }
    };
    
    // Create output with context preserved
    let output = HTTPInterceptOutput {
        context: input.context,
        request: input.request,
        has_response: false,
        ..Default::default()
    };
    
    // Example: Short-circuit health check endpoint
    // Uncomment to test:
    /*
    if input.request.path == "/health" {
        output.has_response = true;
        output.response = Some(HTTPResponse {
            status_code: 200,
            headers: HashMap::new(),
            body: base64::encode(r#"{"status":"ok"}"#),
        });
        return write_string(&serde_json::to_string(&output).unwrap_or_default());
    }
    */
    
    // Pass through
    write_string(&serde_json::to_string(&output).unwrap_or_default())
}

/// Pre-request hook
/// Called before request is sent to the provider.
/// Can modify the request or short-circuit with a response/error.
#[no_mangle]
pub extern "C" fn pre_hook(input_ptr: u32, input_len: u32) -> u64 {
    let input_str = read_string(input_ptr, input_len);
    
    // Parse input
    let input: PreHookInput = match serde_json::from_str(&input_str) {
        Ok(i) => i,
        Err(e) => {
            let output = PreHookOutput {
                error: format!("Failed to parse input: {}", e),
                ..Default::default()
            };
            return write_string(&serde_json::to_string(&output).unwrap_or_default());
        }
    };
    
    // Create output with context preserved
    let mut output = PreHookOutput {
        context: input.context.clone(),
        request: input.request.clone(),
        has_short_circuit: false,
        ..Default::default()
    };
    
    // Add custom values to context for tracking
    output.context.set_value("plugin_processed", serde_json::json!(true));
    output.context.set_value("plugin_name", serde_json::json!("hello-world-wasm-rust"));
    
    // Get provider and model for potential modifications
    let (_provider, model) = input.get_provider_model();
    
    // Example: Short-circuit with mock response for specific model
    // Uncomment to test:
    /*
    if model == "mock-model" {
        output.has_short_circuit = true;
        
        let mock_response = BifrostResponse {
            chat_response: Some(BifrostChatResponse {
                id: format!("mock-{}", input.context.request_id.unwrap_or_default()),
                model: "mock-model".to_string(),
                choices: vec![ResponseChoice {
                    index: 0,
                    message: Some(ChatMessage {
                        role: ChatMessageRole::Assistant,
                        content: Some(ChatMessageContent::Text(
                            "This is a mock response from the Rust WASM plugin!".to_string()
                        )),
                        ..Default::default()
                    }),
                    finish_reason: Some("stop".to_string()),
                    ..Default::default()
                }],
                usage: Some(LLMUsage {
                    prompt_tokens: 10,
                    completion_tokens: 15,
                    total_tokens: 25,
                    ..Default::default()
                }),
                ..Default::default()
            }),
            ..Default::default()
        };
        
        output.short_circuit = Some(PluginShortCircuit {
            response: Some(mock_response),
            error: None,
        });
        
        return write_string(&serde_json::to_string(&output).unwrap_or_default());
    }
    */
    
    // Example: Short-circuit with rate limit error
    // Uncomment to test:
    /*
    if should_rate_limit(&input.context) {
        output.has_short_circuit = true;
        output.short_circuit = Some(PluginShortCircuit {
            response: None,
            error: Some(
                BifrostError::new("Rate limit exceeded")
                    .with_type("rate_limit")
                    .with_code("429")
                    .with_status(429)
            ),
        });
        return write_string(&serde_json::to_string(&output).unwrap_or_default());
    }
    */

    // Silence unused variable warning in example code
    let _ = model;
    
    // Pass through - empty request means use original
    write_string(&serde_json::to_string(&output).unwrap_or_default())
}

/// Post-response hook
/// Called after response is received from provider.
/// Can modify the response or error.
#[no_mangle]
pub extern "C" fn post_hook(input_ptr: u32, input_len: u32) -> u64 {
    let input_str = read_string(input_ptr, input_len);
    
    // Parse input
    let input: PostHookInput = match serde_json::from_str(&input_str) {
        Ok(i) => i,
        Err(e) => {
            let output = PostHookOutput {
                hook_error: format!("Failed to parse input: {}", e),
                ..Default::default()
            };
            return write_string(&serde_json::to_string(&output).unwrap_or_default());
        }
    };
    
    // Create output with context preserved
    let mut output = PostHookOutput {
        context: input.context.clone(),
        response: serde_json::json!({}),
        error: serde_json::json!({}),
        has_error: false,
        hook_error: String::new(),
    };
    
    // Check if our plugin processed this request
    let plugin_processed = input.context.get_bool("plugin_processed").unwrap_or(false);
    
    if plugin_processed {
        // Plugin was involved in pre_hook, add completion marker
        output.context.set_value("post_hook_completed", serde_json::json!(true));
    }
    
    // Handle error case
    if input.has_error {
        output.has_error = true;
        output.error = input.error.clone();
        
        // Example: Modify error message
        // Uncomment to test:
        /*
        if let Some(mut error) = input.parse_error() {
            error.error.message = format!("{} (processed by Rust WASM plugin)", error.error.message);
            output.error = serde_json::to_value(&error).unwrap_or_default();
        }
        */
        
        return write_string(&serde_json::to_string(&output).unwrap_or_default());
    }
    
    // Handle success case - pass through response
    output.response = input.response;
    
    // Example: Modify response
    // Uncomment to test:
    /*
    if let Some(mut response) = input.parse_response() {
        // Add custom metadata, modify model name, etc.
        if let Some(ref mut chat) = response.chat_response {
            // Add a marker to the model name
            chat.model = format!("{} (via rust-wasm)", chat.model);
        }
        output.response = serde_json::to_value(&response).unwrap_or_default();
    }
    */
    
    write_string(&serde_json::to_string(&output).unwrap_or_default())
}

/// Cleanup resources
/// Called when plugin is being unloaded.
/// Returns 0 on success, non-zero on error
#[no_mangle]
pub extern "C" fn cleanup() -> i32 {
    // Clear stored configuration
    unsafe {
        PLUGIN_CONFIG = None;
    }
    
    0 // Success
}

// =============================================================================
// Helper Functions
// =============================================================================

/// Example rate limit check function
#[allow(dead_code)]
fn should_rate_limit(_context: &BifrostContext) -> bool {
    // Implement your rate limiting logic here
    false
}
