package claudeagent

import "encoding/json"

// SDKMessage is the interface implemented by all message types streamed from the SDK.
// Use a type switch to handle specific message types.
type SDKMessage interface {
	sdkMessage() // unexported marker method
	// MessageType returns the wire "type" field value.
	MessageType() string
}

// --- Usage types ---

// NonNullableUsage contains token usage information with all fields non-nullable.
type NonNullableUsage struct {
	InputTokens              int            `json:"input_tokens"`
	OutputTokens             int            `json:"output_tokens"`
	CacheCreationInputTokens int            `json:"cache_creation_input_tokens"`
	CacheReadInputTokens     int            `json:"cache_read_input_tokens"`
	ServerToolUse            *ServerToolUse `json:"server_tool_use"`
	ServiceTier              *string        `json:"service_tier"`
	CacheCreation            *CacheCreation `json:"cache_creation"`
}

// ServerToolUse contains server-side tool use counts.
type ServerToolUse struct {
	WebSearchRequests int `json:"web_search_requests"`
	WebFetchRequests  int `json:"web_fetch_requests"`
}

// CacheCreation contains cache creation token counts.
type CacheCreation struct {
	Ephemeral1hInputTokens int `json:"ephemeral_1h_input_tokens"`
	Ephemeral5mInputTokens int `json:"ephemeral_5m_input_tokens"`
}

// ModelUsage contains per-model usage statistics.
type ModelUsage struct {
	InputTokens              int     `json:"inputTokens"`
	OutputTokens             int     `json:"outputTokens"`
	CacheReadInputTokens     int     `json:"cacheReadInputTokens"`
	CacheCreationInputTokens int     `json:"cacheCreationInputTokens"`
	WebSearchRequests        int     `json:"webSearchRequests"`
	CostUSD                  float64 `json:"costUSD"`
	ContextWindow            int     `json:"contextWindow"`
	MaxOutputTokens          int     `json:"maxOutputTokens"`
}

// SDKPermissionDenial records a denied tool use.
type SDKPermissionDenial struct {
	ToolName  string                 `json:"tool_name"`
	ToolUseID string                 `json:"tool_use_id"`
	ToolInput map[string]interface{} `json:"tool_input"`
}

// --- Assistant Message ---

// SDKAssistantMessage represents an assistant response message.
type SDKAssistantMessage struct {
	Type            string                    `json:"type"` // "assistant"
	Message         json.RawMessage           `json:"message"`
	ParentToolUseID *string                   `json:"parent_tool_use_id"`
	Error           *SDKAssistantMessageError `json:"error,omitempty"`
	UUID            string                    `json:"uuid"`
	SessionID       string                    `json:"session_id"`
}

func (m *SDKAssistantMessage) sdkMessage()        {}
func (m *SDKAssistantMessage) MessageType() string { return "assistant" }

// --- User Messages ---

// SDKUserMessage represents a user input message.
type SDKUserMessage struct {
	Type            string          `json:"type"` // "user"
	Message         json.RawMessage `json:"message"`
	ParentToolUseID *string         `json:"parent_tool_use_id"`
	IsSynthetic     *bool           `json:"isSynthetic,omitempty"`
	ToolUseResult   interface{}     `json:"tool_use_result,omitempty"`
	Priority        *string         `json:"priority,omitempty"` // "now" | "next" | "later"
	Timestamp       *string         `json:"timestamp,omitempty"`
	UUID            *string         `json:"uuid,omitempty"`
	SessionID       string          `json:"session_id"`
}

func (m *SDKUserMessage) sdkMessage()        {}
func (m *SDKUserMessage) MessageType() string { return "user" }

// SDKUserMessageReplay represents a replayed user message (from session history).
type SDKUserMessageReplay struct {
	Type            string          `json:"type"` // "user"
	Message         json.RawMessage `json:"message"`
	ParentToolUseID *string         `json:"parent_tool_use_id"`
	IsSynthetic     *bool           `json:"isSynthetic,omitempty"`
	ToolUseResult   interface{}     `json:"tool_use_result,omitempty"`
	Priority        *string         `json:"priority,omitempty"`
	Timestamp       *string         `json:"timestamp,omitempty"`
	UUID            string          `json:"uuid"`
	SessionID       string          `json:"session_id"`
	IsReplay        bool            `json:"isReplay"` // always true
}

func (m *SDKUserMessageReplay) sdkMessage()        {}
func (m *SDKUserMessageReplay) MessageType() string { return "user" }

// --- Result Messages ---

// SDKResultSuccess represents a successful query result.
type SDKResultSuccess struct {
	Type              string                `json:"type"`    // "result"
	Subtype           string                `json:"subtype"` // "success"
	DurationMs        int                   `json:"duration_ms"`
	DurationAPIMs     int                   `json:"duration_api_ms"`
	IsError           bool                  `json:"is_error"`
	NumTurns          int                   `json:"num_turns"`
	Result            string                `json:"result"`
	StopReason        *string               `json:"stop_reason"`
	TotalCostUSD      float64               `json:"total_cost_usd"`
	Usage             NonNullableUsage      `json:"usage"`
	ModelUsageMap     map[string]ModelUsage `json:"modelUsage"`
	PermissionDenials []SDKPermissionDenial `json:"permission_denials"`
	StructuredOutput  interface{}           `json:"structured_output,omitempty"`
	FastModeState     *FastModeState        `json:"fast_mode_state,omitempty"`
	UUID              string                `json:"uuid"`
	SessionID         string                `json:"session_id"`
}

func (m *SDKResultSuccess) sdkMessage()        {}
func (m *SDKResultSuccess) MessageType() string { return "result" }

// SDKResultError represents a failed query result.
type SDKResultError struct {
	Type              string                `json:"type"`    // "result"
	Subtype           string                `json:"subtype"` // "error_during_execution" | "error_max_turns" | "error_max_budget_usd" | "error_max_structured_output_retries"
	DurationMs        int                   `json:"duration_ms"`
	DurationAPIMs     int                   `json:"duration_api_ms"`
	IsError           bool                  `json:"is_error"`
	NumTurns          int                   `json:"num_turns"`
	StopReason        *string               `json:"stop_reason"`
	TotalCostUSD      float64               `json:"total_cost_usd"`
	Usage             NonNullableUsage      `json:"usage"`
	ModelUsageMap     map[string]ModelUsage `json:"modelUsage"`
	PermissionDenials []SDKPermissionDenial `json:"permission_denials"`
	Errors            []string              `json:"errors"`
	FastModeState     *FastModeState        `json:"fast_mode_state,omitempty"`
	UUID              string                `json:"uuid"`
	SessionID         string                `json:"session_id"`
}

func (m *SDKResultError) sdkMessage()        {}
func (m *SDKResultError) MessageType() string { return "result" }

// --- System Messages ---

// McpServerRef is a reference to an MCP server in the system init message.
type McpServerRef struct {
	Name   string `json:"name"`
	Status string `json:"status"`
}

// PluginRef is a reference to a plugin in the system init message.
type PluginRef struct {
	Name string `json:"name"`
	Path string `json:"path"`
}

// SDKSystemMessage represents the system initialization message.
type SDKSystemMessage struct {
	Type              string         `json:"type"`    // "system"
	Subtype           string         `json:"subtype"` // "init"
	Agents            []string       `json:"agents,omitempty"`
	ApiKeySource      ApiKeySource   `json:"apiKeySource"`
	Betas             []string       `json:"betas,omitempty"`
	ClaudeCodeVersion string         `json:"claude_code_version"`
	Cwd               string         `json:"cwd"`
	Tools             []string       `json:"tools"`
	McpServers        []McpServerRef `json:"mcp_servers"`
	Model             string         `json:"model"`
	PermissionMode    PermissionMode `json:"permissionMode"`
	SlashCommands     []string       `json:"slash_commands"`
	OutputStyle       string         `json:"output_style"`
	Skills            []string       `json:"skills"`
	Plugins           []PluginRef    `json:"plugins"`
	FastModeState     *FastModeState `json:"fast_mode_state,omitempty"`
	UUID              string         `json:"uuid"`
	SessionID         string         `json:"session_id"`
}

func (m *SDKSystemMessage) sdkMessage()        {}
func (m *SDKSystemMessage) MessageType() string { return "system" }

// --- Status Message ---

// SDKStatusMessage represents a system status update.
type SDKStatusMessage struct {
	Type           string          `json:"type"`    // "system"
	Subtype        string          `json:"subtype"` // "status"
	Status         *string         `json:"status"`  // "compacting" | null
	PermissionMode *PermissionMode `json:"permissionMode,omitempty"`
	UUID           string          `json:"uuid"`
	SessionID      string          `json:"session_id"`
}

func (m *SDKStatusMessage) sdkMessage()        {}
func (m *SDKStatusMessage) MessageType() string { return "system" }

// --- API Retry ---

// SDKAPIRetryMessage represents an API retry notification.
type SDKAPIRetryMessage struct {
	Type         string                    `json:"type"`    // "system"
	Subtype      string                    `json:"subtype"` // "api_retry"
	Attempt      int                       `json:"attempt"`
	MaxRetries   int                       `json:"max_retries"`
	RetryDelayMs int                       `json:"retry_delay_ms"`
	ErrorStatus  *int                      `json:"error_status"`
	Error        *SDKAssistantMessageError `json:"error"`
	UUID         string                    `json:"uuid"`
	SessionID    string                    `json:"session_id"`
}

func (m *SDKAPIRetryMessage) sdkMessage()        {}
func (m *SDKAPIRetryMessage) MessageType() string { return "system" }

// --- Compact Boundary ---

// CompactMetadata contains metadata about a context compaction event.
type CompactMetadata struct {
	Trigger          string            `json:"trigger"` // "manual" | "auto"
	PreTokens        int               `json:"pre_tokens"`
	PreservedSegment *PreservedSegment `json:"preserved_segment,omitempty"`
}

// PreservedSegment identifies the preserved message range during compaction.
type PreservedSegment struct {
	HeadUUID   string `json:"head_uuid"`
	AnchorUUID string `json:"anchor_uuid"`
	TailUUID   string `json:"tail_uuid"`
}

// SDKCompactBoundaryMessage represents a context compaction boundary.
type SDKCompactBoundaryMessage struct {
	Type            string          `json:"type"`    // "system"
	Subtype         string          `json:"subtype"` // "compact_boundary"
	CompactMetadata CompactMetadata `json:"compact_metadata"`
	UUID            string          `json:"uuid"`
	SessionID       string          `json:"session_id"`
}

func (m *SDKCompactBoundaryMessage) sdkMessage()        {}
func (m *SDKCompactBoundaryMessage) MessageType() string { return "system" }

// --- Local Command Output ---

// SDKLocalCommandOutputMessage represents output from a local shell command.
type SDKLocalCommandOutputMessage struct {
	Type      string `json:"type"`    // "system"
	Subtype   string `json:"subtype"` // "local_command_output"
	Content   string `json:"content"`
	UUID      string `json:"uuid"`
	SessionID string `json:"session_id"`
}

func (m *SDKLocalCommandOutputMessage) sdkMessage()        {}
func (m *SDKLocalCommandOutputMessage) MessageType() string { return "system" }

// --- Hook Messages ---

// SDKHookStartedMessage represents a hook execution start event.
type SDKHookStartedMessage struct {
	Type      string `json:"type"`    // "system"
	Subtype   string `json:"subtype"` // "hook_started"
	HookID    string `json:"hook_id"`
	HookName  string `json:"hook_name"`
	HookEvent string `json:"hook_event"`
	UUID      string `json:"uuid"`
	SessionID string `json:"session_id"`
}

func (m *SDKHookStartedMessage) sdkMessage()        {}
func (m *SDKHookStartedMessage) MessageType() string { return "system" }

// SDKHookProgressMessage represents hook execution progress.
type SDKHookProgressMessage struct {
	Type      string `json:"type"`    // "system"
	Subtype   string `json:"subtype"` // "hook_progress"
	HookID    string `json:"hook_id"`
	HookName  string `json:"hook_name"`
	HookEvent string `json:"hook_event"`
	Stdout    string `json:"stdout"`
	Stderr    string `json:"stderr"`
	Output    string `json:"output"`
	UUID      string `json:"uuid"`
	SessionID string `json:"session_id"`
}

func (m *SDKHookProgressMessage) sdkMessage()        {}
func (m *SDKHookProgressMessage) MessageType() string { return "system" }

// SDKHookResponseMessage represents the final hook execution result.
type SDKHookResponseMessage struct {
	Type      string `json:"type"`    // "system"
	Subtype   string `json:"subtype"` // "hook_response"
	HookID    string `json:"hook_id"`
	HookName  string `json:"hook_name"`
	HookEvent string `json:"hook_event"`
	Output    string `json:"output"`
	Stdout    string `json:"stdout"`
	Stderr    string `json:"stderr"`
	ExitCode  *int   `json:"exit_code,omitempty"`
	Outcome   string `json:"outcome"` // "success" | "error" | "cancelled"
	UUID      string `json:"uuid"`
	SessionID string `json:"session_id"`
}

func (m *SDKHookResponseMessage) sdkMessage()        {}
func (m *SDKHookResponseMessage) MessageType() string { return "system" }

// --- Stream Event (Partial Assistant Message) ---

// SDKPartialAssistantMessage represents a streaming content delta.
type SDKPartialAssistantMessage struct {
	Type            string          `json:"type"` // "stream_event"
	Event           json.RawMessage `json:"event"`
	ParentToolUseID *string         `json:"parent_tool_use_id"`
	UUID            string          `json:"uuid"`
	SessionID       string          `json:"session_id"`
}

func (m *SDKPartialAssistantMessage) sdkMessage()        {}
func (m *SDKPartialAssistantMessage) MessageType() string { return "stream_event" }

// --- Tool Progress ---

// SDKToolProgressMessage represents progress on an active tool execution.
type SDKToolProgressMessage struct {
	Type               string  `json:"type"` // "tool_progress"
	ToolUseID          string  `json:"tool_use_id"`
	ToolName           string  `json:"tool_name"`
	ParentToolUseID    *string `json:"parent_tool_use_id"`
	ElapsedTimeSeconds float64 `json:"elapsed_time_seconds"`
	TaskID             *string `json:"task_id,omitempty"`
	UUID               string  `json:"uuid"`
	SessionID          string  `json:"session_id"`
}

func (m *SDKToolProgressMessage) sdkMessage()        {}
func (m *SDKToolProgressMessage) MessageType() string { return "tool_progress" }

// --- Tool Use Summary ---

// SDKToolUseSummaryMessage represents a summary of completed tool uses.
type SDKToolUseSummaryMessage struct {
	Type                string   `json:"type"` // "tool_use_summary"
	Summary             string   `json:"summary"`
	PrecedingToolUseIDs []string `json:"preceding_tool_use_ids"`
	UUID                string   `json:"uuid"`
	SessionID           string   `json:"session_id"`
}

func (m *SDKToolUseSummaryMessage) sdkMessage()        {}
func (m *SDKToolUseSummaryMessage) MessageType() string { return "tool_use_summary" }

// --- Auth Status ---

// SDKAuthStatusMessage represents an authentication status update.
type SDKAuthStatusMessage struct {
	Type             string   `json:"type"` // "auth_status"
	IsAuthenticating bool     `json:"isAuthenticating"`
	Output           []string `json:"output"`
	Error            *string  `json:"error,omitempty"`
	UUID             string   `json:"uuid"`
	SessionID        string   `json:"session_id"`
}

func (m *SDKAuthStatusMessage) sdkMessage()        {}
func (m *SDKAuthStatusMessage) MessageType() string { return "auth_status" }

// --- Task Messages ---

// TaskUsage contains usage statistics for a subagent task.
type TaskUsage struct {
	TotalTokens int `json:"total_tokens"`
	ToolUses    int `json:"tool_uses"`
	DurationMs  int `json:"duration_ms"`
}

// SDKTaskNotificationMessage represents a task completion notification.
type SDKTaskNotificationMessage struct {
	Type       string     `json:"type"`    // "system"
	Subtype    string     `json:"subtype"` // "task_notification"
	TaskID     string     `json:"task_id"`
	ToolUseID  *string    `json:"tool_use_id,omitempty"`
	Status     string     `json:"status"` // "completed" | "failed" | "stopped"
	OutputFile string     `json:"output_file"`
	Summary    string     `json:"summary"`
	Usage      *TaskUsage `json:"usage,omitempty"`
	UUID       string     `json:"uuid"`
	SessionID  string     `json:"session_id"`
}

func (m *SDKTaskNotificationMessage) sdkMessage()        {}
func (m *SDKTaskNotificationMessage) MessageType() string { return "system" }

// SDKTaskStartedMessage represents a task start event.
type SDKTaskStartedMessage struct {
	Type        string  `json:"type"`    // "system"
	Subtype     string  `json:"subtype"` // "task_started"
	TaskID      string  `json:"task_id"`
	ToolUseID   *string `json:"tool_use_id,omitempty"`
	Description string  `json:"description"`
	TaskType    *string `json:"task_type,omitempty"`
	Prompt      *string `json:"prompt,omitempty"`
	UUID        string  `json:"uuid"`
	SessionID   string  `json:"session_id"`
}

func (m *SDKTaskStartedMessage) sdkMessage()        {}
func (m *SDKTaskStartedMessage) MessageType() string { return "system" }

// SDKTaskProgressMessage represents a task progress update.
type SDKTaskProgressMessage struct {
	Type         string    `json:"type"`    // "system"
	Subtype      string    `json:"subtype"` // "task_progress"
	TaskID       string    `json:"task_id"`
	ToolUseID    *string   `json:"tool_use_id,omitempty"`
	Description  string    `json:"description"`
	Usage        TaskUsage `json:"usage"`
	LastToolName *string   `json:"last_tool_name,omitempty"`
	Summary      *string   `json:"summary,omitempty"`
	UUID         string    `json:"uuid"`
	SessionID    string    `json:"session_id"`
}

func (m *SDKTaskProgressMessage) sdkMessage()        {}
func (m *SDKTaskProgressMessage) MessageType() string { return "system" }

// --- Files Persisted ---

// PersistedFile represents a successfully persisted file.
type PersistedFile struct {
	Filename string `json:"filename"`
	FileID   string `json:"file_id"`
}

// PersistedFileFail represents a file that failed to persist.
type PersistedFileFail struct {
	Filename string `json:"filename"`
	Error    string `json:"error"`
}

// SDKFilesPersistedEvent represents a files persisted notification.
type SDKFilesPersistedEvent struct {
	Type        string              `json:"type"`    // "system"
	Subtype     string              `json:"subtype"` // "files_persisted"
	Files       []PersistedFile     `json:"files"`
	Failed      []PersistedFileFail `json:"failed"`
	ProcessedAt string              `json:"processed_at"`
	UUID        string              `json:"uuid"`
	SessionID   string              `json:"session_id"`
}

func (m *SDKFilesPersistedEvent) sdkMessage()        {}
func (m *SDKFilesPersistedEvent) MessageType() string { return "system" }

// --- Rate Limit ---

// SDKRateLimitInfo contains rate limit status details.
type SDKRateLimitInfo struct {
	Status                string   `json:"status"` // "allowed" | "allowed_warning" | "rejected"
	ResetsAt              *int64   `json:"resetsAt,omitempty"`
	RateLimitType         *string  `json:"rateLimitType,omitempty"`
	Utilization           *float64 `json:"utilization,omitempty"`
	OverageStatus         *string  `json:"overageStatus,omitempty"`
	OverageResetsAt       *int64   `json:"overageResetsAt,omitempty"`
	OverageDisabledReason *string  `json:"overageDisabledReason,omitempty"`
	IsUsingOverage        *bool    `json:"isUsingOverage,omitempty"`
	SurpassedThreshold    *float64 `json:"surpassedThreshold,omitempty"`
}

// SDKRateLimitEvent represents a rate limit status update.
type SDKRateLimitEvent struct {
	Type          string           `json:"type"` // "rate_limit_event"
	RateLimitInfo SDKRateLimitInfo `json:"rate_limit_info"`
	UUID          string           `json:"uuid"`
	SessionID     string           `json:"session_id"`
}

func (m *SDKRateLimitEvent) sdkMessage()        {}
func (m *SDKRateLimitEvent) MessageType() string { return "rate_limit_event" }

// --- Elicitation Complete ---

// SDKElicitationCompleteMessage represents an elicitation completion event.
type SDKElicitationCompleteMessage struct {
	Type          string `json:"type"`    // "system"
	Subtype       string `json:"subtype"` // "elicitation_complete"
	McpServerName string `json:"mcp_server_name"`
	ElicitationID string `json:"elicitation_id"`
	UUID          string `json:"uuid"`
	SessionID     string `json:"session_id"`
}

func (m *SDKElicitationCompleteMessage) sdkMessage()        {}
func (m *SDKElicitationCompleteMessage) MessageType() string { return "system" }

// --- Prompt Suggestion ---

// SDKPromptSuggestionMessage represents a prompt suggestion from the CLI.
type SDKPromptSuggestionMessage struct {
	Type       string `json:"type"` // "prompt_suggestion"
	Suggestion string `json:"suggestion"`
	UUID       string `json:"uuid"`
	SessionID  string `json:"session_id"`
}

func (m *SDKPromptSuggestionMessage) sdkMessage()        {}
func (m *SDKPromptSuggestionMessage) MessageType() string { return "prompt_suggestion" }

// --- Raw Message (fallthrough for unknown types) ---

// SDKRawMessage holds an unrecognized message type as raw JSON.
// This prevents errors when the CLI emits new or internal message types.
type SDKRawMessage struct {
	RawType    string          `json:"type"`
	RawSubtype string          `json:"subtype,omitempty"`
	Raw        json.RawMessage `json:"-"`
}

func (m *SDKRawMessage) sdkMessage()        {}
func (m *SDKRawMessage) MessageType() string { return m.RawType }

// IsResultMessage returns true if the given SDKMessage is a result message
// (either *SDKResultSuccess or *SDKResultError).
func IsResultMessage(msg SDKMessage) bool {
	switch msg.(type) {
	case *SDKResultSuccess, *SDKResultError:
		return true
	}
	return false
}
