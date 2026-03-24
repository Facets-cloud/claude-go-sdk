package tools

// BashInput is the input for the Bash tool.
type BashInput struct {
	// Command is the command to execute.
	Command string `json:"command"`
	// Timeout is an optional timeout in milliseconds (max 600000).
	Timeout *int `json:"timeout,omitempty"`
	// Description describes what the command does.
	Description *string `json:"description,omitempty"`
	// RunInBackground runs the command in the background when true.
	RunInBackground *bool `json:"run_in_background,omitempty"`
	// DangerouslyDisableSandbox overrides sandbox mode when true.
	DangerouslyDisableSandbox *bool `json:"dangerouslyDisableSandbox,omitempty"`
}

// BashOutput is the output from the Bash tool.
type BashOutput struct {
	// Stdout is the standard output of the command.
	Stdout string `json:"stdout"`
	// Stderr is the standard error output.
	Stderr string `json:"stderr"`
	// RawOutputPath is the path to raw output for large MCP tool outputs.
	RawOutputPath *string `json:"rawOutputPath,omitempty"`
	// Interrupted indicates whether the command was interrupted.
	Interrupted bool `json:"interrupted"`
	// IsImage indicates if stdout contains image data.
	IsImage *bool `json:"isImage,omitempty"`
	// BackgroundTaskID is the ID of the background task.
	BackgroundTaskID *string `json:"backgroundTaskId,omitempty"`
	// BackgroundedByUser is true if the user manually backgrounded the command.
	BackgroundedByUser *bool `json:"backgroundedByUser,omitempty"`
	// AssistantAutoBackgrounded is true if auto-backgrounded by assistant mode.
	AssistantAutoBackgrounded *bool `json:"assistantAutoBackgrounded,omitempty"`
	// DangerouslyDisableSandbox indicates if sandbox mode was overridden.
	DangerouslyDisableSandbox *bool `json:"dangerouslyDisableSandbox,omitempty"`
	// ReturnCodeInterpretation is semantic interpretation for non-error exit codes.
	ReturnCodeInterpretation *string `json:"returnCodeInterpretation,omitempty"`
	// NoOutputExpected indicates the command is expected to produce no output on success.
	NoOutputExpected *bool `json:"noOutputExpected,omitempty"`
	// StructuredContent contains structured content blocks.
	StructuredContent []interface{} `json:"structuredContent,omitempty"`
	// PersistedOutputPath is the path to persisted full output when too large for inline.
	PersistedOutputPath *string `json:"persistedOutputPath,omitempty"`
	// PersistedOutputSize is the total size of persisted output in bytes.
	PersistedOutputSize *int `json:"persistedOutputSize,omitempty"`
	// TokenSaverOutput is compressed output sent to model when token-saver is active.
	TokenSaverOutput *string `json:"tokenSaverOutput,omitempty"`
}
