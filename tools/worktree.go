package tools

// EnterWorktreeInput is the input for the EnterWorktree tool.
type EnterWorktreeInput struct {
	// Name is an optional name for the worktree.
	Name *string `json:"name,omitempty"`
}

// EnterWorktreeOutput is the output from the EnterWorktree tool.
type EnterWorktreeOutput struct {
	// WorktreePath is the path to the created worktree.
	WorktreePath string `json:"worktreePath"`
	// WorktreeBranch is the branch name of the worktree.
	WorktreeBranch *string `json:"worktreeBranch,omitempty"`
	// Message is a status message.
	Message string `json:"message"`
}

// ExitWorktreeInput is the input for the ExitWorktree tool.
type ExitWorktreeInput struct {
	// Action is "keep" or "remove".
	Action string `json:"action"`
	// DiscardChanges forces removal even with uncommitted changes.
	DiscardChanges *bool `json:"discard_changes,omitempty"`
}

// ExitWorktreeOutput is the output from the ExitWorktree tool.
type ExitWorktreeOutput struct {
	// Action is "keep" or "remove".
	Action string `json:"action"`
	// OriginalCwd is the original working directory.
	OriginalCwd string `json:"originalCwd"`
	// WorktreePath is the path to the worktree.
	WorktreePath string `json:"worktreePath"`
	// WorktreeBranch is the branch name of the worktree.
	WorktreeBranch *string `json:"worktreeBranch,omitempty"`
	// TmuxSessionName is the tmux session name if applicable.
	TmuxSessionName *string `json:"tmuxSessionName,omitempty"`
	// DiscardedFiles is the number of files discarded.
	DiscardedFiles *int `json:"discardedFiles,omitempty"`
	// DiscardedCommits is the number of commits discarded.
	DiscardedCommits *int `json:"discardedCommits,omitempty"`
	// Message is a status message.
	Message string `json:"message"`
}
