package claudeagent

// PermissionMode controls how tool executions are handled.
type PermissionMode string

const (
	PermissionModeDefault           PermissionMode = "default"
	PermissionModeAcceptEdits       PermissionMode = "acceptEdits"
	PermissionModeBypassPermissions PermissionMode = "bypassPermissions"
	PermissionModePlan              PermissionMode = "plan"
	PermissionModeDontAsk           PermissionMode = "dontAsk"
)

// ExitReason describes why a session ended.
type ExitReason string

const (
	ExitReasonClear                     ExitReason = "clear"
	ExitReasonResume                    ExitReason = "resume"
	ExitReasonLogout                    ExitReason = "logout"
	ExitReasonPromptInputExit           ExitReason = "prompt_input_exit"
	ExitReasonOther                     ExitReason = "other"
	ExitReasonBypassPermissionsDisabled ExitReason = "bypass_permissions_disabled"
)

// HookEvent identifies a lifecycle event that hooks can intercept.
type HookEvent string

const (
	HookEventPreToolUse         HookEvent = "PreToolUse"
	HookEventPostToolUse        HookEvent = "PostToolUse"
	HookEventPostToolUseFailure HookEvent = "PostToolUseFailure"
	HookEventNotification       HookEvent = "Notification"
	HookEventUserPromptSubmit   HookEvent = "UserPromptSubmit"
	HookEventSessionStart       HookEvent = "SessionStart"
	HookEventSessionEnd         HookEvent = "SessionEnd"
	HookEventStop               HookEvent = "Stop"
	HookEventStopFailure        HookEvent = "StopFailure"
	HookEventSubagentStart      HookEvent = "SubagentStart"
	HookEventSubagentStop       HookEvent = "SubagentStop"
	HookEventPreCompact         HookEvent = "PreCompact"
	HookEventPostCompact        HookEvent = "PostCompact"
	HookEventPermissionRequest  HookEvent = "PermissionRequest"
	HookEventSetup              HookEvent = "Setup"
	HookEventTeammateIdle       HookEvent = "TeammateIdle"
	HookEventTaskCompleted      HookEvent = "TaskCompleted"
	HookEventElicitation        HookEvent = "Elicitation"
	HookEventElicitationResult  HookEvent = "ElicitationResult"
	HookEventConfigChange       HookEvent = "ConfigChange"
	HookEventWorktreeCreate     HookEvent = "WorktreeCreate"
	HookEventWorktreeRemove     HookEvent = "WorktreeRemove"
	HookEventInstructionsLoaded HookEvent = "InstructionsLoaded"
)

// AllHookEvents returns all valid hook event values.
func AllHookEvents() []HookEvent {
	return []HookEvent{
		HookEventPreToolUse, HookEventPostToolUse, HookEventPostToolUseFailure,
		HookEventNotification, HookEventUserPromptSubmit, HookEventSessionStart,
		HookEventSessionEnd, HookEventStop, HookEventStopFailure,
		HookEventSubagentStart, HookEventSubagentStop, HookEventPreCompact,
		HookEventPostCompact, HookEventPermissionRequest, HookEventSetup,
		HookEventTeammateIdle, HookEventTaskCompleted, HookEventElicitation,
		HookEventElicitationResult, HookEventConfigChange, HookEventWorktreeCreate,
		HookEventWorktreeRemove, HookEventInstructionsLoaded,
	}
}

// PermissionBehavior describes how a permission rule acts.
type PermissionBehavior string

const (
	PermissionBehaviorAllow PermissionBehavior = "allow"
	PermissionBehaviorDeny  PermissionBehavior = "deny"
	PermissionBehaviorAsk   PermissionBehavior = "ask"
)

// FastModeState indicates whether fast mode is active.
type FastModeState string

const (
	FastModeStateOff      FastModeState = "off"
	FastModeStateCooldown FastModeState = "cooldown"
	FastModeStateOn       FastModeState = "on"
)

// SDKStatus represents system status states.
type SDKStatus *string

// SDKAssistantMessageError enumerates assistant message error types.
type SDKAssistantMessageError string

const (
	AssistantErrorAuthFailed      SDKAssistantMessageError = "authentication_failed"
	AssistantErrorBilling         SDKAssistantMessageError = "billing_error"
	AssistantErrorRateLimit       SDKAssistantMessageError = "rate_limit"
	AssistantErrorInvalidRequest  SDKAssistantMessageError = "invalid_request"
	AssistantErrorServer          SDKAssistantMessageError = "server_error"
	AssistantErrorUnknown         SDKAssistantMessageError = "unknown"
	AssistantErrorMaxOutputTokens SDKAssistantMessageError = "max_output_tokens"
)

// ApiKeySource identifies where the API key came from.
type ApiKeySource string

const (
	ApiKeySourceUser      ApiKeySource = "user"
	ApiKeySourceProject   ApiKeySource = "project"
	ApiKeySourceOrg       ApiKeySource = "org"
	ApiKeySourceTemporary ApiKeySource = "temporary"
	ApiKeySourceOAuth     ApiKeySource = "oauth"
)

// SettingSource identifies a settings file location.
type SettingSource string

const (
	SettingSourceUser    SettingSource = "user"
	SettingSourceProject SettingSource = "project"
	SettingSourceLocal   SettingSource = "local"
)

// ConfigScope identifies a configuration scope.
type ConfigScope string

const (
	ConfigScopeLocal   ConfigScope = "local"
	ConfigScopeUser    ConfigScope = "user"
	ConfigScopeProject ConfigScope = "project"
)

// OutputFormatType identifies output format types.
type OutputFormatType string

const (
	OutputFormatTypeJSONSchema OutputFormatType = "json_schema"
)

// PermissionUpdateDestination identifies where permission updates are stored.
type PermissionUpdateDestination string

const (
	PermissionDestUserSettings    PermissionUpdateDestination = "userSettings"
	PermissionDestProjectSettings PermissionUpdateDestination = "projectSettings"
	PermissionDestLocalSettings   PermissionUpdateDestination = "localSettings"
	PermissionDestSession         PermissionUpdateDestination = "session"
	PermissionDestCLIArg          PermissionUpdateDestination = "cliArg"
)

// SdkBeta identifies available beta features.
type SdkBeta string

const (
	SdkBetaContext1M SdkBeta = "context-1m-2025-08-07"
)
