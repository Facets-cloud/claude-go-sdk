package claudeagent

import "context"

// CanUseTool is a callback invoked before each tool execution.
// Return a PermissionResult to allow or deny the tool use.
type CanUseTool func(ctx context.Context, toolName string, input map[string]interface{}, opts CanUseToolOptions) (PermissionResult, error)

// CanUseToolOptions provides context for permission decisions.
type CanUseToolOptions struct {
	Suggestions    []PermissionUpdate `json:"suggestions,omitempty"`
	BlockedPath    *string            `json:"blockedPath,omitempty"`
	DecisionReason *string            `json:"decisionReason,omitempty"`
	Title          *string            `json:"title,omitempty"`
	DisplayName    *string            `json:"displayName,omitempty"`
	Description    *string            `json:"description,omitempty"`
	ToolUseID      string             `json:"toolUseID"`
	AgentID        *string            `json:"agentID,omitempty"`
}

// PermissionResult is the interface for permission decisions.
type PermissionResult interface {
	permissionResult()
}

// PermissionResultAllow approves a tool use.
type PermissionResultAllow struct {
	Behavior           PermissionBehavior    `json:"behavior"` // "allow"
	UpdatedInput       map[string]interface{} `json:"updatedInput,omitempty"`
	UpdatedPermissions []PermissionUpdate     `json:"updatedPermissions,omitempty"`
	ToolUseID          *string                `json:"toolUseID,omitempty"`
}

func (r PermissionResultAllow) permissionResult() {}

// PermissionResultDeny denies a tool use.
type PermissionResultDeny struct {
	Behavior  PermissionBehavior `json:"behavior"` // "deny"
	Message   string             `json:"message"`
	Interrupt *bool              `json:"interrupt,omitempty"`
	ToolUseID *string            `json:"toolUseID,omitempty"`
}

func (r PermissionResultDeny) permissionResult() {}

// PermissionRuleValue identifies a permission rule.
type PermissionRuleValue struct {
	ToolName    string  `json:"toolName"`
	RuleContent *string `json:"ruleContent,omitempty"`
}

// PermissionUpdate is the interface for permission update operations.
type PermissionUpdate interface {
	permissionUpdate()
}

// PermissionUpdateAddRules adds new permission rules.
type PermissionUpdateAddRules struct {
	UpdateType  string                      `json:"type"` // "addRules"
	Rules       []PermissionRuleValue       `json:"rules"`
	Behavior    PermissionBehavior          `json:"behavior"`
	Destination PermissionUpdateDestination `json:"destination"`
}

func (u PermissionUpdateAddRules) permissionUpdate() {}

// PermissionUpdateReplaceRules replaces existing permission rules.
type PermissionUpdateReplaceRules struct {
	UpdateType  string                      `json:"type"` // "replaceRules"
	Rules       []PermissionRuleValue       `json:"rules"`
	Behavior    PermissionBehavior          `json:"behavior"`
	Destination PermissionUpdateDestination `json:"destination"`
}

func (u PermissionUpdateReplaceRules) permissionUpdate() {}

// PermissionUpdateRemoveRules removes permission rules.
type PermissionUpdateRemoveRules struct {
	UpdateType  string                      `json:"type"` // "removeRules"
	Rules       []PermissionRuleValue       `json:"rules"`
	Behavior    PermissionBehavior          `json:"behavior"`
	Destination PermissionUpdateDestination `json:"destination"`
}

func (u PermissionUpdateRemoveRules) permissionUpdate() {}

// PermissionUpdateSetMode changes the permission mode.
type PermissionUpdateSetMode struct {
	UpdateType  string                      `json:"type"` // "setMode"
	Mode        PermissionMode              `json:"mode"`
	Destination PermissionUpdateDestination `json:"destination"`
}

func (u PermissionUpdateSetMode) permissionUpdate() {}

// PermissionUpdateAddDirectories adds directories to the permission scope.
type PermissionUpdateAddDirectories struct {
	UpdateType  string                      `json:"type"` // "addDirectories"
	Directories []string                    `json:"directories"`
	Destination PermissionUpdateDestination `json:"destination"`
}

func (u PermissionUpdateAddDirectories) permissionUpdate() {}

// PermissionUpdateRemoveDirectories removes directories from the permission scope.
type PermissionUpdateRemoveDirectories struct {
	UpdateType  string                      `json:"type"` // "removeDirectories"
	Directories []string                    `json:"directories"`
	Destination PermissionUpdateDestination `json:"destination"`
}

func (u PermissionUpdateRemoveDirectories) permissionUpdate() {}
