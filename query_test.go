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

	if len(received) != 1 {
		t.Fatalf("expected 1 message, got %d", len(received))
	}
	if received[0].MessageType() != "assistant" {
		t.Errorf("message type = %q, want 'assistant'", received[0].MessageType())
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

	// Read init request and send matched response.
	initLine := readStdinLine(t, fp)
	var initEnv struct {
		RequestID string `json:"request_id"`
	}
	json.Unmarshal([]byte(initLine), &initEnv)

	respBody := fmt.Sprintf(`{"request_id":%q,"commands":[],"agents":[],"models":[],"account":{},"output_style":"concise","available_output_styles":[]}`, initEnv.RequestID)
	initResp := map[string]interface{}{
		"type":     "control_response",
		"response": json.RawMessage(respBody),
	}
	data, _ := json.Marshal(initResp)
	fp.stdoutW.Write(append(data, '\n'))

	// Send a permission control request from stdout.
	permReq := map[string]interface{}{
		"type":       "control_request",
		"request_id": "perm-1",
		"request":    json.RawMessage(`{"subtype":"permission","tool_name":"Bash","input":{"command":"ls"},"tool_use_id":"tu-1"}`),
	}
	data, _ = json.Marshal(permReq)
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

	initLine := readStdinLine(t, fp)
	var initEnv struct {
		RequestID string `json:"request_id"`
	}
	json.Unmarshal([]byte(initLine), &initEnv)

	// The response field must contain request_id for correlation.
	respBody := fmt.Sprintf(`{"request_id":%q,"output_style":"concise","available_output_styles":["concise","verbose"],"commands":[],"agents":[],"models":[],"account":{}}`, initEnv.RequestID)
	initResp := map[string]interface{}{
		"type":     "control_response",
		"response": json.RawMessage(respBody),
	}
	data, _ := json.Marshal(initResp)
	fp.stdoutW.Write(append(data, '\n'))

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

	initLine := readStdinLine(t, fp)
	var initEnv struct {
		RequestID string `json:"request_id"`
	}
	json.Unmarshal([]byte(initLine), &initEnv)
	sendMatchedInitResponse(t, fp, initEnv.RequestID)

	permReq := map[string]interface{}{
		"type":       "control_request",
		"request_id": "perm-deny-1",
		"request":    json.RawMessage(`{"subtype":"permission","tool_name":"Bash","input":{"command":"rm -rf /"},"tool_use_id":"tu-d1"}`),
	}
	data, _ := json.Marshal(permReq)
	fp.stdoutW.Write(append(data, '\n'))

	line := readStdinLine(t, fp)
	var respEnv struct {
		Type     string          `json:"type"`
		Response json.RawMessage `json:"response"`
	}
	json.Unmarshal([]byte(line), &respEnv)
	if respEnv.Type != "control_response" {
		t.Errorf("type = %q, want 'control_response'", respEnv.Type)
	}

	var resp struct {
		Behavior string `json:"behavior"`
	}
	json.Unmarshal(respEnv.Response, &resp)
	if resp.Behavior != "deny" {
		t.Errorf("behavior = %q, want 'deny'", resp.Behavior)
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

	initLine := readStdinLine(t, fp)
	var initEnv struct {
		RequestID string `json:"request_id"`
	}
	json.Unmarshal([]byte(initLine), &initEnv)
	sendMatchedInitResponse(t, fp, initEnv.RequestID)

	permReq := map[string]interface{}{
		"type":       "control_request",
		"request_id": "perm-err-1",
		"request":    json.RawMessage(`{"subtype":"permission","tool_name":"Bash","input":{},"tool_use_id":"tu-e1"}`),
	}
	data, _ := json.Marshal(permReq)
	fp.stdoutW.Write(append(data, '\n'))

	line := readStdinLine(t, fp)
	var respEnv struct {
		Response json.RawMessage `json:"response"`
	}
	json.Unmarshal([]byte(line), &respEnv)

	var resp struct {
		Behavior string `json:"behavior"`
		Message  string `json:"message"`
	}
	json.Unmarshal(respEnv.Response, &resp)
	if resp.Behavior != "deny" {
		t.Errorf("behavior = %q, want 'deny'", resp.Behavior)
	}
	if resp.Message != "permission check failed" {
		t.Errorf("message = %q, want 'permission check failed'", resp.Message)
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

	initLine := readStdinLine(t, fp)
	var initEnv struct {
		RequestID string `json:"request_id"`
	}
	json.Unmarshal([]byte(initLine), &initEnv)
	sendMatchedInitResponse(t, fp, initEnv.RequestID)

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
	var respEnv struct {
		Type     string          `json:"type"`
		Response json.RawMessage `json:"response"`
	}
	json.Unmarshal([]byte(line), &respEnv)
	if respEnv.Type != "control_response" {
		t.Errorf("type = %q, want 'control_response'", respEnv.Type)
	}

	var resp struct {
		Action  string                 `json:"action"`
		Content map[string]interface{} `json:"content"`
	}
	json.Unmarshal(respEnv.Response, &resp)
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

	initLine := readStdinLine(t, fp)
	var initEnv struct {
		RequestID string `json:"request_id"`
	}
	json.Unmarshal([]byte(initLine), &initEnv)
	sendMatchedInitResponse(t, fp, initEnv.RequestID)

	elicitReq := map[string]interface{}{
		"type":       "control_request",
		"request_id": "elicit-no-handler",
		"request":    json.RawMessage(`{"subtype":"elicitation","mcp_server_name":"server","message":"auth needed"}`),
	}
	data, _ := json.Marshal(elicitReq)
	fp.stdoutW.Write(append(data, '\n'))

	line := readStdinLine(t, fp)
	var respEnv struct {
		Response json.RawMessage `json:"response"`
	}
	json.Unmarshal([]byte(line), &respEnv)

	var resp struct {
		Action string `json:"action"`
	}
	json.Unmarshal(respEnv.Response, &resp)
	if resp.Action != "decline" {
		t.Errorf("action = %q, want 'decline'", resp.Action)
	}
}

func TestQuery_HookCallback_SendsEmptyResponse(t *testing.T) {
	fp := newFakeProcess()
	q := NewQuery(QueryParams{
		Prompt: "test",
		Options: &Options{
			SpawnClaudeCodeProcess:     func(opts SpawnOptions) SpawnedProcess { return fp },
			PathToClaudeCodeExecutable: fakeCLIPath(t),
		},
	})
	defer func() { q.Close(); fp.stdoutW.Close() }()

	initLine := readStdinLine(t, fp)
	var initEnv struct {
		RequestID string `json:"request_id"`
	}
	json.Unmarshal([]byte(initLine), &initEnv)
	sendMatchedInitResponse(t, fp, initEnv.RequestID)

	hookReq := map[string]interface{}{
		"type":       "control_request",
		"request_id": "hook-1",
		"request":    json.RawMessage(`{"subtype":"hook_callback","hook_event":"PreToolUse","input":{}}`),
	}
	data, _ := json.Marshal(hookReq)
	fp.stdoutW.Write(append(data, '\n'))

	line := readStdinLine(t, fp)
	var respEnv struct {
		Type      string `json:"type"`
		RequestID string `json:"request_id"`
	}
	json.Unmarshal([]byte(line), &respEnv)
	if respEnv.Type != "control_response" {
		t.Errorf("type = %q, want 'control_response'", respEnv.Type)
	}
	if respEnv.RequestID != "hook-1" {
		t.Errorf("request_id = %q, want 'hook-1'", respEnv.RequestID)
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

	// Consume init request so run() doesn't block on stdin write.
	consumeStdinLine(t, fp)

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

	if len(received) != 2 {
		t.Fatalf("expected 2 messages, got %d", len(received))
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

	if len(received) != 1 {
		t.Fatalf("expected 1 valid message, got %d", len(received))
	}
}

func TestQuery_SendInitialize_HasCapabilities(t *testing.T) {
	fp := newFakeProcess()
	q := NewQuery(QueryParams{
		Prompt: "test",
		Options: &Options{
			SpawnClaudeCodeProcess:     func(opts SpawnOptions) SpawnedProcess { return fp },
			PathToClaudeCodeExecutable: fakeCLIPath(t),
			CanUseTool: func(ctx context.Context, toolName string, input map[string]interface{}, opts CanUseToolOptions) (PermissionResult, error) {
				return PermissionResultAllow{Behavior: PermissionBehaviorAllow}, nil
			},
			OnElicitation: func(ctx context.Context, req ElicitationRequest) (*ElicitationResult, error) {
				return nil, nil
			},
		},
	})
	defer func() { q.Close(); fp.stdoutW.Close() }()

	initLine := readStdinLine(t, fp)
	var env struct {
		Type    string          `json:"type"`
		Request json.RawMessage `json:"request"`
	}
	json.Unmarshal([]byte(initLine), &env)

	var req struct {
		Subtype        string `json:"subtype"`
		CanUseTool     bool   `json:"canUseTool"`
		HasElicitation bool   `json:"hasElicitation"`
	}
	json.Unmarshal(env.Request, &req)
	if req.Subtype != "initialize" {
		t.Errorf("subtype = %q, want 'initialize'", req.Subtype)
	}
	if !req.CanUseTool {
		t.Error("canUseTool should be true when CanUseTool handler is set")
	}
	if !req.HasElicitation {
		t.Error("hasElicitation should be true when OnElicitation handler is set")
	}
}

// --- test helpers ---

// consumeStdinLine reads and discards one line from the fake process stdin.
func consumeStdinLine(t *testing.T, fp *fakeProcess) {
	t.Helper()
	_ = readStdinLine(t, fp)
}

// doInitHandshake reads the init request and sends a matched init response.
func doInitHandshake(t *testing.T, fp *fakeProcess) {
	t.Helper()
	initLine := readStdinLine(t, fp)
	var initEnv struct {
		RequestID string `json:"request_id"`
	}
	json.Unmarshal([]byte(initLine), &initEnv)
	sendMatchedInitResponse(t, fp, initEnv.RequestID)
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
func sendControlResponse(t *testing.T, fp *fakeProcess, requestID string) {
	t.Helper()
	resp := map[string]interface{}{
		"type":       "control_response",
		"request_id": requestID,
		"response":   json.RawMessage(`{}`),
	}
	data, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("marshal control response: %v", err)
	}
	fp.stdoutW.Write(append(data, '\n'))
}

// sendMatchedInitResponse sends a control_response matching the init request_id.
func sendMatchedInitResponse(t *testing.T, fp *fakeProcess, requestID string) {
	t.Helper()
	resp := map[string]interface{}{
		"type":       "control_response",
		"request_id": requestID,
		"response":   json.RawMessage(`{"commands":[],"agents":[],"models":[],"account":{},"output_style":"concise","available_output_styles":[]}`),
	}
	data, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("marshal init response: %v", err)
	}
	fp.stdoutW.Write(append(data, '\n'))
}
