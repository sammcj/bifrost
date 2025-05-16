# Bifrost Fallback System

Bifrost provides a robust fallback mechanism that allows you to define alternative providers and models to use when the primary provider fails. This ensures high availability and reliability for your AI-powered applications.

## 1. How Fallbacks Work

1. When a request is made to a primary provider, Bifrost first attempts to complete the request using that provider
2. If the primary provider fails after all retry attempts, Bifrost automatically tries the fallback providers in the order specified
3. Each fallback provider uses its own retry settings and configuration set in your account implementation
4. The first successful fallback response is returned to the client

## 2. Configuring Fallbacks

### Basic Fallback Configuration

```golang
result, err := bifrost.ChatCompletionRequest(
    context.Background(), &schemas.BifrostRequest{
        Provider: schemas.OpenAI,
        Model:    "gpt-4",
        Input: schemas.RequestInput{
            ChatCompletionInput: &messages,
        },
        Fallbacks: []schemas.Fallback{
            {
                Provider: schemas.Anthropic,
                Model:    "claude-3-sonnet",
            },
        },
    },
)
```

### Multiple Fallbacks

```golang
result, err := bifrost.ChatCompletionRequest(
    context.Background(), &schemas.BifrostRequest{
        Provider: schemas.OpenAI,
        Model:    "gpt-4",
        Input: schemas.RequestInput{
            ChatCompletionInput: &messages,
        },
        Fallbacks: []schemas.Fallback{
            {
                Provider: schemas.Anthropic,
                Model:    "claude-3-sonnet",
            },
            {
                Provider: schemas.Bedrock,
                Model:    "anthropic.claude-3-sonnet",
            },
            {
                Provider: schemas.Azure,
                Model:    "gpt-4",
            },
        },
    },
)
```

## 3. Important Considerations

### Provider Configuration

- Each fallback provider must be properly configured in your account
- If a fallback provider is not configured, it will be skipped
- Each provider's configuration (retries, timeouts, etc.) is independent

### Model Compatibility

- Ensure that the fallback models support the same capabilities as your primary model
- Consider model-specific parameters and limitations
- Verify that the fallback models are available in your account

### Performance Impact

- Fallbacks add latency when the primary provider fails
- Consider the order of fallbacks based on:
  - Provider reliability
  - Model performance
  - Cost considerations
  - Geographic location

## 4. Best Practices

1. **Provider Selection**

   - Choose fallback providers with different infrastructure
   - Consider geographic distribution for high availability
   - Balance cost and performance in fallback order

2. **Model Selection**

   - Use models with similar capabilities
   - Consider model-specific features (e.g., function calling, streaming)
   - Account for different token limits and pricing

3. **Error Handling**

   - Monitor fallback usage to identify provider issues
   - Set up alerts for frequent fallback activations (can be done using bifrost's plugin interface)
   - Regularly review and update fallback configurations

4. **Testing**
   - Test fallback scenarios in development
   - Verify all fallback providers are properly configured
   - Simulate provider failures to ensure smooth fallback

## 5. HTTP Transport Examples

### Basic HTTP Fallback Request

```json
POST /v1/chat/completions
{
    "provider": "openai",
    "model": "gpt-4",
    "input": {
        "chat_completion_input": [
            {
                "role": "user",
                "content": "Hello, how are you?"
            }
        ]
    },
    "fallbacks": [
        {
            "provider": "anthropic",
            "model": "claude-3-sonnet"
        }
    ]
}
```

### HTTP Request with Multiple Fallbacks

```json
POST /v1/chat/completions
{
    "provider": "openai",
    "model": "gpt-4",
    "input": {
        "chat_completion_input": [
            {
                "role": "user",
                "content": "Explain quantum computing"
            }
        ]
    },
    "fallbacks": [
        {
            "provider": "anthropic",
            "model": "claude-3-sonnet"
        },
        {
            "provider": "bedrock",
            "model": "anthropic.claude-3-sonnet"
        },
        {
            "provider": "azure",
            "model": "gpt-4"
        }
    ],
    "params": {
        "temperature": 0.7,
        "max_tokens": 1000
    }
}
```

### HTTP Response Example

```json
{
  "id": "chatcmpl-123",
  "object": "chat.completion",
  "choices": [
    {
      "index": 0,
      "message": {
        "role": "assistant",
        "content": "Quantum computing is a type of computing..."
      },
      "finish_reason": "stop"
    }
  ],
  "model": "claude-3-sonnet",
  "usage": {
    "prompt_tokens": 10,
    "completion_tokens": 100,
    "total_tokens": 110
  },
  "extra_fields": {
    "provider": "anthropic",
    "latency": 1.234,
    "billed_usage": {
      "prompt_tokens": 10.0,
      "completion_tokens": 100.0
    }
  }
}
```

Note: The response includes metadata about which provider was used (in this case, the fallback provider "anthropic") and its performance metrics.
