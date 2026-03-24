package claudeagent

// SandboxSettings configures the sandboxing behavior for tool execution.
type SandboxSettings struct {
	Enabled                      *bool                      `json:"enabled,omitempty"`
	AutoAllowBashIfSandboxed     *bool                      `json:"autoAllowBashIfSandboxed,omitempty"`
	AllowUnsandboxedCommands     *bool                      `json:"allowUnsandboxedCommands,omitempty"`
	Network                      *SandboxNetworkConfig      `json:"network,omitempty"`
	Filesystem                   *SandboxFilesystemConfig   `json:"filesystem,omitempty"`
	IgnoreViolations             map[string][]string        `json:"ignoreViolations,omitempty"`
	EnableWeakerNestedSandbox    *bool                      `json:"enableWeakerNestedSandbox,omitempty"`
	EnableWeakerNetworkIsolation *bool                      `json:"enableWeakerNetworkIsolation,omitempty"`
	ExcludedCommands             []string                   `json:"excludedCommands,omitempty"`
	Ripgrep                      *SandboxRipgrepConfig      `json:"ripgrep,omitempty"`
}

// SandboxNetworkConfig configures network access within the sandbox.
type SandboxNetworkConfig struct {
	AllowedDomains         []string `json:"allowedDomains,omitempty"`
	AllowManagedDomainsOnly *bool   `json:"allowManagedDomainsOnly,omitempty"`
	AllowUnixSockets       []string `json:"allowUnixSockets,omitempty"`
	AllowAllUnixSockets    *bool    `json:"allowAllUnixSockets,omitempty"`
	AllowLocalBinding      *bool    `json:"allowLocalBinding,omitempty"`
	HttpProxyPort          *int     `json:"httpProxyPort,omitempty"`
	SocksProxyPort         *int     `json:"socksProxyPort,omitempty"`
}

// SandboxFilesystemConfig configures filesystem access within the sandbox.
type SandboxFilesystemConfig struct {
	AllowWrite              []string `json:"allowWrite,omitempty"`
	DenyWrite               []string `json:"denyWrite,omitempty"`
	DenyRead                []string `json:"denyRead,omitempty"`
	AllowRead               []string `json:"allowRead,omitempty"`
	AllowManagedReadPathsOnly *bool  `json:"allowManagedReadPathsOnly,omitempty"`
}

// SandboxRipgrepConfig configures the ripgrep binary used within the sandbox.
type SandboxRipgrepConfig struct {
	Command string   `json:"command"`
	Args    []string `json:"args,omitempty"`
}

// SandboxIgnoreViolations is a map of violation categories to ignored violation patterns.
type SandboxIgnoreViolations = map[string][]string
