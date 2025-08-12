# InProcess MCP Example

This example demonstrates how to use Bifrost's InProcess MCP connection type to embed custom tools directly in your Go application for maximum performance.

## What is InProcess MCP?

InProcess MCP connections allow you to connect directly to an MCP server running in the same process as your application. This provides:

- **Lowest latency** (~0.1ms) - No serialization, deserialization, or IPC overhead
- **Highest security** - Tools run in the same process with full control
- **Simplified deployment** - No external processes or services to manage
- **Perfect for testing** - Test MCP integrations without external dependencies

## Running the Example

1. Set your OpenAI API key:
```bash
export OPENAI_API_KEY="your-api-key-here"
```

2. Run the example:
```bash
cd examples/inprocess-mcp
go mod tidy
go run main.go
```

## Example Tools

This example includes three custom tools:

### Calculator Tool
Performs basic arithmetic operations (add, subtract, multiply, divide).

### Timestamp Tool
Returns the current timestamp in various formats.

### JSON Formatter Tool
Formats data as pretty-printed JSON with custom indentation.

## Code Structure

The example shows how to:

1. Create a custom MCP server using `server.NewMCPServer()`
2. Add tools to the server with proper schemas
3. Configure Bifrost with an InProcess connection
4. Use the tools through the AI model

## Performance Comparison

| Connection Type | Latency   | Use Case                    |
|----------------|-----------|------------------------------|
| InProcess      | ~0.1ms    | Embedded tools, testing      |
| STDIO          | ~1-10ms   | Local CLI tools              |
| HTTP           | ~10-500ms | Remote services              |
| SSE            | Variable  | Real-time streaming          |

## When to Use InProcess

InProcess connections are ideal for:

- **High-performance applications** requiring minimal latency
- **Embedded tools** that are part of your application logic
- **Testing and development** without external dependencies
- **Security-sensitive operations** where tools must run in a controlled environment

## Limitations

- Only available when using the Bifrost Go package directly
- Cannot be configured via JSON or HTTP transport
- Server instance must be created programmatically
- Tools must be implemented in Go

## See Also

- [Bifrost MCP Documentation](../../docs/usage/go-package/mcp.md)
- [MCP Architecture](../../docs/architecture/mcp.md)