package claudeagent

import "encoding/json"

// Settings represents the complete Claude Code settings configuration.
// This struct covers ALL settings.json options including enterprise, hooks,
// marketplace, plugins, sandbox, permissions, and UI settings.
type Settings struct {
	// JSON Schema reference
	Schema *string `json:"$schema,omitempty"`

	// --- Authentication ---

	ApiKeyHelper      *string `json:"apiKeyHelper,omitempty"`
	AwsCredentialExport *string `json:"awsCredentialExport,omitempty"`
	AwsAuthRefresh    *string `json:"awsAuthRefresh,omitempty"`
	GcpAuthRefresh    *string `json:"gcpAuthRefresh,omitempty"`
	ForceLoginMethod  *string `json:"forceLoginMethod,omitempty"`  // "claudeai" | "console"
	ForceLoginOrgUUID *string `json:"forceLoginOrgUUID,omitempty"`
	OtelHeadersHelper *string `json:"otelHeadersHelper,omitempty"`

	// --- File and Project ---

	FileSuggestion   *SettingsFileSuggestion `json:"fileSuggestion,omitempty"`
	RespectGitignore *bool                   `json:"respectGitignore,omitempty"`
	CleanupPeriodDays *int                   `json:"cleanupPeriodDays,omitempty"`

	// --- Environment ---

	Env map[string]string `json:"env,omitempty"`

	// --- Attribution ---

	Attribution      *SettingsAttribution `json:"attribution,omitempty"`
	IncludeCoAuthoredBy *bool             `json:"includeCoAuthoredBy,omitempty"` // deprecated: use attribution
	IncludeGitInstructions *bool           `json:"includeGitInstructions,omitempty"`

	// --- Permissions ---

	Permissions *SettingsPermissions `json:"permissions,omitempty"`

	// --- Model ---

	Model          *string           `json:"model,omitempty"`
	AvailableModels []string         `json:"availableModels,omitempty"`
	ModelOverrides map[string]string `json:"modelOverrides,omitempty"`

	// --- MCP Servers ---

	EnableAllProjectMcpServers *bool              `json:"enableAllProjectMcpServers,omitempty"`
	EnabledMcpjsonServers      []string           `json:"enabledMcpjsonServers,omitempty"`
	DisabledMcpjsonServers     []string           `json:"disabledMcpjsonServers,omitempty"`
	AllowedMcpServers          []McpServerMatcher `json:"allowedMcpServers,omitempty"`
	DeniedMcpServers           []McpServerMatcher `json:"deniedMcpServers,omitempty"`

	// --- Hooks ---

	Hooks           map[string][]SettingsHookMatcher `json:"hooks,omitempty"`
	DisableAllHooks *bool                            `json:"disableAllHooks,omitempty"`
	DefaultShell    *string                          `json:"defaultShell,omitempty"` // "bash" | "powershell"

	// --- Enterprise Hook Controls ---

	AllowManagedHooksOnly           *bool    `json:"allowManagedHooksOnly,omitempty"`
	AllowedHttpHookUrls             []string `json:"allowedHttpHookUrls,omitempty"`
	HttpHookAllowedEnvVars          []string `json:"httpHookAllowedEnvVars,omitempty"`
	AllowManagedPermissionRulesOnly *bool    `json:"allowManagedPermissionRulesOnly,omitempty"`
	AllowManagedMcpServersOnly      *bool    `json:"allowManagedMcpServersOnly,omitempty"`

	// --- Plugin Customization ---

	StrictPluginOnlyCustomization json.RawMessage `json:"strictPluginOnlyCustomization,omitempty"` // bool | []string

	// --- Worktree ---

	Worktree *SettingsWorktree `json:"worktree,omitempty"`

	// --- Status Line ---

	StatusLine *SettingsStatusLine `json:"statusLine,omitempty"`

	// --- Plugins ---

	EnabledPlugins          map[string]json.RawMessage  `json:"enabledPlugins,omitempty"` // bool | []string | object
	ExtraKnownMarketplaces  map[string]MarketplaceEntry `json:"extraKnownMarketplaces,omitempty"`
	StrictKnownMarketplaces []json.RawMessage           `json:"strictKnownMarketplaces,omitempty"` // MarketplaceSource variants
	BlockedMarketplaces     []json.RawMessage           `json:"blockedMarketplaces,omitempty"`     // MarketplaceSource variants
	PluginConfigs           map[string]PluginConfig     `json:"pluginConfigs,omitempty"`
	PluginTrustMessage      *string                     `json:"pluginTrustMessage,omitempty"`

	// --- Sandbox ---

	Sandbox *SandboxSettings `json:"sandbox,omitempty"`

	// --- UI/Behavior ---

	OutputStyle               *string                `json:"outputStyle,omitempty"`
	Language                  *string                `json:"language,omitempty"`
	SyntaxHighlightingDisabled *bool                 `json:"syntaxHighlightingDisabled,omitempty"`
	TerminalTitleFromRename   *bool                  `json:"terminalTitleFromRename,omitempty"`
	AlwaysThinkingEnabled     *bool                  `json:"alwaysThinkingEnabled,omitempty"`
	EffortLevel               *string                `json:"effortLevel,omitempty"` // "low" | "medium" | "high"
	FastMode                  *bool                  `json:"fastMode,omitempty"`
	FastModePerSessionOptIn   *bool                  `json:"fastModePerSessionOptIn,omitempty"`
	PromptSuggestionEnabled   *bool                  `json:"promptSuggestionEnabled,omitempty"`
	ShowClearContextOnPlanAccept *bool                `json:"showClearContextOnPlanAccept,omitempty"`
	FeedbackSurveyRate        *float64               `json:"feedbackSurveyRate,omitempty"`
	SpinnerTipsEnabled        *bool                  `json:"spinnerTipsEnabled,omitempty"`
	SpinnerVerbs              *SettingsSpinnerVerbs   `json:"spinnerVerbs,omitempty"`
	SpinnerTipsOverride       *SettingsSpinnerTips    `json:"spinnerTipsOverride,omitempty"`
	PrefersReducedMotion      *bool                  `json:"prefersReducedMotion,omitempty"`
	ShowThinkingSummaries     *bool                  `json:"showThinkingSummaries,omitempty"`
	SkipDangerousModePermissionPrompt *bool           `json:"skipDangerousModePermissionPrompt,omitempty"`
	DisableAutoMode           *string                `json:"disableAutoMode,omitempty"` // "disable"
	SkipWebFetchPreflight     *bool                  `json:"skipWebFetchPreflight,omitempty"`

	// --- Agent ---

	Agent               *string  `json:"agent,omitempty"`
	CompanyAnnouncements []string `json:"companyAnnouncements,omitempty"`

	// --- Remote ---

	Remote *SettingsRemote `json:"remote,omitempty"`

	// --- Auto-update ---

	AutoUpdatesChannel *string `json:"autoUpdatesChannel,omitempty"` // "latest" | "stable"
	MinimumVersion     *string `json:"minimumVersion,omitempty"`

	// --- Plans ---

	PlansDirectory *string `json:"plansDirectory,omitempty"`

	// --- Memory ---

	AutoMemoryEnabled   *bool   `json:"autoMemoryEnabled,omitempty"`
	AutoMemoryDirectory *string `json:"autoMemoryDirectory,omitempty"`
	AutoDreamEnabled    *bool   `json:"autoDreamEnabled,omitempty"`

	// --- SSH ---

	SSHConfigs []SSHConfig `json:"sshConfigs,omitempty"`

	// --- CLAUDE.md ---

	ClaudeMdExcludes []string `json:"claudeMdExcludes,omitempty"`
}

// --- Nested Settings Types ---

// SettingsFileSuggestion configures file suggestion for @ mentions.
type SettingsFileSuggestion struct {
	Type    string `json:"type"` // "command"
	Command string `json:"command"`
}

// SettingsAttribution configures attribution text for commits/PRs.
type SettingsAttribution struct {
	Commit *string `json:"commit,omitempty"`
	PR     *string `json:"pr,omitempty"`
}

// SettingsPermissions configures tool usage permissions.
type SettingsPermissions struct {
	Allow                        []string `json:"allow,omitempty"`
	Deny                         []string `json:"deny,omitempty"`
	Ask                          []string `json:"ask,omitempty"`
	DefaultMode                  *string  `json:"defaultMode,omitempty"` // PermissionMode values
	DisableBypassPermissionsMode *string  `json:"disableBypassPermissionsMode,omitempty"` // "disable"
	AdditionalDirectories        []string `json:"additionalDirectories,omitempty"`
}

// McpServerMatcher matches MCP servers by name, command, or URL.
type McpServerMatcher struct {
	ServerName    *string  `json:"serverName,omitempty"`
	ServerCommand []string `json:"serverCommand,omitempty"` // [command, ...args]
	ServerUrl     *string  `json:"serverUrl,omitempty"`
}

// SettingsHookMatcher matches hook events and invokes hooks.
type SettingsHookMatcher struct {
	Matcher *string              `json:"matcher,omitempty"`
	Hooks   []SettingsHookConfig `json:"hooks"`
}

// SettingsHookConfig is a union of the 4 hook types: command, prompt, agent, http.
// Use the Type field to determine which fields are relevant.
type SettingsHookConfig struct {
	// Common fields
	Type          string  `json:"type"` // "command" | "prompt" | "agent" | "http"
	Timeout       *int    `json:"timeout,omitempty"`
	StatusMessage *string `json:"statusMessage,omitempty"`
	Once          *bool   `json:"once,omitempty"`

	// Command hook fields
	Command      *string `json:"command,omitempty"`
	Shell        *string `json:"shell,omitempty"` // "bash" | "powershell"
	Async        *bool   `json:"async,omitempty"`
	AsyncRewake  *bool   `json:"asyncRewake,omitempty"`

	// Prompt/Agent hook fields
	Prompt *string `json:"prompt,omitempty"`
	Model  *string `json:"model,omitempty"`

	// HTTP hook fields
	URL            *string           `json:"url,omitempty"`
	Headers        map[string]string `json:"headers,omitempty"`
	AllowedEnvVars []string          `json:"allowedEnvVars,omitempty"`
}

// SettingsWorktree configures git worktree behavior.
type SettingsWorktree struct {
	SymlinkDirectories []string `json:"symlinkDirectories,omitempty"`
	SparsePaths        []string `json:"sparsePaths,omitempty"`
}

// SettingsStatusLine configures the custom status line.
type SettingsStatusLine struct {
	Type    string `json:"type"` // "command"
	Command string `json:"command"`
	Padding *int   `json:"padding,omitempty"`
}

// SettingsSpinnerVerbs customizes spinner verbs.
type SettingsSpinnerVerbs struct {
	Mode  string   `json:"mode"` // "append" | "replace"
	Verbs []string `json:"verbs"`
}

// SettingsSpinnerTips overrides spinner tips.
type SettingsSpinnerTips struct {
	ExcludeDefault *bool    `json:"excludeDefault,omitempty"`
	Tips           []string `json:"tips"`
}

// SettingsRemote configures remote session behavior.
type SettingsRemote struct {
	DefaultEnvironmentID *string `json:"defaultEnvironmentId,omitempty"`
}

// SSHConfig configures an SSH connection for remote environments.
type SSHConfig struct {
	ID              string  `json:"id"`
	Name            string  `json:"name"`
	SSHHost         string  `json:"sshHost"`
	SSHPort         *int    `json:"sshPort,omitempty"`
	SSHIdentityFile *string `json:"sshIdentityFile,omitempty"`
	StartDirectory  *string `json:"startDirectory,omitempty"`
}

// --- Marketplace Types ---

// MarketplaceEntry is an entry in extraKnownMarketplaces.
type MarketplaceEntry struct {
	Source          json.RawMessage `json:"source"`
	InstallLocation *string         `json:"installLocation,omitempty"`
	AutoUpdate      *bool           `json:"autoUpdate,omitempty"`
}

// PluginConfig provides per-plugin configuration including MCP server configs.
type PluginConfig struct {
	McpServers map[string]map[string]interface{} `json:"mcpServers,omitempty"`
	Options    map[string]interface{}            `json:"options,omitempty"`
}
