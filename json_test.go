package claudeagent

import (
	"fmt"
	"testing"
)

func TestRawJSONMessage_Unmarshal(t *testing.T) {
	raw := `{"type":"assistant","message":{},"parent_tool_use_id":null,"uuid":"abc","session_id":"s1"}`
	msg, err := ParseSDKMessage([]byte(raw))
	if err != nil {
		t.Fatalf("ParseSDKMessage: %v", err)
	}
	if _, ok := msg.(*SDKAssistantMessage); !ok {
		t.Errorf("expected *SDKAssistantMessage, got %T", msg)
	}
}

func TestRawJSONMessage_UnmarshalResult(t *testing.T) {
	raw := `{"type":"result","subtype":"success","duration_ms":100,"duration_api_ms":50,"is_error":false,"num_turns":1,"result":"done","stop_reason":null,"total_cost_usd":0.01,"usage":{"input_tokens":10,"output_tokens":20,"cache_creation_input_tokens":0,"cache_read_input_tokens":0,"server_tool_use":null,"service_tier":null,"cache_creation":null},"modelUsage":{},"permission_denials":[],"uuid":"abc","session_id":"s1"}`
	msg, err := ParseSDKMessage([]byte(raw))
	if err != nil {
		t.Fatalf("ParseSDKMessage: %v", err)
	}
	if result, ok := msg.(*SDKResultSuccess); !ok {
		t.Errorf("expected *SDKResultSuccess, got %T", msg)
	} else if result.Result != "done" {
		t.Errorf("result = %q, want %q", result.Result, "done")
	}
}

func TestRawJSONMessage_UnmarshalSystem(t *testing.T) {
	raw := `{"type":"system","subtype":"init","agents":[],"apiKeySource":"user","claude_code_version":"2.1.81","cwd":"/tmp","tools":["Bash"],"mcp_servers":[],"model":"claude-sonnet-4-6","permissionMode":"default","slash_commands":[],"output_style":"normal","skills":[],"plugins":[],"uuid":"abc","session_id":"s1"}`
	msg, err := ParseSDKMessage([]byte(raw))
	if err != nil {
		t.Fatalf("ParseSDKMessage: %v", err)
	}
	if sys, ok := msg.(*SDKSystemMessage); !ok {
		t.Errorf("expected *SDKSystemMessage, got %T", msg)
	} else if sys.ClaudeCodeVersion != "2.1.81" {
		t.Errorf("version = %q, want %q", sys.ClaudeCodeVersion, "2.1.81")
	}
}

func TestParseSDKMessage_UnknownType(t *testing.T) {
	raw := `{"type":"future_type","some_field":"value"}`
	msg, err := ParseSDKMessage([]byte(raw))
	if err != nil {
		t.Fatalf("ParseSDKMessage should not error on unknown types: %v", err)
	}
	rawMsg, ok := msg.(*SDKRawMessage)
	if !ok {
		t.Fatalf("expected *SDKRawMessage, got %T", msg)
	}
	if rawMsg.RawType != "future_type" {
		t.Errorf("RawType = %q, want %q", rawMsg.RawType, "future_type")
	}
}

func TestParseSDKMessage_UnknownSystemSubtype(t *testing.T) {
	raw := `{"type":"system","subtype":"future_subtype","uuid":"abc","session_id":"s1"}`
	msg, err := ParseSDKMessage([]byte(raw))
	if err != nil {
		t.Fatalf("ParseSDKMessage should not error on unknown system subtypes: %v", err)
	}
	rawMsg, ok := msg.(*SDKRawMessage)
	if !ok {
		t.Fatalf("expected *SDKRawMessage, got %T", msg)
	}
	if rawMsg.RawType != "system" {
		t.Errorf("RawType = %q, want %q", rawMsg.RawType, "system")
	}
	if rawMsg.RawSubtype != "future_subtype" {
		t.Errorf("RawSubtype = %q, want %q", rawMsg.RawSubtype, "future_subtype")
	}
}

func TestParseSDKMessage_AllTypes(t *testing.T) {
	tests := []struct {
		name     string
		json     string
		wantType string
	}{
		{"assistant", `{"type":"assistant","message":{},"uuid":"a","session_id":"s"}`, "*claudeagent.SDKAssistantMessage"},
		{"user", `{"type":"user","message":{},"session_id":"s"}`, "*claudeagent.SDKUserMessage"},
		{"user_replay", `{"type":"user","message":{},"isReplay":true,"uuid":"a","session_id":"s"}`, "*claudeagent.SDKUserMessageReplay"},
		{"result_success", `{"type":"result","subtype":"success","duration_ms":0,"duration_api_ms":0,"is_error":false,"num_turns":0,"result":"","total_cost_usd":0,"usage":{"input_tokens":0,"output_tokens":0,"cache_creation_input_tokens":0,"cache_read_input_tokens":0},"modelUsage":{},"permission_denials":[],"uuid":"a","session_id":"s"}`, "*claudeagent.SDKResultSuccess"},
		{"result_error", `{"type":"result","subtype":"error_during_execution","duration_ms":0,"duration_api_ms":0,"is_error":true,"num_turns":0,"total_cost_usd":0,"usage":{"input_tokens":0,"output_tokens":0,"cache_creation_input_tokens":0,"cache_read_input_tokens":0},"modelUsage":{},"permission_denials":[],"errors":[],"uuid":"a","session_id":"s"}`, "*claudeagent.SDKResultError"},
		{"system_init", `{"type":"system","subtype":"init","agents":[],"apiKeySource":"user","claude_code_version":"1.0","cwd":"/","tools":[],"mcp_servers":[],"model":"m","permissionMode":"default","slash_commands":[],"output_style":"normal","skills":[],"plugins":[],"uuid":"a","session_id":"s"}`, "*claudeagent.SDKSystemMessage"},
		{"system_status", `{"type":"system","subtype":"status","status":null,"uuid":"a","session_id":"s"}`, "*claudeagent.SDKStatusMessage"},
		{"system_api_retry", `{"type":"system","subtype":"api_retry","attempt":1,"max_retries":3,"retry_delay_ms":1000,"uuid":"a","session_id":"s"}`, "*claudeagent.SDKAPIRetryMessage"},
		{"system_compact_boundary", `{"type":"system","subtype":"compact_boundary","compact_metadata":{"trigger":"auto","pre_tokens":100},"uuid":"a","session_id":"s"}`, "*claudeagent.SDKCompactBoundaryMessage"},
		{"system_local_command_output", `{"type":"system","subtype":"local_command_output","content":"output","uuid":"a","session_id":"s"}`, "*claudeagent.SDKLocalCommandOutputMessage"},
		{"system_hook_started", `{"type":"system","subtype":"hook_started","hook_id":"h1","hook_name":"test","hook_event":"PreToolUse","uuid":"a","session_id":"s"}`, "*claudeagent.SDKHookStartedMessage"},
		{"system_hook_progress", `{"type":"system","subtype":"hook_progress","hook_id":"h1","hook_name":"test","hook_event":"PreToolUse","stdout":"","stderr":"","output":"","uuid":"a","session_id":"s"}`, "*claudeagent.SDKHookProgressMessage"},
		{"system_hook_response", `{"type":"system","subtype":"hook_response","hook_id":"h1","hook_name":"test","hook_event":"PreToolUse","output":"","stdout":"","stderr":"","outcome":"success","uuid":"a","session_id":"s"}`, "*claudeagent.SDKHookResponseMessage"},
		{"system_task_notification", `{"type":"system","subtype":"task_notification","task_id":"t1","status":"completed","output_file":"out","summary":"done","uuid":"a","session_id":"s"}`, "*claudeagent.SDKTaskNotificationMessage"},
		{"system_task_started", `{"type":"system","subtype":"task_started","task_id":"t1","description":"test","uuid":"a","session_id":"s"}`, "*claudeagent.SDKTaskStartedMessage"},
		{"system_task_progress", `{"type":"system","subtype":"task_progress","task_id":"t1","description":"test","usage":{"total_tokens":0,"tool_uses":0,"duration_ms":0},"uuid":"a","session_id":"s"}`, "*claudeagent.SDKTaskProgressMessage"},
		{"system_files_persisted", `{"type":"system","subtype":"files_persisted","files":[],"failed":[],"processed_at":"2024-01-01","uuid":"a","session_id":"s"}`, "*claudeagent.SDKFilesPersistedEvent"},
		{"system_elicitation_complete", `{"type":"system","subtype":"elicitation_complete","mcp_server_name":"test","elicitation_id":"e1","uuid":"a","session_id":"s"}`, "*claudeagent.SDKElicitationCompleteMessage"},
		{"stream_event", `{"type":"stream_event","event":{},"uuid":"a","session_id":"s"}`, "*claudeagent.SDKPartialAssistantMessage"},
		{"tool_progress", `{"type":"tool_progress","tool_use_id":"t1","tool_name":"Bash","elapsed_time_seconds":1.0,"uuid":"a","session_id":"s"}`, "*claudeagent.SDKToolProgressMessage"},
		{"tool_use_summary", `{"type":"tool_use_summary","summary":"done","preceding_tool_use_ids":[],"uuid":"a","session_id":"s"}`, "*claudeagent.SDKToolUseSummaryMessage"},
		{"auth_status", `{"type":"auth_status","isAuthenticating":false,"output":[],"uuid":"a","session_id":"s"}`, "*claudeagent.SDKAuthStatusMessage"},
		{"rate_limit_event", `{"type":"rate_limit_event","rate_limit_info":{"status":"allowed"},"uuid":"a","session_id":"s"}`, "*claudeagent.SDKRateLimitEvent"},
		{"prompt_suggestion", `{"type":"prompt_suggestion","suggestion":"try this","uuid":"a","session_id":"s"}`, "*claudeagent.SDKPromptSuggestionMessage"},
		{"unknown_type", `{"type":"brand_new","data":123}`, "*claudeagent.SDKRawMessage"},
		{"unknown_system_subtype", `{"type":"system","subtype":"brand_new","uuid":"a","session_id":"s"}`, "*claudeagent.SDKRawMessage"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg, err := ParseSDKMessage([]byte(tt.json))
			if err != nil {
				t.Fatalf("ParseSDKMessage: %v", err)
			}
			gotType := fmt.Sprintf("%T", msg)
			if gotType != tt.wantType {
				t.Errorf("got type %s, want %s", gotType, tt.wantType)
			}
		})
	}
}
