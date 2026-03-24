package tools

// AllowedPrompt describes a category of actions permitted in a plan.
type AllowedPrompt struct {
	// Tool is the tool this prompt applies to.
	Tool string `json:"tool"`
	// Prompt is a semantic description of the action.
	Prompt string `json:"prompt"`
}

// ExitPlanModeInput is the input for the ExitPlanMode tool.
type ExitPlanModeInput struct {
	// AllowedPrompts lists the permissions needed to implement the plan.
	AllowedPrompts []AllowedPrompt `json:"allowedPrompts,omitempty"`
}

// ExitPlanModeOutput is the output from the ExitPlanMode tool.
type ExitPlanModeOutput struct {
	// Plan is the plan that was presented to the user.
	Plan *string `json:"plan"`
	// IsAgent indicates whether this is an agent context.
	IsAgent bool `json:"isAgent"`
	// FilePath is where the plan was saved.
	FilePath *string `json:"filePath,omitempty"`
	// HasTaskTool indicates whether the Agent tool is available.
	HasTaskTool *bool `json:"hasTaskTool,omitempty"`
	// AwaitingLeaderApproval indicates the teammate sent a plan approval request.
	AwaitingLeaderApproval *bool `json:"awaitingLeaderApproval,omitempty"`
	// RequestID is the unique identifier for the plan approval request.
	RequestID *string `json:"requestId,omitempty"`
}
