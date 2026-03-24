package claudeagent

import (
	"encoding/json"
	"testing"
)

func TestPermissionMode_JSON(t *testing.T) {
	tests := []struct {
		mode PermissionMode
		json string
	}{
		{PermissionModeDefault, `"default"`},
		{PermissionModeAcceptEdits, `"acceptEdits"`},
		{PermissionModeBypassPermissions, `"bypassPermissions"`},
		{PermissionModePlan, `"plan"`},
		{PermissionModeDontAsk, `"dontAsk"`},
	}
	for _, tt := range tests {
		b, err := json.Marshal(tt.mode)
		if err != nil {
			t.Fatalf("Marshal(%v): %v", tt.mode, err)
		}
		if string(b) != tt.json {
			t.Errorf("Marshal(%v) = %s, want %s", tt.mode, b, tt.json)
		}
		var got PermissionMode
		if err := json.Unmarshal([]byte(tt.json), &got); err != nil {
			t.Fatalf("Unmarshal(%s): %v", tt.json, err)
		}
		if got != tt.mode {
			t.Errorf("Unmarshal(%s) = %v, want %v", tt.json, got, tt.mode)
		}
	}
}

func TestExitReason_Values(t *testing.T) {
	expected := []ExitReason{
		ExitReasonClear, ExitReasonResume, ExitReasonLogout,
		ExitReasonPromptInputExit, ExitReasonOther, ExitReasonBypassPermissionsDisabled,
	}
	if len(expected) != 6 {
		t.Errorf("expected 6 exit reasons, got %d", len(expected))
	}
}

func TestHookEvent_Values(t *testing.T) {
	events := AllHookEvents()
	if len(events) != 23 {
		t.Errorf("expected 23 hook events, got %d", len(events))
	}
}

func TestPermissionBehavior_JSON(t *testing.T) {
	b, err := json.Marshal(PermissionBehaviorAllow)
	if err != nil {
		t.Fatal(err)
	}
	if string(b) != `"allow"` {
		t.Errorf("got %s, want %q", b, "allow")
	}
}

func TestFastModeState_JSON(t *testing.T) {
	b, err := json.Marshal(FastModeStateOff)
	if err != nil {
		t.Fatal(err)
	}
	if string(b) != `"off"` {
		t.Errorf("got %s, want %q", b, "off")
	}
}
