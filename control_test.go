package claudeagent

import (
	"context"
	"encoding/json"
	"testing"
)

func TestSDKControlRequest_JSON(t *testing.T) {
	raw := `{
		"type": "control_request",
		"request_id": "req-123",
		"request": {
			"subtype": "interrupt"
		}
	}`
	var req SDKControlRequest
	if err := json.Unmarshal([]byte(raw), &req); err != nil {
		t.Fatal(err)
	}
	if req.Type != "control_request" {
		t.Errorf("got type %q, want %q", req.Type, "control_request")
	}
	if req.RequestID != "req-123" {
		t.Errorf("got request_id %q, want %q", req.RequestID, "req-123")
	}
}

func TestSDKControlResponse_Success(t *testing.T) {
	raw := `{
		"type": "control_response",
		"response": {
			"subtype": "success",
			"request_id": "req-456",
			"response": {"key": "value"}
		}
	}`
	var resp SDKControlResponse
	if err := json.Unmarshal([]byte(raw), &resp); err != nil {
		t.Fatal(err)
	}
	if resp.Type != "control_response" {
		t.Errorf("got type %q, want %q", resp.Type, "control_response")
	}

	var inner ControlResponse
	if err := json.Unmarshal(resp.Response, &inner); err != nil {
		t.Fatal(err)
	}
	if inner.Subtype != "success" {
		t.Errorf("got subtype %q, want %q", inner.Subtype, "success")
	}
	if inner.RequestID != "req-456" {
		t.Errorf("got request_id %q, want %q", inner.RequestID, "req-456")
	}
}

func TestSDKControlResponse_Error(t *testing.T) {
	raw := `{
		"type": "control_response",
		"response": {
			"subtype": "error",
			"request_id": "req-789",
			"error": "something went wrong"
		}
	}`
	var resp SDKControlResponse
	if err := json.Unmarshal([]byte(raw), &resp); err != nil {
		t.Fatal(err)
	}

	var inner ControlErrorResponse
	if err := json.Unmarshal(resp.Response, &inner); err != nil {
		t.Fatal(err)
	}
	if inner.Subtype != "error" {
		t.Errorf("got subtype %q, want %q", inner.Subtype, "error")
	}
	if inner.Error != "something went wrong" {
		t.Errorf("got error %q, want %q", inner.Error, "something went wrong")
	}
}

func TestSDKControlCancelRequest_JSON(t *testing.T) {
	raw := `{
		"type": "control_cancel_request",
		"request_id": "req-cancel"
	}`
	var req SDKControlCancelRequest
	if err := json.Unmarshal([]byte(raw), &req); err != nil {
		t.Fatal(err)
	}
	if req.Type != "control_cancel_request" {
		t.Errorf("got type %q, want %q", req.Type, "control_cancel_request")
	}
	if req.RequestID != "req-cancel" {
		t.Errorf("got request_id %q, want %q", req.RequestID, "req-cancel")
	}
}

func TestSDKControlInitializeRequest_JSON(t *testing.T) {
	req := SDKControlInitializeRequest{
		Subtype:            "initialize",
		SystemPrompt:       strPtr("You are a helpful assistant"),
		AppendSystemPrompt: strPtr("Always be polite"),
		PromptSuggestions:  boolPtr(true),
	}
	data, err := json.Marshal(req)
	if err != nil {
		t.Fatal(err)
	}
	var got SDKControlInitializeRequest
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatal(err)
	}
	if got.Subtype != "initialize" {
		t.Errorf("got subtype %q, want %q", got.Subtype, "initialize")
	}
	if *got.SystemPrompt != "You are a helpful assistant" {
		t.Errorf("got system_prompt %q", *got.SystemPrompt)
	}
	if *got.PromptSuggestions != true {
		t.Error("expected prompt_suggestions true")
	}
}

func TestSDKControlPermissionRequest_JSON(t *testing.T) {
	raw := `{
		"subtype": "can_use_tool",
		"tool_name": "Bash",
		"input": {"command": "ls"},
		"tool_use_id": "tu-123"
	}`
	var req SDKControlPermissionRequest
	if err := json.Unmarshal([]byte(raw), &req); err != nil {
		t.Fatal(err)
	}
	if req.Subtype != "can_use_tool" {
		t.Errorf("got subtype %q, want %q", req.Subtype, "can_use_tool")
	}
	if req.ToolName != "Bash" {
		t.Errorf("got tool_name %q, want %q", req.ToolName, "Bash")
	}
	if req.ToolUseID != "tu-123" {
		t.Errorf("got tool_use_id %q, want %q", req.ToolUseID, "tu-123")
	}
}

func TestSDKControlElicitationRequest_JSON(t *testing.T) {
	raw := `{
		"subtype": "elicitation",
		"mcp_server_name": "my-server",
		"message": "Please enter your name",
		"mode": "form"
	}`
	var req SDKControlElicitationRequest
	if err := json.Unmarshal([]byte(raw), &req); err != nil {
		t.Fatal(err)
	}
	if req.Subtype != "elicitation" {
		t.Errorf("got subtype %q, want %q", req.Subtype, "elicitation")
	}
	if req.McpServerName != "my-server" {
		t.Errorf("got mcp_server_name %q", req.McpServerName)
	}
	if *req.Mode != "form" {
		t.Errorf("got mode %q", *req.Mode)
	}
}

func TestControlRequestSubtypes_Marshal(t *testing.T) {
	tests := []struct {
		name    string
		req     interface{}
		subtype string
	}{
		{"interrupt", SDKControlInterruptRequest{Subtype: "interrupt"}, "interrupt"},
		{"set_model", SDKControlSetModelRequest{Subtype: "set_model", Model: "claude-sonnet-4-6"}, "set_model"},
		{"set_permission_mode", SDKControlSetPermissionModeRequest{Subtype: "set_permission_mode", Mode: PermissionModeDefault}, "set_permission_mode"},
		{"set_max_thinking_tokens", SDKControlSetMaxThinkingTokensRequest{Subtype: "set_max_thinking_tokens", MaxThinkingTokens: 1000}, "set_max_thinking_tokens"},
		{"mcp_status", SDKControlMcpStatusRequest{Subtype: "mcp_status"}, "mcp_status"},
		{"mcp_reconnect", SDKControlMcpReconnectRequest{Subtype: "mcp_reconnect", ServerName: "s1"}, "mcp_reconnect"},
		{"mcp_toggle", SDKControlMcpToggleRequest{Subtype: "mcp_toggle", ServerName: "s1", Enabled: true}, "mcp_toggle"},
		{"rewind_files", SDKControlRewindFilesRequest{Subtype: "rewind_files", UserMessageID: "msg-1"}, "rewind_files"},
		{"stop_task", SDKControlStopTaskRequest{Subtype: "stop_task", TaskID: "t-1"}, "stop_task"},
		{"apply_flag_settings", SDKControlApplyFlagSettingsRequest{Subtype: "apply_flag_settings"}, "apply_flag_settings"},
		{"get_settings", SDKControlGetSettingsRequest{Subtype: "get_settings"}, "get_settings"},
		{"cancel_async_message", SDKControlCancelAsyncMessageRequest{Subtype: "cancel_async_message", MessageUUID: "m-1"}, "cancel_async_message"},
		{"end_session", SDKControlEndSessionRequest{Subtype: "end_session"}, "end_session"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := json.Marshal(tt.req)
			if err != nil {
				t.Fatal(err)
			}
			var env struct {
				Subtype string `json:"subtype"`
			}
			if err := json.Unmarshal(data, &env); err != nil {
				t.Fatal(err)
			}
			if env.Subtype != tt.subtype {
				t.Errorf("got subtype %q, want %q", env.Subtype, tt.subtype)
			}
		})
	}
}

func TestMissingControlRequestTypes_Marshal(t *testing.T) {
	tests := []struct {
		name    string
		req     interface{}
		subtype string
	}{
		{"mcp_authenticate", SDKControlMcpAuthenticateRequest{Subtype: "mcp_authenticate", ServerName: "s1"}, "mcp_authenticate"},
		{"mcp_clear_auth", SDKControlMcpClearAuthRequest{Subtype: "mcp_clear_auth", ServerName: "s1"}, "mcp_clear_auth"},
		{"mcp_oauth_callback_url", SDKControlMcpOAuthCallbackUrlRequest{Subtype: "mcp_oauth_callback_url", ServerName: "s1", CallbackUrl: "http://cb"}, "mcp_oauth_callback_url"},
		{"claude_authenticate", SDKControlClaudeAuthenticateRequest{Subtype: "claude_authenticate"}, "claude_authenticate"},
		{"claude_oauth_callback", SDKControlClaudeOAuthCallbackRequest{Subtype: "claude_oauth_callback", CallbackUrl: "http://cb"}, "claude_oauth_callback"},
		{"claude_oauth_wait", SDKControlClaudeOAuthWaitForCompletionRequest{Subtype: "claude_oauth_wait_for_completion"}, "claude_oauth_wait_for_completion"},
		{"remote_control", SDKControlRemoteControlRequest{Subtype: "remote_control", Action: "start"}, "remote_control"},
		{"set_proactive", SDKControlSetProactiveRequest{Subtype: "set_proactive", Proactive: true}, "set_proactive"},
		{"generate_session_title", SDKControlGenerateSessionTitleRequest{Subtype: "generate_session_title"}, "generate_session_title"},
		{"side_question", SDKControlSideQuestionRequest{Subtype: "side_question", Question: "what?"}, "side_question"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := json.Marshal(tt.req)
			if err != nil {
				t.Fatal(err)
			}
			var env struct {
				Subtype string `json:"subtype"`
			}
			if err := json.Unmarshal(data, &env); err != nil {
				t.Fatal(err)
			}
			if env.Subtype != tt.subtype {
				t.Errorf("got subtype %q, want %q", env.Subtype, tt.subtype)
			}
		})
	}
}

func TestSDKControlMcpMessageRequest_JSON(t *testing.T) {
	raw := `{
		"subtype": "mcp_message",
		"server_name": "my-server",
		"message": {"jsonrpc": "2.0", "method": "test"}
	}`
	var req SDKControlMcpMessageRequest
	if err := json.Unmarshal([]byte(raw), &req); err != nil {
		t.Fatal(err)
	}
	if req.Subtype != "mcp_message" {
		t.Errorf("got subtype %q, want %q", req.Subtype, "mcp_message")
	}
	if req.ServerName != "my-server" {
		t.Errorf("got server_name %q", req.ServerName)
	}
}

func TestSDKHookCallbackRequest_JSON(t *testing.T) {
	raw := `{
		"subtype": "hook_callback",
		"hook_event": "PreToolUse",
		"input": {"tool_name": "Bash"}
	}`
	var req SDKHookCallbackRequest
	if err := json.Unmarshal([]byte(raw), &req); err != nil {
		t.Fatal(err)
	}
	if req.Subtype != "hook_callback" {
		t.Errorf("got subtype %q", req.Subtype)
	}
	if req.HookEvent != "PreToolUse" {
		t.Errorf("got hook_event %q", req.HookEvent)
	}
}

func TestSDKControlInitializeResponse_JSON(t *testing.T) {
	raw := `{
		"commands": [{"name": "commit", "description": "Create a commit", "argumentHint": ""}],
		"agents": [],
		"output_style": "default",
		"available_output_styles": ["default", "concise"],
		"models": [{"id": "claude-sonnet-4-6", "name": "Claude Sonnet 4.6"}],
		"account": {"accountUuid": "acc-1"},
		"fast_mode_state": "off"
	}`
	var resp SDKControlInitializeResponse
	if err := json.Unmarshal([]byte(raw), &resp); err != nil {
		t.Fatal(err)
	}
	if len(resp.Commands) != 1 {
		t.Errorf("got %d commands, want 1", len(resp.Commands))
	}
	if resp.Commands[0].Name != "commit" {
		t.Errorf("got command name %q", resp.Commands[0].Name)
	}
	if resp.OutputStyle != "default" {
		t.Errorf("got output_style %q", resp.OutputStyle)
	}
	if *resp.FastModeState != FastModeStateOff {
		t.Errorf("got fast_mode_state %q", *resp.FastModeState)
	}
}

func TestCorrelationEngine_SendAndReceive(t *testing.T) {
	engine := NewCorrelationEngine()
	defer engine.Close()

	// Register a pending request
	ch := engine.Register("req-1")

	// Simulate a response arriving
	resp := &SDKControlResponse{
		Type: "control_response",
	}
	raw, _ := json.Marshal(ControlResponse{
		Subtype:   "success",
		RequestID: "req-1",
	})
	resp.Response = raw

	engine.Deliver("req-1", resp)

	// Receive the response
	ctx := context.Background()
	got, err := WaitForResponse(ctx, ch)
	if err != nil {
		t.Fatal(err)
	}
	if got.Type != "control_response" {
		t.Errorf("got type %q", got.Type)
	}
}

func TestCorrelationEngine_Timeout(t *testing.T) {
	engine := NewCorrelationEngine()
	defer engine.Close()

	ch := engine.Register("req-timeout")

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // immediately cancel

	_, err := WaitForResponse(ctx, ch)
	if err == nil {
		t.Error("expected error from cancelled context")
	}
}

func TestCorrelationEngine_Close(t *testing.T) {
	engine := NewCorrelationEngine()
	ch := engine.Register("req-close")
	engine.Close()

	// Channel should be closed after engine.Close()
	_, ok := <-ch
	if ok {
		t.Error("expected channel to be closed")
	}
}

func TestParseControlRequestSubtype(t *testing.T) {
	raw := `{"subtype": "interrupt"}`
	subtype, err := ParseControlRequestSubtype(json.RawMessage(raw))
	if err != nil {
		t.Fatal(err)
	}
	if subtype != "interrupt" {
		t.Errorf("got %q, want %q", subtype, "interrupt")
	}
}

func TestParseControlResponseSubtype(t *testing.T) {
	tests := []struct {
		raw     string
		subtype string
	}{
		{`{"subtype": "success", "request_id": "r1"}`, "success"},
		{`{"subtype": "error", "request_id": "r2", "error": "oops"}`, "error"},
	}
	for _, tt := range tests {
		subtype, err := ParseControlResponseSubtype(json.RawMessage(tt.raw))
		if err != nil {
			t.Fatal(err)
		}
		if subtype != tt.subtype {
			t.Errorf("got %q, want %q", subtype, tt.subtype)
		}
	}
}

// --- Error path and edge case tests ---

func TestParseControlRequestSubtype_InvalidJSON(t *testing.T) {
	_, err := ParseControlRequestSubtype(json.RawMessage(`{invalid`))
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}

func TestParseControlResponseSubtype_InvalidJSON(t *testing.T) {
	_, err := ParseControlResponseSubtype(json.RawMessage(`not json`))
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}

func TestCorrelationEngine_RegisterOnClosed(t *testing.T) {
	engine := NewCorrelationEngine()
	engine.Close()

	ch := engine.Register("req-after-close")
	// Channel should be immediately closed
	_, ok := <-ch
	if ok {
		t.Error("expected closed channel from Register on closed engine")
	}
}

func TestCorrelationEngine_DeliverNoMatch(t *testing.T) {
	engine := NewCorrelationEngine()
	defer engine.Close()

	ok := engine.Deliver("nonexistent", &SDKControlResponse{Type: "control_response"})
	if ok {
		t.Error("expected false for delivery to nonexistent request ID")
	}
}

func TestWaitForResponse_ClosedChannel(t *testing.T) {
	ch := make(chan *SDKControlResponse, 1)
	close(ch)

	_, err := WaitForResponse(context.Background(), ch)
	if err == nil {
		t.Error("expected error from closed channel")
	}
}

// --- JSON round-trip tests for all control request types ---

func TestControlRequestTypes_RoundTrip(t *testing.T) {
	tests := []struct {
		name string
		req  interface{}
	}{
		{"interrupt", SDKControlInterruptRequest{Subtype: "interrupt"}},
		{"permission", SDKControlPermissionRequest{
			Subtype:   "can_use_tool",
			ToolName:  "Bash",
			Input:     map[string]interface{}{"command": "ls"},
			ToolInput: map[string]interface{}{"command": "ls"},
			ToolUseID: "tu-1",
			AgentID:   strPtr("agent-1"),
		}},
		{"initialize", SDKControlInitializeRequest{
			Subtype:            "initialize",
			ProtocolVersion:    1,
			CanUseTool:         true,
			HasHooks:           true,
			HasElicitation:     true,
			SystemPrompt:       strPtr("test"),
			AppendSystemPrompt: strPtr("append"),
			PromptSuggestions:  boolPtr(true),
		}},
		{"set_model", SDKControlSetModelRequest{Subtype: "set_model", Model: "claude-sonnet-4-6"}},
		{"set_permission_mode", SDKControlSetPermissionModeRequest{Subtype: "set_permission_mode", Mode: PermissionModeDefault}},
		{"set_max_thinking_tokens", SDKControlSetMaxThinkingTokensRequest{Subtype: "set_max_thinking_tokens", MaxThinkingTokens: 1000}},
		{"mcp_status", SDKControlMcpStatusRequest{Subtype: "mcp_status"}},
		{"mcp_message", SDKControlMcpMessageRequest{
			Subtype:    "mcp_message",
			ServerName: "s1",
			Message:    json.RawMessage(`{"jsonrpc":"2.0"}`),
		}},
		{"mcp_set_servers", SDKControlMcpSetServersRequest{
			Subtype: "mcp_set_servers",
			Servers: map[string]interface{}{"s1": map[string]interface{}{"command": "node"}},
		}},
		{"mcp_reconnect", SDKControlMcpReconnectRequest{Subtype: "mcp_reconnect", ServerName: "s1"}},
		{"mcp_toggle", SDKControlMcpToggleRequest{Subtype: "mcp_toggle", ServerName: "s1", Enabled: true}},
		{"rewind_files", SDKControlRewindFilesRequest{
			Subtype:       "rewind_files",
			UserMessageID: "msg-1",
			DryRun:        boolPtr(true),
			FilePaths:     []string{"a.go", "b.go"},
		}},
		{"stop_task", SDKControlStopTaskRequest{Subtype: "stop_task", TaskID: "t-1"}},
		{"apply_flag_settings", SDKControlApplyFlagSettingsRequest{
			Subtype:  "apply_flag_settings",
			Settings: map[string]interface{}{"key": "val"},
		}},
		{"get_settings", SDKControlGetSettingsRequest{Subtype: "get_settings"}},
		{"elicitation", SDKControlElicitationRequest{
			Subtype:       "elicitation",
			McpServerName: "srv",
			Message:       "enter name",
			Mode:          strPtr("form"),
			URL:           strPtr("http://example.com"),
			ElicitationID: strPtr("e-1"),
		}},
		{"cancel_async_message", SDKControlCancelAsyncMessageRequest{Subtype: "cancel_async_message", MessageUUID: "m-1"}},
		{"end_session", SDKControlEndSessionRequest{Subtype: "end_session"}},
		{"mcp_authenticate", SDKControlMcpAuthenticateRequest{Subtype: "mcp_authenticate", ServerName: "s1"}},
		{"mcp_clear_auth", SDKControlMcpClearAuthRequest{Subtype: "mcp_clear_auth", ServerName: "s1"}},
		{"mcp_oauth_callback_url", SDKControlMcpOAuthCallbackUrlRequest{
			Subtype: "mcp_oauth_callback_url", ServerName: "s1", CallbackUrl: "http://cb",
		}},
		{"claude_authenticate", SDKControlClaudeAuthenticateRequest{Subtype: "claude_authenticate"}},
		{"claude_oauth_callback", SDKControlClaudeOAuthCallbackRequest{Subtype: "claude_oauth_callback", CallbackUrl: "http://cb"}},
		{"claude_oauth_wait", SDKControlClaudeOAuthWaitForCompletionRequest{Subtype: "claude_oauth_wait_for_completion"}},
		{"remote_control", SDKControlRemoteControlRequest{
			Subtype: "remote_control", Action: "start",
			Data: map[string]interface{}{"key": "val"},
		}},
		{"set_proactive", SDKControlSetProactiveRequest{Subtype: "set_proactive", Proactive: true}},
		{"generate_session_title", SDKControlGenerateSessionTitleRequest{Subtype: "generate_session_title"}},
		{"side_question", SDKControlSideQuestionRequest{Subtype: "side_question", Question: "what?"}},
		{"hook_callback", SDKHookCallbackRequest{
			Subtype:   "hook_callback",
			HookEvent: "PreToolUse",
			Input:     json.RawMessage(`{"tool_name":"Bash"}`),
		}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := json.Marshal(tt.req)
			if err != nil {
				t.Fatalf("marshal: %v", err)
			}

			// Verify round-trip by unmarshaling back into a generic map
			var m map[string]interface{}
			if err := json.Unmarshal(data, &m); err != nil {
				t.Fatalf("unmarshal to map: %v", err)
			}

			// Verify subtype field is present
			if sub, ok := m["subtype"]; ok {
				if sub == "" {
					t.Error("subtype is empty")
				}
			}
		})
	}
}

// --- Envelope types round-trip ---

func TestSDKControlRequest_RoundTrip(t *testing.T) {
	req := SDKControlRequest{
		Type:      "control_request",
		RequestID: "req-rt",
		Request:   json.RawMessage(`{"subtype":"interrupt"}`),
	}
	data, err := json.Marshal(req)
	if err != nil {
		t.Fatal(err)
	}
	var got SDKControlRequest
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatal(err)
	}
	if got.Type != req.Type || got.RequestID != req.RequestID {
		t.Errorf("round-trip mismatch: got %+v", got)
	}
}

func TestSDKControlResponse_RoundTrip(t *testing.T) {
	resp := SDKControlResponse{
		Type:     "control_response",
		Response: json.RawMessage(`{"subtype":"success","request_id":"r1"}`),
	}
	data, err := json.Marshal(resp)
	if err != nil {
		t.Fatal(err)
	}
	var got SDKControlResponse
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatal(err)
	}
	if got.Type != resp.Type {
		t.Errorf("got type %q", got.Type)
	}
}

func TestSDKControlCancelRequest_RoundTrip(t *testing.T) {
	req := SDKControlCancelRequest{
		Type:      "control_cancel_request",
		RequestID: "req-c",
	}
	data, err := json.Marshal(req)
	if err != nil {
		t.Fatal(err)
	}
	var got SDKControlCancelRequest
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatal(err)
	}
	if got.Type != req.Type || got.RequestID != req.RequestID {
		t.Errorf("round-trip mismatch: got %+v", got)
	}
}

func TestControlErrorResponse_RoundTrip(t *testing.T) {
	resp := ControlErrorResponse{
		Subtype:   "error",
		RequestID: "r-err",
		Error:     "bad request",
		PendingPermissionRequests: []SDKControlRequest{
			{Type: "control_request", RequestID: "p1"},
		},
	}
	data, err := json.Marshal(resp)
	if err != nil {
		t.Fatal(err)
	}
	var got ControlErrorResponse
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatal(err)
	}
	if got.Error != "bad request" {
		t.Errorf("got error %q", got.Error)
	}
	if len(got.PendingPermissionRequests) != 1 {
		t.Errorf("got %d pending requests", len(got.PendingPermissionRequests))
	}
}

func TestSDKControlInitializeResponse_RoundTrip(t *testing.T) {
	fms := FastModeStateOff
	resp := SDKControlInitializeResponse{
		Commands:              []SlashCommand{{Name: "commit", Description: "Create a commit"}},
		Agents:                []json.RawMessage{json.RawMessage(`{"id":"a1"}`)},
		OutputStyle:           "default",
		AvailableOutputStyles: []string{"default", "concise"},
		Models:                []json.RawMessage{json.RawMessage(`{"id":"m1"}`)},
		Account:               json.RawMessage(`{"accountUuid":"acc-1"}`),
		FastModeState:         &fms,
	}
	data, err := json.Marshal(resp)
	if err != nil {
		t.Fatal(err)
	}
	var got SDKControlInitializeResponse
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatal(err)
	}
	if len(got.Commands) != 1 || got.Commands[0].Name != "commit" {
		t.Errorf("commands mismatch")
	}
	if got.OutputStyle != "default" {
		t.Errorf("got output_style %q", got.OutputStyle)
	}
	if *got.FastModeState != FastModeStateOff {
		t.Errorf("got fast_mode_state %q", *got.FastModeState)
	}
}

func TestSDKHookCallbackMatcher_JSON(t *testing.T) {
	raw := `{"matcher":"Bash","hooks":[{"type":"command"}]}`
	var m SDKHookCallbackMatcher
	if err := json.Unmarshal([]byte(raw), &m); err != nil {
		t.Fatal(err)
	}
	if *m.Matcher != "Bash" {
		t.Errorf("got matcher %q", *m.Matcher)
	}
}

// --- Edge cases ---

func TestSDKControlPermissionRequest_NilOptionalFields(t *testing.T) {
	req := SDKControlPermissionRequest{
		Subtype:   "can_use_tool",
		ToolName:  "Bash",
		Input:     map[string]interface{}{"command": "ls"},
		ToolUseID: "tu-1",
		// All optional fields nil
	}
	data, err := json.Marshal(req)
	if err != nil {
		t.Fatal(err)
	}
	var got SDKControlPermissionRequest
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatal(err)
	}
	if got.BlockedPath != nil || got.AgentID != nil || got.Description != nil {
		t.Error("expected nil optional fields")
	}
}

func TestSDKControlInitializeRequest_MinimalFields(t *testing.T) {
	req := SDKControlInitializeRequest{Subtype: "initialize"}
	data, err := json.Marshal(req)
	if err != nil {
		t.Fatal(err)
	}
	var got SDKControlInitializeRequest
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatal(err)
	}
	if got.Subtype != "initialize" {
		t.Errorf("got subtype %q", got.Subtype)
	}
	if got.SystemPrompt != nil {
		t.Error("expected nil SystemPrompt")
	}
}

func TestSDKControlRewindFilesRequest_EmptyFilePaths(t *testing.T) {
	req := SDKControlRewindFilesRequest{
		Subtype:       "rewind_files",
		UserMessageID: "msg-1",
	}
	data, err := json.Marshal(req)
	if err != nil {
		t.Fatal(err)
	}
	var got SDKControlRewindFilesRequest
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatal(err)
	}
	if got.FilePaths != nil {
		t.Error("expected nil FilePaths for empty request")
	}
}

// helpers
func strPtr(s string) *string { return &s }
func boolPtr(b bool) *bool    { return &b }
func intPtr(i int) *int       { return &i }
