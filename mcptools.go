package claudeagent

import (
	"context"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// SdkMcpToolDefinition defines a custom MCP tool with a typed handler.
// This is the Go equivalent of the TypeScript SDK's tool() function.
//
// Example:
//
//	weatherTool := claudeagent.Tool("get_weather",
//	    "Get current weather for a city",
//	    mcp.WithString("city", mcp.Required(), mcp.Description("City name")),
//	    mcp.WithAnnotations(mcp.ToolAnnotation{ReadOnlyHint: claudeagent.Bool(true)}),
//	)
//	weatherTool.Handler = func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
//	    city := req.GetArguments()["city"].(string)
//	    return &mcp.CallToolResult{
//	        Content: []mcp.Content{mcp.NewTextContent("Sunny in " + city)},
//	    }, nil
//	}
type SdkMcpToolDefinition struct {
	// Name is the tool name exposed to Claude.
	Name string

	// Description is a human-readable description.
	Description string

	// Tool is the underlying mcp-go Tool definition with input schema.
	Tool mcp.Tool

	// Annotations provides MCP tool hints (readOnly, destructive, openWorld, idempotent).
	Annotations *mcp.ToolAnnotation

	// Handler is called when Claude invokes this tool.
	Handler server.ToolHandlerFunc
}

// Tool creates a new SdkMcpToolDefinition with the given name, description,
// and mcp-go tool options (WithString, WithNumber, WithBoolean, etc.).
//
// This is the Go equivalent of the TypeScript SDK's tool() helper.
//
// Example:
//
//	t := claudeagent.Tool("calculator", "Perform arithmetic",
//	    mcp.WithString("expression", mcp.Required(), mcp.Description("Math expression")),
//	)
//	t.Handler = func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
//	    expr := req.GetArguments()["expression"].(string)
//	    return mcp.NewToolResultText(expr + " = 42"), nil
//	}
func Tool(name, description string, opts ...mcp.ToolOption) SdkMcpToolDefinition {
	allOpts := append([]mcp.ToolOption{mcp.WithDescription(description)}, opts...)
	return SdkMcpToolDefinition{
		Name:        name,
		Description: description,
		Tool:        mcp.NewTool(name, allOpts...),
	}
}

// CreateSdkMcpServerOptions configures an in-process MCP server.
type CreateSdkMcpServerOptions struct {
	// Name is the server name (shown in MCP server status).
	Name string

	// Version is the server version (default "1.0.0").
	Version string

	// Tools are the tool definitions to register.
	Tools []SdkMcpToolDefinition
}

// CreateSdkMcpServer creates an in-process MCP server backed by mcp-go.
// Returns an McpSdkServerConfigWithInstance.
//
// IMPORTANT: True in-process MCP (where the server runs in the same Go process)
// requires the bidirectional --input-format stream-json control protocol, which
// the Go SDK does not yet implement (it currently uses --print mode).
//
// For now, use one of these approaches:
//
// 1. External stdio server (RECOMMENDED — works today):
//
//	q := claudeagent.NewQuery(claudeagent.QueryParams{
//	    Prompt: "Use my tool",
//	    Options: &claudeagent.Options{
//	        McpServers: map[string]interface{}{
//	            "my-tools": claudeagent.McpStdioServerConfig{
//	                Command: "go", Args: []string{"run", "./my-mcp-server"},
//	            },
//	        },
//	    },
//	})
//
// 2. Build a standalone server binary using CreateSdkMcpServer + mcp-go stdio:
//
//	// In my-mcp-server/main.go:
//	srv := claudeagent.CreateSdkMcpServer(opts)
//	mcpServer := srv.Instance.(*server.MCPServer)
//	stdio := server.NewStdioServer(mcpServer)
//	stdio.Listen(ctx, os.Stdin, os.Stdout)
//
// See demos/mcp-inline/ for a complete working example.
func CreateSdkMcpServer(opts CreateSdkMcpServerOptions) McpSdkServerConfigWithInstance {
	version := opts.Version
	if version == "" {
		version = "1.0.0"
	}

	mcpServer := server.NewMCPServer(opts.Name, version)

	for _, toolDef := range opts.Tools {
		if toolDef.Handler != nil {
			mcpServer.AddTool(toolDef.Tool, toolDef.Handler)
		}
	}

	return McpSdkServerConfigWithInstance{
		McpSdkServerConfig: McpSdkServerConfig{
			Type: "sdk",
			Name: opts.Name,
		},
		Instance: mcpServer,
	}
}

// ToolResult helpers — convenience wrappers around mcp-go result constructors.

// NewToolResultText creates a tool result with a single text content block.
func NewToolResultText(text string) *mcp.CallToolResult {
	return &mcp.CallToolResult{
		Content: []mcp.Content{mcp.NewTextContent(text)},
	}
}

// NewToolResultError creates a tool result indicating an error.
func NewToolResultError(errMsg string) *mcp.CallToolResult {
	return &mcp.CallToolResult{
		Content: []mcp.Content{mcp.NewTextContent(errMsg)},
		IsError: true,
	}
}

// ToolHandlerFunc is the function signature for MCP tool handlers.
// Re-exported from mcp-go for convenience.
type ToolHandlerFunc = server.ToolHandlerFunc

// CallToolRequest is re-exported from mcp-go for convenience.
type CallToolRequest = mcp.CallToolRequest

// CallToolResult is re-exported from mcp-go for convenience.
type CallToolResult = mcp.CallToolResult

// ToolAnnotation is re-exported from mcp-go for convenience.
type ToolAnnotation = mcp.ToolAnnotation

// Ensure the handler is assignable via a compile-time check.
var _ ToolHandlerFunc = func(ctx context.Context, req CallToolRequest) (*CallToolResult, error) {
	return nil, nil
}
