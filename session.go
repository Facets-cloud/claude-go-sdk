package claudeagent

import "encoding/json"

// SDKSessionInfo contains session metadata returned by ListSessions and GetSessionInfo.
type SDKSessionInfo struct {
	SessionID    string  `json:"sessionId"`
	Summary      string  `json:"summary"`
	LastModified int64   `json:"lastModified"`
	FileSize     *int64  `json:"fileSize,omitempty"`
	CustomTitle  *string `json:"customTitle,omitempty"`
	FirstPrompt  *string `json:"firstPrompt,omitempty"`
	GitBranch    *string `json:"gitBranch,omitempty"`
	Cwd          *string `json:"cwd,omitempty"`
	Tag          *string `json:"tag,omitempty"`
	CreatedAt    *int64  `json:"createdAt,omitempty"`
}

// SessionMessage is a user or assistant message from a session transcript.
type SessionMessage struct {
	Type            string          `json:"type"` // "user" | "assistant"
	UUID            string          `json:"uuid"`
	SessionID       string          `json:"session_id"`
	Message         json.RawMessage `json:"message"`
	ParentToolUseID *string         `json:"parent_tool_use_id"`
}

// ListSessionsOptions configures session listing.
type ListSessionsOptions struct {
	Dir              *string `json:"dir,omitempty"`
	Limit            *int    `json:"limit,omitempty"`
	Offset           *int    `json:"offset,omitempty"`
	IncludeWorktrees *bool   `json:"includeWorktrees,omitempty"`
}

// GetSessionInfoOptions configures single session info retrieval.
type GetSessionInfoOptions struct {
	Dir *string `json:"dir,omitempty"`
}

// GetSessionMessagesOptions configures session message retrieval.
type GetSessionMessagesOptions struct {
	Dir    *string `json:"dir,omitempty"`
	Limit  *int    `json:"limit,omitempty"`
	Offset *int    `json:"offset,omitempty"`
}

// SessionMutationOptions is shared by renameSession, tagSession, deleteSession, forkSession.
type SessionMutationOptions struct {
	Dir *string `json:"dir,omitempty"`
}

// ForkSessionOptions configures forking a session into a new branch.
type ForkSessionOptions struct {
	SessionMutationOptions
	UpToMessageID *string `json:"upToMessageId,omitempty"`
	Title         *string `json:"title,omitempty"`
}

// ForkSessionResult contains the result of a fork operation.
type ForkSessionResult struct {
	SessionID string `json:"sessionId"`
}

// SDKSessionOptions configures the V2 session API (unstable/alpha).
type SDKSessionOptions struct {
	Model            string            `json:"model"`
	PathToClaudeCode *string           `json:"pathToClaudeCodeExecutable,omitempty"`
	Executable       *string           `json:"executable,omitempty"` // "node" | "bun"
	ExecutableArgs   []string          `json:"executableArgs,omitempty"`
	Env              map[string]string `json:"env,omitempty"`
	AllowedTools     []string          `json:"allowedTools,omitempty"`
	DisallowedTools  []string          `json:"disallowedTools,omitempty"`
	PermissionMode   *PermissionMode   `json:"permissionMode,omitempty"`
}
