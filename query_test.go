package claudeagent

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// fakeCLIPath creates a temp file that passes os.Stat checks for CLIPath.
func fakeCLIPath(t *testing.T) *string {
	t.Helper()
	dir := t.TempDir()
	p := filepath.Join(dir, "claude")
	if err := os.WriteFile(p, []byte("#!/bin/sh\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	return &p
}

func TestNewQuery_ReturnsQuery(t *testing.T) {
	fp := newFakeProcess()
	q := NewQuery(QueryParams{
		Prompt: "hello",
		Options: &Options{
			SpawnClaudeCodeProcess: func(opts SpawnOptions) SpawnedProcess {
				return fp
			},
			PathToClaudeCodeExecutable: fakeCLIPath(t),
		},
	})
	defer q.Close()

	if q == nil {
		t.Fatal("NewQuery returned nil")
	}
	if q.messages == nil {
		t.Error("messages channel is nil")
	}
}

func TestQuery_Messages_ReceivesAssistantMessage(t *testing.T) {
	fp := newFakeProcess()
	q := NewQuery(QueryParams{
		Prompt: "hello",
		Options: &Options{
			SpawnClaudeCodeProcess: func(opts SpawnOptions) SpawnedProcess {
				return fp
			},
			PathToClaudeCodeExecutable: fakeCLIPath(t),
		},
	})
	defer q.Close()

	// Complete the init handshake.
	doInitHandshake(t, fp)

	// Write an assistant message to stdout.
	fp.stdoutW.Write([]byte(`{"type":"assistant","message":{"role":"assistant","content":[{"type":"text","text":"Hi"}]}}` + "\n"))

	// Close stdout to signal end.
	fp.stdoutW.Close()

	var received []SDKMessage
	for msg := range q.Messages() {
		received = append(received, msg)
	}

	// In stream-json mode, init is handled via control_response (not a stream message).
	if len(received) != 1 {
		t.Fatalf("expected 1 message (assistant), got %d", len(received))
	}
	if received[0].MessageType() != "assistant" {
		t.Errorf("message[0] type = %q, want 'assistant'", received[0].MessageType())
	}
}

func TestQuery_Close_StopsMessageChannel(t *testing.T) {
	fp := newFakeProcess()
	q := NewQuery(QueryParams{
		Prompt: "test",
		Options: &Options{
			SpawnClaudeCodeProcess: func(opts SpawnOptions) SpawnedProcess {
				return fp
			},
			PathToClaudeCodeExecutable: fakeCLIPath(t),
		},
	})

	// Close stdin read-end so the init write unblocks with an error.
	fp.stdinR.Close()
	fp.stdoutW.Close()

	q.Close()

	// Drain messages to confirm channel closes.
	for range q.Messages() {
	}
}

func TestQuery_Interrupt_SendsControlRequest(t *testing.T) {
	fp := newFakeProcess()
	q := NewQuery(QueryParams{
		Prompt: "test",
		Options: &Options{
			SpawnClaudeCodeProcess: func(opts SpawnOptions) SpawnedProcess {
				return fp
			},
			PathToClaudeCodeExecutable: fakeCLIPath(t),
		},
	})
	defer func() {
		q.Close()
		fp.stdoutW.Close()
	}()

	// Complete the init handshake.
	doInitHandshake(t, fp)

	// Call Interrupt in a goroutine (it waits for a response).
	ctx := context.Background()
	errCh := make(chan error, 1)
	go func() {
		errCh <- q.Interrupt(ctx)
	}()

	// Read the interrupt control request from stdin.
	line := readStdinLine(t, fp)
	var env struct {
		Type    string          `json:"type"`
		Request json.RawMessage `json:"request"`
	}
	if err := json.Unmarshal([]byte(line), &env); err != nil {
		t.Fatalf("unmarshal interrupt request: %v", err)
	}
	if env.Type != "control_request" {
		t.Fatalf("type = %q, want 'control_request'", env.Type)
	}

	var req struct {
		Subtype string `json:"subtype"`
	}
	json.Unmarshal(env.Request, &req)
	if req.Subtype != "interrupt" {
		t.Errorf("subtype = %q, want 'interrupt'", req.Subtype)
	}
}

func TestQuery_SetPermissionMode_SendsControlRequest(t *testing.T) {
	fp := newFakeProcess()
	q := NewQuery(QueryParams{
		Prompt: "test",
		Options: &Options{
			SpawnClaudeCodeProcess: func(opts SpawnOptions) SpawnedProcess {
				return fp
			},
			PathToClaudeCodeExecutable: fakeCLIPath(t),
		},
	})
	defer func() {
		q.Close()
		fp.stdoutW.Close()
	}()

	doInitHandshake(t, fp)

	ctx := context.Background()
	go func() {
		_ = q.SetPermissionMode(ctx, PermissionModeBypassPermissions)
	}()

	line := readStdinLine(t, fp)
	var env struct {
		Type    string          `json:"type"`
		Request json.RawMessage `json:"request"`
	}
	json.Unmarshal([]byte(line), &env)

	var req struct {
		Subtype string `json:"subtype"`
		Mode    string `json:"mode"`
	}
	json.Unmarshal(env.Request, &req)
	if req.Subtype != "set_permission_mode" {
		t.Errorf("subtype = %q, want 'set_permission_mode'", req.Subtype)
	}
	if req.Mode != "bypassPermissions" {
		t.Errorf("mode = %q, want 'bypassPermissions'", req.Mode)
	}
}

func TestQuery_PermissionCallback(t *testing.T) {
	fp := newFakeProcess()
	permCalled := make(chan bool, 1)

	q := NewQuery(QueryParams{
		Prompt: "test",
		Options: &Options{
			SpawnClaudeCodeProcess: func(opts SpawnOptions) SpawnedProcess {
				return fp
			},
			PathToClaudeCodeExecutable: fakeCLIPath(t),
			CanUseTool: func(ctx context.Context, toolName string, input map[string]interface{}, opts CanUseToolOptions) (PermissionResult, error) {
				permCalled <- true
				return PermissionResultAllow{
					Behavior: PermissionBehaviorAllow,
				}, nil
			},
		},
	})
	defer func() {
		q.Close()
		fp.stdoutW.Close()
	}()

	doInitHandshake(t, fp)

	// Send a permission control request from stdout.
	permReq := map[string]interface{}{
		"type":       "control_request",
		"request_id": "perm-1",
		"request":    json.RawMessage(`{"subtype":"can_use_tool","tool_name":"Bash","input":{"command":"ls"},"tool_use_id":"tu-1"}`),
	}
	data, _ := json.Marshal(permReq)
	fp.stdoutW.Write(append(data, '\n'))

	// Wait for the permission callback.
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	select {
	case called := <-permCalled:
		if !called {
			t.Error("permission callback not called")
		}
	case <-ctx.Done():
		t.Fatal("timed out waiting for permission callback")
	}

	// Read the permission response from stdin.
	line := readStdinLine(t, fp)
	var respEnv struct {
		Type     string          `json:"type"`
		Response json.RawMessage `json:"response"`
	}
	json.Unmarshal([]byte(line), &respEnv)
	if respEnv.Type != "control_response" {
		t.Errorf("type = %q, want 'control_response'", respEnv.Type)
	}
}

func TestQuery_InitializationResult(t *testing.T) {
	fp := newFakeProcess()
	q := NewQuery(QueryParams{
		Prompt: "test",
		Options: &Options{
			SpawnClaudeCodeProcess: func(opts SpawnOptions) SpawnedProcess {
				return fp
			},
			PathToClaudeCodeExecutable: fakeCLIPath(t),
		},
	})
	defer func() {
		q.Close()
		fp.stdoutW.Close()
	}()

	doInitHandshake(t, fp)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	result, err := q.InitializationResult(ctx)
	if err != nil {
		t.Fatalf("InitializationResult: %v", err)
	}
	if result == nil {
		t.Fatal("InitializationResult returned nil")
	}
	if result.OutputStyle != "concise" {
		t.Errorf("OutputStyle = %q, want 'concise'", result.OutputStyle)
	}
}

func TestQuery_SetModel_SendsControlRequest(t *testing.T) {
	fp := newFakeProcess()
	q := NewQuery(QueryParams{
		Prompt: "test",
		Options: &Options{
			SpawnClaudeCodeProcess:     func(opts SpawnOptions) SpawnedProcess { return fp },
			PathToClaudeCodeExecutable: fakeCLIPath(t),
		},
	})
	defer func() { q.Close(); fp.stdoutW.Close() }()

	doInitHandshake(t, fp)

	go func() {
		_ = q.SetModel(context.Background(), String("claude-opus-4-6"))
	}()

	line := readStdinLine(t, fp)
	var env struct {
		Type    string          `json:"type"`
		Request json.RawMessage `json:"request"`
	}
	json.Unmarshal([]byte(line), &env)
	if env.Type != "control_request" {
		t.Fatalf("type = %q, want 'control_request'", env.Type)
	}

	var req struct {
		Subtype string `json:"subtype"`
		Model   string `json:"model"`
	}
	json.Unmarshal(env.Request, &req)
	if req.Subtype != "set_model" {
		t.Errorf("subtype = %q, want 'set_model'", req.Subtype)
	}
	if req.Model != "claude-opus-4-6" {
		t.Errorf("model = %q, want 'claude-opus-4-6'", req.Model)
	}
}

func TestQuery_SetMaxThinkingTokens_SendsControlRequest(t *testing.T) {
	fp := newFakeProcess()
	q := NewQuery(QueryParams{
		Prompt: "test",
		Options: &Options{
			SpawnClaudeCodeProcess:     func(opts SpawnOptions) SpawnedProcess { return fp },
			PathToClaudeCodeExecutable: fakeCLIPath(t),
		},
	})
	defer func() { q.Close(); fp.stdoutW.Close() }()

	doInitHandshake(t, fp)

	go func() {
		_ = q.SetMaxThinkingTokens(context.Background(), Int(8192))
	}()

	line := readStdinLine(t, fp)
	var env struct {
		Type    string          `json:"type"`
		Request json.RawMessage `json:"request"`
	}
	json.Unmarshal([]byte(line), &env)

	var req struct {
		Subtype           string `json:"subtype"`
		MaxThinkingTokens int    `json:"max_thinking_tokens"`
	}
	json.Unmarshal(env.Request, &req)
	if req.Subtype != "set_max_thinking_tokens" {
		t.Errorf("subtype = %q, want 'set_max_thinking_tokens'", req.Subtype)
	}
	if req.MaxThinkingTokens != 8192 {
		t.Errorf("max_thinking_tokens = %d, want 8192", req.MaxThinkingTokens)
	}
}

func TestQuery_StopTask_SendsControlRequest(t *testing.T) {
	fp := newFakeProcess()
	q := NewQuery(QueryParams{
		Prompt: "test",
		Options: &Options{
			SpawnClaudeCodeProcess:     func(opts SpawnOptions) SpawnedProcess { return fp },
			PathToClaudeCodeExecutable: fakeCLIPath(t),
		},
	})
	defer func() { q.Close(); fp.stdoutW.Close() }()

	doInitHandshake(t, fp)

	go func() {
		_ = q.StopTask(context.Background(), "task-42")
	}()

	line := readStdinLine(t, fp)
	var env struct {
		Type    string          `json:"type"`
		Request json.RawMessage `json:"request"`
	}
	json.Unmarshal([]byte(line), &env)

	var req struct {
		Subtype string `json:"subtype"`
		TaskID  string `json:"task_id"`
	}
	json.Unmarshal(env.Request, &req)
	if req.Subtype != "stop_task" {
		t.Errorf("subtype = %q, want 'stop_task'", req.Subtype)
	}
	if req.TaskID != "task-42" {
		t.Errorf("task_id = %q, want 'task-42'", req.TaskID)
	}
}

func TestQuery_PermissionDeny_NoHandler(t *testing.T) {
	fp := newFakeProcess()
	q := NewQuery(QueryParams{
		Prompt: "test",
		Options: &Options{
			SpawnClaudeCodeProcess:     func(opts SpawnOptions) SpawnedProcess { return fp },
			PathToClaudeCodeExecutable: fakeCLIPath(t),
		},
	})
	defer func() { q.Close(); fp.stdoutW.Close() }()

	doInitHandshake(t, fp)

	permReq := map[string]interface{}{
		"type":       "control_request",
		"request_id": "perm-deny-1",
		"request":    json.RawMessage(`{"subtype":"can_use_tool","tool_name":"Bash","input":{"command":"rm -rf /"},"tool_use_id":"tu-d1"}`),
	}
	data, _ := json.Marshal(permReq)
	fp.stdoutW.Write(append(data, '\n'))

	line := readStdinLine(t, fp)
	_, _, isError, errMsg := parseResponsePayload(t, line)
	// No handler: should send error response
	if !isError {
		t.Error("expected error response when no handler configured")
	}
	if errMsg == "" {
		t.Error("expected non-empty error message")
	}
}

func TestQuery_PermissionCallback_Error(t *testing.T) {
	fp := newFakeProcess()
	q := NewQuery(QueryParams{
		Prompt: "test",
		Options: &Options{
			SpawnClaudeCodeProcess:     func(opts SpawnOptions) SpawnedProcess { return fp },
			PathToClaudeCodeExecutable: fakeCLIPath(t),
			CanUseTool: func(ctx context.Context, toolName string, input map[string]interface{}, opts CanUseToolOptions) (PermissionResult, error) {
				return nil, fmt.Errorf("permission check failed")
			},
		},
	})
	defer func() { q.Close(); fp.stdoutW.Close() }()

	doInitHandshake(t, fp)

	permReq := map[string]interface{}{
		"type":       "control_request",
		"request_id": "perm-err-1",
		"request":    json.RawMessage(`{"subtype":"can_use_tool","tool_name":"Bash","input":{},"tool_use_id":"tu-e1"}`),
	}
	data, _ := json.Marshal(permReq)
	fp.stdoutW.Write(append(data, '\n'))

	line := readStdinLine(t, fp)
	_, _, isError, errMsg := parseResponsePayload(t, line)
	if !isError {
		t.Error("expected error response from failed permission callback")
	}
	if errMsg != "permission check failed" {
		t.Errorf("error = %q, want 'permission check failed'", errMsg)
	}
}

func TestQuery_Elicitation_WithHandler(t *testing.T) {
	fp := newFakeProcess()
	elicitCalled := make(chan bool, 1)

	q := NewQuery(QueryParams{
		Prompt: "test",
		Options: &Options{
			SpawnClaudeCodeProcess:     func(opts SpawnOptions) SpawnedProcess { return fp },
			PathToClaudeCodeExecutable: fakeCLIPath(t),
			OnElicitation: func(ctx context.Context, req ElicitationRequest) (*ElicitationResult, error) {
				elicitCalled <- true
				if req.ServerName != "my-mcp" {
					t.Errorf("server name = %q, want 'my-mcp'", req.ServerName)
				}
				return &ElicitationResult{
					Action:  "accept",
					Content: map[string]interface{}{"token": "abc123"},
				}, nil
			},
		},
	})
	defer func() { q.Close(); fp.stdoutW.Close() }()

	doInitHandshake(t, fp)

	elicitReq := map[string]interface{}{
		"type":       "control_request",
		"request_id": "elicit-1",
		"request":    json.RawMessage(`{"subtype":"elicitation","mcp_server_name":"my-mcp","message":"Please authenticate"}`),
	}
	data, _ := json.Marshal(elicitReq)
	fp.stdoutW.Write(append(data, '\n'))

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	select {
	case <-elicitCalled:
	case <-ctx.Done():
		t.Fatal("timed out waiting for elicitation callback")
	}

	line := readStdinLine(t, fp)
	_, payload, isError, _ := parseResponsePayload(t, line)
	if isError {
		t.Fatal("expected success response for elicitation")
	}
	var resp struct {
		Action  string                 `json:"action"`
		Content map[string]interface{} `json:"content"`
	}
	json.Unmarshal(payload, &resp)
	if resp.Action != "accept" {
		t.Errorf("action = %q, want 'accept'", resp.Action)
	}
}

func TestQuery_Elicitation_NoHandler_Declines(t *testing.T) {
	fp := newFakeProcess()
	q := NewQuery(QueryParams{
		Prompt: "test",
		Options: &Options{
			SpawnClaudeCodeProcess:     func(opts SpawnOptions) SpawnedProcess { return fp },
			PathToClaudeCodeExecutable: fakeCLIPath(t),
		},
	})
	defer func() { q.Close(); fp.stdoutW.Close() }()

	doInitHandshake(t, fp)

	elicitReq := map[string]interface{}{
		"type":       "control_request",
		"request_id": "elicit-no-handler",
		"request":    json.RawMessage(`{"subtype":"elicitation","mcp_server_name":"server","message":"auth needed"}`),
	}
	data, _ := json.Marshal(elicitReq)
	fp.stdoutW.Write(append(data, '\n'))

	line := readStdinLine(t, fp)
	_, payload, isError, _ := parseResponsePayload(t, line)
	if isError {
		t.Fatal("expected success response for elicitation decline")
	}
	var resp struct {
		Action string `json:"action"`
	}
	json.Unmarshal(payload, &resp)
	if resp.Action != "decline" {
		t.Errorf("action = %q, want 'decline'", resp.Action)
	}
}

func TestQuery_HookCallback_UnknownCallbackID_Error(t *testing.T) {
	fp := newFakeProcess()
	q := NewQuery(QueryParams{
		Prompt: "test",
		Options: &Options{
			SpawnClaudeCodeProcess:     func(opts SpawnOptions) SpawnedProcess { return fp },
			PathToClaudeCodeExecutable: fakeCLIPath(t),
		},
	})
	defer func() { q.Close(); fp.stdoutW.Close() }()

	doInitHandshake(t, fp)

	hookReq := map[string]interface{}{
		"type":       "control_request",
		"request_id": "hook-1",
		"request":    json.RawMessage(`{"subtype":"hook_callback","callback_id":"nonexistent","input":{}}`),
	}
	data, _ := json.Marshal(hookReq)
	fp.stdoutW.Write(append(data, '\n'))

	line := readStdinLine(t, fp)
	reqID, _, isError, errMsg := parseResponsePayload(t, line)
	if reqID != "hook-1" {
		t.Errorf("request_id = %q, want 'hook-1'", reqID)
	}
	if !isError {
		t.Error("expected error response for unknown callback ID")
	}
	if errMsg == "" {
		t.Error("expected non-empty error message")
	}
}

func TestQuery_HookCallback_WithRegisteredHook(t *testing.T) {
	fp := newFakeProcess()
	hookCalled := make(chan bool, 1)

	q := NewQuery(QueryParams{
		Prompt: "test",
		Options: &Options{
			SpawnClaudeCodeProcess:     func(opts SpawnOptions) SpawnedProcess { return fp },
			PathToClaudeCodeExecutable: fakeCLIPath(t),
			Hooks: map[HookEvent][]HookCallbackMatcher{
				"PreToolUse": {
					{
						Matcher: strPtr("Bash"),
						Hooks: []HookCallback{
							func(ctx context.Context, input HookInput, toolUseID *string) (HookJSONOutput, error) {
								hookCalled <- true
								return SyncHookJSONOutput{
									Decision: strPtr("approve"),
								}, nil
							},
						},
					},
				},
			},
		},
	})
	defer func() { q.Close(); fp.stdoutW.Close() }()

	doInitHandshake(t, fp)

	// The hook was registered as "hook_0" during initialize.
	// Send a hook_callback with that callback_id.
	hookReq := map[string]interface{}{
		"type":       "control_request",
		"request_id": "hook-reg-1",
		"request":    json.RawMessage(`{"subtype":"hook_callback","callback_id":"hook_0","input":{"tool_name":"Bash"}}`),
	}
	data, _ := json.Marshal(hookReq)
	fp.stdoutW.Write(append(data, '\n'))

	// Wait for the hook callback to be called.
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	select {
	case <-hookCalled:
		// Good
	case <-ctx.Done():
		t.Fatal("timed out waiting for hook callback")
	}

	// Read the response.
	line := readStdinLine(t, fp)
	reqID, payload, isError, _ := parseResponsePayload(t, line)
	if reqID != "hook-reg-1" {
		t.Errorf("request_id = %q, want 'hook-reg-1'", reqID)
	}
	if isError {
		t.Error("expected success response for registered hook callback")
	}
	var resp struct {
		Decision string `json:"decision"`
	}
	json.Unmarshal(payload, &resp)
	if resp.Decision != "approve" {
		t.Errorf("decision = %q, want 'approve'", resp.Decision)
	}
}

func TestQuery_ControlCancelRequest(t *testing.T) {
	fp := newFakeProcess()
	hookBlocked := make(chan struct{})

	q := NewQuery(QueryParams{
		Prompt: "test",
		Options: &Options{
			SpawnClaudeCodeProcess:     func(opts SpawnOptions) SpawnedProcess { return fp },
			PathToClaudeCodeExecutable: fakeCLIPath(t),
			Hooks: map[HookEvent][]HookCallbackMatcher{
				"PreToolUse": {
					{
						Hooks: []HookCallback{
							func(ctx context.Context, input HookInput, toolUseID *string) (HookJSONOutput, error) {
								close(hookBlocked)
								// Block until cancelled
								<-ctx.Done()
								return nil, ctx.Err()
							},
						},
					},
				},
			},
		},
	})
	defer func() { q.Close(); fp.stdoutW.Close() }()

	doInitHandshake(t, fp)

	// Send a hook_callback that will block
	hookReq := map[string]interface{}{
		"type":       "control_request",
		"request_id": "hook-cancel-1",
		"request":    json.RawMessage(`{"subtype":"hook_callback","callback_id":"hook_0","input":{}}`),
	}
	data, _ := json.Marshal(hookReq)
	fp.stdoutW.Write(append(data, '\n'))

	// Wait until the hook is actually blocked
	<-hookBlocked

	// Send cancel request
	cancelReq := map[string]interface{}{
		"type":       "control_cancel_request",
		"request_id": "hook-cancel-1",
	}
	cancelData, _ := json.Marshal(cancelReq)
	fp.stdoutW.Write(append(cancelData, '\n'))

	// Read the error response (context cancelled)
	line := readStdinLine(t, fp)
	_, _, isError, _ := parseResponsePayload(t, line)
	if !isError {
		t.Error("expected error response from cancelled hook callback")
	}
}

func TestQuery_InitializationResult_Timeout(t *testing.T) {
	fp := newFakeProcess()
	q := NewQuery(QueryParams{
		Prompt: "test",
		Options: &Options{
			SpawnClaudeCodeProcess:     func(opts SpawnOptions) SpawnedProcess { return fp },
			PathToClaudeCodeExecutable: fakeCLIPath(t),
		},
	})
	defer func() {
		q.Close()
		fp.stdinR.Close()
		fp.stdoutW.Close()
	}()

	// No init message sent on stdout — InitializationResult should timeout.
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	_, err := q.InitializationResult(ctx)
	if err == nil {
		t.Fatal("expected timeout error")
	}
	if err != context.DeadlineExceeded {
		t.Errorf("err = %v, want context.DeadlineExceeded", err)
	}
}

func TestQuery_Close_Idempotent(t *testing.T) {
	fp := newFakeProcess()
	q := NewQuery(QueryParams{
		Prompt: "test",
		Options: &Options{
			SpawnClaudeCodeProcess:     func(opts SpawnOptions) SpawnedProcess { return fp },
			PathToClaudeCodeExecutable: fakeCLIPath(t),
		},
	})

	fp.stdinR.Close()
	fp.stdoutW.Close()

	q.Close()
	q.Close()
	q.Close()

	for range q.Messages() {
	}
}

func TestQuery_MultipleMessages(t *testing.T) {
	fp := newFakeProcess()
	q := NewQuery(QueryParams{
		Prompt: "test",
		Options: &Options{
			SpawnClaudeCodeProcess:     func(opts SpawnOptions) SpawnedProcess { return fp },
			PathToClaudeCodeExecutable: fakeCLIPath(t),
		},
	})
	defer q.Close()

	doInitHandshake(t, fp)

	fp.stdoutW.Write([]byte(`{"type":"assistant","message":{"role":"assistant","content":[]}}` + "\n"))
	fp.stdoutW.Write([]byte(`{"type":"result","subtype":"success","result":"done","is_error":false}` + "\n"))
	fp.stdoutW.Close()

	var received []SDKMessage
	for msg := range q.Messages() {
		received = append(received, msg)
	}

	// Init is now via control_response, not a stream message.
	if len(received) != 2 {
		t.Fatalf("expected 2 messages (assistant + result), got %d", len(received))
	}
	if received[0].MessageType() != "assistant" {
		t.Errorf("msg[0] type = %q, want 'assistant'", received[0].MessageType())
	}
	if received[1].MessageType() != "result" {
		t.Errorf("msg[1] type = %q, want 'result'", received[1].MessageType())
	}
}

func TestQuery_MalformedLines_Skipped(t *testing.T) {
	fp := newFakeProcess()
	q := NewQuery(QueryParams{
		Prompt: "test",
		Options: &Options{
			SpawnClaudeCodeProcess:     func(opts SpawnOptions) SpawnedProcess { return fp },
			PathToClaudeCodeExecutable: fakeCLIPath(t),
		},
	})
	defer q.Close()

	doInitHandshake(t, fp)

	fp.stdoutW.Write([]byte("not json at all\n"))
	fp.stdoutW.Write([]byte("\n"))
	fp.stdoutW.Write([]byte(`{"type":"assistant","message":{"role":"assistant","content":[]}}` + "\n"))
	fp.stdoutW.Write([]byte("{bad json\n"))
	fp.stdoutW.Close()

	var received []SDKMessage
	for msg := range q.Messages() {
		received = append(received, msg)
	}

	// Init is via control_response, not a stream message.
	if len(received) != 1 {
		t.Fatalf("expected 1 valid message (assistant), got %d", len(received))
	}
}

func TestQuery_InitFromControlResponse_CapturesInitResult(t *testing.T) {
	fp := newFakeProcess()
	q := NewQuery(QueryParams{
		Prompt: "test",
		Options: &Options{
			SpawnClaudeCodeProcess:     func(opts SpawnOptions) SpawnedProcess { return fp },
			PathToClaudeCodeExecutable: fakeCLIPath(t),
		},
	})
	defer func() { q.Close(); fp.stdoutW.Close() }()

	doInitHandshake(t, fp)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	result, err := q.InitializationResult(ctx)
	if err != nil {
		t.Fatalf("InitializationResult: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil init result")
	}
	if result.OutputStyle != "concise" {
		t.Errorf("output_style = %q, want 'concise'", result.OutputStyle)
	}
}

// --- test helpers ---

// parseResponsePayload extracts the inner response payload from the nested
// control_response envelope: {type:"control_response", response:{subtype:"success", request_id:X, response:{...}}}
func parseResponsePayload(t *testing.T, line string) (requestID string, payload json.RawMessage, isError bool, errMsg string) {
	t.Helper()
	var outer struct {
		Type     string          `json:"type"`
		Response json.RawMessage `json:"response"`
	}
	if err := json.Unmarshal([]byte(line), &outer); err != nil {
		t.Fatalf("parse outer envelope: %v", err)
	}
	if outer.Type != "control_response" {
		t.Fatalf("type = %q, want 'control_response'", outer.Type)
	}
	var inner struct {
		Subtype   string          `json:"subtype"`
		RequestID string          `json:"request_id"`
		Response  json.RawMessage `json:"response"`
		Error     string          `json:"error"`
	}
	if err := json.Unmarshal(outer.Response, &inner); err != nil {
		t.Fatalf("parse inner envelope: %v", err)
	}
	if inner.Subtype == "error" {
		return inner.RequestID, nil, true, inner.Error
	}
	return inner.RequestID, inner.Response, false, ""
}

// consumeStdinLine reads and discards one line from the fake process stdin.
func consumeStdinLine(t *testing.T, fp *fakeProcess) {
	t.Helper()
	_ = readStdinLine(t, fp)
}

// doInitHandshake reads the user message and initialize request from stdin,
// then sends back a control_response with matching request_id on stdout.
func doInitHandshake(t *testing.T, fp *fakeProcess) {
	t.Helper()

	// 1. Read the user message written to stdin
	consumeStdinLine(t, fp)

	// 2. Read the initialize control_request from stdin
	initLine := readStdinLine(t, fp)
	var envelope struct {
		Type      string `json:"type"`
		RequestID string `json:"request_id"`
	}
	if err := json.Unmarshal([]byte(initLine), &envelope); err != nil {
		t.Fatalf("parse init request: %v (line: %s)", err, initLine)
	}
	if envelope.Type != "control_request" {
		t.Fatalf("expected control_request, got %q", envelope.Type)
	}

	// 3. Send back a control_response with the init response payload
	sendMatchedInitResponse(t, fp, envelope.RequestID)

	// Small delay to let the goroutine process the response
	time.Sleep(10 * time.Millisecond)
}

// readStdinLine reads one newline-terminated line from the fake process stdin pipe.
func readStdinLine(t *testing.T, fp *fakeProcess) string {
	t.Helper()
	buf := make([]byte, 0, 4096)
	tmp := make([]byte, 1)
	for {
		n, err := fp.stdinR.Read(tmp)
		if err != nil {
			t.Fatalf("reading stdin: %v", err)
		}
		if n > 0 {
			if tmp[0] == '\n' {
				return string(buf)
			}
			buf = append(buf, tmp[0])
		}
	}
}

// sendControlResponse writes a generic control_response to the fake process stdout.
// Uses the nested envelope format matching the TS SDK.
func sendControlResponse(t *testing.T, fp *fakeProcess, requestID string) {
	t.Helper()
	innerResp := map[string]interface{}{
		"subtype":    "success",
		"request_id": requestID,
		"response":   json.RawMessage(`{}`),
	}
	innerJSON, _ := json.Marshal(innerResp)
	resp := map[string]interface{}{
		"type":     "control_response",
		"response": json.RawMessage(innerJSON),
	}
	data, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("marshal control response: %v", err)
	}
	fp.stdoutW.Write(append(data, '\n'))
}

// sendMatchedInitResponse sends a control_response matching the init request_id.
// Uses the nested envelope format: {type:"control_response", response:{subtype:"success", request_id:X, response:{...}}}
func sendMatchedInitResponse(t *testing.T, fp *fakeProcess, requestID string) {
	t.Helper()
	innerResp := map[string]interface{}{
		"subtype":    "success",
		"request_id": requestID,
		"response":   json.RawMessage(`{"commands":[],"agents":[],"models":[],"account":{},"output_style":"concise","available_output_styles":[]}`),
	}
	innerJSON, _ := json.Marshal(innerResp)
	resp := map[string]interface{}{
		"type":     "control_response",
		"response": json.RawMessage(innerJSON),
	}
	data, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("marshal init response: %v", err)
	}
	fp.stdoutW.Write(append(data, '\n'))
}
