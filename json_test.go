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

func TestParseSDKMessage_InvalidJSON(t *testing.T) {
	_, err := ParseSDKMessage([]byte(`{not valid json`))
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}

func TestParseSDKMessage_EmptyType(t *testing.T) {
	// Empty type should fall through to SDKRawMessage
	msg, err := ParseSDKMessage([]byte(`{"type":""}`))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := msg.(*SDKRawMessage); !ok {
		t.Errorf("expected *SDKRawMessage for empty type, got %T", msg)
	}
}

func TestParseSDKMessage_ResultErrorSubtypes(t *testing.T) {
	subtypes := []string{
		"error_during_execution",
		"error_max_turns",
		"error_max_budget_usd",
		"error_max_structured_output_retries",
	}
	for _, subtype := range subtypes {
		t.Run(subtype, func(t *testing.T) {
			raw := fmt.Sprintf(`{"type":"result","subtype":"%s","duration_ms":0,"duration_api_ms":0,"is_error":true,"num_turns":0,"total_cost_usd":0,"usage":{"input_tokens":0,"output_tokens":0,"cache_creation_input_tokens":0,"cache_read_input_tokens":0},"modelUsage":{},"permission_denials":[],"errors":["fail"],"uuid":"a","session_id":"s"}`, subtype)
			msg, err := ParseSDKMessage([]byte(raw))
			if err != nil {
				t.Fatalf("ParseSDKMessage: %v", err)
			}
			errMsg, ok := msg.(*SDKResultError)
			if !ok {
				t.Fatalf("expected *SDKResultError, got %T", msg)
			}
			if errMsg.Subtype != subtype {
				t.Errorf("Subtype = %q, want %q", errMsg.Subtype, subtype)
			}
		})
	}
}

func TestParseSDKMessage_AllTypes(t *testing.T) {
	tests := []struct {
		name        string
		json        string
		wantType    string
		wantMsgType string
	}{
		{"assistant", `{"type":"assistant","message":{},"uuid":"a","session_id":"s"}`, "*claudeagent.SDKAssistantMessage", "assistant"},
		{"user", `{"type":"user","message":{},"session_id":"s"}`, "*claudeagent.SDKUserMessage", "user"},
		{"user_replay", `{"type":"user","message":{},"isReplay":true,"uuid":"a","session_id":"s"}`, "*claudeagent.SDKUserMessageReplay", "user"},
		{"result_success", `{"type":"result","subtype":"success","duration_ms":0,"duration_api_ms":0,"is_error":false,"num_turns":0,"result":"","total_cost_usd":0,"usage":{"input_tokens":0,"output_tokens":0,"cache_creation_input_tokens":0,"cache_read_input_tokens":0},"modelUsage":{},"permission_denials":[],"uuid":"a","session_id":"s"}`, "*claudeagent.SDKResultSuccess", "result"},
		{"result_error", `{"type":"result","subtype":"error_during_execution","duration_ms":0,"duration_api_ms":0,"is_error":true,"num_turns":0,"total_cost_usd":0,"usage":{"input_tokens":0,"output_tokens":0,"cache_creation_input_tokens":0,"cache_read_input_tokens":0},"modelUsage":{},"permission_denials":[],"errors":[],"uuid":"a","session_id":"s"}`, "*claudeagent.SDKResultError", "result"},
		{"system_init", `{"type":"system","subtype":"init","agents":[],"apiKeySource":"user","claude_code_version":"1.0","cwd":"/","tools":[],"mcp_servers":[],"model":"m","permissionMode":"default","slash_commands":[],"output_style":"normal","skills":[],"plugins":[],"uuid":"a","session_id":"s"}`, "*claudeagent.SDKSystemMessage", "system"},
		{"system_status", `{"type":"system","subtype":"status","status":null,"uuid":"a","session_id":"s"}`, "*claudeagent.SDKStatusMessage", "system"},
		{"system_api_retry", `{"type":"system","subtype":"api_retry","attempt":1,"max_retries":3,"retry_delay_ms":1000,"uuid":"a","session_id":"s"}`, "*claudeagent.SDKAPIRetryMessage", "system"},
		{"system_compact_boundary", `{"type":"system","subtype":"compact_boundary","compact_metadata":{"trigger":"auto","pre_tokens":100},"uuid":"a","session_id":"s"}`, "*claudeagent.SDKCompactBoundaryMessage", "system"},
		{"system_local_command_output", `{"type":"system","subtype":"local_command_output","content":"output","uuid":"a","session_id":"s"}`, "*claudeagent.SDKLocalCommandOutputMessage", "system"},
		{"system_hook_started", `{"type":"system","subtype":"hook_started","hook_id":"h1","hook_name":"test","hook_event":"PreToolUse","uuid":"a","session_id":"s"}`, "*claudeagent.SDKHookStartedMessage", "system"},
		{"system_hook_progress", `{"type":"system","subtype":"hook_progress","hook_id":"h1","hook_name":"test","hook_event":"PreToolUse","stdout":"","stderr":"","output":"","uuid":"a","session_id":"s"}`, "*claudeagent.SDKHookProgressMessage", "system"},
		{"system_hook_response", `{"type":"system","subtype":"hook_response","hook_id":"h1","hook_name":"test","hook_event":"PreToolUse","output":"","stdout":"","stderr":"","outcome":"success","uuid":"a","session_id":"s"}`, "*claudeagent.SDKHookResponseMessage", "system"},
		{"system_task_notification", `{"type":"system","subtype":"task_notification","task_id":"t1","status":"completed","output_file":"out","summary":"done","uuid":"a","session_id":"s"}`, "*claudeagent.SDKTaskNotificationMessage", "system"},
		{"system_task_started", `{"type":"system","subtype":"task_started","task_id":"t1","description":"test","uuid":"a","session_id":"s"}`, "*claudeagent.SDKTaskStartedMessage", "system"},
		{"system_task_progress", `{"type":"system","subtype":"task_progress","task_id":"t1","description":"test","usage":{"total_tokens":0,"tool_uses":0,"duration_ms":0},"uuid":"a","session_id":"s"}`, "*claudeagent.SDKTaskProgressMessage", "system"},
		{"system_files_persisted", `{"type":"system","subtype":"files_persisted","files":[],"failed":[],"processed_at":"2024-01-01","uuid":"a","session_id":"s"}`, "*claudeagent.SDKFilesPersistedEvent", "system"},
		{"system_elicitation_complete", `{"type":"system","subtype":"elicitation_complete","mcp_server_name":"test","elicitation_id":"e1","uuid":"a","session_id":"s"}`, "*claudeagent.SDKElicitationCompleteMessage", "system"},
		{"stream_event", `{"type":"stream_event","event":{},"uuid":"a","session_id":"s"}`, "*claudeagent.SDKPartialAssistantMessage", "stream_event"},
		{"tool_progress", `{"type":"tool_progress","tool_use_id":"t1","tool_name":"Bash","elapsed_time_seconds":1.0,"uuid":"a","session_id":"s"}`, "*claudeagent.SDKToolProgressMessage", "tool_progress"},
		{"tool_use_summary", `{"type":"tool_use_summary","summary":"done","preceding_tool_use_ids":[],"uuid":"a","session_id":"s"}`, "*claudeagent.SDKToolUseSummaryMessage", "tool_use_summary"},
		{"auth_status", `{"type":"auth_status","isAuthenticating":false,"output":[],"uuid":"a","session_id":"s"}`, "*claudeagent.SDKAuthStatusMessage", "auth_status"},
		{"rate_limit_event", `{"type":"rate_limit_event","rate_limit_info":{"status":"allowed"},"uuid":"a","session_id":"s"}`, "*claudeagent.SDKRateLimitEvent", "rate_limit_event"},
		{"prompt_suggestion", `{"type":"prompt_suggestion","suggestion":"try this","uuid":"a","session_id":"s"}`, "*claudeagent.SDKPromptSuggestionMessage", "prompt_suggestion"},
		{"unknown_type", `{"type":"brand_new","data":123}`, "*claudeagent.SDKRawMessage", "brand_new"},
		{"unknown_system_subtype", `{"type":"system","subtype":"brand_new","uuid":"a","session_id":"s"}`, "*claudeagent.SDKRawMessage", "system"},
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
			// Verify MessageType() returns the expected wire type
			if msg.MessageType() != tt.wantMsgType {
				t.Errorf("MessageType() = %q, want %q", msg.MessageType(), tt.wantMsgType)
			}
		})
	}
}

func TestParseSDKMessage_SDKRawMessage_PreservesRawJSON(t *testing.T) {
	raw := `{"type":"new_type","subtype":"sub","data":{"nested":true}}`
	msg, err := ParseSDKMessage([]byte(raw))
	if err != nil {
		t.Fatalf("ParseSDKMessage: %v", err)
	}
	rawMsg, ok := msg.(*SDKRawMessage)
	if !ok {
		t.Fatalf("expected *SDKRawMessage, got %T", msg)
	}
	if rawMsg.RawType != "new_type" {
		t.Errorf("RawType = %q", rawMsg.RawType)
	}
	if rawMsg.RawSubtype != "sub" {
		t.Errorf("RawSubtype = %q", rawMsg.RawSubtype)
	}
	if rawMsg.Raw == nil {
		t.Error("Raw should contain the original JSON data")
	}
	if rawMsg.MessageType() != "new_type" {
		t.Errorf("MessageType() = %q, want %q", rawMsg.MessageType(), "new_type")
	}
}
