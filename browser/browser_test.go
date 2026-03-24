package browser

import (
	"context"
	"encoding/json"
	"testing"

	claudeagent "github.com/anthropics/claude-agent-sdk-go"
)

func TestOAuthCredential_JSON(t *testing.T) {
	cred := OAuthCredential{
		Type:  "oauth",
		Token: "test-token-123",
	}
	data, err := json.Marshal(cred)
	if err != nil {
		t.Fatal(err)
	}
	var got OAuthCredential
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatal(err)
	}
	if got.Type != "oauth" {
		t.Errorf("Type = %q, want 'oauth'", got.Type)
	}
	if got.Token != "test-token-123" {
		t.Errorf("Token = %q, want 'test-token-123'", got.Token)
	}
}

func TestAuthMessage_JSON(t *testing.T) {
	msg := AuthMessage{
		Type: "auth",
		Credential: OAuthCredential{
			Type:  "oauth",
			Token: "my-token",
		},
	}
	data, err := json.Marshal(msg)
	if err != nil {
		t.Fatal(err)
	}
	var got AuthMessage
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatal(err)
	}
	if got.Type != "auth" {
		t.Errorf("Type = %q, want 'auth'", got.Type)
	}
	if got.Credential.Token != "my-token" {
		t.Errorf("Credential.Token = %q", got.Credential.Token)
	}
}

func TestWebSocketOptions_JSON(t *testing.T) {
	raw := `{"url":"wss://example.com/ws","headers":{"Authorization":"Bearer tok"},"authMessage":{"type":"auth","credential":{"type":"oauth","token":"t"}}}`
	var opts WebSocketOptions
	if err := json.Unmarshal([]byte(raw), &opts); err != nil {
		t.Fatal(err)
	}
	if opts.URL != "wss://example.com/ws" {
		t.Errorf("URL = %q", opts.URL)
	}
	if opts.Headers["Authorization"] != "Bearer tok" {
		t.Errorf("Headers = %v", opts.Headers)
	}
	if opts.AuthMessage == nil {
		t.Fatal("AuthMessage is nil")
	}
	if opts.AuthMessage.Credential.Token != "t" {
		t.Errorf("AuthMessage.Credential.Token = %q", opts.AuthMessage.Credential.Token)
	}
}

func TestWebSocketOptions_MinimalJSON(t *testing.T) {
	raw := `{"url":"wss://example.com"}`
	var opts WebSocketOptions
	if err := json.Unmarshal([]byte(raw), &opts); err != nil {
		t.Fatal(err)
	}
	if opts.URL != "wss://example.com" {
		t.Errorf("URL = %q", opts.URL)
	}
	if opts.Headers != nil {
		t.Errorf("expected nil headers, got %v", opts.Headers)
	}
	if opts.AuthMessage != nil {
		t.Error("expected nil AuthMessage")
	}
}

func TestBrowserQueryOptions_Fields(t *testing.T) {
	// Verify BrowserQueryOptions can be constructed with all fields.
	opts := BrowserQueryOptions{
		Prompt: make(<-chan claudeagent.SDKUserMessage),
		WebSocket: WebSocketOptions{
			URL: "wss://example.com/claude",
		},
		McpServers: map[string]interface{}{
			"server1": map[string]interface{}{"command": "node"},
		},
		JSONSchema: map[string]interface{}{
			"type": "object",
		},
	}
	if opts.WebSocket.URL != "wss://example.com/claude" {
		t.Errorf("WebSocket.URL = %q", opts.WebSocket.URL)
	}
	if opts.McpServers == nil {
		t.Error("McpServers is nil")
	}
	if opts.JSONSchema == nil {
		t.Error("JSONSchema is nil")
	}
}

func TestBrowserQueryOptions_WithCanUseTool(t *testing.T) {
	opts := BrowserQueryOptions{
		WebSocket: WebSocketOptions{URL: "wss://example.com"},
		CanUseTool: func(ctx context.Context, toolName string, input map[string]interface{}, canUseOpts claudeagent.CanUseToolOptions) (claudeagent.PermissionResult, error) {
			return claudeagent.PermissionResultAllow{
				Behavior: claudeagent.PermissionBehaviorAllow,
			}, nil
		},
	}
	if opts.CanUseTool == nil {
		t.Error("CanUseTool is nil")
	}
}

func TestBrowserQueryOptions_WithHooks(t *testing.T) {
	opts := BrowserQueryOptions{
		WebSocket: WebSocketOptions{URL: "wss://example.com"},
		Hooks: map[claudeagent.HookEvent][]claudeagent.HookCallbackMatcher{
			claudeagent.HookEventPreToolUse: {},
		},
	}
	if opts.Hooks == nil {
		t.Error("Hooks is nil")
	}
	if _, ok := opts.Hooks[claudeagent.HookEventPreToolUse]; !ok {
		t.Error("missing PreToolUse hook")
	}
}
