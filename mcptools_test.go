package claudeagent

import (
	"context"
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
	// Should not panic with empty tools
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
	// Handler is nil — should not panic

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
	// Verify the re-exported type is usable
	var handler ToolHandlerFunc = func(ctx context.Context, req CallToolRequest) (*CallToolResult, error) {
		return NewToolResultText("ok"), nil
	}
	_ = handler
}
