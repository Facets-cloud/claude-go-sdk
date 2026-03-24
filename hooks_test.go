package claudeagent

import (
	"encoding/json"
	"testing"
)

func TestBaseHookInput_JSON(t *testing.T) {
	raw := `{"session_id":"s1","transcript_path":"/tmp/transcript","cwd":"/home","permission_mode":"default","agent_id":"a1","agent_type":"general"}`
	var b BaseHookInput
	if err := json.Unmarshal([]byte(raw), &b); err != nil {
		t.Fatal(err)
	}
	if b.SessionID != "s1" {
		t.Errorf("SessionID = %q", b.SessionID)
	}
	if b.Cwd != "/home" {
		t.Errorf("Cwd = %q", b.Cwd)
	}
}

func TestPreToolUseHookInput_JSON(t *testing.T) {
	raw := `{"session_id":"s1","transcript_path":"/tmp/t","cwd":"/","hook_event_name":"PreToolUse","tool_name":"Bash","tool_input":{"command":"ls"},"tool_use_id":"tu1"}`
	var h PreToolUseHookInput
	if err := json.Unmarshal([]byte(raw), &h); err != nil {
		t.Fatal(err)
	}
	if h.ToolName != "Bash" {
		t.Errorf("ToolName = %q", h.ToolName)
	}
	if h.ToolUseID != "tu1" {
		t.Errorf("ToolUseID = %q", h.ToolUseID)
	}
}

func TestPostToolUseHookInput_JSON(t *testing.T) {
	raw := `{"session_id":"s1","transcript_path":"/tmp/t","cwd":"/","hook_event_name":"PostToolUse","tool_name":"Read","tool_input":{},"tool_response":"content","tool_use_id":"tu2"}`
	var h PostToolUseHookInput
	if err := json.Unmarshal([]byte(raw), &h); err != nil {
		t.Fatal(err)
	}
	if h.ToolName != "Read" {
		t.Errorf("ToolName = %q", h.ToolName)
	}
}

func TestSessionStartHookInput_JSON(t *testing.T) {
	raw := `{"session_id":"s1","transcript_path":"/tmp/t","cwd":"/","hook_event_name":"SessionStart","source":"startup","model":"claude-sonnet-4-6"}`
	var h SessionStartHookInput
	if err := json.Unmarshal([]byte(raw), &h); err != nil {
		t.Fatal(err)
	}
	if h.Source != "startup" {
		t.Errorf("Source = %q", h.Source)
	}
}

func TestSessionEndHookInput_JSON(t *testing.T) {
	raw := `{"session_id":"s1","transcript_path":"/tmp/t","cwd":"/","hook_event_name":"SessionEnd","reason":"clear"}`
	var h SessionEndHookInput
	if err := json.Unmarshal([]byte(raw), &h); err != nil {
		t.Fatal(err)
	}
	if h.Reason != ExitReasonClear {
		t.Errorf("Reason = %q", h.Reason)
	}
}

func TestElicitationHookInput_JSON(t *testing.T) {
	raw := `{"session_id":"s1","transcript_path":"/tmp/t","cwd":"/","hook_event_name":"Elicitation","mcp_server_name":"srv","message":"Enter API key","mode":"form"}`
	var h ElicitationHookInput
	if err := json.Unmarshal([]byte(raw), &h); err != nil {
		t.Fatal(err)
	}
	if h.McpServerName != "srv" {
		t.Errorf("McpServerName = %q", h.McpServerName)
	}
}

func TestInstructionsLoadedHookInput_JSON(t *testing.T) {
	raw := `{"session_id":"s1","transcript_path":"/tmp/t","cwd":"/","hook_event_name":"InstructionsLoaded","file_path":"/project/CLAUDE.md","memory_type":"Project","load_reason":"session_start"}`
	var h InstructionsLoadedHookInput
	if err := json.Unmarshal([]byte(raw), &h); err != nil {
		t.Fatal(err)
	}
	if h.MemoryType != "Project" {
		t.Errorf("MemoryType = %q", h.MemoryType)
	}
}

func TestSyncHookJSONOutput_JSON(t *testing.T) {
	cont := true
	o := SyncHookJSONOutput{
		Continue:   &cont,
		StopReason: strPtr("done"),
	}
	b, err := json.Marshal(o)
	if err != nil {
		t.Fatal(err)
	}
	var got map[string]interface{}
	if err := json.Unmarshal(b, &got); err != nil {
		t.Fatal(err)
	}
	if got["continue"] != true {
		t.Errorf("continue = %v", got["continue"])
	}
}

func TestAsyncHookJSONOutput_JSON(t *testing.T) {
	o := AsyncHookJSONOutput{
		Async:        true,
		AsyncTimeout: intPtr(30),
	}
	b, err := json.Marshal(o)
	if err != nil {
		t.Fatal(err)
	}
	var got map[string]interface{}
	if err := json.Unmarshal(b, &got); err != nil {
		t.Fatal(err)
	}
	if got["async"] != true {
		t.Errorf("async = %v", got["async"])
	}
}

func TestHookJSONOutput_Interface(t *testing.T) {
	var _ HookJSONOutput = SyncHookJSONOutput{}
	var _ HookJSONOutput = AsyncHookJSONOutput{}
}

func TestAllHookInputTypes_Compile(t *testing.T) {
	// Verify all 23 hook input types exist and compile
	_ = PreToolUseHookInput{}
	_ = PostToolUseHookInput{}
	_ = PostToolUseFailureHookInput{}
	_ = NotificationHookInput{}
	_ = UserPromptSubmitHookInput{}
	_ = SessionStartHookInput{}
	_ = SessionEndHookInput{}
	_ = StopHookInput{}
	_ = StopFailureHookInput{}
	_ = SubagentStartHookInput{}
	_ = SubagentStopHookInput{}
	_ = PreCompactHookInput{}
	_ = PostCompactHookInput{}
	_ = PermissionRequestHookInput{}
	_ = SetupHookInput{}
	_ = TeammateIdleHookInput{}
	_ = TaskCompletedHookInput{}
	_ = ElicitationHookInput{}
	_ = ElicitationResultHookInput{}
	_ = ConfigChangeHookInput{}
	_ = InstructionsLoadedHookInput{}
	_ = WorktreeCreateHookInput{}
	_ = WorktreeRemoveHookInput{}
}

func TestAllHookSpecificOutputs_Compile(t *testing.T) {
	_ = PreToolUseHookSpecificOutput{}
	_ = PostToolUseHookSpecificOutput{}
	_ = PostToolUseFailureHookSpecificOutput{}
	_ = NotificationHookSpecificOutput{}
	_ = UserPromptSubmitHookSpecificOutput{}
	_ = SessionStartHookSpecificOutput{}
	_ = SetupHookSpecificOutput{}
	_ = SubagentStartHookSpecificOutput{}
	_ = PermissionRequestHookSpecificOutput{}
	_ = ElicitationHookSpecificOutput{}
	_ = ElicitationResultHookSpecificOutput{}
}

// helpers — strPtr and intPtr defined in control_test.go
