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

	// Consume the init request from stdin.
	consumeStdinLine(t, fp)

	// Send a fake init response (unmatched request_id is fine, init goroutine handles it).
	sendControlResponse(t, fp, "")

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

	// Consume init request.
	consumeStdinLine(t, fp)
	sendControlResponse(t, fp, "")

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

	consumeStdinLine(t, fp)
	sendControlResponse(t, fp, "")

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

// --- test helpers ---

// consumeStdinLine reads and discards one line from the fake process stdin.
func consumeStdinLine(t *testing.T, fp *fakeProcess) {
	t.Helper()
	_ = readStdinLine(t, fp)
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
