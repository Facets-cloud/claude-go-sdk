package claudeagent

// McpStdioServerConfig defines a stdio-based MCP server.
type McpStdioServerConfig struct {
	Type    *string           `json:"type,omitempty"` // "stdio" or omitted
	Command string            `json:"command"`
	Args    []string          `json:"args,omitempty"`
	Env     map[string]string `json:"env,omitempty"`
}

// McpSSEServerConfig defines an SSE-based MCP server.
type McpSSEServerConfig struct {
	Type    string            `json:"type"` // "sse"
	URL     string            `json:"url"`
	Headers map[string]string `json:"headers,omitempty"`
}

// McpHttpServerConfig defines an HTTP-based MCP server.
type McpHttpServerConfig struct {
	Type    string            `json:"type"` // "http"
	URL     string            `json:"url"`
	Headers map[string]string `json:"headers,omitempty"`
}

// McpSdkServerConfig defines an in-process SDK MCP server (serializable config only).
type McpSdkServerConfig struct {
	Type string `json:"type"` // "sdk"
	Name string `json:"name"`
}

// McpSdkServerConfigWithInstance extends McpSdkServerConfig with a live server instance.
// Not serializable — contains a runtime reference.
type McpSdkServerConfigWithInstance struct {
	McpSdkServerConfig
	Instance interface{} `json:"-"` // runtime McpServer instance, not serializable
}

// McpClaudeAIProxyServerConfig defines a claude.ai proxy MCP server.
type McpClaudeAIProxyServerConfig struct {
	Type string `json:"type"` // "claudeai-proxy"
	URL  string `json:"url"`
	ID   string `json:"id"`
}

// McpServerConfig is a union of all MCP server configuration types including non-serializable instances.
// Use json.RawMessage and inspect the "type" field to determine the concrete type.
type McpServerConfig = interface{}

// McpServerConfigForProcessTransport is a union of serializable MCP server config types
// (excludes McpSdkServerConfigWithInstance).
type McpServerConfigForProcessTransport = interface{}

// McpServerStatusConfig is a union of process transport configs plus claude.ai proxy.
type McpServerStatusConfig = interface{}

// McpServerStatus describes the current status of an MCP server connection.
type McpServerStatus struct {
	Name       string              `json:"name"`
	Status     string              `json:"status"` // "connected" | "failed" | "needs-auth" | "pending" | "disabled"
	ServerInfo *McpServerInfo      `json:"serverInfo,omitempty"`
	Error      *string             `json:"error,omitempty"`
	Config     interface{}         `json:"config,omitempty"` // McpServerStatusConfig
	Scope      *string             `json:"scope,omitempty"`
	Tools      []McpServerToolInfo `json:"tools,omitempty"`
}

// McpServerInfo contains name and version of a connected MCP server.
type McpServerInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

// McpServerToolInfo describes a tool provided by an MCP server.
type McpServerToolInfo struct {
	Name        string              `json:"name"`
	Description *string             `json:"description,omitempty"`
	Annotations *McpToolAnnotations `json:"annotations,omitempty"`
}

// McpToolAnnotations provides metadata about a tool's behavior.
type McpToolAnnotations struct {
	ReadOnly    *bool `json:"readOnly,omitempty"`
	Destructive *bool `json:"destructive,omitempty"`
	OpenWorld   *bool `json:"openWorld,omitempty"`
}

// McpSetServersResult is the result of a setMcpServers operation.
type McpSetServersResult struct {
	Added   []string          `json:"added"`
	Removed []string          `json:"removed"`
	Errors  map[string]string `json:"errors"`
}
