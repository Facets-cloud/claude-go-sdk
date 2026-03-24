package claudeagent

import "context"

// BaseHookInput contains fields common to all hook event inputs.
type BaseHookInput struct {
	SessionID      string  `json:"session_id"`
	TranscriptPath string  `json:"transcript_path"`
	Cwd            string  `json:"cwd"`
	PermissionMode *string `json:"permission_mode,omitempty"`
	AgentID        *string `json:"agent_id,omitempty"`
	AgentType      *string `json:"agent_type,omitempty"`
}

// --- Hook Event Input Types (all 23) ---

// PreToolUseHookInput is fired before a tool is executed.
type PreToolUseHookInput struct {
	BaseHookInput
	HookEventName string      `json:"hook_event_name"` // "PreToolUse"
	ToolName      string      `json:"tool_name"`
	ToolInput     interface{} `json:"tool_input"`
	ToolUseID     string      `json:"tool_use_id"`
}

// PostToolUseHookInput is fired after a tool completes successfully.
type PostToolUseHookInput struct {
	BaseHookInput
	HookEventName string      `json:"hook_event_name"` // "PostToolUse"
	ToolName      string      `json:"tool_name"`
	ToolInput     interface{} `json:"tool_input"`
	ToolResponse  interface{} `json:"tool_response"`
	ToolUseID     string      `json:"tool_use_id"`
}

// PostToolUseFailureHookInput is fired after a tool execution fails.
type PostToolUseFailureHookInput struct {
	BaseHookInput
	HookEventName string      `json:"hook_event_name"` // "PostToolUseFailure"
	ToolName      string      `json:"tool_name"`
	ToolInput     interface{} `json:"tool_input"`
	ToolUseID     string      `json:"tool_use_id"`
	Error         string      `json:"error"`
	IsInterrupt   *bool       `json:"is_interrupt,omitempty"`
}

// NotificationHookInput is fired when a notification occurs.
type NotificationHookInput struct {
	BaseHookInput
	HookEventName    string `json:"hook_event_name"` // "Notification"
	Message          string `json:"message"`
	Title            *string `json:"title,omitempty"`
	NotificationType string `json:"notification_type"`
}

// UserPromptSubmitHookInput is fired when a user submits a prompt.
type UserPromptSubmitHookInput struct {
	BaseHookInput
	HookEventName string `json:"hook_event_name"` // "UserPromptSubmit"
	Prompt        string `json:"prompt"`
}

// SessionStartHookInput is fired when a session starts.
type SessionStartHookInput struct {
	BaseHookInput
	HookEventName string  `json:"hook_event_name"` // "SessionStart"
	Source        string  `json:"source"`           // "startup" | "resume" | "clear" | "compact"
	AgentType     *string `json:"agent_type,omitempty"`
	Model         *string `json:"model,omitempty"`
}

// SessionEndHookInput is fired when a session ends.
type SessionEndHookInput struct {
	BaseHookInput
	HookEventName string     `json:"hook_event_name"` // "SessionEnd"
	Reason        ExitReason `json:"reason"`
}

// StopHookInput is fired when the assistant stops.
type StopHookInput struct {
	BaseHookInput
	HookEventName        string  `json:"hook_event_name"` // "Stop"
	StopHookActive       bool    `json:"stop_hook_active"`
	LastAssistantMessage *string `json:"last_assistant_message,omitempty"`
}

// StopFailureHookInput is fired when the assistant stops due to an error.
type StopFailureHookInput struct {
	BaseHookInput
	HookEventName        string                   `json:"hook_event_name"` // "StopFailure"
	Error                SDKAssistantMessageError `json:"error"`
	ErrorDetails         *string                  `json:"error_details,omitempty"`
	LastAssistantMessage *string                  `json:"last_assistant_message,omitempty"`
}

// SubagentStartHookInput is fired when a subagent starts.
type SubagentStartHookInput struct {
	BaseHookInput
	HookEventName string `json:"hook_event_name"` // "SubagentStart"
	AgentID       string `json:"agent_id"`
	AgentType     string `json:"agent_type"`
}

// SubagentStopHookInput is fired when a subagent stops.
type SubagentStopHookInput struct {
	BaseHookInput
	HookEventName        string  `json:"hook_event_name"` // "SubagentStop"
	StopHookActive       bool    `json:"stop_hook_active"`
	AgentID              string  `json:"agent_id"`
	AgentTranscriptPath  string  `json:"agent_transcript_path"`
	AgentType            string  `json:"agent_type"`
	LastAssistantMessage *string `json:"last_assistant_message,omitempty"`
}

// PreCompactHookInput is fired before context compaction.
type PreCompactHookInput struct {
	BaseHookInput
	HookEventName      string  `json:"hook_event_name"` // "PreCompact"
	Trigger            string  `json:"trigger"`          // "manual" | "auto"
	CustomInstructions *string `json:"custom_instructions"`
}

// PostCompactHookInput is fired after context compaction.
type PostCompactHookInput struct {
	BaseHookInput
	HookEventName  string `json:"hook_event_name"` // "PostCompact"
	Trigger        string `json:"trigger"`          // "manual" | "auto"
	CompactSummary string `json:"compact_summary"`
}

// PermissionRequestHookInput is fired when a tool requests permission.
type PermissionRequestHookInput struct {
	BaseHookInput
	HookEventName        string             `json:"hook_event_name"` // "PermissionRequest"
	ToolName             string             `json:"tool_name"`
	ToolInput            interface{}        `json:"tool_input"`
	PermissionSuggestions []PermissionUpdate `json:"permission_suggestions,omitempty"`
}

// SetupHookInput is fired during initialization or maintenance.
type SetupHookInput struct {
	BaseHookInput
	HookEventName string `json:"hook_event_name"` // "Setup"
	Trigger       string `json:"trigger"`          // "init" | "maintenance"
}

// TeammateIdleHookInput is fired when a teammate becomes idle.
type TeammateIdleHookInput struct {
	BaseHookInput
	HookEventName string `json:"hook_event_name"` // "TeammateIdle"
	TeammateName  string `json:"teammate_name"`
	TeamName      string `json:"team_name"`
}

// TaskCompletedHookInput is fired when a task completes.
type TaskCompletedHookInput struct {
	BaseHookInput
	HookEventName   string  `json:"hook_event_name"` // "TaskCompleted"
	TaskID          string  `json:"task_id"`
	TaskSubject     string  `json:"task_subject"`
	TaskDescription *string `json:"task_description,omitempty"`
	TeammateName    *string `json:"teammate_name,omitempty"`
	TeamName        *string `json:"team_name,omitempty"`
}

// ElicitationHookInput is fired when an MCP server requests user input.
type ElicitationHookInput struct {
	BaseHookInput
	HookEventName   string                 `json:"hook_event_name"` // "Elicitation"
	McpServerName   string                 `json:"mcp_server_name"`
	Message         string                 `json:"message"`
	Mode            *string                `json:"mode,omitempty"` // "form" | "url"
	URL             *string                `json:"url,omitempty"`
	ElicitationID   *string                `json:"elicitation_id,omitempty"`
	RequestedSchema map[string]interface{} `json:"requested_schema,omitempty"`
}

// ElicitationResultHookInput is fired after the user responds to an elicitation.
type ElicitationResultHookInput struct {
	BaseHookInput
	HookEventName string                 `json:"hook_event_name"` // "ElicitationResult"
	McpServerName string                 `json:"mcp_server_name"`
	ElicitationID *string                `json:"elicitation_id,omitempty"`
	Mode          *string                `json:"mode,omitempty"` // "form" | "url"
	Action        string                 `json:"action"`         // "accept" | "decline" | "cancel"
	Content       map[string]interface{} `json:"content,omitempty"`
}

// ConfigChangeHookInput is fired when a configuration file changes.
type ConfigChangeHookInput struct {
	BaseHookInput
	HookEventName string  `json:"hook_event_name"` // "ConfigChange"
	Source        string  `json:"source"`           // "user_settings" | "project_settings" | "local_settings" | "policy_settings" | "skills"
	FilePath      *string `json:"file_path,omitempty"`
}

// InstructionsLoadedHookInput is fired when a CLAUDE.md file is loaded.
type InstructionsLoadedHookInput struct {
	BaseHookInput
	HookEventName   string   `json:"hook_event_name"` // "InstructionsLoaded"
	FilePath        string   `json:"file_path"`
	MemoryType      string   `json:"memory_type"`  // "User" | "Project" | "Local" | "Managed"
	LoadReason      string   `json:"load_reason"`  // "session_start" | "nested_traversal" | "path_glob_match" | "include" | "compact"
	Globs           []string `json:"globs,omitempty"`
	TriggerFilePath *string  `json:"trigger_file_path,omitempty"`
	ParentFilePath  *string  `json:"parent_file_path,omitempty"`
}

// WorktreeCreateHookInput is fired when a git worktree is created.
type WorktreeCreateHookInput struct {
	BaseHookInput
	HookEventName string `json:"hook_event_name"` // "WorktreeCreate"
	Name          string `json:"name"`
}

// WorktreeRemoveHookInput is fired when a git worktree is removed.
type WorktreeRemoveHookInput struct {
	BaseHookInput
	HookEventName string `json:"hook_event_name"` // "WorktreeRemove"
	WorktreePath  string `json:"worktree_path"`
}

// HookInput is the union of all hook event input types.
type HookInput interface{}

// --- Hook-Specific Output Types ---

// PreToolUseHookSpecificOutput is the hook-specific output for PreToolUse events.
type PreToolUseHookSpecificOutput struct {
	HookEventName          string                 `json:"hookEventName"` // "PreToolUse"
	PermissionDecision     *string                `json:"permissionDecision,omitempty"` // "allow" | "deny" | "ask"
	PermissionDecisionReason *string              `json:"permissionDecisionReason,omitempty"`
	UpdatedInput           map[string]interface{} `json:"updatedInput,omitempty"`
	AdditionalContext      *string                `json:"additionalContext,omitempty"`
}

// PostToolUseHookSpecificOutput is the hook-specific output for PostToolUse events.
type PostToolUseHookSpecificOutput struct {
	HookEventName      string      `json:"hookEventName"` // "PostToolUse"
	AdditionalContext  *string     `json:"additionalContext,omitempty"`
	UpdatedMCPToolOutput interface{} `json:"updatedMCPToolOutput,omitempty"`
}

// PostToolUseFailureHookSpecificOutput is the hook-specific output for PostToolUseFailure events.
type PostToolUseFailureHookSpecificOutput struct {
	HookEventName     string  `json:"hookEventName"` // "PostToolUseFailure"
	AdditionalContext *string `json:"additionalContext,omitempty"`
}

// NotificationHookSpecificOutput is the hook-specific output for Notification events.
type NotificationHookSpecificOutput struct {
	HookEventName     string  `json:"hookEventName"` // "Notification"
	AdditionalContext *string `json:"additionalContext,omitempty"`
}

// UserPromptSubmitHookSpecificOutput is the hook-specific output for UserPromptSubmit events.
type UserPromptSubmitHookSpecificOutput struct {
	HookEventName     string  `json:"hookEventName"` // "UserPromptSubmit"
	AdditionalContext *string `json:"additionalContext,omitempty"`
}

// SessionStartHookSpecificOutput is the hook-specific output for SessionStart events.
type SessionStartHookSpecificOutput struct {
	HookEventName       string  `json:"hookEventName"` // "SessionStart"
	AdditionalContext   *string `json:"additionalContext,omitempty"`
	InitialUserMessage  *string `json:"initialUserMessage,omitempty"`
}

// SetupHookSpecificOutput is the hook-specific output for Setup events.
type SetupHookSpecificOutput struct {
	HookEventName     string  `json:"hookEventName"` // "Setup"
	AdditionalContext *string `json:"additionalContext,omitempty"`
}

// SubagentStartHookSpecificOutput is the hook-specific output for SubagentStart events.
type SubagentStartHookSpecificOutput struct {
	HookEventName     string  `json:"hookEventName"` // "SubagentStart"
	AdditionalContext *string `json:"additionalContext,omitempty"`
}

// PermissionRequestHookSpecificOutput is the hook-specific output for PermissionRequest events.
type PermissionRequestHookSpecificOutput struct {
	HookEventName string                            `json:"hookEventName"` // "PermissionRequest"
	Decision      PermissionRequestHookDecision     `json:"decision"`
}

// PermissionRequestHookDecision is the decision for a permission request hook.
// Use PermissionRequestAllow or PermissionRequestDeny.
type PermissionRequestHookDecision interface {
	permissionRequestHookDecision()
}

// PermissionRequestAllow approves a permission request from a hook.
type PermissionRequestAllow struct {
	Behavior           string                 `json:"behavior"` // "allow"
	UpdatedInput       map[string]interface{} `json:"updatedInput,omitempty"`
	UpdatedPermissions []PermissionUpdate     `json:"updatedPermissions,omitempty"`
}

func (d PermissionRequestAllow) permissionRequestHookDecision() {}

// PermissionRequestDeny denies a permission request from a hook.
type PermissionRequestDeny struct {
	Behavior  string  `json:"behavior"` // "deny"
	Message   *string `json:"message,omitempty"`
	Interrupt *bool   `json:"interrupt,omitempty"`
}

func (d PermissionRequestDeny) permissionRequestHookDecision() {}

// ElicitationHookSpecificOutput is the hook-specific output for Elicitation events.
type ElicitationHookSpecificOutput struct {
	HookEventName string                 `json:"hookEventName"` // "Elicitation"
	Action        *string                `json:"action,omitempty"` // "accept" | "decline" | "cancel"
	Content       map[string]interface{} `json:"content,omitempty"`
}

// ElicitationResultHookSpecificOutput is the hook-specific output for ElicitationResult events.
type ElicitationResultHookSpecificOutput struct {
	HookEventName string                 `json:"hookEventName"` // "ElicitationResult"
	Action        *string                `json:"action,omitempty"` // "accept" | "decline" | "cancel"
	Content       map[string]interface{} `json:"content,omitempty"`
}

// --- Hook Callback Types ---

// HookCallback is a function called when a hook event fires.
type HookCallback func(ctx context.Context, input HookInput, toolUseID *string) (HookJSONOutput, error)

// HookCallbackMatcher contains hook callbacks with optional pattern matching.
type HookCallbackMatcher struct {
	Matcher *string        `json:"matcher,omitempty"`
	Hooks   []HookCallback `json:"-"` // not serializable
	Timeout *int           `json:"timeout,omitempty"` // seconds
}

// HookJSONOutput is the union of sync and async hook outputs.
type HookJSONOutput interface {
	hookJSONOutput()
}

// SyncHookJSONOutput is the synchronous hook output.
type SyncHookJSONOutput struct {
	Continue       *bool       `json:"continue,omitempty"`
	SuppressOutput *bool       `json:"suppressOutput,omitempty"`
	StopReason     *string     `json:"stopReason,omitempty"`
	Decision       *string     `json:"decision,omitempty"` // "approve" | "block"
	SystemMessage  *string     `json:"systemMessage,omitempty"`
	Reason         *string     `json:"reason,omitempty"`
	HookSpecificOutput interface{} `json:"hookSpecificOutput,omitempty"`
}

func (o SyncHookJSONOutput) hookJSONOutput() {}

// AsyncHookJSONOutput indicates that the hook runs asynchronously.
type AsyncHookJSONOutput struct {
	Async        bool `json:"async"` // always true
	AsyncTimeout *int `json:"asyncTimeout,omitempty"`
}

func (o AsyncHookJSONOutput) hookJSONOutput() {}
