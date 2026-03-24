package claudeagent

import (
	"encoding/json"
	"testing"
)

func TestSettings_BasicFields_JSON(t *testing.T) {
	raw := `{
		"$schema": "https://json.schemastore.org/claude-code-settings.json",
		"model": "claude-sonnet-4-6",
		"permissions": {
			"allow": ["Bash(*)"],
			"deny": ["Write(/etc/*)"],
			"defaultMode": "default"
		},
		"env": {"MY_VAR": "value"},
		"includeCoAuthoredBy": true,
		"includeGitInstructions": true,
		"cleanupPeriodDays": 30,
		"respectGitignore": true
	}`
	var s Settings
	if err := json.Unmarshal([]byte(raw), &s); err != nil {
		t.Fatal(err)
	}
	if *s.Model != "claude-sonnet-4-6" {
		t.Errorf("got model %q", *s.Model)
	}
	if s.Permissions == nil {
		t.Fatal("expected permissions")
	}
	if len(s.Permissions.Allow) != 1 || s.Permissions.Allow[0] != "Bash(*)" {
		t.Errorf("got allow %v", s.Permissions.Allow)
	}
	if s.Env == nil || s.Env["MY_VAR"] != "value" {
		t.Error("expected env")
	}
	if *s.CleanupPeriodDays != 30 {
		t.Errorf("got cleanupPeriodDays %d", *s.CleanupPeriodDays)
	}
}

func TestSettings_Hooks_JSON(t *testing.T) {
	raw := `{
		"hooks": {
			"PreToolUse": [{
				"matcher": "Write",
				"hooks": [{
					"type": "command",
					"command": "echo hello",
					"shell": "bash",
					"timeout": 10,
					"once": false,
					"async": false
				}]
			}]
		}
	}`
	var s Settings
	if err := json.Unmarshal([]byte(raw), &s); err != nil {
		t.Fatal(err)
	}
	if s.Hooks == nil {
		t.Fatal("expected hooks")
	}
	matchers, ok := s.Hooks["PreToolUse"]
	if !ok || len(matchers) != 1 {
		t.Fatal("expected 1 PreToolUse matcher")
	}
	if *matchers[0].Matcher != "Write" {
		t.Errorf("got matcher %q", *matchers[0].Matcher)
	}
	if len(matchers[0].Hooks) != 1 {
		t.Fatal("expected 1 hook")
	}
	h := matchers[0].Hooks[0]
	if h.Type != "command" {
		t.Errorf("got hook type %q", h.Type)
	}
	if *h.Command != "echo hello" {
		t.Errorf("got hook command %q", *h.Command)
	}
}

func TestSettings_HookTypes_JSON(t *testing.T) {
	tests := []struct {
		name string
		raw  string
		typ  string
	}{
		{
			"prompt hook",
			`{"type":"prompt","prompt":"check $ARGUMENTS","model":"claude-sonnet-4-6"}`,
			"prompt",
		},
		{
			"agent hook",
			`{"type":"agent","prompt":"verify tests passed","timeout":60}`,
			"agent",
		},
		{
			"http hook",
			`{"type":"http","url":"https://hooks.example.com/api","headers":{"Authorization":"Bearer token"}}`,
			"http",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var h SettingsHookConfig
			if err := json.Unmarshal([]byte(tt.raw), &h); err != nil {
				t.Fatal(err)
			}
			if h.Type != tt.typ {
				t.Errorf("got type %q, want %q", h.Type, tt.typ)
			}
		})
	}
}

func TestSettings_Sandbox_JSON(t *testing.T) {
	raw := `{
		"sandbox": {
			"enabled": true,
			"autoAllowBashIfSandboxed": true,
			"network": {
				"allowedDomains": ["*.example.com"]
			}
		}
	}`
	var s Settings
	if err := json.Unmarshal([]byte(raw), &s); err != nil {
		t.Fatal(err)
	}
	if s.Sandbox == nil || !*s.Sandbox.Enabled {
		t.Error("expected sandbox enabled")
	}
}

func TestSettings_Marketplace_JSON(t *testing.T) {
	raw := `{
		"extraKnownMarketplaces": {
			"my-marketplace": {
				"source": {
					"source": "github",
					"repo": "org/repo"
				},
				"autoUpdate": true
			}
		}
	}`
	var s Settings
	if err := json.Unmarshal([]byte(raw), &s); err != nil {
		t.Fatal(err)
	}
	if s.ExtraKnownMarketplaces == nil {
		t.Fatal("expected extraKnownMarketplaces")
	}
	mp, ok := s.ExtraKnownMarketplaces["my-marketplace"]
	if !ok {
		t.Fatal("expected my-marketplace")
	}
	if mp.AutoUpdate == nil || !*mp.AutoUpdate {
		t.Error("expected autoUpdate true")
	}
}

func TestSettings_AllEnterpriseFields(t *testing.T) {
	raw := `{
		"allowManagedHooksOnly": true,
		"allowManagedPermissionRulesOnly": true,
		"allowManagedMcpServersOnly": true,
		"disableAllHooks": true,
		"forceLoginMethod": "console",
		"forceLoginOrgUUID": "org-123"
	}`
	var s Settings
	if err := json.Unmarshal([]byte(raw), &s); err != nil {
		t.Fatal(err)
	}
	if !*s.AllowManagedHooksOnly {
		t.Error("expected allowManagedHooksOnly")
	}
	if !*s.AllowManagedPermissionRulesOnly {
		t.Error("expected allowManagedPermissionRulesOnly")
	}
	if !*s.AllowManagedMcpServersOnly {
		t.Error("expected allowManagedMcpServersOnly")
	}
	if !*s.DisableAllHooks {
		t.Error("expected disableAllHooks")
	}
	if *s.ForceLoginMethod != "console" {
		t.Errorf("got forceLoginMethod %q", *s.ForceLoginMethod)
	}
}

func TestSettings_UIAndBehavior(t *testing.T) {
	raw := `{
		"syntaxHighlightingDisabled": true,
		"terminalTitleFromRename": false,
		"alwaysThinkingEnabled": true,
		"effortLevel": "high",
		"fastMode": true,
		"fastModePerSessionOptIn": true,
		"promptSuggestionEnabled": false,
		"showClearContextOnPlanAccept": true,
		"spinnerTipsEnabled": true,
		"feedbackSurveyRate": 0.05,
		"prefersReducedMotion": true,
		"autoMemoryEnabled": true,
		"autoDreamEnabled": false,
		"showThinkingSummaries": true,
		"disableAutoMode": "disable"
	}`
	var s Settings
	if err := json.Unmarshal([]byte(raw), &s); err != nil {
		t.Fatal(err)
	}
	if !*s.SyntaxHighlightingDisabled {
		t.Error("expected syntaxHighlightingDisabled")
	}
	if *s.TerminalTitleFromRename {
		t.Error("expected terminalTitleFromRename false")
	}
	if *s.EffortLevel != "high" {
		t.Errorf("got effortLevel %q", *s.EffortLevel)
	}
	if !*s.FastMode {
		t.Error("expected fastMode")
	}
	if *s.FeedbackSurveyRate != 0.05 {
		t.Errorf("got feedbackSurveyRate %f", *s.FeedbackSurveyRate)
	}
	if *s.DisableAutoMode != "disable" {
		t.Errorf("got disableAutoMode %q", *s.DisableAutoMode)
	}
}

func TestSettings_RoundTrip(t *testing.T) {
	s := Settings{
		Model: strPtr("claude-opus-4-6"),
		Permissions: &SettingsPermissions{
			Allow:       []string{"Read(*)"},
			DefaultMode: strPtr("default"),
		},
		Sandbox: &SandboxSettings{
			Enabled: boolPtr(true),
		},
	}
	data, err := json.Marshal(s)
	if err != nil {
		t.Fatal(err)
	}
	var s2 Settings
	if err := json.Unmarshal(data, &s2); err != nil {
		t.Fatal(err)
	}
	if *s2.Model != "claude-opus-4-6" {
		t.Errorf("round-trip model: %q", *s2.Model)
	}
	if s2.Permissions == nil || len(s2.Permissions.Allow) != 1 {
		t.Error("round-trip permissions failed")
	}
	if s2.Sandbox == nil || !*s2.Sandbox.Enabled {
		t.Error("round-trip sandbox failed")
	}
}

func TestSettings_SSHConfigs(t *testing.T) {
	raw := `{
		"sshConfigs": [{
			"id": "ssh-1",
			"name": "Dev Server",
			"sshHost": "dev@example.com",
			"sshPort": 2222,
			"startDirectory": "~/projects"
		}]
	}`
	var s Settings
	if err := json.Unmarshal([]byte(raw), &s); err != nil {
		t.Fatal(err)
	}
	if len(s.SSHConfigs) != 1 {
		t.Fatalf("got %d ssh configs", len(s.SSHConfigs))
	}
	if s.SSHConfigs[0].ID != "ssh-1" {
		t.Errorf("got id %q", s.SSHConfigs[0].ID)
	}
	if *s.SSHConfigs[0].SSHPort != 2222 {
		t.Errorf("got sshPort %d", *s.SSHConfigs[0].SSHPort)
	}
}
