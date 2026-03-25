package claudeagent

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
)

// --- Top-level Control Envelope Types ---

// SDKControlRequest is the outer envelope for all control requests.
type SDKControlRequest struct {
	Type      string          `json:"type"`       // "control_request"
	RequestID string          `json:"request_id"`
	Request   json.RawMessage `json:"request"`
}

// SDKControlResponse is the outer envelope for all control responses.
type SDKControlResponse struct {
	Type     string          `json:"type"` // "control_response"
	Response json.RawMessage `json:"response"`
}

// SDKControlCancelRequest cancels a pending control request.
// NOTE: This has type "control_cancel_request" (NOT "control_request").
type SDKControlCancelRequest struct {
	Type      string `json:"type"`       // "control_cancel_request"
	RequestID string `json:"request_id"`
}

// ControlResponse is a successful control response.
type ControlResponse struct {
	Subtype   string                 `json:"subtype"`    // "success"
	RequestID string                 `json:"request_id"`
	Response  map[string]interface{} `json:"response,omitempty"`
}

// ControlErrorResponse is an error control response.
type ControlErrorResponse struct {
	Subtype                   string              `json:"subtype"`    // "error"
	RequestID                 string              `json:"request_id"`
	Error                     string              `json:"error"`
	PendingPermissionRequests []SDKControlRequest `json:"pending_permission_requests,omitempty"`
}

// --- Control Request Inner Types ---

// SDKControlInterruptRequest interrupts the currently running conversation turn.
type SDKControlInterruptRequest struct {
	Subtype string `json:"subtype"` // "interrupt"
}

// SDKControlPermissionRequest requests permission to use a tool.
type SDKControlPermissionRequest struct {
	Subtype               string                 `json:"subtype"` // "can_use_tool"
	ToolName              string                 `json:"tool_name"`
	Input                 map[string]interface{} `json:"input"`
	PermissionSuggestions json.RawMessage        `json:"permission_suggestions,omitempty"` // []PermissionUpdate
	BlockedPath           *string                `json:"blocked_path,omitempty"`
	DecisionReason        *string                `json:"decision_reason,omitempty"`
	Title                 *string                `json:"title,omitempty"`
	DisplayName           *string                `json:"display_name,omitempty"`
	ToolUseID             string                 `json:"tool_use_id"`
	AgentID               *string                `json:"agent_id,omitempty"`
	Description           *string                `json:"description,omitempty"`
}

// SDKControlInitializeRequest initializes the SDK session.
type SDKControlInitializeRequest struct {
	Subtype                string                 `json:"subtype"` // "initialize"
	ProtocolVersion        int                    `json:"protocolVersion,omitempty"`
	CanUseTool             bool                   `json:"canUseTool,omitempty"`
	HasHooks               bool                   `json:"hasHooks,omitempty"`
	HasElicitation         bool                   `json:"hasElicitation,omitempty"`
	Hooks                  map[string]interface{} `json:"hooks,omitempty"`
	SdkMcpServers          []string               `json:"sdkMcpServers,omitempty"`
	JSONSchema             map[string]interface{} `json:"jsonSchema,omitempty"`
	SystemPrompt           *string                `json:"systemPrompt,omitempty"`
	AppendSystemPrompt     *string                `json:"appendSystemPrompt,omitempty"`
	Agents                 map[string]interface{} `json:"agents,omitempty"`
	PromptSuggestions      *bool                  `json:"promptSuggestions,omitempty"`
	AgentProgressSummaries *bool                  `json:"agentProgressSummaries,omitempty"`
}

// SDKControlInitializeResponse is the response from session initialization.
type SDKControlInitializeResponse struct {
	Commands              []SlashCommand `json:"commands"`
	Agents                []AgentInfo    `json:"agents"`
	OutputStyle           string         `json:"output_style"`
	AvailableOutputStyles []string       `json:"available_output_styles"`
	Models                []ModelInfo    `json:"models"`
	Account               *AccountInfo   `json:"account"`
	FastModeState         *FastModeState `json:"fast_mode_state,omitempty"`
}

// SDKControlSetPermissionModeRequest sets the permission mode.
type SDKControlSetPermissionModeRequest struct {
	Subtype string         `json:"subtype"` // "set_permission_mode"
	Mode    PermissionMode `json:"mode"`
}

// SDKControlSetModelRequest sets the model.
// Model is *string: nil reverts to the default model.
type SDKControlSetModelRequest struct {
	Subtype string  `json:"subtype"` // "set_model"
	Model   *string `json:"model,omitempty"`
}

// SDKControlSetMaxThinkingTokensRequest sets thinking token budget.
// MaxThinkingTokens is *int: nil clears the limit (reverts to default).
type SDKControlSetMaxThinkingTokensRequest struct {
	Subtype           string `json:"subtype"` // "set_max_thinking_tokens"
	MaxThinkingTokens *int   `json:"max_thinking_tokens"`
}

// SDKControlMcpStatusRequest requests MCP server statuses.
type SDKControlMcpStatusRequest struct {
	Subtype string `json:"subtype"` // "mcp_status"
}

// SDKControlMcpMessageRequest sends a JSON-RPC message to an MCP server.
type SDKControlMcpMessageRequest struct {
	Subtype    string          `json:"subtype"` // "mcp_message"
	ServerName string          `json:"server_name"`
	Message    json.RawMessage `json:"message"` // JSONRPCMessage
}

// SDKControlMcpSetServersRequest replaces dynamically managed MCP servers.
type SDKControlMcpSetServersRequest struct {
	Subtype string                 `json:"subtype"` // "mcp_set_servers"
	Servers map[string]interface{} `json:"servers"`
}

// SDKControlMcpReconnectRequest reconnects a disconnected MCP server.
type SDKControlMcpReconnectRequest struct {
	Subtype    string `json:"subtype"` // "mcp_reconnect"
	ServerName string `json:"serverName"`
}

// SDKControlMcpToggleRequest enables or disables an MCP server.
type SDKControlMcpToggleRequest struct {
	Subtype    string `json:"subtype"` // "mcp_toggle"
	ServerName string `json:"serverName"`
	Enabled    bool   `json:"enabled"`
}

// SDKControlRewindFilesRequest rewinds file changes since a specific user message.
type SDKControlRewindFilesRequest struct {
	Subtype       string `json:"subtype"` // "rewind_files"
	UserMessageID string `json:"user_message_id"`
	DryRun        *bool  `json:"dry_run,omitempty"`
}

// SDKControlStopTaskRequest stops a running task.
type SDKControlStopTaskRequest struct {
	Subtype string `json:"subtype"` // "stop_task"
	TaskID  string `json:"task_id"`
}

// SDKControlApplyFlagSettingsRequest merges settings into the flag settings layer.
type SDKControlApplyFlagSettingsRequest struct {
	Subtype  string                 `json:"subtype"` // "apply_flag_settings"
	Settings map[string]interface{} `json:"settings,omitempty"`
}

// SDKControlGetSettingsRequest returns effective merged settings.
type SDKControlGetSettingsRequest struct {
	Subtype string `json:"subtype"` // "get_settings"
}

// SDKControlElicitationRequest handles MCP elicitation.
type SDKControlElicitationRequest struct {
	Subtype         string                 `json:"subtype"` // "elicitation"
	McpServerName   string                 `json:"mcp_server_name"`
	Message         string                 `json:"message"`
	Mode            *string                `json:"mode,omitempty"` // "form" | "url"
	URL             *string                `json:"url,omitempty"`
	ElicitationID   *string                `json:"elicitation_id,omitempty"`
	RequestedSchema map[string]interface{} `json:"requested_schema,omitempty"`
}

// SDKControlCancelAsyncMessageRequest cancels a pending async user message.
type SDKControlCancelAsyncMessageRequest struct {
	Subtype     string `json:"subtype"` // "cancel_async_message"
	MessageUUID string `json:"message_uuid"`
}

// SDKControlEndSessionRequest ends the current session.
type SDKControlEndSessionRequest struct {
	Subtype string `json:"subtype"` // "end_session"
}

// --- Hook Callback Types ---

// SDKHookCallbackMatcher matches hook events and invokes callbacks.
type SDKHookCallbackMatcher struct {
	Matcher *string         `json:"matcher,omitempty"`
	Hooks   json.RawMessage `json:"hooks"` // array of hook callback functions (opaque in Go)
}

// SDKHookCallbackRequest is a control request for a hook callback.
type SDKHookCallbackRequest struct {
	Subtype    string          `json:"subtype"` // "hook_callback"
	CallbackID string          `json:"callback_id"`
	Input      json.RawMessage `json:"input"`
	ToolUseID  *string         `json:"tool_use_id,omitempty"`
}

// --- Missing Control Request Types (Task 29 addendum) ---

// SDKControlMcpAuthenticateRequest initiates MCP server authentication.
type SDKControlMcpAuthenticateRequest struct {
	Subtype    string `json:"subtype"` // "mcp_authenticate"
	ServerName string `json:"serverName"`
}

// SDKControlMcpClearAuthRequest clears MCP server auth credentials.
type SDKControlMcpClearAuthRequest struct {
	Subtype    string `json:"subtype"` // "mcp_clear_auth"
	ServerName string `json:"serverName"`
}

// SDKControlMcpOAuthCallbackUrlRequest provides an OAuth callback URL.
type SDKControlMcpOAuthCallbackUrlRequest struct {
	Subtype     string `json:"subtype"` // "mcp_oauth_callback_url"
	ServerName  string `json:"serverName"`
	CallbackUrl string `json:"callbackUrl"`
}

// SDKControlClaudeAuthenticateRequest initiates Claude authentication.
type SDKControlClaudeAuthenticateRequest struct {
	Subtype string `json:"subtype"` // "claude_authenticate"
}

// SDKControlClaudeOAuthCallbackRequest provides a Claude OAuth callback.
type SDKControlClaudeOAuthCallbackRequest struct {
	Subtype     string `json:"subtype"` // "claude_oauth_callback"
	CallbackUrl string `json:"callbackUrl"`
}

// SDKControlClaudeOAuthWaitForCompletionRequest waits for OAuth to complete.
type SDKControlClaudeOAuthWaitForCompletionRequest struct {
	Subtype string `json:"subtype"` // "claude_oauth_wait_for_completion"
}

// SDKControlRemoteControlRequest sends a remote control command.
type SDKControlRemoteControlRequest struct {
	Subtype string                 `json:"subtype"` // "remote_control"
	Action  string                 `json:"action"`
	Data    map[string]interface{} `json:"data,omitempty"`
}

// SDKControlSetProactiveRequest toggles proactive behavior.
type SDKControlSetProactiveRequest struct {
	Subtype   string `json:"subtype"` // "set_proactive"
	Proactive bool   `json:"proactive"`
}

// SDKControlGenerateSessionTitleRequest requests AI-generated session title.
type SDKControlGenerateSessionTitleRequest struct {
	Subtype string `json:"subtype"` // "generate_session_title"
}

// SDKControlSideQuestionRequest sends a side question outside the main turn.
type SDKControlSideQuestionRequest struct {
	Subtype  string `json:"subtype"` // "side_question"
	Question string `json:"question"`
}

// --- Parsing Helpers ---

// ParseControlRequestSubtype extracts the subtype from a raw control request inner JSON.
func ParseControlRequestSubtype(data json.RawMessage) (string, error) {
	var env struct {
		Subtype string `json:"subtype"`
	}
	if err := json.Unmarshal(data, &env); err != nil {
		return "", fmt.Errorf("parse control request subtype: %w", err)
	}
	return env.Subtype, nil
}

// ParseControlResponseSubtype extracts the subtype from a raw control response JSON.
func ParseControlResponseSubtype(data json.RawMessage) (string, error) {
	var env struct {
		Subtype string `json:"subtype"`
	}
	if err := json.Unmarshal(data, &env); err != nil {
		return "", fmt.Errorf("parse control response subtype: %w", err)
	}
	return env.Subtype, nil
}

// --- Correlation Engine ---

// CorrelationEngine tracks pending control requests and delivers responses
// to the correct caller via channels keyed by request_id.
type CorrelationEngine struct {
	mu       sync.Mutex
	pending  map[string]chan *SDKControlResponse
	closed   bool
}

// NewCorrelationEngine creates a new correlation engine.
func NewCorrelationEngine() *CorrelationEngine {
	return &CorrelationEngine{
		pending: make(map[string]chan *SDKControlResponse),
	}
}

// Register creates a response channel for the given request ID.
// The caller should read from the returned channel to receive the response.
func (e *CorrelationEngine) Register(requestID string) <-chan *SDKControlResponse {
	ch := make(chan *SDKControlResponse, 1)
	e.mu.Lock()
	defer e.mu.Unlock()
	if e.closed {
		close(ch)
		return ch
	}
	e.pending[requestID] = ch
	return ch
}

// Deliver sends a response to the channel registered for the given request ID.
// Returns true if a pending request was found and delivered, false otherwise.
func (e *CorrelationEngine) Deliver(requestID string, resp *SDKControlResponse) bool {
	e.mu.Lock()
	ch, ok := e.pending[requestID]
	if ok {
		delete(e.pending, requestID)
	}
	e.mu.Unlock()

	if !ok {
		return false
	}
	ch <- resp
	close(ch)
	return true
}

// Close closes all pending channels and prevents new registrations.
func (e *CorrelationEngine) Close() {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.closed = true
	for id, ch := range e.pending {
		close(ch)
		delete(e.pending, id)
	}
}

// WaitForResponse waits for a control response on the given channel,
// respecting context cancellation.
func WaitForResponse(ctx context.Context, ch <-chan *SDKControlResponse) (*SDKControlResponse, error) {
	select {
	case resp, ok := <-ch:
		if !ok {
			return nil, fmt.Errorf("response channel closed")
		}
		return resp, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}
