# 📖 Usage Documentation

Complete API reference and usage guides for both Go package and HTTP transport integration methods.

## 🎯 Choose Your Integration Method

| Method                                   | Description                         | Best For                      | Documentation                   |
| ---------------------------------------- | ----------------------------------- | ----------------------------- | ------------------------------- |
| **[🔧 Go Package](go-package/)**         | Direct Go integration               | Go applications, custom logic | Complete Go API reference       |
| **[🌐 HTTP Transport](http-transport/)** | REST API with drop-in compatibility | Any language, microservices   | HTTP endpoints and integrations |

---

## 🔧 [Go Package Usage](go-package/)

**Direct integration for Go applications**

### Core Topics

- **[📋 Overview](go-package/README.md)** - Getting started with the Go package
- **[🎯 Bifrost Client](go-package/bifrost-client.md)** - Main client methods and configuration
- **[👤 Account Management](go-package/account.md)** - API key management and authentication
- **[🔌 Plugins](go-package/plugins.md)** - Custom middleware and request processing
- **[🛠️ MCP Integration](go-package/mcp.md)** - Model Context Protocol usage
- **[📝 Logging](go-package/logging.md)** - Logging configuration and best practices
- **[📊 Schemas](go-package/schemas.md)** - Data structures and interfaces

### Quick Links

- **[⚡ Quick Start](../quickstart/go-package.md)** - 30-second setup
- **[💡 Examples](../examples/)** - Practical code examples
- **[🏛️ Architecture](../architecture/)** - How it works internally

---

## 🌐 [HTTP Transport Usage](http-transport/)

**REST API with drop-in compatibility for existing provider SDKs**

### Core Topics

- **[📋 Overview](http-transport/README.md)** - Getting started with HTTP transport
- **[🎯 Endpoints](http-transport/endpoints.md)** - Native Bifrost REST API
- **[🔧 Configuration](http-transport/configuration/)** - JSON configuration for providers, plugins, and MCP
- **[🔄 Integrations](http-transport/integrations/)** - Drop-in replacements for OpenAI, Anthropic, GenAI

### Configuration

- **[🔗 Providers](http-transport/configuration/providers.md)** - Provider setup and configuration
- **[🛠️ MCP](http-transport/configuration/mcp.md)** - Model Context Protocol configuration
- **[🔌 Plugins](http-transport/configuration/plugins.md)** - Plugin configuration and custom plugins

### Drop-in Integrations

- **[🤖 OpenAI Compatible](http-transport/integrations/openai-compatible.md)** - Replace OpenAI API calls
- **[🧠 Anthropic Compatible](http-transport/integrations/anthropic-compatible.md)** - Replace Anthropic API calls
- **[🔍 GenAI Compatible](http-transport/integrations/genai-compatible.md)** - Replace Google GenAI API calls
- **[🔄 Migration Guide](http-transport/integrations/migration-guide.md)** - Step-by-step migration from existing providers

### Quick Links

- **[⚡ Quick Start](../quickstart/http-transport.md)** - 30-second setup
- **[💡 Examples](../examples/)** - Practical usage examples
- **[📊 OpenAPI Spec](http-transport/openapi.json)** - Machine-readable API specification

---

## 🔧 Universal Concepts

These concepts apply to both Go package and HTTP transport usage:

| Concept                                                | Description                                           | Documentation                                         |
| ------------------------------------------------------ | ----------------------------------------------------- | ----------------------------------------------------- |
| **[🔗 Providers](providers.md)**                       | Multi-provider support and advanced configurations    | Provider-specific settings, fallbacks, load balancing |
| **[🔑 Key Management](usage/key-management.md)**       | API key rotation and weighted distribution            | Key rotation strategies, security best practices      |
| **[⚡ Memory Management](usage/memory-management.md)** | Performance optimization and resource management      | Memory usage patterns, optimization techniques        |
| **[🌐 Networking](usage/networking.md)**               | Proxies, timeouts, retries, and connection management | Network configuration, proxy settings, retry policies |
| **[❌ Error Handling](errors.md)**                     | Error types, codes, and troubleshooting               | Comprehensive error reference and resolution guide    |

---

## 🚀 Getting Started

### New to Bifrost?

1. **[⚡ Quick Start](../quickstart/)** - Choose your integration method
2. **[📋 Core Concepts](../README.md#core-concepts)** - Understand key concepts
3. **[💡 Examples](../examples/)** - See practical use cases

### Migrating from Another Provider?

1. **[🔄 Migration Guide](http-transport/integrations/migration-guide.md)** - Step-by-step migration
2. **[🤖 OpenAI Users](http-transport/integrations/openai-compatible.md)** - Drop-in replacement
3. **[🧠 Anthropic Users](http-transport/integrations/anthropic-compatible.md)** - Drop-in replacement

### Need Advanced Features?

1. **[🔌 Plugins](go-package/plugins.md)** - Custom middleware
2. **[🛠️ MCP Integration](go-package/mcp.md)** - External tools
3. **[🏛️ Architecture](../architecture/)** - Understand internals

---

## 💡 Need Help?

- **[🔍 Troubleshooting](../troubleshooting.md)** - Common issues and solutions
- **[❓ FAQ](../faq.md)** - Frequently asked questions
- **[📖 Main Documentation](../README.md)** - Complete documentation hub
