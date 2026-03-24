package claudeagent

import "encoding/json"

// Settings represents the full Claude Code settings configuration.
type Settings struct {
	Schema                         *string                                `json:"$schema,omitempty"`
	ApiKeyHelper                   *string                                `json:"apiKeyHelper,omitempty"`
	AwsCredentialExport            *string                                `json:"awsCredentialExport,omitempty"`
	AwsAuthRefresh                 *string                                `json:"awsAuthRefresh,omitempty"`
	GcpAuthRefresh                 *string                                `json:"gcpAuthRefresh,omitempty"`
	FileSuggestion                 *SettingsFileSuggestion                `json:"fileSuggestion,omitempty"`
	RespectGitignore               *bool                                  `json:"respectGitignore,omitempty"`
	CleanupPeriodDays              *int                                   `json:"cleanupPeriodDays,omitempty"`
	Env                            map[string]string                      `json:"env,omitempty"`
	Attribution                    *SettingsAttribution                   `json:"attribution,omitempty"`
	IncludeCoAuthoredBy            *bool                                  `json:"includeCoAuthoredBy,omitempty"`
	IncludeGitInstructions         *bool                                  `json:"includeGitInstructions,omitempty"`
	Permissions                    *SettingsPermissions                   `json:"permissions,omitempty"`
	Model                          *string                                `json:"model,omitempty"`
	AvailableModels                []string                               `json:"availableModels,omitempty"`
	ModelOverrides                 map[string]string                      `json:"modelOverrides,omitempty"`
	EnableAllProjectMcpServers     *bool                                  `json:"enableAllProjectMcpServers,omitempty"`
	EnabledMcpjsonServers          []string                               `json:"enabledMcpjsonServers,omitempty"`
	DisabledMcpjsonServers         []string                               `json:"disabledMcpjsonServers,omitempty"`
	AllowedMcpServers              []SettingsMcpServerRule                `json:"allowedMcpServers,omitempty"`
	DeniedMcpServers               []SettingsMcpServerRule                `json:"deniedMcpServers,omitempty"`
	Hooks                          map[string][]SettingsHookMatcher       `json:"hooks,omitempty"`
	Worktree                       *SettingsWorktree                      `json:"worktree,omitempty"`
	DisableAllHooks                *bool                                  `json:"disableAllHooks,omitempty"`
	DefaultShell                   *string                                `json:"defaultShell,omitempty"`
	AllowManagedHooksOnly          *bool                                  `json:"allowManagedHooksOnly,omitempty"`
	AllowedHttpHookUrls            []string                               `json:"allowedHttpHookUrls,omitempty"`
	HttpHookAllowedEnvVars         []string                               `json:"httpHookAllowedEnvVars,omitempty"`
	AllowManagedPermissionRulesOnly *bool                                 `json:"allowManagedPermissionRulesOnly,omitempty"`
	AllowManagedMcpServersOnly     *bool                                  `json:"allowManagedMcpServersOnly,omitempty"`
	StrictPluginOnlyCustomization  json.RawMessage                        `json:"strictPluginOnlyCustomization,omitempty"`
	StatusLine                     *SettingsStatusLine                    `json:"statusLine,omitempty"`
	EnabledPlugins                 map[string]json.RawMessage             `json:"enabledPlugins,omitempty"`
	ExtraKnownMarketplaces         map[string]SettingsMarketplaceConfig   `json:"extraKnownMarketplaces,omitempty"`
	StrictKnownMarketplaces        []json.RawMessage                      `json:"strictKnownMarketplaces,omitempty"`
	BlockedMarketplaces            []json.RawMessage                      `json:"blockedMarketplaces,omitempty"`
	ForceLoginMethod               *string                                `json:"forceLoginMethod,omitempty"`
	ForceLoginOrgUUID              *string                                `json:"forceLoginOrgUUID,omitempty"`
	OtelHeadersHelper              *string                                `json:"otelHeadersHelper,omitempty"`
	OutputStyle                    *string                                `json:"outputStyle,omitempty"`
	Language                       *string                                `json:"language,omitempty"`
	SkipWebFetchPreflight          *bool                                  `json:"skipWebFetchPreflight,omitempty"`
	Sandbox                        *SandboxSettings                       `json:"sandbox,omitempty"`
	FeedbackSurveyRate             *float64                               `json:"feedbackSurveyRate,omitempty"`
	SpinnerTipsEnabled             *bool                                  `json:"spinnerTipsEnabled,omitempty"`
	SpinnerVerbs                   *SettingsSpinnerVerbs                  `json:"spinnerVerbs,omitempty"`
	SpinnerTipsOverride            *SettingsSpinnerTipsOverride           `json:"spinnerTipsOverride,omitempty"`
	SyntaxHighlightingDisabled     *bool                                  `json:"syntaxHighlightingDisabled,omitempty"`
	TerminalTitleFromRename        *bool                                  `json:"terminalTitleFromRename,omitempty"`
	AlwaysThinkingEnabled          *bool                                  `json:"alwaysThinkingEnabled,omitempty"`
	EffortLevel                    *string                                `json:"effortLevel,omitempty"`
	FastMode                       *bool                                  `json:"fastMode,omitempty"`
	FastModePerSessionOptIn        *bool                                  `json:"fastModePerSessionOptIn,omitempty"`
	PromptSuggestionEnabled        *bool                                  `json:"promptSuggestionEnabled,omitempty"`
	ShowClearContextOnPlanAccept   *bool                                  `json:"showClearContextOnPlanAccept,omitempty"`
	Agent                          *string                                `json:"agent,omitempty"`
	CompanyAnnouncements           []string                               `json:"companyAnnouncements,omitempty"`
	PluginConfigs                  map[string]json.RawMessage             `json:"pluginConfigs,omitempty"`
	Remote                         *SettingsRemote                        `json:"remote,omitempty"`
	AutoUpdatesChannel             *string                                `json:"autoUpdatesChannel,omitempty"`
	MinimumVersion                 *string                                `json:"minimumVersion,omitempty"`
	PlansDirectory                 *string                                `json:"plansDirectory,omitempty"`
	PrefersReducedMotion           *bool                                  `json:"prefersReducedMotion,omitempty"`
	AutoMemoryEnabled              *bool                                  `json:"autoMemoryEnabled,omitempty"`
	AutoMemoryDirectory            *string                                `json:"autoMemoryDirectory,omitempty"`
	AutoDreamEnabled               *bool                                  `json:"autoDreamEnabled,omitempty"`
	ShowThinkingSummaries          *bool                                  `json:"showThinkingSummaries,omitempty"`
	SkipDangerousModePermissionPrompt *bool                               `json:"skipDangerousModePermissionPrompt,omitempty"`
	DisableAutoMode                *string                                `json:"disableAutoMode,omitempty"`
	SSHConfigs                     []SettingsSSHConfig                    `json:"sshConfigs,omitempty"`
	ClaudeMdExcludes               []string                               `json:"claudeMdExcludes,omitempty"`
}

// SettingsPermissions configures tool usage permissions.
type SettingsPermissions struct {
	Allow                         []string `json:"allow,omitempty"`
	Deny                          []string `json:"deny,omitempty"`
	Ask                           []string `json:"ask,omitempty"`
	DefaultMode                   *string  `json:"defaultMode,omitempty"`
	DisableBypassPermissionsMode  *string  `json:"disableBypassPermissionsMode,omitempty"`
	AdditionalDirectories         []string `json:"additionalDirectories,omitempty"`
}

// SettingsFileSuggestion configures custom file suggestion for @ mentions.
type SettingsFileSuggestion struct {
	Type    string `json:"type"`    // "command"
	Command string `json:"command"`
}

// SettingsAttribution configures attribution text for commits and PRs.
type SettingsAttribution struct {
	Commit *string `json:"commit,omitempty"`
	PR     *string `json:"pr,omitempty"`
}

// SettingsMcpServerRule identifies an allowed or denied MCP server.
type SettingsMcpServerRule struct {
	ServerName    *string  `json:"serverName,omitempty"`
	ServerCommand []string `json:"serverCommand,omitempty"`
	ServerUrl     *string  `json:"serverUrl,omitempty"`
}

// SettingsHookMatcher contains a pattern matcher and associated hooks.
type SettingsHookMatcher struct {
	Matcher *string              `json:"matcher,omitempty"`
	Hooks   []SettingsHookConfig `json:"hooks"`
}

// SettingsHookConfig represents a hook configuration (command, prompt, agent, or http).
type SettingsHookConfig struct {
	Type          string            `json:"type"`                    // "command" | "prompt" | "agent" | "http"
	Command       *string           `json:"command,omitempty"`       // for type="command"
	Shell         *string           `json:"shell,omitempty"`         // for type="command"
	Prompt        *string           `json:"prompt,omitempty"`        // for type="prompt" or "agent"
	URL           *string           `json:"url,omitempty"`           // for type="http"
	Headers       map[string]string `json:"headers,omitempty"`       // for type="http"
	AllowedEnvVars []string         `json:"allowedEnvVars,omitempty"` // for type="http"
	Timeout       *int              `json:"timeout,omitempty"`
	StatusMessage *string           `json:"statusMessage,omitempty"`
	Once          *bool             `json:"once,omitempty"`
	Async         *bool             `json:"async,omitempty"`         // for type="command"
	AsyncRewake   *bool             `json:"asyncRewake,omitempty"`   // for type="command"
	Model         *string           `json:"model,omitempty"`         // for type="prompt" or "agent"
}

// SettingsWorktree configures git worktree behavior.
type SettingsWorktree struct {
	SymlinkDirectories []string `json:"symlinkDirectories,omitempty"`
	SparsePaths        []string `json:"sparsePaths,omitempty"`
}

// SettingsStatusLine configures a custom status line display.
type SettingsStatusLine struct {
	Type    string `json:"type"` // "command"
	Command string `json:"command"`
	Padding *int   `json:"padding,omitempty"`
}

// SettingsMarketplaceConfig configures an additional marketplace.
type SettingsMarketplaceConfig struct {
	Source          json.RawMessage `json:"source"`
	InstallLocation *string        `json:"installLocation,omitempty"`
	AutoUpdate      *bool          `json:"autoUpdate,omitempty"`
}

// SettingsSpinnerVerbs configures custom spinner verbs.
type SettingsSpinnerVerbs struct {
	Mode  string   `json:"mode"` // "append" | "replace"
	Verbs []string `json:"verbs"`
}

// SettingsSpinnerTipsOverride configures custom spinner tips.
type SettingsSpinnerTipsOverride struct {
	ExcludeDefault *bool    `json:"excludeDefault,omitempty"`
	Tips           []string `json:"tips"`
}

// SettingsRemote configures remote session behavior.
type SettingsRemote struct {
	DefaultEnvironmentId *string `json:"defaultEnvironmentId,omitempty"`
}

// SettingsSSHConfig configures an SSH connection for remote environments.
type SettingsSSHConfig struct {
	ID              string  `json:"id"`
	Name            string  `json:"name"`
	SSHHost         string  `json:"sshHost"`
	SSHPort         *int    `json:"sshPort,omitempty"`
	SSHIdentityFile *string `json:"sshIdentityFile,omitempty"`
	StartDirectory  *string `json:"startDirectory,omitempty"`
}