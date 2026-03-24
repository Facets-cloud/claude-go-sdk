package claudeagent

import (
	"encoding/json"
	"testing"
)

func TestPermissionResult_Allow_JSON(t *testing.T) {
	r := PermissionResultAllow{
		Behavior: PermissionBehaviorAllow,
	}
	b, err := json.Marshal(r)
	if err != nil {
		t.Fatal(err)
	}
	if string(b) == "" {
		t.Error("empty marshal")
	}
}

func TestPermissionResult_Deny_JSON(t *testing.T) {
	r := PermissionResultDeny{
		Behavior: PermissionBehaviorDeny,
		Message:  "not allowed",
	}
	b, err := json.Marshal(r)
	if err != nil {
		t.Fatal(err)
	}
	var got map[string]interface{}
	if err := json.Unmarshal(b, &got); err != nil {
		t.Fatal(err)
	}
	if got["message"] != "not allowed" {
		t.Errorf("message = %v", got["message"])
	}
}

func TestPermissionUpdate_AddRules_JSON(t *testing.T) {
	raw := `{"type":"addRules","rules":[{"toolName":"Bash","ruleContent":"npm *"}],"behavior":"allow","destination":"session"}`
	var u PermissionUpdateAddRules
	if err := json.Unmarshal([]byte(raw), &u); err != nil {
		t.Fatal(err)
	}
	if u.UpdateType != "addRules" {
		t.Errorf("Type = %q", u.UpdateType)
	}
	if len(u.Rules) != 1 {
		t.Errorf("Rules = %v", u.Rules)
	}
}

func TestPermissionUpdate_SetMode_JSON(t *testing.T) {
	raw := `{"type":"setMode","mode":"bypassPermissions","destination":"session"}`
	var u PermissionUpdateSetMode
	if err := json.Unmarshal([]byte(raw), &u); err != nil {
		t.Fatal(err)
	}
	if u.Mode != PermissionModeBypassPermissions {
		t.Errorf("Mode = %q", u.Mode)
	}
}

func TestPermissionUpdate_AddDirectories_JSON(t *testing.T) {
	raw := `{"type":"addDirectories","directories":["/tmp","/home"],"destination":"userSettings"}`
	var u PermissionUpdateAddDirectories
	if err := json.Unmarshal([]byte(raw), &u); err != nil {
		t.Fatal(err)
	}
	if len(u.Directories) != 2 {
		t.Errorf("Directories = %v", u.Directories)
	}
}

func TestPermissionRuleValue_JSON(t *testing.T) {
	raw := `{"toolName":"Bash","ruleContent":"npm test"}`
	var r PermissionRuleValue
	if err := json.Unmarshal([]byte(raw), &r); err != nil {
		t.Fatal(err)
	}
	if r.ToolName != "Bash" {
		t.Errorf("ToolName = %q", r.ToolName)
	}
}

func TestPermissionResult_Interface(t *testing.T) {
	var _ PermissionResult = PermissionResultAllow{}
	var _ PermissionResult = PermissionResultDeny{}
}

func TestPermissionUpdate_Interface(t *testing.T) {
	var _ PermissionUpdate = PermissionUpdateAddRules{}
	var _ PermissionUpdate = PermissionUpdateReplaceRules{}
	var _ PermissionUpdate = PermissionUpdateRemoveRules{}
	var _ PermissionUpdate = PermissionUpdateSetMode{}
	var _ PermissionUpdate = PermissionUpdateAddDirectories{}
	var _ PermissionUpdate = PermissionUpdateRemoveDirectories{}
}
