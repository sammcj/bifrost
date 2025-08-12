package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	bifrost "github.com/maximhq/bifrost/core"
	"github.com/maximhq/bifrost/core/schemas"
)

// SimpleAccount implements the schemas.Account interface
type SimpleAccount struct {
	APIKey string
}

func (a *SimpleAccount) GetByProvider(provider schemas.Provider) (*string, error) {
	return &a.APIKey, nil
}

// Custom tool handlers
func calculatorTool(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := request.GetArguments()

	operation, ok := args["operation"].(string)
	if !ok {
		return mcp.NewToolResultError("operation parameter required"), nil
	}

	a, aOk := args["a"].(float64)
	b, bOk := args["b"].(float64)

	if !aOk || !bOk {
		return mcp.NewToolResultError("parameters 'a' and 'b' must be numbers"), nil
	}

	var result float64
	switch operation {
	case "add":
		result = a + b
	case "subtract":
		result = a - b
	case "multiply":
		result = a * b
	case "divide":
		if b == 0 {
			return mcp.NewToolResultError("division by zero"), nil
		}
		result = a / b
	default:
		return mcp.NewToolResultError(fmt.Sprintf("unknown operation: %s", operation)), nil
	}

	return mcp.NewToolResultText(fmt.Sprintf("%.2f", result)), nil
}

func timestampTool(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := request.GetArguments()

	format := "2006-01-02 15:04:05"
	if customFormat, ok := args["format"].(string); ok {
		format = customFormat
	}

	return mcp.NewToolResultText(time.Now().Format(format)), nil
}

func jsonFormatterTool(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := request.GetArguments()

	data, ok := args["data"]
	if !ok {
		return mcp.NewToolResultError("data parameter required"), nil
	}

	indent := "  "
	if customIndent, ok := args["indent"].(string); ok {
		indent = customIndent
	}

	jsonBytes, err := json.MarshalIndent(data, "", indent)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to format JSON: %v", err)), nil
	}

	return mcp.NewToolResultText(string(jsonBytes)), nil
}

func main() {
	// Create a custom MCP server with various tools
	customServer := server.NewMCPServer(
		"CustomToolServer",
		"1.0.0",
		server.WithToolCapabilities(true),
	)

	// Add calculator tool
	customServer.AddTool(
		mcp.NewTool("calculator",
			mcp.WithDescription("Perform basic arithmetic operations"),
			mcp.WithInputSchema(mcp.ToolInputSchema{
				Type: "object",
				Properties: map[string]interface{}{
					"operation": map[string]interface{}{
						"type":        "string",
						"enum":        []string{"add", "subtract", "multiply", "divide"},
						"description": "The arithmetic operation to perform",
					},
					"a": map[string]interface{}{
						"type":        "number",
						"description": "First number",
					},
					"b": map[string]interface{}{
						"type":        "number",
						"description": "Second number",
					},
				},
				Required: []string{"operation", "a", "b"},
			}),
		),
		calculatorTool,
	)

	// Add timestamp tool
	customServer.AddTool(
		mcp.NewTool("get_timestamp",
			mcp.WithDescription("Get the current timestamp"),
			mcp.WithInputSchema(mcp.ToolInputSchema{
				Type: "object",
				Properties: map[string]interface{}{
					"format": map[string]interface{}{
						"type":        "string",
						"description": "Go time format string (optional)",
					},
				},
			}),
		),
		timestampTool,
	)

	// Add JSON formatter tool
	customServer.AddTool(
		mcp.NewTool("format_json",
			mcp.WithDescription("Format data as pretty-printed JSON"),
			mcp.WithInputSchema(mcp.ToolInputSchema{
				Type: "object",
				Properties: map[string]interface{}{
					"data": map[string]interface{}{
						"type":        "object",
						"description": "The data to format as JSON",
					},
					"indent": map[string]interface{}{
						"type":        "string",
						"description": "Indentation string (default: two spaces)",
					},
				},
				Required: []string{"data"},
			}),
		),
		jsonFormatterTool,
	)

	// Initialize Bifrost with InProcess MCP connection
	client, err := bifrost.Init(schemas.BifrostConfig{
		Account: &SimpleAccount{APIKey: "your-api-key-here"},
		MCPConfig: &schemas.MCPConfig{
			ClientConfigs: []schemas.MCPClientConfig{
				{
					Name:            "custom-tools",
					ConnectionType:  schemas.MCPConnectionTypeInProcess,
					InProcessServer: customServer,
				},
			},
		},
		Logger: bifrost.NewDefaultLogger(schemas.LogLevelInfo),
	})
	if err != nil {
		log.Fatalf("Failed to initialize Bifrost: %v", err)
	}
	defer client.Cleanup()

	fmt.Println("🚀 Bifrost initialized with InProcess MCP tools!")
	fmt.Println("\nAvailable tools:")
	fmt.Println("  - calculator: Perform arithmetic operations")
	fmt.Println("  - get_timestamp: Get current timestamp")
	fmt.Println("  - format_json: Format data as JSON")

	// Example 1: Using the calculator tool
	fmt.Println("\n📊 Example 1: Calculator")
	response, err := client.ChatCompletionRequest(context.Background(), &schemas.BifrostRequest{
		Provider: schemas.OpenAI,
		Model:    "gpt-4o-mini",
		Input: schemas.RequestInput{
			ChatCompletionInput: &[]schemas.BifrostMessage{
				{
					Role: schemas.ModelChatMessageRoleUser,
					Content: schemas.MessageContent{
						ContentStr: &[]string{"What is 42 multiplied by 3.14?"}[0],
					},
				},
			},
		},
	})

	if err != nil {
		log.Printf("Request failed: %v", err)
	} else if len(response.Choices) > 0 {
		handleResponse(client, response)
	}

	// Example 2: Using the timestamp tool
	fmt.Println("\n⏰ Example 2: Timestamp")
	response, err = client.ChatCompletionRequest(context.Background(), &schemas.BifrostRequest{
		Provider: schemas.OpenAI,
		Model:    "gpt-4o-mini",
		Input: schemas.RequestInput{
			ChatCompletionInput: &[]schemas.BifrostMessage{
				{
					Role: schemas.ModelChatMessageRoleUser,
					Content: schemas.MessageContent{
						ContentStr: &[]string{"What is the current date and time?"}[0],
					},
				},
			},
		},
	})

	if err != nil {
		log.Printf("Request failed: %v", err)
	} else if len(response.Choices) > 0 {
		handleResponse(client, response)
	}

	// Example 3: Complex calculation with multiple tool calls
	fmt.Println("\n🔢 Example 3: Complex Calculation")
	response, err = client.ChatCompletionRequest(context.Background(), &schemas.BifrostRequest{
		Provider: schemas.OpenAI,
		Model:    "gpt-4o-mini",
		Input: schemas.RequestInput{
			ChatCompletionInput: &[]schemas.BifrostMessage{
				{
					Role: schemas.ModelChatMessageRoleUser,
					Content: schemas.MessageContent{
						ContentStr: &[]string{"Calculate (10 + 20) * 3 - 15 / 5. Show each step."}[0],
					},
				},
			},
		},
	})

	if err != nil {
		log.Printf("Request failed: %v", err)
	} else if len(response.Choices) > 0 {
		handleResponse(client, response)
	}
}

func handleResponse(client *bifrost.Bifrost, response *schemas.BifrostResponse) {
	choice := response.Choices[0]

	// Check if the model wants to use tools
	if choice.Message.ToolCalls != nil {
		fmt.Println("  Model is using tools...")

		// Execute each tool call
		conversation := []schemas.BifrostMessage{choice.Message}

		for _, toolCall := range *choice.Message.ToolCalls {
			if toolCall.Function.Name != nil {
				fmt.Printf("  Executing tool: %s\n", *toolCall.Function.Name)

				// Execute the tool
				toolResult, err := client.ExecuteMCPTool(context.Background(), toolCall)
				if err != nil {
					log.Printf("  Tool execution failed: %v", err)
					continue
				}

				// Add tool result to conversation
				conversation = append(conversation, *toolResult)

				if toolResult.Content.ContentStr != nil {
					fmt.Printf("  Tool result: %s\n", *toolResult.Content.ContentStr)
				}
			}
		}

		// Continue the conversation with tool results
		followUp, err := client.ChatCompletionRequest(context.Background(), &schemas.BifrostRequest{
			Provider: schemas.OpenAI,
			Model:    "gpt-4o-mini",
			Input: schemas.RequestInput{
				ChatCompletionInput: &conversation,
			},
		})

		if err != nil {
			log.Printf("Follow-up request failed: %v", err)
		} else if len(followUp.Choices) > 0 && followUp.Choices[0].Message.Content.ContentStr != nil {
			fmt.Printf("  Final answer: %s\n", *followUp.Choices[0].Message.Content.ContentStr)
		}
	} else if choice.Message.Content.ContentStr != nil {
		fmt.Printf("  Response: %s\n", *choice.Message.Content.ContentStr)
	}
}
