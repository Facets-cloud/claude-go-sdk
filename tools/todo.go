package tools

// TodoItem represents a single todo item.
type TodoItem struct {
	// Content is the todo item text.
	Content string `json:"content"`
	// Status is "pending", "in_progress", or "completed".
	Status string `json:"status"`
	// ActiveForm is the present continuous form shown in spinner.
	ActiveForm string `json:"activeForm"`
}

// TodoWriteInput is the input for the TodoWrite tool.
type TodoWriteInput struct {
	// Todos is the updated todo list.
	Todos []TodoItem `json:"todos"`
}

// TodoWriteOutput is the output from the TodoWrite tool.
type TodoWriteOutput struct {
	// OldTodos is the todo list before the update.
	OldTodos []TodoItem `json:"oldTodos"`
	// NewTodos is the todo list after the update.
	NewTodos []TodoItem `json:"newTodos"`
	// VerificationNudgeNeeded indicates if verification is needed.
	VerificationNudgeNeeded *bool `json:"verificationNudgeNeeded,omitempty"`
}
