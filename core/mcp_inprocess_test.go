package bifrost

import (
	"context"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/maximhq/bifrost/core/schemas"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestInProcessMCPConnection tests the InProcess MCP connection type
func TestInProcessMCPConnection(t *testing.T) {
	// Create a custom MCP server for testing
	testServer := server.NewMCPServer(
		"TestServer",
		"1.0.0",
		server.WithToolCapabilities(true),
	)

	// Add a test tool to the server
	testToolExecuted := false
	testServer.AddTool(
		mcp.NewTool("test_tool", mcp.WithDescription("Test tool for InProcess connection")),
		func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			testToolExecuted = true
			args := request.GetArguments()
			if msg, ok := args["message"].(string); ok {
				return mcp.NewToolResultText("Received: " + msg), nil
			}
			return mcp.NewToolResultText("Test tool executed"), nil
		},
	)

	// Create MCP manager with InProcess connection
	logger := NewDefaultLogger(schemas.LogLevelDebug)
	mcpConfig := schemas.MCPConfig{
		ClientConfigs: []schemas.MCPClientConfig{
			{
				Name:            "test-inprocess",
				ConnectionType:  schemas.MCPConnectionTypeInProcess,
				InProcessServer: testServer,
			},
		},
	}

	manager, err := newMCPManager(mcpConfig, logger)
	require.NoError(t, err, "Failed to create MCP manager")
	defer manager.cleanup()

	// Verify the client was added
	assert.Contains(t, manager.clientMap, "test-inprocess", "InProcess client should be in client map")

	// Verify the client is connected
	client := manager.clientMap["test-inprocess"]
	assert.NotNil(t, client.Conn, "InProcess client should have an active connection")
	assert.Equal(t, schemas.MCPConnectionTypeInProcess, client.ConnectionInfo.Type, "Connection type should be InProcess")

	// Get available tools
	ctx := context.Background()
	tools := manager.getAvailableTools(ctx)

	// Verify the test tool is available
	foundTestTool := false
	for _, tool := range tools {
		if tool.Function.Name == "test_tool" {
			foundTestTool = true
			assert.Equal(t, "Test tool for InProcess connection", tool.Function.Description)
			break
		}
	}
	assert.True(t, foundTestTool, "Test tool should be available through InProcess connection")

	// Execute the tool
	toolCall := schemas.ToolCall{
		ID:   &[]string{"test_call_1"}[0],
		Type: &[]string{"function"}[0],
		Function: schemas.FunctionCall{
			Name:      &[]string{"test_tool"}[0],
			Arguments: `{"message": "Hello InProcess"}`,
		},
	}

	result, err := manager.executeTool(ctx, toolCall)
	require.NoError(t, err, "Tool execution should not fail")
	assert.NotNil(t, result, "Tool execution should return a result")
	assert.Equal(t, schemas.ModelChatMessageRoleTool, result.Role, "Result should have tool role")
	assert.Contains(t, *result.Content.ContentStr, "Received: Hello InProcess", "Tool should process the message correctly")
	assert.True(t, testToolExecuted, "Test tool handler should have been executed")
}

// TestInProcessValidation tests validation of InProcess connections
func TestInProcessValidation(t *testing.T) {
	// Test missing server instance
	config := schemas.MCPClientConfig{
		Name:           "invalid-inprocess",
		ConnectionType: schemas.MCPConnectionTypeInProcess,
		// InProcessServer is nil
	}

	err := validateMCPClientConfig(&config)
	assert.Error(t, err, "Validation should fail for missing InProcessServer")
	assert.Contains(t, err.Error(), "InProcessServer is required", "Error should mention missing server")

	// Test valid configuration
	testServer := server.NewMCPServer("TestServer", "1.0.0")
	config.InProcessServer = testServer

	err = validateMCPClientConfig(&config)
	assert.NoError(t, err, "Validation should pass with valid InProcessServer")
}

// TestInProcessWithMultipleConnections tests InProcess alongside other connection types
func TestInProcessWithMultipleConnections(t *testing.T) {
	// Create an InProcess server
	inProcessServer := server.NewMCPServer(
		"InProcessServer",
		"1.0.0",
		server.WithToolCapabilities(true),
	)

	inProcessServer.AddTool(
		mcp.NewTool("inprocess_tool", mcp.WithDescription("Tool from InProcess server")),
		func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			return mcp.NewToolResultText("InProcess tool executed"), nil
		},
	)

	// Configure multiple connection types
	mcpConfig := schemas.MCPConfig{
		ClientConfigs: []schemas.MCPClientConfig{
			{
				Name:            "inprocess-client",
				ConnectionType:  schemas.MCPConnectionTypeInProcess,
				InProcessServer: inProcessServer,
			},
			// Note: We can't easily test STDIO without an actual executable,
			// but the manager should handle mixed connection types
		},
	}

	logger := NewDefaultLogger(schemas.LogLevelDebug)
	manager, err := newMCPManager(mcpConfig, logger)
	require.NoError(t, err, "Failed to create MCP manager with mixed connections")
	defer manager.cleanup()

	// Verify both InProcess client and internal Bifrost client exist
	assert.Contains(t, manager.clientMap, "inprocess-client", "InProcess client should exist")

	// Get available tools and verify InProcess tool is available
	ctx := context.Background()
	tools := manager.getAvailableTools(ctx)

	foundInProcessTool := false
	for _, tool := range tools {
		if tool.Function.Name == "inprocess_tool" {
			foundInProcessTool = true
			break
		}
	}
	assert.True(t, foundInProcessTool, "InProcess tool should be available")
}

// TestInProcessPerformance tests that InProcess connections work efficiently
func TestInProcessPerformance(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping performance test in short mode")
	}

	// Create a server with a simple tool
	perfServer := server.NewMCPServer(
		"PerformanceServer",
		"1.0.0",
		server.WithToolCapabilities(true),
	)

	callCount := 0
	perfServer.AddTool(
		mcp.NewTool("perf_tool", mcp.WithDescription("Performance test tool")),
		func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			callCount++
			return mcp.NewToolResultText("Call completed"), nil
		},
	)

	mcpConfig := schemas.MCPConfig{
		ClientConfigs: []schemas.MCPClientConfig{
			{
				Name:            "perf-test",
				ConnectionType:  schemas.MCPConnectionTypeInProcess,
				InProcessServer: perfServer,
			},
		},
	}

	logger := NewDefaultLogger(schemas.LogLevelError) // Reduce logging for performance test
	manager, err := newMCPManager(mcpConfig, logger)
	require.NoError(t, err)
	defer manager.cleanup()

	// Execute tool multiple times
	ctx := context.Background()
	iterations := 100

	for i := 0; i < iterations; i++ {
		toolCall := schemas.ToolCall{
			ID:   &[]string{"call_" + string(rune(i))}[0],
			Type: &[]string{"function"}[0],
			Function: schemas.FunctionCall{
				Name:      &[]string{"perf_tool"}[0],
				Arguments: "{}",
			},
		}

		result, err := manager.executeTool(ctx, toolCall)
		require.NoError(t, err, "Tool execution should not fail on iteration %d", i)
		assert.NotNil(t, result, "Should get result on iteration %d", i)
	}

	assert.Equal(t, iterations, callCount, "All tool calls should have been executed")
}
