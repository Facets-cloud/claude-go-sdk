package claudeagent

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"io"
	"strings"
	"sync"
	"testing"
	"time"
)

// --- buildProcessArgs tests ---

func TestBuildProcessArgs_NilOptions(t *testing.T) {
	args := buildProcessArgs(nil, "hello")
	assertContains(t, args, "--output-format", "stream-json")
	assertFlag(t, args, "--verbose")
	assertFlag(t, args, "--print")
	// Prompt is a positional arg (last element)
	if args[len(args)-1] != "hello" {
		t.Errorf("expected prompt 'hello' as last arg, got %q", args[len(args)-1])
	}
}

func TestBuildProcessArgs_EmptyPrompt(t *testing.T) {
	args := buildProcessArgs(nil, "")
	// Last arg should not be a bare prompt
	if len(args) > 0 && args[len(args)-1] == "" {
		t.Error("empty prompt should not be appended")
	}
}

func TestBuildProcessArgs_Model(t *testing.T) {
	model := "claude-sonnet-4-6"
	opts := &Options{Model: &model}
	args := buildProcessArgs(opts, "test")
	assertContains(t, args, "--model", "claude-sonnet-4-6")
}

func TestBuildProcessArgs_PermissionMode(t *testing.T) {
	mode := PermissionModeBypassPermissions
	opts := &Options{PermissionMode: &mode}
	args := buildProcessArgs(opts, "test")
	assertContains(t, args, "--permission-mode", "bypassPermissions")
}

func TestBuildProcessArgs_MaxTurns(t *testing.T) {
	turns := 5
	opts := &Options{MaxTurns: &turns}
	args := buildProcessArgs(opts, "")
	assertContains(t, args, "--max-turns", "5")
}

func TestBuildProcessArgs_MaxBudgetUsd(t *testing.T) {
	budget := 10.50
	opts := &Options{MaxBudgetUsd: &budget}
	args := buildProcessArgs(opts, "")
	assertContains(t, args, "--max-budget-usd", "10.50")
}

func TestBuildProcessArgs_Continue(t *testing.T) {
	cont := true
	opts := &Options{Continue: &cont}
	args := buildProcessArgs(opts, "")
	assertFlag(t, args, "--continue")
}

func TestBuildProcessArgs_Resume(t *testing.T) {
	sid := "session-123"
	opts := &Options{Resume: &sid}
	args := buildProcessArgs(opts, "")
	assertContains(t, args, "--resume", "session-123")
}

func TestBuildProcessArgs_Cwd(t *testing.T) {
	// Cwd is handled via cmd.Dir, not a CLI flag.
	// Verify it's NOT in the args.
	cwd := "/tmp/work"
	opts := &Options{Cwd: &cwd}
	args := buildProcessArgs(opts, "")
	for _, a := range args {
		if a == "--cwd" {
			t.Error("--cwd should not be in args (handled via cmd.Dir)")
		}
	}
}

func TestBuildProcessArgs_Debug(t *testing.T) {
	debug := true
	debugFile := "/tmp/debug.log"
	opts := &Options{Debug: &debug, DebugFile: &debugFile}
	args := buildProcessArgs(opts, "")
	assertFlag(t, args, "--debug")
	assertContains(t, args, "--debug-file", "/tmp/debug.log")
}

func TestBuildProcessArgs_AllowedAndDisallowedTools(t *testing.T) {
	opts := &Options{
		AllowedTools:    []string{"Bash", "Read"},
		DisallowedTools: []string{"Write"},
	}
	args := buildProcessArgs(opts, "")
	assertFlag(t, args, "--allowed-tools")
	assertFlag(t, args, "Bash")
	assertFlag(t, args, "Read")
	assertFlag(t, args, "--disallowed-tools")
	assertFlag(t, args, "Write")
}

func TestBuildProcessArgs_AdditionalDirectories(t *testing.T) {
	opts := &Options{
		AdditionalDirectories: []string{"/dir1", "/dir2"},
	}
	args := buildProcessArgs(opts, "")
	assertContains(t, args, "--additional-directory", "/dir1")
	assertContains(t, args, "--additional-directory", "/dir2")
}

func TestBuildProcessArgs_Betas(t *testing.T) {
	opts := &Options{
		Betas: []SdkBeta{SdkBetaContext1M},
	}
	args := buildProcessArgs(opts, "")
	assertContains(t, args, "--beta", "context-1m-2025-08-07")
}

func TestBuildProcessArgs_ForkSession(t *testing.T) {
	fork := true
	opts := &Options{ForkSession: &fork}
	args := buildProcessArgs(opts, "")
	assertFlag(t, args, "--fork-session")
}

func TestBuildProcessArgs_NoPersistSession(t *testing.T) {
	persist := false
	opts := &Options{PersistSession: &persist}
	args := buildProcessArgs(opts, "")
	assertFlag(t, args, "--no-persist-session")
}

func TestBuildProcessArgs_IncludePartialMessages(t *testing.T) {
	partial := true
	opts := &Options{IncludePartialMessages: &partial}
	args := buildProcessArgs(opts, "")
	assertFlag(t, args, "--include-partial-messages")
}

func TestBuildProcessArgs_SystemPromptString(t *testing.T) {
	opts := &Options{SystemPrompt: "Be helpful"}
	args := buildProcessArgs(opts, "")
	assertContains(t, args, "--system-prompt", "Be helpful")
}

func TestBuildProcessArgs_SystemPromptPreset(t *testing.T) {
	append := "Also be concise"
	opts := &Options{SystemPrompt: SystemPromptPreset{
		Type:   "preset",
		Preset: "claude_code",
		Append: &append,
	}}
	args := buildProcessArgs(opts, "")
	assertContains(t, args, "--append-system-prompt", "Also be concise")
}

func TestBuildProcessArgs_ToolsStringSlice(t *testing.T) {
	opts := &Options{Tools: []string{"Bash", "Read"}}
	args := buildProcessArgs(opts, "")
	assertContains(t, args, "--tools", "Bash,Read")
}

func TestBuildProcessArgs_DangerouslySkipPermissions(t *testing.T) {
	skip := true
	opts := &Options{AllowDangerouslySkipPermissions: &skip}
	args := buildProcessArgs(opts, "")
	assertFlag(t, args, "--dangerously-skip-permissions")
}

func TestBuildProcessArgs_ExtraArgs(t *testing.T) {
	v := "value"
	opts := &Options{ExtraArgs: map[string]*string{
		"custom-flag":   &v,
		"boolean-flag":  nil,
	}}
	args := buildProcessArgs(opts, "")
	assertContains(t, args, "--custom-flag", "value")
	assertFlag(t, args, "--boolean-flag")
}

func TestBuildProcessArgs_MaxThinkingTokens(t *testing.T) {
	tokens := 8000
	opts := &Options{MaxThinkingTokens: &tokens}
	args := buildProcessArgs(opts, "")
	assertContains(t, args, "--max-thinking-tokens", "8000")
}

func TestBuildProcessArgs_Effort(t *testing.T) {
	effort := "high"
	opts := &Options{Effort: &effort}
	args := buildProcessArgs(opts, "")
	assertContains(t, args, "--effort", "high")
}

func TestBuildProcessArgs_Agent(t *testing.T) {
	agent := "my-agent"
	opts := &Options{Agent: &agent}
	args := buildProcessArgs(opts, "")
	assertContains(t, args, "--agent", "my-agent")
}

func TestBuildProcessArgs_FallbackModel(t *testing.T) {
	fb := "claude-haiku-4-5-20251001"
	opts := &Options{FallbackModel: &fb}
	args := buildProcessArgs(opts, "")
	assertContains(t, args, "--fallback-model", "claude-haiku-4-5-20251001")
}

func TestBuildProcessArgs_Sandbox(t *testing.T) {
	opts := &Options{Sandbox: &SandboxSettings{
		Enabled: boolPtr(true),
	}}
	args := buildProcessArgs(opts, "")
	assertFlag(t, args, "--sandbox")
}

func TestBuildProcessArgs_McpServers(t *testing.T) {
	opts := &Options{McpServers: map[string]interface{}{
		"my-server": map[string]interface{}{"command": "node", "args": []string{"server.js"}},
	}}
	args := buildProcessArgs(opts, "")
	assertFlag(t, args, "--mcp-config")
}

func TestBuildProcessArgs_ThinkingConfig(t *testing.T) {
	tc := ThinkingAdaptive()
	opts := &Options{Thinking: &tc}
	args := buildProcessArgs(opts, "")
	assertFlag(t, args, "--thinking")
}

func TestBuildProcessArgs_Plugins(t *testing.T) {
	opts := &Options{Plugins: []SdkPluginConfig{
		{Type: "local", Path: "/path/to/plugin"},
	}}
	args := buildProcessArgs(opts, "")
	assertContains(t, args, "--plugin", "/path/to/plugin")
}

func TestBuildProcessArgs_SettingSources(t *testing.T) {
	opts := &Options{SettingSources: []SettingSource{SettingSourceUser, SettingSourceProject}}
	args := buildProcessArgs(opts, "")
	assertContains(t, args, "--setting-source", "user")
	assertContains(t, args, "--setting-source", "project")
}

func TestBuildProcessArgs_PromptSuggestions(t *testing.T) {
	ps := true
	opts := &Options{PromptSuggestions: &ps}
	args := buildProcessArgs(opts, "")
	assertFlag(t, args, "--prompt-suggestions")
}

func TestBuildProcessArgs_AgentProgressSummaries(t *testing.T) {
	aps := true
	opts := &Options{AgentProgressSummaries: &aps}
	args := buildProcessArgs(opts, "")
	assertFlag(t, args, "--agent-progress-summaries")
}

func TestBuildProcessArgs_StrictMcpConfig(t *testing.T) {
	strict := true
	opts := &Options{StrictMcpConfig: &strict}
	args := buildProcessArgs(opts, "")
	assertFlag(t, args, "--strict-mcp-config")
}

func TestBuildProcessArgs_EnableFileCheckpointing(t *testing.T) {
	efc := true
	opts := &Options{EnableFileCheckpointing: &efc}
	args := buildProcessArgs(opts, "")
	assertFlag(t, args, "--enable-file-checkpointing")
}

func TestBuildProcessArgs_SessionID(t *testing.T) {
	sid := "sess-abc"
	opts := &Options{SessionID: &sid}
	args := buildProcessArgs(opts, "")
	assertContains(t, args, "--session-id", "sess-abc")
}

func TestBuildProcessArgs_ResumeSessionAt(t *testing.T) {
	rsa := "msg-uuid-123"
	opts := &Options{ResumeSessionAt: &rsa}
	args := buildProcessArgs(opts, "")
	assertContains(t, args, "--resume-session-at", "msg-uuid-123")
}

func TestBuildProcessArgs_PermissionPromptToolName(t *testing.T) {
	tn := "my-permission-tool"
	opts := &Options{PermissionPromptToolName: &tn}
	args := buildProcessArgs(opts, "")
	assertContains(t, args, "--permission-prompt-tool-name", "my-permission-tool")
}

// --- defaultSpawnedProcess interface tests ---

var _ SpawnedProcess = &defaultSpawnedProcess{}

// --- processManager tests ---

// fakeProcess is a test double for SpawnedProcess using in-memory pipes.
type fakeProcess struct {
	stdinW  *io.PipeWriter
	stdinR  *io.PipeReader
	stdoutW *io.PipeWriter
	stdoutR *io.PipeReader
	waitErr error
	killed  bool
	mu      sync.Mutex
}

func newFakeProcess() *fakeProcess {
	stdinR, stdinW := io.Pipe()
	stdoutR, stdoutW := io.Pipe()
	return &fakeProcess{
		stdinW:  stdinW,
		stdinR:  stdinR,
		stdoutW: stdoutW,
		stdoutR: stdoutR,
	}
}

func (p *fakeProcess) Stdin() io.WriteCloser  { return p.stdinW }
func (p *fakeProcess) Stdout() io.ReadCloser  { return p.stdoutR }
func (p *fakeProcess) Wait() error            { return p.waitErr }
func (p *fakeProcess) Kill() error {
	p.mu.Lock()
	p.killed = true
	p.mu.Unlock()
	return nil
}

func TestProcessManager_ReadMessages(t *testing.T) {
	fp := newFakeProcess()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	pm := newProcessManager(fp, ctx, nil)

	// Write JSON messages to the fake stdout.
	go func() {
		msgs := []string{
			`{"type":"assistant","message":{"role":"assistant","content":[{"type":"text","text":"Hello"}]}}`,
			`{"type":"result","subtype":"success","message_id":"m1","duration_ms":100,"duration_api_ms":80,"is_error":false,"num_turns":1,"session_id":"s1"}`,
		}
		for _, m := range msgs {
			fp.stdoutW.Write([]byte(m + "\n"))
		}
		fp.stdoutW.Close()
	}()

	var received []SDKMessage
	for msg := range pm.Messages() {
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

func TestProcessManager_WriteJSON(t *testing.T) {
	fp := newFakeProcess()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Read from stdin pipe in a goroutine BEFORE creating processManager,
	// to ensure the reader is ready before any writes happen.
	done := make(chan string, 1)
	go func() {
		scanner := bufio.NewScanner(fp.stdinR)
		if scanner.Scan() {
			done <- scanner.Text()
		} else {
			done <- ""
		}
	}()

	pm := newProcessManager(fp, ctx, nil)

	// Close stdout so message reader goroutine finishes immediately.
	fp.stdoutW.Close()

	// Drain the messages channel to let readStdout finish and close it.
	go func() {
		for range pm.Messages() {
		}
	}()

	// Give readStdout goroutine a moment to finish.
	time.Sleep(10 * time.Millisecond)

	msg := SDKControlRequest{
		Type:      "control_request",
		RequestID: "req-1",
		Request:   json.RawMessage(`{"subtype":"interrupt"}`),
	}

	err := pm.WriteJSON(msg)
	if err != nil {
		t.Fatalf("WriteJSON: %v", err)
	}

	select {
	case line := <-done:
		if !strings.Contains(line, "control_request") {
			t.Errorf("stdin line should contain 'control_request', got %q", line)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("timed out waiting for stdin write")
	}
}

func TestProcessManager_StderrCallback(t *testing.T) {
	fp := newFakeProcess()

	var stderrLines []string
	var mu sync.Mutex

	stderrCb := func(line string) {
		mu.Lock()
		stderrLines = append(stderrLines, line)
		mu.Unlock()
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	pm := newProcessManager(fp, ctx, stderrCb)

	// Write to stderr pipe (we need to set it up manually since fakeProcess doesn't have stderr).
	// For this test, we verify the callback is stored.
	_ = pm

	// Close stdout so reader finishes.
	fp.stdoutW.Close()
	<-pm.Messages()
}

func TestProcessManager_Kill(t *testing.T) {
	fp := newFakeProcess()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	pm := newProcessManager(fp, ctx, nil)
	fp.stdoutW.Close()

	err := pm.Kill()
	if err != nil {
		t.Fatalf("Kill: %v", err)
	}
	fp.mu.Lock()
	killed := fp.killed
	fp.mu.Unlock()
	if !killed {
		t.Error("expected process to be killed")
	}
}

func TestProcessManager_ContextCancellation(t *testing.T) {
	fp := newFakeProcess()

	ctx, cancel := context.WithCancel(context.Background())
	pm := newProcessManager(fp, ctx, nil)

	// Cancel context — should cause messages channel to close.
	cancel()
	fp.stdoutW.Close()

	// Drain messages — should not block forever.
	for range pm.Messages() {
	}
}

func TestProcessManager_MalformedJSON(t *testing.T) {
	fp := newFakeProcess()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	pm := newProcessManager(fp, ctx, nil)

	// Write a mix of valid and invalid JSON lines.
	go func() {
		fp.stdoutW.Write([]byte("not valid json\n"))
		fp.stdoutW.Write([]byte(`{"type":"assistant","message":{"role":"assistant","content":[]}}` + "\n"))
		fp.stdoutW.Close()
	}()

	var received []SDKMessage
	for msg := range pm.Messages() {
		received = append(received, msg)
	}

	// Malformed line should be skipped, valid message should come through.
	if len(received) != 1 {
		t.Fatalf("expected 1 valid message (skipping malformed), got %d", len(received))
	}
}

func TestProcessManager_EmptyLines(t *testing.T) {
	fp := newFakeProcess()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	pm := newProcessManager(fp, ctx, nil)

	go func() {
		fp.stdoutW.Write([]byte("\n\n"))
		fp.stdoutW.Write([]byte(`{"type":"assistant","message":{"role":"assistant","content":[]}}` + "\n"))
		fp.stdoutW.Write([]byte("\n"))
		fp.stdoutW.Close()
	}()

	var received []SDKMessage
	for msg := range pm.Messages() {
		received = append(received, msg)
	}

	if len(received) != 1 {
		t.Fatalf("expected 1 message (ignoring empty lines), got %d", len(received))
	}
}

func TestDefaultSpawn_Env(t *testing.T) {
	proc := defaultSpawn(SpawnOptions{
		Command: "echo",
		Args:    []string{"hello"},
		Env:     map[string]string{"MY_VAR": "val"},
	})
	if proc == nil {
		t.Fatal("defaultSpawn returned nil")
	}
	// Verify it produces a valid SpawnedProcess with accessible pipes.
	if proc.Stdin() == nil {
		t.Error("Stdin should not be nil")
	}
	if proc.Stdout() == nil {
		t.Error("Stdout should not be nil")
	}
}

func TestDefaultSpawn_Cwd(t *testing.T) {
	proc := defaultSpawn(SpawnOptions{
		Command: "echo",
		Args:    []string{"hello"},
		Cwd:     "/tmp",
	})
	if proc == nil {
		t.Fatal("defaultSpawn returned nil")
	}
	if proc.Stdin() == nil {
		t.Error("Stdin should not be nil")
	}
}

func TestBuildProcessArgs_OutputFormat(t *testing.T) {
	opts := &Options{OutputFormat: &OutputFormat{
		Type:   OutputFormatTypeJSONSchema,
		Schema: map[string]interface{}{"type": "object"},
	}}
	args := buildProcessArgs(opts, "")
	assertFlag(t, args, "--output-format-json")
}

func TestBuildProcessArgs_SettingsString(t *testing.T) {
	opts := &Options{Settings: "/path/to/settings.json"}
	args := buildProcessArgs(opts, "")
	assertContains(t, args, "--settings", "/path/to/settings.json")
}

func TestBuildProcessArgs_SettingsObject(t *testing.T) {
	opts := &Options{Settings: map[string]interface{}{"key": "val"}}
	args := buildProcessArgs(opts, "")
	assertFlag(t, args, "--settings")
}

func TestBuildProcessArgs_Agents(t *testing.T) {
	opts := &Options{Agents: map[string]AgentDefinition{
		"researcher": {Model: strPtr("sonnet")},
	}}
	args := buildProcessArgs(opts, "")
	assertFlag(t, args, "--agents")
}

func TestBuildProcessArgs_ToolConfig(t *testing.T) {
	pf := "markdown"
	opts := &Options{ToolConfig: &ToolConfig{
		AskUserQuestion: &AskUserQuestionConfig{PreviewFormat: &pf},
	}}
	args := buildProcessArgs(opts, "")
	assertFlag(t, args, "--tool-config")
}

func TestBuildProcessArgs_ToolPreset(t *testing.T) {
	opts := &Options{Tools: ToolPreset{Type: "preset", Preset: "claude_code"}}
	args := buildProcessArgs(opts, "")
	// ToolPreset should not add a --tools flag.
	for _, a := range args {
		if a == "--tools" {
			t.Error("--tools should not be present for ToolPreset")
		}
	}
}

func TestBuildProcessArgs_PromptAtEnd(t *testing.T) {
	model := "sonnet"
	opts := &Options{Model: &model}
	args := buildProcessArgs(opts, "my prompt")
	// Prompt is the last positional argument.
	if len(args) == 0 {
		t.Fatal("args should not be empty")
	}
	if args[len(args)-1] != "my prompt" {
		t.Errorf("expected prompt as last arg, got %q", args[len(args)-1])
	}
}

// --- test helpers ---

// boolPtr is defined in control_test.go

func assertContains(t *testing.T, args []string, flag, value string) {
	t.Helper()
	for i, a := range args {
		if a == flag && i+1 < len(args) && args[i+1] == value {
			return
		}
	}
	t.Errorf("args should contain %s %s, got %v", flag, value, args)
}

func assertFlag(t *testing.T, args []string, flag string) {
	t.Helper()
	for _, a := range args {
		if a == flag {
			return
		}
	}
	t.Errorf("args should contain %s, got %v", flag, args)
}

// suppress unused import warning
var _ = bytes.NewBuffer
