// MCP Server — runs as a stdio MCP server process.
//
// This demonstrates using claudeagent.Tool() and claudeagent.CreateSdkMcpServer()
// to define custom tools, then serving them over stdio so the Claude CLI can use them.
//
// Usage: go run ./demos/mcp-inline/server
// (The Claude CLI spawns this as a subprocess via McpStdioServerConfig)
package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	claudeagent "github.com/anthropics/claude-agent-sdk-go"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func main() {
	// Define tools using the SDK's Tool() helper
	greetTool := claudeagent.Tool("greet",
		"Greet a person by name. Returns a friendly greeting.",
		mcp.WithString("name",
			mcp.Required(),
			mcp.Description("The person's name"),
		),
	)
	greetTool.Handler = func(ctx context.Context, req claudeagent.CallToolRequest) (*claudeagent.CallToolResult, error) {
		name, _ := req.GetArguments()["name"].(string)
		return claudeagent.NewToolResultText(fmt.Sprintf("Hello, %s! Welcome to the Go SDK.", name)), nil
	}

	reverseTool := claudeagent.Tool("reverse_text",
		"Reverse a string of text.",
		mcp.WithString("text",
			mcp.Required(),
			mcp.Description("The text to reverse"),
		),
	)
	reverseTool.Handler = func(ctx context.Context, req claudeagent.CallToolRequest) (*claudeagent.CallToolResult, error) {
		text, _ := req.GetArguments()["text"].(string)
		runes := []rune(text)
		for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
			runes[i], runes[j] = runes[j], runes[i]
		}
		return claudeagent.NewToolResultText(string(runes)), nil
	}

	calcTool := claudeagent.Tool("calculate",
		"Perform a simple arithmetic calculation. Supports +, -, *, /.",
		mcp.WithNumber("a", mcp.Required(), mcp.Description("First number")),
		mcp.WithNumber("b", mcp.Required(), mcp.Description("Second number")),
		mcp.WithString("operation", mcp.Required(), mcp.Description("Operation: add, subtract, multiply, divide")),
	)
	calcTool.Handler = func(ctx context.Context, req claudeagent.CallToolRequest) (*claudeagent.CallToolResult, error) {
		args := req.GetArguments()
		a, _ := args["a"].(float64)
		b, _ := args["b"].(float64)
		op, _ := args["operation"].(string)

		var result float64
		switch strings.ToLower(op) {
		case "add", "+":
			result = a + b
		case "subtract", "-":
			result = a - b
		case "multiply", "*":
			result = a * b
		case "divide", "/":
			if b == 0 {
				return claudeagent.NewToolResultError("division by zero"), nil
			}
			result = a / b
		default:
			return claudeagent.NewToolResultError(fmt.Sprintf("unknown operation: %s", op)), nil
		}

		return claudeagent.NewToolResultText(fmt.Sprintf("%.6g", result)), nil
	}

	// Create the MCP server
	srv := claudeagent.CreateSdkMcpServer(claudeagent.CreateSdkMcpServerOptions{
		Name:    "go-sdk-tools",
		Version: "1.0.0",
		Tools: []claudeagent.SdkMcpToolDefinition{
			greetTool,
			reverseTool,
			calcTool,
		},
	})

	// Get the underlying mcp-go MCPServer and serve over stdio
	mcpServer := srv.Instance.(*server.MCPServer)
	stdio := server.NewStdioServer(mcpServer)

	if err := stdio.Listen(context.Background(), os.Stdin, os.Stdout); err != nil {
		fmt.Fprintf(os.Stderr, "MCP server error: %v\n", err)
		os.Exit(1)
	}
}
