package claudeagent

import (
	"context"
	"io"
)

// ThinkingConfig controls Claude's thinking/reasoning behavior.
type ThinkingConfig struct {
	Type         string `json:"type"`                   // "adaptive" | "enabled" | "disabled"
	BudgetTokens *int   `json:"budgetTokens,omitempty"` // only for "enabled"
}

// ThinkingAdaptive creates a ThinkingConfig for adaptive thinking (Opus 4.6+).
func ThinkingAdaptive() ThinkingConfig { return ThinkingConfig{Type: "adaptive"} }

// ThinkingEnabled creates a ThinkingConfig with a fixed token budget.
func ThinkingEnabled(budgetTokens int) ThinkingConfig {
	return ThinkingConfig{Type: "enabled", BudgetTokens: &budgetTokens}
}

// ThinkingDisabledConfig creates a ThinkingConfig that disables thinking.
func ThinkingDisabledConfig() ThinkingConfig { return ThinkingConfig{Type: "disabled"} }

// ToolConfig provides per-tool configuration for built-in tools.
type ToolConfig struct {
	AskUserQuestion *AskUserQuestionConfig `json:"askUserQuestion,omitempty"`
}

// AskUserQuestionConfig configures the AskUserQuestion tool.
type AskUserQuestionConfig struct {
	PreviewFormat *string `json:"previewFormat,omitempty"` // "markdown" | "html"
}

// OutputFormat configures structured output.
type OutputFormat struct {
	Type   OutputFormatType       `json:"type"`
	Schema map[string]interface{} `json:"schema"`
}

// SystemPromptPreset represents the default Claude Code system prompt with optional append.
type SystemPromptPreset struct {
	Type   string  `json:"type"`   // "preset"
	Preset string  `json:"preset"` // "claude_code"
	Append *string `json:"append,omitempty"`
}

// ToolPreset represents the preset tool set.
type ToolPreset struct {
	Type   string `json:"type"`   // "preset"
	Preset string `json:"preset"` // "claude_code"
}

// SdkPluginConfig configures a local plugin.
type SdkPluginConfig struct {
	Type string `json:"type"` // "local"
	Path string `json:"path"`
}

// ElicitationResult is the response to an elicitation request.
type ElicitationResult struct {
	Action  string                 `json:"action"` // "accept" | "decline" | "cancel"
	Content map[string]interface{} `json:"content,omitempty"`
}

// ElicitationRequest is a request from an MCP server for user input.
type ElicitationRequest struct {
	ServerName      string                 `json:"serverName"`
	Message         string                 `json:"message"`
	Mode            *string                `json:"mode,omitempty"` // "form" | "url"
	URL             *string                `json:"url,omitempty"`
	ElicitationID   *string                `json:"elicitationId,omitempty"`
	RequestedSchema map[string]interface{} `json:"requestedSchema,omitempty"`
}

// OnElicitation is a callback for handling MCP elicitation requests.
type OnElicitation func(ctx context.Context, request ElicitationRequest) (*ElicitationResult, error)

// SpawnOptions configures how the Claude Code process is spawned.
type SpawnOptions struct {
	Command string
	Args    []string
	Cwd     string
	Env     map[string]string
	Signal  <-chan struct{}
}

// SpawnedProcess wraps a running Claude Code subprocess.
type SpawnedProcess interface {
	Stdin() io.WriteCloser
	Stdout() io.ReadCloser
	Wait() error
	Kill() error
}

// SpawnFunc allows custom process spawning (VMs, containers, remote).
type SpawnFunc func(opts SpawnOptions) SpawnedProcess

// Options configures a query to the Claude Code SDK.
type Options struct {
	// AbortContext is a context for cancelling the query.
	AbortContext context.Context

	// AdditionalDirectories are extra directories Claude can access.
	AdditionalDirectories []string

	// Agent is the agent name for the main thread.
	Agent *string

	// Agents defines custom subagents keyed by name.
	Agents map[string]AgentDefinition

	// AllowedTools are tools auto-allowed without permission prompts.
	AllowedTools []string

	// CanUseTool is a custom permission handler.
	CanUseTool CanUseTool

	// Continue resumes the most recent conversation.
	Continue *bool

	// Cwd is the working directory for the session.
	Cwd *string

	// DisallowedTools are tools explicitly disallowed.
	DisallowedTools []string

	// Tools specifies the base set of available built-in tools.
	// Use []string for specific tool names, or ToolPreset for a preset.
	Tools interface{}

	// Env contains environment variables for the Claude Code process.
	Env map[string]string

	// Executable is the JS runtime to use ("bun", "deno", "node").
	Executable *string

	// ExecutableArgs are additional arguments for the JS runtime.
	ExecutableArgs []string

	// ExtraArgs are additional CLI arguments (keys without --, nil values for flags).
	ExtraArgs map[string]*string

	// FallbackModel is a fallback model if the primary fails.
	FallbackModel *string

	// EnableFileCheckpointing enables file change tracking.
	EnableFileCheckpointing *bool

	// ToolConfig provides per-tool configuration.
	ToolConfig *ToolConfig

	// ForkSession forks to a new session ID when resuming.
	ForkSession *bool

	// Betas enables beta features.
	Betas []SdkBeta

	// Hooks are callbacks for lifecycle events.
	Hooks map[HookEvent][]HookCallbackMatcher

	// OnElicitation handles MCP elicitation requests.
	OnElicitation OnElicitation

	// PersistSession controls whether sessions are saved to disk.
	PersistSession *bool

	// IncludePartialMessages enables streaming delta events.
	IncludePartialMessages *bool

	// Thinking controls Claude's reasoning behavior.
	Thinking *ThinkingConfig

	// Effort controls how much effort Claude puts in ("low", "medium", "high", "max").
	Effort *string

	// MaxThinkingTokens limits thinking tokens (deprecated, use Thinking).
	MaxThinkingTokens *int

	// MaxTurns limits conversation turns.
	MaxTurns *int

	// MaxBudgetUsd limits spending in USD.
	MaxBudgetUsd *float64

	// McpServers configures MCP servers keyed by name.
	McpServers map[string]interface{}

	// Model is the Claude model to use.
	Model *string

	// OutputFormat configures structured output.
	OutputFormat *OutputFormat

	// PathToClaudeCodeExecutable overrides the CLI path.
	PathToClaudeCodeExecutable *string

	// PermissionMode sets the permission mode.
	PermissionMode *PermissionMode

	// AllowDangerouslySkipPermissions must be true for bypassPermissions mode.
	AllowDangerouslySkipPermissions *bool

	// PermissionPromptToolName routes permission prompts through an MCP tool.
	PermissionPromptToolName *string

	// Plugins configures local plugins.
	Plugins []SdkPluginConfig

	// PromptSuggestions enables prompt suggestions after each turn.
	PromptSuggestions *bool

	// AgentProgressSummaries enables periodic AI-generated progress summaries.
	AgentProgressSummaries *bool

	// Resume loads a session by ID.
	Resume *string

	// SessionID uses a specific session ID instead of auto-generated.
	SessionID *string

	// ResumeSessionAt resumes up to a specific message UUID.
	ResumeSessionAt *string

	// Sandbox configures command execution isolation.
	Sandbox *SandboxSettings

	// Settings provides additional settings (string path or *Settings).
	Settings interface{}

	// SettingSources controls which filesystem settings to load.
	SettingSources []SettingSource

	// Debug enables verbose debug logging.
	Debug *bool

	// DebugFile writes debug logs to a file.
	DebugFile *string

	// Stderr is a callback for stderr output.
	Stderr func(string)

	// StrictMcpConfig enforces strict MCP config validation.
	StrictMcpConfig *bool

	// SystemPrompt is a custom system prompt (string or SystemPromptPreset).
	SystemPrompt interface{}

	// SpawnClaudeCodeProcess customizes how the CLI process is spawned.
	SpawnClaudeCodeProcess SpawnFunc
}
