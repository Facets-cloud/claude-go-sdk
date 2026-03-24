package tools

import "encoding/json"

// AgentInput is the input for the Agent tool, which spawns a sub-agent.
type AgentInput struct {
	// Description is a short (3-5 word) description of the task.
	Description string `json:"description"`
	// Prompt is the task for the agent to perform.
	Prompt string `json:"prompt"`
	// SubagentType is the type of specialized agent to use.
	SubagentType *string `json:"subagent_type,omitempty"`
	// Model overrides the agent's model. One of "sonnet", "opus", "haiku".
	Model *string `json:"model,omitempty"`
	// RunInBackground runs this agent in the background when true.
	RunInBackground *bool `json:"run_in_background,omitempty"`
	// Name makes the agent addressable via SendMessage.
	Name *string `json:"name,omitempty"`
	// TeamName is the team name for spawning.
	TeamName *string `json:"team_name,omitempty"`
	// Mode is the permission mode for the spawned teammate.
	Mode *string `json:"mode,omitempty"`
	// Isolation mode. "worktree" creates a temporary git worktree.
	Isolation *string `json:"isolation,omitempty"`
}

// AgentOutput is a union type for agent tool results.
// It can be either AgentOutputCompleted or AgentOutputAsyncLaunched.
type AgentOutput interface {
	agentOutput()
}

// AgentUsage contains token usage information.
type AgentUsage struct {
	InputTokens                int              `json:"input_tokens"`
	OutputTokens               int              `json:"output_tokens"`
	CacheCreationInputTokens   *int             `json:"cache_creation_input_tokens"`
	CacheReadInputTokens       *int             `json:"cache_read_input_tokens"`
	ServerToolUse              *ServerToolUse    `json:"server_tool_use"`
	ServiceTier                *string           `json:"service_tier"`
	CacheCreation              *CacheCreation    `json:"cache_creation"`
}

// ServerToolUse tracks server-side tool use counts.
type ServerToolUse struct {
	WebSearchRequests int `json:"web_search_requests"`
	WebFetchRequests  int `json:"web_fetch_requests"`
}

// CacheCreation tracks cache creation token counts.
type CacheCreation struct {
	Ephemeral1hInputTokens int `json:"ephemeral_1h_input_tokens"`
	Ephemeral5mInputTokens int `json:"ephemeral_5m_input_tokens"`
}

// AgentContentBlock is a content block in a completed agent output.
type AgentContentBlock struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

// AgentOutputCompleted is returned when an agent finishes its work.
type AgentOutputCompleted struct {
	AgentID           string              `json:"agentId"`
	Content           []AgentContentBlock `json:"content"`
	TotalToolUseCount int                 `json:"totalToolUseCount"`
	TotalDurationMs   int                 `json:"totalDurationMs"`
	TotalTokens       int                 `json:"totalTokens"`
	Usage             AgentUsage          `json:"usage"`
	Status            string              `json:"status"` // always "completed"
	Prompt            string              `json:"prompt"`
}

func (AgentOutputCompleted) agentOutput() {}

// AgentOutputAsyncLaunched is returned when a background agent is launched.
type AgentOutputAsyncLaunched struct {
	Status            string `json:"status"` // always "async_launched"
	AgentID           string `json:"agentId"`
	Description       string `json:"description"`
	Prompt            string `json:"prompt"`
	OutputFile        string `json:"outputFile"`
	CanReadOutputFile *bool  `json:"canReadOutputFile,omitempty"`
}

func (AgentOutputAsyncLaunched) agentOutput() {}

// UnmarshalAgentOutput unmarshals JSON into the correct AgentOutput variant.
func UnmarshalAgentOutput(data []byte) (AgentOutput, error) {
	var probe struct {
		Status string `json:"status"`
	}
	if err := json.Unmarshal(data, &probe); err != nil {
		return nil, err
	}
	switch probe.Status {
	case "async_launched":
		var out AgentOutputAsyncLaunched
		if err := json.Unmarshal(data, &out); err != nil {
			return nil, err
		}
		return out, nil
	default:
		var out AgentOutputCompleted
		if err := json.Unmarshal(data, &out); err != nil {
			return nil, err
		}
		return out, nil
	}
}
