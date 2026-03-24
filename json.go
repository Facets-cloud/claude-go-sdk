package claudeagent

import (
	"encoding/json"
	"fmt"
)

// messageEnvelope is used to peek at type/subtype before full deserialization.
type messageEnvelope struct {
	Type    string `json:"type"`
	Subtype string `json:"subtype,omitempty"`
}

// ParseSDKMessage deserializes raw JSON into the appropriate SDKMessage concrete type.
func ParseSDKMessage(data []byte) (SDKMessage, error) {
	var env messageEnvelope
	if err := json.Unmarshal(data, &env); err != nil {
		return nil, fmt.Errorf("parse message envelope: %w", err)
	}

	switch env.Type {
	case "assistant":
		var m SDKAssistantMessage
		if err := json.Unmarshal(data, &m); err != nil {
			return nil, fmt.Errorf("parse assistant message: %w", err)
		}
		return &m, nil

	case "user":
		var peek struct {
			IsReplay bool `json:"isReplay"`
		}
		_ = json.Unmarshal(data, &peek)
		if peek.IsReplay {
			var m SDKUserMessageReplay
			if err := json.Unmarshal(data, &m); err != nil {
				return nil, fmt.Errorf("parse user replay message: %w", err)
			}
			return &m, nil
		}
		var m SDKUserMessage
		if err := json.Unmarshal(data, &m); err != nil {
			return nil, fmt.Errorf("parse user message: %w", err)
		}
		return &m, nil

	case "result":
		switch env.Subtype {
		case "success":
			var m SDKResultSuccess
			if err := json.Unmarshal(data, &m); err != nil {
				return nil, fmt.Errorf("parse result success: %w", err)
			}
			return &m, nil
		default:
			var m SDKResultError
			if err := json.Unmarshal(data, &m); err != nil {
				return nil, fmt.Errorf("parse result error: %w", err)
			}
			return &m, nil
		}

	case "system":
		return parseSystemMessage(env.Subtype, data)

	case "stream_event":
		var m SDKPartialAssistantMessage
		if err := json.Unmarshal(data, &m); err != nil {
			return nil, fmt.Errorf("parse stream event: %w", err)
		}
		return &m, nil

	case "tool_progress":
		var m SDKToolProgressMessage
		if err := json.Unmarshal(data, &m); err != nil {
			return nil, fmt.Errorf("parse tool progress: %w", err)
		}
		return &m, nil

	case "tool_use_summary":
		var m SDKToolUseSummaryMessage
		if err := json.Unmarshal(data, &m); err != nil {
			return nil, fmt.Errorf("parse tool use summary: %w", err)
		}
		return &m, nil

	case "auth_status":
		var m SDKAuthStatusMessage
		if err := json.Unmarshal(data, &m); err != nil {
			return nil, fmt.Errorf("parse auth status: %w", err)
		}
		return &m, nil

	case "rate_limit_event":
		var m SDKRateLimitEvent
		if err := json.Unmarshal(data, &m); err != nil {
			return nil, fmt.Errorf("parse rate limit event: %w", err)
		}
		return &m, nil

	case "prompt_suggestion":
		var m SDKPromptSuggestionMessage
		if err := json.Unmarshal(data, &m); err != nil {
			return nil, fmt.Errorf("parse prompt suggestion: %w", err)
		}
		return &m, nil

	default:
		return &SDKRawMessage{RawType: env.Type, Raw: json.RawMessage(data)}, nil
	}
}

func parseSystemMessage(subtype string, data []byte) (SDKMessage, error) {
	switch subtype {
	case "init":
		var m SDKSystemMessage
		if err := json.Unmarshal(data, &m); err != nil {
			return nil, fmt.Errorf("parse system init: %w", err)
		}
		return &m, nil
	case "status":
		var m SDKStatusMessage
		if err := json.Unmarshal(data, &m); err != nil {
			return nil, fmt.Errorf("parse status: %w", err)
		}
		return &m, nil
	case "api_retry":
		var m SDKAPIRetryMessage
		if err := json.Unmarshal(data, &m); err != nil {
			return nil, fmt.Errorf("parse api retry: %w", err)
		}
		return &m, nil
	case "compact_boundary":
		var m SDKCompactBoundaryMessage
		if err := json.Unmarshal(data, &m); err != nil {
			return nil, fmt.Errorf("parse compact boundary: %w", err)
		}
		return &m, nil
	case "local_command_output":
		var m SDKLocalCommandOutputMessage
		if err := json.Unmarshal(data, &m); err != nil {
			return nil, fmt.Errorf("parse local command output: %w", err)
		}
		return &m, nil
	case "hook_started":
		var m SDKHookStartedMessage
		if err := json.Unmarshal(data, &m); err != nil {
			return nil, fmt.Errorf("parse hook started: %w", err)
		}
		return &m, nil
	case "hook_progress":
		var m SDKHookProgressMessage
		if err := json.Unmarshal(data, &m); err != nil {
			return nil, fmt.Errorf("parse hook progress: %w", err)
		}
		return &m, nil
	case "hook_response":
		var m SDKHookResponseMessage
		if err := json.Unmarshal(data, &m); err != nil {
			return nil, fmt.Errorf("parse hook response: %w", err)
		}
		return &m, nil
	case "task_notification":
		var m SDKTaskNotificationMessage
		if err := json.Unmarshal(data, &m); err != nil {
			return nil, fmt.Errorf("parse task notification: %w", err)
		}
		return &m, nil
	case "task_started":
		var m SDKTaskStartedMessage
		if err := json.Unmarshal(data, &m); err != nil {
			return nil, fmt.Errorf("parse task started: %w", err)
		}
		return &m, nil
	case "task_progress":
		var m SDKTaskProgressMessage
		if err := json.Unmarshal(data, &m); err != nil {
			return nil, fmt.Errorf("parse task progress: %w", err)
		}
		return &m, nil
	case "files_persisted":
		var m SDKFilesPersistedEvent
		if err := json.Unmarshal(data, &m); err != nil {
			return nil, fmt.Errorf("parse files persisted: %w", err)
		}
		return &m, nil
	case "elicitation_complete":
		var m SDKElicitationCompleteMessage
		if err := json.Unmarshal(data, &m); err != nil {
			return nil, fmt.Errorf("parse elicitation complete: %w", err)
		}
		return &m, nil
	default:
		return &SDKRawMessage{RawType: "system", RawSubtype: subtype, Raw: json.RawMessage(data)}, nil
	}
}
