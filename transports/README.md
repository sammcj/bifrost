# 🌐 Bifrost Transports

Bifrost Transports let you run Bifrost as a blazing-fast HTTP API or integrate it directly as a Go package. Connect to any AI provider (OpenAI, Anthropic, Bedrock, and more) in seconds with automatic fallbacks and advanced features.

📖 **Complete documentation**: [docs/usage/http-transport/](../docs/usage/http-transport/)

---

## 🚀 Quick Start

### Docker

```bash
# Pull and run Bifrost HTTP API
docker pull maximhq/bifrost
docker run -p 8080:8080 \
  -v $(pwd)/config.json:/app/config/config.json \
  -e OPENAI_API_KEY \
  maximhq/bifrost
```

### Go Binary

```bash
# Install and run locally (Make sure Go is in your PATH)
go install github.com/maximhq/bifrost/transports/bifrost-http@latest
bifrost-http -config config.json -port 8080
```

**Ready in 30 seconds!** See [HTTP Transport Quickstart](../docs/quickstart/http-transport.md) for detailed setup.

---

## 🌟 Key Features

| Feature                       | Description                                                         | Learn More                                                 |
| ----------------------------- | ------------------------------------------------------------------- | ---------------------------------------------------------- |
| **🔄 Multi-Provider Support** | OpenAI, Anthropic, Bedrock, Vertex, Cohere, Mistral, Ollama         | [Provider Setup](../docs/usage/providers.md)               |
| **🔌 Drop-in Compatibility**  | Replace OpenAI/Anthropic/GenAI APIs with zero code changes          | [Integrations](../docs/usage/http-transport/integrations/) |
| **🛠️ MCP Tool Calling**       | Enable AI models to use external tools (filesystem, web, databases) | [MCP Guide](../docs/mcp.md)                                |
| **⚡ Plugin System**          | Add analytics, caching, rate limiting, custom logic                 | [Plugin System](../docs/plugins.md)                        |
| **📊 Built-in Monitoring**    | Prometheus metrics at `/metrics` endpoint                           | [Monitoring](../docs/usage/http-transport/endpoints.md)    |
| **🔀 Automatic Fallbacks**    | Seamless failover between providers and models                      | [Fallback Config](../docs/usage/providers.md)              |

---

## 🎯 Usage Examples

### Basic Chat Completion

```bash
curl -X POST http://localhost:8080/v1/chat/completions \
  -H "Content-Type: application/json" \
  -d '{
    "provider": "openai",
    "model": "gpt-4o-mini",
    "messages": [{"role": "user", "content": "Hello!"}]
  }'
```

### Drop-in OpenAI Replacement

```python
import openai

# Just change the base_url - everything else stays the same!
client = openai.OpenAI(
    base_url="http://localhost:8080/openai",
    api_key="dummy-key"  # Handled by Bifrost
)

response = client.chat.completions.create(
    model="gpt-4o-mini",
    messages=[{"role": "user", "content": "Hello from Bifrost!"}]
)
```

### Multi-Provider Fallbacks

```bash
curl -X POST http://localhost:8080/v1/chat/completions \
  -H "Content-Type: application/json" \
  -d '{
    "provider": "openai",
    "model": "gpt-4o-mini",
    "messages": [{"role": "user", "content": "Hello!"}],
    "fallbacks": [
      {"provider": "anthropic", "model": "claude-3-5-sonnet-20241022"},
      {"provider": "bedrock", "model": "anthropic.claude-3-sonnet-20240229-v1:0"}
    ]
  }'
```

**More examples**: [HTTP Transport Endpoints](../docs/usage/http-transport/endpoints.md)

---

## ⚙️ Configuration

### Minimal Config

```json
{
  "providers": {
    "openai": {
      "keys": [
        {
          "value": "env.OPENAI_API_KEY",
          "models": ["gpt-4o-mini"],
          "weight": 1.0
        }
      ]
    }
  }
}
```

**Learn More:**

- [Provider Setup Guide](../docs/usage/http-transport/configuration/providers.md)
- [MCP Configuration](../docs/usage/http-transport/configuration/mcp.md)
- [Plugin Configuration](../docs/usage/http-transport/configuration/plugins.md)
- [Complete Examples](config.example.json)

---

## 🛠️ Advanced Features

### MCP Tool Integration

Enable AI models to use external tools like filesystem operations, web search, and databases:

```bash
# AI automatically uses configured tools
curl -X POST http://localhost:8080/v1/chat/completions \
  -d '{"provider": "openai", "model": "gpt-4o-mini",
       "messages": [{"role": "user", "content": "List files in /tmp"}]}'
```

**Learn more**: [MCP Integration Guide](../docs/mcp.md)

### Plugin System

Add custom middleware for analytics, caching, rate limiting:

```bash
# Run with plugins
bifrost-http -config config.json -plugins "maxim,redis"
```

**Available plugins**: [Plugin Repository](https://github.com/maximhq/bifrost/tree/main/plugins) | [Plugin Guide](../docs/plugins.md)

### Prometheus Monitoring

Built-in metrics collection at `/metrics`:

```bash
curl http://localhost:8080/metrics
```

**Custom labels**: `-prometheus-labels team-id,task-id`

---

## 🔧 Runtime Configuration

### For Go Binary

| Flag                 | Default | Description                 |
| -------------------- | ------- | --------------------------- |
| `-config`            | -       | Path to config.json file    |
| `-port`              | 8080    | HTTP server port            |
| `-pool-size`         | 300     | Connection pool size        |
| `-plugins`           | -       | Comma-separated plugin list |
| `-prometheus-labels` | -       | Custom Prometheus labels    |

### For Docker

| Variable                   | Description                           |
| -------------------------- | ------------------------------------- |
| `APP_PORT`                 | Server port override                  |
| `APP_POOL_SIZE`            | Pool size override                    |
| `APP_PLUGINS`              | Plugin list override                  |
| `APP_DROP_EXCESS_REQUESTS` | Drop excess requests when buffer full |

---

## 📚 Documentation

### 🎯 Getting Started

- **[⚡ 30-Second Quickstart](../docs/quickstart/http-transport.md)** - Get running immediately
- **[🔧 Configuration Guide](../docs/usage/http-transport/configuration/)** - Providers, MCP, plugins
- **[🔄 Migration Guide](../docs/usage/http-transport/integrations/migration-guide.md)** - Migrate from existing providers

### 🚀 Core Features

- **[🔗 Multi-Provider Support](../docs/usage/providers.md)** - 8+ AI providers with fallbacks
- **[🛠️ MCP Integration](../docs/mcp.md)** - External tool calling for AI models
- **[🔌 Plugin System](../docs/plugins.md)** - Extensible middleware architecture

### 🌐 API Integrations

- **[🤖 OpenAI Compatible](../docs/usage/http-transport/integrations/openai-compatible.md)** - Drop-in OpenAI replacement
- **[🧠 Anthropic Compatible](../docs/usage/http-transport/integrations/anthropic-compatible.md)** - Drop-in Anthropic replacement
- **[🔍 GenAI Compatible](../docs/usage/http-transport/integrations/genai-compatible.md)** - Drop-in Google GenAI replacement

### 🏛️ Architecture & Performance

- **[📊 Benchmarks](../docs/benchmarks.md)** - Performance metrics and optimization
- **[🏗️ Architecture](../docs/architecture/)** - System design and internals
- **[💡 Examples](../docs/examples/)** - Real-world usage patterns

---

## 🎉 Ready to Scale?

🚀 **Production Deployment**: [Production Guide](../docs/usage/http-transport/configuration/)  
📈 **Performance Tuning**: [Benchmarks & Optimization](../docs/benchmarks.md)  
🔍 **Troubleshooting**: [Common Issues](../docs/usage/errors.md)

---

_Built with ❤️ by [Maxim](https://github.com/maximhq)_
