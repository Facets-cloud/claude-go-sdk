package claudeagent

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
)

func TestTool_CreatesDefinition(t *testing.T) {
	td := Tool("get_weather", "Get weather for a city",
		mcp.WithString("city",
			mcp.Required(),
			mcp.Description("City name"),
		),
	)

	if td.Name != "get_weather" {
		t.Errorf("Name = %q, want %q", td.Name, "get_weather")
	}
	if td.Description != "Get weather for a city" {
		t.Errorf("Description = %q", td.Description)
	}
	if td.Tool.Name != "get_weather" {
		t.Errorf("Tool.Name = %q", td.Tool.Name)
	}
}

func TestTool_WithHandler(t *testing.T) {
	td := Tool("echo", "Echo input back")
	td.Handler = func(ctx context.Context, req CallToolRequest) (*CallToolResult, error) {
		args := req.GetArguments()
		text, _ := args["text"].(string)
		return NewToolResultText("echo: " + text), nil
	}

	if td.Handler == nil {
		t.Fatal("Handler should be set")
	}
}

func TestCreateSdkMcpServer_Basic(t *testing.T) {
	td := Tool("greet", "Say hello",
		mcp.WithString("name", mcp.Required()),
	)
	td.Handler = func(ctx context.Context, req CallToolRequest) (*CallToolResult, error) {
		name := req.GetArguments()["name"].(string)
		return NewToolResultText("Hello, " + name + "!"), nil
	}

	srv := CreateSdkMcpServer(CreateSdkMcpServerOptions{
		Name:  "greeting-server",
		Tools: []SdkMcpToolDefinition{td},
	})

	if srv.Name != "greeting-server" {
		t.Errorf("Name = %q", srv.Name)
	}
	if srv.Type != "sdk" {
		t.Errorf("Type = %q", srv.Type)
	}
	if srv.Instance == nil {
		t.Error("Instance should not be nil")
	}
}

func TestCreateSdkMcpServer_DefaultVersion(t *testing.T) {
	srv := CreateSdkMcpServer(CreateSdkMcpServerOptions{
		Name: "test",
	})
	if srv.Instance == nil {
		t.Error("Instance should not be nil")
	}
}

func TestCreateSdkMcpServer_CustomVersion(t *testing.T) {
	srv := CreateSdkMcpServer(CreateSdkMcpServerOptions{
		Name:    "test",
		Version: "2.0.0",
	})
	if srv.Instance == nil {
		t.Error("Instance should not be nil")
	}
}

func TestCreateSdkMcpServer_MultipleTools(t *testing.T) {
	t1 := Tool("tool1", "First tool")
	t1.Handler = func(ctx context.Context, req CallToolRequest) (*CallToolResult, error) {
		return NewToolResultText("t1"), nil
	}

	t2 := Tool("tool2", "Second tool",
		mcp.WithNumber("count"),
	)
	t2.Handler = func(ctx context.Context, req CallToolRequest) (*CallToolResult, error) {
		return NewToolResultText("t2"), nil
	}

	srv := CreateSdkMcpServer(CreateSdkMcpServerOptions{
		Name:  "multi",
		Tools: []SdkMcpToolDefinition{t1, t2},
	})

	if srv.Instance == nil {
		t.Error("Instance should not be nil")
	}
}

func TestCreateSdkMcpServer_SkipsNilHandler(t *testing.T) {
	td := Tool("no-handler", "Tool without handler")

	srv := CreateSdkMcpServer(CreateSdkMcpServerOptions{
		Name:  "test",
		Tools: []SdkMcpToolDefinition{td},
	})

	if srv.Instance == nil {
		t.Error("Instance should not be nil")
	}
}

func TestNewToolResultText(t *testing.T) {
	result := NewToolResultText("hello")
	if len(result.Content) != 1 {
		t.Fatalf("Content len = %d", len(result.Content))
	}
	if result.IsError {
		t.Error("should not be error")
	}
}

func TestNewToolResultError(t *testing.T) {
	result := NewToolResultError("something broke")
	if len(result.Content) != 1 {
		t.Fatalf("Content len = %d", len(result.Content))
	}
	if !result.IsError {
		t.Error("should be error")
	}
}

func TestToolHandlerFunc_Assignable(t *testing.T) {
	var handler ToolHandlerFunc = func(ctx context.Context, req CallToolRequest) (*CallToolResult, error) {
		return NewToolResultText("ok"), nil
	}
	_ = handler
}

func TestMcpSdkServerConfigWithInstance_SerializesAsSdkType(t *testing.T) {
	// Verify that McpSdkServerConfigWithInstance serializes to {"type":"sdk","name":"X"}
	// The Instance field is NOT serialized — it's a live object.
	srv := CreateSdkMcpServer(CreateSdkMcpServerOptions{
		Name: "my-server",
	})

	data, err := json.Marshal(srv)
	if err != nil {
		t.Fatal(err)
	}

	var parsed map[string]interface{}
	json.Unmarshal(data, &parsed)

	if parsed["type"] != "sdk" {
		t.Errorf("type = %v, want 'sdk'", parsed["type"])
	}
	if parsed["name"] != "my-server" {
		t.Errorf("name = %v, want 'my-server'", parsed["name"])
	}
	// Instance should NOT appear in JSON
	if _, exists := parsed["instance"]; exists {
		t.Error("Instance should not be serialized to JSON")
	}
}

func TestMcpSdkServerConfigWithInstance_NotSuitableForPrintMode(t *testing.T) {
	// IMPORTANT: In --print mode, McpSdkServerConfigWithInstance cannot work
	// because the CLI doesn't know how to connect to an in-process server.
	//
	// The CLI receives {"type":"sdk","name":"X"} via --mcp-config which it
	// can't use — it expects stdio/sse/http configs.
	//
	// True in-process MCP requires --input-format stream-json with the
	// bidirectional control protocol (initialize with sdkMcpServers, then
	// mcp_message control requests for JSON-RPC routing).
	//
	// For --print mode, use McpStdioServerConfig with an external server
	// binary instead. See demos/mcp-inline/ for the pattern.

	srv := CreateSdkMcpServer(CreateSdkMcpServerOptions{
		Name: "test-server",
	})

	// When serialized for --mcp-config, the Instance is lost
	mcpConfig := map[string]interface{}{
		"mcpServers": map[string]interface{}{
			"test-server": srv,
		},
	}
	data, _ := json.Marshal(mcpConfig)
	configStr := string(data)

	// The JSON won't contain any callable server info — just {"type":"sdk","name":"test-server"}
	if !contains(configStr, `"type":"sdk"`) {
		t.Error("should contain sdk type")
	}
	// This config will NOT work with the CLI in --print mode
	// The test documents this known limitation
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
