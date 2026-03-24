package claudeagent

// ModelInfo describes an available model.
type ModelInfo struct {
	Value                    string   `json:"value"`
	DisplayName              string   `json:"displayName"`
	Description              string   `json:"description"`
	SupportsEffort           *bool    `json:"supportsEffort,omitempty"`
	SupportedEffortLevels    []string `json:"supportedEffortLevels,omitempty"`
	SupportsAdaptiveThinking *bool    `json:"supportsAdaptiveThinking,omitempty"`
	SupportsFastMode         *bool    `json:"supportsFastMode,omitempty"`
	SupportsAutoMode         *bool    `json:"supportsAutoMode,omitempty"`
}

// AccountInfo describes the authenticated user's account.
type AccountInfo struct {
	Email            *string `json:"email,omitempty"`
	Organization     *string `json:"organization,omitempty"`
	SubscriptionType *string `json:"subscriptionType,omitempty"`
	TokenSource      *string `json:"tokenSource,omitempty"`
	ApiKeySource     *string `json:"apiKeySource,omitempty"`
	ApiProvider      *string `json:"apiProvider,omitempty"` // "firstParty" | "bedrock" | "vertex" | "foundry"
}

// AgentDefinition defines a custom subagent.
type AgentDefinition struct {
	Description            string        `json:"description"`
	Tools                  []string      `json:"tools,omitempty"`
	DisallowedTools        []string      `json:"disallowedTools,omitempty"`
	Prompt                 string        `json:"prompt"`
	Model                  *string       `json:"model,omitempty"`
	McpServers             []interface{} `json:"mcpServers,omitempty"` // string or map[string]McpServerConfigForProcessTransport
	CriticalSystemReminder *string       `json:"criticalSystemReminder_EXPERIMENTAL,omitempty"`
	Skills                 []string      `json:"skills,omitempty"`
	MaxTurns               *int          `json:"maxTurns,omitempty"`
}

// AgentInfo describes an available subagent.
type AgentInfo struct {
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Model       *string `json:"model,omitempty"`
}

// SlashCommand describes an available slash command/skill.
type SlashCommand struct {
	Name         string `json:"name"`
	Description  string `json:"description"`
	ArgumentHint string `json:"argumentHint"`
}

// PromptRequest is a prompt request from the CLI asking the SDK consumer to choose.
type PromptRequest struct {
	Prompt  string                `json:"prompt"`
	Message string                `json:"message"`
	Options []PromptRequestOption `json:"options"`
}

// PromptRequestOption is an option in a PromptRequest.
type PromptRequestOption struct {
	Key         string  `json:"key"`
	Label       string  `json:"label"`
	Description *string `json:"description,omitempty"`
}

// PromptResponse is a response to a PromptRequest.
type PromptResponse struct {
	PromptResponse string `json:"prompt_response"`
	Selected       string `json:"selected"`
}
