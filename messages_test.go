package claudeagent

import (
	"encoding/json"
	"testing"
)

func TestSDKAssistantMessage_JSON(t *testing.T) {
	raw := `{"type":"assistant","message":{"id":"msg_01","type":"message","role":"assistant","content":[{"type":"text","text":"hello"}],"model":"claude-sonnet-4-6","stop_reason":"end_turn","usage":{"input_tokens":10,"output_tokens":5}},"parent_tool_use_id":null,"uuid":"550e8400-e29b-41d4-a716-446655440000","session_id":"sess1"}`
	var msg SDKAssistantMessage
	if err := json.Unmarshal([]byte(raw), &msg); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if msg.UUID != "550e8400-e29b-41d4-a716-446655440000" {
		t.Errorf("UUID = %q", msg.UUID)
	}
	if msg.SessionID != "sess1" {
		t.Errorf("SessionID = %q", msg.SessionID)
	}
}

func TestSDKSystemMessage_JSON(t *testing.T) {
	raw := `{"type":"system","subtype":"init","agents":["general"],"apiKeySource":"user","claude_code_version":"2.1.81","cwd":"/tmp","tools":["Bash","Read"],"mcp_servers":[],"model":"claude-sonnet-4-6","permissionMode":"default","slash_commands":[],"output_style":"normal","skills":[],"plugins":[],"uuid":"abc","session_id":"s1"}`
	var msg SDKSystemMessage
	if err := json.Unmarshal([]byte(raw), &msg); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if msg.ClaudeCodeVersion != "2.1.81" {
		t.Errorf("version = %q", msg.ClaudeCodeVersion)
	}
	if len(msg.Tools) != 2 {
		t.Errorf("tools = %v", msg.Tools)
	}
}

func TestSDKResultSuccess_JSON(t *testing.T) {
	raw := `{"type":"result","subtype":"success","duration_ms":1000,"duration_api_ms":800,"is_error":false,"num_turns":3,"result":"All done","stop_reason":"end_turn","total_cost_usd":0.05,"usage":{"input_tokens":100,"output_tokens":50,"cache_creation_input_tokens":0,"cache_read_input_tokens":0,"server_tool_use":null,"service_tier":null,"cache_creation":null},"modelUsage":{},"permission_denials":[],"uuid":"abc","session_id":"s1"}`
	var msg SDKResultSuccess
	if err := json.Unmarshal([]byte(raw), &msg); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if msg.Result != "All done" {
		t.Errorf("result = %q", msg.Result)
	}
	if msg.NumTurns != 3 {
		t.Errorf("num_turns = %d", msg.NumTurns)
	}
}

func TestSDKUserMessage_JSON(t *testing.T) {
	raw := `{"type":"user","message":{"role":"user","content":"hello"},"parent_tool_use_id":null,"session_id":"s1"}`
	var msg SDKUserMessage
	if err := json.Unmarshal([]byte(raw), &msg); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if msg.SessionID != "s1" {
		t.Errorf("session_id = %q", msg.SessionID)
	}
}

func TestSDKMessage_Interface(t *testing.T) {
	// Verify all message types satisfy the SDKMessage interface
	var _ SDKMessage = &SDKAssistantMessage{}
	var _ SDKMessage = &SDKUserMessage{}
	var _ SDKMessage = &SDKUserMessageReplay{}
	var _ SDKMessage = &SDKResultSuccess{}
	var _ SDKMessage = &SDKResultError{}
	var _ SDKMessage = &SDKSystemMessage{}
	var _ SDKMessage = &SDKPartialAssistantMessage{}
	var _ SDKMessage = &SDKCompactBoundaryMessage{}
	var _ SDKMessage = &SDKStatusMessage{}
	var _ SDKMessage = &SDKAPIRetryMessage{}
	var _ SDKMessage = &SDKLocalCommandOutputMessage{}
	var _ SDKMessage = &SDKHookStartedMessage{}
	var _ SDKMessage = &SDKHookProgressMessage{}
	var _ SDKMessage = &SDKHookResponseMessage{}
	var _ SDKMessage = &SDKToolProgressMessage{}
	var _ SDKMessage = &SDKAuthStatusMessage{}
	var _ SDKMessage = &SDKTaskNotificationMessage{}
	var _ SDKMessage = &SDKTaskStartedMessage{}
	var _ SDKMessage = &SDKTaskProgressMessage{}
	var _ SDKMessage = &SDKFilesPersistedEvent{}
	var _ SDKMessage = &SDKToolUseSummaryMessage{}
	var _ SDKMessage = &SDKRateLimitEvent{}
	var _ SDKMessage = &SDKElicitationCompleteMessage{}
	var _ SDKMessage = &SDKPromptSuggestionMessage{}
	var _ SDKMessage = &SDKRawMessage{}
}
