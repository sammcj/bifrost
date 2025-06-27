# Bifrost Core Package Documentation

This guide covers how to use Bifrost as a Go package in your applications, providing direct integration without the need for external transports.

![Bifrost Package Demo](./media/package-demo.mp4)

## 📑 Table of Contents

- [Bifrost Core Package Documentation](#bifrost-core-package-documentation)
  - [📑 Table of Contents](#-table-of-contents)
  - [Package Structure](#package-structure)
  - [Getting Started](#getting-started)
  - [Basic Usage](#basic-usage)
    - [Implementing Your Account Interface](#implementing-your-account-interface)
    - [Initializing Bifrost](#initializing-bifrost)
    - [Making Your First LLM Call](#making-your-first-llm-call)
  - [Advanced Configuration](#advanced-configuration)
  - [Additional Features](#additional-features)
    - [🧠 Memory Management](#-memory-management)
    - [📝 Logger](#-logger)
    - [🔌 Plugins](#-plugins)
    - [⚙️ Provider Configurations](#️-provider-configurations)
    - [🔄 Fallbacks](#-fallbacks)
    - [🛠️ MCP Integration](#️-mcp-integration)
  - [Next Steps](#next-steps)

---

## Package Structure

Bifrost is built with a modular architecture where the core functionality is separated from transport layers:

```text
bifrost/
├── core/                 # Core functionality and shared components
│   ├── providers/        # Provider-specific implementations
│   ├── schemas/          # Interfaces and structs used in bifrost
│   ├── bifrost.go        # Main Bifrost implementation
│   ├── logger.go         # Logging functionality
│   ├── mcp.go           # Model Context Protocol support
│   └── utils.go         # Utility functions
```

All interfaces are defined in `core/schemas/` and can be used as a reference for contributions and custom implementations.

---

## Getting Started

To use Bifrost as a Go package in your application:

```bash
go get github.com/maximhq/bifrost/core
```

---

## Basic Usage

### Implementing Your Account Interface

First, create an account that follows [Bifrost's account interface](https://github.com/maximhq/bifrost/blob/main/core/schemas/account.go):

```golang
package main

import (
    "os"
    "github.com/maximhq/bifrost/core/schemas"
)

type BaseAccount struct{}

func (baseAccount *BaseAccount) GetConfiguredProviders() ([]schemas.ModelProvider, error) {
    return []schemas.ModelProvider{schemas.OpenAI}, nil
}

func (baseAccount *BaseAccount) GetKeysForProvider(providerKey schemas.ModelProvider) ([]schemas.Key, error) {
    return []schemas.Key{
        {
            Value:  os.Getenv("OPENAI_API_KEY"),
            Models: []string{"gpt-4o-mini"},
            Weight: 1.0,
        },
    }, nil
}

func (baseAccount *BaseAccount) GetConfigForProvider(providerKey schemas.ModelProvider) (*schemas.ProviderConfig, error) {
    return &schemas.ProviderConfig{
        NetworkConfig:            schemas.DefaultNetworkConfig,
        ConcurrencyAndBufferSize: schemas.DefaultConcurrencyAndBufferSize,
    }, nil
}
```

Bifrost uses these methods to get all the keys and configurations it needs to call the providers.

### Initializing Bifrost

Set up the Bifrost instance by providing your account implementation:

```golang
package main

import (
    "context"
    "github.com/maximhq/bifrost/core"
    "github.com/maximhq/bifrost/core/schemas"
)

func main() {
    account := BaseAccount{}

    client, err := bifrost.Init(schemas.BifrostConfig{
        Account: &account,
    })
    if err != nil {
        panic(err)
    }
}
```

### Making Your First LLM Call

```golang
bifrostResult, bifrostErr := client.ChatCompletionRequest(
    context.Background(),
    &schemas.BifrostRequest{
        Provider: schemas.OpenAI,
        Model: "gpt-4o-mini", // make sure you have configured gpt-4o-mini in your account interface
        Input: schemas.RequestInput{
            ChatCompletionInput: bifrost.Ptr([]schemas.BifrostMessage{{
                Role: schemas.ModelChatMessageRoleUser,
                Content: schemas.MessageContent{
                    ContentStr: bifrost.Ptr("What is a LLM gateway?"),
                },
            }}),
        },
    },
)

if bifrostErr != nil {
    panic(bifrostErr)
}

// Handle the response
fmt.Println(bifrostResult.Response)
```

You can add model parameters by including `Params: &schemas.ModelParameters{...yourParams}` in ChatCompletionRequest.

---

## Advanced Configuration

Bifrost offers extensive configuration options to customize behavior for your specific needs. You can configure various aspects through the account interface and initialization parameters.

For detailed configuration options, see the [Provider Configurations](./providers.md) documentation.

---

## Additional Features

Bifrost provides several advanced features to enhance your AI application development:

### 🧠 Memory Management

Optimize memory usage and performance with configurable buffer sizes and connection pooling.

- **Documentation**: [Memory Management](./memory-management.md)

### 📝 Logger

Built-in logging system with configurable levels and output formats.

- **Documentation**: [Logger](./logger.md)

### 🔌 Plugins

Extend Bifrost functionality with custom plugins using the plugin-first architecture.

- **Documentation**: [Plugins](./plugins.md)

### ⚙️ Provider Configurations

Fine-tune provider-specific settings including retry logic, timeouts, and concurrency limits.

- **Documentation**: [Provider Configurations](./providers.md)

### 🔄 Fallbacks

Implement robust fallback mechanisms for high availability across multiple providers and models.

- **Documentation**: [Fallbacks](./fallbacks.md)

### 🛠️ MCP Integration

Leverage Model Context Protocol (MCP) for external tool integration and execution.

- **Documentation**: [MCP Integration](./mcp.md)

---

## Next Steps

- Explore the [HTTP Transport](../transports/README.md) for API-based integration
- Check out [example implementations](../tests/core-chatbot/) for real-world usage patterns
- Review the [system architecture](./system-architecture.md) for understanding Bifrost's internal design
