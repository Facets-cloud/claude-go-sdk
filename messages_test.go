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
	// Verify all message types satisfy the SDKMessage interface at compile time
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

// TestSDKMessage_MarkerAndType exercises sdkMessage() and MessageType() on all concrete types.
func TestSDKMessage_MarkerAndType(t *testing.T) {
	tests := []struct {
		name     string
		msg      SDKMessage
		wantType string
	}{
		{"SDKAssistantMessage", &SDKAssistantMessage{}, "assistant"},
		{"SDKUserMessage", &SDKUserMessage{}, "user"},
		{"SDKUserMessageReplay", &SDKUserMessageReplay{}, "user"},
		{"SDKResultSuccess", &SDKResultSuccess{}, "result"},
		{"SDKResultError", &SDKResultError{}, "result"},
		{"SDKSystemMessage", &SDKSystemMessage{}, "system"},
		{"SDKStatusMessage", &SDKStatusMessage{}, "system"},
		{"SDKAPIRetryMessage", &SDKAPIRetryMessage{}, "system"},
		{"SDKCompactBoundaryMessage", &SDKCompactBoundaryMessage{}, "system"},
		{"SDKLocalCommandOutputMessage", &SDKLocalCommandOutputMessage{}, "system"},
		{"SDKHookStartedMessage", &SDKHookStartedMessage{}, "system"},
		{"SDKHookProgressMessage", &SDKHookProgressMessage{}, "system"},
		{"SDKHookResponseMessage", &SDKHookResponseMessage{}, "system"},
		{"SDKTaskNotificationMessage", &SDKTaskNotificationMessage{}, "system"},
		{"SDKTaskStartedMessage", &SDKTaskStartedMessage{}, "system"},
		{"SDKTaskProgressMessage", &SDKTaskProgressMessage{}, "system"},
		{"SDKFilesPersistedEvent", &SDKFilesPersistedEvent{}, "system"},
		{"SDKElicitationCompleteMessage", &SDKElicitationCompleteMessage{}, "system"},
		{"SDKPartialAssistantMessage", &SDKPartialAssistantMessage{}, "stream_event"},
		{"SDKToolProgressMessage", &SDKToolProgressMessage{}, "tool_progress"},
		{"SDKToolUseSummaryMessage", &SDKToolUseSummaryMessage{}, "tool_use_summary"},
		{"SDKAuthStatusMessage", &SDKAuthStatusMessage{}, "auth_status"},
		{"SDKRateLimitEvent", &SDKRateLimitEvent{}, "rate_limit_event"},
		{"SDKPromptSuggestionMessage", &SDKPromptSuggestionMessage{}, "prompt_suggestion"},
		{"SDKRawMessage", &SDKRawMessage{RawType: "custom"}, "custom"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Exercise sdkMessage() marker (must not panic)
			tt.msg.sdkMessage()
			if got := tt.msg.MessageType(); got != tt.wantType {
				t.Errorf("MessageType() = %q, want %q", got, tt.wantType)
			}
		})
	}
}

func TestSDKResultError_JSON(t *testing.T) {
	raw := `{"type":"result","subtype":"error_during_execution","duration_ms":500,"duration_api_ms":300,"is_error":true,"num_turns":1,"total_cost_usd":0.01,"usage":{"input_tokens":10,"output_tokens":5,"cache_creation_input_tokens":0,"cache_read_input_tokens":0},"modelUsage":{},"permission_denials":[],"errors":["something broke","another error"],"uuid":"abc","session_id":"s1"}`
	var msg SDKResultError
	if err := json.Unmarshal([]byte(raw), &msg); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if msg.Subtype != "error_during_execution" {
		t.Errorf("Subtype = %q", msg.Subtype)
	}
	if !msg.IsError {
		t.Error("IsError should be true")
	}
	if len(msg.Errors) != 2 {
		t.Errorf("Errors = %v", msg.Errors)
	}
}

func TestSDKUserMessageReplay_JSON(t *testing.T) {
	raw := `{"type":"user","message":{"role":"user","content":"replay"},"isReplay":true,"uuid":"abc","session_id":"s1"}`
	var msg SDKUserMessageReplay
	if err := json.Unmarshal([]byte(raw), &msg); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if !msg.IsReplay {
		t.Error("IsReplay should be true")
	}
}

func TestSDKStatusMessage_JSON(t *testing.T) {
	raw := `{"type":"system","subtype":"status","status":null,"uuid":"a","session_id":"s"}`
	var msg SDKStatusMessage
	if err := json.Unmarshal([]byte(raw), &msg); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if msg.UUID != "a" {
		t.Errorf("UUID = %q", msg.UUID)
	}
}

func TestSDKAPIRetryMessage_JSON(t *testing.T) {
	raw := `{"type":"system","subtype":"api_retry","attempt":2,"max_retries":5,"retry_delay_ms":2000,"uuid":"a","session_id":"s"}`
	var msg SDKAPIRetryMessage
	if err := json.Unmarshal([]byte(raw), &msg); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if msg.Attempt != 2 {
		t.Errorf("Attempt = %d", msg.Attempt)
	}
	if msg.MaxRetries != 5 {
		t.Errorf("MaxRetries = %d", msg.MaxRetries)
	}
}

func TestSDKCompactBoundaryMessage_JSON(t *testing.T) {
	raw := `{"type":"system","subtype":"compact_boundary","compact_metadata":{"trigger":"auto","pre_tokens":500},"uuid":"a","session_id":"s"}`
	var msg SDKCompactBoundaryMessage
	if err := json.Unmarshal([]byte(raw), &msg); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if msg.CompactMetadata.Trigger != "auto" {
		t.Errorf("Trigger = %q", msg.CompactMetadata.Trigger)
	}
	if msg.CompactMetadata.PreTokens != 500 {
		t.Errorf("PreTokens = %d", msg.CompactMetadata.PreTokens)
	}
}

func TestSDKToolProgressMessage_JSON(t *testing.T) {
	raw := `{"type":"tool_progress","tool_use_id":"tu_1","tool_name":"Bash","elapsed_time_seconds":5.5,"uuid":"a","session_id":"s"}`
	var msg SDKToolProgressMessage
	if err := json.Unmarshal([]byte(raw), &msg); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if msg.ToolName != "Bash" {
		t.Errorf("ToolName = %q", msg.ToolName)
	}
	if msg.ElapsedTimeSeconds != 5.5 {
		t.Errorf("ElapsedTimeSeconds = %f", msg.ElapsedTimeSeconds)
	}
}

func TestSDKToolUseSummaryMessage_JSON(t *testing.T) {
	raw := `{"type":"tool_use_summary","summary":"ran 3 commands","preceding_tool_use_ids":["t1","t2"],"uuid":"a","session_id":"s"}`
	var msg SDKToolUseSummaryMessage
	if err := json.Unmarshal([]byte(raw), &msg); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if msg.Summary != "ran 3 commands" {
		t.Errorf("Summary = %q", msg.Summary)
	}
	if len(msg.PrecedingToolUseIDs) != 2 {
		t.Errorf("PrecedingToolUseIDs count = %d", len(msg.PrecedingToolUseIDs))
	}
}

func TestSDKAuthStatusMessage_JSON(t *testing.T) {
	raw := `{"type":"auth_status","isAuthenticating":true,"output":["please authenticate"],"uuid":"a","session_id":"s"}`
	var msg SDKAuthStatusMessage
	if err := json.Unmarshal([]byte(raw), &msg); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if !msg.IsAuthenticating {
		t.Error("IsAuthenticating should be true")
	}
	if len(msg.Output) != 1 {
		t.Errorf("Output = %v", msg.Output)
	}
}

func TestSDKTaskNotificationMessage_JSON(t *testing.T) {
	raw := `{"type":"system","subtype":"task_notification","task_id":"t1","status":"completed","output_file":"out.txt","summary":"done","uuid":"a","session_id":"s"}`
	var msg SDKTaskNotificationMessage
	if err := json.Unmarshal([]byte(raw), &msg); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if msg.TaskID != "t1" {
		t.Errorf("TaskID = %q", msg.TaskID)
	}
	if msg.Status != "completed" {
		t.Errorf("Status = %q", msg.Status)
	}
}

func TestSDKTaskStartedMessage_JSON(t *testing.T) {
	raw := `{"type":"system","subtype":"task_started","task_id":"t1","description":"building","uuid":"a","session_id":"s"}`
	var msg SDKTaskStartedMessage
	if err := json.Unmarshal([]byte(raw), &msg); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if msg.Description != "building" {
		t.Errorf("Description = %q", msg.Description)
	}
}

func TestSDKTaskProgressMessage_JSON(t *testing.T) {
	raw := `{"type":"system","subtype":"task_progress","task_id":"t1","description":"50%","usage":{"total_tokens":100,"tool_uses":5,"duration_ms":1000},"uuid":"a","session_id":"s"}`
	var msg SDKTaskProgressMessage
	if err := json.Unmarshal([]byte(raw), &msg); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if msg.Usage.TotalTokens != 100 {
		t.Errorf("TotalTokens = %d", msg.Usage.TotalTokens)
	}
}

func TestSDKFilesPersistedEvent_JSON(t *testing.T) {
	raw := `{"type":"system","subtype":"files_persisted","files":[{"path":"/tmp/a.txt","hash":"abc"}],"failed":[{"path":"/tmp/b.txt","error":"permission denied"}],"processed_at":"2024-01-01T00:00:00Z","uuid":"a","session_id":"s"}`
	var msg SDKFilesPersistedEvent
	if err := json.Unmarshal([]byte(raw), &msg); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if len(msg.Files) != 1 {
		t.Errorf("Files count = %d", len(msg.Files))
	}
	if len(msg.Failed) != 1 {
		t.Errorf("Failed count = %d", len(msg.Failed))
	}
}

func TestSDKRateLimitEvent_JSON(t *testing.T) {
	raw := `{"type":"rate_limit_event","rate_limit_info":{"status":"allowed"},"uuid":"a","session_id":"s"}`
	var msg SDKRateLimitEvent
	if err := json.Unmarshal([]byte(raw), &msg); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if msg.RateLimitInfo.Status != "allowed" {
		t.Errorf("Status = %q", msg.RateLimitInfo.Status)
	}
}

func TestSDKElicitationCompleteMessage_JSON(t *testing.T) {
	raw := `{"type":"system","subtype":"elicitation_complete","mcp_server_name":"test","elicitation_id":"e1","uuid":"a","session_id":"s"}`
	var msg SDKElicitationCompleteMessage
	if err := json.Unmarshal([]byte(raw), &msg); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if msg.McpServerName != "test" {
		t.Errorf("McpServerName = %q", msg.McpServerName)
	}
}

func TestSDKPromptSuggestionMessage_JSON(t *testing.T) {
	raw := `{"type":"prompt_suggestion","suggestion":"try asking about X","uuid":"a","session_id":"s"}`
	var msg SDKPromptSuggestionMessage
	if err := json.Unmarshal([]byte(raw), &msg); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if msg.Suggestion != "try asking about X" {
		t.Errorf("Suggestion = %q", msg.Suggestion)
	}
}

func TestSDKHookStartedMessage_JSON(t *testing.T) {
	raw := `{"type":"system","subtype":"hook_started","hook_id":"h1","hook_name":"pre-check","hook_event":"PreToolUse","uuid":"a","session_id":"s"}`
	var msg SDKHookStartedMessage
	if err := json.Unmarshal([]byte(raw), &msg); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if msg.HookName != "pre-check" {
		t.Errorf("HookName = %q", msg.HookName)
	}
}

func TestSDKHookProgressMessage_JSON(t *testing.T) {
	raw := `{"type":"system","subtype":"hook_progress","hook_id":"h1","hook_name":"test","hook_event":"PreToolUse","stdout":"out","stderr":"err","output":"combined","uuid":"a","session_id":"s"}`
	var msg SDKHookProgressMessage
	if err := json.Unmarshal([]byte(raw), &msg); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if msg.Stdout != "out" {
		t.Errorf("Stdout = %q", msg.Stdout)
	}
	if msg.Stderr != "err" {
		t.Errorf("Stderr = %q", msg.Stderr)
	}
}

func TestSDKHookResponseMessage_JSON(t *testing.T) {
	raw := `{"type":"system","subtype":"hook_response","hook_id":"h1","hook_name":"test","hook_event":"PreToolUse","output":"done","stdout":"","stderr":"","outcome":"success","uuid":"a","session_id":"s"}`
	var msg SDKHookResponseMessage
	if err := json.Unmarshal([]byte(raw), &msg); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if msg.Outcome != "success" {
		t.Errorf("Outcome = %q", msg.Outcome)
	}
}

func TestSDKLocalCommandOutputMessage_JSON(t *testing.T) {
	raw := `{"type":"system","subtype":"local_command_output","content":"command output here","uuid":"a","session_id":"s"}`
	var msg SDKLocalCommandOutputMessage
	if err := json.Unmarshal([]byte(raw), &msg); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if msg.Content != "command output here" {
		t.Errorf("Content = %q", msg.Content)
	}
}

func TestNonNullableUsage_JSON(t *testing.T) {
	raw := `{"input_tokens":100,"output_tokens":50,"cache_creation_input_tokens":10,"cache_read_input_tokens":5,"server_tool_use":{"web_search_requests":3},"service_tier":"standard","cache_creation":{"num_segments":2}}`
	var u NonNullableUsage
	if err := json.Unmarshal([]byte(raw), &u); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if u.InputTokens != 100 {
		t.Errorf("InputTokens = %d", u.InputTokens)
	}
	if u.ServerToolUse == nil {
		t.Error("ServerToolUse should not be nil")
	}
}

func TestNonNullableUsage_NullOptionalFields(t *testing.T) {
	raw := `{"input_tokens":10,"output_tokens":5,"cache_creation_input_tokens":0,"cache_read_input_tokens":0,"server_tool_use":null,"service_tier":null,"cache_creation":null}`
	var u NonNullableUsage
	if err := json.Unmarshal([]byte(raw), &u); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if u.ServerToolUse != nil {
		t.Error("ServerToolUse should be nil")
	}
	if u.ServiceTier != nil {
		t.Error("ServiceTier should be nil")
	}
}

func TestSDKPartialAssistantMessage_JSON(t *testing.T) {
	raw := `{"type":"stream_event","event":{"type":"content_block_delta"},"uuid":"a","session_id":"s"}`
	var msg SDKPartialAssistantMessage
	if err := json.Unmarshal([]byte(raw), &msg); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if msg.Event == nil {
		t.Error("Event should not be nil")
	}
}

func TestSDKPermissionDenial_JSON(t *testing.T) {
	raw := `{"tool_name":"Bash","tool_use_id":"tu1","tool_input":{"command":"ls"}}`
	var d SDKPermissionDenial
	if err := json.Unmarshal([]byte(raw), &d); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if d.ToolName != "Bash" {
		t.Errorf("ToolName = %q", d.ToolName)
	}
	if d.ToolUseID != "tu1" {
		t.Errorf("ToolUseID = %q", d.ToolUseID)
	}
}

func TestSDKResultSuccess_RoundTrip(t *testing.T) {
	original := SDKResultSuccess{
		DurationMs:    1000,
		DurationAPIMs: 800,
		IsError:       false,
		NumTurns:      3,
		Result:        "done",
		TotalCostUSD:  0.05,
		UUID:          "abc",
		SessionID:     "s1",
	}
	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	var got SDKResultSuccess
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if got.Result != "done" {
		t.Errorf("Result = %q", got.Result)
	}
	if got.NumTurns != 3 {
		t.Errorf("NumTurns = %d", got.NumTurns)
	}
}

func TestIsResultMessage(t *testing.T) {
	tests := []struct {
		name string
		msg  SDKMessage
		want bool
	}{
		{"SDKResultSuccess", &SDKResultSuccess{}, true},
		{"SDKResultError", &SDKResultError{}, true},
		{"SDKAssistantMessage", &SDKAssistantMessage{}, false},
		{"SDKUserMessage", &SDKUserMessage{}, false},
		{"SDKSystemMessage", &SDKSystemMessage{}, false},
		{"SDKRawMessage", &SDKRawMessage{}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsResultMessage(tt.msg); got != tt.want {
				t.Errorf("IsResultMessage() = %v, want %v", got, tt.want)
			}
		})
	}
}
