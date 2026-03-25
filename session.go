package claudeagent

import (
	"encoding/json"
	"fmt"
	"os/exec"
)

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
	Model            string                                   `json:"model"`
	PathToClaudeCode *string                                  `json:"pathToClaudeCodeExecutable,omitempty"`
	Executable       *string                                  `json:"executable,omitempty"` // "node" | "bun"
	ExecutableArgs   []string                                 `json:"executableArgs,omitempty"`
	Env              map[string]string                        `json:"env,omitempty"`
	AllowedTools     []string                                 `json:"allowedTools,omitempty"`
	DisallowedTools  []string                                 `json:"disallowedTools,omitempty"`
	CanUseTool       CanUseTool                               `json:"-"` // not serializable
	Hooks            map[HookEvent][]HookCallbackMatcher      `json:"-"` // not serializable
	PermissionMode   *PermissionMode                          `json:"permissionMode,omitempty"`
}

// --- Session Management Functions ---
// These invoke the Claude Code CLI with --print-session* flags and parse the output.

// runCLICommand runs the claude CLI with given args and returns stdout.
func runCLICommand(args ...string) ([]byte, error) {
	cliPath, err := CLIPath(nil)
	if err != nil {
		return nil, err
	}
	cmd := exec.Command(cliPath, args...)
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("claude cli: %w", err)
	}
	return out, nil
}

// ListSessions returns session metadata.
func ListSessions(opts *ListSessionsOptions) ([]SDKSessionInfo, error) {
	args := []string{"--output-format", "json", "--print-sessions"}
	if opts != nil {
		if opts.Dir != nil {
			args = append(args, "--cwd", *opts.Dir)
		}
		if opts.Limit != nil {
			args = append(args, "--limit", fmt.Sprintf("%d", *opts.Limit))
		}
		if opts.Offset != nil {
			args = append(args, "--offset", fmt.Sprintf("%d", *opts.Offset))
		}
	}
	out, err := runCLICommand(args...)
	if err != nil {
		return nil, err
	}
	var sessions []SDKSessionInfo
	if err := json.Unmarshal(out, &sessions); err != nil {
		return nil, fmt.Errorf("parse sessions: %w", err)
	}
	return sessions, nil
}

// GetSessionInfo returns metadata for a single session.
func GetSessionInfo(sessionID string, opts *GetSessionInfoOptions) (*SDKSessionInfo, error) {
	args := []string{"--output-format", "json", "--print-session-info", sessionID}
	if opts != nil && opts.Dir != nil {
		args = append(args, "--cwd", *opts.Dir)
	}
	out, err := runCLICommand(args...)
	if err != nil {
		return nil, err
	}
	var info SDKSessionInfo
	if err := json.Unmarshal(out, &info); err != nil {
		return nil, fmt.Errorf("parse session info: %w", err)
	}
	return &info, nil
}

// GetSessionMessages reads conversation messages from a session transcript.
func GetSessionMessages(sessionID string, opts *GetSessionMessagesOptions) ([]SessionMessage, error) {
	args := []string{"--output-format", "json", "--print-session-messages", sessionID}
	if opts != nil {
		if opts.Dir != nil {
			args = append(args, "--cwd", *opts.Dir)
		}
		if opts.Limit != nil {
			args = append(args, "--limit", fmt.Sprintf("%d", *opts.Limit))
		}
		if opts.Offset != nil {
			args = append(args, "--offset", fmt.Sprintf("%d", *opts.Offset))
		}
	}
	out, err := runCLICommand(args...)
	if err != nil {
		return nil, err
	}
	var messages []SessionMessage
	if err := json.Unmarshal(out, &messages); err != nil {
		return nil, fmt.Errorf("parse session messages: %w", err)
	}
	return messages, nil
}

// ForkSession creates a new session branched from an existing one.
func ForkSession(sessionID string, opts *ForkSessionOptions) (*ForkSessionResult, error) {
	args := []string{"--output-format", "json", "--fork-session", sessionID}
	if opts != nil {
		if opts.UpToMessageID != nil {
			args = append(args, "--up-to-message-id", *opts.UpToMessageID)
		}
		if opts.Title != nil {
			args = append(args, "--title", *opts.Title)
		}
		if opts.Dir != nil {
			args = append(args, "--cwd", *opts.Dir)
		}
	}
	out, err := runCLICommand(args...)
	if err != nil {
		return nil, err
	}
	var result ForkSessionResult
	if err := json.Unmarshal(out, &result); err != nil {
		return nil, fmt.Errorf("parse fork result: %w", err)
	}
	return &result, nil
}

// RenameSession changes a session's title.
func RenameSession(sessionID string, title string, opts *SessionMutationOptions) error {
	args := []string{"--output-format", "json", "--rename-session", sessionID, "--title", title}
	if opts != nil && opts.Dir != nil {
		args = append(args, "--cwd", *opts.Dir)
	}
	_, err := runCLICommand(args...)
	return err
}

// TagSession adds or clears a tag on a session. Pass nil to clear the tag.
func TagSession(sessionID string, tag *string, opts *SessionMutationOptions) error {
	args := []string{"--output-format", "json", "--tag-session", sessionID}
	if tag != nil {
		args = append(args, "--tag", *tag)
	} else {
		args = append(args, "--tag", "")
	}
	if opts != nil && opts.Dir != nil {
		args = append(args, "--cwd", *opts.Dir)
	}
	_, err := runCLICommand(args...)
	return err
}
