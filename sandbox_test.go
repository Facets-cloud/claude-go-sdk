package claudeagent

import (
	"encoding/json"
	"testing"
)

func TestSandboxSettings_JSON(t *testing.T) {
	raw := `{
		"enabled": true,
		"autoAllowBashIfSandboxed": true,
		"allowUnsandboxedCommands": false,
		"network": {
			"allowedDomains": ["example.com"],
			"allowManagedDomainsOnly": false,
			"allowUnixSockets": ["/var/run/docker.sock"],
			"allowAllUnixSockets": false,
			"allowLocalBinding": true,
			"httpProxyPort": 8080,
			"socksProxyPort": 1080
		},
		"filesystem": {
			"allowWrite": ["/tmp"],
			"denyWrite": ["/etc"],
			"denyRead": ["/secret"],
			"allowRead": ["/public"],
			"allowManagedReadPathsOnly": false
		},
		"ignoreViolations": {
			"network": ["allowed-violation"]
		},
		"enableWeakerNestedSandbox": false,
		"enableWeakerNetworkIsolation": false,
		"excludedCommands": ["dangerous-cmd"],
		"ripgrep": {
			"command": "/usr/local/bin/rg",
			"args": ["--hidden"]
		}
	}`
	var s SandboxSettings
	if err := json.Unmarshal([]byte(raw), &s); err != nil {
		t.Fatal(err)
	}
	if !*s.Enabled {
		t.Error("expected enabled=true")
	}
	if !*s.AutoAllowBashIfSandboxed {
		t.Error("expected autoAllowBashIfSandboxed=true")
	}
	if s.Network == nil {
		t.Fatal("expected network config")
	}
	if len(s.Network.AllowedDomains) != 1 || s.Network.AllowedDomains[0] != "example.com" {
		t.Errorf("got allowedDomains %v", s.Network.AllowedDomains)
	}
	if *s.Network.HttpProxyPort != 8080 {
		t.Errorf("got httpProxyPort %d", *s.Network.HttpProxyPort)
	}
	if s.Filesystem == nil {
		t.Fatal("expected filesystem config")
	}
	if len(s.Filesystem.AllowWrite) != 1 || s.Filesystem.AllowWrite[0] != "/tmp" {
		t.Errorf("got allowWrite %v", s.Filesystem.AllowWrite)
	}
	if len(s.IgnoreViolations) != 1 {
		t.Errorf("got %d ignore violations", len(s.IgnoreViolations))
	}
	if len(s.ExcludedCommands) != 1 {
		t.Errorf("got %d excluded commands", len(s.ExcludedCommands))
	}
	if s.Ripgrep == nil || s.Ripgrep.Command != "/usr/local/bin/rg" {
		t.Error("expected ripgrep config")
	}

	// Round-trip
	data, err := json.Marshal(s)
	if err != nil {
		t.Fatal(err)
	}
	var s2 SandboxSettings
	if err := json.Unmarshal(data, &s2); err != nil {
		t.Fatal(err)
	}
	if !*s2.Enabled {
		t.Error("round-trip: expected enabled=true")
	}
}

func TestSandboxNetworkConfig_Empty(t *testing.T) {
	raw := `{}`
	var n SandboxNetworkConfig
	if err := json.Unmarshal([]byte(raw), &n); err != nil {
		t.Fatal(err)
	}
	if n.AllowedDomains != nil {
		t.Error("expected nil allowedDomains")
	}
}

func TestSandboxFilesystemConfig_Empty(t *testing.T) {
	raw := `{}`
	var f SandboxFilesystemConfig
	if err := json.Unmarshal([]byte(raw), &f); err != nil {
		t.Fatal(err)
	}
	if f.AllowWrite != nil {
		t.Error("expected nil allowWrite")
	}
}
