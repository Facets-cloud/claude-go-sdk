package tools

// TaskOutputInput is the input for the TaskOutput tool (reading background task output).
type TaskOutputInput struct {
	// TaskID is the task ID to get output from.
	TaskID string `json:"task_id"`
	// Block indicates whether to wait for completion.
	Block bool `json:"block"`
	// Timeout is the max wait time in milliseconds.
	Timeout int `json:"timeout"`
}

// TaskStopInput is the input for the TaskStop tool.
type TaskStopInput struct {
	// TaskID is the ID of the background task to stop.
	TaskID *string `json:"task_id,omitempty"`
	// ShellID is deprecated; use TaskID instead.
	ShellID *string `json:"shell_id,omitempty"`
}

// TaskStopOutput is the output from the TaskStop tool.
type TaskStopOutput struct {
	// Message is the status message about the operation.
	Message string `json:"message"`
	// TaskID is the ID of the task that was stopped.
	TaskID string `json:"task_id"`
	// TaskType is the type of the task that was stopped.
	TaskType string `json:"task_type"`
	// Command is the command or description of the stopped task.
	Command *string `json:"command,omitempty"`
}
