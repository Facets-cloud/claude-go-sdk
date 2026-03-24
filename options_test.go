package claudeagent

import (
	"encoding/json"
	"testing"
)

func TestThinkingConfig_Adaptive_JSON(t *testing.T) {
	tc := ThinkingAdaptive()
	b, err := json.Marshal(tc)
	if err != nil {
		t.Fatal(err)
	}
	var got ThinkingConfig
	if err := json.Unmarshal(b, &got); err != nil {
		t.Fatal(err)
	}
	if got.Type != "adaptive" {
		t.Errorf("Type = %q, want %q", got.Type, "adaptive")
	}
}

func TestThinkingConfig_Enabled_JSON(t *testing.T) {
	tc := ThinkingEnabled(10000)
	b, err := json.Marshal(tc)
	if err != nil {
		t.Fatal(err)
	}
	var got ThinkingConfig
	if err := json.Unmarshal(b, &got); err != nil {
		t.Fatal(err)
	}
	if got.Type != "enabled" {
		t.Errorf("Type = %q, want %q", got.Type, "enabled")
	}
	if got.BudgetTokens == nil || *got.BudgetTokens != 10000 {
		t.Errorf("BudgetTokens = %v, want 10000", got.BudgetTokens)
	}
}

func TestThinkingConfig_Disabled_JSON(t *testing.T) {
	tc := ThinkingDisabledConfig()
	if tc.Type != "disabled" {
		t.Errorf("Type = %q, want %q", tc.Type, "disabled")
	}
}

func TestToolConfig_JSON(t *testing.T) {
	raw := `{"askUserQuestion":{"previewFormat":"html"}}`
	var tc ToolConfig
	if err := json.Unmarshal([]byte(raw), &tc); err != nil {
		t.Fatal(err)
	}
	if tc.AskUserQuestion == nil {
		t.Fatal("AskUserQuestion is nil")
	}
	if tc.AskUserQuestion.PreviewFormat == nil || *tc.AskUserQuestion.PreviewFormat != "html" {
		t.Errorf("PreviewFormat = %v", tc.AskUserQuestion.PreviewFormat)
	}
}

func TestOutputFormat_JSON(t *testing.T) {
	raw := `{"type":"json_schema","schema":{"type":"object","properties":{"result":{"type":"string"}}}}`
	var of OutputFormat
	if err := json.Unmarshal([]byte(raw), &of); err != nil {
		t.Fatal(err)
	}
	if of.Type != OutputFormatTypeJSONSchema {
		t.Errorf("Type = %q", of.Type)
	}
}

func TestSdkPluginConfig_JSON(t *testing.T) {
	raw := `{"type":"local","path":"./my-plugin"}`
	var p SdkPluginConfig
	if err := json.Unmarshal([]byte(raw), &p); err != nil {
		t.Fatal(err)
	}
	if p.Path != "./my-plugin" {
		t.Errorf("Path = %q", p.Path)
	}
}

func TestElicitationRequest_JSON(t *testing.T) {
	raw := `{"serverName":"test","message":"Enter key","mode":"form"}`
	var r ElicitationRequest
	if err := json.Unmarshal([]byte(raw), &r); err != nil {
		t.Fatal(err)
	}
	if r.ServerName != "test" {
		t.Errorf("ServerName = %q", r.ServerName)
	}
}

func TestOptions_DefaultZeroValue(t *testing.T) {
	// Verify Options can be constructed with zero values
	opts := Options{}
	if opts.Model != nil {
		t.Error("Model should be nil by default")
	}
	if opts.PermissionMode != nil {
		t.Error("PermissionMode should be nil by default")
	}
}
