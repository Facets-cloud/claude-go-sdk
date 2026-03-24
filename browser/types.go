package browser

import (
	"context"

	claudeagent "github.com/anthropics/claude-agent-sdk-go"
)

// OAuthCredential represents an OAuth token for WebSocket authentication.
type OAuthCredential struct {
	Type  string `json:"type"` // "oauth"
	Token string `json:"token"`
}

// AuthMessage is an authentication message sent over the WebSocket.
type AuthMessage struct {
	Type       string          `json:"type"` // "auth"
	Credential OAuthCredential `json:"credential"`
}

// WebSocketOptions configures the WebSocket connection.
type WebSocketOptions struct {
	URL         string            `json:"url"`
	Headers     map[string]string `json:"headers,omitempty"`
	AuthMessage *AuthMessage      `json:"authMessage,omitempty"`
}

// BrowserQueryOptions configures a browser-based query via WebSocket.
type BrowserQueryOptions struct {
	// Prompt is a channel of user messages for multi-turn conversations.
	Prompt <-chan claudeagent.SDKUserMessage

	// WebSocket configures the WebSocket connection.
	WebSocket WebSocketOptions

	// AbortContext cancels the query when done.
	AbortContext context.Context

	// CanUseTool is a custom permission handler.
	CanUseTool claudeagent.CanUseTool

	// Hooks are callbacks for lifecycle events.
	Hooks map[claudeagent.HookEvent][]claudeagent.HookCallbackMatcher

	// McpServers configures MCP servers keyed by name.
	McpServers map[string]interface{}

	// JSONSchema configures structured output.
	JSONSchema map[string]interface{}
}
